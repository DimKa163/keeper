CREATE TABLE IF NOT EXISTS users(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    login VARCHAR(25) NOT NULL,
    password BYTEA NOT NULL,
    salt BYTEA NOT NULL
);

CREATE UNIQUE INDEX users_login_uix ON public.users(login);