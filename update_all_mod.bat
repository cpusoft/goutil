@echo off
set pwd=%cd% 

echo %pwd%

cd %pwd%\asn1util
call:goUpdateFunc asn1util

cd %pwd%\asnutil
call:goUpdateFunc asnutil

cd %pwd%\base64util
call:goUpdateFunc base64util

cd %pwd%\beconfig
call:goUpdateFunc beconfig

cd %pwd%\belogs
call:goUpdateFunc belogs

cd %pwd%\bitutil
call:goUpdateFunc bitutil

cd %pwd%\certutil
call:goUpdateFunc certutil

cd %pwd%\conf
call:goUpdateFunc conf

cd %pwd%\convert
call:goUpdateFunc convert

cd %pwd%\datetime
call:goUpdateFunc datetime

cd %pwd%\db
call:goUpdateFunc db

cd %pwd%\dbconf
call:goUpdateFunc dbconf

cd %pwd%\errorutil
call:goUpdateFunc errorutil

cd %pwd%\executil
call:goUpdateFunc executil

cd %pwd%\fileutil
call:goUpdateFunc fileutil

cd %pwd%\ginserver
call:goUpdateFunc ginserver

cd %pwd%\ginsession
call:goUpdateFunc ginsession

cd %pwd%\hashutil
call:goUpdateFunc hashutil

cd %pwd%\httpclient
call:goUpdateFunc httpclient

cd %pwd%\httpserver
call:goUpdateFunc httpserver

cd %pwd%\iputil
call:goUpdateFunc iputil

cd %pwd%\jsonutil
call:goUpdateFunc jsonutil

cd %pwd%\logs
call:goUpdateFunc logs

cd %pwd%\opensslutil
call:goUpdateFunc opensslutil

cd %pwd%\osutil
call:goUpdateFunc osutil

cd %pwd%\passwordutil
call:goUpdateFunc passwordutil

cd %pwd%\randutil
call:goUpdateFunc randutil

cd %pwd%\redisutil
call:goUpdateFunc redisutil

cd %pwd%\regexputil
call:goUpdateFunc regexputil

cd %pwd%\rrdputil
call:goUpdateFunc rrdputil

cd %pwd%\rsyncutil
call:goUpdateFunc rsyncutil

cd %pwd%\stringutil
call:goUpdateFunc stringutil

cd %pwd%\talutil
call:goUpdateFunc talutil

cd %pwd%\tcpserver
call:goUpdateFunc tcpserver

cd %pwd%\tcpudputil
call:goUpdateFunc tcpudputil

cd %pwd%\urlutil
call:goUpdateFunc urlutil

cd %pwd%\uuidutil
call:goUpdateFunc uuidutil

cd %pwd%\xmlutil
call:goUpdateFunc xmlutil

cd %pwd%\xormdb
call:goUpdateFunc xormdb


cd %pwd%
echo "go mod tidy"
go mod tidy

EXIT /B %ERRORLEVEL% 

:goUpdateFunc
echo "go get && go build" %~1 
go get -u
go build
EXIT /B 0
	

