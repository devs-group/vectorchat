package api

import (
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	apperrors "github.com/yourusername/vectorchat/internal/errors"
	"github.com/yourusername/vectorchat/internal/middleware"
	"github.com/yourusername/vectorchat/internal/services"
	"github.com/yourusername/vectorchat/pkg/models"
)

type SharedKnowledgeBaseHandler struct {
	AuthMiddleware *middleware.AuthMiddleware
	Service        *services.SharedKnowledgeBaseService
}

func NewSharedKnowledgeBaseHandler(auth *middleware.AuthMiddleware, service *services.SharedKnowledgeBaseService) *SharedKnowledgeBaseHandler {
	return &SharedKnowledgeBaseHandler{
		AuthMiddleware: auth,
		Service:        service,
	}
}

func (h *SharedKnowledgeBaseHandler) RegisterRoutes(app *fiber.App) {
	group := app.Group("/knowledge-bases", h.AuthMiddleware.RequireAuth)

	group.Get("/", h.GET_ListKnowledgeBases)
	group.Post("/", h.POST_CreateKnowledgeBase)
	group.Get("/:id", h.GET_KnowledgeBase)
	group.Put("/:id", h.PUT_UpdateKnowledgeBase)
	group.Delete("/:id", h.DELETE_KnowledgeBase)

	group.Post("/:id/upload", h.POST_UploadFile)
	group.Post("/:id/text", h.POST_UploadText)
	group.Post("/:id/website", h.POST_UploadWebsite)
	group.Get("/:id/files", h.GET_Files)
	group.Delete("/:id/files/:filename", h.DELETE_File)
	group.Get("/:id/text", h.GET_TextSources)
	group.Delete("/:id/text/:sourceId", h.DELETE_TextSource)
}

// @Summary List shared knowledge bases
// @Description Retrieve all shared knowledge bases owned by the authenticated user
// @Tags sharedKnowledgeBase
// @Accept json
// @Produce json
// @Success 200 {object} models.SharedKnowledgeBaseListResponse
// @Failure 401 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Security ApiKeyAuth
// @Router /knowledge-bases [get]
func (h *SharedKnowledgeBaseHandler) GET_ListKnowledgeBases(c *fiber.Ctx) error {
	user, err := GetUser(c)
	if err != nil {
		return err
	}

	resp, err := h.Service.List(c.Context(), user.ID)
	if err != nil {
		return ErrorResponse(c, "Failed to list knowledge bases", err)
	}

	return c.JSON(resp)
}

// @Summary Create shared knowledge base
// @Description Create a new shared knowledge base for the authenticated user
// @Tags sharedKnowledgeBase
// @Accept json
// @Produce json
// @Param body body models.SharedKnowledgeBaseCreateRequest true "Knowledge base payload"
// @Success 201 {object} models.SharedKnowledgeBaseResponse
// @Failure 400 {object} models.APIResponse
// @Failure 401 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Security ApiKeyAuth
// @Router /knowledge-bases [post]
func (h *SharedKnowledgeBaseHandler) POST_CreateKnowledgeBase(c *fiber.Ctx) error {
	user, err := GetUser(c)
	if err != nil {
		return err
	}

	var req models.SharedKnowledgeBaseCreateRequest
	if err := c.BodyParser(&req); err != nil {
		return ErrorResponse(c, "Invalid request body", err, http.StatusBadRequest)
	}

	resp, err := h.Service.Create(c.Context(), user.ID, &req)
	if err != nil {
		status := http.StatusInternalServerError
		if apperrors.Is(err, apperrors.ErrInvalidChatbotParameters) {
			status = http.StatusBadRequest
		}
		return ErrorResponse(c, "Failed to create knowledge base", err, status)
	}

	return c.Status(http.StatusCreated).JSON(resp)
}

// @Summary Get shared knowledge base
// @Description Fetch a shared knowledge base by ID for the authenticated user
// @Tags sharedKnowledgeBase
// @Accept json
// @Produce json
// @Param id path string true "Knowledge base ID (UUID)"
// @Success 200 {object} models.SharedKnowledgeBaseResponse
// @Failure 400 {object} models.APIResponse
// @Failure 401 {object} models.APIResponse
// @Failure 403 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Security ApiKeyAuth
// @Router /knowledge-bases/{id} [get]
func (h *SharedKnowledgeBaseHandler) GET_KnowledgeBase(c *fiber.Ctx) error {
	user, err := GetUser(c)
	if err != nil {
		return err
	}

	kbID, err := parseUUIDParam(c, "id")
	if err != nil {
		return ErrorResponse(c, "Invalid knowledge base id", err, http.StatusBadRequest)
	}

	resp, err := h.Service.Get(c.Context(), user.ID, kbID)
	if err != nil {
		status := http.StatusInternalServerError
		if apperrors.Is(err, apperrors.ErrSharedKnowledgeBaseNotFound) {
			status = http.StatusNotFound
		} else if apperrors.Is(err, apperrors.ErrUnauthorizedKnowledgeBaseAccess) {
			status = http.StatusForbidden
		}
		return ErrorResponse(c, "Failed to fetch knowledge base", err, status)
	}

	return c.JSON(resp)
}

