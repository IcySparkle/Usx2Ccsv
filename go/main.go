package main

import (
	"encoding/csv"
	"encoding/json"
	"encoding/xml"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

type stringSlice []string

func (s *stringSlice) String() string {
	return strings.Join(*s, ",")
}

func (s *stringSlice) Set(value string) error {
	*s = append(*s, value)
	return nil
}

type nodeType int

const (
	nodeElement nodeType = iota
	nodeText
)

type node struct {
	Type     nodeType
	Name     string
	Attrs    map[string]string
	Children []*node
	Text     string
}

type row struct {
	Book       string
	Chapter    string
	Verse      string
	TextPlain  string
	TextStyled string
	Footnotes  string
	Crossrefs  string
	Subtitle   string
}

type fileResult struct {
	Input  string `json:"input"`
	Output string `json:"output"`
	Format string `json:"format"`
	Rows   int    `json:"rows"`
}

type summary struct {
	Files []fileResult `json:"files"`
}

type usxState struct {
	bookCode        string
	currentChapter  string
	currentVerse    string
	currentPlain    string
	currentStyled   string
	currentFootnote []string
	currentCrossref []string
	currentSubtitle string
	rows            []row
}

func main() {
	var inputs stringSlice
	output := flag.String("output", "", "Output folder (optional)")
	help := flag.Bool("help", false, "Show help")
	quiet := flag.Bool("quiet", false, "Suppress progress output")
	jsonOut := flag.Bool("json", false, "Output JSON summary to stdout")
	flag.Var(&inputs, "input", "Input file/folder/wildcard path (repeatable)")
	flag.Parse()

	if *help || len(inputs) == 0 {
		showUsage()
		return
	}

	inputItems, err := resolveInputItems(inputs)
	if err != nil {
		fail(err.Error(), *jsonOut)
	}

	files, err := collectFiles(inputItems)
	if err != nil {
		fail(err.Error(), *jsonOut)
	}

	if len(files) == 0 {
		fail("No .usx, .usfm, or .sfm files found.", *jsonOut)
	}

	if *output != "" {
		if err := os.MkdirAll(*output, 0o755); err != nil {
			fail(err.Error(), *jsonOut)
		}
	}

	runSummary := summary{}
	for _, path := range files {
		ext := strings.ToLower(filepath.Ext(path))
		csvPath := outputPath(path, *output)

		switch ext {
		case ".usx":
			rows, err := convertUsxToCsv(path, csvPath, *quiet)
			if err != nil {
				fail(err.Error(), *jsonOut)
			}
			runSummary.Files = append(runSummary.Files, fileResult{
				Input:  path,
				Output: csvPath,
				Format: "usx",
				Rows:   rows,
			})
		case ".usfm", ".sfm":
			rows, err := convertUsfmToCsv(path, csvPath, *quiet)
			if err != nil {
				fail(err.Error(), *jsonOut)
			}
			format := strings.TrimPrefix(ext, ".")
			runSummary.Files = append(runSummary.Files, fileResult{
				Input:  path,
				Output: csvPath,
				Format: format,
				Rows:   rows,
			})
		default:
			continue
		}
	}

	if *jsonOut {
		writeJSONSummary(runSummary)
		return
	}

	fmt.Println("All conversions completed.")
}

func showUsage() {
	fmt.Println("usxtocsv (Go) - Convert USX/USFM/SFM to CSV")
	fmt.Println("")
	fmt.Println("Usage:")
	fmt.Println("  usxtocsv -input <file|folder|wildcard> [-output <folder>]")
	fmt.Println("  usxtocsv -input <path1> -input <path2>")
	fmt.Println("  usxtocsv -quiet -json")
	fmt.Println("  usxtocsv -help")
}

func resolveInputItems(inputs []string) ([]string, error) {
	var items []string

	for _, raw := range inputs {
		parts := strings.Split(raw, ",")
		for _, part := range parts {
			p := strings.TrimSpace(part)
			if p == "" {
				continue
			}

			if hasWildcard(p) {
				matches, err := filepath.Glob(p)
				if err != nil {
					return nil, err
				}
				if len(matches) == 0 {
					return nil, fmt.Errorf("Input path not found: %s", p)
				}
				items = append(items, matches...)
				continue
			}

			if _, err := os.Stat(p); err != nil {
				return nil, fmt.Errorf("Input path not found: %s", p)
			}
			items = append(items, p)
		}
	}

	return items, nil
}

func collectFiles(items []string) ([]string, error) {
	var files []string

	for _, item := range items {
		info, err := os.Stat(item)
		if err != nil {
			return nil, fmt.Errorf("Input path not found: %s", item)
		}

		if info.IsDir() {
			entries, err := os.ReadDir(item)
			if err != nil {
				return nil, err
			}
			for _, entry := range entries {
				if entry.IsDir() {
					continue
				}
				ext := strings.ToLower(filepath.Ext(entry.Name()))
				if isSupportedExt(ext) {
					files = append(files, filepath.Join(item, entry.Name()))
				}
			}
			continue
		}

		ext := strings.ToLower(filepath.Ext(item))
		if !isSupportedExt(ext) {
			return nil, errors.New("Input must be a .usx, .usfm, or .sfm file, or a folder containing them.")
		}
		files = append(files, item)
	}

	return files, nil
}

func outputPath(inputPath, outputFolder string) string {
	if outputFolder != "" {
		base := strings.TrimSuffix(filepath.Base(inputPath), filepath.Ext(inputPath))
		return filepath.Join(outputFolder, base+".csv")
	}
	return strings.TrimSuffix(inputPath, filepath.Ext(inputPath)) + ".csv"
}

func isSupportedExt(ext string) bool {
	switch ext {
	case ".usx", ".usfm", ".sfm":
		return true
	default:
		return false
	}
}

func hasWildcard(path string) bool {
	return strings.ContainsAny(path, "*?[]")
}

func convertUsxToCsv(usxPath, csvPath string, quiet bool) (int, error) {
	if !quiet {
		fmt.Fprintf(os.Stderr, "Processing (USX) %s\n", usxPath)
	}
	root, err := parseXML(usxPath)
	if err != nil {
		return 0, err
	}

	if root == nil || root.Name != "usx" {
		return 0, fmt.Errorf("No <usx> root found in %s", usxPath)
	}

	bookNode := findFirstChild(root, "book")
	if bookNode == nil {
		return 0, fmt.Errorf("No <book> found in %s", usxPath)
	}

	state := &usxState{
		bookCode: getAttrValue(bookNode, "code"),
	}

	for _, child := range root.Children {
		processUsxNode(child, state)
	}

	sortRows(state.rows)
	if err := writeCsv(csvPath, state.rows); err != nil {
		return 0, err
	}

	if !quiet {
		fmt.Fprintf(os.Stderr, "Created CSV: %s\n", csvPath)
	}
	return len(state.rows), nil
}

func convertUsfmToCsv(usfmPath, csvPath string, quiet bool) (int, error) {
	if !quiet {
		fmt.Fprintf(os.Stderr, "Processing (USFM/SFM) %s\n", usfmPath)
	}
	data, err := os.ReadFile(usfmPath)
	if err != nil {
		return 0, err
	}

	content := strings.ReplaceAll(string(data), "\r\n", "\n")
	lines := strings.Split(content, "\n")

	bookCode := strings.TrimSuffix(filepath.Base(usfmPath), filepath.Ext(usfmPath))
	reID := regexp.MustCompile(`(?i)^\\id\s+(\S+)`)
	for _, line := range lines {
		l := strings.TrimSpace(line)
		if l == "" {
			continue
		}
		if m := reID.FindStringSubmatch(l); len(m) > 1 {
			bookCode = m[1]
			break
		}
	}

	var rows []row
	currentChapter := ""
	currentVerse := ""
	currentPlain := ""
	currentStyled := ""
	currentFootnotes := []string{}
	currentCrossrefs := []string{}
	currentSubtitle := ""

	reChapter := regexp.MustCompile(`(?i)^\\c\s+(\d+)\b`)
	reHeading := regexp.MustCompile(`(?i)^\\(s[0-3]?|sp|ms|mr|mt[12]?)\s*(.*)$`)
	reVerse := regexp.MustCompile(`(?i)^\\v\s+(\d+)\s*(.*)$`)
	rePara := regexp.MustCompile(`(?i)^\\(m|p|pi|q[0-4]?|qt[0-4]?)\s*(.*)$`)

	for _, line := range lines {
		l := strings.TrimSpace(line)
		if l == "" {
			continue
		}

		if m := reChapter.FindStringSubmatch(l); len(m) > 1 {
			if currentVerse != "" {
				addCurrentVerseUsfm(&rows, bookCode, currentChapter, currentVerse, currentPlain, currentStyled, currentFootnotes, currentCrossrefs, currentSubtitle)
			}
			currentVerse = ""
			currentPlain = ""
			currentStyled = ""
			currentFootnotes = []string{}
			currentCrossrefs = []string{}

			currentChapter = m[1]
			continue
		}

		if m := reHeading.FindStringSubmatch(l); len(m) > 1 {
			headText := m[2]
			headText = extractNotesFromUsfmSegment(headText, &currentFootnotes, &currentCrossrefs)
			headText = regexp.MustCompile(`(?i)\\\+?[a-z0-9]+\*?`).ReplaceAllString(headText, " ")
			headText = normalizeWhitespace(headText)
			if headText != "" {
				currentSubtitle = headText
			}
			continue
		}

		if m := reVerse.FindStringSubmatch(l); len(m) > 1 {
			if currentVerse != "" {
				addCurrentVerseUsfm(&rows, bookCode, currentChapter, currentVerse, currentPlain, currentStyled, currentFootnotes, currentCrossrefs, currentSubtitle)
			}

			currentVerse = m[1]
			currentPlain = ""
			currentStyled = ""
			currentFootnotes = []string{}
			currentCrossrefs = []string{}

			rest := m[2]
			if rest != "" {
				processUsfmContentSegment(rest, &currentPlain, &currentStyled, &currentFootnotes, &currentCrossrefs)
			}
			continue
		}

		if m := rePara.FindStringSubmatch(l); len(m) > 1 {
			rest := m[2]
			if currentVerse != "" && rest != "" {
				processUsfmContentSegment(rest, &currentPlain, &currentStyled, &currentFootnotes, &currentCrossrefs)
			}
			continue
		}

		if currentVerse != "" {
			processUsfmContentSegment(l, &currentPlain, &currentStyled, &currentFootnotes, &currentCrossrefs)
		}
	}

	if currentVerse != "" {
		addCurrentVerseUsfm(&rows, bookCode, currentChapter, currentVerse, currentPlain, currentStyled, currentFootnotes, currentCrossrefs, currentSubtitle)
	}

	sortRows(rows)
	if err := writeCsv(csvPath, rows); err != nil {
		return 0, err
	}

	if !quiet {
		fmt.Fprintf(os.Stderr, "Created CSV: %s\n", csvPath)
	}
	return len(rows), nil
}

func writeJSONSummary(runSummary summary) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	_ = enc.Encode(runSummary)
}

