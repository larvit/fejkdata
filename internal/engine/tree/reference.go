package tree

// ResolvedRef is what a reference resolves to. head is "/" and the category's path from the root,
// the key the category is held under; tail is the path read into the category.
type ResolvedRef struct {
	head string
	tail []string
}

func NewResolvedRef(head string, tail []string) ResolvedRef { return ResolvedRef{head, tail} }

func (r ResolvedRef) Head() string { return r.head }

func (r ResolvedRef) Tail() []string { return r.tail }
