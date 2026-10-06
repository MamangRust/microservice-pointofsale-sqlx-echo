package repository

import (
	"context"

	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
	"github.com/MamangRust/microservice-point-of-sale-shared/errors/transaction_errors"
	db "github.com/MamangRust/microservice-point-of-sale-transacton/database/schema"
	"github.com/jmoiron/sqlx"
)

const getMonthlyAmountTransactionFailed = `WITH
    monthly_data AS (
        SELECT
            EXTRACT(
                YEAR
                FROM t.created_at
            )::integer AS year,
            EXTRACT(
                MONTH
                FROM t.created_at
            )::integer AS month,
            COUNT(*) AS total_failed,
            COALESCE(SUM(t.amount), 0)::integer AS total_amount
        FROM transactions t
        WHERE
            t.deleted_at IS NULL
            AND t.payment_status = 'failed'
            AND (
                (
                    t.created_at >= $1::timestamp
                    AND t.created_at <= $2::timestamp
                )
                OR (
                    t.created_at >= $3::timestamp
                    AND t.created_at <= $4::timestamp
                )
            )
        GROUP BY
            EXTRACT(
                YEAR
                FROM t.created_at
            ),
            EXTRACT(
                MONTH
                FROM t.created_at
            )
    ),
    formatted_data AS (
        SELECT
            year::text,
            TO_CHAR(
                TO_DATE(month::text, 'MM'),
                'Mon'
            ) AS month,
            total_failed,
            total_amount
        FROM monthly_data
        UNION ALL
        SELECT
            EXTRACT(
                YEAR
                FROM $1::timestamp
            )::text AS year,
            TO_CHAR($1::timestamp, 'Mon') AS month,
            0 AS total_failed,
            0 AS total_amount
        WHERE
            NOT EXISTS (
                SELECT 1
                FROM monthly_data
                WHERE
                    year = EXTRACT(
                        YEAR
                        FROM $1::timestamp
                    )::integer
                    AND month = EXTRACT(
                        MONTH
                        FROM $1::timestamp
                    )::integer
            )
        UNION ALL
        SELECT
            EXTRACT(
                YEAR
                FROM $3::timestamp
            )::text AS year,
            TO_CHAR($3::timestamp, 'Mon') AS month,
            0 AS total_failed,
            0 AS total_amount
        WHERE
            NOT EXISTS (
                SELECT 1
                FROM monthly_data
                WHERE
                    year = EXTRACT(
                        YEAR
                        FROM $3::timestamp
                    )::integer
                    AND month = EXTRACT(
                        MONTH
                        FROM $3::timestamp
                    )::integer
            )
    )
SELECT year, month, total_failed, total_amount
FROM formatted_data
ORDER BY year DESC, TO_DATE(month, 'Mon') DESC
`

const getMonthlyAmountTransactionFailedByMerchant = `WITH
    monthly_data AS (
        SELECT
            EXTRACT(
                YEAR
                FROM t.created_at
            )::integer AS year,
            EXTRACT(
                MONTH
                FROM t.created_at
            )::integer AS month,
            COUNT(*) AS total_failed,
            COALESCE(SUM(t.amount), 0)::integer AS total_amount
        FROM transactions t
        WHERE
            t.deleted_at IS NULL
            AND t.payment_status = 'failed'
            AND t.merchant_id = $5
            AND (
                (
                    t.created_at >= $1::timestamp
                    AND t.created_at <= $2::timestamp
                )
                OR (
                    t.created_at >= $3::timestamp
                    AND t.created_at <= $4::timestamp
                )
            )
        GROUP BY
            EXTRACT(
                YEAR
                FROM t.created_at
            ),
            EXTRACT(
                MONTH
                FROM t.created_at
            )
    ),
    formatted_data AS (
        SELECT
            year::text,
            TO_CHAR(
                TO_DATE(month::text, 'MM'),
                'Mon'
            ) AS month,
            total_failed,
            total_amount
        FROM monthly_data
        UNION ALL
        SELECT
            EXTRACT(
                YEAR
                FROM $1::timestamp
            )::text AS year,
            TO_CHAR($1::timestamp, 'Mon') AS month,
            0 AS total_failed,
            0 AS total_amount
        WHERE
            NOT EXISTS (
                SELECT 1
                FROM monthly_data
                WHERE
                    year = EXTRACT(
                        YEAR
                        FROM $1::timestamp
                    )::integer
                    AND month = EXTRACT(
                        MONTH
                        FROM $1::timestamp
                    )::integer
            )
        UNION ALL
        SELECT
            EXTRACT(
                YEAR
                FROM $3::timestamp
            )::text AS year,
            TO_CHAR($3::timestamp, 'Mon') AS month,
            0 AS total_failed,
            0 AS total_amount
        WHERE
            NOT EXISTS (
                SELECT 1
                FROM monthly_data
                WHERE
                    year = EXTRACT(
                        YEAR
                        FROM $3::timestamp
                    )::integer
                    AND month = EXTRACT(
                        MONTH
                        FROM $3::timestamp
                    )::integer
            )
    )
SELECT year, month, total_failed, total_amount
FROM formatted_data
ORDER BY year DESC, TO_DATE(month, 'Mon') DESC
`

