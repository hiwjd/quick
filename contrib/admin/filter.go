package admin

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/hiwjd/quick"
	"github.com/hiwjd/quick/support/dataperm"
	"github.com/hiwjd/quick/support/session"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// FnCanAccessAPI 检查adminID是否有有个接口的访问权限
type FnCanAccessAPI func(ctx context.Context, adminID uint, method, path string) bool

// FnBuildDataPermAppliers 根据adminID构造出数据权限适配器
type FnBuildDataPermAppliers func(ctx context.Context, adminID uint) []dataperm.Applier

// AdminSessionCheck
func AdminSessionCheck(storage session.Storage, fnCanAccessAPI FnCanAccessAPI, logf quick.Logf) echo.MiddlewareFunc {
	keyAuthConfig := middleware.DefaultKeyAuthConfig
	keyAuthConfig.Validator = func(key string, c echo.Context) (bool, error) {
		req := c.Request()
		method := req.Method
		uri := req.URL.Path
		if uri == "/ana/admin/logout" {
			if err := storage.Del(key); err != nil {
				logf("[ERROR] 登出时删除会话出错: %s, key=%s", err.Error(), key)
			}
			return true, nil
		}

		var session Session
		if err := storage.Get(key, &session); err != nil {
			return false, echo.NewHTTPError(http.StatusUnauthorized).SetInternal(err)
		}

		if !fnCanAccessAPI(req.Context(), session.ID, method, uri) {
			return false, echo.NewHTTPError(http.StatusUnauthorized)
		}

		// appliers := fnBuildDataPermAppliers(req.Context(), session.ID)
		// ctx := dataperm.SetAppliers(req.Context(), appliers)
		// c.SetRequest(req.WithContext(ctx))

		if !session.Remember {
			if er := storage.RefreshTTL(key, 30*time.Minute); er != nil {
				logf("[ERROR] 刷新会话存活时间失败: %s", er.Error())
			}
		}
		c.Set(AdminSessionID, session)

		return true, nil
	}
	keyAuthConfig.Skipper = func(c echo.Context) bool {
		uri := c.Request().RequestURI
		return !strings.HasPrefix(uri, "/ana/admin/")
	}

	return middleware.KeyAuthWithConfig(keyAuthConfig)
}
