package types

import "github.com/coreos/go-semver/semver"

const (
	ClientVersion    = "0.3.2"
	ClientMinVersion = "0.3.0"
	ServerVersion    = "0.3.2"
)

type VersionResponse struct {
	LatestClientVersion string `json:"latest_client_version"`
	ServerVersion       string `json:"server_version"`
	MinClientVersion    string `json:"min_client_version"`
	UpgradeScriptUrl    string `json:"upgrade_script"`
}

func IsVersionLess(resp *VersionResponse) bool {
	client := semver.New(ClientVersion)
	curClient := semver.New(resp.LatestClientVersion)
	if client.LessThan(*curClient) {
		return true
	}
	return false
}
func IsVersionMin(resp *VersionResponse) bool {
	client := semver.New(ClientVersion)
	minClient := semver.New(resp.MinClientVersion)
	if client.LessThan(*minClient) {
		return false
	}
	return true
}
