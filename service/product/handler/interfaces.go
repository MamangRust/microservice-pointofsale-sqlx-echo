package handler

import (
	pbproduct "github.com/MamangRust/microservice-point-of-sale-pb/product"
)

type ProductHandleGrpc interface {
	pbproduct.ProductServiceServer
}
