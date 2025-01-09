package controllers

import (
	"fmt"
	"net/http"
	"onichan/database"
	"onichan/model"
	"onichan/services"
	"onichan/utils"
	"os"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
)

type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

type ChangeEmailRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required"`
}

type ResetForgottenPasswordRequest struct {
	Token string `json:"token" binding:"required"`
}

type ChangeAvatarRequest struct {
	AvatarURL string `json:"avatar_url" binding:"required"`
}

var EMAIL_EXPIRATION, _ = strconv.Atoi(os.Getenv("EMAIL_EXPIRATION"))

// Register godoc
// @Summary      Register a new user
// @Description  Creates a new user with the provided username, email, and password.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        registerRequest  body      RegisterRequest  true  "Register user"
// @Success      200  {object}    map[string]interface{}  "{"message": "User created successfully"}"
// @Failure      400  {object}    map[string]interface{}  "{"error": "Password must contain at least 8 characters"}"
// @Failure      409  {object}    map[string]interface{}  "{"error": "Email already in use"}"
// @Failure      500  {object}    map[string]interface{}  "{"error": "Could not hash password"}"
// @Router       /auth/register [post]
func Register(c *gin.Context) {
	var payload RegisterRequest

	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(payload.Password) < 8 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Password must contain atleast 8 characters"})
		return
	}

	var emailUser model.User
	if result := database.Database.First(&emailUser, "email = ?", payload.Email); result.Error == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Email already in use"})
		return
	}

	salt := utils.GetToken(32)
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(salt+payload.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not hash password"})
		return
	}

	randomAvatar := utils.GetRandomAvatar()

	user := model.User{
		Username:     payload.Username,
		Email:        payload.Email,
		PasswordHash: string(hashedPassword),
		Salt:         salt,
		AvatarURL:    &randomAvatar,
	}

	if result := database.Database.Create(&user); result.Error != nil {
		if strings.Contains(result.Error.Error(), "duplicate key") {
			c.JSON(http.StatusConflict, gin.H{"error": "Username already exists"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User created successfully"})
}

// Login godoc
// @Summary      Login a user
// @Description  Authenticates a user with username and password and returns a JWT token.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        loginRequest  body      LoginRequest  true  "Login user"
// @Success      200  {object}  map[string]interface{}  "{"token": "your.jwt.token"}"
// @Failure      400  {object}  map[string]interface{}  "{"error": "Invalid username or password"}"
// @Failure      500  {object}  map[string]interface{}  "{"error": "Failed to generate token"}"
// @Router       /auth/login [post]
func Login(c *gin.Context) {
	var payload LoginRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user model.User
	if err := database.Database.First(&user, "username = ?", payload.Username).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid username or password"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(user.Salt+payload.Password)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid username or password"})
		return
	}

	token, err := utils.GenerateJWT(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

// ChangePassword godoc
// @Summary      Change the current user's password
// @Description  Allows a logged-in user to change their current password by providing the old password and a new password.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        changePasswordRequest  body      ChangePasswordRequest  true  "Change Password"
// @Success      200  {object}  map[string]interface{}  "{"message": "Password updated successfully"}"
// @Failure      400  {object}  map[string]interface{}  "{"error": "Invalid old password"}"
// @Failure      404  {object}  map[string]interface{}  "{"error": "User not found"}"
// @Failure      500  {object}  map[string]interface{}  "{"error": "Could not hash password"}"
// @Security     ApiKeyAuth
// @Router       /auth/change-password [patch]
func ChangePassword(c *gin.Context) {
	var payload ChangePasswordRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := uint(c.MustGet("user_id").(float64))

	var user model.User
	if err := database.Database.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(user.Salt+payload.OldPassword)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid old password"})
		return
	}

	salt := utils.GetToken(32)
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(salt+payload.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not hash password"})
		return
	}

	if err := database.Database.Model(&user).Updates(map[string]interface{}{
		"password_hash": string(hashedPassword),
		"salt":          salt,
	}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update password"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password updated successfully"})
}

