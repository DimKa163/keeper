DO $$
BEGIN
IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'secret_type') THEN
CREATE TYPE secret_type AS ENUM ('login_pass', 'text', 'bank_card', 'other');
END IF;
END $$;
CREATE TABLE IF NOT EXISTS secret(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    modified_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    user_id UUID NOT NULL,
    big_data boolean NOT NULL,
    secret_type secret_type NOT NULL,
    payload BYTEA,
    dek BYTEA,
    path TEXT,
    version INT,
    deleted boolean DEFAULT FALSE
);

CREATE TABLE IF NOT EXISTS sync_state(
    id  VARCHAR(25),
    user_id UUID NOT NULL,
    value INT,

    PRIMARY KEY(id, user_id)
);
