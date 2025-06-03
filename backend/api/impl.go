package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/pgtype"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

// ensure that we've conformed to the `ServerInterface` with a compile-time check
var _ ServerInterface = (*Server)(nil)

type Server struct {
	dbpool *pgxpool.Pool
	logger *slog.Logger
}

func NewServer(dbpool *pgxpool.Pool, logger *slog.Logger) *Server {
	return &Server{
		dbpool: dbpool,
		logger: logger,
	}
}

func (s *Server) LandlordsList(w http.ResponseWriter, r *http.Request, params LandlordsListParams) {
	landlords := []Landlord{}

	limit, page, offset := handlePaginationParams(params)

	conditions := map[string]interface{}{
		"name":          params.Name,
		"archived_only": params.ArchivedOnly,
	}

	whereClause, queryParams, paramCount := buildWhereClause(conditions)

	// Get total count
	countSQL := fmt.Sprintf(`
		SELECT COUNT(*) 
		FROM landlords
		%s`, whereClause)

	var total int
	err := s.dbpool.QueryRow(context.Background(), countSQL, queryParams...).Scan(&total)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get paginated results
	queryParams = append(queryParams, limit, offset)
	listSQL := fmt.Sprintf(`
		SELECT 
			id, 
			name, 
			email, 
			mobile, 
			phone, 
			address_line_1,
			address_line_2,
			suburb,
			postcode,
			state,
			country,
			is_archived,
			created_at,
			updated_at
		FROM landlords
		%s
		ORDER BY name
		LIMIT $%d 
		OFFSET $%d`, whereClause, paramCount, paramCount+1)

	rows, err := s.dbpool.Query(context.Background(), listSQL, queryParams...)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var landlord Landlord

		err := rows.Scan(
			&landlord.Id,
			&landlord.Name,
			&landlord.Email,
			&landlord.Mobile,
			&landlord.Phone,
			&landlord.AddressLine1,
			&landlord.AddressLine2,
			&landlord.Suburb,
			&landlord.Postcode,
			&landlord.State,
			&landlord.Country,
			&landlord.IsArchived,
			&landlord.CreatedAt,
			&landlord.UpdatedAt,
		)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		landlords = append(landlords, landlord)
	}

	if err := rows.Err(); err != nil {
		apiError := handleLandlordErrors(err)

		w.WriteHeader(int(apiError.Code))
		json.NewEncoder(w).Encode(apiError)
		return
	}

	resp := LandlordList{
		Items: landlords,
		Pagination: PaginatedMetadata{
			Total:       int32(total),
			Count:       int32(len(landlords)),
			PerPage:     int32(limit),
			CurrentPage: int32(page),
			TotalPages:  int32(math.Ceil(float64(total) / float64(limit))),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(resp)

	if err != nil {
		s.logger.Error("error encoding response", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Error{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error",
		})
		return
	}

	s.logger.Debug("Landlords List Response", "response", resp)
}

func (s *Server) LandlordsCreate(w http.ResponseWriter, r *http.Request) {
	var payload CreateLandlord
	err := json.NewDecoder(r.Body).Decode(&payload)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Error{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	id, err := uuid.NewV7()

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Error{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error",
		})
		return
	}

	sql := `
		INSERT INTO landlords (
			id,
			name,
			email,
			mobile,
			phone,
			address_line_1,
			address_line_2,
			suburb,
			postcode,
			state,
			country
		) VALUES (
			$1,
			$2,
			$3,
			$4,
			$5,
			$6,
			$7,
			$8,
			$9,
			$10,
			$11
		) RETURNING 
		 	id,
			name,
			email,
			mobile,
			phone,
			address_line_1,
			address_line_2,
			suburb,
			postcode,
			state,
			country,
			is_archived,
			created_at,
			updated_at
		`

	row := s.dbpool.QueryRow(
		context.Background(),
		sql,
		id.String(),
		payload.Name,
		payload.Email,
		payload.Mobile,
		payload.Phone,
		payload.AddressLine1,
		payload.AddressLine2,
		payload.Suburb,
		payload.Postcode,
		payload.State,
		payload.Country,
	)

	var createdLandlord Landlord

	err = row.Scan(
		&createdLandlord.Id,
		&createdLandlord.Name,
		&createdLandlord.Email,
		&createdLandlord.Mobile,
		&createdLandlord.Phone,
		&createdLandlord.AddressLine1,
		&createdLandlord.AddressLine2,
		&createdLandlord.Suburb,
		&createdLandlord.Postcode,
		&createdLandlord.State,
		&createdLandlord.Country,
		&createdLandlord.IsArchived,
		&createdLandlord.CreatedAt,
		&createdLandlord.UpdatedAt,
	)

	if err != nil {
		apiError := handleLandlordErrors(err)

		w.WriteHeader(int(apiError.Code))
		json.NewEncoder(w).Encode(apiError)
		return
	}

	s.logger.Debug("Landlord Created", "landlord", createdLandlord)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdLandlord)
}

