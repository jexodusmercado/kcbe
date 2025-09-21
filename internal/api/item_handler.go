package api

import (
	"encoding/json"
	"errors"
	"kabancount/internal/middleware"
	"kabancount/internal/store"
	"kabancount/internal/utils"
	"log"
	"net/http"

	"github.com/google/uuid"
)

type ItemHandler struct {
	itemStore store.ItemStore
	logger    *log.Logger
}

func NewItemHandler(itemStore store.ItemStore, logger *log.Logger) *ItemHandler {
	return &ItemHandler{
		itemStore: itemStore,
		logger:    logger,
	}
}

func (ih *ItemHandler) HandleCreateItem(w http.ResponseWriter, r *http.Request) {
	var req store.Item

	user := middleware.GetUser(r)
	if user.IsAnonymous() {
		utils.WriteJSON(w, http.StatusUnauthorized, utils.Envelope{"error": "Unauthorized"})
		return
	}

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		ih.logger.Printf("Error decoding request body: %v", err)
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{"error": "Invalid request payload"})
		return
	}

	if err := ih.validateCreateItemRequest(&req); err != nil {
		ih.logger.Printf("Validation error: %v", err)
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{"error": err.Error()})
		return
	}

	req.OrganizationID = user.OrganizationID

	createdItem, err := ih.itemStore.CreateItem(&req)
	if err != nil {
		ih.logger.Printf("Error creating item: %v", err)
		utils.WriteJSON(w, http.StatusInternalServerError, utils.Envelope{"error": "Failed to create item"})
		return
	}

	utils.WriteJSON(w, http.StatusCreated, utils.Envelope{"item": createdItem})
}

func (ih *ItemHandler) HandleGetItemByID(w http.ResponseWriter, r *http.Request) {
	itemID, err := utils.ReadIDParam(r)
	if err != nil {
		ih.logger.Printf("Error reading ID parameter: %v", err)
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{"error": "Invalid ID parameter"})
		return
	}

	item, err := ih.itemStore.GetItemByID(*itemID)
	if err != nil {
		ih.logger.Printf("Error fetching item: %v", err)
		utils.WriteJSON(w, http.StatusNotFound, utils.Envelope{"error": "Item not found"})
		return
	}

	utils.WriteJSON(w, http.StatusOK, utils.Envelope{"item": item})
}

func (ih *ItemHandler) HandleUpdateItem(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r)
	if user.IsAnonymous() {
		utils.WriteJSON(w, http.StatusUnauthorized, utils.Envelope{"error": "Unauthorized"})
		return
	}

	itemID, err := utils.ReadIDParam(r)
	if err != nil {
		ih.logger.Printf("Error reading ID parameter: %v", err)
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{"error": "Invalid ID parameter"})
		return
	}

	existingItem, err := ih.itemStore.GetItemByID(*itemID)
	if err != nil {
		ih.logger.Printf("Error retrieving item: %v", err)
		utils.WriteJSON(w, http.StatusInternalServerError, utils.Envelope{"error": "Failed to retrieve item"})
		return
	}

	if existingItem == nil {
		utils.WriteJSON(w, http.StatusNotFound, utils.Envelope{"error": "Item not found"})
		return
	}

	var paramItem store.Item
	err = json.NewDecoder(r.Body).Decode(&paramItem)
	if err != nil {
		ih.logger.Printf("Error decoding request body: %v", err)
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{"error": "Invalid request payload"})
		return
	}

	if err := ih.validateCreateItemRequest(&paramItem); err != nil {
		ih.logger.Printf("Validation error: %v", err)
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{"error": err.Error()})
		return
	}

	if existingItem.OrganizationID != user.OrganizationID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	updatedItem, err := ih.itemStore.UpdateItem(existingItem)
	if err != nil {
		ih.logger.Printf("Error updating item: %v", err)
		utils.WriteJSON(w, http.StatusInternalServerError, utils.Envelope{"error": "Failed to update item"})
		return
	}

	utils.WriteJSON(w, http.StatusOK, utils.Envelope{"item": updatedItem})
}

func (ih *ItemHandler) HandleDeleteItem(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r)
	if user.IsAnonymous() {
		utils.WriteJSON(w, http.StatusUnauthorized, utils.Envelope{"error": "Unauthorized"})
		return
	}

	itemID, err := utils.ReadIDParam(r)
	if err != nil {
		ih.logger.Printf("Error reading ID parameter: %v", err)
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{"error": "Invalid ID parameter"})
		return
	}

	existingItem, err := ih.itemStore.GetItemByID(*itemID)
	if err != nil {
		ih.logger.Printf("Error retrieving item: %v", err)
		utils.WriteJSON(w, http.StatusInternalServerError, utils.Envelope{"error": "Failed to retrieve item"})
		return
	}

	if existingItem == nil {
		utils.WriteJSON(w, http.StatusNotFound, utils.Envelope{"error": "Item not found"})
		return
	}

	if existingItem.OrganizationID != user.OrganizationID {
		utils.WriteJSON(w, http.StatusForbidden, utils.Envelope{"error": "Forbidden"})
		return
	}

	err = ih.itemStore.DeleteItem(*itemID)
	if err != nil {
		ih.logger.Printf("Error deleting item: %v", err)
		utils.WriteJSON(w, http.StatusInternalServerError, utils.Envelope{"error": "Failed to delete item"})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (ih *ItemHandler) HandleGetItemsByOrganization(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r)
	if user.IsAnonymous() {
		utils.WriteJSON(w, http.StatusUnauthorized, utils.Envelope{"error": "Unauthorized"})
		return
	}

	pageSize, page := utils.PaginationParams(r)

	items, err := ih.itemStore.GetItemsByOrganization(page, pageSize, user.OrganizationID)
	if err != nil {
		ih.logger.Printf("Error fetching items: %v", err)
		utils.WriteJSON(w, http.StatusInternalServerError, utils.Envelope{"error": "Failed to fetch items"})
		return
	}

	ih.logger.Printf("Fetched %d items for organization %s", len(items), user.OrganizationID)
	ih.logger.Printf("Page: %d, Page Size: %d", page, pageSize)

	totalItems, err := ih.itemStore.CountItemsByOrganization(user.OrganizationID)
	if err != nil {
		ih.logger.Printf("Error counting items: %v", err)
		utils.WriteJSON(w, http.StatusInternalServerError, utils.Envelope{"error": "Failed to count items"})
		return
	}

	envelope := utils.Envelope{
		"data":      items,
		"count":     len(items),
		"total":     totalItems,
		"page":      page,
		"page_size": pageSize,
	}
	utils.WriteJSON(w, http.StatusOK, envelope)
}

func (ih *ItemHandler) validateCreateItemRequest(req *store.Item) error {
	if req.CategoryID == uuid.Nil {
		return errors.New("category_id is required")
	}

	if req.Name == "" {
		return errors.New("name is required")
	}

	if req.UnitPrice < 0 {
		return errors.New("unit_price is required")
	}

	if len(req.Stock) == 0 {
		return errors.New("stock cannot be empty")
	}

	for _, r := range req.Stock {
		if r.LocationID == uuid.Nil {
			return errors.New("location_id is required for stock entries")
		}

		if r.QuantityAvailable <= 0 {
			return errors.New("quantity_available must be greater than zero for stock entries")
		}

	}

	return nil

}
