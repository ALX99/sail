package util

import (
	"path/filepath"
	"strings"
)

// Common Nerd Font icons
const (
	IconDirectory  = ""
	IconFile       = ""
	IconGo         = ""
	IconZip        = ""
	IconGit        = ""
	IconDocker     = ""
	IconJavascript = ""
	IconTypescript = ""
	IconHtml       = ""
	IconCss        = ""
	IconJson       = ""
	IconMd         = ""
	IconImage      = ""
	IconVideo      = ""
	IconAudio      = "cn"
	IconShell      = ""
)

func GetIcon(name string, isDir bool) string {
	if isDir {
		return IconDirectory
	}

	ext := strings.ToLower(filepath.Ext(name))
	switch ext {
	case ".go":
		return IconGo
	case ".zip", ".tar", ".gz", ".xz", ".7z", ".rar":
		return IconZip
	case ".js", ".cjs", ".mjs":
		return IconJavascript
	case ".ts", ".tsx":
		return IconTypescript
	case ".html", ".htm":
		return IconHtml
	case ".css", ".scss", ".sass":
		return IconCss
	case ".json", ".yaml", ".yml", ".toml":
		return IconJson
	case ".md", ".markdown":
		return IconMd
	case ".png", ".jpg", ".jpeg", ".gif", ".webp", ".svg", ".bmp":
		return IconImage
	case ".mp4", ".mkv", ".mov", ".avi", ".webm":
		return IconVideo
	case ".mp3", ".wav", ".flac", ".ogg":
		return IconAudio
	case ".sh", ".bash", ".zsh":
		return IconShell
	case ".dockerfile":
		return IconDocker
	case ".gitignore", ".gitattributes":
		return IconGit
	default:
		if strings.Contains(strings.ToLower(name), "docker") {
			return IconDocker
		}
		if strings.HasPrefix(name, "git") {
			return IconGit
		}
		return IconFile
	}
}
