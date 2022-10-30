CREATE TABLE users
(
    id           uuid PRIMARY KEY,
    username     varchar(32),
    email        varchar(32),
    passwordHash varchar(60)
);