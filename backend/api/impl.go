package api

import (
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
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte("Not yet implemented"))
}

func (s *Server) LandlordsCreate(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte("Not yet implemented"))
	
}

func (s *Server) LandlordsArchive(w http.ResponseWriter, r *http.Request, id string) {
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte("Not yet implemented"))
}

func (s *Server) LandlordsGet(w http.ResponseWriter, r *http.Request, id string) {
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte("Not yet implemented"))
}

func (s *Server) LandlordsUpdate(w http.ResponseWriter, r *http.Request, id string) {
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte("Not yet implemented"))
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