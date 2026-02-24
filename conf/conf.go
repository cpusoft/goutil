package conf

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	config "github.com/cpusoft/goutil/beconfig"
	"github.com/cpusoft/goutil/osutil"
)

var configure config.Configer

// load configure file
func init() {
	var err error
	var conf string

	// 修复：遍历所有参数，支持任意位置的 -conf/--conf/conf 参数
	for _, arg := range os.Args[1:] {
		parts := strings.SplitN(arg, "=", 2) // 仅分割第一个=，避免值中包含=
		if len(parts) != 2 {
			continue
		}
		key := parts[0]
		value := parts[1]
		if key == "conf" || key == "-conf" || key == "--conf" {
			conf = value
			break
		}
	}

	// 修复：优先使用命令行参数，无则使用默认路径
	if conf == "" {
		confPath, currentPath, err := osutil.GetConfOrLogPath("conf")
		if err != nil {
			fmt.Printf("conf(): GetConfOrLogPath failed: %v\n", err)
		}
		if confPath == "" {
			fmt.Printf("conf(): confPath not found, use currentPath: %s\n", currentPath)
			// 修复：使用 filepath.Join 保证跨平台兼容性
			conf = filepath.Join(currentPath, "project.conf")
		} else {
			conf = filepath.Join(confPath, "project.conf")
		}
	}

	fmt.Printf("conf file is: %s\n", conf)

	// 修复：检查配置文件是否存在
	if _, err := os.Stat(conf); err != nil {
		fmt.Printf("conf(): config file not exist: %s, err: %v\n", conf, err)
		configure = nil
		return
	}

	configure, err = config.NewConfig("ini", conf)
	if err != nil {
		fmt.Printf("NewConfig failed, file %s is not ini format: %v\n", conf, err)
		configure = nil
		return
	}
	fmt.Printf("NewConfig success, file: %s\n", conf)
}

// 以下基础配置读取函数保持原有逻辑，仅优化注释
func String(key string) string {
	if configure != nil {
		s, _ := configure.String(key)
		return s
	}
	return ""
}

func DefaultString(key string, defaultVal string) string {
	if configure != nil {
		return configure.DefaultString(key, defaultVal)
	}
	return defaultVal
}

func Int(key string) int {
	if configure != nil {
		i, _ := configure.Int(key)
		return i
	}
	return 0
}

func DefaultInt(key string, defaultVal int) int {
	if configure != nil {
		return configure.DefaultInt(key, defaultVal)
	}
	return defaultVal
}

// 新增：兼容beconfig的Strings分隔符问题，手动按逗号分割
func Strings(key string) []string {
	if configure != nil {
		s, err := configure.Strings(key)
		// 若解析失败/返回长度为1，手动按逗号分割（兼容常见场景）
		if err != nil || (len(s) == 1 && strings.Contains(s[0], ",")) {
			raw := String(key)
			if len(raw) > 0 {
				return strings.Split(raw, ",")
			}
			return nil
		}
		return s
	}
	return nil
}

func DefaultStrings(key string, defaultValues []string) []string {
	if configure != nil {
		return configure.DefaultStrings(key, defaultValues)
	}
	return defaultValues
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
	return defaultVal
}

/* close
// VariableString 解析带变量占位符的配置值（如 ${rpstir2::datadir}/rsyncrepo）
// 不支持嵌套占位符 （如 ${a::${b::c}}）
// Deprecated: 建议使用更健壮的变量替换方案
// VariableString 解析带变量占位符的配置值（仅支持单层占位符，如 ${rpstir2::datadir}/rsyncrepo）
// 注意：
// 1. 不支持嵌套占位符（如 ${a::${b::c}}），此类场景会返回原始值；
// 2. 占位符格式必须为 ${key}，且 key 必须存在配置值，否则返回原始值；
// Deprecated: 建议使用更健壮的变量替换方案
func VariableString(key string) string {
	// 边界校验：key或配置值为空时返回空
	if len(key) == 0 {
		return ""
	}
	value := String(key)
	if len(value) == 0 {
		return ""
	}

	start := strings.Index(value, "${")
	end := strings.Index(value, "}")

	// 边界校验，避免切片越界
	if start < 0 || end <= start || start+2 > end { // ${ 占2个字符，需保证 start+2 <= end
		return value // 无合法占位符，返回原始值
	}

	// 提取占位符内的key（如 ${rpstir2::datadir} -> rpstir2::datadir）
	replaceKey := value[start+2 : end]
	replaceValue := String(replaceKey)
	if len(replaceValue) == 0 {
		return value // 占位符key无配置值，返回原始值
	}

	// 拼接新值：前缀 + 替换值 + 后缀
	prefix := value[:start]
	suffix := ""
	if end+1 < len(value) {
		suffix = value[end+1:]
	}
	newValue := prefix + replaceValue + suffix
	return newValue
}
*/

// SetString 设置配置值
// 修复：增加key/value合法性校验
func SetString(key, value string) error {
	if configure == nil {
		return errors.New("configure is nil")
	}
	if len(key) == 0 {
		return errors.New("key is empty")
	}
	return configure.Set(key, value)
}
