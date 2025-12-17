package test

import (
	"encoding/json"
	"net/http/httptest"
	"os"
	"testing"
	"github.com/WedhaWS/uasgosmt5/app/model"
	"github.com/WedhaWS/uasgosmt5/middleware"
	"github.com/WedhaWS/uasgosmt5/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

func TestAuthMiddleware_AuthRequired(t *testing.T) {
	// Set test JWT secret
	os.Setenv("JWT_SECRET", "test_secret_key_for_unit_testing")
	defer os.Unsetenv("JWT_SECRET")

	tests := []struct {
		name           string
		authHeader     string
		expectedStatus int
		expectedError  string
		shouldCallNext bool
	}{
		{
			name: "Valid token",
			authHeader: func() string {
				token, _ := utils.GenerateToken("user-123", "Admin", []string{"user:manage"})
				return "Bearer " + token
			}(),
			expectedStatus: 200,
			shouldCallNext: true,
		},
		{
			name:           "Missing authorization header",
			authHeader:     "",
			expectedStatus: 401,
			expectedError:  "Missing authorization header",
			shouldCallNext: false,
		},
		{
			name:           "Invalid token format - no Bearer",
			authHeader:     "InvalidToken",
			expectedStatus: 401,
			expectedError:  "Invalid token format",
			shouldCallNext: false,
		},
		{
			name:           "Invalid token format - malformed",
			authHeader:     "Bearer",
			expectedStatus: 401,
			expectedError:  "Invalid token format",
			shouldCallNext: false,
		},
		{
			name:           "Invalid token - expired/malformed",
			authHeader:     "Bearer invalid.token.here",
			expectedStatus: 401,
			expectedError:  "Invalid or expired token",
			shouldCallNext: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create middleware with nil repository (AuthRequired doesn't use it)
			authMiddleware := middleware.NewAuthMiddleware(nil)

			// Setup Fiber app
			app := fiber.New()
			nextCalled := false

			// Add middleware and test route
			app.Use(authMiddleware.AuthRequired())
			app.Get("/test", func(c *fiber.Ctx) error {
				nextCalled = true
				return c.JSON(fiber.Map{"message": "success"})
			})

			// Create request
			req := httptest.NewRequest("GET", "/test", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			// Execute request
			resp, err := app.Test(req)
			assert.NoError(t, err)

			// Assertions
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
			assert.Equal(t, tt.shouldCallNext, nextCalled)

			if tt.expectedError != "" {
				// Parse response body to check error message
				var response model.WebResponse
				err := json.NewDecoder(resp.Body).Decode(&response)
				assert.NoError(t, err)
				assert.Contains(t, response.Message, tt.expectedError)
			}

			// If token is valid, check if user data is set in context
			if tt.shouldCallNext && tt.expectedStatus == 200 {
				assert.True(t, nextCalled)
			}
		})
	}
}

func TestAuthMiddleware_PermissionRequired(t *testing.T) {
	// Set test JWT secret
	os.Setenv("JWT_SECRET", "test_secret_key_for_unit_testing")
	defer os.Unsetenv("JWT_SECRET")

	tests := []struct {
		name            string
		requiredPerm    string
		userPermissions []string
		expectedStatus  int
		expectedError   string
		shouldCallNext  bool
	}{
		{
			name:            "User has required permission",
			requiredPerm:    "user:manage",
			userPermissions: []string{"user:manage", "achievement:verify"},
			expectedStatus:  200,
			shouldCallNext:  true,
		},
		{
			name:            "User missing required permission",
			requiredPerm:    "admin:delete",
			userPermissions: []string{"user:manage", "achievement:verify"},
			expectedStatus:  403,
			expectedError:   "Access denied. Missing permission: admin:delete",
			shouldCallNext:  false,
		},
		{
			name:            "User has no permissions",
			requiredPerm:    "user:manage",
			userPermissions: []string{},
			expectedStatus:  403,
			expectedError:   "Access denied",
			shouldCallNext:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create middleware with nil repository (PermissionRequired doesn't use it)
			authMiddleware := middleware.NewAuthMiddleware(nil)

			// Setup Fiber app
			app := fiber.New()
			nextCalled := false

			// Add middleware that sets permissions in context
			app.Use(func(c *fiber.Ctx) error {
				c.Locals("permissions", tt.userPermissions)
				return c.Next()
			})

			// Add permission middleware and test route
			app.Use(authMiddleware.PermissionRequired(tt.requiredPerm))
			app.Get("/test", func(c *fiber.Ctx) error {
				nextCalled = true
				return c.JSON(fiber.Map{"message": "success"})
			})

			// Create request
			req := httptest.NewRequest("GET", "/test", nil)

			// Execute request
			resp, err := app.Test(req)
			assert.NoError(t, err)

			// Assertions
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
			assert.Equal(t, tt.shouldCallNext, nextCalled)

			if tt.expectedError != "" {
				// Parse response body to check error message
				var response model.WebResponse
				err := json.NewDecoder(resp.Body).Decode(&response)
				assert.NoError(t, err)
				assert.Contains(t, response.Message, tt.expectedError)
			}
		})
	}
}

func TestAuthMiddleware_PermissionRequired_NoPermissionsInContext(t *testing.T) {
	// Create middleware with nil repository
	authMiddleware := middleware.NewAuthMiddleware(nil)

	// Setup Fiber app
	app := fiber.New()
	nextCalled := false

	// Add permission middleware without setting permissions in context
	app.Use(authMiddleware.PermissionRequired("user:manage"))
	app.Get("/test", func(c *fiber.Ctx) error {
		nextCalled = true
		return c.JSON(fiber.Map{"message": "success"})
	})

	// Create request
	req := httptest.NewRequest("GET", "/test", nil)

	// Execute request
	resp, err := app.Test(req)
	assert.NoError(t, err)

	// Assertions
	assert.Equal(t, 403, resp.StatusCode)
	assert.False(t, nextCalled)

	// Parse response body to check error message
	var response model.WebResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Contains(t, response.Message, "No permissions found")
}

func TestAuthMiddleware_IntegrationFlow(t *testing.T) {
	// Set test JWT secret
	os.Setenv("JWT_SECRET", "test_secret_key_for_unit_testing")
	defer os.Unsetenv("JWT_SECRET")

	// Generate valid token
	token, err := utils.GenerateToken("user-123", "Admin", []string{"user:manage", "achievement:verify"})
	assert.NoError(t, err)

	// Create middleware with nil repository
	authMiddleware := middleware.NewAuthMiddleware(nil)

	// Setup Fiber app with both middlewares
	app := fiber.New()
	var capturedUserID, capturedRole string
	var capturedPermissions []string

	app.Use(authMiddleware.AuthRequired())
	app.Use(authMiddleware.PermissionRequired("user:manage"))
	app.Get("/test", func(c *fiber.Ctx) error {
		capturedUserID = c.Locals("user_id").(string)
		capturedRole = c.Locals("role").(string)

		// Handle permissions type assertion
		permsInterface := c.Locals("permissions")
		switch v := permsInterface.(type) {
		case []string:
			capturedPermissions = v
		case []interface{}:
			for _, item := range v {
				if s, ok := item.(string); ok {
					capturedPermissions = append(capturedPermissions, s)
				}
			}
		}

		return c.JSON(fiber.Map{
			"message":     "success",
			"user_id":     capturedUserID,
			"role":        capturedRole,
			"permissions": capturedPermissions,
		})
	})

	// Create request with valid token
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	// Execute request
	resp, err := app.Test(req)
	assert.NoError(t, err)

	// Assertions
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "user-123", capturedUserID)
	assert.Equal(t, "Admin", capturedRole)
	assert.Contains(t, capturedPermissions, "user:manage")
	assert.Contains(t, capturedPermissions, "achievement:verify")
}