const getMonthlyAmountTransactionSuccess = `WITH
    monthly_data AS (
        SELECT
            EXTRACT(
                YEAR
                FROM t.created_at
            )::integer AS year,
            EXTRACT(
                MONTH
                FROM t.created_at
            )::integer AS month,
            COUNT(*) AS total_success,
            COALESCE(SUM(t.amount), 0)::integer AS total_amount
        FROM transactions t
        WHERE
            t.deleted_at IS NULL
            AND t.payment_status = 'success'
            AND (
                (
                    t.created_at >= $1::timestamp
                    AND t.created_at <= $2::timestamp
                )
                OR (
                    t.created_at >= $3::timestamp
                    AND t.created_at <= $4::timestamp
                )
            )
        GROUP BY
            EXTRACT(
                YEAR
                FROM t.created_at
            ),
            EXTRACT(
                MONTH
                FROM t.created_at
            )
    ),
    formatted_data AS (
        SELECT
            year::text,
            TO_CHAR(
                TO_DATE(month::text, 'MM'),
                'Mon'
            ) AS month,
            total_success,
            total_amount
        FROM monthly_data
        UNION ALL
        SELECT
            EXTRACT(
                YEAR
                FROM $1::timestamp
            )::text AS year,
            TO_CHAR($1::timestamp, 'Mon') AS month,
            0 AS total_success,
            0 AS total_amount
        WHERE
            NOT EXISTS (
                SELECT 1
                FROM monthly_data
                WHERE
                    year = EXTRACT(
                        YEAR
                        FROM $1::timestamp
                    )::integer
                    AND month = EXTRACT(
                        MONTH
                        FROM $1::timestamp
                    )::integer
            )
        UNION ALL
        SELECT
            EXTRACT(
                YEAR
                FROM $3::timestamp
            )::text AS year,
            TO_CHAR($3::timestamp, 'Mon') AS month,
            0 AS total_success,
            0 AS total_amount
        WHERE
            NOT EXISTS (
                SELECT 1
                FROM monthly_data
                WHERE
                    year = EXTRACT(
                        YEAR
                        FROM $3::timestamp
                    )::integer
                    AND month = EXTRACT(
                        MONTH
                        FROM $3::timestamp
                    )::integer
            )
    )
SELECT year, month, total_success, total_amount
FROM formatted_data
ORDER BY year DESC, TO_DATE(month, 'Mon') DESC
`

const getMonthlyAmountTransactionSuccessByMerchant = `WITH
    monthly_data AS (
        SELECT
            EXTRACT(
                YEAR
                FROM t.created_at
            )::integer AS year,
            EXTRACT(
                MONTH
                FROM t.created_at
            )::integer AS month,
            COUNT(*) AS total_success,
            COALESCE(SUM(t.amount), 0)::integer AS total_amount
        FROM transactions t
        WHERE
            t.deleted_at IS NULL
            AND t.payment_status = 'success'
            AND t.merchant_id = $5
            AND (
                (
                    t.created_at >= $1::timestamp
                    AND t.created_at <= $2::timestamp
                )
                OR (
                    t.created_at >= $3::timestamp
                    AND t.created_at <= $4::timestamp
                )
            )
        GROUP BY
            EXTRACT(
                YEAR
                FROM t.created_at
            ),
            EXTRACT(
                MONTH
                FROM t.created_at
            )
    ),
    formatted_data AS (
        SELECT
            year::text,
            TO_CHAR(
                TO_DATE(month::text, 'MM'),
                'Mon'
            ) AS month,
            total_success,
            total_amount
        FROM monthly_data
        UNION ALL
        SELECT
            EXTRACT(
                YEAR
                FROM $1::timestamp
            )::text AS year,
            TO_CHAR($1::timestamp, 'Mon') AS month,
            0 AS total_success,
            0 AS total_amount
        WHERE
            NOT EXISTS (
                SELECT 1
                FROM monthly_data
                WHERE
                    year = EXTRACT(
                        YEAR
                        FROM $1::timestamp
                    )::integer
                    AND month = EXTRACT(
                        MONTH
                        FROM $1::timestamp
                    )::integer
            )
        UNION ALL
        SELECT
            EXTRACT(
                YEAR
                FROM $3::timestamp
            )::text AS year,
            TO_CHAR($3::timestamp, 'Mon') AS month,
            0 AS total_success,
            0 AS total_amount
        WHERE
            NOT EXISTS (
                SELECT 1
                FROM monthly_data
                WHERE
                    year = EXTRACT(
                        YEAR
                        FROM $3::timestamp
                    )::integer
                    AND month = EXTRACT(
                        MONTH
                        FROM $3::timestamp
                    )::integer
            )
    )
SELECT year, month, total_success, total_amount
FROM formatted_data
ORDER BY year DESC, TO_DATE(month, 'Mon') DESC
`

