package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/paulochiaradia/dashtrack/internal/handlers"
	"github.com/paulochiaradia/dashtrack/internal/models"
	"github.com/paulochiaradia/dashtrack/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRoleHandlers(t *testing.T) {
	// Setup mock database
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	roleRepo := repository.NewRoleRepository(db)
	roleHandler := handlers.NewRoleHandler(roleRepo)

	t.Run("GetRoles", func(t *testing.T) {
		// Mock the database query
		rows := sqlmock.NewRows([]string{"id", "name", "description", "created_at", "updated_at"})

		adminID := uuid.New()
		driverID := uuid.New()
		helperID := uuid.New()

		rows.AddRow(adminID, "admin", "System administrator with full access", "2023-01-01T00:00:00Z", "2023-01-01T00:00:00Z")
		rows.AddRow(driverID, "driver", "Vehicle driver with access to assigned vehicles", "2023-01-01T00:00:00Z", "2023-01-01T00:00:00Z")
		rows.AddRow(helperID, "helper", "Driver assistant with limited access to assigned vehicles", "2023-01-01T00:00:00Z", "2023-01-01T00:00:00Z")

		mock.ExpectQuery("SELECT (.+) FROM roles").WillReturnRows(rows)

		req, err := http.NewRequest("GET", "/roles", nil)
		require.NoError(t, err)

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(roleHandler.ListRoles)
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var roles []models.Role
		err = json.Unmarshal(rr.Body.Bytes(), &roles)
		require.NoError(t, err)
		assert.Len(t, roles, 3)
		assert.Equal(t, "admin", roles[0].Name)
		assert.Equal(t, "driver", roles[1].Name)
		assert.Equal(t, "helper", roles[2].Name)
	})

	// Ensure all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}
