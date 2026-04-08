// ABOUTME: JWT validation middleware using Auth0's go-jwt-middleware/v2 with JWKS caching.
// ABOUTME: Extracts sub claim, sanitizes to tenant ID, injects into request context.
package auth

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/auth0/go-jwt-middleware/v2/jwks"
	"github.com/auth0/go-jwt-middleware/v2/validator"
)

// TenantProvisioner is called when a tenant ID is seen for the first time.
type TenantProvisioner func(tenantID string) error

// NewJWTMiddleware creates an HTTP middleware that validates Auth0 JWT tokens.
// It extracts the sub claim, sanitizes it to a tenant ID, checks provisioning,
// and injects the tenant ID into the request context.
func NewJWTMiddleware(issuerURL, audience string, provisioner TenantProvisioner) func(http.Handler) http.Handler {
	issuer, err := url.Parse(issuerURL)
	if err != nil {
		log.Fatalf("auth: invalid issuer URL %q: %v", issuerURL, err)
	}

	provider := jwks.NewCachingProvider(issuer, 5*time.Minute)

	jwtValidator, err := validator.New(
		provider.KeyFunc,
		validator.RS256,
		issuerURL,
		[]string{audience},
	)
	if err != nil {
		log.Fatalf("auth: failed to create JWT validator: %v", err)
	}

	mw := jwtmiddleware.New(jwtValidator.ValidateToken,
		jwtmiddleware.WithErrorHandler(jwtErrorHandler),
	)

	// In-memory cache of provisioned tenants
	var mu sync.RWMutex
	provisioned := make(map[string]bool)

	return func(next http.Handler) http.Handler {
		return mw.CheckJWT(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := r.Context().Value(jwtmiddleware.ContextKey{}).(*validator.ValidatedClaims)
			if !ok || claims == nil {
				writeAuthError(w, http.StatusUnauthorized, "invalid token")
				return
			}

			sub := claims.RegisteredClaims.Subject
			if sub == "" {
				writeAuthError(w, http.StatusUnauthorized, "missing sub claim")
				return
			}

			tenantID := SanitizeTenantID(sub)

			// Check provisioning cache; provision on first encounter
			if provisioner != nil {
				mu.RLock()
				known := provisioned[tenantID]
				mu.RUnlock()

				if !known {
					if err := provisioner(tenantID); err != nil {
						log.Printf("auth: failed to provision tenant %s: %v", tenantID, err)
						writeAuthError(w, http.StatusServiceUnavailable, "tenant provisioning failed")
						return
					}
					mu.Lock()
					provisioned[tenantID] = true
					mu.Unlock()
					log.Printf("auth: provisioned tenant %s", tenantID)
				}
			}

			ctx := SetTenantID(r.Context(), tenantID)
			next.ServeHTTP(w, r.WithContext(ctx))
		}))
	}
}

// ExemptPaths wraps a handler to skip authentication for specific path prefixes.
// Also exempts OPTIONS requests (CORS preflight).
func ExemptPaths(paths []string, protected http.Handler, unprotected http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			unprotected.ServeHTTP(w, r)
			return
		}
		for _, p := range paths {
			if strings.HasPrefix(r.URL.Path, p) {
				unprotected.ServeHTTP(w, r)
				return
			}
		}
		protected.ServeHTTP(w, r)
	})
}

func jwtErrorHandler(w http.ResponseWriter, r *http.Request, err error) {
	msg := "invalid token"
	if strings.Contains(err.Error(), "expired") {
		msg = "token expired"
	} else if strings.Contains(err.Error(), "missing") {
		msg = "missing authorization header"
	}
	writeAuthError(w, http.StatusUnauthorized, msg)
}

func writeAuthError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

// ProvisionCheck is a helper that checks if a tenant is provisioned using
// an in-memory cache, calling the provisioner only on cache miss.
// This is used by the local middleware when PostgreSQL is in use.
func ProvisionCheck(provisioner TenantProvisioner) TenantProvisioner {
	if provisioner == nil {
		return nil
	}
	var mu sync.RWMutex
	cache := make(map[string]bool)
	return func(tenantID string) error {
		mu.RLock()
		known := cache[tenantID]
		mu.RUnlock()
		if known {
			return nil
		}
		if err := provisioner(tenantID); err != nil {
			return err
		}
		mu.Lock()
		cache[tenantID] = true
		mu.Unlock()
		return nil
	}
}

// contextKeyForProvisioner is used to pass provisioner through context in local mode.
type contextKeyForProvisioner struct{}

// WithProvisioner attaches a TenantProvisioner to the context.
func WithProvisioner(ctx context.Context, p TenantProvisioner) context.Context {
	return context.WithValue(ctx, contextKeyForProvisioner{}, p)
}
