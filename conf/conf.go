package conf

import (
	config "github.com/astaxie/beego/config"
	"os"
)

var Configure config.Configer

func init() {
	/*
		iniFile := config.NewINIFile(util.GetParentPath() + string(os.PathSeparator) + "conf/slurm.conf")
		Configure = config.NewConfig([]config.Provider{iniFile})
		if err := Configure.Load(); err != nil {
			fmt.Println("conf:", err)
		}
		fmt.Println("conf:", *Configure)
	*/
	Configure, _ = config.NewConfig("ini", util.GetParentPath()+string(os.PathSeparator)+"conf" + string(os.PathSeparator) +"project.conf")

}
