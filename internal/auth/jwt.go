package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID    int64  `json:"user_id"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	TokenType string `json:"token_type"`
	jwt.RegisteredClaims
}

type JWTManager struct {
	secret []byte
	issuer string
}

func NewJWTManager(secret, issuer string) (*JWTManager, error) {
	if secret == "" {
		return nil, errors.New("jwt secret must not be empty")
	}
	if issuer == "" {
		return nil, errors.New("jwt issuer must not be empty")
	}
	return &JWTManager{
		secret: []byte(secret),
		issuer: issuer,
	}, nil
}

func (m *JWTManager) GenerateToken(userID int64, email, role string, ttl time.Duration) (string, error) {
	return m.generateToken(userID, email, role, "access", "", ttl)
}

func (m *JWTManager) GenerateRefreshToken(userID int64, email, role string, ttl time.Duration) (token string, tokenID string, err error) {
	tokenID, err = randomTokenID()
	if err != nil {
		return "", "", err
	}
	token, err = m.generateToken(userID, email, role, "refresh", tokenID, ttl)
	if err != nil {
		return "", "", err
	}
	return token, tokenID, nil
}

func (m *JWTManager) generateToken(userID int64, email, role, tokenType, tokenID string, ttl time.Duration) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID:    userID,
		Email:     email,
		Role:      role,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    m.issuer,
			Subject:   fmt.Sprintf("%d", userID),
			ID:        tokenID,
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secret)
}

func (m *JWTManager) ParseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return m.secret, nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token claims")
	}
	return claims, nil
}

func randomTokenID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
