package api

import (
	"github.com/application-research/edge-ur/core"
	"github.com/labstack/echo/v4"
	"strings"
	"time"
)

type BucketsResponse struct {
	BucketUUID  string    `json:"bucket_uuid"`
	PieceCid    string    `json:"piece_cid"`
	PayloadCid  string    `json:"payload_cid"`
	DirCid      string    `json:"dir_cid"`
	PieceSize   int64     `json:"piece_size"`
	DownloadUrl string    `json:"download_url"`
	Status      string    `json:"status"`
	Size        int64     `json:"size"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func ConfigureBucketsRouter(e *echo.Group, node *core.LightNode) {
	//var DeltaUploadApi = node.Config.Delta.ApiUrl
	buckets := e.Group("/buckets")
	buckets.GET("/get-open", handleGetOpenBuckets(node))
	buckets.DELETE("/:uuid", handleDeleteBucket(node))

}
func handleDeleteBucket(node *core.LightNode) func(c echo.Context) error {
	return func(c echo.Context) error {

		// check if its being called by the admin api key
		authorizationString := c.Request().Header.Get("Authorization")
		authParts := strings.Split(authorizationString, " ")
		if authParts[1] != node.Config.Node.AdminApiKey {
			return c.JSON(401, map[string]interface{}{
				"message": "Unauthorized",
			})
		}

		node.DB.Model(&core.Bucket{}).Where("uuid = ?", c.Param("uuid")).Update("status", "deleted")
		return c.JSON(200, map[string]interface{}{
			"message": "Bucket deleted",
			"bucket":  c.Param("uuid"),
		})
	}
}
func handleGetOpenBuckets(node *core.LightNode) func(c echo.Context) error {
	return func(c echo.Context) error {
		var buckets []core.Bucket
		node.DB.Model(&core.Bucket{}).Where("status = ?", "ready-for-deal-making").Find(&buckets)

		var bucketsResponse []BucketsResponse
		for _, bucket := range buckets {
			bucketsResponse = append(bucketsResponse, BucketsResponse{
				BucketUUID: bucket.Uuid,
				PieceCid:   bucket.PieceCid,
				PieceSize:  bucket.PieceSize,
				PayloadCid: bucket.Cid,
				DirCid:     bucket.DirCid,
				//DownloadUrl: "<a href=/gw/" + bucket.Cid + ">" + bucket.PieceCid + "</a>",
				DownloadUrl: "/gw/" + bucket.Cid,
				Status:      bucket.Status,
				Size:        bucket.Size,
				CreatedAt:   bucket.CreatedAt,
				UpdatedAt:   bucket.UpdatedAt,
			})
		}

		if len(bucketsResponse) == 0 {
			return c.JSON(404, map[string]interface{}{
				"message":     "No open buckets found.",
				"description": "This means that there are no buckets that are ready for deal making.",
			})
		}
		return c.JSON(200, bucketsResponse)
	}
}
