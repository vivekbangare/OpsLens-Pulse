package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTManager struct {
	secret   []byte
	issuer   string
	audience string
}

func NewJWTManager(secret, issuer, audience string) *JWTManager {
	return &JWTManager{
		secret:   []byte(secret),
		issuer:   issuer,
		audience: audience,
	}
}

type Claims struct {
	UserID       string   `json:"user_id"`
	Username     string   `json:"username"`
	TenantID     string   `json:"tenant_id"`
	Permissions  []string `json:"permissions"`
	IsSuperAdmin bool     `json:"is_super_admin"`
	jwt.RegisteredClaims
}

func (j *JWTManager) Generate(
	userID string,
	username string,
	tenantID string,
	permissions []string,
	isSuperAdmin bool,
) (string, error) {

	claims := Claims{
		UserID:       userID,
		Username:     username,
		TenantID:     tenantID,
		Permissions:  permissions,
		IsSuperAdmin: isSuperAdmin,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    j.issuer,
			Audience:  []string{j.audience},
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.secret)
}

func (j *JWTManager) Validate(tokenStr string) (*Claims, error) {

	token, err := jwt.ParseWithClaims(
		tokenStr,
		&Claims{},
		func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return j.secret, nil
		},
		jwt.WithIssuer(j.issuer),
		jwt.WithAudience(j.audience),
		jwt.WithLeeway(2*time.Minute), // clock skew tolerance
	)

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}
