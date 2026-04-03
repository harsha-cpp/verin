package app

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/smtp"
	"os"
	"strings"
	"time"

	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgtype"
	pdflib "github.com/ledongthuc/pdf"

	"github.com/verin/dms/apps/backend/internal/dbgen"
	"github.com/verin/dms/apps/backend/internal/jobs"
)

func (s *Server) WorkerMux() *asynq.ServeMux {
	mux := asynq.NewServeMux()
	mux.HandleFunc(jobs.TypeOCR, s.handleOCRTask)
	mux.HandleFunc(jobs.TypePreview, s.handlePreviewTask)
	mux.HandleFunc(jobs.TypeNotify, s.handleNotificationTask)
	mux.HandleFunc(jobs.TypeAuditExport, s.handleAuditExportTask)
	mux.HandleFunc(jobs.TypeRetention, s.handleRetentionSweepTask)
	mux.HandleFunc(jobs.TypeOrphanClean, s.handleOrphanCleanupTask)
	return mux
}

func (s *Server) handleOCRTask(ctx context.Context, task *asynq.Task) error {
	var payload jobs.OCRPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return err
	}

	versionID := MustPGUUID(payload.DocumentVersionID)
	document, err := s.Queries.GetDocumentByVersionID(ctx, versionID)
	if err != nil {
		return err
	}
	version, err := s.Queries.GetDocumentVersionByID(ctx, versionID)
	if err != nil {
		return err
	}

	taskID, _ := asynq.GetTaskID(ctx)
	retryCount, _ := asynq.GetRetryCount(ctx)
	_ = s.Queries.UpdateJobStatus(ctx, dbgen.UpdateJobStatusParams{
		TaskID:       taskID,
		Status:       "processing",
		AttemptCount: int32(retryCount),
		ErrorMessage: pgtype.Text{},
	})

	object, err := s.Storage.GetObject(ctx, version.StorageKey)
	if err != nil {
		return s.failJob(ctx, task, err)
	}
	defer object.Close()

	content, err := io.ReadAll(object)
	if err != nil {
		return s.failJob(ctx, task, err)
	}

	var extractedText string
	if strings.HasPrefix(version.MimeType, "application/pdf") {
		extractedText, err = extractPDFText(content)
		if err != nil {
			s.Logger.Warn().Err(err).Msg("pdf text extraction failed, storing empty")
			extractedText = ""
		}
	} else if strings.HasPrefix(version.MimeType, "text/") {
		extractedText = string(content)
	}

	confidence := pgtype.Numeric{}
	_ = confidence.Scan("90")
	_, err = s.Queries.UpsertOCRText(ctx, dbgen.UpsertOCRTextParams{
		DocumentVersionID: versionID,
		Content:           extractedText,
		ExtractionStatus:  "completed",
		ExtractedAt:       timestamptz(time.Now()),
		Language:          "en",
		ConfidenceScore:   confidence,
	})
	if err != nil {
		return s.failJob(ctx, task, err)
	}

	_ = s.Queries.SetDocumentStatus(ctx, dbgen.SetDocumentStatusParams{
		ID:     document.ID,
		Status: "ready",
	})
	_ = s.Queries.UpdateJobStatus(ctx, dbgen.UpdateJobStatusParams{
		TaskID:       taskID,
		Status:       "completed",
		AttemptCount: int32(retryCount),
		ErrorMessage: pgtype.Text{},
	})
	return nil
}

func (s *Server) handlePreviewTask(ctx context.Context, task *asynq.Task) error {
	var payload jobs.PreviewPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return err
	}
	versionID := MustPGUUID(payload.DocumentVersionID)
	document, err := s.Queries.GetDocumentByVersionID(ctx, versionID)
	if err != nil {
		return s.failJob(ctx, task, err)
	}
	if err := s.Queries.UpsertPreviewAsset(ctx, dbgen.UpsertPreviewAssetParams{
		DocumentVersionID: versionID,
		AssetType:         "preview",
		StorageKey:        payload.StorageKey,
		Status:            "ready",
	}); err != nil {
		return s.failJob(ctx, task, err)
	}
	_ = s.Queries.SetDocumentStatus(ctx, dbgen.SetDocumentStatusParams{
		ID:     document.ID,
		Status: "ready",
	})
	_ = s.Queries.UpdateJobStatus(ctx, dbgen.UpdateJobStatusParams{
		TaskID:       mustTaskID(ctx),
		Status:       "completed",
		AttemptCount: int32(mustRetryCount(ctx)),
		ErrorMessage: pgtype.Text{},
	})
	return nil
}

func (s *Server) handleNotificationTask(ctx context.Context, task *asynq.Task) error {
	var payload jobs.NotificationPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return err
	}
	userID := MustPGUUID(payload.UserID)
	user, err := s.Queries.GetUserByID(ctx, userID)
	if err != nil {
		return s.failJob(ctx, task, err)
	}
	_, err = s.Queries.CreateNotification(ctx, dbgen.CreateNotificationParams{
		OrgID:       user.OrgID,
		UserID:      userID,
		Kind:        payload.Kind,
		Title:       payload.Title,
		Body:        payload.Body,
		PayloadJson: JSONBytes(payload.Data),
	})
	if err != nil {
		return s.failJob(ctx, task, err)
	}

	if s.Config.SMTPHost != "" {
		go s.sendEmailNotification(user.Email, payload.Title, payload.Body)
	}

	_ = s.Queries.UpdateJobStatus(ctx, dbgen.UpdateJobStatusParams{
		TaskID:       mustTaskID(ctx),
		Status:       "completed",
		AttemptCount: int32(mustRetryCount(ctx)),
		ErrorMessage: pgtype.Text{},
	})
	return nil
}

