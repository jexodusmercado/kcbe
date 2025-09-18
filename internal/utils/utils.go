package utils

import (
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type Envelope map[string]any

func WriteJSON(w http.ResponseWriter, status int, data Envelope) error {

	js, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	js = append(js, '\n')

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(js)
	return nil
}

func ReadIDParam(r *http.Request) (*uuid.UUID, error) {
	paramsID := chi.URLParam(r, "id")
	if paramsID == "" {
		return nil, errors.New("invalid or missing id parameter")
	}

	ID, err := uuid.Parse(paramsID)
	if err != nil {
		return nil, errors.New("invalid id parameter")
	}

	return &ID, nil
}

func PaginationParams(r *http.Request) (limit, offset int) {
	query := r.URL.Query()

	limit = 20
	offset = 0

	if l := query.Get("limit"); l != "" {
		var err error
		limit, err = strconv.Atoi(l)
		if err != nil || limit < 1 {
			limit = 20
		}
	}

	if o := query.Get("offset"); o != "" {
		var err error
		offset, err = strconv.Atoi(o)
		if err != nil || offset < 0 {
			offset = 0
		}
	}

	return limit, offset
}

func IsValidEmail(email string) bool {
	// A very simple email validation
	if len(email) < 3 || len(email) > 254 {
		return false
	}
	// More complex regex can be added here for better validation
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

func IsPasswordStrong(password string) bool {
	var (
		hasMinLen  = false
		hasUpper   = false
		hasLower   = false
		hasNumber  = false
		hasSpecial = false
	)

	if len(password) >= 8 {
		hasMinLen = true
	}

	for _, char := range password {
		switch {
		case 'A' <= char && char <= 'Z':
			hasUpper = true
		case 'a' <= char && char <= 'z':
			hasLower = true
		case '0' <= char && char <= '9':
			hasNumber = true
		case (char >= 33 && char <= 47) || (char >= 58 && char <= 64) ||
			(char >= 91 && char <= 96) || (char >= 123 && char <= 126):
			hasSpecial = true
		}

		if hasMinLen && hasUpper && hasLower && hasNumber && hasSpecial {
			return true
		}

	}

	return false
}
