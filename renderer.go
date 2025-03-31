package main

import (
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/lozovskaya/go-commonmark"
)

type RDXBlob struct {
	Tokens []Token
}

type Token struct {
	Type    string  `json:"type"`
	Value   string  `json:"value"`
	Attrs   map[string]string  `json:"attrs"`
}

// RenderRDX converts blocks to RDXBlob
func RenderRDX(blocks []*commonmark.RootBlock, refMap commonmark.ReferenceMap) RDXBlob {
	var blob RDXBlob
	for _, b := range blocks {
		appendRDX(&blob, b.Source, refMap, &b.Block)
	}
	return blob
}

func appendRDX(blob *RDXBlob, source []byte, refMap commonmark.ReferenceMap, block *commonmark.Block) {
	switch block.Kind() {
	case commonmark.ParagraphKind:
		blob.Tokens = append(blob.Tokens, Token{Type: "paragraph_open"})
		appendChildrenRDX(blob, source, refMap, block, false)
		blob.Tokens = append(blob.Tokens, Token{Type: "paragraph_close"})
	case commonmark.ThematicBreakKind:
		blob.Tokens = append(blob.Tokens, Token{Type: "thematic_break"})
	case commonmark.ATXHeadingKind, commonmark.SetextHeadingKind:
		level := strconv.Itoa(block.HeadingLevel())
		blob.Tokens = append(blob.Tokens, Token{
			Type:  "heading_open",
			Attrs: map[string]string{"level": level},
		})
		appendChildrenRDX(blob, source, refMap, block, false)
		blob.Tokens = append(blob.Tokens, Token{
			Type:  "heading_close",
			Attrs: map[string]string{"level": level},
		})
	case commonmark.IndentedCodeBlockKind, commonmark.FencedCodeBlockKind:
		attrs := make(map[string]string)
		if info := block.InfoString(); info != nil {
			words := strings.Fields(info.Text(source))
			if len(words) > 0 {
				attrs["info"] = words[0]
			}
		}
		blob.Tokens = append(blob.Tokens, Token{Type: "code_block_open", Attrs: attrs})
		appendChildrenRDX(blob, source, refMap, block, false)
		blob.Tokens = append(blob.Tokens, Token{Type: "code_block_close"})
	case commonmark.BlockQuoteKind:
		blob.Tokens = append(blob.Tokens, Token{Type: "blockquote_open"})
		appendChildrenRDX(blob, source, refMap, block, false)
		blob.Tokens = append(blob.Tokens, Token{Type: "blockquote_close"})
	case commonmark.ListKind:
		attrs := make(map[string]string)
		if block.IsOrderedList() {
			attrs["type"] = "ordered"
		} else {
			attrs["type"] = "bullet"
		}
		blob.Tokens = append(blob.Tokens, Token{Type: "list_open", Attrs: attrs})
		appendChildrenRDX(blob, source, refMap, block, false)
		blob.Tokens = append(blob.Tokens, Token{Type: "list_close"})
	case commonmark.ListItemKind:
		blob.Tokens = append(blob.Tokens, Token{Type: "list_item_open"})
		appendChildrenRDX(blob, source, refMap, block, block.IsTightList())
		blob.Tokens = append(blob.Tokens, Token{Type: "list_item_close"})
	case commonmark.HTMLBlockKind:
		appendChildrenRDX(blob, source, refMap, block, false)
	}
}

func appendChildrenRDX(blob *RDXBlob, source []byte, refMap commonmark.ReferenceMap, parent *commonmark.Block, tight bool) {
	switch {
	case parent != nil && len(parent.InlineChildren()) > 0:
		for _, c := range parent.InlineChildren() {
			appendInlineRDX(blob, source, refMap, c)
		}
	case parent != nil && len(parent.BlockChildren()) > 0:
		for _, c := range parent.BlockChildren() {
			if tight && c.Kind() == commonmark.ParagraphKind {
				appendChildrenRDX(blob, source, refMap, c, false)
			} else {
				appendRDX(blob, source, refMap, c)
			}
		}
	}
}

