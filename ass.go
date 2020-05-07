package ass

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"text/template"
)

// Event is a single subtitle item
// Layer, Start, End, Style, Name, MarginL, MarginR, MarginV, Effect, Text
type Event struct {
	Layer   int    `json:"layer"`
	Start   string `json:"start"` // 0:00:00:00 h:mm:ss:msms
	End     string `json:"end"`   // 0:00:00:00 h:mm:ss:msms
	Style   string `json:"style"`
	Name    string `json:"name"` // The speaker name, just a placeholder
	MarginL uint   `json:"marginLeft"`
	MarginR uint   `json:"marginRight"`
	MarginV uint   `json:"marginV"`
	Effect  string `json:"effect"`
	Text    string `json:"text"`
}

var timeReg = regexp.MustCompile(`\d:[0-6]\d:[0-6]\d:\d\d`)

func (evt Event) validate() error {
	if !timeReg.MatchString(evt.Start) {
		return fmt.Errorf("Invalid start time: %s", evt.Start)
	}
	if !timeReg.MatchString(evt.End) {
		return fmt.Errorf("Invalid end time: %s", evt.End)
	}
	return nil
}

// Style is a style for ass subtitle
type Style struct {
	Name         string `json:"name"`
	FontName     string `json:"font"`
	FontSize     int    `json:"fontSize"`
	PrimaryColor string `json:"primaryColor"`
	SecondColor  string `json:"secondColor"`
	OutlineColor string `json:"outlineColor"`
	BackColor    string `json:"backColor"`
	Bold         int    `json:"bold"`
	Italic       int    `json:"italic"`
	Underline    int    `json:"underline"`
	StrikeOut    int    `json:"strikeOut"`
	ScaleX       int    `json:"scaleX"`
	ScaleY       int    `json:"scaleY"`
}

// Check color is ABGR or not
func isValidABGR(color string) bool {
	if len(color) != 8 {
		return false
	}
	for _, c := range color {
		if ('0' <= c && c <= '9') || ('a' <= c && c <= 'f') || ('A' <= c && c <= 'F') {
			continue
		}
		return false
	}
	return true
}

func (style Style) validate() error {
	if style.PrimaryColor != "" && !isValidABGR(style.PrimaryColor) {
		return fmt.Errorf("Invalid primary color: %s", style.PrimaryColor)
	}
	if style.SecondColor != "" && !isValidABGR(style.SecondColor) {
		return fmt.Errorf("Invalid secondary color: %s", style.SecondColor)
	}
	if style.OutlineColor != "" && !isValidABGR(style.OutlineColor) {
		return fmt.Errorf("Invalid outline color: %s", style.OutlineColor)
	}
	if style.BackColor != "" && !isValidABGR(style.BackColor) {
		return fmt.Errorf("Invalid back color: %s", style.BackColor)
	}
	if style.Bold != 0 && style.Bold != -1 {
		return fmt.Errorf("Invalid style bold: %d", style.Bold)
	}
	if style.Italic != 0 && style.Italic != -1 {
		return fmt.Errorf("Invalid style italic: %d", style.Italic)
	}
	if style.Underline != 0 && style.Underline != -1 {
		return fmt.Errorf("Invalid style underline: %d", style.Underline)
	}
	if style.StrikeOut != 0 && style.StrikeOut != -1 {
		return fmt.Errorf("Invalid style StrikeOut: %d", style.StrikeOut)
	}
	return nil
}

// Subtitle the ass subtitle
type Subtitle struct {
	Title        string   `json:"title"`
	OriginScript string   `json:"originScript"`
	PlayerWidth  uint     `json:"playResX"`
	PlayerHeight uint     `json:"playResY"`
	PlayDepth    uint     `json:"playDepth"`
	Timer        float32  `json:"timer"`
	Styles       []*Style `json:"styles"`
	Events       []*Event `json:"events"`
}

// some default values
const (
	defPlayerWidth  = 1920
	defPlayerHeight = 1080
	defFontName     = "Arial"
)

// validate subtitle
func (as *Subtitle) validate() error {

	if as.Timer < 0 {
		return fmt.Errorf("Invalid timer: %f", as.Timer)
	}

	for _, style := range as.Styles {
		if style == nil {
			return fmt.Errorf("Style cannot be nil")
		}
		if err := style.validate(); err != nil {
			return err
		}
	}

	for _, evt := range as.Events {
		if evt == nil {
			return fmt.Errorf("Event cannot be nil")
		}
		if err := evt.validate(); err != nil {
			return err
		}
	}

	return nil
}

// fulfill subtitle with some default values
func (as *Subtitle) fulfill() {
	if as.OriginScript == "" {
		as.OriginScript = "unknown"
	}
	if as.PlayerWidth == 0 && as.PlayerHeight == 0 {
		as.PlayerWidth = defPlayerWidth
		as.PlayerHeight = defPlayerHeight
	} else if as.PlayerWidth == 0 {
		as.PlayerWidth = defPlayerWidth * (as.PlayerHeight / defPlayerHeight)
	} else if as.PlayerHeight == 0 {
		as.PlayerHeight = defPlayerHeight * (as.PlayerWidth / defPlayerWidth)
	}
	for _, style := range as.Styles {
		if style.FontName == "" {
			style.FontName = defFontName
		}
	}
}

const assV4Tpl = `
[Script Info]
Title: {{.Title}}
Original Script: {{.OriginScript}}
ScriptType: v4.00+
Collisions: Normal
PlayResX: {{.PlayerWidth}}
PlayResY: {{.PlayerHeight}}
Timer: {{printf "%.4f" .Timer}}

[V4+ Styles]
Format: Name, Fontname, Fontsize, PrimaryColour, SecondaryColour, OutlineColour, BackColour, Bold, Italic, Underline, StrikeOut, ScaleX, ScaleY, Spacing, Angle, BorderStyle, Outline, Shadow, Alignment, MarginL, MarginR, MarginV, Encoding
{{range .Styles -}}
Style: {{.Name}},{{.FontName}},{{.FontSize}},&H{{.PrimaryColor}},&H{{.SecondColor}},&H{{.OutlineColor}},&H{{.BackColor}},1,0,0,0,100,100,0,0,1,2,0,2,20,20,2,0
{{end}}

[Events]
Format: Layer, Start, End, Style, Name, MarginL, MarginR, MarginV, Effect, Text
{{range .Events -}}
Dialogue: {{.Layer}},{{.Start}},{{.End}},{{.Style}},{{.Name}},{{printf "%04d" .MarginL}},{{printf "%04d" .MarginR}},{{printf "%04d" .MarginV}},{{.Effect}},{{.Text}}
{{end}}
`

// WriteTo write ass subtitle to destination
func (as Subtitle) WriteTo(w io.Writer) (int64, error) {
	err := as.validate()
	if err != nil {
		return 0, err
	}

	// fulfill subtitle, add some default values
	as.fulfill()

	tpl := template.New("ass")
	tpl.Parse(assV4Tpl)

	writer := bufio.NewWriter(w)
	err = tpl.Execute(writer, as)
	if err != nil {
		return 0, err
	}
	n := writer.Buffered()
	err = writer.Flush()
	return int64(n), err
}
