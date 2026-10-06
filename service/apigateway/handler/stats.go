package handler

import (
	"fmt"
	"net/http"
	"time"

	gateway_cache "github.com/MamangRust/microservice-point-of-sale-apigateway/cache/gateway_cache"
	pbcashier "github.com/MamangRust/microservice-point-of-sale-pb/cashier"
	pbcategory "github.com/MamangRust/microservice-point-of-sale-pb/category"
	pborder "github.com/MamangRust/microservice-point-of-sale-pb/order"
	pbstats "github.com/MamangRust/microservice-point-of-sale-pb/stats"
	pbtransaction "github.com/MamangRust/microservice-point-of-sale-pb/transaction"
	"github.com/MamangRust/microservice-point-of-sale-pkg/logger"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/response"
	"github.com/MamangRust/microservice-point-of-sale-shared/errors"
	response_api "github.com/MamangRust/microservice-point-of-sale-shared/mapper/response/api"
	"github.com/labstack/echo/v4"
)

// StatsClients bundles the stats-reader gRPC clients (main, ByMerchant and
// ById variants) consumed by the gateway stats handlers.
type StatsClients struct {
	Order             pbstats.OrderStatsServiceClient
	OrderMerchant     pbstats.OrderStatsByMerchantServiceClient
	OrderById         pbstats.OrderStatsByIdServiceClient
	Cashier           pbstats.CashierStatsServiceClient
	CashierMerchant   pbstats.CashierStatsByMerchantServiceClient
	CashierById       pbstats.CashierStatsByIdServiceClient
	Category          pbstats.CategoryStatsServiceClient
	CategoryMerchant  pbstats.CategoryStatsByMerchantServiceClient
	CategoryById      pbstats.CategoryStatsByIdServiceClient
	TransactionStatus pbstats.TransactionStatsStatusServiceClient
	TransactionMethod pbstats.TransactionStatsMethodServiceClient
}

// statsHandleApi exposes the stats-reader (ClickHouse) gRPC contracts as REST
// endpoints (F4 §7.4). All cache keys use the "apigw:" namespace so the gateway
// can never overwrite domain-service or stats-reader cache keys (§8.1).
type statsHandleApi struct {
	clients    StatsClients
	logger     logger.LoggerInterface
	mapping    response_api.StatsResponseMapper
	apiHandler errors.ApiHandler
	cache      *gateway_cache.GatewayCache
}

func NewHandlerStats(
	router *echo.Echo,
	clients StatsClients,
	logger logger.LoggerInterface,
	mapping response_api.StatsResponseMapper,
	apiHandler errors.ApiHandler,
	cache *gateway_cache.GatewayCache,
) *statsHandleApi {
	statsHandler := &statsHandleApi{
		clients:    clients,
		logger:     logger,
		mapping:    mapping,
		apiHandler: apiHandler,
		cache:      cache,
	}

	orderStats := router.Group("/api/order-stats")
	orderStats.GET("/monthly-total-revenue", apiHandler.Handle("get-order-stats-monthly-total-revenue", statsHandler.FindMonthlyTotalRevenue))
	orderStats.GET("/yearly-total-revenue", apiHandler.Handle("get-order-stats-yearly-total-revenue", statsHandler.FindYearlyTotalRevenue))
	orderStats.GET("/monthly-revenue", apiHandler.Handle("get-order-stats-monthly-revenue", statsHandler.FindMonthlyRevenue))
	orderStats.GET("/yearly-revenue", apiHandler.Handle("get-order-stats-yearly-revenue", statsHandler.FindYearlyRevenue))
	orderStats.GET("/by-merchant/monthly-total-revenue", apiHandler.Handle("get-order-stats-merchant-monthly-total-revenue", statsHandler.FindMonthlyTotalRevenueByMerchant))
	orderStats.GET("/by-merchant/yearly-total-revenue", apiHandler.Handle("get-order-stats-merchant-yearly-total-revenue", statsHandler.FindYearlyTotalRevenueByMerchant))
	orderStats.GET("/by-merchant/monthly-revenue", apiHandler.Handle("get-order-stats-merchant-monthly-revenue", statsHandler.FindMonthlyRevenueByMerchant))
	orderStats.GET("/by-merchant/yearly-revenue", apiHandler.Handle("get-order-stats-merchant-yearly-revenue", statsHandler.FindYearlyRevenueByMerchant))
	orderStats.GET("/by-id/monthly-total-revenue", apiHandler.Handle("get-order-stats-id-monthly-total-revenue", statsHandler.FindMonthlyTotalRevenueById))
	orderStats.GET("/by-id/yearly-total-revenue", apiHandler.Handle("get-order-stats-id-yearly-total-revenue", statsHandler.FindYearlyTotalRevenueById))

	cashierStats := router.Group("/api/cashier-stats")
	cashierStats.GET("/monthly-total-sales", apiHandler.Handle("get-cashier-stats-monthly-total-sales", statsHandler.FindCashierMonthlyTotalSales))
	cashierStats.GET("/yearly-total-sales", apiHandler.Handle("get-cashier-stats-yearly-total-sales", statsHandler.FindCashierYearlyTotalSales))
	cashierStats.GET("/month-sales", apiHandler.Handle("get-cashier-stats-month-sales", statsHandler.FindCashierMonthSales))
	cashierStats.GET("/year-sales", apiHandler.Handle("get-cashier-stats-year-sales", statsHandler.FindCashierYearSales))
	cashierStats.GET("/by-merchant/monthly-total-sales", apiHandler.Handle("get-cashier-stats-merchant-monthly-total-sales", statsHandler.FindCashierMonthlyTotalSalesByMerchant))
	cashierStats.GET("/by-merchant/yearly-total-sales", apiHandler.Handle("get-cashier-stats-merchant-yearly-total-sales", statsHandler.FindCashierYearlyTotalSalesByMerchant))
	cashierStats.GET("/by-merchant/month-sales", apiHandler.Handle("get-cashier-stats-merchant-month-sales", statsHandler.FindCashierMonthSalesByMerchant))
	cashierStats.GET("/by-merchant/year-sales", apiHandler.Handle("get-cashier-stats-merchant-year-sales", statsHandler.FindCashierYearSalesByMerchant))
	cashierStats.GET("/by-id/monthly-total-sales", apiHandler.Handle("get-cashier-stats-id-monthly-total-sales", statsHandler.FindCashierMonthlyTotalSalesById))
	cashierStats.GET("/by-id/yearly-total-sales", apiHandler.Handle("get-cashier-stats-id-yearly-total-sales", statsHandler.FindCashierYearlyTotalSalesById))
	cashierStats.GET("/by-id/month-sales", apiHandler.Handle("get-cashier-stats-id-month-sales", statsHandler.FindCashierMonthSalesById))
	cashierStats.GET("/by-id/year-sales", apiHandler.Handle("get-cashier-stats-id-year-sales", statsHandler.FindCashierYearSalesById))

	categoryStats := router.Group("/api/category-stats")
	categoryStats.GET("/monthly-total-prices", apiHandler.Handle("get-category-stats-monthly-total-prices", statsHandler.FindCategoryMonthlyTotalPrices))
	categoryStats.GET("/yearly-total-prices", apiHandler.Handle("get-category-stats-yearly-total-prices", statsHandler.FindCategoryYearlyTotalPrices))
	categoryStats.GET("/month-price", apiHandler.Handle("get-category-stats-month-price", statsHandler.FindCategoryMonthPrice))
	categoryStats.GET("/year-price", apiHandler.Handle("get-category-stats-year-price", statsHandler.FindCategoryYearPrice))
	categoryStats.GET("/by-merchant/monthly-total-prices", apiHandler.Handle("get-category-stats-merchant-monthly-total-prices", statsHandler.FindCategoryMonthlyTotalPricesByMerchant))
	categoryStats.GET("/by-merchant/yearly-total-prices", apiHandler.Handle("get-category-stats-merchant-yearly-total-prices", statsHandler.FindCategoryYearlyTotalPricesByMerchant))
	categoryStats.GET("/by-merchant/month-price", apiHandler.Handle("get-category-stats-merchant-month-price", statsHandler.FindCategoryMonthPriceByMerchant))
	categoryStats.GET("/by-merchant/year-price", apiHandler.Handle("get-category-stats-merchant-year-price", statsHandler.FindCategoryYearPriceByMerchant))
	categoryStats.GET("/by-id/monthly-total-prices", apiHandler.Handle("get-category-stats-id-monthly-total-prices", statsHandler.FindCategoryMonthlyTotalPricesById))
	categoryStats.GET("/by-id/yearly-total-prices", apiHandler.Handle("get-category-stats-id-yearly-total-prices", statsHandler.FindCategoryYearlyTotalPricesById))
	categoryStats.GET("/by-id/month-price", apiHandler.Handle("get-category-stats-id-month-price", statsHandler.FindCategoryMonthPriceById))
	categoryStats.GET("/by-id/year-price", apiHandler.Handle("get-category-stats-id-year-price", statsHandler.FindCategoryYearPriceById))

	transactionStats := router.Group("/api/transaction-stats")
	transactionStats.GET("/month-status-success", apiHandler.Handle("get-transaction-stats-month-status-success", statsHandler.FindTransactionMonthStatusSuccess))
	transactionStats.GET("/year-status-success", apiHandler.Handle("get-transaction-stats-year-status-success", statsHandler.FindTransactionYearStatusSuccess))
	transactionStats.GET("/month-status-failed", apiHandler.Handle("get-transaction-stats-month-status-failed", statsHandler.FindTransactionMonthStatusFailed))
	transactionStats.GET("/year-status-failed", apiHandler.Handle("get-transaction-stats-year-status-failed", statsHandler.FindTransactionYearStatusFailed))
	transactionStats.GET("/month-method-success", apiHandler.Handle("get-transaction-stats-month-method-success", statsHandler.FindTransactionMonthMethodSuccess))
	transactionStats.GET("/year-method-success", apiHandler.Handle("get-transaction-stats-year-method-success", statsHandler.FindTransactionYearMethodSuccess))
	transactionStats.GET("/month-method-failed", apiHandler.Handle("get-transaction-stats-month-method-failed", statsHandler.FindTransactionMonthMethodFailed))
	transactionStats.GET("/year-method-failed", apiHandler.Handle("get-transaction-stats-year-method-failed", statsHandler.FindTransactionYearMethodFailed))
	transactionStats.GET("/by-merchant/month-status-success", apiHandler.Handle("get-transaction-stats-merchant-month-status-success", statsHandler.FindTransactionMonthStatusSuccessByMerchant))
	transactionStats.GET("/by-merchant/year-status-success", apiHandler.Handle("get-transaction-stats-merchant-year-status-success", statsHandler.FindTransactionYearStatusSuccessByMerchant))
	transactionStats.GET("/by-merchant/month-status-failed", apiHandler.Handle("get-transaction-stats-merchant-month-status-failed", statsHandler.FindTransactionMonthStatusFailedByMerchant))
	transactionStats.GET("/by-merchant/year-status-failed", apiHandler.Handle("get-transaction-stats-merchant-year-status-failed", statsHandler.FindTransactionYearStatusFailedByMerchant))
	transactionStats.GET("/by-merchant/month-method-success", apiHandler.Handle("get-transaction-stats-merchant-month-method-success", statsHandler.FindTransactionMonthMethodSuccessByMerchant))
	transactionStats.GET("/by-merchant/year-method-success", apiHandler.Handle("get-transaction-stats-merchant-year-method-success", statsHandler.FindTransactionYearMethodSuccessByMerchant))
	transactionStats.GET("/by-merchant/month-method-failed", apiHandler.Handle("get-transaction-stats-merchant-month-method-failed", statsHandler.FindTransactionMonthMethodFailedByMerchant))
	transactionStats.GET("/by-merchant/year-method-failed", apiHandler.Handle("get-transaction-stats-merchant-year-method-failed", statsHandler.FindTransactionYearMethodFailedByMerchant))

	return statsHandler
}