const getMonthlyTransactionMethodsByMerchantFailed = `WITH
    date_ranges AS (
        SELECT
            $1::timestamp AS range1_start,
            $2::timestamp AS range1_end,
            $3::timestamp AS range2_start,
            $4::timestamp AS range2_end
    ),
    payment_methods AS (
        SELECT DISTINCT
            payment_method
        FROM transactions
        WHERE deleted_at IS NULL
    ),
    all_months AS (
        SELECT generate_series(
            date_trunc('month', LEAST(
                (SELECT range1_start FROM date_ranges),
                (SELECT range2_start FROM date_ranges)
            )),
            date_trunc('month', GREATEST(
                (SELECT range1_end FROM date_ranges),
                (SELECT range2_end FROM date_ranges)
            )),
            interval '1 month'
        )::date AS activity_month
    ),
    all_combinations AS (
        SELECT 
            am.activity_month,
            pm.payment_method
        FROM all_months am
        CROSS JOIN payment_methods pm
    ),
    monthly_transactions AS (
        SELECT
            date_trunc('month', t.created_at)::date AS activity_month,
            t.payment_method,
            COUNT(t.transaction_id) AS total_transactions,
            COALESCE(SUM(t.amount), 0)::NUMERIC AS total_amount
        FROM transactions t
        JOIN date_ranges dr ON (
            t.created_at BETWEEN dr.range1_start AND dr.range1_end
            OR t.created_at BETWEEN dr.range2_start AND dr.range2_end
        )
        WHERE
            t.deleted_at IS NULL
            AND t.payment_status = 'failed'
            AND t.merchant_id = $5  
        GROUP BY
            date_trunc('month', t.created_at),
            t.payment_method
    )
SELECT 
    TO_CHAR(ac.activity_month, 'Mon') AS month,
    ac.payment_method,
    COALESCE(mt.total_transactions, 0) AS total_transactions,
    COALESCE(mt.total_amount, 0) AS total_amount
FROM all_combinations ac
LEFT JOIN monthly_transactions mt ON 
    ac.activity_month = mt.activity_month
    AND ac.payment_method = mt.payment_method
ORDER BY 
    ac.activity_month, 
    ac.payment_method
`

const getMonthlyTransactionMethodsByMerchantSuccess = `WITH
    date_ranges AS (
        SELECT
            $1::timestamp AS range1_start,
            $2::timestamp AS range1_end,
            $3::timestamp AS range2_start,
            $4::timestamp AS range2_end
    ),
    payment_methods AS (
        SELECT DISTINCT
            payment_method
        FROM transactions
        WHERE deleted_at IS NULL
    ),
    all_months AS (
        SELECT generate_series(
            date_trunc('month', LEAST(
                (SELECT range1_start FROM date_ranges),
                (SELECT range2_start FROM date_ranges)
            )),
            date_trunc('month', GREATEST(
                (SELECT range1_end FROM date_ranges),
                (SELECT range2_end FROM date_ranges)
            )),
            interval '1 month'
        )::date AS activity_month
    ),
    all_combinations AS (
        SELECT 
            am.activity_month,
            pm.payment_method
        FROM all_months am
        CROSS JOIN payment_methods pm
    ),
    monthly_transactions AS (
        SELECT
            date_trunc('month', t.created_at)::date AS activity_month,
            t.payment_method,
            COUNT(t.transaction_id) AS total_transactions,
            COALESCE(SUM(t.amount), 0)::NUMERIC AS total_amount
        FROM transactions t
        JOIN date_ranges dr ON (
            t.created_at BETWEEN dr.range1_start AND dr.range1_end
            OR t.created_at BETWEEN dr.range2_start AND dr.range2_end
        )
        WHERE
            t.deleted_at IS NULL
            AND t.payment_status = 'success'
            AND t.merchant_id = $5  
        GROUP BY
            date_trunc('month', t.created_at),
            t.payment_method
    )
SELECT 
    TO_CHAR(ac.activity_month, 'Mon') AS month,
    ac.payment_method,
    COALESCE(mt.total_transactions, 0) AS total_transactions,
    COALESCE(mt.total_amount, 0) AS total_amount
FROM all_combinations ac
LEFT JOIN monthly_transactions mt ON 
    ac.activity_month = mt.activity_month
    AND ac.payment_method = mt.payment_method
ORDER BY 
    ac.activity_month, 
    ac.payment_method
`