// @Summary Update shared knowledge base
// @Description Update the details of a shared knowledge base owned by the authenticated user
// @Tags sharedKnowledgeBase
// @Accept json
// @Produce json
// @Param id path string true "Knowledge base ID (UUID)"
// @Param body body models.SharedKnowledgeBaseUpdateRequest true "Knowledge base updates"
// @Success 200 {object} models.SharedKnowledgeBaseResponse
// @Failure 400 {object} models.APIResponse
// @Failure 401 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Security ApiKeyAuth
// @Router /knowledge-bases/{id} [put]
func (h *SharedKnowledgeBaseHandler) PUT_UpdateKnowledgeBase(c *fiber.Ctx) error {
	user, err := GetUser(c)
	if err != nil {
		return err
	}

	kbID, err := parseUUIDParam(c, "id")
	if err != nil {
		return ErrorResponse(c, "Invalid knowledge base id", err, http.StatusBadRequest)
	}

	var req models.SharedKnowledgeBaseUpdateRequest
	if err := c.BodyParser(&req); err != nil {
		return ErrorResponse(c, "Invalid request body", err, http.StatusBadRequest)
	}

	resp, err := h.Service.Update(c.Context(), user.ID, kbID, &req)
	if err != nil {
		status := http.StatusInternalServerError
		if apperrors.Is(err, apperrors.ErrSharedKnowledgeBaseNotFound) {
			status = http.StatusNotFound
		} else if apperrors.Is(err, apperrors.ErrUnauthorizedKnowledgeBaseAccess) || apperrors.Is(err, apperrors.ErrInvalidChatbotParameters) {
			status = http.StatusBadRequest
		}
		return ErrorResponse(c, "Failed to update knowledge base", err, status)
	}

	return c.JSON(resp)
}

// @Summary Delete shared knowledge base
// @Description Delete a shared knowledge base owned by the authenticated user
// @Tags sharedKnowledgeBase
// @Accept json
// @Produce json
// @Param id path string true "Knowledge base ID (UUID)"
// @Success 204 {string} string ""
// @Failure 400 {object} models.APIResponse
// @Failure 401 {object} models.APIResponse
// @Failure 403 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Security ApiKeyAuth
// @Router /knowledge-bases/{id} [delete]
func (h *SharedKnowledgeBaseHandler) DELETE_KnowledgeBase(c *fiber.Ctx) error {
	user, err := GetUser(c)
	if err != nil {
		return err
	}

	kbID, err := parseUUIDParam(c, "id")
	if err != nil {
		return ErrorResponse(c, "Invalid knowledge base id", err, http.StatusBadRequest)
	}

	if err := h.Service.Delete(c.Context(), user.ID, kbID); err != nil {
		status := http.StatusInternalServerError
		if apperrors.Is(err, apperrors.ErrSharedKnowledgeBaseNotFound) {
			status = http.StatusNotFound
		} else if apperrors.Is(err, apperrors.ErrUnauthorizedKnowledgeBaseAccess) {
			status = http.StatusForbidden
		}
		return ErrorResponse(c, "Failed to delete knowledge base", err, status)
	}

	return c.SendStatus(http.StatusNoContent)
}

// @Summary Upload file to shared knowledge base
// @Description Upload a file to be processed into the shared knowledge base
// @Tags sharedKnowledgeBase
// @Accept multipart/form-data
// @Produce json
// @Param id path string true "Knowledge base ID (UUID)"
// @Param file formData file true "File to upload"
// @Success 200 {object} models.SharedKnowledgeBaseFileUploadResponse
// @Failure 400 {object} models.APIResponse
// @Failure 401 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Security ApiKeyAuth
// @Router /knowledge-bases/{id}/upload [post]
func (h *SharedKnowledgeBaseHandler) POST_UploadFile(c *fiber.Ctx) error {
	user, err := GetUser(c)
	if err != nil {
		return err
	}

	kbID, err := parseUUIDParam(c, "id")
	if err != nil {
		return ErrorResponse(c, "Invalid knowledge base id", err, http.StatusBadRequest)
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		return ErrorResponse(c, "No file uploaded", err, http.StatusBadRequest)
	}

	resp, err := h.Service.ProcessFileUpload(c.Context(), user.ID, kbID, fileHeader)
	if err != nil {
		status := http.StatusInternalServerError
		if apperrors.Is(err, apperrors.ErrInvalidChatbotParameters) {
			status = http.StatusBadRequest
		}
		return ErrorResponse(c, "Failed to upload file", err, status)
	}

	return c.JSON(resp)
}

