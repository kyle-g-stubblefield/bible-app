package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	_ "github.com/joho/godotenv/autoload"
	"golang.org/x/net/html"
)

type ApiResponse struct {
	Query       string  `json:"query"`
	Canonical   string  `json:"canonical"`
	Parsed      [][]int `json:"parsed"`
	PassageMeta []struct {
		Canonical    string `json:"canonical"`
		ChapterStart []int  `json:"chapter_start"`
		ChapterEnd   []int  `json:"chapter_end"`
		PrevVerse    int    `json:"prev_verse"`
		NextVerse    int    `json:"next_verse"`
		PrevChapter  []int  `json:"prev_chapter"`
		NextChapter  []int  `json:"next_chapter"`
	} `json:"passage_meta"`
	Passages []string `json:"passages"`
}

type SearchResponse struct {
	Page    int `json:"page"`
	Total   int `json:"total_results"`
	Results []struct {
		Reference string `json:"reference"`
		Content   string `json:"content"`
	} `json:"results"`
}

type Verse struct {
	Reference string `json:"reference"`
	Color     string `json:"color"`
}

type HighlightedVersesResponse struct {
	Verses []Verse `json:"verses"`
} // {verses: [{"reference": "v43003016", "color": "bg-red-600"}]}

func HighlightedVersesHandler(w http.ResponseWriter, r *http.Request) {

	result := HighlightedVersesResponse{
		Verses: []Verse{
			Verse{Reference: "v43003016", Color: "bg-red-600"},
			Verse{Reference: "v43003017", Color: "bg-violet-600"}},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)

}

func SearchRequestHandler(w http.ResponseWriter, r *http.Request) {
	p := r.URL.Query()
	search := p.Get("search")
	page := "1"
	pageSize := "20"
	params := url.Values{}
	params.Add("q", search)
	params.Add("page", page)
	params.Add("page-size", pageSize)
	api_key := os.Getenv("API_KEY")
	baseURL := "https://api.esv.org/v3/passage/search/"

	// ------------ This is the same as the ApiRequestHandler function ------------- \\

	urlWithParams := baseURL + "?" + params.Encode()

	//log.Println(urlWithParams)
	req, err := http.NewRequest("GET", urlWithParams, nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}

	req.Header.Add("Authorization", "Token "+api_key)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error making request:", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	var result SearchResponse                             // Needs to be changed to SearchResponse from ApiResponse
	if err := json.Unmarshal(body, &result); err != nil { // Parse []byte to go struct pointer
		fmt.Println("Can not unmarshal JSON")
	}
	//w.Header().Set("Content-Type", "text/html")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)

}

func ApiRequestHandler(w http.ResponseWriter, r *http.Request) {
	p := r.URL.Query()
	verse := "John+3:16-21"
	numbers := "false"
	headings := "false"
	extras := "false"

	params := url.Values{}
	if len(p.Get("verse")) > 0 {
		verse = p.Get("verse")
	}
	if p.Get("numbers") == "true" {
		numbers = "true"
	}
	if p.Get("headings") == "true" {
		headings = "true"
	}
	if p.Get("extras") == "true" {
		extras = "true"
	}
	baseURL := "https://api.esv.org/v3/passage/html/"
	api_key := os.Getenv("API_KEY")
	params.Add("q", verse)
	params.Add("include-passage-references", "true")
	params.Add("include-verse-anchors", "true")
	params.Add("include-chapter-numbers", numbers)
	params.Add("include-verse-numbers", numbers)
	params.Add("include-headings", headings)
	params.Add("include-subheadings", headings)
	params.Add("include-footnotes", extras)
	params.Add("include-audio-link", extras)
	// Encode the query parameters and append them to the base URL

	urlWithParams := baseURL + "?" + params.Encode()

	//log.Println(urlWithParams)
	req, err := http.NewRequest("GET", urlWithParams, nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}

	req.Header.Add("Authorization", "Token "+api_key)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error making request:", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	var result ApiResponse
	if err := json.Unmarshal(body, &result); err != nil { // Parse []byte to go struct pointer
		fmt.Println("Can not unmarshal JSON")
	}
	text := strings.Join(result.Passages, "")
	//fmt.Println(text)
	wrapped, err := wrapVerses(text)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	//fmt.Println(wrapped)
	result.Passages = []string{wrapped}
	//w.Header().Set("Content-Type", "text/html")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)
}



