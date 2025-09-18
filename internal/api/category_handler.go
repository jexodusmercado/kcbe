package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"kabancount/internal/middleware"
	"kabancount/internal/store"
	"kabancount/internal/utils"
	"log"
	"net/http"
)

type CategoryHandler struct {
	categoryStore store.CategoryStore
	logger        *log.Logger
}

func NewCategoryHandler(categoryStore store.CategoryStore, logger *log.Logger) *CategoryHandler {
	return &CategoryHandler{
		categoryStore: categoryStore,
		logger:        logger,
	}
}

func (ch *CategoryHandler) HandleCreateCategory(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r)

	if user.IsAnonymous() {
		utils.WriteJSON(w, http.StatusUnauthorized, utils.Envelope{"error": "Unauthorized"})
		return
	}

	var req store.Category
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		ch.logger.Printf("Error decoding request body: %v", err)
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{"error": "Invalid request payload"})
		return
	}

	if err := ch.validateCreateCategoryRequest(&req); err != nil {
		ch.logger.Printf("Validation error: %v", err)
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{"error": err.Error()})
		return
	}
	req.OrganizationID = user.OrganizationID

	createdCategory, err := ch.categoryStore.CreateCategory(&req)
	if err != nil {
		ch.logger.Printf("Error creating category: %v", err)
		utils.WriteJSON(w, http.StatusInternalServerError, utils.Envelope{"error": "Failed to create category"})
		return
	}

	utils.WriteJSON(w, http.StatusCreated, utils.Envelope{"category": createdCategory})
}

func (ch *CategoryHandler) HandleGetCategoryByID(w http.ResponseWriter, r *http.Request) {
	categoryID, err := utils.ReadIDParam(r)
	if err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{"error": "Invalid category ID"})
		return
	}

	category, err := ch.categoryStore.GetCategoryByID(*categoryID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			utils.WriteJSON(w, http.StatusNotFound, utils.Envelope{"error": "Category not found"})
			return
		}

		ch.logger.Printf("Error fetching category: %v", err)
		utils.WriteJSON(w, http.StatusInternalServerError, utils.Envelope{"error": "Failed to fetch category"})
		return
	}

	utils.WriteJSON(w, http.StatusOK, utils.Envelope{"category": category})
}

func (ch *CategoryHandler) HandleUpdateCategory(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r)
	if user.IsAnonymous() {
		utils.WriteJSON(w, http.StatusUnauthorized, utils.Envelope{"error": "Unauthorized"})
		return
	}

	categoryID, err := utils.ReadIDParam(r)
	if err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{"error": "Invalid category ID"})
		return
	}

	existingCategory, err := ch.categoryStore.GetCategoryByID(*categoryID)
	if err != nil {
		ch.logger.Printf("Error retrieving category: %v", err)
		utils.WriteJSON(w, http.StatusInternalServerError, utils.Envelope{"error": "Failed to retrieve category"})
		return
	}

	if existingCategory == nil {
		utils.WriteJSON(w, http.StatusNotFound, utils.Envelope{"error": "Category not found"})
		return
	}

	if existingCategory.OrganizationID != user.OrganizationID {
		utils.WriteJSON(w, http.StatusForbidden, utils.Envelope{"error": "Forbidden"})
		return
	}

	var req store.Category
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		ch.logger.Printf("Error decoding request body: %v", err)
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{"error": "Invalid request payload"})
		return
	}

	if req.Name != "" {
		existingCategory.Name = req.Name
	}

	if req.Description != nil {
		existingCategory.Description = req.Description
	}

	if user.OrganizationID != existingCategory.OrganizationID {
		utils.WriteJSON(w, http.StatusForbidden, utils.Envelope{"error": "You do not have permission to update this category"})
		return
	}

	updatedCategory, err := ch.categoryStore.UpdateCategory(existingCategory)
	if err != nil {
		ch.logger.Printf("Error updating category: %v", err)
		utils.WriteJSON(w, http.StatusInternalServerError, utils.Envelope{"error": "Failed to update category"})
		return
	}

	utils.WriteJSON(w, http.StatusOK, utils.Envelope{"category": updatedCategory})
}

func (ch *CategoryHandler) HandleDeleteCategory(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r)
	if user.IsAnonymous() {
		utils.WriteJSON(w, http.StatusUnauthorized, utils.Envelope{"error": "Unauthorized"})
		return
	}

	categoryID, err := utils.ReadIDParam(r)
	if err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{"error": "Invalid category ID"})
		return
	}

	existingCategory, err := ch.categoryStore.GetCategoryByID(*categoryID)
	if err != nil {
		ch.logger.Printf("Error retrieving category: %v", err)
		utils.WriteJSON(w, http.StatusInternalServerError, utils.Envelope{"error": "Failed to retrieve category"})
		return
	}

	if existingCategory == nil {
		utils.WriteJSON(w, http.StatusNotFound, utils.Envelope{"error": "Category not found"})
		return
	}

	if existingCategory.OrganizationID != user.OrganizationID {
		utils.WriteJSON(w, http.StatusForbidden, utils.Envelope{"error": "You do not have permission to delete this category"})
		return
	}

	err = ch.categoryStore.DeleteCategory(*categoryID)
	if err != nil {
		ch.logger.Printf("Error deleting category: %v", err)
		utils.WriteJSON(w, http.StatusInternalServerError, utils.Envelope{"error": "Failed to delete category"})
		return
	}

	utils.WriteJSON(w, http.StatusNoContent, nil)
}

func (ch *CategoryHandler) HandleGetCategoriesByOrganization(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r)
	if user.IsAnonymous() {
		utils.WriteJSON(w, http.StatusUnauthorized, utils.Envelope{"error": "Unauthorized"})
		return
	}

	pageSize, page := utils.PaginationParams(r)

	categories, err := ch.categoryStore.GetCategoryByOrganization(page, pageSize, user.OrganizationID)
	if err != nil {
		ch.logger.Printf("Error fetching categories: %v", err)
		utils.WriteJSON(w, http.StatusInternalServerError, utils.Envelope{"error": "Failed to fetch categories"})
		return
	}

	totalCategories, err := ch.categoryStore.CountCategoriesByOrganization(user.OrganizationID)
	if err != nil {
		ch.logger.Printf("Error counting categories: %v", err)
		utils.WriteJSON(w, http.StatusInternalServerError, utils.Envelope{"error": "Failed to count categories"})
		return
	}

	utils.WriteJSON(w, http.StatusOK, utils.Envelope{
		"items":     categories,
		"count":     len(categories),
		"total":     totalCategories,
		"page":      page,
		"page_size": pageSize,
	})
}

func (ch *CategoryHandler) validateCreateCategoryRequest(req *store.Category) error {
	if req.Name == "" {
		return errors.New("name is required")
	}

	return nil
}
