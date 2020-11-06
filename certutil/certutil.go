package cert

import (
	"bytes"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/pem"
	"errors"
	"io/ioutil"
	"os/exec"
	"strconv"
	"strings"
	"time"

	belogs "github.com/astaxie/beego/logs"
	convert "github.com/cpusoft/goutil/convert"
	osutil "github.com/cpusoft/goutil/osutil"
)

// if cert cannot pass verify, just log info level
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

// if cert cannot pass verify, just log info level
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

// if cert cannot pass verify, just log info level
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
	result, err = VerifyCerByteByX509(fatherFileByte, childFileByte)
	if err != nil {
		belogs.Info("VerifyCerByX509():VerifyCerByteByX509 fail, fatherCertFile:", fatherCertFile, "    childCertFile:", childCertFile, "  err:", err)
		return "fail", err
	}
	return result, nil
}

// if cert cannot pass verify, just log info level
func VerifyEeCertByX509(fatherCertFile string, mftRoaFile string, eeCertStart, eeCertEnd uint64) (result string, err error) {
	fatherFileByte, err := ioutil.ReadFile(fatherCertFile)
	if err != nil {
		belogs.Error("VerifyEeCertByX509():read father, fatherCertFile:", fatherCertFile, "   mftRoaFile:", mftRoaFile, "     ReadFile err:", err)
		return "fail", err
	}
	mftRoaFileByte, err := ioutil.ReadFile(mftRoaFile)
	if err != nil {
		belogs.Error("VerifyEeCertByX509():ReadFile fail, fatherCertFile:", fatherCertFile, "    mftRoaFile:", mftRoaFile, "   ReadFile err:", err)
		return "fail", err
	}
	if eeCertStart >= eeCertEnd || len(mftRoaFileByte) < int(eeCertEnd) {
		belogs.Error("VerifyEeCertByX509():mftRoaFileByte[eeCertStart:eeCertEnd] fatherCertFile:", fatherCertFile, "    mftRoaFile:", mftRoaFile,
			"   eeCertStart :", eeCertStart, "   eeCertEnd:", eeCertEnd, "   len(mftRoaFileByte):", len(mftRoaFileByte))
		return "fail", errors.New("get eecert fail,  eeCertStart is " + strconv.Itoa(int(eeCertStart)) +
			",   eeCertEnd is " + strconv.Itoa(int(eeCertEnd)) + ", but len(mftRoaFileByte) is " + strconv.Itoa(len(mftRoaFileByte)))
	}
	//belogs.Debug("VerifyEeCertByX509():mftRoaFileByte[eeCertStart:eeCertEnd]:", mftRoaFile,
	//	"   eeCertStart :", eeCertStart, "   eeCertEnd:", eeCertEnd, "   len(mftRoaFileByte):", len(mftRoaFileByte))
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
		belogs.Error("VerifyCerByteByX509():parse fatherCertFile fail, ParseCertificate err:", err)
		return "fail", err
	}
	faterCert.UnhandledCriticalExtensions = make([]asn1.ObjectIdentifier, 0)
	belogs.Debug("VerifyCerByteByX509():father issuer:", faterCert.Issuer.String(), "   subject:", faterCert.Subject.String())

	fatherPool.AddCert(faterCert)

	childCert, err := x509.ParseCertificate(childCertByte)
	if err != nil {
		belogs.Error("VerifyCerByteByX509():parse childCertFile fail:  father issuer:", faterCert.Issuer.String(), " ParseCertificate err:", err)
		return "fail", err
	}
	childCert.UnhandledCriticalExtensions = make([]asn1.ObjectIdentifier, 0)
	belogs.Debug("VerifyCerByteByX509():child issuer:", childCert.Issuer.String(), "   childCert:", childCert.Subject.String())

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
	if _, err := childCert.Verify(opts); err != nil {

		if strings.Contains(err.Error(), "issuer name does not match subject from issuing certificate") {
			// compare by subject and issuer by string
			if strings.Compare(faterCert.Subject.String(), childCert.Issuer.String()) == 0 {
				belogs.Debug("VerifyCerByteByX509():Verify fail, subject and issuer by string is equal, but by raw bytes is not equal, will return ok:",
					"\nfather subject:`"+faterCert.Subject.String()+"`",
					"\nchild issuer:`"+childCert.Issuer.String()+"`",
					"\nfather subject == child issuer:", strings.Compare(faterCert.Subject.String(), childCert.Issuer.String()),

					"\nfather rawsubject:"+convert.Bytes2String(faterCert.RawSubject),
					"\nchild  rawissuer:"+convert.Bytes2String(childCert.RawIssuer),
					"\nfather rawsubject == child rawissuer:", bytes.Compare(faterCert.RawSubject, childCert.RawIssuer),

					"\nVerify err:", err)
				return "ok", nil
			} else {
				belogs.Info("VerifyCerByteByX509():Verify fail, subject and issuer by string is not equal:",
					"\nfather subject:`"+faterCert.Subject.String()+"`",
					"\nchild issuer:`"+childCert.Issuer.String()+"`",
					"\nfather subject == child issuer:", strings.Compare(faterCert.Subject.String(), childCert.Issuer.String()),
					"\nVerify err:", err)
				err = errors.New(err.Error() + ".  Father subject is '" + faterCert.Subject.String() + "', and child issuer is '" + childCert.Issuer.String() + "'")
				return "fail", err
			}
		} else if strings.Contains(err.Error(), "certificate has expired or is not yet valid") {
			belogs.Info("VerifyCerByteByX509():Verify fail, certificate has expired or is not yet valid. ",
				"\nNow:", convert.Time2StringZone(time.Now()),
				"\nchild NotBefore:", convert.Time2StringZone(childCert.NotBefore),
				"\nchild NotAfter:", convert.Time2StringZone(childCert.NotAfter), "\nVerify err:", err)
			err = errors.New(err.Error() + ".   NotBefore is '" + convert.Time2StringZone(childCert.NotBefore) + "', and NotAfter is '" + convert.Time2StringZone(childCert.NotAfter) + "'")
			return "fail", err
		} else {
			belogs.Info("VerifyCerByteByX509():Verify fail ",
				"\nfather subject:`"+faterCert.Subject.String()+"`",
				"\nchild issuer:`"+childCert.Issuer.String()+"`",
				"\nVerify err:", err)
			return "fail", err
		}

	}
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
		belogs.Info("VerifyRootCerByOpenssl(): verify pem fail: err: ", string(output), rootFile)
		return "fail", errors.New("verify pem fail")
	}
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
		return "fail", err
	}

	crl, err := ReadFileToCrl(crlFile)
	if err != nil {
		belogs.Error("VerifyCrlByX509(): ReadFileToCrl fail: err: ", err, crlFile)
		return "fail", err
	}

	err = cer.CheckCRLSignature(crl)
	if err != nil {
		belogs.Info("VerifyCrlByX509(): CheckCRLSignature fail: err: ", err, cerFile, crlFile)
		return "fail", err
	}
	return "ok", nil
}

//deprecated
func JudgeBelongNic(repoDestPath, filePath string) (nicName string) {
	if !strings.HasPrefix(filePath, repoDestPath) {
		return ""
	}
	tmp := strings.Replace(filePath, repoDestPath+"/", "", -1)
	return strings.Split(tmp, "/")[0]
}
