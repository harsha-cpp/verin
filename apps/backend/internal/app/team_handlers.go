package app

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/verin/dms/apps/backend/internal/dbgen"
)

type createTeamRequest struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type joinTeamRequest struct {
	Code string `json:"code"`
}

type createInviteRequest struct {
	MaxUses        int32 `json:"maxUses"`
	ExpiresInHours int   `json:"expiresInHours"`
}

var roleDefinitions = []struct {
	Key         string
	Name        string
	Description string
	Permissions []string
}{
	{"owner", "Owner", "Full workspace access", []string{"documents.read", "documents.write", "documents.admin", "audit.read", "admin.manage"}},
	{"editor", "Editor", "Upload, organize, and share documents", []string{"documents.read", "documents.write"}},
	{"viewer", "Viewer", "Read shared documents and collaborate in comments", []string{"documents.read"}},
}

func (s *Server) handleCreateTeam(w http.ResponseWriter, r *http.Request) {
	var req createTeamRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, r, http.StatusBadRequest, "INVALID_BODY", "Invalid request payload", nil)
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	req.Slug = strings.TrimSpace(strings.ToLower(req.Slug))
	if req.Name == "" || req.Slug == "" {
		writeError(w, r, http.StatusBadRequest, "MISSING_FIELDS", "Team name and slug are required", nil)
		return
	}

	authCtx, _ := AuthFromContext(r.Context())
	if authCtx.OrgID != "" {
		writeError(w, r, http.StatusConflict, "ALREADY_IN_TEAM", "You are already a member of a team", nil)
		return
	}

	ctx := r.Context()
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "TX_FAILED", "Could not start transaction", nil)
		return
	}
	defer tx.Rollback(ctx)

	txq := dbgen.New(tx)

	if err := txq.EnsureGlobalPermissions(ctx); err != nil {
		writeError(w, r, http.StatusInternalServerError, "PERMISSIONS_FAILED", "Could not ensure permissions", nil)
		return
	}

	org, err := txq.CreateOrganization(ctx, dbgen.CreateOrganizationParams{Name: req.Name, Slug: req.Slug})
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			writeError(w, r, http.StatusConflict, "SLUG_TAKEN", "That team slug is already taken", nil)
			return
		}
		writeError(w, r, http.StatusInternalServerError, "ORG_CREATE_FAILED", "Could not create team", nil)
		return
	}

	allPerms, err := txq.ListAllPermissions(ctx)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "PERMS_FAILED", "Could not load permissions", nil)
		return
	}
	permByKey := make(map[string]pgtype.UUID, len(allPerms))
	for _, p := range allPerms {
		permByKey[p.Key] = p.ID
	}

	var ownerRoleID pgtype.UUID
	for _, rd := range roleDefinitions {
		role, err := txq.CreateRole(ctx, dbgen.CreateRoleParams{
			OrgID:       org.ID,
			Key:         rd.Key,
			Name:        rd.Name,
			Description: rd.Description,
		})
		if err != nil {
			writeError(w, r, http.StatusInternalServerError, "ROLE_CREATE_FAILED", "Could not create roles", nil)
			return
		}
		if rd.Key == "owner" {
			ownerRoleID = role.ID
		}
		for _, permKey := range rd.Permissions {
			if permID, ok := permByKey[permKey]; ok {
				_ = txq.AddRolePermission(ctx, dbgen.AddRolePermissionParams{RoleID: role.ID, PermissionID: permID})
			}
		}
	}

	userID, err := ToPGUUID(authCtx.UserID)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "INVALID_USER", "Invalid user ID", nil)
		return
	}

	if err := txq.SetUserOrg(ctx, dbgen.SetUserOrgParams{ID: userID, OrgID: org.ID}); err != nil {
		writeError(w, r, http.StatusInternalServerError, "SET_ORG_FAILED", "Could not assign team", nil)
		return
	}

	if err := txq.AddUserRole(ctx, dbgen.AddUserRoleParams{UserID: userID, RoleID: ownerRoleID}); err != nil {
		writeError(w, r, http.StatusInternalServerError, "ROLE_ASSIGN_FAILED", "Could not assign owner role", nil)
		return
	}

	if _, err := txq.CreateDefaultQuota(ctx, org.ID); err != nil {
		writeError(w, r, http.StatusInternalServerError, "QUOTA_FAILED", "Could not create quota", nil)
		return
	}

	if err := tx.Commit(ctx); err != nil {
		writeError(w, r, http.StatusInternalServerError, "COMMIT_FAILED", "Could not finalize team creation", nil)
		return
	}

	s.refreshSessionOrg(w, r, UUIDString(org.ID))
	s.invalidateWorkspaceCache(UUIDString(org.ID))

	writeJSON(w, http.StatusCreated, map[string]any{
		"teamId": UUIDString(org.ID),
		"name":   org.Name,
		"slug":   org.Slug,
	})
}

