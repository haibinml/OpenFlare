// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/Rain-kl/Wavelet/internal/apps/oauth"
	"github.com/Rain-kl/Wavelet/internal/apps/openflare/compat"
	"github.com/Rain-kl/Wavelet/internal/buildinfo"
	"github.com/Rain-kl/Wavelet/internal/config"
	"github.com/Rain-kl/Wavelet/internal/db"
	"github.com/Rain-kl/Wavelet/internal/db/idgen"
	"github.com/Rain-kl/Wavelet/internal/listener"
	"github.com/Rain-kl/Wavelet/internal/model"
	"github.com/Rain-kl/Wavelet/internal/repository"
	"github.com/Rain-kl/Wavelet/internal/task"
	"github.com/Rain-kl/Wavelet/pkg/logger"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const (
	legacyTokenName       = "openflare-legacy"
	minPasswordLength     = 8
	legacyItemsPerPage    = 10
	passwordResetKeyFmt   = "of_password_reset:%s"
	passwordResetExpiry   = 15 * time.Minute
	verificationCodeRange = 900000
	verificationOffset    = 100000
)

// LoginInput holds legacy login credentials.
type LoginInput struct {
	Username string
	Password string
	Code     string
}

// RegisterInput holds legacy registration fields.
type RegisterInput struct {
	Username    string
	Password    string
	Nickname    string
	DisplayName string
	Email       string
	Code        string
}

// ManageUserInput holds legacy user management actions.
type ManageUserInput struct {
	Username string
	Action   string
}

// UpdateUserInput holds legacy admin user update fields.
type UpdateUserInput struct {
	ID          int
	Username    string
	Password    string
	DisplayName string
	Role        int
	Email       string
}

// UpdateSelfInput holds legacy self-update fields.
type UpdateSelfInput struct {
	Username    string
	Password    string
	DisplayName string
	Email       string
}

// CreateUserInput holds legacy admin create-user fields.
type CreateUserInput struct {
	Username    string
	Password    string
	DisplayName string
	Role        int
	Email       string
}

// PasswordResetInput holds password reset confirmation.
type PasswordResetInput struct {
	Email string
	Token string
}

// LinkExistingInput binds a pending OAuth account to an existing user.
type LinkExistingInput struct {
	Username string
	Password string
}

// Login authenticates a user and returns the legacy user shape with an access token.
func Login(ctx context.Context, c *gin.Context, input LoginInput) (LegacyUser, error) {
	if !isPasswordLoginEnabled(ctx) {
		return LegacyUser{}, errors.New(errPasswordLoginDisabled)
	}

	input.Username = strings.TrimSpace(input.Username)
	if input.Username == "" || input.Password == "" {
		return LegacyUser{}, errors.New(errInvalidParams)
	}

	var user model.User
	if err := db.DB(ctx).Where("username = ? OR email = ?", input.Username, input.Username).First(&user).Error; err != nil {
		logger.WarnF(ctx, "[LoginAudit] failed login attempt (username not found) for input: %s, IP: %s", input.Username, c.ClientIP())
		return LegacyUser{}, errors.New(errUsernameOrPasswordWrong)
	}
	if !user.IsActive {
		return LegacyUser{}, errors.New(errBannedAccount)
	}
	if !user.CheckPassword(input.Password) {
		logger.WarnF(ctx, "[LoginAudit] failed login attempt (incorrect password) for username: %s, ID: %d, IP: %s", user.Username, user.ID, c.ClientIP())
		return LegacyUser{}, errors.New(errUsernameOrPasswordWrong)
	}

	if isEmailLoginVerificationEnabled(ctx) {
		if err := verifyLoginEmailCode(ctx, user.Email, input.Code); err != nil {
			return LegacyUser{}, err
		}
	}

	user.LastLoginAt = time.Now()
	if err := db.DB(ctx).Model(&user).Update("last_login_at", user.LastLoginAt).Error; err != nil {
		return LegacyUser{}, err
	}
	if err := setLoginSession(ctx, c, &user); err != nil {
		return LegacyUser{}, errors.New(errSaveSessionFailed)
	}

	token, err := issueLegacyAccessToken(ctx, &user)
	if err != nil {
		return LegacyUser{}, err
	}

	logger.InfoF(ctx, "[LoginAudit] successful legacy login for user: %s, ID: %d, IP: %s", user.Username, user.ID, c.ClientIP())
	listener.EmitAdminLoggedIn(ctx, &user, c.ClientIP())

	return ToLegacyUser(&user, token), nil
}

