package handler

import (
	"context"
	"math"

	db "github.com/MamangRust/microservice-point-of-sale-category/database/schema"
	"github.com/MamangRust/microservice-point-of-sale-category/service"
	pbcategory "github.com/MamangRust/microservice-point-of-sale-pb/category"
	pbcommon "github.com/MamangRust/microservice-point-of-sale-pb/common"
	"github.com/MamangRust/microservice-point-of-sale-pkg/logger"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
	"github.com/MamangRust/microservice-point-of-sale-shared/errors"
	"github.com/MamangRust/microservice-point-of-sale-shared/errors/category_errors"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type categoryHandleGrpc struct {
	pbcategory.UnimplementedCategoryServiceServer
	categoryQuery   service.CategoryQueryService
	categoryCommand service.CategoryCommandService
	logger          logger.LoggerInterface
}

func NewCategoryHandleGrpc(
	service *service.Service,
	logger logger.LoggerInterface,
) pbcategory.CategoryServiceServer {
	return &categoryHandleGrpc{
		categoryQuery:   service.CategoryQuery,
		categoryCommand: service.CategoryCommand,
		logger:          logger,
	}
}

func (s *categoryHandleGrpc) FindAll(ctx context.Context, request *pbcategory.FindAllCategoryRequest) (*pbcategory.ApiResponsePaginationCategory, error) {
	s.logger.Info("FindAll categories called", zap.Int32("page", request.GetPage()))

	page := int(request.GetPage())
	pageSize := int(request.GetPageSize())
	search := request.GetSearch()

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	reqService := requests.FindAllCategory{
		Page:     page,
		PageSize: pageSize,
		Search:   search,
	}

	category, totalRecords, err := s.categoryQuery.FindAll(ctx, &reqService)
	if err != nil {
		s.logger.Error("FindAll categories failed", zap.Error(err))
		return nil, errors.ToGrpcError(err)
	}

	totalPages := int(math.Ceil(float64(*totalRecords) / float64(pageSize)))

	paginationMeta := &pbcommon.PaginationMeta{
		CurrentPage:  int32(page),
		PageSize:     int32(pageSize),
		TotalPages:   int32(totalPages),
		TotalRecords: int32(*totalRecords),
	}

	s.logger.Info("FindAll categories success", zap.Int("count", len(category)))

	return &pbcategory.ApiResponsePaginationCategory{
		Status:     "success",
		Message:    "Successfully fetched categories",
		Data:       mapResponsesCategory(category),
		Pagination: mapPaginationMeta(paginationMeta),
	}, nil
}

func (s *categoryHandleGrpc) FindById(ctx context.Context, request *pbcategory.FindByIdCategoryRequest) (*pbcategory.ApiResponseCategory, error) {
	s.logger.Info("FindById category called", zap.Int32("id", request.GetId()))

	id := int(request.GetId())
	if id <= 0 {
		return nil, category_errors.ErrGrpcFailedInvalidId
	}

	category, err := s.categoryQuery.FindById(ctx, id)
	if err != nil {
		s.logger.Error("FindById category failed", zap.Error(err))
		return nil, errors.ToGrpcError(err)
	}

	s.logger.Info("FindById category success", zap.Int("id", id))

	return &pbcategory.ApiResponseCategory{
		Status:  "success",
		Message: "Successfully fetched category",
		Data:    mapResponseCategory(category),
	}, nil
}

func (s *categoryHandleGrpc) FindByName(ctx context.Context, request *pbcategory.FindByNameCategoryRequest) (*pbcategory.ApiResponseCategory, error) {
	s.logger.Info("FindByName category called", zap.String("name", request.GetName()))

	name := request.GetName()
	if name == "" {
		return nil, category_errors.ErrGrpcFailedInvalidName
	}

	category, err := s.categoryQuery.FindByName(ctx, name)
	if err != nil {
		s.logger.Error("FindByName category failed", zap.Error(err))
		return nil, errors.ToGrpcError(err)
	}

	s.logger.Info("FindByName category success", zap.String("name", name))

	return &pbcategory.ApiResponseCategory{
		Status:  "success",
		Message: "Successfully fetched category",
		Data:    mapResponseCategory(category),
	}, nil
}

func (s *categoryHandleGrpc) FindByIds(ctx context.Context, request *pbcategory.FindByIdsCategoryRequest) (*pbcategory.ApiResponsesCategory, error) {
	s.logger.Info("FindByIds categories called", zap.Int("count", len(request.GetIds())))

	ids := make([]int, 0, len(request.GetIds()))
	for _, id := range request.GetIds() {
		if id > 0 {
			ids = append(ids, int(id))
		}
	}

	categories, err := s.categoryQuery.FindByIds(ctx, ids)
	if err != nil {
		s.logger.Error("FindByIds categories failed", zap.Error(err))
		return nil, errors.ToGrpcError(err)
	}

	s.logger.Info("FindByIds categories success", zap.Int("count", len(categories)))

	return &pbcategory.ApiResponsesCategory{
		Status:  "success",
		Message: "Successfully fetched categories",
		Data:    mapResponsesCategoryModels(categories),
	}, nil
}

func (s *categoryHandleGrpc) FindByActive(ctx context.Context, request *pbcategory.FindAllCategoryRequest) (*pbcategory.ApiResponsePaginationCategoryDeleteAt, error) {	s.logger.Info("FindByActive categories called", zap.Int32("page", request.GetPage()))

	page := int(request.GetPage())
	pageSize := int(request.GetPageSize())
	search := request.GetSearch()

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	reqService := requests.FindAllCategory{
		Page:     page,
		PageSize: pageSize,
		Search:   search,
	}

	categories, totalRecords, err := s.categoryQuery.FindByActive(ctx, &reqService)
	if err != nil {
		s.logger.Error("FindByActive categories failed", zap.Error(err))
		return nil, errors.ToGrpcError(err)
	}

	totalPages := int(math.Ceil(float64(*totalRecords) / float64(pageSize)))

	paginationMeta := &pbcommon.PaginationMeta{
		CurrentPage:  int32(page),
		PageSize:     int32(pageSize),
		TotalPages:   int32(totalPages),
		TotalRecords: int32(*totalRecords),
	}

	s.logger.Info("FindByActive categories success")

	return &pbcategory.ApiResponsePaginationCategoryDeleteAt{
		Status:     "success",
		Message:    "Successfully fetched active categories",
		Data:       mapResponsesCategoryActive(categories),
		Pagination: mapPaginationMeta(paginationMeta),
	}, nil
}

func (s *categoryHandleGrpc) FindByTrashed(ctx context.Context, request *pbcategory.FindAllCategoryRequest) (*pbcategory.ApiResponsePaginationCategoryDeleteAt, error) {
	s.logger.Info("FindByTrashed categories called")

	page := int(request.GetPage())
	pageSize := int(request.GetPageSize())
	search := request.GetSearch()

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	reqService := requests.FindAllCategory{
		Page:     page,
		PageSize: pageSize,
		Search:   search,
	}

	categories, totalRecords, err := s.categoryQuery.FindByTrashed(ctx, &reqService)
	if err != nil {
		s.logger.Error("FindByTrashed categories failed", zap.Error(err))
		return nil, errors.ToGrpcError(err)
	}

	totalPages := int(math.Ceil(float64(*totalRecords) / float64(pageSize)))

	paginationMeta := &pbcommon.PaginationMeta{
		CurrentPage:  int32(page),
		PageSize:     int32(pageSize),
		TotalPages:   int32(totalPages),
		TotalRecords: int32(*totalRecords),
	}

	s.logger.Info("FindByTrashed categories success")

	return &pbcategory.ApiResponsePaginationCategoryDeleteAt{
		Status:     "success",
		Message:    "Successfully fetched trashed categories",
		Data:       mapResponsesCategoryTrashed(categories),
		Pagination: mapPaginationMeta(paginationMeta),
	}, nil
}

func (s *categoryHandleGrpc) Create(ctx context.Context, request *pbcategory.CreateCategoryRequest) (*pbcategory.ApiResponseCategory, error) {
	s.logger.Info("Create category called", zap.String("name", request.GetName()))

	req := &requests.CreateCategoryRequest{
		Name:        request.GetName(),
		Description: request.GetDescription(),
	}

	if err := req.Validate(); err != nil {
		return nil, category_errors.ErrGrpcValidateCreateCategory
	}

	category, err := s.categoryCommand.CreateCategory(ctx, req)
	if err != nil {
		s.logger.Error("Create category failed", zap.Error(err))
		return nil, errors.ToGrpcError(err)
	}

	s.logger.Info("Create category success")

	return &pbcategory.ApiResponseCategory{
		Status:  "success",
		Message: "Successfully created category",
		Data:    mapResponseCategory(category),
	}, nil
}

func (s *categoryHandleGrpc) Update(ctx context.Context, request *pbcategory.UpdateCategoryRequest) (*pbcategory.ApiResponseCategory, error) {
	s.logger.Info("Update category called", zap.Int32("id", request.GetCategoryId()))

	id := int(request.GetCategoryId())
	if id <= 0 {
		return nil, category_errors.ErrGrpcFailedInvalidId
	}

	req := &requests.UpdateCategoryRequest{
		CategoryID:  &id,
		Name:        request.GetName(),
		Description: request.GetDescription(),
	}

	if err := req.Validate(); err != nil {
		return nil, category_errors.ErrGrpcValidateUpdateCategory
	}

	category, err := s.categoryCommand.UpdateCategory(ctx, req)
	if err != nil {
		s.logger.Error("Update category failed", zap.Error(err))
		return nil, errors.ToGrpcError(err)
	}

	s.logger.Info("Update category success")

	return &pbcategory.ApiResponseCategory{
		Status:  "success",
		Message: "Successfully updated category",
		Data:    mapResponseCategory(category),
	}, nil
}

func (s *categoryHandleGrpc) TrashedCategory(ctx context.Context, request *pbcategory.FindByIdCategoryRequest) (*pbcategory.ApiResponseCategoryDeleteAt, error) {
	s.logger.Info("TrashedCategory called", zap.Int32("id", request.GetId()))

	id := int(request.GetId())
	if id <= 0 {
		return nil, category_errors.ErrGrpcFailedInvalidId
	}

	category, err := s.categoryCommand.TrashedCategory(ctx, id)
	if err != nil {
		s.logger.Error("TrashedCategory failed", zap.Error(err))
		return nil, errors.ToGrpcError(err)
	}

	s.logger.Info("TrashedCategory success")

	return &pbcategory.ApiResponseCategoryDeleteAt{
		Status:  "success",
		Message: "Successfully trashed category",
		Data:    mapResponseCategoryDeleteAt(category),
	}, nil
}

func (s *categoryHandleGrpc) RestoreCategory(ctx context.Context, request *pbcategory.FindByIdCategoryRequest) (*pbcategory.ApiResponseCategoryDeleteAt, error) {
	s.logger.Info("RestoreCategory called", zap.Int32("id", request.GetId()))

	id := int(request.GetId())
	if id <= 0 {
		return nil, category_errors.ErrGrpcFailedInvalidId
	}

	category, err := s.categoryCommand.RestoreCategory(ctx, id)
	if err != nil {
		s.logger.Error("RestoreCategory failed", zap.Error(err))
		return nil, errors.ToGrpcError(err)
	}

	s.logger.Info("RestoreCategory success")

	return &pbcategory.ApiResponseCategoryDeleteAt{
		Status:  "success",
		Message: "Successfully restored category",
		Data:    mapResponseCategoryDeleteAt(category),
	}, nil
}

func (s *categoryHandleGrpc) DeleteCategoryPermanent(ctx context.Context, request *pbcategory.FindByIdCategoryRequest) (*pbcategory.ApiResponseCategoryDelete, error) {
	s.logger.Info("DeleteCategoryPermanent called", zap.Int32("id", request.GetId()))

	id := int(request.GetId())
	if id <= 0 {
		return nil, category_errors.ErrGrpcFailedInvalidId
	}

	_, err := s.categoryCommand.DeleteCategoryPermanent(ctx, id)
	if err != nil {
		s.logger.Error("DeleteCategoryPermanent failed", zap.Error(err))
		return nil, errors.ToGrpcError(err)
	}

	s.logger.Info("DeleteCategoryPermanent success")

	return &pbcategory.ApiResponseCategoryDelete{
		Status:  "success",
		Message: "Successfully deleted category permanently",
	}, nil
}

func (s *categoryHandleGrpc) RestoreAllCategory(ctx context.Context, _ *emptypb.Empty) (*pbcategory.ApiResponseCategoryAll, error) {
	s.logger.Info("RestoreAllCategory called")

	_, err := s.categoryCommand.RestoreAllCategories(ctx)
	if err != nil {
		s.logger.Error("RestoreAllCategory failed", zap.Error(err))
		return nil, errors.ToGrpcError(err)
	}

	s.logger.Info("RestoreAllCategory success")

	return &pbcategory.ApiResponseCategoryAll{
		Status:  "success",
		Message: "Successfully restore all category",
	}, nil
}

func (s *categoryHandleGrpc) DeleteAllCategoryPermanent(ctx context.Context, _ *emptypb.Empty) (*pbcategory.ApiResponseCategoryAll, error) {
	s.logger.Info("DeleteAllCategoryPermanent called")

	_, err := s.categoryCommand.DeleteAllCategoriesPermanent(ctx)
	if err != nil {
		s.logger.Error("DeleteAllCategoryPermanent failed", zap.Error(err))
		return nil, errors.ToGrpcError(err)
	}

	s.logger.Info("DeleteAllCategoryPermanent success")

	return &pbcategory.ApiResponseCategoryAll{
		Status:  "success",
		Message: "Successfully delete category permanen",
	}, nil
}

// Internal map helpers
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

func mapResponseCategory(category *db.Category) *pbcategory.CategoryResponse {
	if category == nil {
		return nil
	}
	var description, slugCategory string
	if category.Description != nil {
		description = *category.Description
	}
	if category.SlugCategory != nil {
		slugCategory = *category.SlugCategory
	}
	var createdAtStr, updatedAtStr string
	if category.CreatedAt.Valid {
		createdAtStr = category.CreatedAt.Time.Format("2006-01-02 15:04:05")
	}
	if category.UpdatedAt.Valid {
		updatedAtStr = category.UpdatedAt.Time.Format("2006-01-02 15:04:05")
	}
	return &pbcategory.CategoryResponse{
		Id:           int32(category.CategoryID),
		Name:         category.Name,
		Description:  description,
		SlugCategory: slugCategory,
		CreatedAt:    createdAtStr,
		UpdatedAt:    updatedAtStr,
	}
}

func mapResponsesCategoryModels(categories []*db.Category) []*pbcategory.CategoryResponse {
	var mappedCategories []*pbcategory.CategoryResponse
	for _, category := range categories {
		mappedCategories = append(mappedCategories, mapResponseCategory(category))
	}
	return mappedCategories
}

func mapResponseGetCategory(category *db.GetCategoriesRow) *pbcategory.CategoryResponse {
	if category == nil {
		return nil
	}
	var description, slugCategory string
	if category.Description != nil {
		description = *category.Description
	}
	if category.SlugCategory != nil {
		slugCategory = *category.SlugCategory
	}
	var createdAtStr, updatedAtStr string
	if category.CreatedAt.Valid {
		createdAtStr = category.CreatedAt.Time.Format("2006-01-02 15:04:05")
	}
	if category.UpdatedAt.Valid {
		updatedAtStr = category.UpdatedAt.Time.Format("2006-01-02 15:04:05")
	}
	return &pbcategory.CategoryResponse{
		Id:           int32(category.CategoryID),
		Name:         category.Name,
		Description:  description,
		SlugCategory: slugCategory,
		CreatedAt:    createdAtStr,
		UpdatedAt:    updatedAtStr,
	}
}

func mapResponsesCategory(categories []*db.GetCategoriesRow) []*pbcategory.CategoryResponse {
	var mappedCategories []*pbcategory.CategoryResponse
	for _, category := range categories {
		mappedCategories = append(mappedCategories, mapResponseGetCategory(category))
	}
	return mappedCategories
}

func mapResponseCategoryDeleteAt(category *db.Category) *pbcategory.CategoryResponseDeleteAt {
	if category == nil {
		return nil
	}
	var description, slugCategory string
	if category.Description != nil {
		description = *category.Description
	}
	if category.SlugCategory != nil {
		slugCategory = *category.SlugCategory
	}
	var createdAtStr, updatedAtStr string
	if category.CreatedAt.Valid {
		createdAtStr = category.CreatedAt.Time.Format("2006-01-02 15:04:05")
	}
	if category.UpdatedAt.Valid {
		updatedAtStr = category.UpdatedAt.Time.Format("2006-01-02 15:04:05")
	}
	var deletedAt *wrapperspb.StringValue
	if category.DeletedAt.Valid {
		deletedAt = wrapperspb.String(category.DeletedAt.Time.Format("2006-01-02 15:04:05"))
	}

	return &pbcategory.CategoryResponseDeleteAt{
		Id:           int32(category.CategoryID),
		Name:         category.Name,
		Description:  description,
		SlugCategory: slugCategory,
		CreatedAt:    createdAtStr,
		UpdatedAt:    updatedAtStr,
		DeletedAt:    deletedAt,
	}
}

func mapResponseGetCategoryActive(category *db.GetCategoriesActiveRow) *pbcategory.CategoryResponseDeleteAt {
	if category == nil {
		return nil
	}
	var description, slugCategory string
	if category.Description != nil {
		description = *category.Description
	}
	if category.SlugCategory != nil {
		slugCategory = *category.SlugCategory
	}
	var createdAtStr, updatedAtStr string
	if category.CreatedAt.Valid {
		createdAtStr = category.CreatedAt.Time.Format("2006-01-02 15:04:05")
	}
	if category.UpdatedAt.Valid {
		updatedAtStr = category.UpdatedAt.Time.Format("2006-01-02 15:04:05")
	}
	var deletedAt *wrapperspb.StringValue
	if category.DeletedAt.Valid {
		deletedAt = wrapperspb.String(category.DeletedAt.Time.Format("2006-01-02 15:04:05"))
	}

	return &pbcategory.CategoryResponseDeleteAt{
		Id:           int32(category.CategoryID),
		Name:         category.Name,
		Description:  description,
		SlugCategory: slugCategory,
		CreatedAt:    createdAtStr,
		UpdatedAt:    updatedAtStr,
		DeletedAt:    deletedAt,
	}
}

func mapResponsesCategoryActive(categories []*db.GetCategoriesActiveRow) []*pbcategory.CategoryResponseDeleteAt {
	var mappedCategories []*pbcategory.CategoryResponseDeleteAt
	for _, category := range categories {
		mappedCategories = append(mappedCategories, mapResponseGetCategoryActive(category))
	}
	return mappedCategories
}

func mapResponseGetCategoryTrashed(category *db.GetCategoriesTrashedRow) *pbcategory.CategoryResponseDeleteAt {
	if category == nil {
		return nil
	}
	var description, slugCategory string
	if category.Description != nil {
		description = *category.Description
	}
	if category.SlugCategory != nil {
		slugCategory = *category.SlugCategory
	}
	var createdAtStr, updatedAtStr string
	if category.CreatedAt.Valid {
		createdAtStr = category.CreatedAt.Time.Format("2006-01-02 15:04:05")
	}
	if category.UpdatedAt.Valid {
		updatedAtStr = category.UpdatedAt.Time.Format("2006-01-02 15:04:05")
	}
	var deletedAt *wrapperspb.StringValue
	if category.DeletedAt.Valid {
		deletedAt = wrapperspb.String(category.DeletedAt.Time.Format("2006-01-02 15:04:05"))
	}

	return &pbcategory.CategoryResponseDeleteAt{
		Id:           int32(category.CategoryID),
		Name:         category.Name,
		Description:  description,
		SlugCategory: slugCategory,
		CreatedAt:    createdAtStr,
		UpdatedAt:    updatedAtStr,
		DeletedAt:    deletedAt,
	}
}

func mapResponsesCategoryTrashed(categories []*db.GetCategoriesTrashedRow) []*pbcategory.CategoryResponseDeleteAt {
	var mappedCategories []*pbcategory.CategoryResponseDeleteAt
	for _, category := range categories {
		mappedCategories = append(mappedCategories, mapResponseGetCategoryTrashed(category))
	}
	return mappedCategories
}