func (s *Server) handleJoinTeam(w http.ResponseWriter, r *http.Request) {
	var req joinTeamRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, r, http.StatusBadRequest, "INVALID_BODY", "Invalid request payload", nil)
		return
	}

	req.Code = strings.TrimSpace(strings.ToUpper(req.Code))
	if req.Code == "" {
		writeError(w, r, http.StatusBadRequest, "MISSING_CODE", "Invite code is required", nil)
		return
	}

	authCtx, _ := AuthFromContext(r.Context())
	if authCtx.OrgID != "" {
		writeError(w, r, http.StatusConflict, "ALREADY_IN_TEAM", "You are already a member of a team", nil)
		return
	}

	ctx := r.Context()
	invite, err := s.Queries.ClaimTeamInvite(ctx, req.Code)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "INVALID_INVITE", "Invite code is invalid, expired, or exhausted", nil)
		return
	}

	userID, _ := ToPGUUID(authCtx.UserID)
	if err := s.Queries.SetUserOrg(ctx, dbgen.SetUserOrgParams{ID: userID, OrgID: invite.OrgID}); err != nil {
		writeError(w, r, http.StatusInternalServerError, "SET_ORG_FAILED", "Could not assign team", nil)
		return
	}

	roles, _ := s.Queries.ListRoles(ctx, invite.OrgID)
	for _, role := range roles {
		if role.Key == "editor" {
			_ = s.Queries.AddUserRole(ctx, dbgen.AddUserRoleParams{UserID: userID, RoleID: role.ID})
			break
		}
	}

	org, _ := s.Queries.GetOrganizationByID(ctx, invite.OrgID)
	s.refreshSessionOrg(w, r, UUIDString(invite.OrgID))
	s.invalidateWorkspaceCache(UUIDString(invite.OrgID))

	writeJSON(w, http.StatusOK, map[string]any{
		"teamId":   UUIDString(invite.OrgID),
		"teamName": org.Name,
	})
}

func (s *Server) handleCreateInvite(w http.ResponseWriter, r *http.Request) {
	var req createInviteRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, r, http.StatusBadRequest, "INVALID_BODY", "Invalid request payload", nil)
		return
	}

	if req.ExpiresInHours <= 0 {
		req.ExpiresInHours = 24
	}
	if req.MaxUses < 0 {
		req.MaxUses = 0
	}

	authCtx, _ := AuthFromContext(r.Context())
	orgID, _ := ToPGUUID(authCtx.OrgID)
	userID, _ := ToPGUUID(authCtx.UserID)

	code := generateInviteCode(8)
	invite, err := s.Queries.CreateTeamInvite(r.Context(), dbgen.CreateTeamInviteParams{
		OrgID:     orgID,
		Code:      code,
		CreatedBy: userID,
		ExpiresAt: pgtype.Timestamptz{Time: time.Now().Add(time.Duration(req.ExpiresInHours) * time.Hour), Valid: true},
		MaxUses:   req.MaxUses,
	})
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "INVITE_FAILED", "Could not create invite", nil)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"code":      invite.Code,
		"expiresAt": invite.ExpiresAt.Time.Format(time.RFC3339),
		"maxUses":   invite.MaxUses,
	})
}

