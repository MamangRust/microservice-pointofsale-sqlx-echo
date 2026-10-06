package category_test

import (
	"context"
	"testing"

	cat_cache "github.com/MamangRust/microservice-point-of-sale-category/cache"
	cat_handler "github.com/MamangRust/microservice-point-of-sale-category/handler"
	cat_repo "github.com/MamangRust/microservice-point-of-sale-category/repository"
	cat_service "github.com/MamangRust/microservice-point-of-sale-category/service"
	pbcategory "github.com/MamangRust/microservice-point-of-sale-pb/category"
	"github.com/MamangRust/microservice-point-of-sale-shared/cache"
	"github.com/MamangRust/microservice-point-of-sale-shared/observability"
	tests "github.com/MamangRust/microservice-point-of-sale-test"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type CategoryGapiTestSuite struct {
	tests.BaseTestSuite
	client pbcategory.CategoryServiceClient
}

func (s *CategoryGapiTestSuite) SetupSuite() {
	s.BaseTestSuite.SetupSuite()

	// Infrastructure
	cacheMetrics, _ := observability.NewCacheMetrics("test")
	cacheStore := cache.NewCacheStore(s.RedisClient(), s.Log, cacheMetrics)
	// Category dependencies
	mencache := cat_cache.NewMencache(cacheStore)
	repos := cat_repo.NewRepositories(s.SQLxDB())
	svc := cat_service.NewService(&cat_service.Deps{
		Ctx:           context.Background(),
		Mencache:      mencache,
		Repositories:  repos,
		Logger:        s.Log,
		Observability: s.Obs,
	})

	// Handler
	handler := cat_handler.NewHandler(&cat_handler.Deps{
		Service: svc,
		Logger:  s.Log,
	})

	// GRPC Server
	server := grpc.NewServer()
	pbcategory.RegisterCategoryServiceServer(server, handler.Category)

	addr := s.RegisterServer(server)
	conn := s.GetConnection(addr)

	s.client = pbcategory.NewCategoryServiceClient(conn)
}

func (s *CategoryGapiTestSuite) TestCategoryGapiLifecycle() {
	ctx := context.Background()

	// 1. Create
	createRes, err := s.client.Create(ctx, &pbcategory.CreateCategoryRequest{
		Name:        "GAPI Category",
		Description: "Testing via GRPC",
	})
	s.NoError(err)
	s.NotNil(createRes)
	catID := createRes.Data.Id

	// 2. FindById
	getRes, err := s.client.FindById(ctx, &pbcategory.FindByIdCategoryRequest{Id: catID})
	s.NoError(err)
	s.Equal("GAPI Category", getRes.Data.Name)

	// 3. FindAll
	allRes, err := s.client.FindAll(ctx, &pbcategory.FindAllCategoryRequest{Page: 1, PageSize: 10})
	s.NoError(err)
	s.NotEmpty(allRes.Data)

	// 4. FindByActive
	activeRes, err := s.client.FindByActive(ctx, &pbcategory.FindAllCategoryRequest{Page: 1, PageSize: 10})
	s.NoError(err)
	s.NotEmpty(activeRes.Data)

	// 5. Update
	updateRes, err := s.client.Update(ctx, &pbcategory.UpdateCategoryRequest{
		CategoryId:  catID,
		Name:        "GAPI Category Updated",
		Description: "Updated via GRPC",
	})
	s.NoError(err)
	s.Equal("GAPI Category Updated", updateRes.Data.Name)

	// 6. Trash
	_, err = s.client.TrashedCategory(ctx, &pbcategory.FindByIdCategoryRequest{Id: catID})
	s.NoError(err)

	// 7. FindByTrashed
	trashedRes, err := s.client.FindByTrashed(ctx, &pbcategory.FindAllCategoryRequest{Page: 1, PageSize: 10})
	s.NoError(err)
	s.NotEmpty(trashedRes.Data)

	// 8. Restore
	_, err = s.client.RestoreCategory(ctx, &pbcategory.FindByIdCategoryRequest{Id: catID})
	s.NoError(err)

	// 9. DeletePermanent
	_, _ = s.client.TrashedCategory(ctx, &pbcategory.FindByIdCategoryRequest{Id: catID})
	_, err = s.client.DeleteCategoryPermanent(ctx, &pbcategory.FindByIdCategoryRequest{Id: catID})
	s.NoError(err)

	// 10. RestoreAll
	_, err = s.client.RestoreAllCategory(ctx, &emptypb.Empty{})
	s.NoError(err)

	// 11. DeleteAll
	_, err = s.client.DeleteAllCategoryPermanent(ctx, &emptypb.Empty{})
	s.NoError(err)
}

func TestCategoryGapiSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(CategoryGapiTestSuite))
}