func (s *Server) LandlordsArchive(w http.ResponseWriter, r *http.Request, id string) {
	var archivedLandlord Landlord

	sql := `
		UPDATE landlords 
		SET 
			is_archived = NOW(),  
			updated_at = NOW()
		WHERE id = $1
		RETURNING 
			id, 
			name, 
			email, 
			mobile, 
			phone, 
			address_line_1,
			address_line_2,
			suburb,
			postcode,
			state,
			country,
			is_archived,
			created_at,
			updated_at
	`

	err := s.dbpool.QueryRow(context.Background(), sql, id).Scan(
		&archivedLandlord.Id,
		&archivedLandlord.Name,
		&archivedLandlord.Email,
		&archivedLandlord.Mobile,
		&archivedLandlord.Phone,
		&archivedLandlord.AddressLine1,
		&archivedLandlord.AddressLine2,
		&archivedLandlord.Suburb,
		&archivedLandlord.Postcode,
		&archivedLandlord.State,
		&archivedLandlord.Country,
		&archivedLandlord.IsArchived,
		&archivedLandlord.CreatedAt,
		&archivedLandlord.UpdatedAt,
	)

	if err != nil {
		apiError := handleLandlordErrors(err)

		w.WriteHeader(int(apiError.Code))
		json.NewEncoder(w).Encode(apiError)
		return
	}

	s.logger.Debug("Landlord Archived", "landlord", archivedLandlord)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(archivedLandlord)
}

func (s *Server) LandlordsGet(w http.ResponseWriter, r *http.Request, id string) {
	var landlord Landlord

	sql := `
		SELECT 
			id, 
			name, 
			email, 
			mobile, 
			phone, 
			address_line_1,
			address_line_2,
			suburb,
			postcode,
			state,
			country,
			is_archived,
			created_at,
			updated_at
		FROM landlords WHERE id = $1
	`

	err := s.dbpool.QueryRow(context.Background(), sql, id).Scan(
		&landlord.Id,
		&landlord.Name,
		&landlord.Email,
		&landlord.Mobile,
		&landlord.Phone,
		&landlord.AddressLine1,
		&landlord.AddressLine2,
		&landlord.Suburb,
		&landlord.Postcode,
		&landlord.State,
		&landlord.Country,
		&landlord.IsArchived,
		&landlord.CreatedAt,
		&landlord.UpdatedAt,
	)

	w.Header().Set("Content-Type", "application/json")

	if err != nil {
		apiError := handleLandlordErrors(err)

		w.WriteHeader(int(apiError.Code))
		json.NewEncoder(w).Encode(apiError)
		return
	}

	s.logger.Debug("Landlord Retrieved", "landlord", landlord)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(landlord)
}

