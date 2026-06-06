package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"net/http"

	"github.com/beetrack/backend/internal/service"
	"github.com/beetrack/backend/pkg/respond"
)

// AuthHandler handles authentication-related HTTP requests.
type AuthHandler struct {
	auth *service.AuthService
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(auth *service.AuthService) *AuthHandler {
	return &AuthHandler{auth: auth}
}

// --- HTML page helpers ---

type staticPage struct {
	icon    string
	title   string
	heading string
	body    string
}

var verifySuccessPages = map[string]staticPage{
	"en": {"✅", "BeeTrack — Email Verified", "Email Verified", "Your account is now active. You can log in to BeeTrack."},
	"pl": {"✅", "BeeTrack — E-mail zweryfikowany", "Adres e-mail zweryfikowany", "Twoje konto jest aktywne. Możesz się zalogować do BeeTrack."},
}

var verifyFailPages = map[string]staticPage{
	"en": {"❌", "BeeTrack — Verification Failed", "Verification Failed", "This link has expired or has already been used. Please request a new verification email from the app."},
	"pl": {"❌", "BeeTrack — Weryfikacja nieudana", "Weryfikacja nieudana", "Link wygasł lub był już użyty. Poproś o nowy e-mail weryfikacyjny w aplikacji."},
}

var resetSuccessPages = map[string]staticPage{
	"en": {"✅", "BeeTrack — Password Changed", "Password Changed", "Your password has been updated. You can now log in to BeeTrack."},
	"pl": {"✅", "BeeTrack — Hasło zmienione", "Hasło zmienione", "Twoje hasło zostało zaktualizowane. Możesz teraz zalogować się do BeeTrack."},
}

var resetExpiredPages = map[string]staticPage{
	"en": {"❌", "BeeTrack — Link Expired", "Link Expired", "This link has expired or has already been used. Please request a new password reset from the app."},
	"pl": {"❌", "BeeTrack — Link wygasł", "Link wygasł", "Link wygasł lub był już użyty. Poproś o nowy reset hasła w aplikacji."},
}

func resolvePage(pages map[string]staticPage, lang string) staticPage {
	if p, ok := pages[lang]; ok {
		return p
	}
	return pages["en"]
}

func writeStaticPage(w http.ResponseWriter, status int, p staticPage) {
	html := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width,initial-scale=1">
  <title>%s</title>
  <style>
    *{box-sizing:border-box;margin:0;padding:0}
    body{font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif;background:#FFFBF2;min-height:100vh;display:flex;align-items:center;justify-content:center}
    .card{background:#fff;border-radius:16px;padding:48px 40px;max-width:420px;width:100%%;text-align:center;box-shadow:0 2px 16px rgba(0,0,0,.08)}
    .icon{font-size:56px;margin-bottom:20px}
    h1{font-size:24px;color:#1a1a1a;margin-bottom:12px}
    p{color:#6F6961;line-height:1.6}
  </style>
</head>
<body>
  <div class="card">
    <div class="icon">%s</div>
    <h1>%s</h1>
    <p>%s</p>
  </div>
</body>
</html>`, p.title, p.icon, p.heading, p.body)
	w.Header().Set("Content-Type", "text/html; charset=UTF-8")
	w.WriteHeader(status)
	w.Write([]byte(html)) //nolint:errcheck
}

var resetFormTmpl = template.Must(template.New("resetForm").Parse(`<!DOCTYPE html>
<html>
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width,initial-scale=1">
  <title>{{.Title}}</title>
  <style>
    *{box-sizing:border-box;margin:0;padding:0}
    body{font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif;background:#FFFBF2;min-height:100vh;display:flex;align-items:center;justify-content:center}
    .card{background:#fff;border-radius:16px;padding:48px 40px;max-width:420px;width:100%;text-align:center;box-shadow:0 2px 16px rgba(0,0,0,.08)}
    h1{font-size:24px;color:#1a1a1a;margin-bottom:24px}
    input[type=password]{width:100%;padding:12px 16px;border:1px solid #e0dbd4;border-radius:8px;font-size:16px;margin-bottom:8px;outline:none}
    input[type=password]:focus{border-color:#FBBF24}
    .error{color:#dc2626;font-size:14px;margin-bottom:12px}
    button{width:100%;padding:12px;background:#FBBF24;color:#fff;border:none;border-radius:8px;font-size:16px;font-weight:600;cursor:pointer;margin-top:8px}
    button:hover{background:#f59e0b}
  </style>
</head>
<body>
  <div class="card">
    <h1>{{.Heading}}</h1>
    <form method="POST" action="/api/v1/auth/reset-password-form">
      <input type="hidden" name="token" value="{{.Token}}">
      <input type="hidden" name="lang" value="{{.Lang}}">
      {{if .Error}}<p class="error">{{.Error}}</p>{{end}}
      <input type="password" name="password" placeholder="{{.PasswordLabel}}" required minlength="8" autofocus>
      <button type="submit">{{.SubmitLabel}}</button>
    </form>
  </div>
</body>
</html>`))

type resetFormData struct {
	Error         string
	Heading       string
	Lang          string
	PasswordLabel string
	SubmitLabel   string
	Title         string
	Token         string
}

var resetFormLabels = map[string]resetFormData{
	"en": {Title: "BeeTrack — Reset Password", Heading: "Reset Password", PasswordLabel: "New password", SubmitLabel: "Save password"},
	"pl": {Title: "BeeTrack — Resetowanie hasła", Heading: "Resetowanie hasła", PasswordLabel: "Nowe hasło", SubmitLabel: "Zapisz hasło"},
}

var weakPasswordMsg = map[string]string{
	"en": "Password must be at least 8 characters",
	"pl": "Hasło musi mieć co najmniej 8 znaków",
}

// --- Handlers ---

// ForgotPassword handles POST /api/v1/auth/forgot-password — initiates a password reset
// by sending a reset link to the given email. Always returns 204 to avoid email enumeration.
func (h *AuthHandler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email string `json:"email"`
		Lang  string `json:"lang"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "INVALID_BODY", "invalid request body")
		return
	}

	if err := h.auth.ForgotPassword(r.Context(), req.Email, req.Lang); err != nil {
		respond.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Login handles POST /api/v1/auth/login — authenticates a user and returns a token pair.
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "INVALID_BODY", "invalid request body")
		return
	}

	accessToken, refreshToken, name, err := h.auth.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidPassword):
			respond.Error(w, http.StatusUnauthorized, "INVALID_CREDENTIALS", err.Error())
		case errors.Is(err, service.ErrEmailNotVerified):
			respond.Error(w, http.StatusForbidden, "EMAIL_NOT_VERIFIED", err.Error())
		default:
			respond.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		}
		return
	}

	respond.JSON(w, http.StatusOK, map[string]string{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"name":          name,
	})
}

// Logout handles POST /api/v1/auth/logout — revokes the refresh token.
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "INVALID_BODY", "invalid request body")
		return
	}

	if err := h.auth.Logout(r.Context(), req.RefreshToken); err != nil {
		respond.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Refresh handles POST /api/v1/auth/refresh — exchanges a refresh token for a new token pair.
func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "INVALID_BODY", "invalid request body")
		return
	}

	accessToken, refreshToken, err := h.auth.Refresh(r.Context(), req.RefreshToken)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidRefreshToken):
			respond.Error(w, http.StatusUnauthorized, "INVALID_REFRESH_TOKEN", err.Error())
		case errors.Is(err, service.ErrTokenExpired):
			respond.Error(w, http.StatusUnauthorized, "TOKEN_EXPIRED", err.Error())
		default:
			respond.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		}
		return
	}

	respond.JSON(w, http.StatusOK, map[string]string{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}

// Register handles POST /api/v1/auth/register — creates a new user account and sends
// a verification email in the requested language.
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Lang     string `json:"lang"`
		Name     string `json:"name"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "INVALID_BODY", "invalid request body")
		return
	}

	user, err := h.auth.Register(r.Context(), req.Email, req.Name, req.Password, req.Lang)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrEmailTaken):
			respond.Error(w, http.StatusConflict, "EMAIL_TAKEN", err.Error())
		case errors.Is(err, service.ErrInvalidEmail):
			respond.Error(w, http.StatusBadRequest, "INVALID_EMAIL", err.Error())
		case errors.Is(err, service.ErrWeakPassword):
			respond.Error(w, http.StatusBadRequest, "WEAK_PASSWORD", err.Error())
		default:
			respond.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		}
		return
	}

	respond.JSON(w, http.StatusCreated, map[string]any{
		"id":         user.ID,
		"email":      user.Email,
		"name":       user.Name,
		"created_at": user.CreatedAt,
	})
}

// ResendVerification handles POST /api/v1/auth/resend-verification — sends a new verification
// email. Always returns 204 to avoid email enumeration.
func (h *AuthHandler) ResendVerification(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email string `json:"email"`
		Lang  string `json:"lang"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "INVALID_BODY", "invalid request body")
		return
	}

	if err := h.auth.ResendVerification(r.Context(), req.Email, req.Lang); err != nil {
		respond.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ResetPassword handles POST /api/v1/auth/reset-password — validates the reset token and
// updates the user's password. Used by API clients (mobile).
func (h *AuthHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Password string `json:"password"`
		Token    string `json:"token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "INVALID_BODY", "invalid request body")
		return
	}

	if err := h.auth.ResetPassword(r.Context(), req.Token, req.Password); err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidResetToken):
			respond.Error(w, http.StatusBadRequest, "INVALID_RESET_TOKEN", err.Error())
		case errors.Is(err, service.ErrWeakPassword):
			respond.Error(w, http.StatusBadRequest, "WEAK_PASSWORD", err.Error())
		default:
			respond.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ResetPasswordForm handles GET /api/v1/auth/reset-password-form — serves the HTML
// password reset form, and POST — processes the form submission.
func (h *AuthHandler) ResetPasswordForm(w http.ResponseWriter, r *http.Request) {
	lang := r.URL.Query().Get("lang")
	if lang == "" {
		lang = "en"
	}

	if r.Method == http.MethodGet {
		token := r.URL.Query().Get("token")
		if token == "" {
			writeStaticPage(w, http.StatusBadRequest, resolvePage(resetExpiredPages, lang))
			return
		}
		h.renderResetForm(w, token, lang, "")
		return
	}

	// POST
	if err := r.ParseForm(); err != nil {
		respond.Error(w, http.StatusBadRequest, "INVALID_BODY", "invalid form")
		return
	}
	rawToken := r.FormValue("token")
	password := r.FormValue("password")
	lang = r.FormValue("lang")
	if lang == "" {
		lang = "en"
	}

	if err := h.auth.ResetPassword(r.Context(), rawToken, password); err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidResetToken):
			writeStaticPage(w, http.StatusBadRequest, resolvePage(resetExpiredPages, lang))
		case errors.Is(err, service.ErrWeakPassword):
			msg := weakPasswordMsg[lang]
			if msg == "" {
				msg = weakPasswordMsg["en"]
			}
			h.renderResetForm(w, rawToken, lang, msg)
		default:
			writeStaticPage(w, http.StatusInternalServerError, resolvePage(resetExpiredPages, lang))
		}
		return
	}

	writeStaticPage(w, http.StatusOK, resolvePage(resetSuccessPages, lang))
}

// VerifyEmail handles GET /api/v1/auth/verify-email — validates the verification token,
// marks the account as verified, and returns a localized HTML confirmation page.
func (h *AuthHandler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	lang := r.URL.Query().Get("lang")
	if lang == "" {
		lang = "en"
	}
	rawToken := r.URL.Query().Get("token")

	if rawToken == "" {
		writeStaticPage(w, http.StatusBadRequest, resolvePage(verifyFailPages, lang))
		return
	}

	if err := h.auth.VerifyEmail(r.Context(), rawToken); err != nil {
		writeStaticPage(w, http.StatusBadRequest, resolvePage(verifyFailPages, lang))
		return
	}

	writeStaticPage(w, http.StatusOK, resolvePage(verifySuccessPages, lang))
}

func (h *AuthHandler) renderResetForm(w http.ResponseWriter, token, lang, errMsg string) {
	labels, ok := resetFormLabels[lang]
	if !ok {
		labels = resetFormLabels["en"]
	}
	data := resetFormData{
		Error:         errMsg,
		Heading:       labels.Heading,
		Lang:          lang,
		PasswordLabel: labels.PasswordLabel,
		SubmitLabel:   labels.SubmitLabel,
		Title:         labels.Title,
		Token:         token,
	}
	w.Header().Set("Content-Type", "text/html; charset=UTF-8")
	resetFormTmpl.Execute(w, data) //nolint:errcheck
}