// Register creates a new user and logs them in.
func Register(ctx context.Context, c *gin.Context, input RegisterInput) (LegacyUser, error) {
	if !isRegistrationEnabled(ctx) || !isPasswordRegisterEnabled(ctx) {
		return LegacyUser{}, errors.New(errRegistrationDisabled)
	}

	input.Username = strings.TrimSpace(input.Username)
	input.Password = strings.TrimSpace(input.Password)
	input.Nickname = strings.TrimSpace(input.Nickname)
	input.DisplayName = strings.TrimSpace(input.DisplayName)
	input.Email = strings.TrimSpace(input.Email)
	input.Code = strings.TrimSpace(input.Code)

	if input.Username == "" || input.Password == "" {
		return LegacyUser{}, errors.New(errInvalidParams)
	}
	if len(input.Password) < minPasswordLength {
		return LegacyUser{}, errors.New(errPasswordTooShort)
	}
	if input.Email == "" {
		return LegacyUser{}, errors.New(errEmailRequired)
	}
	if isEmailRegisterVerificationEnabled(ctx) {
		if input.Code == "" || !verifyEmailCode(ctx, input.Email, "register", input.Code) {
			return LegacyUser{}, errors.New(errEmailCodeInvalid)
		}
	}

	user := model.User{
		ID:          idgen.NextUint64ID(),
		Username:    input.Username,
		Nickname:    input.Nickname,
		Email:       input.Email,
		IsActive:    true,
		IsAdmin:     false,
		LastLoginAt: time.Now(),
	}
	if user.Nickname == "" {
		user.Nickname = input.DisplayName
	}
	if user.Nickname == "" {
		user.Nickname = input.Username
	}
	if err := user.SetEncryptedPassword(input.Password); err != nil {
		return LegacyUser{}, err
	}
	if err := user.RegisterUser(ctx, db.DB(ctx)); err != nil {
		return LegacyUser{}, err
	}
	if err := setLoginSession(ctx, c, &user); err != nil {
		return LegacyUser{}, errors.New(errSaveSessionFailed)
	}

	token, err := issueLegacyAccessToken(ctx, &user)
	if err != nil {
		return LegacyUser{}, err
	}
	return ToLegacyUser(&user, token), nil
}

// Logout clears session and revokes the legacy access token when provided.
func Logout(ctx context.Context, c *gin.Context) error {
	token := strings.TrimSpace(c.GetHeader(compat.OpenFlareTokenHeader()))
	if token == "" {
		token = strings.TrimSpace(c.GetHeader("X-Access-Token"))
	}
	if token != "" {
		tokenHash := model.HashToken(token)
		_ = db.DB(ctx).Where("token_hash = ? AND name = ?", tokenHash, legacyTokenName).Delete(&model.AccessToken{}).Error
	}

	session := sessions.Default(c)
	session.Options(oauth.GetSessionOptions(-1))
	session.Clear()
	return session.Save()
}

// GetSelf returns the current user's legacy profile.
func GetSelf(ctx context.Context, userID uint64) (LegacyUser, error) {
	user, err := repository.GetUserByID(ctx, userID)
	if err != nil {
		return LegacyUser{}, errors.New(errUserNotFound)
	}
	return ToLegacyUser(&user, ""), nil
}

// GenerateUserToken issues a fresh legacy access token for the user.
func GenerateUserToken(ctx context.Context, userID uint64) (string, error) {
	user, err := repository.GetUserByID(ctx, userID)
	if err != nil {
		return "", errors.New(errUserNotFound)
	}
	return issueLegacyAccessToken(ctx, &user)
}

// UpdateSelf updates the logged-in user's profile.
func UpdateSelf(ctx context.Context, userID uint64, input UpdateSelfInput) error {
	user, err := repository.GetUserByID(ctx, userID)
	if err != nil {
		return errors.New(errUserNotFound)
	}

	if input.DisplayName != "" {
		user.Nickname = strings.TrimSpace(input.DisplayName)
	}
	if input.Username != "" {
		user.Username = strings.TrimSpace(input.Username)
	}
	if input.Email != "" {
		user.Email = strings.TrimSpace(input.Email)
	}
	if input.Password != "" {
		if len(input.Password) < minPasswordLength {
			return errors.New(errPasswordTooShort)
		}
		if err := user.SetEncryptedPassword(input.Password); err != nil {
			return err
		}
	}
	return db.DB(ctx).Save(&user).Error
}

