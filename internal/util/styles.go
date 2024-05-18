package util

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/rs/zerolog/log"
)

const defaultColours = "di=01;34:ln=01;36:mh=00:pi=40;33:so=01;35:do=01;35:bd=40;33;01:cd=40;33;01:or=40;31;01:mi=00:su=37;41:sg=30;43:ca=00:tw=30;42:ow=34;42:st=37;44:ex=01;32:*.tar=01;31:*.tgz=01;31:*.arc=01;31:*.arj=01;31:*.taz=01;31:*.lha=01;31:*.lz4=01;31:*.lzh=01;31:*.lzma=01;31:*.tlz=01;31:*.txz=01;31:*.tzo=01;31:*.t7z=01;31:*.zip=01;31:*.z=01;31:*.dz=01;31:*.gz=01;31:*.lrz=01;31:*.lz=01;31:*.lzo=01;31:*.xz=01;31:*.zst=01;31:*.tzst=01;31:*.bz2=01;31:*.bz=01;31:*.tbz=01;31:*.tbz2=01;31:*.tz=01;31:*.deb=01;31:*.rpm=01;31:*.jar=01;31:*.war=01;31:*.ear=01;31:*.sar=01;31:*.rar=01;31:*.alz=01;31:*.ace=01;31:*.zoo=01;31:*.cpio=01;31:*.7z=01;31:*.rz=01;31:*.cab=01;31:*.wim=01;31:*.swm=01;31:*.dwm=01;31:*.esd=01;31:*.avif=01;35:*.jpg=01;35:*.jpeg=01;35:*.mjpg=01;35:*.mjpeg=01;35:*.gif=01;35:*.bmp=01;35:*.pbm=01;35:*.pgm=01;35:*.ppm=01;35:*.tga=01;35:*.xbm=01;35:*.xpm=01;35:*.tif=01;35:*.tiff=01;35:*.png=01;35:*.svg=01;35:*.svgz=01;35:*.mng=01;35:*.pcx=01;35:*.mov=01;35:*.mpg=01;35:*.mpeg=01;35:*.m2v=01;35:*.mkv=01;35:*.webm=01;35:*.webp=01;35:*.ogm=01;35:*.mp4=01;35:*.m4v=01;35:*.mp4v=01;35:*.vob=01;35:*.qt=01;35:*.nuv=01;35:*.wmv=01;35:*.asf=01;35:*.rm=01;35:*.rmvb=01;35:*.flc=01;35:*.avi=01;35:*.fli=01;35:*.flv=01;35:*.gl=01;35:*.dl=01;35:*.xcf=01;35:*.xwd=01;35:*.yuv=01;35:*.cgm=01;35:*.emf=01;35:*.ogv=01;35:*.ogx=01;35:*.aac=00;36:*.au=00;36:*.flac=00;36:*.m4a=00;36:*.mid=00;36:*.midi=00;36:*.mka=00;36:*.mp3=00;36:*.mpc=00;36:*.ogg=00;36:*.ra=00;36:*.wav=00;36:*.oga=00;36:*.opus=00;36:*.spx=00;36:*.xspf=00;36:*~=00;90:*#=00;90:*.bak=00;90:*.old=00;90:*.orig=00;90:*.part=00;90:*.rej=00;90:*.swp=00;90:*.tmp=00;90:*.dpkg-dist=00;90:*.dpkg-old=00;90:*.ucf-dist=00;90:*.ucf-new=00;90:*.ucf-old=00;90:*.rpmnew=00;90:*.rpmorig=00;90:*.rpmsave=00;90:"

