package middleware

import (
	"context"
	"encoding/json"
	"strings"

	"crow/internal/biz"

	kratosmiddleware "github.com/go-kratos/kratos/v3/middleware"
	kratoshttp "github.com/go-kratos/kratos/v3/transport/http"
)

// NewAdminOperationLog writes successful admin mutation requests into admin_operation_log.
func NewAdminOperationLog(uc *biz.AdminOperationLogUsecase) kratosmiddleware.Middleware {
	return func(next kratosmiddleware.Handler) kratosmiddleware.Handler {
		return func(ctx context.Context, req any) (any, error) {
			reply, err := next(ctx, req)
			if err != nil || uc == nil {
				return reply, err
			}

			httpReq, ok := kratoshttp.RequestFromServerContext(ctx)
			if !ok || !shouldLogOperation(httpReq.Method, httpReq.URL.Path) {
				return reply, err
			}

			principal, ok := biz.AdminPrincipalFromContext(ctx)
			if !ok || principal == nil {
				return reply, err
			}

			module, action, description := describeOperation(httpReq.Method, httpReq.URL.Path)
			if module == "" || action == "" {
				return reply, err
			}

			_, _ = uc.Create(ctx, &biz.AdminOperationLog{
				AdminID:       principal.ID,
				AdminName:     principal.Account,
				Module:        module,
				Action:        action,
				Description:   description,
				RequestMethod: httpReq.Method,
				RequestURL:    httpReq.URL.Path,
				RequestParams: marshalRequestParams(req),
			})

			return reply, err
		}
	}
}

func shouldLogOperation(method, path string) bool {
	if !shouldProtectAdminPath(path) {
		return false
	}
	switch strings.ToUpper(method) {
	case "POST", "PUT", "DELETE":
		return true
	default:
		return false
	}
}

func describeOperation(method, path string) (module, action, description string) {
	switch {
	case strings.HasPrefix(path, "/v1/admins"):
		module = "admin"
		description = "管理员"
	case strings.HasPrefix(path, "/v1/roles"):
		module = "role"
		description = "角色"
	case strings.HasPrefix(path, "/v1/permissions"):
		module = "permission"
		description = "权限"
	case strings.HasPrefix(path, "/v1/admin-roles"):
		module = "admin_role"
		description = "管理员角色关联"
	case strings.HasPrefix(path, "/v1/group-permissions"):
		module = "role_permission"
		description = "角色权限关联"
	case strings.HasPrefix(path, "/v1/cps"):
		module = "cp"
		description = "内容提供商"
	case strings.HasPrefix(path, "/v1/sps"):
		module = "sp"
		description = "内容服务商"
	case strings.HasPrefix(path, "/v1/cp-sps"):
		module = "cp_sp"
		description = "内容提供商与内容服务商绑定"
	default:
		return "", "", ""
	}

	switch strings.ToUpper(method) {
	case "POST":
		return module, "create", "创建" + description
	case "PUT":
		return module, "update", "更新" + description
	case "DELETE":
		return module, "delete", "删除" + description
	default:
		return "", "", ""
	}
}

func marshalRequestParams(req any) string {
	if req == nil {
		return ""
	}

	raw, err := json.Marshal(req)
	if err != nil {
		return ""
	}

	var payload any
	if err := json.Unmarshal(raw, &payload); err != nil {
		return string(raw)
	}

	sanitizeSensitiveFields(payload)

	raw, err = json.Marshal(payload)
	if err != nil {
		return ""
	}
	return string(raw)
}

func sanitizeSensitiveFields(value any) {
	switch typed := value.(type) {
	case map[string]any:
		for key, item := range typed {
			if strings.Contains(strings.ToLower(key), "password") {
				typed[key] = "***"
				continue
			}
			sanitizeSensitiveFields(item)
		}
	case []any:
		for _, item := range typed {
			sanitizeSensitiveFields(item)
		}
	}
}
