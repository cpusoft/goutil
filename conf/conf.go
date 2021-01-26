package conf

import (
	"fmt"
	"os"
	"strings"

	config "github.com/astaxie/beego/config"
	osutil "github.com/cpusoft/goutil/osutil"
)

var Configure config.Configer

// load configure file
func init() {
	/*
			cannot use flag in init()
				flagFile := flag.String("conf", "", "")
				flag.Parse()
				fmt.Println("conf file is ", *flagFile, " from args")
				exists, err := osutil.IsExists(*flagFile)
				if err != nil || !exists {
					*flagFile = osutil.GetParentPath() + string(os.PathSeparator) + "conf" + string(os.PathSeparator) + "project.conf"
					fmt.Println("conf file is ", *flagFile, " default")
				}
		so ,use os.Args

	*/
	var err error
	var conf string
	if len(os.Args) > 1 {
		args := strings.Split(os.Args[1], "=")
		if len(args) > 0 && (args[0] == "conf" || args[0] == "-conf" || args[0] == "--conf") {
			conf = args[1]
		}
	}

	// decide by "conf" directory
	if conf == "" {
		path, err := osutil.GetCurrentOrParentAbsolutePath("conf")
		if err != nil {
			panic("found " + path + " failed, " + err.Error())

		}
		conf = path + string(os.PathSeparator) + "project.conf"

	}
	fmt.Println("conf file is ", conf)
	Configure, err = config.NewConfig("ini", conf)
	if err != nil {
		panic("load " + conf + " failed, " + err.Error())

	}

}

func String(key string) string {
	s := Configure.String(key)
	return s
}

func Int(key string) int {
	i, _ := Configure.Int(key)
	return i
}

func Strings(key string) []string {
	s := Configure.Strings(key)
	return s
}

func Bool(key string) bool {
	b, _ := Configure.Bool(key)
	return b
}

func DefaultBool(key string, defaultVal bool) bool {
	return Configure.DefaultBool(key, defaultVal)
}

//destpath=${rpstir2::datadir}/rsyncrepo   --> replace ${rpstir2::datadir}
//-->/root/rpki/data/rsyncrepo --> get /root/rpki/data/rsyncrepo
func VariableString(key string) string {
	if len(key) == 0 || len(String(key)) == 0 {
		return ""
	}
	value := String(key)
	start := strings.Index(value, "${")
	end := strings.Index(value, "}")
	if start >= 0 && end > 0 && start < end {
		//${rpstir2::datadir}/rsyncrepo -->rpstir2::datadir
		replaceKey := string(value[start+len("${") : end])
		if len(replaceKey) == 0 || len(String(replaceKey)) == 0 {
			return value
		}
		//rpstir2::datadir -->get  "/root/rpki/data"
		replaceValue := String(replaceKey)
		prefix := string(value[:start])
		suffix := string(value[end+1:])
		///root/rpki/data/rsyncrepo
		newValue := prefix + replaceValue + suffix
		return newValue
	}
	return ""

}

// key:  "aaa::bbb"
// value: "ccc"
func SetString(key, value string) error {
	return Configure.Set(key, value)
}