// DeleteSelf removes the logged-in user.
func DeleteSelf(ctx context.Context, userID uint64) error {
	return repository.DeleteUserWithRelations(ctx, userID)
}

// ListUsers returns a paginated legacy user list.
func ListUsers(ctx context.Context, page int) ([]LegacyUser, error) {
	if page < 0 {
		page = 0
	}
	_, users, err := repository.ListAdminUsers(ctx, repository.AdminUserListFilter{
		Page:     page + 1,
		PageSize: legacyItemsPerPage,
	})
	if err != nil {
		return nil, err
	}
	return ToLegacyUsers(users), nil
}

// SearchUsers searches users by keyword.
func SearchUsers(ctx context.Context, keyword string) ([]LegacyUser, error) {
	keyword = strings.TrimSpace(keyword)
	var users []model.User
	query := db.DB(ctx).Model(&model.User{}).
		Select("id, username, nickname, email, is_active, is_admin")
	if keyword != "" {
		like := keyword + "%"
		query = query.Where(
			"CAST(id AS TEXT) = ? OR username LIKE ? OR email LIKE ? OR nickname LIKE ?",
			keyword, like, like, like,
		)
	}
	if err := query.Order("id DESC").Find(&users).Error; err != nil {
		return nil, err
	}
	return ToLegacyUsers(users), nil
}

// GetUserByID returns a legacy user if the caller has sufficient role.
func GetUserByID(ctx context.Context, callerRole int, id uint64) (LegacyUser, error) {
	user, err := repository.GetAdminUserDetail(ctx, id)
	if err != nil {
		return LegacyUser{}, errors.New(errUserNotFound)
	}
	targetRole := RoleFromUser(&user)
	if callerRole <= targetRole {
		return LegacyUser{}, errors.New(errInsufficientPermission)
	}
	return ToLegacyUser(&user, ""), nil
}

// CreateUser creates a user from legacy admin input.
func CreateUser(ctx context.Context, callerRole int, input CreateUserInput) error {
	input.Username = strings.TrimSpace(input.Username)
	input.Password = strings.TrimSpace(input.Password)
	input.DisplayName = strings.TrimSpace(input.DisplayName)
	input.Email = strings.TrimSpace(input.Email)

	if input.Username == "" || input.Password == "" {
		return errors.New(errInvalidParams)
	}
	if input.Role >= callerRole {
		return errors.New(errInsufficientPermission)
	}
	if input.Email == "" {
		input.Email = input.Username + "@openflare.local"
	}

	newUser := model.User{
		ID:       idgen.NextUint64ID(),
		Username: input.Username,
		Nickname: input.DisplayName,
		Email:    input.Email,
		IsActive: true,
		IsAdmin:  input.Role >= compat.RoleAdminUser,
	}
	if newUser.Nickname == "" {
		newUser.Nickname = input.Username
	}
	if err := newUser.SetEncryptedPassword(input.Password); err != nil {
		return err
	}
	return repository.CreateUser(ctx, &newUser)
}

// UpdateUser updates another user with legacy role checks.
func UpdateUser(ctx context.Context, callerRole int, input UpdateUserInput) error {
	if input.ID == 0 {
		return errors.New(errInvalidParams)
	}

	origin, err := repository.GetAdminUserDetail(ctx, uint64(input.ID))
	if err != nil {
		return errors.New(errUserNotFound)
	}
	originRole := RoleFromUser(&origin)
	if callerRole <= originRole {
		return errors.New(errInsufficientPermission)
	}
	if input.Role > 0 && callerRole <= input.Role {
		return errors.New(errInsufficientPermission)
	}

	if trimmed := strings.TrimSpace(input.Username); trimmed != "" {
		origin.Username = trimmed
	}
	if input.DisplayName != "" {
		origin.Nickname = strings.TrimSpace(input.DisplayName)
	}
	if input.Email != "" {
		origin.Email = strings.TrimSpace(input.Email)
	}
	if input.Role > 0 {
		origin.IsAdmin = input.Role >= compat.RoleAdminUser
	}
	if input.Password != "" {
		if len(input.Password) < minPasswordLength {
			return errors.New(errPasswordTooShort)
		}
		if err := origin.SetEncryptedPassword(input.Password); err != nil {
			return err
		}
	}
	return db.DB(ctx).Save(&origin).Error
}

