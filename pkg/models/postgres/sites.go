package postgres

import (
	"database/sql"
	"time"
)

type SiteModel struct {
	DB *sql.DB
}

func (s *SiteModel) Insert(URL string, interval time.Duration, pattern string) (int, error) {
	var siteID int
	stmt := `INSERT INTO sites (url, period, pattern, created) VALUES($1, $2, $3, $4) RETURNING id`
	err := s.DB.QueryRow(stmt, URL, interval.Seconds(), pattern, time.Now()).Scan(&siteID)
	if err != nil {
		return 0, err
	}
	return siteID, nil
}
