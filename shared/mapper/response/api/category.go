package response_api

import (
	pbcategory "github.com/MamangRust/microservice-point-of-sale-pb/category"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/response"
)

type categoryResponseMapper struct{}

func NewCategoryResponseMapper() *categoryResponseMapper {
	return &categoryResponseMapper{}
}

func (c *categoryResponseMapper) ToResponseCategory(category *pbcategory.CategoryResponse) *response.CategoryResponse {
	return &response.CategoryResponse{
		ID:            int(category.Id),
		Name:          category.Name,
		Description:   category.Description,
		SlugCategory:  category.SlugCategory,
		ImageCategory: category.ImageCategory,
		CreatedAt:     category.CreatedAt,
		UpdatedAt:     category.UpdatedAt,
	}
}

func (c *categoryResponseMapper) ToResponsesCategory(categories []*pbcategory.CategoryResponse) []*response.CategoryResponse {
	var mappedCategories []*response.CategoryResponse

	for _, category := range categories {
		mappedCategories = append(mappedCategories, c.ToResponseCategory(category))
	}

	return mappedCategories
}

func (c *categoryResponseMapper) ToResponseCategoryDelete(category *pbcategory.CategoryResponseDeleteAt) *response.CategoryResponseDeleteAt {
	var deletedAt string
	if category.DeletedAt != nil {
		deletedAt = category.DeletedAt.Value
	}

	return &response.CategoryResponseDeleteAt{
		ID:            int(category.Id),
		Name:          category.Name,
		Description:   category.Description,
		SlugCategory:  category.SlugCategory,
		ImageCategory: category.ImageCategory,
		CreatedAt:     category.CreatedAt,
		UpdatedAt:     category.UpdatedAt,
		DeletedAt:     &deletedAt,
	}
}

func (c *categoryResponseMapper) ToResponsesCategoryDeleteAt(categories []*pbcategory.CategoryResponseDeleteAt) []*response.CategoryResponseDeleteAt {
	var mappedCategories []*response.CategoryResponseDeleteAt

	for _, category := range categories {
		mappedCategories = append(mappedCategories, c.ToResponseCategoryDelete(category))
	}

	return mappedCategories
}

func (c *categoryResponseMapper) ToApiResponseCategory(pbResponse *pbcategory.ApiResponseCategory) *response.ApiResponseCategory {
	return &response.ApiResponseCategory{
		Status:  pbResponse.Status,
		Message: pbResponse.Message,
		Data:    c.ToResponseCategory(pbResponse.Data),
	}
}

func (c *categoryResponseMapper) ToApiResponseCategoryDeleteAt(pbResponse *pbcategory.ApiResponseCategoryDeleteAt) *response.ApiResponseCategoryDeleteAt {
	return &response.ApiResponseCategoryDeleteAt{
		Status:  pbResponse.Status,
		Message: pbResponse.Message,
		Data:    c.ToResponseCategoryDelete(pbResponse.Data),
	}
}

func (c *categoryResponseMapper) ToApiResponsesCategory(pbResponse *pbcategory.ApiResponsesCategory) *response.ApiResponsesCategory {
	return &response.ApiResponsesCategory{
		Status:  pbResponse.Status,
		Message: pbResponse.Message,
		Data:    c.ToResponsesCategory(pbResponse.Data),
	}
}

func (c *categoryResponseMapper) ToApiResponseCategoryDelete(pbResponse *pbcategory.ApiResponseCategoryDelete) *response.ApiResponseCategoryDelete {
	return &response.ApiResponseCategoryDelete{
		Status:  pbResponse.Status,
		Message: pbResponse.Message,
	}
}

func (c *categoryResponseMapper) ToApiResponseCategoryAll(pbResponse *pbcategory.ApiResponseCategoryAll) *response.ApiResponseCategoryAll {
	return &response.ApiResponseCategoryAll{
		Status:  pbResponse.Status,
		Message: pbResponse.Message,
	}
}

func (c *categoryResponseMapper) ToApiResponsePaginationCategoryDeleteAt(pbResponse *pbcategory.ApiResponsePaginationCategoryDeleteAt) *response.ApiResponsePaginationCategoryDeleteAt {
	return &response.ApiResponsePaginationCategoryDeleteAt{
		Status:     pbResponse.Status,
		Message:    pbResponse.Message,
		Data:       c.ToResponsesCategoryDeleteAt(pbResponse.Data),
		Pagination: *mapPaginationMeta(pbResponse.Pagination),
	}
}

func (c *categoryResponseMapper) ToApiResponsePaginationCategory(pbResponse *pbcategory.ApiResponsePaginationCategory) *response.ApiResponsePaginationCategory {
	return &response.ApiResponsePaginationCategory{
		Status:     pbResponse.Status,
		Message:    pbResponse.Message,
		Data:       c.ToResponsesCategory(pbResponse.Data),
		Pagination: *mapPaginationMeta(pbResponse.Pagination),
	}
}
