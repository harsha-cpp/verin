package jobs

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

const (
	TypeOCR         = "document:ocr"
	TypePreview     = "document:preview"
	TypeNotify      = "notifications:dispatch"
	TypeRetention   = "retention:sweep"
	TypeAuditExport = "audit:export"
	TypeOrphanClean = "maintenance:orphan-cleanup"
)

type Queue struct {
	client *asynq.Client
}

func New(client *asynq.Client) *Queue {
	return &Queue{client: client}
}

func (q *Queue) Enqueue(ctx context.Context, taskType string, payload any) (*asynq.TaskInfo, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal task payload: %w", err)
	}

	task := asynq.NewTask(taskType, body)
	return q.client.EnqueueContext(ctx, task)
}

type OCRPayload struct {
	DocumentID        string `json:"document_id"`
	DocumentVersionID string `json:"document_version_id"`
}

type PreviewPayload struct {
	DocumentID        string `json:"document_id"`
	DocumentVersionID string `json:"document_version_id"`
	MimeType          string `json:"mime_type"`
	StorageKey        string `json:"storage_key"`
}

type NotificationPayload struct {
	UserID string         `json:"user_id"`
	Kind   string         `json:"kind"`
	Title  string         `json:"title"`
	Body   string         `json:"body"`
	Data   map[string]any `json:"data"`
}

type AuditExportPayload struct {
	OrgID string `json:"org_id"`
}

func NewID() string {
	return uuid.NewString()
}
