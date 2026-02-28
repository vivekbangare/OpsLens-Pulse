package api

import (
	"database/sql"
	"net/http"

	"golang.org/x/crypto/bcrypt"

	"opslense-pulse/server/auth"
	"opslense-pulse/server/middleware"
	"opslense-pulse/server/store"
	"opslense-pulse/server/utils"
)

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

func LoginHandler(
	db *sql.DB,
	jwtManager *auth.JWTManager,
	pgStore *store.PostgresStore,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		reqID := middleware.GetRequestID(r.Context())

		// -----------------------------------------
		// Parse Request (STRICT JSON)
		// -----------------------------------------
		var req LoginRequest
		if err := utils.DecodeJSONStrict(r, &req); err != nil {
			utils.WriteError(w, http.StatusBadRequest, "invalid_json", "invalid request body", reqID)
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

		if err == sql.ErrNoRows {

			go pgStore.InsertAuditLog(r.Context(), store.AuditLog{
				TenantID:  "unknown", // unknown at this point
				UserID:    nil,
				Action:    "login",
				Status:    "failure",
				IPAddress: utils.GetClientIP(r),
				UserAgent: r.UserAgent(),
				Metadata: map[string]interface{}{
					"username": req.Username,
					"reason":   "user_not_found",
				},
			})

			utils.WriteError(w, http.StatusUnauthorized, "unauthorized", "invalid credentials", reqID)
			return
		}

		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, "internal_error", "database error", reqID)
			return
		}

		if bcrypt.CompareHashAndPassword(
			[]byte(passwordHash),
			[]byte(req.Password),
		) != nil {

			go pgStore.InsertAuditLog(r.Context(), store.AuditLog{
				TenantID:  "unknown", // still unknown until tenant resolution
				UserID:    &userID,
				Action:    "login",
				Status:    "failure",
				IPAddress: utils.GetClientIP(r),
				UserAgent: r.UserAgent(),
				Metadata: map[string]interface{}{
					"reason": "invalid_password",
				},
			})

			utils.WriteError(w, http.StatusUnauthorized, "unauthorized", "invalid credentials", reqID)
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

			go pgStore.InsertAuditLog(r.Context(), store.AuditLog{
				TenantID:  "unknown",
				UserID:    &userID,
				Action:    "login",
				Status:    "failure",
				IPAddress: utils.GetClientIP(r),
				UserAgent: r.UserAgent(),
				Metadata: map[string]interface{}{
					"reason": "no_active_tenant",
				},
			})

			utils.WriteError(w, http.StatusForbidden, "forbidden", "no active tenant assigned", reqID)
			return
		}

		// -----------------------------------------
		// Load Permissions For Tenant
		// -----------------------------------------
		permsMap, err := store.LoadUserPermissions(db, userID, tenantID)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, "internal_error", "failed to load permissions", reqID)
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
			utils.WriteError(w, http.StatusInternalServerError, "internal_error", "failed to generate token", reqID)
			return
		}
		// -----------------------------------------
		// Audit: Successful Login
		// -----------------------------------------
		go pgStore.InsertAuditLog(r.Context(), store.AuditLog{
			TenantID:  tenantID,
			UserID:    &userID,
			Action:    "login",
			Status:    "success",
			IPAddress: utils.GetClientIP(r),
			UserAgent: r.UserAgent(),
			Metadata: map[string]interface{}{
				"is_super_admin": isSuperAdmin,
			},
		})

		// -----------------------------------------
		// Return Token (SAME RESPONSE FORMAT)
		// -----------------------------------------
		utils.WriteSuccess(
			w,
			http.StatusOK,
			LoginResponse{Token: token},
			"login successful",
			reqID,
		)
	}
}
