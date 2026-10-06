package response_api

import (
	pbauth "github.com/MamangRust/microservice-point-of-sale-pb/auth"
	pbcashier "github.com/MamangRust/microservice-point-of-sale-pb/cashier"
	pbcategory "github.com/MamangRust/microservice-point-of-sale-pb/category"
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

type AuthResponseMapper interface {
	ToResponseVerifyCode(res *pbauth.ApiResponseVerifyCode) *response.ApiResponseVerifyCode
	ToResponseForgotPassword(res *pbauth.ApiResponseForgotPassword) *response.ApiResponseForgotPassword
	ToResponseResetPassword(res *pbauth.ApiResponseResetPassword) *response.ApiResponseResetPassword
	ToResponseLogin(res *pbauth.ApiResponseLogin) *response.ApiResponseLogin
	ToResponseRegister(res *pbauth.ApiResponseRegister) *response.ApiResponseRegister
	ToResponseRefreshToken(res *pbauth.ApiResponseRefreshToken) *response.ApiResponseRefreshToken
	ToResponseGetMe(res *pbauth.ApiResponseGetMe) *response.ApiResponseGetMe
}

type RoleResponseMapper interface {
	ToApiResponseRoleAll(pbResponse *pbrole.ApiResponseRoleAll) *response.ApiResponseRoleAll
	ToApiResponseRoleDelete(pbResponse *pbrole.ApiResponseRoleDelete) *response.ApiResponseRoleDelete
	ToApiResponseRole(pbResponse *pbrole.ApiResponseRole) *response.ApiResponseRole
	ToApiResponsesRole(pbResponse *pbrole.ApiResponsesRole) *response.ApiResponsesRole
	ToApiResponsePaginationRole(pbResponse *pbrole.ApiResponsePaginationRole) *response.ApiResponsePaginationRole
	ToApiResponsePaginationRoleDeleteAt(pbResponse *pbrole.ApiResponsePaginationRoleDeleteAt) *response.ApiResponsePaginationRoleDeleteAt
}

type UserResponseMapper interface {
	ToApiResponseUserDeleteAt(pbResponse *pbuser.ApiResponseUserDeleteAt) *response.ApiResponseUserDeleteAt
	ToApiResponseUser(pbResponse *pbuser.ApiResponseUser) *response.ApiResponseUser
	ToApiResponsesUser(pbResponse *pbuser.ApiResponsesUser) *response.ApiResponsesUser

	ToApiResponseUserDelete(pbResponse *pbuser.ApiResponseUserDelete) *response.ApiResponseUserDelete
	ToApiResponseUserAll(pbResponse *pbuser.ApiResponseUserAll) *response.ApiResponseUserAll
	ToApiResponsePaginationUserDeleteAt(pbResponse *pbuser.ApiResponsePaginationUserDeleteAt) *response.ApiResponsePaginationUserDeleteAt
	ToApiResponsePaginationUser(pbResponse *pbuser.ApiResponsePaginationUser) *response.ApiResponsePaginationUser
}

type CategoryResponseMapper interface {
	ToApiResponseCategory(pbResponse *pbcategory.ApiResponseCategory) *response.ApiResponseCategory
	ToApiResponseCategoryDeleteAt(pbResponse *pbcategory.ApiResponseCategoryDeleteAt) *response.ApiResponseCategoryDeleteAt
	ToApiResponsesCategory(pbResponse *pbcategory.ApiResponsesCategory) *response.ApiResponsesCategory
	ToApiResponseCategoryDelete(pbResponse *pbcategory.ApiResponseCategoryDelete) *response.ApiResponseCategoryDelete
	ToApiResponseCategoryAll(pbResponse *pbcategory.ApiResponseCategoryAll) *response.ApiResponseCategoryAll
	ToApiResponsePaginationCategoryDeleteAt(pbResponse *pbcategory.ApiResponsePaginationCategoryDeleteAt) *response.ApiResponsePaginationCategoryDeleteAt
	ToApiResponsePaginationCategory(pbResponse *pbcategory.ApiResponsePaginationCategory) *response.ApiResponsePaginationCategory
}

type CashierResponseMapper interface {
	ToApiResponseCashier(pbResponse *pbcashier.ApiResponseCashier) *response.ApiResponseCashier
	ToApiResponsesCashier(pbResponse *pbcashier.ApiResponsesCashier) *response.ApiResponsesCashier
	ToApiResponseCashierAll(pbResponse *pbcashier.ApiResponseCashierAll) *response.ApiResponseCashierAll
	ToApiResponseCashierDelete(pbResponse *pbcashier.ApiResponseCashierDelete) *response.ApiResponseCashierDelete
	ToApiResponseCashierDeleteAt(pbResponse *pbcashier.ApiResponseCashierDeleteAt) *response.ApiResponseCashierDeleteAt
	ToApiResponsePaginationCashierDeleteAt(pbResponse *pbcashier.ApiResponsePaginationCashierDeleteAt) *response.ApiResponsePaginationCashierDeleteAt
	ToApiResponsePaginationCashier(pbResponse *pbcashier.ApiResponsePaginationCashier) *response.ApiResponsePaginationCashier
}

type MerchantResponseMapper interface {
	ToApiResponseMerchant(pbResponse *pbmerchant.ApiResponseMerchant) *response.ApiResponseMerchant

	ToApiResponseMerchantDeleteAt(pbResponse *pbmerchant.ApiResponseMerchantDeleteAt) *response.ApiResponseMerchantDeleteAt
	ToApiResponsesMerchant(pbResponse *pbmerchant.ApiResponsesMerchant) *response.ApiResponsesMerchant
	ToApiResponseMerchantDelete(pbResponse *pbmerchant.ApiResponseMerchantDelete) *response.ApiResponseMerchantDelete
	ToApiResponseMerchantAll(pbResponse *pbmerchant.ApiResponseMerchantAll) *response.ApiResponseMerchantAll
	ToApiResponsePaginationMerchantDeleteAt(pbResponse *pbmerchant.ApiResponsePaginationMerchantDeleteAt) *response.ApiResponsePaginationMerchantDeleteAt
	ToApiResponsePaginationMerchant(pbResponse *pbmerchant.ApiResponsePaginationMerchant) *response.ApiResponsePaginationMerchant
}

type MerchantDocumentResponseMapper interface {
	ToApiResponseMerchantDocument(doc *pbmerchantdoc.ApiResponseMerchantDocument) *response.ApiResponseMerchantDocument
	ToApiResponsesMerchantDocument(docs *pbmerchantdoc.ApiResponsesMerchantDocument) *response.ApiResponsesMerchantDocument

	ToApiResponsePaginationMerchantDocument(docs *pbmerchantdoc.ApiResponsePaginationMerchantDocument) *response.ApiResponsePaginationMerchantDocument
	ToApiResponsePaginationMerchantDocumentDeleteAt(docs *pbmerchantdoc.ApiResponsePaginationMerchantDocumentAt) *response.ApiResponsePaginationMerchantDocumentDeleteAt

	ToApiResponseMerchantDocumentAll(resp *pbmerchantdoc.ApiResponseMerchantDocumentAll) *response.ApiResponseMerchantDocumentAll
	ToApiResponseMerchantDocumentDeleteAt(resp *pbmerchantdoc.ApiResponseMerchantDocumentDelete) *response.ApiResponseMerchantDocumentDelete
}

type OrderItemResponseMapper interface {
	ToApiResponseOrderItem(pbResponse *pborderitem.ApiResponseOrderItem) *response.ApiResponseOrderItem
	ToApiResponsesOrderItem(pbResponse *pborderitem.ApiResponsesOrderItem) *response.ApiResponsesOrderItem
	ToApiResponseOrderItemDelete(pbResponse *pborderitem.ApiResponseOrderItemDelete) *response.ApiResponseOrderItemDelete
	ToApiResponseOrderItemAll(pbResponse *pborderitem.ApiResponseOrderItemAll) *response.ApiResponseOrderItemAll
	ToApiResponsePaginationOrderItemDeleteAt(pbResponse *pborderitem.ApiResponsePaginationOrderItemDeleteAt) *response.ApiResponsePaginationOrderItemDeleteAt
	ToApiResponsePaginationOrderItem(pbResponse *pborderitem.ApiResponsePaginationOrderItem) *response.ApiResponsePaginationOrderItem
}

type OrderResponseMapper interface {
	ToApiResponseOrder(pbResponse *pborder.ApiResponseOrder) *response.ApiResponseOrder
	ToApiResponseOrderDeleteAt(pbResponse *pborder.ApiResponseOrderDeleteAt) *response.ApiResponseOrderDeleteAt
	ToApiResponsesOrder(pbResponse *pborder.ApiResponsesOrder) *response.ApiResponsesOrder
	ToApiResponseOrderDelete(pbResponse *pborder.ApiResponseOrderDelete) *response.ApiResponseOrderDelete
	ToApiResponseOrderAll(pbResponse *pborder.ApiResponseOrderAll) *response.ApiResponseOrderAll
	ToApiResponsePaginationOrderDeleteAt(pbResponse *pborder.ApiResponsePaginationOrderDeleteAt) *response.ApiResponsePaginationOrderDeleteAt
	ToApiResponsePaginationOrder(pbResponse *pborder.ApiResponsePaginationOrder) *response.ApiResponsePaginationOrder
}

type ProductResponseMapper interface {
	ToApiResponseProduct(pbResponse *pbproduct.ApiResponseProduct) *response.ApiResponseProduct
	ToApiResponsesProductDeleteAt(pbResponse *pbproduct.ApiResponseProductDeleteAt) *response.ApiResponseProductDeleteAt
	ToApiResponsesProduct(pbResponse *pbproduct.ApiResponsesProduct) *response.ApiResponsesProduct
	ToApiResponseProductDelete(pbResponse *pbproduct.ApiResponseProductDelete) *response.ApiResponseProductDelete
	ToApiResponseProductAll(pbResponse *pbproduct.ApiResponseProductAll) *response.ApiResponseProductAll
	ToApiResponsePaginationProductDeleteAt(pbResponse *pbproduct.ApiResponsePaginationProductDeleteAt) *response.ApiResponsePaginationProductDeleteAt
	ToApiResponsePaginationProduct(pbResponse *pbproduct.ApiResponsePaginationProduct) *response.ApiResponsePaginationProduct
}

type TransactionResponseMapper interface {
	ToApiResponseTransaction(pbResponse *pbtransaction.ApiResponseTransaction) *response.ApiResponseTransaction
	ToApiResponseTransactionDeleteAt(pbResponse *pbtransaction.ApiResponseTransactionDeleteAt) *response.ApiResponseTransactionDeleteAt
	ToApiResponsesTransaction(pbResponse *pbtransaction.ApiResponsesTransaction) *response.ApiResponsesTransaction
	ToApiResponseTransactionDelete(pbResponse *pbtransaction.ApiResponseTransactionDelete) *response.ApiResponseTransactionDelete
	ToApiResponseTransactionAll(pbResponse *pbtransaction.ApiResponseTransactionAll) *response.ApiResponseTransactionAll
	ToApiResponsePaginationTransactionDeleteAt(pbResponse *pbtransaction.ApiResponsePaginationTransactionDeleteAt) *response.ApiResponsePaginationTransactionDeleteAt
	ToApiResponsePaginationTransaction(pbResponse *pbtransaction.ApiResponsePaginationTransaction) *response.ApiResponsePaginationTransaction
}
