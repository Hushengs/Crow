package server

import (
	"context"
	"fmt"
	"io"
	gohttp "net/http"
	"os"
	"path/filepath"
	"strings"

	kratoserrors "github.com/go-kratos/kratos/v3/errors"
	khttp "github.com/go-kratos/kratos/v3/transport/http"
	"github.com/google/uuid"
)

const posterUploadMaxBytes int64 = 10 << 20

type posterUploadResponse struct {
	URL string `json:"url"`
}

func registerPosterRoutes(srv *khttp.Server) {
	route := srv.Route("/")
	route.Handle("POST", "/v1/videos/posters/upload", uploadPosterHTTPHandler())
	srv.HandlePrefix("/uploads/", gohttp.StripPrefix("/uploads/", gohttp.FileServer(gohttp.Dir("uploads"))))
}

func uploadPosterHTTPHandler() func(ctx khttp.Context) error {
	return func(ctx khttp.Context) error {
		khttp.SetOperation(ctx, "/crow.vod.UploadPoster")
		h := ctx.Middleware(func(inner context.Context, _ any) (any, error) {
			return savePosterUpload(ctx)
		})
		out, err := h(ctx, nil)
		if err != nil {
			return err
		}
		return ctx.JSON(gohttp.StatusOK, out)
	}
}

func savePosterUpload(ctx khttp.Context) (*posterUploadResponse, error) {
	req := ctx.Request()
	req.Body = gohttp.MaxBytesReader(ctx.Response(), req.Body, posterUploadMaxBytes)
	if err := req.ParseMultipartForm(posterUploadMaxBytes); err != nil {
		return nil, kratoserrors.BadRequest("VOD_INVALID_ARGUMENT", "poster upload request is invalid")
	}
	if req.MultipartForm != nil {
		defer req.MultipartForm.RemoveAll()
	}

	file, header, err := req.FormFile("file")
	if err != nil {
		return nil, kratoserrors.BadRequest("VOD_INVALID_ARGUMENT", "poster image is required")
	}
	defer file.Close()

	sniff := make([]byte, 512)
	n, err := file.Read(sniff)
	if err != nil && err != io.EOF {
		return nil, kratoserrors.InternalServer("VOD_UPLOAD_READ_FAILED", "failed to read poster image")
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return nil, kratoserrors.InternalServer("VOD_UPLOAD_READ_FAILED", "failed to read poster image")
	}

	ext, err := posterExtension(header.Header.Get("Content-Type"), sniff[:n], header.Filename)
	if err != nil {
		return nil, err
	}

	dir := filepath.Join("uploads", "posters")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, kratoserrors.InternalServer("VOD_UPLOAD_SAVE_FAILED", "failed to prepare poster storage")
	}

	filename := uuid.NewString() + ext
	dstPath := filepath.Join(dir, filename)
	dst, err := os.OpenFile(dstPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return nil, kratoserrors.InternalServer("VOD_UPLOAD_SAVE_FAILED", "failed to save poster image")
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		return nil, kratoserrors.InternalServer("VOD_UPLOAD_SAVE_FAILED", "failed to save poster image")
	}

	return &posterUploadResponse{URL: "/uploads/posters/" + filename}, nil
}

func posterExtension(contentType string, sniff []byte, originalName string) (string, error) {
	normalized := strings.ToLower(strings.TrimSpace(contentType))
	if normalized == "" {
		normalized = strings.ToLower(gohttp.DetectContentType(sniff))
	}
	switch normalized {
	case "image/jpeg", "image/jpg":
		return ".jpg", nil
	case "image/png":
		return ".png", nil
	case "image/webp":
		return ".webp", nil
	case "image/gif":
		return ".gif", nil
	}

	switch strings.ToLower(filepath.Ext(originalName)) {
	case ".jpg", ".jpeg":
		return ".jpg", nil
	case ".png":
		return ".png", nil
	case ".webp":
		return ".webp", nil
	case ".gif":
		return ".gif", nil
	default:
		return "", kratoserrors.BadRequest("VOD_INVALID_ARGUMENT", fmt.Sprintf("unsupported poster image type %q", contentType))
	}
}
