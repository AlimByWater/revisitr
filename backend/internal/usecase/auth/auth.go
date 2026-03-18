package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log/slog"
	"time"

	"revisitr/internal/entity"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

const (
	bcryptCost      = 10
	refreshTokenLen = 32
	refreshTokenTTL = 30 * 24 * time.Hour
)

type config interface {
	GetJWTSecret() string
	GetTokenTTL() time.Duration
}

type usersRepo interface {
	Create(ctx context.Context, user *entity.User) error
	GetByID(ctx context.Context, id int) (*entity.User, error)
	GetByEmail(ctx context.Context, email string) (*entity.User, error)
	CreateOrganization(ctx context.Context, org *entity.Organization) error
	UpdateOrganizationOwner(ctx context.Context, orgID, ownerID int) error
}

type sessionsRepo interface {
	StoreRefreshToken(ctx context.Context, userID int, tokenHash string, ttl time.Duration) error
	GetUserIDByToken(ctx context.Context, tokenHash string) (int, error)
	DeleteRefreshToken(ctx context.Context, tokenHash string) error
	DeleteUserSessions(ctx context.Context, userID int) error
}

type Usecase struct {
	cfg      config
	users    usersRepo
	sessions sessionsRepo
	logger   *slog.Logger
}

func New(cfg config, users usersRepo, sessions sessionsRepo) *Usecase {
	return &Usecase{
		cfg:      cfg,
		users:    users,
		sessions: sessions,
	}
}

func (uc *Usecase) Init(_ context.Context, logger *slog.Logger) error {
	uc.logger = logger
	return nil
}

func (uc *Usecase) Register(ctx context.Context, req *entity.RegisterRequest) (*entity.AuthResponse, error) {
	existing, _ := uc.users.GetByEmail(ctx, req.Email)
	if existing != nil {
		return nil, ErrUserExists
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcryptCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	org := &entity.Organization{
		Name: req.Organization,
	}
	if err := uc.users.CreateOrganization(ctx, org); err != nil {
		return nil, fmt.Errorf("create organization: %w", err)
	}

	user := &entity.User{
		Email:        req.Email,
		Phone:        req.Phone,
		Name:         req.Name,
		PasswordHash: string(hash),
		Role:         "owner",
		OrgID:        org.ID,
	}
	if err := uc.users.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	if err := uc.users.UpdateOrganizationOwner(ctx, org.ID, user.ID); err != nil {
		return nil, fmt.Errorf("update organization owner: %w", err)
	}

	tokens, err := uc.generateTokenPair(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("generate tokens: %w", err)
	}

	return &entity.AuthResponse{
		User:   *user,
		Tokens: *tokens,
	}, nil
}

func (uc *Usecase) Login(ctx context.Context, req *entity.LoginRequest) (*entity.AuthResponse, error) {
	user, err := uc.users.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	tokens, err := uc.generateTokenPair(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("generate tokens: %w", err)
	}

	return &entity.AuthResponse{
		User:   *user,
		Tokens: *tokens,
	}, nil
}

func (uc *Usecase) Refresh(ctx context.Context, refreshToken string) (*entity.TokenPair, error) {
	tokenHash := hashToken(refreshToken)

	userID, err := uc.sessions.GetUserIDByToken(ctx, tokenHash)
	if err != nil {
		return nil, ErrTokenExpired
	}

	if err := uc.sessions.DeleteRefreshToken(ctx, tokenHash); err != nil {
		uc.logger.Error("failed to delete old refresh token", "error", err)
	}

	user, err := uc.users.GetByID(ctx, userID)
	if err != nil {
		return nil, ErrTokenExpired
	}

	tokens, err := uc.generateTokenPair(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("generate tokens: %w", err)
	}

	return tokens, nil
}

func (uc *Usecase) Logout(ctx context.Context, refreshToken string) error {
	tokenHash := hashToken(refreshToken)

	if err := uc.sessions.DeleteRefreshToken(ctx, tokenHash); err != nil {
		return fmt.Errorf("delete refresh token: %w", err)
	}

	return nil
}

func (uc *Usecase) generateTokenPair(ctx context.Context, user *entity.User) (*entity.TokenPair, error) {
	ttl := uc.cfg.GetTokenTTL()
	now := time.Now()

	claims := jwt.MapClaims{
		"user_id": user.ID,
		"org_id":  user.OrgID,
		"role":    user.Role,
		"exp":     now.Add(ttl).Unix(),
		"iat":     now.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	accessToken, err := token.SignedString([]byte(uc.cfg.GetJWTSecret()))
	if err != nil {
		return nil, fmt.Errorf("sign access token: %w", err)
	}

	refreshBytes := make([]byte, refreshTokenLen)
	if _, err := rand.Read(refreshBytes); err != nil {
		return nil, fmt.Errorf("generate refresh token: %w", err)
	}
	rawRefreshToken := hex.EncodeToString(refreshBytes)

	tokenHash := hashToken(rawRefreshToken)
	if err := uc.sessions.StoreRefreshToken(ctx, user.ID, tokenHash, refreshTokenTTL); err != nil {
		return nil, fmt.Errorf("store refresh token: %w", err)
	}

	return &entity.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: rawRefreshToken,
		ExpiresIn:    int64(ttl.Seconds()),
	}, nil
}

func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}
