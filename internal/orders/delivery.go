package orders

import "github.com/gin-gonic/gin"

type Handlers interface {
	GetById() gin.HandlerFunc
}
