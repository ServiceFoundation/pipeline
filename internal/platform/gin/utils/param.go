// Copyright © 2018 Banzai Cloud
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ginutils

import (
	"net/http"
	"strconv"

	"github.com/banzaicloud/pipeline/pkg/common"
	"github.com/gin-gonic/gin"
)

// UintParam returns a parameter parsed as uint or responds with an error.
func UintParam(ctx *gin.Context, paramName string) (uint, bool) {
	value := ctx.Param(paramName)
	uintValue, err := strconv.ParseUint(value, 10, 0)
	if err != nil {
		ReplyWithErrorResponse(ctx, &common.ErrorResponse{
			Code:    http.StatusBadRequest,
			Error:   "Invalid parameter",
			Message: "Parameter 'id' must be a positive, numeric value",
		})

		return 0, false
	}

	return uint(uintValue), true
}
