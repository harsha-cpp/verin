package app

import (
	"context"
	"crypto/sha256"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/netip"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/verin/dms/apps/backend/internal/dbgen"
	"github.com/verin/dms/apps/backend/internal/jobs"
	"github.com/verin/dms/apps/backend/internal/storage"
)

type initUploadRequest struct {
	FileName       string `json:"fileName"`
	MimeType       string `json:"mimeType"`
	SizeBytes      int64  `json:"sizeBytes"`
	ChecksumSha256 string `json:"checksumSha256"`
}

type metadataRequest struct {
	SchemaKey    string         `json:"schemaKey"`
	ValueText    string         `json:"valueText"`
	ValueNumber  *float64       `json:"valueNumber"`
	ValueBoolean *bool          `json:"valueBoolean"`
	ValueDate    string         `json:"valueDate"`
	ValueJSON    map[string]any `json:"valueJson"`
}

type completeUploadRequest struct {
	UploadID      string            `json:"uploadId"`
	Title         string            `json:"title"`
	CollectionID  string            `json:"collectionId"`
	ChangeSummary string            `json:"changeSummary"`
	Metadata      []metadataRequest `json:"metadata"`
	Tags          []string          `json:"tags"`
}

type updateDocumentRequest struct {
	Title        string            `json:"title"`
	CollectionID string            `json:"collectionId"`
	Status       string            `json:"status"`
	Metadata     []metadataRequest `json:"metadata"`
	Tags         []string          `json:"tags"`
}

type createCommentRequest struct {
	Body string `json:"body"`
}

type advancedSearchRequest struct {
	Query    string `json:"query"`
	Status   string `json:"status"`
	MimeType string `json:"mimeType"`
	DateFrom string `json:"dateFrom"`
	DateTo   string `json:"dateTo"`
	Limit    int32  `json:"limit"`
	Offset   int32  `json:"offset"`
}

type createSavedSearchRequest struct {
	Name      string         `json:"name"`
	QueryText string         `json:"queryText"`
	Filters   map[string]any `json:"filters"`
}

type assignRolesRequest struct {
	RoleIDs []string `json:"roleIds"`
}

type quotaRequest struct {
	TargetType       string `json:"targetType"`
	TargetID         string `json:"targetId"`
	MaxStorageBytes  int64  `json:"maxStorageBytes"`
	MaxDocumentCount int32  `json:"maxDocumentCount"`
}

type retentionRequest struct {
	ID                    string `json:"id"`
	Name                  string `json:"name"`
	AppliesToCollectionID string `json:"appliesToCollectionId"`
	RetentionDays         int32  `json:"retentionDays"`
	ArchiveAfterDays      *int32 `json:"archiveAfterDays"`
	DeleteAfterDays       *int32 `json:"deleteAfterDays"`
	Enabled               *bool  `json:"enabled"`
}

type settingRequest struct {
	SettingKey   string         `json:"settingKey"`
	SettingValue map[string]any `json:"settingValue"`
}

var allowedMIMEs = map[string]bool{
	"application/pdf": true,
	"image/png":       true,
	"image/jpeg":      true,
	"image/tiff":      true,
	"text/plain":      true,
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document": true,
}

func (s *Server) handleListDocuments(w http.ResponseWriter, r *http.Request) {
	authContext, _ := AuthFromContext(r.Context())
	limit, offset := parsePagination(r)
	searchQuery := strings.TrimSpace(r.URL.Query().Get("q"))

	items, err := s.Queries.ListAccessibleDocuments(r.Context(), dbgen.ListAccessibleDocumentsParams{
		OrgID:       MustPGUUID(authContext.OrgID),
		OwnerUserID: MustPGUUID(authContext.UserID),
		Column3:     authContext.RoleIDs,
		Column4:     authContext.IsAdmin,
		Column5:     searchQuery,
		Limit:       limit,
		Offset:      offset,
	})
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "DOCUMENT_LIST_FAILED", "Could not load documents", nil)
		return
	}

	total, err := s.Queries.CountAccessibleDocuments(r.Context(), dbgen.CountAccessibleDocumentsParams{
		OrgID:       MustPGUUID(authContext.OrgID),
		OwnerUserID: MustPGUUID(authContext.UserID),
		Column3:     authContext.RoleIDs,
		Column4:     authContext.IsAdmin,
		Column5:     searchQuery,
	})
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "DOCUMENT_COUNT_FAILED", "Could not count documents", nil)
		return
	}

	response := make([]map[string]any, 0, len(items))
	for _, item := range items {
		response = append(response, presentDocumentSummaryRow(item))
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"items": response,
		"total": total,
	})
}

func (s *Server) handleInitUpload(w http.ResponseWriter, r *http.Request) {
	authContext, _ := AuthFromContext(r.Context())

	var request initUploadRequest
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, r, http.StatusBadRequest, "INVALID_BODY", "Invalid upload payload", nil)
		return
	}

	if !allowedMIMEs[request.MimeType] {
		writeError(w, r, http.StatusBadRequest, "MIME_NOT_ALLOWED", "Unsupported file type", map[string]any{
			"mimeType": request.MimeType,
		})
		return
	}

	futureDocumentID := uuid.NewString()
	futureVersionID := uuid.NewString()
	objectKey := storage.ObjectKey(authContext.OrgID, futureDocumentID, futureVersionID, "original")
	expiresAt := time.Now().Add(s.Config.SignedURLTTL)

	upload, err := s.Queries.CreateUpload(r.Context(), dbgen.CreateUploadParams{
		OrgID:            MustPGUUID(authContext.OrgID),
		UserID:           MustPGUUID(authContext.UserID),
		ObjectKey:        objectKey,
		OriginalFilename: request.FileName,
		MimeType:         request.MimeType,
		SizeBytes:        request.SizeBytes,
		ChecksumSha256:   request.ChecksumSha256,
		Status:           "pending",
		ExpiresAt:        timestamptz(expiresAt),
	})
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "UPLOAD_INIT_FAILED", "Could not initialize upload", nil)
		return
	}

	signedURL, err := s.Storage.CreateSignedUploadURL(r.Context(), objectKey, s.Config.SignedURLTTL, request.MimeType)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "SIGNED_URL_FAILED", "Could not create signed upload URL", nil)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"upload": map[string]any{
			"id":        UUIDString(upload.ID),
			"uploadUrl": signedURL,
			"objectKey": objectKey,
			"expiresAt": expiresAt.Format(time.RFC3339),
		},
	})
}

const maxUploadSize = 50 << 20

func (s *Server) handleDirectUpload(w http.ResponseWriter, r *http.Request) {
	authContext, _ := AuthFromContext(r.Context())

	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		writeError(w, r, http.StatusBadRequest, "FILE_TOO_LARGE", "File exceeds 50 MB limit", nil)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "NO_FILE", "No file provided", nil)
		return
	}
	defer file.Close()

	title := strings.TrimSpace(r.FormValue("title"))
	if title == "" {
		title = trimExtension(header.Filename)
	}

	mimeType := header.Header.Get("Content-Type")
	if mimeType == "" || mimeType == "application/octet-stream" {
		mimeType = detectMIME(header.Filename)
	}

	if !allowedMIMEs[mimeType] {
		writeError(w, r, http.StatusBadRequest, "MIME_NOT_ALLOWED", "Unsupported file type", map[string]any{"mimeType": mimeType})
		return
	}

	hasher := sha256.New()
	reader := io.TeeReader(file, hasher)

	documentID := uuid.NewString()
	versionID := uuid.NewString()
	objectKey := storage.ObjectKey(authContext.OrgID, documentID, versionID, "original")

	if err := s.Storage.PutObject(r.Context(), objectKey, reader, header.Size, mimeType); err != nil {
		writeError(w, r, http.StatusInternalServerError, "UPLOAD_FAILED", "Could not store file", nil)
		return
	}

	checksum := hex.EncodeToString(hasher.Sum(nil))
	collectionID, _ := optionalUUID(r.FormValue("collectionId"))
	changeSummary := fallback(r.FormValue("changeSummary"), "Initial upload")
	tagsRaw := r.FormValue("tags")
	var tags []string
	if tagsRaw != "" {
		for _, tag := range strings.Split(tagsRaw, ",") {
			if trimmed := strings.TrimSpace(tag); trimmed != "" {
				tags = append(tags, trimmed)
			}
		}
	}

	department := r.FormValue("department")
	var metadata []metadataRequest
	if department != "" {
		metadata = append(metadata, metadataRequest{SchemaKey: "department", ValueText: department})
	}

	pgDocID := MustPGUUID(documentID)
	pgVerID := MustPGUUID(versionID)

	if err := s.withTx(r.Context(), func(ctx context.Context, queries *dbgen.Queries) error {
		document, err := queries.CreateDocument(ctx, dbgen.CreateDocumentParams{
			ID:               pgDocID,
			OrgID:            MustPGUUID(authContext.OrgID),
			Title:            title,
			OriginalFilename: header.Filename,
			MimeType:         mimeType,
			SizeBytes:        header.Size,
			ChecksumSha256:   checksum,
			OwnerUserID:      MustPGUUID(authContext.UserID),
			CollectionID:     collectionID,
			Status:           "processing",
		})
		if err != nil {
			return err
		}

		version, err := queries.CreateDocumentVersion(ctx, dbgen.CreateDocumentVersionParams{
			ID:             pgVerID,
			DocumentID:     document.ID,
			VersionNumber:  1,
			StorageKey:     objectKey,
			SizeBytes:      header.Size,
			ChecksumSha256: checksum,
			MimeType:       mimeType,
			UploadedBy:     MustPGUUID(authContext.UserID),
			ChangeSummary:  changeSummary,
		})
		if err != nil {
			return err
		}

		if err := queries.SetDocumentCurrentVersion(ctx, dbgen.SetDocumentCurrentVersionParams{
			ID:               document.ID,
			CurrentVersionID: version.ID,
			SizeBytes:        version.SizeBytes,
			MimeType:         version.MimeType,
			ChecksumSha256:   version.ChecksumSha256,
		}); err != nil {
			return err
		}

		return saveMetadataAndTags(ctx, queries, document.ID, MustPGUUID(authContext.OrgID), metadata, tags)
	}); err != nil {
		s.Logger.Error().Err(err).Msg("direct-upload transaction failed")
		writeError(w, r, http.StatusInternalServerError, "UPLOAD_FAILED", "Could not create document", nil)
		return
	}

	_ = s.recordAudit(r.Context(), authContext, "document.uploaded", "document", documentID, nil)

	_, _ = s.enqueueJob(r.Context(), jobs.TypeOCR, jobs.OCRPayload{
		DocumentID:        documentID,
		DocumentVersionID: versionID,
	}, versionID)

	s.respondDocumentDetail(w, r, documentID)
}

