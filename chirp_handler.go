package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/AymaneIsmail/chirpy/internal/database"
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
		Body   string    `json:"body"`
		UserID uuid.UUID `json:"user_id"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	const maxChirpLength = 140
	if len(params.Body) > maxChirpLength {
		jsonError(w, http.StatusBadRequest, "Chirp is too long", nil)
		return
	}

	words := strings.Split(params.Body, " ")

	cleanedWords := validateWords(words)
	s := strings.Join(cleanedWords, " ")

	lastUser, err := cfg.db.GetLastUser(r.Context())
	if err != nil {
		jsonError(w, http.StatusInternalServerError, fmt.Sprintf("Cannot get the last user created: %v\n", err), err)
		return
	}

	createChirpParams := database.CreateChirpParams{
		Body:   s,
		UserID: lastUser.ID,
	}

	chirp, err := cfg.db.CreateChirp(r.Context(), createChirpParams)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, fmt.Sprintf("Cannot create Chirp: %v\n", err), err)
		return
	}

	jsonResponse(w, http.StatusCreated, Chirp{
		ID:          chirp.ID,
		CreatedAt:   chirp.CreatedAt,
		UpdatedAt:   chirp.UpdatedAt,
		CleanedBody: s,
		UserID:      chirp.UserID,
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
