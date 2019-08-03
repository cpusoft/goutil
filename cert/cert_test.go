package cert

import (
	"fmt"
	"testing"
)

func TestXYZ(t *testing.T) {
	path := `G:\Download\cert\verify\2\`
	fatherFile := path + `inter.cer`
	childFile := path + `A9.cer`

	result, err := VerifyCertsByX509(fatherFile, childFile)
	fmt.Println(result, err)
}