func detectMIME(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".pdf":
		return "application/pdf"
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".tif", ".tiff":
		return "image/tiff"
	case ".txt":
		return "text/plain"
	case ".docx":
		return "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	default:
		return "application/octet-stream"
	}
}

func (s *Server) handleCompleteUpload(w http.ResponseWriter, r *http.Request) {
	authContext, _ := AuthFromContext(r.Context())

	var request completeUploadRequest
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, r, http.StatusBadRequest, "INVALID_BODY", "Invalid upload completion payload", nil)
		return
	}

	uploadID, err := ToPGUUID(request.UploadID)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "INVALID_UPLOAD", "Invalid upload identifier", nil)
		return
	}

	upload, err := s.Queries.GetUploadByID(r.Context(), uploadID)
	if err != nil {
		writeError(w, r, http.StatusNotFound, "UPLOAD_NOT_FOUND", "Upload not found", nil)
		return
	}

	if UUIDString(upload.UserID) != authContext.UserID {
		writeError(w, r, http.StatusForbidden, "FORBIDDEN", "Upload belongs to another user", nil)
		return
	}

	if err := s.Storage.StatObject(r.Context(), upload.ObjectKey); err != nil {
		writeError(w, r, http.StatusBadRequest, "UPLOAD_MISSING", "Uploaded file not found in storage", nil)
		return
	}

	documentIDString, versionIDString, err := parseObjectKey(upload.ObjectKey)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "UPLOAD_KEY_INVALID", "Upload key is invalid", nil)
		return
	}

	documentID := MustPGUUID(documentIDString)
	versionID := MustPGUUID(versionIDString)
	collectionID, err := optionalUUID(request.CollectionID)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "INVALID_COLLECTION", "Invalid collection identifier", nil)
		return
	}

	if err := s.withTx(r.Context(), func(ctx context.Context, queries *dbgen.Queries) error {
		document, err := queries.CreateDocument(ctx, dbgen.CreateDocumentParams{
			ID:               documentID,
			OrgID:            MustPGUUID(authContext.OrgID),
			Title:            fallback(request.Title, trimExtension(upload.OriginalFilename)),
			OriginalFilename: upload.OriginalFilename,
			MimeType:         upload.MimeType,
			SizeBytes:        upload.SizeBytes,
			ChecksumSha256:   upload.ChecksumSha256,
			OwnerUserID:      MustPGUUID(authContext.UserID),
			CollectionID:     collectionID,
			Status:           "processing",
		})
		if err != nil {
			return err
		}

		version, err := queries.CreateDocumentVersion(ctx, dbgen.CreateDocumentVersionParams{
			ID:             versionID,
			DocumentID:     document.ID,
			VersionNumber:  1,
			StorageKey:     upload.ObjectKey,
			SizeBytes:      upload.SizeBytes,
			ChecksumSha256: upload.ChecksumSha256,
			MimeType:       upload.MimeType,
			UploadedBy:     MustPGUUID(authContext.UserID),
			ChangeSummary:  fallback(request.ChangeSummary, "Initial upload"),
		})
		if err != nil {
			return err
		}

		if err := queries.SetDocumentCurrentVersion(ctx, dbgen.SetDocumentCurrentVersionParams{
			ID:               document.ID,
			CurrentVersionID: version.ID,
			SizeBytes:        version.SizeBytes,
			MimeType:         version.MimeType,
			ChecksumSha256:   version.ChecksumSha256,
		}); err != nil {
			return err
		}

		if err := queries.CompleteUpload(ctx, dbgen.CompleteUploadParams{
			ID:         upload.ID,
			Status:     "completed",
			DocumentID: document.ID,
		}); err != nil {
			return err
		}

		if err := saveMetadataAndTags(ctx, queries, document.ID, MustPGUUID(authContext.OrgID), request.Metadata, request.Tags); err != nil {
			return err
		}

		return nil
	}); err != nil {
		s.Logger.Error().Err(err).Msg("complete-upload transaction failed")
		writeError(w, r, http.StatusInternalServerError, "UPLOAD_COMPLETE_FAILED", "Could not finalize upload", nil)
		return
	}

	_ = s.recordAudit(r.Context(), authContext, "document.upload.completed", "document", documentIDString, map[string]any{
		"uploadId": upload.ID.String(),
	})

	_, _ = s.Queue.Enqueue(r.Context(), jobs.TypeNotify, jobs.NotificationPayload{
		UserID: authContext.UserID,
		Kind:   "upload.complete",
		Title:  "Upload accepted",
		Body:   "Your document is being processed for OCR and previews.",
		Data:   map[string]any{"documentId": documentIDString},
	})

	_, _ = s.enqueueJob(r.Context(), jobs.TypeOCR, jobs.OCRPayload{
		DocumentID:        documentIDString,
		DocumentVersionID: versionIDString,
	}, versionIDString)

	_, _ = s.enqueueJob(r.Context(), jobs.TypePreview, jobs.PreviewPayload{
		DocumentID:        documentIDString,
		DocumentVersionID: versionIDString,
		MimeType:          upload.MimeType,
		StorageKey:        upload.ObjectKey,
	}, versionIDString)

	s.respondDocumentDetail(w, r, documentIDString)
}

func (s *Server) handleGetDocument(w http.ResponseWriter, r *http.Request) {
	s.respondDocumentDetail(w, r, chi.URLParam(r, "documentID"))
}

func (s *Server) respondDocumentDetail(w http.ResponseWriter, r *http.Request, documentIDString string) {
	authContext, _ := AuthFromContext(r.Context())
	documentID, err := ToPGUUID(documentIDString)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "INVALID_DOCUMENT", "Invalid document identifier", nil)
		return
	}

	document, err := s.Queries.GetDocumentByID(r.Context(), documentID)
	if err != nil {
		writeError(w, r, http.StatusNotFound, "DOCUMENT_NOT_FOUND", "Document not found", nil)
		return
	}

	allowed, err := s.canAccessDocument(r.Context(), authContext, document)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "PERMISSION_CHECK_FAILED", "Could not evaluate access", nil)
		return
	}
	if !allowed {
		writeError(w, r, http.StatusForbidden, "FORBIDDEN", "You do not have access to this document", nil)
		return
	}

	metadata, _ := s.Queries.ListMetadataByDocumentID(r.Context(), document.ID)
	tags, _ := s.Queries.ListTagsByDocumentID(r.Context(), document.ID)
	versions, _ := s.Queries.ListDocumentVersions(r.Context(), document.ID)
	comments, _ := s.Queries.ListCommentsByDocumentID(r.Context(), document.ID)

	downloadURL := ""
	ocrStatus := "pending"
	if document.CurrentVersionID.Valid {
		currentVersion, err := s.Queries.GetDocumentVersionByID(r.Context(), document.CurrentVersionID)
		if err == nil {
			downloadURL, _ = s.Storage.CreateSignedDownloadURL(r.Context(), currentVersion.StorageKey, s.Config.SignedURLTTL)
			if ocrText, err := s.Queries.GetOCRTextByVersionID(r.Context(), currentVersion.ID); err == nil {
				ocrStatus = ocrText.ExtractionStatus
			}
		}
	}

	collectionName := ""
	if document.CollectionID.Valid {
		if collection, err := s.Queries.GetCollectionByID(r.Context(), document.CollectionID); err == nil {
			collectionName = collection.Name
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"document": map[string]any{
			"id":                   UUIDString(document.ID),
			"title":                document.Title,
			"originalFilename":     document.OriginalFilename,
			"mimeType":             document.MimeType,
			"sizeBytes":            document.SizeBytes,
			"status":               document.Status,
			"updatedAt":            timestamp(document.UpdatedAt),
			"currentVersionNumber": currentVersionNumber(versions),
			"metadata":             presentMetadata(metadata),
			"tags":                 presentTags(tags),
			"versions":             presentVersions(versions),
			"comments":             presentComments(comments),
			"downloadUrl":          downloadURL,
			"collectionName":       collectionName,
			"ocrStatus":            ocrStatus,
		},
	})
}

