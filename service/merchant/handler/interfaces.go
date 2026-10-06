package handler

import (
	pbmerchant "github.com/MamangRust/microservice-point-of-sale-pb/merchant"
	pbmerchantdoc "github.com/MamangRust/microservice-point-of-sale-pb/merchant_document"
)

type MerchantDocumentHandleGrpc interface {
	pbmerchantdoc.MerchantDocumentServiceServer
}

type MerchantHandleGrpc interface {
	pbmerchant.MerchantServiceServer
}
