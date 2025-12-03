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

type OrganizationHandler struct {
	auth          *middleware.AuthMiddleware
	orgService    *services.OrganizationService
	orgMiddleware *middleware.OrganizationMiddleware
}

func NewOrganizationHandler(auth *middleware.AuthMiddleware, orgService *services.OrganizationService, orgMiddleware *middleware.OrganizationMiddleware) *OrganizationHandler {
	return &OrganizationHandler{
		auth:          auth,
		orgService:    orgService,
		orgMiddleware: orgMiddleware,
	}
}

func (h *OrganizationHandler) RegisterRoutes(app *fiber.App) {
	group := app.Group("/orgs", h.auth.RequireAuth)

	group.Get("/", h.listOrganizations)
	group.Post("/", h.createOrganization)
	group.Get("/current", h.orgMiddleware.Attach, h.getCurrentContext)
	group.Get("/:id", h.getOrganization)
	group.Patch("/:id", h.updateOrganization)
	group.Delete("/:id", h.deleteOrganization)

	group.Get("/:id/members", h.listMembers)
	group.Patch("/:id/members/:userID", h.updateMemberRole)
	group.Delete("/:id/members/:userID", h.removeMember)

	group.Post("/:id/invites", h.createInvite)
	group.Get("/:id/invites", h.listInvites)

	app.Post("/org-invites/accept", h.auth.RequireAuth, h.acceptInvite)
}

func (h *OrganizationHandler) listOrganizations(c *fiber.Ctx) error {
	user, err := GetUser(c)
	if err != nil {
		return err
	}

	resp, err := h.orgService.ListForUser(c.Context(), user.ID)
	if err != nil {
		return ErrorResponse(c, "Failed to list organizations", err)
	}
	return c.JSON(resp)
}

func (h *OrganizationHandler) createOrganization(c *fiber.Ctx) error {
	user, err := GetUser(c)
	if err != nil {
		return err
	}

	var req models.OrganizationCreateRequest
	if err := c.BodyParser(&req); err != nil {
		return ErrorResponse(c, "Invalid request body", err, http.StatusBadRequest)
	}

	resp, err := h.orgService.Create(c.Context(), user.ID, &req)
	if err != nil {
		status := http.StatusInternalServerError
		if apperrors.Is(err, apperrors.ErrInvalidUserData) || apperrors.Is(err, apperrors.ErrOrganizationAlreadyExists) {
			status = http.StatusBadRequest
		}
		return ErrorResponse(c, "Failed to create organization", err, status)
	}
	return c.Status(http.StatusCreated).JSON(resp)
}

func (h *OrganizationHandler) getOrganization(c *fiber.Ctx) error {
	user, err := GetUser(c)
	if err != nil {
		return err
	}
	orgID, parseErr := parseUUIDParam(c, "id")
	if parseErr != nil {
		return ErrorResponse(c, "Invalid organization id", parseErr, http.StatusBadRequest)
	}
	resp, err := h.orgService.Get(c.Context(), orgID, user.ID)
	if err != nil {
		status := http.StatusInternalServerError
		if apperrors.Is(err, apperrors.ErrOrganizationNotFound) {
			status = http.StatusNotFound
		} else if apperrors.Is(err, apperrors.ErrUnauthorizedOrganizationAccess) {
			status = http.StatusForbidden
		}
		return ErrorResponse(c, "Failed to fetch organization", err, status)
	}
	return c.JSON(resp)
}