func fail(message string, jsonOut bool) {
	if jsonOut {
		out := map[string]string{"error": message}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(out)
		os.Exit(1)
	}
	fmt.Fprintln(os.Stderr, message)
	os.Exit(1)
}

func addCurrentVerseUsfm(rows *[]row, book, chapter, verse, plain, styled string, footnotes, crossrefs []string, subtitle string) {
	plain = strings.TrimSpace(plain)
	styled = strings.TrimSpace(styled)
	subtitle = strings.TrimSpace(subtitle)

	if book != "" && chapter != "" && verse != "" && plain != "" {
		*rows = append(*rows, row{
			Book:       book,
			Chapter:    chapter,
			Verse:      verse,
			TextPlain:  plain,
			TextStyled: styled,
			Footnotes:  strings.Join(footnotes, " | "),
			Crossrefs:  strings.Join(crossrefs, " | "),
			Subtitle:   subtitle,
		})
	}
}

func processUsfmContentSegment(segment string, currentPlain, currentStyled *string, footnotes, crossrefs *[]string) {
	if strings.TrimSpace(segment) == "" {
		return
	}

	reSup := regexp.MustCompile(`(?is)\\\+?sup\b.*?\\\+?sup\*`)
	seg := reSup.ReplaceAllString(segment, " ")

	seg = extractNotesFromUsfmSegment(seg, footnotes, crossrefs)
	if strings.TrimSpace(seg) == "" {
		return
	}

	styleMap := map[string]string{
		"wj":   "wj",
		"add":  "add",
		"nd":   "nd",
		"it":   "i",
		"bd":   "b",
		"bdit": "bdit",
	}

	styled := seg
	for key, tag := range styleMap {
		reOpen := regexp.MustCompile(fmt.Sprintf(`(?i)\\\+?%s\b\s*`, key))
		reClose := regexp.MustCompile(fmt.Sprintf(`(?i)\\\+?%s\*\s*`, key))
		styled = reOpen.ReplaceAllString(styled, "<"+tag+">")
		styled = reClose.ReplaceAllString(styled, "</"+tag+">")
	}

	reUnknown := regexp.MustCompile(`(?i)\\\+?[a-z0-9]+\*?`)
	styled = normalizeWhitespace(reUnknown.ReplaceAllString(styled, " "))
	plain := normalizeWhitespace(reUnknown.ReplaceAllString(seg, " "))

	if plain != "" {
		if *currentPlain != "" {
			*currentPlain += " "
			*currentStyled += " "
		}
		*currentPlain += plain
		*currentStyled += styled
	}
}

