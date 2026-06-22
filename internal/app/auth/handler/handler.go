package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Udang-Keju/shrimpy-discord-bot/internal/api/middleware"
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/app/auth/model"
	settings_svc "github.com/Udang-Keju/shrimpy-discord-bot/internal/app/settings/service"
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/pkg/apiutil"
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/pkg/crypto"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// SettingsProvider is satisfied by the settings service and provides live OAuth2 credentials.
type SettingsProvider interface {
	GetDecryptedCredentials(ctx context.Context, id string) (token, clientID, clientSecret, redirectURI string, err error)
	List(ctx context.Context) ([]settings_svc.DiscordAppDTO, error)
}

// AuthUserRepo defines the user database operations needed by AuthHandler.
type AuthUserRepo interface {
	Upsert(ctx context.Context, u *model.User) error
	GetByID(ctx context.Context, userID int64) (*model.User, error)
}

// AuthHandler coordinates session token generation and OAuth2 validation callbacks.
type AuthHandler struct {
	userRepo    AuthUserRepo
	jwtSecret   []byte
	tokenEncKey []byte
	settings    SettingsProvider
}

// NewAuthHandler constructs a new AuthHandler.
func NewAuthHandler(userRepo AuthUserRepo, jwtSecret []byte, tokenEncKey []byte, settings SettingsProvider) *AuthHandler {
	return &AuthHandler{
		userRepo:    userRepo,
		jwtSecret:   jwtSecret,
		tokenEncKey: tokenEncKey,
		settings:    settings,
	}
}

type discordUser struct {
	ID            string  `json:"id"`
	Username      string  `json:"username"`
	Discriminator *string `json:"discriminator"`
	Avatar        *string `json:"avatar"`
}

type discordGuild struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Permissions string `json:"permissions"`
	Owner       bool   `json:"owner"`
}

type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

// Login redirects the user's browser to the Discord authorization page.
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	appID := r.URL.Query().Get("app_id")
	if appID == "" {
		apps, err := h.settings.List(r.Context())
		if err != nil || len(apps) == 0 {
			apiutil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "No Discord applications configured")
			return
		}
		appID = apps[0].ID
	}

	_, clientID, _, redirectURI, err := h.settings.GetDecryptedCredentials(r.Context(), appID)
	if err != nil {
		apiutil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve credentials for application")
		return
	}

	discordURL := fmt.Sprintf(
		"https://discord.com/api/oauth2/authorize?client_id=%s&redirect_uri=%s&response_type=code&scope=identify+guilds+guilds.members.read&state=%s",
		clientID,
		url.QueryEscape(redirectURI),
		appID,
	)

	http.Redirect(w, r, discordURL, http.StatusTemporaryRedirect)
}

