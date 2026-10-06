package category_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	apicache "github.com/MamangRust/microservice-point-of-sale-apigateway/cache/gateway_cache"
	categoryhandler "github.com/MamangRust/microservice-point-of-sale-apigateway/handler"
	pbcategory "github.com/MamangRust/microservice-point-of-sale-pb/category"
	"github.com/MamangRust/microservice-point-of-sale-shared/errors"
	response_api "github.com/MamangRust/microservice-point-of-sale-shared/mapper/response/api"
	tests "github.com/MamangRust/microservice-point-of-sale-test"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/suite"
)

type CategoryApiTestSuite struct {
	tests.BaseTestSuite
	echo       *echo.Echo
	categoryID int
}

func (s *CategoryApiTestSuite) SetupSuite() {
	s.BaseTestSuite.SetupSuite()
	s.SetupCategoryService()

	s.echo = echo.New()
	apiHandler := errors.NewApiHandler(s.Obs, s.Log)

	catMapper := response_api.NewCategoryResponseMapper()
	gatewayCache := apicache.NewGatewayCache(s.GetCacheStore())
	client := pbcategory.NewCategoryServiceClient(s.Conns["category"])
	categoryhandler.NewHandlerCategory(
		s.echo,
		client,
		s.Log,
		catMapper,
		apiHandler,
		gatewayCache,
	)
}

func (s *CategoryApiTestSuite) TestCategoryApiLifecycle() {
	// 1. Create
	fields := map[string]interface{}{
		"name":        "Test Category",
		"description": "Test Description",
	}
	body, _ := json.Marshal(fields)
	req := httptest.NewRequest(http.MethodPost, "/api/category/create", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, "application/json")
	rec := httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	s.Require().Equal(http.StatusCreated, rec.Code, rec.Body.String())
	var res map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &res)
	data := res["data"].(map[string]interface{})
	s.Equal(fields["name"], data["name"])
	s.categoryID = int(data["id"].(float64))

	// 2. FindById
	req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/category/%d", s.categoryID), nil)
	rec = httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	s.Equal(http.StatusOK, rec.Code)
	json.Unmarshal(rec.Body.Bytes(), &res)
	data = res["data"].(map[string]interface{})
	s.Equal(float64(s.categoryID), data["id"])

	// 3. FindAll
	req = httptest.NewRequest(http.MethodGet, "/api/category", nil)
	rec = httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	s.Equal(http.StatusOK, rec.Code)

	// 4. FindByActive
	req = httptest.NewRequest(http.MethodGet, "/api/category/active", nil)
	rec = httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	s.Equal(http.StatusOK, rec.Code)

	// 5. Update
	updFields := map[string]interface{}{
		"name":        "Updated Category",
		"description": "Updated Description",
	}
	updBody, _ := json.Marshal(updFields)
	req = httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/category/update/%d", s.categoryID), bytes.NewReader(updBody))
	req.Header.Set(echo.HeaderContentType, "application/json")
	rec = httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	s.Equal(http.StatusOK, rec.Code, rec.Body.String())
	json.Unmarshal(rec.Body.Bytes(), &res)
	data = res["data"].(map[string]interface{})
	s.Equal(updFields["name"], data["name"])

	// 6. Trash
	req = httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/category/trashed/%d", s.categoryID), nil)
	rec = httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	s.Equal(http.StatusOK, rec.Code)

	// 7. FindByTrashed
	req = httptest.NewRequest(http.MethodGet, "/api/category/trashed", nil)
	rec = httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	s.Equal(http.StatusOK, rec.Code)

	// 8. Restore
	req = httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/category/restore/%d", s.categoryID), nil)
	rec = httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	s.Equal(http.StatusOK, rec.Code)

	// 9. DeletePermanent
	req = httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/category/trashed/%d", s.categoryID), nil)
	rec = httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	req = httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/category/permanent/%d", s.categoryID), nil)
	rec = httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	s.Equal(http.StatusOK, rec.Code)

	// 10. RestoreAll
	req = httptest.NewRequest(http.MethodPost, "/api/category/restore/all", nil)
	rec = httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	s.Equal(http.StatusOK, rec.Code)

	// 11. DeleteAll
	req = httptest.NewRequest(http.MethodPost, "/api/category/permanent/all", nil)
	rec = httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	s.Equal(http.StatusOK, rec.Code)
}

func TestCategoryApiSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(CategoryApiTestSuite))
}
