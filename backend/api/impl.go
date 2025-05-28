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
		"name": params.Name,
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

	fmt.Println(resp)

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

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
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

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(landlord)
}

func (s *Server) LandlordsUpdate(w http.ResponseWriter, r *http.Request, id string) {
	var payload Landlord
	err := json.NewDecoder(r.Body).Decode(&payload)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Error{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	sql := `
		UPDATE landlords SET
			name = $1,
			email = $2,
			mobile = $3,
			phone = $4,
			address_line_1 = $5,
			address_line_2 = $6,
			suburb = $7,
			postcode = $8,
			state = $9,
			country = $10,
			updated_at = NOW()
		WHERE id = $11
		RETURNING id,
			name,
			email,
			mobile,
			phone,
			is_archived,
			created_at,
			updated_at
	`

	row := s.dbpool.QueryRow(
		context.Background(),
		sql,
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
		id,
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

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updatedLandlord)
}

func (s *Server) PropertiesList(w http.ResponseWriter, r *http.Request, params PropertiesListParams) {
	properties := []Property{}

	limit, page, offset := handlePaginationParams(params)

	conditions := map[string]interface{}{
		"full_address": params.Address,
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
			address_line_1,
			address_line_2,
			suburb,
			state,
			postcode,
			country,
			landlord_id,
			management_fee,
			is_archived,
			created_at,
			updated_at
		FROM properties
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
		var property Property
		err := rows.Scan(
			&property.Id,
			&property.AddressLine1,
			&property.AddressLine2,
			&property.Suburb,
			&property.State,
			&property.Postcode,
			&property.Country,
			&property.LandlordId,
			&property.ManagementFee,
			&property.IsArchived,
			&property.CreatedAt,
			&property.UpdatedAt,
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
			address_line_1,
			address_line_2,
			suburb,
			state,	
			postcode,
			country,
			landlord_id,
			management_fee
		) VALUES (
			$1,
			$2,
			$3,
			$4,
			$5,
			$6,
			$7,
			$8
		) RETURNING 
			id, 
			address_line_1, 
			address_line_2, 
			suburb, 
			state, 
			postcode, 
			country,
			landlord_id, 
			management_fee, 
			is_archived,
			created_at,
			updated_at
	`

	row := s.dbpool.QueryRow(
		context.Background(),
		sql,
		id.String(),
		payload.AddressLine1,
		payload.AddressLine2,
		payload.Suburb,
		payload.State,
		payload.Postcode,
		payload.Country,
		payload.LandlordId,
		payload.ManagementFee,
	)

	var createdProperty Property

	err = row.Scan(
		&createdProperty.Id,
		&createdProperty.AddressLine1,
		&createdProperty.AddressLine2,
		&createdProperty.Suburb,
		&createdProperty.State,
		&createdProperty.Postcode,
		&createdProperty.Country,
		&createdProperty.LandlordId,
		&createdProperty.ManagementFee,
		&createdProperty.IsArchived,
		&createdProperty.CreatedAt,
		&createdProperty.UpdatedAt,
	)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Error{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error",
		})
		return
	}

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
			address_line_1,
			address_line_2,
			suburb,
			state,
			postcode,
			country,
			landlord_id,
			management_fee,
			is_archived,
			created_at,
			updated_at
	`

	err := s.dbpool.QueryRow(context.Background(), sql, id).Scan(
		&archivedProperty.Id,
		&archivedProperty.AddressLine1,
		&archivedProperty.AddressLine2,
		&archivedProperty.Suburb,
		&archivedProperty.State,
		&archivedProperty.Postcode,
		&archivedProperty.Country,
		&archivedProperty.LandlordId,
		&archivedProperty.ManagementFee,
		&archivedProperty.IsArchived,
		&archivedProperty.CreatedAt,
		&archivedProperty.UpdatedAt,
	)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Error{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(archivedProperty)
}

func (s *Server) PropertiesGet(w http.ResponseWriter, r *http.Request, id string) {
	var property Property

	sql := `
		SELECT id,
			address_line_1,
			address_line_2,
			suburb,
			state,
			postcode,
			country,
			landlord_id,
			management_fee,
			is_archived,
			created_at,
			updated_at
		FROM properties 
		WHERE id = $1
	`

	err := s.dbpool.QueryRow(context.Background(), sql, id).Scan(
		&property.Id,
		&property.AddressLine1,
		&property.AddressLine2,
		&property.Suburb,
		&property.State,
		&property.Postcode,
		&property.Country,
		&property.LandlordId,
		&property.ManagementFee,
		&property.IsArchived,
		&property.CreatedAt,
		&property.UpdatedAt,
	)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Error{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error",
		})
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
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Error{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	sql := `
		UPDATE properties SET
			address_line_1 = $1,
			address_line_2 = $2,
			suburb = $3,
			state = $4,
			postcode = $5,
			country = $6,
			landlord_id = $7,
			management_fee = $8,
			updated_at = NOW()
		WHERE id = $9
		RETURNING 
			id, 
			address_line_1, 
			address_line_2, 
			suburb, 
			state, 
			postcode, 
			country,
			landlord_id, 
			management_fee, 
			is_archived,
			created_at,
			updated_at
	`

	row := s.dbpool.QueryRow(
		context.Background(),
		sql,
		payload.AddressLine1,
		payload.AddressLine2,
		payload.Suburb,
		payload.State,
		payload.Postcode,
		payload.Country,
		payload.LandlordId,
		payload.ManagementFee,
		id,
	)

	var updatedProperty Property

	err = row.Scan(
		&updatedProperty.Id,
		&updatedProperty.AddressLine1,
		&updatedProperty.AddressLine2,
		&updatedProperty.Suburb,
		&updatedProperty.State,
		&updatedProperty.Postcode,
		&updatedProperty.Country,
		&updatedProperty.LandlordId,
		&updatedProperty.ManagementFee,
		&updatedProperty.IsArchived,
		&updatedProperty.CreatedAt,
		&updatedProperty.UpdatedAt,
	)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Error{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updatedProperty)
}

func (s *Server) TenantsList(w http.ResponseWriter, r *http.Request, params TenantsListParams) {
	tenants := []Tenant{}

	limit, page, offset := handlePaginationParams(params)

	conditions := map[string]interface{}{
		"name": params.Name,
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
			$9
		) RETURNING 
		 	id,
			name,
			email,
			mobile,
			phone,
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

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tenant)
}

func (s *Server) TenantsUpdate(w http.ResponseWriter, r *http.Request, id string) {
	var payload Tenant
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Error{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	sql := `
		UPDATE tenants SET
			name = $1,
			email = $2,
			mobile = $3,
			phone = $4,
			original_start_date = $5,
			start_date = $6,
			end_date = $7,
			termination_date = $8,
			termination_reason = $9,
			vacate_date = $10,
			is_archived = $11,
			property_id = $12,
			updated_at = NOW()
		WHERE id = $13
		RETURNING 
			id,
			name,
			email,
			mobile,
			phone,
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

	var updatedTenant Tenant

	row := s.dbpool.QueryRow(context.Background(), sql, payload.Name, payload.Email, payload.Mobile, payload.Phone, payload.OriginalStartDate, payload.StartDate, payload.EndDate, payload.TerminationDate, payload.TerminationReason, payload.VacateDate, payload.IsArchived, payload.PropertyId, id)

	updatedTenant, err = scanTenant(row)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Error{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updatedTenant)
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
