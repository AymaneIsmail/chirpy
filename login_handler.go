package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/AymaneIsmail/chirpy/internal/auth"
	"github.com/AymaneIsmail/chirpy/internal/database"
)

func (cfg *apiConfig) loginHandler(w http.ResponseWriter, r *http.Request) {
	type LoginDTO struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	type response struct {
		User
	}

	decoder := json.NewDecoder(r.Body)
	var loginDTO LoginDTO
	if err := decoder.Decode(&loginDTO); err != nil {
		jsonError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	user, err := cfg.db.GetUserByEmail(r.Context(), loginDTO.Email)
	if err != nil {
		jsonError(w, http.StatusUnauthorized, "Incorrect email or password", err)
		return
	}

	if err := auth.CheckHashedPassword(user.HashedPassword, loginDTO.Password); err != nil {
		jsonError(w, http.StatusUnauthorized, "Incorrect email or password", err)
		return
	}

	tokenStr, err := auth.MakeJWT(user.ID, cfg.JWTSecret, time.Hour)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, "Couldn't generate JWT", err)
		return
	}

	refreshToken, _ := auth.MakeRefreshToken()

	createRefreshToken := database.CreateRefreshTokenParams{
		Token:     refreshToken,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(60 * 24 * time.Hour),
		RevokedAt: sql.NullTime{}, // NULL (Valid=false)
	}

	_, err = cfg.db.CreateRefreshToken(r.Context(), createRefreshToken)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, "Couldn't insert RefreshToken", err)
		return
	}

	IsChirpyRed := false
	if user.IsChirpyRed.Valid {
		IsChirpyRed = user.IsChirpyRed.Bool
	}

	jsonResponse(w, http.StatusOK, response{
		User: User{
			ID:           user.ID,
			CreatedAt:    user.CreatedAt,
			UpdatedAt:    user.UpdatedAt,
			Email:        user.Email,
			Token:        tokenStr,
			RefreshToken: refreshToken,
			IsChirpyRed:  IsChirpyRed,
		},
	})
}
