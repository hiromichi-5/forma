package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/hiromichi-5/forma/backend/internal/entity"
	"github.com/hiromichi-5/forma/backend/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthUseCase_Signup(t *testing.T) {
	t.Run("正常系: ユーザーを新規登録できること", func(t *testing.T) {
		truncate(t)
		uc := newAuthUseCase()
		ctx := context.Background()

		userID, err := uc.Signup(ctx, "test@example.com", "password123", "Test User")
		require.NoError(t, err)
		assert.NotEmpty(t, userID)
	})

	t.Run("正常系: 未認証ユーザーが再度signupするとトークンが再発行されること", func(t *testing.T) {
		truncate(t)
		uc := newAuthUseCase()
		ctx := context.Background()

		userID, err := uc.Signup(ctx, "dup@example.com", "password123", "User1")
		require.NoError(t, err)

		before := testutil.GetEmailVerificationToken(t, ctx, testPool, userID)

		userID2, err := uc.Signup(ctx, "dup@example.com", "password123", "User2")
		require.NoError(t, err)
		assert.Equal(t, userID, userID2)

		after := testutil.GetEmailVerificationToken(t, ctx, testPool, userID)
		assert.NotEqual(t, before, after)
	})

	t.Run("準正常系: 認証済みユーザーの重複メールアドレスで CONFLICT エラーになること", func(t *testing.T) {
		truncate(t)
		uc := newAuthUseCase()
		ctx := context.Background()

		testutil.CreateVerifiedUser(
			t,
			ctx,
			testPool,
			"verified@example.com",
			"password123",
			"User1",
		)

		_, err := uc.Signup(ctx, "verified@example.com", "password123", "User2")
		require.Error(t, err)
		var appErr *entity.Error
		require.True(t, errors.As(err, &appErr))
		assert.Equal(t, entity.CodeConflict, appErr.Code)
	})
}

func TestAuthUseCase_Authenticate(t *testing.T) {
	t.Run("正常系: 認証済みユーザーでログインできること", func(t *testing.T) {
		truncate(t)
		uc := newAuthUseCase()
		ctx := context.Background()

		testutil.CreateVerifiedUser(
			t,
			ctx,
			testPool,
			"auth@example.com",
			"password123",
			"Auth User",
		)

		session, err := uc.Authenticate(ctx, "auth@example.com", "password123")
		require.NoError(t, err)
		assert.NotEmpty(t, session.ID)
	})

	t.Run("準正常系: パスワードが間違っている場合 INVALID_CREDENTIALS エラーになること", func(t *testing.T) {
		truncate(t)
		uc := newAuthUseCase()
		ctx := context.Background()

		testutil.CreateVerifiedUser(
			t,
			ctx,
			testPool,
			"auth@example.com",
			"password123",
			"Auth User",
		)

		_, err := uc.Authenticate(ctx, "auth@example.com", "wrongpassword")
		require.Error(t, err)
		var appErr *entity.Error
		require.True(t, errors.As(err, &appErr))
		assert.Equal(t, entity.CodeInvalidCredentials, appErr.Code)
	})

	t.Run("準正常系: 存在しないメールアドレスで INVALID_CREDENTIALS エラーになること", func(t *testing.T) {
		truncate(t)
		uc := newAuthUseCase()
		ctx := context.Background()

		_, err := uc.Authenticate(ctx, "nobody@example.com", "password123")
		require.Error(t, err)
		var appErr *entity.Error
		require.True(t, errors.As(err, &appErr))
		assert.Equal(t, entity.CodeInvalidCredentials, appErr.Code)
	})

	t.Run("準正常系: メール未認証ユーザーで EMAIL_NOT_VERIFIED エラーになること", func(t *testing.T) {
		truncate(t)
		uc := newAuthUseCase()
		ctx := context.Background()

		_, err := uc.Signup(ctx, "unverified@example.com", "password123", "Unverified")
		require.NoError(t, err)

		_, err = uc.Authenticate(ctx, "unverified@example.com", "password123")
		require.Error(t, err)
		var appErr *entity.Error
		require.True(t, errors.As(err, &appErr))
		assert.Equal(t, entity.CodeEmailNotVerified, appErr.Code)
	})
}

