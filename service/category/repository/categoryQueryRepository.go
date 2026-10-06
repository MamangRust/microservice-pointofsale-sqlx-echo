package repository

import (
	"context"

	db "github.com/MamangRust/microservice-point-of-sale-category/database/schema"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
	"github.com/MamangRust/microservice-point-of-sale-shared/errors/category_errors"
	"github.com/jmoiron/sqlx"
)

const getCategories = `SELECT
    *,
    COUNT(*) OVER() AS total_count
FROM categories
WHERE deleted_at IS NULL
AND ($1::TEXT IS NULL OR name ILIKE '%' || $1 || '%' OR slug_category ILIKE '%' || $1 || '%')
ORDER BY created_at DESC
LIMIT $2 OFFSET $3`

const getCategoriesActive = `SELECT
    *,
    COUNT(*) OVER() AS total_count
FROM categories
WHERE deleted_at IS NULL
AND ($1::TEXT IS NULL OR name ILIKE '%' || $1 || '%' OR slug_category ILIKE '%' || $1 || '%')
ORDER BY created_at DESC
LIMIT $2 OFFSET $3`

const getCategoriesTrashed = `SELECT
    *,
    COUNT(*) OVER() AS total_count
FROM categories
WHERE deleted_at IS NOT NULL
AND ($1::TEXT IS NULL OR name ILIKE '%' || $1 || '%' OR slug_category ILIKE '%' || $1 || '%')
ORDER BY created_at DESC
LIMIT $2 OFFSET $3`

const getMonthlyTotalPrice = `WITH monthly_totals AS (
    SELECT
        EXTRACT(YEAR FROM o.created_at)::TEXT AS year,
        EXTRACT(MONTH FROM o.created_at)::integer AS month,
        COALESCE(SUM(o.total_price), 0)::INTEGER AS total_revenue
    FROM
        orders o
    JOIN
        order_items oi ON o.order_id = oi.order_id
    JOIN
        products p ON oi.product_id = p.product_id
    JOIN
        categories c ON p.category_id = c.category_id
    WHERE
        o.deleted_at IS NULL
        AND oi.deleted_at IS NULL
        AND (
            (o.created_at >= $1 AND o.created_at <= $2)
            OR (o.created_at >= $3 AND o.created_at <= $4)
        )
    GROUP BY
        EXTRACT(YEAR FROM o.created_at),
        EXTRACT(MONTH FROM o.created_at)
),
all_months AS (
    SELECT
        EXTRACT(YEAR FROM $1)::TEXT AS year,
        EXTRACT(MONTH FROM $1)::integer AS month,
        TO_CHAR($1, 'FMMonth') AS month_name

    UNION

    SELECT
        EXTRACT(YEAR FROM $3)::TEXT AS year,
        EXTRACT(MONTH FROM $3)::integer AS month,
        TO_CHAR($3, 'FMMonth') AS month_name
)
SELECT
    COALESCE(am.year, EXTRACT(YEAR FROM $1)::TEXT) AS year,
    COALESCE(am.month_name, TO_CHAR($1, 'FMMonth')) AS month,
    COALESCE(mt.total_revenue, 0) AS total_revenue
FROM
    all_months am
LEFT JOIN
    monthly_totals mt ON am.year = mt.year AND am.month = mt.month
ORDER BY
    am.year::INT DESC,
    am.month DESC`

const getYearlyTotalPrice = `WITH yearly_data AS (
    SELECT
        EXTRACT(YEAR FROM o.created_at)::integer AS year,
        COALESCE(SUM(o.total_price), 0)::INTEGER AS total_revenue
    FROM
        orders o
    JOIN
        order_items oi ON o.order_id = oi.order_id
    JOIN
        products p ON oi.product_id = p.product_id
    JOIN
        categories c ON p.category_id = c.category_id
    WHERE
        o.deleted_at IS NULL
        AND oi.deleted_at IS NULL
        AND p.deleted_at IS NULL
        AND c.deleted_at IS NULL
        AND (
            EXTRACT(YEAR FROM o.created_at) = $1::integer
            OR EXTRACT(YEAR FROM o.created_at) = $1::integer - 1
        )
    GROUP BY
        EXTRACT(YEAR FROM o.created_at)
),
all_years AS (
    SELECT $1 AS year
    UNION
    SELECT $1 - 1 AS year
)
SELECT
    a.year::text AS year,
    COALESCE(yd.total_revenue, 0) AS total_revenue
FROM
    all_years a
LEFT JOIN
    yearly_data yd ON a.year = yd.year
ORDER BY
    a.year DESC`

