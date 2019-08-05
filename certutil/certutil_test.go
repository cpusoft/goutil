package cert

import (
	"fmt"
	"testing"
)

func TestVerifyCertByX509(t *testing.T) {
	path := `G:\Download\cert\verify\2\`
	fatherFile := path + `inter.cer`
	childFile := path + `A9.cer`

	result, err := VerifyCertByX509(fatherFile, childFile)
	fmt.Println(result, err)
}
func TestVerifyRootCertByX509(t *testing.T) {
	path := `G:\Download\cert\verify\2\`
	fatherFile := path + `inter.cer`
	childFile := path + `inter.cer`

	result, err := VerifyCertByX509(fatherFile, childFile)
	fmt.Println(result, err)
}
func TestVerifyEeCertByX509(t *testing.T) {
	/*
		/root/rpki/repo/repo/rpki.ripe.net/repository/DEFAULT/ec/49c449-2d9c-4fc9-b340-51a23ddb6410/1/
		rtpKuIKhDn9Y8Zg6y9HhlQfmPsU.roa
			"eeStart": 159,
			"eeEnd": 1426


			/root/rpki/repo/repo/rpki.ripe.net/repository/DEFAULT/
			ACBRR9OW8JgDvUcuWBka9usiwvU.cer
	*/
	path := `G:\Download\cert\verify\3\`
	fatherFile := path + `ACBRR9OW8JgDvUcuWBka9usiwvU.cer`
	childFile := path + `rtpKuIKhDn9Y8Zg6y9HhlQfmPsU.roa`

	result, err := VerifyEeCertByX509(fatherFile, childFile, 159, 1426)
	fmt.Println(result, err)
}
