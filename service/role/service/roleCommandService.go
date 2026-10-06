package service

import (
	"context"

	db "github.com/MamangRust/microservice-point-of-sale-role/database/schema"
	"github.com/MamangRust/microservice-point-of-sale-pkg/logger"
	mencache "github.com/MamangRust/microservice-point-of-sale-role/cache"
	"github.com/MamangRust/microservice-point-of-sale-role/repository"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
	sharederrorhandler "github.com/MamangRust/microservice-point-of-sale-shared/errorhandler"
	"github.com/MamangRust/microservice-point-of-sale-shared/errors/role_errors"
	userrole_errors "github.com/MamangRust/microservice-point-of-sale-shared/errors/user_role_errors"
	"github.com/MamangRust/microservice-point-of-sale-shared/observability"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
)

type roleCommandService struct {
	mencache        mencache.RoleCommandCache
	roleCommand     repository.RoleCommandRepository
	userRoleCommand repository.UserRoleCommandRepository
	logger          logger.LoggerInterface
	observability   observability.TraceLoggerObservability
}

func NewRoleCommandService(
	mencache mencache.RoleCommandCache,
	roleCommand repository.RoleCommandRepository,
	userRoleCommand repository.UserRoleCommandRepository,
	logger logger.LoggerInterface,
	obs observability.TraceLoggerObservability,
) *roleCommandService {
	return &roleCommandService{
		mencache:        mencache,
		roleCommand:     roleCommand,
		userRoleCommand: userRoleCommand,
		logger:          logger,
		observability:   obs,
	}
}

func (s *roleCommandService) CreateRole(ctx context.Context, request *requests.CreateRoleRequest) (*db.Role, error) {
	const method = "CreateRole"

	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method, attribute.String("name", request.Name))
	defer func() {
		end(status)
	}()

	role, err := s.roleCommand.CreateRole(ctx, request)
	if err != nil {
		status = "error"
		return sharederrorhandler.HandleError[*db.Role](
			s.logger,
			role_errors.ErrFailedCreateRole.WithInternal(err),
			method,
			span,
			zap.Error(err),
		)
	}

	logSuccess("Successfully created role", zap.Int32("role.id", role.RoleID), zap.Bool("success", true))
	return role, nil
}

func (s *roleCommandService) UpdateRole(ctx context.Context, request *requests.UpdateRoleRequest) (*db.Role, error) {
	const method = "UpdateRole"

	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method, attribute.Int("id", *request.ID))
	defer func() {
		end(status)
	}()

	role, err := s.roleCommand.UpdateRole(ctx, request)
	if err != nil {
		status = "error"
		return sharederrorhandler.HandleError[*db.Role](
			s.logger,
			role_errors.ErrFailedUpdateRole.WithInternal(err),
			method,
			span,
			zap.Error(err),
		)
	}

	s.mencache.DeleteCachedRole(ctx, *request.ID)
	logSuccess("Successfully updated role", zap.Int32("role.id", role.RoleID), zap.Bool("success", true))
	return role, nil
}

func (s *roleCommandService) TrashedRole(ctx context.Context, id int) (*db.Role, error) {
	const method = "TrashedRole"

	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method, attribute.Int("id", id))
	defer func() {
		end(status)
	}()

	role, err := s.roleCommand.TrashedRole(ctx, id)
	if err != nil {
		status = "error"
		return sharederrorhandler.HandleError[*db.Role](
			s.logger,
			role_errors.ErrFailedTrashedRole.WithInternal(err),
			method,
			span,
			zap.Error(err),
		)
	}

	s.mencache.DeleteCachedRole(ctx, id)
	logSuccess("Successfully trashed role", zap.Int32("role.id", role.RoleID), zap.Bool("success", true))
	return role, nil
}

func (s *roleCommandService) RestoreRole(ctx context.Context, id int) (*db.Role, error) {
	const method = "RestoreRole"

	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method, attribute.Int("id", id))
	defer func() {
		end(status)
	}()

	role, err := s.roleCommand.RestoreRole(ctx, id)
	if err != nil {
		status = "error"
		return sharederrorhandler.HandleError[*db.Role](
			s.logger,
			role_errors.ErrFailedRestoreRole.WithInternal(err),
			method,
			span,
			zap.Error(err),
		)
	}

	s.mencache.DeleteCachedRole(ctx, id)
	logSuccess("Successfully restored role", zap.Int32("role.id", role.RoleID), zap.Bool("success", true))
	return role, nil
}

