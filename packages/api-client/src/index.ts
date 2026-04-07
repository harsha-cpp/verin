export type ShareDocumentPayload = {
  userId: string;
  accessLevel?: string;
};

export type CreateSpacePayload = {
  name: string;
  description?: string;
};

export type CreateTeamPayload = {
  name: string;
  slug: string;
};

export type JoinTeamPayload = {
  code: string;
};

export type CreateInvitePayload = {
  maxUses?: number;
  expiresInHours?: number;
};

export type UpdateTeamPayload = {
  name: string;
  slug: string;
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

function getCSRFToken(): string {
  if (typeof document === "undefined") return "";
  const match = document.cookie.match(/(?:^|;\s*)verin_csrf=([^;]*)/);
  return match ? decodeURIComponent(match[1]) : "";
}

async function request<T>(path: string, init?: RequestInit & { csrf?: boolean }): Promise<T> {
  const headers = new Headers(init?.headers ?? {});
  if (init?.csrf) {
    headers.set("X-CSRF-Token", getCSRFToken());
  }
  if (init?.body && !(init.body instanceof FormData) && !headers.has("Content-Type")) {
    headers.set("Content-Type", "application/json");
  }

  const res = await fetch(`${resolveBaseUrl()}${path}`, {
    credentials: "include",
    ...init,
    headers,
  });

  if (!res.ok) {
    const body = await res.json().catch(() => null);
    const message = body?.error?.message ?? `Request failed (${res.status})`;
    const error = new Error(message) as Error & { status?: number };
    error.status = res.status;
    throw error;
  }

  if (res.status === 204) {
    return undefined as T;
  }

  return res.json() as Promise<T>;
}

export const api = {
  me() {
    return request("/api/v1/auth/me");
  },
  logout() {
    return request("/api/v1/auth/logout", { method: "POST", csrf: true });
  },
  home() {
    return request("/api/v1/home");
  },
  listDocuments(search = "") {
    const params = new URLSearchParams();
    if (search) params.set("q", search);
    const query = params.toString();
    return request(`/api/v1/documents${query ? `?${query}` : ""}`);
  },
  getDocument(documentId: string) {
    return request(`/api/v1/documents/${documentId}`);
  },
  search(query: string) {
    const params = new URLSearchParams({ q: query });
    return request(`/api/v1/search?${params.toString()}`);
  },
  notifications() {
    return request("/api/v1/notifications");
  },
  markNotificationRead(notificationId: string) {
    return request(`/api/v1/notifications/${notificationId}/read`, { method: "POST", csrf: true });
  },
  async uploadDocument(file: File, metadata: { title?: string; collectionId?: string; changeSummary?: string; department?: string; tags?: string }) {
    const formData = new FormData();
    formData.append("file", file);
    if (metadata.title) formData.append("title", metadata.title);
    if (metadata.collectionId) formData.append("collectionId", metadata.collectionId);
    if (metadata.changeSummary) formData.append("changeSummary", metadata.changeSummary);
    if (metadata.department) formData.append("department", metadata.department);
    if (metadata.tags) formData.append("tags", metadata.tags);

    return request("/api/v1/documents/upload", {
      method: "POST",
      body: formData,
      csrf: true,
    });
  },
  deleteDocument(documentId: string) {
    return request(`/api/v1/documents/${documentId}`, { method: "DELETE", csrf: true });
  },
  addComment(documentId: string, body: string) {
    return request(`/api/v1/documents/${documentId}/comments`, {
      method: "POST",
      body: JSON.stringify({ body }),
      csrf: true,
    });
  },
  shareDocument(documentId: string, payload: ShareDocumentPayload) {
    return request(`/api/v1/documents/${documentId}/share`, {
      method: "POST",
      body: JSON.stringify(payload),
      csrf: true,
    });
  },
  revokeShare(documentId: string, userId: string) {
    return request(`/api/v1/documents/${documentId}/share`, {
      method: "DELETE",
      body: JSON.stringify({ userId }),
      csrf: true,
    });
  },
  listSharedDocuments() {
    return request("/api/v1/shared");
  },
  listSpaces() {
    return request("/api/v1/spaces");
  },
  createSpace(payload: CreateSpacePayload) {
    return request("/api/v1/spaces", {
      method: "POST",
      body: JSON.stringify(payload),
      csrf: true,
    });
  },
  getSpace(spaceId: string) {
    return request(`/api/v1/spaces/${spaceId}`);
  },
  deleteSpace(spaceId: string) {
    return request(`/api/v1/spaces/${spaceId}`, { method: "DELETE", csrf: true });
  },
  createTeam(payload: CreateTeamPayload) {
    return request("/api/v1/teams", {
      method: "POST",
      body: JSON.stringify(payload),
      csrf: true,
    });
  },
  joinTeam(payload: JoinTeamPayload) {
    return request("/api/v1/teams/join", {
      method: "POST",
      body: JSON.stringify(payload),
      csrf: true,
    });
  },
  createInvite(payload: CreateInvitePayload) {
    return request("/api/v1/teams/invite", {
      method: "POST",
      body: JSON.stringify(payload),
      csrf: true,
    });
  },
  listTeamMembers() {
    return request("/api/v1/teams/members");
  },
  getTeamInfo() {
    return request("/api/v1/teams/info");
  },
  updateTeam(payload: UpdateTeamPayload) {
    return request("/api/v1/teams", {
      method: "PUT",
      body: JSON.stringify(payload),
      csrf: true,
    });
  },
  deleteTeam() {
    return request("/api/v1/teams", { method: "DELETE", csrf: true });
  },
  updateMemberRole(memberId: string, roleKey: string) {
    return request(`/api/v1/teams/members/${memberId}/role`, {
      method: "PUT",
      body: JSON.stringify({ roleKey }),
      csrf: true,
    });
  },
  removeMember(memberId: string) {
    return request(`/api/v1/teams/members/${memberId}`, { method: "DELETE", csrf: true });
  },
};
