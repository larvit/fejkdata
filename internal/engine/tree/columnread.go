package tree

// ColumnRead is a record's column that a format reads, where the format only reads that one
// reference or name. The format then counts as the column: it takes the column's datatype and
// null. category and field name the column; category is "" for a column of the reading
// template's own record, which checkColumns reaches on its own. spread is set where the read goes
// through the items of a choice: column is then a choice of the column each item holds, which
// checkColumns checks for agreement, since nothing else reaches those items.
type ColumnRead struct {
	a               Arm
	category, field string
	column          Node
	spread          bool
}

func NewColumnRead(a Arm, category, field string, column Node, spread bool) *ColumnRead {
	return &ColumnRead{a, category, field, column, spread}
}

func (r *ColumnRead) Arm() Arm { return r.a }

func (r *ColumnRead) Category() string { return r.category }

func (r *ColumnRead) Column() Node { return r.column }

func (r *ColumnRead) Field() string { return r.field }

func (r *ColumnRead) Spread() bool { return r.spread }
