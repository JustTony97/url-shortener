package handler

import (
	"context"
	"net/http"
	"time"
)

type DB interface {
	PingContext(ctx context.Context) error
}

type BaseHandler struct {
	db DB
}

func NewBaseHandler(db DB) *BaseHandler {
	return &BaseHandler{db: db}
}

func (bh *BaseHandler) Healthcheck(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), time.Second*3)
	defer cancel()

	err := bh.db.PingContext(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
