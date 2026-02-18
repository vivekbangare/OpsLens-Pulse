package api

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"golang.org/x/crypto/bcrypt"

	"opslense-pulse/server/auth"
	"opslense-pulse/server/store"
)

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

func LoginHandler(db *sql.DB, jwtManager *auth.JWTManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// -----------------------------------------
		// Parse Request
		// -----------------------------------------
		var req LoginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}

		// -----------------------------------------
		// Validate Credentials
		// -----------------------------------------
		var userID string
		var passwordHash string
		var isSuperAdmin bool

		err := db.QueryRow(`
			SELECT id, password_hash, is_super_admin
			FROM users
			WHERE username = $1
			  AND is_active = true
		`, req.Username).Scan(&userID, &passwordHash, &isSuperAdmin)

		if err != nil {
			http.Error(w, "invalid credentials", http.StatusUnauthorized)
			return
		}

		if bcrypt.CompareHashAndPassword(
			[]byte(passwordHash),
			[]byte(req.Password),
		) != nil {
			http.Error(w, "invalid credentials", http.StatusUnauthorized)
			return
		}

		// -----------------------------------------
		// Resolve Default Tenant (Zero Trust)
		// -----------------------------------------
		var tenantID string
		err = db.QueryRow(`
			SELECT t.id
			FROM tenants t
			JOIN user_tenants ut ON ut.tenant_id = t.id
			WHERE ut.user_id = $1
			  AND t.is_active = true
			LIMIT 1
		`, userID).Scan(&tenantID)

		if err != nil {
			http.Error(w, "no active tenant assigned", http.StatusForbidden)
			return
		}

		// -----------------------------------------
		// Load Permissions For Tenant
		// -----------------------------------------
		permsMap, err := store.LoadUserPermissions(db, userID, tenantID)
		if err != nil {
			http.Error(w, "failed to load permissions", http.StatusInternalServerError)
			return
		}

		var permList []string
		for p := range permsMap {
			permList = append(permList, p)
		}

		// -----------------------------------------
		// Generate JWT (Tenant + Permissions inside)
		// -----------------------------------------
		token, err := jwtManager.Generate(
			userID,
			req.Username,
			tenantID,
			permList,
			isSuperAdmin,
		)
		if err != nil {
			http.Error(w, "failed to generate token", http.StatusInternalServerError)
			return
		}

		// -----------------------------------------
		// Return Token
		// -----------------------------------------
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(LoginResponse{
			Token: token,
		})
	}
}
