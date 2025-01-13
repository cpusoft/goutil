package passwordutil

import (
	"strings"
	"sync"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/fileutil"
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

func ForceTestHashPassword(hashPassword, salt string, dictFilePathName string) (password string, err error) {
	lines, err := fileutil.ReadFileToLines(dictFilePathName)
	if err != nil {
		belogs.Error("ForceTestHashPassword(): ReadFileToLines fail, dictFilePathName:", dictFilePathName, err)
		return "", err
	}
	ch := make(chan int, 500)
	var wg sync.WaitGroup
	for _, line := range lines {
		wg.Add(1)
		ch <- 1
		go func(wg1 *sync.WaitGroup, ch1 chan int) {
			defer func() {
				<-ch1
				wg.Done()
			}()

			testPassword := strings.TrimSpace(line)
			//belogs.Debug("ForceTestHashPassword(): test pasword:", testPassword)
			isPass := VerifyHashPassword(testPassword, salt, hashPassword)
			if isPass {
				password = testPassword
				belogs.Debug("ForceTestHashPassword(): found password")
			}
		}(&wg, ch)
	}
	wg.Wait()
	close(ch)
	return password, nil
}
