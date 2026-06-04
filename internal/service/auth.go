package service

import (
	"errors"
	"os"
	"time"

	"spendwise-ms/internal/model"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var jwtSecret = []byte(getJWTSecret())

func getJWTSecret() string {
	if s := os.Getenv("JWT_SECRET"); s != "" {
		return s
	}
	return "spendwise-secret-key"
}

type Claims struct {
	UserID string `json:"userId"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	return string(bytes), err
}

func CheckPassword(password, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

func GenerateTokens(userID, email string) (string, string, error) {
	accessClaims := Claims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ID:        uuid.New().String(),
		},
	}
	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims).SignedString(jwtSecret)
	if err != nil {
		return "", "", err
	}

	refreshClaims := Claims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ID:        uuid.New().String(),
		},
	}
	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString(jwtSecret)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}

func Register(db *gorm.DB, name, email, password string) (*model.User, error) {
	var existing model.User
	if err := db.Where("email = ?", email).First(&existing).Error; err == nil {
		return nil, errors.New("email already registered")
	}

	hash, err := HashPassword(password)
	if err != nil {
		return nil, err
	}

	user := model.User{
		ID:       uuid.New().String(),
		Name:     name,
		Email:    email,
		Password: hash,
	}
	if err := db.Create(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func Login(db *gorm.DB, email, password string) (*model.User, error) {
	var user model.User
	if err := db.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, errors.New("invalid email or password")
	}
	if !CheckPassword(password, user.Password) {
		return nil, errors.New("invalid email or password")
	}
	return &user, nil
}