const getMonthlyTotalPriceByMerchant = `WITH monthly_totals AS (
    SELECT
        EXTRACT(YEAR FROM o.created_at)::TEXT AS year,
        EXTRACT(MONTH FROM o.created_at)::integer AS month,
        COALESCE(SUM(o.total_price), 0)::INTEGER AS total_revenue
    FROM
        orders o
    JOIN
        order_items oi ON o.order_id = oi.order_id
    JOIN
        products p ON oi.product_id = p.product_id
    JOIN
        categories c ON p.category_id = c.category_id
    WHERE
        o.deleted_at IS NULL
        AND oi.deleted_at IS NULL
        AND (
            (o.created_at >= $1 AND o.created_at <= $2)
            OR (o.created_at >= $3 AND o.created_at <= $4)
        )
        AND o.merchant_id = $5
    GROUP BY
        EXTRACT(YEAR FROM o.created_at),
        EXTRACT(MONTH FROM o.created_at)
),
all_months AS (
    SELECT
        EXTRACT(YEAR FROM $1)::TEXT AS year,
        EXTRACT(MONTH FROM $1)::integer AS month,
        TO_CHAR($1, 'FMMonth') AS month_name

    UNION

    SELECT
        EXTRACT(YEAR FROM $3)::TEXT AS year,
        EXTRACT(MONTH FROM $3)::integer AS month,
        TO_CHAR($3, 'FMMonth') AS month_name
)
SELECT
    COALESCE(am.year, EXTRACT(YEAR FROM $1)::TEXT) AS year,
    COALESCE(am.month_name, TO_CHAR($1, 'FMMonth')) AS month,
    COALESCE(mt.total_revenue, 0) AS total_revenue
FROM
    all_months am
LEFT JOIN
    monthly_totals mt ON am.year = mt.year AND am.month = mt.month
ORDER BY
    am.year::INT DESC,
    am.month DESC`

const getYearlyTotalPriceByMerchant = `WITH yearly_data AS (
    SELECT
        EXTRACT(YEAR FROM o.created_at)::integer AS year,
        COALESCE(SUM(o.total_price), 0)::INTEGER AS total_revenue
    FROM
        orders o
    JOIN
        order_items oi ON o.order_id = oi.order_id
    JOIN
        products p ON oi.product_id = p.product_id
    JOIN
        categories c ON p.category_id = c.category_id
    WHERE
        o.deleted_at IS NULL
        AND oi.deleted_at IS NULL
        AND p.deleted_at IS NULL
        AND c.deleted_at IS NULL
        AND (
            EXTRACT(YEAR FROM o.created_at) = $1::integer
            OR EXTRACT(YEAR FROM o.created_at) = $1::integer - 1
        )
        AND o.merchant_id = $2
    GROUP BY
        EXTRACT(YEAR FROM o.created_at)
),
all_years AS (
    SELECT $1 AS year
    UNION
    SELECT $1 - 1 AS year
)
SELECT
    a.year::text AS year,
    COALESCE(yd.total_revenue, 0) AS total_revenue
FROM
    all_years a
LEFT JOIN
    yearly_data yd ON a.year = yd.year
ORDER BY
    a.year DESC`

