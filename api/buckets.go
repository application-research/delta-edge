package api

import (
	"github.com/application-research/edge-ur/core"
	"github.com/labstack/echo/v4"
	"time"
)

type BucketsResponse struct {
	BucketUUID  string    `json:"bucket_uuid"`
	BucketID    int64     `json:"bucket_id"`
	PieceCid    string    `json:"piece_cid"`
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
	buckets.GET("/get-open-buckets", handleGetOpenBuckets(node))

}

func handleGetOpenBuckets(node *core.LightNode) func(c echo.Context) error {
	return func(c echo.Context) error {
		var buckets []core.Bucket
		node.DB.Model(&core.Bucket{}).Where("status = ?", "ready-for-deal-making").Find(&buckets)

		var bucketsResponse []BucketsResponse
		for _, bucket := range buckets {
			bucketsResponse = append(bucketsResponse, BucketsResponse{
				BucketUUID:  bucket.Uuid,
				BucketID:    bucket.ID,
				PieceCid:    bucket.PieceCid,
				PieceSize:   bucket.PieceSize,
				DownloadUrl: "/gw/" + bucket.Cid,
				Status:      bucket.Status,
				Size:        bucket.Size,
				CreatedAt:   bucket.CreatedAt,
				UpdatedAt:   bucket.UpdatedAt,
			})
		}

		return c.JSON(200, bucketsResponse)
	}
}
