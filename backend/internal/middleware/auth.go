package middleware

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/davidtaing/property-management/api"
	"github.com/davidtaing/property-management/internal/types"
	"github.com/gorilla/mux"
)

func AuthMiddleware() mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := clerk.SessionClaimsFromContext(r.Context())

			if !ok {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(api.Error{
					Code:    http.StatusUnauthorized,
					Message: "Unauthorized",
				})
				return
			}

			org := claims.ActiveOrganizationID
			orgRole := claims.ActiveOrganizationRole
			userId := claims.Subject

			ctx := context.WithValue(r.Context(), types.OrgIDKey, org)
			ctx = context.WithValue(ctx, types.OrgRoleKey, orgRole)
			ctx = context.WithValue(ctx, types.UserIDKey, userId)

			r = r.WithContext(ctx)

			// Call the next handler
			next.ServeHTTP(w, r)
		})
	}
}
