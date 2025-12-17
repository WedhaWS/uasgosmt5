package test

import (
	"os"
	"testing"
	"time"
	"github.com/WedhaWS/uasgosmt5/utils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test JWT utilities
func TestJWTUtils_GenerateToken(t *testing.T) {
	// Set test JWT secret
	os.Setenv("JWT_SECRET", "test_secret_key_for_unit_testing")
	defer os.Unsetenv("JWT_SECRET")

	tests := []struct {
		name        string
		userID      string
		role        string
		permissions []string
		expectError bool
	}{
		{
			name:        "Valid admin token generation",
			userID:      "admin-123",
			role:        "Admin",
			permissions: []string{"user:manage", "achievement:verify"},
			expectError: false,
		},
		{
			name:        "Valid student token generation",
			userID:      "student-456",
			role:        "Mahasiswa",
			permissions: []string{"achievement:create", "achievement:update"},
			expectError: false,
		},
		{
			name:        "Valid lecturer token generation",
			userID:      "lecturer-789",
			role:        "Dosen Wali",
			permissions: []string{"achievement:verify"},
			expectError: false,
		},
		{
			name:        "Empty permissions should work",
			userID:      "user-000",
			role:        "Guest",
			permissions: []string{},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := utils.GenerateToken(tt.userID, tt.role, tt.permissions)

			if tt.expectError {
				assert.Error(t, err)
				assert.Empty(t, token)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, token)
				assert.Contains(t, token, ".")

				// Verify token can be parsed back
				claims, parseErr := utils.ParseToken(token)
				require.NoError(t, parseErr)
				assert.Equal(t, tt.userID, claims.UserID)
				assert.Equal(t, tt.role, claims.Role)
				assert.Equal(t, tt.permissions, claims.Permissions)

				// Check expiration is set properly (24 hours from now)
				expectedExp := time.Now().Add(24 * time.Hour)
				assert.WithinDuration(t, expectedExp, claims.ExpiresAt.Time, time.Minute)
			}
		})
	}
}

func TestJWTUtils_ParseToken(t *testing.T) {
	os.Setenv("JWT_SECRET", "test_secret_key_for_unit_testing")
	defer os.Unsetenv("JWT_SECRET")

	// Generate valid tokens for testing
	adminToken, _ := utils.GenerateToken("admin-123", "Admin", []string{"user:manage"})
	studentToken, _ := utils.GenerateToken("student-456", "Mahasiswa", []string{"achievement:create"})

	tests := []struct {
		name          string
		token         string
		expectError   bool
		expectedID    string
		expectedRole  string
		expectedPerms []string
	}{
		{
			name:          "Valid admin token",
			token:         adminToken,
			expectError:   false,
			expectedID:    "admin-123",
			expectedRole:  "Admin",
			expectedPerms: []string{"user:manage"},
		},
		{
			name:          "Valid student token",
			token:         studentToken,
			expectError:   false,
			expectedID:    "student-456",
			expectedRole:  "Mahasiswa",
			expectedPerms: []string{"achievement:create"},
		},
		{
			name:        "Invalid token format",
			token:       "invalid.token.format",
			expectError: true,
		},
		{
			name:        "Empty token",
			token:       "",
			expectError: true,
		},
		{
			name:        "Malformed JWT",
			token:       "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.invalid.signature",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := utils.ParseToken(tt.token)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, claims)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, claims)
				assert.Equal(t, tt.expectedID, claims.UserID)
				assert.Equal(t, tt.expectedRole, claims.Role)
				assert.Equal(t, tt.expectedPerms, claims.Permissions)

				// Check that issued at is reasonable
				assert.True(t, claims.IssuedAt.Time.Before(time.Now().Add(time.Minute)))
				assert.True(t, claims.IssuedAt.Time.After(time.Now().Add(-time.Minute)))
			}
		})
	}
}

// Test Password utilities
func TestPasswordUtils_HashPassword(t *testing.T) {
	tests := []struct {
		name        string
		password    string
		expectError bool
	}{
		{
			name:        "Valid password",
			password:    "password123",
			expectError: false,
		},
		{
			name:        "Empty password",
			password:    "",
			expectError: false, // bcrypt can handle empty strings
		},
		{
			name:        "Long password",
			password:    "this_is_a_very_long_password_that_should_still_work_fine",
			expectError: false,
		},
		{
			name:        "Password with special characters",
			password:    "p@ssw0rd!@#$%^&*()",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := utils.HashPassword(tt.password)

			if tt.expectError {
				assert.Error(t, err)
				assert.Empty(t, hash)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, hash)
				assert.NotEqual(t, tt.password, hash) // Hash should be different from original
				assert.True(t, len(hash) > 50)        // bcrypt hashes are typically 60 characters

				// Verify the hash can be used to check the original password
				isValid := utils.CheckPasswordHash(tt.password, hash)
				assert.True(t, isValid)
			}
		})
	}
}

func TestPasswordUtils_CheckPasswordHash(t *testing.T) {
	// Pre-generate some test hashes
	validPassword := "testpassword123"
	validHash, _ := utils.HashPassword(validPassword)

	tests := []struct {
		name     string
		password string
		hash     string
		expected bool
	}{
		{
			name:     "Correct password and hash",
			password: validPassword,
			hash:     validHash,
			expected: true,
		},
		{
			name:     "Incorrect password",
			password: "wrongpassword",
			hash:     validHash,
			expected: false,
		},
		{
			name:     "Empty password with valid hash",
			password: "",
			hash:     validHash,
			expected: false,
		},
		{
			name:     "Valid password with empty hash",
			password: validPassword,
			hash:     "",
			expected: false,
		},
		{
			name:     "Both empty",
			password: "",
			hash:     "",
			expected: false,
		},
		{
			name:     "Invalid hash format",
			password: validPassword,
			hash:     "invalid_hash_format",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.CheckPasswordHash(tt.password, tt.hash)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Benchmark tests
func BenchmarkHashPassword(b *testing.B) {
	password := "benchmarkpassword123"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := utils.HashPassword(password)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCheckPasswordHash(b *testing.B) {
	password := "benchmarkpassword123"
	hash, _ := utils.HashPassword(password)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		utils.CheckPasswordHash(password, hash)
	}
}
