package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/jmoiron/sqlx"
)

type BaseHandler struct {
	db *sqlx.DB
}

func NewBaseHandler(db *sqlx.DB) *BaseHandler {
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
