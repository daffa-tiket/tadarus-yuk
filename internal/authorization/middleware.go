package authorization

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/gorilla/mux"
)

var secretKey = []byte("your-secret-key")

type CustomClaims struct {
	UserID int    `json:"user_id"`
	Role   string `json:"role"`
	jwt.StandardClaims
}

func AuthenticationMiddleware(role string) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
		return secretKey, nil
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
	signedToken, err := token.SignedString(secretKey)
	if err != nil {
		log.Printf("Error : %v", err.Error())
		return "", err
	}

	return signedToken, nil
}
