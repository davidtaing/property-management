package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"net/http"

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

	sql := `
		SELECT COUNT(*) 
		FROM landlords
	`

	var total int

	err := s.dbpool.QueryRow(context.Background(), sql).Scan(&total)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sql = `
		SELECT 
			id, 
			name, 
			email, 
			mobile, 
			phone, 
			is_archived 
		FROM landlords
		ORDER BY name
		LIMIT $1 
		OFFSET $2
	`

	rows, err := s.dbpool.Query(context.Background(), sql, limit, offset)
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
		apiError := handleLandlordErrors(err)

		w.WriteHeader(int(apiError.Code))
		json.NewEncoder(w).Encode(apiError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

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

	json.NewEncoder(w).Encode(resp)
}

func (s *Server) LandlordsCreate(w http.ResponseWriter, r *http.Request) {
	var payload Landlord
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
		INSERT INTO landlords (
			id,
			name,
			email,
			mobile,
			phone
		) VALUES (
			$1,
			$2,
			$3,
			$4,
			$5
		) RETURNING 
		 	id,
			name,
			email,
			mobile,
			phone,
			is_archived`

	row := s.dbpool.QueryRow(
		context.Background(),
		sql,
		id.String(),
		payload.Name,
		payload.Email,
		payload.Mobile,
		payload.Phone,
	)

	var createdLandlord Landlord

	err = row.Scan(
		&createdLandlord.Id,
		&createdLandlord.Name,
		&createdLandlord.Email,
		&createdLandlord.Mobile,
		&createdLandlord.Phone,
		&createdLandlord.IsArchived,
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
		UPDATE landlords SET is_archived = NOW() WHERE id = $1
		RETURNING 
			id, 
			name, 
			email, 
			mobile, 
			phone, 
			is_archived
	`

	err := s.dbpool.QueryRow(context.Background(), sql, id).Scan(
		&archivedLandlord.Id,
		&archivedLandlord.Name,
		&archivedLandlord.Email,
		&archivedLandlord.Mobile,
		&archivedLandlord.Phone,
		&archivedLandlord.IsArchived,
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
			is_archived 
		FROM landlords WHERE id = $1
	`

	err := s.dbpool.QueryRow(context.Background(), sql, id).Scan(
		&landlord.Id,
		&landlord.Name,
		&landlord.Email,
		&landlord.Mobile,
		&landlord.Phone,
		&landlord.IsArchived,
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
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	sql := `
		UPDATE landlords SET
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
			is_archived
	`

	row := s.dbpool.QueryRow(
		context.Background(),
		sql,
		payload.Name,
		payload.Email,
		payload.Mobile,
		payload.Phone,
		id,
	)

	var updatedLandlord Landlord

	err = row.Scan(&updatedLandlord.Id, &updatedLandlord.Name, &updatedLandlord.Email, &updatedLandlord.Mobile, &updatedLandlord.Phone, &updatedLandlord.IsArchived)

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

	sql := `
		SELECT COUNT(*) 
		FROM properties
	`

	var total int

	err := s.dbpool.QueryRow(context.Background(), sql).Scan(&total)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sql = `
		SELECT 
			id,
			address_line_1,
			address_line_2,
			suburb,
			state,
			postcode,
			landlord_id,
			management_fee,
			is_archived
		FROM properties
		LIMIT $1 
		OFFSET $2
	`

	rows, err := s.dbpool.Query(context.Background(), sql, limit, offset)

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
	var payload Property
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
		INSERT INTO properties (
			id,
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
			$7,
			$8
		) RETURNING 
			id, 
			address_line_1, 
			address_line_2, 
			suburb, 
			state, 
			postcode, 
			landlord_id, 
			management_fee, 
			is_archived
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
		payload.LandlordId,
		payload.ManagementFee,
	)

	var createdProperty Property

	err = row.Scan(&createdProperty.Id, &createdProperty.AddressLine1, &createdProperty.AddressLine2, &createdProperty.Suburb, &createdProperty.State, &createdProperty.Postcode, &createdProperty.LandlordId, &createdProperty.ManagementFee, &createdProperty.IsArchived)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
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
		SET is_archived = NOW() 
		WHERE id = $1
		RETURNING 
			id,
			address_line_1,
			address_line_2,
			suburb,
			state,
			postcode,
			landlord_id,
			management_fee,
			is_archived
	`

	err := s.dbpool.QueryRow(context.Background(), sql, id).Scan(
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

	sql := `
		SELECT id,
			address_line_1,
			address_line_2,
			suburb,
			state,
			postcode,
			landlord_id,
			management_fee,
			is_archived
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

	sql := `
		UPDATE properties SET
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
			is_archived
	`

	row := s.dbpool.QueryRow(
		context.Background(),
		sql,
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

	err = row.Scan(&updatedProperty.Id, &updatedProperty.AddressLine1, &updatedProperty.AddressLine2, &updatedProperty.Suburb, &updatedProperty.State, &updatedProperty.Postcode, &updatedProperty.LandlordId, &updatedProperty.ManagementFee, &updatedProperty.IsArchived)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updatedProperty)
}

func (s *Server) TenantsList(w http.ResponseWriter, r *http.Request, params TenantsListParams) {
	tenants := []Tenant{}

	limit, page, offset := handlePaginationParams(params)

	sql := `
		SELECT COUNT(*) 
		FROM tenants
	`

	var total int

	err := s.dbpool.QueryRow(context.Background(), sql).Scan(&total)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sql = `
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
			property_id
		FROM tenants
		LIMIT $1 
		OFFSET $2
	`

	rows, err := s.dbpool.Query(context.Background(), sql, limit, offset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		tenant, err := scanTenant(rows)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			fmt.Println(err.Error())
			return
		}
		tenants = append(tenants, tenant)
	}

	if err := rows.Err(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
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
	var payload Tenant
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
			property_id
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
		http.Error(w, err.Error(), http.StatusInternalServerError)
		fmt.Println(err.Error())
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
		SET is_archived = NOW() 
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
			property_id
	`

	row := s.dbpool.QueryRow(context.Background(), sql, id)

	archivedTenant, err := scanTenant(row)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
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
			property_id
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
		http.Error(w, err.Error(), http.StatusBadRequest)
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
			property_id = $12
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
			property_id
	`

	var updatedTenant Tenant

	row := s.dbpool.QueryRow(context.Background(), sql, payload.Name, payload.Email, payload.Mobile, payload.Phone, payload.OriginalStartDate, payload.StartDate, payload.EndDate, payload.TerminationDate, payload.TerminationReason, payload.VacateDate, payload.IsArchived, payload.PropertyId, id)

	updatedTenant, err = scanTenant(row)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
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

	offset := (page - 1) * limit
	return limit, page, offset
}
