package usecase

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/hiromichi-5/forma/backend/internal/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApplyStatusUpdate(t *testing.T) {
	t.Parallel()

	strp := func(s string) *string { return &s }
	int32p := func(i int32) *int32 { return &i }
	color := "#E53935"

	base := entity.FormStatus{
		ID:           uuid.New(),
		FormID:       uuid.New(),
		Name:         "未対応",
		Color:        &color,
		DisplayOrder: 1,
		IsDefault:    false,
	}

	tests := []struct {
		name        string
		current     entity.FormStatus
		nameArg     *string
		colorArg    *string
		orderArg    *int32
		wantName    string
		wantColor   *string
		wantOrder   int32
		wantErr     bool
		wantErrCode entity.Code
	}{
		{
			name:      "すべてnilなら変更なし",
			current:   base,
			wantName:  base.Name,
			wantColor: base.Color,
			wantOrder: base.DisplayOrder,
		},
		{
			name:      "名前を更新できること",
			current:   base,
			nameArg:   strp("対応中"),
			wantName:  "対応中",
			wantColor: base.Color,
			wantOrder: base.DisplayOrder,
		},
		{
			name:        "空文字名はバリデーションエラー",
			current:     base,
			nameArg:     strp(""),
			wantErr:     true,
			wantErrCode: entity.CodeValidation,
		},
		{
			name:        "空白のみの名前はバリデーションエラー",
			current:     base,
			nameArg:     strp("   "),
			wantErr:     true,
			wantErrCode: entity.CodeValidation,
		},
		{
			name:      "色を更新できること",
			current:   base,
			colorArg:  strp("#43A047"),
			wantName:  base.Name,
			wantColor: strp("#43A047"),
			wantOrder: base.DisplayOrder,
		},
		{
			name:      "空文字色でnilにリセットできること",
			current:   base,
			colorArg:  strp(""),
			wantName:  base.Name,
			wantColor: nil,
			wantOrder: base.DisplayOrder,
		},
		{
			name:      "displayOrderを更新できること",
			current:   base,
			orderArg:  int32p(5),
			wantName:  base.Name,
			wantColor: base.Color,
			wantOrder: 5,
		},
		{
			name:        "displayOrder 0 はバリデーションエラー",
			current:     base,
			orderArg:    int32p(0),
			wantErr:     true,
			wantErrCode: entity.CodeValidation,
		},
		{
			name:        "displayOrder 負数はバリデーションエラー",
			current:     base,
			orderArg:    int32p(-1),
			wantErr:     true,
			wantErrCode: entity.CodeValidation,
		},
		{
			name:      "displayOrder 1に更新できること",
			current:   base,
			orderArg:  int32p(1),
			wantName:  base.Name,
			wantColor: base.Color,
			wantOrder: 1,
		},
		{
			name:      "すべてのフィールドを同時に更新できること",
			current:   base,
			nameArg:   strp("完了"),
			colorArg:  strp("#00BCD4"),
			orderArg:  int32p(3),
			wantName:  "完了",
			wantColor: strp("#00BCD4"),
			wantOrder: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := applyStatusUpdate(tt.current, tt.nameArg, tt.colorArg, tt.orderArg)
			if tt.wantErr {
				require.Error(t, err)
				var appErr *entity.Error
				require.True(t, errors.As(err, &appErr))
				assert.Equal(t, tt.wantErrCode, appErr.Code)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantName, got.Name)
			assert.Equal(t, tt.wantColor, got.Color)
			assert.Equal(t, tt.wantOrder, got.DisplayOrder)
		})
	}
}
