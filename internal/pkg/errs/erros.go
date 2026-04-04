package errs

import (
	"database/sql"
	"errors"
	"net/http"
)

var (
	ErrNotFound       = errors.New("ressource non trouvée")
	ErrAlreadyExists  = errors.New("la ressource existe déjà")
	ErrInvalidInput   = errors.New("données invalides")
	ErrInternalServer = errors.New("erreur interne du serveur")
	ErrUnauthorized   = errors.New("Vous n'êtes pas autorisé")
	// ErrStockInsuffisant = errors.New("stock insuffisant")
)

func MapHTTPError(err error) int {
	if errors.Is(err, sql.ErrNoRows) {
		return http.StatusNotFound
	}
	if errors.Is(err, ErrNotFound) {
		return http.StatusNotFound
	}
	if errors.Is(err, ErrAlreadyExists) {
		return http.StatusConflict
	}
	if errors.Is(err, ErrInvalidInput) {
		return http.StatusBadRequest
	}
	if errors.Is(err, ErrUnauthorized) {
		return http.StatusUnauthorized
	}
	// if errors.Is(err, ErrStockInsuffisant) {
	// 	return http.StatusUnauthorized
	// }
	return http.StatusConflict
}

// var (
// 	ErrNotFound       = errors.New("ressource non trouvée")
// 	ErrAlreadyExists  = errors.New("la ressource existe déjà")
// 	ErrInvalidInput   = errors.New("données invalides")
// 	ErrInternalServer = errors.New("erreur interne du serveur")
// )

// func MapHTTPError(err error) int {
// 	if errors.Is(err, sql.ErrNoRows) {
// 		return http.StatusNotFound
// 	}
// 	if errors.Is(err, ErrNotFound) {
// 		return http.StatusNotFound
// 	}
// 	if errors.Is(err, ErrAlreadyExists) {
// 		return http.StatusConflict
// 	}
// 	if errors.Is(err, ErrInvalidInput) {
// 		return http.StatusBadRequest
// 	}
// 	return http.StatusInternalServerError
// }