// ChangeEmail godoc
// @Summary      Change the current user's email
// @Description  Allows a logged-in user to change their email by providing current password and the new email.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        changeEmailRequest  body      ChangeEmailRequest  true  "Change Email"
// @Success      200  {object}  map[string]interface{}  "{"message": "Email updated successfully"}"
// @Failure      400  {object}  map[string]interface{}  "{"error": "Invalid password"}"
// @Failure      404  {object}  map[string]interface{}  "{"error": "User not found"}"
// @Failure      409  {object}  map[string]interface{}  "{"error": "Email already in use"}"
// @Failure      500  {object}  map[string]interface{}  "{"error": "Failed to update email"}"
// @Security     ApiKeyAuth
// @Router       /auth/change-email [patch]
func ChangeEmail(c *gin.Context) {
	var payload ChangeEmailRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := uint(c.MustGet("user_id").(float64))

	var user model.User
	if err := database.Database.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(user.Salt+payload.Password)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid password"})
		return
	}

	// Check if new email is already in use
	var existingUser model.User
	if result := database.Database.First(&existingUser, "email = ?", payload.Email); result.Error == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Email already in use"})
		return
	}

	if err := database.Database.Model(&user).Update("email", payload.Email).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update email"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Email updated successfully"})
}

// ChangeAvatar godoc
// @Summary      Change the current user's avatar
// @Description  Allows a logged-in user to change their avatar URL.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        changeAvatarRequest  body      ChangeAvatarRequest  true  "Change Avatar URL"
// @Success      200  {object}  map[string]interface{}  "{"message": "Avatar updated successfully"}"
// @Failure      400  {object}  map[string]interface{}  "{"error": "Bad request"}"
// @Failure      404  {object}  map[string]interface{}  "{"error": "User not found"}"
// @Failure      500  {object}  map[string]interface{}  "{"error": "Failed to update avatar"}"
// @Security     ApiKeyAuth
// @Router       /auth/change-avatar [patch]
func ChangeAvatar(c *gin.Context) {
	var payload ChangeAvatarRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := uint(c.MustGet("user_id").(float64))

	var user model.User
	if err := database.Database.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	if err := database.Database.Model(&user).Update("avatar_url", payload.AvatarURL).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update avatar"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Avatar updated successfully"})
}

// ForgotPassword godoc
// @Summary      Initiate password reset
// @Description  Generates a token and sends a password-reset email to the user
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        ForgotPassword body ForgotPasswordRequest true "Forgot Password Payload"
// @Success      200 {object} gin.H{"message": "Email sent"}
// @Failure      400 {object} gin.H{"error": "Bad request"}
// @Failure      404 {object} gin.H{"error": "User not found"}
// @Failure      500 {object} gin.H{"error": "Internal server error"}
// @Router       /auth/forgot-password [post]
func ForgotPassword(c *gin.Context) {
	var payload ForgotPasswordRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user model.User
	if err := database.Database.First(&user, "email = ?", payload.Email).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	token, err := utils.GenerateJWTForgotPassword(user.ID, user.PasswordHash)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	err = services.SendEmail(user.Email, "Change Password", "Click the link to change your password: "+os.Getenv("FRONTEND_URL")+"/reset-password?token="+token)

	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send email"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Email sent"})
}

// ResetForgottenPassword godoc
// @Summary      Reset a forgotten password
// @Description  Validates the token, generates a new password, updates the password in the database, and returns the new password
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        ResetForgottenPassword body ResetForgottenPasswordRequest true "Reset Forgotten Password Payload"
// @Success      200 {object} gin.H{"password": "newly-generated-password"}
// @Failure      400 {object} gin.H{"error": "Invalid token or bad request"}
// @Failure      404 {object} gin.H{"error": "User not found"}
// @Failure      500 {object} gin.H{"error": "Failed to update password"}
// @Router       /auth/reset-password [post]
func ResetForgottenPassword(c *gin.Context) {
	var payload ResetForgottenPasswordRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, err := utils.ValidateJWT(payload.Token)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid token"})
		return
	}

	claims := token.Claims.(jwt.MapClaims)
	userID := uint(claims["user_id"].(float64))
	hash := claims["hash"].(string)

	fmt.Println(userID, hash)

	var user model.User
	if err := database.Database.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	if hash != user.PasswordHash {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid token"})
		return
	}

	password := utils.GetRandomAlphaString(8)
	salt := utils.GetToken(32)
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(salt+password), bcrypt.DefaultCost)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not hash password"})
		return
	}

	if err := database.Database.Model(&user).Updates(map[string]interface{}{
		"password_hash": string(hashedPassword),
		"salt":          salt,
	}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update password"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"password": password})
}
