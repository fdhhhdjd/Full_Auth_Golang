package helpers

import (
	"context"
	"fmt"
	"log"

	"firebase.google.com/go/auth"
	"github.com/fdhhhdjd/Go_Secure_Auth_Pro/global"
	"github.com/fdhhhdjd/Go_Secure_Auth_Pro/internal/models"
	"github.com/gin-gonic/gin"
)

type UserTest struct {
	Fullname string `json:"fullname"`
	Email    string `json:"email"`
	Picture  string `json:"picture"`
}

// getAuthClient returns an instance of the Firebase Authentication client.
// It takes a Gin context as a parameter and returns a pointer to the auth.Client.
// If an error occurs during the creation of the client, it logs the error and returns nil.
func getAuthClient(c *gin.Context) *auth.Client {
	authClient, err := global.AdminSdk.Auth(c.Request.Context())
	if err != nil {
		errMsg := fmt.Errorf("error creating user: %v", err)
		log.Fatalf(errMsg.Error())
		return nil
	}

	return authClient
}

// GetUserIDToken retrieves the user ID token and returns a SocialResponse object containing user information.
// It takes a gin.Context object and a uid string as parameters.
// If the user ID token cannot be retrieved or the authClient is nil, it returns nil.
// Otherwise, it returns a SocialResponse object with the user's full name, email, and picture.
func GetUserIDToken(c *gin.Context, uid string) *models.SocialResponse {
	authClient := getAuthClient(c)

	userRecord, err := authClient.GetUser(context.Background(), uid)
	if err != nil || authClient == nil {
		return nil
	}

	return &models.SocialResponse{
		Fullname: userRecord.DisplayName,
		Email:    userRecord.Email,
		Picture:  userRecord.PhotoURL,
	}
}

// createUser creates a new user in Firebase Authentication with the provided email and password.
// It returns the created user record or an error if the user creation fails.
func createUser(authClient *auth.Client, email, password string) (*auth.UserRecord, error) {
	params := (&auth.UserToCreate{}).
		Email(email).
		EmailVerified(false).
		Password(password).
		DisplayName("Example User").
		Disabled(false)

	u, err := authClient.CreateUser(context.Background(), params)
	if err != nil {
		return nil, fmt.Errorf("error creating user: %v", err)
	}
	fmt.Printf("Successfully created user: %v\n", u.UID)
	return u, nil
}

// CreateAndGetUidTestFireBase is a function that creates a new user in Firebase
// and retrieves the ID token for the user.
func CreateAndGetUidTestFireBase(c *gin.Context) {
	authClient := getAuthClient(c)
	// Create a new user
	email := RandomEmail()
	password := RandomPassword()
	u, err := createUser(authClient, email, password)
	if err != nil {
		errMsg := fmt.Errorf("error creating user: %v", err)
		log.Fatalf(errMsg.Error())
	}

	// Get the ID Token for the user
	userRecord := GetUserIDToken(c, u.UID)
	if userRecord != nil {
		fmt.Printf("ID userRecord: %s\n", userRecord)
	} else {
		log.Fatalf("error getting ID userRecord")
	}
}
