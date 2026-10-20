package ast

// Walk calls fn for every node in nodes, in source order, descending into
// the bodies of the tags that have one (<go:if>, <go:switch>, <go:for>).
// It is the one place that knows how to reach a whole tree, so anything
// that has to see every node - collecting imports, linting, an editor
// highlighter - goes through it instead of recursing on its own.
//
// Ex: Walk(prog.Nodes, func(n Node) { fmt.Printf("%T\n", n) })
func Walk(nodes []Node, fn func(Node)) {
	for _, n := range nodes {
		fn(n)

		switch node := n.(type) {
		case *If:
			Walk(node.Then, fn)
			for i := range node.Elifs {
				Walk(node.Elifs[i].Body, fn)
			}
			Walk(node.Else, fn)
		case *Switch:
			for i := range node.Cases {
				Walk(node.Cases[i].Body, fn)
			}
			Walk(node.Default, fn)
		case *For:
			Walk(node.Body, fn)
		}
	}
}
