package main

import "github.com/jackc/pgx/v5/pgtype"

func getPgtypeText(s string) pgtype.Text {
	return pgtype.Text{
		String: s,
		Valid:  true,
	}
}

func getPgtypeNumeric(s string) (pgtype.Numeric, error) {
	var numeric pgtype.Numeric
	err := numeric.Scan(s)
	return numeric, err
}
