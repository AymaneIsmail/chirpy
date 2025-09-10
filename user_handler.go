package main

import (
	"encoding/json"
	"github.com/google/uuid"
	"net/http"
	"time"

	"github.com/AymaneIsmail/chirpy/internal/auth"
	"github.com/AymaneIsmail/chirpy/internal/database"
)

type User struct {
	ID           uuid.UUID `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Email        string    `json:"email"`
	Password     string    `json:"-"`
	Token        string    `json:"token"`
	RefreshToken string    `json:"refresh_token"`
	IsChirpyRed  bool      `json:"is_chirpy_red"`
}

func (cfg *apiConfig) createUserHandler(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	type response struct {
		User
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	hashedPassword, err := auth.HashPassword(params.Password)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, "Couldn't hash password", err)
		return
	}

	createUserParams := database.CreateUserParams{
		Email:          params.Email,
		HashedPassword: hashedPassword,
	}

	user, err := cfg.db.CreateUser(r.Context(), createUserParams)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, "Couldn't create user", err)
		return
	}

	IsChirpyRed := false
	if user.IsChirpyRed.Valid {
		IsChirpyRed = user.IsChirpyRed.Bool
	}

	jsonResponse(w, http.StatusCreated, response{
		User: User{
			ID:          user.ID,
			CreatedAt:   user.CreatedAt,
			UpdatedAt:   user.UpdatedAt,
			Email:       user.Email,
			IsChirpyRed: IsChirpyRed,
		},
	})
}

func (cfg *apiConfig) updateUserHandler(w http.ResponseWriter, r *http.Request) {
	type Params struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	type response struct {
		User
	}

	bearerToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		jsonError(w, http.StatusUnauthorized, "Couldn't find JWT", err)
		return
	}

	userID, err := auth.ValidateJWT(bearerToken, cfg.JWTSecret)
	if err != nil {
		jsonError(w, http.StatusUnauthorized, "invalid or expired token", err)
		return
	}

	var p Params
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		jsonError(w, http.StatusBadRequest, "couldn't decode request body", err)
		return
	}

	if p.Email == "" || p.Password == "" {
		jsonError(w, http.StatusBadRequest, "email and password are required", nil)
		return
	}

	hashedPassword, err := auth.HashPassword(p.Password)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, "Couldn't hash password", err)
		return
	}

	user, err := cfg.db.UpdateUserByID(r.Context(), database.UpdateUserByIDParams{
		ID:             userID,
		Email:          p.Email,
		HashedPassword: hashedPassword,
	})
	if err != nil {
		jsonError(w, http.StatusBadRequest, "couldn't update user", err)
		return
	}

	IsChirpyRed := false
	if user.IsChirpyRed.Valid {
		IsChirpyRed = user.IsChirpyRed.Bool
	}

	jsonResponse(w, http.StatusOK, response{
		User: User{
			ID:          user.ID,
			CreatedAt:   user.CreatedAt,
			UpdatedAt:   user.UpdatedAt,
			Email:       user.Email,
			IsChirpyRed: IsChirpyRed,
		},
	})
}