// @Summary Upload text to shared knowledge base
// @Description Upload plain text content to be indexed for the shared knowledge base
// @Tags sharedKnowledgeBase
// @Accept json
// @Produce json
// @Param id path string true "Knowledge base ID (UUID)"
// @Param body body models.TextUploadRequest true "Text payload"
// @Success 200 {object} models.MessageResponse
// @Failure 400 {object} models.APIResponse
// @Failure 401 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Security ApiKeyAuth
// @Router /knowledge-bases/{id}/text [post]
func (h *SharedKnowledgeBaseHandler) POST_UploadText(c *fiber.Ctx) error {
	user, err := GetUser(c)
	if err != nil {
		return err
	}

	kbID, err := parseUUIDParam(c, "id")
	if err != nil {
		return ErrorResponse(c, "Invalid knowledge base id", err, http.StatusBadRequest)
	}

	var req models.TextUploadRequest
	if err := c.BodyParser(&req); err != nil {
		return ErrorResponse(c, "Invalid request body", err, http.StatusBadRequest)
	}

	if err := h.Service.ProcessTextUpload(c.Context(), user.ID, kbID, req.Text); err != nil {
		status := http.StatusInternalServerError
		if apperrors.Is(err, apperrors.ErrInvalidChatbotParameters) {
			status = http.StatusBadRequest
		}
		return ErrorResponse(c, "Failed to add text", err, status)
	}

	return c.JSON(models.MessageResponse{Message: "Text processed successfully"})
}

// @Summary Upload website to shared knowledge base
// @Description Crawl and index a website into the shared knowledge base starting from the provided URL
// @Tags sharedKnowledgeBase
// @Accept json
// @Produce json
// @Param id path string true "Knowledge base ID (UUID)"
// @Param body body models.WebsiteUploadRequest true "Website payload"
// @Success 200 {object} models.MessageResponse
// @Failure 400 {object} models.APIResponse
// @Failure 401 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Security ApiKeyAuth
// @Router /knowledge-bases/{id}/website [post]
func (h *SharedKnowledgeBaseHandler) POST_UploadWebsite(c *fiber.Ctx) error {
	user, err := GetUser(c)
	if err != nil {
		return err
	}

	kbID, err := parseUUIDParam(c, "id")
	if err != nil {
		return ErrorResponse(c, "Invalid knowledge base id", err, http.StatusBadRequest)
	}

	var req models.WebsiteUploadRequest
	if err := c.BodyParser(&req); err != nil || strings.TrimSpace(req.URL) == "" {
		return ErrorResponse(c, "Invalid request body", err, http.StatusBadRequest)
	}

	if err := h.Service.ProcessWebsiteUpload(c.Context(), user.ID, kbID, req.URL); err != nil {
		status := http.StatusInternalServerError
		if apperrors.Is(err, apperrors.ErrInvalidChatbotParameters) {
			status = http.StatusBadRequest
		}
		return ErrorResponse(c, "Failed to index website", err, status)
	}

	return c.JSON(models.MessageResponse{Message: "Website indexed successfully"})
}

// @Summary List files in shared knowledge base
// @Description List non-text files associated with the shared knowledge base
// @Tags sharedKnowledgeBase
// @Accept json
// @Produce json
// @Param id path string true "Knowledge base ID (UUID)"
// @Success 200 {object} models.SharedKnowledgeBaseFilesResponse
// @Failure 400 {object} models.APIResponse
// @Failure 401 {object} models.APIResponse
// @Failure 403 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Security ApiKeyAuth
// @Router /knowledge-bases/{id}/files [get]
func (h *SharedKnowledgeBaseHandler) GET_Files(c *fiber.Ctx) error {
	user, err := GetUser(c)
	if err != nil {
		return err
	}
	kbID, err := parseUUIDParam(c, "id")
	if err != nil {
		return ErrorResponse(c, "Invalid knowledge base id", err, http.StatusBadRequest)
	}

	resp, err := h.Service.ListFiles(c.Context(), user.ID, kbID)
	if err != nil {
		status := http.StatusInternalServerError
		if apperrors.Is(err, apperrors.ErrSharedKnowledgeBaseNotFound) {
			status = http.StatusNotFound
		} else if apperrors.Is(err, apperrors.ErrUnauthorizedKnowledgeBaseAccess) {
			status = http.StatusForbidden
		}
		return ErrorResponse(c, "Failed to list files", err, status)
	}

	return c.JSON(resp)
}