func extractNotesFromUsfmSegment(segment string, footnotes, crossrefs *[]string) string {
	if strings.TrimSpace(segment) == "" {
		return segment
	}

	reFoot := regexp.MustCompile(`(?is)\\f\b(.*?\\f\*)`)
	text := reFoot.ReplaceAllStringFunc(segment, func(m string) string {
		sub := reFoot.FindStringSubmatch(m)
		if len(sub) > 1 {
			full := `\f` + sub[1]
			ftText := extractFtFromUsfmNoteText(full)
			if ftText != "" {
				*footnotes = append(*footnotes, ftText)
			}
		}
		return " "
	})

	reCross := regexp.MustCompile(`(?is)\\x\b(.*?\\x\*)`)
	text = reCross.ReplaceAllStringFunc(text, func(m string) string {
		sub := reCross.FindStringSubmatch(m)
		if len(sub) > 1 {
			full := `\x` + sub[1]
			ftText := extractFtFromUsfmNoteText(full)
			if ftText != "" {
				*crossrefs = append(*crossrefs, ftText)
			}
		}
		return " "
	})

	return text
}

func extractFtFromUsfmNoteText(noteText string) string {
	if strings.TrimSpace(noteText) == "" {
		return ""
	}

	re := regexp.MustCompile(`(?i)\\ft\b([^\\]*)`)
	m := re.FindStringSubmatch(noteText)
	if len(m) < 2 {
		return ""
	}
	return normalizeWhitespace(m[1])
}

