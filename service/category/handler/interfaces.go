package handler

import (
	pbcategory "github.com/MamangRust/microservice-point-of-sale-pb/category"
)

type CategoryHandleGrpc interface {
	pbcategory.CategoryServiceServer
}