const getMonthlyTransactionMethodsFailed = `WITH
    date_ranges AS (
        SELECT
            $1::timestamp AS range1_start,
            $2::timestamp AS range1_end,
            $3::timestamp AS range2_start,
            $4::timestamp AS range2_end
    ),
    payment_methods AS (
        SELECT DISTINCT
            payment_method
        FROM transactions
        WHERE deleted_at IS NULL
    ),
    all_months AS (
        SELECT generate_series(
            date_trunc('month', LEAST(
                (SELECT range1_start FROM date_ranges),
                (SELECT range2_start FROM date_ranges)
            )),
            date_trunc('month', GREATEST(
                (SELECT range1_end FROM date_ranges),
                (SELECT range2_end FROM date_ranges)
            )),
            interval '1 month'
        )::date AS activity_month
    ),
    all_combinations AS (
        SELECT 
            am.activity_month,
            pm.payment_method
        FROM all_months am
        CROSS JOIN payment_methods pm
    ),
    monthly_transactions AS (
        SELECT
            date_trunc('month', t.created_at)::date AS activity_month,
            t.payment_method,
            COUNT(t.transaction_id) AS total_transactions,
            COALESCE(SUM(t.amount), 0)::NUMERIC AS total_amount
        FROM transactions t
        JOIN date_ranges dr ON (
            t.created_at BETWEEN dr.range1_start AND dr.range1_end
            OR t.created_at BETWEEN dr.range2_start AND dr.range2_end
        )
        WHERE
            t.deleted_at IS NULL
            AND t.payment_status = 'failed'
        GROUP BY
            date_trunc('month', t.created_at),
            t.payment_method
    )
SELECT 
    TO_CHAR(ac.activity_month, 'Mon') AS month,
    ac.payment_method,
    COALESCE(mt.total_transactions, 0) AS total_transactions,
    COALESCE(mt.total_amount, 0) AS total_amount
FROM all_combinations ac
LEFT JOIN monthly_transactions mt ON 
    ac.activity_month = mt.activity_month
    AND ac.payment_method = mt.payment_method
ORDER BY 
    ac.activity_month, 
    ac.payment_method
`

const getMonthlyTransactionMethodsSuccess = `WITH
    date_ranges AS (
        SELECT
            $1::timestamp AS range1_start,
            $2::timestamp AS range1_end,
            $3::timestamp AS range2_start,
            $4::timestamp AS range2_end
    ),
    payment_methods AS (
        SELECT DISTINCT
            payment_method
        FROM transactions
        WHERE deleted_at IS NULL
    ),
    all_months AS (
        SELECT generate_series(
            date_trunc('month', LEAST(
                (SELECT range1_start FROM date_ranges),
                (SELECT range2_start FROM date_ranges)
            )),
            date_trunc('month', GREATEST(
                (SELECT range1_end FROM date_ranges),
                (SELECT range2_end FROM date_ranges)
            )),
            interval '1 month'
        )::date AS activity_month
    ),
    all_combinations AS (
        SELECT 
            am.activity_month,
            pm.payment_method
        FROM all_months am
        CROSS JOIN payment_methods pm
    ),
    monthly_transactions AS (
        SELECT
            date_trunc('month', t.created_at)::date AS activity_month,
            t.payment_method,
            COUNT(t.transaction_id) AS total_transactions,
            COALESCE(SUM(t.amount), 0)::NUMERIC AS total_amount
        FROM transactions t
        JOIN date_ranges dr ON (
            t.created_at BETWEEN dr.range1_start AND dr.range1_end
            OR t.created_at BETWEEN dr.range2_start AND dr.range2_end
        )
        WHERE
            t.deleted_at IS NULL
            AND t.payment_status = 'success'
        GROUP BY
            date_trunc('month', t.created_at),
            t.payment_method
    )
SELECT 
    TO_CHAR(ac.activity_month, 'Mon') AS month,
    ac.payment_method,
    COALESCE(mt.total_transactions, 0) AS total_transactions,
    COALESCE(mt.total_amount, 0) AS total_amount
FROM all_combinations ac
LEFT JOIN monthly_transactions mt ON 
    ac.activity_month = mt.activity_month
    AND ac.payment_method = mt.payment_method
ORDER BY 
    ac.activity_month, 
    ac.payment_method
`

const getTransactionByID = `SELECT transaction_id, order_id, merchant_id, payment_method, amount, change_amount, payment_status, created_at, updated_at, deleted_at
FROM transactions
WHERE
    transaction_id = $1
    AND deleted_at IS NULL
`

