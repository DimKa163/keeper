package datatool

import "crypto/rand"

func GenerateDek(len int32) ([]byte, error) {
	dek := make([]byte, len)
	if _, err := rand.Read(dek); err != nil {
		return nil, err
	}
	return dek, nil
}
