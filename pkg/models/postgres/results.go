package postgres

import (
	"database/sql"
	"errors"
	"github.com/dnataraj/healthbee/pkg/models"
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

func (r *ResultModel) Get(id int) (*models.CheckResult, error) {
	res := &models.CheckResult{}
	stmt := `SELECT id, site_id, checked_at, result, matched FROM results WHERE id = $1`
	err := r.DB.QueryRow(stmt, id).Scan(&res.ID, &res.SiteID, &res.At, &res.ResponseCode, &res.MatchedPattern)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.ErrNoRecord
		}
		return nil, err
	}
	return res, nil
}