func (s *Server) handleListMembers(w http.ResponseWriter, r *http.Request) {
	authCtx, _ := AuthFromContext(r.Context())
	orgID, _ := ToPGUUID(authCtx.OrgID)

	members, err := s.Queries.ListTeamMembers(r.Context(), orgID)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "LIST_FAILED", "Could not list members", nil)
		return
	}

	items := make([]map[string]any, 0, len(members))
	for _, m := range members {
		roleKeys := make([]string, 0, len(m.RoleKeys))
		for _, roleKey := range m.RoleKeys {
			roleKeys = append(roleKeys, normalizeRoleKey(roleKey))
		}
		items = append(items, map[string]any{
			"id":        UUIDString(m.ID),
			"email":     m.Email,
			"fullName":  m.FullName,
			"avatarUrl": m.AvatarUrl,
			"status":    m.Status,
			"roles":     normalizeRoles(roleKeys),
			"role":      normalizedPrimaryRole(roleKeys),
			"joinedAt":  m.CreatedAt.Time.Format(time.RFC3339),
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

type updateMemberRoleRequest struct {
	RoleKey string `json:"roleKey"`
}

func (s *Server) handleUpdateMemberRole(w http.ResponseWriter, r *http.Request) {
	var req updateMemberRoleRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, r, http.StatusBadRequest, "INVALID_BODY", "Invalid request payload", nil)
		return
	}

	memberID := chi.URLParam(r, "memberID")
	if memberID == "" || req.RoleKey == "" {
		writeError(w, r, http.StatusBadRequest, "MISSING_FIELDS", "Member ID and role key are required", nil)
		return
	}

	authCtx, _ := AuthFromContext(r.Context())

	if !authCtx.IsAdmin {
		writeError(w, r, http.StatusForbidden, "FORBIDDEN", "Only workspace owners can change roles", nil)
		return
	}

	if memberID == authCtx.UserID {
		writeError(w, r, http.StatusBadRequest, "CANNOT_CHANGE_OWN_ROLE", "You cannot change your own role", nil)
		return
	}

	targetUser, err := s.Queries.GetUserByID(r.Context(), MustPGUUID(memberID))
	if err != nil {
		writeError(w, r, http.StatusNotFound, "MEMBER_NOT_FOUND", "Member not found", nil)
		return
	}

	targetRoles, _ := s.Queries.ListUserRoles(r.Context(), targetUser.ID)
	for _, tr := range targetRoles {
		if tr.Key == "owner" {
			writeError(w, r, http.StatusForbidden, "CANNOT_CHANGE_OWNER", "Owner role cannot be changed", nil)
			return
		}
	}

	orgID, _ := ToPGUUID(authCtx.OrgID)
	userID, _ := ToPGUUID(memberID)

	role, err := resolveRoleByKey(r.Context(), s, orgID, normalizeRoleKey(req.RoleKey))
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "INVALID_ROLE", "Role not found", nil)
		return
	}

	_ = s.Queries.DeleteUserRolesByUserID(r.Context(), userID)
	_ = s.Queries.AddUserRole(r.Context(), dbgen.AddUserRoleParams{UserID: userID, RoleID: role.ID})
	s.invalidateWorkspaceCache(authCtx.OrgID)

	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (s *Server) handleRemoveMember(w http.ResponseWriter, r *http.Request) {
	memberID := chi.URLParam(r, "memberID")
	if memberID == "" {
		writeError(w, r, http.StatusBadRequest, "MISSING_MEMBER", "Member ID is required", nil)
		return
	}

	authCtx, _ := AuthFromContext(r.Context())

	if !authCtx.IsAdmin {
		writeError(w, r, http.StatusForbidden, "FORBIDDEN", "Only workspace owners can remove members", nil)
		return
	}

	if memberID == authCtx.UserID {
		writeError(w, r, http.StatusBadRequest, "CANNOT_REMOVE_SELF", "You cannot remove yourself from the team", nil)
		return
	}

	orgID, _ := ToPGUUID(authCtx.OrgID)
	userID, _ := ToPGUUID(memberID)

	_ = s.Queries.DeleteUserRolesByUserID(r.Context(), userID)
	if err := s.Queries.RemoveTeamMember(r.Context(), dbgen.RemoveTeamMemberParams{ID: userID, OrgID: orgID}); err != nil {
		writeError(w, r, http.StatusInternalServerError, "REMOVE_FAILED", "Could not remove member", nil)
		return
	}
	s.invalidateWorkspaceCache(authCtx.OrgID)

	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (s *Server) handleGetTeamInfo(w http.ResponseWriter, r *http.Request) {
	authCtx, _ := AuthFromContext(r.Context())
	orgID, _ := ToPGUUID(authCtx.OrgID)

	org, err := s.Queries.GetOrganizationByID(r.Context(), orgID)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "ORG_NOT_FOUND", "Team not found", nil)
		return
	}

	members, _ := s.Queries.ListTeamMembers(r.Context(), orgID)
	roles, _ := s.Queries.ListRoles(r.Context(), orgID)
	roleList := make([]map[string]any, 0, len(roles))
	for _, role := range roles {
		normalized := normalizeRoleKey(role.Key)
		roleList = append(roleList, map[string]any{
			"id":          UUIDString(role.ID),
			"key":         normalized,
			"name":        strings.ToUpper(normalized[:1]) + normalized[1:],
			"description": role.Description,
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"id":          UUIDString(org.ID),
		"name":        org.Name,
		"slug":        org.Slug,
		"memberCount": len(members),
		"roles":       roleList,
	})
}

