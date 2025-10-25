package main

import "github.com/jackc/pgx/v5/pgtype"

func getPgtypeText(s string) pgtype.Text {
	return pgtype.Text{
		String: s,
		Valid:  true,
	}
}
