package api

import (
	"encoding/json"
	"net/http"

	"opslense-pulse/server/store"
)

func HostsHandler(st store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		accountID := r.URL.Query().Get("account_id")
		if accountID == "" {
			accountID = "default"
		}

		agents, err := st.ListAgents(accountID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(agents)
	}
}
