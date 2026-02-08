package api

import (
	"encoding/json"
	"net/http"

	"opslense-pulse/server/store"
)

func HostsHandler(s store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Optionally parse query params
		accountID := r.URL.Query().Get("account_id")

		filters := make(map[string]string)
		// parse filters from query if needed

		hosts, err := s.GetFiltered(accountID, filters)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(hosts)
	}
}
