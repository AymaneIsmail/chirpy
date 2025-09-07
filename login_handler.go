package main

import (
	"encoding/json"
	"net/http"

	"github.com/AymaneIsmail/chirpy/internal/auth"
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

	err = auth.CheckHashedPassword(user.HashedPassword, loginDTO.Password)
	if err != nil {
		jsonError(w, http.StatusUnauthorized, "Incorrect email or password", err)
		return
	}

	jsonResponse(w, http.StatusOK, response{
		User: User{
			ID:        user.ID,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
			Email:     user.Email,
		},
	})
	return
}
