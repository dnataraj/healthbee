INSERT INTO sites(site_hash, url, period, pattern, created)
    VALUES (md5('https://www.example.com'), 'https://www.example.com', 5, 'content', CURRENT_TIMESTAMP);
INSERT INTO sites(site_hash, url, period, pattern, created)
    VALUES (md5('https://www.example.org'), 'https://www.example.org', 3, 'content', CURRENT_TIMESTAMP);

INSERT INTO results(site_id, checked_at, response_time, result, matched)
    VALUES (1, CURRENT_TIMESTAMP, 600, 200, true);
INSERT INTO results(site_id, checked_at, response_time, result, matched)
    VALUES (2, CURRENT_TIMESTAMP, 1200, 400, false);
INSERT INTO results(site_id, checked_at, response_time, result, matched)
    VALUES (2, CURRENT_TIMESTAMP, 200, 400, false);
