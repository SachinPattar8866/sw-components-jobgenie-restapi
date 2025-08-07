package services

import (
	"context"
	"log"
	"os"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/auth"
	"google.golang.org/api/option"
)

var firebaseApp *firebase.App

// InitFirebase initializes the Firebase Admin SDK
func InitFirebase() {
	serviceAccountFile := os.Getenv("FIREBASE_SERVICE_ACCOUNT_PATH")
	if serviceAccountFile == "" {
		log.Fatal("FIREBASE_SERVICE_ACCOUNT_PATH not set in .env file")
	}

	opt := option.WithCredentialsFile(serviceAccountFile)

	var err error
	firebaseApp, err = firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		log.Fatalf("error initializing firebase app: %v", err)
	}
	log.Println("Firebase Admin SDK initialized successfully.")
}

// VerifyFirebaseToken verifies the token received from the client
func VerifyFirebaseToken(ctx context.Context, idToken string) (*auth.Token, error) { // Note the change here
	client, err := firebaseApp.Auth(ctx)
	if err != nil {
		return nil, err
	}
	return client.VerifyIDToken(ctx, idToken)
}

func GetFirebaseAuthClient(ctx context.Context) (*auth.Client, error) {
	return firebaseApp.Auth(ctx)
}
