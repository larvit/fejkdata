package tree

// ColumnRead is a record's column that a format reads, where the format only reads that one
// reference or name. The format then counts as the column: it takes the column's datatype and
// null.
type ColumnRead struct{ spec ColumnReadSpec }

// ColumnReadSpec is what NewColumnRead builds a ColumnRead of.
type ColumnReadSpec struct {
	Arm Arm // the format's one read
	// Category and Field name the column. Category is "" for a column of the reading template's
	// own record, which a load already checks with that record's columns.
	Category, Field string
	// Column is the column read. Where Spread is set, the read goes through a choice, and Column is
	// a choice of the column each item holds; a load checks those agree, since nothing else
	// reaches those items.
	Column Node
	Spread bool
}

func NewColumnRead(s ColumnReadSpec) *ColumnRead { return &ColumnRead{s} }

func (c *ColumnRead) Arm() Arm { return c.spec.Arm }

func (c *ColumnRead) Category() string { return c.spec.Category }

func (c *ColumnRead) Column() Node { return c.spec.Column }

func (c *ColumnRead) Field() string { return c.spec.Field }

func (c *ColumnRead) Spread() bool { return c.spec.Spread }
