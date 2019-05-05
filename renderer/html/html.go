package html

import (
	"bytes"
	"fmt"
	"strconv"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
)

// A Config struct has configurations for the HTML based renderers.
type Config struct {
	Writer    Writer
	HardWraps bool
	XHTML     bool
	Unsafe    bool
}

// NewConfig returns a new Config with defaults.
func NewConfig() Config {
	return Config{
		Writer:    DefaultWriter,
		HardWraps: false,
		XHTML:     false,
		Unsafe:    false,
	}
}

// SetOption implements renderer.NodeRenderer.SetOption.
func (c *Config) SetOption(name renderer.OptionName, value interface{}) {
	switch name {
	case HardWraps:
		c.HardWraps = value.(bool)
	case XHTML:
		c.XHTML = value.(bool)
	case Unsafe:
		c.Unsafe = value.(bool)
	case TextWriter:
		c.Writer = value.(Writer)
	}
}

// An Option interface sets options for HTML based renderers.
type Option interface {
	SetHTMLOption(*Config)
}

// TextWriter is an option name used in WithWriter.
const TextWriter renderer.OptionName = "Writer"

type withWriter struct {
	value Writer
}

func (o *withWriter) SetConfig(c *renderer.Config) {
	c.Options[TextWriter] = o.value
}

func (o *withWriter) SetHTMLOption(c *Config) {
	c.Writer = o.value
}

// WithWriter is a functional option that allow you to set the given writer to
// the renderer.
func WithWriter(writer Writer) interface {
	renderer.Option
	Option
} {
	return &withWriter{writer}
}

// HardWraps is an option name used in WithHardWraps.
const HardWraps renderer.OptionName = "HardWraps"

type withHardWraps struct {
}

func (o *withHardWraps) SetConfig(c *renderer.Config) {
	c.Options[HardWraps] = true
}

func (o *withHardWraps) SetHTMLOption(c *Config) {
	c.HardWraps = true
}

// WithHardWraps is a functional option that indicates whether softline breaks
// should be rendered as '<br>'.
func WithHardWraps() interface {
	renderer.Option
	Option
} {
	return &withHardWraps{}
}

// XHTML is an option name used in WithXHTML.
const XHTML renderer.OptionName = "XHTML"

type withXHTML struct {
}

func (o *withXHTML) SetConfig(c *renderer.Config) {
	c.Options[XHTML] = true
}

func (o *withXHTML) SetHTMLOption(c *Config) {
	c.XHTML = true
}

// WithXHTML is a functional option indicates that nodes should be rendered in
// xhtml instead of HTML5.
func WithXHTML() interface {
	Option
	renderer.Option
} {
	return &withXHTML{}
}

// Unsafe is an option name used in WithUnsafe.
const Unsafe renderer.OptionName = "Unsafe"

type withUnsafe struct {
}

func (o *withUnsafe) SetConfig(c *renderer.Config) {
	c.Options[Unsafe] = true
}

func (o *withUnsafe) SetHTMLOption(c *Config) {
	c.Unsafe = true
}

// WithUnsafe is a functional option that renders dangerous contents
// (raw htmls and potentially dangerous links) as it is.
func WithUnsafe() interface {
	renderer.Option
	Option
} {
	return &withUnsafe{}
}

// A Renderer struct is an implementation of renderer.NodeRenderer that renders
// nodes as (X)HTML.
type Renderer struct {
	Config
}

// NewRenderer returns a new Renderer with given options.
func NewRenderer(opts ...Option) renderer.NodeRenderer {
	r := &Renderer{
		Config: NewConfig(),
	}

	for _, opt := range opts {
		opt.SetHTMLOption(&r.Config)
	}
	return r
}

// RegisterFuncs implements NodeRenderer.RegisterFuncs .
func (r *Renderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	// blocks

	reg.Register(ast.KindDocument, r.renderDocument)
	reg.Register(ast.KindHeading, r.renderHeading)
	reg.Register(ast.KindBlockquote, r.renderBlockquote)
	reg.Register(ast.KindCodeBlock, r.renderCodeBlock)
	reg.Register(ast.KindFencedCodeBlock, r.renderFencedCodeBlock)
	reg.Register(ast.KindHTMLBlock, r.renderHTMLBlock)
	reg.Register(ast.KindList, r.renderList)
	reg.Register(ast.KindListItem, r.renderListItem)
	reg.Register(ast.KindParagraph, r.renderParagraph)
	reg.Register(ast.KindTextBlock, r.renderTextBlock)
	reg.Register(ast.KindThemanticBreak, r.renderThemanticBreak)

	// inlines

	reg.Register(ast.KindAutoLink, r.renderAutoLink)
	reg.Register(ast.KindCodeSpan, r.renderCodeSpan)
	reg.Register(ast.KindEmphasis, r.renderEmphasis)
	reg.Register(ast.KindImage, r.renderImage)
	reg.Register(ast.KindLink, r.renderLink)
	reg.Register(ast.KindRawHTML, r.renderRawHTML)
	reg.Register(ast.KindText, r.renderText)
}

