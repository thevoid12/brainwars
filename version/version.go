package version

import "fmt"

const VersionMajor = 1
const VersionMinor = 1
const VersionPatch = 2
const ReleaseDate = "30-Oct-2024"

func GetLatestVersion() string {
	return "Version: " + fmt.Sprint(VersionMajor) + "." + fmt.Sprint(VersionMinor) + "." + fmt.Sprint(VersionPatch) + " " + ReleaseDate
}
