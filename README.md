# Forma

## 概要

Forma（フォーマ）は、Googleフォームの回答をチケット化して管理するWebアプリケーションです。

## ドキュメント

このリポジトリのドキュメントは`/docs`にあります。詳細は以下の通りです。

| ディレクトリ | 内容 |
| --- | --- |
| guides/ | 開発ガイド（getting-started、デプロイ等） |
| adr/ | Architecture Decision Records |
| api/ | APIドキュメント |
| design/ | 機能別デザインドキュメント |
| rules/ | コーディング規約・レビュー観点・テスト方針等 |
| images/ | ドキュメント用画像 |

なお、MarkdownファイルはCIで`markdownlint-cli2`により検証しています。ただし、`MD013`と`MD024`については除外しています。詳細は<https://github.com/DavidAnson/markdownlint/tree/main>を参照してください。
