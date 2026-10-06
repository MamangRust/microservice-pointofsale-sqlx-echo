package handler

import (
	"context"
	"math"

	pbcommon "github.com/MamangRust/microservice-point-of-sale-pb/common"
	pbuser "github.com/MamangRust/microservice-point-of-sale-pb/user"
	"github.com/MamangRust/microservice-point-of-sale-pkg/logger"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
	"github.com/MamangRust/microservice-point-of-sale-shared/errors"
	"github.com/MamangRust/microservice-point-of-sale-shared/errors/user_errors"
	db "github.com/MamangRust/microservice-point-of-sale-user/database/schema"
	"github.com/MamangRust/microservice-point-of-sale-user/service"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type userHandleGrpc struct {
	pbuser.UnimplementedUserServiceServer
	userQueryService   service.UserQueryService
	userCommandService service.UserCommandService
	logger             logger.LoggerInterface
}

func NewUserHandleGrpc(
	user *service.Service,
	logger logger.LoggerInterface,
) pbuser.UserServiceServer {
	return &userHandleGrpc{
		userQueryService:   user.UserQuery,
		userCommandService: user.UserCommand,
		logger:             logger,
	}
}

func (s *userHandleGrpc) FindAll(ctx context.Context, request *pbuser.FindAllUserRequest) (*pbuser.ApiResponsePaginationUser, error) {
	s.logger.Info("FindAll users called", zap.Int32("page", request.GetPage()))

	page := int(request.GetPage())
	pageSize := int(request.GetPageSize())
	search := request.GetSearch()

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	reqService := requests.FindAllUsers{
		Page:     page,
		PageSize: pageSize,
		Search:   search,
	}

	users, totalRecords, err := s.userQueryService.FindAll(ctx, &reqService)
	if err != nil {
		s.logger.Error("FindAll users failed", zap.Error(err))
		return nil, errors.ToGrpcError(err)
	}

	var totalRecordsVal int
	if totalRecords != nil {
		totalRecordsVal = *totalRecords
	}
	totalPages := int(math.Ceil(float64(totalRecordsVal) / float64(pageSize)))

	paginationMeta := &pbcommon.PaginationMeta{
		CurrentPage:  int32(page),
		PageSize:     int32(pageSize),
		TotalPages:   int32(totalPages),
		TotalRecords: int32(totalRecordsVal),
	}

	s.logger.Info("FindAll users success")

	return &pbuser.ApiResponsePaginationUser{
		Status:     "success",
		Message:    "Successfully fetched users",
		Data:       mapGetUsersRowsToProto(users),
		Pagination: mapPaginationMeta(paginationMeta),
	}, nil
}

func (s *userHandleGrpc) FindById(ctx context.Context, request *pbuser.FindByIdUserRequest) (*pbuser.ApiResponseUser, error) {
	s.logger.Info("FindById user called", zap.Int32("id", request.GetId()))

	id := int(request.GetId())
	if id <= 0 {
		return nil, user_errors.ErrGrpcUserNotFound
	}

	user, err := s.userQueryService.FindByID(ctx, id)
	if err != nil {
		s.logger.Error("FindById user failed", zap.Error(err))
		return nil, errors.ToGrpcError(err)
	}

	s.logger.Info("FindById user success")

	return &pbuser.ApiResponseUser{
		Status:  "success",
		Message: "Successfully fetched user",
		Data:    mapUserToProto(user),
	}, nil
}

func (s *userHandleGrpc) FindByActive(ctx context.Context, request *pbuser.FindAllUserRequest) (*pbuser.ApiResponsePaginationUserDeleteAt, error) {
	s.logger.Info("FindByActive users called", zap.Int32("page", request.GetPage()))

	page := int(request.GetPage())
	pageSize := int(request.GetPageSize())
	search := request.GetSearch()

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	reqService := requests.FindAllUsers{
		Page:     page,
		PageSize: pageSize,
		Search:   search,
	}

	users, totalRecords, err := s.userQueryService.FindByActive(ctx, &reqService)
	if err != nil {
		s.logger.Error("FindByActive users failed", zap.Error(err))
		return nil, errors.ToGrpcError(err)
	}

	var totalRecordsVal int
	if totalRecords != nil {
		totalRecordsVal = *totalRecords
	}
	totalPages := int(math.Ceil(float64(totalRecordsVal) / float64(pageSize)))

	paginationMeta := &pbcommon.PaginationMeta{
		CurrentPage:  int32(page),
		PageSize:     int32(pageSize),
		TotalPages:   int32(totalPages),
		TotalRecords: int32(totalRecordsVal),
	}

	s.logger.Info("FindByActive users success")

	return &pbuser.ApiResponsePaginationUserDeleteAt{
		Status:     "success",
		Message:    "Successfully fetched active users",
		Data:       mapActiveUsersToProto(users),
		Pagination: mapPaginationMeta(paginationMeta),
	}, nil
}

func (s *userHandleGrpc) FindByTrashed(ctx context.Context, request *pbuser.FindAllUserRequest) (*pbuser.ApiResponsePaginationUserDeleteAt, error) {
	s.logger.Info("FindByTrashed users called")

	page := int(request.GetPage())
	pageSize := int(request.GetPageSize())
	search := request.GetSearch()

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	reqService := requests.FindAllUsers{
		Page:     page,
		PageSize: pageSize,
		Search:   search,
	}

	users, totalRecords, err := s.userQueryService.FindByTrashed(ctx, &reqService)
	if err != nil {
		s.logger.Error("FindByTrashed users failed", zap.Error(err))
		return nil, errors.ToGrpcError(err)
	}

	var totalRecordsVal int
	if totalRecords != nil {
		totalRecordsVal = *totalRecords
	}
	totalPages := int(math.Ceil(float64(totalRecordsVal) / float64(pageSize)))

	paginationMeta := &pbcommon.PaginationMeta{
		CurrentPage:  int32(page),
		PageSize:     int32(pageSize),
		TotalPages:   int32(totalPages),
		TotalRecords: int32(totalRecordsVal),
	}

	s.logger.Info("FindByTrashed users success")

	return &pbuser.ApiResponsePaginationUserDeleteAt{
		Status:     "success",
		Message:    "Successfully fetched trashed users",
		Data:       mapTrashedUsersToProto(users),
		Pagination: mapPaginationMeta(paginationMeta),
	}, nil
}

func (s *userHandleGrpc) Create(ctx context.Context, request *pbuser.CreateUserRequest) (*pbuser.ApiResponseUser, error) {
	s.logger.Info("Create user called", zap.String("email", request.GetEmail()))

	req := &requests.CreateUserRequest{
		FirstName:       request.GetFirstname(),
		LastName:        request.GetLastname(),
		Email:           request.GetEmail(),
		Password:        request.GetPassword(),
		ConfirmPassword: request.GetConfirmPassword(),
	}

	if err := req.Validate(); err != nil {
		return nil, user_errors.ErrGrpcValidateCreateUser
	}

	user, err := s.userCommandService.CreateUser(ctx, req)
	if err != nil {
		s.logger.Error("Create user failed", zap.Error(err))
		return nil, errors.ToGrpcError(err)
	}

	s.logger.Info("Create user success")

	return &pbuser.ApiResponseUser{
		Status:  "success",
		Message: "Successfully created user",
		Data:    mapUserToProto(user),
	}, nil
}

func (s *userHandleGrpc) Update(ctx context.Context, request *pbuser.UpdateUserRequest) (*pbuser.ApiResponseUser, error) {
	s.logger.Info("Update user called", zap.Int32("id", request.GetId()))

	id := int(request.GetId())
	if id <= 0 {
		return nil, user_errors.ErrGrpcUserInvalidId
	}

	req := &requests.UpdateUserRequest{
		UserID:          &id,
		FirstName:       request.GetFirstname(),
		LastName:        request.GetLastname(),
		Email:           request.GetEmail(),
		Password:        request.GetPassword(),
		ConfirmPassword: request.GetConfirmPassword(),
	}

	if err := req.Validate(); err != nil {
		return nil, user_errors.ErrGrpcValidateCreateUser
	}

	user, err := s.userCommandService.UpdateUser(ctx, req)
	if err != nil {
		s.logger.Error("Update user failed", zap.Error(err))
		return nil, errors.ToGrpcError(err)
	}

	s.logger.Info("Update user success")

	return &pbuser.ApiResponseUser{
		Status:  "success",
		Message: "Successfully updated user",
		Data:    mapUserToProto(user),
	}, nil
}

func (s *userHandleGrpc) TrashedUser(ctx context.Context, request *pbuser.FindByIdUserRequest) (*pbuser.ApiResponseUserDeleteAt, error) {
	s.logger.Info("TrashedUser called", zap.Int32("id", request.GetId()))

	id := int(request.GetId())
	if id <= 0 {
		return nil, user_errors.ErrGrpcUserInvalidId
	}

	user, err := s.userCommandService.TrashedUser(ctx, id)
	if err != nil {
		s.logger.Error("TrashedUser failed", zap.Error(err))
		return nil, errors.ToGrpcError(err)
	}

	s.logger.Info("TrashedUser success")

	return &pbuser.ApiResponseUserDeleteAt{
		Status:  "success",
		Message: "Successfully trashed user",
		Data:    mapUserDeleteAtToProto(user),
	}, nil
}

func (s *userHandleGrpc) RestoreUser(ctx context.Context, request *pbuser.FindByIdUserRequest) (*pbuser.ApiResponseUserDeleteAt, error) {
	s.logger.Info("RestoreUser called", zap.Int32("id", request.GetId()))

	id := int(request.GetId())
	if id <= 0 {
		return nil, user_errors.ErrGrpcUserInvalidId
	}

	user, err := s.userCommandService.RestoreUser(ctx, id)
	if err != nil {
		s.logger.Error("RestoreUser failed", zap.Error(err))
		return nil, errors.ToGrpcError(err)
	}

	s.logger.Info("RestoreUser success")

	return &pbuser.ApiResponseUserDeleteAt{
		Status:  "success",
		Message: "Successfully restored user",
		Data:    mapUserDeleteAtToProto(user),
	}, nil
}

func (s *userHandleGrpc) DeleteUserPermanent(ctx context.Context, request *pbuser.FindByIdUserRequest) (*pbuser.ApiResponseUserDelete, error) {
	s.logger.Info("DeleteUserPermanent called", zap.Int32("id", request.GetId()))

	id := int(request.GetId())
	if id <= 0 {
		return nil, user_errors.ErrGrpcUserInvalidId
	}

	_, err := s.userCommandService.DeleteUserPermanent(ctx, id)
	if err != nil {
		s.logger.Error("DeleteUserPermanent failed", zap.Error(err))
		return nil, errors.ToGrpcError(err)
	}

	s.logger.Info("DeleteUserPermanent success")

	return &pbuser.ApiResponseUserDelete{
		Status:  "success",
		Message: "Successfully deleted user permanently",
	}, nil
}

func (s *userHandleGrpc) RestoreAllUser(ctx context.Context, _ *emptypb.Empty) (*pbuser.ApiResponseUserAll, error) {
	s.logger.Info("RestoreAllUser called")

	_, err := s.userCommandService.RestoreAllUser(ctx)
	if err != nil {
		s.logger.Error("RestoreAllUser failed", zap.Error(err))
		return nil, errors.ToGrpcError(err)
	}

	s.logger.Info("RestoreAllUser success")

	return &pbuser.ApiResponseUserAll{
		Status:  "success",
		Message: "Successfully restore all user",
	}, nil
}

func (s *userHandleGrpc) DeleteAllUserPermanent(ctx context.Context, _ *emptypb.Empty) (*pbuser.ApiResponseUserAll, error) {
	s.logger.Info("DeleteAllUserPermanent called")

	_, err := s.userCommandService.DeleteAllUserPermanent(ctx)
	if err != nil {
		s.logger.Error("DeleteAllUserPermanent failed", zap.Error(err))
		return nil, errors.ToGrpcError(err)
	}

	s.logger.Info("DeleteAllUserPermanent success")

	return &pbuser.ApiResponseUserAll{
		Status:  "success",
		Message: "Successfully delete user permanen",
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

func mapUserToProto(user *db.User) *pbuser.UserResponse {
	if user == nil {
		return nil
	}
	var createdAtStr, updatedAtStr string
	if user.CreatedAt.Valid {
		createdAtStr = user.CreatedAt.Time.Format("2006-01-02 15:04:05")
	}
	if user.UpdatedAt.Valid {
		updatedAtStr = user.UpdatedAt.Time.Format("2006-01-02 15:04:05")
	}
	return &pbuser.UserResponse{
		Id:        int32(user.UserID),
		Firstname: user.Firstname,
		Lastname:  user.Lastname,
		Email:     user.Email,
		CreatedAt: createdAtStr,
		UpdatedAt: updatedAtStr,
	}
}

func mapGetUsersRowToProto(user *db.GetUsersRow) *pbuser.UserResponse {
	if user == nil {
		return nil
	}
	var createdAtStr, updatedAtStr string
	if user.CreatedAt.Valid {
		createdAtStr = user.CreatedAt.Time.Format("2006-01-02 15:04:05")
	}
	if user.UpdatedAt.Valid {
		updatedAtStr = user.UpdatedAt.Time.Format("2006-01-02 15:04:05")
	}
	return &pbuser.UserResponse{
		Id:        int32(user.UserID),
		Firstname: user.Firstname,
		Lastname:  user.Lastname,
		Email:     user.Email,
		CreatedAt: createdAtStr,
		UpdatedAt: updatedAtStr,
	}
}

func mapGetUsersRowsToProto(users []*db.GetUsersRow) []*pbuser.UserResponse {
	var responseUsers []*pbuser.UserResponse
	for _, u := range users {
		responseUsers = append(responseUsers, mapGetUsersRowToProto(u))
	}
	return responseUsers
}

func mapUserDeleteAtToProto(user *db.User) *pbuser.UserResponseDeleteAt {
	if user == nil {
		return nil
	}
	var createdAtStr, updatedAtStr string
	var deletedAt *wrapperspb.StringValue
	if user.CreatedAt.Valid {
		createdAtStr = user.CreatedAt.Time.Format("2006-01-02 15:04:05")
	}
	if user.UpdatedAt.Valid {
		updatedAtStr = user.UpdatedAt.Time.Format("2006-01-02 15:04:05")
	}
	if user.DeletedAt.Valid {
		deletedAt = wrapperspb.String(user.DeletedAt.Time.Format("2006-01-02 15:04:05"))
	}
	return &pbuser.UserResponseDeleteAt{
		Id:        int32(user.UserID),
		Firstname: user.Firstname,
		Lastname:  user.Lastname,
		Email:     user.Email,
		CreatedAt: createdAtStr,
		UpdatedAt: updatedAtStr,
		DeletedAt: deletedAt,
	}
}

func mapActiveUserToProto(user *db.GetUsersActiveRow) *pbuser.UserResponseDeleteAt {
	if user == nil {
		return nil
	}
	var createdAtStr, updatedAtStr string
	var deletedAt *wrapperspb.StringValue
	if user.CreatedAt.Valid {
		createdAtStr = user.CreatedAt.Time.Format("2006-01-02 15:04:05")
	}
	if user.UpdatedAt.Valid {
		updatedAtStr = user.UpdatedAt.Time.Format("2006-01-02 15:04:05")
	}
	if user.DeletedAt.Valid {
		deletedAt = wrapperspb.String(user.DeletedAt.Time.Format("2006-01-02 15:04:05"))
	}
	return &pbuser.UserResponseDeleteAt{
		Id:        int32(user.UserID),
		Firstname: user.Firstname,
		Lastname:  user.Lastname,
		Email:     user.Email,
		CreatedAt: createdAtStr,
		UpdatedAt: updatedAtStr,
		DeletedAt: deletedAt,
	}
}

func mapActiveUsersToProto(users []*db.GetUsersActiveRow) []*pbuser.UserResponseDeleteAt {
	var responseUsers []*pbuser.UserResponseDeleteAt
	for _, u := range users {
		responseUsers = append(responseUsers, mapActiveUserToProto(u))
	}
	return responseUsers
}

func mapTrashedUserToProto(user *db.GetUserTrashedRow) *pbuser.UserResponseDeleteAt {
	if user == nil {
		return nil
	}
	var createdAtStr, updatedAtStr string
	var deletedAt *wrapperspb.StringValue
	if user.CreatedAt.Valid {
		createdAtStr = user.CreatedAt.Time.Format("2006-01-02 15:04:05")
	}
	if user.UpdatedAt.Valid {
		updatedAtStr = user.UpdatedAt.Time.Format("2006-01-02 15:04:05")
	}
	if user.DeletedAt.Valid {
		deletedAt = wrapperspb.String(user.DeletedAt.Time.Format("2006-01-02 15:04:05"))
	}
	return &pbuser.UserResponseDeleteAt{
		Id:        int32(user.UserID),
		Firstname: user.Firstname,
		Lastname:  user.Lastname,
		Email:     user.Email,
		CreatedAt: createdAtStr,
		UpdatedAt: updatedAtStr,
		DeletedAt: deletedAt,
	}
}

func mapTrashedUsersToProto(users []*db.GetUserTrashedRow) []*pbuser.UserResponseDeleteAt {
	var responseUsers []*pbuser.UserResponseDeleteAt
	for _, u := range users {
		responseUsers = append(responseUsers, mapTrashedUserToProto(u))
	}
	return responseUsers
}
