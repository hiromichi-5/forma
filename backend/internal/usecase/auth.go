package usecase

import (
	"context"
	"crypto/rand"
	"encoding/base32"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/hiromichi-5/forma/backend/internal/entity"
	"github.com/hiromichi-5/forma/backend/internal/logger"
	"github.com/hiromichi-5/forma/backend/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

const tokenTTL = 24 * time.Hour

type AuthUseCase struct {
	userRepo        repository.UserRepository
	uow             repository.UnitOfWork[repository.AuthRepos]
	emailSender     repository.EmailSender
	frontendBaseURL string
	now             func() time.Time
	generateToken   func() (string, error)
}

func NewAuthUseCase(
	userRepo repository.UserRepository,
	uow repository.UnitOfWork[repository.AuthRepos],
	emailSender repository.EmailSender,
	frontendBaseURL string,
) *AuthUseCase {
	return &AuthUseCase{
		userRepo:        userRepo,
		uow:             uow,
		emailSender:     emailSender,
		frontendBaseURL: frontendBaseURL,
		now:             time.Now,
		generateToken:   defaultToken,
	}
}

func defaultToken() (string, error) {
	buf := make([]byte, 24)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	token := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(buf)
	return strings.ToLower(token), nil
}

func (uc *AuthUseCase) Authenticate(
	ctx context.Context,
	email, password string,
) (entity.Session, error) {
	if email == "" || password == "" {
		return entity.Session{}, entity.NewError(entity.CodeValidation)
	}

	user, err := uc.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return entity.Session{}, entity.NewError(entity.CodeInvalidCredentials)
		}
		return entity.Session{}, err
	}

	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) != nil {
		return entity.Session{}, entity.NewError(entity.CodeInvalidCredentials)
	}

	if user.VerifiedAt == nil {
		return entity.Session{}, entity.NewError(entity.CodeEmailNotVerified)
	}

	session, err := uc.userRepo.CreateSession(ctx, entity.Session{
		ID:     uuid.New(),
		UserID: user.ID,
	})
	if err != nil {
		return entity.Session{}, err
	}

	logger.From(ctx).Info("user authenticated", "user_id", user.ID.String())

	return session, nil
}

func (uc *AuthUseCase) Signup(
	ctx context.Context,
	email, password, displayName string,
) (uuid.UUID, error) {
	if email == "" || password == "" || displayName == "" {
		return uuid.UUID{}, entity.NewError(entity.CodeValidation)
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return uuid.UUID{}, err
	}

	// 既存ユーザーが未認証の場合はトークンを再発行する
	existing, err := uc.userRepo.GetByEmail(ctx, email)
	if err == nil {
		if existing.VerifiedAt != nil {
			return uuid.UUID{}, entity.NewError(entity.CodeConflict)
		}

		var tokenStr string
		if err := uc.uow.Do(ctx, func(repos repository.AuthRepos) error {
			if err := repos.User.DeleteEmailVerificationTokensByUser(ctx, existing.ID); err != nil {
				return err
			}
			tokenStr, err = uc.issueEmailVerificationTokenTx(ctx, repos.User, existing.ID)
			return err
		}); err != nil {
			return uuid.UUID{}, err
		}

		if err := uc.sendEmailVerification(ctx, email, tokenStr); err != nil {
			return uuid.UUID{}, err
		}

		logger.From(ctx).Info("user signed up (resend)", "user_id", existing.ID.String())

		return existing.ID, nil
	}
	if !errors.Is(err, repository.ErrNotFound) {
		return uuid.UUID{}, err
	}

	var (
		createdUserID uuid.UUID
		tokenStr      string
	)
	if err := uc.uow.Do(ctx, func(repos repository.AuthRepos) error {
		user, err := repos.User.Create(ctx, entity.User{
			ID:           uuid.New(),
			Email:        email,
			PasswordHash: string(hashed),
			DisplayName:  displayName,
		})
		if err != nil {
			if errors.Is(err, repository.ErrConflict) {
				return entity.NewError(entity.CodeConflict)
			}
			return err
		}
		createdUserID = user.ID

		tokenStr, err = uc.issueEmailVerificationTokenTx(ctx, repos.User, user.ID)
		return err
	}); err != nil {
		return uuid.UUID{}, err
	}

	if err := uc.sendEmailVerification(ctx, email, tokenStr); err != nil {
		return uuid.UUID{}, err
	}

	logger.From(ctx).Info("user signed up", "user_id", createdUserID.String())

	return createdUserID, nil
}

func (uc *AuthUseCase) Logout(ctx context.Context, sessionID uuid.UUID) error {
	err := uc.userRepo.DeleteSession(ctx, sessionID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return entity.NewError(entity.CodeInvalidSession)
		}
		return err
	}
	return nil
}

