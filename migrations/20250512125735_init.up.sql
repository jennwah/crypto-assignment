CREATE SCHEMA crypto;

-- UUID generation
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Enums
CREATE TYPE crypto.transaction_type AS ENUM ('deposit', 'withdraw', 'transfer');
CREATE TYPE crypto.transaction_status AS ENUM ('success', 'failed');

-- wallets table
CREATE TABLE crypto.wallets (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID UNIQUE NOT NULL,
    balance BIGINT NOT NULL CHECK (balance >= 0),
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- transactions table
CREATE TABLE crypto.transactions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    initiator_wallet_id UUID NOT NULL REFERENCES crypto.wallets(id),
    type crypto.transaction_type NOT NULL,
    status crypto.transaction_status NOT NULL,
    amount BIGINT NOT NULL CHECK (amount > 0),
    recipient_wallet_id UUID REFERENCES crypto.wallets(id),
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);