const getMonthlyTotalPriceById = `WITH monthly_totals AS (
    SELECT
        EXTRACT(YEAR FROM o.created_at)::TEXT AS year,
        EXTRACT(MONTH FROM o.created_at)::integer AS month,
        COALESCE(SUM(o.total_price), 0)::INTEGER AS total_revenue
    FROM
        orders o
    JOIN
        order_items oi ON o.order_id = oi.order_id
    JOIN
        products p ON oi.product_id = p.product_id
    JOIN
        categories c ON p.category_id = c.category_id
    WHERE
        o.deleted_at IS NULL
        AND oi.deleted_at IS NULL
        AND (
            (o.created_at >= $1 AND o.created_at <= $2)
            OR (o.created_at >= $3 AND o.created_at <= $4)
        )
        AND c.category_id = $5
    GROUP BY
        EXTRACT(YEAR FROM o.created_at),
        EXTRACT(MONTH FROM o.created_at)
),
all_months AS (
    SELECT
        EXTRACT(YEAR FROM $1)::TEXT AS year,
        EXTRACT(MONTH FROM $1)::integer AS month,
        TO_CHAR($1, 'FMMonth') AS month_name

    UNION

    SELECT
        EXTRACT(YEAR FROM $3)::TEXT AS year,
        EXTRACT(MONTH FROM $3)::integer AS month,
        TO_CHAR($3, 'FMMonth') AS month_name
)
SELECT
    COALESCE(am.year, EXTRACT(YEAR FROM $1)::TEXT) AS year,
    COALESCE(am.month_name, TO_CHAR($1, 'FMMonth')) AS month,
    COALESCE(mt.total_revenue, 0) AS total_revenue
FROM
    all_months am
LEFT JOIN
    monthly_totals mt ON am.year = mt.year AND am.month = mt.month
ORDER BY
    am.year::INT DESC,
    am.month DESC`

const getYearlyTotalPriceById = `WITH yearly_data AS (
    SELECT
        EXTRACT(YEAR FROM o.created_at)::integer AS year,
        COALESCE(SUM(o.total_price), 0)::INTEGER AS total_revenue
    FROM
        orders o
    JOIN
        order_items oi ON o.order_id = oi.order_id
    JOIN
        products p ON oi.product_id = p.product_id
    JOIN
        categories c ON p.category_id = c.category_id
    WHERE
        o.deleted_at IS NULL
        AND oi.deleted_at IS NULL
        AND p.deleted_at IS NULL
        AND c.deleted_at IS NULL
        AND (
            EXTRACT(YEAR FROM o.created_at) = $1::integer
            OR EXTRACT(YEAR FROM o.created_at) = $1::integer - 1
        )
        AND c.category_id = $2
    GROUP BY
        EXTRACT(YEAR FROM o.created_at)
),
all_years AS (
    SELECT $1 AS year
    UNION
    SELECT $1 - 1 AS year
)
SELECT
    a.year::text AS year,
    COALESCE(yd.total_revenue, 0) AS total_revenue
FROM
    all_years a
LEFT JOIN
    yearly_data yd ON a.year = yd.year
ORDER BY
    a.year DESC`

const getMonthlyCategory = `WITH date_range AS (
    SELECT
        date_trunc('month', $1::timestamp) AS start_date,
        date_trunc('month', $1::timestamp) + interval '1 year' - interval '1 day' AS end_date
),
monthly_category_stats AS (
    SELECT
        c.category_id,
        c.name AS category_name,
        date_trunc('month', o.created_at) AS activity_month,
        COUNT(DISTINCT o.order_id) AS order_count,
        SUM(oi.quantity) AS items_sold,
        COALESCE(SUM(o.total_price), 0)::INTEGER AS total_revenue
    FROM
        orders o
    JOIN
        order_items oi ON o.order_id = oi.order_id
    JOIN
        products p ON oi.product_id = p.product_id
    JOIN
        categories c ON p.category_id = c.category_id
    WHERE
        o.deleted_at IS NULL
        AND oi.deleted_at IS NULL
        AND p.deleted_at IS NULL
        AND c.deleted_at IS NULL
        AND o.created_at BETWEEN (SELECT start_date FROM date_range)
                             AND (SELECT end_date FROM date_range)
    GROUP BY
        c.category_id, c.name, activity_month
)
SELECT
    TO_CHAR(mcs.activity_month, 'Mon') AS month,
    mcs.category_id,
    mcs.category_name,
    mcs.order_count,
    mcs.items_sold,
    mcs.total_revenue
FROM
    monthly_category_stats mcs
ORDER BY
    mcs.activity_month, mcs.total_revenue DESC`

