package main

import "net/http"

func (cfg *apiConfig) resetUserHandler(w http.ResponseWriter, r *http.Request) {

	if cfg.Platform != "dev" {
		jsonError(w, http.StatusForbidden, "Reset is only allowed in dev environment.", nil)
		return
	}

	cfg.fileServerHits.Store(0)
	err := cfg.db.Reset(r.Context())
	if err != nil {
		jsonError(w, http.StatusInternalServerError, "Failed to reset the database: ", err)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hits reset to 0 and database reset to initial state."))
}
