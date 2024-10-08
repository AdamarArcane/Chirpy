package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/adamararcane/chirpy/internal/auth"
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

	userJWT, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Printf("User does not have correct prefix in auth header: %s", err)
		WriteErrorResponse(w, 401, "Authorization header does not have correct prefix")
		return
	}

	userUUID, err := auth.ValidateJWT(userJWT, cfg.JWT_SECRET)
	if err != nil {
		log.Printf("Error validating user token: %s", err)
		WriteErrorResponse(w, 401, "JWT token failed to validate")
		return
	}

	if len(params.Body) > 140 {
		log.Printf("Chirp over 140 characters")
		WriteErrorResponse(w, 400, "Chirp is too long")
		return
	}

	cleanedChirp := cleanChirp(params.Body)

	newChirp, err := cfg.db.CreateChirp(r.Context(), database.CreateChirpParams{Body: cleanedChirp, UserID: userUUID})
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
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(201)
	w.Write(dat)
}

func (cfg *apiConfig) handlerGetChirps(w http.ResponseWriter, r *http.Request) {
	userIDstr := r.URL.Query().Get("author_id")
	sortQuery := r.URL.Query().Get("sort")

	if userIDstr != "" {
		userID, err := uuid.Parse(userIDstr)
		if err != nil {
			log.Printf("Error parsing userID string: %s", err)
			WriteErrorResponse(w, 400, "Invalid UUID")
			return
		}
		var userChirps []database.Chirp
		if sortQuery == "desc" {
			userChirps, err = cfg.db.GetChirpsFromUserIDDESC(r.Context(), userID)
			if err != nil {
				log.Printf("User does not exist: %s", err)
				WriteErrorResponse(w, 404, "User not found or user has no chirps")
				return
			}
		} else {
			userChirps, err = cfg.db.GetChirpsFromUserIDASC(r.Context(), userID)
			if err != nil {
				log.Printf("User does not exist: %s", err)
				WriteErrorResponse(w, 404, "User not found or user has no chirps")
				return
			}
		}

		var allChirpsJSON []ChirpResp
		for _, chirp := range userChirps {
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
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write(dat)

	} else {
		var err error
		var allChirps []database.Chirp
		if sortQuery == "desc" {
			allChirps, err = cfg.db.GetAllChirpsDESC(r.Context())
			if err != nil {
				log.Printf("User does not exist: %s", err)
				WriteErrorResponse(w, 404, "User not found or user has no chirps")
				return
			}
		} else {
			allChirps, err = cfg.db.GetAllChirpsASC(r.Context())
			if err != nil {
				log.Printf("User does not exist: %s", err)
				WriteErrorResponse(w, 404, "User not found or user has no chirps")
				return
			}
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
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write(dat)
	}
}

func (cfg *apiConfig) handlerGetChirpByID(w http.ResponseWriter, r *http.Request) {
	chirpIDstring := r.PathValue("chirpID")
	fmt.Println(chirpIDstring)
	chirpID, err := uuid.Parse(chirpIDstring)
	if err != nil {
		log.Printf("Error parsing UUID: %s", err)
		WriteErrorResponse(w, 400, "Chirp ID is not valid UUID")
		return
	}

	chirp, err := cfg.db.GetChirpByID(r.Context(), chirpID)
	if err != nil {
		log.Printf("Chirp not found: %s", err)
		WriteErrorResponse(w, 404, "Chirp not found")
		return
	}

	chirpResp := ChirpResp{
		ID:         chirp.ID,
		Created_at: chirp.CreatedAt,
		Updated_at: chirp.UpdatedAt,
		Body:       chirp.Body,
		User_id:    chirp.UserID,
	}

	dat, err := json.Marshal(chirpResp)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		WriteErrorResponse(w, 500, "Something went wrong")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(dat)
}

func (cfg *apiConfig) handlerDeleteChirpByID(w http.ResponseWriter, r *http.Request) {
	accessToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Printf("User header is malformed: %s", err)
		WriteErrorResponse(w, 401, "Bearer <token> authentication token not found")
		return
	}

	userID, err := auth.ValidateJWT(accessToken, cfg.JWT_SECRET)
	if err != nil {
		log.Printf("User token is invalid or expried")
		WriteErrorResponse(w, 403, "Token is malformed or missing")
		return
	}

	chirpIDstring := r.PathValue("chirpID")
	fmt.Println(chirpIDstring)
	chirpID, err := uuid.Parse(chirpIDstring)
	if err != nil {
		log.Printf("Error parsing UUID: %s", err)
		WriteErrorResponse(w, 400, "Chirp ID is not valid UUID")
		return
	}

	chirp, err := cfg.db.GetChirpByID(r.Context(), chirpID)
	if err != nil {
		log.Printf("Chirp not found: %s", err)
		WriteErrorResponse(w, 404, "Chirp not found")
		return
	}

	if chirp.UserID != userID {
		log.Printf("Chirp UserID and UserID do not match: %s", err)
		WriteErrorResponse(w, 403, "You may not delete chirps that are not yours")
		return
	}

	err = cfg.db.DeleteChirpByID(r.Context(), chirpID)
	if err != nil {
		log.Printf("Error deleteing chirp from DB: %s", err)
		WriteErrorResponse(w, 500, "Error deleting chirp from database")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(204)
}