func appendInlineRDX(blob *RDXBlob, source []byte, refMap commonmark.ReferenceMap, inline *commonmark.Inline) {
	switch inline.Kind() {
	case commonmark.TextKind, commonmark.UnparsedKind, commonmark.CharacterReferenceKind:
		blob.Tokens = append(blob.Tokens, Token{
			Type:    "text",
			Value: inline.Text(source),
		})
	case commonmark.RawHTMLKind:
		blob.Tokens = append(blob.Tokens, Token{
			Type:    "html_inline",
			Value: inline.Text(source),
		})
	case commonmark.SoftLineBreakKind:
		blob.Tokens = append(blob.Tokens, Token{Type: "softbreak"})
	case commonmark.HardLineBreakKind:
		blob.Tokens = append(blob.Tokens, Token{Type: "hardbreak"})
	case commonmark.EmphasisKind:
		blob.Tokens = append(blob.Tokens, Token{Type: "em_open"})
		for _, c := range inline.Children() {
			appendInlineRDX(blob, source, refMap, c)
		}
		blob.Tokens = append(blob.Tokens, Token{Type: "em_close"})
	case commonmark.StrongKind:
		blob.Tokens = append(blob.Tokens, Token{Type: "strong_open"})
		for _, c := range inline.Children() {
			appendInlineRDX(blob, source, refMap, c)
		}
		blob.Tokens = append(blob.Tokens, Token{Type: "strong_close"})
	case commonmark.CodeSpanKind:
		blob.Tokens = append(blob.Tokens, Token{Type: "code_inline_open"})
		for _, c := range inline.Children() {
			appendInlineRDX(blob, source, refMap, c)
		}
		blob.Tokens = append(blob.Tokens, Token{Type: "code_inline_close"})
	case commonmark.LinkKind:
		var def commonmark.LinkDefinition
		if ref := inline.LinkReference(); ref != "" {
			def = refMap[ref]
		} else {
			title := inline.LinkTitle()
			def = commonmark.LinkDefinition{
				Destination:  inline.LinkDestination().Text(source),
				Title:        title.Text(source),
				TitlePresent: title != nil,
			}
		}
		attrs := map[string]string{
			"href": NormalizeURI(def.Destination),
		}
		if def.TitlePresent {
			attrs["title"] = def.Title
		}
		blob.Tokens = append(blob.Tokens, Token{Type: "link_open", Attrs: attrs})
		for _, c := range inline.Children() {
			appendInlineRDX(blob, source, refMap, c)
		}
		blob.Tokens = append(blob.Tokens, Token{Type: "link_close"})
	case commonmark.ImageKind:
		var def commonmark.LinkDefinition
		if ref := inline.LinkReference(); ref != "" {
			def = refMap[ref]
		} else {
			title := inline.LinkTitle()
			def = commonmark.LinkDefinition{
				Destination:  inline.LinkDestination().Text(source),
				Title:        title.Text(source),
				TitlePresent: title != nil,
			}
		}
		attrs := map[string]string{
			"src": NormalizeURI(def.Destination),
		}
		if def.TitlePresent {
			attrs["title"] = def.Title
		}
		attrs["alt"] = getAltText(source, inline)
		blob.Tokens = append(blob.Tokens, Token{Type: "image", Attrs: attrs})
	case commonmark.AutolinkKind:
		destination := inline.Children()[0].Text(source)
		attrs := map[string]string{
			"href": NormalizeURI(destination),
		}
		blob.Tokens = append(blob.Tokens, Token{Type: "link_open", Attrs: attrs})
		blob.Tokens = append(blob.Tokens, Token{
			Type:    "text",
			Value: destination,
		})
		blob.Tokens = append(blob.Tokens, Token{Type: "link_close"})
	case commonmark.IndentKind:
		content := strings.Repeat(" ", inline.IndentWidth())
		blob.Tokens = append(blob.Tokens, Token{
			Type:    "text",
			Value: content,
		})
	case commonmark.HTMLTagKind:
		for _, c := range inline.Children() {
			appendInlineRDX(blob, source, refMap, c)
		}
	}
}

func getAltText(source []byte, parent *commonmark.Inline) string {
	var buf strings.Builder
	stack := []*commonmark.Inline{parent}
	
	for len(stack) > 0 {
		curr := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		switch curr.Kind() {
		case commonmark.TextKind:
			buf.WriteString(curr.Text(source))
		case commonmark.IndentKind, commonmark.SoftLineBreakKind, commonmark.HardLineBreakKind:
			buf.WriteByte(' ')
		case commonmark.LinkDestinationKind, commonmark.LinkTitleKind, commonmark.LinkLabelKind:
			// Ignore
		default:
			for i := len(curr.Children()) - 1; i >= 0; i-- {
				stack = append(stack, curr.Children()[i])
			}
		}
	}
	return buf.String()
}

// NormalizeURI percent-encodes any characters in a string
// that are not reserved or unreserved URI characters.
func NormalizeURI(s string) string {
	const safeSet = `;/?:@&=+$,-_.!~*'()#`

	sb := new(strings.Builder)
	sb.Grow(len(s))
	skip := 0
	var buf [utf8.UTFMax]byte
	for i, c := range s {
		if skip > 0 {
			skip--
			sb.WriteRune(c)
			continue
		}
		switch {
		case c == '%':
			if i+2 < len(s) && isHex(s[i+1]) && isHex(s[i+2]) {
				skip = 2
				sb.WriteByte('%')
			} else {
				sb.WriteString("%25")
			}
		case (c < 0x80 && (isASCIILetter(byte(c)) || isASCIIDigit(byte(c)))) || strings.ContainsRune(safeSet, c):
			sb.WriteRune(c)
		default:
			n := utf8.EncodeRune(buf[:], c)
			for _, b := range buf[:n] {
				sb.WriteByte('%')
				sb.WriteByte(urlHexDigit(b >> 4))
				sb.WriteByte(urlHexDigit(b & 0x0f))
			}
		}
	}
	return sb.String()
}

func isHex(c byte) bool {
	return 'a' <= c && c <= 'f' || 'A' <= c && c <= 'f' || isASCIIDigit(c)
}

func urlHexDigit(x byte) byte {
	switch {
	case x < 0xa:
		return '0' + x
	case x < 0x10:
		return 'A' + x - 0xa
	default:
		panic("out of bounds")
	}
}

func isASCIILetter(c byte) bool {
	return ('a' <= c && c <= 'z') || ('A' <= c && c <= 'Z')
}

func isASCIIDigit(c byte) bool {
	return '0' <= c && c <= '9'
}