func wrapInlineVersesInNode(n *html.Node) {
	// 1) Gather original children
	var origChildren []*html.Node
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		origChildren = append(origChildren, c)
	}

	// 2) Rebuild
	var rebuilt []*html.Node
	i := 0

	for i < len(origChildren) {
		ch := origChildren[i]

		switch {
		// A) verse‐number <b class="verse-num">…</b>: emit as‐is
		case isVerseNumberNode(ch):
			if ch.Parent != nil {
				ch.Parent.RemoveChild(ch)
			}
			rebuilt = append(rebuilt, ch)
			i++

		// B) verse anchor <a class="va" rel="vXXXX">…</a>:
		case isVerseAnchorNode(ch):
			// 1) emit the anchor itself
			if ch.Parent != nil {
				ch.Parent.RemoveChild(ch)
			}
			rebuilt = append(rebuilt, ch)

			// 2) record verseID
			verseID := getVerseID(ch)
			i++

			// 3) collect everything for this verse, but skip wrapping footnotes
			var toWrap []*html.Node

			for i < len(origChildren) {
				sib := origChildren[i]

				// Stop if next anchor or next verse‐number
				if isVerseAnchorNode(sib) || isVerseNumberNode(sib) {
					break
				}

				// B1) Direct <sup class="footnote">
				if sib.Type == html.ElementNode && sib.Data == "sup" {
					isFoot := false
					for _, a := range sib.Attr {
						if a.Key == "class" && strings.Contains(a.Val, "footnote") {
							isFoot = true
							break
						}
					}
					if isFoot {
						// (a) flush current toWrap into a <span class="verse">
						if len(toWrap) > 0 {
							spanNode := &html.Node{
								Type: html.ElementNode,
								Data: "span",
								Attr: []html.Attribute{
									{Key: "class", Val: "verse"},
									{Key: "data-verse", Val: verseID},
								},
							}
							for _, w := range toWrap {
								if w.Parent != nil {
									w.Parent.RemoveChild(w)
								}
								spanNode.AppendChild(w)
							}
							rebuilt = append(rebuilt, spanNode)
							toWrap = nil
						}
						// (b) emit the <sup> itself unwrapped, in correct order
						if sib.Parent != nil {
							sib.Parent.RemoveChild(sib)
						}
						rebuilt = append(rebuilt, sib)
						i++
						continue
					}
				}

				// B2) <span class="woc">…</span> (may contain footnotes inside)
				if sib.Type == html.ElementNode && sib.Data == "span" {
					isWoc := false
					for _, a := range sib.Attr {
						if a.Key == "class" && strings.Contains(a.Val, "woc") {
							isWoc = true
							break
						}
					}
					if isWoc {
						// We'll break this woc into sub‐pieces
						// (a) iterate inner children in order, splitting out footnotes
						var innerToWrap []*html.Node
						for gc := sib.FirstChild; gc != nil; {
							nextGC := gc.NextSibling
							// If gc is a sup.footnote
							if gc.Type == html.ElementNode && gc.Data == "sup" {
								isFoot := false
								for _, a := range gc.Attr {
									if a.Key == "class" && strings.Contains(a.Val, "footnote") {
										isFoot = true
										break
									}
								}
								if isFoot {
									// i) flush any accumulated innerToWrap as a new <span class="woc">
									if len(innerToWrap) > 0 {
										// build a new woc span containing those nodes
										newWoc := &html.Node{
											Type: html.ElementNode,
											Data: "span",
											Attr: []html.Attribute{
												{Key: "class", Val: "woc"},
											},
										}
										for _, w := range innerToWrap {
											if w.Parent != nil {
												w.Parent.RemoveChild(w)
											}
											newWoc.AppendChild(w)
										}
										// now that newWoc is footnote-free, add it to toWrap
										toWrap = append(toWrap, newWoc)
										innerToWrap = nil
									}
									// ii) flush toWrap as verse-span (if any)
									if len(toWrap) > 0 {
										spanNode := &html.Node{
											Type: html.ElementNode,
											Data: "span",
											Attr: []html.Attribute{
												{Key: "class", Val: "verse"},
												{Key: "data-verse", Val: verseID},
											},
										}
										for _, w := range toWrap {
											if w.Parent != nil {
												w.Parent.RemoveChild(w)
											}
											spanNode.AppendChild(w)
										}
										rebuilt = append(rebuilt, spanNode)
										toWrap = nil
									}
									// iii) emit the footnote sup unwrapped
									if gc.Parent != nil {
										gc.Parent.RemoveChild(gc)
									}
									rebuilt = append(rebuilt, gc)
									gc = nextGC
									continue
								}
							}

							// Otherwise, gc is not a footnote: accumulate inside innerToWrap
							if gc.Parent != nil {
								gc.Parent.RemoveChild(gc)
							}
							innerToWrap = append(innerToWrap, gc)
							gc = nextGC
						}

						// (b) after processing all grandchildren of span.woc,
						//     if any innerToWrap remains, crank out one final newWoc
						if len(innerToWrap) > 0 {
							newWoc := &html.Node{
								Type: html.ElementNode,
								Data: "span",
								Attr: []html.Attribute{
									{Key: "class", Val: "woc"},
								},
							}
							for _, w := range innerToWrap {
								if w.Parent != nil {
									w.Parent.RemoveChild(w)
								}
								newWoc.AppendChild(w)
							}
							toWrap = append(toWrap, newWoc)
							innerToWrap = nil
						}

						// (c) remove the original span.woc
						if sib.Parent != nil {
							sib.Parent.RemoveChild(sib)
						}
						i++
						continue
					}
				}

				// B3) Anything else: just collect under toWrap
				toWrap = append(toWrap, sib)
				i++
			}

			// 4) Flush remaining toWrap as a <span class="verse">
			if len(toWrap) > 0 {
				spanNode := &html.Node{
					Type: html.ElementNode,
					Data: "span",
					Attr: []html.Attribute{
						{Key: "class", Val: "verse"},
						{Key: "data-verse", Val: verseID},
					},
				}
				for _, w := range toWrap {
					if w.Parent != nil {
						w.Parent.RemoveChild(w)
					}
					spanNode.AppendChild(w)
				}
				rebuilt = append(rebuilt, spanNode)
				toWrap = nil
			}

		// C) Anything else: emit unchanged
		default:
			if ch.Parent != nil {
				ch.Parent.RemoveChild(ch)
			}
			rebuilt = append(rebuilt, ch)
			i++
		}
	}

	// 5) Replace n’s children with rebuilt list
	for old := n.FirstChild; old != nil; old = n.FirstChild {
		n.RemoveChild(old)
	}
	for _, c := range rebuilt {
		n.AppendChild(c)
	}
}

