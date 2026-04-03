import createClient from "openapi-fetch";

import type { paths } from "./generated";

export type LoginPayload = {
  email: string;
  password: string;
};

export type MFAVerifyPayload = {
  code: string;
};

export type InitUploadPayload = {
  fileName: string;
  mimeType: string;
  sizeBytes: number;
  checksumSha256: string;
};

export type CompleteUploadPayload = {
  uploadId: string;
  title: string;
  collectionId?: string;
  changeSummary?: string;
  metadata?: Array<{ schemaKey: string; valueText?: string }>;
  tags?: string[];
};

export type SavedSearchPayload = {
  name: string;
  queryText: string;
  filters?: Record<string, unknown>;
};

export type ShareDocumentPayload = {
  userId: string;
  accessLevel?: string;
};

export type CreateCollectionPayload = {
  name: string;
  description?: string;
};

export type AddCollectionMemberPayload = {
  userId: string;
  accessLevel?: string;
};

function resolveBaseUrl() {
  if (typeof window === "undefined") {
    return "";
  }

  const runtimeBaseUrl = (window as Window & { __VERIN_API__?: string }).__VERIN_API__;
  const envBaseUrl = (import.meta as ImportMeta & { env?: Record<string, string | undefined> }).env?.VITE_API_BASE_URL;
  const baseUrl = runtimeBaseUrl || envBaseUrl || "";

  return baseUrl.replace(/\/$/, "");
}

export const client = createClient<paths>({
  baseUrl: resolveBaseUrl(),
  credentials: "include",
});

function getCSRFToken(): string {
  if (typeof document === "undefined") return "";
  const match = document.cookie.match(/(?:^|;\s*)verin_csrf=([^;]*)/);
  return match ? decodeURIComponent(match[1]) : "";
}

type APIResponse<T> = Promise<T>;

async function unwrap<T>(promise: Promise<{ data?: T; error?: unknown }>): APIResponse<T> {
  const response = await promise;
  if (response.error) {
    throw response.error;
  }
  if (response.data === undefined) {
    throw new Error("API response body was empty");
  }
  return response.data as T;
}

