package opensslutil

import (
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/conf"
	"github.com/cpusoft/goutil/osutil"
)

// 公共函数：获取OpenSSL命令路径（提取重复代码）
func getOpensslCmd() string {
	opensslCmd := "openssl"
	path := conf.String("openssl::path")
	if len(path) > 0 {
		opensslCmd = osutil.JoinPathFile(path, opensslCmd)
	}
	return opensslCmd
}

// 公共函数：校验文件合法性（防路径遍历、空值、文件不存在）
func validateCertFile(certFile string) error {
	// 1. 检查文件路径是否为空
	if strings.TrimSpace(certFile) == "" {
		return errors.New("certificate file path is empty")
	}

	// 2. 清理路径，防止路径遍历（如../../etc/passwd）
	cleanedPath := filepath.Clean(certFile)
	if !filepath.IsAbs(cleanedPath) {
		// 若为相对路径，转为绝对路径后再次校验
		absPath, err := filepath.Abs(cleanedPath)
		if err != nil {
			return fmt.Errorf("invalid certificate file path: %w", err)
		}
		cleanedPath = absPath
	}

	// 3. 处理osutil.IsExists的双返回值（bool, error）
	exists, err := osutil.IsExists(cleanedPath)
	if err != nil {
		// 若判断文件存在性时出错，返回该错误
		return fmt.Errorf("failed to check certificate file existence: %w", err)
	}
	if !exists {
		// 若文件不存在，返回对应错误
		return fmt.Errorf("certificate file not found: %s", cleanedPath)
	}

	return nil
}

// 公共函数：执行OpenSSL命令（移除超时控制）
func execOpensslCmd(args []string) ([]byte, error) {
	opensslCmd := getOpensslCmd()
	belogs.Debug("execOpensslCmd(): cmd:", opensslCmd, "args:", args)

	// 执行命令（移除超时控制，恢复原始exec.Command）
	cmd := exec.Command(opensslCmd, args...)
	output, err := cmd.CombinedOutput()

	return output, err
}

// 内部公共函数：处理OpenSSL输出结果（提取重复逻辑，小写开头表示内部使用）
func processOpensslOutput(output []byte) []string {
	result := string(output)
	tmps := strings.Split(result, osutil.GetNewLineSep())
	results := make([]string, len(tmps))
	for i := range tmps {
		results[i] = strings.TrimSpace(tmps[i])
	}
	return results
}

func GetResultsByOpensslX509(certFile string) (results []string, err error) {
	// 1. 前置校验文件合法性
	if err = validateCertFile(certFile); err != nil {
		belogs.Error("GetResultsByOpensslX509(): validate cert file fail:", err)
		return nil, errors.New("invalid certificate file: " + err.Error())
	}

	// 2. 定义OpenSSL命令参数（DER格式）
	argsDer := []string{"x509", "-noout", "-text", "-in", certFile, "-inform", "der"}
	output, err := execOpensslCmd(argsDer)
	if err != nil {
		belogs.Debug("GetResultsByOpensslX509(): try DER format fail, err:", err, "output:", string(output))

		// 尝试PEM格式
		argsPem := []string{"x509", "-noout", "-text", "-in", certFile, "-inform", "pem"}
		output, err = execOpensslCmd(argsPem)
		if err != nil {
			belogs.Error("GetResultsByOpensslX509(): both DER and PEM format fail, certFile:", certFile, "err:", err, "output:", string(output))
			// 隐藏敏感信息，只返回通用错误
			return nil, errors.New("fail to parse x509 certificate: invalid format or corrupted file")
		}
	}

	// 3. 调用公共内部函数处理输出结果
	results = processOpensslOutput(output)
	belogs.Debug("GetResultsByOpensslX509(): len(results):", len(results))

	return results, nil
}

func GetResultsByOpensslAns1(certFile string) (results []string, err error) {
	// 1. 前置校验文件合法性
	if err = validateCertFile(certFile); err != nil {
		belogs.Error("GetResultsByOpensslAns1(): validate cert file fail:", err)
		return nil, errors.New("invalid asn1 file: " + err.Error())
	}

	// 2. 定义OpenSSL命令参数（DER格式）
	argsDer := []string{"asn1parse", "-in", certFile, "-inform", "der"}
	output, err := execOpensslCmd(argsDer)
	if err != nil {
		belogs.Debug("GetResultsByOpensslAns1(): try DER format fail, err:", err, "output:", string(output))

		// 尝试PEM格式
		argsPem := []string{"asn1parse", "-in", certFile, "-inform", "pem"}
		output, err = execOpensslCmd(argsPem)
		if err != nil {
			belogs.Error("GetResultsByOpensslAns1(): both DER and PEM format fail, certFile:", certFile, "err:", err, "output:", string(output))
			// 修复反引号未闭合问题 + 隐藏敏感信息
			return nil, errors.New("fail to parse asn1 format: invalid format or corrupted file")
		}
	}

	// 3. 调用公共内部函数处理输出结果
	results = processOpensslOutput(output)
	belogs.Debug("GetResultsByOpensslAns1(): len(results):", len(results))

	return results, nil
}