func (r *Renderer) writeLines(w util.BufWriter, source []byte, n ast.Node) {
	l := n.Lines().Len()
	for i := 0; i < l; i++ {
		line := n.Lines().At(i)
		r.Writer.RawWrite(w, line.Value(source))
	}
}

func (r *Renderer) renderDocument(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	// nothing to do
	return ast.WalkContinue, nil
}

var attrNameID = []byte("id")

func (r *Renderer) renderHeading(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*ast.Heading)
	if entering {
		w.WriteString("<h")
		w.WriteByte("0123456"[n.Level])
		if n.Attributes() != nil {
			r.RenderAttributes(w, node)
		}
		w.WriteByte('>')
	} else {
		w.WriteString("</h")
		w.WriteByte("0123456"[n.Level])
		w.WriteString(">\n")
	}
	return ast.WalkContinue, nil
}

func (r *Renderer) renderBlockquote(w util.BufWriter, source []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		w.WriteString("<blockquote>\n")
	} else {
		w.WriteString("</blockquote>\n")
	}
	return ast.WalkContinue, nil
}

func (r *Renderer) renderCodeBlock(w util.BufWriter, source []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		w.WriteString("<pre><code>")
		r.writeLines(w, source, n)
	} else {
		w.WriteString("</code></pre>\n")
	}
	return ast.WalkContinue, nil
}

func (r *Renderer) renderFencedCodeBlock(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*ast.FencedCodeBlock)
	if entering {
		w.WriteString("<pre><code")
		if n.Info != nil {
			segment := n.Info.Segment
			info := segment.Value(source)
			i := 0
			for ; i < len(info); i++ {
				if info[i] == ' ' {
					break
				}
			}
			language := info[:i]
			w.WriteString(" class=\"language-")
			r.Writer.Write(w, language)
			w.WriteString("\"")
		}
		w.WriteByte('>')
		r.writeLines(w, source, n)
	} else {
		w.WriteString("</code></pre>\n")
	}
	return ast.WalkContinue, nil
}

func (r *Renderer) renderHTMLBlock(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*ast.HTMLBlock)
	if entering {
		if r.Unsafe {
			l := n.Lines().Len()
			for i := 0; i < l; i++ {
				line := n.Lines().At(i)
				w.Write(line.Value(source))
			}
		} else {
			w.WriteString("<!-- raw HTML omitted -->\n")
		}
	} else {
		if n.HasClosure() {
			if r.Unsafe {
				closure := n.ClosureLine
				w.Write(closure.Value(source))
			} else {
				w.WriteString("<!-- raw HTML omitted -->\n")
			}
		}
	}
	return ast.WalkContinue, nil
}

func (r *Renderer) renderList(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*ast.List)
	tag := "ul"
	if n.IsOrdered() {
		tag = "ol"
	}
	if entering {
		w.WriteByte('<')
		w.WriteString(tag)
		if n.IsOrdered() && n.Start != 1 {
			fmt.Fprintf(w, " start=\"%d\">\n", n.Start)
		} else {
			w.WriteString(">\n")
		}
	} else {
		w.WriteString("</")
		w.WriteString(tag)
		w.WriteString(">\n")
	}
	return ast.WalkContinue, nil
}

func (r *Renderer) renderListItem(w util.BufWriter, source []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		w.WriteString("<li>")
		fc := n.FirstChild()
		if fc != nil {
			if _, ok := fc.(*ast.TextBlock); !ok {
				w.WriteByte('\n')
			}
		}
	} else {
		w.WriteString("</li>\n")
	}
	return ast.WalkContinue, nil
}

func (r *Renderer) renderParagraph(w util.BufWriter, source []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		w.WriteString("<p>")
	} else {
		w.WriteString("</p>\n")
	}
	return ast.WalkContinue, nil
}

func (r *Renderer) renderTextBlock(w util.BufWriter, source []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		if _, ok := n.NextSibling().(ast.Node); ok && n.FirstChild() != nil {
			w.WriteByte('\n')
		}
	}
	return ast.WalkContinue, nil
}

