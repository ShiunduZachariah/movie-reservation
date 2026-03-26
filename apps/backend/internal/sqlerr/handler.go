package sqlerr

import (
	"database/sql"
	"errors"

	"github.com/ShiunduZachariah/movie-reservation/apps/backend/internal/errs"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func HandleError(err error) error {
	if err == nil {
		return nil
	}

	var httpErr *errs.HTTPError
	if errors.As(err, &httpErr) {
		return err
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23503":
			return errs.BadRequest("REFERENCE_NOT_FOUND", "referenced record does not exist", nil)
		case "23505":
			return errs.Conflict("ALREADY_EXISTS", "resource already exists")
		case "23514":
			return errs.BadRequest("CHECK_VIOLATION", "request did not satisfy required constraints", nil)
		default:
			return errs.Internal("database error")
		}
	}

	if errors.Is(err, pgx.ErrNoRows) || errors.Is(err, sql.ErrNoRows) {
		return errs.NotFound("NOT_FOUND", "resource not found")
	}

	return errs.Internal("database error")
}
