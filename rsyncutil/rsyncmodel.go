package rsyncutil

type RsyncFileHash struct {
	FilePath    string `json:"filePath" xorm:"filePath varchar(512)"`
	FileName    string `json:"fileName" xorm:"fileName varchar(128)"`
	FileHash    string `json:"fileHash" xorm:"fileHash varchar(512)"`
	JsonAll     string `json:"jsonAll" xorm:"jsonAll json"`
	LastJsonAll string `json:"lastJsonAll" xorm:"lastJsonAll json"`
	// cer/roa/mft/crl, no dot
	FileType string `json:"fileType" xorm:"fileType  varchar(16)"`
}

const (
	RSYNC_TYPE_ADD       = "add"
	RSYNC_TYPE_DEL       = "del"
	RSYNC_TYPE_UPDATE    = "update"
	RSYNC_TYPE_MKDIR     = "mkdir"
	RSYNC_TYPE_IGNORE    = "ignore"
	RSYNC_TYPE_JUST_SYNC = "justsync" //The file itself is not updated, just used to trigger sync sub-dir , so no need save to db

	RSYNC_LOG_PREFIX    = 12
	RSYNC_LOG_FILE_NAME = "rsync.log"

	RSYNC_TIMEOUT_SEC    = "12"
	RSYNC_CONTIMEOUT_SEC = "12"
)

type RsyncClientConfig struct {
	Timeout    string `json:"timeout"`
	ConTimeout string `json:"conTimeout"`
}

func NewRsyncClientConfig(timeoutSec, conTimeoutSec string) *RsyncClientConfig {
	r := new(RsyncClientConfig)
	r.Timeout = timeoutSec       //RSYNC_TIMEOUT_SEC
	r.ConTimeout = conTimeoutSec //RSYNC_CONTIMEOUT_SEC
	return r
}
