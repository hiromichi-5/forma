# デプロイガイド

## アーキテクチャ

```text
                     ┌─────────────┐
                     │  Cloudflare │
                     └──────┬──────┘
                            │
              ┌─────────────┼─────────────┐
              │             │             │
    forma.hiromichi.dev     │   api.forma.hiromichi.dev
              │             │             │
       ┌──────┴──────┐     │   ┌─────────┴─────────┐
       │   Vercel    │     │   │    Cloud Run       │
       │  (Frontend) │     │   │    (Backend)       │
       └─────────────┘     │   └─────────┬─────────┘
                           │             │
                           │   ┌─────────┴─────────┐
                           │   │    Supabase        │
                           │   │   (PostgreSQL)     │
                           │   └───────────────────┘
                           │
                     ┌─────┴─────┐
                     │  Resend   │
                     └───────────┘
```

## IaC 化

インフラのコード化については以下を参照。

- 現行構成（Cloud Run / Supabase / Vercel / Cloudflare）の Terraform 化: `docs/design/iac-current-infra.md`
- AWS に載せる場合の理想構成: `docs/design/iac-aws-ideal.md`

## WIP
