package passwordutil

import (
	"github.com/cpusoft/goutil/hashutil"
	"github.com/cpusoft/goutil/uuidutil"
)

func GetHashPassword(password string) (hashPassword, salt string) {
	salt = uuidutil.GetUuid()
	hashPassword = hashutil.Sha256([]byte(password + salt))
	return hashPassword, salt
}

func VerifyHashPassword(password, salt, hashPassword string) (isPass bool) {
	hashPassword1 := hashutil.Sha256([]byte(password + salt))
	return hashPassword == hashPassword1
}
