package lib

import (
	"errors"
	"log"
	"net/http"
	"order-service/src/services"

	"github.com/gin-gonic/gin"
)

func HandleServiceError(ctx *gin.Context, err error) {
	var validationErr services.ValidationError
	var stockErr *services.InsufficientStockError
	var transitionErr *services.InvalidTransitionError

	switch {
	case errors.Is(err, services.ErrProductNotFound),
		errors.Is(err, services.ErrCustomerNotFound),
		errors.Is(err, services.ErrOrderNotFound):
		RespondError(ctx, http.StatusNotFound, "not_found", err.Error())
	case errors.Is(err, services.ErrDuplicateSKU),
		errors.Is(err, services.ErrDuplicateEmail):
		RespondError(ctx, http.StatusConflict, "duplicate_entry", err.Error())
	case errors.As(err, &transitionErr):
		RespondError(ctx, http.StatusConflict, "invalid_status_transition", err.Error())
	case errors.As(err, &stockErr):
		RespondError(ctx, http.StatusBadRequest, "insufficient stock", err.Error())
	case errors.As(err, &validationErr):
		RespondError(ctx, http.StatusBadRequest, "validation_error", err.Error())
	default:
		log.Printf("internal error: %v", err)
		RespondError(ctx, http.StatusInternalServerError, "internal_error", "internal server error")
	}
}