const getTransactionByMerchant = `SELECT transaction_id, order_id, merchant_id, payment_method, amount, change_amount, payment_status, created_at, updated_at, deleted_at, COUNT(*) OVER () AS total_count
FROM transactions
WHERE
    deleted_at IS NULL
    AND (
        $1::TEXT IS NULL
        OR payment_method ILIKE '%' || $1 || '%'
        OR payment_status ILIKE '%' || $1 || '%'
    )
    AND (
        $2::INT IS NULL
        OR merchant_id = $2
    )
ORDER BY created_at DESC
LIMIT $3
OFFSET
    $4
`

const getTransactionByOrderID = `SELECT transaction_id, order_id, merchant_id, payment_method, amount, change_amount, payment_status, created_at, updated_at, deleted_at
FROM transactions
WHERE
    order_id = $1
    AND deleted_at IS NULL
`

const getTransactions = `SELECT transaction_id, order_id, merchant_id, payment_method, amount, change_amount, payment_status, created_at, updated_at, deleted_at, COUNT(*) OVER () AS total_count
FROM transactions
WHERE
    deleted_at IS NULL
    AND (
        $1::TEXT IS NULL
        OR payment_method ILIKE '%' || $1 || '%'
        OR payment_status ILIKE '%' || $1 || '%'
    )
ORDER BY created_at DESC
LIMIT $2
OFFSET
    $3
`

const getTransactionsActive = `SELECT transaction_id, order_id, merchant_id, payment_method, amount, change_amount, payment_status, created_at, updated_at, deleted_at, COUNT(*) OVER () AS total_count
FROM transactions
WHERE
    deleted_at IS NULL
    AND (
        $1::TEXT IS NULL
        OR payment_method ILIKE '%' || $1 || '%'
        OR payment_status ILIKE '%' || $1 || '%'
    )
ORDER BY created_at DESC
LIMIT $2
OFFSET
    $3
`

const getTransactionsTrashed = `SELECT transaction_id, order_id, merchant_id, payment_method, amount, change_amount, payment_status, created_at, updated_at, deleted_at, COUNT(*) OVER () AS total_count
FROM transactions
WHERE
    deleted_at IS NOT NULL
    AND (
        $1::TEXT IS NULL
        OR payment_method ILIKE '%' || $1 || '%'
        OR payment_status ILIKE '%' || $1 || '%'
    )
ORDER BY created_at DESC
LIMIT $2
OFFSET
    $3
`

const getYearlyAmountTransactionFailed = `WITH
    yearly_data AS (
        SELECT
            EXTRACT(
                YEAR
                FROM t.created_at
            )::integer AS year,
            COUNT(*) AS total_failed,
            COALESCE(SUM(t.amount), 0)::integer AS total_amount
        FROM transactions t
        WHERE
            t.deleted_at IS NULL
            AND t.payment_status = 'failed'
            AND (
                EXTRACT(
                    YEAR
                    FROM t.created_at
                ) = $1::integer
                OR EXTRACT(
                    YEAR
                    FROM t.created_at
                ) = $1::integer - 1
            )
        GROUP BY
            EXTRACT(
                YEAR
                FROM t.created_at
            )
    ),
    formatted_data AS (
        SELECT
            year::text,
            total_failed::integer,
            total_amount::integer
        FROM yearly_data
        UNION ALL
        SELECT
            $1::text AS year,
            0::integer AS total_failed,
            0::integer AS total_amount
        WHERE
            NOT EXISTS (
                SELECT 1
                FROM yearly_data
                WHERE
                    year = $1::integer
            )
        UNION ALL
        SELECT ($1::integer - 1)::text AS year,
            0::integer AS total_failed,
            0::integer AS total_amount
        WHERE
            NOT EXISTS (
                SELECT 1
                FROM yearly_data
                WHERE
                    year = $1::integer - 1
            )
    )
SELECT year, total_failed, total_amount
FROM formatted_data
ORDER BY year DESC
`

const getYearlyAmountTransactionFailedByMerchant = `
WITH
    yearly_data AS (
        SELECT
            EXTRACT(
                YEAR
                FROM t.created_at
            )::integer AS year,
            COUNT(*) AS total_failed,
            COALESCE(SUM(t.amount), 0)::integer AS total_amount
        FROM transactions t
        WHERE
            t.deleted_at IS NULL
            AND t.payment_status = 'failed'
            AND t.merchant_id = $2
            AND (
                EXTRACT(
                    YEAR
                    FROM t.created_at
                ) = $1::integer
                OR EXTRACT(
                    YEAR
                    FROM t.created_at
                ) = $1::integer - 1
            )
        GROUP BY
            EXTRACT(
                YEAR
                FROM t.created_at
            )
    ),
    formatted_data AS (
        SELECT
            year::text,
            total_failed::integer,
            total_amount::integer
        FROM yearly_data
        UNION ALL
        SELECT
            $1::text AS year,
            0::integer AS total_failed,
            0::integer AS total_amount
        WHERE
            NOT EXISTS (
                SELECT 1
                FROM yearly_data
                WHERE
                    year = $1::integer
            )
        UNION ALL
        SELECT ($1::integer - 1)::text AS year,
            0::integer AS total_failed,
            0::integer AS total_amount
        WHERE
            NOT EXISTS (
                SELECT 1
                FROM yearly_data
                WHERE
                    year = $1::integer - 1
            )
    )
SELECT year, total_failed, total_amount
FROM formatted_data
ORDER BY year DESC
`

