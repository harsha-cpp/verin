package app

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"

	"github.com/verin/dms/apps/backend/internal/auth"
	"github.com/verin/dms/apps/backend/internal/config"
	"github.com/verin/dms/apps/backend/internal/dbgen"
	"github.com/verin/dms/apps/backend/internal/storage"
)

type Server struct {
	Config         config.Config
	Logger         zerolog.Logger
	DB             *pgxpool.Pool
	Queries        *dbgen.Queries
	Redis          *redis.Client
	Storage        storage.Client
	Sessions       *auth.SessionStore
	sessionCache   *MemCache[AuthContext]
	workspaceCache *MemCache[WorkspaceSnapshot]
	taskRunner     *TaskRunner
}

func NewServer(
	cfg config.Config,
	logger zerolog.Logger,
	db *pgxpool.Pool,
	redisClient *redis.Client,
	storageClient storage.Client,
) *Server {
	server := &Server{
		Config:         cfg,
		Logger:         logger,
		DB:             db,
		Queries:        dbgen.New(db),
		Redis:          redisClient,
		Storage:        storageClient,
		Sessions:       auth.NewSessionStore(redisClient),
		sessionCache:   NewMemCache[AuthContext](30 * time.Second),
		workspaceCache: NewMemCache[WorkspaceSnapshot](30 * time.Second),
	}
	server.taskRunner = NewTaskRunner(server, 4)
	return server
}

func (s *Server) StartBackgroundTasks(ctx context.Context) {
	if s.taskRunner == nil {
		return
	}
	s.taskRunner.ResumePending(ctx)
}

func (s *Server) Router() http.Handler {
	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Recoverer)
	router.Use(httprate.LimitByRealIP(120, 1))
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{s.Config.WebOrigin},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		AllowCredentials: true,
	}))
	router.Use(s.logRequests)
	router.Use(s.loadSession)
	router.Use(s.verifyCSRF)

	router.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	router.Route("/api/v1", func(r chi.Router) {
		r.Route("/auth", func(ar chi.Router) {
			ar.Get("/google", s.handleGoogleRedirect)
			ar.Get("/google/callback", s.handleGoogleCallback)
			ar.Post("/demo-login", s.handleDemoLogin)
			ar.Post("/signup", s.handleSignup)
			ar.Post("/login", s.handleLogin)
			ar.Post("/logout", s.handleLogout)
			ar.Get("/me", s.handleMe)
		})

		r.Group(func(authed chi.Router) {
			authed.Use(s.requireAuthenticated)

			authed.Get("/home", s.handleHome)
			authed.Post("/teams", s.handleCreateTeam)
			authed.Post("/teams/join", s.handleJoinTeam)

			authed.Group(func(protected chi.Router) {
				protected.Use(s.requireTeam)

				protected.Post("/teams/invite", s.handleCreateInvite)
				protected.Get("/teams/members", s.handleListMembers)
				protected.Get("/teams/info", s.handleGetTeamInfo)
				protected.Put("/teams", s.handleUpdateTeam)
				protected.Delete("/teams", s.handleDeleteTeam)
				protected.Put("/teams/members/{memberID}/role", s.handleUpdateMemberRole)
				protected.Delete("/teams/members/{memberID}", s.handleRemoveMember)

				protected.Route("/documents", func(dr chi.Router) {
					dr.Get("/", s.handleListDocuments)
					dr.Post("/upload", s.handleDirectUpload)
					dr.Post("/init-upload", s.handleInitUpload)
					dr.Post("/complete-upload", s.handleCompleteUpload)
					dr.Get("/{documentID}", s.handleGetDocument)
					dr.Patch("/{documentID}", s.handleUpdateDocument)
					dr.Delete("/{documentID}", s.handleDeleteDocument)
					dr.Post("/{documentID}/archive", s.handleArchiveDocument)
					dr.Post("/{documentID}/restore", s.handleRestoreDocument)
					dr.Post("/{documentID}/download", s.handleSignedDownload)
					dr.Get("/{documentID}/versions", s.handleListVersions)
					dr.Post("/{documentID}/versions/{versionID}/restore", s.handleRestoreVersion)
					dr.Route("/{documentID}/comments", func(cr chi.Router) {
						cr.Get("/", s.handleListComments)
						cr.Post("/", s.handleCreateComment)
					})
					dr.Post("/{documentID}/share", s.handleShareDocument)
					dr.Delete("/{documentID}/share", s.handleRevokeShare)
				})

				protected.Get("/shared", s.handleListSharedDocuments)

				protected.Route("/spaces", func(cr chi.Router) {
					cr.Get("/", s.handleListCollections)
					cr.Post("/", s.handleCreateCollection)
					cr.Get("/{collectionID}", s.handleGetCollection)
					cr.Patch("/{collectionID}", s.handleUpdateCollection)
					cr.Delete("/{collectionID}", s.handleDeleteCollection)
					cr.Get("/{collectionID}/documents", s.handleListCollectionDocuments)
					cr.Post("/{collectionID}/members", s.handleAddCollectionMember)
					cr.Delete("/{collectionID}/members/{userID}", s.handleRemoveCollectionMember)
				})

				protected.Get("/search", s.handleSearch)

				protected.Get("/notifications", s.handleListNotifications)
				protected.Post("/notifications/{notificationID}/read", s.handleMarkNotificationRead)
			})
		})
	})

	return router
}

func (s *Server) logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		dur := time.Since(start)
		s.Logger.Info().
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Dur("latency", dur).
			Str("request_id", middleware.GetReqID(r.Context())).
			Msg("http request")
	})
}

func (s *Server) verifyCSRF(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet || r.Method == http.MethodOptions {
			next.ServeHTTP(w, r)
			return
		}

		if strings.HasPrefix(r.URL.Path, "/api/v1/auth/") {
			next.ServeHTTP(w, r)
			return
		}

		session, ok := AuthFromContext(r.Context())
		if !ok || session.CSRFToken == "" {
			writeError(w, r, http.StatusUnauthorized, "AUTH_REQUIRED", "Authentication required", nil)
			return
		}

		if r.Header.Get("X-CSRF-Token") != session.CSRFToken {
			writeError(w, r, http.StatusForbidden, "CSRF_INVALID", "Invalid CSRF token", nil)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (s *Server) requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.requireAuthenticated(http.HandlerFunc(next)).ServeHTTP(w, r)
	}
}

func (s *Server) requireAuthenticated(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authContext, ok := AuthFromContext(r.Context())
		if !ok || !authContext.Authenticated {
			writeError(w, r, http.StatusUnauthorized, "AUTH_REQUIRED", "Authentication required", nil)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) requireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authContext, ok := AuthFromContext(r.Context())
		if !ok || !authContext.IsAdmin {
			writeError(w, r, http.StatusForbidden, "FORBIDDEN", "Admin access required", nil)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) requireTeam(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authContext, ok := AuthFromContext(r.Context())
		if !ok || authContext.OrgID == "" {
			writeError(w, r, http.StatusForbidden, "TEAM_REQUIRED", "You must join or create a team first", nil)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) enqueueJob(ctx context.Context, jobType string, payload any, versionID string) (string, error) {
	var versionUUID any
	if strings.TrimSpace(versionID) != "" {
		versionUUID = MustPGUUID(versionID)
	}
	return s.taskRunner.Enqueue(ctx, jobType, payload, versionUUID)
}

func (s *Server) enqueueDetachedJob(ctx context.Context, jobType string, payload any) (string, error) {
	return s.taskRunner.Enqueue(ctx, jobType, payload, nil)
}
