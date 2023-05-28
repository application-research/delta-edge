package api

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"github.com/application-research/edge-ur/core"
	"github.com/application-research/edge-ur/jobs"
	"github.com/filecoin-project/go-data-segment/datasegment"
	"github.com/filecoin-project/go-data-segment/merkletree"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/ipfs/go-cid"
	"github.com/labstack/echo/v4"
	"strconv"
)

type StatusCheckBySubPieceCidResponse struct {
	ContentInfo struct {
		Cid   string `json:"cid"`
		Name  string `json:"name"`
		Size  int64  `json:"size"`
		Miner string `json:"miner"`
	} `json:"content_info,omitempty"`
	SubPieceInfo struct {
		PieceCid string `json:"piece_cid"`
		Size     int64  `json:"size"`
		CommPa   string `json:"comm_pa"`
		SizePa   int64  `json:"size_pa"`
		CommPc   string `json:"comm_pc"`
		SizePc   int64  `json:"size_pc"`
		Status   string `json:"status"`
		//InclusionProof datasegment.InclusionProof        `json:"inclusion_proof"`
		InclusionProof struct {
			ProofIndex struct {
				Index string   `json:"index"`
				Path  []string `json:"path"`
			} `json:"proofIndex"`
			ProofSubtree struct {
				Index string   `json:"index"`
				Path  []string `json:"path"`
			} `json:"proofSubtree"`
		} `json:"inclusion_proof"`
		VerifierData datasegment.InclusionVerifierData `json:"verifier_data"`
	} `json:"sub_piece_info,omitempty"`
	DealInfo DealInfo `json:"deal_info,omitempty"`
	Message  string   `json:"message,omitempty"`
}
type DealInfo struct {
	DealUuid  string `json:"deal_uuid"`
	DealID    int64  `json:"deal_id"`
	Status    string `json:"status"`
	DeltaNode string `json:"delta_node"`
}

