package store

import (
	"database/sql"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("pgx", "host=localhost user=postgres password=password dbname=kabancount_test port=5433 sslmode=disable")
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	err = Migrate(db, "../../migrations")
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	_, err = db.Exec(`TRUNCATE TABLE organizations RESTART IDENTITY CASCADE`)
	if err != nil {
		t.Fatalf("Failed to truncate test database: %v", err)
	}

	return db
}

func TestCreateOrganization(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	store := NewPostgresOrganizationStore(db)

	tests := []struct {
		name    string
		org     *Organization
		wantErr bool
	}{
		{
			name: "Valid Organization",
			org: &Organization{
				Name: "Test Org",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			createdOrg, err := store.CreateOrganization(tt.org)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateOrganization() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, createdOrg.ID)
			assert.Equal(t, tt.org.Name, createdOrg.Name)
			assert.False(t, createdOrg.CreatedAt.IsZero())
			assert.False(t, createdOrg.UpdatedAt.IsZero())

			retrievedOrg, err := store.GetOrganizationByID(createdOrg.ID)
			require.NoError(t, err)
			assert.Equal(t, createdOrg.ID, retrievedOrg.ID)
			assert.Equal(t, createdOrg.Name, retrievedOrg.Name)
		})
	}
}