const getYearlyCategory = `WITH last_five_years AS (
    SELECT
        c.category_id,
        c.name AS category_name,
        EXTRACT(YEAR FROM o.created_at)::text AS year,
        COUNT(DISTINCT o.order_id) AS order_count,
        SUM(oi.quantity) AS items_sold,
        COALESCE(SUM(o.total_price), 0)::INTEGER AS total_revenue,
        COUNT(DISTINCT oi.product_id) AS unique_products_sold
    FROM
        orders o
    JOIN
        order_items oi ON o.order_id = oi.order_id
    JOIN
        products p ON oi.product_id = p.product_id
    JOIN
        categories c ON p.category_id = c.category_id
    WHERE
        o.deleted_at IS NULL
        AND oi.deleted_at IS NULL
        AND p.deleted_at IS NULL
        AND c.deleted_at IS NULL
        AND EXTRACT(YEAR FROM o.created_at) BETWEEN (EXTRACT(YEAR FROM $1::timestamp) - 4) AND EXTRACT(YEAR FROM $1::timestamp)
    GROUP BY
        c.category_id, c.name, EXTRACT(YEAR FROM o.created_at)
)
SELECT
    year,
    category_id,
    category_name,
    order_count,
    items_sold,
    total_revenue,
    unique_products_sold
FROM
    last_five_years
ORDER BY
    year, total_revenue DESC`

const getMonthlyCategoryByMerchant = `WITH date_range AS (
    SELECT
        date_trunc('month', $1::timestamp) AS start_date,
        date_trunc('month', $1::timestamp) + interval '1 year' - interval '1 day' AS end_date
),
monthly_category_stats AS (
    SELECT
        c.category_id,
        c.name AS category_name,
        date_trunc('month', o.created_at) AS activity_month,
        COUNT(DISTINCT o.order_id) AS order_count,
        SUM(oi.quantity) AS items_sold,
        COALESCE(SUM(o.total_price), 0)::INTEGER AS total_revenue
    FROM
        orders o
    JOIN
        order_items oi ON o.order_id = oi.order_id
    JOIN
        products p ON oi.product_id = p.product_id
    JOIN
        categories c ON p.category_id = c.category_id
    WHERE
        o.deleted_at IS NULL
        AND oi.deleted_at IS NULL
        AND p.deleted_at IS NULL
        AND c.deleted_at IS NULL
        AND o.created_at BETWEEN (SELECT start_date FROM date_range)
                             AND (SELECT end_date FROM date_range)
        AND o.merchant_id = $2
    GROUP BY
        c.category_id, c.name, activity_month
)
SELECT
    TO_CHAR(mcs.activity_month, 'Mon') AS month,
    mcs.category_id,
    mcs.category_name,
    mcs.order_count,
    mcs.items_sold,
    mcs.total_revenue
FROM
    monthly_category_stats mcs
ORDER BY
    mcs.activity_month, mcs.total_revenue DESC`

const getYearlyCategoryByMerchant = `WITH last_five_years AS (
    SELECT
        c.category_id,
        c.name AS category_name,
        EXTRACT(YEAR FROM o.created_at)::text AS year,
        COUNT(DISTINCT o.order_id) AS order_count,
        SUM(oi.quantity) AS items_sold,
        COALESCE(SUM(o.total_price), 0)::INTEGER AS total_revenue,
        COUNT(DISTINCT oi.product_id) AS unique_products_sold
    FROM
        orders o
    JOIN
        order_items oi ON o.order_id = oi.order_id
    JOIN
        products p ON oi.product_id = p.product_id
    JOIN
        categories c ON p.category_id = c.category_id
    WHERE
        o.deleted_at IS NULL
        AND oi.deleted_at IS NULL
        AND p.deleted_at IS NULL
        AND c.deleted_at IS NULL
        AND EXTRACT(YEAR FROM o.created_at) BETWEEN (EXTRACT(YEAR FROM $1::timestamp) - 4) AND EXTRACT(YEAR FROM $1::timestamp)
        AND o.merchant_id = $2
    GROUP BY
        c.category_id, c.name, EXTRACT(YEAR FROM o.created_at)
)
SELECT
    year,
    category_id,
    category_name,
    order_count,
    items_sold,
    total_revenue,
    unique_products_sold
FROM
    last_five_years
ORDER BY
    year, total_revenue DESC`