func (s *Server) LandlordsUpdate(w http.ResponseWriter, r *http.Request, id string) {
	var payload UpdateLandlord
	err := json.NewDecoder(r.Body).Decode(&payload)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Error{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	setClause, values, paramCount := buildLandlordUpdateSetClause(payload)
	values = append(values, id)

	sql := fmt.Sprintf(`
		UPDATE landlords 
		%s
		WHERE id = $%d
		RETURNING id,
			name,
			email,
			mobile,
			phone,
			address_line_1,
			address_line_2,
			suburb,
			postcode,
			state,
			country,
			is_archived,
			created_at,
			updated_at
	`, setClause, paramCount+1)

	row := s.dbpool.QueryRow(
		context.Background(),
		sql,
		values...,
	)

	var updatedLandlord Landlord

	err = row.Scan(
		&updatedLandlord.Id,
		&updatedLandlord.Name,
		&updatedLandlord.Email,
		&updatedLandlord.Mobile,
		&updatedLandlord.Phone,
		&updatedLandlord.AddressLine1,
		&updatedLandlord.AddressLine2,
		&updatedLandlord.Suburb,
		&updatedLandlord.Postcode,
		&updatedLandlord.State,
		&updatedLandlord.Country,
		&updatedLandlord.IsArchived,
		&updatedLandlord.CreatedAt,
		&updatedLandlord.UpdatedAt,
	)

	if err != nil {
		apiError := handleLandlordErrors(err)

		w.WriteHeader(int(apiError.Code))
		json.NewEncoder(w).Encode(apiError)
		return
	}

	s.logger.Debug("Landlord Updated", "landlord", updatedLandlord)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updatedLandlord)
}

func (s *Server) PropertiesList(w http.ResponseWriter, r *http.Request, params PropertiesListParams) {
	properties := []Property{}

	limit, page, offset := handlePaginationParams(params)

	conditions := map[string]interface{}{
		"full_address":  params.Address,
		"archived_only": params.ArchivedOnly,
	}

	whereClause, queryParams, paramCount := buildWhereClause(conditions)

	sql := fmt.Sprintf(`
		SELECT COUNT(*) 
		FROM properties
		%s
	`, whereClause)

	var total int

	err := s.dbpool.QueryRow(context.Background(), sql, queryParams...).Scan(&total)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	queryParams = append(queryParams, limit, offset)

	sql = fmt.Sprintf(`
		SELECT 
			id,
			street_number,
			street_name,
			suburb,
			state,
			postcode,
			country,
			landlord_id,
			management_fee,
			management_gained,
			management_lost,
			is_archived,
			created_at,
			updated_at
		FROM properties
		%s
		ORDER BY street_name, street_number
		LIMIT $%d
		OFFSET $%d
	`, whereClause, paramCount, paramCount+1)

	rows, err := s.dbpool.Query(context.Background(), sql, queryParams...)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		property, err := scanProperty(rows)

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

	resp := PropertyList{
		Items: properties,
		Pagination: PaginatedMetadata{
			Total:       int32(total),
			Count:       int32(len(properties)),
			PerPage:     int32(limit),
			CurrentPage: int32(page),
			TotalPages:  int32(math.Ceil(float64(total) / float64(limit))),
		},
	}

	s.logger.Debug("Properties List Response", "response", resp)

	json.NewEncoder(w).Encode(resp)
}

func (s *Server) PropertiesCreate(w http.ResponseWriter, r *http.Request) {
	var payload CreateProperty
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Error{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	id, err := uuid.NewV7()

	if err != nil {
		s.logger.Info("Failed to generate UUID", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Error{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error",
		})
		return
	}

	sql := `
		INSERT INTO properties (
			id,
			street_number,
			street_name,
			suburb,
			state,	
			postcode,
			country,
			landlord_id,
			management_fee,
			management_gained
		) VALUES (
			$1,
			$2,
			$3,
			$4,
			$5,
			$6,
			$7,
			$8,
			$9,
			$10
		) RETURNING 
			id, 
			street_number, 
			street_name, 
			suburb, 
			state, 
			postcode, 
			country,
			landlord_id, 
			management_fee, 
			management_gained,
			management_lost,
			is_archived,
			created_at,
			updated_at
	`

	row := s.dbpool.QueryRow(
		context.Background(),
		sql,
		id.String(),
		payload.StreetNumber,
		payload.StreetName,
		payload.Suburb,
		payload.State,
		payload.Postcode,
		payload.Country,
		payload.LandlordId,
		payload.ManagementFee,
		payload.ManagementGained,
	)

	createdProperty, err := scanProperty(row)

	if err != nil {
		s.logger.Info("Failed to create property", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Error{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error",
		})
		return
	}

	s.logger.Debug("Property Created", "property", createdProperty)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdProperty)
}

