package repository

import "errors"

// ErrNotFound はリソースが存在しない、または操作対象の行がなかった場合に返される。
// infra 層が pgx.ErrNoRows や RowsAffected() == 0 を検知してこのエラーに変換する。
var ErrNotFound = errors.New("not found")

// ErrConflict は一意制約違反が発生した場合に返される。
// infra 層が PG エラーコード 23505 (unique_violation) を検知してこのエラーに変換する。
// usecase は文脈に応じて適切なドメインエラーに変換するか、隠蔽する。
var ErrConflict = errors.New("conflict")

// ErrForbidden は外部 API が 403 を返した場合に返される。
// infra 層が HTTP ステータスコードを検知してこのエラーに変換する。
var ErrForbidden = errors.New("forbidden")