const getMonthlyCategoryById = `WITH date_range AS (
    SELECT
        date_trunc('month', $1::timestamp) AS start_date,
        date_trunc('month', $1::timestamp) + interval '1 year' - interval '1 day' AS end_date
),
monthly_category_stats AS (
    SELECT
        c.category_id,
        c.name AS category_name,
        date_trunc('month', o.created_at) AS activity_month,
        COUNT(DISTINCT o.order_id) AS order_count,
        SUM(oi.quantity) AS items_sold,
        COALESCE(SUM(o.total_price), 0)::INTEGER AS total_revenue
    FROM
        orders o
    JOIN
        order_items oi ON o.order_id = oi.order_id
    JOIN
        products p ON oi.product_id = p.product_id
    JOIN
        categories c ON p.category_id = c.category_id
    WHERE
        o.deleted_at IS NULL
        AND oi.deleted_at IS NULL
        AND p.deleted_at IS NULL
        AND c.deleted_at IS NULL
        AND o.created_at BETWEEN (SELECT start_date FROM date_range)
                             AND (SELECT end_date FROM date_range)
        AND c.category_id = $2
    GROUP BY
        c.category_id, c.name, activity_month
)
SELECT
    TO_CHAR(mcs.activity_month, 'Mon') AS month,
    mcs.category_id,
    mcs.category_name,
    mcs.order_count,
    mcs.items_sold,
    mcs.total_revenue
FROM
    monthly_category_stats mcs
ORDER BY
    mcs.activity_month, mcs.total_revenue DESC`

const getYearlyCategoryById = `WITH last_five_years AS (
    SELECT
        c.category_id,
        c.name AS category_name,
        EXTRACT(YEAR FROM o.created_at)::text AS year,
        COUNT(DISTINCT o.order_id) AS order_count,
        SUM(oi.quantity) AS items_sold,
        COALESCE(SUM(o.total_price), 0)::INTEGER AS total_revenue,
        COUNT(DISTINCT oi.product_id) AS unique_products_sold
    FROM
        orders o
    JOIN
        order_items oi ON o.order_id = oi.order_id
    JOIN
        products p ON oi.product_id = p.product_id
    JOIN
        categories c ON p.category_id = c.category_id
    WHERE
        o.deleted_at IS NULL
        AND oi.deleted_at IS NULL
        AND p.deleted_at IS NULL
        AND c.deleted_at IS NULL
        AND EXTRACT(YEAR FROM o.created_at) BETWEEN (EXTRACT(YEAR FROM $1::timestamp) - 4) AND EXTRACT(YEAR FROM $1::timestamp)
        AND c.category_id = $2
    GROUP BY
        c.category_id, c.name, EXTRACT(YEAR FROM o.created_at)
)
SELECT
    year,
    category_id,
    category_name,
    order_count,
    items_sold,
    total_revenue,
    unique_products_sold
FROM
    last_five_years
ORDER BY
    year, total_revenue DESC`

const getCategoryByID = `SELECT *
FROM categories
WHERE category_id = $1
  AND deleted_at IS NULL`

const getCategoryByIDTrashed = `SELECT *
FROM categories
WHERE category_id = $1
  AND deleted_at IS NOT NULL`

const getCategoryByName = `SELECT *
FROM categories
WHERE name = $1
  AND deleted_at IS NULL`

const getCategoriesByIDs = `SELECT *
FROM categories
WHERE category_id IN (?)
  AND deleted_at IS NULL`

const getCategoryByNameAndId = `SELECT *
FROM categories
WHERE name = $1
  AND category_id = $2
  AND deleted_at IS NULL`

type categoryQueryRepository struct {
	db *sqlx.DB
}

