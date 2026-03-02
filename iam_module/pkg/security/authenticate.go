package security

import (
	"os"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/rdhmuhammad/phisiobook/pkg/clock"
	"github.com/rdhmuhammad/phisiobook/pkg/environment"
)

type Auth struct {
	clock clock.CLOCK
	env   environment.ENV
}

func NewAuth() Auth {
	return Auth{
		clock: clock.Default(),
		env:   environment.NewEnvironment(),
	}
}

/*
GenerateSingleToken fo generating single expiration token
*/
func (receiver Auth) GenerateSingleToken(claim SingleTokenClaim) (string, error) {
	method := jwt.SigningMethodHS256
	claim.ExpiresAt = jwt.NewNumericDate(receiver.clock.NowUTC().Add(time.Hour * time.Duration(receiver.env.GetInt("EXPIRED_TOKEN_JWT", 0))))
	token := &jwt.Token{
		Header: map[string]interface{}{
			"typ": "JWT",
			"alg": method.Alg(),
		},
		Claims: claim,
		Method: method,
	}
	secret := []byte(os.Getenv("SECRET"))
	tokenStr, err := token.SignedString(secret)
	if err != nil {

		return "", err
	}
	return tokenStr, nil
}
