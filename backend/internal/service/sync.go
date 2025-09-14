package service

import (
	"context"
	"encoding/json"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"google.golang.org/api/forms/v1"

	"github.com/hiromichi-5/forma/backend/internal/db"
)

func (s *Service) roleFor(ctx context.Context, formID string, actor uuid.UUID) (string, error) {
	r, err := s.Q.GetUserFormRole(ctx, db.GetUserFormRoleParams{
		UserID: pgtype.UUID{Bytes: actor, Valid: true},
		FormID: formID,
	})
	if err != nil {
		return "", ErrForbidden
	}
	return r, nil
}

func (s *Service) SyncFormOnce(ctx context.Context, formID string, actor uuid.UUID) (synced int, newTickets int, last time.Time, err error) {
	role, err := s.roleFor(ctx, formID, actor)
	if err != nil || (role != "admin" && role != "editor") {
		return 0, 0, time.Time{}, ErrForbidden
	}

	// カーソル決定 - 既存のsync_cursorを取得、なければ7日前を使用
	var cursor time.Time
	syncCursor, err := s.Q.GetFormSyncCursor(ctx, formID)
	if err != nil || !syncCursor.Valid {
		cursor = time.Now().Add(-7 * 24 * time.Hour)
	} else {
		cursor = syncCursor.Time
	}

	formattedCursor := cursor.UTC().Format(time.RFC3339)
	// Validate the formatted timestamp to ensure it matches RFC3339
	if _, err := time.Parse(time.RFC3339, formattedCursor); err != nil {
		return 0, 0, time.Time{}, ErrForbidden
	}
	filter := "timestamp >= " + formattedCursor
	var all []*forms.FormResponse
	token := ""
	for {
		page, e := s.GF.ListResponses(ctx, formID, filter, token)
		if e != nil {
			err = e
			return
		}
		if page.Responses != nil {
			all = append(all, page.Responses...)
		}
		if page.NextPageToken == "" {
			break
		}
		token = page.NextPageToken
	}

	// submittedTime昇順で処理
	type pair struct {
		submitted time.Time
		r         *forms.FormResponse
	}
	ps := make([]pair, 0, len(all))
	for _, r := range all {
		if r.CreateTime == "" && r.LastSubmittedTime == "" {
			continue
		}
		// submitted_atはLastSubmittedTime優先
		ts := r.LastSubmittedTime
		if ts == "" {
			ts = r.CreateTime
		}
		t, e := time.Parse(time.RFC3339, ts)
		if e != nil {
			continue
		}
		ps = append(ps, pair{submitted: t, r: r})
	}
	sort.Slice(ps, func(i, j int) bool { return ps[i].submitted.Before(ps[j].submitted) })

	var maxSubmitted time.Time
	for _, p := range ps {
		if p.submitted.After(maxSubmitted) {
			maxSubmitted = p.submitted
		}

		// 代表キー: responseId
		rid := p.r.ResponseId
		if rid == "" {
			continue
		}

		// payloadはそのまま保持
		payloadData := map[string]any{"answers": p.r.Answers}
		payloadBytes, err := json.Marshal(payloadData)
		if err != nil {
			continue
		}

		// responses挿入（重複は無視）
		rowsAffected, e := s.Q.InsertResponse(ctx, db.InsertResponseParams{
			ResponseID:    rid,
			FormID:        formID,
			SubmittedAt:   pgtype.Timestamptz{Time: p.submitted, Valid: true},
			Payload:       payloadBytes,
			SchemaVersion: 1,
		})
		if e != nil {
			continue
		}

		// 新規挿入された場合のみ
		if rowsAffected > 0 {
			synced++
			// 新規のみチケット作成
			ticketID := uuid.New()
			_, e = s.Q.CreateTicket(ctx, db.CreateTicketParams{
				ID:         pgtype.UUID{Bytes: ticketID, Valid: true},
				FormID:     formID,
				ResponseID: rid,
			})
			if e == nil {
				newTickets++
			}
		}
	}

	if !maxSubmitted.IsZero() {
		_ = s.Q.UpdateSyncCursor(ctx, db.UpdateSyncCursorParams{
			FormID:     formID,
			SyncCursor: pgtype.Timestamptz{Time: maxSubmitted, Valid: true},
		})
	}

	return synced, newTickets, maxSubmitted, nil
}
