package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
	"database/sql"

	"github.com/AymaneIsmail/chirpy/internal/database"
	"github.com/AymaneIsmail/chirpy/internal/auth"
	"github.com/google/uuid"
)

type Chirp struct {
	ID          uuid.UUID `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	UserID      uuid.UUID `json:"user_id"`
	CleanedBody string    `json:"body"`
}

func (cfg *apiConfig) createChirpHandler(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}

	// 1) Auth: Bearer + JWT
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

	// 2) Decode body
	var params parameters
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		jsonError(w, http.StatusBadRequest, "couldn't decode request body", err)
		return
	}

	// 3) Validation + nettoyage
	const maxChirpLength = 140
	if len(params.Body) > maxChirpLength {
		jsonError(w, http.StatusBadRequest, "chirp is too long", nil)
		return
	}

	words := strings.Split(params.Body, " ")
	cleaned := strings.Join(validateWords(words), " ")

	// 4) Création en DB avec l'user issu du JWT
	createParams := database.CreateChirpParams{
		Body:   cleaned,
		UserID: userID,
	}
	chirp, err := cfg.db.CreateChirp(r.Context(), createParams)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, fmt.Sprintf("cannot create chirp: %v", err), err)
		return
	}

	// 5) Réponse
	jsonResponse(w, http.StatusCreated, Chirp{
		ID:          chirp.ID,
		CreatedAt:   chirp.CreatedAt,
		UpdatedAt:   chirp.UpdatedAt,
		UserID:      chirp.UserID,
		CleanedBody: cleaned,
	})
}


func (cfg *apiConfig) GetChirps(w http.ResponseWriter, r *http.Request) {
	dbChirps, err := cfg.db.GetChirps(r.Context())
	if err != nil {
		jsonError(w, http.StatusInternalServerError, "Cannot get the chirps", err)
		return
	}

	chirps := []Chirp{}
	for _, dbChirp := range dbChirps {
		chirps = append(chirps, Chirp{
			ID:          dbChirp.ID,
			CreatedAt:   dbChirp.CreatedAt,
			UpdatedAt:   dbChirp.UpdatedAt,
			UserID:      dbChirp.UserID,
			CleanedBody: dbChirp.Body,
		})
	}
	jsonResponse(w, http.StatusOK, chirps)
}

func (cfg *apiConfig) GetChirp(w http.ResponseWriter, r *http.Request) {
	chirpIDStr := r.PathValue("chirpID")
	if chirpIDStr == "" {
		jsonError(w, http.StatusBadRequest, "No Chirp ID provided", nil)
		return
	}

	chirpID, err := uuid.Parse(chirpIDStr)
	if err != nil {
		jsonError(w, http.StatusBadRequest, "invalid chirp ID (must be UUID)", err)
		return
	}
	rawChirp, err := cfg.db.GetChirp(r.Context(), chirpID)
	if err != nil {
		jsonError(w, http.StatusNotFound, fmt.Sprintf("No Chirp found for id %s", chirpID), err)
		return
	}

	chirp := Chirp{
		ID:          rawChirp.ID,
		CreatedAt:   rawChirp.CreatedAt,
		UpdatedAt:   rawChirp.UpdatedAt,
		UserID:      rawChirp.UserID,
		CleanedBody: rawChirp.Body,
	}

	jsonResponse(w, http.StatusOK, chirp)
}

func (cfg *apiConfig) deleteChirpHandler(w http.ResponseWriter, r *http.Request) {
	bearerToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		jsonError(w, http.StatusUnauthorized, "Couldn't parse Bearer token", err)
		return
	}
	userID, err := auth.ValidateJWT(bearerToken, cfg.JWTSecret)
	if err != nil {
		jsonError(w, http.StatusUnauthorized, "invalid or expired token", err)
		return
	}

	chirpIDStr := r.PathValue("chirpID")
	if chirpIDStr == "" {
		jsonError(w, http.StatusBadRequest, "no chirp ID provided", nil)
		return
	}
	chirpID, err := uuid.Parse(chirpIDStr)
	if err != nil {
		jsonError(w, http.StatusBadRequest, "invalid chirp ID (must be UUID)", err)
		return
	}

	chirp, err := cfg.db.GetChirp(r.Context(), chirpID)
	if err != nil {
		if err == sql.ErrNoRows {
			jsonError(w, http.StatusNotFound, "chirp not found", err)
			return
		}
		jsonError(w, http.StatusInternalServerError, "failed to load chirp", err)
		return
	}

	if chirp.UserID != userID {
		jsonError(w, http.StatusForbidden, "not the author of this chirp", nil)
		return
	}

	if err := cfg.db.DeleteChirp(r.Context(), chirpID); err != nil {
		jsonError(w, http.StatusInternalServerError, "failed to delete chirp", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func validateWords(words []string) []string {

	blackList := map[string]bool{
		"kerfuffle": true,
		"sharbert":  true,
		"fornax":    true,
	}
	output := []string{}

	for _, word := range words {
		if !blackList[strings.ToLower(word)] {
			output = append(output, word)
		} else {
			output = append(output, "****")
		}
	}

	return output
}
