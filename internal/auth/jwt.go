package auth

import (
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"time"
)

var ErrTokenExpired = errors.New("token expired")

type JWTClaims struct {
	UserID int   `json:"user_id"`
	Exp    int64 `json:"exp"`
	jwt.MapClaims
}

func GenerateTokens(secret string, userID int) (string, string, error) {
	if secret == "" {
		return "", "", fmt.Errorf("secret is empty")
	}

	// Create the access token
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["user_id"] = userID
	claims["exp"] = time.Now().Add(time.Minute * 15).Unix() // Access token expires in 15 minutes
	accessToken, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", "", err
	}

	// Create the refresh token
	refreshToken := jwt.New(jwt.SigningMethodHS256)
	refreshClaims := refreshToken.Claims.(jwt.MapClaims)
	refreshClaims["user_id"] = userID
	refreshClaims["exp"] = time.Now().Add(time.Hour * 24 * 7).Unix() // Refresh token expires in 7 days
	refreshTokenString, err := refreshToken.SignedString([]byte(secret))
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshTokenString, nil
}

func ValidateJWTToken(secret string, tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		expirationTime := time.Unix(claims.Exp, 0)
		if expirationTime.Before(time.Now()) {
			return nil, ErrTokenExpired
		}

		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}
