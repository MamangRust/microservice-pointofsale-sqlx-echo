package userrole_errors

import (
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/response"

	"google.golang.org/grpc/codes"
)

var (
	ErrGrpcValidateAssignRole = response.NewGrpcError("error", "validation failed: invalid assign role request", int(codes.InvalidArgument))
	ErrGrpcValidateRemoveRole = response.NewGrpcError("error", "validation failed: invalid remove role request", int(codes.InvalidArgument))
)