func (s *Server) handleUpdateDocument(w http.ResponseWriter, r *http.Request) {
	authContext, _ := AuthFromContext(r.Context())
	documentID, err := ToPGUUID(chi.URLParam(r, "documentID"))
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "INVALID_DOCUMENT", "Invalid document identifier", nil)
		return
	}

	document, err := s.Queries.GetDocumentByID(r.Context(), documentID)
	if err != nil {
		writeError(w, r, http.StatusNotFound, "DOCUMENT_NOT_FOUND", "Document not found", nil)
		return
	}

	if !authContext.IsAdmin && UUIDString(document.OwnerUserID) != authContext.UserID {
		writeError(w, r, http.StatusForbidden, "FORBIDDEN", "Only the owner or an admin can update this document", nil)
		return
	}

	var request updateDocumentRequest
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, r, http.StatusBadRequest, "INVALID_BODY", "Invalid document update payload", nil)
		return
	}

	collectionID, err := optionalUUID(request.CollectionID)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "INVALID_COLLECTION", "Invalid collection identifier", nil)
		return
	}

	if err := s.withTx(r.Context(), func(ctx context.Context, queries *dbgen.Queries) error {
		if _, err := queries.UpdateDocumentCore(ctx, dbgen.UpdateDocumentCoreParams{
			ID:           document.ID,
			Title:        fallback(request.Title, document.Title),
			CollectionID: collectionID,
			Status:       fallback(request.Status, document.Status),
		}); err != nil {
			return err
		}

		if err := saveMetadataAndTags(ctx, queries, document.ID, MustPGUUID(authContext.OrgID), request.Metadata, request.Tags); err != nil {
			return err
		}

		return nil
	}); err != nil {
		writeError(w, r, http.StatusInternalServerError, "DOCUMENT_UPDATE_FAILED", "Could not update document", nil)
		return
	}

	_ = s.recordAudit(r.Context(), authContext, "document.updated", "document", UUIDString(document.ID), nil)
	s.respondDocumentDetail(w, r, UUIDString(document.ID))
}

func (s *Server) handleArchiveDocument(w http.ResponseWriter, r *http.Request) {
	s.updateArchiveStatus(w, r, true)
}

func (s *Server) handleRestoreDocument(w http.ResponseWriter, r *http.Request) {
	s.updateArchiveStatus(w, r, false)
}

func (s *Server) updateArchiveStatus(w http.ResponseWriter, r *http.Request, archive bool) {
	authContext, _ := AuthFromContext(r.Context())
	if !authContext.IsAdmin && !contains(authContext.Permissions, "documents.admin") {
		writeError(w, r, http.StatusForbidden, "FORBIDDEN", "Admin permission required", nil)
		return
	}

	documentID, err := ToPGUUID(chi.URLParam(r, "documentID"))
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "INVALID_DOCUMENT", "Invalid document identifier", nil)
		return
	}

	if archive {
		err = s.Queries.ArchiveDocument(r.Context(), documentID)
	} else {
		err = s.Queries.RestoreDocument(r.Context(), documentID)
	}
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "DOCUMENT_STATUS_FAILED", "Could not update document status", nil)
		return
	}

	action := "document.restored"
	if archive {
		action = "document.archived"
	}
	_ = s.recordAudit(r.Context(), authContext, action, "document", UUIDString(documentID), nil)
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleSignedDownload(w http.ResponseWriter, r *http.Request) {
	authContext, _ := AuthFromContext(r.Context())
	documentID, err := ToPGUUID(chi.URLParam(r, "documentID"))
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "INVALID_DOCUMENT", "Invalid document identifier", nil)
		return
	}

	document, err := s.Queries.GetDocumentByID(r.Context(), documentID)
	if err != nil {
		writeError(w, r, http.StatusNotFound, "DOCUMENT_NOT_FOUND", "Document not found", nil)
		return
	}

	allowed, err := s.canAccessDocument(r.Context(), authContext, document)
	if err != nil || !allowed {
		writeError(w, r, http.StatusForbidden, "FORBIDDEN", "You do not have access to this document", nil)
		return
	}

	version, err := s.Queries.GetDocumentVersionByID(r.Context(), document.CurrentVersionID)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "VERSION_NOT_FOUND", "Current version not found", nil)
		return
	}

	signedURL, err := s.Storage.CreateSignedDownloadURL(r.Context(), version.StorageKey, s.Config.SignedURLTTL)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "SIGNED_URL_FAILED", "Could not create signed download URL", nil)
		return
	}

	_ = s.recordAudit(r.Context(), authContext, "document.download.issued", "document", UUIDString(document.ID), nil)
	writeJSON(w, http.StatusOK, map[string]any{
		"url":       signedURL,
		"expiresAt": time.Now().Add(s.Config.SignedURLTTL).Format(time.RFC3339),
	})
}

func (s *Server) handleListVersions(w http.ResponseWriter, r *http.Request) {
	documentID, err := ToPGUUID(chi.URLParam(r, "documentID"))
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "INVALID_DOCUMENT", "Invalid document identifier", nil)
		return
	}
	versions, err := s.Queries.ListDocumentVersions(r.Context(), documentID)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "VERSION_LIST_FAILED", "Could not load versions", nil)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": presentVersions(versions)})
}

func (s *Server) handleRestoreVersion(w http.ResponseWriter, r *http.Request) {
	authContext, _ := AuthFromContext(r.Context())
	if !authContext.IsAdmin && !contains(authContext.Permissions, "documents.admin") {
		writeError(w, r, http.StatusForbidden, "FORBIDDEN", "Admin permission required", nil)
		return
	}

	documentID, err := ToPGUUID(chi.URLParam(r, "documentID"))
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "INVALID_DOCUMENT", "Invalid document identifier", nil)
		return
	}
	versionID, err := ToPGUUID(chi.URLParam(r, "versionID"))
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "INVALID_VERSION", "Invalid version identifier", nil)
		return
	}
	version, err := s.Queries.GetDocumentVersionByID(r.Context(), versionID)
	if err != nil {
		writeError(w, r, http.StatusNotFound, "VERSION_NOT_FOUND", "Version not found", nil)
		return
	}

	if err := s.Queries.SetDocumentCurrentVersion(r.Context(), dbgen.SetDocumentCurrentVersionParams{
		ID:               documentID,
		CurrentVersionID: version.ID,
		SizeBytes:        version.SizeBytes,
		MimeType:         version.MimeType,
		ChecksumSha256:   version.ChecksumSha256,
	}); err != nil {
		writeError(w, r, http.StatusInternalServerError, "VERSION_RESTORE_FAILED", "Could not restore version", nil)
		return
	}

	_ = s.Queries.SetDocumentStatus(r.Context(), dbgen.SetDocumentStatusParams{
		ID:     documentID,
		Status: "ready",
	})
	_ = s.recordAudit(r.Context(), authContext, "document.version.restored", "document", UUIDString(documentID), map[string]any{
		"versionId": UUIDString(version.ID),
	})
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleListComments(w http.ResponseWriter, r *http.Request) {
	documentID, err := ToPGUUID(chi.URLParam(r, "documentID"))
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "INVALID_DOCUMENT", "Invalid document identifier", nil)
		return
	}
	comments, err := s.Queries.ListCommentsByDocumentID(r.Context(), documentID)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "COMMENTS_LIST_FAILED", "Could not load comments", nil)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": presentComments(comments)})
}

func (s *Server) handleCreateComment(w http.ResponseWriter, r *http.Request) {
	authContext, _ := AuthFromContext(r.Context())
	documentID, err := ToPGUUID(chi.URLParam(r, "documentID"))
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "INVALID_DOCUMENT", "Invalid document identifier", nil)
		return
	}
	var request createCommentRequest
	if err := decodeJSON(r, &request); err != nil || strings.TrimSpace(request.Body) == "" {
		writeError(w, r, http.StatusBadRequest, "INVALID_BODY", "Comment body is required", nil)
		return
	}
	comment, err := s.Queries.CreateComment(r.Context(), dbgen.CreateCommentParams{
		DocumentID:   documentID,
		AuthorUserID: MustPGUUID(authContext.UserID),
		Body:         strings.TrimSpace(request.Body),
	})
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "COMMENT_CREATE_FAILED", "Could not create comment", nil)
		return
	}

	go s.processMentions(r.Context(), authContext, UUIDString(documentID), request.Body)

	writeJSON(w, http.StatusOK, map[string]any{
		"id":         UUIDString(comment.ID),
		"authorName": authContext.FullName,
		"body":       comment.Body,
		"createdAt":  timestamp(comment.CreatedAt),
	})
}

