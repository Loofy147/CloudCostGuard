package api

import (
	"database/sql"
	"net/http"

	"cloudcostguard/backend/internal/api/handlers"
	"cloudcostguard/backend/internal/service"
	"github.com/swaggo/http-swagger"

	"go.uber.org/zap"
)

func NewRouter(estimatorSvc *service.Estimator, logger *zap.Logger, db *sql.DB) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/swagger/", httpSwagger.WrapHandler)

	estimateHandler := handlers.NewEstimateHandler(estimatorSvc, logger)
	mux.Handle("/estimate", estimateHandler)

	statusHandler := handlers.NewStatusHandler(logger, db)
	mux.Handle("/status", statusHandler)

	return mux
}
