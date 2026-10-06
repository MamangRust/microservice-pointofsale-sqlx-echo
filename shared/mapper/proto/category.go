package protomapper

import (
	pbcategory "github.com/MamangRust/microservice-point-of-sale-pb/category"
	pbcommon "github.com/MamangRust/microservice-point-of-sale-pb/common"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/response"

	"google.golang.org/protobuf/types/known/wrapperspb"
)

type categoryProtoMapper struct {
}

func NewCategoryProtoMapper() *categoryProtoMapper {
	return &categoryProtoMapper{}
}

func (c *categoryProtoMapper) ToProtoResponseCategory(status string, message string, pbResponse *response.CategoryResponse) *pbcategory.ApiResponseCategory {
	return &pbcategory.ApiResponseCategory{
		Status:  status,
		Message: message,
		Data:    c.mapResponseCategory(pbResponse),
	}
}

func (c *categoryProtoMapper) ToProtoResponseCategoryDeleteAt(status string, message string, pbResponse *response.CategoryResponseDeleteAt) *pbcategory.ApiResponseCategoryDeleteAt {
	return &pbcategory.ApiResponseCategoryDeleteAt{
		Status:  status,
		Message: message,
		Data:    c.mapResponseCategoryDeleteAt(pbResponse),
	}
}

func (c *categoryProtoMapper) ToProtoResponsesCategory(status string, message string, pbResponse []*response.CategoryResponse) *pbcategory.ApiResponsesCategory {
	return &pbcategory.ApiResponsesCategory{
		Status:  status,
		Message: message,
		Data:    c.mapResponsesCategory(pbResponse),
	}
}

func (c *categoryProtoMapper) ToProtoResponseCategoryDelete(status string, message string) *pbcategory.ApiResponseCategoryDelete {
	return &pbcategory.ApiResponseCategoryDelete{
		Status:  status,
		Message: message,
	}
}

func (c *categoryProtoMapper) ToProtoResponseCategoryAll(status string, message string) *pbcategory.ApiResponseCategoryAll {
	return &pbcategory.ApiResponseCategoryAll{
		Status:  status,
		Message: message,
	}
}

func (c *categoryProtoMapper) ToProtoResponsePaginationCategoryDeleteAt(pagination *pbcommon.PaginationMeta, status string, message string, categories []*response.CategoryResponseDeleteAt) *pbcategory.ApiResponsePaginationCategoryDeleteAt {
	return &pbcategory.ApiResponsePaginationCategoryDeleteAt{
		Status:     status,
		Message:    message,
		Data:       c.mapResponsesCategoryDeleteAt(categories),
		Pagination: mapPaginationMeta(pagination),
	}
}

func (c *categoryProtoMapper) ToProtoResponsePaginationCategory(pagination *pbcommon.PaginationMeta, status string, message string, categories []*response.CategoryResponse) *pbcategory.ApiResponsePaginationCategory {
	return &pbcategory.ApiResponsePaginationCategory{
		Status:     status,
		Message:    message,
		Data:       c.mapResponsesCategory(categories),
		Pagination: mapPaginationMeta(pagination),
	}
}

func (c *categoryProtoMapper) mapResponseCategory(category *response.CategoryResponse) *pbcategory.CategoryResponse {
	return &pbcategory.CategoryResponse{
		Id:            int32(category.ID),
		Name:          category.Name,
		Description:   category.Description,
		SlugCategory:  category.SlugCategory,
		ImageCategory: category.ImageCategory,
		CreatedAt:     category.CreatedAt,
		UpdatedAt:     category.UpdatedAt,
	}
}

func (c *categoryProtoMapper) mapResponsesCategory(categories []*response.CategoryResponse) []*pbcategory.CategoryResponse {
	var mappedCategories []*pbcategory.CategoryResponse

	for _, category := range categories {
		mappedCategories = append(mappedCategories, c.mapResponseCategory(category))
	}

	return mappedCategories
}

func (c *categoryProtoMapper) mapResponseCategoryDeleteAt(category *response.CategoryResponseDeleteAt) *pbcategory.CategoryResponseDeleteAt {
	var deletedAt *wrapperspb.StringValue
	if category.DeletedAt != nil {
		deletedAt = wrapperspb.String(*category.DeletedAt)
	}

	return &pbcategory.CategoryResponseDeleteAt{
		Id:            int32(category.ID),
		Name:          category.Name,
		Description:   category.Description,
		SlugCategory:  category.SlugCategory,
		ImageCategory: category.ImageCategory,
		CreatedAt:     category.CreatedAt,
		UpdatedAt:     category.UpdatedAt,
		DeletedAt:     deletedAt,
	}
}

func (c *categoryProtoMapper) mapResponsesCategoryDeleteAt(categories []*response.CategoryResponseDeleteAt) []*pbcategory.CategoryResponseDeleteAt {
	var mappedCategories []*pbcategory.CategoryResponseDeleteAt

	for _, category := range categories {
		mappedCategories = append(mappedCategories, c.mapResponseCategoryDeleteAt(category))
	}

	return mappedCategories
}