func (s *roleCommandService) DeleteRolePermanent(ctx context.Context, id int) (bool, error) {
	const method = "DeleteRolePermanent"

	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method, attribute.Int("id", id))
	defer func() {
		end(status)
	}()

	success, err := s.roleCommand.DeleteRolePermanent(ctx, id)
	if err != nil {
		status = "error"
		return sharederrorhandler.HandleError[bool](
			s.logger,
			role_errors.ErrFailedDeletePermanent.WithInternal(err),
			method,
			span,
			zap.Error(err),
		)
	}

	s.mencache.DeleteCachedRole(ctx, id)
	logSuccess("Successfully deleted role permanently", zap.Int("role.id", id), zap.Bool("success", success))
	return success, nil
}

func (s *roleCommandService) RestoreAllRole(ctx context.Context) (bool, error) {
	const method = "RestoreAllRole"

	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method)
	defer func() {
		end(status)
	}()

	success, err := s.roleCommand.RestoreAllRole(ctx)
	if err != nil {
		status = "error"
		return sharederrorhandler.HandleError[bool](
			s.logger,
			role_errors.ErrFailedRestoreAll.WithInternal(err),
			method,
			span,
			zap.Error(err),
		)
	}

	s.mencache.DeleteCachedRoleAllCache(ctx)
	logSuccess("Successfully restored all roles", zap.Bool("success", success))
	return success, nil
}

func (s *roleCommandService) DeleteAllRolePermanent(ctx context.Context) (bool, error) {
	const method = "DeleteAllRolePermanent"

	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method)
	defer func() {
		end(status)
	}()

	success, err := s.roleCommand.DeleteAllRolePermanent(ctx)
	if err != nil {
		status = "error"
		return sharederrorhandler.HandleError[bool](
			s.logger,
			role_errors.ErrFailedDeleteAll.WithInternal(err),
			method,
			span,
			zap.Error(err),
		)
	}

	s.mencache.DeleteCachedRoleAllCache(ctx)
	logSuccess("Successfully deleted all roles permanently", zap.Bool("success", success))
	return success, nil
}

func (s *roleCommandService) AssignRoleToUser(ctx context.Context, request *requests.CreateUserRoleRequest) (*db.UserRole, error) {
	const method = "AssignRoleToUser"

	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method,
		attribute.Int("user.id", request.UserId), attribute.Int("role.id", request.RoleId))
	defer func() {
		end(status)
	}()

	userRole, err := s.userRoleCommand.AssignRoleToUser(ctx, request)
	if err != nil {
		status = "error"
		return sharederrorhandler.HandleError[*db.UserRole](
			s.logger,
			userrole_errors.ErrFailedAssignRoleToUser.WithInternal(err),
			method,
			span,
			zap.Int("user.id", request.UserId),
			zap.Int("role.id", request.RoleId),
			zap.Error(err),
		)
	}

	s.mencache.DeleteCachedRoleByUserId(ctx, request.UserId)
	logSuccess("Successfully assigned role to user", zap.Int32("user_role.id", userRole.UserRoleID), zap.Bool("success", true))
	return userRole, nil
}

func (s *roleCommandService) RemoveRoleFromUser(ctx context.Context, request *requests.RemoveUserRoleRequest) error {
	const method = "RemoveRoleFromUser"

	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method,
		attribute.Int("user.id", request.UserId), attribute.Int("role.id", request.RoleId))
	defer func() {
		end(status)
	}()

	if err := s.userRoleCommand.RemoveRoleFromUser(ctx, request); err != nil {
		status = "error"
		_, err = sharederrorhandler.HandleError[error](
			s.logger,
			userrole_errors.ErrFailedRemoveRole.WithInternal(err),
			method,
			span,
			zap.Int("user.id", request.UserId),
			zap.Int("role.id", request.RoleId),
			zap.Error(err),
		)
		return err
	}

	s.mencache.DeleteCachedRoleByUserId(ctx, request.UserId)
	logSuccess("Successfully removed role from user", zap.Int("user.id", request.UserId), zap.Bool("success", true))
	return nil
}