func (uc *AuthUseCase) VerifyEmail(ctx context.Context, token string) error {
	if token == "" {
		return entity.NewError(entity.CodeValidation)
	}

	t, err := uc.userRepo.GetEmailVerificationTokenByToken(ctx, token)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return entity.NewError(entity.CodeTokenNotFound)
		}
		return err
	}

	return uc.uow.Do(ctx, func(repos repository.AuthRepos) error {
		if err := repos.User.UseEmailVerificationToken(ctx, t.ID); err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return entity.NewError(entity.CodeTokenNotFound)
			}
			return err
		}
		if err := repos.User.SetVerifiedAt(ctx, t.UserID, uc.now()); err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return entity.NewError(entity.CodeUserNotFound)
			}
			return err
		}
		return nil
	})
}

func (uc *AuthUseCase) ResendEmailVerification(ctx context.Context, email string) error {
	if email == "" {
		return entity.NewError(entity.CodeValidation)
	}

	user, err := uc.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil
		}
		return err
	}

	if user.VerifiedAt != nil {
		return nil
	}

	var tokenStr string
	if err := uc.uow.Do(ctx, func(repos repository.AuthRepos) error {
		if err := repos.User.DeleteEmailVerificationTokensByUser(ctx, user.ID); err != nil {
			return err
		}
		tokenStr, err = uc.issueEmailVerificationTokenTx(ctx, repos.User, user.ID)
		return err
	}); err != nil {
		return err
	}

	return uc.sendEmailVerification(ctx, email, tokenStr)
}

func (uc *AuthUseCase) RequestPasswordReset(ctx context.Context, email string) error {
	if email == "" {
		return entity.NewError(entity.CodeValidation)
	}

	user, err := uc.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil
		}
		return err
	}

	var tokenStr string
	if err := uc.uow.Do(ctx, func(repos repository.AuthRepos) error {
		if err := repos.User.DeletePasswordResetTokensByUser(ctx, user.ID); err != nil {
			return err
		}
		tokenStr, err = uc.issuePasswordResetTokenTx(ctx, repos.User, user.ID)
		return err
	}); err != nil {
		return err
	}

	resetURL := uc.frontendBaseURL + "/password-reset/confirm?token=" + tokenStr
	return uc.emailSender.SendEmail(ctx, repository.SendEmailInput{
		To:           []string{email},
		TemplateName: repository.TemplatePasswordReset,
		TemplateData: map[string]string{"reset_url": resetURL},
	})
}

func (uc *AuthUseCase) ConfirmPasswordReset(ctx context.Context, token, newPassword string) error {
	if token == "" || newPassword == "" {
		return entity.NewError(entity.CodeValidation)
	}
	if len(newPassword) < 8 {
		return entity.NewError(entity.CodeValidation)
	}

	t, err := uc.userRepo.GetPasswordResetTokenByToken(ctx, token)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return entity.NewError(entity.CodeTokenNotFound)
		}
		return err
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	return uc.uow.Do(ctx, func(repos repository.AuthRepos) error {
		if err := repos.User.UsePasswordResetToken(ctx, t.ID); err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return entity.NewError(entity.CodeTokenNotFound)
			}
			return err
		}
		if err := repos.User.UpdatePasswordHash(ctx, t.UserID, string(hashed)); err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return entity.NewError(entity.CodeUserNotFound)
			}
			return err
		}
		return repos.User.DeletePasswordResetTokensByUser(ctx, t.UserID)
	})
}

func (uc *AuthUseCase) sendEmailVerification(ctx context.Context, email, token string) error {
	verifyURL := uc.frontendBaseURL + "/verify-email?token=" + token
	return uc.emailSender.SendEmail(ctx, repository.SendEmailInput{
		To:           []string{email},
		TemplateName: repository.TemplateEmailVerification,
		TemplateData: map[string]string{"verify_url": verifyURL},
	})
}

func (uc *AuthUseCase) issueEmailVerificationTokenTx(
	ctx context.Context,
	userRepo repository.UserRepository,
	userID uuid.UUID,
) (string, error) {
	token, err := uc.generateToken()
	if err != nil {
		return "", err
	}
	_, err = userRepo.CreateEmailVerificationToken(ctx, entity.EmailVerificationToken{
		ID:        uuid.New(),
		UserID:    userID,
		Token:     token,
		ExpiresAt: uc.now().Add(tokenTTL),
	})
	if err != nil {
		return "", err
	}
	return token, nil
}

func (uc *AuthUseCase) issuePasswordResetTokenTx(
	ctx context.Context,
	userRepo repository.UserRepository,
	userID uuid.UUID,
) (string, error) {
	token, err := uc.generateToken()
	if err != nil {
		return "", err
	}
	_, err = userRepo.CreatePasswordResetToken(ctx, entity.PasswordResetToken{
		ID:        uuid.New(),
		UserID:    userID,
		Token:     token,
		ExpiresAt: uc.now().Add(tokenTTL),
	})
	if err != nil {
		return "", err
	}
	return token, nil
}
