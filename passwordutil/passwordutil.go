package passwordutil

import (
	"github.com/cpusoft/goutil/hashutil"
	"github.com/cpusoft/goutil/uuidutil"
)

func GetHashPasswordAndSalt(password string) (hashPassword, salt string) {
	salt = uuidutil.GetUuid()
	return GetHashPassword(password, salt), salt
}
func GetHashPassword(password, salt string) (hashPassword string) {
	return hashutil.Sha256([]byte(password + salt))
}
func VerifyHashPassword(password, salt, hashPassword string) (isPass bool) {
	hashPassword1 := hashutil.Sha256([]byte(password + salt))
	return hashPassword == hashPassword1
}