const getYearlyAmountTransactionSuccess = `WITH
    yearly_data AS (
        SELECT
            EXTRACT(
                YEAR
                FROM t.created_at
            )::integer AS year,
            COUNT(*) AS total_success,
            COALESCE(SUM(t.amount), 0)::integer AS total_amount
        FROM transactions t
        WHERE
            t.deleted_at IS NULL
            AND t.payment_status = 'success'
            AND (
                EXTRACT(
                    YEAR
                    FROM t.created_at
                ) = $1::integer
                OR EXTRACT(
                    YEAR
                    FROM t.created_at
                ) = $1::integer - 1
            )
        GROUP BY
            EXTRACT(
                YEAR
                FROM t.created_at
            )
    ),
    formatted_data AS (
        SELECT
            year::text,
            total_success::integer,
            total_amount::integer
        FROM yearly_data
        UNION ALL
        SELECT
            $1::text AS year,
            0::integer AS total_success,
            0::integer AS total_amount
        WHERE
            NOT EXISTS (
                SELECT 1
                FROM yearly_data
                WHERE
                    year = $1::integer
            )
        UNION ALL
        SELECT ($1::integer - 1)::text AS year,
            0::integer AS total_success,
            0::integer AS total_amount
        WHERE
            NOT EXISTS (
                SELECT 1
                FROM yearly_data
                WHERE
                    year = $1::integer - 1
            )
    )
SELECT year, total_success, total_amount
FROM formatted_data
ORDER BY year DESC
`

const getYearlyAmountTransactionSuccessByMerchant = `WITH
    yearly_data AS (
        SELECT
            EXTRACT(
                YEAR
                FROM t.created_at
            )::integer AS year,
            COUNT(*) AS total_success,
            COALESCE(SUM(t.amount), 0)::integer AS total_amount
        FROM transactions t
        WHERE
            t.deleted_at IS NULL
            AND t.payment_status = 'success'
            AND t.merchant_id = $2
            AND (
                EXTRACT(
                    YEAR
                    FROM t.created_at
                ) = $1::integer
                OR EXTRACT(
                    YEAR
                    FROM t.created_at
                ) = $1::integer - 1
            )
        GROUP BY
            EXTRACT(
                YEAR
                FROM t.created_at
            )
    ),
    formatted_data AS (
        SELECT
            year::text,
            total_success::integer,
            total_amount::integer
        FROM yearly_data
        UNION ALL
        SELECT
            $1::text AS year,
            0::integer AS total_success,
            0::integer AS total_amount
        WHERE
            NOT EXISTS (
                SELECT 1
                FROM yearly_data
                WHERE
                    year = $1::integer
            )
        UNION ALL
        SELECT ($1::integer - 1)::text AS year,
            0::integer AS total_success,
            0::integer AS total_amount
        WHERE
            NOT EXISTS (
                SELECT 1
                FROM yearly_data
                WHERE
                    year = $1::integer - 1
            )
    )
SELECT year, total_success, total_amount
FROM formatted_data
ORDER BY year DESC
`

const getYearlyTransactionMethodsByMerchantFailed = `WITH
    year_series AS (
        SELECT generate_series(
            EXTRACT(YEAR FROM $1::timestamp)::integer - 1,
            EXTRACT(YEAR FROM $1::timestamp)::integer,
            1
        ) AS year
    ),
    yearly_transactions AS (
        SELECT
            EXTRACT(YEAR FROM t.created_at)::integer AS year,
            t.payment_method,
            COUNT(t.transaction_id) AS total_transactions,
            SUM(t.amount)::NUMERIC AS total_amount
        FROM transactions t
        WHERE
            t.deleted_at IS NULL
            AND t.payment_status = 'failed'
            AND t.merchant_id = $2
            AND EXTRACT(YEAR FROM t.created_at) BETWEEN (EXTRACT(YEAR FROM $1::timestamp) - 1) AND EXTRACT(YEAR FROM $1::timestamp)
        GROUP BY
            year,
            t.payment_method
    ),
    payment_methods AS (
        SELECT DISTINCT payment_method
        FROM transactions
        WHERE deleted_at IS NULL
    )
SELECT
    ys.year::text AS year,
    pm.payment_method,
    COALESCE(yt.total_transactions, 0) AS total_transactions,
    COALESCE(yt.total_amount, 0) AS total_amount
FROM year_series ys
CROSS JOIN payment_methods pm
LEFT JOIN yearly_transactions yt
    ON ys.year = yt.year
    AND pm.payment_method = yt.payment_method
ORDER BY ys.year, pm.payment_method
`