func parseXML(path string) (*node, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	decoder := xml.NewDecoder(file)
	var stack []*node
	var root *node

	for {
		tok, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		switch t := tok.(type) {
		case xml.StartElement:
			n := &node{
				Type:  nodeElement,
				Name:  t.Name.Local,
				Attrs: map[string]string{},
			}
			for _, attr := range t.Attr {
				n.Attrs[attr.Name.Local] = attr.Value
			}
			if len(stack) > 0 {
				parent := stack[len(stack)-1]
				parent.Children = append(parent.Children, n)
			} else {
				root = n
			}
			stack = append(stack, n)
		case xml.EndElement:
			if len(stack) > 0 {
				stack = stack[:len(stack)-1]
			}
		case xml.CharData:
			if len(stack) == 0 {
				continue
			}
			text := string(t)
			if text == "" {
				continue
			}
			n := &node{
				Type: nodeText,
				Text: text,
			}
			parent := stack[len(stack)-1]
			parent.Children = append(parent.Children, n)
		}
	}

	return root, nil
}

func processUsxNode(n *node, state *usxState) {
	if n == nil {
		return
	}

	switch n.Type {
	case nodeElement:
		switch n.Name {
		case "chapter":
			state.currentChapter = getAttrValue(n, "number")
			return
		case "verse":
			sid := getAttrValue(n, "sid")
			eid := getAttrValue(n, "eid")
			if sid != "" {
				state.currentVerse = getAttrValue(n, "number")
				state.currentPlain = ""
				state.currentStyled = ""
				state.currentFootnote = []string{}
				state.currentCrossref = []string{}
				return
			}
			if eid != "" {
				state.addCurrentVerse()
				state.currentVerse = ""
				state.currentPlain = ""
				state.currentStyled = ""
				state.currentFootnote = []string{}
				state.currentCrossref = []string{}
				return
			}
		case "note":
			processUsxNote(n, state)
			return
		case "para":
			style := getAttrValue(n, "style")
			if isSubtitleStyle(style) {
				subText := normalizeWhitespace(innerText(n))
				if subText != "" {
					state.currentSubtitle = subText
				}
			}
		case "char":
			style := getAttrValue(n, "style")
			if style == "sup" {
				return
			}
			tag := ""
			if style != "" {
				tag = getStyledTagName(style)
			}
			if state.currentVerse != "" && tag != "" {
				state.currentStyled += "<" + tag + ">"
			}
			for _, child := range n.Children {
				processUsxNode(child, state)
			}
			if state.currentVerse != "" && tag != "" {
				state.currentStyled += "</" + tag + ">"
			}
			return
		}

		for _, child := range n.Children {
			processUsxNode(child, state)
		}
	case nodeText:
		if state.currentVerse == "" {
			return
		}
		text := normalizeWhitespace(n.Text)
		if text == "" {
			return
		}
		if state.currentPlain != "" {
			state.currentPlain += " "
			state.currentStyled += " "
		}
		state.currentPlain += text
		state.currentStyled += text
	}
}

