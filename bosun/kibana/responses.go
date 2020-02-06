package kibana

import "github.com/Masterminds/semver/v3"

type Info struct {
	Name    string `json:"name"`
	UUID    string `json:"uuid"`
	Version struct {
		Number        string `json:"number"`
		BuildHash     string `json:"build_hash"`
		BuildNumber   int    `json:"build_number"`
		BuildSnapshot bool   `json:"build_snapshot"`
	} `json:"version"`
}

func (i Info) GetSemVar() (*semver.Version, error) {
	return semver.NewVersion(i.Version.Number)
}

type Index struct {
	Name string `json:"index"`
}
