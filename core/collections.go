package core

//
//import (
//	"errors"
//	"fmt"
//	"net/http"
//	"path/filepath"
//	"strings"
//	"time"
//
//	"golang.org/x/xerrors"
//	"gorm.io/gorm"
//)
//
//type CidType string
//
//type CollectionListResponse struct {
//	Name      string    `json:"name"`
//	Type      CidType   `json:"type"`
//	Size      int64     `json:"size"`
//	ContID    uint64    `json:"contId"`
//	Cid       *DbCID    `json:"cid,omitempty"`
//	Dir       string    `json:"dir"`
//	ColUuid   string    `json:"coluuid"`
//	UpdatedAt time.Time `json:"updatedAt"`
//}
//
//const (
//	CidTypeDir  CidType = "directory"
//	CidTypeFile CidType = "file"
//)
//
//type CollectionService struct {
//	node LightNode
//}
//
//func NewCollectionService(node LightNode) *CollectionService {
//	return &CollectionService{
//		node: node,
//	}
//}
//
//func (c CollectionService) GetCollection(coluuid string, requestingApiKey string) (Collection, error) {
//	var collection Collection
//	err := c.node.DB.Model(Collection{}).Where("uuid = ? and requesting_api_key = ?", coluuid, requestingApiKey).First(&collection).Error
//	if err != nil {
//		return Collection{}, errors.New("collection not found")
//	}
//	return collection, nil
//}
//
//// refs = collections.GetContentsInPath(path, s.DB, u)
//func (c CollectionService) GetContentsInPath(coluuid string, path string, requestingApiKey string) ([]CollectionRef, error) {
//	col, err := c.GetCollection(coluuid, requestingApiKey)
//	if err != nil {
//		return []CollectionRef{}, err
//	}
//
//	refs := []CollectionRef{}
//	if err := c.node.DB.Model(CollectionRef{}).
//		Where("collection = ?", col.ID).
//		Scan(&refs).Error; err != nil {
//		return []CollectionRef{}, err
//	}
//
//	if len(refs) == 0 {
//		return []CollectionRef{}, fmt.Errorf("no contents on specified path for collection")
//	}
//
//	var selectedRefs []CollectionRef
//	for _, ref := range refs {
//		if strings.HasPrefix(*ref.Path, path) {
//			selectedRefs = append(selectedRefs, ref)
//		}
//	}
//
//	return selectedRefs, nil
//}
//
//func (c CollectionService) Contains(collection *Collection, fullPath string, db *gorm.DB) bool {
//	var colRef CollectionRef
//	err := db.First(&colRef, "collection = ? and path = ?", collection.ID, fullPath).Error
//	return !errors.Is(err, gorm.ErrRecordNotFound)
//}
//
//func (c CollectionService) AddContentToCollection(coluuid string, contentID string, dir string, overwrite bool, requestingApiKey string) error {
//	// first we get the collection and content
//	col, err := c.GetCollection(coluuid, requestingApiKey)
//	if err != nil {
//		return err
//	}
//	content, err := util.GetContent(contentID, db, u)
//	if err != nil {
//		return err
//	}
//
//	path, err := c.ConstructDirectoryPath(dir)
//	if err != nil {
//		return err
//	}
//	fullPath := filepath.Join(path, content.Name)
//
//	// see if there's already a file with that name/path on that collection
//	pathInCollection := c.Contains(&col, fullPath, db)
//	if pathInCollection && !overwrite {
//		return xerrors.Errorf("file already exists in collection, specify 'overwrite=true' to overwrite")
//	}
//
//	// if there's a duplicate and overwrite has been set to true, then update
//	if pathInCollection && overwrite {
//		if err := c.node.DB.Model(CollectionRef{}).Where("collection = ? and path = ?", col.ID, fullPath).UpdateColumn("content", content.ID).Error; err != nil {
//			return xerrors.Errorf("unable to overwrite file: %w", err)
//		}
//	} else { // else, create collectionRef for new file
//		if err := c.node.DB.Create(&CollectionRef{
//			Collection: col.ID,
//			Content:    content.ID,
//			Path:       &fullPath,
//		}).Error; err != nil {
//			return err
//		}
//	}
//	return nil
//}
//
//func (c CollectionService) GetDirectoryContents(refs []util.ContentWithPath, queryDir, coluuid string) ([]*CollectionListResponse, error) {
//	dirs := make(map[string]bool)
//	var result []*CollectionListResponse
//	for _, r := range refs {
//		directoryContent, err := getDirectoryContent(r, queryDir, coluuid)
//
//		if err != nil {
//			return nil, err
//		}
//
//		if directoryContent != nil { // if there was content
//			if directoryContent.Type == CidTypeDir { // if the content was a directory
//				subDir := directoryContent.Dir
//				if dirs[subDir] { // if the directory had already been added to response, continue
//					continue
//				}
//				dirs[subDir] = true
//			}
//			result = append(result, directoryContent)
//		}
//	}
//	return result, nil
//}
//
//func (c CollectionService) getDirectoryContent(r util.ContentWithPath, queryDir, coluuid string) (*CollectionListResponse, error) {
//	if r.Path == "" || r.Name == "" {
//		return nil, nil
//	}
//
//	if !strings.HasPrefix(r.Path, queryDir) {
//		return nil, nil
//	}
//
//	relp, err := getRelativePath(r, queryDir)
//	if err != nil {
//		return nil, &util.HttpError{
//			Code:    http.StatusInternalServerError,
//			Reason:  util.ERR_INTERNAL_SERVER,
//			Details: fmt.Sprintf("errored while calculating relative contentPath queryDir=%s, contentPath=%s", queryDir, r.Path),
//		}
//	}
//
//	// Query directory has a subdirectory, which contains the actual content.
//	// if relative contentPath has a /, the file is in a subdirectory
//	// print the directory the file is in if we haven't already
//	if strings.Contains(relp, "/") {
//		parts := strings.Split(relp, "/")
//		subDir := parts[0]
//		return &CollectionListResponse{
//			Name:      subDir,
//			Type:      CidTypeDir,
//			Dir:       queryDir,
//			ColUuid:   coluuid,
//			UpdatedAt: r.UpdatedAt,
//		}, nil
//	}
//
//	// trying to list a CID queryDir, not allowed
//	if r.Type == util.Directory {
//		return nil, &util.HttpError{
//			Code:    http.StatusBadRequest,
//			Reason:  util.ERR_BAD_REQUEST,
//			Details: fmt.Sprintf("listing CID directories is not allowed"),
//		}
//	}
//	return &CollectionListResponse{
//		Name:      r.Name,
//		Type:      CidTypeFile,
//		Size:      r.Size,
//		ContID:    r.ID,
//		Cid:       &util.DbCID{CID: r.Cid.CID},
//		Dir:       queryDir,
//		ColUuid:   coluuid,
//		UpdatedAt: r.UpdatedAt,
//	}, nil
//}
//
//func (c CollectionService) getRelativePath(r util.ContentWithPath, queryDir string) (string, error) {
//	contentPath := r.Path
//	relp, err := filepath.Rel(queryDir, contentPath)
//	return relp, err
//}
//
//func (c CollectionService) ConstructDirectoryPath(dir string) (string, error) {
//	defaultPath := "/"
//	path := defaultPath
//	if cp := dir; cp != "" {
//		sp, err := sanitizePath(cp)
//		if err != nil {
//			return "", err
//		}
//
//		path = sp
//	}
//	return path, nil
//}
//
//func sanitizePath(p string) (string, error) {
//	if len(p) == 0 {
//		return "", fmt.Errorf("can't sanitize empty path")
//	}
//
//	if p[0] != '/' {
//		return "", fmt.Errorf("paths must start with /")
//	}
//
//	// TODO: prevent use of special weird characters
//
//	cleanPath := filepath.Clean(p)
//
//	// if original path ends in /, append / to cleaned path
//	// needed for full path vs dir+filename magic to work in handleAddIpfs
//	if strings.HasSuffix(p, "/") {
//		cleanPath = cleanPath + "/"
//	}
//	return cleanPath, nil
//}
