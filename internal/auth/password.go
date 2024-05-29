package auth

import "github.com/matthewhartstonge/argon2"

func Verify(password, encoded string) (bool, error) {
	ok, err := argon2.VerifyEncoded([]byte(password), []byte(encoded))
	if err != nil {
		return false, err
	}

	return ok, nil
}
