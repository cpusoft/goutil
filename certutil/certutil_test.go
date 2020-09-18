package cert

import (
	"fmt"
	"testing"
)

func TestReadFileToCer(t *testing.T) {
	path := `G:\Download\cert\verify\2\`
	fatherFile := path + `inter.pem.cer`
	p, err := ReadFileToCer(fatherFile)
	fmt.Println(p, err)

	fatherFile = `G:\Download\cert\verify\3\rtpKuIKhDn9Y8Zg6y9HhlQfmPsU.roa`
	p, err = ReadFileToCer(fatherFile)
	fmt.Println(p, err)
}
func TestReadFileToCrl(t *testing.T) {
	path := `G:\Download\cert\1\`
	fatherFile := path + `ACBRR9OW8JgDvUcuWBka9usiwvU.crl`
	p, err := ReadFileToCrl(fatherFile)
	fmt.Println(p, err)

	fatherFile = `G:\Download\cert\verify\3\rtpKuIKhDn9Y8Zg6y9HhlQfmPsU.roa`
	p, err = ReadFileToCrl(fatherFile)
	fmt.Println(p, err)
}

func TestReadFileToByte(t *testing.T) {
	path := `G:\Download\cert\verify\2\`
	fatherFile := path + `inter.pem.cer`
	p, by, err := ReadFileToByte(fatherFile)
	fmt.Println(p, by, err)

	fatherFile = path + `inter.cer`
	p, by, err = ReadFileToByte(fatherFile)
	fmt.Println(p, by, err)
}

func TestVerifyCertByX509(t *testing.T) {
	path := `G:\Download\cert\`
	fatherFile := path + `c0793683-aa07-4935-a2c2-ec423ea7dd0b.father.cer`
	childFile := path + `c922abf8-95b1-37f0-90cd-bdb125467e8e.ee.cer`

	result, err := VerifyCerByX509(fatherFile, childFile)
	fmt.Println(result, err)
}
func TestVerifyRootCertByX509(t *testing.T) {
	path := `E:\Go\common-util\src\certutil\example\`
	root := path + `root.cer`
	//childFile := path + `inter.cer`

	result, err := VerifyRootCerByOpenssl(root)
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
	path := `G:\Download\cert\`
	fatherFile := path + `ohcWJIUz0QduJriNGOlBlT-lB9c.cer`
	childFile := path + `db42e932-926a-42bd-afdb-63320fa7ec40.roa`

	result, err := VerifyEeCertByX509(fatherFile, childFile, 838969, 1019659)
	fmt.Println(result, err)
}

func TestVerifyCrlByX509(t *testing.T) {
	path := `G:\Download\cert\verify\4\`

	//cerFile := path + `inter.cer` //err
	cerFile := path + `bW-_qXU9uNhGQz21NR2ansB8lr0.cer` //ok
	crlFile := path + `bW-_qXU9uNhGQz21NR2ansB8lr0.crl`

	result, err := VerifyCrlByX509(cerFile, crlFile)
	fmt.Println(result, err)
}
