package handler

import (
	"github.com/MamangRust/microservice-point-of-sale-merchant/service"
	pbmerchant "github.com/MamangRust/microservice-point-of-sale-pb/merchant"
	pbmerchantdoc "github.com/MamangRust/microservice-point-of-sale-pb/merchant_document"
	"github.com/MamangRust/microservice-point-of-sale-pkg/logger"
)

type Deps struct {
	Service *service.Service
	Logger  logger.LoggerInterface
}

type Handler struct {
	Merchant         pbmerchant.MerchantServiceServer
	MerchantDocument pbmerchantdoc.MerchantDocumentServiceServer
}

func NewHandler(deps *Deps) *Handler {
	return &Handler{
		Merchant: NewMerchantHandleGrpc(
			deps.Service,
			deps.Logger,
		),
		MerchantDocument: NewMerchantDocumentHandleGrpc(
			deps.Service,
			deps.Logger,
		),
	}
}
