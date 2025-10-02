package api

import (
	"encoding/json"
	"errors"
	"kabancount/internal/middleware"
	"kabancount/internal/store"
	"kabancount/internal/utils"
	"log"
	"net/http"
)

type LocationHandler struct {
	locationStore store.LocationStore
	logger        *log.Logger
}

func NewLocationHandler(locationStore store.LocationStore, logger *log.Logger) *LocationHandler {
	return &LocationHandler{
		locationStore: locationStore,
		logger:        logger,
	}
}

func (lh *LocationHandler) HandleCreateLocation(w http.ResponseWriter, r *http.Request) {

	user := middleware.GetUser(r)
	if user.IsAnonymous() {
		utils.WriteJSON(w, http.StatusUnauthorized, utils.Envelope{"error": "Unauthorized"})
		return
	}

	var req store.Location

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		lh.logger.Printf("Error decoding request body: %v", err)
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if err := lh.validateCreateLocationRequest(&req); err != nil {
		lh.logger.Printf("Validation error: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	req.OrganizationID = user.OrganizationID

	createdLocation, err := lh.locationStore.CreateLocation(&req)
	if err != nil {
		lh.logger.Printf("Error creating location: %v", err)
		http.Error(w, "Failed to create location", http.StatusInternalServerError)
		return
	}

	utils.WriteJSON(w, http.StatusCreated, utils.Envelope{"data": createdLocation})

}

func (lh *LocationHandler) HandleGetLocationsByOrganization(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r)
	if user.IsAnonymous() {
		utils.WriteJSON(w, http.StatusUnauthorized, utils.Envelope{"error": "Unauthorized"})
		return
	}

	locations, err := lh.locationStore.GetLocationsByOrganization(user.OrganizationID)
	if err != nil {
		lh.logger.Printf("Error fetching locations: %v", err)
		http.Error(w, "Failed to fetch locations", http.StatusInternalServerError)
		return
	}

	utils.WriteJSON(w, http.StatusOK, utils.Envelope{"data": locations})

}

func (lh *LocationHandler) validateCreateLocationRequest(location *store.Location) error {
	if location.Name == "" {
		return errors.New("name is required")
	}

	return nil
}
