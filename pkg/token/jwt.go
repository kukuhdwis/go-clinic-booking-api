package token

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Claims struct {
	UserID uuid.UUID `json:"user_id"`
	Email  string    `json:"email"`
	Role   string    `json:"role"`
	jwt.RegisteredClaims
}

type JWTManager struct {
	secretKey     []byte
	tokenDuration time.Duration
}

func NewJWTManager(secretKey string, durationHours int) *JWTManager {
	return &JWTManager{
		secretKey:     []byte(secretKey),
		tokenDuration: time.Duration(durationHours) * time.Hour,
	}
}

func (m *JWTManager) GenerateToken(userID uuid.UUID, email, role string) (string, time.Time, error) {
	expiresAt := time.Now().Add(m.tokenDuration)
	claims := Claims{
		UserID: userID,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := t.SignedString(m.secretKey)
	if err != nil {
		return "", time.Time{}, err
	}

	return tokenString, expiresAt, nil
}

func (m *JWTManager) ValidateToken(tokenString string) (*Claims, error) {
	parsedToken, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return m.secretKey, nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := parsedToken.Claims.(*Claims)
	if !ok || !parsedToken.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}
