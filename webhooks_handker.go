package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/AymaneIsmail/chirpy/internal/auth"
	"github.com/google/uuid"
)

type Event struct {
	Event string `json:"event"`
	Data  struct {
		UserID uuid.UUID `json:"user_id"`
	} `json:"data"`
}

func (cfg *apiConfig) webhooks(w http.ResponseWriter, r *http.Request) {

	apiKey, err := auth.GetAPIKey(r.Header)
	if err != nil {
		jsonError(w, http.StatusUnauthorized, "", err)
		return
	}

	if apiKey != cfg.PolkaKey {
		jsonError(w, http.StatusUnauthorized, "", err)
		return
	}

	var ev struct {
		Event string `json:"event"`
		Data  struct {
			UserID uuid.UUID `json:"user_id"`
		} `json:"data"`
	}

	if err := json.NewDecoder(r.Body).Decode(&ev); err != nil {
		jsonError(w, http.StatusBadRequest, "couldn't decode request body", err)
		return
	}

	if ev.Event != "user.upgraded" {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	fmt.Printf("UserID: %s", ev.Data.UserID)

	if _, err := cfg.db.UpgradeToChirpyRed(r.Context(), ev.Data.UserID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		jsonError(w, http.StatusInternalServerError, "failed to upgrade user", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
