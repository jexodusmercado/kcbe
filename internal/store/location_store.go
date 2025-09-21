package store

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type Location struct {
	ID             uuid.UUID `json:"id"`
	OrganizationID uuid.UUID `json:"organization_id"`
	Name           string    `json:"name"`
	Description    string    `json:"description,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type PostgresLocationStore struct {
	db *sql.DB
}

func NewPostgresLocationStore(db *sql.DB) *PostgresLocationStore {
	return &PostgresLocationStore{db: db}
}

type LocationStore interface {
	CreateLocation(location *Location) (*Location, error)
	GetLocationsByOrganization(organizationID uuid.UUID) ([]Location, error)
}

func (s *PostgresLocationStore) CreateLocation(location *Location) (*Location, error) {
	query := `
		INSERT INTO locations (organization_id, name, description)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at
	`
	row := s.db.QueryRow(
		query,
		location.OrganizationID,
		location.Name,
		location.Description,
	)

	err := row.Scan(&location.ID, &location.CreatedAt, &location.UpdatedAt)

	if err != nil {
		return nil, err
	}
	return location, nil
}

func (s *PostgresLocationStore) GetLocationsByOrganization(organizationID uuid.UUID) ([]Location, error) {
	query := `
		SELECT id, organization_id, name, description, created_at, updated_at
		FROM locations
		WHERE organization_id = $1
	`
	rows, err := s.db.Query(query, organizationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var locations []Location
	for rows.Next() {
		var location Location

		if err := rows.Scan(
			&location.ID,
			&location.OrganizationID,
			&location.Name,
			&location.Description,
			&location.CreatedAt,
			&location.UpdatedAt,
		); err != nil {
			return nil, err
		}
		locations = append(locations, location)

	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return locations, nil
}