func ConfigureOpenStatusCheckRouter(e *echo.Group, node *core.LightNode) {
	e.GET("/status/content/:id", func(c echo.Context) error {
		var response StatusCheckBySubPieceCidResponse

		var content core.Content
		node.DB.Model(&core.Content{}).Where("id = ?", c.Param("id")).First(&content)

		response.ContentInfo.Cid = content.Cid
		response.ContentInfo.Name = content.Name
		response.ContentInfo.Size = content.Size
		response.ContentInfo.Miner = content.Miner

		// get the inclusion proof to get the aggregate piece cid
		if content.PieceCid != "" {
			ip := new(datasegment.InclusionProof)
			vd := new(datasegment.InclusionVerifierData)

			readerIp := bytes.NewReader(content.InclusionProof)
			readerVd := bytes.NewReader(content.VerifierData)
			ip.UnmarshalCBOR(readerIp)
			vd.UnmarshalCBOR(readerVd)

			response.SubPieceInfo.PieceCid = content.PieceCid
			response.SubPieceInfo.Size = content.PieceSize
			response.SubPieceInfo.CommPa = content.CommPa
			response.SubPieceInfo.SizePa = content.SizePa
			response.SubPieceInfo.CommPc = content.CommPc
			response.SubPieceInfo.SizePc = content.SizePc
			response.SubPieceInfo.Status = content.Status
			//response.SubPieceInfo.InclusionProof = *ip
			response.SubPieceInfo.VerifierData = *vd
			response.SubPieceInfo.InclusionProof.ProofIndex.Index = "0x" + hex.EncodeToString(uint64ToBytes(ip.ProofIndex.Index))

			var proofIndexPath []string
			for _, n := range ip.ProofIndex.Path {
				i := [merkletree.NodeSize]byte(n)
				proofIndexPath = append(proofIndexPath, "0x"+hex.EncodeToString(i[:]))
			}
			response.SubPieceInfo.InclusionProof.ProofIndex.Path = proofIndexPath

			response.SubPieceInfo.InclusionProof.ProofSubtree.Index = "0x" + hex.EncodeToString(uint64ToBytes(ip.ProofSubtree.Index))
			var proofSubtreePath []string
			for _, n := range ip.ProofSubtree.Path {
				i := [merkletree.NodeSize]byte(n)
				proofSubtreePath = append(proofSubtreePath, "0x"+hex.EncodeToString(i[:]))
			}
			response.SubPieceInfo.InclusionProof.ProofSubtree.Path = proofSubtreePath

			contentPieceCid, err := cid.Decode(content.PieceCid)
			if err != nil {
				return c.JSON(500, map[string]interface{}{
					"message": "failed to decode piece cid",
				})
			}
			_, err = ip.ComputeExpectedAuxData(datasegment.VerifierDataForPieceInfo(abi.PieceInfo{
				PieceCID: contentPieceCid,
				Size:     abi.PaddedPieceSize(content.PieceSize),
			}))

			if err != nil {
				return c.JSON(500, map[string]interface{}{
					"message": "failed to compute expected aux data",
				})
			}
		}
		if content.BucketUuid != "" {
			var bucket core.Bucket
			node.DB.Model(&core.Bucket{}).Where("uuid = ?", content.BucketUuid).Scan(&bucket)
			job := jobs.CreateNewDispatcher()
			job.AddJob(jobs.NewBucketChecker(node, bucket))
			job.Start(1)
			bucket.RequestingApiKey = ""
		} else {
			job := jobs.CreateNewDispatcher()
			job.AddJob(jobs.NewDealItemChecker(node, content))
			job.Start(1)
		}

		response.DealInfo = DealInfo{
			DealID:    content.DealId,
			Status:    content.Status,
			DeltaNode: content.DeltaNodeUrl,
		}
		return c.JSON(200, map[string]interface{}{
			"message": "success",
			"data":    response,
		})
	})
	e.GET("/status/content/cid/:cid", func(c echo.Context) error {

		var responses []StatusCheckBySubPieceCidResponse

		var contents []core.Content
		node.DB.Model(&core.Content{}).Where("cid = ?", c.Param("cid")).Find(&contents)

		for _, content := range contents {
			var response StatusCheckBySubPieceCidResponse

			response.ContentInfo.Cid = content.Cid
			response.ContentInfo.Name = content.Name
			response.ContentInfo.Size = content.Size
			response.ContentInfo.Miner = content.Miner

			if content.PieceCid != "" {

				// get the inclusion proof to get the aggregate piece cid
				ip := new(datasegment.InclusionProof)
				vd := new(datasegment.InclusionVerifierData)

				readerIp := bytes.NewReader(content.InclusionProof)
				readerVd := bytes.NewReader(content.VerifierData)
				ip.UnmarshalCBOR(readerIp)
				vd.UnmarshalCBOR(readerVd)

				response.SubPieceInfo.PieceCid = content.PieceCid
				response.SubPieceInfo.Size = content.PieceSize
				response.SubPieceInfo.CommPa = content.CommPa
				response.SubPieceInfo.SizePa = content.SizePa
				response.SubPieceInfo.CommPc = content.CommPc
				response.SubPieceInfo.SizePc = content.SizePc
				response.SubPieceInfo.Status = content.Status
				//response.SubPieceInfo.InclusionProof = *ip
				response.SubPieceInfo.VerifierData = *vd

				response.SubPieceInfo.InclusionProof.ProofIndex.Index = "0x" + hex.EncodeToString(uint64ToBytes(ip.ProofIndex.Index))

				var proofIndexPath []string
				for _, n := range ip.ProofIndex.Path {
					i := [merkletree.NodeSize]byte(n)
					proofIndexPath = append(proofIndexPath, "0x"+hex.EncodeToString(i[:]))
				}
				response.SubPieceInfo.InclusionProof.ProofIndex.Path = proofIndexPath

				response.SubPieceInfo.InclusionProof.ProofSubtree.Index = "0x" + hex.EncodeToString(uint64ToBytes(ip.ProofSubtree.Index))
				var proofSubtreePath []string
				for _, n := range ip.ProofSubtree.Path {
					i := [merkletree.NodeSize]byte(n)
					proofSubtreePath = append(proofSubtreePath, "0x"+hex.EncodeToString(i[:]))
				}
				response.SubPieceInfo.InclusionProof.ProofSubtree.Path = proofSubtreePath

				contentPieceCid, err := cid.Decode(content.PieceCid)
				if err != nil {
					return c.JSON(500, map[string]interface{}{
						"message": "failed to decode piece cid",
					})
				}
				_, err = ip.ComputeExpectedAuxData(datasegment.VerifierDataForPieceInfo(abi.PieceInfo{
					PieceCID: contentPieceCid,
					Size:     abi.PaddedPieceSize(content.PieceSize),
				}))

				if err != nil {
					fmt.Println("failed to compute expected aux data", err)
					return c.JSON(500, map[string]interface{}{
						"message": "failed to compute expected aux data",
					})
				}
			}

			if content.BucketUuid != "" {
				var bucket core.Bucket
				node.DB.Model(&core.Bucket{}).Where("uuid = ?", content.BucketUuid).Scan(&bucket)
				job := jobs.CreateNewDispatcher()
				job.AddJob(jobs.NewBucketChecker(node, bucket))
				job.Start(1)
				bucket.RequestingApiKey = ""
			} else {
				job := jobs.CreateNewDispatcher()
				job.AddJob(jobs.NewDealItemChecker(node, content))
				job.Start(1)
			}

			response.DealInfo = DealInfo{
				DealID:    content.DealId,
				Status:    content.Status,
				DeltaNode: content.DeltaNodeUrl,
			}

			responses = append(responses, response)
		}

		return c.JSON(200, map[string]interface{}{
			"data": responses,
		})
	})
	e.GET("/status/content/piece/:piece_cid", func(c echo.Context) error {

		var response StatusCheckBySubPieceCidResponse

		var content core.Content
		node.DB.Model(&core.Content{}).Where("piece_cid = ?", c.Param("piece_cid")).First(&content)

		if content.ID == 0 {
			return c.JSON(500, map[string]interface{}{
				"message": "piece cid is not found",
			})
		}

		response.ContentInfo.Cid = content.Cid
		response.ContentInfo.Name = content.Name
		response.ContentInfo.Size = content.Size
		response.ContentInfo.Miner = content.Miner

		// get the inclusion proof to get the aggregate piece cid
		ip := new(datasegment.InclusionProof)
		vd := new(datasegment.InclusionVerifierData)

		readerIp := bytes.NewReader(content.InclusionProof)
		readerVd := bytes.NewReader(content.VerifierData)
		ip.UnmarshalCBOR(readerIp)
		vd.UnmarshalCBOR(readerVd)

		response.SubPieceInfo.PieceCid = content.PieceCid
		response.SubPieceInfo.Size = content.PieceSize
		response.SubPieceInfo.CommPa = content.CommPa
		response.SubPieceInfo.SizePa = content.SizePa
		response.SubPieceInfo.CommPc = content.CommPc
		response.SubPieceInfo.SizePc = content.SizePc
		response.SubPieceInfo.Status = content.Status
		//response.SubPieceInfo.InclusionProof = *ip
		response.SubPieceInfo.VerifierData = *vd
		response.SubPieceInfo.InclusionProof.ProofIndex.Index = "0x" + hex.EncodeToString(uint64ToBytes(ip.ProofIndex.Index))

		var proofIndexPath []string
		for _, n := range ip.ProofIndex.Path {
			i := [merkletree.NodeSize]byte(n)
			proofIndexPath = append(proofIndexPath, "0x"+hex.EncodeToString(i[:]))
		}
		response.SubPieceInfo.InclusionProof.ProofIndex.Path = proofIndexPath

		response.SubPieceInfo.InclusionProof.ProofSubtree.Index = "0x" + hex.EncodeToString(uint64ToBytes(ip.ProofSubtree.Index))
		var proofSubtreePath []string
		for _, n := range ip.ProofSubtree.Path {
			i := [merkletree.NodeSize]byte(n)
			proofSubtreePath = append(proofSubtreePath, "0x"+hex.EncodeToString(i[:]))
		}
		response.SubPieceInfo.InclusionProof.ProofSubtree.Path = proofSubtreePath

		response.DealInfo = DealInfo{
			DealID:    content.DealId,
			Status:    content.Status,
			DeltaNode: content.DeltaNodeUrl,
		}

		contentPieceCid, err := cid.Decode(content.PieceCid)
		if err != nil {
			return c.JSON(500, map[string]interface{}{
				"message": "failed to decode piece cid",
			})
		}
		_, err = ip.ComputeExpectedAuxData(datasegment.VerifierDataForPieceInfo(abi.PieceInfo{
			PieceCID: contentPieceCid,
			Size:     abi.PaddedPieceSize(content.PieceSize),
		}))
		if err != nil {
			fmt.Println("failed to compute expected aux data", err)
			return c.JSON(500, map[string]interface{}{
				"message": "failed to compute expected aux data",
			})
		}

		var bucket core.Bucket
		node.DB.Model(&core.Bucket{}).Where("uuid = ?", content.BucketUuid).Scan(&bucket)
		job := jobs.CreateNewDispatcher()
		job.AddJob(jobs.NewBucketChecker(node, bucket))
		job.Start(1)

		bucket.RequestingApiKey = ""
		return c.JSON(200, map[string]interface{}{
			"message": "success",
			"data":    response,
		})
	})
	e.GET("/status/content/by-cid/:cid", func(c echo.Context) error {

		var content core.Content
		node.DB.Raw("select * from contents as c where cid = ?", c.Param("cid")).Scan(&content)
		content.RequestingApiKey = ""

		var buckets []core.Bucket
		node.DB.Model(&core.Bucket{}).Where("uuid = ?", content.BucketUuid).Find(&buckets)

		//var bundles []core.Bundle
		//for _, bucket := range buckets {
		//	bucket.RequestingApiKey = ""
		//	node.DB.Model(&core.Bundle{}).Where("	uuid = ?", bucket.BundleUuid).Find(&bundles)
		//	for _, bundle := range bundles {
		//		bundle.RequestingApiKey = ""
		//	}
		//}

		if content.ID == 0 {
			return c.JSON(404, map[string]interface{}{
				"message": "Content not found. Please check if you have the proper API key or if the content id is valid",
			})
		}
		// associated bundle

		// trigger status check
		job := jobs.CreateNewDispatcher()
		job.AddJob(jobs.NewDealItemChecker(node, content))
		job.Start(1)

		return c.JSON(200, map[string]interface{}{
			"content": content,
			"buckets": buckets,
		})
	})

	//e.GET("/status/bundle/:uuid", func(c echo.Context) error {
	//	var bundle core.Bundle
	//	node.DB.Model(&core.Bundle{}).Where("uuid = ?", c.Param("uuid")).Scan(&bundle)
	//	if bundle.ID == 0 {
	//		return c.JSON(404, map[string]interface{}{
	//			"message": "Bundle not found. Please check if you have the proper API key or if the bucket uuid is valid",
	//		})
	//	}
	//
	//	job := jobs.CreateNewDispatcher()
	//	job.AddJob(jobs.NewBundleChecker(node, bundle))
	//	job.Start(1)
	//
	//	bundle.RequestingApiKey = ""
	//	return c.JSON(200, map[string]interface{}{
	//		"bundle": bundle,
	//	})
	//	return nil
	//})
	//
	//e.GET("/status/bundle/buckets/:uuid", func(c echo.Context) error {
	//
	//	var bundle core.Bundle
	//	node.DB.Model(&core.Bundle{}).Where("uuid = ?", c.Param("uuid")).Scan(&bundle)
	//	if bundle.ID == 0 {
	//		return c.JSON(404, map[string]interface{}{
	//			"message": "Bundle not found. Please check if you have the proper API key or if the bucket UUID is valid",
	//		})
	//	}
	//
	//	var buckets []core.Bucket
	//	node.DB.Model(&core.Bucket{}).Where("bundle_uuid = ?", c.Param("uuid")).Scan(&buckets)
	//
	//	// Get the query parameters for page and perPage
	//	pageStr := c.QueryParam("page")
	//	perPageStr := c.QueryParam("per_page")
	//
	//	// Convert the query parameters to integers
	//	page, err := strconv.Atoi(pageStr)
	//	if err != nil || page <= 0 {
	//		// If the page parameter is invalid, set it to the default value of 1
	//		page = 1
	//	}
	//
	//	perPage, err := strconv.Atoi(perPageStr)
	//	if err != nil || perPage <= 0 {
	//		// If the perPage parameter is invalid, set it to the default value of 10
	//		perPage = 10
	//	}
	//
	//	// Retrieve the total count of contents for the bucket
	//	var totalCount int64
	//	node.DB.Model(&core.Content{}).Where("bucket_uuid = ?", c.Param("uuid")).Count(&totalCount)
	//
	//	// Calculate the offset and limit for the current page
	//	offset := (page - 1) * perPage
	//	limit := perPage
	//
	//	// Retrieve the contents with pagination
	//	var contents []core.Content
	//	node.DB.Model(&core.Content{}).Where("bucket_uuid = ?", c.Param("uuid")).Offset(offset).Limit(limit).Scan(&contents)
	//
	//	if len(contents) == 0 {
	//		return c.JSON(404, map[string]interface{}{
	//			"message": "Bucket has no contents",
	//		})
	//	}
	//
	//	var contentResponse []core.Content
	//	for _, content := range contents {
	//		content.RequestingApiKey = ""
	//		contentResponse = append(contentResponse, content)
	//		job := jobs.CreateNewDispatcher()
	//		job.AddJob(jobs.NewDealItemChecker(node, content))
	//	}
	//
	//	job := jobs.CreateNewDispatcher()
	//	job.AddJob(jobs.NewBundleChecker(node, bundle))
	//	job.Start(len(contents) + 1)
	//
	//	bundle.RequestingApiKey = ""
	//	return c.JSON(200, map[string]interface{}{
	//		"bundle":          bundle,
	//		"buckets":         buckets,
	//		"content_entries": contentResponse,
	//		"pagination": map[string]interface{}{
	//			"page":       page,
	//			"perPage":    perPage,
	//			"totalCount": totalCount,
	//		},
	//	})
	//})

	e.GET("/status/bucket/:uuid", func(c echo.Context) error {
		var bucket core.Bucket
		node.DB.Model(&core.Bucket{}).Where("uuid = ?", c.Param("uuid")).Scan(&bucket)
		if bucket.ID == 0 {
			return c.JSON(404, map[string]interface{}{
				"message": "Bucket not found. Please check if you have the proper API key or if the bucket uuid is valid",
			})
		}

		job := jobs.CreateNewDispatcher()
		job.AddJob(jobs.NewBucketChecker(node, bucket))
		job.Start(1)

		bucket.RequestingApiKey = ""
		return c.JSON(200, map[string]interface{}{
			"bucket": bucket,
		})
		return nil
	})
	e.GET("/status/bucket/contents/:uuid", func(c echo.Context) error {
		var bucket core.Bucket
		node.DB.Model(&core.Bucket{}).Where("uuid = ?", c.Param("uuid")).Scan(&bucket)
		if bucket.ID == 0 {
			return c.JSON(404, map[string]interface{}{
				"message": "Bucket not found. Please check if you have the proper API key or if the bucket UUID is valid",
			})
		}

		// Get the query parameters for page and perPage
		pageStr := c.QueryParam("page")
		perPageStr := c.QueryParam("per_page")

		// Convert the query parameters to integers
		page, err := strconv.Atoi(pageStr)
		if err != nil || page <= 0 {
			// If the page parameter is invalid, set it to the default value of 1
			page = 1
		}

		perPage, err := strconv.Atoi(perPageStr)
		if err != nil || perPage <= 0 {
			// If the perPage parameter is invalid, set it to the default value of 10
			perPage = 10
		}

		// Retrieve the total count of contents for the bucket
		var totalCount int64
		node.DB.Model(&core.Content{}).Where("bucket_uuid = ?", c.Param("uuid")).Count(&totalCount)

		// Calculate the offset and limit for the current page
		offset := (page - 1) * perPage
		limit := perPage

		// Retrieve the contents with pagination
		var contents []core.Content
		node.DB.Model(&core.Content{}).Where("bucket_uuid = ?", c.Param("uuid")).Offset(offset).Limit(limit).Scan(&contents)

		if len(contents) == 0 {
			return c.JSON(404, map[string]interface{}{
				"message": "Bucket has no contents",
			})
		}

		var contentResponse []core.Content
		for _, content := range contents {
			content.RequestingApiKey = ""
			contentResponse = append(contentResponse, content)
			job := jobs.CreateNewDispatcher()
			job.AddJob(jobs.NewDealItemChecker(node, content))
		}

		job := jobs.CreateNewDispatcher()
		job.AddJob(jobs.NewBucketChecker(node, bucket))
		job.Start(len(contents) + 1)

		bucket.RequestingApiKey = ""
		return c.JSON(200, map[string]interface{}{
			"bucket":          bucket,
			"content_entries": contentResponse,
			"pagination": map[string]interface{}{
				"page":       page,
				"perPage":    perPage,
				"totalCount": totalCount,
			},
		})
	})
	e.GET("/status/bucket/dag/:uuid", func(c echo.Context) error {

		var bucket core.Bucket
		node.DB.Model(&core.Bucket{}).Where("uuid = ?", c.Param("uuid")).Scan(&bucket)

		if bucket.ID == 0 {
			return c.JSON(404, map[string]interface{}{
				"message": "Bucket not found or has no DAGs. Please check if you have the proper API key or if the bucket uuid is valid",
			})
		}

		// get the cid
		bucketCid, err := cid.Decode(bucket.Cid)
		if err != nil {
			return c.JSON(404, map[string]interface{}{
				"message": "Bucket not found. Please check if you have the proper API key or if the bucket uuid is valid",
			})
		}
		dirNdRaw, err := node.Node.DAGService.Get(context.Background(), bucketCid)
		if err != nil {
			return c.JSON(404, map[string]interface{}{
				"message": "Bucket not found. Please check if you have the proper API key or if the bucket id is valid",
			})
		}

		var contents []core.Content
		node.DB.Model(&core.Content{}).Where("bucket_uuid = ?", c.Param("uuid")).Scan(&contents)

		var contentResponse []core.Content
		for _, content := range contents {
			content.RequestingApiKey = ""

			for _, link := range dirNdRaw.Links() {
				cidFromDb, err := cid.Decode(content.Cid)
				if err != nil {
					continue
				}
				if link.Cid == cidFromDb {
					content.Cid = link.Cid.String()
					content.RequestingApiKey = ""
					contentResponse = append(contentResponse, content)
				}

			}
			job := jobs.CreateNewDispatcher()
			job.AddJob(jobs.NewDealItemChecker(node, content))

		}
		// trigger status check
		job := jobs.CreateNewDispatcher()
		job.AddJob(jobs.NewBucketChecker(node, bucket))
		job.Start(len(contents) + 1)

		bucket.RequestingApiKey = ""
		return c.JSON(200, map[string]interface{}{
			"bucket":        bucket,
			"content_links": contentResponse,
		})
	})

}

func uint64ToBytes(number uint64) []byte {
	byteArray := make([]byte, 8)
	for i := 0; i < 8; i++ {
		byteArray[i] = byte(number >> ((7 - i) * 8))
	}
	return byteArray
}
