DO $$
BEGIN
IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'stored_data_type') THEN
CREATE TYPE stored_data_type AS ENUM ('login_pass', 'text', 'bank_card', 'other');
END IF;
END $$;
CREATE TABLE IF NOT EXISTS stored_data(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    name TEXT,
    user_id UUID NOT NULL,
    large boolean,
    data_type stored_data_type NOT NULL,
    data BYTEA,
    data_nonce BYTEA,
    dek BYTEA,
    dek_nonce BYTEA,
    version INT
);

