package service

import (
	"errors"
	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	storage "github.com/mahadeva604/audio-storage"
	"github.com/mahadeva604/audio-storage/pkg/repository"
	"time"
)

type tokenClaims struct {
	jwt.StandardClaims
	UserId int `json:"user_id"`
}

type AuthService struct {
	repo            repository.Authorization
	secretKey       []byte
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
}

func NewAuthService(repo repository.Authorization, secretKey []byte, accessTokenTTL, refreshTokenTTL time.Duration) *AuthService {
	return &AuthService{repo: repo, secretKey: secretKey, accessTokenTTL: accessTokenTTL, refreshTokenTTL: refreshTokenTTL}
}

func (s *AuthService) CreateUser(user storage.User) (int, error) {
	return s.repo.CreateUser(user)
}

func (s *AuthService) GenerateAccessToken(username, password string) (int, string, error) {
	user, err := s.repo.GetUser(username, password)
	if err != nil {
		return 0, "", err
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &tokenClaims{
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(s.accessTokenTTL).Unix(),
			IssuedAt:  time.Now().Unix(),
		},
		user.Id,
	})

	signedToken, err := token.SignedString(s.secretKey)

	return user.Id, signedToken, err
}

func (s *AuthService) UpdateAccessToken(userId int) (string, error) {

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &tokenClaims{
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(s.accessTokenTTL).Unix(),
			IssuedAt:  time.Now().Unix(),
		},
		userId,
	})

	return token.SignedString(s.secretKey)
}

func (s *AuthService) GenerateRefreshToken(userId int) (string, error) {

	refreshToken := uuid.New().String()

	err := s.repo.SetRefreshToken(userId, refreshToken, s.refreshTokenTTL)
	if err != nil {
		return "", err
	}
	return refreshToken, nil
}

func (s *AuthService) UpdateRefreshToken(oldRefreshToken string) (int, string, error) {

	newRefreshToken := uuid.New().String()

	userId, err := s.repo.UpdateRefreshToken(oldRefreshToken, newRefreshToken, s.refreshTokenTTL)
	if err != nil {
		return 0, "", err
	}
	return userId, newRefreshToken, nil
}

func (s *AuthService) ParseToken(accessToken string) (int, error) {
	token, err := jwt.ParseWithClaims(accessToken, &tokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return []byte(s.secretKey), nil
	})
	if err != nil {
		return 0, err
	}

	claims, ok := token.Claims.(*tokenClaims)
	if !ok {
		return 0, errors.New("token claims are not of type *tokenClaims")
	}

	return claims.UserId, nil
}
