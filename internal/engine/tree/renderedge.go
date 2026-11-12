package tree

import (
	"fmt"

	"github.com/larvit/fejkdata/internal/invariant"
)

// A render steps from a node into a child through this file only. It draws one child with the
// Rendered method of the node's kind, or lands on a leaf of a read of a template's ops, and
// RenderEdges lists every child those can give. So a child missing here is one a render cannot
// draw, and no fence walks a list of its own.

// RenderedItem is item i of c, which a render of c draws one of.
func (c *Choice) RenderedItem(i int) Node { return c.items[i] }

// RenderedFormat is the template a render of t expands, in the row it draws.
func (t *Table) RenderedFormat() *Template { return t.formatTemplate }

// RenderedFormat is the template a render of r expands, in the row r selects.
func (r *TableRow) RenderedFormat() *Template { return r.t.formatTemplate }

// RenderedCell is the template c's cell in row renders, or nil where the cell is plain text.
func (c *TableColumn) RenderedCell(row int) *Template { return c.t.CellTemplate(row, c.i) }

// RenderEdge is a child a node renders into, labelled by what reaches it (a field name,
// reference, or choice index) for a readable cycle report.
type RenderEdge struct {
	To    Node
	Label string
}

// RenderEdges lists every child a render of n can step into. It reads what l found, since the
// cycle check runs before Finish.
func RenderEdges(n Node, l *LoadFinds) []RenderEdge {
	switch n := n.(type) {
	case *Choice:
		es := make([]RenderEdge, len(n.items))
		for i := range n.items {
			es[i] = RenderEdge{To: n.RenderedItem(i), Label: fmt.Sprintf("[%d]", i)}
		}
		return es
	case *Template:
		var es []RenderEdge
		for _, o := range l.Of(n).Compiled().Ops() {
			for _, a := range o.AllReads() {
				for _, leaf := range a.Leaves() {
					es = append(es, RenderEdge{leaf, a.Spelling()})
				}
			}
		}
		return es
	case *Table:
		return []RenderEdge{{To: n.RenderedFormat(), Label: "format"}}
	case *TableRow:
		return []RenderEdge{{To: n.RenderedFormat(), Label: "format"}}
	case *TableColumn:
		if len(n.t.cellTemplates) == 0 {
			return nil
		}
		var es []RenderEdge
		for row := 0; row < n.t.rows.Len(); row++ {
			if cell := n.RenderedCell(row); cell != nil {
				es = append(es, RenderEdge{To: cell, Label: n.Name()})
			}
		}
		return es
	case *Folder, *NullItem:
		return nil
	default:
		panic(invariant.Broken("RenderEdges has no case for node %T", n))
	}
}
