package protomapper

import (
	pbauth "github.com/MamangRust/microservice-point-of-sale-pb/auth"
	pbcashier "github.com/MamangRust/microservice-point-of-sale-pb/cashier"
	pbcategory "github.com/MamangRust/microservice-point-of-sale-pb/category"
	pbcommon "github.com/MamangRust/microservice-point-of-sale-pb/common"
	pbmerchant "github.com/MamangRust/microservice-point-of-sale-pb/merchant"
	pbmerchantdoc "github.com/MamangRust/microservice-point-of-sale-pb/merchant_document"
	pborder "github.com/MamangRust/microservice-point-of-sale-pb/order"
	pborderitem "github.com/MamangRust/microservice-point-of-sale-pb/order_item"
	pbproduct "github.com/MamangRust/microservice-point-of-sale-pb/product"
	pbrole "github.com/MamangRust/microservice-point-of-sale-pb/role"
	pbtransaction "github.com/MamangRust/microservice-point-of-sale-pb/transaction"
	pbuser "github.com/MamangRust/microservice-point-of-sale-pb/user"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/response"
)

type AuthProtoMapper interface {
	ToProtoResponseVerifyCode(status string, message string) *pbauth.ApiResponseVerifyCode
	ToProtoResponseForgotPassword(status string, message string) *pbauth.ApiResponseForgotPassword
	ToProtoResponseResetPassword(status string, message string) *pbauth.ApiResponseResetPassword
	ToProtoResponseLogin(status string, message string, response *response.TokenResponse) *pbauth.ApiResponseLogin
	ToProtoResponseRegister(status string, message string, response *response.UserResponse) *pbauth.ApiResponseRegister
	ToProtoResponseRefreshToken(status string, message string, response *response.TokenResponse) *pbauth.ApiResponseRefreshToken
	ToProtoResponseGetMe(status string, message string, response *response.UserResponse) *pbauth.ApiResponseGetMe
}

type UserProtoMapper interface {
	ToProtoResponseUserDeleteAt(status string, message string, pbResponse *response.UserResponseDeleteAt) *pbuser.ApiResponseUserDeleteAt
	ToProtoResponsesUser(status string, message string, pbResponse []*response.UserResponse) *pbuser.ApiResponsesUser
	ToProtoResponseUser(status string, message string, pbResponse *response.UserResponse) *pbuser.ApiResponseUser
	ToProtoResponseUserDelete(status string, message string) *pbuser.ApiResponseUserDelete
	ToProtoResponseUserAll(status string, message string) *pbuser.ApiResponseUserAll
	ToProtoResponsePaginationUserDeleteAt(pagination *pbcommon.PaginationMeta, status string, message string, users []*response.UserResponseDeleteAt) *pbuser.ApiResponsePaginationUserDeleteAt
	ToProtoResponsePaginationUser(pagination *pbcommon.PaginationMeta, status string, message string, users []*response.UserResponse) *pbuser.ApiResponsePaginationUser
}

type RoleProtoMapper interface {
	ToProtoResponseRoleAll(status string, message string) *pbrole.ApiResponseRoleAll
	ToProtoResponseRoleDelete(status string, message string) *pbrole.ApiResponseRoleDelete
	ToProtoResponseRole(status string, message string, pbResponse *response.RoleResponse) *pbrole.ApiResponseRole
	ToProtoResponsesRole(status string, message string, pbResponse []*response.RoleResponse) *pbrole.ApiResponsesRole
	ToProtoResponsePaginationRole(pagination *pbcommon.PaginationMeta, status string, message string, pbResponse []*response.RoleResponse) *pbrole.ApiResponsePaginationRole
	ToProtoResponsePaginationRoleDeleteAt(pagination *pbcommon.PaginationMeta, status string, message string, pbResponse []*response.RoleResponseDeleteAt) *pbrole.ApiResponsePaginationRoleDeleteAt
}

