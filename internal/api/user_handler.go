package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"kabancount/internal/store"
	"kabancount/internal/utils"
	"log"
	"net/http"
)

type registerUserRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Bio      string `json:"bio,omitempty"`
	Role     string `json:"role,omitempty"`
}

type UserHandler struct {
	userStore store.UserStore
	logger    *log.Logger
}

func NewUserHandler(userStore store.UserStore, logger *log.Logger) *UserHandler {
	return &UserHandler{
		userStore: userStore,
		logger:    logger,
	}
}

func (u *UserHandler) HandleGetUserByID(w http.ResponseWriter, r *http.Request) {
	userID, err := utils.ReadIDParam(r)
	if err != nil {
		u.logger.Printf("Error reading ID parameter: %v", err)
		http.Error(w, "Invalid ID parameter", http.StatusBadRequest)
		return
	}

	fmt.Fprintf(w, "User ID: %s\n", userID.String())
}

func (u *UserHandler) HandleCreateUser(w http.ResponseWriter, r *http.Request) {
	var req registerUserRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		u.logger.Printf("Error decoding request body: %v", err)
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	err = u.validateRegisterRequest(&req)
	if err != nil {
		u.logger.Printf("Validation error: %v", err)
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{"error": "invalid request payload"})
		return
	}

	newUser := &store.User{
		Username: req.Username,
		Email:    req.Email,
	}

	if req.Bio != "" {
		newUser.Bio = req.Bio
	}

	if req.Role != "" {
		newUser.Role = req.Role
	}

	err = newUser.PasswordHash.Set(req.Password)
	if err != nil {
		u.logger.Printf("Error setting password: %v", err)
		utils.WriteJSON(w, http.StatusInternalServerError, utils.Envelope{"error": "failed to create user"})
		return
	}

	createdUser, err := u.userStore.CreateUser(newUser)
	if err != nil {
		u.logger.Printf("Error creating user: %v", err)
		utils.WriteJSON(w, http.StatusInternalServerError, utils.Envelope{"error": "failed to create user"})
		return
	}

	utils.WriteJSON(w, http.StatusCreated, utils.Envelope{"user": createdUser})

}

func (u *UserHandler) validateRegisterRequest(req *registerUserRequest) error {
	if req.Username == "" {
		return errors.New("username is required")
	}

	if len(req.Username) < 3 || len(req.Username) > 30 {
		return errors.New("username must be between 3 and 30 characters")
	}

	if req.Email == "" {
		return errors.New("email is required")
	}

	if !utils.IsValidEmail(req.Email) {
		return errors.New("invalid email format")
	}

	if req.Password == "" {
		return errors.New("password is required")
	}

	if !utils.IsPasswordStrong(req.Password) {
		return errors.New("password ensures minimum 8 characters, at least one uppercase letter, one lowercase letter, one number and one special character")
	}

	return nil
}
