CREATE TABLE refresh_sessions
(
    refresh_token uuid                   NOT NULL PRIMARY KEY,
    user_id       uuid                   NOT NULL,
    ua            character varying(200) NOT NULL,
    fingerprint   character varying(200) NOT NULL,
    ip            character varying(15)  NOT NULL,
    expires_at    bigint                 NOT NULL
);