// DeleteUserByID deletes a user when the caller has sufficient role.
func DeleteUserByID(ctx context.Context, callerRole int, id uint64) error {
	origin, err := repository.GetAdminUserDetail(ctx, id)
	if err != nil {
		return errors.New(errUserNotFound)
	}
	if callerRole <= RoleFromUser(&origin) {
		return errors.New(errInsufficientPermission)
	}
	if RoleFromUser(&origin) >= compat.RoleRootUser {
		return errors.New(errCannotDeleteRoot)
	}
	return repository.DeleteUserWithRelations(ctx, id)
}

// ManageUser performs enable/disable/delete/promote/demote actions.
func ManageUser(ctx context.Context, callerRole int, input ManageUserInput) (LegacyUser, error) {
	input.Username = strings.TrimSpace(input.Username)
	input.Action = strings.TrimSpace(input.Action)
	if input.Username == "" || input.Action == "" {
		return LegacyUser{}, errors.New(errInvalidParams)
	}

	user, err := repository.GetUserByUsername(ctx, input.Username)
	if err != nil {
		return LegacyUser{}, errors.New(errUserNotFound)
	}
	targetRole := RoleFromUser(&user)
	if callerRole <= targetRole && callerRole != compat.RoleRootUser {
		return LegacyUser{}, errors.New(errInsufficientPermission)
	}

	switch input.Action {
	case "disable":
		if targetRole >= compat.RoleRootUser {
			return LegacyUser{}, errors.New(errCannotDisableRoot)
		}
		user.IsActive = false
	case "enable":
		user.IsActive = true
	case "delete":
		if targetRole >= compat.RoleRootUser {
			return LegacyUser{}, errors.New(errCannotDeleteRoot)
		}
		if err := repository.DeleteUserWithRelations(ctx, user.ID); err != nil {
			return LegacyUser{}, err
		}
		return LegacyUser{Role: compat.RoleCommonUser, Status: legacyUserStatusDisabled}, nil
	case "promote":
		if callerRole != compat.RoleRootUser {
			return LegacyUser{}, errors.New(errCannotPromoteAdmin)
		}
		if user.IsAdmin {
			return LegacyUser{}, errors.New(errAlreadyAdmin)
		}
		user.IsAdmin = true
	case "demote":
		if targetRole >= compat.RoleRootUser {
			return LegacyUser{}, errors.New(errCannotDisableRoot)
		}
		if !user.IsAdmin {
			return LegacyUser{}, errors.New(errAlreadyCommonUser)
		}
		user.IsAdmin = false
	default:
		return LegacyUser{}, errors.New(errInvalidParams)
	}

	if err := db.DB(ctx).Save(&user).Error; err != nil {
		return LegacyUser{}, err
	}
	return LegacyUser{
		Role:   RoleFromUser(&user),
		Status: StatusFromUser(&user),
	}, nil
}

// SendRegisterVerificationEmail sends a registration verification code.
func SendRegisterVerificationEmail(ctx context.Context, email string) error {
	email = strings.TrimSpace(email)
	if email == "" || !strings.Contains(email, "@") {
		return errors.New(errInvalidParams)
	}
	var count int64
	if err := db.DB(ctx).Model(&model.User{}).Where("email = ?", email).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return errors.New(errEmailAlreadyRegistered)
	}
	return sendEmailVerificationCode(ctx, email, "register", "register_email")
}

// SendPasswordResetEmail stores a reset token and emails the user.
func SendPasswordResetEmail(ctx context.Context, email string) error {
	email = strings.TrimSpace(email)
	if email == "" || !strings.Contains(email, "@") {
		return errors.New(errInvalidParams)
	}
	var user model.User
	if err := db.DB(ctx).Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New(errEmailNotRegistered)
		}
		return err
	}

	token, err := generateResetToken()
	if err != nil {
		return err
	}
	key := fmt.Sprintf(passwordResetKeyFmt, email)
	if err := db.SetJSON(ctx, key, token, passwordResetExpiry); err != nil {
		return err
	}

	serverAddr, _ := repository.GetSystemConfigByKey(ctx, model.ConfigKeyServerAddress)
	base := strings.TrimRight(serverAddr.Value, "/")
	link := fmt.Sprintf("%s/user/reset?email=%s&token=%s", base, email, token)
	body := fmt.Sprintf("<p>您好，你正在进行密码重置。</p><p>点击<a href='%s'>此处</a>进行密码重置。</p>", link)
	return dispatchEmail(ctx, email, "密码重置", body)
}

