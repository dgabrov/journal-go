package controller

import (
	"encoding/json"
	"errors"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/disintegration/imaging"
)

const cookieName = "jou12"

type ErrNoAuth struct{}

func (e *ErrNoAuth) Error() string {
	return "no authentication token found"
}

type Success struct {
	Success string `json:"success"`
}

func NewSuccess() *Success {
	return &Success{Success: "success"}
}

type ErrorResponse struct {
	Message string   `json:"message"`
	Items   []string `json:"items"`
}

func writeJsonResponse(writer http.ResponseWriter, payload any, err error) {
	status := http.StatusOK
	processPayload := payload

	if err != nil {
		processPayload = ErrorResponse{
			Message: err.Error(),
			Items:   make([]string, 0),
		}

		status = http.StatusBadRequest
	}

	header := writer.Header()
	header.Set("Content-Type", "application/json")
	header.Set("Cache-Control", "no-cache")

	writer.WriteHeader(status)

	// write the actual object here
	bts, err := json.Marshal(processPayload)
	if err != nil {
		slog.Error("error marshalling payload")
	}

	_, err = writer.Write(bts)
	if err != nil {
		slog.Error("error writing payload")
	}
}

func parseJsonRequest(r *http.Request, v any) error {
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		return fmt.Errorf("failed to parse request body: %w", err)
	}
	return nil
}

func getToken(r *http.Request) (string, error) {
	// Check Authorization header first
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		lower := strings.ToLower(authHeader)
		if strings.HasPrefix(lower, "bearer ") {
			return authHeader[7:], nil
		}

		return authHeader, nil
	}

	// Check for cookie if Authorization header not found
	cookie, err := r.Cookie(cookieName)
	if err == nil {
		return cookie.Value, nil
	}

	// No token found
	return "", &ErrNoAuth{}
}

func isValidImageFile(file io.ReadSeeker) bool {
	if _, err := file.Seek(0, 0); err != nil {
		return false
	}

	header := make([]byte, 12)
	if _, err := file.Read(header); err != nil {
		return false
	}

	if _, err := file.Seek(0, 0); err != nil {
		return false
	}

	isJPEG := len(header) >= 2 && header[0] == 0xFF && header[1] == 0xD8
	isPNG := len(header) >= 8 && header[0] == 0x89 && header[1] == 0x50 && header[2] == 0x4E && header[3] == 0x47

	return isJPEG || isPNG
}

func detectImageType(header []byte) string {
	isJPEG := len(header) >= 2 && header[0] == 0xFF && header[1] == 0xD8
	isPNG := len(header) >= 8 && header[0] == 0x89 && header[1] == 0x50 && header[2] == 0x4E && header[3] == 0x47

	if isJPEG {
		return "jpeg"
	}
	if isPNG {
		return "png"
	}
	return ""
}

func getImageExtension(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	header := make([]byte, 12)
	if _, err := file.Read(header); err != nil {
		return "", err
	}

	imageType := detectImageType(header)
	if imageType == "" {
		return "", errors.New("unknown image type")
	}

	if imageType == "jpeg" {
		return ".jpg", nil
	}
	return ".png", nil
}

func getImageContentType(file io.ReadSeeker) (string, error) {
	if _, err := file.Seek(0, 0); err != nil {
		return "", err
	}

	header := make([]byte, 12)
	if _, err := file.Read(header); err != nil {
		return "", err
	}

	if _, err := file.Seek(0, 0); err != nil {
		return "", err
	}

	imageType := detectImageType(header)
	if imageType == "" {
		return "", errors.New("unknown image type")
	}

	if imageType == "jpeg" {
		return "image/jpeg", nil
	}
	return "image/png", nil
}

func saveAttachmentFile(file io.ReadSeeker, attachmentID string, regularFolder string) (string, error) {
	contentType, err := getImageContentType(file)
	if err != nil {
		return "", err
	}

	filename := attachmentID + ".dat"
	filePath := filepath.Join(regularFolder, filename)

	dest, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer dest.Close()

	if _, err := io.Copy(dest, file); err != nil {
		os.Remove(filePath)
		return "", err
	}

	return contentType, nil
}

func deleteAttachmentFileFromFolder(folderPath string, filename string, attachmentID string) error {
	filePath := filepath.Join(folderPath, filename)
	if err := os.Remove(filePath); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			slog.Error("failed to delete attachment file", "id", attachmentID, "path", filePath, "error", err)
			return err
		}
	}
	return nil
}

func createThumbnail(id string, regularFolder string, dimension int, smallFolder string) {
	filename := id + ".dat"
	slog.Info("starting processing thumbnail", slog.String("id", id))
	sourcePath := filepath.Join(regularFolder, filename)

	ext, err := getImageExtension(sourcePath)
	if err != nil {
		slog.Error("failed to detect image extension", "id", id, "error", err)
		return
	}

	img, err := imaging.Open(sourcePath)
	if err != nil {
		slog.Error("failed to open image for thumbnail", "id", id, "error", err)
		return
	}

	slog.Info("opened source path for thumbnail processing")

	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	destPath := filepath.Join(smallFolder, filename)
	_ = os.Remove(destPath)

	if width <= dimension && height <= dimension {
		sourceFile, err := os.Open(sourcePath)
		if err != nil {
			slog.Error("failed to open source file for copying", "id", id, "error", err)
			return
		}
		defer sourceFile.Close()

		destFile, err := os.Create(destPath)
		if err != nil {
			slog.Error("failed to create destination file", "id", id, "error", err)
			return
		}
		defer destFile.Close()

		if _, err := io.Copy(destFile, sourceFile); err != nil {
			slog.Error("failed to copy file to small folder", "id", id, "error", err)
			return
		}

		slog.Info("thumbnail copied (no resizing needed)", "id", id)
		return
	}

	var scaledImg image.Image
	if width > height {
		scaledImg = imaging.Resize(img, dimension, 0, imaging.Lanczos)
	} else {
		scaledImg = imaging.Resize(img, 0, dimension, imaging.Lanczos)
	}

	slog.Info("thumbnail resizing complete")

	tempPath := filepath.Join(smallFolder, id+ext)
	if err := imaging.Save(scaledImg, tempPath); err != nil {
		slog.Error("failed to save thumbnail", "id", id, "error", err)
		return
	}

	if err := os.Rename(tempPath, destPath); err != nil {
		slog.Error("failed to rename thumbnail to .dat", "id", id, "error", err)
		os.Remove(tempPath)
		return
	}

	slog.Info("thumbnail created", "id", id)
}
