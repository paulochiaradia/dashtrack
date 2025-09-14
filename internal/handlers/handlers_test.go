package handlers_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/paulochiaradia/dashtrack/internal/handlers"
	"github.com/paulochiaradia/dashtrack/internal/models"
	"github.com/paulochiaradia/dashtrack/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHealthHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/health", nil)
	require.NoError(t, err)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"status":  "ok",
			"message": "API is running",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "ok", response["status"])
}

func TestUserHandlers(t *testing.T) {
	// Setup mock database
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	userRepo := repository.NewUserRepository(db)
	roleRepo := repository.NewRoleRepository(db)
	userHandler := handlers.NewUserHandler(userRepo, roleRepo)

	t.Run("GetUsers", func(t *testing.T) {
		// Mock the database query
		rows := sqlmock.NewRows([]string{
			"id", "name", "email", "phone", "cpf", "avatar", "role_id",
			"active", "last_login", "dashboard_config", "login_attempts",
			"blocked_until", "password_changed_at", "created_at", "updated_at",
			"id", "name", "description", "created_at", "updated_at",
		})

		userID := uuid.New()
		roleID := uuid.New()

		rows.AddRow(
			userID, "Test User", "test@example.com", "123456789", "12345678901", "avatar.jpg", roleID,
			true, nil, nil, 0, nil, "2023-01-01T00:00:00Z", "2023-01-01T00:00:00Z", "2023-01-01T00:00:00Z",
			roleID, "Admin", "Administrator role", "2023-01-01T00:00:00Z", "2023-01-01T00:00:00Z",
		)

		mock.ExpectQuery(".*").
			WillReturnRows(rows)

		req, err := http.NewRequest("GET", "/users", nil)
		require.NoError(t, err)

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(userHandler.ListUsers)
		handler.ServeHTTP(rr, req)

		// Debug: print response body
		t.Logf("Response status: %d", rr.Code)
		t.Logf("Response body: %s", rr.Body.String())

		assert.Equal(t, http.StatusOK, rr.Code)

		var response struct {
			Users      []models.User `json:"users"`
			Pagination struct {
				Limit  int `json:"limit"`
				Offset int `json:"offset"`
				Count  int `json:"count"`
			} `json:"pagination"`
		}
		err = json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Len(t, response.Users, 1)
		assert.Equal(t, "Test User", response.Users[0].Name)
	})

	t.Run("CreateUser", func(t *testing.T) {
		userRequest := models.CreateUserRequest{
			Name:     "New User",
			Email:    "new@example.com",
			Password: "password123",
			Phone:    "987654321",
			RoleID:   uuid.New().String(),
		}

		// Mock role validation query
		roleRows := sqlmock.NewRows([]string{"id", "name", "description", "created_at", "updated_at"})
		roleRows.AddRow(userRequest.RoleID, "admin", "Administrator", "2023-01-01T00:00:00Z", "2023-01-01T00:00:00Z")
		mock.ExpectQuery("SELECT (.+) FROM roles WHERE id = ?").WithArgs(userRequest.RoleID).WillReturnRows(roleRows)

		// Mock user creation
		mock.ExpectQuery("INSERT INTO users").
			WithArgs(
				sqlmock.AnyArg(), // id (UUID)
				userRequest.Name,
				userRequest.Email,
				sqlmock.AnyArg(), // hashed password
				userRequest.Phone,
				sql.NullString{}, // cpf
				sql.NullString{}, // avatar
				userRequest.RoleID,
				true,             // active
				sqlmock.AnyArg(), // created_at
				sqlmock.AnyArg(), // updated_at
			).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "name", "email", "password", "phone", "cpf", "avatar",
				"role_id", "active", "last_login", "dashboard_config", "api_token",
				"login_attempts", "blocked_until", "password_changed_at",
				"created_at", "updated_at",
			}).AddRow(
				uuid.New(), userRequest.Name, userRequest.Email, "hashed_password",
				userRequest.Phone, nil, nil, userRequest.RoleID, true,
				nil, nil, nil, 0, nil, "2023-01-01T00:00:00Z",
				"2023-01-01T00:00:00Z", "2023-01-01T00:00:00Z",
			))

		body, err := json.Marshal(userRequest)
		require.NoError(t, err)

		req, err := http.NewRequest("POST", "/users", bytes.NewBuffer(body))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(userHandler.CreateUser)
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusCreated, rr.Code)

		var user models.User
		err = json.Unmarshal(rr.Body.Bytes(), &user)
		require.NoError(t, err)
		assert.Equal(t, userRequest.Name, user.Name)
		assert.Equal(t, userRequest.Email, user.Email)
	})

	t.Run("GetUserByID", func(t *testing.T) {
		userID := uuid.New()
		roleID := uuid.New()

		rows := sqlmock.NewRows([]string{
			"id", "name", "email", "password", "phone", "cpf", "avatar",
			"role_id", "active", "last_login", "dashboard_config", "api_token",
			"login_attempts", "blocked_until", "password_changed_at",
			"created_at", "updated_at",
		})

		rows.AddRow(
			userID, "Test User", "test@example.com", "hashed_password",
			"123456789", "12345678901", "avatar.jpg", roleID, true,
			nil, nil, nil, 0, nil, "2023-01-01T00:00:00Z",
			"2023-01-01T00:00:00Z", "2023-01-01T00:00:00Z",
		)

		mock.ExpectQuery("SELECT (.+) FROM users WHERE id = ?").
			WithArgs(userID).
			WillReturnRows(rows)

		req, err := http.NewRequest("GET", "/users/"+userID.String(), nil)
		require.NoError(t, err)

		rr := httptest.NewRecorder()

		// Setup router to handle path parameter
		router := mux.NewRouter()
		router.HandleFunc("/users/{id}", userHandler.GetUser).Methods("GET")
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var user models.User
		err = json.Unmarshal(rr.Body.Bytes(), &user)
		require.NoError(t, err)
		assert.Equal(t, userID, user.ID)
		assert.Equal(t, "Test User", user.Name)
	})

	t.Run("UpdateUser", func(t *testing.T) {
		userID := uuid.New()
		roleID := uuid.New()

		updateRequest := models.UpdateUserRequest{
			Name:  "Updated User",
			Email: "updated@example.com",
			Phone: "555666777",
		}

		// Mock role validation
		roleRows := sqlmock.NewRows([]string{"id", "name", "description", "created_at", "updated_at"})
		roleRows.AddRow(roleID, "admin", "Administrator", "2023-01-01T00:00:00Z", "2023-01-01T00:00:00Z")
		mock.ExpectQuery("SELECT (.+) FROM roles WHERE id = ?").WithArgs(roleID).WillReturnRows(roleRows)

		// Mock user update
		mock.ExpectQuery("UPDATE users SET").
			WithArgs(
				updateRequest.Name,
				updateRequest.Email,
				updateRequest.Phone,
				sqlmock.AnyArg(), // updated_at
				userID,
			).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "name", "email", "password", "phone", "cpf", "avatar",
				"role_id", "active", "last_login", "dashboard_config", "api_token",
				"login_attempts", "blocked_until", "password_changed_at",
				"created_at", "updated_at",
			}).AddRow(
				userID, updateRequest.Name, updateRequest.Email, "hashed_password",
				updateRequest.Phone, nil, nil, roleID, true,
				nil, nil, nil, 0, nil, "2023-01-01T00:00:00Z",
				"2023-01-01T00:00:00Z", "2023-01-01T00:00:00Z",
			))

		body, err := json.Marshal(updateRequest)
		require.NoError(t, err)

		req, err := http.NewRequest("PUT", "/users/"+userID.String(), bytes.NewBuffer(body))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()

		// Setup router to handle path parameter
		router := mux.NewRouter()
		router.HandleFunc("/users/{id}", userHandler.UpdateUser).Methods("PUT")
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var user models.User
		err = json.Unmarshal(rr.Body.Bytes(), &user)
		require.NoError(t, err)
		assert.Equal(t, updateRequest.Name, user.Name)
		assert.Equal(t, updateRequest.Email, user.Email)
	})

	t.Run("DeleteUser", func(t *testing.T) {
		userID := uuid.New()

		mock.ExpectExec("DELETE FROM users WHERE id = ?").
			WithArgs(userID).
			WillReturnResult(sqlmock.NewResult(1, 1))

		req, err := http.NewRequest("DELETE", "/users/"+userID.String(), nil)
		require.NoError(t, err)

		rr := httptest.NewRecorder()

		// Setup router to handle path parameter
		router := mux.NewRouter()
		router.HandleFunc("/users/{id}", userHandler.DeleteUser).Methods("DELETE")
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var response map[string]interface{}
		err = json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "User deleted successfully", response["message"])
	})

	// Ensure all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}
