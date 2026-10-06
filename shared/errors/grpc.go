package errors

import (
	"encoding/json"

	pbcommon "github.com/MamangRust/microservice-point-of-sale-pb/common"
)

func GrpcErrorToJson(err *pbcommon.ErrorResponse) string {
	jsonData, _ := json.Marshal(err)
	return string(jsonData)
}
