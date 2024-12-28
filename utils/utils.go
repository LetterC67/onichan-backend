package utils

import (
	"encoding/base32"
	"log"
	"onichan/database"
	"onichan/model"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/joho/godotenv"
	"golang.org/x/exp/rand"
)

var jwtSecret []byte
var jwtTTL int

func LoadEnv() {
	err := godotenv.Load(".env")

	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
}

func LoadJWT() {
	jwtSecret = []byte(os.Getenv("JWT_SECRET_KEY"))
	jwtTTL, _ = strconv.Atoi(os.Getenv("TOKEN_TTL"))
}

func GenerateJWT(userID uint) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Duration(jwtTTL) * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString(jwtSecret)
}

func ValidateJWT(tokenString string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return jwtSecret, nil
	})

	return token, err
}

func GetRandomAvatar() string {
	var avatar model.Avatar

	rand.Seed(uint64(time.Now().UnixNano()))

	if err := database.Database.Order("RANDOM()").First(&avatar).Error; err != nil {
		return "https://www.svgrepo.com/svg/411748/solve"
	}

	return avatar.AvatarURL
}

func GetPostPage(post model.Post) int {
	var count int64
	pageSize, _ := strconv.Atoi(os.Getenv("PAGE_SIZE"))
	database.Database.Model(&model.Post{}).Where("created_at < ? AND parent_post_id = ?", post.CreatedAt, post.ParentPostID).Count(&count)
	return int((int(count) + pageSize + 1) / pageSize)
}

func GetToken(length int) string {
	randomBytes := make([]byte, 32)
	_, err := rand.Read(randomBytes)
	if err != nil {
		panic(err)
	}
	return base32.StdEncoding.EncodeToString(randomBytes)[:length]
}
