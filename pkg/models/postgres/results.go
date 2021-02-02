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

// Insert adds an availability metric to the Results table
func (r *ResultModel) Insert(siteID int, checkedAt time.Time, code int, matched bool) (int, error) {
	var id int
	stmt := `INSERT INTO results (site_id, checked_at, result, matched) VALUES ($1, $2, $3, $4) RETURNING id`
	err := r.DB.QueryRow(stmt, siteID, checkedAt, code, matched).Scan(&id)
	if err != nil {
		return 0, nil
	}
	return id, nil
}

// Get fetches an availability metric from the Results table given a metric ID
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

// GetResultsForSite fetches the latest 20 site availability metrics for a given Site ID
// Results are ordered by the check timestamp.
func (r *ResultModel) GetResultsForSite(siteID int) ([]*models.CheckResult, error) {
	metrics := make([]*models.CheckResult, 0)
	stmt := `SELECT id, site_id, checked_at, result, matched FROM results WHERE site_id = $1 ORDER BY checked_at DESC LIMIT 20`
	rows, err := r.DB.Query(stmt, siteID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		res := &models.CheckResult{}
		if err := rows.Scan(&res.ID, &res.SiteID, &res.At, &res.ResponseCode, &res.MatchedPattern); err != nil {
			return nil, err
		}
		metrics = append(metrics, res)
	}

	return metrics, nil
}
