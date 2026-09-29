package usecase

import (
	"testing"

	"github.com/google/uuid"
	"github.com/hiromichi-5/forma/backend/internal/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJoinValues(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		values []string
		want   string
	}{
		{"複数値はカンマ区切り", []string{"a", "b", "c"}, "a, b, c"},
		{"単一値はそのまま", []string{"hello"}, "hello"},
		{"空スライスは空文字", []string{}, ""},
		{"nilは空文字", nil, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, joinValues(tt.values))
		})
	}
}

func TestNewFormQuestionSet(t *testing.T) {
	t.Parallel()

	t.Run("空リストならdefaultTitleIDは空文字", func(t *testing.T) {
		t.Parallel()
		set := newFormQuestionSet(nil)
		assert.Empty(t, set.defaultTitleID)
		assert.Empty(t, set.ordered)
	})

	t.Run("先頭のQuestionIDがdefaultTitleIDになること", func(t *testing.T) {
		t.Parallel()
		qs := []entity.FormQuestion{
			{QuestionID: "q1", Title: "名前"},
			{QuestionID: "q2", Title: "メール"},
		}
		set := newFormQuestionSet(qs)
		assert.Equal(t, "q1", set.defaultTitleID)
		assert.Len(t, set.ordered, 2)
	})
}

func TestExtractTextValues(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		ans  storedAnswer
		want []string
	}{
		{
			name: "TextAnswersがnilなら空スライス",
			ans:  storedAnswer{},
			want: []string{},
		},
		{
			name: "値を抽出できること",
			ans: storedAnswer{
				TextAnswers: &storedTextBlock{
					Answers: []storedTextAnswer{
						{Value: "回答1"},
						{Value: "回答2"},
					},
				},
			},
			want: []string{"回答1", "回答2"},
		},
		{
			name: "空のAnswersなら空スライス",
			ans: storedAnswer{
				TextAnswers: &storedTextBlock{Answers: []storedTextAnswer{}},
			},
			want: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, extractTextValues(tt.ans))
		})
	}
}

func TestParseResponseAnswers(t *testing.T) {
	t.Parallel()

	t.Run("空バイトは空マップを返すこと", func(t *testing.T) {
		t.Parallel()
		got, err := parseResponseAnswers(nil)
		require.NoError(t, err)
		assert.Empty(t, got)
	})

	t.Run("空バイト列は空マップを返すこと", func(t *testing.T) {
		t.Parallel()
		got, err := parseResponseAnswers([]byte{})
		require.NoError(t, err)
		assert.Empty(t, got)
	})

	t.Run("Google Forms answers map形式をパースできること", func(t *testing.T) {
		t.Parallel()
		payload := []byte(`{
			"q1": {
				"questionId": "q1",
				"textAnswers": {
					"answers": [{"value": "回答A"}]
				}
			}
		}`)
		got, err := parseResponseAnswers(payload)
		require.NoError(t, err)
		assert.Equal(t, []string{"回答A"}, got["q1"])
	})

	t.Run("不正なJSONはエラーを返すこと", func(t *testing.T) {
		t.Parallel()
		_, err := parseResponseAnswers([]byte(`{invalid`))
		require.Error(t, err)
	})
}

func TestDeriveTitle(t *testing.T) {
	t.Parallel()

	questions := formQuestionSet{
		ordered: []entity.FormQuestion{
			{QuestionID: "q1", Title: "名前"},
			{QuestionID: "q2", Title: "メール"},
		},
		defaultTitleID: "q1",
	}

	t.Run("titleQIDの回答があればそれを返すこと", func(t *testing.T) {
		t.Parallel()
		answers := map[string][]string{
			"q1": {"田中太郎"},
			"q2": {"taro@example.com"},
		}
		got := deriveTitle("q1", answers, questions, "フォーム名", "resp-001")
		assert.Equal(t, "田中太郎", got)
	})

	t.Run("titleQIDの回答が空なら他のquestionにフォールバックすること", func(t *testing.T) {
		t.Parallel()
		answers := map[string][]string{
			"q1": {},
			"q2": {"taro@example.com"},
		}
		got := deriveTitle("q1", answers, questions, "フォーム名", "resp-001")
		assert.Equal(t, "taro@example.com", got)
	})

	t.Run("orderedにない回答にフォールバックすること", func(t *testing.T) {
		t.Parallel()
		answers := map[string][]string{
			"q1":    {},
			"q2":    {},
			"extra": {"追加回答"},
		}
		got := deriveTitle("q1", answers, questions, "フォーム名", "resp-001")
		assert.Equal(t, "追加回答", got)
	})

	t.Run("回答がすべて空ならformTitleにフォールバックすること", func(t *testing.T) {
		t.Parallel()
		got := deriveTitle("q1", map[string][]string{}, questions, "フォーム名", "resp-001")
		assert.Equal(t, "フォーム名", got)
	})

	t.Run("formTitleも空ならresponseIDにフォールバックすること", func(t *testing.T) {
		t.Parallel()
		got := deriveTitle("q1", map[string][]string{}, questions, "", "resp-001")
		assert.Equal(t, "resp-001", got)
	})
}