func (s *Server) sendEmailNotification(to string, subject string, body string) {
	from := s.Config.SMTPFromAddress
	addr := fmt.Sprintf("%s:%s", s.Config.SMTPHost, s.Config.SMTPPort)

	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nContent-Type: text/plain; charset=utf-8\r\n\r\n%s",
		from, to, subject, body)

	if err := smtp.SendMail(addr, nil, from, []string{to}, []byte(msg)); err != nil {
		s.Logger.Warn().Err(err).Str("to", to).Msg("failed to send email notification")
	}
}

func (s *Server) handleOrphanCleanupTask(ctx context.Context, task *asynq.Task) error {
	orphans, err := s.Queries.ListOrphanUploads(ctx, 200)
	if err != nil {
		return s.failJob(ctx, task, err)
	}

	for _, orphan := range orphans {
		_ = s.Storage.DeleteObject(ctx, orphan.ObjectKey)
		_ = s.Queries.DeleteUpload(ctx, orphan.ID)
	}

	_ = s.Queries.UpdateJobStatus(ctx, dbgen.UpdateJobStatusParams{
		TaskID:       mustTaskID(ctx),
		Status:       "completed",
		AttemptCount: int32(mustRetryCount(ctx)),
		ErrorMessage: pgtype.Text{},
	})
	return nil
}

func (s *Server) handleAuditExportTask(ctx context.Context, task *asynq.Task) error {
	var payload jobs.AuditExportPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return err
	}
	filePath, err := s.exportAuditCSV(ctx, payload.OrgID)
	if err != nil {
		return s.failJob(ctx, task, err)
	}
	defer os.Remove(filePath)

	file, err := os.Open(filePath)
	if err != nil {
		return s.failJob(ctx, task, err)
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return s.failJob(ctx, task, err)
	}

	exportKey := fmt.Sprintf("orgs/%s/exports/%s.csv", payload.OrgID, jobs.NewID())
	if err := s.Storage.PutObject(ctx, exportKey, file, info.Size(), "text/csv"); err != nil {
		return s.failJob(ctx, task, err)
	}
	_ = s.Queries.UpdateJobStatus(ctx, dbgen.UpdateJobStatusParams{
		TaskID:       mustTaskID(ctx),
		Status:       "completed",
		AttemptCount: int32(mustRetryCount(ctx)),
		ErrorMessage: pgtype.Text{},
	})
	return nil
}

func (s *Server) handleRetentionSweepTask(ctx context.Context, task *asynq.Task) error {
	orgs, err := s.Queries.ListOrganizations(ctx)
	if err != nil {
		return s.failJob(ctx, task, err)
	}

	for _, org := range orgs {
		policies, err := s.Queries.ListRetentionPolicies(ctx, org.ID)
		if err != nil {
			continue
		}
		for _, policy := range policies {
			if !policy.Enabled {
				continue
			}
			if policy.ArchiveAfterDays.Valid && policy.ArchiveAfterDays.Int32 > 0 {
				cutoff := time.Now().AddDate(0, 0, -int(policy.ArchiveAfterDays.Int32))
				_ = s.Queries.ArchiveExpiredDocuments(ctx, dbgen.ArchiveExpiredDocumentsParams{
					OrgID:     org.ID,
					UpdatedAt: pgtype.Timestamptz{Time: cutoff, Valid: true},
				})
			}
		}
	}

	orphans, err := s.Queries.ListOrphanUploads(ctx, 100)
	if err == nil {
		for _, orphan := range orphans {
			_ = s.Storage.DeleteObject(ctx, orphan.ObjectKey)
			_ = s.Queries.DeleteUpload(ctx, orphan.ID)
		}
	}

	_ = s.Queries.UpdateJobStatus(ctx, dbgen.UpdateJobStatusParams{
		TaskID:       mustTaskID(ctx),
		Status:       "completed",
		AttemptCount: int32(mustRetryCount(ctx)),
		ErrorMessage: pgtype.Text{},
	})
	return nil
}

func (s *Server) failJob(ctx context.Context, task *asynq.Task, err error) error {
	_ = s.Queries.UpdateJobStatus(ctx, dbgen.UpdateJobStatusParams{
		TaskID:       mustTaskID(ctx),
		Status:       "failed",
		AttemptCount: int32(mustRetryCount(ctx)),
		ErrorMessage: pgtype.Text{String: err.Error(), Valid: true},
	})
	return err
}

func mustTaskID(ctx context.Context) string {
	taskID, _ := asynq.GetTaskID(ctx)
	return taskID
}

func mustRetryCount(ctx context.Context) int {
	retryCount, _ := asynq.GetRetryCount(ctx)
	return retryCount
}

func extractPDFText(data []byte) (string, error) {
	reader := bytes.NewReader(data)
	pdfReader, err := pdflib.NewReader(reader, int64(len(data)))
	if err != nil {
		return "", fmt.Errorf("open pdf: %w", err)
	}

	var sb strings.Builder
	for i := 1; i <= pdfReader.NumPage(); i++ {
		page := pdfReader.Page(i)
		if page.V.IsNull() {
			continue
		}
		text, err := page.GetPlainText(nil)
		if err != nil {
			continue
		}
		sb.WriteString(text)
		sb.WriteString("\n")
	}
	return strings.TrimSpace(sb.String()), nil
}
