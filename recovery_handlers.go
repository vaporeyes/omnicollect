// ABOUTME: Authenticated deletion-recovery routes scoped through requestStore.
// ABOUTME: Lost acknowledgements never promise rollback; clients inspect persisted batches.
package main

import (
	"database/sql"
	"net/http"
)

func (s *Server) handleDiscardDeletion(w http.ResponseWriter, r *http.Request) {
	if err := s.requestStore(r).DiscardDeletion(r.PathValue("id")); err != nil {
		writeError(w, http.StatusConflict, "Discard was not confirmed; reload the recovery list: "+err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
func (s *Server) handleListDeletions(w http.ResponseWriter, r *http.Request) {
	batches, err := s.requestStore(r).ListDeletions()
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, batches)
}
func (s *Server) handleRecoverDeletion(w http.ResponseWriter, r *http.Request) {
	count, err := s.requestStore(r).RecoverDeletion(r.PathValue("id"))
	if err != nil {
		status := http.StatusConflict
		if err == sql.ErrNoRows {
			status = http.StatusNotFound
		}
		writeError(w, status, "Recovery was not confirmed. Refresh the collection and recovery list before retrying: "+err.Error())
		return
	}
	writeJSON(w, 200, map[string]int{"restored": count})
}
