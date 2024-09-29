package conf

import (
	"errors"
	"fmt"
	"os"
	"strings"

	config "github.com/cpusoft/goutil/beconfig"
	"github.com/cpusoft/goutil/osutil"
)

var configure config.Configer

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
		if len(args) > 1 && (args[0] == "conf" || args[0] == "-conf" || args[0] == "--conf") {
			conf = args[1]
		}
	}

	// decide by "conf" directory
	if conf == "" {
		path, err := osutil.GetCurrentOrParentAbsolutePath("conf")
		if err != nil {
			fmt.Println("found " + path + " failed, " + err.Error())
			return
		}
		conf = path + string(os.PathSeparator) + "project.conf"
	}
	//fmt.Println("conf file is ", conf)
	configure, err = config.NewConfig("ini", conf)
	if err != nil {
		fmt.Println("Loaded configuration file " + conf + " is not in ini format. " + err.Error())
		return
	}

}

func String(key string) string {
	if configure != nil {
		s, _ := configure.String(key)
		return s
	}
	return ""
}

func Int(key string) int {
	if configure != nil {
		i, _ := configure.Int(key)
		return i
	}
	return 0
}

func Strings(key string) []string {
	if configure != nil {
		s, _ := configure.Strings(key)
		return s
	}
	return nil
}

func Bool(key string) bool {
	if configure != nil {
		b, _ := configure.Bool(key)
		return b
	}
	return false
}

func DefaultBool(key string, defaultVal bool) bool {
	if configure != nil {
		return configure.DefaultBool(key, defaultVal)
	}
	return false
}

// destpath=${rpstir2::datadir}/rsyncrepo   --> replace ${rpstir2::datadir}
// -->/root/rpki/data/rsyncrepo --> get /root/rpki/data/rsyncrepo
// Deprecated
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
		suffix := ""
		if end+1 < len(value) {
			suffix = string(value[end+1:])
		}
		///root/rpki/data/rsyncrepo
		newValue := prefix + replaceValue + suffix
		return newValue
	}
	return ""

}

// key:  "aaa::bbb"
// value: "ccc"
func SetString(key, value string) error {
	if configure != nil {
		return configure.Set(key, value)
	}
	return errors.New("configure is nil")
}