// List of predefined keys for LS_COLORS
var keys = []string{
	"no", // NORMAL, NORM	Global default, although everything should be something
	"fi", // FILE	Normal file
	"di", // DIR	Directory
	"ln", // SYMLINK, LINK, LNK	Symbolic link. If you set this to 'target' instead of a numerical value, the colour is as for the file pointed to.
	"pi", // FIFO, PIPE	Named pipe
	"do", // DOOR	Door
	"bd", // BLOCK, BLK	Block device
	"cd", // CHAR, CHR	Character device
	"or", // ORPHAN	Symbolic link pointing to a non-existent file
	"so", // SOCK	Socket
	"su", // SETUID	File that is setuid (u+s)
	"sg", // SETGID	File that is setgid (g+s)
	"tw", // STICKY_OTHER_WRITABLE	Directory that is sticky and other-writable (+t,o+w)
	"ow", // OTHER_WRITABLE	Directory that is other-writable (o+w) and not sticky
	"st", // STICKY	Directory with the sticky bit set (+t) and not other-writable
	"ex", // EXEC	Executable file (i.e. has 'x' set in permissions)
	"mi", // MISSING	Non-existent file pointed to by a symbolic link (visible when you type ls -l)
	"lc", // LEFTCODE, LEFT	Opening terminal code
	"rc", // RIGHTCODE, RIGHT	Closing terminal code
	"ec", // ENDCODE, END	Non-filename text
}

var styles map[string]lipgloss.Style

func SetupStyles() {
	var colours string
	if e := os.Getenv("LS_COLORS"); e != "" {
		colours = e
	} else {
		colours = defaultColours
	}

	styles = make(map[string]lipgloss.Style)
	for _, s := range strings.Split(colours, ":") {
		if strings.TrimSpace(s) == "" {
			continue
		}
		entry, st, err := parseStyle(s)
		if err != nil {
			log.Warn().Err(err).Msg("Error while parsing LS_COLORS")
			continue
		}
		styles[entry] = st
	}
}

// https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_parameters
func parseStyle(s string) (string, lipgloss.Style, error) {
	st := lipgloss.NewStyle()
	split := strings.Split(s, "=")

	// Currently there is no support for real regex matching such as
	// "*package-lock.json=0;38;2;122;112;112". This might be added in the future
	// with a map[regex]style variable.
	if len(split) != 2 {
		return "", st, errors.New("invalid entry" + s)
	} else if !slices.Contains(keys, split[0]) && split[0][0:2] != "*." && split[0][0] != '*' {
		return "", st, errors.New("unsupported entry " + s)
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
	for _, s1 := range strings.Split(s, ";") {
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

// GetStyle returns the style corresponding to
// the given file
func GetStyle(dirEntry fs.DirEntry) lipgloss.Style {
	fInfo, err := dirEntry.Info()
	if err != nil {
		return lipgloss.NewStyle()
	}

	m := fInfo.Mode()
	var k string
	switch {
	case dirEntry.IsDir():
		k = "di"
		switch {
		case m&os.ModeSticky != 0 && m&0o002 != 0:
			k = "tw"
		case m&os.ModeSticky != 0:
			k = "st"
		case m&0o002 != 0:
			k = "ow"
		}
	case m&os.ModeSymlink != 0:
		// todo identify broken symlinks ("or" key)
		k = "ln"
	case m&os.ModeSetuid != 0:
		k = "su"
	case m&os.ModeSetgid != 0:
		k = "sg"
	case m&os.ModeSticky != 0:
		k = "st"
	case m&os.ModeSocket != 0:
		k = "so"
	case m&os.ModeDevice != 0:
		k = "bd"
	case m.IsRegular() && m&0o111 != 0:
		k = "ex"
	case m&os.ModeNamedPipe != 0:
		k = "pi"
	case m&os.ModeCharDevice != 0:
		k = "cd"
	default:
		k = "*" + filepath.Ext(dirEntry.Name()) // Try to match against *.pattern
	}

	if s, ok := styles[k]; ok {
		return s
	}

	// try to match against *pattern
	k = "*" + dirEntry.Name()
	if s, ok := styles[k]; ok {
		return s
	}

	return lipgloss.NewStyle()
}

func rgbToHex(r, g, b int) string {
	return fmt.Sprintf("#%02x%02x%02x", r, g, b)
}