func (s *Server) PropertiesArchive(w http.ResponseWriter, r *http.Request, id string) {
	var archivedProperty Property

	sql := `
		UPDATE properties 
		SET 
			is_archived = NOW(),
			updated_at = NOW()
		WHERE id = $1
		RETURNING 
			id,
			street_number,
			street_name,
			suburb,
			state,
			postcode,
			country,
			landlord_id,
			management_fee,
			management_gained,
			management_lost,
			is_archived,
			created_at,
			updated_at
	`

	row := s.dbpool.QueryRow(context.Background(), sql, id)

	archivedProperty, err := scanProperty(row)

	if err != nil {
		s.logger.Info("Failed to archive property", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Error{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error",
		})
		return
	}

	s.logger.Debug("Property Archived", "property", archivedProperty)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(archivedProperty)
}

func (s *Server) PropertiesGet(w http.ResponseWriter, r *http.Request, id string) {
	var property Property

	sql := `
		SELECT 
			id,
			street_number,
			street_name,
			suburb,
			state,
			postcode,
			country,
			landlord_id,
			management_fee,
			management_gained,
			management_lost,
			is_archived,
			created_at,
			updated_at
		FROM properties 
		WHERE id = $1
	`

	row := s.dbpool.QueryRow(context.Background(), sql, id)

	property, err := scanProperty(row)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Error{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error",
		})
		return
	}

	s.logger.Debug("Property Retrieved", "property", property)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(property)
}

func (s *Server) PropertiesUpdate(w http.ResponseWriter, r *http.Request, id string) {
	var payload UpdateProperty
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Error{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	setClause, values, paramCount := buildPropertyUpdateSetClause(payload)
	values = append(values, id)

	sql := fmt.Sprintf(`
		UPDATE properties 
		%s
		WHERE id = $%d
		RETURNING 
			id, 
			street_number, 
			street_name, 
			suburb, 
			state, 
			postcode, 
			country,
			landlord_id, 
			management_fee, 
			management_gained,
			management_lost,
			is_archived,
			created_at,
			updated_at
	`, setClause, paramCount+1)

	row := s.dbpool.QueryRow(
		context.Background(),
		sql,
		values...,
	)

	updatedProperty, err := scanProperty(row)

	if err != nil {
		s.logger.Info("Failed to update property", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Error{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error",
		})
		return
	}

	s.logger.Debug("Property Updated", "property", updatedProperty)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updatedProperty)
}

func (s *Server) TenantsList(w http.ResponseWriter, r *http.Request, params TenantsListParams) {
	tenants := []Tenant{}

	limit, page, offset := handlePaginationParams(params)

	conditions := map[string]interface{}{
		"name":          params.Name,
		"archived_only": params.ArchivedOnly,
	}

	whereClause, queryParams, paramCount := buildWhereClause(conditions)

	sql := fmt.Sprintf(`
		SELECT COUNT(*) 
		FROM tenants 
		%s
	`, whereClause)

	var total int

	err := s.dbpool.QueryRow(context.Background(), sql, queryParams...).Scan(&total)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Error{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error",
		})
		return
	}

	queryParams = append(queryParams, limit, offset)

	sql = fmt.Sprintf(`
		SELECT 
			id,
			name,
			email,
			mobile,
			phone,
			paid_from,
			paid_to,
			rental_amount,
			frequency,
			original_start_date,
			start_date,
			end_date,
			termination_date,
			termination_reason,
			vacate_date,
			is_archived,
			property_id,
			created_at,
			updated_at
		FROM tenants
		%s
		LIMIT $%d
		OFFSET $%d
	`, whereClause, paramCount, paramCount+1)

	rows, err := s.dbpool.Query(context.Background(), sql, queryParams...)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		tenant, err := scanTenant(rows)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(Error{
				Code:    http.StatusInternalServerError,
				Message: "Internal server error",
			})
			return
		}
		tenants = append(tenants, tenant)
	}

	if err := rows.Err(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Error{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	resp := TenantList{
		Items: tenants,
		Pagination: PaginatedMetadata{
			Total:       int32(total),
			Count:       int32(len(tenants)),
			PerPage:     int32(limit),
			CurrentPage: int32(page),
			TotalPages:  int32(math.Ceil(float64(total) / float64(limit))),
		},
	}

	s.logger.Debug("Tenants List Response", "response", resp)

	json.NewEncoder(w).Encode(resp)
}

func (s *Server) TenantsCreate(w http.ResponseWriter, r *http.Request) {
	var payload CreateTenant
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	id, err := uuid.NewV7()

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sql := `
		INSERT INTO tenants (
			id,
			name,
			email,
			mobile,
			phone,
			paid_from,
			paid_to,
			rental_amount,
			frequency,
			original_start_date,
			start_date,
			end_date,
			property_id
		)
		VALUES (
			$1,
			$2,
			$3,
			$4,
			$5,
			$6,
			$7,
			$8,
			$9,
			$10,
			$11,
			$12,
			$13
		) RETURNING 
		 	id,
			name,
			email,
			mobile,
			phone,
			paid_from,
			paid_to,
			rental_amount,
			frequency,
			original_start_date,
			start_date,
			end_date,
			termination_date,
			termination_reason,
			vacate_date,
			is_archived,
			property_id,
			created_at,
			updated_at
	`

	row := s.dbpool.QueryRow(
		context.Background(),
		sql,
		id.String(),
		payload.Name,
		payload.Email,
		payload.Mobile,
		payload.Phone,
		payload.PaidTo,
		payload.PaidTo,
		payload.RentalAmount,
		payload.Frequency,
		payload.OriginalStartDate,
		payload.StartDate,
		payload.EndDate,
		payload.PropertyId,
	)

	createdTenant, err := scanTenant(row)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Error{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error",
		})
		return
	}

	s.logger.Debug("Tenant Created", "tenant", createdTenant)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdTenant)
}

