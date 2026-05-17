package admin

import (
	"context"
	"errors"
	"net/mail"
	"strings"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrEmailRequired        = errors.New("email is required")
	ErrInvalidEmail         = errors.New("email is invalid")
	ErrPasswordRequired     = errors.New("password is required")
	ErrPasswordTooShort     = errors.New("password must be at least 8 characters")
	ErrDisplayNameRequired  = errors.New("display name is required")
	ErrInvalidCredentials   = errors.New("invalid email or password")
	ErrInactiveAdmin        = errors.New("admin is inactive")
	ErrForbidden            = errors.New("forbidden")
	ErrCannotDeactivateSelf = errors.New("owner cannot deactivate themselves")
)

type Service struct {
	repository *Repository
	ownerEmail string
}

func NewService(repository *Repository, ownerEmail string) *Service {
	return &Service{
		repository: repository,
		ownerEmail: strings.ToLower(strings.TrimSpace(ownerEmail)),
	}
}

func (s *Service) BootstrapOwner(ctx context.Context, input BootstrapOwnerInput) (Admin, error) {
	input.Email = normalizeEmail(input.Email)
	input.DisplayName = strings.TrimSpace(input.DisplayName)

	if err := validateAdminInput(input.Email, input.Password, input.DisplayName); err != nil {
		return Admin{}, err
	}

	if input.Email != s.ownerEmail {
		return Admin{}, ErrOwnerNotAllowed
	}

	owners, err := s.repository.CountOwners(ctx)
	if err != nil {
		return Admin{}, err
	}

	if owners > 0 {
		return Admin{}, ErrOwnerExists
	}

	passwordHash, err := hashPassword(input.Password)
	if err != nil {
		return Admin{}, err
	}

	return s.repository.CreateOwner(ctx, input.Email, passwordHash, input.DisplayName)
}

func (s *Service) Login(ctx context.Context, input LoginInput) (Admin, error) {
	email := normalizeEmail(input.Email)

	if email == "" || strings.TrimSpace(input.Password) == "" {
		return Admin{}, ErrInvalidCredentials
	}

	a, err := s.repository.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, ErrAdminNotFound) {
			return Admin{}, ErrInvalidCredentials
		}

		return Admin{}, err
	}

	if !a.IsActive {
		return Admin{}, ErrInactiveAdmin
	}

	if err := bcrypt.CompareHashAndPassword([]byte(a.PasswordHash), []byte(input.Password)); err != nil {
		return Admin{}, ErrInvalidCredentials
	}

	if err := s.repository.TouchLastLogin(ctx, a.ID); err != nil {
		return Admin{}, err
	}

	return a.Admin, nil
}

func (s *Service) CreateAdmin(ctx context.Context, actor Admin, input CreateAdminInput) (Admin, error) {
	if actor.Role != RoleOwner {
		return Admin{}, ErrForbidden
	}

	input.Email = normalizeEmail(input.Email)
	input.DisplayName = strings.TrimSpace(input.DisplayName)

	if err := validateAdminInput(input.Email, input.Password, input.DisplayName); err != nil {
		return Admin{}, err
	}

	passwordHash, err := hashPassword(input.Password)
	if err != nil {
		return Admin{}, err
	}

	return s.repository.CreateAdmin(ctx, input.Email, passwordHash, input.DisplayName, actor.ID)
}

func (s *Service) ListAdmins(ctx context.Context, actor Admin) ([]Admin, error) {
	if actor.Role != RoleOwner {
		return nil, ErrForbidden
	}

	return s.repository.List(ctx)
}

func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (Admin, error) {
	return s.repository.GetByID(ctx, id)
}

func validateAdminInput(email string, password string, displayName string) error {
	if email == "" {
		return ErrEmailRequired
	}

	if _, err := mail.ParseAddress(email); err != nil {
		return ErrInvalidEmail
	}

	if strings.TrimSpace(password) == "" {
		return ErrPasswordRequired
	}

	if len(password) < 8 {
		return ErrPasswordTooShort
	}

	if displayName == "" {
		return ErrDisplayNameRequired
	}

	return nil
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(hash), nil
}

func (s *Service) UpdateAdminStatus(
	ctx context.Context,
	actor Admin,
	adminID uuid.UUID,
	input UpdateAdminStatusInput,
) (Admin, error) {
	if actor.Role != RoleOwner {
		return Admin{}, ErrForbidden
	}

	if actor.ID == adminID && !input.IsActive {
		return Admin{}, ErrCannotDeactivateSelf
	}

	targetAdmin, err := s.repository.GetByID(ctx, adminID)
	if err != nil {
		return Admin{}, err
	}

	if targetAdmin.Role == RoleOwner && !input.IsActive {
		return Admin{}, ErrCannotDeactivateSelf
	}

	return s.repository.UpdateStatus(ctx, adminID, input.IsActive)
}
