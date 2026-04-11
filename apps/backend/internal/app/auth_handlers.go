package app

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/redis/go-redis/v9"

	verauth "github.com/verin/dms/apps/backend/internal/auth"
	"github.com/verin/dms/apps/backend/internal/dbgen"
)

const sessionCacheTTL = 5 * time.Minute

func (s *Server) loadSession(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(s.Config.SessionCookieName)
		if err != nil || cookie.Value == "" {
			next.ServeHTTP(w, r)
			return
		}

		sessionID := cookie.Value

		if cached, ok := s.sessionCache.Get(sessionID); ok {
			ctx := context.WithValue(r.Context(), AuthContextKey, cached)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		session, err := s.Sessions.Get(r.Context(), sessionID)
		if err != nil {
			if !errors.Is(err, redis.Nil) {
				s.Logger.Warn().Err(err).Msg("load session")
			}
			next.ServeHTTP(w, r)
			return
		}

		if !session.Authenticated {
			next.ServeHTTP(w, r)
			return
		}

		needsRefresh := session.CachedAt == 0 || time.Since(time.Unix(session.CachedAt, 0)) > sessionCacheTTL
		if needsRefresh {
			if err := s.refreshSessionCache(r.Context(), &session); err != nil {
				next.ServeHTTP(w, r)
				return
			}
		}

		authContext := AuthContext{
			UserID:        session.UserID,
			OrgID:         session.OrgID,
			Email:         session.Email,
			FullName:      session.FullName,
			AvatarURL:     session.AvatarURL,
			Roles:         session.Roles,
			Permissions:   session.Permissions,
			IsAdmin:       session.IsAdmin,
			CSRFToken:     session.CSRFTok,
			Authenticated: session.Authenticated,
			RoleIDs:       make([]pgtype.UUID, 0),
		}

		s.sessionCache.Set(sessionID, authContext)

		http.SetCookie(w, &http.Cookie{
			Name:     s.Config.CSRFCookieName,
			Value:    session.CSRFTok,
			Path:     "/",
			HttpOnly: false,
			SameSite: cookieSameSite(s.Config.AppEnv),
			Secure:   s.Config.AppEnv != "development",
			Expires:  session.ExpiresAt,
		})

		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), AuthContextKey, authContext)))
	})
}

func (s *Server) refreshSessionCache(ctx context.Context, session *verauth.Session) error {
	userID, err := ToPGUUID(session.UserID)
	if err != nil {
		return err
	}

	user, err := s.Queries.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}

	roles, _ := s.Queries.ListUserRoles(ctx, userID)
	roleNames := make([]string, 0, len(roles))
	pgRoleIDs := make([]pgtype.UUID, 0, len(roles))
	for _, role := range roles {
		roleNames = append(roleNames, role.Key)
		pgRoleIDs = append(pgRoleIDs, role.ID)
	}

	permissions, _ := s.Queries.ListPermissionsByRoleIDs(ctx, pgRoleIDs)
	permKeys := make([]string, 0, len(permissions))
	isAdmin := false
	for _, p := range permissions {
		permKeys = append(permKeys, p.Key)
		if p.Key == "admin.manage" {
			isAdmin = true
		}
	}

	session.Email = user.Email
	session.FullName = user.FullName
	session.AvatarURL = user.AvatarUrl
	session.OrgID = UUIDString(user.OrgID)
	session.Roles = roleNames
	session.Permissions = permKeys
	session.IsAdmin = isAdmin
	session.CachedAt = time.Now().Unix()

	_ = s.Sessions.Save(ctx, *session, s.Config.SessionTTL)
	return nil
}