// ResetPassword validates a reset token and returns a new random password.
func ResetPassword(ctx context.Context, input PasswordResetInput) (string, error) {
	input.Email = strings.TrimSpace(input.Email)
	input.Token = strings.TrimSpace(input.Token)
	if input.Email == "" || input.Token == "" {
		return "", errors.New(errInvalidParams)
	}

	key := fmt.Sprintf(passwordResetKeyFmt, input.Email)
	var stored string
	if err := db.GetJSON(ctx, key, &stored); err != nil || stored != input.Token {
		return "", errors.New(errResetLinkInvalid)
	}

	password, err := generateResetToken()
	if err != nil {
		return "", err
	}
	if len(password) < minPasswordLength {
		password = password + "Aa1!"
	}

	var user model.User
	if err := db.DB(ctx).Where("email = ?", input.Email).First(&user).Error; err != nil {
		return "", errors.New(errEmailNotRegistered)
	}
	if err := user.SetEncryptedPassword(password); err != nil {
		return "", err
	}
	if err := db.DB(ctx).Model(&user).Update("password", user.Password).Error; err != nil {
		return "", err
	}
	_ = db.Redis.Del(ctx, db.PrefixedKey(key)).Err()
	return password, nil
}

// BuildPublicStatus assembles the legacy /api/status payload.
func BuildPublicStatus(ctx context.Context) (map[string]any, error) {
	authSources, err := publicAuthSources(ctx, "/api")
	if err != nil {
		authSources = []map[string]any{}
	}

	siteName, _ := repository.GetSystemConfigByKey(ctx, model.ConfigKeySiteName)
	serverAddr, _ := repository.GetSystemConfigByKey(ctx, model.ConfigKeyServerAddress)

	return map[string]any{
		"version":                   buildVersion(),
		"start_time":                appStartUnix(),
		"email_verification":        isEmailRegisterVerificationEnabled(ctx),
		"github_oauth":              false,
		"github_client_id":          "",
		"system_name":               siteName.Value,
		"home_page_link":            "",
		"footer_html":               "",
		"wechat_qrcode":             "",
		"wechat_login":              false,
		"server_address":            serverAddr.Value,
		"password_register_enabled": isPasswordRegisterEnabled(ctx),
		"cap_login_enabled":         capLoginEnabled(ctx),
		"auth_sources":              authSources,
	}, nil
}

// GetNotice returns the legacy notice option value.
func GetNotice(ctx context.Context) string {
	return getOptionValue(ctx, "notice")
}

// GetAbout returns the legacy about option value.
func GetAbout(ctx context.Context) string {
	return getOptionValue(ctx, "about")
}

func issueLegacyAccessToken(ctx context.Context, user *model.User) (string, error) {
	tokenStr, err := model.GenerateTokenString()
	if err != nil {
		return "", errors.New(errGenerateTokenFailed)
	}
	record := model.AccessToken{
		UserID:      user.ID,
		Name:        legacyTokenName,
		TokenHash:   model.HashToken(tokenStr),
		MaskedToken: model.MaskTokenString(tokenStr),
		IsAdmin:     user.IsAdmin,
	}
	if err := db.DB(ctx).Create(&record).Error; err != nil {
		return "", err
	}
	return tokenStr, nil
}

func setLoginSession(ctx context.Context, c *gin.Context, user *model.User) error {
	session := sessions.Default(c)
	session.Set(oauth.UserIDKey, user.ID)
	session.Set(oauth.UserNameKey, user.Username)
	session.Set(oauth.PasswordHashKey, user.Password)

	maxAge := config.Config.App.SessionAge
	isSessionCookie := false
	ttlHours, err := repository.GetIntByKey(ctx, model.ConfigKeyLoginSessionTTLHours)
	if err == nil {
		switch {
		case ttlHours == -1:
			maxAge = 10 * 365 * 24 * 3600
		case ttlHours > 0:
			maxAge = ttlHours * 3600
		case ttlHours == 0:
			isSessionCookie = true
		}
	}
	session.Options(oauth.GetSessionOptions(maxAge))
	if err := session.Save(); err != nil {
		return err
	}
	if isSessionCookie {
		oauth.StripCookieMaxAgeAndExpires(c.Writer.Header(), config.Config.App.SessionCookieName)
	}
	return nil
}

func isPasswordLoginEnabled(ctx context.Context) bool {
	enabled, err := repository.GetBoolByKey(ctx, model.ConfigKeyPasswordLoginEnabled)
	return err != nil || enabled
}

func isPasswordRegisterEnabled(ctx context.Context) bool {
	enabled, err := repository.GetBoolByKey(ctx, model.ConfigKeyPasswordRegisterEnabled)
	return err != nil || enabled
}

