package api

import (
	"github.com/application-research/edge-ur/core"
	"github.com/labstack/echo/v4"
	"github.com/patrickmn/go-cache"
	"time"
)

var cacheStats = cache.New(5*time.Minute, 10*time.Minute)

type Stats struct {
	TotalContentCount int `json:"total_content_count"`
	TotalSize         int `json:"total_size"`
	TotalSignedUrls   int `json:"total_signed_urls"`
	TotalApiKeys      int `json:"total_api_keys"`
}

func ConfigureStatsRouter(e *echo.Group, node *core.LightNode) {
	e.GET("/stats", func(c echo.Context) error {

		stats, found := cacheStats.Get("stats")
		if found {
			return c.JSON(200, stats)
		}

		var s Stats
		err := node.DB.Raw("select count(*) as total_content_count, sum(size) as total_size from contents").Scan(&s).Error
		if err != nil {
			return c.JSON(500, err)
		}

		err = node.DB.Raw("select count(*) as content_signature_meta").Scan(&s.TotalSignedUrls).Error
		if err != nil {
			return c.JSON(500, err)
		}

		err = node.DB.Raw("select sum(sum) from(select count(*) as sum from contents group by requesting_api_key)").Scan(&s.TotalApiKeys).Error
		if err != nil {
			return c.JSON(500, err)
		}

		cacheStats.Set("stats", &s, cache.DefaultExpiration)
		return c.JSON(200, s)
	})
}