func TestBuildAssignee(t *testing.T) {
	t.Parallel()

	memberID := uuid.New()
	members := map[uuid.UUID]entity.Member{
		memberID: {
			UserRef: entity.UserRef{
				ID:          memberID,
				Email:       "user@example.com",
				DisplayName: "テストユーザー",
			},
			Role: entity.RoleEditor,
		},
	}

	t.Run("assigneeIDがnilならnilを返すこと", func(t *testing.T) {
		t.Parallel()
		got := buildAssignee(nil, members)
		assert.Nil(t, got)
	})

	t.Run("メンバーに存在しないIDならnilを返すこと", func(t *testing.T) {
		t.Parallel()
		unknownID := uuid.New()
		got := buildAssignee(&unknownID, members)
		assert.Nil(t, got)
	})

	t.Run("メンバーが見つかればAssigneeを返すこと", func(t *testing.T) {
		t.Parallel()
		got := buildAssignee(&memberID, members)
		require.NotNil(t, got)
		assert.Equal(t, memberID, got.ID)
		assert.Equal(t, "テストユーザー", got.DisplayName)
		assert.Equal(t, "user@example.com", got.Email)
	})
}

func TestBuildAnswerList(t *testing.T) {
	t.Parallel()

	t.Run("questionsの順序で回答リストが構築されること", func(t *testing.T) {
		t.Parallel()
		questions := formQuestionSet{
			ordered: []entity.FormQuestion{
				{QuestionID: "q1", Title: "名前", QuestionType: "TEXT"},
				{QuestionID: "q2", Title: "メール", QuestionType: "TEXT"},
			},
		}
		answers := map[string][]string{
			"q1": {"田中"},
			"q2": {"tanaka@example.com"},
		}
		got := buildAnswerList(answers, questions)
		require.Len(t, got, 2)
		assert.Equal(t, "q1", got[0].QuestionID)
		assert.Equal(t, "名前", got[0].QuestionTitle)
		assert.Equal(t, []string{"田中"}, got[0].Values)
		assert.Equal(t, "q2", got[1].QuestionID)
	})

	t.Run("questionsにない回答がアルファベット順で末尾に追加されること", func(t *testing.T) {
		t.Parallel()
		questions := formQuestionSet{
			ordered: []entity.FormQuestion{
				{QuestionID: "q1", Title: "名前", QuestionType: "TEXT"},
			},
		}
		answers := map[string][]string{
			"q1":     {"田中"},
			"extra2": {"追加B"},
			"extra1": {"追加A"},
		}
		got := buildAnswerList(answers, questions)
		require.Len(t, got, 3)
		assert.Equal(t, "q1", got[0].QuestionID)
		assert.Equal(t, "extra1", got[1].QuestionID)
		assert.Equal(t, "unknown", got[1].QuestionType)
		assert.Equal(t, "extra2", got[2].QuestionID)
	})

	t.Run("回答がないquestionは空スライスで含まれること", func(t *testing.T) {
		t.Parallel()
		questions := formQuestionSet{
			ordered: []entity.FormQuestion{
				{QuestionID: "q1", Title: "名前", QuestionType: "TEXT"},
			},
		}
		got := buildAnswerList(map[string][]string{}, questions)
		require.Len(t, got, 1)
		assert.Equal(t, []string{}, got[0].Values)
	})
}
