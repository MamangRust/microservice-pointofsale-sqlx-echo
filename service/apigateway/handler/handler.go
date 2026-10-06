package handler

import (
	"fmt"
	"strconv"

	auth_cache "github.com/MamangRust/microservice-point-of-sale-apigateway/cache/auth"
	gateway_cache "github.com/MamangRust/microservice-point-of-sale-apigateway/cache/gateway_cache"
	pbauth "github.com/MamangRust/microservice-point-of-sale-pb/auth"
	pbcashier "github.com/MamangRust/microservice-point-of-sale-pb/cashier"
	pbcategory "github.com/MamangRust/microservice-point-of-sale-pb/category"
	pbmerchant "github.com/MamangRust/microservice-point-of-sale-pb/merchant"
	pbmerchantdoc "github.com/MamangRust/microservice-point-of-sale-pb/merchant_document"
	pborder "github.com/MamangRust/microservice-point-of-sale-pb/order"
	pborderitem "github.com/MamangRust/microservice-point-of-sale-pb/order_item"
	pbproduct "github.com/MamangRust/microservice-point-of-sale-pb/product"
	pbrole "github.com/MamangRust/microservice-point-of-sale-pb/role"
	pbstats "github.com/MamangRust/microservice-point-of-sale-pb/stats"
	pbtransaction "github.com/MamangRust/microservice-point-of-sale-pb/transaction"
	pbuser "github.com/MamangRust/microservice-point-of-sale-pb/user"
	pbuserrole "github.com/MamangRust/microservice-point-of-sale-pb/user_role"
	"github.com/MamangRust/microservice-point-of-sale-pkg/auth"
	"github.com/MamangRust/microservice-point-of-sale-pkg/logger"
	"github.com/MamangRust/microservice-point-of-sale-pkg/upload_image"
	"github.com/MamangRust/microservice-point-of-sale-shared/errors"
	"github.com/labstack/echo/v4"
	"google.golang.org/grpc"

	response_api "github.com/MamangRust/microservice-point-of-sale-shared/mapper/response/api"
)

type ServiceConnections struct {
	Auth        *grpc.ClientConn
	Role        *grpc.ClientConn
	User        *grpc.ClientConn
	Cashier     *grpc.ClientConn
	Category    *grpc.ClientConn
	Merchant    *grpc.ClientConn
	OrderItem   *grpc.ClientConn
	Order       *grpc.ClientConn
	Product     *grpc.ClientConn
	Transaction *grpc.ClientConn
	StatsReader *grpc.ClientConn
}

type Deps struct {
	Token              auth.TokenManager
	E                  *echo.Echo
	Logger             logger.LoggerInterface
	Mapping            *response_api.ResponseApiMapper
	ImageUpload        upload_image.ImageUploads
	ServiceConnections *ServiceConnections
	ApiHandler         errors.ApiHandler
	AuthCache          auth_cache.AuthMencache
	GatewayCache       *gateway_cache.GatewayCache
}