func (h *OrganizationHandler) updateOrganization(c *fiber.Ctx) error {
	user, err := GetUser(c)
	if err != nil {
		return err
	}
	orgID, parseErr := parseUUIDParam(c, "id")
	if parseErr != nil {
		return ErrorResponse(c, "Invalid organization id", parseErr, http.StatusBadRequest)
	}

	var req models.OrganizationUpdateRequest
	if err := c.BodyParser(&req); err != nil {
		return ErrorResponse(c, "Invalid request body", err, http.StatusBadRequest)
	}

	resp, err := h.orgService.Update(c.Context(), orgID, user.ID, &req)
	if err != nil {
		status := http.StatusInternalServerError
		if apperrors.Is(err, apperrors.ErrOrganizationNotFound) {
			status = http.StatusNotFound
		} else if apperrors.Is(err, apperrors.ErrUnauthorizedOrganizationAccess) || apperrors.Is(err, apperrors.ErrInvalidUserData) {
			status = http.StatusBadRequest
		}
		return ErrorResponse(c, "Failed to update organization", err, status)
	}
	return c.JSON(resp)
}

func (h *OrganizationHandler) deleteOrganization(c *fiber.Ctx) error {
	user, err := GetUser(c)
	if err != nil {
		return err
	}
	orgID, parseErr := parseUUIDParam(c, "id")
	if parseErr != nil {
		return ErrorResponse(c, "Invalid organization id", parseErr, http.StatusBadRequest)
	}

	if err := h.orgService.Delete(c.Context(), orgID, user.ID); err != nil {
		status := http.StatusInternalServerError
		if apperrors.Is(err, apperrors.ErrOrganizationNotFound) {
			status = http.StatusNotFound
		} else if apperrors.Is(err, apperrors.ErrUnauthorizedOrganizationAccess) {
			status = http.StatusForbidden
		}
		return ErrorResponse(c, "Failed to delete organization", err, status)
	}
	return c.SendStatus(http.StatusNoContent)
}

func (h *OrganizationHandler) listMembers(c *fiber.Ctx) error {
	user, err := GetUser(c)
	if err != nil {
		return err
	}
	orgID, parseErr := parseUUIDParam(c, "id")
	if parseErr != nil {
		return ErrorResponse(c, "Invalid organization id", parseErr, http.StatusBadRequest)
	}

	members, err := h.orgService.ListMembers(c.Context(), orgID, user.ID)
	if err != nil {
		status := http.StatusInternalServerError
		if apperrors.Is(err, apperrors.ErrUnauthorizedOrganizationAccess) {
			status = http.StatusForbidden
		}
		return ErrorResponse(c, "Failed to list members", err, status)
	}
	return c.JSON(fiber.Map{"members": members})
}

func (h *OrganizationHandler) updateMemberRole(c *fiber.Ctx) error {
	user, err := GetUser(c)
	if err != nil {
		return err
	}
	orgID, parseErr := parseUUIDParam(c, "id")
	if parseErr != nil {
		return ErrorResponse(c, "Invalid organization id", parseErr, http.StatusBadRequest)
	}
	targetUserID := strings.TrimSpace(c.Params("userID"))
	if targetUserID == "" {
		return ErrorResponse(c, "user id is required", nil, http.StatusBadRequest)
	}

	var body struct {
		Role string `json:"role"`
	}
	if err := c.BodyParser(&body); err != nil {
		return ErrorResponse(c, "Invalid request body", err, http.StatusBadRequest)
	}

	if err := h.orgService.UpdateMemberRole(c.Context(), orgID, targetUserID, user.ID, body.Role); err != nil {
		status := http.StatusInternalServerError
		if apperrors.Is(err, apperrors.ErrUnauthorizedOrganizationAccess) || apperrors.Is(err, apperrors.ErrInvalidUserData) {
			status = http.StatusBadRequest
		}
		return ErrorResponse(c, "Failed to update member role", err, status)
	}
	return c.SendStatus(http.StatusNoContent)
}

