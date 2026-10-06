package handler

import (
	"context"
	"math"

	pbcommon "github.com/MamangRust/microservice-point-of-sale-pb/common"
	pbrole "github.com/MamangRust/microservice-point-of-sale-pb/role"
	"github.com/MamangRust/microservice-point-of-sale-pkg/logger"
	db "github.com/MamangRust/microservice-point-of-sale-role/database/schema"
	"github.com/MamangRust/microservice-point-of-sale-role/service"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
	"github.com/MamangRust/microservice-point-of-sale-shared/errors"
	"github.com/MamangRust/microservice-point-of-sale-shared/errors/role_errors"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/emptypb"
)

type roleHandleGrpc struct {
	pbrole.UnimplementedRoleServiceServer
	roleQuery   service.RoleQueryService
	roleCommand service.RoleCommandService
	logger      logger.LoggerInterface
}

func NewRoleHandleGrpc(
	service *service.Service,
	logger logger.LoggerInterface,
) pbrole.RoleServiceServer {
	return &roleHandleGrpc{
		roleQuery:   service.RoleQuery,
		roleCommand: service.RoleCommand,
		logger:      logger,
	}
}

func (s *roleHandleGrpc) FindAllRole(ctx context.Context, req *pbrole.FindAllRoleRequest) (*pbrole.ApiResponsePaginationRole, error) {
	s.logger.Info("FindAllRole called", zap.Int32("page", req.GetPage()))

	page := int(req.GetPage())
	pageSize := int(req.GetPageSize())
	search := req.GetSearch()

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	reqService := requests.FindAllRoles{
		Page:     page,
		PageSize: pageSize,
		Search:   search,
	}

	roles, totalRecords, err := s.roleQuery.FindAll(ctx, &reqService)
	if err != nil {
		s.logger.Error("FindAllRoles failed", zap.Error(err))
		return nil, errors.ToGrpcError(err)
	}

	totalPages := int(math.Ceil(float64(*totalRecords) / float64(pageSize)))

	paginationMeta := &pbcommon.PaginationMeta{
		CurrentPage:  int32(page),
		PageSize:     int32(pageSize),
		TotalPages:   int32(totalPages),
		TotalRecords: int32(*totalRecords),
	}

	s.logger.Info("FindAllRoles success")

	return &pbrole.ApiResponsePaginationRole{
		Status:     "success",
		Message:    "Successfully fetched role records",
		Data:       mapResponsesRole(roles),
		Pagination: mapPaginationMeta(paginationMeta),
	}, nil
}

func (s *roleHandleGrpc) FindByIdRole(ctx context.Context, req *pbrole.FindByIdRoleRequest) (*pbrole.ApiResponseRole, error) {
	s.logger.Info("FindByIdRole called", zap.Int32("roleId", req.GetRoleId()))

	roleID := int(req.GetRoleId())
	if roleID <= 0 {
		return nil, role_errors.ErrGrpcRoleInvalidId
	}

	role, err := s.roleQuery.FindById(ctx, roleID)
	if err != nil {
		s.logger.Error("FindByIdRole failed", zap.Error(err))
		return nil, errors.ToGrpcError(err)
	}

	s.logger.Info("FindByIdRole success")

	return &pbrole.ApiResponseRole{
		Status:  "success",
		Message: "Successfully fetched role",
		Data:    mapResponseRole(role),
	}, nil
}

func (s *roleHandleGrpc) FindByNameRole(ctx context.Context, req *pbrole.FindByNameRoleRequest) (*pbrole.ApiResponseRole, error) {
	s.logger.Info("FindByNameRole called", zap.String("name", req.GetName()))

	name := req.GetName()
	if name == "" {
		return nil, role_errors.ErrGrpcRoleInvalidName
	}

	role, err := s.roleQuery.FindByName(ctx, name)
	if err != nil {
		s.logger.Error("FindByNameRole failed", zap.Error(err))
		return nil, errors.ToGrpcError(err)
	}

	s.logger.Info("FindByNameRole success")

	return &pbrole.ApiResponseRole{
		Status:  "success",
		Message: "Successfully fetched role",
		Data:    mapResponseRole(role),
	}, nil
}