func (r *Renderer) renderThemanticBreak(w util.BufWriter, source []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}
	if r.XHTML {
		w.WriteString("<hr />\n")
	} else {
		w.WriteString("<hr>\n")
	}
	return ast.WalkContinue, nil
}

func (r *Renderer) renderAutoLink(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*ast.AutoLink)
	if !entering {
		return ast.WalkContinue, nil
	}
	w.WriteString(`<a href="`)
	segment := n.Value.Segment
	value := segment.Value(source)
	if n.AutoLinkType == ast.AutoLinkEmail && !bytes.HasPrefix(bytes.ToLower(value), []byte("mailto:")) {
		w.WriteString("mailto:")
	}
	w.Write(util.EscapeHTML(util.URLEscape(value, false)))
	w.WriteString(`">`)
	w.Write(util.EscapeHTML(value))
	w.WriteString(`</a>`)
	return ast.WalkContinue, nil
}

func (r *Renderer) renderCodeSpan(w util.BufWriter, source []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		w.WriteString("<code>")
		for c := n.FirstChild(); c != nil; c = c.NextSibling() {
			segment := c.(*ast.Text).Segment
			value := segment.Value(source)
			if bytes.HasSuffix(value, []byte("\n")) {
				r.Writer.RawWrite(w, value[:len(value)-1])
				if c != n.LastChild() {
					r.Writer.RawWrite(w, []byte(" "))
				}
			} else {
				r.Writer.RawWrite(w, value)
			}
		}
		return ast.WalkSkipChildren, nil
	}
	w.WriteString("</code>")
	return ast.WalkContinue, nil
}

func (r *Renderer) renderEmphasis(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*ast.Emphasis)
	tag := "em"
	if n.Level == 2 {
		tag = "strong"
	}
	if entering {
		w.WriteByte('<')
		w.WriteString(tag)
		w.WriteByte('>')
	} else {
		w.WriteString("</")
		w.WriteString(tag)
		w.WriteByte('>')
	}
	return ast.WalkContinue, nil
}

func (r *Renderer) renderLink(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*ast.Link)
	if entering {
		w.WriteString("<a href=\"")
		if r.Unsafe || !IsDangerousURL(n.Destination) {
			w.Write(util.EscapeHTML(util.URLEscape(n.Destination, true)))
		}
		w.WriteByte('"')
		if n.Title != nil {
			w.WriteString(` title="`)
			r.Writer.Write(w, n.Title)
			w.WriteByte('"')
		}
		w.WriteByte('>')
	} else {
		w.WriteString("</a>")
	}
	return ast.WalkContinue, nil
}
func (r *Renderer) renderImage(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}
	n := node.(*ast.Image)
	w.WriteString("<img src=\"")
	if r.Unsafe || !IsDangerousURL(n.Destination) {
		w.Write(util.EscapeHTML(util.URLEscape(n.Destination, true)))
	}
	w.WriteString(`" alt="`)
	w.Write(n.Text(source))
	w.WriteByte('"')
	if n.Title != nil {
		w.WriteString(` title="`)
		r.Writer.Write(w, n.Title)
		w.WriteByte('"')
	}
	if r.XHTML {
		w.WriteString(" />")
	} else {
		w.WriteString(">")
	}
	return ast.WalkSkipChildren, nil
}

func (r *Renderer) renderRawHTML(w util.BufWriter, source []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if r.Unsafe {
		return ast.WalkContinue, nil
	}
	w.WriteString("<!-- raw HTML omitted -->")
	return ast.WalkSkipChildren, nil
}

func (r *Renderer) renderText(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}
	n := node.(*ast.Text)
	segment := n.Segment
	if n.IsRaw() {
		w.Write(segment.Value(source))
	} else {
		r.Writer.Write(w, segment.Value(source))
		if n.HardLineBreak() || (n.SoftLineBreak() && r.HardWraps) {
			if r.XHTML {
				w.WriteString("<br />\n")
			} else {
				w.WriteString("<br>\n")
			}
		} else if n.SoftLineBreak() {
			w.WriteByte('\n')
		}
	}
	return ast.WalkContinue, nil
}

// RenderAttributes renders given node's attributes.
func (r *Renderer) RenderAttributes(w util.BufWriter, node ast.Node) {
	for _, attr := range node.Attributes() {
		w.WriteString(" ")
		w.Write(attr.Name)
		w.WriteString(`="`)
		w.Write(attr.Value)
		w.WriteByte('"')
	}
}

