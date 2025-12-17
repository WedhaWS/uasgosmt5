package middleware

import (
	"github.com/WedhaWS/uasgosmt5/app/model"
	"github.com/WedhaWS/uasgosmt5/app/repository"
	"github.com/WedhaWS/uasgosmt5/utils"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type AuthMiddleware struct {
	roleRepo *repository.RoleRepository
}

func NewAuthMiddleware(roleRepo *repository.RoleRepository) *AuthMiddleware {
	return &AuthMiddleware{roleRepo: roleRepo}
}

// ---------------------------------------------------------------------
// 1. AuthRequired (Authentication)
// Tugas: Cek apakah user sudah login (punya token valid)
// Flow FR-002: Step 1 (Ekstrak), Step 2 (Validasi), Step 3 (Load Perms)
// ---------------------------------------------------------------------
func (m *AuthMiddleware) AuthRequired() fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(401).JSON(model.WebResponse{Code: 401, Status: "error", Message: "Missing authorization header"})
		}

		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			return c.Status(401).JSON(model.WebResponse{Code: 401, Status: "error", Message: "Invalid token format"})
		}

		// Validasi Token & Ambil Claims
		claims, err := utils.ParseToken(tokenParts[1])
		if err != nil {
			return c.Status(401).JSON(model.WebResponse{Code: 401, Status: "error", Message: "Invalid or expired token"})
		}

		// Simpan data user ke Context (Locals) agar bisa dipakai di next handler
		c.Locals("user_id", claims.UserID)
		c.Locals("role", claims.Role)
		c.Locals("permissions", claims.Permissions) // LOAD Permissions dari Token (Cache)

		return c.Next()
	}
}

// ---------------------------------------------------------------------
// 2. PermissionRequired (Authorization / RBAC)
// Tugas: Cek apakah user punya hak akses spesifik
// Flow FR-002: Step 4 (Check), Step 5 (Allow/Deny)
// ---------------------------------------------------------------------
func (m *AuthMiddleware) PermissionRequired(requiredPerm string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Ambil permissions dari Context (yang diset oleh AuthRequired)
		userPermsInterface := c.Locals("permissions")
		if userPermsInterface == nil {
			return c.Status(403).JSON(model.WebResponse{Code: 403, Status: "error", Message: "No permissions found"})
		}

		// Casting ke []string
		// Karena dari JWT parsing, kadang terdeteksi sebagai []interface{}, kita handle keduanya
		var userPerms []string
		
		switch v := userPermsInterface.(type) {
		case []string:
			userPerms = v
		case []interface{}:
			for _, item := range v {
				if s, ok := item.(string); ok {
					userPerms = append(userPerms, s)
				}
			}
		}

		// Cek apakah requiredPerm ada di daftar permission user
		hasPermission := false
		for _, p := range userPerms {
			if p == requiredPerm {
				hasPermission = true
				break
			}
		}

		// Deny Request
		if !hasPermission {
			return c.Status(403).JSON(model.WebResponse{
				Code:    403,
				Status:  "error",
				Message: "Access denied. Missing permission: " + requiredPerm,
			})
		}

		// Allow Request
		return c.Next()
	}
}