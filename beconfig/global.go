package config

// We use this to simply application's development
// for most users, they only need to use those methods
var globalInstance Configer

// InitGlobalInstance will ini the global instance
// If you want to use specific implementation, don't forget to import it.
// err := InitGlobalInstance("etcd", "someconfig")
func InitGlobalInstance(name string, cfg string) error {
	var err error
	globalInstance, err = NewConfig(name, cfg)
	return err
}

// support section::key type in given key when using ini type.
func Set(key, val string) error {
	return globalInstance.Set(key, val)
}

// support section::key type in key string when using ini and json type; Int,Int64,Bool,Float,DIY are same.
func String(key string) (string, error) {
	return globalInstance.String(key)
}

// get string slice
func Strings(key string) ([]string, error) {
	return globalInstance.Strings(key)
}

func Int(key string) (int, error) {
	return globalInstance.Int(key)
}

func Int64(key string) (int64, error) {
	return globalInstance.Int64(key)
}

func Bool(key string) (bool, error) {
	return globalInstance.Bool(key)
}

func Float(key string) (float64, error) {
	return globalInstance.Float(key)
}

// support section::key type in key string when using ini and json type; Int,Int64,Bool,Float,DIY are same.
func DefaultString(key string, defaultVal string) string {
	return globalInstance.DefaultString(key, defaultVal)
}

// get string slice
func DefaultStrings(key string, defaultVal []string) []string {
	return globalInstance.DefaultStrings(key, defaultVal)
}

func DefaultInt(key string, defaultVal int) int {
	return globalInstance.DefaultInt(key, defaultVal)
}

func DefaultInt64(key string, defaultVal int64) int64 {
	return globalInstance.DefaultInt64(key, defaultVal)
}

func DefaultBool(key string, defaultVal bool) bool {
	return globalInstance.DefaultBool(key, defaultVal)
}

func DefaultFloat(key string, defaultVal float64) float64 {
	return globalInstance.DefaultFloat(key, defaultVal)
}

// DIY return the original value
func DIY(key string) (interface{}, error) {
	return globalInstance.DIY(key)
}

func GetSection(section string) (map[string]string, error) {
	return globalInstance.GetSection(section)
}

func Unmarshaler(prefix string, obj interface{}, opt ...DecodeOption) error {
	return globalInstance.Unmarshaler(prefix, obj, opt...)
}

func Sub(key string) (Configer, error) {
	return globalInstance.Sub(key)
}

func OnChange(key string, fn func(value string)) {
	globalInstance.OnChange(key, fn)
}

func SaveConfigFile(filename string) error {
	return globalInstance.SaveConfigFile(filename)
}
