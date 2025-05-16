package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ensure that we've conformed to the `ServerInterface` with a compile-time check
var _ ServerInterface = (*Server)(nil)

type Server struct{
	dbpool *pgxpool.Pool
}

func NewServer(dbpool *pgxpool.Pool) *Server {
	return &Server{
		dbpool: dbpool,
	}
}

func (s *Server) LandlordsList(w http.ResponseWriter, r *http.Request) {
	landlords := []Landlord{}

	rows, err := s.dbpool.Query(context.Background(), "SELECT * FROM landlords")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var landlord Landlord
		err := rows.Scan(&landlord.Id, &landlord.Name, &landlord.Email, &landlord.Mobile, &landlord.Phone, &landlord.IsArchived)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}	
		landlords = append(landlords, landlord)
	}

	if err := rows.Err(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(LandlordList{Items: landlords})
}

func (s *Server) LandlordsCreate(w http.ResponseWriter, r *http.Request) {
	var payload Landlord
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	rows, err := s.dbpool.Query(
		context.Background(),
		`INSERT INTO landlords (
			name,
			email,
			mobile,
			phone
		) VALUES (
			$1,
			$2,
			$3,
			$4
		) RETURNING *`,
		payload.Name,
		payload.Email,
		payload.Mobile,
		payload.Phone,
	)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var createdLandlord Landlord

	if rows.Next() {
		err = rows.Scan(&createdLandlord.Id, &createdLandlord.Name, &createdLandlord.Email, &createdLandlord.Mobile, &createdLandlord.Phone, &createdLandlord.IsArchived)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdLandlord)
	
}

func (s *Server) LandlordsArchive(w http.ResponseWriter, r *http.Request, id string) {
	var archivedLandlord Landlord

	
	rows, err := s.dbpool.Exec(context.Background(), "UPDATE landlords SET is_archived = NOW() WHERE id = $1", id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if rows.RowsAffected() == 0 {
		http.Error(w, "Landlord not found", http.StatusNotFound)
		return
	}

	err = s.dbpool.QueryRow(context.Background(), "SELECT * FROM landlords WHERE id = $1", id).Scan(&archivedLandlord.Id, &archivedLandlord.Name, &archivedLandlord.Email, &archivedLandlord.Mobile, &archivedLandlord.Phone, &archivedLandlord.IsArchived)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(archivedLandlord)
}

func (s *Server) LandlordsGet(w http.ResponseWriter, r *http.Request, id string) {
	var landlord Landlord

	err := s.dbpool.QueryRow(context.Background(), "SELECT * FROM landlords WHERE id = $1", id).Scan(&landlord.Id, &landlord.Name, &landlord.Email, &landlord.Mobile, &landlord.Phone, &landlord.IsArchived)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(landlord)
}

func (s *Server) LandlordsUpdate(w http.ResponseWriter, r *http.Request, id string) {
	var payload Landlord
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	rows, err := s.dbpool.Query(
		context.Background(),
		`UPDATE landlords SET
			name = $1,
			email = $2,
			mobile = $3,
			phone = $4
		WHERE id = $5
		RETURNING *`,
		payload.Name,
		payload.Email,
		payload.Mobile,
		payload.Phone,
		id,
	)

	var updatedLandlord Landlord

	if rows.Next() {
		err = rows.Scan(&updatedLandlord.Id, &updatedLandlord.Name, &updatedLandlord.Email, &updatedLandlord.Mobile, &updatedLandlord.Phone, &updatedLandlord.IsArchived)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updatedLandlord)
}

func (s *Server) PropertiesList(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte("Not yet implemented"))
}

func (s *Server) PropertiesCreate(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte("Not yet implemented"))
}

func (s *Server) PropertiesArchive(w http.ResponseWriter, r *http.Request, id string) {
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte("Not yet implemented"))
}

func (s *Server) PropertiesGet(w http.ResponseWriter, r *http.Request, id string) {
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte("Not yet implemented"))
}	

func (s *Server) PropertiesUpdate(w http.ResponseWriter, r *http.Request, id string) {
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte("Not yet implemented"))
}

func (s *Server) TenantsList(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte("Not yet implemented"))
}

func (s *Server) TenantsCreate(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte("Not yet implemented"))
}	

func (s *Server) TenantsArchive(w http.ResponseWriter, r *http.Request, id string) {
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte("Not yet implemented"))
}

func (s *Server) TenantsGet(w http.ResponseWriter, r *http.Request, id string) {
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte("Not yet implemented"))
}

func (s *Server) TenantsUpdate(w http.ResponseWriter, r *http.Request, id string) {
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte("Not yet implemented"))
}	