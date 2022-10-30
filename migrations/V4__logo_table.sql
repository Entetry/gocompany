CREATE TABLE logo
(
    id         uuid    NOT NULL PRIMARY KEY,
    company_id uuid    NOT NULL,
    image      varchar NOT NULL
);