type CategoryProtoMapper interface {
	ToProtoResponsesCategory(status string, message string, pbResponse []*response.CategoryResponse) *pbcategory.ApiResponsesCategory
	ToProtoResponseCategoryDeleteAt(status string, message string, pbResponse *response.CategoryResponseDeleteAt) *pbcategory.ApiResponseCategoryDeleteAt

	ToProtoResponseCategoryAll(status string, message string) *pbcategory.ApiResponseCategoryAll
	ToProtoResponseCategory(status string, message string, pbResponse *response.CategoryResponse) *pbcategory.ApiResponseCategory
	ToProtoResponseCategoryDelete(status string, message string) *pbcategory.ApiResponseCategoryDelete
	ToProtoResponsePaginationCategoryDeleteAt(pagination *pbcommon.PaginationMeta, status string, message string, categories []*response.CategoryResponseDeleteAt) *pbcategory.ApiResponsePaginationCategoryDeleteAt
	ToProtoResponsePaginationCategory(pagination *pbcommon.PaginationMeta, status string, message string, categories []*response.CategoryResponse) *pbcategory.ApiResponsePaginationCategory
}

type CashierProtoMapper interface {
	ToProtoResponseCashier(status string, message string, pbResponse *response.CashierResponse) *pbcashier.ApiResponseCashier
	ToProtoResponseCashierDeleteAt(status string, message string, pbResponse *response.CashierResponseDeleteAt) *pbcashier.ApiResponseCashierDeleteAt
	ToProtoResponsesCashier(status string, message string, pbResponse []*response.CashierResponse) *pbcashier.ApiResponsesCashier
	ToProtoResponseCashierDelete(status string, message string) *pbcashier.ApiResponseCashierDelete
	ToProtoResponseCashierAll(status string, message string) *pbcashier.ApiResponseCashierAll
	ToProtoResponsePaginationCashierDeleteAt(pagination *pbcommon.PaginationMeta, status string, message string, users []*response.CashierResponseDeleteAt) *pbcashier.ApiResponsePaginationCashierDeleteAt
	ToProtoResponsePaginationCashier(pagination *pbcommon.PaginationMeta, status string, message string, users []*response.CashierResponse) *pbcashier.ApiResponsePaginationCashier
}

type MerchantProtoMapper interface {
	ToProtoResponseMerchant(status string, message string, pbResponse *response.MerchantResponse) *pbmerchant.ApiResponseMerchant
	ToProtoResponseMerchantDeleteAt(status string, message string, pbResponse *response.MerchantResponseDeleteAt) *pbmerchant.ApiResponseMerchantDeleteAt

	ToProtoResponsesMerchant(status string, message string, pbResponse []*response.MerchantResponse) *pbmerchant.ApiResponsesMerchant
	ToProtoResponseMerchantDelete(status string, message string) *pbmerchant.ApiResponseMerchantDelete
	ToProtoResponseMerchantAll(status string, message string) *pbmerchant.ApiResponseMerchantAll
	ToProtoResponsePaginationMerchantDeleteAt(pagination *pbcommon.PaginationMeta, status string, message string, merchants []*response.MerchantResponseDeleteAt) *pbmerchant.ApiResponsePaginationMerchantDeleteAt
	ToProtoResponsePaginationMerchant(pagination *pbcommon.PaginationMeta, status string, message string, merchants []*response.MerchantResponse) *pbmerchant.ApiResponsePaginationMerchant
}

type MerchantDocumentProtoMapper interface {
	ToProtoResponseMerchantDocument(status string, message string, doc *response.MerchantDocumentResponse) *pbmerchantdoc.ApiResponseMerchantDocument
	ToProtoResponsesMerchantDocument(status string, message string, docs []*response.MerchantDocumentResponse) *pbmerchantdoc.ApiResponsesMerchantDocument

	ToProtoResponsePaginationMerchantDocument(pagination *pbcommon.PaginationMeta, status string, message string, docs []*response.MerchantDocumentResponse) *pbmerchantdoc.ApiResponsePaginationMerchantDocument
	ToProtoResponsePaginationMerchantDocumentDeleteAt(pagination *pbcommon.PaginationMeta, status string, message string, docs []*response.MerchantDocumentResponseDeleteAt) *pbmerchantdoc.ApiResponsePaginationMerchantDocumentAt

	ToProtoResponseMerchantDocumentDelete(status string, message string) *pbmerchantdoc.ApiResponseMerchantDocumentDelete

	ToProtoResponseMerchantDocumentAll(status string, message string) *pbmerchantdoc.ApiResponseMerchantDocumentAll
}

