package postgres

import (
	"database/sql"
	"time"
)

type ResultModel struct {
	DB *sql.DB
}

func (r *ResultModel) Insert(siteID int, checkedAt time.Time, code int, matched bool) (int, error) {
	var id int
	stmt := `INSERT INTO results (site_id, checked_at, result, matched) VALUES ($1, $2, $3, $4) RETURNING id`
	err := r.DB.QueryRow(stmt, siteID, checkedAt, code, matched).Scan(&id)
	if err != nil {
		return 0, nil
	}
	return id, nil
}