func (h *statsHandleApi) parseYearMonth(c echo.Context) (int, int, error) {
	year, err := parseQueryIntWithValidation(c, "year", 1, 9999)
	if err != nil {
		return 0, 0, errors.NewBadRequestError("invalid year")
	}
	month, err := parseQueryIntWithValidation(c, "month", 1, 12)
	if err != nil {
		return 0, 0, errors.NewBadRequestError("invalid month")
	}
	return year, month, nil
}

func (h *statsHandleApi) parseYear(c echo.Context) (int, error) {
	year, err := parseQueryIntWithValidation(c, "year", 1, 9999)
	if err != nil {
		return 0, errors.NewBadRequestError("invalid year")
	}
	return year, nil
}

func (h *statsHandleApi) parseID(c echo.Context, key string) (int, error) {
	id, err := parseQueryIntWithValidation(c, key, 1, 999999)
	if err != nil {
		return 0, errors.NewBadRequestError("invalid " + key)
	}
	return id, nil
}

// --- Order stats ---

// @Summary Monthly total revenue (order)
// @Tags Order Stats
// @Accept json
// @Produce json
// @Param year query int true "Year (e.g. 2026)"
// @Param month query int true "Month (1-12)"
// @Success 200 {object} response.ApiResponseOrderMonthlyTotalRevenue
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/order-stats/monthly-total-revenue [get]
// @Security BearerAuth
func (h *statsHandleApi) FindMonthlyTotalRevenue(c echo.Context) error {
	ctx := c.Request().Context()
	year, month, err := h.parseYearMonth(c)
	if err != nil {
		return err
	}

	cacheKey := fmt.Sprintf("apigw:order-stats:monthly-total-revenue:%d:%d", year, month)
	if cached, found := gateway_cache.Get[response.ApiResponseOrderMonthlyTotalRevenue](ctx, h.cache, cacheKey); found && cached != nil {
		return c.JSON(http.StatusOK, cached)
	}

	res, err := h.clients.Order.FindMonthlyTotalRevenue(ctx, &pborder.FindYearMonthTotalRevenue{Year: int32(year), Month: int32(month)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponseOrderMonthlyTotalRevenue(res)
	gateway_cache.Set(ctx, h.cache, cacheKey, so, 5*time.Minute)
	return c.JSON(http.StatusOK, so)
}

// @Summary Yearly total revenue (order)
// @Tags Order Stats
// @Accept json
// @Produce json
// @Param year query int true "Year (e.g. 2026)"
// @Success 200 {object} response.ApiResponseOrderYearlyTotalRevenue
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/order-stats/yearly-total-revenue [get]
// @Security BearerAuth
func (h *statsHandleApi) FindYearlyTotalRevenue(c echo.Context) error {
	ctx := c.Request().Context()
	year, err := h.parseYear(c)
	if err != nil {
		return err
	}

	cacheKey := fmt.Sprintf("apigw:order-stats:yearly-total-revenue:%d", year)
	if cached, found := gateway_cache.Get[response.ApiResponseOrderYearlyTotalRevenue](ctx, h.cache, cacheKey); found && cached != nil {
		return c.JSON(http.StatusOK, cached)
	}

	res, err := h.clients.Order.FindYearlyTotalRevenue(ctx, &pborder.FindYearTotalRevenue{Year: int32(year)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponseOrderYearlyTotalRevenue(res)
	gateway_cache.Set(ctx, h.cache, cacheKey, so, 5*time.Minute)
	return c.JSON(http.StatusOK, so)
}

// @Summary Monthly revenue per month (order)
// @Tags Order Stats
// @Accept json
// @Produce json
// @Param year query int true "Year (e.g. 2026)"
// @Success 200 {object} response.ApiResponseOrderMonthly
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/order-stats/monthly-revenue [get]
// @Security BearerAuth
func (h *statsHandleApi) FindMonthlyRevenue(c echo.Context) error {
	ctx := c.Request().Context()
	year, err := h.parseYear(c)
	if err != nil {
		return err
	}

	cacheKey := fmt.Sprintf("apigw:order-stats:monthly-revenue:%d", year)
	if cached, found := gateway_cache.Get[response.ApiResponseOrderMonthly](ctx, h.cache, cacheKey); found && cached != nil {
		return c.JSON(http.StatusOK, cached)
	}

	res, err := h.clients.Order.FindMonthlyRevenue(ctx, &pborder.FindYearOrder{Year: int32(year)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponseOrderMonthly(res)
	gateway_cache.Set(ctx, h.cache, cacheKey, so, 5*time.Minute)
	return c.JSON(http.StatusOK, so)
}

// @Summary Yearly revenue summary (order)
// @Tags Order Stats
// @Accept json
// @Produce json
// @Param year query int true "Year (e.g. 2026)"
// @Success 200 {object} response.ApiResponseOrderYearly
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/order-stats/yearly-revenue [get]
// @Security BearerAuth
func (h *statsHandleApi) FindYearlyRevenue(c echo.Context) error {
	ctx := c.Request().Context()
	year, err := h.parseYear(c)
	if err != nil {
		return err
	}

	cacheKey := fmt.Sprintf("apigw:order-stats:yearly-revenue:%d", year)
	if cached, found := gateway_cache.Get[response.ApiResponseOrderYearly](ctx, h.cache, cacheKey); found && cached != nil {
		return c.JSON(http.StatusOK, cached)
	}

	res, err := h.clients.Order.FindYearlyRevenue(ctx, &pborder.FindYearOrder{Year: int32(year)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponseOrderYearly(res)
	gateway_cache.Set(ctx, h.cache, cacheKey, so, 5*time.Minute)
	return c.JSON(http.StatusOK, so)
}

// @Summary Monthly total revenue by merchant (order)
// @Tags Order Stats
// @Accept json
// @Produce json
// @Param merchant_id query int true "Merchant ID"
// @Param year query int true "Year (e.g. 2026)"
// @Param month query int true "Month (1-12)"
// @Success 200 {object} response.ApiResponseOrderMonthlyTotalRevenue
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/order-stats/by-merchant/monthly-total-revenue [get]
// @Security BearerAuth
func (h *statsHandleApi) FindMonthlyTotalRevenueByMerchant(c echo.Context) error {
	ctx := c.Request().Context()
	merchantID, err := h.parseID(c, "merchant_id")
	if err != nil {
		return err
	}
	year, month, err := h.parseYearMonth(c)
	if err != nil {
		return err
	}

	cacheKey := fmt.Sprintf("apigw:order-stats:monthly-total-revenue:merchant:%d:%d:%d", merchantID, year, month)
	if cached, found := gateway_cache.Get[response.ApiResponseOrderMonthlyTotalRevenue](ctx, h.cache, cacheKey); found && cached != nil {
		return c.JSON(http.StatusOK, cached)
	}

	res, err := h.clients.OrderMerchant.FindMonthlyTotalRevenueByMerchant(ctx, &pborder.FindYearMonthTotalRevenueByMerchant{Year: int32(year), Month: int32(month), MerchantId: int32(merchantID)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponseOrderMonthlyTotalRevenue(res)
	gateway_cache.Set(ctx, h.cache, cacheKey, so, 5*time.Minute)
	return c.JSON(http.StatusOK, so)
}

// @Summary Yearly total revenue by merchant (order)
// @Tags Order Stats
// @Accept json
// @Produce json
// @Param merchant_id query int true "Merchant ID"
// @Param year query int true "Year (e.g. 2026)"
// @Success 200 {object} response.ApiResponseOrderYearlyTotalRevenue
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/order-stats/by-merchant/yearly-total-revenue [get]
// @Security BearerAuth
func (h *statsHandleApi) FindYearlyTotalRevenueByMerchant(c echo.Context) error {
	ctx := c.Request().Context()
	merchantID, err := h.parseID(c, "merchant_id")
	if err != nil {
		return err
	}
	year, err := h.parseYear(c)
	if err != nil {
		return err
	}

	cacheKey := fmt.Sprintf("apigw:order-stats:yearly-total-revenue:merchant:%d:%d", merchantID, year)
	if cached, found := gateway_cache.Get[response.ApiResponseOrderYearlyTotalRevenue](ctx, h.cache, cacheKey); found && cached != nil {
		return c.JSON(http.StatusOK, cached)
	}

	res, err := h.clients.OrderMerchant.FindYearlyTotalRevenueByMerchant(ctx, &pborder.FindYearTotalRevenueByMerchant{Year: int32(year), MerchantId: int32(merchantID)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponseOrderYearlyTotalRevenue(res)
	gateway_cache.Set(ctx, h.cache, cacheKey, so, 5*time.Minute)
	return c.JSON(http.StatusOK, so)
}

// @Summary Monthly revenue per month by merchant (order)
// @Tags Order Stats
// @Accept json
// @Produce json
// @Param merchant_id query int true "Merchant ID"
// @Param year query int true "Year (e.g. 2026)"
// @Success 200 {object} response.ApiResponseOrderMonthly
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/order-stats/by-merchant/monthly-revenue [get]
// @Security BearerAuth
func (h *statsHandleApi) FindMonthlyRevenueByMerchant(c echo.Context) error {
	ctx := c.Request().Context()
	merchantID, err := h.parseID(c, "merchant_id")
	if err != nil {
		return err
	}
	year, err := h.parseYear(c)
	if err != nil {
		return err
	}

	cacheKey := fmt.Sprintf("apigw:order-stats:monthly-revenue:merchant:%d:%d", merchantID, year)
	if cached, found := gateway_cache.Get[response.ApiResponseOrderMonthly](ctx, h.cache, cacheKey); found && cached != nil {
		return c.JSON(http.StatusOK, cached)
	}

	res, err := h.clients.OrderMerchant.FindMonthlyRevenueByMerchant(ctx, &pborder.FindYearOrderByMerchant{Year: int32(year), MerchantId: int32(merchantID)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponseOrderMonthly(res)
	gateway_cache.Set(ctx, h.cache, cacheKey, so, 5*time.Minute)
	return c.JSON(http.StatusOK, so)
}

// @Summary Yearly revenue summary by merchant (order)
// @Tags Order Stats
// @Accept json
// @Produce json
// @Param merchant_id query int true "Merchant ID"
// @Param year query int true "Year (e.g. 2026)"
// @Success 200 {object} response.ApiResponseOrderYearly
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/order-stats/by-merchant/yearly-revenue [get]
// @Security BearerAuth
func (h *statsHandleApi) FindYearlyRevenueByMerchant(c echo.Context) error {
	ctx := c.Request().Context()
	merchantID, err := h.parseID(c, "merchant_id")
	if err != nil {
		return err
	}
	year, err := h.parseYear(c)
	if err != nil {
		return err
	}

	cacheKey := fmt.Sprintf("apigw:order-stats:yearly-revenue:merchant:%d:%d", merchantID, year)
	if cached, found := gateway_cache.Get[response.ApiResponseOrderYearly](ctx, h.cache, cacheKey); found && cached != nil {
		return c.JSON(http.StatusOK, cached)
	}

	res, err := h.clients.OrderMerchant.FindYearlyRevenueByMerchant(ctx, &pborder.FindYearOrderByMerchant{Year: int32(year), MerchantId: int32(merchantID)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponseOrderYearly(res)
	gateway_cache.Set(ctx, h.cache, cacheKey, so, 5*time.Minute)
	return c.JSON(http.StatusOK, so)
}

// @Summary Monthly total revenue by order id (order)
// @Tags Order Stats
// @Accept json
// @Produce json
// @Param order_id query int true "Order ID"
// @Param year query int true "Year (e.g. 2026)"
// @Param month query int true "Month (1-12)"
// @Success 200 {object} response.ApiResponseOrderMonthlyTotalRevenue
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/order-stats/by-id/monthly-total-revenue [get]
// @Security BearerAuth
func (h *statsHandleApi) FindMonthlyTotalRevenueById(c echo.Context) error {
	ctx := c.Request().Context()
	orderID, err := h.parseID(c, "order_id")
	if err != nil {
		return err
	}
	year, month, err := h.parseYearMonth(c)
	if err != nil {
		return err
	}

	cacheKey := fmt.Sprintf("apigw:order-stats:monthly-total-revenue:order:%d:%d:%d", orderID, year, month)
	if cached, found := gateway_cache.Get[response.ApiResponseOrderMonthlyTotalRevenue](ctx, h.cache, cacheKey); found && cached != nil {
		return c.JSON(http.StatusOK, cached)
	}

	res, err := h.clients.OrderById.FindMonthlyTotalRevenueById(ctx, &pborder.FindYearMonthTotalRevenueById{Year: int32(year), Month: int32(month), OrderId: int32(orderID)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponseOrderMonthlyTotalRevenue(res)
	gateway_cache.Set(ctx, h.cache, cacheKey, so, 5*time.Minute)
	return c.JSON(http.StatusOK, so)
}

// @Summary Yearly total revenue by order id (order)
// @Tags Order Stats
// @Accept json
// @Produce json
// @Param order_id query int true "Order ID"
// @Param year query int true "Year (e.g. 2026)"
// @Success 200 {object} response.ApiResponseOrderYearlyTotalRevenue
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/order-stats/by-id/yearly-total-revenue [get]
// @Security BearerAuth
func (h *statsHandleApi) FindYearlyTotalRevenueById(c echo.Context) error {
	ctx := c.Request().Context()
	orderID, err := h.parseID(c, "order_id")
	if err != nil {
		return err
	}
	year, err := h.parseYear(c)
	if err != nil {
		return err
	}

	cacheKey := fmt.Sprintf("apigw:order-stats:yearly-total-revenue:order:%d:%d", orderID, year)
	if cached, found := gateway_cache.Get[response.ApiResponseOrderYearlyTotalRevenue](ctx, h.cache, cacheKey); found && cached != nil {
		return c.JSON(http.StatusOK, cached)
	}

	res, err := h.clients.OrderById.FindYearlyTotalRevenueById(ctx, &pborder.FindYearTotalRevenueById{Year: int32(year), OrderId: int32(orderID)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponseOrderYearlyTotalRevenue(res)
	gateway_cache.Set(ctx, h.cache, cacheKey, so, 5*time.Minute)
	return c.JSON(http.StatusOK, so)
}

// --- Cashier stats ---

// @Summary Monthly total sales (cashier)
// @Tags Cashier Stats
// @Accept json
// @Produce json
// @Param year query int true "Year (e.g. 2026)"
// @Param month query int true "Month (1-12)"
// @Success 200 {object} response.ApiResponseCashierMonthlyTotalSales
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/cashier-stats/monthly-total-sales [get]
// @Security BearerAuth
func (h *statsHandleApi) FindCashierMonthlyTotalSales(c echo.Context) error {
	ctx := c.Request().Context()
	year, month, err := h.parseYearMonth(c)
	if err != nil {
		return err
	}

	cacheKey := fmt.Sprintf("apigw:cashier-stats:monthly-total-sales:%d:%d", year, month)
	if cached, found := gateway_cache.Get[response.ApiResponseCashierMonthlyTotalSales](ctx, h.cache, cacheKey); found && cached != nil {
		return c.JSON(http.StatusOK, cached)
	}

	res, err := h.clients.Cashier.FindMonthlyTotalSales(ctx, &pbcashier.FindYearMonthTotalSales{Year: int32(year), Month: int32(month)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponseCashierMonthlyTotalSales(res)
	gateway_cache.Set(ctx, h.cache, cacheKey, so, 5*time.Minute)
	return c.JSON(http.StatusOK, so)
}

// @Summary Yearly total sales (cashier)
// @Tags Cashier Stats
// @Accept json
// @Produce json
// @Param year query int true "Year (e.g. 2026)"
// @Success 200 {object} response.ApiResponseCashierYearlyTotalSales
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/cashier-stats/yearly-total-sales [get]
// @Security BearerAuth
func (h *statsHandleApi) FindCashierYearlyTotalSales(c echo.Context) error {
	ctx := c.Request().Context()
	year, err := h.parseYear(c)
	if err != nil {
		return err
	}

	cacheKey := fmt.Sprintf("apigw:cashier-stats:yearly-total-sales:%d", year)
	if cached, found := gateway_cache.Get[response.ApiResponseCashierYearlyTotalSales](ctx, h.cache, cacheKey); found && cached != nil {
		return c.JSON(http.StatusOK, cached)
	}

	res, err := h.clients.Cashier.FindYearlyTotalSales(ctx, &pbcashier.FindYearTotalSales{Year: int32(year)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponseCashierYearlyTotalSales(res)
	gateway_cache.Set(ctx, h.cache, cacheKey, so, 5*time.Minute)
	return c.JSON(http.StatusOK, so)
}

// @Summary Monthly sales per cashier (cashier)
// @Tags Cashier Stats
// @Accept json
// @Produce json
// @Param year query int true "Year (e.g. 2026)"
// @Success 200 {object} response.ApiResponseCashierMonthSales
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/cashier-stats/month-sales [get]
// @Security BearerAuth
func (h *statsHandleApi) FindCashierMonthSales(c echo.Context) error {
	ctx := c.Request().Context()
	year, err := h.parseYear(c)
	if err != nil {
		return err
	}

	cacheKey := fmt.Sprintf("apigw:cashier-stats:month-sales:%d", year)
	if cached, found := gateway_cache.Get[response.ApiResponseCashierMonthSales](ctx, h.cache, cacheKey); found && cached != nil {
		return c.JSON(http.StatusOK, cached)
	}

	res, err := h.clients.Cashier.FindMonthSales(ctx, &pbcashier.FindYearCashier{Year: int32(year)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponseCashierMonthSales(res)
	gateway_cache.Set(ctx, h.cache, cacheKey, so, 5*time.Minute)
	return c.JSON(http.StatusOK, so)
}

// @Summary Yearly sales per cashier (cashier)
// @Tags Cashier Stats
// @Accept json
// @Produce json
// @Param year query int true "Year (e.g. 2026)"
// @Success 200 {object} response.ApiResponseCashierYearSales
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/cashier-stats/year-sales [get]
// @Security BearerAuth
func (h *statsHandleApi) FindCashierYearSales(c echo.Context) error {
	ctx := c.Request().Context()
	year, err := h.parseYear(c)
	if err != nil {
		return err
	}

	cacheKey := fmt.Sprintf("apigw:cashier-stats:year-sales:%d", year)
	if cached, found := gateway_cache.Get[response.ApiResponseCashierYearSales](ctx, h.cache, cacheKey); found && cached != nil {
		return c.JSON(http.StatusOK, cached)
	}

	res, err := h.clients.Cashier.FindYearSales(ctx, &pbcashier.FindYearCashier{Year: int32(year)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponseCashierYearSales(res)
	gateway_cache.Set(ctx, h.cache, cacheKey, so, 5*time.Minute)
	return c.JSON(http.StatusOK, so)
}

// @Summary Monthly total sales by merchant (cashier)
// @Tags Cashier Stats
// @Accept json
// @Produce json
// @Param merchant_id query int true "Merchant ID"
// @Param year query int true "Year (e.g. 2026)"
// @Param month query int true "Month (1-12)"
// @Success 200 {object} response.ApiResponseCashierMonthlyTotalSales
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/cashier-stats/by-merchant/monthly-total-sales [get]
// @Security BearerAuth
func (h *statsHandleApi) FindCashierMonthlyTotalSalesByMerchant(c echo.Context) error {
	ctx := c.Request().Context()
	merchantID, err := h.parseID(c, "merchant_id")
	if err != nil {
		return err
	}
	year, month, err := h.parseYearMonth(c)
	if err != nil {
		return err
	}

	cacheKey := fmt.Sprintf("apigw:cashier-stats:monthly-total-sales:merchant:%d:%d:%d", merchantID, year, month)
	if cached, found := gateway_cache.Get[response.ApiResponseCashierMonthlyTotalSales](ctx, h.cache, cacheKey); found && cached != nil {
		return c.JSON(http.StatusOK, cached)
	}

	res, err := h.clients.CashierMerchant.FindMonthlyTotalSalesByMerchant(ctx, &pbcashier.FindYearMonthTotalSalesByMerchant{Year: int32(year), Month: int32(month), MerchantId: int32(merchantID)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponseCashierMonthlyTotalSales(res)
	gateway_cache.Set(ctx, h.cache, cacheKey, so, 5*time.Minute)
	return c.JSON(http.StatusOK, so)
}

// @Summary Yearly total sales by merchant (cashier)
// @Tags Cashier Stats
// @Accept json
// @Produce json
// @Param merchant_id query int true "Merchant ID"
// @Param year query int true "Year (e.g. 2026)"
// @Success 200 {object} response.ApiResponseCashierYearlyTotalSales
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/cashier-stats/by-merchant/yearly-total-sales [get]
// @Security BearerAuth
func (h *statsHandleApi) FindCashierYearlyTotalSalesByMerchant(c echo.Context) error {
	ctx := c.Request().Context()
	merchantID, err := h.parseID(c, "merchant_id")
	if err != nil {
		return err
	}
	year, err := h.parseYear(c)
	if err != nil {
		return err
	}

	cacheKey := fmt.Sprintf("apigw:cashier-stats:yearly-total-sales:merchant:%d:%d", merchantID, year)
	if cached, found := gateway_cache.Get[response.ApiResponseCashierYearlyTotalSales](ctx, h.cache, cacheKey); found && cached != nil {
		return c.JSON(http.StatusOK, cached)
	}

	res, err := h.clients.CashierMerchant.FindYearlyTotalSalesByMerchant(ctx, &pbcashier.FindYearTotalSalesByMerchant{Year: int32(year), MerchantId: int32(merchantID)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponseCashierYearlyTotalSales(res)
	gateway_cache.Set(ctx, h.cache, cacheKey, so, 5*time.Minute)
	return c.JSON(http.StatusOK, so)
}

// @Summary Monthly sales per cashier by merchant (cashier)
// @Tags Cashier Stats
// @Accept json
// @Produce json
// @Param merchant_id query int true "Merchant ID"
// @Param year query int true "Year (e.g. 2026)"
// @Success 200 {object} response.ApiResponseCashierMonthSales
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/cashier-stats/by-merchant/month-sales [get]
// @Security BearerAuth
func (h *statsHandleApi) FindCashierMonthSalesByMerchant(c echo.Context) error {
	ctx := c.Request().Context()
	merchantID, err := h.parseID(c, "merchant_id")
	if err != nil {
		return err
	}
	year, err := h.parseYear(c)
	if err != nil {
		return err
	}

	cacheKey := fmt.Sprintf("apigw:cashier-stats:month-sales:merchant:%d:%d", merchantID, year)
	if cached, found := gateway_cache.Get[response.ApiResponseCashierMonthSales](ctx, h.cache, cacheKey); found && cached != nil {
		return c.JSON(http.StatusOK, cached)
	}

	res, err := h.clients.CashierMerchant.FindMonthSalesByMerchant(ctx, &pbcashier.FindYearCashierByMerchant{Year: int32(year), MerchantId: int32(merchantID)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponseCashierMonthSales(res)
	gateway_cache.Set(ctx, h.cache, cacheKey, so, 5*time.Minute)
	return c.JSON(http.StatusOK, so)
}

// @Summary Yearly sales per cashier by merchant (cashier)
// @Tags Cashier Stats
// @Accept json
// @Produce json
// @Param merchant_id query int true "Merchant ID"
// @Param year query int true "Year (e.g. 2026)"
// @Success 200 {object} response.ApiResponseCashierYearSales
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/cashier-stats/by-merchant/year-sales [get]
// @Security BearerAuth
func (h *statsHandleApi) FindCashierYearSalesByMerchant(c echo.Context) error {
	ctx := c.Request().Context()
	merchantID, err := h.parseID(c, "merchant_id")
	if err != nil {
		return err
	}
	year, err := h.parseYear(c)
	if err != nil {
		return err
	}

	cacheKey := fmt.Sprintf("apigw:cashier-stats:year-sales:merchant:%d:%d", merchantID, year)
	if cached, found := gateway_cache.Get[response.ApiResponseCashierYearSales](ctx, h.cache, cacheKey); found && cached != nil {
		return c.JSON(http.StatusOK, cached)
	}

	res, err := h.clients.CashierMerchant.FindYearSalesByMerchant(ctx, &pbcashier.FindYearCashierByMerchant{Year: int32(year), MerchantId: int32(merchantID)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponseCashierYearSales(res)
	gateway_cache.Set(ctx, h.cache, cacheKey, so, 5*time.Minute)
	return c.JSON(http.StatusOK, so)
}

// @Summary Monthly total sales by cashier id (cashier)
// @Tags Cashier Stats
// @Accept json
// @Produce json
// @Param cashier_id query int true "Cashier ID"
// @Param year query int true "Year (e.g. 2026)"
// @Param month query int true "Month (1-12)"
// @Success 200 {object} response.ApiResponseCashierMonthlyTotalSales
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/cashier-stats/by-id/monthly-total-sales [get]
// @Security BearerAuth
func (h *statsHandleApi) FindCashierMonthlyTotalSalesById(c echo.Context) error {
	ctx := c.Request().Context()
	cashierID, err := h.parseID(c, "cashier_id")
	if err != nil {
		return err
	}
	year, month, err := h.parseYearMonth(c)
	if err != nil {
		return err
	}

	cacheKey := fmt.Sprintf("apigw:cashier-stats:monthly-total-sales:cashier:%d:%d:%d", cashierID, year, month)
	if cached, found := gateway_cache.Get[response.ApiResponseCashierMonthlyTotalSales](ctx, h.cache, cacheKey); found && cached != nil {
		return c.JSON(http.StatusOK, cached)
	}

	res, err := h.clients.CashierById.FindMonthlyTotalSalesById(ctx, &pbcashier.FindYearMonthTotalSalesById{Year: int32(year), Month: int32(month), CashierId: int32(cashierID)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponseCashierMonthlyTotalSales(res)
	gateway_cache.Set(ctx, h.cache, cacheKey, so, 5*time.Minute)
	return c.JSON(http.StatusOK, so)
}

// @Summary Yearly total sales by cashier id (cashier)
// @Tags Cashier Stats
// @Accept json
// @Produce json
// @Param cashier_id query int true "Cashier ID"
// @Param year query int true "Year (e.g. 2026)"
// @Success 200 {object} response.ApiResponseCashierYearlyTotalSales
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/cashier-stats/by-id/yearly-total-sales [get]
// @Security BearerAuth
func (h *statsHandleApi) FindCashierYearlyTotalSalesById(c echo.Context) error {
	ctx := c.Request().Context()
	cashierID, err := h.parseID(c, "cashier_id")
	if err != nil {
		return err
	}
	year, err := h.parseYear(c)
	if err != nil {
		return err
	}

	cacheKey := fmt.Sprintf("apigw:cashier-stats:yearly-total-sales:cashier:%d:%d", cashierID, year)
	if cached, found := gateway_cache.Get[response.ApiResponseCashierYearlyTotalSales](ctx, h.cache, cacheKey); found && cached != nil {
		return c.JSON(http.StatusOK, cached)
	}

	res, err := h.clients.CashierById.FindYearlyTotalSalesById(ctx, &pbcashier.FindYearTotalSalesById{Year: int32(year), CashierId: int32(cashierID)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponseCashierYearlyTotalSales(res)
	gateway_cache.Set(ctx, h.cache, cacheKey, so, 5*time.Minute)
	return c.JSON(http.StatusOK, so)
}

// @Summary Monthly sales by cashier id (cashier)
// @Tags Cashier Stats
// @Accept json
// @Produce json
// @Param cashier_id query int true "Cashier ID"
// @Param year query int true "Year (e.g. 2026)"
// @Success 200 {object} response.ApiResponseCashierMonthSales
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/cashier-stats/by-id/month-sales [get]
// @Security BearerAuth
func (h *statsHandleApi) FindCashierMonthSalesById(c echo.Context) error {
	ctx := c.Request().Context()
	cashierID, err := h.parseID(c, "cashier_id")
	if err != nil {
		return err
	}
	year, err := h.parseYear(c)
	if err != nil {
		return err
	}

	cacheKey := fmt.Sprintf("apigw:cashier-stats:month-sales:cashier:%d:%d", cashierID, year)
	if cached, found := gateway_cache.Get[response.ApiResponseCashierMonthSales](ctx, h.cache, cacheKey); found && cached != nil {
		return c.JSON(http.StatusOK, cached)
	}

	res, err := h.clients.CashierById.FindMonthSalesById(ctx, &pbcashier.FindYearCashierById{Year: int32(year), CashierId: int32(cashierID)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponseCashierMonthSales(res)
	gateway_cache.Set(ctx, h.cache, cacheKey, so, 5*time.Minute)
	return c.JSON(http.StatusOK, so)
}

// @Summary Yearly sales by cashier id (cashier)
// @Tags Cashier Stats
// @Accept json
// @Produce json
// @Param cashier_id query int true "Cashier ID"
// @Param year query int true "Year (e.g. 2026)"
// @Success 200 {object} response.ApiResponseCashierYearSales
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/cashier-stats/by-id/year-sales [get]
// @Security BearerAuth
func (h *statsHandleApi) FindCashierYearSalesById(c echo.Context) error {
	ctx := c.Request().Context()
	cashierID, err := h.parseID(c, "cashier_id")
	if err != nil {
		return err
	}
	year, err := h.parseYear(c)
	if err != nil {
		return err
	}

	cacheKey := fmt.Sprintf("apigw:cashier-stats:year-sales:cashier:%d:%d", cashierID, year)
	if cached, found := gateway_cache.Get[response.ApiResponseCashierYearSales](ctx, h.cache, cacheKey); found && cached != nil {
		return c.JSON(http.StatusOK, cached)
	}

	res, err := h.clients.CashierById.FindYearSalesById(ctx, &pbcashier.FindYearCashierById{Year: int32(year), CashierId: int32(cashierID)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponseCashierYearSales(res)
	gateway_cache.Set(ctx, h.cache, cacheKey, so, 5*time.Minute)
	return c.JSON(http.StatusOK, so)
}

// --- Category stats ---

// @Summary Monthly total prices (category)
// @Tags Category Stats
// @Accept json
// @Produce json
// @Param year query int true "Year (e.g. 2026)"
// @Param month query int true "Month (1-12)"
// @Success 200 {object} response.ApiResponseCategoryMonthlyTotalPrice
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/category-stats/monthly-total-prices [get]
// @Security BearerAuth
func (h *statsHandleApi) FindCategoryMonthlyTotalPrices(c echo.Context) error {
	ctx := c.Request().Context()
	year, month, err := h.parseYearMonth(c)
	if err != nil {
		return err
	}

	cacheKey := fmt.Sprintf("apigw:category-stats:monthly-total-prices:%d:%d", year, month)
	if cached, found := gateway_cache.Get[response.ApiResponseCategoryMonthlyTotalPrice](ctx, h.cache, cacheKey); found && cached != nil {
		return c.JSON(http.StatusOK, cached)
	}

	res, err := h.clients.Category.FindMonthlyTotalPrices(ctx, &pbcategory.FindYearMonthTotalPrices{Year: int32(year), Month: int32(month)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponseCategoryMonthlyTotalPrice(res)
	gateway_cache.Set(ctx, h.cache, cacheKey, so, 5*time.Minute)
	return c.JSON(http.StatusOK, so)
}

// @Summary Yearly total prices (category)
// @Tags Category Stats
// @Accept json
// @Produce json
// @Param year query int true "Year (e.g. 2026)"
// @Success 200 {object} response.ApiResponseCategoryYearlyTotalPrice
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/category-stats/yearly-total-prices [get]
// @Security BearerAuth
func (h *statsHandleApi) FindCategoryYearlyTotalPrices(c echo.Context) error {
	ctx := c.Request().Context()
	year, err := h.parseYear(c)
	if err != nil {
		return err
	}

	cacheKey := fmt.Sprintf("apigw:category-stats:yearly-total-prices:%d", year)
	if cached, found := gateway_cache.Get[response.ApiResponseCategoryYearlyTotalPrice](ctx, h.cache, cacheKey); found && cached != nil {
		return c.JSON(http.StatusOK, cached)
	}

	res, err := h.clients.Category.FindYearlyTotalPrices(ctx, &pbcategory.FindYearTotalPrices{Year: int32(year)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponseCategoryYearlyTotalPrice(res)
	gateway_cache.Set(ctx, h.cache, cacheKey, so, 5*time.Minute)
	return c.JSON(http.StatusOK, so)
}

// @Summary Monthly price per category (category)
// @Tags Category Stats
// @Accept json
// @Produce json
// @Param year query int true "Year (e.g. 2026)"
// @Success 200 {object} response.ApiResponseCategoryMonthPrice
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/category-stats/month-price [get]
// @Security BearerAuth
func (h *statsHandleApi) FindCategoryMonthPrice(c echo.Context) error {
	ctx := c.Request().Context()
	year, err := h.parseYear(c)
	if err != nil {
		return err
	}

	cacheKey := fmt.Sprintf("apigw:category-stats:month-price:%d", year)
	if cached, found := gateway_cache.Get[response.ApiResponseCategoryMonthPrice](ctx, h.cache, cacheKey); found && cached != nil {
		return c.JSON(http.StatusOK, cached)
	}

	res, err := h.clients.Category.FindMonthPrice(ctx, &pbcategory.FindYearCategory{Year: int32(year)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponseCategoryMonthPrice(res)
	gateway_cache.Set(ctx, h.cache, cacheKey, so, 5*time.Minute)
	return c.JSON(http.StatusOK, so)
}

// @Summary Yearly price per category (category)
// @Tags Category Stats
// @Accept json
// @Produce json
// @Param year query int true "Year (e.g. 2026)"
// @Success 200 {object} response.ApiResponseCategoryYearPrice
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/category-stats/year-price [get]
// @Security BearerAuth
func (h *statsHandleApi) FindCategoryYearPrice(c echo.Context) error {
	ctx := c.Request().Context()
	year, err := h.parseYear(c)
	if err != nil {
		return err
	}

	cacheKey := fmt.Sprintf("apigw:category-stats:year-price:%d", year)
	if cached, found := gateway_cache.Get[response.ApiResponseCategoryYearPrice](ctx, h.cache, cacheKey); found && cached != nil {
		return c.JSON(http.StatusOK, cached)
	}

	res, err := h.clients.Category.FindYearPrice(ctx, &pbcategory.FindYearCategory{Year: int32(year)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponseCategoryYearPrice(res)
	gateway_cache.Set(ctx, h.cache, cacheKey, so, 5*time.Minute)
	return c.JSON(http.StatusOK, so)
}

// @Summary Monthly total prices by merchant (category)
// @Tags Category Stats
// @Accept json
// @Produce json
// @Param merchant_id query int true "Merchant ID"
// @Param year query int true "Year (e.g. 2026)"
// @Param month query int true "Month (1-12)"
// @Success 200 {object} response.ApiResponseCategoryMonthlyTotalPrice
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/category-stats/by-merchant/monthly-total-prices [get]
// @Security BearerAuth
func (h *statsHandleApi) FindCategoryMonthlyTotalPricesByMerchant(c echo.Context) error {
	ctx := c.Request().Context()
	merchantID, err := h.parseID(c, "merchant_id")
	if err != nil {
		return err
	}
	year, month, err := h.parseYearMonth(c)
	if err != nil {
		return err
	}

	cacheKey := fmt.Sprintf("apigw:category-stats:monthly-total-prices:merchant:%d:%d:%d", merchantID, year, month)
	if cached, found := gateway_cache.Get[response.ApiResponseCategoryMonthlyTotalPrice](ctx, h.cache, cacheKey); found && cached != nil {
		return c.JSON(http.StatusOK, cached)
	}

	res, err := h.clients.CategoryMerchant.FindMonthlyTotalPricesByMerchant(ctx, &pbcategory.FindYearMonthTotalPriceByMerchant{Year: int32(year), Month: int32(month), MerchantId: int32(merchantID)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponseCategoryMonthlyTotalPrice(res)
	gateway_cache.Set(ctx, h.cache, cacheKey, so, 5*time.Minute)
	return c.JSON(http.StatusOK, so)
}

// @Summary Yearly total prices by merchant (category)
// @Tags Category Stats
// @Accept json
// @Produce json
// @Param merchant_id query int true "Merchant ID"
// @Param year query int true "Year (e.g. 2026)"
// @Success 200 {object} response.ApiResponseCategoryYearlyTotalPrice
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/category-stats/by-merchant/yearly-total-prices [get]
// @Security BearerAuth
func (h *statsHandleApi) FindCategoryYearlyTotalPricesByMerchant(c echo.Context) error {
	ctx := c.Request().Context()
	merchantID, err := h.parseID(c, "merchant_id")
	if err != nil {
		return err
	}
	year, err := h.parseYear(c)
	if err != nil {
		return err
	}

	cacheKey := fmt.Sprintf("apigw:category-stats:yearly-total-prices:merchant:%d:%d", merchantID, year)
	if cached, found := gateway_cache.Get[response.ApiResponseCategoryYearlyTotalPrice](ctx, h.cache, cacheKey); found && cached != nil {
		return c.JSON(http.StatusOK, cached)
	}

	res, err := h.clients.CategoryMerchant.FindYearlyTotalPricesByMerchant(ctx, &pbcategory.FindYearTotalPriceByMerchant{Year: int32(year), MerchantId: int32(merchantID)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponseCategoryYearlyTotalPrice(res)
	gateway_cache.Set(ctx, h.cache, cacheKey, so, 5*time.Minute)
	return c.JSON(http.StatusOK, so)
}

// @Summary Monthly price per category by merchant (category)
// @Tags Category Stats
// @Accept json
// @Produce json
// @Param merchant_id query int true "Merchant ID"
// @Param year query int true "Year (e.g. 2026)"
// @Success 200 {object} response.ApiResponseCategoryMonthPrice
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/category-stats/by-merchant/month-price [get]
// @Security BearerAuth
func (h *statsHandleApi) FindCategoryMonthPriceByMerchant(c echo.Context) error {
	ctx := c.Request().Context()
	merchantID, err := h.parseID(c, "merchant_id")
	if err != nil {
		return err
	}
	year, err := h.parseYear(c)
	if err != nil {
		return err
	}

	cacheKey := fmt.Sprintf("apigw:category-stats:month-price:merchant:%d:%d", merchantID, year)
	if cached, found := gateway_cache.Get[response.ApiResponseCategoryMonthPrice](ctx, h.cache, cacheKey); found && cached != nil {
		return c.JSON(http.StatusOK, cached)
	}

	res, err := h.clients.CategoryMerchant.FindMonthPriceByMerchant(ctx, &pbcategory.FindYearCategoryByMerchant{Year: int32(year), MerchantId: int32(merchantID)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponseCategoryMonthPrice(res)
	gateway_cache.Set(ctx, h.cache, cacheKey, so, 5*time.Minute)
	return c.JSON(http.StatusOK, so)
}

// @Summary Yearly price per category by merchant (category)
// @Tags Category Stats
// @Accept json
// @Produce json
// @Param merchant_id query int true "Merchant ID"
// @Param year query int true "Year (e.g. 2026)"
// @Success 200 {object} response.ApiResponseCategoryYearPrice
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/category-stats/by-merchant/year-price [get]
// @Security BearerAuth
func (h *statsHandleApi) FindCategoryYearPriceByMerchant(c echo.Context) error {
	ctx := c.Request().Context()
	merchantID, err := h.parseID(c, "merchant_id")
	if err != nil {
		return err
	}
	year, err := h.parseYear(c)
	if err != nil {
		return err
	}

	cacheKey := fmt.Sprintf("apigw:category-stats:year-price:merchant:%d:%d", merchantID, year)
	if cached, found := gateway_cache.Get[response.ApiResponseCategoryYearPrice](ctx, h.cache, cacheKey); found && cached != nil {
		return c.JSON(http.StatusOK, cached)
	}

	res, err := h.clients.CategoryMerchant.FindYearPriceByMerchant(ctx, &pbcategory.FindYearCategoryByMerchant{Year: int32(year), MerchantId: int32(merchantID)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponseCategoryYearPrice(res)
	gateway_cache.Set(ctx, h.cache, cacheKey, so, 5*time.Minute)
	return c.JSON(http.StatusOK, so)
}

// @Summary Monthly total prices by category id (category)
// @Tags Category Stats
// @Accept json
// @Produce json
// @Param category_id query int true "Category ID"
// @Param year query int true "Year (e.g. 2026)"
// @Param month query int true "Month (1-12)"
// @Success 200 {object} response.ApiResponseCategoryMonthlyTotalPrice
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/category-stats/by-id/monthly-total-prices [get]
// @Security BearerAuth
func (h *statsHandleApi) FindCategoryMonthlyTotalPricesById(c echo.Context) error {
	ctx := c.Request().Context()
	categoryID, err := h.parseID(c, "category_id")
	if err != nil {
		return err
	}
	year, month, err := h.parseYearMonth(c)
	if err != nil {
		return err
	}

	cacheKey := fmt.Sprintf("apigw:category-stats:monthly-total-prices:category:%d:%d:%d", categoryID, year, month)
	if cached, found := gateway_cache.Get[response.ApiResponseCategoryMonthlyTotalPrice](ctx, h.cache, cacheKey); found && cached != nil {
		return c.JSON(http.StatusOK, cached)
	}

	res, err := h.clients.CategoryById.FindMonthlyTotalPricesById(ctx, &pbcategory.FindYearMonthTotalPriceById{Year: int32(year), Month: int32(month), CategoryId: int32(categoryID)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponseCategoryMonthlyTotalPrice(res)
	gateway_cache.Set(ctx, h.cache, cacheKey, so, 5*time.Minute)
	return c.JSON(http.StatusOK, so)
}

// @Summary Yearly total prices by category id (category)
// @Tags Category Stats
// @Accept json
// @Produce json
// @Param category_id query int true "Category ID"
// @Param year query int true "Year (e.g. 2026)"
// @Success 200 {object} response.ApiResponseCategoryYearlyTotalPrice
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/category-stats/by-id/yearly-total-prices [get]
// @Security BearerAuth
func (h *statsHandleApi) FindCategoryYearlyTotalPricesById(c echo.Context) error {
	ctx := c.Request().Context()
	categoryID, err := h.parseID(c, "category_id")
	if err != nil {
		return err
	}
	year, err := h.parseYear(c)
	if err != nil {
		return err
	}

	cacheKey := fmt.Sprintf("apigw:category-stats:yearly-total-prices:category:%d:%d", categoryID, year)
	if cached, found := gateway_cache.Get[response.ApiResponseCategoryYearlyTotalPrice](ctx, h.cache, cacheKey); found && cached != nil {
		return c.JSON(http.StatusOK, cached)
	}

	res, err := h.clients.CategoryById.FindYearlyTotalPricesById(ctx, &pbcategory.FindYearTotalPriceById{Year: int32(year), CategoryId: int32(categoryID)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponseCategoryYearlyTotalPrice(res)
	gateway_cache.Set(ctx, h.cache, cacheKey, so, 5*time.Minute)
	return c.JSON(http.StatusOK, so)
}

// @Summary Monthly price by category id (category)
// @Tags Category Stats
// @Accept json
// @Produce json
// @Param category_id query int true "Category ID"
// @Param year query int true "Year (e.g. 2026)"
// @Success 200 {object} response.ApiResponseCategoryMonthPrice
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/category-stats/by-id/month-price [get]
// @Security BearerAuth
func (h *statsHandleApi) FindCategoryMonthPriceById(c echo.Context) error {
	ctx := c.Request().Context()
	categoryID, err := h.parseID(c, "category_id")
	if err != nil {
		return err
	}
	year, err := h.parseYear(c)
	if err != nil {
		return err
	}

	cacheKey := fmt.Sprintf("apigw:category-stats:month-price:category:%d:%d", categoryID, year)
	if cached, found := gateway_cache.Get[response.ApiResponseCategoryMonthPrice](ctx, h.cache, cacheKey); found && cached != nil {
		return c.JSON(http.StatusOK, cached)
	}

	res, err := h.clients.CategoryById.FindMonthPriceById(ctx, &pbcategory.FindYearCategoryById{Year: int32(year), CategoryId: int32(categoryID)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponseCategoryMonthPrice(res)
	gateway_cache.Set(ctx, h.cache, cacheKey, so, 5*time.Minute)
	return c.JSON(http.StatusOK, so)
}

// @Summary Yearly price by category id (category)
// @Tags Category Stats
// @Accept json
// @Produce json
// @Param category_id query int true "Category ID"
// @Param year query int true "Year (e.g. 2026)"
// @Success 200 {object} response.ApiResponseCategoryYearPrice
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/category-stats/by-id/year-price [get]
// @Security BearerAuth
func (h *statsHandleApi) FindCategoryYearPriceById(c echo.Context) error {
	ctx := c.Request().Context()
	categoryID, err := h.parseID(c, "category_id")
	if err != nil {
		return err
	}
	year, err := h.parseYear(c)
	if err != nil {
		return err
	}

	cacheKey := fmt.Sprintf("apigw:category-stats:year-price:category:%d:%d", categoryID, year)
	if cached, found := gateway_cache.Get[response.ApiResponseCategoryYearPrice](ctx, h.cache, cacheKey); found && cached != nil {
		return c.JSON(http.StatusOK, cached)
	}

	res, err := h.clients.CategoryById.FindYearPriceById(ctx, &pbcategory.FindYearCategoryById{Year: int32(year), CategoryId: int32(categoryID)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponseCategoryYearPrice(res)
	gateway_cache.Set(ctx, h.cache, cacheKey, so, 5*time.Minute)
	return c.JSON(http.StatusOK, so)
}

// --- Transaction stats ---

// @Summary Monthly successful transactions (transaction)
// @Tags Transaction Stats
// @Accept json
// @Produce json
// @Param year query int true "Year (e.g. 2026)"
// @Param month query int true "Month (1-12)"
// @Success 200 {object} response.ApiResponsesTransactionMonthSuccess
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/transaction-stats/month-status-success [get]
// @Security BearerAuth
func (h *statsHandleApi) FindTransactionMonthStatusSuccess(c echo.Context) error {
	ctx := c.Request().Context()
	year, month, err := h.parseYearMonth(c)
	if err != nil {
		return err
	}

	cacheKey := fmt.Sprintf("apigw:transaction-stats:month-status-success:%d:%d", year, month)
	if cached, found := gateway_cache.Get[response.ApiResponsesTransactionMonthSuccess](ctx, h.cache, cacheKey); found && cached != nil {
		return c.JSON(http.StatusOK, cached)
	}

	res, err := h.clients.TransactionStatus.FindMonthStatusSuccess(ctx, &pbtransaction.FindMonthlyTransactionStatus{Year: int32(year), Month: int32(month)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponsesTransactionMonthSuccess(res)
	gateway_cache.Set(ctx, h.cache, cacheKey, so, 5*time.Minute)
	return c.JSON(http.StatusOK, so)
}

// @Summary Yearly successful transactions (transaction)
// @Tags Transaction Stats
// @Accept json
// @Produce json
// @Param year query int true "Year (e.g. 2026)"
// @Success 200 {object} response.ApiResponsesTransactionYearSuccess
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/transaction-stats/year-status-success [get]
// @Security BearerAuth
func (h *statsHandleApi) FindTransactionYearStatusSuccess(c echo.Context) error {
	ctx := c.Request().Context()
	year, err := h.parseYear(c)
	if err != nil {
		return err
	}

	cacheKey := fmt.Sprintf("apigw:transaction-stats:year-status-success:%d", year)
	if cached, found := gateway_cache.Get[response.ApiResponsesTransactionYearSuccess](ctx, h.cache, cacheKey); found && cached != nil {
		return c.JSON(http.StatusOK, cached)
	}

	res, err := h.clients.TransactionStatus.FindYearStatusSuccess(ctx, &pbtransaction.FindYearlyTransactionStatus{Year: int32(year)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponsesTransactionYearSuccess(res)
	gateway_cache.Set(ctx, h.cache, cacheKey, so, 5*time.Minute)
	return c.JSON(http.StatusOK, so)
}

// @Summary Monthly failed transactions (transaction)
// @Tags Transaction Stats
// @Accept json
// @Produce json
// @Param year query int true "Year (e.g. 2026)"
// @Param month query int true "Month (1-12)"
// @Success 200 {object} response.ApiResponsesTransactionMonthFailed
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/transaction-stats/month-status-failed [get]
// @Security BearerAuth
func (h *statsHandleApi) FindTransactionMonthStatusFailed(c echo.Context) error {
	ctx := c.Request().Context()
	year, month, err := h.parseYearMonth(c)
	if err != nil {
		return err
	}

	cacheKey := fmt.Sprintf("apigw:transaction-stats:month-status-failed:%d:%d", year, month)
	if cached, found := gateway_cache.Get[response.ApiResponsesTransactionMonthFailed](ctx, h.cache, cacheKey); found && cached != nil {
		return c.JSON(http.StatusOK, cached)
	}

	res, err := h.clients.TransactionStatus.FindMonthStatusFailed(ctx, &pbtransaction.FindMonthlyTransactionStatus{Year: int32(year), Month: int32(month)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponsesTransactionMonthFailed(res)
	gateway_cache.Set(ctx, h.cache, cacheKey, so, 5*time.Minute)
	return c.JSON(http.StatusOK, so)
}

// @Summary Yearly failed transactions (transaction)
// @Tags Transaction Stats
// @Accept json
// @Produce json
// @Param year query int true "Year (e.g. 2026)"
// @Success 200 {object} response.ApiResponsesTransactionYearFailed
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/transaction-stats/year-status-failed [get]
// @Security BearerAuth
func (h *statsHandleApi) FindTransactionYearStatusFailed(c echo.Context) error {
	ctx := c.Request().Context()
	year, err := h.parseYear(c)
	if err != nil {
		return err
	}

	cacheKey := fmt.Sprintf("apigw:transaction-stats:year-status-failed:%d", year)
	if cached, found := gateway_cache.Get[response.ApiResponsesTransactionYearFailed](ctx, h.cache, cacheKey); found && cached != nil {
		return c.JSON(http.StatusOK, cached)
	}

	res, err := h.clients.TransactionStatus.FindYearStatusFailed(ctx, &pbtransaction.FindYearlyTransactionStatus{Year: int32(year)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponsesTransactionYearFailed(res)
	gateway_cache.Set(ctx, h.cache, cacheKey, so, 5*time.Minute)
	return c.JSON(http.StatusOK, so)
}

// @Summary Monthly payment methods for successful transactions (transaction)
// @Tags Transaction Stats
// @Accept json
// @Produce json
// @Param year query int true "Year (e.g. 2026)"
// @Param month query int true "Month (1-12)"
// @Success 200 {object} response.ApiResponsesTransactionMonthMethod
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/transaction-stats/month-method-success [get]
// @Security BearerAuth
func (h *statsHandleApi) FindTransactionMonthMethodSuccess(c echo.Context) error {
	ctx := c.Request().Context()
	year, month, err := h.parseYearMonth(c)
	if err != nil {
		return err
	}

	cacheKey := fmt.Sprintf("apigw:transaction-stats:month-method-success:%d:%d", year, month)
	if cached, found := gateway_cache.Get[response.ApiResponsesTransactionMonthMethod](ctx, h.cache, cacheKey); found && cached != nil {
		return c.JSON(http.StatusOK, cached)
	}

	res, err := h.clients.TransactionMethod.FindMonthMethodSuccess(ctx, &pbtransaction.MonthTransactionMethod{Year: int32(year), Month: int32(month)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponsesTransactionMonthMethod(res)
	gateway_cache.Set(ctx, h.cache, cacheKey, so, 5*time.Minute)
	return c.JSON(http.StatusOK, so)
}

// @Summary Yearly payment methods for successful transactions (transaction)
// @Tags Transaction Stats
// @Accept json
// @Produce json
// @Param year query int true "Year (e.g. 2026)"
// @Success 200 {object} response.ApiResponsesTransactionYearMethod
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/transaction-stats/year-method-success [get]
// @Security BearerAuth
func (h *statsHandleApi) FindTransactionYearMethodSuccess(c echo.Context) error {
	ctx := c.Request().Context()
	year, err := h.parseYear(c)
	if err != nil {
		return err
	}

	cacheKey := fmt.Sprintf("apigw:transaction-stats:year-method-success:%d", year)
	if cached, found := gateway_cache.Get[response.ApiResponsesTransactionYearMethod](ctx, h.cache, cacheKey); found && cached != nil {
		return c.JSON(http.StatusOK, cached)
	}

	res, err := h.clients.TransactionMethod.FindYearMethodSuccess(ctx, &pbtransaction.YearTransactionMethod{Year: int32(year)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponsesTransactionYearMethod(res)
	gateway_cache.Set(ctx, h.cache, cacheKey, so, 5*time.Minute)
	return c.JSON(http.StatusOK, so)
}

// @Summary Monthly payment methods for failed transactions (transaction)
// @Tags Transaction Stats
// @Accept json
// @Produce json
// @Param year query int true "Year (e.g. 2026)"
// @Param month query int true "Month (1-12)"
// @Success 200 {object} response.ApiResponsesTransactionMonthMethod
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/transaction-stats/month-method-failed [get]
// @Security BearerAuth
func (h *statsHandleApi) FindTransactionMonthMethodFailed(c echo.Context) error {
	ctx := c.Request().Context()
	year, month, err := h.parseYearMonth(c)
	if err != nil {
		return err
	}

	cacheKey := fmt.Sprintf("apigw:transaction-stats:month-method-failed:%d:%d", year, month)
	if cached, found := gateway_cache.Get[response.ApiResponsesTransactionMonthMethod](ctx, h.cache, cacheKey); found && cached != nil {
		return c.JSON(http.StatusOK, cached)
	}

	res, err := h.clients.TransactionMethod.FindMonthMethodFailed(ctx, &pbtransaction.MonthTransactionMethod{Year: int32(year), Month: int32(month)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponsesTransactionMonthMethod(res)
	gateway_cache.Set(ctx, h.cache, cacheKey, so, 5*time.Minute)
	return c.JSON(http.StatusOK, so)
}

// @Summary Yearly payment methods for failed transactions (transaction)
// @Tags Transaction Stats
// @Accept json
// @Produce json
// @Param year query int true "Year (e.g. 2026)"
// @Success 200 {object} response.ApiResponsesTransactionYearMethod
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/transaction-stats/year-method-failed [get]
// @Security BearerAuth
func (h *statsHandleApi) FindTransactionYearMethodFailed(c echo.Context) error {
	ctx := c.Request().Context()
	year, err := h.parseYear(c)
	if err != nil {
		return err
	}

	cacheKey := fmt.Sprintf("apigw:transaction-stats:year-method-failed:%d", year)
	if cached, found := gateway_cache.Get[response.ApiResponsesTransactionYearMethod](ctx, h.cache, cacheKey); found && cached != nil {
		return c.JSON(http.StatusOK, cached)
	}

	res, err := h.clients.TransactionMethod.FindYearMethodFailed(ctx, &pbtransaction.YearTransactionMethod{Year: int32(year)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponsesTransactionYearMethod(res)
	gateway_cache.Set(ctx, h.cache, cacheKey, so, 5*time.Minute)
	return c.JSON(http.StatusOK, so)
}

// @Summary Monthly successful transactions by merchant (transaction)
// @Tags Transaction Stats
// @Accept json
// @Produce json
// @Param merchant_id query int true "Merchant ID"
// @Param year query int true "Year (e.g. 2026)"
// @Param month query int true "Month (1-12)"
// @Success 200 {object} response.ApiResponsesTransactionMonthSuccess
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/transaction-stats/by-merchant/month-status-success [get]
// @Security BearerAuth
func (h *statsHandleApi) FindTransactionMonthStatusSuccessByMerchant(c echo.Context) error {
	ctx := c.Request().Context()
	merchantID, err := h.parseID(c, "merchant_id")
	if err != nil {
		return err
	}
	year, month, err := h.parseYearMonth(c)
	if err != nil {
		return err
	}

	cacheKey := fmt.Sprintf("apigw:transaction-stats:month-status-success:merchant:%d:%d:%d", merchantID, year, month)
	if cached, found := gateway_cache.Get[response.ApiResponsesTransactionMonthSuccess](ctx, h.cache, cacheKey); found && cached != nil {
		return c.JSON(http.StatusOK, cached)
	}

	res, err := h.clients.TransactionStatus.FindMonthStatusSuccessByMerchant(ctx, &pbtransaction.FindMonthlyTransactionStatusByMerchant{Year: int32(year), Month: int32(month), MerchantId: int32(merchantID)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponsesTransactionMonthSuccess(res)
	gateway_cache.Set(ctx, h.cache, cacheKey, so, 5*time.Minute)
	return c.JSON(http.StatusOK, so)
}

// @Summary Yearly successful transactions by merchant (transaction)
// @Tags Transaction Stats
// @Accept json
// @Produce json
// @Param merchant_id query int true "Merchant ID"
// @Param year query int true "Year (e.g. 2026)"
// @Success 200 {object} response.ApiResponsesTransactionYearSuccess
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/transaction-stats/by-merchant/year-status-success [get]
// @Security BearerAuth
func (h *statsHandleApi) FindTransactionYearStatusSuccessByMerchant(c echo.Context) error {
	ctx := c.Request().Context()
	merchantID, err := h.parseID(c, "merchant_id")
	if err != nil {
		return err
	}
	year, err := h.parseYear(c)
	if err != nil {
		return err
	}

	cacheKey := fmt.Sprintf("apigw:transaction-stats:year-status-success:merchant:%d:%d", merchantID, year)
	if cached, found := gateway_cache.Get[response.ApiResponsesTransactionYearSuccess](ctx, h.cache, cacheKey); found && cached != nil {
		return c.JSON(http.StatusOK, cached)
	}

	res, err := h.clients.TransactionStatus.FindYearStatusSuccessByMerchant(ctx, &pbtransaction.FindYearlyTransactionStatusByMerchant{Year: int32(year), MerchantId: int32(merchantID)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponsesTransactionYearSuccess(res)
	gateway_cache.Set(ctx, h.cache, cacheKey, so, 5*time.Minute)
	return c.JSON(http.StatusOK, so)
}

// @Summary Monthly failed transactions by merchant (transaction)
// @Tags Transaction Stats
// @Accept json
// @Produce json
// @Param merchant_id query int true "Merchant ID"
// @Param year query int true "Year (e.g. 2026)"
// @Param month query int true "Month (1-12)"
// @Success 200 {object} response.ApiResponsesTransactionMonthFailed
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/transaction-stats/by-merchant/month-status-failed [get]
// @Security BearerAuth
func (h *statsHandleApi) FindTransactionMonthStatusFailedByMerchant(c echo.Context) error {
	ctx := c.Request().Context()
	merchantID, err := h.parseID(c, "merchant_id")
	if err != nil {
		return err
	}
	year, month, err := h.parseYearMonth(c)
	if err != nil {
		return err
	}

	cacheKey := fmt.Sprintf("apigw:transaction-stats:month-status-failed:merchant:%d:%d:%d", merchantID, year, month)
	if cached, found := gateway_cache.Get[response.ApiResponsesTransactionMonthFailed](ctx, h.cache, cacheKey); found && cached != nil {
		return c.JSON(http.StatusOK, cached)
	}

	res, err := h.clients.TransactionStatus.FindMonthStatusFailedByMerchant(ctx, &pbtransaction.FindMonthlyTransactionStatusByMerchant{Year: int32(year), Month: int32(month), MerchantId: int32(merchantID)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponsesTransactionMonthFailed(res)
	gateway_cache.Set(ctx, h.cache, cacheKey, so, 5*time.Minute)
	return c.JSON(http.StatusOK, so)
}

// @Summary Yearly failed transactions by merchant (transaction)
// @Tags Transaction Stats
// @Accept json
// @Produce json
// @Param merchant_id query int true "Merchant ID"
// @Param year query int true "Year (e.g. 2026)"
// @Success 200 {object} response.ApiResponsesTransactionYearFailed
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/transaction-stats/by-merchant/year-status-failed [get]
// @Security BearerAuth
func (h *statsHandleApi) FindTransactionYearStatusFailedByMerchant(c echo.Context) error {
	ctx := c.Request().Context()
	merchantID, err := h.parseID(c, "merchant_id")
	if err != nil {
		return err
	}
	year, err := h.parseYear(c)
	if err != nil {
		return err
	}

	cacheKey := fmt.Sprintf("apigw:transaction-stats:year-status-failed:merchant:%d:%d", merchantID, year)
	if cached, found := gateway_cache.Get[response.ApiResponsesTransactionYearFailed](ctx, h.cache, cacheKey); found && cached != nil {
		return c.JSON(http.StatusOK, cached)
	}

	res, err := h.clients.TransactionStatus.FindYearStatusFailedByMerchant(ctx, &pbtransaction.FindYearlyTransactionStatusByMerchant{Year: int32(year), MerchantId: int32(merchantID)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponsesTransactionYearFailed(res)
	gateway_cache.Set(ctx, h.cache, cacheKey, so, 5*time.Minute)
	return c.JSON(http.StatusOK, so)
}

// @Summary Monthly payment methods (success) by merchant (transaction)
// @Tags Transaction Stats
// @Accept json
// @Produce json
// @Param merchant_id query int true "Merchant ID"
// @Param year query int true "Year (e.g. 2026)"
// @Param month query int true "Month (1-12)"
// @Success 200 {object} response.ApiResponsesTransactionMonthMethod
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/transaction-stats/by-merchant/month-method-success [get]
// @Security BearerAuth
func (h *statsHandleApi) FindTransactionMonthMethodSuccessByMerchant(c echo.Context) error {
	ctx := c.Request().Context()
	merchantID, err := h.parseID(c, "merchant_id")
	if err != nil {
		return err
	}
	year, month, err := h.parseYearMonth(c)
	if err != nil {
		return err
	}

	cacheKey := fmt.Sprintf("apigw:transaction-stats:month-method-success:merchant:%d:%d:%d", merchantID, year, month)
	if cached, found := gateway_cache.Get[response.ApiResponsesTransactionMonthMethod](ctx, h.cache, cacheKey); found && cached != nil {
		return c.JSON(http.StatusOK, cached)
	}

	res, err := h.clients.TransactionMethod.FindMonthMethodByMerchantSuccess(ctx, &pbtransaction.MonthTransactionMethodByMerchant{Year: int32(year), Month: int32(month), MerchantId: int32(merchantID)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponsesTransactionMonthMethod(res)
	gateway_cache.Set(ctx, h.cache, cacheKey, so, 5*time.Minute)
	return c.JSON(http.StatusOK, so)
}

// @Summary Yearly payment methods (success) by merchant (transaction)
// @Tags Transaction Stats
// @Accept json
// @Produce json
// @Param merchant_id query int true "Merchant ID"
// @Param year query int true "Year (e.g. 2026)"
// @Success 200 {object} response.ApiResponsesTransactionYearMethod
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/transaction-stats/by-merchant/year-method-success [get]
// @Security BearerAuth
func (h *statsHandleApi) FindTransactionYearMethodSuccessByMerchant(c echo.Context) error {
	ctx := c.Request().Context()
	merchantID, err := h.parseID(c, "merchant_id")
	if err != nil {
		return err
	}
	year, err := h.parseYear(c)
	if err != nil {
		return err
	}

	cacheKey := fmt.Sprintf("apigw:transaction-stats:year-method-success:merchant:%d:%d", merchantID, year)
	if cached, found := gateway_cache.Get[response.ApiResponsesTransactionYearMethod](ctx, h.cache, cacheKey); found && cached != nil {
		return c.JSON(http.StatusOK, cached)
	}

	res, err := h.clients.TransactionMethod.FindYearMethodByMerchantSuccess(ctx, &pbtransaction.YearTransactionMethodByMerchant{Year: int32(year), MerchantId: int32(merchantID)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponsesTransactionYearMethod(res)
	gateway_cache.Set(ctx, h.cache, cacheKey, so, 5*time.Minute)
	return c.JSON(http.StatusOK, so)
}

// @Summary Monthly payment methods (failed) by merchant (transaction)
// @Tags Transaction Stats
// @Accept json
// @Produce json
// @Param merchant_id query int true "Merchant ID"
// @Param year query int true "Year (e.g. 2026)"
// @Param month query int true "Month (1-12)"
// @Success 200 {object} response.ApiResponsesTransactionMonthMethod
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/transaction-stats/by-merchant/month-method-failed [get]
// @Security BearerAuth
func (h *statsHandleApi) FindTransactionMonthMethodFailedByMerchant(c echo.Context) error {
	ctx := c.Request().Context()
	merchantID, err := h.parseID(c, "merchant_id")
	if err != nil {
		return err
	}
	year, month, err := h.parseYearMonth(c)
	if err != nil {
		return err
	}

	cacheKey := fmt.Sprintf("apigw:transaction-stats:month-method-failed:merchant:%d:%d:%d", merchantID, year, month)
	if cached, found := gateway_cache.Get[response.ApiResponsesTransactionMonthMethod](ctx, h.cache, cacheKey); found && cached != nil {
		return c.JSON(http.StatusOK, cached)
	}

	res, err := h.clients.TransactionMethod.FindMonthMethodByMerchantFailed(ctx, &pbtransaction.MonthTransactionMethodByMerchant{Year: int32(year), Month: int32(month), MerchantId: int32(merchantID)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponsesTransactionMonthMethod(res)
	gateway_cache.Set(ctx, h.cache, cacheKey, so, 5*time.Minute)
	return c.JSON(http.StatusOK, so)
}

// @Summary Yearly payment methods (failed) by merchant (transaction)
// @Tags Transaction Stats
// @Accept json
// @Produce json
// @Param merchant_id query int true "Merchant ID"
// @Param year query int true "Year (e.g. 2026)"
// @Success 200 {object} response.ApiResponsesTransactionYearMethod
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/transaction-stats/by-merchant/year-method-failed [get]
// @Security BearerAuth
func (h *statsHandleApi) FindTransactionYearMethodFailedByMerchant(c echo.Context) error {
	ctx := c.Request().Context()
	merchantID, err := h.parseID(c, "merchant_id")
	if err != nil {
		return err
	}
	year, err := h.parseYear(c)
	if err != nil {
		return err
	}

	cacheKey := fmt.Sprintf("apigw:transaction-stats:year-method-failed:merchant:%d:%d", merchantID, year)
	if cached, found := gateway_cache.Get[response.ApiResponsesTransactionYearMethod](ctx, h.cache, cacheKey); found && cached != nil {
		return c.JSON(http.StatusOK, cached)
	}

	res, err := h.clients.TransactionMethod.FindYearMethodByMerchantFailed(ctx, &pbtransaction.YearTransactionMethodByMerchant{Year: int32(year), MerchantId: int32(merchantID)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponsesTransactionYearMethod(res)
	gateway_cache.Set(ctx, h.cache, cacheKey, so, 5*time.Minute)
	return c.JSON(http.StatusOK, so)
}
