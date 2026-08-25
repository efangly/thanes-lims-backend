package user

import (
	"time"

	"github.com/efangly/thanes-lims-backend/internal/adapters/http/middleware"
	"github.com/efangly/thanes-lims-backend/internal/adapters/http/response"
	"github.com/efangly/thanes-lims-backend/internal/adapters/http/validate"
	applicationuser "github.com/efangly/thanes-lims-backend/internal/application/user"
	domainuser "github.com/efangly/thanes-lims-backend/internal/domain/user"
	"github.com/gofiber/fiber/v3"
)

type Handler struct {
	login        *applicationuser.LoginUseCase
	refresh      *applicationuser.RefreshUseCase
	logout       *applicationuser.LogoutUseCase
	logoutAll    *applicationuser.LogoutAllUseCase
	create       *applicationuser.CreateUserUseCase
	list         *applicationuser.ListUsersUseCase
	get          *applicationuser.GetUserUseCase
	update       *applicationuser.UpdateUserUseCase
	cookieSecure bool
}

func NewHandler(
	login *applicationuser.LoginUseCase,
	refresh *applicationuser.RefreshUseCase,
	logout *applicationuser.LogoutUseCase,
	logoutAll *applicationuser.LogoutAllUseCase,
	create *applicationuser.CreateUserUseCase,
	list *applicationuser.ListUsersUseCase,
	get *applicationuser.GetUserUseCase,
	update *applicationuser.UpdateUserUseCase,
	cookieSecure bool,
) *Handler {
	return &Handler{
		login:        login,
		refresh:      refresh,
		logout:       logout,
		logoutAll:    logoutAll,
		create:       create,
		list:         list,
		get:          get,
		update:       update,
		cookieSecure: cookieSecure,
	}
}

// setRefreshCookie sets the httpOnly Refresh Cookie (see ADR 0004).
func (h *Handler) setRefreshCookie(c fiber.Ctx, token string, expiresAt time.Time) {
	c.Cookie(&fiber.Cookie{
		Name:     middleware.RefreshCookieName,
		Value:    token,
		Path:     middleware.RefreshCookiePath,
		Expires:  expiresAt,
		Secure:   h.cookieSecure,
		HTTPOnly: true,
		SameSite: fiber.CookieSameSiteStrictMode,
	})
}

// clearRefreshCookie expires the Refresh Cookie client-side. Set explicitly
// with the same Path/attributes used when the cookie was set, rather than
// via Fiber's ClearCookie helper, since that assumes the default Path "/".
func (h *Handler) clearRefreshCookie(c fiber.Ctx) {
	c.Cookie(&fiber.Cookie{
		Name:     middleware.RefreshCookieName,
		Value:    "",
		Path:     middleware.RefreshCookiePath,
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		Secure:   h.cookieSecure,
		HTTPOnly: true,
		SameSite: fiber.CookieSameSiteStrictMode,
	})
}

// Login godoc
//
//	@Summary		เข้าสู่ระบบ
//	@Description	ตรวจสอบอีเมล/รหัสผ่าน คืน access token ใน body และตั้ง refresh token เป็น httpOnly cookie
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		LoginRequest	true	"อีเมลและรหัสผ่าน"
//	@Success		200		{object}	response.Envelope{data=AccessTokenResponse}
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

	pair, err := h.login.Execute(c.Context(), req.Email, req.Password, string(c.Request().Header.UserAgent()), c.IP())
	if err != nil {
		return err
	}
	h.setRefreshCookie(c, pair.RefreshToken, pair.RefreshExpiresAt)
	return response.OK(c, AccessTokenResponse{AccessToken: pair.AccessToken})
}

// Refresh godoc
//
//	@Summary		ต่ออายุ token
//	@Description	แลก refresh token เก่า (จาก cookie หรือ body) เป็น access token ใหม่ และหมุน refresh token cookie
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		RefreshRequest	false	"Refresh token (ใช้เมื่อไม่มี cookie)"
//	@Success		200		{object}	response.Envelope{data=TokenPairResponse}
//	@Failure		400		{object}	response.Envelope
//	@Failure		401		{object}	response.Envelope
//	@Router			/auth/refresh [post]
func (h *Handler) Refresh(c fiber.Ctx) error {
	raw := c.Cookies(middleware.RefreshCookieName)
	usedCookie := raw != ""
	if !usedCookie {
		var req RefreshRequest
		if err := c.Bind().Body(&req); err != nil {
			return err
		}
		if err := validate.Struct(req); err != nil {
			return err
		}
		raw = req.RefreshToken
	}

	pair, err := h.refresh.Execute(c.Context(), raw, string(c.Request().Header.UserAgent()), c.IP())
	if err != nil {
		h.clearRefreshCookie(c)
		return err
	}

	h.setRefreshCookie(c, pair.RefreshToken, pair.RefreshExpiresAt)
	resp := TokenPairResponse{AccessToken: pair.AccessToken}
	if !usedCookie {
		resp.RefreshToken = pair.RefreshToken
	}
	return response.OK(c, resp)
}

// Logout godoc
//
//	@Summary		ออกจากระบบ
//	@Description	เพิกถอน refresh token ของ session ปัจจุบัน (จาก cookie หรือ body) แล้วล้าง cookie
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		RefreshRequest	false	"Refresh token ที่ต้องการเพิกถอน (ใช้เมื่อไม่มี cookie)"
//	@Success		200		{object}	response.Envelope
//	@Failure		400		{object}	response.Envelope
//	@Router			/auth/logout [post]
func (h *Handler) Logout(c fiber.Ctx) error {
	raw := c.Cookies(middleware.RefreshCookieName)
	if raw == "" {
		var req RefreshRequest
		if err := c.Bind().Body(&req); err == nil {
			raw = req.RefreshToken
		}
	}
	if raw != "" {
		if err := h.logout.Execute(c.Context(), raw); err != nil {
			return err
		}
	}
	h.clearRefreshCookie(c)
	return response.OK(c, fiber.Map{"logged_out": true})
}

// LogoutAll godoc
//
//	@Summary		ออกจากระบบทุกอุปกรณ์
//	@Description	เพิกถอน refresh token ของทุก session ของผู้ใช้ปัจจุบัน แล้วล้าง cookie
//	@Tags			auth
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	response.Envelope
//	@Failure		401	{object}	response.Envelope
//	@Router			/auth/logout-all [post]
func (h *Handler) LogoutAll(c fiber.Ctx) error {
	userID := fiber.Locals[int64](c, middleware.LocalsUserID)
	if err := h.logoutAll.Execute(c.Context(), userID); err != nil {
		return err
	}
	h.clearRefreshCookie(c)
	return response.OK(c, fiber.Map{"logged_out_all": true})
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
	// Snapshot the response DTO, not the domain User - it has no
	// PasswordHash field, so the bcrypt hash never lands in Metadata.
	c.Locals(middleware.LocalsAuditChangeSet, middleware.Snapshot(toUserResponse(u)))
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

	before, err := h.get.Execute(c.Context(), id)
	if err != nil {
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
	c.Locals(middleware.LocalsAuditChangeSet, middleware.ChangeSet(toUserResponse(before), toUserResponse(u)))
	return response.OK(c, toUserResponse(u))
}
