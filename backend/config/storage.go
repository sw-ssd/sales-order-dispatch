package config

// Storage 檔案儲存設定。
type Storage struct {
	StorageRoot string `envconfig:"STORAGE_ROOT" default:"./data/files"`
}
