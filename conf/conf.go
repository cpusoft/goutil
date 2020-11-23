package conf

import (
	"flag"
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
		iniFile := config.NewINIFile(util.GetParentPath() + string(os.PathSeparator) + "conf/slurm.conf")
		Configure = config.NewConfig([]config.Provider{iniFile})
		if err := Configure.Load(); err != nil {
			fmt.Println("conf:", err)
		}
		fmt.Println("conf:", *Configure)
	*/

	flagFile := flag.String("conf",
		osutil.GetParentPath()+string(os.PathSeparator)+"conf"+string(os.PathSeparator)+"project.conf", "")
	fmt.Println("conf file is ", *flagFile)
	exists, err := osutil.IsExists(*flagFile)
	if err != nil {
		panic(*flagFile + "conf init failed, " + err.Error())
	}
	if !exists {
		panic(*flagFile + " is not exists")
	}

	Configure, err = config.NewConfig("ini", *flagFile)
	if err != nil {
		panic("load " + *flagFile + " failed, " + err.Error())
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
