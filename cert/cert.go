package cert

import (
	"crypto/x509"
	"encoding/asn1"
	"io/ioutil"

	belogs "github.com/astaxie/beego/logs"
)

func VerifyCertByX509(fatherCertFile string, childCertFile string) (result string, err error) {
	fatherFileByte, err := ioutil.ReadFile(fatherCertFile)
	if err != nil {
		belogs.Debug("VerifyCertsByX509():fatherCertFile:", fatherCertFile, "   ReadFile err:", err)
		return "fail", err
	}
	childFileByte, err := ioutil.ReadFile(childCertFile)
	if err != nil {
		belogs.Debug("VerifyCertsByX509():childCertFile:", childCertFile, "   ReadFile err:", err)
		return "fail", err
	}
	return VerifyCertByteByX509(fatherFileByte, childFileByte)
}

func VerifyEeCertByX509(fatherCertFile string, mftRoaFile string, eeCertStart, eeCertEnd uint64) (result string, err error) {
	fatherFileByte, err := ioutil.ReadFile(fatherCertFile)
	if err != nil {
		belogs.Debug("VerifyCertsByX509():fatherCertFile:", fatherCertFile, "   ReadFile err:", err)
		return "fail", err
	}
	mftRoaFileByte, err := ioutil.ReadFile(mftRoaFile)
	if err != nil {
		belogs.Debug("VerifyCertsByX509():childCertFile:", mftRoaFile, "   ReadFile err:", err)
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
		belogs.Debug("VerifyCertsByX509():fatherCertFile   ParseCertificate err:", err)
		return "fail", err
	}
	faterCert.UnhandledCriticalExtensions = make([]asn1.ObjectIdentifier, 0)
	belogs.Debug("VerifyCertsByX509():father issuer:", faterCert.Issuer.String(), "   subject:", faterCert.Subject.String())
	fatherPool.AddCert(faterCert)

	childCert, err := x509.ParseCertificate(childCertByte)
	if err != nil {
		belogs.Debug("VerifyCertsByX509():childCertFile:  ParseCertificate err:", err)
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
		belogs.Debug("VerifyCertsByX509():Verify err:", err)
		return "fail", err
	}
	return "ok", nil
}
