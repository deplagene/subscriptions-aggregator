-- name: CreateSubscription :one
INSERT INTO subscriptions (
    service_name,
    price,
    user_id,
    started_at,
    ended_at
) VALUES (
    sqlc.arg('service_name'),
    sqlc.arg('price'),
    sqlc.arg('user_id'),
    sqlc.arg('started_at'),
    sqlc.narg('ended_at')
)
RETURNING id;

-- name: GetSubscriptionByID :one
SELECT id, service_name, price, user_id, started_at, ended_at, created_at, updated_at
FROM subscriptions
WHERE id = sqlc.arg('id')
LIMIT 1;

-- name: ListSubscriptions :many
SELECT id, service_name, price, user_id, started_at, ended_at, created_at, updated_at
FROM subscriptions
WHERE (sqlc.narg('user_id')::uuid IS NULL OR user_id = sqlc.narg('user_id')::uuid)
  AND (sqlc.narg('service_name')::text IS NULL OR service_name = sqlc.narg('service_name')::text)
ORDER BY created_at DESC, id DESC
LIMIT sqlc.arg('limit')
OFFSET sqlc.arg('offset');

-- name: UpdateSubscription :execrows
UPDATE subscriptions
SET
    service_name = sqlc.arg('service_name'),
    price = sqlc.arg('price'),
    user_id = sqlc.arg('user_id'),
    started_at = sqlc.arg('started_at'),
    ended_at = sqlc.narg('ended_at'),
    updated_at = NOW()
WHERE id = sqlc.arg('id');

-- name: DeleteSubscription :execrows
DELETE FROM subscriptions
WHERE id = sqlc.arg('id');

-- name: CalculateSubscriptionsTotal :one
SELECT COALESCE(
    SUM(
        s.price::bigint * (
            (
                EXTRACT(YEAR FROM LEAST(COALESCE(s.ended_at, sqlc.arg('to')::date), sqlc.arg('to')::date))::bigint
                - EXTRACT(YEAR FROM GREATEST(s.started_at, sqlc.arg('from')::date))::bigint
            ) * 12
            + (
                EXTRACT(MONTH FROM LEAST(COALESCE(s.ended_at, sqlc.arg('to')::date), sqlc.arg('to')::date))::bigint
                - EXTRACT(MONTH FROM GREATEST(s.started_at, sqlc.arg('from')::date))::bigint
            ) + 1
        )
    ),
    0
)::bigint AS total
FROM subscriptions AS s
WHERE s.started_at <= sqlc.arg('to')::date
  AND COALESCE(s.ended_at, sqlc.arg('to')::date) >= sqlc.arg('from')::date
  AND (sqlc.narg('user_id')::uuid IS NULL OR s.user_id = sqlc.narg('user_id')::uuid)
  AND (sqlc.narg('service_name')::text IS NULL OR s.service_name = sqlc.narg('service_name')::text);