func (s *roleHandleGrpc) FindByActive(ctx context.Context, req *pbrole.FindAllRoleRequest) (*pbrole.ApiResponsePaginationRoleDeleteAt, error) {
	s.logger.Info("FindByActive roles called", zap.Int32("page", req.GetPage()))

	page := int(req.GetPage())
	pageSize := int(req.GetPageSize())
	search := req.GetSearch()

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	reqService := requests.FindAllRoles{
		Page:     page,
		PageSize: pageSize,
		Search:   search,
	}

	roles, totalRecords, err := s.roleQuery.FindByActiveRole(ctx, &reqService)
	if err != nil {
		s.logger.Error("FindByActive roles failed", zap.Error(err))
		return nil, errors.ToGrpcError(err)
	}

	totalPages := int(math.Ceil(float64(*totalRecords) / float64(pageSize)))

	paginationMeta := &pbcommon.PaginationMeta{
		CurrentPage:  int32(page),
		PageSize:     int32(pageSize),
		TotalPages:   int32(totalPages),
		TotalRecords: int32(*totalRecords),
	}

	s.logger.Info("FindByActive roles success")

	return &pbrole.ApiResponsePaginationRoleDeleteAt{
		Status:     "success",
		Message:    "Successfully fetched active roles",
		Data:       mapResponsesRoleFromActive(roles),
		Pagination: mapPaginationMeta(paginationMeta),
	}, nil
}

func (s *roleHandleGrpc) FindByTrashed(ctx context.Context, req *pbrole.FindAllRoleRequest) (*pbrole.ApiResponsePaginationRoleDeleteAt, error) {
	s.logger.Info("FindByTrashed roles called")

	page := int(req.GetPage())
	pageSize := int(req.GetPageSize())
	search := req.GetSearch()

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	reqService := requests.FindAllRoles{
		Page:     page,
		PageSize: pageSize,
		Search:   search,
	}

	roles, totalRecords, err := s.roleQuery.FindByTrashedRole(ctx, &reqService)
	if err != nil {
		s.logger.Error("FindByTrashed roles failed", zap.Error(err))
		return nil, errors.ToGrpcError(err)
	}

	totalPages := int(math.Ceil(float64(*totalRecords) / float64(pageSize)))

	paginationMeta := &pbcommon.PaginationMeta{
		CurrentPage:  int32(page),
		PageSize:     int32(pageSize),
		TotalPages:   int32(totalPages),
		TotalRecords: int32(*totalRecords),
	}

	s.logger.Info("FindByTrashed roles success")

	return &pbrole.ApiResponsePaginationRoleDeleteAt{
		Status:     "success",
		Message:    "Successfully fetched trashed roles",
		Data:       mapResponsesRoleFromTrashed(roles),
		Pagination: mapPaginationMeta(paginationMeta),
	}, nil
}

func (s *roleHandleGrpc) CreateRole(ctx context.Context, reqPb *pbrole.CreateRoleRequest) (*pbrole.ApiResponseRole, error) {
	s.logger.Info("CreateRole called", zap.String("name", reqPb.Name))

	req := &requests.CreateRoleRequest{
		Name: reqPb.Name,
	}

	if err := req.Validate(); err != nil {
		return nil, role_errors.ErrGrpcValidateCreateRole
	}

	role, err := s.roleCommand.CreateRole(ctx, req)
	if err != nil {
		s.logger.Error("CreateRole failed", zap.Error(err))
		return nil, errors.ToGrpcError(err)
	}

	s.logger.Info("CreateRole success")

	return &pbrole.ApiResponseRole{
		Status:  "success",
		Message: "Successfully created role",
		Data:    mapResponseRole(role),
	}, nil
}

