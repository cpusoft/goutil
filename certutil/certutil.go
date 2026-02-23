package certutil

import (
	"bytes"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/conf"
	"github.com/cpusoft/goutil/convert"
	"github.com/cpusoft/goutil/osutil"
)

// 最大文件读取大小限制 (50MB)
const maxFileSize = 50 * 1024 * 1024

// if cert cannot pass verify, just log info level
// cerFile may be PEM
func ReadFileToCer(fileName string) (*x509.Certificate, error) {
	belogs.Debug("ReadFileToCer(): fileName:", fileName)

	// 检查文件大小（保留安全限制）
	fileInfo, err := os.Stat(fileName)
	if err != nil {
		belogs.Error("ReadFileToCer(): Stat fail, fileName:", fileName, err)
		return nil, err
	}
	if fileInfo.Size() > maxFileSize {
		belogs.Error("ReadFileToCer(): File too large, fileName:", fileName, " size:", fileInfo.Size())
		return nil, errors.New("file size exceeds maximum limit")
	}

	buf, err := os.ReadFile(fileName)
	if err != nil {
		belogs.Error("ReadFileToCer(): ReadFile fail, fileName:", fileName, err)
		return nil, err
	}
	belogs.Debug("ReadFileToCer(): ReadFile fileName:", fileName, "  len(buf):", len(buf))

	// default
	fileByte := buf

	// no PEM data is found, p is nil and the whole of the input is returned in
	p, _ := pem.Decode(buf)
	if p != nil {
		fileByte = p.Bytes
		belogs.Debug("ReadFileToCer(): is pem, pem.Decode ok, fileName:", fileName,
			"  len(fileByte):", len(fileByte))
	}
	belogs.Debug("ReadFileToCer(): ok, fileName:", fileName,
		"  len(fileByte):", len(fileByte))
	return x509.ParseCertificate(fileByte)
}

// if cert cannot pass verify, just log info level
func ReadFileToCrl(fileName string) (*pkix.CertificateList, error) {
	belogs.Debug("ReadFileToCrl(): fileName:", fileName)

	// 检查文件大小（保留安全限制）
	fileInfo, err := os.Stat(fileName)
	if err != nil {
		belogs.Error("ReadFileToCrl(): Stat fail, fileName:", fileName, err)
		return nil, err
	}
	if fileInfo.Size() > maxFileSize {
		belogs.Error("ReadFileToCrl(): File too large, fileName:", fileName, " size:", fileInfo.Size())
		return nil, errors.New("file size exceeds maximum limit")
	}

	fileByte, err := os.ReadFile(fileName)
	if err != nil {
		belogs.Error("ReadFileToCrl(): ReadFile fail, fileName:", fileName, err)
		return nil, err
	}

	if len(fileByte) == 0 {
		belogs.Error("ReadFileToCrl(): len(fileByte) is zero fail, fileName:", fileName,
			"  len(fileByte):", len(fileByte))
		return nil, errors.New("this file is empty, not a valid CRL")
	}

	// 尝试解码PEM格式的CRL
	p, _ := pem.Decode(fileByte)
	if p != nil && (p.Type == "X509 CRL" || p.Type == "CRL") {
		fileByte = p.Bytes
	}

	belogs.Debug("ReadFileToCrl(): ok, fileName:", fileName,
		"  len(fileByte):", len(fileByte), err)

	crl, err := x509.ParseCRL(fileByte)
	if err != nil {
		belogs.Error("ReadFileToCrl(): ParseCRL fail, fileName:", fileName, err)
		return nil, fmt.Errorf("failed to parse CRL: %w", err)
	}

	return crl, nil
}

// if cert cannot pass verify, just log info level
func VerifyCerByX509(fatherCertFile string, childCertFile string) (result string, err error) {
	fatherFileByte, err := os.ReadFile(fatherCertFile)
	if err != nil {
		belogs.Error("VerifyCerByX509():fatherCertFile:", fatherCertFile, "   ReadFile err:", err)
		return "fail", fmt.Errorf("failed to read father certificate: %w", err)
	}
	childFileByte, err := os.ReadFile(childCertFile)
	if err != nil {
		belogs.Error("VerifyCerByX509():childCertFile:", childCertFile, "   ReadFile err:", err)
		return "fail", fmt.Errorf("failed to read child certificate: %w", err)
	}
	result, err = VerifyCerByteByX509(fatherFileByte, childFileByte)
	if err != nil {
		belogs.Info("VerifyCerByX509():VerifyCerByteByX509 fail, fatherCertFile:", fatherCertFile, "    childCertFile:", childCertFile, "  err:", err)
		return "fail", err
	}
	return result, nil
}

