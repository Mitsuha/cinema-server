package model

type ClientVersion struct {
	Version     string `json:"version" yaml:"version"`
	BuildNumber int    `json:"buildNumber" yaml:"buildNumber"`
	Description string `json:"description" yaml:"description"`
	DownloadUrl string `json:"downloadUrl" yaml:"downloadUrl"`
	Forcibly    bool   `json:"forcibly" yaml:"forcibly"`
}
