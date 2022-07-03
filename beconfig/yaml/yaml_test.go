package yaml

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	config "github.com/cpusoft/goutil/beconfig"
)

func TestYaml(t *testing.T) {
	var (
		yamlcontext = `
"appname": beeapi
"httpport": 8080
"mysqlport": 3600
"PI": 3.1415976
"runmode": dev
"autorender": false
"copyrequestbody": true
"PATH": GOPATH
"path1": ${GOPATH}
"path2": ${GOPATH||/home/go}
"empty": "" 
"user":
  "name": "tom"
  "age": 13
`

		keyValue = map[string]interface{}{
			"appname":         "beeapi",
			"httpport":        8080,
			"mysqlport":       int64(3600),
			"PI":              3.1415976,
			"runmode":         "dev",
			"autorender":      false,
			"copyrequestbody": true,
			"PATH":            "GOPATH",
			"path1":           os.Getenv("GOPATH"),
			"path2":           os.Getenv("GOPATH"),
			"error":           "",
			"emptystrings":    []string{},
		}
	)
	f, err := os.Create("testyaml.conf")
	if err != nil {
		t.Fatal(err)
	}
	_, err = f.WriteString(yamlcontext)
	if err != nil {
		f.Close()
		t.Fatal(err)
	}
	f.Close()
	defer os.Remove("testyaml.conf")
	yamlconf, err := config.NewConfig("yaml", "testyaml.conf")
	if err != nil {
		t.Fatal(err)
	}

	res, _ := yamlconf.String("appname")
	if res != "beeapi" {
		t.Fatal("appname not equal to beeapi")
	}

	for k, v := range keyValue {

		var (
			value interface{}
			err   error
		)

		switch v.(type) {
		case int:
			value, err = yamlconf.Int(k)
		case int64:
			value, err = yamlconf.Int64(k)
		case float64:
			value, err = yamlconf.Float(k)
		case bool:
			value, err = yamlconf.Bool(k)
		case []string:
			value, err = yamlconf.Strings(k)
		case string:
			value, err = yamlconf.String(k)
		default:
			value, err = yamlconf.DIY(k)
		}
		if err != nil {
			t.Errorf("get key %q value fatal,%v err %s", k, v, err)
		} else if fmt.Sprintf("%v", v) != fmt.Sprintf("%v", value) {
			t.Errorf("get key %q value, want %v got %v .", k, v, value)
		}

	}

	if err = yamlconf.Set("name", "astaxie"); err != nil {
		t.Fatal(err)
	}
	res, _ = yamlconf.String("name")
	if res != "astaxie" {
		t.Fatal("get name error")
	}

	sub, err := yamlconf.Sub("user")
	assert.Nil(t, err)
	assert.NotNil(t, sub)
	name, err := sub.String("name")
	assert.Nil(t, err)
	assert.Equal(t, "tom", name)

	age, err := sub.Int("age")
	assert.Nil(t, err)
	assert.Equal(t, 13, age)

	user := &User{}

	err = sub.Unmarshaler("", user)
	assert.Nil(t, err)
	assert.Equal(t, "tom", user.Name)
	assert.Equal(t, 13, user.Age)

	user = &User{}

	err = yamlconf.Unmarshaler("user", user)
	assert.Nil(t, err)
	assert.Equal(t, "tom", user.Name)
	assert.Equal(t, 13, user.Age)
}

type User struct {
	Name string `yaml:"name"`
	Age  int    `yaml:"age"`
}