// if cert cannot pass verify, just log info level
func VerifyEeCertByX509(fatherCertFile string, mftRoaFile string, eeCertStart, eeCertEnd uint64) (result string, err error) {
	fatherFileByte, err := os.ReadFile(fatherCertFile)
	if err != nil {
		belogs.Error("VerifyEeCertByX509():read father, fatherCertFile:", fatherCertFile, "   mftRoaFile:", mftRoaFile, "     ReadFile err:", err)
		return "fail", fmt.Errorf("failed to read father certificate: %w", err)
	}
	mftRoaFileByte, err := os.ReadFile(mftRoaFile)
	if err != nil {
		belogs.Error("VerifyEeCertByX509():ReadFile fail, fatherCertFile:", fatherCertFile, "    mftRoaFile:", mftRoaFile, "   ReadFile err:", err)
		return "fail", fmt.Errorf("failed to read MFT ROA file: %w", err)
	}
	if eeCertStart >= eeCertEnd || len(mftRoaFileByte) < int(eeCertEnd) {
		belogs.Error("VerifyEeCertByX509():mftRoaFileByte[eeCertStart:eeCertEnd] fatherCertFile:", fatherCertFile, "    mftRoaFile:", mftRoaFile,
			"   eeCertStart :", eeCertStart, "   eeCertEnd:", eeCertEnd, "   len(mftRoaFileByte):", len(mftRoaFileByte))
		return "fail", fmt.Errorf("invalid EE certificate range: start=%d, end=%d, file length=%d",
			eeCertStart, eeCertEnd, len(mftRoaFileByte))
	}

	b := mftRoaFileByte[eeCertStart:eeCertEnd]
	result, err = VerifyCerByteByX509(fatherFileByte, b)
	if err != nil {
		belogs.Info("VerifyEeCertByX509():VerifyCerByteByX509 fail, fatherCertFile:", fatherCertFile, "    mftRoaFile:", mftRoaFile,
			"   eeCertStart, eeCertEnd:", eeCertStart, eeCertEnd, "  err:", err)
		return "fail", err
	}
	return result, nil
}

