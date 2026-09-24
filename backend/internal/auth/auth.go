package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"commercial-diving-decompression-control/backend/internal/util"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const (
	RolePlanner    = "planner"
	RoleSupervisor = "supervisor"
	RoleAdmin      = "admin"
)

type User struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Username     string    `gorm:"size:64;not null;uniqueIndex" json:"username"`
	PasswordHash string    `gorm:"size:128;not null" json:"-"`
	DisplayName  string    `gorm:"size:100;not null" json:"display_name"`
	Role         string    `gorm:"size:24;not null;index" json:"role"`
	Active       bool      `gorm:"not null;default:true" json:"active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (User) TableName() string { return "users" }

type LoginRequest struct {
	Username string `json:"username" binding:"required,min=3,max=64"`
	Password string `json:"password" binding:"required,min=8,max=128"`
}

type UserResponse struct {
	ID          uint   `json:"id"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	Role        string `json:"role"`
}

type LoginResponse struct {
	Token     string       `json:"token"`
	ExpiresAt time.Time    `json:"expires_at"`
	User      UserResponse `json:"user"`
}

type Claims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

type Repository struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) *Repository { return &Repository{db: db} }

func (r *Repository) FindByUsername(ctx context.Context, username string) (User, error) {
	var user User
	err := r.db.WithContext(ctx).Where("LOWER(username) = ?", strings.ToLower(strings.TrimSpace(username))).First(&user).Error
	if err != nil {
		return User{}, fmt.Errorf("find user by username: %w", err)
	}
	return user, nil
}

func (r *Repository) FindByID(ctx context.Context, id uint) (User, error) {
	var user User
	if err := r.db.WithContext(ctx).First(&user, id).Error; err != nil {
		return User{}, fmt.Errorf("find user by id: %w", err)
	}
	return user, nil
}

type Service struct {
	repo   *Repository
	secret []byte
	ttl    time.Duration
}

func NewService(repo *Repository, secret string, ttl time.Duration) *Service {
	return &Service{repo: repo, secret: []byte(secret), ttl: ttl}
}

func (s *Service) Login(ctx context.Context, req LoginRequest) (LoginResponse, error) {
	user, err := s.repo.FindByUsername(ctx, req.Username)
	if err != nil {
		return LoginResponse{}, util.Unauthorized("username or password is incorrect")
	}
	if !user.Active {
		return LoginResponse{}, util.Forbidden("account is inactive")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return LoginResponse{}, util.Unauthorized("username or password is incorrect")
	}
	expiresAt := time.Now().UTC().Add(s.ttl)
	claims := Claims{
		UserID: user.ID, Username: user.Username, Role: user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   fmt.Sprintf("%d", user.ID),
			Issuer:    "commercial-diving-decompression-control",
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(s.secret)
	if err != nil {
		return LoginResponse{}, util.Internal(fmt.Errorf("sign jwt: %w", err))
	}
	return LoginResponse{
		Token: token, ExpiresAt: expiresAt,
		User: UserResponse{ID: user.ID, Username: user.Username, DisplayName: user.DisplayName, Role: user.Role},
	}, nil
}

func (s *Service) Parse(raw string) (Claims, error) {
	parsed, err := jwt.ParseWithClaims(raw, &Claims{}, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, errors.New("unexpected jwt signing method")
		}
		return s.secret, nil
	})
	if err != nil || !parsed.Valid {
		return Claims{}, util.Unauthorized("session token is invalid or expired")
	}
	claims, ok := parsed.Claims.(*Claims)
	if !ok || claims.UserID == 0 || claims.Role == "" {
		return Claims{}, util.Unauthorized("session token is missing claims")
	}
	return *claims, nil
}

// CurrentPrincipal revalidates the account after JWT verification so role and
// active status changes take effect before every protected request.
func (s *Service) CurrentPrincipal(ctx context.Context, id uint) (User, error) {
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return User{}, util.Unauthorized("session account is unavailable")
	}
	if !user.Active {
		return User{}, util.Unauthorized("session account is inactive")
	}
	return user, nil
}

type Handler struct{ service *Service }

func NewHandler(service *Service) *Handler { return &Handler{service: service} }

func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if !util.BindJSON(c, &req) {
		return
	}
	response, err := h.service.Login(c.Request.Context(), req)
	if err != nil {
		util.Fail(c, err)
		return
	}
	util.OK(c, response)
}

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}
	return string(hash), nil
}

func SeedUsers(db *gorm.DB) error {
	seeds := []struct{ Username, DisplayName, Role, Password string }{
		{"planner", "Training Plan Author", RolePlanner, "planner123"},
		{"supervisor", "Dive Training Supervisor", RoleSupervisor, "supervisor123"},
		{"admin", "Control Administrator", RoleAdmin, "admin123"},
	}
	for _, seed := range seeds {
		var count int64
		if err := db.Model(&User{}).Where("username = ?", seed.Username).Count(&count).Error; err != nil {
			return fmt.Errorf("count seed user %s: %w", seed.Username, err)
		}
		if count > 0 {
			continue
		}
		hash, err := HashPassword(seed.Password)
		if err != nil {
			return err
		}
		user := User{Username: seed.Username, DisplayName: seed.DisplayName, Role: seed.Role, PasswordHash: hash, Active: true}
		if err := db.Create(&user).Error; err != nil {
			return fmt.Errorf("create seed user %s: %w", seed.Username, err)
		}
	}
	return nil
}
