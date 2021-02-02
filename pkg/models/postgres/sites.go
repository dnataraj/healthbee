package postgres

import (
	"database/sql"
	"errors"
	"github.com/dnataraj/healthbee/pkg/models"
	"github.com/lib/pq"
	"time"
)

type SiteModel struct {
	DB *sql.DB
}

// Insert adds an entry to the Sites table
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

// Get fetches a registered Site from the Site table
func (s *SiteModel) Get(id int) (*models.Site, error) {
	site := &models.Site{}
	// We handle the interval separately here to maintain its unit (i.e. seconds)
	var p int
	stmt := `SELECT id, url, period, pattern, created FROM sites WHERE id = $1`
	err := s.DB.QueryRow(stmt, id).Scan(&site.ID, &site.URL, &p, &site.Interval, &site.Created)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.ErrNoRecord
		}
		return nil, err
	}
	site.Interval = models.Period(time.Duration(p) * time.Second)
	return site, nil
}

// GetAll fetches the latest 20 registered sites from the site table
func (s *SiteModel) GetAll() ([]*models.Site, error) {
	sites := make([]*models.Site, 0)
	stmt := `SELECT id, url, period, pattern, created FROM sites ORDER BY created DESC LIMIT 20`
	rows, err := s.DB.Query(stmt)
	defer rows.Close()
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		site := &models.Site{}
		var p int
		if err := rows.Scan(&site.ID, &site.URL, &p, &site.Pattern, &site.Created); err != nil {
			// For now, we'll simple return on any failure rather than serve partials
			return nil, err
		}
		site.Interval = models.Period(time.Duration(p) * time.Second)
		sites = append(sites, site)
	}

	return sites, nil
}
