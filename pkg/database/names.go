package database

// Logical database names for the five bounded contexts. Each context runs on
// its own PostgreSQL instance; the database keeps the context name so the same
// identifier works in env keys, migrations and dashboards. A service selects
// its own through the cluster prefix (see the Cluster* constants) so that one
// table has exactly one owner.
const (
	IdentityDB = "pos_identity"
	CatalogDB  = "pos_catalog"
	SalesDB    = "pos_sales"
	MerchantDB = "pos_merchant"
	EmailDB    = "pos_email"
)

// Cluster* are the env prefixes passed to NewClientWithPrefix. Each prefix
// resolves its own <Prefix>_HOST, _PORT, _USERNAME, _NAME and _PASSWORD — one
// PostgreSQL instance per bounded context, fronted by its own PgBouncer. Only
// the username and password fall back to the generic DB_* keys.
const (
	IdentityCluster = "DB_IDENTITY"
	CatalogCluster  = "DB_CATALOG"
	SalesCluster    = "DB_SALES"
	MerchantCluster = "DB_MERCHANT"
	EmailCluster    = "DB_EMAIL"
)