const getYearlyTransactionMethodsByMerchantSuccess = `WITH
    year_series AS (
        SELECT generate_series(
            EXTRACT(YEAR FROM $1::timestamp)::integer - 2,
            EXTRACT(YEAR FROM $1::timestamp)::integer,
            1
        ) AS year
    ),
    yearly_transactions AS (
        SELECT
            EXTRACT(YEAR FROM t.created_at)::integer AS year,
            t.payment_method,
            COUNT(t.transaction_id) AS total_transactions,
            SUM(t.amount)::NUMERIC AS total_amount
        FROM transactions t
        WHERE
            t.deleted_at IS NULL
            AND t.payment_status = 'success'
            AND t.merchant_id = $2
            AND EXTRACT(YEAR FROM t.created_at) BETWEEN (EXTRACT(YEAR FROM $1::timestamp) - 1) AND EXTRACT(YEAR FROM $1::timestamp)
        GROUP BY
            year,
            t.payment_method
    ),
    payment_methods AS (
        SELECT DISTINCT payment_method
        FROM transactions
        WHERE deleted_at IS NULL
    )
SELECT
    ys.year::text AS year,
    pm.payment_method,
    COALESCE(yt.total_transactions, 0) AS total_transactions,
    COALESCE(yt.total_amount, 0) AS total_amount
FROM year_series ys
CROSS JOIN payment_methods pm
LEFT JOIN yearly_transactions yt
    ON ys.year = yt.year
    AND pm.payment_method = yt.payment_method
ORDER BY ys.year, pm.payment_method
`

const getYearlyTransactionMethodsFailed = `WITH
    year_range AS (
        SELECT 
            EXTRACT(YEAR FROM $1::timestamp)::int - 1 AS start_year,
            EXTRACT(YEAR FROM $1::timestamp)::int AS end_year
    ),
    payment_methods AS (
        SELECT DISTINCT
            payment_method
        FROM transactions
        WHERE deleted_at IS NULL
    ),
    all_years AS (
        SELECT generate_series(
            (SELECT start_year FROM year_range),
            (SELECT end_year FROM year_range)
        )::int AS year
    ),
    all_combinations AS (
        SELECT 
            ay.year::text AS year,  
            pm.payment_method
        FROM all_years ay
        CROSS JOIN payment_methods pm
    ),
    yearly_transactions AS (
        SELECT
            EXTRACT(YEAR FROM t.created_at)::text AS year,
            t.payment_method,
            COUNT(t.transaction_id) AS total_transactions,
            COALESCE(SUM(t.amount), 0)::NUMERIC AS total_amount
        FROM transactions t
        WHERE
            t.deleted_at IS NULL
            AND t.payment_status = 'failed'
            AND EXTRACT(YEAR FROM t.created_at) BETWEEN (SELECT start_year FROM year_range) AND (SELECT end_year FROM year_range)
        GROUP BY
            EXTRACT(YEAR FROM t.created_at),
            t.payment_method
    )
SELECT 
    ac.year, 
    ac.payment_method,
    COALESCE(yt.total_transactions, 0) AS total_transactions,
    COALESCE(yt.total_amount, 0) AS total_amount
FROM all_combinations ac
LEFT JOIN yearly_transactions yt ON 
    ac.year = yt.year
    AND ac.payment_method = yt.payment_method
ORDER BY 
    ac.year,  
    ac.payment_method
`

const getYearlyTransactionMethodsSuccess = `WITH
    year_range AS (
        SELECT 
            EXTRACT(YEAR FROM $1::timestamp)::int - 1 AS start_year,
            EXTRACT(YEAR FROM $1::timestamp)::int AS end_year
    ),
    payment_methods AS (
        SELECT DISTINCT
            payment_method
        FROM transactions
        WHERE deleted_at IS NULL
    ),
    all_years AS (
        SELECT generate_series(
            (SELECT start_year FROM year_range),
            (SELECT end_year FROM year_range)
        )::int AS year
    ),
    all_combinations AS (
        SELECT 
            ay.year::text AS year,  
            pm.payment_method
        FROM all_years ay
        CROSS JOIN payment_methods pm
    ),
    yearly_transactions AS (
        SELECT
            EXTRACT(YEAR FROM t.created_at)::text AS year,
            t.payment_method,
            COUNT(t.transaction_id) AS total_transactions,
            COALESCE(SUM(t.amount), 0)::NUMERIC AS total_amount
        FROM transactions t
        WHERE
            t.deleted_at IS NULL
            AND t.payment_status = 'success'
            AND EXTRACT(YEAR FROM t.created_at) BETWEEN (SELECT start_year FROM year_range) AND (SELECT end_year FROM year_range)
        GROUP BY
            EXTRACT(YEAR FROM t.created_at),
            t.payment_method
    )
SELECT 
    ac.year,  
    ac.payment_method,
    COALESCE(yt.total_transactions, 0) AS total_transactions,
    COALESCE(yt.total_amount, 0) AS total_amount
FROM all_combinations ac
LEFT JOIN yearly_transactions yt ON 
    ac.year = yt.year
    AND ac.payment_method = yt.payment_method
ORDER BY 
    ac.year,
    ac.payment_method
`