func (h *OrganizationHandler) removeMember(c *fiber.Ctx) error {
	user, err := GetUser(c)
	if err != nil {
		return err
	}
	orgID, parseErr := parseUUIDParam(c, "id")
	if parseErr != nil {
		return ErrorResponse(c, "Invalid organization id", parseErr, http.StatusBadRequest)
	}
	targetUserID := strings.TrimSpace(c.Params("userID"))
	if targetUserID == "" {
		return ErrorResponse(c, "user id is required", nil, http.StatusBadRequest)
	}

	if err := h.orgService.RemoveMember(c.Context(), orgID, targetUserID, user.ID); err != nil {
		status := http.StatusInternalServerError
		if apperrors.Is(err, apperrors.ErrUnauthorizedOrganizationAccess) {
			status = http.StatusForbidden
		}
		return ErrorResponse(c, "Failed to remove member", err, status)
	}
	return c.SendStatus(http.StatusNoContent)
}

func (h *OrganizationHandler) createInvite(c *fiber.Ctx) error {
	user, err := GetUser(c)
	if err != nil {
		return err
	}
	orgID, parseErr := parseUUIDParam(c, "id")
	if parseErr != nil {
		return ErrorResponse(c, "Invalid organization id", parseErr, http.StatusBadRequest)
	}
	var req models.OrganizationInviteRequest
	if err := c.BodyParser(&req); err != nil {
		return ErrorResponse(c, "Invalid request body", err, http.StatusBadRequest)
	}

	invite, token, err := h.orgService.CreateInvite(c.Context(), orgID, user.ID, &req)
	if err != nil {
		status := http.StatusInternalServerError
		if apperrors.Is(err, apperrors.ErrUnauthorizedOrganizationAccess) || apperrors.Is(err, apperrors.ErrInvalidUserData) {
			status = http.StatusBadRequest
		}
		return ErrorResponse(c, "Failed to create invite", err, status)
	}

	return c.Status(http.StatusCreated).JSON(fiber.Map{
		"invite": invite,
		"token":  token,
	})
}

func (h *OrganizationHandler) listInvites(c *fiber.Ctx) error {
	user, err := GetUser(c)
	if err != nil {
		return err
	}
	orgID, parseErr := parseUUIDParam(c, "id")
	if parseErr != nil {
		return ErrorResponse(c, "Invalid organization id", parseErr, http.StatusBadRequest)
	}

	invites, err := h.orgService.ListInvites(c.Context(), orgID, user.ID)
	if err != nil {
		status := http.StatusInternalServerError
		if apperrors.Is(err, apperrors.ErrUnauthorizedOrganizationAccess) {
			status = http.StatusForbidden
		}
		return ErrorResponse(c, "Failed to list invites", err, status)
	}
	return c.JSON(fiber.Map{"invites": invites})
}

func (h *OrganizationHandler) acceptInvite(c *fiber.Ctx) error {
	user, err := GetUser(c)
	if err != nil {
		return err
	}
	var body struct {
		Token string `json:"token"`
	}
	if err := c.BodyParser(&body); err != nil {
		return ErrorResponse(c, "Invalid request body", err, http.StatusBadRequest)
	}
	resp, err := h.orgService.AcceptInvite(c.Context(), strings.TrimSpace(body.Token), user.ID)
	if err != nil {
		status := http.StatusInternalServerError
		if apperrors.Is(err, apperrors.ErrOrganizationInviteInvalid) {
			status = http.StatusBadRequest
		}
		return ErrorResponse(c, "Failed to accept invite", err, status)
	}
	return c.JSON(resp)
}

func (h *OrganizationHandler) getCurrentContext(c *fiber.Ctx) error {
	org := GetOrgContext(c)
	user, err := GetUser(c)
	if err != nil {
		return err
	}
	if org == nil || org.ID == nil {
		return c.JSON(fiber.Map{
			"organization": models.OrganizationResponse{
				ID:        uuid.Nil,
				Name:      "Personal",
				Slug:      "personal",
				PlanTier:  "free",
				Role:      services.OrgRolePersonal,
				CreatedBy: user.ID,
			},
		})
	}
	resp, err := h.orgService.Get(c.Context(), *org.ID, user.ID)
	if err != nil {
		return ErrorResponse(c, "Failed to load current organization", err)
	}
	return c.JSON(fiber.Map{"organization": resp})
}