// A Writer interface wirtes textual contents to a writer.
type Writer interface {
	// Write writes the given source to writer with resolving references and unescaping
	// backslash escaped characters.
	Write(writer util.BufWriter, source []byte)

	// RawWrite wirtes the given source to writer without resolving references and
	// unescaping backslash escaped characters.
	RawWrite(writer util.BufWriter, source []byte)
}

type defaultWriter struct {
}

func escapeRune(writer util.BufWriter, r rune) {
	if r < 256 {
		v := util.EscapeHTMLByte(byte(r))
		if v != nil {
			writer.Write(v)
			return
		}
	}
	writer.WriteRune(util.ToValidRune(r))
}

func (d *defaultWriter) RawWrite(writer util.BufWriter, source []byte) {
	n := 0
	l := len(source)
	for i := 0; i < l; i++ {
		v := util.EscapeHTMLByte(source[i])
		if v != nil {
			writer.Write(source[i-n : i])
			n = 0
			writer.Write(v)
			continue
		}
		n++
	}
	if n != 0 {
		writer.Write(source[l-n:])
	}
}

func (d *defaultWriter) Write(writer util.BufWriter, source []byte) {
	escaped := false
	ok := false
	limit := len(source)
	n := 0
	for i := 0; i < limit; i++ {
		c := source[i]
		if escaped {
			if util.IsPunct(c) {
				d.RawWrite(writer, source[n:i-1])
				n = i
				escaped = false
				continue
			}
		}
		if c == '&' {
			pos := i
			next := i + 1
			if next < limit && source[next] == '#' {
				nnext := next + 1
				nc := source[nnext]
				// code point like #x22;
				if nnext < limit && nc == 'x' || nc == 'X' {
					start := nnext + 1
					i, ok = util.ReadWhile(source, [2]int{start, limit}, util.IsHexDecimal)
					if ok && i < limit && source[i] == ';' {
						v, _ := strconv.ParseUint(util.BytesToReadOnlyString(source[start:i]), 16, 32)
						d.RawWrite(writer, source[n:pos])
						n = i + 1
						escapeRune(writer, rune(v))
						continue
					}
					// code point like #1234;
				} else if nc >= '0' && nc <= '9' {
					start := nnext
					i, ok = util.ReadWhile(source, [2]int{start, limit}, util.IsNumeric)
					if ok && i < limit && i-start < 8 && source[i] == ';' {
						v, _ := strconv.ParseUint(util.BytesToReadOnlyString(source[start:i]), 0, 32)
						d.RawWrite(writer, source[n:pos])
						n = i + 1
						escapeRune(writer, rune(v))
						continue
					}
				}
			} else {
				start := next
				i, ok = util.ReadWhile(source, [2]int{start, limit}, util.IsAlphaNumeric)
				// entity reference
				if ok && i < limit && source[i] == ';' {
					name := util.BytesToReadOnlyString(source[start:i])
					entity, ok := util.LookUpHTML5EntityByName(name)
					if ok {
						d.RawWrite(writer, source[n:pos])
						n = i + 1
						d.RawWrite(writer, entity.Characters)
						continue
					}
				}
			}
			i = next - 1
		}
		if c == '\\' {
			escaped = true
			continue
		}
		escaped = false
	}
	d.RawWrite(writer, source[n:len(source)])
}

// DefaultWriter is a default implementation of the Writer.
var DefaultWriter = &defaultWriter{}

var bDataImage = []byte("data:image/")
var bPng = []byte("png;")
var bGif = []byte("gif;")
var bJpeg = []byte("jpeg;")
var bWebp = []byte("webp;")
var bJs = []byte("javascript:")
var bVb = []byte("vbscript:")
var bFile = []byte("file:")
var bData = []byte("data:")

// IsDangerousURL returns true if the given url seems a potentially dangerous url,
// otherwise false.
func IsDangerousURL(url []byte) bool {
	if bytes.HasPrefix(url, bDataImage) && len(url) >= 11 {
		v := url[11:]
		if bytes.HasPrefix(v, bPng) || bytes.HasPrefix(v, bGif) ||
			bytes.HasPrefix(v, bJpeg) || bytes.HasPrefix(v, bWebp) {
			return false
		}
		return true
	}
	return bytes.HasPrefix(url, bJs) || bytes.HasPrefix(url, bVb) ||
		bytes.HasPrefix(url, bFile) || bytes.HasPrefix(url, bData)
}
