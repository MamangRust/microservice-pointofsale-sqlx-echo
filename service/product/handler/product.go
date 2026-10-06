package handler

import (
	"context"
	stderrors "errors"
	"math"

	pbcommon "github.com/MamangRust/microservice-point-of-sale-pb/common"
	pbproduct "github.com/MamangRust/microservice-point-of-sale-pb/product"
	"github.com/MamangRust/microservice-point-of-sale-pkg/logger"
	db "github.com/MamangRust/microservice-point-of-sale-product/database/schema"
	"github.com/MamangRust/microservice-point-of-sale-product/service"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
	"github.com/MamangRust/microservice-point-of-sale-shared/errors"
	"github.com/MamangRust/microservice-point-of-sale-shared/errors/product_errors"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type productHandleGrpc struct {
	pbproduct.UnimplementedProductServiceServer
	productQueryService   service.ProductQueryService
	productCommandService service.ProductCommandService
	logger                logger.LoggerInterface
}

func NewProductHandleGrpc(
	service *service.Service,
	logger logger.LoggerInterface,
) pbproduct.ProductServiceServer {
	return &productHandleGrpc{
		productQueryService:   service.ProductQuery,
		productCommandService: service.ProductCommand,
		logger:                logger,
	}
}

func (s *productHandleGrpc) FindAll(ctx context.Context, request *pbproduct.FindAllProductRequest) (*pbproduct.ApiResponsePaginationProduct, error) {
	s.logger.Info("FindAll products called", zap.Int32("page", request.GetPage()))

	page := int(request.GetPage())
	pageSize := int(request.GetPageSize())
	search := request.GetSearch()

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	reqService := requests.FindAllProducts{
		Page:     page,
		PageSize: pageSize,
		Search:   search,
	}

	product, totalRecords, err := s.productQueryService.FindAll(ctx, &reqService)
	if err != nil {
		s.logger.Error("FindAll products failed", zap.Error(err))
		return nil, errors.ToGrpcError(err)
	}

	totalPages := int(math.Ceil(float64(*totalRecords) / float64(pageSize)))

	paginationMeta := &pbcommon.PaginationMeta{
		CurrentPage:  int32(page),
		PageSize:     int32(pageSize),
		TotalPages:   int32(totalPages),
		TotalRecords: int32(*totalRecords),
	}

	s.logger.Info("FindAll products success")

	return &pbproduct.ApiResponsePaginationProduct{
		Status:     "success",
		Message:    "Successfully fetched product",
		Data:       mapResponsesProduct(product),
		Pagination: mapPaginationMeta(paginationMeta),
	}, nil
}

func (s *productHandleGrpc) FindByMerchant(ctx context.Context, request *pbproduct.FindAllProductMerchantRequest) (*pbproduct.ApiResponsePaginationProduct, error) {
	s.logger.Info("FindByMerchant products called", zap.Int32("merchantId", request.GetMerchantId()))

	page := int(request.GetPage())
	pageSize := int(request.GetPageSize())
	search := request.GetSearch()
	merchant_id := int(request.GetMerchantId())
	min_price := int(request.GetMinPrice())
	max_price := int(request.GetMaxPrice())

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	if min_price <= 0 {
		min_price = 0
	}
	if max_price <= 0 {
		max_price = 0
	}

	reqService := requests.ProductByMerchantRequest{
		MerchantID: merchant_id,
		Page:       page,
		PageSize:   pageSize,
		Search:     search,
		MinPrice:   &min_price,
		MaxPrice:   &max_price,
	}

	product, totalRecords, err := s.productQueryService.FindByMerchant(ctx, &reqService)
	if err != nil {
		s.logger.Error("FindByMerchant products failed", zap.Error(err))
		return nil, errors.ToGrpcError(err)
	}

	totalPages := int(math.Ceil(float64(*totalRecords) / float64(pageSize)))

	paginationMeta := &pbcommon.PaginationMeta{
		CurrentPage:  int32(page),
		PageSize:     int32(pageSize),
		TotalPages:   int32(totalPages),
		TotalRecords: int32(*totalRecords),
	}

	s.logger.Info("FindByMerchant products success")

	return &pbproduct.ApiResponsePaginationProduct{
		Status:     "success",
		Message:    "Successfully fetched product",
		Data:       mapResponsesProductByMerchant(product, int32(merchant_id)),
		Pagination: mapPaginationMeta(paginationMeta),
	}, nil
}

func (s *productHandleGrpc) FindByCategory(ctx context.Context, request *pbproduct.FindAllProductCategoryRequest) (*pbproduct.ApiResponsePaginationProduct, error) {
	s.logger.Info("FindByCategory products called", zap.String("categoryName", request.GetCategoryName()))

	page := int(request.GetPage())
	pageSize := int(request.GetPageSize())
	search := request.GetSearch()
	category_name := request.GetCategoryName()
	min_price := int(request.GetMinprice())
	max_price := int(request.GetMaxprice())

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	if min_price <= 0 {
		min_price = 0
	}
	if max_price <= 0 {
		max_price = 0
	}

	reqService := requests.ProductByCategoryRequest{
		Page:         page,
		PageSize:     pageSize,
		Search:       search,
		CategoryName: category_name,
		MinPrice:     &min_price,
		MaxPrice:     &max_price,
	}

	product, totalRecords, err := s.productQueryService.FindByCategory(ctx, &reqService)
	if err != nil {
		s.logger.Error("FindByCategory products failed", zap.Error(err))
		return nil, errors.ToGrpcError(err)
	}

	totalPages := int(math.Ceil(float64(*totalRecords) / float64(pageSize)))

	paginationMeta := &pbcommon.PaginationMeta{
		CurrentPage:  int32(page),
		PageSize:     int32(pageSize),
		TotalPages:   int32(totalPages),
		TotalRecords: int32(*totalRecords),
	}

	s.logger.Info("FindByCategory products success")

	return &pbproduct.ApiResponsePaginationProduct{
		Status:     "success",
		Message:    "Successfully fetched product",
		Data:       mapResponsesProductByCategory(product),
		Pagination: mapPaginationMeta(paginationMeta),
	}, nil
}

func (s *productHandleGrpc) FindById(ctx context.Context, request *pbproduct.FindByIdProductRequest) (*pbproduct.ApiResponseProduct, error) {
	s.logger.Info("FindById product called", zap.Int32("id", request.GetId()))

	id := int(request.GetId())
	if id <= 0 {
		return nil, product_errors.ErrGrpcInvalidID
	}

	product, err := s.productQueryService.FindById(ctx, id)
	if err != nil {
		s.logger.Error("FindById product failed", zap.Error(err))
		return nil, errors.ToGrpcError(err)
	}

	s.logger.Info("FindById product success")

	return &pbproduct.ApiResponseProduct{
		Status:  "success",
		Message: "Successfully fetched product",
		Data:    mapResponseProduct(product),
	}, nil
}

func (s *productHandleGrpc) FindByActive(ctx context.Context, request *pbproduct.FindAllProductRequest) (*pbproduct.ApiResponsePaginationProductDeleteAt, error) {
	s.logger.Info("FindByActive products called", zap.Int32("page", request.GetPage()))

	page := int(request.GetPage())
	pageSize := int(request.GetPageSize())
	search := request.GetSearch()

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	reqService := requests.FindAllProducts{
		Page:     page,
		PageSize: pageSize,
		Search:   search,
	}

	product, totalRecords, err := s.productQueryService.FindByActive(ctx, &reqService)
	if err != nil {
		s.logger.Error("FindByActive products failed", zap.Error(err))
		return nil, errors.ToGrpcError(err)
	}

	totalPages := int(math.Ceil(float64(*totalRecords) / float64(pageSize)))

	paginationMeta := &pbcommon.PaginationMeta{
		CurrentPage:  int32(page),
		PageSize:     int32(pageSize),
		TotalPages:   int32(totalPages),
		TotalRecords: int32(*totalRecords),
	}

	s.logger.Info("FindByActive products success")

	return &pbproduct.ApiResponsePaginationProductDeleteAt{
		Status:     "success",
		Message:    "Successfully fetched active product",
		Data:       mapResponsesProductActive(product),
		Pagination: mapPaginationMeta(paginationMeta),
	}, nil
}

func (s *productHandleGrpc) FindByTrashed(ctx context.Context, request *pbproduct.FindAllProductRequest) (*pbproduct.ApiResponsePaginationProductDeleteAt, error) {
	s.logger.Info("FindByTrashed products called")

	page := int(request.GetPage())
	pageSize := int(request.GetPageSize())
	search := request.GetSearch()

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	reqService := requests.FindAllProducts{
		Page:     page,
		PageSize: pageSize,
		Search:   search,
	}

	users, totalRecords, err := s.productQueryService.FindByTrashed(ctx, &reqService)
	if err != nil {
		s.logger.Error("FindByTrashed products failed", zap.Error(err))
		return nil, errors.ToGrpcError(err)
	}

	totalPages := int(math.Ceil(float64(*totalRecords) / float64(pageSize)))

	paginationMeta := &pbcommon.PaginationMeta{
		CurrentPage:  int32(page),
		PageSize:     int32(pageSize),
		TotalPages:   int32(totalPages),
		TotalRecords: int32(*totalRecords),
	}

	s.logger.Info("FindByTrashed products success")

	return &pbproduct.ApiResponsePaginationProductDeleteAt{
		Status:     "success",
		Message:    "Successfully fetched trashed product",
		Data:       mapResponsesProductTrashed(users),
		Pagination: mapPaginationMeta(paginationMeta),
	}, nil
}

func (s *productHandleGrpc) Create(ctx context.Context, request *pbproduct.CreateProductRequest) (*pbproduct.ApiResponseProduct, error) {
	s.logger.Info("Create product called", zap.String("name", request.GetName()))

	req := &requests.CreateProductRequest{
		MerchantID:   int(request.GetMerchantId()),
		CategoryID:   int(request.GetCategoryId()),
		Name:         request.GetName(),
		Description:  request.GetDescription(),
		Price:        int(request.GetPrice()),
		CountInStock: int(request.GetCountInStock()),
		Brand:        request.GetBrand(),
		Weight:       int(request.GetWeight()),
		ImageProduct: request.GetImageProduct(),
	}

	if err := req.Validate(); err != nil {
		return nil, product_errors.ErrGrpcValidateCreateProduct
	}

	product, err := s.productCommandService.CreateProduct(ctx, req)
	if err != nil {
		s.logger.Error("Create product failed", zap.Error(err))
		return nil, errors.ToGrpcError(err)
	}

	s.logger.Info("Create product success")

	return &pbproduct.ApiResponseProduct{
		Status:  "success",
		Message: "Successfully created product",
		Data:    mapResponseProduct(product),
	}, nil
}

func (s *productHandleGrpc) Update(ctx context.Context, request *pbproduct.UpdateProductRequest) (*pbproduct.ApiResponseProduct, error) {
	s.logger.Info("Update product called", zap.Int32("id", request.GetProductId()))

	id := int(request.GetProductId())
	if id <= 0 {
		return nil, product_errors.ErrGrpcInvalidID
	}

	req := &requests.UpdateProductRequest{
		ProductID:    &id,
		MerchantID:   int(request.GetMerchantId()),
		CategoryID:   int(request.GetCategoryId()),
		Name:         request.GetName(),
		Description:  request.GetDescription(),
		Price:        int(request.GetPrice()),
		CountInStock: int(request.GetCountInStock()),
		Brand:        request.GetBrand(),
		Weight:       int(request.GetWeight()),
		ImageProduct: request.GetImageProduct(),
	}

	if err := req.Validate(); err != nil {
		return nil, product_errors.ErrGrpcValidateUpdateProduct
	}

	product, err := s.productCommandService.UpdateProduct(ctx, req)
	if err != nil {
		s.logger.Error("Update product failed", zap.Error(err))
		return nil, errors.ToGrpcError(err)
	}

	s.logger.Info("Update product success")

	return &pbproduct.ApiResponseProduct{
		Status:  "success",
		Message: "Successfully updated product",
		Data:    mapResponseProduct(product),
	}, nil
}

func (s *productHandleGrpc) TrashedProduct(ctx context.Context, request *pbproduct.FindByIdProductRequest) (*pbproduct.ApiResponseProductDeleteAt, error) {
	s.logger.Info("TrashedProduct called", zap.Int32("id", request.GetId()))

	id := int(request.GetId())
	if id <= 0 {
		return nil, product_errors.ErrGrpcInvalidID
	}

	product, err := s.productCommandService.TrashProduct(ctx, id)
	if err != nil {
		s.logger.Error("TrashedProduct failed", zap.Error(err))
		return nil, errors.ToGrpcError(err)
	}

	s.logger.Info("TrashedProduct success")

	return &pbproduct.ApiResponseProductDeleteAt{
		Status:  "success",
		Message: "Successfully trashed product",
		Data:    mapResponseProductDeleteAt(product),
	}, nil
}

func (s *productHandleGrpc) RestoreProduct(ctx context.Context, request *pbproduct.FindByIdProductRequest) (*pbproduct.ApiResponseProductDeleteAt, error) {
	s.logger.Info("RestoreProduct called", zap.Int32("id", request.GetId()))

	id := int(request.GetId())
	if id <= 0 {
		return nil, product_errors.ErrGrpcInvalidID
	}

	product, err := s.productCommandService.RestoreProduct(ctx, id)
	if err != nil {
		s.logger.Error("RestoreProduct failed", zap.Error(err))
		return nil, errors.ToGrpcError(err)
	}

	s.logger.Info("RestoreProduct success")

	return &pbproduct.ApiResponseProductDeleteAt{
		Status:  "success",
		Message: "Successfully restored product",
		Data:    mapResponseProductDeleteAt(product),
	}, nil
}

func (s *productHandleGrpc) DeleteProductPermanent(ctx context.Context, request *pbproduct.FindByIdProductRequest) (*pbproduct.ApiResponseProductDelete, error) {
	s.logger.Info("DeleteProductPermanent called", zap.Int32("id", request.GetId()))

	id := int(request.GetId())
	if id <= 0 {
		return nil, product_errors.ErrGrpcInvalidID
	}

	_, err := s.productCommandService.DeleteProductPermanent(ctx, id)
	if err != nil {
		s.logger.Error("DeleteProductPermanent failed", zap.Error(err))
		return nil, errors.ToGrpcError(err)
	}

	s.logger.Info("DeleteProductPermanent success")

	return &pbproduct.ApiResponseProductDelete{
		Status:  "success",
		Message: "Successfully deleted Product permanently",
	}, nil
}

func (s *productHandleGrpc) RestoreAllProduct(ctx context.Context, _ *emptypb.Empty) (*pbproduct.ApiResponseProductAll, error) {
	s.logger.Info("RestoreAllProduct called")

	_, err := s.productCommandService.RestoreAllProducts(ctx)
	if err != nil {
		s.logger.Error("RestoreAllProduct failed", zap.Error(err))
		return nil, errors.ToGrpcError(err)
	}

	s.logger.Info("RestoreAllProduct success")

	return &pbproduct.ApiResponseProductAll{
		Status:  "success",
		Message: "Successfully restore all Product",
	}, nil
}

func (s *productHandleGrpc) DeleteAllProductPermanent(ctx context.Context, _ *emptypb.Empty) (*pbproduct.ApiResponseProductAll, error) {
	s.logger.Info("DeleteAllProductPermanent called")

	_, err := s.productCommandService.DeleteAllProductsPermanent(ctx)
	if err != nil {
		s.logger.Error("DeleteAllProductPermanent failed", zap.Error(err))
		return nil, errors.ToGrpcError(err)
	}

	s.logger.Info("DeleteAllProductPermanent success")

	return &pbproduct.ApiResponseProductAll{
		Status:  "success",
		Message: "Successfully delete Product permanen",
	}, nil
}

func (s *productHandleGrpc) DecrementStock(ctx context.Context, request *pbproduct.AdjustProductStockRequest) (*pbproduct.ApiResponseProduct, error) {
	return s.adjustStock(ctx, request, true)
}

func (s *productHandleGrpc) IncrementStock(ctx context.Context, request *pbproduct.AdjustProductStockRequest) (*pbproduct.ApiResponseProduct, error) {
	return s.adjustStock(ctx, request, false)
}

// adjustStock delegates the stock mutation to the owning service. Insufficient
// stock is surfaced as codes.FailedPrecondition so callers (order-service) can
// distinguish a business rejection from a transport failure.
func (s *productHandleGrpc) adjustStock(ctx context.Context, request *pbproduct.AdjustProductStockRequest, decrement bool) (*pbproduct.ApiResponseProduct, error) {
	productID := int(request.GetProductId())
	quantity := int(request.GetQuantity())

	s.logger.Info("Adjust product stock called",
		zap.Int("product.id", productID),
		zap.Int("quantity", quantity),
		zap.Bool("decrement", decrement),
	)

	if productID <= 0 {
		return nil, product_errors.ErrGrpcInvalidID
	}
	if quantity <= 0 {
		return nil, product_errors.ErrGrpcInvalidQuantity
	}

	var (
		product *db.Product
		err     error
	)
	if decrement {
		product, err = s.productCommandService.DecrementProductCountStock(ctx, productID, quantity)
	} else {
		product, err = s.productCommandService.IncrementProductCountStock(ctx, productID, quantity)
	}
	if err != nil {
		s.logger.Error("Adjust product stock failed",
			zap.Int("product.id", productID),
			zap.Bool("decrement", decrement),
			zap.Error(err),
		)
		if stderrors.Is(err, product_errors.ErrInsufficientProductStock) {
			return nil, product_errors.ErrGrpcInsufficientStock
		}
		return nil, errors.ToGrpcError(err)
	}

	s.logger.Info("Adjust product stock success", zap.Int("product.id", productID))

	return &pbproduct.ApiResponseProduct{
		Status:  "success",
		Message: "Successfully adjusted product stock",
		Data:    mapResponseProduct(product),
	}, nil
}

// Map helpers
func parseStrPointer(s *string) string {
	if s != nil {
		return *s
	}
	return ""
}

func parseInt32Pointer(i *int32) int32 {
	if i != nil {
		return *i
	}
	return 0
}

func mapPaginationMeta(meta *pbcommon.PaginationMeta) *pbcommon.PaginationMeta {
	if meta == nil {
		return nil
	}
	return &pbcommon.PaginationMeta{
		CurrentPage:  meta.CurrentPage,
		PageSize:     meta.PageSize,
		TotalPages:   meta.TotalPages,
		TotalRecords: meta.TotalRecords,
	}
}

func mapResponseProduct(product *db.Product) *pbproduct.ProductResponse {
	if product == nil {
		return nil
	}
	var createdAtStr, updatedAtStr string
	if product.CreatedAt.Valid {
		createdAtStr = product.CreatedAt.Time.Format("2006-01-02 15:04:05")
	}
	if product.UpdatedAt.Valid {
		updatedAtStr = product.UpdatedAt.Time.Format("2006-01-02 15:04:05")
	}
	return &pbproduct.ProductResponse{
		Id:           int32(product.ProductID),
		MerchantId:   int32(product.MerchantID),
		CategoryId:   int32(product.CategoryID),
		Name:         product.Name,
		Description:  parseStrPointer(product.Description),
		Price:        int32(product.Price),
		CountInStock: int32(product.CountInStock),
		Brand:        parseStrPointer(product.Brand),
		Weight:       parseInt32Pointer(product.Weight),
		SlugProduct:  parseStrPointer(product.SlugProduct),
		ImageProduct: parseStrPointer(product.ImageProduct),
		Barcode:      parseStrPointer(product.Barcode),
		CreatedAt:    createdAtStr,
		UpdatedAt:    updatedAtStr,
	}
}

func mapResponsesProduct(products []*db.GetProductsRow) []*pbproduct.ProductResponse {
	var mappedProducts []*pbproduct.ProductResponse
	for _, p := range products {
		if p == nil {
			continue
		}
		var createdAtStr, updatedAtStr string
		if p.CreatedAt.Valid {
			createdAtStr = p.CreatedAt.Time.Format("2006-01-02 15:04:05")
		}
		if p.UpdatedAt.Valid {
			updatedAtStr = p.UpdatedAt.Time.Format("2006-01-02 15:04:05")
		}
		mappedProducts = append(mappedProducts, &pbproduct.ProductResponse{
			Id:           int32(p.ProductID),
			MerchantId:   int32(p.MerchantID),
			CategoryId:   int32(p.CategoryID),
			Name:         p.Name,
			Description:  parseStrPointer(p.Description),
			Price:        int32(p.Price),
			CountInStock: int32(p.CountInStock),
			Brand:        parseStrPointer(p.Brand),
			Weight:       parseInt32Pointer(p.Weight),
			SlugProduct:  parseStrPointer(p.SlugProduct),
			ImageProduct: parseStrPointer(p.ImageProduct),
			Barcode:      parseStrPointer(p.Barcode),
			CreatedAt:    createdAtStr,
			UpdatedAt:    updatedAtStr,
		})
	}
	return mappedProducts
}

func mapResponsesProductByMerchant(products []*db.GetProductsByMerchantRow, merchantId int32) []*pbproduct.ProductResponse {
	var mappedProducts []*pbproduct.ProductResponse
	for _, p := range products {
		if p == nil {
			continue
		}
		var createdAtStr string
		if p.CreatedAt.Valid {
			createdAtStr = p.CreatedAt.Time.Format("2006-01-02 15:04:05")
		}
		mappedProducts = append(mappedProducts, &pbproduct.ProductResponse{
			Id:           int32(p.ProductID),
			MerchantId:   merchantId,
			Name:         p.Name,
			Description:  parseStrPointer(p.Description),
			Price:        int32(p.Price),
			CountInStock: int32(p.CountInStock),
			Brand:        parseStrPointer(p.Brand),
			ImageProduct: parseStrPointer(p.ImageProduct),
			CreatedAt:    createdAtStr,
		})
	}
	return mappedProducts
}

func mapResponsesProductByCategory(products []*db.GetProductsByCategoryNameRow) []*pbproduct.ProductResponse {
	var mappedProducts []*pbproduct.ProductResponse
	for _, p := range products {
		if p == nil {
			continue
		}
		var createdAtStr, updatedAtStr string
		if p.CreatedAt.Valid {
			createdAtStr = p.CreatedAt.Time.Format("2006-01-02 15:04:05")
		}
		if p.UpdatedAt.Valid {
			updatedAtStr = p.UpdatedAt.Time.Format("2006-01-02 15:04:05")
		}
		mappedProducts = append(mappedProducts, &pbproduct.ProductResponse{
			Id:           int32(p.ProductID),
			MerchantId:   int32(p.MerchantID),
			CategoryId:   int32(p.CategoryID),
			Name:         p.Name,
			Description:  parseStrPointer(p.Description),
			Price:        int32(p.Price),
			CountInStock: int32(p.CountInStock),
			Brand:        parseStrPointer(p.Brand),
			Weight:       parseInt32Pointer(p.Weight),
			SlugProduct:  parseStrPointer(p.SlugProduct),
			ImageProduct: parseStrPointer(p.ImageProduct),
			Barcode:      parseStrPointer(p.Barcode),
			CreatedAt:    createdAtStr,
			UpdatedAt:    updatedAtStr,
		})
	}
	return mappedProducts
}

func mapResponseProductDeleteAt(product *db.Product) *pbproduct.ProductResponseDeleteAt {
	if product == nil {
		return nil
	}
	var createdAtStr, updatedAtStr string
	if product.CreatedAt.Valid {
		createdAtStr = product.CreatedAt.Time.Format("2006-01-02 15:04:05")
	}
	if product.UpdatedAt.Valid {
		updatedAtStr = product.UpdatedAt.Time.Format("2006-01-02 15:04:05")
	}
	var deletedAt *wrapperspb.StringValue
	if product.DeletedAt.Valid {
		deletedAt = wrapperspb.String(product.DeletedAt.Time.Format("2006-01-02 15:04:05"))
	}

	return &pbproduct.ProductResponseDeleteAt{
		Id:           int32(product.ProductID),
		MerchantId:   int32(product.MerchantID),
		CategoryId:   int32(product.CategoryID),
		Name:         product.Name,
		Description:  parseStrPointer(product.Description),
		Price:        int32(product.Price),
		CountInStock: int32(product.CountInStock),
		Brand:        parseStrPointer(product.Brand),
		Weight:       parseInt32Pointer(product.Weight),
		SlugProduct:  parseStrPointer(product.SlugProduct),
		ImageProduct: parseStrPointer(product.ImageProduct),
		Barcode:      parseStrPointer(product.Barcode),
		CreatedAt:    createdAtStr,
		UpdatedAt:    updatedAtStr,
		DeletedAt:    deletedAt,
	}
}

func mapResponsesProductActive(products []*db.GetProductsActiveRow) []*pbproduct.ProductResponseDeleteAt {
	var mappedProducts []*pbproduct.ProductResponseDeleteAt
	for _, p := range products {
		if p == nil {
			continue
		}
		var createdAtStr, updatedAtStr string
		if p.CreatedAt.Valid {
			createdAtStr = p.CreatedAt.Time.Format("2006-01-02 15:04:05")
		}
		if p.UpdatedAt.Valid {
			updatedAtStr = p.UpdatedAt.Time.Format("2006-01-02 15:04:05")
		}
		var deletedAt *wrapperspb.StringValue
		if p.DeletedAt.Valid {
			deletedAt = wrapperspb.String(p.DeletedAt.Time.Format("2006-01-02 15:04:05"))
		}

		mappedProducts = append(mappedProducts, &pbproduct.ProductResponseDeleteAt{
			Id:           int32(p.ProductID),
			MerchantId:   int32(p.MerchantID),
			CategoryId:   int32(p.CategoryID),
			Name:         p.Name,
			Description:  parseStrPointer(p.Description),
			Price:        int32(p.Price),
			CountInStock: int32(p.CountInStock),
			Brand:        parseStrPointer(p.Brand),
			Weight:       parseInt32Pointer(p.Weight),
			SlugProduct:  parseStrPointer(p.SlugProduct),
			ImageProduct: parseStrPointer(p.ImageProduct),
			Barcode:      parseStrPointer(p.Barcode),
			CreatedAt:    createdAtStr,
			UpdatedAt:    updatedAtStr,
			DeletedAt:    deletedAt,
		})
	}
	return mappedProducts
}

func mapResponsesProductTrashed(products []*db.GetProductsTrashedRow) []*pbproduct.ProductResponseDeleteAt {
	var mappedProducts []*pbproduct.ProductResponseDeleteAt
	for _, p := range products {
		if p == nil {
			continue
		}
		var createdAtStr, updatedAtStr string
		if p.CreatedAt.Valid {
			createdAtStr = p.CreatedAt.Time.Format("2006-01-02 15:04:05")
		}
		if p.UpdatedAt.Valid {
			updatedAtStr = p.UpdatedAt.Time.Format("2006-01-02 15:04:05")
		}
		var deletedAt *wrapperspb.StringValue
		if p.DeletedAt.Valid {
			deletedAt = wrapperspb.String(p.DeletedAt.Time.Format("2006-01-02 15:04:05"))
		}

		mappedProducts = append(mappedProducts, &pbproduct.ProductResponseDeleteAt{
			Id:           int32(p.ProductID),
			MerchantId:   int32(p.MerchantID),
			CategoryId:   int32(p.CategoryID),
			Name:         p.Name,
			Description:  parseStrPointer(p.Description),
			Price:        int32(p.Price),
			CountInStock: int32(p.CountInStock),
			Brand:        parseStrPointer(p.Brand),
			Weight:       parseInt32Pointer(p.Weight),
			SlugProduct:  parseStrPointer(p.SlugProduct),
			ImageProduct: parseStrPointer(p.ImageProduct),
			Barcode:      parseStrPointer(p.Barcode),
			CreatedAt:    createdAtStr,
			UpdatedAt:    updatedAtStr,
			DeletedAt:    deletedAt,
		})
	}
	return mappedProducts
}