func (s *Server) TenantsArchive(w http.ResponseWriter, r *http.Request, id string) {
	var archivedTenant Tenant

	sql := `
		UPDATE tenants 
		SET 
			is_archived = NOW(),
			updated_at = NOW()
		WHERE id = $1
		RETURNING 
			id,
			name,
			email,
			mobile,
			phone,
			paid_from,
			paid_to,
			rental_amount,
			frequency,
			original_start_date,
			start_date,
			end_date,
			termination_date,
			termination_reason,
			vacate_date,
			is_archived,
			property_id,
			created_at,
			updated_at
	`

	row := s.dbpool.QueryRow(context.Background(), sql, id)

	archivedTenant, err := scanTenant(row)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Error{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error",
		})
		return
	}

	s.logger.Debug("Tenant Archived", "tenant", archivedTenant)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(archivedTenant)
}

func (s *Server) TenantsGet(w http.ResponseWriter, r *http.Request, id string) {
	var tenant Tenant

	sql := `
		SELECT id,
			name,
			email,
			mobile,
			phone,
			paid_from,
			paid_to,
			rental_amount,
			frequency,
			original_start_date,
			start_date,
			end_date,
			termination_date,
			termination_reason,
			vacate_date,
			is_archived,
			property_id,
			created_at,
			updated_at
		FROM tenants 
		WHERE id = $1
	`

	row := s.dbpool.QueryRow(context.Background(), sql, id)

	tenant, err := scanTenant(row)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.logger.Debug("Tenant Retrieved", "tenant", tenant)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tenant)
}

func (s *Server) TenantsUpdate(w http.ResponseWriter, r *http.Request, id string) {
	var payload UpdateTenant
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Error{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	setClause, values, paramCount := buildTenantUpdateSetClause(payload)
	values = append(values, id)

	sql := fmt.Sprintf(`
		UPDATE tenants 
		%s
		WHERE id = $%d
		RETURNING 
			id,
			name,
			email,
			mobile,
			phone,
			paid_from,
			paid_to,
			rental_amount,
			frequency,
			original_start_date,
			start_date,
			end_date,
			termination_date,
			termination_reason,
			vacate_date,
			is_archived,
			property_id,
			created_at,
			updated_at
	`, setClause, paramCount)

	var updatedTenant Tenant

	row := s.dbpool.QueryRow(
		context.Background(),
		sql,
		values...,
	)

	updatedTenant, err = scanTenant(row)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Error{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error",
		})
		return
	}

	s.logger.Debug("Tenant Updated", "tenant", updatedTenant)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updatedTenant)
}