func (s *Server) invalidateSessionCache(sessionID string) {
	s.sessionCache.Delete(sessionID)
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(s.Config.SessionCookieName)
	if err == nil && cookie.Value != "" {
		s.invalidateSessionCache(cookie.Value)
		_ = s.Sessions.Delete(r.Context(), cookie.Value)
	}
	s.clearSessionCookies(w)
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleMe(w http.ResponseWriter, r *http.Request) {
	authContext, ok := AuthFromContext(r.Context())
	if !ok || !authContext.Authenticated {
		writeJSON(w, http.StatusOK, map[string]any{
			"authenticated": false,
			"csrfToken":     "",
		})
		return
	}

	hasWorkspace := authContext.OrgID != ""
	workspace := map[string]any(nil)
	if hasWorkspace {
		snapshot, err := s.getWorkspaceSnapshot(r.Context(), authContext.OrgID)
		if err == nil {
			workspace = map[string]any{
				"id":          snapshot.ID,
				"name":        snapshot.Name,
				"slug":        snapshot.Slug,
				"memberCount": snapshot.MemberCount,
			}
		}
	}

	normalizedRoles := normalizeRoles(authContext.Roles)

	writeJSON(w, http.StatusOK, map[string]any{
		"authenticated": true,
		"csrfToken":     authContext.CSRFToken,
		"hasWorkspace":  hasWorkspace,
		"workspace":     workspace,
		"user": map[string]any{
			"id":        authContext.UserID,
			"email":     authContext.Email,
			"fullName":  authContext.FullName,
			"avatarUrl": authContext.AvatarURL,
			"status":    "active",
			"role":      normalizedPrimaryRole(authContext.Roles),
			"roles":     normalizedRoles,
		},
	})
}

func (s *Server) setSessionCookies(w http.ResponseWriter, session verauth.Session) {
	sameSite := cookieSameSite(s.Config.AppEnv)
	secure := s.Config.AppEnv != "development"
	http.SetCookie(w, &http.Cookie{
		Name:     s.Config.SessionCookieName,
		Value:    session.ID,
		Path:     "/",
		HttpOnly: true,
		SameSite: sameSite,
		Secure:   secure,
		Expires:  session.ExpiresAt,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     s.Config.CSRFCookieName,
		Value:    session.CSRFTok,
		Path:     "/",
		HttpOnly: false,
		SameSite: sameSite,
		Secure:   secure,
		Expires:  session.ExpiresAt,
	})
}

func cookieSameSite(appEnv string) http.SameSite {
	if appEnv == "development" {
		return http.SameSiteLaxMode
	}
	return http.SameSiteNoneMode
}

func (s *Server) clearSessionCookies(w http.ResponseWriter) {
	expired := time.Unix(0, 0)
	http.SetCookie(w, &http.Cookie{
		Name:     s.Config.SessionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Expires:  expired,
		MaxAge:   -1,
	})
	http.SetCookie(w, &http.Cookie{
		Name:    s.Config.CSRFCookieName,
		Value:   "",
		Path:    "/",
		Expires: expired,
		MaxAge:  -1,
	})
}

func (s *Server) handleDemoLogin(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	email := "demo@verin.app"

	user, err := s.Queries.GetUserByEmail(ctx, email)
	if err != nil {
		user, err = s.Queries.CreateGoogleUser(ctx, dbgen.CreateGoogleUserParams{
			Email:    email,
			FullName: "Demo User",
			GoogleID: pgtype.Text{Valid: false},
		})
		if err != nil {
			writeError(w, r, http.StatusInternalServerError, "SEED_FAILED", "Could not create demo user", nil)
			return
		}
	}

	session := verauth.Session{
		UserID:        UUIDString(user.ID),
		OrgID:         UUIDString(user.OrgID),
		CSRFTok:       uuid.NewString(),
		Authenticated: true,
	}

	session, err = s.Sessions.Create(ctx, session, s.Config.SessionTTL)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "SESSION_FAILED", "Could not create session", nil)
		return
	}

	_ = s.Queries.UpdateUserLastLogin(ctx, user.ID)
	s.setSessionCookies(w, session)
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":       true,
		"redirect": "/home",
	})
}