func (s *roleHandleGrpc) UpdateRole(ctx context.Context, reqPb *pbrole.UpdateRoleRequest) (*pbrole.ApiResponseRole, error) {
	s.logger.Info("UpdateRole called", zap.Int32("id", reqPb.GetId()))

	roleID := int(reqPb.GetId())
	if roleID <= 0 {
		return nil, role_errors.ErrGrpcRoleInvalidId
	}

	name := reqPb.GetName()
	req := &requests.UpdateRoleRequest{
		ID:   &roleID,
		Name: name,
	}

	if err := req.Validate(); err != nil {
		return nil, role_errors.ErrGrpcValidateUpdateRole
	}

	role, err := s.roleCommand.UpdateRole(ctx, req)
	if err != nil {
		s.logger.Error("UpdateRole failed", zap.Error(err))
		return nil, errors.ToGrpcError(err)
	}

	s.logger.Info("UpdateRole success")

	return &pbrole.ApiResponseRole{
		Status:  "success",
		Message: "Successfully updated role",
		Data:    mapResponseRole(role),
	}, nil
}

func (s *roleHandleGrpc) TrashedRole(ctx context.Context, req *pbrole.FindByIdRoleRequest) (*pbrole.ApiResponseRole, error) {
	s.logger.Info("TrashedRole called", zap.Int32("id", req.GetRoleId()))

	roleID := int(req.GetRoleId())
	if roleID <= 0 {
		return nil, role_errors.ErrGrpcRoleInvalidId
	}

	role, err := s.roleCommand.TrashedRole(ctx, roleID)
	if err != nil {
		s.logger.Error("TrashedRole failed", zap.Error(err))
		return nil, errors.ToGrpcError(err)
	}

	s.logger.Info("TrashedRole success")

	return &pbrole.ApiResponseRole{
		Status:  "success",
		Message: "Successfully trashed role",
		Data:    mapResponseRole(role),
	}, nil
}

func (s *roleHandleGrpc) RestoreRole(ctx context.Context, req *pbrole.FindByIdRoleRequest) (*pbrole.ApiResponseRole, error) {
	s.logger.Info("RestoreRole called", zap.Int32("id", req.GetRoleId()))

	roleID := int(req.GetRoleId())
	if roleID <= 0 {
		return nil, role_errors.ErrGrpcRoleInvalidId
	}

	role, err := s.roleCommand.RestoreRole(ctx, roleID)
	if err != nil {
		s.logger.Error("RestoreRole failed", zap.Error(err))
		return nil, errors.ToGrpcError(err)
	}

	s.logger.Info("RestoreRole success")

	return &pbrole.ApiResponseRole{
		Status:  "success",
		Message: "Successfully restored role",
		Data:    mapResponseRole(role),
	}, nil
}

func (s *roleHandleGrpc) DeleteRolePermanent(ctx context.Context, req *pbrole.FindByIdRoleRequest) (*pbrole.ApiResponseRoleDelete, error) {
	s.logger.Info("DeleteRolePermanent called", zap.Int32("id", req.GetRoleId()))

	id := int(req.GetRoleId())
	if id <= 0 {
		return nil, role_errors.ErrGrpcRoleInvalidId
	}

	_, err := s.roleCommand.DeleteRolePermanent(ctx, id)
	if err != nil {
		s.logger.Error("DeleteRolePermanent failed", zap.Error(err))
		return nil, errors.ToGrpcError(err)
	}

	s.logger.Info("DeleteRolePermanent success")

	return &pbrole.ApiResponseRoleDelete{
		Status:  "success",
		Message: "Successfully deleted role permanently",
	}, nil
}

func (s *roleHandleGrpc) RestoreAllRole(ctx context.Context, req *emptypb.Empty) (*pbrole.ApiResponseRoleAll, error) {
	s.logger.Info("RestoreAllRole called")

	_, err := s.roleCommand.RestoreAllRole(ctx)
	if err != nil {
		s.logger.Error("RestoreAllRole failed", zap.Error(err))
		return nil, errors.ToGrpcError(err)
	}

	s.logger.Info("RestoreAllRole success")

	return &pbrole.ApiResponseRoleAll{
		Status:  "success",
		Message: "Successfully restored all roles",
	}, nil
}

func (s *roleHandleGrpc) DeleteAllRolePermanent(ctx context.Context, req *emptypb.Empty) (*pbrole.ApiResponseRoleAll, error) {
	s.logger.Info("DeleteAllRolePermanent called")

	_, err := s.roleCommand.DeleteAllRolePermanent(ctx)
	if err != nil {
		s.logger.Error("DeleteAllRolePermanent failed", zap.Error(err))
		return nil, errors.ToGrpcError(err)
	}

	s.logger.Info("DeleteAllRolePermanent success")

	return &pbrole.ApiResponseRoleAll{
		Status:  "success",
		Message: "Successfully deleted all roles",
	}, nil
}