func scanProperty(scanner interface {
	Scan(dest ...interface{}) error
}) (Property, error) {
	var property Property

	var managementGained pgtype.Date
	var managementLost *pgtype.Date

	err := scanner.Scan(
		&property.Id,
		&property.StreetNumber,
		&property.StreetName,
		&property.Suburb,
		&property.State,
		&property.Postcode,
		&property.Country,
		&property.LandlordId,
		&property.ManagementFee,
		&managementGained,
		&managementLost,
		&property.IsArchived,
		&property.CreatedAt,
		&property.UpdatedAt,
	)

	if err != nil {
		return property, err
	}

	property.ManagementGained = openapi_types.Date{
		Time: managementGained.Time,
	}

	if managementLost != nil {
		property.ManagementLost = &openapi_types.Date{
			Time: managementLost.Time,
		}
	}

	return property, nil
}

func scanTenant(scanner interface {
	Scan(dest ...interface{}) error
}) (Tenant, error) {
	var tenant Tenant
	var originalStartDate pgtype.Date
	var startDate pgtype.Date
	var endDate pgtype.Date
	var terminationDate *pgtype.Date
	var vacateDate *pgtype.Date

	err := scanner.Scan(
		&tenant.Id,
		&tenant.Name,
		&tenant.Email,
		&tenant.Mobile,
		&tenant.Phone,
		&tenant.PaidFrom,
		&tenant.PaidTo,
		&tenant.RentalAmount,
		&tenant.Frequency,
		&originalStartDate,
		&startDate,
		&endDate,
		&terminationDate,
		&tenant.TerminationReason,
		&vacateDate,
		&tenant.IsArchived,
		&tenant.PropertyId,
		&tenant.CreatedAt,
		&tenant.UpdatedAt,
	)

	tenant.OriginalStartDate.Time = originalStartDate.Time
	tenant.StartDate.Time = startDate.Time
	tenant.EndDate.Time = endDate.Time

	if terminationDate != nil {
		tenant.TerminationDate = &openapi_types.Date{Time: terminationDate.Time}
	}

	if vacateDate != nil {
		tenant.VacateDate = &openapi_types.Date{Time: vacateDate.Time}
	}

	return tenant, err
}

func handleLandlordErrors(err error) Error {
	if err == pgx.ErrNoRows {
		return Error{Message: "No landlord found with the specified ID", Code: http.StatusNotFound}
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if pgErr.Code == "22P02" {
			return Error{Message: "Invalid Landlord ID format - must be a valid UUID", Code: http.StatusBadRequest}
		}
	}

	return Error{Message: "Internal server error", Code: http.StatusInternalServerError}
}

func handlePaginationParams(params any) (int, int, int) {
	var pagePtr *int32
	var limitPtr *int32

	switch p := params.(type) {
	case LandlordsListParams:
		pagePtr = p.Page
		limitPtr = p.Limit
	case TenantsListParams:
		pagePtr = p.Page
		limitPtr = p.Limit
	case PropertiesListParams:
		pagePtr = p.Page
		limitPtr = p.Limit
	default:
		// Optional: handle unexpected types
		return 20, 1, 0 // default: page=1, limit=20
	}

	page := 1
	if pagePtr != nil {
		page = int(*pagePtr)
	}

	limit := 20
	if limitPtr != nil {
		limit = int(*limitPtr)
	}

	if limit > 100 {
		limit = 100
	}

	offset := (page - 1) * limit
	return limit, page, offset
}

// buildWhereClause constructs a SQL WHERE clause from a map of conditions
// conditions is a map of column names to their search values.
//
// Returns the WHERE clause string, query parameters, and the next parameter number
func buildWhereClause(conditions map[string]interface{}) (string, []interface{}, int) {
	if len(conditions) == 0 {
		return "", []interface{}{}, 1
	}

	clauses := []string{}
	params := []interface{}{}
	paramCount := 1

	for column, value := range conditions {
		if value == nil {
			continue
		}

		switch column {
		case "name", "full_address":
			if v, ok := value.(*string); ok && v != nil {
				clauses = append(clauses, fmt.Sprintf("%s ILIKE $%d", column, paramCount))
				params = append(params, "%"+*v+"%")
				paramCount++
			}
		case "archived_only":
			if v, ok := value.(*bool); ok && v != nil && *v {
				clauses = append(clauses, "is_archived is not null")
			} else {
				clauses = append(clauses, "is_archived is null")
			}
		default:
			clauses = append(clauses, fmt.Sprintf("%s = $%d", column, paramCount))
			params = append(params, value)
			paramCount++
		}
	}

	if len(clauses) == 0 {
		return "", []interface{}{}, 1
	}

	return "WHERE " + strings.Join(clauses, " AND "), params, paramCount
}

