package main

import (
	"github.com/AymaneIsmail/chirpy/internal/auth"
	"net/http"
	"time"
)

type RefreshTokenResponse struct {
	Token string `json:"token"`
}

func (cfg *apiConfig) refreshTokenHandler(w http.ResponseWriter, r *http.Request) {
	bearerToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		jsonError(w, http.StatusUnauthorized, "Couldn't parse Bearer token", err)
		return
	}

	rt, err := cfg.db.GetRefreshTokenByToken(r.Context(), bearerToken)
	if err != nil {
		jsonError(w, http.StatusUnauthorized, "No refresh token found", err)
		return
	}

	// 1) Révoqué ?
	if rt.RevokedAt.Valid {
		jsonError(w, http.StatusUnauthorized, "Token revoked", nil)
		return
	}

	// 2) Expiré ?
	if time.Now().After(rt.ExpiresAt) {
		jsonError(w, http.StatusUnauthorized, "Refresh token expired", nil)
		return
	}

	user, err := cfg.db.GetUserByRefreshToken(r.Context(), rt.Token)
	if err != nil {
		jsonError(w, http.StatusUnauthorized, "No user found", err)
		return
	}

	tokenStr, err := auth.MakeJWT(user.ID, cfg.JWTSecret, time.Hour)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, "Couldn't generate JWT", err)
		return
	}

	jsonResponse(w, http.StatusOK, RefreshTokenResponse{Token: tokenStr})
}

func (cfg *apiConfig) revokeRefreshTokenHandler(w http.ResponseWriter, r *http.Request) {
	bearerToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		jsonError(w, http.StatusUnauthorized, "Couldn't parse Bearer token", err)
		return
	}

	// Optionnel : vérifier l’existence du token pour renvoyer 401 s’il n’existe pas
	if _, err := cfg.db.GetRefreshTokenByToken(r.Context(), bearerToken); err != nil {
		jsonError(w, http.StatusUnauthorized, "Refresh token not found", err)
		return
	}

	// Révoquer (revoked_at = NOW(), updated_at = NOW())
	if err := cfg.db.RevokeRefreshToken(r.Context(), bearerToken); err != nil {
		jsonError(w, http.StatusInternalServerError, "Failed to revoke token", err)
		return
	}

	// 204 No Content
	w.WriteHeader(http.StatusNoContent)
}
