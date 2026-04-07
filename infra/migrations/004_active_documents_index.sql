CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_documents_active_by_org
    ON documents (org_id, updated_at DESC)
    WHERE deleted_at IS NULL;
