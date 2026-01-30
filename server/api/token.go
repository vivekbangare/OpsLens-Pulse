package api

var authToken string

func SetAuthToken(t string) {
	authToken = t
}

func GetAuthToken() string {
	return authToken
}