var mentionRegex = regexp.MustCompile(`@([\w.+-]+@[\w.-]+\.\w{2,})`)

func (s *Server) processMentions(ctx context.Context, author AuthContext, documentID string, body string) {
	matches := mentionRegex.FindAllStringSubmatch(body, -1)
	seen := map[string]bool{}
	for _, match := range matches {
		email := match[1]
		if seen[email] || email == author.Email {
			continue
		}
		seen[email] = true

		user, err := s.Queries.GetUserByEmail(ctx, email)
		if err != nil {
			continue
		}

		_, _ = s.Queue.Enqueue(ctx, jobs.TypeNotify, jobs.NotificationPayload{
			UserID: UUIDString(user.ID),
			Kind:   "comment.mention",
			Title:  "You were mentioned in a comment",
			Body:   fmt.Sprintf("%s mentioned you in a comment.", author.FullName),
			Data:   map[string]any{"documentId": documentID},
		})
	}
}

func (s *Server) handleSearch(w http.ResponseWriter, r *http.Request) {
	authContext, _ := AuthFromContext(r.Context())
	query := strings.TrimSpace(r.URL.Query().Get("q"))
	if query == "" {
		writeJSON(w, http.StatusOK, map[string]any{"items": []any{}})
		return
	}

	results, err := s.Queries.SearchDocuments(r.Context(), dbgen.SearchDocumentsParams{
		OrgID:          MustPGUUID(authContext.OrgID),
		OwnerUserID:    MustPGUUID(authContext.UserID),
		Column3:        authContext.RoleIDs,
		Column4:        authContext.IsAdmin,
		PlaintoTsquery: query,
		Limit:          25,
		Offset:         0,
	})
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "SEARCH_FAILED", "Could not search documents", nil)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": presentSearchResults(results)})
}

func (s *Server) handleAdvancedSearch(w http.ResponseWriter, r *http.Request) {
	authContext, _ := AuthFromContext(r.Context())
	var request advancedSearchRequest
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, r, http.StatusBadRequest, "INVALID_BODY", "Invalid search request", nil)
		return
	}
	limit := request.Limit
	if limit == 0 {
		limit = 25
	}

	var dateFrom, dateTo pgtype.Timestamptz
	if request.DateFrom != "" {
		if parsed, err := time.Parse(time.RFC3339, request.DateFrom); err == nil {
			dateFrom = pgtype.Timestamptz{Time: parsed, Valid: true}
		}
	}
	if request.DateTo != "" {
		if parsed, err := time.Parse(time.RFC3339, request.DateTo); err == nil {
			dateTo = pgtype.Timestamptz{Time: parsed, Valid: true}
		}
	}

	results, err := s.Queries.SearchDocumentsWithFilters(r.Context(), dbgen.SearchDocumentsWithFiltersParams{
		OrgID:          MustPGUUID(authContext.OrgID),
		OwnerUserID:    MustPGUUID(authContext.UserID),
		Column3:        authContext.RoleIDs,
		Column4:        authContext.IsAdmin,
		PlaintoTsquery: request.Query,
		Limit:          limit,
		Offset:         request.Offset,
		Column8:        request.Status,
		Column9:        request.MimeType,
		Column10:       dateFrom,
		Column11:       dateTo,
	})
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "SEARCH_FAILED", "Could not search documents", nil)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": presentFilteredSearchResults(results)})
}

func (s *Server) handleListSavedSearches(w http.ResponseWriter, r *http.Request) {
	authContext, _ := AuthFromContext(r.Context())
	items, err := s.Queries.ListSavedSearches(r.Context(), MustPGUUID(authContext.UserID))
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "SAVED_SEARCHES_FAILED", "Could not load saved searches", nil)
		return
	}
	response := make([]map[string]any, 0, len(items))
	for _, item := range items {
		response = append(response, map[string]any{
			"id":        UUIDString(item.ID),
			"name":      item.Name,
			"queryText": item.QueryText,
			"filters":   decodeJSONBytes(item.FiltersJson),
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": response})
}

func (s *Server) handleCreateSavedSearch(w http.ResponseWriter, r *http.Request) {
	authContext, _ := AuthFromContext(r.Context())
	var request createSavedSearchRequest
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, r, http.StatusBadRequest, "INVALID_BODY", "Invalid saved search request", nil)
		return
	}
	saved, err := s.Queries.CreateSavedSearch(r.Context(), dbgen.CreateSavedSearchParams{
		OrgID:       MustPGUUID(authContext.OrgID),
		UserID:      MustPGUUID(authContext.UserID),
		Name:        request.Name,
		QueryText:   request.QueryText,
		FiltersJson: JSONBytes(request.Filters),
	})
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "SAVED_SEARCH_CREATE_FAILED", "Could not save search", nil)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"id":        UUIDString(saved.ID),
		"name":      saved.Name,
		"queryText": saved.QueryText,
		"filters":   decodeJSONBytes(saved.FiltersJson),
	})
}

func (s *Server) handleListAuditEvents(w http.ResponseWriter, r *http.Request) {
	authContext, _ := AuthFromContext(r.Context())
	if !authContext.IsAdmin && !contains(authContext.Permissions, "audit.read") {
		writeError(w, r, http.StatusForbidden, "FORBIDDEN", "Audit permission required", nil)
		return
	}

	limit, offset := parsePagination(r)
	action := strings.TrimSpace(r.URL.Query().Get("action"))
	fromStr := strings.TrimSpace(r.URL.Query().Get("from"))
	toStr := strings.TrimSpace(r.URL.Query().Get("to"))

	var fromTime, toTime pgtype.Timestamptz
	if fromStr != "" {
		if parsed, err := time.Parse(time.RFC3339, fromStr); err == nil {
			fromTime = pgtype.Timestamptz{Time: parsed, Valid: true}
		}
	}
	if toStr != "" {
		if parsed, err := time.Parse(time.RFC3339, toStr); err == nil {
			toTime = pgtype.Timestamptz{Time: parsed, Valid: true}
		}
	}

	items, err := s.Queries.ListAuditEventsFiltered(r.Context(), dbgen.ListAuditEventsFilteredParams{
		OrgID:   MustPGUUID(authContext.OrgID),
		Limit:   limit,
		Offset:  offset,
		Column4: action,
		Column5: fromTime,
		Column6: toTime,
	})
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "AUDIT_LIST_FAILED", "Could not load audit events", nil)
		return
	}
	response := make([]map[string]any, 0, len(items))
	for _, item := range items {
		response = append(response, map[string]any{
			"id":           UUIDString(item.ID),
			"action":       item.Action,
			"resourceType": item.ResourceType,
			"requestId":    item.RequestID,
			"actorRole":    item.ActorRole,
			"createdAt":    timestamp(item.CreatedAt),
			"payload":      decodeJSONBytes(item.PayloadJson),
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": response})
}

func (s *Server) handleAuditExport(w http.ResponseWriter, r *http.Request) {
	authContext, _ := AuthFromContext(r.Context())
	if !authContext.IsAdmin && !contains(authContext.Permissions, "audit.read") {
		writeError(w, r, http.StatusForbidden, "FORBIDDEN", "Audit permission required", nil)
		return
	}
	taskInfo, err := s.Queue.Enqueue(r.Context(), jobs.TypeAuditExport, jobs.AuditExportPayload{OrgID: authContext.OrgID})
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "AUDIT_EXPORT_FAILED", "Could not queue audit export", nil)
		return
	}
	_, _ = s.Queries.CreateJob(r.Context(), dbgen.CreateJobParams{
		DocumentVersionID: pgtype.UUID{},
		JobType:           jobs.TypeAuditExport,
		TaskID:            taskInfo.ID,
		Status:            "queued",
		PayloadJson:       JSONBytes(map[string]any{"orgId": authContext.OrgID}),
	})
	writeJSON(w, http.StatusAccepted, map[string]any{
		"jobId":  taskInfo.ID,
		"status": "queued",
	})
}

