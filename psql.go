package main

import (
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func getPgtypeUuid(uuid uuid.UUID) pgtype.UUID {
	return pgtype.UUID{
		Bytes: uuid,
		Valid: true,
	}
}
