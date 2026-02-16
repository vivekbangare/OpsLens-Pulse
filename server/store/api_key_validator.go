package store

type APIKeyValidator interface {
	ValidateAPIKey(rawKey string) (bool, string, error)
}