type transactionQueryRepository struct {
	db *sqlx.DB
}

func NewTransactionQueryRepository(db *sqlx.DB) *transactionQueryRepository {
	return &transactionQueryRepository{
		db: db,
	}
}

func (r *transactionQueryRepository) FindAllTransactions(ctx context.Context, req *requests.FindAllTransaction) ([]*db.GetTransactionsRow, *int, error) {
	offset := (req.Page - 1) * req.PageSize

	var res []*db.GetTransactionsRow
	err := sqlx.SelectContext(ctx, r.db, &res, getTransactions,
		req.Search,
		int32(req.PageSize),
		int32(offset),
	)
	if err != nil {
		return nil, nil, transaction_errors.ErrFindAllTransactions
	}

	var totalCount int
	if len(res) > 0 {
		totalCount = int(res[0].TotalCount)
	} else {
		totalCount = 0
	}

	return res, &totalCount, nil
}

func (r *transactionQueryRepository) FindByActive(ctx context.Context, req *requests.FindAllTransaction) ([]*db.GetTransactionsActiveRow, *int, error) {
	offset := (req.Page - 1) * req.PageSize

	var res []*db.GetTransactionsActiveRow
	err := sqlx.SelectContext(ctx, r.db, &res, getTransactionsActive,
		req.Search,
		int32(req.PageSize),
		int32(offset),
	)
	if err != nil {
		return nil, nil, transaction_errors.ErrFindByActive
	}

	var totalCount int
	if len(res) > 0 {
		totalCount = int(res[0].TotalCount)
	} else {
		totalCount = 0
	}

	return res, &totalCount, nil
}

func (r *transactionQueryRepository) FindByTrashed(ctx context.Context, req *requests.FindAllTransaction) ([]*db.GetTransactionsTrashedRow, *int, error) {
	offset := (req.Page - 1) * req.PageSize

	var res []*db.GetTransactionsTrashedRow
	err := sqlx.SelectContext(ctx, r.db, &res, getTransactionsTrashed,
		req.Search,
		int32(req.PageSize),
		int32(offset),
	)
	if err != nil {
		return nil, nil, transaction_errors.ErrFindByTrashed
	}

	var totalCount int
	if len(res) > 0 {
		totalCount = int(res[0].TotalCount)
	} else {
		totalCount = 0
	}

	return res, &totalCount, nil
}

func (r *transactionQueryRepository) FindByMerchant(
	ctx context.Context,
	req *requests.FindAllTransactionByMerchant,
) ([]*db.GetTransactionByMerchantRow, *int, error) {
	offset := (req.Page - 1) * req.PageSize

	var res []*db.GetTransactionByMerchantRow
	err := sqlx.SelectContext(ctx, r.db, &res, getTransactionByMerchant,
		req.Search,
		int32(req.MerchantID),
		int32(req.PageSize),
		int32(offset),
	)
	if err != nil {
		return nil, nil, transaction_errors.ErrFindByMerchant
	}

	var totalCount int
	if len(res) > 0 {
		totalCount = int(res[0].TotalCount)
	} else {
		totalCount = 0
	}

	return res, &totalCount, nil
}

func (r *transactionQueryRepository) FindById(ctx context.Context, transaction_id int) (*db.Transaction, error) {
	var transaction db.Transaction
	err := sqlx.GetContext(ctx, r.db, &transaction, getTransactionByID, int32(transaction_id))
	if err != nil {
		return nil, transaction_errors.ErrFindById
	}

	return &transaction, nil
}

func (r *transactionQueryRepository) FindByOrderId(ctx context.Context, order_id int) (*db.Transaction, error) {
	var transaction db.Transaction
	err := sqlx.GetContext(ctx, r.db, &transaction, getTransactionByOrderID, int32(order_id))
	if err != nil {
		return nil, transaction_errors.ErrFindByOrderId
	}

	return &transaction, nil
}
