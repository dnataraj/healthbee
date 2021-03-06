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
func (r *ResultModel) Insert(siteID int, checkedAt time.Time, responseTime models.Period, code int, matched bool) (int, error) {
	var id int
	stmt := `INSERT INTO results (site_id, checked_at, response_time, result, matched) VALUES ($1, $2, $3, $4, $5) RETURNING id`
	err := r.DB.QueryRow(stmt, siteID, checkedAt, responseTime.Duration().Milliseconds(), code, matched).Scan(&id)
	if err != nil {
		return -1, nil
	}
	return id, nil
}

// Get fetches an availability metric from the Results table given a metric ID
func (r *ResultModel) Get(id int) (*models.CheckResult, error) {
	res := &models.CheckResult{}
	var rt int
	stmt := `SELECT id, site_id, checked_at, response_time, result, matched FROM results WHERE id = $1`
	err := r.DB.QueryRow(stmt, id).Scan(&res.ID, &res.SiteID, &res.At, &rt, &res.ResponseCode, &res.MatchedPattern)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.ErrNoRecord
		}
		return nil, err
	}
	res.ResponseTime = models.Period(time.Duration(rt) * time.Millisecond)
	return res, nil
}

// GetResultsForSite fetches the latest 20 site availability metrics for a given Site ID
// Results are ordered by the check timestamp.
func (r *ResultModel) GetResultsForSite(siteID int) ([]*models.CheckResult, error) {
	metrics := make([]*models.CheckResult, 0)
	stmt := `SELECT id, site_id, checked_at, response_time, result, matched FROM results WHERE site_id = $1 ORDER BY checked_at DESC LIMIT 20`
	rows, err := r.DB.Query(stmt, siteID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		res := &models.CheckResult{}
		var rt int
		if err := rows.Scan(&res.ID, &res.SiteID, &res.At, &rt, &res.ResponseCode, &res.MatchedPattern); err != nil {
			// It's odd that Scan doesn't return sql.ErrNoRows as described here:
			// https://pkg.go.dev/database/sql#ErrNoRows
			if errors.Is(err, sql.ErrNoRows) {
				return nil, models.ErrNoRecord
			}
			return nil, err
		}
		res.ResponseTime = models.Period(time.Duration(rt) * time.Millisecond)
		metrics = append(metrics, res)
	}
	if len(metrics) == 0 {
		// We did not find the site
		return nil, models.ErrNoRecord
	}

	return metrics, nil
}
