package user

import (
	"github.com/efangly/thanes-lims-backend/internal/adapters/http/middleware"
	"github.com/efangly/thanes-lims-backend/internal/adapters/http/response"
	"github.com/efangly/thanes-lims-backend/internal/adapters/http/validate"
	applicationuser "github.com/efangly/thanes-lims-backend/internal/application/user"
	domainuser "github.com/efangly/thanes-lims-backend/internal/domain/user"
	"github.com/gofiber/fiber/v3"
)

type Handler struct {
	login   *applicationuser.LoginUseCase
	refresh *applicationuser.RefreshUseCase
	logout  *applicationuser.LogoutUseCase
	create  *applicationuser.CreateUserUseCase
	list    *applicationuser.ListUsersUseCase
	get     *applicationuser.GetUserUseCase
	update  *applicationuser.UpdateUserUseCase
}

func NewHandler(
	login *applicationuser.LoginUseCase,
	refresh *applicationuser.RefreshUseCase,
	logout *applicationuser.LogoutUseCase,
	create *applicationuser.CreateUserUseCase,
	list *applicationuser.ListUsersUseCase,
	get *applicationuser.GetUserUseCase,
	update *applicationuser.UpdateUserUseCase,
) *Handler {
	return &Handler{login: login, refresh: refresh, logout: logout, create: create, list: list, get: get, update: update}
}

func (h *Handler) Login(c fiber.Ctx) error {
	var req LoginRequest
	if err := c.Bind().Body(&req); err != nil {
		return err
	}
	if err := validate.Struct(req); err != nil {
		return err
	}

	pair, err := h.login.Execute(c.Context(), req.Email, req.Password)
	if err != nil {
		return err
	}
	return response.OK(c, TokenPairResponse{AccessToken: pair.AccessToken, RefreshToken: pair.RefreshToken})
}

func (h *Handler) Refresh(c fiber.Ctx) error {
	var req RefreshRequest
	if err := c.Bind().Body(&req); err != nil {
		return err
	}
	if err := validate.Struct(req); err != nil {
		return err
	}

	pair, err := h.refresh.Execute(c.Context(), req.RefreshToken)
	if err != nil {
		return err
	}
	return response.OK(c, TokenPairResponse{AccessToken: pair.AccessToken, RefreshToken: pair.RefreshToken})
}

func (h *Handler) Logout(c fiber.Ctx) error {
	var req RefreshRequest
	if err := c.Bind().Body(&req); err != nil {
		return err
	}
	if err := h.logout.Execute(c.Context(), req.RefreshToken); err != nil {
		return err
	}
	return response.OK(c, fiber.Map{"logged_out": true})
}

func (h *Handler) Me(c fiber.Ctx) error {
	userID := fiber.Locals[int64](c, middleware.LocalsUserID)
	u, err := h.get.Execute(c.Context(), userID)
	if err != nil {
		return err
	}
	return response.OK(c, toUserResponse(u))
}

func (h *Handler) ListUsers(c fiber.Ctx) error {
	users, err := h.list.Execute(c.Context())
	if err != nil {
		return err
	}
	out := make([]UserResponse, len(users))
	for i, u := range users {
		out[i] = toUserResponse(u)
	}
	return response.OK(c, out)
}

func (h *Handler) CreateUser(c fiber.Ctx) error {
	var req CreateUserRequest
	if err := c.Bind().Body(&req); err != nil {
		return err
	}
	if err := validate.Struct(req); err != nil {
		return err
	}

	u, err := h.create.Execute(c.Context(), applicationuser.CreateUserInput{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
		Role:     domainuser.Role(req.Role),
	})
	if err != nil {
		return err
	}
	return response.Created(c, toUserResponse(u))
}

func (h *Handler) UpdateUser(c fiber.Ctx) error {
	id := fiber.Params[int64](c, "id")

	var req UpdateUserRequest
	if err := c.Bind().Body(&req); err != nil {
		return err
	}
	if err := validate.Struct(req); err != nil {
		return err
	}

	u, err := h.update.Execute(c.Context(), applicationuser.UpdateUserInput{
		ID:   id,
		Name: req.Name,
		Role: domainuser.Role(req.Role),
	})
	if err != nil {
		return err
	}
	return response.OK(c, toUserResponse(u))
}