func buildLandlordUpdateSetClause(payload UpdateLandlord) (string, []interface{}, int) {
	fields := []string{}
	values := []interface{}{}
	paramCount := 0

	if payload.Name != nil && *payload.Name != "" {
		paramCount++
		fields = append(fields, fmt.Sprintf("name = $%d", paramCount))
		values = append(values, *payload.Name)
	}

	if payload.Email != nil {
		paramCount++
		fields = append(fields, fmt.Sprintf("email = $%d", paramCount))
		values = append(values, *payload.Email)
	}

	if payload.Mobile != nil && *payload.Mobile != "" {
		paramCount++
		fields = append(fields, fmt.Sprintf("mobile = $%d", paramCount))
		values = append(values, *payload.Mobile)
	}

	if payload.Phone != nil && *payload.Phone != "" {
		paramCount++
		fields = append(fields, fmt.Sprintf("phone = $%d", paramCount))
		values = append(values, *payload.Phone)
	}

	if payload.AddressLine1 != nil && *payload.AddressLine1 != "" {
		paramCount++
		fields = append(fields, fmt.Sprintf("address_line_1 = $%d", paramCount))
		values = append(values, *payload.AddressLine1)
	}

	if payload.AddressLine2 != nil && *payload.AddressLine2 != "" {
		paramCount++
		fields = append(fields, fmt.Sprintf("address_line_2 = $%d", paramCount))
		values = append(values, *payload.AddressLine2)
	}

	if payload.Suburb != nil && *payload.Suburb != "" {
		paramCount++
		fields = append(fields, fmt.Sprintf("suburb = $%d", paramCount))
		values = append(values, *payload.Suburb)
	}

	if payload.Postcode != nil && *payload.Postcode != "" {
		paramCount++
		fields = append(fields, fmt.Sprintf("postcode = $%d", paramCount))
		values = append(values, *payload.Postcode)
	}

	if payload.State != nil && *payload.State != "" {
		paramCount++
		fields = append(fields, fmt.Sprintf("state = $%d", paramCount))
		values = append(values, *payload.State)
	}

	if payload.Country != nil && *payload.Country != "" {
		paramCount++
		fields = append(fields, fmt.Sprintf("country = $%d", paramCount))
		values = append(values, *payload.Country)
	}

	if payload.IsArchived == nil {
		fields = append(fields, "is_archived = null")
	}

	if len(fields) > 0 {
		fields = append(fields, "updated_at = NOW()")
	}

	setClause := "SET\n" + strings.Join(fields, ",\n")

	return setClause, values, paramCount
}

func buildPropertyUpdateSetClause(payload UpdateProperty) (string, []interface{}, int) {
	fields := []string{}
	values := []interface{}{}
	paramCount := 0

	if payload.StreetNumber != nil && *payload.StreetNumber != "" {
		paramCount++
		fields = append(fields, fmt.Sprintf("street_number = $%d", paramCount))
		values = append(values, *payload.StreetNumber)
	}

	if payload.StreetName != nil && *payload.StreetName != "" {
		paramCount++
		fields = append(fields, fmt.Sprintf("street_name = $%d", paramCount))
		values = append(values, *payload.StreetName)
	}

	if payload.Suburb != nil && *payload.Suburb != "" {
		paramCount++
		fields = append(fields, fmt.Sprintf("suburb = $%d", paramCount))
		values = append(values, *payload.Suburb)
	}

	if payload.Postcode != nil && *payload.Postcode != "" {
		paramCount++
		fields = append(fields, fmt.Sprintf("postcode = $%d", paramCount))
		values = append(values, *payload.Postcode)
	}

	if payload.State != nil && *payload.State != "" {
		paramCount++
		fields = append(fields, fmt.Sprintf("state = $%d", paramCount))
		values = append(values, *payload.State)
	}

	if payload.Country != nil && *payload.Country != "" {
		paramCount++
		fields = append(fields, fmt.Sprintf("country = $%d", paramCount))
		values = append(values, *payload.Country)
	}

	if payload.ManagementFee != nil {
		paramCount++
		fields = append(fields, fmt.Sprintf("management_fee = $%d", paramCount))
		values = append(values, *payload.ManagementFee)
	}

	if payload.ManagementGained != nil {
		paramCount++
		fields = append(fields, fmt.Sprintf("management_gained = $%d", paramCount))
		values = append(values, *payload.ManagementGained)
	}

	if payload.ManagementLost != nil {
		paramCount++
		fields = append(fields, fmt.Sprintf("management_lost = $%d", paramCount))
		values = append(values, *payload.ManagementLost)
	}

	if payload.IsArchived == nil {
		fields = append(fields, "is_archived = null")
	}

	if len(fields) > 0 {
		fields = append(fields, "updated_at = NOW()")
	}

	setClause := "SET\n" + strings.Join(fields, ",\n")

	return setClause, values, paramCount
}

