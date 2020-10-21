package bfmdrenderer

import (
	"io"
	"log"
	"strconv"

	bf "github.com/russross/blackfriday/v2"
)

// Option defines the functional option type
type Option func(r *Renderer)

// NewRenderer will return a new renderer with sane defaults
func NewRenderer(options ...Option) *Renderer {
	r := &Renderer{}
	for _, option := range options {
		option(r)
	}
	return r
}

// Renderer is a custom Blackfriday renderer
type Renderer struct {
	paragraphDecoration  []byte
	nestedListLevel      int
	nestedListDecoration []byte
	orderedListCounters  []int
}

// Taken from the black friday HTML renderer
func skipParagraphTags(node *bf.Node) bool {
	parent := node.Parent
	if parent != nil && parent.Type == bf.BlockQuote {
		return true
	}

	grandparent := node.Parent.Parent
	if grandparent == nil || grandparent.Type != bf.List {
		return false
	}

	return grandparent.Type == bf.List && grandparent.Tight
}

// RenderNode satisfies the Renderer interface
func (r *Renderer) RenderNode(w io.Writer, node *bf.Node, entering bool) bf.WalkStatus {
	switch node.Type {
	case bf.Document:
		return bf.GoToNext
	case bf.BlockQuote:
		if entering {
			r.paragraphDecoration = append(r.paragraphDecoration, byte('>'), byte(' '))
		} else {
			r.paragraphDecoration = r.paragraphDecoration[:len(r.paragraphDecoration)-2]
		}
		return bf.GoToNext
	case bf.List:
		if entering {
			r.orderedListCounters = append(r.orderedListCounters, 0)
			r.nestedListLevel++
			if r.nestedListLevel > 1 {
				r.nestedListDecoration = append(r.nestedListDecoration, byte(' '), byte(' '))
			}
		} else {
			if r.nestedListLevel > 1 {
				r.nestedListDecoration = r.nestedListDecoration[:len(r.nestedListDecoration)-2]
			} else {
				w.Write([]byte("\n"))
			}
			r.nestedListLevel--
			r.orderedListCounters = r.orderedListCounters[:len(r.orderedListCounters)-1]
		}

		return bf.GoToNext
	case bf.Item:
		if entering {
			w.Write(r.nestedListDecoration)
			if node.Parent.ListFlags&bf.ListTypeOrdered != 0 {
				r.orderedListCounters[len(r.orderedListCounters)-1]++
				w.Write([]byte(strconv.Itoa(r.orderedListCounters[len(r.orderedListCounters)-1])))
				w.Write([]byte{node.ListData.Delimiter})
				w.Write([]byte(" "))
			} else if node.Parent.ListFlags&bf.ListTypeTerm != 0 {
				log.Println("Definition lists not implemented by Renderer")
			} else {
				w.Write([]byte{node.ListData.BulletChar})
				w.Write([]byte(" "))
			}
		}
		return bf.GoToNext
	case bf.Paragraph:
		if entering {
			w.Write(r.paragraphDecoration)
		} else {
			w.Write([]byte("\n"))
			if !skipParagraphTags(node) {
				w.Write([]byte("\n"))
			}
		}
		return bf.GoToNext
	case bf.Heading:
		if entering {
			for i := 0; i < node.Level; i++ {
				w.Write([]byte("#"))
			}
			w.Write([]byte(" "))
		} else {
			w.Write([]byte("\n\n"))
		}
		return bf.GoToNext
	case bf.HorizontalRule:
		w.Write([]byte("---\n\n"))
		return bf.GoToNext
	case bf.Emph:
		w.Write([]byte("*"))
		return bf.GoToNext
	case bf.Strong:
		w.Write([]byte("**"))
		return bf.GoToNext
	case bf.Del:
		w.Write([]byte("~~"))
		return bf.GoToNext
	case bf.Link:
		if entering {
			w.Write([]byte("["))
		} else {
			w.Write([]byte("]("))
			w.Write(node.LinkData.Destination)
			w.Write([]byte(")"))
		}
		return bf.GoToNext
	case bf.Image:
		if entering {
			w.Write([]byte("!["))
		} else {
			w.Write([]byte("]("))
			w.Write(node.LinkData.Destination)
			w.Write([]byte(")"))
		}
		return bf.GoToNext
	case bf.Code:
		w.Write([]byte("`"))
		w.Write(node.Literal)
		w.Write([]byte("`"))
		return bf.GoToNext
	case bf.Text:
		w.Write(node.Literal)
		return bf.GoToNext
	case bf.CodeBlock:
		w.Write([]byte("```"))
		w.Write(node.CodeBlockData.Info)
		w.Write([]byte("\n"))
		w.Write(node.Literal)
		w.Write([]byte("```\n\n"))
		return bf.GoToNext
	case bf.Softbreak:
		log.Println("Soft breaks not implemented by renderer")
	case bf.Hardbreak:
		w.Write([]byte("  \n"))
		return bf.GoToNext
	case bf.HTMLBlock:
		fallthrough
	case bf.HTMLSpan:
		log.Println("HTML elements not implemented by renderer")
	case bf.Table:
		fallthrough
	case bf.TableCell:
		fallthrough
	case bf.TableHead:
		fallthrough
	case bf.TableBody:
		fallthrough
	case bf.TableRow:
		log.Println("Markdown tables not implemented by renderer")
	default:
		log.Printf("Unknown BlackFriday Node type '%s'\n", node.Type)
	}

	return bf.SkipChildren
}

// RenderHeader satisfies the Renderer interface
func (r *Renderer) RenderHeader(w io.Writer, ast *bf.Node) {
	// Nothing required here
}

// RenderFooter satisfies the Renderer interface
func (r *Renderer) RenderFooter(w io.Writer, ast *bf.Node) {
	// Nothing required here
}
