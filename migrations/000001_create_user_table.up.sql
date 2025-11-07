CREATE TABLE IF NOT EXISTS users(
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    login VARCHAR(25) NOT NULL,
    password BYTEA NOT NULL,
    salt BYTEA NOT NULL,
    encrypt_salt BYTEA NOT NULL
);

CREATE UNIQUE INDEX users_login_uix ON public.users(login);