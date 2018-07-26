package patrol

import (
	"crypto/rand"
	"fmt"
)

func uuidV4() (
	string,
	error,
) {
	// https://en.wikipedia.org/wiki/Universally_unique_identifier#Version_4_(random)
	// 123e4567-e89b-12d3-a456-426655440000
	// xxxxxxxx-xxxx-Mxxx-Nxxx-xxxxxxxxxxxx
	// The four bits of digit M indicate the UUID version, and the one to three most significant bits of digit N indicate the UUID variant.
	uuid := make([]byte, 16)
	_, err := rand.Read(uuid)
	if err != nil {
		return "", err
	}
	// https://tools.ietf.org/html/rfc4122#section-4.1.1
	// variant bits; see section 4.1.1
	uuid[8] = uuid[8]&^0xc0 | 0x80
	// https://tools.ietf.org/html/rfc4122#section-4.1.3
	// version 4 (pseudo-random); see section 4.1.3
	uuid[6] = uuid[6]&^0xf0 | 0x40
	return fmt.Sprintf("%x-%x-%x-%x-%x", uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:]), nil
}
func uuidMust(
	uuid string,
	err error,
) string {
	// idea taken from: https://github.com/satori/go.uuid/blob/master/uuid.go
	if err != nil {
		panic(err)
	}
	return uuid
}
