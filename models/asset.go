package models

import (
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"gitlab.com/slsurvey/slsurvey-srv/helpers"

	"github.com/Sirupsen/logrus"
	"github.com/kennygrant/sanitize"
	uuid "github.com/satori/go.uuid"
)

var (
	//AssetTypeImage used for all type of images
	AssetTypeImage = "image"
	//AssetSystemPathImage defines the storage location for images
	AssetSystemPathImage = "./assets/images"
	//AssetPathImage defines the access location for images
	AssetPathImage = ""
)

var (
	//ErrAssetTypeNotFound used if type is unknown
	ErrAssetTypeNotFound = errors.New("assets_typenotfound")
	//ErrAssetInvalidFileContent used if type is invalid
	ErrAssetInvalidFileContent = errors.New("assets_invalidfilecontent")
	//ErrAssetFileNotSaved used if there was an error while saving the file
	ErrAssetFileNotSaved = errors.New("assets_filenotsaved")
)

//Asset is returned on upload
type Asset struct {
	Path              string                `json:"path"`
	Type              string                `json:"type"`
	FileExtension     string                `json:"fileExtension"`
	FileName          string                `json:"fileName"`
	SystemPath        string                `json:"-"`
	StorageFolderPath string                `json:"-"`
	ContentType       string                `json:"-"`
	FileHeader        *multipart.FileHeader `json:"-"`
	File              multipart.File        `json:"-"`
	IsValid           bool                  `json:"isValid"`
}

func (asset Asset) getLogger() *logrus.Entry {
	return helpers.GetLogger().WithFields(logrus.Fields{
		"model": "Asset",
	})
}

//DetectContentType detecs the file content type
func (asset *Asset) DetectContentType() error {
	buff := make([]byte, 512)
	if _, err := asset.File.Read(buff); err != nil {
		asset.getLogger().WithFields(logrus.Fields{
			"function":   "DetectContentType",
			"fileHeader": asset.FileHeader,
		}).Info(ErrRecordsFetch)
		return err
	}
	asset.ContentType = http.DetectContentType(buff)
	asset.File.Seek(0, 0)
	return nil
}

// image formats and magic numbers
var imageAllowedTypesLookupTable = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/gif":  true,
}

//ValidateContentType checks if the content type is in the lookup table
func (asset *Asset) ValidateContentType() error {
	if err := asset.DetectContentType(); err != nil {
		return err
	}

	if asset.Type == AssetTypeImage {
		if !imageAllowedTypesLookupTable[asset.ContentType] {
			asset.getLogger().WithFields(logrus.Fields{
				"function":            "ValidateContentType",
				"assetType":           asset.Type,
				"detectedContentType": asset.ContentType,
			}).Info(ErrAssetInvalidFileContent)

			asset.IsValid = false
			return ErrAssetInvalidFileContent
		}
		asset.IsValid = true
		return nil
	}
	return ErrAssetInvalidFileContent
}

//Create generate an asset used for uploading the file
func (asset *Asset) Create(assetType string, fileHeader *multipart.FileHeader, file multipart.File) error {
	asset.IsValid = false
	asset.FileHeader = fileHeader
	asset.File = file
	asset.Type = assetType

	if err := asset.ValidateContentType(); err != nil {
		return err
	}

	if assetType == AssetTypeImage {
		asset.CreateImage()
		return nil
	}
	return ErrAssetTypeNotFound
}

//CreateImage generate an asset used for uploading the file
func (asset *Asset) CreateImage() {
	uniqueFileID := uuid.NewV4()
	asset.Type = AssetTypeImage
	asset.FileExtension = filepath.Ext(asset.FileHeader.Filename)
	asset.FileName = sanitize.Path(uniqueFileID.String())
	asset.Path = AssetPathImage + "/" + asset.FileName + asset.FileExtension
	asset.SystemPath = AssetSystemPathImage + "/" + asset.FileName + asset.FileExtension
	asset.StorageFolderPath = AssetSystemPathImage
}

//Upload uploads the file if it was validated
func (asset *Asset) Upload() error {
	os.Mkdir(asset.StorageFolderPath, 0600)

	destinationFile, err := os.OpenFile(asset.SystemPath, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		asset.getLogger().WithFields(logrus.Fields{
			"function":            "Upload",
			"assetType":           asset.Type,
			"detectedContentType": asset.ContentType,
		}).Error(err)
		return ErrAssetFileNotSaved
	}
	defer destinationFile.Close()

	io.Copy(destinationFile, asset.File)
	return nil
}