type updateTeamRequest struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

func (s *Server) handleUpdateTeam(w http.ResponseWriter, r *http.Request) {
	var req updateTeamRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, r, http.StatusBadRequest, "INVALID_BODY", "Invalid request payload", nil)
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	req.Slug = strings.TrimSpace(strings.ToLower(req.Slug))
	if req.Name == "" || req.Slug == "" {
		writeError(w, r, http.StatusBadRequest, "MISSING_FIELDS", "Team name and slug are required", nil)
		return
	}

	authCtx, _ := AuthFromContext(r.Context())
	orgID, _ := ToPGUUID(authCtx.OrgID)

	org, err := s.Queries.UpdateOrganization(r.Context(), dbgen.UpdateOrganizationParams{
		ID:   orgID,
		Name: req.Name,
		Slug: req.Slug,
	})
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			writeError(w, r, http.StatusConflict, "SLUG_TAKEN", "That team slug is already taken", nil)
			return
		}
		writeError(w, r, http.StatusInternalServerError, "UPDATE_FAILED", "Could not update team", nil)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"id":   UUIDString(org.ID),
		"name": org.Name,
		"slug": org.Slug,
	})
	s.invalidateWorkspaceCache(authCtx.OrgID)
}

func (s *Server) handleDeleteTeam(w http.ResponseWriter, r *http.Request) {
	authCtx, _ := AuthFromContext(r.Context())
	orgID, _ := ToPGUUID(authCtx.OrgID)

	ctx := r.Context()
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "TX_FAILED", "Could not start transaction", nil)
		return
	}
	defer tx.Rollback(ctx)

	txq := dbgen.New(tx)

	_ = txq.DeleteInvitesByOrg(ctx, orgID)
	_ = txq.DeleteQuotasByOrg(ctx, orgID)
	_ = txq.ClearOrgUsers(ctx, orgID)
	_ = txq.DeleteRolesByOrg(ctx, orgID)

	if err := txq.DeleteOrganization(ctx, orgID); err != nil {
		writeError(w, r, http.StatusInternalServerError, "DELETE_FAILED", "Could not delete team", nil)
		return
	}

	if err := tx.Commit(ctx); err != nil {
		writeError(w, r, http.StatusInternalServerError, "COMMIT_FAILED", "Could not finalize deletion", nil)
		return
	}

	s.refreshSessionOrg(w, r, "")
	s.invalidateWorkspaceCache(authCtx.OrgID)

	w.WriteHeader(http.StatusNoContent)
}

func resolveRoleByKey(ctx context.Context, s *Server, orgID pgtype.UUID, roleKey string) (dbgen.Role, error) {
	candidates := []string{roleKey}
	switch roleKey {
	case "owner":
		candidates = append(candidates, "admin")
	case "viewer":
		candidates = append(candidates, "auditor")
	}

	for _, candidate := range candidates {
		role, err := s.Queries.GetRoleByOrgAndKey(ctx, dbgen.GetRoleByOrgAndKeyParams{OrgID: orgID, Key: candidate})
		if err == nil {
			return role, nil
		}
	}
	return dbgen.Role{}, fmt.Errorf("role %s not found", roleKey)
}

func (s *Server) refreshSessionOrg(w http.ResponseWriter, r *http.Request, orgID string) {
	cookie, err := r.Cookie(s.Config.SessionCookieName)
	if err != nil {
		return
	}
	s.sessionCache.Delete(cookie.Value)
	session, err := s.Sessions.Get(r.Context(), cookie.Value)
	if err != nil {
		return
	}
	session.OrgID = orgID
	session.CachedAt = 0
	_ = s.Sessions.Save(r.Context(), session, s.Config.SessionTTL)
	s.setSessionCookies(w, session)
}

const inviteAlphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"

func generateInviteCode(length int) string {
	result := make([]byte, length)
	for i := range result {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(inviteAlphabet))))
		result[i] = inviteAlphabet[n.Int64()]
	}
	return string(result)
}
