DO $$
BEGIN
IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'data_type') THEN
CREATE TYPE data_type AS ENUM ('login_pass', 'text', 'bank_card', 'other');
END IF;
END $$;
CREATE TABLE IF NOT EXISTS data(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    name TEXT,
    user_id UUID NOT NULL,
    large boolean,
    data_type data_type NOT NULL,
    payload BYTEA,
    payload_nonce BYTEA,
    dek BYTEA,
    dek_nonce BYTEA,
    version INT
);

