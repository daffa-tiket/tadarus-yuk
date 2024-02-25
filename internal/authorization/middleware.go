package authorization

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
	"io"

	"github.com/golang-jwt/jwt"
	"github.com/gorilla/mux"
)

type CustomClaims struct {
	UserID int    `json:"user_id"`
	Role   string `json:"role"`
	jwt.StandardClaims
}

var (
	JwtSecretKey []byte
	CipherSecretKey []byte
)

func InitSecret() {
	JwtSecretKey = []byte(os.Getenv("JWT_SECRET_KEY"))
	CipherSecretKey = []byte(os.Getenv("CIPHER_SECRET_KEY"))
}

func AuthenticationMiddleware(role string) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Printf("Incoming request: %s %s %v", r.Method, r.URL.Path, r.Header)
			// Perform authentication logic here based on the role
			// For example, check if a valid token is present in the request headers
			tokenString := r.Header.Get("Authorization")
			if tokenString == "" {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Parse and validate the token
			claims, err := parseAndValidateToken(tokenString)
			if err != nil {
				log.Printf("Error : %v", err.Error())
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			fmt.Println(role)
			// Perform additional role-specific checks as needed
			if role == "user" && claims.Role != "user" {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			} else if role == "admin" && claims.Role != "admin" {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			// Set CORS headers
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

			// If authentication and role checks are successful, proceed to the next handler
			next.ServeHTTP(w, r)
		})
	}
}

func parseAndValidateToken(tokenString string) (*CustomClaims, error) {
	// Parse the token
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return JwtSecretKey, nil
	})

	if err != nil {
		log.Printf("Error : %v", err.Error())
		return nil, err
	}

	// Validate the token
	if !token.Valid {
		return nil, jwt.ErrSignatureInvalid
	}

	// Extract claims
	claims, ok := token.Claims.(*CustomClaims)
	if !ok {
		return nil, jwt.ErrInvalidKey
	}

	return claims, nil
}

func GenerateAuthToken(userID int, role string) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["user_id"] = userID
	claims["role"] = role
	claims["exp"] = time.Now().Add(time.Hour * 24).Unix()

	// Sign the token with a secret key
	signedToken, err := token.SignedString(JwtSecretKey)
	if err != nil {
		log.Printf("Error : %v", err.Error())
		return "", err
	}

	return signedToken, nil
}

func EncryptUserID(userID int) (string, error) {
	block, err := aes.NewCipher(CipherSecretKey)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, 12)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	userIDBytes := []byte(fmt.Sprintf("%d", userID))

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	ciphertext := aesGCM.Seal(nil, nonce, userIDBytes, nil)

	encryptedUserID := append(nonce, ciphertext...)

	return base64.URLEncoding.EncodeToString(encryptedUserID), nil
}

func DecryptUserID(encryptedUserID string) (int, error) {
	encryptedUserIDBytes, err := base64.URLEncoding.DecodeString(encryptedUserID)
	if err != nil {
		return 0, err
	}

	block, err := aes.NewCipher(CipherSecretKey)
	if err != nil {
		return 0, err
	}

	nonce := encryptedUserIDBytes[:12]
	ciphertext := encryptedUserIDBytes[12:]

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return 0, err
	}

	decryptedUserIDBytes, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return 0, err
	}

	decryptedUserID, err := strconv.Atoi(string(decryptedUserIDBytes))
	if err != nil {
		return 0, err
	}

	return decryptedUserID, nil
}

func EncryptEmail(email string) (string, error) {
	block, err := aes.NewCipher(CipherSecretKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(email), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func DecryptEmail(encryptedEmail string) (string, error) {
	encryptedEmailBytes, err := base64.URLEncoding.DecodeString(encryptedEmail)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(CipherSecretKey)
	if err != nil {
		return "", err
	}

	nonce := encryptedEmailBytes[:12]
	ciphertext := encryptedEmailBytes[12:]

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	decryptedEmailBytes, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(decryptedEmailBytes), nil
}