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

// Login godoc
//
//	@Summary		เข้าสู่ระบบ
//	@Description	ตรวจสอบอีเมล/รหัสผ่าน แล้วคืน access/refresh token pair
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		LoginRequest	true	"อีเมลและรหัสผ่าน"
//	@Success		200		{object}	response.Envelope{data=TokenPairResponse}
//	@Failure		400		{object}	response.Envelope
//	@Failure		401		{object}	response.Envelope
//	@Router			/auth/login [post]
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

// Refresh godoc
//
//	@Summary		ต่ออายุ token
//	@Description	แลก refresh token เก่าเป็น access/refresh token pair ใหม่
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		RefreshRequest	true	"Refresh token"
//	@Success		200		{object}	response.Envelope{data=TokenPairResponse}
//	@Failure		400		{object}	response.Envelope
//	@Failure		401		{object}	response.Envelope
//	@Router			/auth/refresh [post]
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

// Logout godoc
//
//	@Summary		ออกจากระบบ
//	@Description	เพิกถอน refresh token ที่ระบุ
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		RefreshRequest	true	"Refresh token ที่ต้องการเพิกถอน"
//	@Success		200		{object}	response.Envelope
//	@Failure		400		{object}	response.Envelope
//	@Router			/auth/logout [post]
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

// Me godoc
//
//	@Summary		ข้อมูลผู้ใช้ปัจจุบัน
//	@Description	คืนข้อมูลผู้ใช้ที่ authenticate อยู่ (จาก access token)
//	@Tags			users
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	response.Envelope{data=UserResponse}
//	@Failure		401	{object}	response.Envelope
//	@Router			/users/me [get]
func (h *Handler) Me(c fiber.Ctx) error {
	userID := fiber.Locals[int64](c, middleware.LocalsUserID)
	u, err := h.get.Execute(c.Context(), userID)
	if err != nil {
		return err
	}
	return response.OK(c, toUserResponse(u))
}

// ListUsers godoc
//
//	@Summary		รายชื่อผู้ใช้ทั้งหมด
//	@Description	เฉพาะ admin เท่านั้น
//	@Tags			users
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	response.Envelope{data=[]UserResponse}
//	@Failure		401	{object}	response.Envelope
//	@Failure		403	{object}	response.Envelope
//	@Router			/users [get]
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

// CreateUser godoc
//
//	@Summary		สร้างผู้ใช้ใหม่
//	@Description	เฉพาะ admin เท่านั้น
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		CreateUserRequest	true	"ข้อมูลผู้ใช้ใหม่"
//	@Success		201		{object}	response.Envelope{data=UserResponse}
//	@Failure		400		{object}	response.Envelope
//	@Failure		401		{object}	response.Envelope
//	@Failure		403		{object}	response.Envelope
//	@Failure		409		{object}	response.Envelope
//	@Router			/users [post]
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

// UpdateUser godoc
//
//	@Summary		แก้ไขผู้ใช้
//	@Description	เฉพาะ admin เท่านั้น
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		int					true	"User ID"
//	@Param			request	body		UpdateUserRequest	true	"ข้อมูลที่ต้องการแก้ไข"
//	@Success		200		{object}	response.Envelope{data=UserResponse}
//	@Failure		400		{object}	response.Envelope
//	@Failure		401		{object}	response.Envelope
//	@Failure		403		{object}	response.Envelope
//	@Failure		404		{object}	response.Envelope
//	@Router			/users/{id} [patch]
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
