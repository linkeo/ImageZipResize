package flatten

import (
	"ImageZipResize/util/fileutil"
	"path/filepath"
	"strings"
)

const tag = ".flatten"
const delimiter = "-#-"

func flattenPath(path string) (result string, changed bool) {
	ext := filepath.Ext(path)
	if strings.HasSuffix(path, tag+ext) {
		return
	}
	result = path
	result = strings.ReplaceAll(result, "#", "##")
	result = strings.ReplaceAll(result, fileutil.Separator, delimiter)
	result = strings.TrimSuffix(result, ext) + tag + ext
	changed = true
	return
}

func expandPath(path string) (result string, changed bool) {
	ext := filepath.Ext(path)
	if !strings.HasSuffix(path, tag+ext) {
		return
	}
	result = path
	result = strings.TrimSuffix(result, tag+ext) + ext
	result = strings.ReplaceAll(result, delimiter, fileutil.Separator)
	result = strings.ReplaceAll(result, "##", "#")
	changed = true
	return
}