func NewHandler(deps *Deps) {
	clientAuth := pbauth.NewAuthServiceClient(deps.ServiceConnections.Auth)
	clientRole := pbrole.NewRoleServiceClient(deps.ServiceConnections.Role)
	clientUserRole := pbuserrole.NewUserRoleServiceClient(deps.ServiceConnections.Role)
	clientUser := pbuser.NewUserServiceClient(deps.ServiceConnections.User)
	clientCategory := pbcategory.NewCategoryServiceClient(deps.ServiceConnections.Category)
	clientCashier := pbcashier.NewCashierServiceClient(deps.ServiceConnections.Cashier)
	clientMerchant := pbmerchant.NewMerchantServiceClient(deps.ServiceConnections.Merchant)
	clientMerchantDocument := pbmerchantdoc.NewMerchantDocumentServiceClient(deps.ServiceConnections.Merchant)
	clientOrderItem := pborderitem.NewOrderItemServiceClient(deps.ServiceConnections.OrderItem)
	clientOrder := pborder.NewOrderServiceClient(deps.ServiceConnections.Order)
	clientProduct := pbproduct.NewProductServiceClient(deps.ServiceConnections.Product)
	clientTransaction := pbtransaction.NewTransactionServiceClient(deps.ServiceConnections.Transaction)

	NewHandlerAuth(deps.E, clientAuth, deps.Logger, deps.Mapping.AuthResponseMapper, deps.ApiHandler, deps.AuthCache)
	NewHandlerRole(deps.E, clientRole, clientUserRole, deps.Logger, deps.Mapping.RoleResponseMapper, deps.ApiHandler, deps.GatewayCache)
	NewHandlerUser(deps.E, clientUser, deps.Logger, deps.Mapping.UserResponseMapper, deps.ApiHandler, deps.GatewayCache)
	NewHandlerCategory(deps.E, clientCategory, deps.Logger, deps.Mapping.CategoryResponseMapper, deps.ApiHandler, deps.GatewayCache)
	NewHandlerCashier(deps.E, clientCashier, deps.Logger, deps.Mapping.CashierResponseMapper, deps.ApiHandler, deps.GatewayCache)
	NewHandlerMerchant(deps.E, clientMerchant, deps.Logger, deps.Mapping.MerchantResponseMapper, deps.ApiHandler, deps.GatewayCache)
	NewHandlerMerchantDocument(deps.E, clientMerchantDocument, deps.Logger, deps.Mapping.MerchantDocumentProMapper, deps.ApiHandler, deps.GatewayCache)
	NewHandlerOrderItem(deps.E, clientOrderItem, deps.Logger, deps.Mapping.OrderItemResponseMapper, deps.ApiHandler, deps.GatewayCache)
	NewHandlerOrder(deps.E, clientOrder, deps.Logger, deps.Mapping.OrderResponseMapper, deps.ApiHandler, deps.GatewayCache)
	NewHandlerProduct(deps.E, clientProduct, deps.Logger, deps.Mapping.ProductResponseMapper, deps.ImageUpload, deps.ApiHandler, deps.GatewayCache)
	NewHandlerTransaction(deps.E, clientTransaction, deps.Logger, deps.Mapping.TransactionResponseMapper, deps.ApiHandler, deps.GatewayCache)

	statsClients := StatsClients{
		Order:             pbstats.NewOrderStatsServiceClient(deps.ServiceConnections.StatsReader),
		OrderMerchant:     pbstats.NewOrderStatsByMerchantServiceClient(deps.ServiceConnections.StatsReader),
		OrderById:         pbstats.NewOrderStatsByIdServiceClient(deps.ServiceConnections.StatsReader),
		Cashier:           pbstats.NewCashierStatsServiceClient(deps.ServiceConnections.StatsReader),
		CashierMerchant:   pbstats.NewCashierStatsByMerchantServiceClient(deps.ServiceConnections.StatsReader),
		CashierById:       pbstats.NewCashierStatsByIdServiceClient(deps.ServiceConnections.StatsReader),
		Category:          pbstats.NewCategoryStatsServiceClient(deps.ServiceConnections.StatsReader),
		CategoryMerchant:  pbstats.NewCategoryStatsByMerchantServiceClient(deps.ServiceConnections.StatsReader),
		CategoryById:      pbstats.NewCategoryStatsByIdServiceClient(deps.ServiceConnections.StatsReader),
		TransactionStatus: pbstats.NewTransactionStatsStatusServiceClient(deps.ServiceConnections.StatsReader),
		TransactionMethod: pbstats.NewTransactionStatsMethodServiceClient(deps.ServiceConnections.StatsReader),
	}

	NewHandlerStats(deps.E, statsClients, deps.Logger, deps.Mapping.StatsResponseMapper, deps.ApiHandler, deps.GatewayCache)
}

func parseQueryInt(c echo.Context, key string, defaultValue int) int {
	val, err := strconv.Atoi(c.QueryParam(key))
	if err != nil || val <= 0 {
		return defaultValue
	}
	return val
}

func parseQueryIntWithValidation(c echo.Context, key string, min, max int) (int, error) {
	valStr := c.QueryParam(key)
	val, err := strconv.Atoi(valStr)
	if err != nil || val < min || val > max {
		return 0, fmt.Errorf("invalid %s: %s", key, valStr)
	}
	return val, nil
}
