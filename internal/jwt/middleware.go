package jwt

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func TokenAuthMiddleware(h *ProfileHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		err := TokenValid(c.Request)
		if err != nil {
			c.JSON(http.StatusUnauthorized, "unauthorized")
			c.Abort()
			return
		}

		metadata, err := h.TK.ExtractTokenMetadata(c.Request)
		if err != nil {
			c.JSON(http.StatusUnauthorized, "unauthorized")
			return
		}
		// MemberID, err := h.RD.FetchAuth(metadata.TokenUuid)
		// if err != nil {
		// 	c.JSON(http.StatusUnauthorized, "unauthorized")
		// 	return
		// }

		c.Set(MemberKey, metadata.UserId)

		c.Next()
	}
}

// func TokenAuthNoErrorMiddleware(h *ProfileHandler) gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		err := TokenValid(c.Request)
// 		if err != nil {
// 			c.Set(MemberKey, "unauthorized")
// 			c.Abort()
// 			return
// 		}

// 		metadata, err := h.TK.ExtractTokenMetadata(c.Request)
// 		if err != nil {
// 			c.Set(MemberKey, "unauthorized")
// 			return
// 		}
// 		MemberID, err := h.RD.FetchAuth(metadata.TokenUuid)
// 		if err != nil {
// 			c.Set(MemberKey, "unauthorized")
// 			return
// 		}

// 		c.Set(MemberKey, MemberID)

// 		c.Next()
// 	}
// }
