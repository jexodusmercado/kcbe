package api

import (
	"encoding/json"
	"errors"
	"kabancount/internal/store"
	"kabancount/internal/utils"
	"log"
	"net/http"
)

type registerRequest struct {
	CompanyName string `json:"company_name"`
	Username    string `json:"username"`
	Email       string `json:"email"`
	Password    string `json:"password"`
	Bio         string `json:"bio,omitempty"`
	Role        string `json:"role,omitempty"`
}

type AuthHandler struct {
	organizationStore store.OrganizationStore
	userStore         store.UserStore
	logger            *log.Logger
}

func NewAuthHandler(organizationStore store.OrganizationStore, userStore store.UserStore, logger *log.Logger) *AuthHandler {
	return &AuthHandler{
		organizationStore: organizationStore,
		userStore:         userStore,
		logger:            logger,
	}
}

func (ah *AuthHandler) HandleRegister(w http.ResponseWriter, r *http.Request) {

	var req registerRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		ah.logger.Printf("Error decoding request body: %v", err)
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{"error": "Invalid request payload"})
		return
	}

	err = ah.validateRegisterRequest(&req)
	if err != nil {
		ah.logger.Printf("Validation error: %v", err)
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{"error": err.Error()})
		return
	}

	createdOrg, err := ah.organizationStore.CreateOrganization(&store.Organization{
		Name: req.CompanyName,
	})
	if err != nil {
		ah.logger.Printf("Error creating organization: %v", err)
		utils.WriteJSON(w, http.StatusInternalServerError, utils.Envelope{"error": "Failed to create organization"})
		return
	}

	newUser := &store.User{
		OrganizationID: createdOrg.ID,
		Username:       req.Username,
		Email:          req.Email,
		Bio:            req.Bio,
		Role:           "admin",
	}

	err = newUser.PasswordHash.Set(req.Password)
	if err != nil {
		ah.logger.Printf("Error setting user password: %v", err)
		utils.WriteJSON(w, http.StatusInternalServerError, utils.Envelope{"error": "Failed to set user password"})
		return
	}

	createdUser, err := ah.userStore.CreateUser(newUser)
	if err != nil {
		ah.logger.Printf("Error creating user: %v", err)
		utils.WriteJSON(w, http.StatusInternalServerError, utils.Envelope{"error": "Failed to create user"})
		return
	}

	utils.WriteJSON(w, http.StatusCreated, utils.Envelope{
		"organization": createdOrg,
		"user":         createdUser,
	})

}

func (ah *AuthHandler) validateRegisterRequest(req *registerRequest) error {
	if req.CompanyName == "" || req.Username == "" || req.Email == "" || req.Password == "" {
		return errors.New("all fields are required")
	}

	if len(req.CompanyName) < 3 {
		return errors.New("company name must be at least 3 characters long")
	}

	if len(req.Username) < 3 {
		return errors.New("username must be at least 3 characters long")
	}

	if !utils.IsValidEmail(req.Email) {
		return errors.New("invalid email format")
	}

	if !utils.IsPasswordStrong(req.Password) {
		return errors.New("password must be at least 8 characters long and contain a mix of letters, numbers, and symbols")
	}

	return nil

}
