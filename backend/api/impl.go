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

	rows, err := s.dbpool.Query(context.Background(), "SELECT id, name, email, mobile, phone, is_archived FROM landlords")
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
		) RETURNING id,
			name,
			email,
			mobile,
			phone,
			is_archived`,
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
		RETURNING id,
			name,
			email,
			mobile,
			phone,
			is_archived`,
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
	properties := []Property{}

	rows, err := s.dbpool.Query(
		context.Background(),
		`SELECT 
			id,
			address_line_1,
			address_line_2,
			suburb,
			state,
			postcode,
			landlord_id,
			management_fee,
			is_archived
		FROM properties`,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var property Property
		err := rows.Scan(
			&property.Id,
			&property.AddressLine1,
			&property.AddressLine2,
			&property.Suburb,
			&property.State,
			&property.Postcode,
			&property.LandlordId,
			&property.ManagementFee,
			&property.IsArchived,
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		properties = append(properties, property)
	}

	if err := rows.Err(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(PropertyList{Items: properties})
}

func (s *Server) PropertiesCreate(w http.ResponseWriter, r *http.Request) {
	var payload Property
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	rows, err := s.dbpool.Query(
		context.Background(),
		`INSERT INTO properties (
			address_line_1,
			address_line_2,
			suburb,
			state,	
			postcode,
			landlord_id,
			management_fee
		) VALUES (
			$1,
			$2,
			$3,
			$4,
			$5,
			$6,
			$7
		) RETURNING 
			id, 
			address_line_1, 
			address_line_2, 
			suburb, 
			state, 
			postcode, 
			landlord_id, 
			management_fee, 
			is_archived`,
		payload.AddressLine1,
		payload.AddressLine2,
		payload.Suburb,
		payload.State,
		payload.Postcode,
		payload.LandlordId,
		payload.ManagementFee,
	)

	var createdProperty Property

	if rows.Next() {
		err = rows.Scan(&createdProperty.Id, &createdProperty.AddressLine1, &createdProperty.AddressLine2, &createdProperty.Suburb, &createdProperty.State, &createdProperty.Postcode, &createdProperty.LandlordId, &createdProperty.ManagementFee, &createdProperty.IsArchived)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdProperty)
}

func (s *Server) PropertiesArchive(w http.ResponseWriter, r *http.Request, id string) {
	var archivedProperty Property

	rows, err := s.dbpool.Exec(context.Background(), 
		`UPDATE properties 
			SET is_archived = NOW() 
			WHERE id = $1`,
		id,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if rows.RowsAffected() == 0 {
		http.Error(w, "Property not found", http.StatusNotFound)
		return
	}

	err = s.dbpool.QueryRow(context.Background(), 
		`SELECT id,
			address_line_1,
			address_line_2,
			suburb,
			state,
			postcode,
			landlord_id,
			management_fee,
			is_archived
		FROM properties 
		WHERE id = $1`, 
		id,
	).Scan(
		&archivedProperty.Id, 
		&archivedProperty.AddressLine1, 
		&archivedProperty.AddressLine2, 
		&archivedProperty.Suburb, 
		&archivedProperty.State, 
		&archivedProperty.Postcode, 
		&archivedProperty.LandlordId, 
		&archivedProperty.ManagementFee, 
		&archivedProperty.IsArchived,
	)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(archivedProperty)
}

func (s *Server) PropertiesGet(w http.ResponseWriter, r *http.Request, id string) {
	var property Property
	err := s.dbpool.QueryRow(context.Background(), 
		`SELECT id,
			address_line_1,
			address_line_2,
			suburb,
			state,
			postcode,
			landlord_id,
			management_fee,
			is_archived
		FROM properties 
		WHERE id = $1`, 
		id,
	).Scan(
		&property.Id, 
		&property.AddressLine1, 
		&property.AddressLine2, 
		&property.Suburb, 
		&property.State, 
		&property.Postcode, 
		&property.LandlordId, 
		&property.ManagementFee, 
		&property.IsArchived,
	)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(property)
}	

func (s *Server) PropertiesUpdate(w http.ResponseWriter, r *http.Request, id string) {
	var payload Property
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	rows, err := s.dbpool.Query(
		context.Background(),
		`UPDATE properties SET
			address_line_1 = $1,
			address_line_2 = $2,
			suburb = $3,
			state = $4,
			postcode = $5,
			landlord_id = $6,
			management_fee = $7
		WHERE id = $8
		RETURNING 
			id, 
			address_line_1, 
			address_line_2, 
			suburb, 
			state, 
			postcode, 
			landlord_id, 
			management_fee, 
			is_archived`,
		payload.AddressLine1,
		payload.AddressLine2,
		payload.Suburb,
		payload.State,
		payload.Postcode,
		payload.LandlordId,
		payload.ManagementFee,
		id,
	)

	var updatedProperty Property

	if rows.Next() {
		err = rows.Scan(&updatedProperty.Id, &updatedProperty.AddressLine1, &updatedProperty.AddressLine2, &updatedProperty.Suburb, &updatedProperty.State, &updatedProperty.Postcode, &updatedProperty.LandlordId, &updatedProperty.ManagementFee, &updatedProperty.IsArchived)
	}
	
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updatedProperty)
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