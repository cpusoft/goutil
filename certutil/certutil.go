package cert

import (
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/pem"
	"errors"
	"io/ioutil"
	"os/exec"
	"strconv"
	"strings"

	belogs "github.com/astaxie/beego/logs"
	osutil "github.com/cpusoft/goutil/osutil"
)

func ReadFileToCer(fileName string) (*x509.Certificate, error) {
	p, fileByte, err := ReadFileToByte(fileName)
	if err != nil {
		return nil, err
	}
	if len(fileByte) > 0 {
		return x509.ParseCertificate(fileByte)
	} else if p != nil {
		return x509.ParseCertificate(p.Bytes)
	}
	return nil, errors.New("unknown cert type")
}

func ReadFileToCrl(fileName string) (*pkix.CertificateList, error) {
	_, fileByte, err := ReadFileToByte(fileName)
	if err != nil {
		return nil, err
	}
	if len(fileByte) > 0 {
		return x509.ParseCRL(fileByte)
	}
	return nil, errors.New("unknown cert type")
}

//no PEM data is found, p is nil and the whole of the input is returned in
func ReadFileToByte(fileName string) (p *pem.Block, fileByte []byte, err error) {
	buf, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, nil, err
	}

	p, fileByte = pem.Decode(buf)
	return p, fileByte, nil
}

func VerifyCerByX509(fatherCertFile string, childCertFile string) (result string, err error) {
	fatherFileByte, err := ioutil.ReadFile(fatherCertFile)
	if err != nil {
		belogs.Error("VerifyCerByX509():fatherCertFile:", fatherCertFile, "   ReadFile err:", err)
		return "fail", err
	}
	childFileByte, err := ioutil.ReadFile(childCertFile)
	if err != nil {
		belogs.Error("VerifyCerByX509():childCertFile:", childCertFile, "   ReadFile err:", err)
		return "fail", err
	}
	return VerifyCerByteByX509(fatherFileByte, childFileByte)
}

func VerifyEeCertByX509(fatherCertFile string, mftRoaFile string, eeCertStart, eeCertEnd uint64) (result string, err error) {
	fatherFileByte, err := ioutil.ReadFile(fatherCertFile)
	if err != nil {
		belogs.Error("VerifyEeCertByX509():fatherCertFile:", fatherCertFile, "   ReadFile err:", err)
		return "fail", err
	}
	mftRoaFileByte, err := ioutil.ReadFile(mftRoaFile)
	if err != nil {
		belogs.Error("VerifyEeCertByX509():childCertFile:", mftRoaFile, "   ReadFile err:", err)
		return "fail", err
	}
	if eeCertStart >= eeCertEnd || len(mftRoaFileByte) < int(eeCertEnd) {
		belogs.Error("VerifyEeCertByX509():mftRoaFileByte[eeCertStart:eeCertEnd] err:", mftRoaFile,
			"   eeCertStart :", eeCertStart, "   eeCertEnd:", eeCertEnd, "   len(mftRoaFileByte):", len(mftRoaFileByte))
		return "fail", errors.New("get eecert fail,  eeCertStart is " + strconv.Itoa(int(eeCertStart)) +
			",   eeCertEnd is " + strconv.Itoa(int(eeCertEnd)) + ", but len(mftRoaFileByte) is " + strconv.Itoa(len(mftRoaFileByte)))
	}
	belogs.Debug("VerifyEeCertByX509():mftRoaFileByte[eeCertStart:eeCertEnd]:", mftRoaFile,
		"   eeCertStart :", eeCertStart, "   eeCertEnd:", eeCertEnd, "   len(mftRoaFileByte):", len(mftRoaFileByte))
	b := mftRoaFileByte[eeCertStart:eeCertEnd]
	return VerifyCerByteByX509(fatherFileByte, b)
}

