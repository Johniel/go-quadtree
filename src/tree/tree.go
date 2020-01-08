package tree

import (
	"strings"
)

// 座標
type Point struct {
	X float64
	Y float64
}

// 木の頂点。1つの領域を示す。
type Node struct {
	Min   *Point
	Max   *Point
	Depth int32
}

func (n *Node) Mid() *Point {
	return &Point{
		X: (n.Max.X-n.Min.X)/2.0 + n.Min.X,
		Y: (n.Max.Y-n.Min.Y)/2.0 + n.Min.Y,
	}
}

func (n *Node) IsInside(p *Point) bool {
	return n.Min.X <= p.X && n.Min.Y <= p.Y && p.X < n.Max.X && p.Y < n.Max.Y
}

// その頂点の子を返す
func (n *Node) Children() []*Node {
	dx := (n.Max.X - n.Min.X) / 2.0
	dy := (n.Max.Y - n.Min.Y) / 2.0

	children := make([]*Node, 1<<2)
	for idx := range children {
		ch := &Node{
			Min: &Point{
				X: n.Min.X,
				Y: n.Min.Y,
			},
			Max: &Point{
				X: n.Min.X + dx,
				Y: n.Min.Y + dy,
			},
			Depth: n.Depth + 1,
		}
		if (idx & (1 << 0)) != 0 {
			ch.Min.X += dx
			ch.Max.X += dx
		}
		if (idx & (1 << 1)) != 0 {
			ch.Min.Y += dy
			ch.Max.Y += dy
		}
		children[idx] = ch
	}
	return children
}

// その頂点と同じ深さの8近傍を返す。
func (n *Node) Adjacent() []*Node {
	dx := n.Max.X - n.Min.X
	dy := n.Max.Y - n.Min.Y

	dirX := []float64{-1, -1, -1, 0, 0, +1, +1, +1}
	dirY := []float64{-1, 0, +1, -1, +1, -1, 0, +1}

	adjacent := make([]*Node, len(dirX))
	for d := 0; d < len(dirX); d++ {
		m := &Node{
			Min: &Point{
				X: n.Min.X,
				Y: n.Min.Y,
			},
			Max: &Point{
				X: n.Max.X,
				Y: n.Max.Y,
			},
			Depth: n.Depth,
		}
		m.Min.X += dirX[d] * dx
		m.Min.Y += dirY[d] * dx
		m.Max.X += dirX[d] * dy
		m.Max.Y += dirY[d] * dy
		adjacent[d] = m
	}

	return adjacent
}

type Tree struct {
	min *Point
	max *Point
}

func NewTree(min *Point, max *Point) *Tree {
	return &Tree{
		min: min,
		max: max,
	}
}

func (t *Tree) Path(p *Point, depth int32) (*Node, string) {
	node := &Node{
		Min: t.min,
		Max: t.max,
	}

	builder := &strings.Builder{}
	label := "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz{|}"
	for node.Depth < depth {
		for idx, ch := range node.Children() {
			if ch.IsInside(p) {
				node = ch
				builder.WriteByte(label[idx])
				break
			}
		}
	}
	return node, builder.String()
}