func TestAuthUseCase_Logout(t *testing.T) {
	t.Run("正常系: セッションを削除できること", func(t *testing.T) {
		truncate(t)
		uc := newAuthUseCase()
		ctx := context.Background()

		testutil.CreateVerifiedUser(t, ctx, testPool, "logout@example.com", "password123", "User")
		session, err := uc.Authenticate(ctx, "logout@example.com", "password123")
		require.NoError(t, err)

		err = uc.Logout(ctx, session.ID)
		require.NoError(t, err)
	})

	t.Run("準正常系: 存在しないセッションで INVALID_SESSION エラーになること", func(t *testing.T) {
		truncate(t)
		uc := newAuthUseCase()
		ctx := context.Background()

		err := uc.Logout(ctx, testutil.RandomUUID())
		require.Error(t, err)
		var appErr *entity.Error
		require.True(t, errors.As(err, &appErr))
		assert.Equal(t, entity.CodeInvalidSession, appErr.Code)
	})
}

func TestAuthUseCase_VerifyEmail(t *testing.T) {
	t.Run("正常系: メール認証トークンで認証できること", func(t *testing.T) {
		truncate(t)
		uc := newAuthUseCase()
		ctx := context.Background()

		userID, err := uc.Signup(ctx, "verify@example.com", "password123", "Verify")
		require.NoError(t, err)

		token := testutil.GetEmailVerificationToken(t, ctx, testPool, userID)
		err = uc.VerifyEmail(ctx, token)
		require.NoError(t, err)

		// 認証後にログインできることを確認
		session, err := uc.Authenticate(ctx, "verify@example.com", "password123")
		require.NoError(t, err)
		assert.NotEmpty(t, session.ID)
	})

	t.Run("準正常系: 無効なトークンで TOKEN_NOT_FOUND エラーになること", func(t *testing.T) {
		truncate(t)
		uc := newAuthUseCase()
		ctx := context.Background()

		err := uc.VerifyEmail(ctx, "invalid-token")
		require.Error(t, err)
		var appErr *entity.Error
		require.True(t, errors.As(err, &appErr))
		assert.Equal(t, entity.CodeTokenNotFound, appErr.Code)
	})
}

func TestAuthUseCase_ConfirmPasswordReset(t *testing.T) {
	t.Run("正常系: パスワードリセットが完了すること", func(t *testing.T) {
		truncate(t)
		uc := newAuthUseCase()
		ctx := context.Background()

		userID := testutil.CreateVerifiedUser(
			t,
			ctx,
			testPool,
			"reset@example.com",
			"oldpass123",
			"User",
		)

		err := uc.RequestPasswordReset(ctx, "reset@example.com")
		require.NoError(t, err)

		token := testutil.GetPasswordResetToken(t, ctx, testPool, userID)
		err = uc.ConfirmPasswordReset(ctx, token, "newpass123")
		require.NoError(t, err)

		// 新しいパスワードでログインできることを確認
		session, err := uc.Authenticate(ctx, "reset@example.com", "newpass123")
		require.NoError(t, err)
		assert.NotEmpty(t, session.ID)
	})

	t.Run("準正常系: 無効なトークンで TOKEN_NOT_FOUND エラーになること", func(t *testing.T) {
		truncate(t)
		uc := newAuthUseCase()
		ctx := context.Background()

		err := uc.ConfirmPasswordReset(ctx, "invalid-token", "newpass123")
		require.Error(t, err)
		var appErr *entity.Error
		require.True(t, errors.As(err, &appErr))
		assert.Equal(t, entity.CodeTokenNotFound, appErr.Code)
	})
}

func TestAuthUseCase_RequestPasswordReset(t *testing.T) {
	t.Run("準正常系: 存在しないメールアドレスでもエラーにならないこと", func(t *testing.T) {
		truncate(t)
		uc := newAuthUseCase()
		ctx := context.Background()

		err := uc.RequestPasswordReset(ctx, "nobody@example.com")
		require.NoError(t, err)
	})
}

func TestAuthUseCase_ResendEmailVerification(t *testing.T) {
	t.Run("正常系: 未認証ユーザーに再送できること", func(t *testing.T) {
		truncate(t)
		uc := newAuthUseCase()
		ctx := context.Background()

		userID, err := uc.Signup(ctx, "resend@example.com", "password123", "Resend")
		require.NoError(t, err)

		before := testutil.GetEmailVerificationToken(t, ctx, testPool, userID)
		err = uc.ResendEmailVerification(ctx, "resend@example.com")
		require.NoError(t, err)

		after := testutil.GetEmailVerificationToken(t, ctx, testPool, userID)
		assert.NotEmpty(t, after)
		assert.NotEqual(t, before, after)
	})

	t.Run("準正常系: 認証済みユーザーでもエラーにならないこと", func(t *testing.T) {
		truncate(t)
		uc := newAuthUseCase()
		ctx := context.Background()

		testutil.CreateVerifiedUser(t, ctx, testPool, "verified@example.com", "password123", "User")

		err := uc.ResendEmailVerification(ctx, "verified@example.com")
		require.NoError(t, err)
	})
}