func NewCategoryQueryRepository(db *sqlx.DB) CategoryQueryRepository {
	return &categoryQueryRepository{
		db: db,
	}
}

func (r *categoryQueryRepository) FindAllCategory(ctx context.Context, req *requests.FindAllCategory) ([]*db.GetCategoriesRow, *int, error) {
	offset := (req.Page - 1) * req.PageSize

	var categories []*db.GetCategoriesRow
	err := r.db.SelectContext(ctx, &categories, getCategories,
		req.Search,
		int32(req.PageSize),
		int32(offset),
	)
	if err != nil {
		return nil, nil, category_errors.ErrFindAllCategory
	}

	totalCount := 0
	if len(categories) > 0 {
		totalCount = int(categories[0].TotalCount)
	}

	return categories, &totalCount, nil
}

func (r *categoryQueryRepository) FindByActive(ctx context.Context, req *requests.FindAllCategory) ([]*db.GetCategoriesActiveRow, *int, error) {
	offset := (req.Page - 1) * req.PageSize

	var categories []*db.GetCategoriesActiveRow
	err := r.db.SelectContext(ctx, &categories, getCategoriesActive,
		req.Search,
		int32(req.PageSize),
		int32(offset),
	)
	if err != nil {
		return nil, nil, category_errors.ErrFindByActive
	}

	totalCount := 0
	if len(categories) > 0 {
		totalCount = int(categories[0].TotalCount)
	}

	return categories, &totalCount, nil
}

func (r *categoryQueryRepository) FindByTrashed(ctx context.Context, req *requests.FindAllCategory) ([]*db.GetCategoriesTrashedRow, *int, error) {
	offset := (req.Page - 1) * req.PageSize

	var categories []*db.GetCategoriesTrashedRow
	err := r.db.SelectContext(ctx, &categories, getCategoriesTrashed,
		req.Search,
		int32(req.PageSize),
		int32(offset),
	)
	if err != nil {
		return nil, nil, category_errors.ErrFindByTrashed
	}

	totalCount := 0
	if len(categories) > 0 {
		totalCount = int(categories[0].TotalCount)
	}

	return categories, &totalCount, nil
}

func (r *categoryQueryRepository) FindById(ctx context.Context, category_id int) (*db.Category, error) {
	var category db.Category
	err := r.db.GetContext(ctx, &category, getCategoryByID, int32(category_id))
	if err != nil {
		return nil, category_errors.ErrFindById
	}

	return &category, nil
}

func (r *categoryQueryRepository) FindByIds(ctx context.Context, ids []int) ([]*db.Category, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	intIDs := make([]int32, 0, len(ids))
	for _, id := range ids {
		intIDs = append(intIDs, int32(id))
	}

	query, args, err := sqlx.In(getCategoriesByIDs, intIDs)
	if err != nil {
		return nil, category_errors.ErrFindByIds
	}

	var categories []*db.Category
	if err := r.db.SelectContext(ctx, &categories, r.db.Rebind(query), args...); err != nil {
		return nil, category_errors.ErrFindByIds
	}

	return categories, nil
}

func (r *categoryQueryRepository) FindByIdTrashed(ctx context.Context, category_id int) (*db.Category, error) {	var category db.Category
	err := r.db.GetContext(ctx, &category, getCategoryByIDTrashed, int32(category_id))
	if err != nil {
		return nil, category_errors.ErrFindById
	}

	return &category, nil
}

func (r *categoryQueryRepository) FindByName(ctx context.Context, name string) (*db.Category, error) {
	var category db.Category
	err := r.db.GetContext(ctx, &category, getCategoryByName, name)
	if err != nil {
		return nil, category_errors.ErrFindByName
	}

	return &category, nil
}

func (r *categoryQueryRepository) FindByNameAndId(ctx context.Context, req *requests.CategoryNameAndId) (*db.Category, error) {
	var category db.Category
	err := r.db.GetContext(ctx, &category, getCategoryByNameAndId,
		req.Name,
		int32(req.CategoryID),
	)
	if err != nil {
		return nil, category_errors.ErrFindByNameAndId
	}

	return &category, nil
}
