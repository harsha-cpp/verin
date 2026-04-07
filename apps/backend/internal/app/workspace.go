package app

import (
	"context"
)

func (s *Server) getWorkspaceSnapshot(ctx context.Context, orgID string) (WorkspaceSnapshot, error) {
	if cached, ok := s.workspaceCache.Get(orgID); ok {
		return cached, nil
	}

	pgOrgID, err := ToPGUUID(orgID)
	if err != nil {
		return WorkspaceSnapshot{}, err
	}

	org, err := s.Queries.GetOrganizationByID(ctx, pgOrgID)
	if err != nil {
		return WorkspaceSnapshot{}, err
	}

	members, err := s.Queries.ListTeamMembers(ctx, pgOrgID)
	if err != nil {
		return WorkspaceSnapshot{}, err
	}

	snapshot := WorkspaceSnapshot{
		ID:          UUIDString(org.ID),
		Name:        org.Name,
		Slug:        org.Slug,
		MemberCount: len(members),
	}
	s.workspaceCache.Set(orgID, snapshot)
	return snapshot, nil
}

// getWorkspaceSnapshotWithMemberCount builds a workspace snapshot using a
// pre-fetched member count, avoiding a redundant ListTeamMembers query when
// the caller already has the member list.
func (s *Server) getWorkspaceSnapshotWithMemberCount(ctx context.Context, orgID string, memberCount int) (WorkspaceSnapshot, error) {
	if cached, ok := s.workspaceCache.Get(orgID); ok {
		return cached, nil
	}

	pgOrgID, err := ToPGUUID(orgID)
	if err != nil {
		return WorkspaceSnapshot{}, err
	}

	org, err := s.Queries.GetOrganizationByID(ctx, pgOrgID)
	if err != nil {
		return WorkspaceSnapshot{}, err
	}

	snapshot := WorkspaceSnapshot{
		ID:          UUIDString(org.ID),
		Name:        org.Name,
		Slug:        org.Slug,
		MemberCount: memberCount,
	}
	s.workspaceCache.Set(orgID, snapshot)
	return snapshot, nil
}

func (s *Server) invalidateWorkspaceCache(orgID string) {
	if orgID == "" {
		return
	}
	s.workspaceCache.Delete(orgID)
}

func normalizeRoleKey(role string) string {
	switch role {
	case "admin", "owner":
		return "owner"
	case "editor":
		return "editor"
	case "auditor", "viewer":
		return "viewer"
	default:
		return role
	}
}

func normalizeRoles(items []string) []string {
	seen := make(map[string]bool, len(items))
	result := make([]string, 0, len(items))
	for _, item := range items {
		role := normalizeRoleKey(item)
		if role == "" || seen[role] {
			continue
		}
		seen[role] = true
		result = append(result, role)
	}
	return result
}

func normalizedPrimaryRole(items []string) string {
	normalized := normalizeRoles(items)
	if len(normalized) == 0 {
		return "viewer"
	}
	return normalized[0]
}