// if cert cannot pass verify, just log info level. .
// fatherCerFile is as root, childCerFile is to be verified. .
// result: ok/fail
func VerifyCerByteByX509(fatherCertByte []byte, childCertByte []byte) (result string, err error) {
	//belogs.Debug("VerifyCerByteByX509():fatherCertByte:", len(fatherCertByte), "   childCertByte:", len(childCertByte))

	fatherPool := x509.NewCertPool()

	faterCert, err := x509.ParseCertificate(fatherCertByte)
	if err != nil {
		belogs.Info("VerifyCerByteByX509():parse fatherCertFile fail, ParseCertificate err:", err)
		return "fail", fmt.Errorf("failed to parse father certificate: %w", err)
	}
	faterCert.UnhandledCriticalExtensions = make([]asn1.ObjectIdentifier, 0)
	belogs.Debug("VerifyCerByteByX509():father issuer:", faterCert.Issuer.String(), "   subject:", faterCert.Subject.String())

	fatherPool.AddCert(faterCert)

	childCert, err := x509.ParseCertificate(childCertByte)
	if err != nil {
		belogs.Info("VerifyCerByteByX509():parse childCertFile fail:  father issuer:", faterCert.Issuer.String(), " ParseCertificate err:", err)
		return "fail", fmt.Errorf("failed to parse child certificate: %w", err)
	}
	childCert.UnhandledCriticalExtensions = make([]asn1.ObjectIdentifier, 0)
	belogs.Debug("VerifyCerByteByX509():child issuer:", childCert.Issuer.String(), "   childCert:", childCert.Subject.String())

	// 保持KeyUsages为ExtKeyUsageClientAuth不变（按你的要求）
	/*
		https://tools.ietf.org/html/rfc5280#section-4.2.1.12
		   id-kp OBJECT IDENTIFIER ::= { id-pkix 3 }

		   id-kp-serverAuth             OBJECT IDENTIFIER ::= { id-kp 1 }
		   -- TLS WWW server authentication
		   -- Key usage bits that may be consistent: digitalSignature,
		   -- keyEncipherment or keyAgreement

		   id-kp-clientAuth             OBJECT IDENTIFIER ::= { id-kp 2 }
		   -- TLS WWW client authentication
		   -- Key usage bits that may be consistent: digitalSignature
		   -- and/or keyAgreement

		   id-kp-codeSigning             OBJECT IDENTIFIER ::= { id-kp 3 }
		   -- Signing of downloadable executable code
		   -- Key usage bits that may be consistent: digitalSignature

		   id-kp-emailProtection         OBJECT IDENTIFIER ::= { id-kp 4 }
		   -- Email protection
		   -- Key usage bits that may be consistent: digitalSignature,
		   -- nonRepudiation, and/or (keyEncipherment or keyAgreement)

		   id-kp-timeStamping            OBJECT IDENTIFIER ::= { id-kp 8 }
		   -- Binding the hash of an object to a time
		   -- Key usage bits that may be consistent: digitalSignature
		   -- and/or nonRepudiation

		   id-kp-OCSPSigning            OBJECT IDENTIFIER ::= { id-kp 9 }
		   -- Signing OCSP responses
		   -- Key usage bits that may be consistent: digitalSignature
		   -- and/or nonRepudiation
	*/
	opts := x509.VerifyOptions{
		Roots: fatherPool,
		//Intermediates: inter,
		//KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageAny},
		KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		//KeyUsages: []x509.ExtKeyUsage{x509.KeyUsageCertSign},
	}

	// 执行严格的证书验证
	chains, err := childCert.Verify(opts)
	if err != nil {
		// 只在特定情况下放宽验证（谨慎使用）
		if strings.Contains(err.Error(), "issuer name does not match subject from issuing certificate") {
			// 双重检查：字符串匹配且原始字节不匹配
			if bytes.Compare(faterCert.RawSubject, childCert.RawIssuer) != 0 &&
				strings.Compare(faterCert.Subject.String(), childCert.Issuer.String()) == 0 {

				belogs.Debug("VerifyCerByteByX509():Verify fail, subject and issuer by string is equal, but by raw bytes is not equal, will return ok:",
					"   father subject:`"+faterCert.Subject.String()+"`",
					"   child issuer:`"+childCert.Issuer.String()+"`",
					"   father subject == child issuer:", strings.Compare(faterCert.Subject.String(), childCert.Issuer.String()),

					"   father rawsubject:"+convert.Bytes2String(faterCert.RawSubject),
					"   child  rawissuer:"+convert.Bytes2String(childCert.RawIssuer),
					"   father rawsubject == child rawissuer:", bytes.Compare(faterCert.RawSubject, childCert.RawIssuer),

					"   Verify err:", err)
				// 仍然记录警告，但返回ok
				return "ok", nil
			} else {
				belogs.Info("VerifyCerByteByX509():Verify fail, subject and issuer mismatch:",
					"   father subject:`"+faterCert.Subject.String()+"`",
					"   child issuer:`"+childCert.Issuer.String()+"`",
					"   Verify err:", err)
				return "fail", fmt.Errorf("issuer/subject mismatch: %w", err)
			}
		} else if strings.Contains(err.Error(), "certificate has expired or is not yet valid") {
			belogs.Info("VerifyCerByteByX509():Verify fail, certificate has expired or is not yet valid. ",
				"   Now:", convert.Time2StringZone(time.Now()),
				"   child NotBefore:", convert.Time2StringZone(childCert.NotBefore),
				"   child NotAfter:", convert.Time2StringZone(childCert.NotAfter), "   Verify err:", err)
			return "fail", fmt.Errorf("certificate validity error: %w", err)
		} else {
			belogs.Info("VerifyCerByteByX509():Verify fail ",
				"   father subject:`"+faterCert.Subject.String()+"`",
				"   child issuer:`"+childCert.Issuer.String()+"`",
				"   Verify err:", err)
			return "fail", fmt.Errorf("certificate verification failed: %w", err)
		}
	}

	belogs.Debug("VerifyCerByteByX509(): Certificate verified successfully, chain length:", len(chains))
	return "ok", nil
}

