package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/beetrack/backend/internal/middleware"
	"github.com/beetrack/backend/pkg/respond"
)

// requireAuth extracts the authenticated user ID from the request context,
// responding with 401 MISSING_TOKEN and returning ok=false if it's absent.
// Collapses the auth-context check every authenticated handler needs.
func requireAuth(w http.ResponseWriter, r *http.Request) (userID int64, ok bool) {
	userID, ok = middleware.UserIDFromContext(r.Context())
	if !ok {
		respond.Error(w, http.StatusUnauthorized, "MISSING_TOKEN", "authorization token required")
	}
	return userID, ok
}

// decodeJSON decodes r's JSON body into v, responding with 400 INVALID_BODY
// and returning false if the body is missing or malformed.
func decodeJSON(w http.ResponseWriter, r *http.Request, v any) bool {
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		respond.Error(w, http.StatusBadRequest, "INVALID_BODY", "invalid request body")
		return false
	}
	return true
}

// parsePathID parses the named path parameter as an int64, responding with
// 400 INVALID_ID and message if it's missing or malformed.
func parsePathID(w http.ResponseWriter, r *http.Request, param, message string) (id int64, ok bool) {
	id, err := strconv.ParseInt(r.PathValue(param), 10, 64)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "INVALID_ID", message)
		return 0, false
	}
	return id, true
}
