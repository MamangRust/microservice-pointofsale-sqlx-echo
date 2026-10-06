package handler

import (
	"context"

	pbrole "github.com/MamangRust/microservice-point-of-sale-pb/role"
	pbuserrole "github.com/MamangRust/microservice-point-of-sale-pb/user_role"
	"github.com/MamangRust/microservice-point-of-sale-pkg/logger"
	"github.com/MamangRust/microservice-point-of-sale-role/service"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
	"github.com/MamangRust/microservice-point-of-sale-shared/errors"
	"github.com/MamangRust/microservice-point-of-sale-shared/errors/role_errors"
	userrole_errors "github.com/MamangRust/microservice-point-of-sale-shared/errors/user_role_errors"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/emptypb"
)

// userRoleHandleGrpc serves the UserRoleService. Role assignment moved out of the
// role proto, so this handler owns the user<->role write path while reusing the
// role service's query/command implementations.
type userRoleHandleGrpc struct {
	pbuserrole.UnimplementedUserRoleServiceServer
	roleQuery   service.RoleQueryService
	roleCommand service.RoleCommandService
	logger      logger.LoggerInterface
}

func NewUserRoleHandleGrpc(
	service *service.Service,
	logger logger.LoggerInterface,
) pbuserrole.UserRoleServiceServer {
	return &userRoleHandleGrpc{
		roleQuery:   service.RoleQuery,
		roleCommand: service.RoleCommand,
		logger:      logger,
	}
}

func (s *userRoleHandleGrpc) FindByUserId(ctx context.Context, req *pbuserrole.FindByIdUserRoleRequest) (*pbrole.ApiResponsesRole, error) {
	s.logger.Info("FindByUserId called", zap.Int32("userId", req.GetUserId()))

	userID := int(req.GetUserId())
	if userID <= 0 {
		return nil, role_errors.ErrGrpcRoleInvalidId
	}

	roles, err := s.roleQuery.FindByUserId(ctx, userID)
	if err != nil {
		s.logger.Error("FindByUserId failed", zap.Error(err))
		return nil, errors.ToGrpcError(err)
	}

	s.logger.Info("FindByUserId success")

	return &pbrole.ApiResponsesRole{
		Status:  "success",
		Message: "Successfully fetched role by user ID",
		Data:    mapResponsesRoleFromDB(roles),
	}, nil
}

func (s *userRoleHandleGrpc) AssignRoleToUser(ctx context.Context, req *pbuserrole.AssignRoleToUserRequest) (*pbuserrole.ApiResponseUserRole, error) {
	s.logger.Info("AssignRoleToUser called", zap.Int32("userId", req.GetUserId()), zap.Int32("roleId", req.GetRoleId()))

	userID := int(req.GetUserId())
	roleID := int(req.GetRoleId())
	if userID <= 0 || roleID <= 0 {
		return nil, userrole_errors.ErrGrpcValidateAssignRole
	}

	userRole, err := s.roleCommand.AssignRoleToUser(ctx, &requests.CreateUserRoleRequest{
		UserId: userID,
		RoleId: roleID,
	})
	if err != nil {
		s.logger.Error("AssignRoleToUser failed", zap.Error(err))
		return nil, errors.ToGrpcError(err)
	}

	s.logger.Info("AssignRoleToUser success")

	return &pbuserrole.ApiResponseUserRole{
		Status:  "success",
		Message: "Successfully assigned role to user",
		Data: &pbuserrole.UserRoleResponse{
			UserRoleId: userRole.UserRoleID,
			UserId:     userRole.UserID,
			RoleId:     userRole.RoleID,
		},
	}, nil
}

func (s *userRoleHandleGrpc) RemoveRoleFromUser(ctx context.Context, req *pbuserrole.RemoveRoleFromUserRequest) (*emptypb.Empty, error) {
	s.logger.Info("RemoveRoleFromUser called", zap.Int32("userId", req.GetUserId()), zap.Int32("roleId", req.GetRoleId()))

	userID := int(req.GetUserId())
	roleID := int(req.GetRoleId())
	if userID <= 0 || roleID <= 0 {
		return nil, userrole_errors.ErrGrpcValidateRemoveRole
	}

	if err := s.roleCommand.RemoveRoleFromUser(ctx, &requests.RemoveUserRoleRequest{
		UserId: userID,
		RoleId: roleID,
	}); err != nil {
		s.logger.Error("RemoveRoleFromUser failed", zap.Error(err))
		return nil, errors.ToGrpcError(err)
	}

	s.logger.Info("RemoveRoleFromUser success")

	return &emptypb.Empty{}, nil
}
