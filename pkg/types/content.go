package types

import "strings"

type ContentType struct {
	FileExtension string
	LongName      string
}

var (
	html = ContentType{
		FileExtension: "html",
		LongName:      "html",
	}
	markdown = ContentType{
		FileExtension: "md",
		LongName:      "markdown",
	}
	bash = ContentType{
		FileExtension: "sh",
		LongName:      "bash",
	}
	python = ContentType{
		FileExtension: "py",
		LongName:      "python",
	}
	javascript = ContentType{
		FileExtension: "js",
		LongName:      "javascript",
	}
	golang = ContentType{
		FileExtension: "go",
		LongName:      "golang",
	}

	contentTypes []ContentType = []ContentType{html, markdown, bash, python, golang, javascript}
)

func ResolveContentType(userType string) *ContentType {
	for _, ct := range contentTypes {
		stl := strings.ToLower(userType)
		if stl == ct.FileExtension || stl == ct.LongName {
			return &ct
		}
	}
	return nil
}
