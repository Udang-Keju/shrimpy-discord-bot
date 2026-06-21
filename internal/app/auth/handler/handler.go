package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/Udang-Keju/shrimpy-discord-bot/internal/api/crypto"
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/api/middleware"
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/app/auth/model"
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/pkg/apiutil"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

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
}

// NewAuthHandler constructs a new AuthHandler.
func NewAuthHandler(userRepo AuthUserRepo, jwtSecret []byte, tokenEncKey []byte) *AuthHandler {
	return &AuthHandler{
		userRepo:    userRepo,
		jwtSecret:   jwtSecret,
		tokenEncKey: tokenEncKey,
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

type callbackPayload struct {
	User         discordUser    `json:"discord_user"`
	Guilds       []discordGuild `json:"guilds"`
	AccessToken  string         `json:"access_token"`
	RefreshToken string         `json:"refresh_token"`
	ExpiresIn    int64          `json:"expires_in"`
}

// Callback handles the final step of the OAuth2 flow where the dashboard posts the token data.
func (h *AuthHandler) Callback(w http.ResponseWriter, r *http.Request) {
	var payload callbackPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		apiutil.WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid payload")
		return
	}

	userID, err := strconv.ParseInt(payload.User.ID, 10, 64)
	if err != nil {
		apiutil.WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid Discord user ID")
		return
	}

	req, _ := http.NewRequestWithContext(r.Context(), "GET", "https://discord.com/api/users/@me", nil)
	req.Header.Set("Authorization", "Bearer "+payload.AccessToken)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		apiutil.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Failed to verify Discord access token")
		return
	}
	defer resp.Body.Close()

	var managedGuilds []string
	for _, g := range payload.Guilds {
		perms, err := strconv.ParseInt(g.Permissions, 10, 64)
		if err == nil {
			if g.Owner || (perms&0x8) != 0 || (perms&0x20) != 0 {
				managedGuilds = append(managedGuilds, g.ID)
			}
		}
	}

	encAccess, err := crypto.Encrypt([]byte(payload.AccessToken), h.tokenEncKey)
	if err != nil {
		apiutil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Token encryption failed")
		return
	}

	var encRefresh []byte
	if payload.RefreshToken != "" {
		encRefresh, err = crypto.Encrypt([]byte(payload.RefreshToken), h.tokenEncKey)
		if err != nil {
			apiutil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Token encryption failed")
			return
		}
	}

	expiresAt := time.Now().Add(time.Duration(payload.ExpiresIn) * time.Second)

	dbUser := &model.User{
		UserID:                 userID,
		Username:               payload.User.Username,
		Discriminator:          payload.User.Discriminator,
		AvatarHash:             payload.User.Avatar,
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
			Subject:   payload.User.ID,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ID:        uuid.NewString(),
		},
		Username:      payload.User.Username,
		Avatar:        getStringValue(payload.User.Avatar),
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

	apiutil.WriteJSON(w, http.StatusOK, apiutil.JSONResponse{
		"token": tokenString,
		"user": apiutil.JSONResponse{
			"id":       payload.User.ID,
			"username": payload.User.Username,
			"avatar":   payload.User.Avatar,
		},
	})
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

func getStringValue(s *string) string {
	if s != nil {
		return *s
	}
	return ""
}
