package api

import (
	"encoding/json"
	"kabancount/internal/store"
	"kabancount/internal/tokens"
	"kabancount/internal/utils"
	"log"
	"net/http"
	"time"
)

type TokenHandler struct {
	tokenStore store.TokenStore
	userStore  store.UserStore
	logger     *log.Logger
}

type createTokenRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func NewTokenHandler(tokenStore store.TokenStore, userStore store.UserStore, logger *log.Logger) *TokenHandler {
	return &TokenHandler{
		tokenStore: tokenStore,
		userStore:  userStore,
		logger:     logger,
	}
}

func (h *TokenHandler) HandleCreateToken(w http.ResponseWriter, r *http.Request) {
	var req createTokenRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		h.logger.Printf("Error decoding request body: %v", err)
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{"error": "Invalid request payload"})
		return
	}

	userData, err := h.userStore.GetUserByUsername(req.Username)
	if err != nil || userData == nil {
		h.logger.Printf("Error retrieving user: %v", err)
		utils.WriteJSON(w, http.StatusUnauthorized, utils.Envelope{"error": "Invalid username or password"})
		return
	}

	passwordDoMatch, err := userData.PasswordHash.Matches(req.Password)
	if err != nil || !passwordDoMatch {
		h.logger.Printf("Password mismatch: %v", err)
		utils.WriteJSON(w, http.StatusUnauthorized, utils.Envelope{"error": "Invalid username or password"})
		return
	}

	token, err := h.tokenStore.CreateNewToken(userData.ID, userData.OrganizationID, 24*time.Hour, tokens.ScopeAuth)
	if err != nil {
		h.logger.Printf("Error creating token: %v", err)
		utils.WriteJSON(w, http.StatusInternalServerError, utils.Envelope{"error": "Failed to create token"})
		return
	}

	utils.WriteJSON(w, http.StatusOK, utils.Envelope{"token": token})
}
