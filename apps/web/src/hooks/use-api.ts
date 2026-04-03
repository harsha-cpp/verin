import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { api } from "@verin/api-client";

import {
  mockAdmin,
  mockAudit,
  mockDocumentDetail,
  mockDocuments,
  mockNotifications,
  mockSearch,
  mockSession,
} from "@/mocks/data";

async function withFallback<T>(request: () => Promise<T>, fallback: T) {
  try {
    return await request();
  } catch {
    return fallback;
  }
}

export function useSession() {
  return useQuery({
    queryKey: ["session"],
    queryFn: () => withFallback(() => api.me(), mockSession),
  });
}

export function useDocuments(search = "") {
  return useQuery({
    queryKey: ["documents", search],
    queryFn: () => withFallback(() => api.listDocuments(search), mockDocuments),
  });
}

export function useDocument(documentId: string) {
  return useQuery({
    queryKey: ["document", documentId],
    queryFn: () => withFallback(() => api.getDocument(documentId), mockDocumentDetail),
  });
}

export function useSearch(query: string) {
  return useQuery({
    queryKey: ["search", query],
    queryFn: () => withFallback(() => api.search(query), mockSearch),
    enabled: query.length > 0,
  });
}

export function useAudit() {
  return useQuery({
    queryKey: ["audit"],
    queryFn: () => withFallback(() => api.auditEvents(), mockAudit),
  });
}

export function useNotifications() {
  return useQuery({
    queryKey: ["notifications"],
    queryFn: () => withFallback(() => api.notifications(), mockNotifications),
  });
}

export function useAdminData() {
  const users = useQuery({
    queryKey: ["admin", "users"],
    queryFn: () => withFallback(() => api.adminUsers(), mockAdmin.users),
  });
  const roles = useQuery({
    queryKey: ["admin", "roles"],
    queryFn: () => withFallback(() => api.adminRoles(), mockAdmin.roles),
  });
  const quotas = useQuery({
    queryKey: ["admin", "quotas"],
    queryFn: () => withFallback(() => api.adminQuotas(), mockAdmin.quotas),
  });
  const retention = useQuery({
    queryKey: ["admin", "retention"],
    queryFn: () => withFallback(() => api.adminRetention(), mockAdmin.retention),
  });
  const settings = useQuery({
    queryKey: ["admin", "settings"],
    queryFn: () => withFallback(() => api.adminSettings(), mockAdmin.settings),
  });
  const jobs = useQuery({
    queryKey: ["admin", "jobs"],
    queryFn: () => withFallback(() => api.adminJobs(), mockAdmin.jobs),
  });
  const usage = useQuery({
    queryKey: ["admin", "usage"],
    queryFn: () => withFallback(() => api.adminUsage(), mockAdmin.usage),
  });
  const health = useQuery({
    queryKey: ["admin", "health"],
    queryFn: () => withFallback(() => api.adminHealth(), mockAdmin.health),
  });

  return { users, roles, quotas, retention, settings, jobs, usage, health };
}

export function useLogin() {
  return useMutation({
    mutationFn: api.login,
  });
}

export function useInitUpload() {
  return useMutation({
    mutationFn: api.initUpload,
  });
}

export function useCompleteUpload() {
  return useMutation({
    mutationFn: api.completeUpload,
  });
}

export function useUploadDocument() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ file, metadata }: { file: File; metadata: { title?: string; collectionId?: string; changeSummary?: string; department?: string; tags?: string } }) =>
      api.uploadDocument(file, metadata),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["documents"] }),
  });
}

export function useCollections() {
  return useQuery({
    queryKey: ["collections"],
    queryFn: () => withFallback(() => api.listCollections(), { items: [] }),
  });
}

export function useSharedDocuments() {
  return useQuery({
    queryKey: ["shared-documents"],
    queryFn: () => withFallback(() => api.listSharedDocuments(), { items: [] }),
  });
}

export function useDeleteDocument() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: api.deleteDocument,
    onSuccess: () => qc.invalidateQueries({ queryKey: ["documents"] }),
  });
}

export function useCreateCollection() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: api.createCollection,
    onSuccess: () => qc.invalidateQueries({ queryKey: ["collections"] }),
  });
}

export function useDeleteCollection() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: api.deleteCollection,
    onSuccess: () => qc.invalidateQueries({ queryKey: ["collections"] }),
  });
}
