package postgres

import (
	"database/sql"
	"github.com/dnataraj/healthbee/pkg/models"
	"github.com/lib/pq"
	"time"
)

type SiteModel struct {
	DB *sql.DB
}

func (s *SiteModel) Insert(URL string, interval time.Duration, pattern string) (int, error) {
	var siteID int
	stmt := `INSERT INTO sites (site_hash, url, period, pattern, created) VALUES (md5($1), $1, $2, $3, $4) RETURNING id`
	err := s.DB.QueryRow(stmt, URL, interval.Seconds(), pattern, time.Now()).Scan(&siteID)
	if err != nil {
		if perr, ok := err.(*pq.Error); ok {
			if perr.Code == uniquenessViolation {
				return 0, models.ErrDuplicateSite
			}
		}
		return 0, err
	}
	return siteID, nil
}
