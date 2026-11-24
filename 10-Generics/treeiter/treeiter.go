//go:build !solution

package treeiter

type Tree[T any] interface {
	*T
	Left() *T
	Right() *T
}

func DoInOrder[T any, P Tree[T]](root P, f func(t P)) {
	if root == nil {
		return
	}

	DoInOrder[T, P](root.Left(), f)
	f(root)
	DoInOrder[T, P](root.Right(), f)
}