// if cert cannot pass verify, just log info level
func VerifyRootCerByOpenssl(rootFile string) (result string, err error) {
	/*

		openssl verify -check_ss_sig -CAfile root.pem.cer  inter.pem.cer
		inter.pem.cer: OK

		openssl verify -check_ss_sig -CAfile inter.pem.cer  inter.pem.cer
		CN = apnic-rpki-root-intermediate, serialNumber = 98142C9D0B41A3B9FB603D769848236FD1F31924
		error 20 at 0 depth lookup: unable to get local issuer certificate
		error inter.pem.cer: verification failed

		openssl x509 -inform DER -in AfriNIC.cer -out AfriNIC.cer.pem
		openssl verify -check_ss_sig -Cafile AfriNIC.cer.pem AfriNIC.cer.pem
	*/
	belogs.Debug("VerifyRootCerByOpenssl():rootFile", rootFile)
	_, file := osutil.Split(rootFile)
	// 创建安全的临时文件（保留安全改进）
	pemFile, err := os.CreateTemp("", file+".pem") // temp file
	if err != nil {
		belogs.Error("VerifyRootCerByOpenssl(): failed to create temp file: ", err, rootFile)
		return "fail", fmt.Errorf("failed to create temp file: %w", err)
	}
	// 显式设置临时文件权限为0600
	if err := os.Chmod(pemFile.Name(), 0600); err != nil {
		pemFile.Close()
		os.Remove(pemFile.Name())
		return "fail", fmt.Errorf("failed to set temp file permissions: %w", err)
	}
	// 安全的延迟清理
	defer func() {
		pemFile.Close()
		os.Remove(pemFile.Name())
	}()

	// cer --> pem（移除safeCommand，恢复原命令执行方式）
	opensslCmd := "openssl"
	path := conf.String("openssl::path")
	if len(path) > 0 {
		opensslCmd = osutil.JoinPathFile(path, opensslCmd)
	}
	belogs.Debug("VerifyRootCerByOpenssl(): cmd: opensslCmd", opensslCmd, "x509", "-inform", "der", "-in", rootFile, "-out", pemFile.Name())
	cmd := exec.Command(opensslCmd, "x509", "-inform", "der", "-in", rootFile, "-out", pemFile.Name())
	output, err := cmd.CombinedOutput()
	if err != nil {
		belogs.Error("VerifyRootCerByOpenssl(): exec x509: err: ", err, ": "+string(output), rootFile)
		return "fail", fmt.Errorf("x509 command failed: %w (output: %s)", err, string(output))
	}
	if len(output) != 0 {
		belogs.Error("VerifyRootCerByOpenssl(): x509 command output: ", string(output), rootFile)
		return "fail", errors.New("convert cer to pem fail")
	}

	// verify（移除safeCommand，恢复原命令执行方式）
	belogs.Debug("VerifyRootCerByOpenssl(): cmd: openssl", opensslCmd, "verify", "-check_ss_sig", "-CAfile", pemFile.Name(), pemFile.Name())
	cmd = exec.Command(opensslCmd, "verify", "-check_ss_sig", "-CAfile", pemFile.Name(), pemFile.Name())
	output, err = cmd.CombinedOutput()
	if err != nil {
		belogs.Error("VerifyRootCerByOpenssl(): exec verify: err: ", err, ": "+string(output), rootFile)
		return "fail", fmt.Errorf("verify command failed: %w (output: %s)", err, string(output))
	}

	out := string(output)
	if !strings.Contains(out, "OK") {
		belogs.Info("VerifyRootCerByOpenssl(): verify pem fail: output: ", out, rootFile)
		return "fail", fmt.Errorf("openssl verify failed: %s", out)
	}

	belogs.Debug("VerifyRootCerByOpenssl(): certificate verified successfully")
	return "ok", nil
}

// if cert cannot pass verify, just log info level
func VerifyCrlByX509(cerFile, crlFile string) (result string, err error) {
	/*
		openssl crl -inform DER -in crl.der -outform PEM -out crl.pem
			openssl verify -crl_check -CAfile crl_chain.pem wikipedia.pem

	*/
	cer, err := ReadFileToCer(cerFile)
	if err != nil {
		belogs.Error("VerifyCrlByX509(): ReadFileToCer fail: err: ", err, cerFile)
		return "fail", fmt.Errorf("failed to read certificate: %w", err)
	}

	crl, err := ReadFileToCrl(crlFile)
	if err != nil {
		belogs.Error("VerifyCrlByX509(): ReadFileToCrl fail: err: ", err, crlFile)
		return "fail", fmt.Errorf("failed to read CRL: %w", err)
	}

	// 验证CRL签名
	err = cer.CheckCRLSignature(crl)
	if err != nil {
		belogs.Info("VerifyCrlByX509(): CheckCRLSignature fail: err: ", err, cerFile, crlFile)
		return "fail", fmt.Errorf("CRL signature verification failed: %w", err)
	}

	// 额外检查：CRL的颁发者应该与证书的颁发者匹配
	certIssuer := cer.Issuer.String()
	crlIssuer := crl.TBSCertList.Issuer.String()
	if certIssuer != crlIssuer {
		belogs.Warn("VerifyCrlByX509(): CRL issuer does not match certificate issuer",
			" certIssuer:", certIssuer, " crlIssuer:", crlIssuer)
		return "fail", fmt.Errorf("CRL issuer mismatch: cert issuer=%s, crl issuer=%s", certIssuer, crlIssuer)
	}

	// 检查CRL是否在有效期内
	now := time.Now()
	if crl.TBSCertList.NextUpdate.Before(now) {
		belogs.Error("VerifyCrlByX509(): CRL has expired")
		return "fail", errors.New("CRL has expired")
	}

	belogs.Debug("VerifyCrlByX509(): CRL verified successfully")
	return "ok", nil
}

// deprecated
func JudgeBelongNic(repoDestPath, filePath string) (nicName string) {
	if repoDestPath == "" || filePath == "" {
		return ""
	}
	if !strings.HasPrefix(filePath, repoDestPath) {
		return ""
	}
	tmp := strings.TrimPrefix(filePath, repoDestPath)
	tmp = strings.TrimPrefix(tmp, "/")
	parts := strings.SplitN(tmp, "/", 2)
	if len(parts) == 0 {
		return ""
	}
	return parts[0]
}