func wrapInlineVersesInIndentBlock(p *html.Node) {
	for c := p.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode && c.Data == "span" {
			for _, a := range c.Attr {
				if a.Key == "class" && strings.Contains(a.Val, "line") {
					wrapInlineVersesInNode(c)
					break
				}
			}
		}
	}
}

func wrapVerses(htmlInput string) (string, error) {
	// 1) Parse into a full document tree
	doc, err := html.Parse(strings.NewReader(htmlInput))
	if err != nil {
		return "", err
	}

	// 2) Recursively walk
	var recurse func(n *html.Node)
	recurse = func(n *html.Node) {
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			// If it’s a <p> element:
			if c.Type == html.ElementNode && c.Data == "p" {
				// Check if class="block-indent"
				isIndent := false
				for _, a := range c.Attr {
					if a.Key == "class" && strings.Contains(a.Val, "block-indent") {
						isIndent = true
						break
					}
				}

				if isIndent {
					// First recurse inside so nested spans are processed
					recurse(c)
					// Then apply verse‐wrapping to each span.line inside
					wrapInlineVersesInIndentBlock(c)
				} else {
					// Non‐indented <p>: wrap directly, then recurse deeper for footnotes, etc.
					wrapInlineVersesInNode(c)
					recurse(c)
				}

			} else {
				// Any other node (including <h3>) is left alone; just recurse into children
				recurse(c)
			}
		}
	}

	recurse(doc)

	// 3) Find the <body> node
	var body *html.Node
	var findBody func(n *html.Node)
	findBody = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "body" {
			body = n
			return
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if body == nil {
				findBody(c)
			}
		}
	}
	findBody(doc)
	if body == nil {
		// In the unlikely event there's no <body>, just render entire doc
		var fullBuf bytes.Buffer
		if err := html.Render(&fullBuf, doc); err != nil {
			return "", err
		}
		return fullBuf.String(), nil
	}

	// 4) Render only each child of <body> (so no <html> or <body> wrappers)
	var buf bytes.Buffer
	for c := body.FirstChild; c != nil; c = c.NextSibling {
		html.Render(&buf, c)
	}
	return buf.String(), nil
}

func isVerseAnchorNode(n *html.Node) bool {
	if n.Type == html.ElementNode && n.Data == "a" {
		for _, attr := range n.Attr {
			if attr.Key == "class" && strings.Contains(attr.Val, "va") {
				for _, a := range n.Attr {
					if a.Key == "rel" && strings.HasPrefix(a.Val, "v") {
						return true
					}
				}
			}
		}
	}
	return false
}

func isVerseNumberNode(n *html.Node) bool {
	if n.Type == html.ElementNode && n.Data == "b" {
		for _, attr := range n.Attr {
			if attr.Key == "class" && strings.Contains(attr.Val, "verse-num") {
				return true
			}
		}
	}
	return false
}

func getVerseID(n *html.Node) string {
	for _, attr := range n.Attr {
		if attr.Key == "rel" && strings.HasPrefix(attr.Val, "v") {
			return attr.Val
		}
	}
	return ""
}