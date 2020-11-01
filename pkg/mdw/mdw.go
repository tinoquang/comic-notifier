package mdw

import (
	"fmt"

	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo/v4"
	"github.com/tinoquang/comic-notifier/pkg/util"
)

// CheckLoginStatus validate jwt cookie in request
func CheckLoginStatus(next echo.HandlerFunc) echo.HandlerFunc {
	key := "cc57a4c9-9751-4418-a498-ecf7c3c1d7fd"
	return func(c echo.Context) (err error) {
		jwtCookie, err := c.Cookie("_session")
		if err != nil {
			util.Info("err", err)
			return echo.ErrUnauthorized
		}

		// Validate cookie
		_, err = jwt.Parse(jwtCookie.Value, func(token *jwt.Token) (interface{}, error) {

			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(key), nil
		})
		if err != nil {
			util.Danger(err)
			return echo.ErrUnauthorized
		}

		next(c)
		return
	}
}
