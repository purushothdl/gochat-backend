package http

import (
    "net/http"

    "github.com/gorilla/mux"
    "github.com/purushothdl/gochat-backend/internal/domain/auth"
    "github.com/purushothdl/gochat-backend/internal/domain/user"
    "github.com/purushothdl/gochat-backend/internal/transport/http/middleware"
)

type Router struct {
    authHandler *auth.Handler
    userHandler *user.Handler
    authMw      *middleware.AuthMiddleware
}

func NewRouter(authHandler *auth.Handler, userHandler *user.Handler, authMw *middleware.AuthMiddleware) *Router {
    return &Router{
        authHandler: authHandler,
        userHandler: userHandler,
        authMw:      authMw,
    }
}

func (rt *Router) SetupRoutes() *mux.Router {
    r := mux.NewRouter()

    // API prefix
    api := r.PathPrefix("/api").Subrouter()

    // Health check
    api.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusOK)
        w.Write([]byte(`{"status": "ok", "service": "gochat-backend"}`))
    }).Methods("GET")

    // Auth routes (public)
    authRoutes := api.PathPrefix("/auth").Subrouter()
    authRoutes.HandleFunc("/register", rt.authHandler.Register).Methods("POST")
    authRoutes.HandleFunc("/login", rt.authHandler.Login).Methods("POST")
    authRoutes.HandleFunc("/refresh", rt.authHandler.RefreshToken).Methods("POST")
    authRoutes.HandleFunc("/logout", rt.authHandler.Logout).Methods("POST")
    
    // Protected auth routes
    authRoutes.HandleFunc("/me", rt.authMw.RequireAuth(rt.authHandler.Me)).Methods("GET")

    // User routes (all protected)
    userRoutes := api.PathPrefix("/user").Subrouter()
    userRoutes.HandleFunc("/profile", rt.authMw.RequireAuth(rt.userHandler.GetProfile)).Methods("GET")
    userRoutes.HandleFunc("/profile", rt.authMw.RequireAuth(rt.userHandler.UpdateProfile)).Methods("PUT")
    userRoutes.HandleFunc("/settings", rt.authMw.RequireAuth(rt.userHandler.UpdateSettings)).Methods("PUT")
    userRoutes.HandleFunc("/password", rt.authMw.RequireAuth(rt.userHandler.ChangePassword)).Methods("PUT")

    // CORS middleware for development
    r.Use(func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            w.Header().Set("Access-Control-Allow-Origin", "*")
            w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
            w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
            w.Header().Set("Access-Control-Allow-Credentials", "true")

            if r.Method == "OPTIONS" {
                w.WriteHeader(http.StatusOK)
                return
            }

            next.ServeHTTP(w, r)
        })
    })

    return r
}