func (s *Server) handleListNotifications(w http.ResponseWriter, r *http.Request) {
	authContext, _ := AuthFromContext(r.Context())
	items, err := s.Queries.ListNotificationsByUserID(r.Context(), dbgen.ListNotificationsByUserIDParams{
		UserID: MustPGUUID(authContext.UserID),
		Limit:  50,
	})
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "NOTIFICATIONS_FAILED", "Could not load notifications", nil)
		return
	}
	response := make([]map[string]any, 0, len(items))
	for _, item := range items {
		response = append(response, map[string]any{
			"id":        UUIDString(item.ID),
			"kind":      item.Kind,
			"title":     item.Title,
			"body":      item.Body,
			"readAt":    timestamp(item.ReadAt),
			"createdAt": timestamp(item.CreatedAt),
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": response})
}

func (s *Server) handleMarkNotificationRead(w http.ResponseWriter, r *http.Request) {
	authContext, _ := AuthFromContext(r.Context())
	notificationID, err := ToPGUUID(chi.URLParam(r, "notificationID"))
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "INVALID_NOTIFICATION", "Invalid notification identifier", nil)
		return
	}
	if err := s.Queries.MarkNotificationRead(r.Context(), dbgen.MarkNotificationReadParams{
		ID:     notificationID,
		UserID: MustPGUUID(authContext.UserID),
	}); err != nil {
		writeError(w, r, http.StatusInternalServerError, "NOTIFICATION_UPDATE_FAILED", "Could not update notification", nil)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleListUsers(w http.ResponseWriter, r *http.Request) {
	authContext, _ := AuthFromContext(r.Context())
	rows, err := s.Queries.ListUsers(r.Context(), MustPGUUID(authContext.OrgID))
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "USERS_FAILED", "Could not load users", nil)
		return
	}
	items := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		roleObjects := make([]map[string]any, 0, len(row.RoleNames))
		for _, roleName := range row.RoleNames {
			roleObjects = append(roleObjects, map[string]any{"name": roleName})
		}
		items = append(items, map[string]any{
			"id":         UUIDString(row.ID),
			"email":      row.Email,
			"fullName":   row.FullName,
			"status":     row.Status,
			"mfaEnabled": row.MfaEnabled,
			"roles":      roleObjects,
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (s *Server) handleAssignRoles(w http.ResponseWriter, r *http.Request) {
	userID, err := ToPGUUID(chi.URLParam(r, "userID"))
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "INVALID_USER", "Invalid user identifier", nil)
		return
	}
	var request assignRolesRequest
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, r, http.StatusBadRequest, "INVALID_BODY", "Invalid role assignment request", nil)
		return
	}
	if err := s.withTx(r.Context(), func(ctx context.Context, queries *dbgen.Queries) error {
		if err := queries.DeleteUserRolesByUserID(ctx, userID); err != nil {
			return err
		}
		for _, roleIDString := range request.RoleIDs {
			roleID, err := ToPGUUID(roleIDString)
			if err != nil {
				return err
			}
			if err := queries.AddUserRole(ctx, dbgen.AddUserRoleParams{
				UserID: userID,
				RoleID: roleID,
			}); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		writeError(w, r, http.StatusInternalServerError, "ROLE_ASSIGN_FAILED", "Could not assign roles", nil)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleListRoles(w http.ResponseWriter, r *http.Request) {
	authContext, _ := AuthFromContext(r.Context())
	rows, err := s.Queries.ListRoles(r.Context(), MustPGUUID(authContext.OrgID))
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "ROLES_FAILED", "Could not load roles", nil)
		return
	}
	items := make([]map[string]any, 0, len(rows))
	for _, role := range rows {
		items = append(items, map[string]any{
			"id":          UUIDString(role.ID),
			"key":         role.Key,
			"name":        role.Name,
			"description": role.Description,
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (s *Server) handleListQuotas(w http.ResponseWriter, r *http.Request) {
	authContext, _ := AuthFromContext(r.Context())
	rows, err := s.Queries.ListQuotas(r.Context(), MustPGUUID(authContext.OrgID))
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "QUOTAS_FAILED", "Could not load quotas", nil)
		return
	}
	items := make([]map[string]any, 0, len(rows))
	for _, quota := range rows {
		items = append(items, map[string]any{
			"id":               UUIDString(quota.ID),
			"targetType":       quota.TargetType,
			"targetId":         UUIDString(quota.TargetID),
			"maxStorageBytes":  quota.MaxStorageBytes,
			"maxDocumentCount": quota.MaxDocumentCount,
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (s *Server) handleUpsertQuota(w http.ResponseWriter, r *http.Request) {
	authContext, _ := AuthFromContext(r.Context())
	var request quotaRequest
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, r, http.StatusBadRequest, "INVALID_BODY", "Invalid quota payload", nil)
		return
	}
	targetID, err := optionalUUID(request.TargetID)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "INVALID_TARGET", "Invalid target identifier", nil)
		return
	}
	quota, err := s.Queries.UpsertQuota(r.Context(), dbgen.UpsertQuotaParams{
		OrgID:            MustPGUUID(authContext.OrgID),
		TargetType:       request.TargetType,
		TargetID:         targetID,
		MaxStorageBytes:  request.MaxStorageBytes,
		MaxDocumentCount: request.MaxDocumentCount,
	})
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "QUOTA_SAVE_FAILED", "Could not save quota", nil)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"id":               UUIDString(quota.ID),
		"targetType":       quota.TargetType,
		"targetId":         UUIDString(quota.TargetID),
		"maxStorageBytes":  quota.MaxStorageBytes,
		"maxDocumentCount": quota.MaxDocumentCount,
	})
}

func (s *Server) handleListRetentionPolicies(w http.ResponseWriter, r *http.Request) {
	authContext, _ := AuthFromContext(r.Context())
	rows, err := s.Queries.ListRetentionPolicies(r.Context(), MustPGUUID(authContext.OrgID))
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "RETENTION_FAILED", "Could not load retention policies", nil)
		return
	}
	items := make([]map[string]any, 0, len(rows))
	for _, policy := range rows {
		items = append(items, map[string]any{
			"id":                    UUIDString(policy.ID),
			"name":                  policy.Name,
			"appliesToCollectionId": UUIDString(policy.AppliesToCollectionID),
			"retentionDays":         policy.RetentionDays,
			"archiveAfterDays":      int4(policy.ArchiveAfterDays),
			"deleteAfterDays":       int4(policy.DeleteAfterDays),
			"enabled":               policy.Enabled,
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (s *Server) handleUpsertRetentionPolicy(w http.ResponseWriter, r *http.Request) {
	authContext, _ := AuthFromContext(r.Context())
	var request retentionRequest
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, r, http.StatusBadRequest, "INVALID_BODY", "Invalid retention payload", nil)
		return
	}
	policyID := request.ID
	if policyID == "" {
		policyID = uuid.NewString()
	}
	collectionID, err := optionalUUID(request.AppliesToCollectionID)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "INVALID_COLLECTION", "Invalid collection identifier", nil)
		return
	}
	enabled := true
	if request.Enabled != nil {
		enabled = *request.Enabled
	}
	policy, err := s.Queries.UpsertRetentionPolicy(r.Context(), dbgen.UpsertRetentionPolicyParams{
		ID:                    MustPGUUID(policyID),
		OrgID:                 MustPGUUID(authContext.OrgID),
		Name:                  request.Name,
		AppliesToCollectionID: collectionID,
		RetentionDays:         request.RetentionDays,
		ArchiveAfterDays:      int4Ptr(request.ArchiveAfterDays),
		DeleteAfterDays:       int4Ptr(request.DeleteAfterDays),
		Enabled:               enabled,
	})
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "RETENTION_SAVE_FAILED", "Could not save retention policy", nil)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"id":                    UUIDString(policy.ID),
		"name":                  policy.Name,
		"appliesToCollectionId": UUIDString(policy.AppliesToCollectionID),
		"retentionDays":         policy.RetentionDays,
		"archiveAfterDays":      int4(policy.ArchiveAfterDays),
		"deleteAfterDays":       int4(policy.DeleteAfterDays),
		"enabled":               policy.Enabled,
	})
}

func (s *Server) handleListSettings(w http.ResponseWriter, r *http.Request) {
	authContext, _ := AuthFromContext(r.Context())
	rows, err := s.Queries.ListSystemSettings(r.Context(), MustPGUUID(authContext.OrgID))
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "SETTINGS_FAILED", "Could not load settings", nil)
		return
	}
	items := make([]map[string]any, 0, len(rows))
	for _, setting := range rows {
		items = append(items, map[string]any{
			"settingKey":   setting.SettingKey,
			"settingValue": decodeJSONBytes(setting.SettingValue),
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (s *Server) handleUpsertSetting(w http.ResponseWriter, r *http.Request) {
	authContext, _ := AuthFromContext(r.Context())
	var request settingRequest
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, r, http.StatusBadRequest, "INVALID_BODY", "Invalid setting payload", nil)
		return
	}
	setting, err := s.Queries.UpsertSystemSetting(r.Context(), dbgen.UpsertSystemSettingParams{
		OrgID:        MustPGUUID(authContext.OrgID),
		SettingKey:   request.SettingKey,
		SettingValue: JSONBytes(request.SettingValue),
	})
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "SETTING_SAVE_FAILED", "Could not save setting", nil)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"settingKey":   setting.SettingKey,
		"settingValue": decodeJSONBytes(setting.SettingValue),
	})
}

func (s *Server) handleListJobs(w http.ResponseWriter, r *http.Request) {
	rows, err := s.Queries.ListJobs(r.Context(), 100)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "JOBS_FAILED", "Could not load jobs", nil)
		return
	}
	items := make([]map[string]any, 0, len(rows))
	for _, job := range rows {
		items = append(items, map[string]any{
			"id":           UUIDString(job.ID),
			"jobType":      job.JobType,
			"status":       job.Status,
			"errorMessage": text(job.ErrorMessage),
			"createdAt":    timestamp(job.CreatedAt),
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (s *Server) handleRetryJob(w http.ResponseWriter, r *http.Request) {
	jobID, err := ToPGUUID(chi.URLParam(r, "jobID"))
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "INVALID_JOB", "Invalid job identifier", nil)
		return
	}
	job, err := s.Queries.GetJobByID(r.Context(), jobID)
	if err != nil {
		writeError(w, r, http.StatusNotFound, "JOB_NOT_FOUND", "Job not found", nil)
		return
	}
	info, err := s.Queue.Enqueue(r.Context(), job.JobType, decodeJSONBytes(job.PayloadJson))
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "JOB_RETRY_FAILED", "Could not retry job", nil)
		return
	}
	_, _ = s.Queries.CreateJob(r.Context(), dbgen.CreateJobParams{
		DocumentVersionID: job.DocumentVersionID,
		JobType:           job.JobType,
		TaskID:            info.ID,
		Status:            "queued",
		PayloadJson:       job.PayloadJson,
	})
	writeJSON(w, http.StatusAccepted, map[string]any{"jobId": info.ID, "status": "queued"})
}

