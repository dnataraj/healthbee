DROP TABLE IF EXISTS sites CASCADE;
DROP TABLE IF EXISTS results;

CREATE TABLE sites (
    id INT GENERATED ALWAYS AS IDENTITY,
    site_hash TEXT UNIQUE NOT NULL,
    url VARCHAR(2000) NOT NULL,
    period INT NOT NULL,
    pattern VARCHAR(100) NOT NULL,
    created TIMESTAMPTZ,
    PRIMARY KEY(id)
);

CREATE INDEX idx_site ON sites(id);
CREATE INDEX idx_site_hash ON sites (site_hash);

CREATE TABLE results (
    id INT GENERATED ALWAYS AS IDENTITY,
    site_id INT NOT NULL ,
    checked_at TIMESTAMPTZ,
    response_time INT,
    result INT,
    matched BOOLEAN NOT NULL,
    CONSTRAINT fk_sites
        FOREIGN KEY(site_id)
            REFERENCES sites(id) ON DELETE CASCADE
);

CREATE INDEX idx_site_id ON results(site_id)