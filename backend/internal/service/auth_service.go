package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"ragserver/backend/internal/dto"
	"ragserver/backend/internal/model"
	"ragserver/backend/internal/repository"
)

type AuthService struct {
	users        *repository.UserRepository
	jwtSecret    []byte
	expiresHours int
}

type claims struct {
	UserID uint64 `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func NewAuthService(users *repository.UserRepository, jwtSecret string, expiresHours int) *AuthService {
	if expiresHours <= 0 {
		expiresHours = 24
	}
	return &AuthService{users: users, jwtSecret: []byte(jwtSecret), expiresHours: expiresHours}
}

func (s *AuthService) Register(ctx context.Context, req dto.RegisterRequest) (*dto.UserResponse, error) {
	email := strings.ToLower(strings.TrimSpace(req.Email))
	if email == "" {
		return nil, fmt.Errorf("email is required")
	}
	if len(req.Password) < 6 {
		return nil, fmt.Errorf("password must be at least 6 characters")
	}
	role := strings.TrimSpace(req.Role)
	if role != model.RoleTeacher && role != model.RoleStudent {
		return nil, fmt.Errorf("role must be teacher or student")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	user := &model.User{
		Email:        email,
		PasswordHash: string(hash),
		DisplayName:  strings.TrimSpace(req.DisplayName),
		Role:         role,
		Status:       model.StatusActive,
	}
	if user.DisplayName == "" {
		user.DisplayName = email
	}
	if err := s.users.Create(ctx, user); err != nil {
		return nil, err
	}
	resp := userToDTO(*user)
	return &resp, nil
}

func (s *AuthService) Login(ctx context.Context, req dto.LoginRequest) (*dto.LoginResponse, error) {
	user, err := s.users.GetByEmail(ctx, strings.ToLower(strings.TrimSpace(req.Email)))
	if err != nil {
		return nil, fmt.Errorf("invalid email or password")
	}
	if user.Status != model.StatusActive {
		return nil, fmt.Errorf("user is disabled")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, fmt.Errorf("invalid email or password")
	}
	now := time.Now()
	user.LastLoginAt = &now
	_ = s.users.Update(ctx, user)
	token, expiresAt, err := s.issueToken(*user)
	if err != nil {
		return nil, err
	}
	return &dto.LoginResponse{Token: token, ExpiresAt: expiresAt, User: userToDTO(*user)}, nil
}

func (s *AuthService) ParseToken(tokenValue string) (*model.User, error) {
	parsed, err := jwt.ParseWithClaims(tokenValue, &claims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return s.jwtSecret, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := parsed.Claims.(*claims)
	if !ok || !parsed.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	user, err := s.users.Get(context.Background(), claims.UserID)
	if err != nil {
		return nil, err
	}
	if user.Status != model.StatusActive {
		return nil, fmt.Errorf("user is disabled")
	}
	return user, nil
}

func (s *AuthService) issueToken(user model.User) (string, time.Time, error) {
	expiresAt := time.Now().Add(time.Duration(s.expiresHours) * time.Hour)
	claims := claims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   fmt.Sprintf("%d", user.ID),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(s.jwtSecret)
	return signed, expiresAt, err
}

type UserService struct {
	users *repository.UserRepository
}

func NewUserService(users *repository.UserRepository) *UserService {
	return &UserService{users: users}
}

func (s *UserService) ListActive(ctx context.Context) ([]dto.UserResponse, error) {
	users, err := s.users.ListActive(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]dto.UserResponse, 0, len(users))
	for _, user := range users {
		out = append(out, userToDTO(user))
	}
	return out, nil
}

func userToDTO(user model.User) dto.UserResponse {
	return dto.UserResponse{
		ID:          user.ID,
		Email:       user.Email,
		DisplayName: user.DisplayName,
		Role:        user.Role,
		Status:      user.Status,
		LastLoginAt: user.LastLoginAt,
	}
}
