package style

import (
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

const defaultColours = "di=01;34:ln=01;36:mh=00:pi=40;33:so=01;35:do=01;35:bd=40;33;01:cd=40;33;01:or=40;31;01:mi=00:su=37;41:sg=30;43:ca=00:tw=30;42:ow=34;42:st=37;44:ex=01;32:*.tar=01;31:*.tgz=01;31:*.arc=01;31:*.arj=01;31:*.taz=01;31:*.lha=01;31:*.lz4=01;31:*.lzh=01;31:*.lzma=01;31:*.tlz=01;31:*.txz=01;31:*.tzo=01;31:*.t7z=01;31:*.zip=01;31:*.z=01;31:*.dz=01;31:*.gz=01;31:*.lrz=01;31:*.lz=01;31:*.lzo=01;31:*.xz=01;31:*.zst=01;31:*.tzst=01;31:*.bz2=01;31:*.bz=01;31:*.tbz=01;31:*.tbz2=01;31:*.tz=01;31:*.deb=01;31:*.rpm=01;31:*.jar=01;31:*.war=01;31:*.ear=01;31:*.sar=01;31:*.rar=01;31:*.alz=01;31:*.ace=01;31:*.zoo=01;31:*.cpio=01;31:*.7z=01;31:*.rz=01;31:*.cab=01;31:*.wim=01;31:*.swm=01;31:*.dwm=01;31:*.esd=01;31:*.avif=01;35:*.jpg=01;35:*.jpeg=01;35:*.mjpg=01;35:*.mjpeg=01;35:*.gif=01;35:*.bmp=01;35:*.pbm=01;35:*.pgm=01;35:*.ppm=01;35:*.tga=01;35:*.xbm=01;35:*.xpm=01;35:*.tif=01;35:*.tiff=01;35:*.png=01;35:*.svg=01;35:*.svgz=01;35:*.mng=01;35:*.pcx=01;35:*.mov=01;35:*.mpg=01;35:*.mpeg=01;35:*.m2v=01;35:*.mkv=01;35:*.webm=01;35:*.webp=01;35:*.ogm=01;35:*.mp4=01;35:*.m4v=01;35:*.mp4v=01;35:*.vob=01;35:*.qt=01;35:*.nuv=01;35:*.wmv=01;35:*.asf=01;35:*.rm=01;35:*.rmvb=01;35:*.flc=01;35:*.avi=01;35:*.fli=01;35:*.flv=01;35:*.gl=01;35:*.dl=01;35:*.xcf=01;35:*.xwd=01;35:*.yuv=01;35:*.cgm=01;35:*.emf=01;35:*.ogv=01;35:*.ogx=01;35:*.aac=00;36:*.au=00;36:*.flac=00;36:*.m4a=00;36:*.mid=00;36:*.midi=00;36:*.mka=00;36:*.mp3=00;36:*.mpc=00;36:*.ogg=00;36:*.ra=00;36:*.wav=00;36:*.oga=00;36:*.opus=00;36:*.spx=00;36:*.xspf=00;36:*~=00;90:*#=00;90:*.bak=00;90:*.old=00;90:*.orig=00;90:*.part=00;90:*.rej=00;90:*.swp=00;90:*.tmp=00;90:*.dpkg-dist=00;90:*.dpkg-old=00;90:*.ucf-dist=00;90:*.ucf-new=00;90:*.ucf-old=00;90:*.rpmnew=00;90:*.rpmorig=00;90:*.rpmsave=00;90:"

var knownKeys = map[string]bool{
	"no": true, "fi": true, "di": true, "ln": true, "pi": true,
	"do": true, "bd": true, "cd": true, "or": true, "so": true,
	"su": true, "sg": true, "tw": true, "ow": true, "st": true,
	"ex": true, "mi": true, "mh": true, "ca": true,
}

type Styles struct {
	keys       map[string]lipgloss.Style
	extensions []extension
}

type extension struct {
	pattern string
	style   lipgloss.Style
}

func NewStyles(colours string) *Styles {
	if colours == "" {
		colours = defaultColours
	}

	s := Styles{
		keys: make(map[string]lipgloss.Style),
	}

	for str := range strings.SplitSeq(colours, ":") {
		if strings.TrimSpace(str) == "" {
			continue
		}
		entry, st, err := parseStyle(str)
		if err != nil {
			slog.Warn("Error while parsing LS_COLORS", "entry", entry, "error", err)
			continue
		}

		if knownKeys[entry] {
			s.keys[entry] = st
		} else {
			s.extensions = append(s.extensions, extension{pattern: entry, style: st})
		}
	}
	return &s
}

// GetStyle returns the style corresponding to the given file.
// It implements precedence logic similar to 'ls'.
func (s *Styles) GetStyle(dirEntry fs.DirEntry) lipgloss.Style {
	fInfo, err := dirEntry.Info()
	if err != nil {
		slog.Warn("Could not get file info", "file", dirEntry.Name(), "error", err)
		return lipgloss.NewStyle()
	}

	name := dirEntry.Name()
	mode := fInfo.Mode()
	isDir := dirEntry.IsDir()

	// 1. Directories
	if isDir {
		if mode&os.ModeSticky != 0 && mode&0o002 != 0 {
			if st, ok := s.keys["tw"]; ok {
				return st
			}
		}
		if mode&0o002 != 0 && mode&os.ModeSticky == 0 {
			if st, ok := s.keys["ow"]; ok {
				return st
			}
		}
		if mode&os.ModeSticky != 0 && mode&0o002 == 0 {
			if st, ok := s.keys["st"]; ok {
				return st
			}
		}
		if st, ok := s.keys["di"]; ok {
			return st
		}

		// Fallback to no style if 'di' not defined (shouldn't happen usually)
		return lipgloss.NewStyle()
	}

	// 2. Symlinks
	if mode&os.ModeSymlink != 0 {
		// TODO: Check 'or' (orphan) and 'mi' (missing) if we had target info easily available
		// TODO: Support ln=target
		if st, ok := s.keys["ln"]; ok {
			return st
		}
	}

	// 3. Special Files
	if mode&os.ModeNamedPipe != 0 {
		if st, ok := s.keys["pi"]; ok {
			return st
		}
	}
	if mode&os.ModeSocket != 0 {
		if st, ok := s.keys["so"]; ok {
			return st
		}
	}
	if mode&os.ModeDevice != 0 && mode&os.ModeCharDevice == 0 { // Block Device
		if st, ok := s.keys["bd"]; ok {
			return st
		}
	}
	if mode&os.ModeCharDevice != 0 {
		if st, ok := s.keys["cd"]; ok {
			return st
		}
	}
	if mode&os.ModeSetuid != 0 {
		if st, ok := s.keys["su"]; ok {
			return st
		}
	}
	if mode&os.ModeSetgid != 0 {
		if st, ok := s.keys["sg"]; ok {
			return st
		}
	}

	// 4. Executable (Regular file with +x)
	// Note: 'ls' checks this before extensions for coloring purposes usually,
	// unless an extension is defined?
	// Actually, 'ls' precedence is tricky. Often extension overrides 'ex'.
	// But 'ex' overrides 'fi'.
	// Let's check 'man dir_colors': "EXEC is for files with execute permission... usually green. This is matched if the file is executable and does not match any of the extensions."
	// So Extensions > Executable > File.

	// 5. Extensions / Patterns
	for _, rule := range s.extensions {
		if patternMatches(rule.pattern, name) {
			return rule.style
		}
	}

	// 4. Executable (Fallback if no extension matched)
	if mode.IsRegular() && mode&0o111 != 0 {
		if st, ok := s.keys["ex"]; ok {
			return st
		}
	}

	// 6. Regular File
	if mode.IsRegular() {
		if st, ok := s.keys["fi"]; ok {
			return st
		}
	}

	// 7. Normal / Global Fallback
	if st, ok := s.keys["no"]; ok {
		return st
	}

	return lipgloss.NewStyle()
}

func patternMatches(pattern, name string) bool {
	// Match extension
	if strings.HasPrefix(pattern, "*.") {
		ext := filepath.Ext(name)
		if len(ext) > 0 && ext[0] == '.' { // Remove leading dot from extension
			ext = ext[1:]
		}
		return ext == pattern[2:] // Compare with "*.ext" pattern
	}

	// Match filename suffix
	if strings.HasPrefix(pattern, "*") {
		suffix := pattern[1:] // remove '*'
		return strings.HasSuffix(name, suffix)
	}
	// Direct filename match
	return name == pattern
}

// https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_parameters
func parseStyle(s string) (string, lipgloss.Style, error) {
	st := lipgloss.NewStyle()
	split := strings.Split(s, "=")

	if len(split) != 2 {
		return "", st, errors.New("invalid entry" + s)
	}

	nums := parseNums(split[1])
	for i := 0; i < len(nums); i++ {
		n := nums[i]
		switch {
		case n == 0:
			st = lipgloss.NewStyle()
		case n == 1:
			st = st.Bold(true)
		case n == 2:
			st = st.Faint(true)
		case n == 3:
			st = st.Italic(true)
		case n == 4:
			st = st.Underline(true)
		case n == 5 || n == 6:
			st = st.Blink(true)
		case n == 7:
			st = st.Reverse(true)
		case n == 9:
			st = st.Strikethrough(true)
		// Not sure if 22-29 is of any use.
		// Applying underline for example and then turning it
		// off would have no effect overall
		case n == 22:
			st = st.Bold(false).Faint(false)
		case n == 23:
			st = st.Italic(false)
		case n == 24:
			st = st.Underline(false)
		case n == 25:
			st = st.Blink(false)
		case n == 27:
			st = st.Reverse(false)
		case n == 29:
			st = st.Strikethrough(false)
		case n >= 30 && n <= 37:
			// 0-7 default colours
			st = st.Foreground(lipgloss.Color(strconv.Itoa(n - 30)))
		case n == 38:
			if nums[i+1] == 2 && i+5 <= len(nums) { // RGB
				st = st.Foreground(lipgloss.Color(rgbToHex(nums[i+2], nums[i+3], nums[i+4])))
				i += 4
			} else if nums[i+1] == 5 && i+3 <= len(nums) {
				st = st.Foreground(lipgloss.Color(strconv.Itoa(nums[i+2])))
				i += 2
			} else {
				return "", st, errors.New("badly formatted 38 entry " + s)
			}
		case n >= 40 && n <= 47:
			// 0-7 default colours
			st = st.Background(lipgloss.Color(strconv.Itoa(n - 40)))
		case n == 48:
			if nums[i+1] == 2 && i+5 <= len(nums) { // RGB
				st = st.Background(lipgloss.Color(rgbToHex(nums[i+2], nums[i+3], nums[i+4])))
				i += 4
			} else if nums[i+1] == 5 && i+3 <= len(nums) {
				st = st.Background(lipgloss.Color(strconv.Itoa(nums[i+2])))
				i += 2
			} else {
				return "", st, errors.New("badly formatted 48 entry " + s)
			}
		}
	}
	return split[0], st, nil
}

func parseNums(s string) []int {
	var nums []int
	for s1 := range strings.SplitSeq(s, ";") {
		if s1 == "" {
			// todo maybe we should just throw an error here instead?
			nums = append(nums, 0)
			continue
		}
		n, err := strconv.Atoi(s1)
		if err != nil {
			continue
		}
		nums = append(nums, n)
	}
	return nums
}

func rgbToHex(r, g, b int) string {
	return fmt.Sprintf("#%02x%02x%02x", r, g, b)
}
