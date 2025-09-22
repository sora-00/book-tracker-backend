# Book Tracker Backend

本のトラッキングアプリのバックエンドAPI

## セットアップ

### 1. 環境変数の設定

`.env`ファイルを作成し、以下の環境変数を設定してください：

```bash
# .envファイルを作成
cp .env.example .env

# .envファイルを編集
GOOGLE_CLOUD_PROJECT=your-actual-project-id
GOOGLE_APPLICATION_CREDENTIALS=path/to/your/service-account-key.json
```

### 2. 依存関係のインストール

```bash
go mod tidy
```

### 3. サーバーの起動

```bash
go run .
```

## API エンドポイント

- `GET /health` - ヘルスチェック
- `GET /books` - 本の一覧取得（認証必要）
- `POST /books` - 本の追加（認証必要）

## 認証

Firebase Authenticationを使用しています。認証トークンは`Authorization: Bearer YOUR_FIREBASE_ID_TOKEN`ヘッダーで送信してください。
