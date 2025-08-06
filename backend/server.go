package main

//go:generate sqlc generate

import (
	"context"
	"crypto/rand"
	"database/sql"
	"embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/png"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/amacneil/dbmate/v2/pkg/dbmate"
	_ "github.com/amacneil/dbmate/v2/pkg/driver/sqlite"
	"github.com/chanbakjsd/CCDSQuickShop/backend/db"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/google"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stripe/stripe-go/v81/client"
)

//go:embed db/migrations/*.sql
var migrationFS embed.FS

type Server struct {
	Config  *ServerConfig
	DB      *sql.DB
	Queries *db.Queries
	Stripe  *client.API
}

func NewServer(cfg *ServerConfig) (*Server, error) {
	if cfg.authenticationOK() {
		goth.UseProviders(google.New(cfg.GoogleClientID, cfg.GoogleClientSecret, cfg.FrontendURL+"/api/v0/auth/callback"))
	} else {
		slog.Warn("client id or secret missing, admin authentication will not work")
	}
	connURL, err := url.Parse(cfg.Sqlite3ConnStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse connection URL: %w", err)
	}
	connURL.Scheme = "sqlite3"
	dbmateDB := dbmate.New(connURL)
	dbmateDB.FS = migrationFS
	if err := dbmateDB.CreateAndMigrate(); err != nil {
		return nil, fmt.Errorf("failed to create and migrate database: %w", err)
	}
	sqlDB, err := sql.Open("sqlite3", cfg.Sqlite3ConnStr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	var stripe *client.API
	if cfg.StripeSecretKey == "" {
		slog.Warn("stripe secret missing, payment will be skipped")
	} else {
		stripe = &client.API{}
		stripe.Init(cfg.StripeSecretKey, nil)
	}
	return &Server{
		Config:  cfg,
		DB:      sqlDB,
		Queries: db.New(sqlDB),
		Stripe:  stripe,
	}, nil
}

func (s *Server) HTTPMux() *http.ServeMux {
	mux := http.NewServeMux()
	// Called from frontend.
	mux.HandleFunc("GET /api/v0/sales/{sale_id}/coupons", s.Coupons)
	mux.HandleFunc("GET /api/v0/sales/{sale_id}/coupons/{id}", s.CouponLookup)
	mux.HandleFunc("GET /api/v0/sales/{sale_id}/products", s.Products)
	mux.HandleFunc("GET /api/v0/orders/{id}", s.OrderLookup)
	mux.HandleFunc("POST /api/v0/checkout", s.Checkout)
	mux.HandleFunc("POST /api/v0/checkout/stripe", s.StripeWebhook)
	mux.HandleFunc("GET /api/v0/checkout/complete", s.CheckoutComplete)
	// Admin paths.
	mux.HandleFunc("GET /api/v0/auth", s.Auth)
	mux.HandleFunc("GET /api/v0/auth/callback", s.AuthCallback)
	mux.HandleFunc("POST /api/v0/sales/{sale_id}/coupons", s.SaveCoupon)
	mux.HandleFunc("POST /api/v0/sales/{sale_id}/products", s.SaveProduct)
	mux.HandleFunc("POST /api/v0/image_upload", s.ImageUpload)
	mux.HandleFunc("POST /api/v0/orders/{id}/collect", s.OrderCollect)
	mux.HandleFunc("POST /api/v0/orders/{id}/cancel", s.OrderCancel)
	mux.HandleFunc("GET /api/v0/perm_check", s.PermissionCheck)
	mux.HandleFunc("GET /api/v0/users", s.AdminUsers)
	mux.HandleFunc("POST /api/v0/users", s.CreateAdminUser)
	mux.HandleFunc("DELETE /api/v0/users", s.DeleteAdminUser)
	mux.HandleFunc("GET /api/v0/closures", s.StoreClosures)
	mux.HandleFunc("POST /api/v0/closures", s.SaveStoreClosure)
	mux.HandleFunc("GET /api/v0/sales", s.SalePeriods)
	mux.HandleFunc("POST /api/v0/sales", s.SaveSalePeriod)
	mux.HandleFunc("DELETE /api/v0/closures/{id}", s.DeleteStoreClosure)
	mux.HandleFunc("GET /api/v0/sales/{sale_id}/order_summary", s.OrderSummary)
	mux.Handle("/api/", http.NotFoundHandler())
	if s.Config.ImageDir != "" {
		mux.Handle("GET /content/", http.StripPrefix("/content/", noDirListing(http.FileServer(http.Dir(s.Config.ImageDir)))))
	} else {
		slog.Warn("not serving image content, image directory not configured")
	}
	switch {
	case s.Config.Forwarder != nil && s.Config.StaticDir != nil:
		panic("forwarder and static directory both declared")
	case s.Config.Forwarder != nil:
		mux.Handle("/", s.Config.Forwarder)
	case s.Config.StaticDir != nil:
		mux.Handle("/", http.FileServer(singlePageAppFS{s.Config.StaticDir}))
	}
	return mux
}

type (
	User               string
	AdminUsersResponse struct {
		Users []User `json:"users"`
	}
)

func (s *Server) AdminUsers(w http.ResponseWriter, req *http.Request) {
	if !s.authCheck(w, req) {
		return
	}
	ctx := req.Context()
	dbUsers, err := s.Queries.ListAdminUsers(ctx)
	if err != nil {
		slog.Error("error fetching users", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	users := make([]User, 0, len(dbUsers))
	for _, u := range dbUsers {
		users = append(users, User(u))
	}
	if err := json.NewEncoder(w).Encode(AdminUsersResponse{
		Users: users,
	}); err != nil {
		slog.Error("error writing users response", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func (s *Server) CreateAdminUser(w http.ResponseWriter, req *http.Request) {
	if !s.authCheck(w, req) {
		return
	}
	var user User
	if err := json.NewDecoder(req.Body).Decode(&user); err != nil {
		slog.Error("error parsing request", "err", err)
		http.Error(w, "Invalid Body", http.StatusBadRequest)
		return
	}
	ctx := req.Context()
	if err := s.Queries.CreateAdminUser(ctx, string(user)); err != nil {
		slog.Error("error creating admin user", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) DeleteAdminUser(w http.ResponseWriter, req *http.Request) {
	if !s.authCheck(w, req) {
		return
	}
	var user User
	if err := json.NewDecoder(req.Body).Decode(&user); err != nil {
		slog.Error("error parsing request", "err", err)
		http.Error(w, "Invalid Body", http.StatusBadRequest)
		return
	}
	ctx := req.Context()
	if err := s.Queries.DeleteAdminUser(ctx, string(user)); err != nil {
		slog.Error("error deleting admin user", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

const (
	MaxImageSizeBytes        = 8 * 1024 * 1024
	MaxImageSizePixel        = 16384
	MaxImagePixelAfterResize = 1024
)

type ImageUploadResponse struct {
	URL string `json:"url"`
}

func (s *Server) ImageUpload(w http.ResponseWriter, req *http.Request) {
	if !s.authCheck(w, req) {
		return
	}
	skipSquare := false
	switch req.FormValue("raw") {
	case "1":
		skipSquare = true
	case "0", "":
		skipSquare = false
	default:
		http.Error(w, "Invalid raw value, only 0 and 1 are allowed", http.StatusBadRequest)
		return
	}
	req.Body = http.MaxBytesReader(w, req.Body, MaxImageSizeBytes)
	if err := req.ParseMultipartForm(MaxImageSizeBytes); err != nil {
		slog.Error("error parsing multipart form", "err", err)
		http.Error(w, "Invalid Request (is file too large?)", http.StatusBadRequest)
		return
	}
	file, _, err := req.FormFile("file")
	if err != nil {
		slog.Error("error parsing file field", "err", err)
		http.Error(w, "Invalid Request", http.StatusBadRequest)
		return
	}
	cfg, _, err := image.DecodeConfig(file)
	if err != nil {
		slog.Error("error decoding image", "err", err)
		http.Error(w, "Invalid Request", http.StatusBadRequest)
		return
	}
	if cfg.Width > MaxImageSizePixel || cfg.Height > MaxImageSizePixel {
		slog.Error("image size too big", "width", cfg.Width, "height", cfg.Height, "max", MaxImageSizePixel)
		http.Error(w, "Invalid Request", http.StatusBadRequest)
		return
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		slog.Error("error seeking to start", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	img, _, err := image.Decode(file)
	if err != nil {
		slog.Error("error decoding image", "err", err)
		http.Error(w, "Invalid Request", http.StatusBadRequest)
		return
	}
	imageID, err := randomImageID()
	if err != nil {
		slog.Error("error generating random image ID", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if skipSquare {
		img = scaleImage(img, MaxImagePixelAfterResize)
	} else {
		img = squareImage(img, MaxImagePixelAfterResize)
	}
	imageName := imageID + ".png"
	f, err := os.Create(path.Join(s.Config.ImageDir, imageName))
	if err != nil {
		slog.Error("error creating image file", "dir", s.Config.ImageDir, "image_name", imageName, "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer func() {
		err := f.Close()
		if err != nil {
			slog.Error("error closing image file", "dir", s.Config.ImageDir, "image_name", imageName, "err", err)
		}
	}()
	if err := png.Encode(f, img); err != nil {
		slog.Error("error encoding PNG file", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if err := json.NewEncoder(w).Encode(ImageUploadResponse{
		URL: s.Config.FrontendURL + "/content/" + imageName,
	}); err != nil {
		slog.Error("error writing image upload response", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func (s *Server) PermissionCheck(w http.ResponseWriter, req *http.Request) {
	if !s.authCheck(w, req) {
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) Auth(w http.ResponseWriter, req *http.Request) {
	if err := s.completeAuth(w, req); err == nil {
		return
	}
	if !s.Config.authenticationOK() {
		slog.Error("denying authentication anyways as authentication was previously configured")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	req = req.WithContext(context.WithValue(req.Context(), gothic.ProviderParamKey, "google"))
	gothic.BeginAuthHandler(w, req)
}

func (s *Server) AuthCallback(w http.ResponseWriter, req *http.Request) {
	err := s.completeAuth(w, req)
	if err != nil {
		slog.Error("error authenticating on callback", "err", err)
		http.Error(w, "Invalid authentication request", http.StatusBadRequest)
		return
	}
}

func (s *Server) authCheck(w http.ResponseWriter, req *http.Request) bool {
	if _, err := gothic.GetFromSession("user", req); err != nil {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return false
	}
	return true
}

func (s *Server) completeAuth(w http.ResponseWriter, req *http.Request) error {
	email := "GOOGLE_AUTH_NOT_CONFIGURED"
	if s.Config.authenticationOK() {
		req = req.WithContext(context.WithValue(req.Context(), gothic.ProviderParamKey, "google"))
		user, err := gothic.CompleteUserAuth(w, req)
		if err != nil {
			return err
		}
		email = user.Email
	} else {
		slog.Warn("potentially allowing admin login due to misconfigured Google authentication")
	}
	ctx := req.Context()
	ok, err := s.validAdminUser(ctx, email)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("invalid user: %s", email)
	}
	if err := gothic.StoreInSession("user", email, req, w); err != nil {
		return fmt.Errorf("failed to store session: %s", email)
	}
	http.Redirect(w, req, s.Config.FrontendURL+"/admin", http.StatusTemporaryRedirect)
	return nil
}

func (s *Server) validAdminUser(ctx context.Context, email string) (bool, error) {
	for range 5 {
		ok := false
		tx, err := s.DB.Begin()
		if err != nil {
			return false, fmt.Errorf("cannot start transaction: %w", err)
		}
		defer func() {
			if !ok {
				_ = tx.Rollback()
			}
		}()
		queries := s.Queries.WithTx(tx)
		count, err := queries.CountAdminUsers(ctx)
		if err != nil {
			return false, fmt.Errorf("cannot count admin users: %w", err)
		}
		if count == 0 {
			// Auto-create first admin user.
			if err := queries.CreateAdminUser(ctx, email); err != nil {
				return false, fmt.Errorf("cannot auto-create admin user: %w", err)
			}
		}
		_, authErr := queries.AuthAdminUser(ctx, email)
		if err := tx.Commit(); err != nil {
			slog.Warn("error commiting validAdminUser: %w", "err", err)
			continue
		}
		ok = true
		switch {
		case errors.Is(authErr, sql.ErrNoRows):
			return false, nil
		case authErr != nil:
			return false, authErr
		}
		return true, nil
	}
	return false, fmt.Errorf("too many transaction failures")
}

func (s *Server) resolveSalePeriod(w http.ResponseWriter, req *http.Request, period string) (periodID int64, ok bool) {
	var err error
	if period == "current" {
		periodID, err = s.Queries.CurrentSalePeriod(req.Context())
		if err != nil {
			slog.Error("error fetching current sale period", "err", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return 0, false
		}
		return periodID, true
	}
	if !s.authCheck(w, req) {
		return 0, false
	}
	periodID, err = strconv.ParseInt(period, 10, 64)
	if err != nil {
		http.Error(w, "Invalid sales period", http.StatusBadRequest)
		return 0, false
	}
	return periodID, true
}

func (cfg *ServerConfig) authenticationOK() bool {
	return cfg.GoogleClientID != "" || cfg.GoogleClientSecret != ""
}

func censorBack(s string, charCount int, maxAsterisk int, delimiter byte) string {
	idx := strings.IndexByte(s, delimiter)
	switch {
	case idx != -1:
		return s[:idx+1] + strings.Repeat("*", min(len(s)-idx-1, maxAsterisk))
	case len(s) < charCount:
		return s
	default:
		return s[:charCount] + strings.Repeat("*", min(len(s)-charCount, maxAsterisk))
	}
}

func censorFront(s string, charCount int, maxAsterisk int, delimiter byte) string {
	idx := strings.LastIndexByte(s, delimiter)
	switch {
	case idx != -1:
		return strings.Repeat("*", min(idx, maxAsterisk)) + s[idx:]
	case len(s) < charCount:
		return s
	default:
		remainingIdx := len(s) - charCount
		return strings.Repeat("*", min(remainingIdx, maxAsterisk)) + s[remainingIdx:]
	}
}

var _ http.FileSystem = singlePageAppFS{}

type singlePageAppFS struct {
	fs http.FileSystem
}

func (f singlePageAppFS) Open(path string) (http.File, error) {
	if f, err := f.fs.Open(path); err == nil {
		return f, nil
	}
	return f.fs.Open("index.html")
}

func noDirListing(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		if strings.HasSuffix(req.URL.Path, "/") || req.URL.Path == "" {
			http.NotFound(w, req)
			return
		}
		next.ServeHTTP(w, req)
	}
}

func randomImageID() (string, error) {
	var buf [8]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf[:]), nil
}
