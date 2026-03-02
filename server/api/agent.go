package api

import (
	"context"
	"net/http"

	"opslense-pulse/server/middleware"
	"opslense-pulse/server/utils"
	"opslense-pulse/shared"
)

type AgentRegistrar interface {
	UpsertAgentMetadata(ctx context.Context, m shared.HostMetrics) error
}

func RegisterAgentHandler(st AgentRegistrar) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		reqID := middleware.GetRequestID(r.Context())

		var a shared.AgentInfo
		if err := utils.DecodeJSONStrict(r, &a); err != nil {
			utils.WriteError(w, 400, "invalid_json", err.Error(), reqID)
			return
		}

		tenantID, ok := r.Context().Value(middleware.CtxTenantID).(string)
		if !ok || tenantID == "" {
			utils.WriteError(w, http.StatusUnauthorized, "unauthorized", "tenant missing", reqID)
			return
		}

		// 🔄 Convert AgentInfo → HostMetrics (minimal fields)
		m := shared.HostMetrics{
			TenantID:   tenantID,
			AgentID:    a.AgentID,
			Hostname:   a.Hostname,
			PrivateIP:  a.PrivateIP,
			PublicIP:   a.PublicIP,
			RemoteIP:   extractPublicIP(r),
			K8sNodeIP:  a.K8sNodeIP,
			OS:         a.OS,
			Version:    a.Version,
			Tags:       a.Tags,
			SystemTags: a.SystemTags,
		}

		if err := st.UpsertAgentMetadata(r.Context(), m); err != nil {
			utils.WriteError(w, 500, "internal_error", err.Error(), reqID)
			return
		}

		utils.WriteSuccess(w, http.StatusOK, nil, "registered", reqID)
	}
}