func (s *Server) handleUsage(w http.ResponseWriter, r *http.Request) {
	authContext, _ := AuthFromContext(r.Context())
	summary, err := s.Queries.GetUsageSummary(r.Context(), MustPGUUID(authContext.OrgID))
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "USAGE_FAILED", "Could not load usage summary", nil)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"documentCount": summary.DocumentCount,
		"storageBytes":  summary.StorageBytes,
		"userCount":     summary.UserCount,
	})
}

func (s *Server) handleAdminHealth(w http.ResponseWriter, r *http.Request) {
	status := map[string]any{
		"api":      "ok",
		"database": "ok",
		"redis":    "ok",
		"storage":  "ok",
	}
	if err := s.DB.Ping(r.Context()); err != nil {
		status["database"] = "down"
	}
	if err := s.Redis.Ping(r.Context()).Err(); err != nil {
		status["redis"] = "down"
	}
	if _, err := s.Storage.CreateSignedDownloadURL(r.Context(), "healthcheck/probe", time.Minute); err != nil {
		status["storage"] = "degraded"
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"api":        status["api"],
		"database":   status["database"],
		"redis":      status["redis"],
		"storage":    status["storage"],
		"queueDepth": 0,
	})
}

type shareDocumentRequest struct {
	UserID      string `json:"userId"`
	AccessLevel string `json:"accessLevel"`
}

type createCollectionRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type updateCollectionRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type addCollectionMemberRequest struct {
	UserID      string `json:"userId"`
	AccessLevel string `json:"accessLevel"`
}

func (s *Server) handleDeleteDocument(w http.ResponseWriter, r *http.Request) {
	authContext, _ := AuthFromContext(r.Context())
	documentID, err := ToPGUUID(chi.URLParam(r, "documentID"))
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "INVALID_DOCUMENT", "Invalid document identifier", nil)
		return
	}

	document, err := s.Queries.GetDocumentByID(r.Context(), documentID)
	if err != nil {
		writeError(w, r, http.StatusNotFound, "DOCUMENT_NOT_FOUND", "Document not found", nil)
		return
	}

	if !authContext.IsAdmin && UUIDString(document.OwnerUserID) != authContext.UserID {
		writeError(w, r, http.StatusForbidden, "FORBIDDEN", "Only the owner or an admin can delete this document", nil)
		return
	}

	if err := s.Queries.SoftDeleteDocument(r.Context(), documentID); err != nil {
		writeError(w, r, http.StatusInternalServerError, "DOCUMENT_DELETE_FAILED", "Could not delete document", nil)
		return
	}

	_ = s.recordAudit(r.Context(), authContext, "document.deleted", "document", UUIDString(documentID), nil)
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleShareDocument(w http.ResponseWriter, r *http.Request) {
	authContext, _ := AuthFromContext(r.Context())
	documentID, err := ToPGUUID(chi.URLParam(r, "documentID"))
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "INVALID_DOCUMENT", "Invalid document identifier", nil)
		return
	}

	document, err := s.Queries.GetDocumentByID(r.Context(), documentID)
	if err != nil {
		writeError(w, r, http.StatusNotFound, "DOCUMENT_NOT_FOUND", "Document not found", nil)
		return
	}

	if !authContext.IsAdmin && UUIDString(document.OwnerUserID) != authContext.UserID {
		writeError(w, r, http.StatusForbidden, "FORBIDDEN", "Only the owner or an admin can share this document", nil)
		return
	}

	var request shareDocumentRequest
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, r, http.StatusBadRequest, "INVALID_BODY", "Invalid share request", nil)
		return
	}

	targetUserID, err := ToPGUUID(request.UserID)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "INVALID_USER", "Invalid user identifier", nil)
		return
	}

	accessLevel := fallback(request.AccessLevel, "viewer")
	if err := s.Queries.ShareDocument(r.Context(), dbgen.ShareDocumentParams{
		DocumentID:  documentID,
		SubjectType: "user",
		SubjectID:   targetUserID,
		AccessLevel: accessLevel,
	}); err != nil {
		writeError(w, r, http.StatusInternalServerError, "SHARE_FAILED", "Could not share document", nil)
		return
	}

	_ = s.recordAudit(r.Context(), authContext, "document.shared", "document", UUIDString(documentID), map[string]any{
		"sharedWith": request.UserID,
		"access":     accessLevel,
	})

	_, _ = s.Queue.Enqueue(r.Context(), jobs.TypeNotify, jobs.NotificationPayload{
		UserID: request.UserID,
		Kind:   "document.shared",
		Title:  "Document shared with you",
		Body:   fmt.Sprintf("%s shared a document with you.", authContext.FullName),
		Data:   map[string]any{"documentId": UUIDString(documentID)},
	})

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleRevokeShare(w http.ResponseWriter, r *http.Request) {
	authContext, _ := AuthFromContext(r.Context())
	documentID, err := ToPGUUID(chi.URLParam(r, "documentID"))
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "INVALID_DOCUMENT", "Invalid document identifier", nil)
		return
	}

	document, err := s.Queries.GetDocumentByID(r.Context(), documentID)
	if err != nil {
		writeError(w, r, http.StatusNotFound, "DOCUMENT_NOT_FOUND", "Document not found", nil)
		return
	}

	if !authContext.IsAdmin && UUIDString(document.OwnerUserID) != authContext.UserID {
		writeError(w, r, http.StatusForbidden, "FORBIDDEN", "Only the owner or an admin can revoke document shares", nil)
		return
	}

	var request shareDocumentRequest
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, r, http.StatusBadRequest, "INVALID_BODY", "Invalid revoke request", nil)
		return
	}

	targetUserID, err := ToPGUUID(request.UserID)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "INVALID_USER", "Invalid user identifier", nil)
		return
	}

	if err := s.Queries.RevokeDocumentShare(r.Context(), dbgen.RevokeDocumentShareParams{
		DocumentID:  documentID,
		SubjectType: "user",
		SubjectID:   targetUserID,
	}); err != nil {
		writeError(w, r, http.StatusInternalServerError, "REVOKE_FAILED", "Could not revoke document share", nil)
		return
	}

	_ = s.recordAudit(r.Context(), authContext, "document.share.revoked", "document", UUIDString(documentID), map[string]any{
		"revokedFrom": request.UserID,
	})
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleListSharedDocuments(w http.ResponseWriter, r *http.Request) {
	authContext, _ := AuthFromContext(r.Context())
	limit, offset := parsePagination(r)

	items, err := s.Queries.ListSharedDocuments(r.Context(), dbgen.ListSharedDocumentsParams{
		SubjectID: MustPGUUID(authContext.UserID),
		Limit:     limit,
		Offset:    offset,
	})
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "SHARED_LIST_FAILED", "Could not load shared documents", nil)
		return
	}

	response := make([]map[string]any, 0, len(items))
	for _, item := range items {
		response = append(response, map[string]any{
			"id":                   UUIDString(item.ID),
			"title":                item.Title,
			"originalFilename":     item.OriginalFilename,
			"mimeType":             item.MimeType,
			"sizeBytes":            item.SizeBytes,
			"status":               item.Status,
			"updatedAt":            timestamp(item.UpdatedAt),
			"currentVersionNumber": int4(item.CurrentVersionNumber),
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": response})
}

func (s *Server) handleListCollections(w http.ResponseWriter, r *http.Request) {
	authContext, _ := AuthFromContext(r.Context())
	items, err := s.Queries.ListCollections(r.Context(), MustPGUUID(authContext.OrgID))
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "COLLECTIONS_FAILED", "Could not load collections", nil)
		return
	}
	response := make([]map[string]any, 0, len(items))
	for _, item := range items {
		response = append(response, map[string]any{
			"id":          UUIDString(item.ID),
			"name":        item.Name,
			"description": item.Description,
			"createdAt":   timestamp(item.CreatedAt),
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": response})
}

func (s *Server) handleCreateCollection(w http.ResponseWriter, r *http.Request) {
	authContext, _ := AuthFromContext(r.Context())
	var request createCollectionRequest
	if err := decodeJSON(r, &request); err != nil || strings.TrimSpace(request.Name) == "" {
		writeError(w, r, http.StatusBadRequest, "INVALID_BODY", "Collection name is required", nil)
		return
	}

	collection, err := s.Queries.CreateCollection(r.Context(), dbgen.CreateCollectionParams{
		OrgID:       MustPGUUID(authContext.OrgID),
		Name:        strings.TrimSpace(request.Name),
		Description: strings.TrimSpace(request.Description),
		CreatedBy:   MustPGUUID(authContext.UserID),
	})
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "COLLECTION_CREATE_FAILED", "Could not create collection", nil)
		return
	}

	_ = s.Queries.AddCollectionMember(r.Context(), dbgen.AddCollectionMemberParams{
		CollectionID: collection.ID,
		UserID:       MustPGUUID(authContext.UserID),
		AccessLevel:  "owner",
	})

	_ = s.recordAudit(r.Context(), authContext, "collection.created", "collection", UUIDString(collection.ID), nil)
	writeJSON(w, http.StatusOK, map[string]any{
		"id":          UUIDString(collection.ID),
		"name":        collection.Name,
		"description": collection.Description,
		"createdAt":   timestamp(collection.CreatedAt),
	})
}

func (s *Server) handleGetCollection(w http.ResponseWriter, r *http.Request) {
	collectionID, err := ToPGUUID(chi.URLParam(r, "collectionID"))
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "INVALID_COLLECTION", "Invalid collection identifier", nil)
		return
	}

	collection, err := s.Queries.GetCollectionByID(r.Context(), collectionID)
	if err != nil {
		writeError(w, r, http.StatusNotFound, "COLLECTION_NOT_FOUND", "Collection not found", nil)
		return
	}

	members, _ := s.Queries.ListCollectionMembers(r.Context(), collectionID)
	memberList := make([]map[string]any, 0, len(members))
	for _, m := range members {
		memberList = append(memberList, map[string]any{
			"userId":      UUIDString(m.UserID),
			"fullName":    m.FullName,
			"email":       m.Email,
			"accessLevel": m.AccessLevel,
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"id":          UUIDString(collection.ID),
		"name":        collection.Name,
		"description": collection.Description,
		"createdAt":   timestamp(collection.CreatedAt),
		"members":     memberList,
	})
}