func (s *usxState) addCurrentVerse() {
	plain := strings.TrimSpace(s.currentPlain)
	styled := strings.TrimSpace(s.currentStyled)
	subText := strings.TrimSpace(s.currentSubtitle)

	if s.bookCode != "" && s.currentChapter != "" && s.currentVerse != "" && plain != "" {
		s.rows = append(s.rows, row{
			Book:       s.bookCode,
			Chapter:    s.currentChapter,
			Verse:      s.currentVerse,
			TextPlain:  plain,
			TextStyled: styled,
			Footnotes:  strings.Join(s.currentFootnote, " | "),
			Crossrefs:  strings.Join(s.currentCrossref, " | "),
			Subtitle:   subText,
		})
	}
}

func processUsxNote(noteNode *node, state *usxState) {
	style := getAttrValue(noteNode, "style")
	ft := extractFtFromNote(noteNode)
	if ft == "" {
		return
	}

	if strings.HasPrefix(style, "x") {
		state.currentCrossref = append(state.currentCrossref, ft)
	} else {
		state.currentFootnote = append(state.currentFootnote, ft)
	}
}

func extractFtFromNote(noteNode *node) string {
	ftNode := findFtNode(noteNode)
	if ftNode == nil {
		return ""
	}
	raw := innerText(ftNode)
	if strings.TrimSpace(raw) == "" {
		return ""
	}
	return normalizeWhitespace(raw)
}

func findFtNode(n *node) *node {
	if n == nil {
		return nil
	}
	if n.Type == nodeElement && n.Name == "char" {
		if getAttrValue(n, "style") == "ft" {
			return n
		}
	}
	for _, child := range n.Children {
		if found := findFtNode(child); found != nil {
			return found
		}
	}
	return nil
}

func innerText(n *node) string {
	if n == nil {
		return ""
	}
	if n.Type == nodeText {
		return n.Text
	}
	var b strings.Builder
	for _, child := range n.Children {
		b.WriteString(innerText(child))
	}
	return b.String()
}

func findFirstChild(n *node, name string) *node {
	for _, child := range n.Children {
		if child.Type == nodeElement && child.Name == name {
			return child
		}
	}
	return nil
}

func getAttrValue(n *node, name string) string {
	if n == nil || n.Attrs == nil {
		return ""
	}
	return n.Attrs[name]
}

func normalizeWhitespace(text string) string {
	if strings.TrimSpace(text) == "" {
		return ""
	}
	return strings.Join(strings.Fields(text), " ")
}

func getStyledTagName(style string) string {
	switch style {
	case "wj":
		return "wj"
	case "add":
		return "add"
	case "nd":
		return "nd"
	case "bdit":
		return "bdit"
	case "it":
		return "i"
	case "bd":
		return "b"
	default:
		return "span"
	}
}

func isSubtitleStyle(style string) bool {
	switch style {
	case "s", "s1", "s2", "s3", "sp", "ms", "mr", "mt", "mt1", "mt2":
		return true
	default:
		return false
	}
}

func sortRows(rows []row) {
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].Book != rows[j].Book {
			return rows[i].Book < rows[j].Book
		}
		ci := parseInt(rows[i].Chapter)
		cj := parseInt(rows[j].Chapter)
		if ci != cj {
			return ci < cj
		}
		return rows[i].Verse < rows[j].Verse
	})
}

func parseInt(v string) int {
	n, err := strconv.Atoi(v)
	if err != nil {
		return 0
	}
	return n
}

func writeCsv(path string, rows []row) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	if err := writer.Write([]string{"Book", "Chapter", "Verse", "TextPlain", "TextStyled", "Footnotes", "Crossrefs", "Subtitle"}); err != nil {
		return err
	}
	for _, r := range rows {
		if err := writer.Write([]string{r.Book, r.Chapter, r.Verse, r.TextPlain, r.TextStyled, r.Footnotes, r.Crossrefs, r.Subtitle}); err != nil {
			return err
		}
	}
	writer.Flush()
	return writer.Error()
}
