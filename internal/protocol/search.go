package protocol

// search.go holds the search modality's offer rules (features/search/offer_modality.feature)
// and the one text-flattening routine every search surface shares: the broker cleaning a
// station's reply, the node cleaning its engine's reply, and the harness rendering results.

import (
	"fmt"
	"html"
	"strings"
	"unicode"
)

// ScopeMax is the cap, in characters, on a search offer's scope line.
const ScopeMax = 120

// ValidEngine reports whether e is one of the closed set of search engines.
func ValidEngine(e string) bool {
	switch e {
	case EngineBrave, EngineSearXNG, EngineNative:
		return true
	}
	return false
}

// OfferError is the registration-time check of an offer's modality and engine: "" when the
// offer is acceptable, else the reason it is refused. The modality enum is closed, a search
// offer must name a known engine, and an engine on any other offer is a label with nothing
// behind it.
func (o ModelOffer) OfferError() string {
	if !o.ValidModality() {
		return fmt.Sprintf("offer %q has unknown modality %q", o.Model, o.Modality)
	}
	if o.Modality == ModalitySearch {
		if !ValidEngine(o.Engine) {
			return fmt.Sprintf("search offer %q needs engine brave, searxng or native (got %q)", o.Model, o.Engine)
		}
		return ""
	}
	if o.Engine != "" {
		return fmt.Sprintf("offer %q is not a search offer and cannot carry engine %q", o.Model, o.Engine)
	}
	return ""
}

// CanonicalScope is a search offer's scope line as it may be displayed: flattened to one line
// of plain text and cut at ScopeMax characters.
func CanonicalScope(s string) string { return boundedText(FlattenText(s), ScopeMax) }

// FlattenText cleans one field of provider text: markup out, entities decoded, all whitespace
// (including newlines) collapsed to single spaces.
//
// Search engines wrap query-term matches in <strong>; markup spends a reader's context on
// nothing, and tag-shaped text handed to an agent is worse than noise. The newline collapse is
// load-bearing separately: a newline inside a title or snippet would forge an extra
// "[n] Title / URL" pair that a citation reader would then bind to somebody else's URL.
func FlattenText(s string) string {
	// Strip, then decode, then NEUTRALIZE what decoding may have re-formed: a snippet
	// carrying &lt;strong&gt; would otherwise decode into literal tag-shaped text after the
	// stripper had already run. Decoding first instead would eat legitimate prose ("a < b"),
	// so the angle brackets are blanked rather than the order reversed.
	//
	// CleanText runs LAST, after decoding: an entity can encode a control or format character
	// (&#27; is ESC, &#x202E; is a bidi override), and cleaning before the decode let it back in.
	return CleanText(angleBrackets.Replace(html.UnescapeString(stripTags(s))))
}

// CleanText makes untrusted text safe to show and screen on one line: control characters
// (ESC, NUL, newline) become spaces, format characters (bidi overrides, zero-width
// characters) are deleted, and runs of whitespace collapse to single spaces. Deleting a
// zero-width character rather than spacing it keeps "ch\u200bild" one word for the screen.
func CleanText(s string) string {
	return strings.Join(strings.Fields(strings.Map(func(r rune) rune {
		switch {
		case unicode.IsControl(r):
			return ' '
		case unicode.Is(unicode.Cf, r):
			return -1
		}
		return r
	}, s)), " ")
}

// angleBrackets blanks any angle bracket surviving the strip+decode pass.
var angleBrackets = strings.NewReplacer("<", " ", ">", " ")

// stripTags removes anything between angle brackets. Provider snippets are HTML fragments,
// not documents, so this is deliberately blunt: no parser, no partial-tag ambiguity, and an
// unclosed "<" simply truncates there rather than leaking the rest as markup.
func stripTags(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for {
		i := strings.IndexByte(s, '<')
		if i < 0 {
			b.WriteString(s)
			return b.String()
		}
		b.WriteString(s[:i])
		j := strings.IndexByte(s[i:], '>')
		if j < 0 {
			return b.String()
		}
		s = s[i+j+1:]
	}
}
