DROP TABLE IF EXISTS sites CASCADE;
DROP TABLE IF EXISTS results;

CREATE TABLE sites (
    id INT GENERATED ALWAYS AS IDENTITY,
    url VARCHAR(2000) NOT NULL,
    period INT NOT NULL,
    pattern VARCHAR(100) NOT NULL,
    created TIMESTAMPTZ,
    PRIMARY KEY(id)
);

CREATE TABLE results (
    id INT GENERATED ALWAYS AS IDENTITY,
    site_id INT,
    checked_at TIMESTAMPTZ,
    response_code INT,
    found_pattern BOOLEAN NOT NULL,
    CONSTRAINT fk_sites
        FOREIGN KEY(site_id)
            REFERENCES sites(id)
);