func (s *Server) handleUpdateCollection(w http.ResponseWriter, r *http.Request) {
	authContext, _ := AuthFromContext(r.Context())
	collectionID, err := ToPGUUID(chi.URLParam(r, "collectionID"))
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "INVALID_COLLECTION", "Invalid collection identifier", nil)
		return
	}

	existing, err := s.Queries.GetCollectionByID(r.Context(), collectionID)
	if err != nil {
		writeError(w, r, http.StatusNotFound, "COLLECTION_NOT_FOUND", "Collection not found", nil)
		return
	}

	var request updateCollectionRequest
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, r, http.StatusBadRequest, "INVALID_BODY", "Invalid collection update payload", nil)
		return
	}

	collection, err := s.Queries.UpdateCollection(r.Context(), dbgen.UpdateCollectionParams{
		ID:          collectionID,
		Name:        fallback(request.Name, existing.Name),
		Description: fallback(request.Description, existing.Description),
	})
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "COLLECTION_UPDATE_FAILED", "Could not update collection", nil)
		return
	}

	_ = s.recordAudit(r.Context(), authContext, "collection.updated", "collection", UUIDString(collection.ID), nil)
	writeJSON(w, http.StatusOK, map[string]any{
		"id":          UUIDString(collection.ID),
		"name":        collection.Name,
		"description": collection.Description,
	})
}

func (s *Server) handleDeleteCollection(w http.ResponseWriter, r *http.Request) {
	authContext, _ := AuthFromContext(r.Context())
	collectionID, err := ToPGUUID(chi.URLParam(r, "collectionID"))
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "INVALID_COLLECTION", "Invalid collection identifier", nil)
		return
	}

	if err := s.Queries.DeleteCollection(r.Context(), collectionID); err != nil {
		writeError(w, r, http.StatusInternalServerError, "COLLECTION_DELETE_FAILED", "Could not delete collection", nil)
		return
	}

	_ = s.recordAudit(r.Context(), authContext, "collection.deleted", "collection", UUIDString(collectionID), nil)
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleListCollectionDocuments(w http.ResponseWriter, r *http.Request) {
	collectionID, err := ToPGUUID(chi.URLParam(r, "collectionID"))
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "INVALID_COLLECTION", "Invalid collection identifier", nil)
		return
	}

	limit, offset := parsePagination(r)
	items, err := s.Queries.ListDocumentsByCollectionID(r.Context(), dbgen.ListDocumentsByCollectionIDParams{
		CollectionID: collectionID,
		Limit:        limit,
		Offset:       offset,
	})
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "COLLECTION_DOCS_FAILED", "Could not load collection documents", nil)
		return
	}

	response := make([]map[string]any, 0, len(items))
	for _, item := range items {
		response = append(response, map[string]any{
			"id":                   UUIDString(item.ID),
			"title":                item.Title,
			"originalFilename":     item.OriginalFilename,
			"mimeType":             item.MimeType,
			"sizeBytes":            item.SizeBytes,
			"status":               item.Status,
			"updatedAt":            timestamp(item.UpdatedAt),
			"currentVersionNumber": int4(item.CurrentVersionNumber),
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": response})
}

func (s *Server) handleAddCollectionMember(w http.ResponseWriter, r *http.Request) {
	authContext, _ := AuthFromContext(r.Context())
	collectionID, err := ToPGUUID(chi.URLParam(r, "collectionID"))
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "INVALID_COLLECTION", "Invalid collection identifier", nil)
		return
	}

	var request addCollectionMemberRequest
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, r, http.StatusBadRequest, "INVALID_BODY", "Invalid member request", nil)
		return
	}

	userID, err := ToPGUUID(request.UserID)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "INVALID_USER", "Invalid user identifier", nil)
		return
	}

	accessLevel := fallback(request.AccessLevel, "member")
	if err := s.Queries.AddCollectionMember(r.Context(), dbgen.AddCollectionMemberParams{
		CollectionID: collectionID,
		UserID:       userID,
		AccessLevel:  accessLevel,
	}); err != nil {
		writeError(w, r, http.StatusInternalServerError, "MEMBER_ADD_FAILED", "Could not add member", nil)
		return
	}

	_ = s.recordAudit(r.Context(), authContext, "collection.member.added", "collection", UUIDString(collectionID), map[string]any{
		"userId": request.UserID,
	})
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleRemoveCollectionMember(w http.ResponseWriter, r *http.Request) {
	authContext, _ := AuthFromContext(r.Context())
	collectionID, err := ToPGUUID(chi.URLParam(r, "collectionID"))
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "INVALID_COLLECTION", "Invalid collection identifier", nil)
		return
	}

	userID, err := ToPGUUID(chi.URLParam(r, "userID"))
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "INVALID_USER", "Invalid user identifier", nil)
		return
	}

	if err := s.Queries.RemoveCollectionMember(r.Context(), dbgen.RemoveCollectionMemberParams{
		CollectionID: collectionID,
		UserID:       userID,
	}); err != nil {
		writeError(w, r, http.StatusInternalServerError, "MEMBER_REMOVE_FAILED", "Could not remove member", nil)
		return
	}

	_ = s.recordAudit(r.Context(), authContext, "collection.member.removed", "collection", UUIDString(collectionID), map[string]any{
		"userId": UUIDString(userID),
	})
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) canAccessDocument(ctx context.Context, authContext AuthContext, document dbgen.Document) (bool, error) {
	if authContext.IsAdmin || UUIDString(document.OwnerUserID) == authContext.UserID {
		return true, nil
	}
	if document.CollectionID.Valid {
		isMember, err := s.Queries.IsCollectionMember(ctx, dbgen.IsCollectionMemberParams{
			CollectionID: document.CollectionID,
			UserID:       MustPGUUID(authContext.UserID),
		})
		if err != nil {
			return false, err
		}
		if isMember {
			return true, nil
		}
	}
	permissions, err := s.Queries.ListDocumentPermissions(ctx, document.ID)
	if err != nil {
		return false, err
	}
	for _, permission := range permissions {
		if !permission.Allow {
			continue
		}
		if permission.SubjectType == "user" && UUIDString(permission.SubjectID) == authContext.UserID {
			return true, nil
		}
		if permission.SubjectType == "role" {
			for _, roleID := range authContext.RoleIDs {
				if UUIDString(roleID) == UUIDString(permission.SubjectID) {
					return true, nil
				}
			}
		}
	}
	return false, nil
}