// fatherCerFile is as root, childCerFile is to be verified
// result: ok/fail
func VerifyCerByteByX509(fatherCertByte []byte, childCertByte []byte) (result string, err error) {
	belogs.Debug("VerifyCerByteByX509():fatherCertByte:", len(fatherCertByte), "   childCertByte:", len(childCertByte))

	fatherPool := x509.NewCertPool()

	faterCert, err := x509.ParseCertificate(fatherCertByte)
	if err != nil {
		belogs.Error("VerifyCerByteByX509():fatherCertFile   ParseCertificate err:", err)
		return "fail", err
	}
	faterCert.UnhandledCriticalExtensions = make([]asn1.ObjectIdentifier, 0)
	belogs.Debug("VerifyCerByteByX509():father issuer:", faterCert.Issuer.String(), "   subject:", faterCert.Subject.String())
	fatherPool.AddCert(faterCert)

	childCert, err := x509.ParseCertificate(childCertByte)
	if err != nil {
		belogs.Error("VerifyCerByteByX509():childCertFile:  ParseCertificate err:", err)
		return "fail", err
	}
	childCert.UnhandledCriticalExtensions = make([]asn1.ObjectIdentifier, 0)
	belogs.Debug("VerifyCerByteByX509():child issuer:", childCert.Issuer.String(), "   childCert:", childCert.Subject.String())

	opts := x509.VerifyOptions{
		Roots: fatherPool,
		//Intermediates: inter,
		KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageAny},
		//KeyUsages: []x509.ExtKeyUsage{x509.KeyUsageCertSign},
	}
	if _, err := childCert.Verify(opts); err != nil {
		belogs.Error("VerifyCerByteByX509():Verify err:", err)
		return "fail", err
	}
	return "ok", nil
}

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
	pemFile, err := ioutil.TempFile("", file+".pem") // temp file
	if err != nil {
		belogs.Error("VerifyRootCerByOpenssl(): exec.Command: err: ", err, rootFile)
		return "fail", err
	}
	defer osutil.CloseAndRemoveFile(pemFile)

	// cer --> pem
	belogs.Debug("VerifyRootCerByOpenssl(): cmd: openssl", "x509", "-inform", "der", "-in", rootFile, "-out", pemFile)
	cmd := exec.Command("openssl", "x509", "-inform", "der", "-in", rootFile, "-out", pemFile.Name())
	output, err := cmd.CombinedOutput()
	if err != nil {
		belogs.Error("VerifyRootCerByOpenssl(): exec x509: err: ", err, ": "+string(output), rootFile)
		return "fail", err
	}
	if len(output) != 0 {
		belogs.Error("VerifyRootCerByOpenssl(): convert cer to pem fail: err:", string(output), rootFile)
		return "fail", errors.New("convert cer to pem fail")
	}

	// verify
	belogs.Debug("VerifyRootCerByOpenssl(): cmd: openssl", "verify", "-check_ss_sig", "-CAfile", pemFile.Name(), pemFile.Name())
	cmd = exec.Command("openssl", "verify", "-check_ss_sig", "-CAfile", pemFile.Name(), pemFile.Name())
	output, err = cmd.CombinedOutput()
	if err != nil {
		belogs.Error("VerifyRootCerByOpenssl(): exec verify: err: ", err, ": "+string(output), rootFile)
		return "fail", err
	}
	out := string(output)
	if !strings.Contains(out, "OK") {
		belogs.Error("VerifyRootCerByOpenssl(): verify pem fail: err: ", string(output), rootFile)
		return "fail", errors.New("verify pem fail")
	}
	return "ok", nil
}

func VerifyCrlByX509(cerFile, crlFile string) (result string, err error) {
	/*
		openssl crl -inform DER -in crl.der -outform PEM -out crl.pem
			openssl verify -crl_check -CAfile crl_chain.pem wikipedia.pem

	*/

	cer, err := ReadFileToCer(cerFile)
	if err != nil {
		belogs.Error("VerifyCrlByX509(): ReadFileToCer fail: err: ", err, cerFile)
		return "fail", err
	}

	crl, err := ReadFileToCrl(crlFile)
	if err != nil {
		belogs.Error("VerifyCrlByX509(): ReadFileToCrl fail: err: ", err, crlFile)
		return "fail", err
	}

	err = cer.CheckCRLSignature(crl)
	if err != nil {
		belogs.Error("VerifyCrlByX509(): CheckCRLSignature fail: err: ", err, cerFile, crlFile)
		return "fail", err
	}
	return "ok", nil
}