export const api = {
  login(payload: LoginPayload) {
    return unwrap(client.POST("/api/v1/auth/login", { body: payload }));
  },
  logout() {
    return client.POST("/api/v1/auth/logout");
  },
  me() {
    return unwrap(client.GET("/api/v1/auth/me"));
  },
  setupMFA() {
    return unwrap(client.POST("/api/v1/auth/mfa/setup"));
  },
  verifyMFA(payload: MFAVerifyPayload) {
    return unwrap(client.POST("/api/v1/auth/mfa/verify", { body: payload }));
  },
  listDocuments(search = "") {
    return unwrap(
      client.GET("/api/v1/documents", {
        params: { query: { q: search } },
      }),
    );
  },
  getDocument(documentId: string) {
    return unwrap(client.GET("/api/v1/documents/{documentId}", { params: { path: { documentId } } }));
  },
  initUpload(payload: InitUploadPayload) {
    return unwrap(client.POST("/api/v1/documents/init-upload", { body: payload }));
  },
  completeUpload(payload: CompleteUploadPayload) {
    return unwrap(client.POST("/api/v1/documents/complete-upload", { body: payload }));
  },
  search(query: string) {
    return unwrap(client.GET("/api/v1/search", { params: { query: { q: query } } }));
  },
  advancedSearch(payload: paths["/api/v1/search/advanced"]["post"]["requestBody"]["content"]["application/json"]) {
    return unwrap(client.POST("/api/v1/search/advanced", { body: payload }));
  },
  savedSearches() {
    return unwrap(client.GET("/api/v1/search/saved"));
  },
  createSavedSearch(payload: SavedSearchPayload) {
    return unwrap(client.POST("/api/v1/search/saved", { body: payload }));
  },
  auditEvents() {
    return unwrap(client.GET("/api/v1/audit/events"));
  },
  exportAudit() {
    return unwrap(client.POST("/api/v1/audit/reports/export"));
  },
  notifications() {
    return unwrap(client.GET("/api/v1/notifications"));
  },
  markNotificationRead(notificationId: string) {
    return client.POST("/api/v1/notifications/{notificationId}/read", { params: { path: { notificationId } } });
  },
  adminUsers() {
    return unwrap(client.GET("/api/v1/admin/users"));
  },
  adminRoles() {
    return unwrap(client.GET("/api/v1/admin/roles"));
  },
  adminQuotas() {
    return unwrap(client.GET("/api/v1/admin/quotas"));
  },
  upsertQuota(payload: paths["/api/v1/admin/quotas"]["post"]["requestBody"]["content"]["application/json"]) {
    return unwrap(client.POST("/api/v1/admin/quotas", { body: payload }));
  },
  adminRetention() {
    return unwrap(client.GET("/api/v1/admin/retention"));
  },
  upsertRetention(payload: paths["/api/v1/admin/retention"]["post"]["requestBody"]["content"]["application/json"]) {
    return unwrap(client.POST("/api/v1/admin/retention", { body: payload }));
  },
  adminSettings() {
    return unwrap(client.GET("/api/v1/admin/settings"));
  },
  upsertSetting(payload: paths["/api/v1/admin/settings"]["post"]["requestBody"]["content"]["application/json"]) {
    return unwrap(client.POST("/api/v1/admin/settings", { body: payload }));
  },
  adminJobs() {
    return unwrap(client.GET("/api/v1/admin/jobs"));
  },
  adminUsage() {
    return unwrap(client.GET("/api/v1/admin/usage"));
  },
  adminHealth() {
    return unwrap(client.GET("/api/v1/admin/health"));
  },

  async uploadDocument(file: File, metadata: { title?: string; collectionId?: string; changeSummary?: string; department?: string; tags?: string }) {
    const base = resolveBaseUrl();
    const formData = new FormData();
    formData.append("file", file);
    if (metadata.title) formData.append("title", metadata.title);
    if (metadata.collectionId) formData.append("collectionId", metadata.collectionId);
    if (metadata.changeSummary) formData.append("changeSummary", metadata.changeSummary);
    if (metadata.department) formData.append("department", metadata.department);
    if (metadata.tags) formData.append("tags", metadata.tags);

    const res = await fetch(`${base}/api/v1/documents/upload`, {
      method: "POST",
      credentials: "include",
      headers: { "X-CSRF-Token": getCSRFToken() },
      body: formData,
    });
    if (!res.ok) {
      const err = await res.json().catch(() => ({ error: { message: "Upload failed" } }));
      throw new Error(err.error?.message ?? "Upload failed");
    }
    return res.json();
  },

  async deleteDocument(documentId: string) {
    const base = resolveBaseUrl();
    const res = await fetch(`${base}/api/v1/documents/${documentId}`, {
      method: "DELETE",
      credentials: "include",
      headers: { "X-CSRF-Token": getCSRFToken() },
    });
    if (!res.ok) throw new Error("Failed to delete document");
  },
  async shareDocument(documentId: string, payload: ShareDocumentPayload) {
    const base = resolveBaseUrl();
    const res = await fetch(`${base}/api/v1/documents/${documentId}/share`, {
      method: "POST",
      credentials: "include",
      headers: { "Content-Type": "application/json", "X-CSRF-Token": getCSRFToken() },
      body: JSON.stringify(payload),
    });
    if (!res.ok) throw new Error("Failed to share document");
  },
  async revokeShare(documentId: string, userId: string) {
    const base = resolveBaseUrl();
    const res = await fetch(`${base}/api/v1/documents/${documentId}/share`, {
      method: "DELETE",
      credentials: "include",
      headers: { "Content-Type": "application/json", "X-CSRF-Token": getCSRFToken() },
      body: JSON.stringify({ userId }),
    });
    if (!res.ok) throw new Error("Failed to revoke share");
  },
  async listSharedDocuments() {
    const base = resolveBaseUrl();
    const res = await fetch(`${base}/api/v1/shared`, { credentials: "include" });
    if (!res.ok) throw new Error("Failed to load shared documents");
    return res.json();
  },
  async listCollections() {
    const base = resolveBaseUrl();
    const res = await fetch(`${base}/api/v1/collections`, { credentials: "include" });
    if (!res.ok) throw new Error("Failed to load collections");
    return res.json();
  },
  async createCollection(payload: CreateCollectionPayload) {
    const base = resolveBaseUrl();
    const res = await fetch(`${base}/api/v1/collections`, {
      method: "POST",
      credentials: "include",
      headers: { "Content-Type": "application/json", "X-CSRF-Token": getCSRFToken() },
      body: JSON.stringify(payload),
    });
    if (!res.ok) throw new Error("Failed to create collection");
    return res.json();
  },
  async getCollection(collectionId: string) {
    const base = resolveBaseUrl();
    const res = await fetch(`${base}/api/v1/collections/${collectionId}`, { credentials: "include" });
    if (!res.ok) throw new Error("Failed to load collection");
    return res.json();
  },
  async deleteCollection(collectionId: string) {
    const base = resolveBaseUrl();
    const res = await fetch(`${base}/api/v1/collections/${collectionId}`, {
      method: "DELETE",
      credentials: "include",
      headers: { "X-CSRF-Token": getCSRFToken() },
    });
    if (!res.ok) throw new Error("Failed to delete collection");
  },
  async addCollectionMember(collectionId: string, payload: AddCollectionMemberPayload) {
    const base = resolveBaseUrl();
    const res = await fetch(`${base}/api/v1/collections/${collectionId}/members`, {
      method: "POST",
      credentials: "include",
      headers: { "Content-Type": "application/json", "X-CSRF-Token": getCSRFToken() },
      body: JSON.stringify(payload),
    });
    if (!res.ok) throw new Error("Failed to add member");
  },
};

export type SessionResponse = Awaited<ReturnType<typeof api.me>>;
export type DocumentsResponse = Awaited<ReturnType<typeof api.listDocuments>>;
export type DocumentDetailResponse = Awaited<ReturnType<typeof api.getDocument>>;