func (s *Server) withTx(ctx context.Context, fn func(context.Context, *dbgen.Queries) error) error {
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	queries := dbgen.New(tx)
	if err := fn(ctx, queries); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func saveMetadataAndTags(ctx context.Context, queries *dbgen.Queries, documentID pgtype.UUID, orgID pgtype.UUID, metadata []metadataRequest, tags []string) error {
	if err := queries.DeleteMetadataForDocument(ctx, documentID); err != nil {
		return err
	}
	for _, item := range metadata {
		params, err := buildMetadataParams(documentID, item)
		if err != nil {
			return err
		}
		if err := queries.InsertMetadata(ctx, params); err != nil {
			return err
		}
	}

	if err := queries.DeleteDocumentTags(ctx, documentID); err != nil {
		return err
	}
	for _, rawTag := range tags {
		name := strings.TrimSpace(rawTag)
		if name == "" {
			continue
		}
		tag, err := queries.UpsertTag(ctx, dbgen.UpsertTagParams{
			OrgID: orgID,
			Name:  name,
			Color: "slate",
		})
		if err != nil {
			return err
		}
		if err := queries.AttachTag(ctx, dbgen.AttachTagParams{
			DocumentID: documentID,
			TagID:      tag.ID,
		}); err != nil {
			return err
		}
	}

	return nil
}

func buildMetadataParams(documentID pgtype.UUID, item metadataRequest) (dbgen.InsertMetadataParams, error) {
	params := dbgen.InsertMetadataParams{
		DocumentID: documentID,
		SchemaKey:  item.SchemaKey,
		ValueJsonb: JSONBytes(item.ValueJSON),
	}
	if item.ValueText != "" {
		params.ValueText = pgtype.Text{String: item.ValueText, Valid: true}
	}
	if item.ValueNumber != nil {
		if err := params.ValueNumber.Scan(fmt.Sprintf("%f", *item.ValueNumber)); err != nil {
			return dbgen.InsertMetadataParams{}, err
		}
	}
	if item.ValueBoolean != nil {
		params.ValueBoolean = pgtype.Bool{Bool: *item.ValueBoolean, Valid: true}
	}
	if item.ValueDate != "" {
		parsed, err := time.Parse("2006-01-02", item.ValueDate)
		if err != nil {
			return dbgen.InsertMetadataParams{}, err
		}
		params.ValueDate = pgtype.Date{Time: parsed, Valid: true}
	}
	return params, nil
}

func (s *Server) recordAudit(ctx context.Context, authContext AuthContext, action string, resourceType string, resourceID string, payload map[string]any) error {
	var ip *netip.Addr
	if addr, err := netip.ParseAddr("127.0.0.1"); err == nil {
		ip = &addr
	}

	var resourceUUID pgtype.UUID
	if resourceID != "" {
		resourceUUID = MustPGUUID(resourceID)
	}

	role := ""
	if len(authContext.Roles) > 0 {
		role = authContext.Roles[0]
	}
	_, err := s.Queries.InsertAuditEvent(ctx, dbgen.InsertAuditEventParams{
		OrgID:        MustPGUUID(authContext.OrgID),
		ActorUserID:  MustPGUUID(authContext.UserID),
		ActorRole:    role,
		Action:       action,
		ResourceType: resourceType,
		ResourceID:   resourceUUID,
		IpAddress:    ip,
		UserAgent:    "",
		RequestID:    "",
		PayloadJson:  JSONBytes(payload),
	})
	return err
}

func parsePagination(r *http.Request) (int32, int32) {
	limit := int32(20)
	offset := int32(0)
	if value := r.URL.Query().Get("limit"); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			limit = int32(parsed)
		}
	}
	if value := r.URL.Query().Get("offset"); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			offset = int32(parsed)
		}
	}
	return limit, offset
}

func parseObjectKey(objectKey string) (string, string, error) {
	parts := strings.Split(objectKey, "/")
	if len(parts) < 7 {
		return "", "", fmt.Errorf("invalid object key")
	}
	return parts[3], parts[5], nil
}

func currentVersionNumber(versions []dbgen.DocumentVersion) int32 {
	if len(versions) == 0 {
		return 0
	}
	return versions[0].VersionNumber
}

func presentDocumentSummaryRow(item dbgen.ListAccessibleDocumentsRow) map[string]any {
	return map[string]any{
		"id":                   UUIDString(item.ID),
		"title":                item.Title,
		"originalFilename":     item.OriginalFilename,
		"mimeType":             item.MimeType,
		"sizeBytes":            item.SizeBytes,
		"status":               item.Status,
		"updatedAt":            timestamp(item.UpdatedAt),
		"currentVersionNumber": int4(item.CurrentVersionNumber),
		"previewStorageKey":    item.PreviewStorageKey,
	}
}

func presentMetadata(items []dbgen.DocumentMetadatum) []map[string]any {
	response := make([]map[string]any, 0, len(items))
	for _, item := range items {
		entry := map[string]any{
			"schemaKey": item.SchemaKey,
		}
		if item.ValueText.Valid {
			entry["valueText"] = item.ValueText.String
		}
		if item.ValueBoolean.Valid {
			entry["valueBoolean"] = item.ValueBoolean.Bool
		}
		if item.ValueDate.Valid {
			entry["valueDate"] = item.ValueDate.Time.Format("2006-01-02")
		}
		if item.ValueJsonb != nil {
			entry["valueJson"] = decodeJSONBytes(item.ValueJsonb)
		}
		response = append(response, entry)
	}
	return response
}

func presentTags(items []dbgen.Tag) []map[string]any {
	response := make([]map[string]any, 0, len(items))
	for _, item := range items {
		response = append(response, map[string]any{
			"id":    UUIDString(item.ID),
			"name":  item.Name,
			"color": item.Color,
		})
	}
	return response
}

func presentVersions(items []dbgen.DocumentVersion) []map[string]any {
	response := make([]map[string]any, 0, len(items))
	for _, item := range items {
		response = append(response, map[string]any{
			"id":             UUIDString(item.ID),
			"versionNumber":  item.VersionNumber,
			"mimeType":       item.MimeType,
			"sizeBytes":      item.SizeBytes,
			"checksumSha256": item.ChecksumSha256,
			"createdAt":      timestamp(item.CreatedAt),
			"changeSummary":  item.ChangeSummary,
		})
	}
	return response
}

func presentComments(items []dbgen.ListCommentsByDocumentIDRow) []map[string]any {
	response := make([]map[string]any, 0, len(items))
	for _, item := range items {
		response = append(response, map[string]any{
			"id":         UUIDString(item.ID),
			"authorName": item.AuthorName,
			"body":       item.Body,
			"createdAt":  timestamp(item.CreatedAt),
		})
	}
	return response
}

func presentSearchResults(items []dbgen.SearchDocumentsRow) []map[string]any {
	response := make([]map[string]any, 0, len(items))
	for _, item := range items {
		response = append(response, map[string]any{
			"documentId":       UUIDString(item.ID),
			"title":            item.Title,
			"originalFilename": item.OriginalFilename,
			"status":           item.Status,
			"updatedAt":        timestamp(item.UpdatedAt),
			"rank":             item.Rank,
			"snippet":          item.Snippet,
		})
	}
	return response
}

func presentFilteredSearchResults(items []dbgen.SearchDocumentsWithFiltersRow) []map[string]any {
	response := make([]map[string]any, 0, len(items))
	for _, item := range items {
		response = append(response, map[string]any{
			"documentId":       UUIDString(item.ID),
			"title":            item.Title,
			"originalFilename": item.OriginalFilename,
			"status":           item.Status,
			"mimeType":         item.MimeType,
			"updatedAt":        timestamp(item.UpdatedAt),
			"rank":             item.Rank,
			"snippet":          item.Snippet,
		})
	}
	return response
}

func decodeJSONBytes(value []byte) map[string]any {
	if len(value) == 0 {
		return map[string]any{}
	}
	var decoded map[string]any
	if err := json.Unmarshal(value, &decoded); err != nil {
		return map[string]any{}
	}
	return decoded
}

func fallback(value string, defaultValue string) string {
	if strings.TrimSpace(value) == "" {
		return defaultValue
	}
	return strings.TrimSpace(value)
}

func trimExtension(name string) string {
	return strings.TrimSuffix(name, filepath.Ext(name))
}

func timestamp(value pgtype.Timestamptz) string {
	if !value.Valid {
		return ""
	}
	return value.Time.Format(time.RFC3339)
}

func int4(value pgtype.Int4) any {
	if !value.Valid {
		return nil
	}
	return value.Int32
}

func text(value pgtype.Text) string {
	if !value.Valid {
		return ""
	}
	return value.String
}

func timestamptz(value time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: value, Valid: true}
}

func optionalUUID(value string) (pgtype.UUID, error) {
	if strings.TrimSpace(value) == "" {
		return pgtype.UUID{}, nil
	}
	return ToPGUUID(value)
}

func int4Ptr(value *int32) pgtype.Int4 {
	if value == nil {
		return pgtype.Int4{}
	}
	return pgtype.Int4{Int32: *value, Valid: true}
}

func contains(values []string, needle string) bool {
	for _, value := range values {
		if value == needle {
			return true
		}
	}
	return false
}

func (s *Server) exportAuditCSV(ctx context.Context, orgID string) (string, error) {
	rows, err := s.Queries.ListAuditEvents(ctx, dbgen.ListAuditEventsParams{
		OrgID:  MustPGUUID(orgID),
		Limit:  1000,
		Offset: 0,
	})
	if err != nil {
		return "", err
	}

	filePath := filepath.Join(os.TempDir(), "verin-audit-export-"+uuid.NewString()+".csv")
	file, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	_ = writer.Write([]string{"id", "action", "resource_type", "request_id", "actor_role", "created_at"})
	for _, row := range rows {
		_ = writer.Write([]string{
			UUIDString(row.ID),
			row.Action,
			row.ResourceType,
			row.RequestID,
			row.ActorRole,
			timestamp(row.CreatedAt),
		})
	}

	return filePath, writer.Error()
}
