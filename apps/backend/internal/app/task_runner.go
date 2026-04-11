package app

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/verin/dms/apps/backend/internal/dbgen"
	"github.com/verin/dms/apps/backend/internal/jobs"
)

type WorkspaceSnapshot struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	MemberCount int    `json:"memberCount"`
}

type TaskRunner struct {
	server      *Server
	concurrency chan struct{}
	wg          sync.WaitGroup
}

func NewTaskRunner(server *Server, concurrency int) *TaskRunner {
	if concurrency < 1 {
		concurrency = 1
	}
	return &TaskRunner{
		server:      server,
		concurrency: make(chan struct{}, concurrency),
	}
}

func (r *TaskRunner) Enqueue(ctx context.Context, jobType string, payload any, version any) (string, error) {
	taskID := uuid.NewString()
	versionID := pgtype.UUID{}
	if cast, ok := version.(pgtype.UUID); ok {
		versionID = cast
	}

	_, err := r.server.Queries.CreateJob(ctx, dbgen.CreateJobParams{
		DocumentVersionID: versionID,
		JobType:           jobType,
		TaskID:            taskID,
		Status:            "queued",
		PayloadJson:       JSONBytes(payload),
	})
	if err != nil {
		return "", fmt.Errorf("create job row: %w", err)
	}

	r.start(taskID, jobType, payload)
	return taskID, nil
}

func (r *TaskRunner) ResumePending(ctx context.Context) {
	items, err := r.server.Queries.ListJobs(ctx, 200)
	if err != nil {
		r.server.Logger.Warn().Err(err).Msg("resume pending jobs")
		return
	}

	for _, item := range items {
		if item.Status != "queued" && item.Status != "processing" {
			continue
		}
		var payload any
		if err := json.Unmarshal(item.PayloadJson, &payload); err != nil {
			r.fail(item.TaskID, 0, err)
			continue
		}
		r.start(item.TaskID, item.JobType, payload)
	}
}

func (r *TaskRunner) start(taskID string, jobType string, payload any) {
	r.wg.Add(1)
	go func() {
		defer r.wg.Done()
		defer func() {
			if rec := recover(); rec != nil {
				r.server.Logger.Error().Interface("panic", rec).Str("task_id", taskID).Msg("task runner recovered from panic")
				r.fail(taskID, 1, fmt.Errorf("panic: %v", rec))
			}
		}()
		r.concurrency <- struct{}{}
		defer func() { <-r.concurrency }()

		if err := r.server.Queries.UpdateJobStatus(context.Background(), dbgen.UpdateJobStatusParams{
			TaskID:       taskID,
			Status:       "processing",
			AttemptCount: 0,
			ErrorMessage: pgtype.Text{},
		}); err != nil {
			r.server.Logger.Warn().Err(err).Str("task_id", taskID).Msg("mark job processing")
		}

		if err := r.execute(context.Background(), taskID, jobType, payload); err != nil {
			r.fail(taskID, 1, err)
			return
		}

		if err := r.server.Queries.UpdateJobStatus(context.Background(), dbgen.UpdateJobStatusParams{
			TaskID:       taskID,
			Status:       "completed",
			AttemptCount: 1,
			ErrorMessage: pgtype.Text{},
		}); err != nil {
			r.server.Logger.Warn().Err(err).Str("task_id", taskID).Msg("mark job completed")
		}
	}()
}

func (r *TaskRunner) fail(taskID string, attemptCount int32, err error) {
	r.server.Logger.Warn().Err(err).Str("task_id", taskID).Msg("background task failed")
	_ = r.server.Queries.UpdateJobStatus(context.Background(), dbgen.UpdateJobStatusParams{
		TaskID:       taskID,
		Status:       "failed",
		AttemptCount: attemptCount,
		ErrorMessage: pgtype.Text{
			String: strings.TrimSpace(err.Error()),
			Valid:  true,
		},
	})
}

func (r *TaskRunner) execute(ctx context.Context, taskID string, jobType string, payload any) error {
	switch jobType {
	case jobs.TypeOCR:
		data, err := decodePayload[jobs.OCRPayload](payload)
		if err != nil {
			return err
		}
		return r.server.processOCR(ctx, taskID, data)
	case jobs.TypePreview:
		data, err := decodePayload[jobs.PreviewPayload](payload)
		if err != nil {
			return err
		}
		return r.server.processPreview(ctx, taskID, data)
	case jobs.TypeNotify:
		data, err := decodePayload[jobs.NotificationPayload](payload)
		if err != nil {
			return err
		}
		return r.server.processNotification(ctx, taskID, data)
	case jobs.TypeAuditExport:
		data, err := decodePayload[jobs.AuditExportPayload](payload)
		if err != nil {
			return err
		}
		return r.server.processAuditExport(ctx, taskID, data)
	case jobs.TypeOrphanClean:
		return r.server.processOrphanCleanup(ctx, taskID)
	default:
		return fmt.Errorf("unsupported job type %s", jobType)
	}
}

func decodePayload[T any](payload any) (T, error) {
	var result T
	body, err := json.Marshal(payload)
	if err != nil {
		return result, err
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return result, err
	}
	return result, nil
}

func (r *TaskRunner) Wait(timeout time.Duration) bool {
	done := make(chan struct{})
	go func() {
		defer close(done)
		r.wg.Wait()
	}()

	select {
	case <-done:
		return true
	case <-time.After(timeout):
		return false
	}
}
