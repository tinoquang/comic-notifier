package mdw

import (
	"fmt"

	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo/v4"
	"github.com/tinoquang/comic-notifier/pkg/conf"
	"github.com/tinoquang/comic-notifier/pkg/util"
)

var (
	jwtKey string
)

// SetConfig set jwtKey value for parsing JWT cookie
func SetConfig(cfg *conf.Config) {
	jwtKey = cfg.JWT
}

// CheckLoginStatus validate jwt cookie in request
func CheckLoginStatus(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) (err error) {
		jwtCookie, err := c.Cookie("_session")
		if err != nil {
			return echo.ErrUnauthorized
		}

		// Validate cookie
		_, err = jwt.Parse(jwtCookie.Value, func(token *jwt.Token) (interface{}, error) {

			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(jwtKey), nil
		})
		if err != nil {
			util.Danger(err)
			return echo.ErrUnauthorized
		}

		next(c)
		return
	}
}
