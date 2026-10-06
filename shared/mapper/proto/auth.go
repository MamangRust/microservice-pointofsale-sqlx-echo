package protomapper

import (
	pbauth "github.com/MamangRust/microservice-point-of-sale-pb/auth"
	pbuser "github.com/MamangRust/microservice-point-of-sale-pb/user"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/response"
)

type authProtoMapper struct {
}

func NewAuthProtoMapper() *authProtoMapper {
	return &authProtoMapper{}
}

func (s *authProtoMapper) ToProtoResponseVerifyCode(status string, message string) *pbauth.ApiResponseVerifyCode {
	return &pbauth.ApiResponseVerifyCode{
		Status:  status,
		Message: message,
	}
}

func (s *authProtoMapper) ToProtoResponseForgotPassword(status string, message string) *pbauth.ApiResponseForgotPassword {
	return &pbauth.ApiResponseForgotPassword{
		Status:  status,
		Message: message,
	}
}

func (s *authProtoMapper) ToProtoResponseResetPassword(status string, message string) *pbauth.ApiResponseResetPassword {
	return &pbauth.ApiResponseResetPassword{
		Status:  status,
		Message: message,
	}
}

func (s *authProtoMapper) ToProtoResponseLogin(status string, message string, response *response.TokenResponse) *pbauth.ApiResponseLogin {
	return &pbauth.ApiResponseLogin{
		Status:  status,
		Message: message,
		Data: &pbauth.TokenResponse{
			AccessToken:  response.AccessToken,
			RefreshToken: response.RefreshToken,
		},
	}
}

func (s *authProtoMapper) ToProtoResponseRegister(status string, message string, response *response.UserResponse) *pbauth.ApiResponseRegister {
	return &pbauth.ApiResponseRegister{
		Status:  status,
		Message: message,
		Data: &pbuser.UserResponse{
			Id:        int32(response.ID),
			Firstname: response.FirstName,
			Lastname:  response.LastName,
			Email:     response.Email,
			CreatedAt: response.CreatedAt,
			UpdatedAt: response.UpdatedAt,
		},
	}
}

func (s *authProtoMapper) ToProtoResponseRefreshToken(status string, message string, response *response.TokenResponse) *pbauth.ApiResponseRefreshToken {
	return &pbauth.ApiResponseRefreshToken{
		Status:  status,
		Message: message,
		Data: &pbauth.TokenResponse{
			AccessToken:  response.AccessToken,
			RefreshToken: response.RefreshToken,
		},
	}
}

func (s *authProtoMapper) ToProtoResponseGetMe(status string, message string, response *response.UserResponse) *pbauth.ApiResponseGetMe {
	return &pbauth.ApiResponseGetMe{
		Status:  status,
		Message: message,
		Data: &pbuser.UserResponse{
			Id:        int32(response.ID),
			Firstname: response.FirstName,
			Lastname:  response.LastName,
			Email:     response.Email,
			CreatedAt: response.CreatedAt,
			UpdatedAt: response.UpdatedAt,
		},
	}
}
