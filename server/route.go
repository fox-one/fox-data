package server

import (
	"fmt"

	"github.com/fox-one/gin-contrib/session"
	"github.com/gin-gonic/contrib/cors"
	"github.com/gin-gonic/gin"
)

// Run run a http server on port
func Run(s *session.Session, port int) error {
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(cors.New(cors.Config{
		AllowOriginFunc: func(orgin string) bool {
			return true
		},
	}))
	router.Use(gin.Recovery())
	router.Use(bindFoxSession(s))

	RegisterPriceRoute(router)

	return router.Run(fmt.Sprintf(":%d", port))
}

const (
	foxSessionContextKey = "fox.session.context.key"
)

func bindFoxSession(s *session.Session) gin.HandlerFunc {
	return func(c *gin.Context) {
		s := s.WithContext(c.Request.Context())
		c.Set(foxSessionContextKey, s)
	}
}

// FoxSession get fox session in gin context
func FoxSession(c *gin.Context) *session.Session {
	return c.MustGet(foxSessionContextKey).(*session.Session)
}