type OrderItemProtoMapper interface {
	ToProtoResponseOrderItem(status string, message string, pbResponse *response.OrderItemResponse) *pborderitem.ApiResponseOrderItem
	ToProtoResponsesOrderItem(status string, message string, pbResponse []*response.OrderItemResponse) *pborderitem.ApiResponsesOrderItem
	ToProtoResponseOrderItemDelete(status string, message string) *pborderitem.ApiResponseOrderItemDelete
	ToProtoResponseOrderItemAll(status string, message string) *pborderitem.ApiResponseOrderItemAll
	ToProtoResponsePaginationOrderItemDeleteAt(pagination *pbcommon.PaginationMeta, status string, message string, orderItems []*response.OrderItemResponseDeleteAt) *pborderitem.ApiResponsePaginationOrderItemDeleteAt
	ToProtoResponsePaginationOrderItem(pagination *pbcommon.PaginationMeta, status string, message string, orderItems []*response.OrderItemResponse) *pborderitem.ApiResponsePaginationOrderItem
}

type OrderProtoMapper interface {
	ToProtoResponseOrder(status string, message string, pbResponse *response.OrderResponse) *pborder.ApiResponseOrder
	ToProtoResponseOrderDeleteAt(status string, message string, pbResponse *response.OrderResponseDeleteAt) *pborder.ApiResponseOrderDeleteAt
	ToProtoResponsesOrder(status string, message string, pbResponse []*response.OrderResponse) *pborder.ApiResponsesOrder
	ToProtoResponseOrderDelete(status string, message string) *pborder.ApiResponseOrderDelete
	ToProtoResponseOrderAll(status string, message string) *pborder.ApiResponseOrderAll
	ToProtoResponsePaginationOrderDeleteAt(pagination *pbcommon.PaginationMeta, status string, message string, orders []*response.OrderResponseDeleteAt) *pborder.ApiResponsePaginationOrderDeleteAt
	ToProtoResponsePaginationOrder(pagination *pbcommon.PaginationMeta, status string, message string, orders []*response.OrderResponse) *pborder.ApiResponsePaginationOrder
}

type ProductProtoMapper interface {
	ToProtoResponseProduct(status string, message string, pbResponse *response.ProductResponse) *pbproduct.ApiResponseProduct
	ToProtoResponseProductDeleteAt(status string, message string, pbResponse *response.ProductResponseDeleteAt) *pbproduct.ApiResponseProductDeleteAt

	ToProtoResponsesProduct(status string, message string, pbResponse []*response.ProductResponse) *pbproduct.ApiResponsesProduct
	ToProtoResponseProductDelete(status string, message string) *pbproduct.ApiResponseProductDelete
	ToProtoResponseProductAll(status string, message string) *pbproduct.ApiResponseProductAll
	ToProtoResponsePaginationProductDeleteAt(pagination *pbcommon.PaginationMeta, status string, message string, products []*response.ProductResponseDeleteAt) *pbproduct.ApiResponsePaginationProductDeleteAt
	ToProtoResponsePaginationProduct(pagination *pbcommon.PaginationMeta, status string, message string, products []*response.ProductResponse) *pbproduct.ApiResponsePaginationProduct
}

type TransactionProtoMapper interface {
	ToProtoResponseTransaction(status string, message string, trans *response.TransactionResponse) *pbtransaction.ApiResponseTransaction
	ToProtoResponseTransactionDeleteAt(status string, message string, trans *response.TransactionResponseDeleteAt) *pbtransaction.ApiResponseTransactionDeleteAt
	ToProtoResponsesTransaction(status string, message string, transList []*response.TransactionResponse) *pbtransaction.ApiResponsesTransaction
	ToProtoResponseTransactionDelete(status string, message string) *pbtransaction.ApiResponseTransactionDelete
	ToProtoResponseTransactionAll(status string, message string) *pbtransaction.ApiResponseTransactionAll
	ToProtoResponsePaginationTransactionDeleteAt(pagination *pbcommon.PaginationMeta, status string, message string, transactions []*response.TransactionResponseDeleteAt) *pbtransaction.ApiResponsePaginationTransactionDeleteAt
	ToProtoResponsePaginationTransaction(pagination *pbcommon.PaginationMeta, status string, message string, transactions []*response.TransactionResponse) *pbtransaction.ApiResponsePaginationTransaction
}
