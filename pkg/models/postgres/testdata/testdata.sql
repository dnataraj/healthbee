INSERT INTO sites(site_hash, url, period, pattern, created)
    VALUES (md5('http://site1/test'), 'http://site1/test', 5, 'content', CURRENT_TIMESTAMP);
INSERT INTO sites(site_hash, url, period, pattern, created)
    VALUES (md5('http://site2/test'), 'http://site2/test', 3, 'content', CURRENT_TIMESTAMP);
INSERT INTO sites(site_hash, url, period, pattern, created)
VALUES (md5('http://site3/test'), 'http://site3/test', 8, 'content', CURRENT_TIMESTAMP);

INSERT INTO results(site_id, checked_at, response_time, result, matched)
    VALUES (1, CURRENT_TIMESTAMP, 600, 200, true);
INSERT INTO results(site_id, checked_at, response_time, result, matched)
    VALUES (2, CURRENT_TIMESTAMP, 1200, 400, false);
INSERT INTO results(site_id, checked_at, response_time, result, matched)
VALUES (3, CURRENT_TIMESTAMP, 200, 400, false);
