package app

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	verauth "github.com/verin/dms/apps/backend/internal/auth"
	"github.com/verin/dms/apps/backend/internal/dbgen"
)

func (s *Server) googleOAuthConfig() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     s.Config.GoogleClientID,
		ClientSecret: s.Config.GoogleClientSecret,
		RedirectURL:  s.Config.GoogleRedirectURL,
		Scopes:       []string{"openid", "email", "profile"},
		Endpoint:     google.Endpoint,
	}
}

func (s *Server) handleGoogleRedirect(w http.ResponseWriter, r *http.Request) {
	stateBytes := make([]byte, 32)
	if _, err := rand.Read(stateBytes); err != nil {
		writeError(w, r, http.StatusInternalServerError, "STATE_FAILED", "Could not generate state", nil)
		return
	}
	state := base64.URLEncoding.EncodeToString(stateBytes)

	if err := s.Redis.Set(r.Context(), "verin:oauth:state:"+state, "1", 10*time.Minute).Err(); err != nil {
		writeError(w, r, http.StatusInternalServerError, "STATE_FAILED", "Could not store state", nil)
		return
	}

	url := s.googleOAuthConfig().AuthCodeURL(state, oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

type googleUserInfo struct {
	ID      string `json:"id"`
	Email   string `json:"email"`
	Name    string `json:"name"`
	Picture string `json:"picture"`
}

func (s *Server) handleGoogleCallback(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	state := r.URL.Query().Get("state")
	if state == "" {
		s.redirectWithError(w, r, "missing_state")
		return
	}

	result, err := s.Redis.GetDel(ctx, "verin:oauth:state:"+state).Result()
	if err != nil || result == "" {
		s.redirectWithError(w, r, "invalid_state")
		return
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		s.redirectWithError(w, r, "missing_code")
		return
	}

	token, err := s.googleOAuthConfig().Exchange(ctx, code)
	if err != nil {
		s.Logger.Error().Err(err).Msg("oauth code exchange failed")
		s.redirectWithError(w, r, "exchange_failed")
		return
	}

	userInfo, err := fetchGoogleUserInfo(ctx, token.AccessToken)
	if err != nil {
		s.Logger.Error().Err(err).Msg("fetch google user info failed")
		s.redirectWithError(w, r, "userinfo_failed")
		return
	}

	googleIDText := pgtype.Text{String: userInfo.ID, Valid: true}
	user, err := s.Queries.GetUserByGoogleID(ctx, googleIDText)
	if err != nil {
		user, err = s.Queries.CreateGoogleUser(ctx, dbgen.CreateGoogleUserParams{
			Email:     userInfo.Email,
			FullName:  userInfo.Name,
			GoogleID:  googleIDText,
			AvatarUrl: userInfo.Picture,
		})
		if err != nil {
			s.Logger.Error().Err(err).Msg("create google user failed")
			s.redirectWithError(w, r, "create_user_failed")
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
		s.redirectWithError(w, r, "session_failed")
		return
	}

	_ = s.Queries.UpdateUserLastLogin(ctx, user.ID)
	s.setSessionCookies(w, session)

	redirectPath := "/home"
	if !user.OrgID.Valid {
		redirectPath = "/onboarding"
	}
	http.Redirect(w, r, s.Config.WebOrigin+redirectPath, http.StatusTemporaryRedirect)
}

func (s *Server) redirectWithError(w http.ResponseWriter, r *http.Request, errorCode string) {
	http.Redirect(w, r, s.Config.WebOrigin+"/login?error="+errorCode, http.StatusTemporaryRedirect)
}

func fetchGoogleUserInfo(ctx context.Context, accessToken string) (*googleUserInfo, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://www.googleapis.com/oauth2/v2/userinfo", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("google userinfo returned %d: %s", resp.StatusCode, string(body))
	}

	var info googleUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, err
	}
	return &info, nil
}
