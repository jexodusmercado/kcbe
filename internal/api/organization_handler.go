package api

import (
	"encoding/json"
	"kabancount/internal/store"
	"kabancount/internal/utils"
	"log"
	"net/http"
)

type OrganizationHandler struct {
	organizationStore store.OrganizationStore
	logger            *log.Logger
}

func NewOrganizationHandler(organizationStore store.OrganizationStore, logger *log.Logger) *OrganizationHandler {
	return &OrganizationHandler{
		organizationStore: organizationStore,
		logger:            logger,
	}
}

func (oh *OrganizationHandler) HandleCreateOrganization(w http.ResponseWriter, r *http.Request) {
	var organization store.Organization

	var paramOrganization struct {
		Name *string `json:"name"`
	}

	err := json.NewDecoder(r.Body).Decode(&paramOrganization)
	if err != nil {
		oh.logger.Printf("Error decoding request body: %v", err)
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{"error": "Invalid request payload"})
		return
	}

	if paramOrganization.Name != nil {
		organization.Name = *paramOrganization.Name
	} else {
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{"error": "Name is required"})
		return
	}

	createdOrganization, err := oh.organizationStore.CreateOrganization(&organization)
	if err != nil {
		oh.logger.Printf("Error creating organization: %v", err)
		utils.WriteJSON(w, http.StatusInternalServerError, utils.Envelope{"error": "Failed to create organization"})
		return
	}

	utils.WriteJSON(w, http.StatusCreated, utils.Envelope{"organization": createdOrganization})
}

func (oh *OrganizationHandler) HandleGetOrganizationByID(w http.ResponseWriter, r *http.Request) {
	orgID, err := utils.ReadIDParam(r)
	if err != nil {
		oh.logger.Printf("Error reading ID parameter: %v", err)
		http.Error(w, "Invalid ID parameter", http.StatusBadRequest)
		return
	}

	orgData, err := oh.organizationStore.GetOrganizationByID(*orgID)
	if err != nil {
		oh.logger.Printf("Error retrieving organization: %v", err)
		utils.WriteJSON(w, http.StatusInternalServerError, utils.Envelope{"error": "Failed to retrieve organization"})
		return
	}

	if orgData == nil {
		http.NotFound(w, r)
		return
	}

	utils.WriteJSON(w, http.StatusOK, utils.Envelope{"organization": orgData})
}

func (oh *OrganizationHandler) HandleUpdateOrganization(w http.ResponseWriter, r *http.Request) {
	orgID, err := utils.ReadIDParam(r)
	if err != nil {
		oh.logger.Printf("Error reading ID parameter: %v", err)
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{"error": "Invalid ID parameter"})
		return
	}

	existingOrg, err := oh.organizationStore.GetOrganizationByID(*orgID)
	if err != nil {
		oh.logger.Printf("Error retrieving organization: %v", err)
		utils.WriteJSON(w, http.StatusInternalServerError, utils.Envelope{"error": "Failed to retrieve organization"})
		return
	}

	if existingOrg == nil {
		http.NotFound(w, r)
		return
	}

	var paramOrganization store.Organization
	err = json.NewDecoder(r.Body).Decode(&paramOrganization)
	if err != nil {
		oh.logger.Printf("Error decoding request body: %v", err)
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{"error": "Invalid request payload"})
		return
	}

	if paramOrganization.Name != "" {
		existingOrg.Name = paramOrganization.Name
	}

	updatedOrg, err := oh.organizationStore.UpdateOrganization(existingOrg)
	if err != nil {
		oh.logger.Printf("Error updating organization: %v", err)
		utils.WriteJSON(w, http.StatusInternalServerError, utils.Envelope{"error": "Failed to update organization"})
		return
	}

	utils.WriteJSON(w, http.StatusOK, utils.Envelope{"organization": updatedOrg})
}

func (oh *OrganizationHandler) HandleDeleteOrganization(w http.ResponseWriter, r *http.Request) {
	orgID, err := utils.ReadIDParam(r)
	if err != nil {
		oh.logger.Printf("Error reading ID parameter: %v", err)
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{"error": "Invalid ID parameter"})
		return
	}

	existingOrg, err := oh.organizationStore.GetOrganizationByID(*orgID)
	if err != nil {
		oh.logger.Printf("Error retrieving organization: %v", err)
		utils.WriteJSON(w, http.StatusInternalServerError, utils.Envelope{"error": "Failed to retrieve organization"})
		return
	}

	if existingOrg == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	err = oh.organizationStore.DeleteOrganization(*orgID)
	if err != nil {
		oh.logger.Printf("Error deleting organization: %v", err)
		utils.WriteJSON(w, http.StatusInternalServerError, utils.Envelope{"error": "Failed to delete organization"})
		return
	}

	utils.WriteJSON(w, http.StatusOK, utils.Envelope{"message": "Organization deleted successfully"})
}
