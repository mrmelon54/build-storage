package mc

import "archive/zip"

func DetectMcPlatforms(zr *zip.Reader) []string {
	// check for forge, fabric and quilt metadata files
	platforms := make([]string, 0, 3)
	if hasFileInZip(zr, "META-INF/mods.toml") {
		platforms = append(platforms, "forge")
	}
	if hasFileInZip(zr, "fabric.mod.json") {
		platforms = append(platforms, "fabric")
	}
	if hasFileInZip(zr, "quilt.mod.json") {
		platforms = append(platforms, "quilt")
	}
	return platforms
}

func hasFileInZip(zr *zip.Reader, p string) bool {
	_, err := zr.Open(p)
	return err == nil
}