// @Summary Delete file from shared knowledge base
// @Description Delete a specific file associated with the shared knowledge base
// @Tags sharedKnowledgeBase
// @Accept json
// @Produce json
// @Param id path string true "Knowledge base ID (UUID)"
// @Param filename path string true "Filename"
// @Success 200 {object} models.MessageResponse
// @Failure 400 {object} models.APIResponse
// @Failure 401 {object} models.APIResponse
// @Failure 403 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Security ApiKeyAuth
// @Router /knowledge-bases/{id}/files/{filename} [delete]
func (h *SharedKnowledgeBaseHandler) DELETE_File(c *fiber.Ctx) error {
	user, err := GetUser(c)
	if err != nil {
		return err
	}
	kbID, err := parseUUIDParam(c, "id")
	if err != nil {
		return ErrorResponse(c, "Invalid knowledge base id", err, http.StatusBadRequest)
	}
	filename := c.Params("filename")

	if err := h.Service.DeleteFile(c.Context(), user.ID, kbID, filename); err != nil {
		status := http.StatusInternalServerError
		if apperrors.Is(err, apperrors.ErrInvalidChatbotParameters) {
			status = http.StatusBadRequest
		} else if apperrors.Is(err, apperrors.ErrSharedKnowledgeBaseNotFound) {
			status = http.StatusNotFound
		} else if apperrors.Is(err, apperrors.ErrUnauthorizedKnowledgeBaseAccess) {
			status = http.StatusForbidden
		}
		return ErrorResponse(c, "Failed to delete file", err, status)
	}

	return c.JSON(models.MessageResponse{Message: "File deleted successfully"})
}

// @Summary List text sources in shared knowledge base
// @Description List indexed text sources associated with the shared knowledge base
// @Tags sharedKnowledgeBase
// @Accept json
// @Produce json
// @Param id path string true "Knowledge base ID (UUID)"
// @Success 200 {object} models.SharedKnowledgeBaseTextSourcesResponse
// @Failure 400 {object} models.APIResponse
// @Failure 401 {object} models.APIResponse
// @Failure 403 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Security ApiKeyAuth
// @Router /knowledge-bases/{id}/text [get]
func (h *SharedKnowledgeBaseHandler) GET_TextSources(c *fiber.Ctx) error {
	user, err := GetUser(c)
	if err != nil {
		return err
	}
	kbID, err := parseUUIDParam(c, "id")
	if err != nil {
		return ErrorResponse(c, "Invalid knowledge base id", err, http.StatusBadRequest)
	}

	resp, err := h.Service.ListTextSources(c.Context(), user.ID, kbID)
	if err != nil {
		status := http.StatusInternalServerError
		if apperrors.Is(err, apperrors.ErrSharedKnowledgeBaseNotFound) {
			status = http.StatusNotFound
		} else if apperrors.Is(err, apperrors.ErrUnauthorizedKnowledgeBaseAccess) {
			status = http.StatusForbidden
		}
		return ErrorResponse(c, "Failed to list text sources", err, status)
	}

	return c.JSON(resp)
}

// @Summary Delete text source from shared knowledge base
// @Description Delete a text source previously indexed for the shared knowledge base
// @Tags sharedKnowledgeBase
// @Accept json
// @Produce json
// @Param id path string true "Knowledge base ID (UUID)"
// @Param sourceId path string true "Text source ID (UUID)"
// @Success 200 {object} models.MessageResponse
// @Failure 400 {object} models.APIResponse
// @Failure 401 {object} models.APIResponse
// @Failure 403 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Security ApiKeyAuth
// @Router /knowledge-bases/{id}/text/{sourceId} [delete]
func (h *SharedKnowledgeBaseHandler) DELETE_TextSource(c *fiber.Ctx) error {
	user, err := GetUser(c)
	if err != nil {
		return err
	}
	kbID, err := parseUUIDParam(c, "id")
	if err != nil {
		return ErrorResponse(c, "Invalid knowledge base id", err, http.StatusBadRequest)
	}
	sourceID := c.Params("sourceId")

	if err := h.Service.DeleteTextSource(c.Context(), user.ID, kbID, sourceID); err != nil {
		status := http.StatusInternalServerError
		if apperrors.Is(err, apperrors.ErrInvalidChatbotParameters) {
			status = http.StatusBadRequest
		} else if apperrors.Is(err, apperrors.ErrSharedKnowledgeBaseNotFound) {
			status = http.StatusNotFound
		} else if apperrors.Is(err, apperrors.ErrUnauthorizedKnowledgeBaseAccess) {
			status = http.StatusForbidden
		}
		return ErrorResponse(c, "Failed to delete text source", err, status)
	}

	return c.JSON(models.MessageResponse{Message: "Text source deleted successfully"})
}

func parseUUIDParam(c *fiber.Ctx, key string) (uuid.UUID, error) {
	value := c.Params(key)
	if strings.TrimSpace(value) == "" {
		return uuid.Nil, apperrors.Wrap(apperrors.ErrInvalidChatbotParameters, "missing id")
	}
	return uuid.Parse(value)
}
