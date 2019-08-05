package cert

import (
	"crypto/x509"
	"encoding/asn1"
	"errors"
	"io/ioutil"
	"os/exec"
	"strings"

	belogs "github.com/astaxie/beego/logs"
	osutil "github.com/cpusoft/goutil/osutil"
)

func VerifyCertByX509(fatherCertFile string, childCertFile string) (result string, err error) {
	fatherFileByte, err := ioutil.ReadFile(fatherCertFile)
	if err != nil {
		belogs.Error("VerifyCertsByX509():fatherCertFile:", fatherCertFile, "   ReadFile err:", err)
		return "fail", err
	}
	childFileByte, err := ioutil.ReadFile(childCertFile)
	if err != nil {
		belogs.Error("VerifyCertsByX509():childCertFile:", childCertFile, "   ReadFile err:", err)
		return "fail", err
	}
	return VerifyCertByteByX509(fatherFileByte, childFileByte)
}

func VerifyEeCertByX509(fatherCertFile string, mftRoaFile string, eeCertStart, eeCertEnd uint64) (result string, err error) {
	fatherFileByte, err := ioutil.ReadFile(fatherCertFile)
	if err != nil {
		belogs.Error("VerifyCertsByX509():fatherCertFile:", fatherCertFile, "   ReadFile err:", err)
		return "fail", err
	}
	mftRoaFileByte, err := ioutil.ReadFile(mftRoaFile)
	if err != nil {
		belogs.Error("VerifyCertsByX509():childCertFile:", mftRoaFile, "   ReadFile err:", err)
		return "fail", err
	}
	b := mftRoaFileByte[eeCertStart:eeCertEnd]
	return VerifyCertByteByX509(fatherFileByte, b)
}

// fatherCerFile is as root, childCerFile is to be verified
// result: ok/fail
func VerifyCertByteByX509(fatherCertByte []byte, childCertByte []byte) (result string, err error) {
	belogs.Debug("VerifyCertBytesByX509():fatherCertByte:", len(fatherCertByte), "   childCertByte:", len(childCertByte))

	fatherPool := x509.NewCertPool()

	faterCert, err := x509.ParseCertificate(fatherCertByte)
	if err != nil {
		belogs.Error("VerifyCertsByX509():fatherCertFile   ParseCertificate err:", err)
		return "fail", err
	}
	faterCert.UnhandledCriticalExtensions = make([]asn1.ObjectIdentifier, 0)
	belogs.Debug("VerifyCertsByX509():father issuer:", faterCert.Issuer.String(), "   subject:", faterCert.Subject.String())
	fatherPool.AddCert(faterCert)

	childCert, err := x509.ParseCertificate(childCertByte)
	if err != nil {
		belogs.Error("VerifyCertsByX509():childCertFile:  ParseCertificate err:", err)
		return "fail", err
	}
	childCert.UnhandledCriticalExtensions = make([]asn1.ObjectIdentifier, 0)
	belogs.Debug("VerifyCertsByX509():child issuer:", childCert.Issuer.String(), "   childCert:", childCert.Subject.String())

	opts := x509.VerifyOptions{
		Roots: fatherPool,
		//Intermediates: inter,
		KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageAny},
		//KeyUsages: []x509.ExtKeyUsage{x509.KeyUsageCertSign},
	}
	if _, err := childCert.Verify(opts); err != nil {
		belogs.Error("VerifyCertsByX509():Verify err:", err)
		return "fail", err
	}
	return "ok", nil
}

func VerifyRootCertByOpenssl(rootFile string) (result string, err error) {
	/*

			G:\Download\cert\verify\2>openssl verify -check_ss_sig -CAfile root.pem.cer  inter.pem.cer
		inter.pem.cer: OK

		G:\Download\cert\verify\2>openssl verify -check_ss_sig -CAfile inter.pem.cer  inter.pem.cer
		CN = apnic-rpki-root-intermediate, serialNumber = 98142C9D0B41A3B9FB603D769848236FD1F31924
		error 20 at 0 depth lookup: unable to get local issuer certificate
		error inter.pem.cer: verification failed
	*/
	/*
		openssl x509 -inform DER -in AfriNIC.cer -out AfriNIC.cer.pem
		openssl verify -check_ss_sig -Cafile AfriNIC.cer.pem AfriNIC.cer.pem
	*/
	belogs.Debug("VerifyRootCertByOpenssl():rootFile", rootFile)
	_, file := osutil.Split(rootFile)
	pemFile, err := ioutil.TempFile("", file+".pem") // temp file
	if err != nil {
		belogs.Error("VerifyRootCertByOpenssl(): exec.Command: err: ", err, rootFile)
		return "fail", err
	}
	defer osutil.CloseAndRemoveFile(pemFile)

	// cer --> pem
	belogs.Debug("VerifyRootCertByOpenssl(): cmd: openssl", "x509", "-inform", "der", "-in", rootFile, "-out", pemFile)
	cmd := exec.Command("openssl", "x509", "-inform", "der", "-in", rootFile, "-out", pemFile.Name())
	output, err := cmd.CombinedOutput()
	if err != nil {
		belogs.Error("VerifyRootCertByOpenssl(): exec x509: err: ", err, ": "+string(output), rootFile)
		return "fail", err
	}
	if len(output) != 0 {
		belogs.Error("VerifyRootCertByOpenssl(): convert cer to pem fail: err:", string(output), rootFile)
		return "fail", errors.New("convert cer to pem fail")
	}

	// verify
	belogs.Debug("VerifyRootCertByOpenssl(): cmd: openssl", "verify", "-check_ss_sig", "-CAfile", pemFile.Name(), pemFile.Name())
	cmd = exec.Command("openssl", "verify", "-check_ss_sig", "-CAfile", pemFile.Name(), pemFile.Name())
	output, err = cmd.CombinedOutput()
	if err != nil {
		belogs.Error("VerifyRootCertByOpenssl(): exec verify: err: ", err, ": "+string(output), rootFile)
		return "fail", err
	}
	out := string(output)
	if !strings.Contains(out, "OK") {
		belogs.Error("VerifyRootCertByOpenssl(): verify pem fail: err: ", string(output), rootFile)
		return "fail", errors.New("verify pem fail")
	}
	return "ok", nil
}
