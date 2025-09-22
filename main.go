package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"github.com/gin-gonic/gin"
	"google.golang.org/api/option"
)

func main() {
	ctx := context.Background()

	app, authClient, fsClient := mustInitFirebase(ctx)
	defer fsClient.Close()

	r := gin.Default()
	r.Use(cors())

	// ヘルスチェック
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// 認証必須ルート
	authed := r.Group("/")
	authed.Use(verifyIDToken(authClient))

	// 本の一覧取得
	authed.GET("/books", func(c *gin.Context) {
		uid := c.GetString("uid")
		iter := fsClient.Collection("users").Doc(uid).Collection("books").Documents(c)
		defer iter.Stop()
		
		var books []map[string]interface{}
		for {
			doc, err := iter.Next()
			if err != nil {
				break
			}
			book := doc.Data()
			book["id"] = doc.Ref.ID
			books = append(books, book)
		}
		
		c.JSON(http.StatusOK, gin.H{"books": books})
	})

	// 本の追加
	authed.POST("/books", func(c *gin.Context) {
		uid := c.GetString("uid")
		var body struct {
			Title  string `json:"title"`
			Status string `json:"status"`
		}
		if err := c.BindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
			return
		}

		doc := map[string]interface{}{
			"title":     body.Title,
			"status":    body.Status,
			"createdAt": time.Now(),
		}
		_, _, err := fsClient.Collection("users").Doc(uid).Collection("books").Add(c, doc)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "firestore write failed"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "book saved"})
	})

	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
	_ = app // 今後使う場合に備えて保持
}

func mustInitFirebase(ctx context.Context) (*firebase.App, *auth.Client, *firestore.Client) {
	cred := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
	
	// デフォルト値を設定（開発用）
	if projectID == "" {
		projectID = "your-actual-project-id" // ここに実際のProject IDを入力
	}
	
	var app *firebase.App
	var err error
	
	config := &firebase.Config{
		ProjectID: projectID,
	}
	
	if cred != "" {
		app, err = firebase.NewApp(ctx, config, option.WithCredentialsFile(cred))
	} else {
		// GCP 環境など Application Default Credentials を使う場合
		app, err = firebase.NewApp(ctx, config)
	}
	if err != nil {
		log.Fatalf("firebase init error: %v", err)
	}

	authClient, err := app.Auth(ctx)
	if err != nil {
		log.Fatalf("auth init error: %v", err)
	}

	fsClient, err := app.Firestore(ctx)
	if err != nil {
		log.Fatalf("firestore init error: %v", err)
	}
	return app, authClient, fsClient
}

func verifyIDToken(authClient *auth.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		authz := c.GetHeader("Authorization")
		if authz == "" || !strings.HasPrefix(authz, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing bearer token"})
			return
		}
		idToken := strings.TrimPrefix(authz, "Bearer ")
		token, err := authClient.VerifyIDToken(c, idToken)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}
		c.Set("uid", token.UID)
		c.Next()
	}
}

func cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Expo/React Native からのアクセス用。必要に応じて Origin を限定してください。
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}