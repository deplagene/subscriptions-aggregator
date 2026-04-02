-- +goose Up
CREATE TABLE subscriptions (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    service_name TEXT NOT NULL,
    price INTEGER NOT NULL CHECK (price > 0),
    user_id uuid NOT NULL,
    started_at DATE NOT NULL,
    ended_at DATE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT subscriptions_period_check
        CHECK (ended_at IS NULL OR ended_at >= started_at)
);

-- +goose Down
DROP TABLE IF EXISTS subscriptions;
