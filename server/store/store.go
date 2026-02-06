package store

import "opslense-pulse/shared"

type Store interface {
	SaveMetrics(m shared.HostMetrics) error
	UpdateHeartbeat(hostname string) error
	//	GetFiltered(tags map[string]string) ([]map[string]interface{}, error)

	//API Key management
	CountAPIKeys() (int, error)
	InsertAPIKey(APIKey) error
	ValidateAPIKey(rawKey string) (bool, error)
}
