package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/panhao/url-shortener/internal/model"
	"github.com/panhao/url-shortener/internal/repository"
	"github.com/panhao/url-shortener/internal/util"
)

type AuthService struct {
	userRepo      *repository.UserRepo
	jwtSecret     []byte
	jwtExpireHours int
}

func NewAuthService(userRepo *repository.UserRepo, jwtSecret string, jwtExpireHours int) *AuthService {
	return &AuthService{
		userRepo:       userRepo,
		jwtSecret:      []byte(jwtSecret),
		jwtExpireHours: jwtExpireHours,
	}
}

func (s *AuthService) Register(ctx context.Context, req model.RegisterReq) (*model.AuthResp, error) {
	username := strings.TrimSpace(req.Username)
	if len(username) < 2 || len(username) > 30 {
		return nil, fmt.Errorf("username must be 2-30 characters")
	}
	password := req.Password
	if len(password) < 6 {
		return nil, fmt.Errorf("password must be at least 6 characters")
	}

	existing, err := s.userRepo.GetByUsername(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("check username: %w", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("username already taken")
	}

	hash, err := util.HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	user := &model.User{
		Username:     username,
		Email:        req.Email,
		PasswordHash: hash,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	token, err := s.generateJWT(user.ID, user.Username)
	if err != nil {
		return nil, fmt.Errorf("generate token: %w", err)
	}

	return &model.AuthResp{
		Token:    token,
		Username: user.Username,
		UserID:   user.ID,
	}, nil
}

func (s *AuthService) Login(ctx context.Context, req model.LoginReq) (*model.AuthResp, error) {
	username := strings.TrimSpace(req.Username)
	if username == "" {
		return nil, fmt.Errorf("username is required")
	}

	user, err := s.userRepo.GetByUsername(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("invalid username or password")
	}

	if !util.CheckPassword(req.Password, user.PasswordHash) {
		return nil, fmt.Errorf("invalid username or password")
	}

	token, err := s.generateJWT(user.ID, user.Username)
	if err != nil {
		return nil, fmt.Errorf("generate token: %w", err)
	}

	return &model.AuthResp{
		Token:    token,
		Username: user.Username,
		UserID:   user.ID,
	}, nil
}

func (s *AuthService) ValidateToken(tokenString string) (int64, string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return s.jwtSecret, nil
	})

	if err != nil {
		return 0, "", err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return 0, "", fmt.Errorf("invalid token")
	}

	userID := int64(claims["user_id"].(float64))
	username := claims["username"].(string)
	return userID, username, nil
}

func (s *AuthService) generateJWT(userID int64, username string) (string, error) {
	claims := jwt.MapClaims{
		"user_id":  userID,
		"username": username,
		"exp":      time.Now().Add(time.Duration(s.jwtExpireHours) * time.Hour).Unix(),
		"iat":      time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}