func isRegistrationEnabled(ctx context.Context) bool {
	enabled, err := repository.GetBoolByKey(ctx, model.ConfigKeyRegistrationEnabled)
	return err != nil || enabled
}

func isEmailLoginVerificationEnabled(ctx context.Context) bool {
	enabled, err := repository.GetBoolByKey(ctx, model.ConfigKeyEmailLoginVerificationEnabled)
	return err == nil && enabled
}

func isEmailRegisterVerificationEnabled(ctx context.Context) bool {
	enabled, err := repository.GetBoolByKey(ctx, model.ConfigKeyEmailRegisterVerificationEnabled)
	return err == nil && enabled
}

func capLoginEnabled(ctx context.Context) bool {
	enabled, err := repository.GetBoolByKey(ctx, model.ConfigKeyCapLoginEnabled)
	return err == nil && enabled
}

func verifyLoginEmailCode(ctx context.Context, email, code string) error {
	if code == "" {
		return errors.New("need_email_code:" + email)
	}
	if !verifyEmailCode(ctx, email, "login", code) {
		return errors.New(errEmailCodeInvalid)
	}
	return nil
}

func verifyEmailCode(ctx context.Context, email, scene, code string) bool {
	key := fmt.Sprintf("email_code:%s:%s", scene, email)
	var stored string
	if err := db.GetJSON(ctx, key, &stored); err != nil {
		return false
	}
	if stored != code {
		return false
	}
	_ = db.Redis.Del(ctx, db.PrefixedKey(key)).Err()
	return true
}

func sendEmailVerificationCode(ctx context.Context, email, scene, templateName string) error {
	scHost, errHost := repository.GetSystemConfigByKey(ctx, model.ConfigKeySMTPHost)
	scPort, errPort := repository.GetSystemConfigByKey(ctx, model.ConfigKeySMTPPort)
	scUser, errUser := repository.GetSystemConfigByKey(ctx, model.ConfigKeySMTPUsername)
	scPass, errPass := repository.GetSystemConfigByKey(ctx, model.ConfigKeySMTPPassword)
	if errHost != nil || errPort != nil || errUser != nil || errPass != nil ||
		scHost.Value == "" || scPort.Value == "" || scUser.Value == "" || scPass.Value == "" {
		return errors.New("系统 SMTP 邮件服务配置不完整")
	}

	code, err := generateVerificationCode()
	if err != nil {
		return err
	}
	codeKey := fmt.Sprintf("email_code:%s:%s", scene, email)
	if err := db.SetJSON(ctx, codeKey, code, 5*time.Minute); err != nil {
		return err
	}

	tmpl, err := repository.GetTemplateByKey(ctx, templateName)
	if err != nil {
		body := fmt.Sprintf("<p>您的验证码为: <strong>%s</strong></p>", code)
		return dispatchEmail(ctx, email, "邮箱验证", body)
	}
	subject, body, err := tmpl.Render(map[string]any{"Code": code})
	if err != nil {
		return err
	}
	return dispatchEmail(ctx, email, subject, body)
}

type sendEmailPayload struct {
	To      string `json:"to"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

func dispatchEmail(ctx context.Context, to, subject, body string) error {
	payload := sendEmailPayload{To: to, Subject: subject, Body: body}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	_, err = task.DispatchTask(ctx, "mail:send", payloadBytes, "system")
	return err
}

func generateVerificationCode() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(verificationCodeRange))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()+verificationOffset), nil
}

func generateResetToken() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1<<62))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", n.Int64()), nil
}

func publicAuthSources(ctx context.Context, baseAPIPath string) ([]map[string]any, error) {
	sources, err := model.GetActiveAuthSources(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]map[string]any, 0, len(sources))
	base := strings.TrimRight(baseAPIPath, "/")
	for _, source := range sources {
		result = append(result, map[string]any{
			"id":            source.ID,
			"name":          source.Name,
			"type":          source.Type,
			"display_name":  source.DisplayName,
			"authorize_url": fmt.Sprintf("%s/oauth/%s/authorize", base, source.Name),
			"icon_url":      source.IconURL,
		})
	}
	return result, nil
}

func getOptionValue(ctx context.Context, key string) string {
	sc, err := repository.GetSystemConfigByKey(ctx, key)
	if err != nil {
		return ""
	}
	return sc.Value
}

var appStart = time.Now()

func appStartUnix() int64 {
	return appStart.Unix()
}

func buildVersion() string {
	if buildinfo.Version != "" {
		return buildinfo.Version
	}
	return "dev"
}
