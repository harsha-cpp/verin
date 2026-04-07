import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { api } from "@verin/api-client";

export function useSession() {
  return useQuery({
    queryKey: ["session"],
    queryFn: () => api.me(),
    retry: false,
  });
}

export function useHome() {
  return useQuery({
    queryKey: ["home"],
    queryFn: () => api.home(),
  });
}

export function useDocuments(search = "") {
  return useQuery({
    queryKey: ["documents", search],
    queryFn: () => api.listDocuments(search),
  });
}

export function useDocument(documentId: string) {
  return useQuery({
    queryKey: ["document", documentId],
    queryFn: () => api.getDocument(documentId),
    enabled: Boolean(documentId),
    staleTime: 0,
    gcTime: 0,
  });
}

export function useSearch(query: string) {
  return useQuery({
    queryKey: ["search", query],
    queryFn: () => api.search(query),
    enabled: query.trim().length > 0,
  });
}

export function useNotifications() {
  return useQuery({
    queryKey: ["notifications"],
    queryFn: () => api.notifications(),
  });
}

export function useUploadDocument() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ file, metadata }: { file: File; metadata: { title?: string; collectionId?: string; changeSummary?: string; department?: string; tags?: string } }) =>
      api.uploadDocument(file, metadata),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["documents"] });
      qc.invalidateQueries({ queryKey: ["home"] });
      toast.success("Document uploaded");
    },
    onError: (err: Error) => {
      toast.error(err.message || "Upload failed");
    },
  });
}

export function useDeleteDocument() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: api.deleteDocument,
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["documents"] });
      qc.invalidateQueries({ queryKey: ["home"] });
      toast.success("Document deleted");
    },
    onError: () => {
      toast.error("Could not delete document");
    },
  });
}

export function useAddComment() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ documentId, body }: { documentId: string; body: string }) => api.addComment(documentId, body),
    onSuccess: (_data, variables) => {
      qc.invalidateQueries({ queryKey: ["document", variables.documentId] });
      qc.invalidateQueries({ queryKey: ["home"] });
      toast.success("Comment added");
    },
    onError: () => {
      toast.error("Could not post comment");
    },
  });
}

export function useShareDocument() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ documentId, userId, accessLevel }: { documentId: string; userId: string; accessLevel?: string }) =>
      api.shareDocument(documentId, { userId, accessLevel }),
    onSuccess: (_data, variables) => {
      qc.invalidateQueries({ queryKey: ["document", variables.documentId] });
      toast.success("Shared successfully");
    },
    onError: (err: Error) => {
      toast.error(err.message || "Could not share document");
    },
  });
}

export function useRevokeShare() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ documentId, userId }: { documentId: string; userId: string }) => api.revokeShare(documentId, userId),
    onSuccess: (_data, variables) => {
      qc.invalidateQueries({ queryKey: ["document", variables.documentId] });
      toast.success("Access removed");
    },
    onError: () => {
      toast.error("Could not revoke access");
    },
  });
}

export function useSharedDocuments() {
  return useQuery({
    queryKey: ["shared-documents"],
    queryFn: () => api.listSharedDocuments(),
  });
}

export function useSpaces() {
  return useQuery({
    queryKey: ["spaces"],
    queryFn: () => api.listSpaces(),
  });
}

export function useCreateSpace() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: api.createSpace,
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["spaces"] });
      qc.invalidateQueries({ queryKey: ["home"] });
      toast.success("Space created");
    },
    onError: () => {
      toast.error("Could not create space");
    },
  });
}

export function useDeleteSpace() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: api.deleteSpace,
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["spaces"] });
      qc.invalidateQueries({ queryKey: ["documents"] });
      qc.invalidateQueries({ queryKey: ["home"] });
      toast.success("Space removed");
    },
    onError: () => {
      toast.error("Could not remove space");
    },
  });
}

export function useCreateTeam() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: api.createTeam,
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["session"] });
      toast.success("Workspace created");
    },
    onError: (err: Error) => {
      toast.error(err.message || "Could not create workspace");
    },
  });
}

export function useJoinTeam() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: api.joinTeam,
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["session"] });
      toast.success("Joined workspace");
    },
    onError: (err: Error) => {
      toast.error(err.message || "Invalid invite code");
    },
  });
}

export function useCreateInvite() {
  return useMutation({
    mutationFn: api.createInvite,
    onSuccess: () => {
      toast.success("Invite generated");
    },
    onError: () => {
      toast.error("Could not generate invite");
    },
  });
}

export function useTeamMembers() {
  return useQuery({
    queryKey: ["team-members"],
    queryFn: () => api.listTeamMembers(),
  });
}

export function useTeamInfo() {
  return useQuery({
    queryKey: ["team-info"],
    queryFn: () => api.getTeamInfo(),
  });
}

export function useUpdateTeam() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: api.updateTeam,
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["team-info"] });
      qc.invalidateQueries({ queryKey: ["session"] });
      qc.invalidateQueries({ queryKey: ["home"] });
      toast.success("Settings saved");
    },
    onError: () => {
      toast.error("Could not save settings");
    },
  });
}

export function useDeleteTeam() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: api.deleteTeam,
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["session"] });
      qc.invalidateQueries({ queryKey: ["home"] });
      toast.success("Workspace deleted");
    },
    onError: () => {
      toast.error("Could not delete workspace");
    },
  });
}

export function useUpdateMemberRole() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ memberId, roleKey }: { memberId: string; roleKey: string }) => api.updateMemberRole(memberId, roleKey),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["team-members"] });
      qc.invalidateQueries({ queryKey: ["team-info"] });
      toast.success("Role updated");
    },
    onError: () => {
      toast.error("Could not update role");
    },
  });
}

export function useRemoveMember() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: api.removeMember,
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["team-members"] });
      qc.invalidateQueries({ queryKey: ["team-info"] });
      qc.invalidateQueries({ queryKey: ["home"] });
      toast.success("Member removed");
    },
    onError: () => {
      toast.error("Could not remove member");
    },
  });
}