func buildTenantUpdateSetClause(payload UpdateTenant) (string, []interface{}, int) {
	fields := []string{}
	values := []interface{}{}
	paramCount := 0

	if payload.Name != nil && *payload.Name != "" {
		paramCount++
		fields = append(fields, fmt.Sprintf("name = $%d", paramCount))
		values = append(values, *payload.Name)
	}

	if payload.Email != nil && *payload.Email != "" {
		paramCount++
		fields = append(fields, fmt.Sprintf("email = $%d", paramCount))
		values = append(values, *payload.Email)
	}

	if payload.Mobile != nil && *payload.Mobile != "" {
		paramCount++
		fields = append(fields, fmt.Sprintf("mobile = $%d", paramCount))
		values = append(values, *payload.Mobile)
	}

	if payload.Phone != nil && *payload.Phone != "" {
		paramCount++
		fields = append(fields, fmt.Sprintf("phone = $%d", paramCount))
		values = append(values, *payload.Phone)
	}

	if payload.PaidFrom != nil {
		paramCount++
		fields = append(fields, fmt.Sprintf("paid_from = $%d", paramCount))
		values = append(values, *payload.PaidFrom)
	}

	if payload.PaidTo != nil {
		paramCount++
		fields = append(fields, fmt.Sprintf("paid_to = $%d", paramCount))
		values = append(values, *payload.PaidTo)
	}

	if payload.RentalAmount != nil {
		paramCount++
		fields = append(fields, fmt.Sprintf("rental_amount = $%d", paramCount))
		values = append(values, *payload.RentalAmount)
	}

	if payload.Frequency != nil && *payload.Frequency != "" {
		paramCount++
		fields = append(fields, fmt.Sprintf("frequency = $%d", paramCount))
		values = append(values, *payload.Frequency)
	}

	if payload.OriginalStartDate != nil {
		paramCount++
		fields = append(fields, fmt.Sprintf("original_start_date = $%d", paramCount))
		values = append(values, *payload.OriginalStartDate)
	}

	if payload.StartDate != nil {
		paramCount++
		fields = append(fields, fmt.Sprintf("start_date = $%d", paramCount))
		values = append(values, *payload.StartDate)
	}

	if payload.EndDate != nil {
		paramCount++
		fields = append(fields, fmt.Sprintf("end_date = $%d", paramCount))
		values = append(values, *payload.EndDate)
	}

	if payload.TerminationDate != nil {
		paramCount++
		fields = append(fields, fmt.Sprintf("termination_date = $%d", paramCount))
		values = append(values, *payload.TerminationDate)
	}

	if payload.TerminationReason != nil && *payload.TerminationReason != "" {
		paramCount++
		fields = append(fields, fmt.Sprintf("termination_reason = $%d", paramCount))
		values = append(values, *payload.TerminationReason)
	}

	if payload.VacateDate != nil {
		paramCount++
		fields = append(fields, fmt.Sprintf("vacate_date = $%d", paramCount))
		values = append(values, *payload.VacateDate)
	}

	if payload.IsArchived == nil {
		fields = append(fields, "is_archived = null")
	}

	if len(fields) > 0 {
		fields = append(fields, "updated_at = NOW()")
	}

	setClause := "SET\n" + strings.Join(fields, ",\n")

	return setClause, values, paramCount
}