// Map helpers
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

func mapResponseRole(role *db.Role) *pbrole.RoleResponse {
	if role == nil {
		return nil
	}
	var createdAtStr string
	if role.CreatedAt.Valid {
		createdAtStr = role.CreatedAt.Time.Format("2006-01-02 15:04:05")
	}
	var updatedAtStr string
	if role.UpdatedAt.Valid {
		updatedAtStr = role.UpdatedAt.Time.Format("2006-01-02 15:04:05")
	}
	return &pbrole.RoleResponse{
		Id:        role.RoleID,
		Name:      role.RoleName,
		CreatedAt: createdAtStr,
		UpdatedAt: updatedAtStr,
	}
}

func mapResponsesRole(roles []*db.GetRolesRow) []*pbrole.RoleResponse {
	var responseRoles []*pbrole.RoleResponse
	for _, role := range roles {
		if role == nil {
			continue
		}
		var createdAtStr string
		if role.CreatedAt.Valid {
			createdAtStr = role.CreatedAt.Time.Format("2006-01-02 15:04:05")
		}
		var updatedAtStr string
		if role.UpdatedAt.Valid {
			updatedAtStr = role.UpdatedAt.Time.Format("2006-01-02 15:04:05")
		}
		responseRoles = append(responseRoles, &pbrole.RoleResponse{
			Id:        role.RoleID,
			Name:      role.RoleName,
			CreatedAt: createdAtStr,
			UpdatedAt: updatedAtStr,
		})
	}
	return responseRoles
}

func mapResponsesRoleFromDB(roles []*db.Role) []*pbrole.RoleResponse {
	var responseRoles []*pbrole.RoleResponse
	for _, role := range roles {
		responseRoles = append(responseRoles, mapResponseRole(role))
	}
	return responseRoles
}

func mapResponsesRoleFromActive(roles []*db.GetActiveRolesRow) []*pbrole.RoleResponseDeleteAt {
	var responseRoles []*pbrole.RoleResponseDeleteAt
	for _, role := range roles {
		if role == nil {
			continue
		}
		var createdAtStr string
		if role.CreatedAt.Valid {
			createdAtStr = role.CreatedAt.Time.Format("2006-01-02 15:04:05")
		}
		var updatedAtStr string
		if role.UpdatedAt.Valid {
			updatedAtStr = role.UpdatedAt.Time.Format("2006-01-02 15:04:05")
		}
		var deletedAtStr string
		if role.DeletedAt.Valid {
			deletedAtStr = role.DeletedAt.Time.Format("2006-01-02 15:04:05")
		}
		responseRoles = append(responseRoles, &pbrole.RoleResponseDeleteAt{
			Id:        role.RoleID,
			Name:      role.RoleName,
			CreatedAt: createdAtStr,
			UpdatedAt: updatedAtStr,
			DeletedAt: deletedAtStr,
		})
	}
	return responseRoles
}

func mapResponsesRoleFromTrashed(roles []*db.GetTrashedRolesRow) []*pbrole.RoleResponseDeleteAt {
	var responseRoles []*pbrole.RoleResponseDeleteAt
	for _, role := range roles {
		if role == nil {
			continue
		}
		var createdAtStr string
		if role.CreatedAt.Valid {
			createdAtStr = role.CreatedAt.Time.Format("2006-01-02 15:04:05")
		}
		var updatedAtStr string
		if role.UpdatedAt.Valid {
			updatedAtStr = role.UpdatedAt.Time.Format("2006-01-02 15:04:05")
		}
		var deletedAtStr string
		if role.DeletedAt.Valid {
			deletedAtStr = role.DeletedAt.Time.Format("2006-01-02 15:04:05")
		}
		responseRoles = append(responseRoles, &pbrole.RoleResponseDeleteAt{
			Id:        role.RoleID,
			Name:      role.RoleName,
			CreatedAt: createdAtStr,
			UpdatedAt: updatedAtStr,
			DeletedAt: deletedAtStr,
		})
	}
	return responseRoles
}
