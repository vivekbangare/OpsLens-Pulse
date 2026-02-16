package middleware

import "net/http"

func RequirePermission(permission string) func(http.Handler) http.Handler {

	return func(next http.Handler) http.Handler {

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			perms, ok := r.Context().Value(CtxPermissions).(map[string]bool)

			if !ok {
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}

			if !perms[permission] {
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