// Callback handles the final step of the OAuth2 flow where Discord redirects the user with a code.
func (h *AuthHandler) Callback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	appID := r.URL.Query().Get("state")

	if code == "" || appID == "" {
		apiutil.WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "Missing authorization code or state")
		return
	}

	_, clientID, clientSecret, redirectURI, err := h.settings.GetDecryptedCredentials(r.Context(), appID)
	if err != nil {
		apiutil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve credentials for application")
		return
	}

	tokenData, err := h.exchangeCodeForToken(r.Context(), clientID, clientSecret, redirectURI, code)
	if err != nil {
		apiutil.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "OAuth2 token exchange failed: "+err.Error())
		return
	}

	discUser, err := h.fetchDiscordUser(r.Context(), tokenData.AccessToken)
	if err != nil {
		apiutil.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Failed to fetch user details from Discord")
		return
	}

	discGuilds, err := h.fetchDiscordGuilds(r.Context(), tokenData.AccessToken)
	if err != nil {
		apiutil.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Failed to fetch user guilds from Discord")
		return
	}

	userID, err := strconv.ParseInt(discUser.ID, 10, 64)
	if err != nil {
		apiutil.WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid Discord user ID")
		return
	}

	var managedGuilds []string
	for _, g := range discGuilds {
		perms, err := strconv.ParseInt(g.Permissions, 10, 64)
		if err == nil {
			if g.Owner || (perms&0x8) != 0 || (perms&0x20) != 0 {
				managedGuilds = append(managedGuilds, g.ID)
			}
		}
	}

	encAccess, err := crypto.Encrypt([]byte(tokenData.AccessToken), h.tokenEncKey)
	if err != nil {
		apiutil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Token encryption failed")
		return
	}

	var encRefresh []byte
	if tokenData.RefreshToken != "" {
		encRefresh, err = crypto.Encrypt([]byte(tokenData.RefreshToken), h.tokenEncKey)
		if err != nil {
			apiutil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Token encryption failed")
			return
		}
	}

	expiresAt := time.Now().Add(time.Duration(tokenData.ExpiresIn) * time.Second)

	dbUser := &model.User{
		UserID:                 userID,
		Username:               discUser.Username,
		Discriminator:          discUser.Discriminator,
		AvatarHash:             discUser.Avatar,
		DiscordAccessTokenEnc:  encAccess,
		DiscordRefreshTokenEnc: encRefresh,
		TokenExpiresAt:         &expiresAt,
		LastSeen:               time.Now(),
	}

	if err = h.userRepo.Upsert(r.Context(), dbUser); err != nil {
		apiutil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to upsert user record")
		return
	}

	claims := &middleware.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   discUser.ID,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ID:        uuid.NewString(),
		},
		Username:      discUser.Username,
		Avatar:        getStringValue(discUser.Avatar),
		ManagedGuilds: managedGuilds,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(h.jwtSecret)
	if err != nil {
		apiutil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to sign session token")
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    tokenString,
		Path:     "/",
		Expires:  time.Now().Add(7 * 24 * time.Hour),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})

	dashboardURL := os.Getenv("DASHBOARD_URL")
	if dashboardURL == "" {
		dashboardURL = "http://localhost:3000"
	}
	http.Redirect(w, r, dashboardURL+"/dashboard", http.StatusTemporaryRedirect)
}

// Me returns the active user session data decoded from JWT.
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID := apiutil.GetUserID(r.Context())
	managedGuilds := apiutil.GetManagedGuilds(r.Context())

	if userID == "" {
		apiutil.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Active session not found")
		return
	}

	idInt, _ := strconv.ParseInt(userID, 10, 64)
	u, err := h.userRepo.GetByID(r.Context(), idInt)
	if err != nil {
		apiutil.WriteError(w, http.StatusNotFound, "NOT_FOUND", "User not found")
		return
	}

	apiutil.WriteJSON(w, http.StatusOK, apiutil.JSONResponse{
		"id":             userID,
		"username":       u.Username,
		"avatar":         u.AvatarHash,
		"managed_guilds": managedGuilds,
	})
}

// Logout invalidates the HttpOnly session cookie.
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})
	apiutil.WriteJSON(w, http.StatusOK, apiutil.JSONResponse{"success": true})
}

func (h *AuthHandler) exchangeCodeForToken(ctx context.Context, clientID, clientSecret, redirectURI, code string) (*tokenResponse, error) {
	data := url.Values{}
	data.Set("client_id", clientID)
	data.Set("client_secret", clientSecret)
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", redirectURI)

	req, err := http.NewRequestWithContext(ctx, "POST", "https://discord.com/api/oauth2/token", strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errData map[string]interface{}
		_ = json.NewDecoder(resp.Body).Decode(&errData)
		return nil, fmt.Errorf("discord OAuth2 token exchange returned status %d: %v", resp.StatusCode, errData)
	}

	var tr tokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tr); err != nil {
		return nil, err
	}
	return &tr, nil
}

func (h *AuthHandler) fetchDiscordUser(ctx context.Context, accessToken string) (*discordUser, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://discord.com/api/users/@me", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch user profile, status: %d", resp.StatusCode)
	}

	var du discordUser
	if err := json.NewDecoder(resp.Body).Decode(&du); err != nil {
		return nil, err
	}
	return &du, nil
}

func (h *AuthHandler) fetchDiscordGuilds(ctx context.Context, accessToken string) ([]discordGuild, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://discord.com/api/users/@me/guilds", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch user guilds, status: %d", resp.StatusCode)
	}

	var dg []discordGuild
	if err := json.NewDecoder(resp.Body).Decode(&dg); err != nil {
		return nil, err
	}
	return dg, nil
}

func getStringValue(s *string) string {
	if s != nil {
		return *s
	}
	return ""
}
