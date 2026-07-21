package controller

import (
	"bytes"
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"slices"

	"github.com/amanagement24/journal-go/internal/data"
	"github.com/amanagement24/journal-go/internal/server"
	"github.com/google/uuid"
)

type LoginHandler struct {
	Config *data.ConfigData
	DB     *sql.DB
}

func NewLoginHandler(config *data.ConfigData, db *sql.DB) *LoginHandler {
	return &LoginHandler{
		Config: config,
		DB:     db,
	}
}

func (h *LoginHandler) process(ctx context.Context, login *data.Login) (*data.LoginResponse, string, error) {
	// login against authenticator first
	auth, err := h.processAuth(h.Config.Auth, login)
	if err != nil {
		return nil, "", err
	}

	// check the rights
	expectedAccess := h.Config.Access
	rights := auth.Rights

	if !slices.Contains(rights, expectedAccess) {
		return nil, "", errors.New("sorry, login successful but you do not have access to this application")
	}

	servr := server.New(h.DB, h.Config)

	// check the user or create if not present
	providedId := auth.Id
	userID, err := servr.GetUserByProvidedId(ctx, providedId)
	if err != nil {
		return nil, "", err
	}

	if userID == "" {
		userID, err = servr.CreateUserByProvidedID(ctx, providedId, auth.Login, auth.Name)
		if err != nil {
			return nil, "", err
		}
	}

	// create the session entry and save the token
	token := h.createRandomToken()

	err = servr.CreateSessionForUser(ctx, userID, token)
	if err != nil {
		return nil, "", err
	}

	res, err := servr.GetLoginResponse(ctx, userID)
	res.User.ProvidedId = providedId

	return res, token, err
}

func (h *LoginHandler) processAuth(url string, login *data.Login) (*data.Authentication, error) {
	payload, err := json.Marshal(login)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal login: %w", err)
	}

	resp, err := http.Post(url, "application/json", bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("authentication failed: %s", string(body))
	}

	var auth data.Authentication
	if err := json.Unmarshal(body, &auth); err != nil {
		return nil, fmt.Errorf("failed to parse authentication response: %w", err)
	}

	return &auth, nil
}

func (h *LoginHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	var login data.Login
	var response *data.LoginResponse
	token := ""

	err := parseJsonRequest(r, &login)
	if err == nil {
		response, token, err = h.process(ctx, &login)
	}

	if err == nil && token != "" {
		http.SetCookie(w, &http.Cookie{
			Name:     cookieName,
			Value:    token,
			Path:     "/",
			HttpOnly: true,
		})
	}

	writeJsonResponse(w, response, err)
}

func (h *LoginHandler) createRandomToken() string {
	u7 := uuid.Must(uuid.NewV7())
	randomBytes := make([]byte, 14)
	_, _ = rand.Read(randomBytes)
	return fmt.Sprintf("%s%s", u7.String(), hex.EncodeToString(randomBytes))[:64]
}
