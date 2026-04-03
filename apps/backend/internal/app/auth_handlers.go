package app

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pquerna/otp/totp"
	"github.com/redis/go-redis/v9"

	verauth "github.com/verin/dms/apps/backend/internal/auth"
	"github.com/verin/dms/apps/backend/internal/dbgen"
)

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type mfaVerifyRequest struct {
	Code string `json:"code"`
}

func (s *Server) loadSession(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(s.Config.SessionCookieName)
		if err != nil || cookie.Value == "" {
			next.ServeHTTP(w, r)
			return
		}

		session, err := s.Sessions.Get(r.Context(), cookie.Value)
		if err != nil {
			if !errors.Is(err, redis.Nil) {
				s.Logger.Warn().Err(err).Msg("load session")
			}
			next.ServeHTTP(w, r)
			return
		}

		userID, err := ToPGUUID(session.UserID)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}

		user, err := s.Queries.GetUserByID(r.Context(), userID)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}

		roles, _ := s.Queries.ListUserRoles(r.Context(), userID)
		roleIDs := make([]pgtype.UUID, 0, len(roles))
		roleNames := make([]string, 0, len(roles))
		for _, role := range roles {
			roleIDs = append(roleIDs, role.ID)
			roleNames = append(roleNames, role.Key)
		}

		permissions, _ := s.Queries.ListPermissionsByRoleIDs(r.Context(), roleIDs)
		permissionKeys := make([]string, 0, len(permissions))
		isAdmin := false
		for _, permission := range permissions {
			permissionKeys = append(permissionKeys, permission.Key)
			if permission.Key == "admin.manage" {
				isAdmin = true
			}
		}

		authContext := AuthContext{
			UserID:        session.UserID,
			OrgID:         session.OrgID,
			Email:         user.Email,
			FullName:      user.FullName,
			Roles:         roleNames,
			RoleIDs:       roleIDs,
			Permissions:   permissionKeys,
			IsAdmin:       isAdmin,
			CSRFToken:     session.CSRFTok,
			Authenticated: session.Authenticated,
		}

		http.SetCookie(w, &http.Cookie{
			Name:     s.Config.CSRFCookieName,
			Value:    session.CSRFTok,
			Path:     "/",
			HttpOnly: false,
			SameSite: http.SameSiteLaxMode,
			Secure:   s.Config.AppEnv != "development",
			Expires:  session.ExpiresAt,
		})

		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), AuthContextKey, authContext)))
	})
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	var request loginRequest
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, r, http.StatusBadRequest, "INVALID_BODY", "Invalid request payload", nil)
		return
	}

	user, err := s.Queries.GetUserByEmail(r.Context(), request.Email)
	if err != nil {
		writeError(w, r, http.StatusUnauthorized, "INVALID_CREDENTIALS", "Invalid email or password", nil)
		return
	}

	if user.LockedUntil.Valid && user.LockedUntil.Time.After(time.Now()) {
		writeError(w, r, http.StatusForbidden, "ACCOUNT_LOCKED", "Account is temporarily locked due to too many failed attempts", map[string]any{
			"lockedUntil": user.LockedUntil.Time.Format(time.RFC3339),
		})
		return
	}

	if !verauth.VerifyPassword(user.PasswordHash, request.Password) {
		_ = s.Queries.IncrementFailedLoginAttempts(r.Context(), user.ID)
		if user.FailedLoginAttempts+1 >= s.Config.LockoutThreshold {
			lockUntil := time.Now().Add(s.Config.LockoutDuration)
			_ = s.Queries.LockUserAccount(r.Context(), dbgen.LockUserAccountParams{
				ID:          user.ID,
				LockedUntil: pgtype.Timestamptz{Time: lockUntil, Valid: true},
			})
		}
		writeError(w, r, http.StatusUnauthorized, "INVALID_CREDENTIALS", "Invalid email or password", nil)
		return
	}

	_ = s.Queries.ResetFailedLoginAttempts(r.Context(), user.ID)

	roles, _ := s.Queries.ListUserRoles(r.Context(), user.ID)
	session := verauth.Session{
		UserID:        UUIDString(user.ID),
		OrgID:         UUIDString(user.OrgID),
		CSRFTok:       uuid.NewString(),
		Authenticated: !user.MfaEnabled,
		PendingMFA:    user.MfaEnabled,
	}

	ttl := s.Config.SessionTTL
	if user.MfaEnabled {
		ttl = s.Config.AuthPendingTTL
	}

	session, err = s.Sessions.Create(r.Context(), session, ttl)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "SESSION_CREATE_FAILED", "Could not create session", nil)
		return
	}

	s.setSessionCookies(w, session)
	if !user.MfaEnabled {
		_ = s.Queries.UpdateUserLastLogin(r.Context(), user.ID)
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"user":        presentUser(user, roles),
		"requiresMfa": user.MfaEnabled,
		"csrfToken":   session.CSRFTok,
	})
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(s.Config.SessionCookieName)
	if err == nil && cookie.Value != "" {
		_ = s.Sessions.Delete(r.Context(), cookie.Value)
	}

	s.clearSessionCookies(w)
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleMe(w http.ResponseWriter, r *http.Request) {
	authContext, ok := AuthFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusOK, map[string]any{
			"authenticated": false,
			"csrfToken":     "",
		})
		return
	}

	userID, err := ToPGUUID(authContext.UserID)
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"authenticated": false,
			"csrfToken":     "",
		})
		return
	}

	user, err := s.Queries.GetUserByID(r.Context(), userID)
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"authenticated": false,
			"csrfToken":     "",
		})
		return
	}

	roles, _ := s.Queries.ListUserRoles(r.Context(), userID)
	writeJSON(w, http.StatusOK, map[string]any{
		"authenticated": authContext.Authenticated,
		"csrfToken":     authContext.CSRFToken,
		"user":          presentUser(user, roles),
	})
}

