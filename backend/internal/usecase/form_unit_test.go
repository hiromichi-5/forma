package usecase

import (
	"errors"
	"testing"

	"github.com/hiromichi-5/forma/backend/internal/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtractFormID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:  "Google Forms URL からIDを抽出できること",
			input: "https://docs.google.com/forms/d/1FAIpQLSfABC123_xyz-456/viewform",
			want:  "1FAIpQLSfABC123_xyz-456",
		},
		{
			name:  "末尾スラッシュ付きURLからIDを抽出できること",
			input: "https://docs.google.com/forms/d/1FAIpQLSfABC123_xyz-456/viewform?usp=sf_link",
			want:  "1FAIpQLSfABC123_xyz-456",
		},
		{
			name:  "20文字以上のID文字列をそのまま返すこと",
			input: "1FAIpQLSfABC123_xyz-456",
			want:  "1FAIpQLSfABC123_xyz-456",
		},
		{
			name:    "20文字未満のスラッシュなし文字列はバリデーションエラー",
			input:   "short",
			wantErr: true,
		},
		{
			name:    "空文字列はバリデーションエラー",
			input:   "",
			wantErr: true,
		},
		{
			name:    "パス形式だがフォームIDを含まないURLはバリデーションエラー",
			input:   "https://example.com/other/path",
			wantErr: true,
		},
		{
			name:  "ちょうど20文字のIDを受け付けること",
			input: "12345678901234567890",
			want:  "12345678901234567890",
		},
		{
			name:    "19文字のIDはバリデーションエラー",
			input:   "1234567890123456789",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := extractFormID(tt.input)
			if tt.wantErr {
				require.Error(t, err)
				var appErr *entity.Error
				require.True(t, errors.As(err, &appErr))
				assert.Equal(t, entity.CodeValidation, appErr.Code)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
