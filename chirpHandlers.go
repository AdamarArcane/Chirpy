package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/adamararcane/chirpy/internal/database"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerCreateChirp(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body   string    `json:"body"`
		UserID uuid.UUID `json:"user_id"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		WriteErrorResponse(w, 500, "Something went wrong")
		return
	}

	if len(params.Body) > 140 {
		log.Printf("Chirp over 140 characters")
		WriteErrorResponse(w, 400, "Chirp is too long")
		return
	}

	cleanedChirp := cleanChirp(params.Body)

	newChirp, err := cfg.db.CreateChirp(r.Context(), database.CreateChirpParams{Body: cleanedChirp, UserID: params.UserID})
	if err != nil {
		log.Printf("Errror creating chirp: %s", err)
		WriteErrorResponse(w, 500, "Error creating chirp")
	}

	respBody := ChirpResp{
		ID:         newChirp.ID,
		Created_at: newChirp.CreatedAt,
		Updated_at: newChirp.UpdatedAt,
		Body:       newChirp.Body,
		User_id:    newChirp.UserID,
	}

	dat, err := json.Marshal(respBody)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		WriteErrorResponse(w, 500, "Something went wrong")
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(201)
	w.Write(dat)
}

func (cfg *apiConfig) handlerGetAllChrips(w http.ResponseWriter, r *http.Request) {
	allChirps, err := cfg.db.GetAllChirps(r.Context())
	if err != nil {
		log.Printf("Error getting all chirps: %s", err)
		WriteErrorResponse(w, 500, "Error getting all chirps")
	}

	var allChirpsJSON []ChirpResp
	for _, chirp := range allChirps {
		respItem := ChirpResp{
			ID:         chirp.ID,
			Created_at: chirp.CreatedAt,
			Updated_at: chirp.UpdatedAt,
			Body:       chirp.Body,
			User_id:    chirp.UserID,
		}
		allChirpsJSON = append(allChirpsJSON, respItem)
	}

	dat, err := json.Marshal(allChirpsJSON)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		WriteErrorResponse(w, 500, "Something went wrong")
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(dat)
}