func (s *Server) handleMFASetup(w http.ResponseWriter, r *http.Request) {
	authContext, _ := AuthFromContext(r.Context())
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "Verin DMS",
		AccountName: authContext.Email,
	})
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "MFA_SETUP_FAILED", "Could not initialize MFA", nil)
		return
	}

	cookie, err := r.Cookie(s.Config.SessionCookieName)
	if err != nil {
		writeError(w, r, http.StatusUnauthorized, "AUTH_REQUIRED", "Authentication required", nil)
		return
	}

	session, err := s.Sessions.Get(r.Context(), cookie.Value)
	if err != nil {
		writeError(w, r, http.StatusUnauthorized, "AUTH_REQUIRED", "Authentication required", nil)
		return
	}

	session.PendingMFASecret = key.Secret()
	session.PendingMFASetup = true
	if err := s.Sessions.Save(r.Context(), session, s.Config.SessionTTL); err != nil {
		writeError(w, r, http.StatusInternalServerError, "SESSION_SAVE_FAILED", "Could not persist MFA setup", nil)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"secret":     key.Secret(),
		"otpauthUrl": key.URL(),
	})
}

func (s *Server) handleMFAVerify(w http.ResponseWriter, r *http.Request) {
	var request mfaVerifyRequest
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, r, http.StatusBadRequest, "INVALID_BODY", "Invalid request payload", nil)
		return
	}

	cookie, err := r.Cookie(s.Config.SessionCookieName)
	if err != nil {
		writeError(w, r, http.StatusUnauthorized, "AUTH_REQUIRED", "Authentication required", nil)
		return
	}

	session, err := s.Sessions.Get(r.Context(), cookie.Value)
	if err != nil {
		writeError(w, r, http.StatusUnauthorized, "AUTH_REQUIRED", "Authentication required", nil)
		return
	}

	userID, err := ToPGUUID(session.UserID)
	if err != nil {
		writeError(w, r, http.StatusUnauthorized, "AUTH_REQUIRED", "Authentication required", nil)
		return
	}

	user, err := s.Queries.GetUserByID(r.Context(), userID)
	if err != nil {
		writeError(w, r, http.StatusUnauthorized, "AUTH_REQUIRED", "Authentication required", nil)
		return
	}

	secret := session.PendingMFASecret
	if secret == "" {
		secret, err = verauth.DecryptSecret(s.Config.MFAEncryptionKey, user.MfaSecretEncrypted)
		if err != nil {
			writeError(w, r, http.StatusBadRequest, "MFA_NOT_CONFIGURED", "MFA is not configured for this account", nil)
			return
		}
	}

	if !totp.Validate(request.Code, secret) {
		writeError(w, r, http.StatusUnauthorized, "MFA_INVALID", "Invalid MFA code", nil)
		return
	}

	if session.PendingMFASetup {
		encrypted, err := verauth.EncryptSecret(s.Config.MFAEncryptionKey, secret)
		if err != nil {
			writeError(w, r, http.StatusInternalServerError, "MFA_STORE_FAILED", "Could not save MFA secret", nil)
			return
		}

		if err := s.Queries.UpdateUserMFA(r.Context(), dbgen.UpdateUserMFAParams{
			ID:                 user.ID,
			MfaEnabled:         true,
			MfaSecretEncrypted: encrypted,
		}); err != nil {
			writeError(w, r, http.StatusInternalServerError, "MFA_STORE_FAILED", "Could not save MFA secret", nil)
			return
		}

		user.MfaEnabled = true
		user.MfaSecretEncrypted = encrypted
	}

	session.Authenticated = true
	session.PendingMFA = false
	session.PendingMFASecret = ""
	session.PendingMFASetup = false
	session.ExpiresAt = time.Now().Add(s.Config.SessionTTL)
	if err := s.Sessions.Save(r.Context(), session, s.Config.SessionTTL); err != nil {
		writeError(w, r, http.StatusInternalServerError, "SESSION_SAVE_FAILED", "Could not update session", nil)
		return
	}

	_ = s.Queries.UpdateUserLastLogin(r.Context(), user.ID)
	roles, _ := s.Queries.ListUserRoles(r.Context(), user.ID)
	s.setSessionCookies(w, session)

	writeJSON(w, http.StatusOK, map[string]any{
		"user":        presentUser(user, roles),
		"requiresMfa": false,
		"csrfToken":   session.CSRFTok,
	})
}

func (s *Server) setSessionCookies(w http.ResponseWriter, session verauth.Session) {
	http.SetCookie(w, &http.Cookie{
		Name:     s.Config.SessionCookieName,
		Value:    session.ID,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   s.Config.AppEnv != "development",
		Expires:  session.ExpiresAt,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     s.Config.CSRFCookieName,
		Value:    session.CSRFTok,
		Path:     "/",
		HttpOnly: false,
		SameSite: http.SameSiteLaxMode,
		Secure:   s.Config.AppEnv != "development",
		Expires:  session.ExpiresAt,
	})
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

func presentUser(user dbgen.User, roles []dbgen.Role) map[string]any {
	serializedRoles := make([]map[string]any, 0, len(roles))
	for _, role := range roles {
		serializedRoles = append(serializedRoles, map[string]any{
			"id":          UUIDString(role.ID),
			"key":         role.Key,
			"name":        role.Name,
			"description": role.Description,
		})
	}

	return map[string]any{
		"id":         UUIDString(user.ID),
		"email":      user.Email,
		"fullName":   user.FullName,
		"status":     user.Status,
		"mfaEnabled": user.MfaEnabled,
		"roles":      serializedRoles,
	}
}
