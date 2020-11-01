package mdw

import (
	"github.com/labstack/echo/v4"
	"github.com/tinoquang/comic-notifier/pkg/util"
)

// CheckLoginStatus validate jwt cookie in request
func CheckLoginStatus(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) (err error) {

		_, err = c.Cookie("_session")
		if err != nil {
			util.Info("err", err)
			return echo.ErrUnauthorized
		}

		next(c)
		return
	}
}
