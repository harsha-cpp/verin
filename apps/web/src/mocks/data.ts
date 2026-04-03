export const mockSession = {
  authenticated: true,
  csrfToken: "local-csrf-token",
  user: {
    id: "40000000-0000-0000-0000-000000000001",
    email: "admin@verin.local",
    fullName: "Asha Rao",
    status: "active",
    mfaEnabled: true,
    roles: [{ id: "30000000-0000-0000-0000-000000000001", key: "admin", name: "Admin" }],
  },
};

export const mockDocuments = {
  total: 3,
  items: [
    {
      id: "90000000-0000-0000-0000-000000000001",
      title: "Quarterly vendor compliance pack",
      originalFilename: "vendor-compliance-q1.pdf",
      mimeType: "application/pdf",
      sizeBytes: 3298391,
      status: "ready",
      updatedAt: "2026-04-02T09:00:00Z",
      currentVersionNumber: 3,
      previewStorageKey: "orgs/default/documents/doc-1/versions/ver-1/preview",
    },
    {
      id: "90000000-0000-0000-0000-000000000002",
      title: "Operations runbook",
      originalFilename: "operations-runbook.docx",
      mimeType: "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
      sizeBytes: 884120,
      status: "processing",
      updatedAt: "2026-04-02T08:16:00Z",
      currentVersionNumber: 1,
      previewStorageKey: "",
    },
    {
      id: "90000000-0000-0000-0000-000000000003",
      title: "Archive intake manifest",
      originalFilename: "archive-intake.csv",
      mimeType: "text/plain",
      sizeBytes: 121032,
      status: "archived",
      updatedAt: "2026-04-01T17:45:00Z",
      currentVersionNumber: 4,
      previewStorageKey: "",
    },
  ],
};

export const mockDocumentDetail = {
  document: {
    ...mockDocuments.items[0],
    collectionName: "Compliance",
    ocrStatus: "completed",
    downloadUrl: "#",
    metadata: [
      { schemaKey: "department", valueText: "Compliance" },
      { schemaKey: "retentionClass", valueText: "7-year" },
      { schemaKey: "effectiveDate", valueDate: "2026-01-01" },
    ],
    tags: [
      { id: "1", name: "vendor", color: "slate" },
      { id: "2", name: "quarterly", color: "slate" },
    ],
    versions: [
      {
        id: "v3",
        versionNumber: 3,
        mimeType: "application/pdf",
        sizeBytes: 3298391,
        checksumSha256: "9df2...",
        createdAt: "2026-04-02T09:00:00Z",
        changeSummary: "Uploaded signed revision",
      },
      {
        id: "v2",
        versionNumber: 2,
        mimeType: "application/pdf",
        sizeBytes: 3121140,
        checksumSha256: "21ae...",
        createdAt: "2026-03-20T13:10:00Z",
        changeSummary: "Updated compliance appendix",
      },
    ],
    comments: [
      {
        id: "c1",
        authorName: "Leena Iyer",
        body: "Retention class verified against the latest policy.",
        createdAt: "2026-04-02T10:04:00Z",
      },
    ],
  },
};

export const mockSearch = {
  items: [
    {
      documentId: mockDocuments.items[0].id,
      title: mockDocuments.items[0].title,
      originalFilename: mockDocuments.items[0].originalFilename,
      status: mockDocuments.items[0].status,
      updatedAt: mockDocuments.items[0].updatedAt,
      rank: 92,
    },
  ],
};

export const mockAudit = {
  items: [
    {
      id: "a1",
      action: "document.upload.completed",
      resourceType: "document",
      requestId: "req_a1",
      actorRole: "admin",
      createdAt: "2026-04-02T09:01:00Z",
      payload: { documentId: mockDocuments.items[0].id },
    },
    {
      id: "a2",
      action: "document.version.restored",
      resourceType: "document",
      requestId: "req_a2",
      actorRole: "admin",
      createdAt: "2026-04-01T16:12:00Z",
      payload: { versionId: "v2" },
    },
  ],
};

export const mockNotifications = {
  items: [
    {
      id: "n1",
      kind: "upload.complete",
      title: "Upload accepted",
      body: "Your file is queued for OCR and preview generation.",
      createdAt: "2026-04-02T09:02:00Z",
    },
  ],
};

export const mockAdmin = {
  users: {
    items: [
      mockSession.user,
      {
        id: "40000000-0000-0000-0000-000000000002",
        email: "editor@verin.local",
        fullName: "Milan Dutta",
        status: "active",
        mfaEnabled: false,
        roles: [{ id: "30000000-0000-0000-0000-000000000002", key: "editor", name: "Editor" }],
      },
    ],
  },
  roles: {
    items: [
      { id: "30000000-0000-0000-0000-000000000001", key: "admin", name: "Admin", description: "Full access" },
      { id: "30000000-0000-0000-0000-000000000002", key: "editor", name: "Editor", description: "Upload and manage" },
      { id: "30000000-0000-0000-0000-000000000003", key: "auditor", name: "Auditor", description: "Review audit and retention" },
    ],
  },
  quotas: {
    items: [
      { id: "q1", targetType: "org", maxStorageBytes: 10737418240, maxDocumentCount: 50000 },
    ],
  },
  retention: {
    items: [
      { id: "r1", name: "Operations default", retentionDays: 365, archiveAfterDays: 180, enabled: true },
      { id: "r2", name: "Compliance hold", retentionDays: 2555, enabled: true },
    ],
  },
  settings: {
    items: [{ settingKey: "branding", settingValue: { productName: "Verin DMS", accent: "slate" } }],
  },
  jobs: {
    items: [
      { id: "j1", jobType: "document:ocr", status: "completed", createdAt: "2026-04-02T09:02:12Z" },
      { id: "j2", jobType: "document:preview", status: "processing", createdAt: "2026-04-02T09:02:15Z" },
    ],
  },
  usage: { documentCount: 1284, storageBytes: 869730877, userCount: 18 },
  health: { api: "ok", database: "ok", redis: "ok", storage: "ok", queueDepth: 1 },
};
