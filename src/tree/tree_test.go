package tree

import (
	"testing"
)

func TestNode_Children(t *testing.T) {
	parent := &Node{
		Min: &Point{
			X: 0,
			Y: 0,
		},
		Max: &Point{
			X: 128.0,
			Y: 256.0,
		},
		Depth: 1,
	}

	children := parent.Children()

	for _, ch := range children {
		if (parent.Max.X - parent.Min.X) != (ch.Max.X-ch.Min.X)*2.0 {
			t.Errorf("invalid x size: expected %v, but actual %v", parent.Max.X-parent.Min.X, (ch.Max.X-ch.Min.X)*2.0)
		}
		if (parent.Max.Y - parent.Min.Y) != (ch.Max.Y-ch.Min.Y)*2.0 {
			t.Errorf("invalid y size: expected %v, but actual %v", parent.Max.Y-parent.Min.Y, (ch.Max.Y-ch.Min.Y)*2.0)
		}

		if !parent.IsInside(ch.Min) {
			t.Errorf("invalid position: %v", ch)
		}

		if parent.Depth+1 != ch.Depth {
			t.Errorf("invalid depth: expected %v, but actual %v", parent.Depth, ch.Depth)
		}
	}
}

func TestNode_IsInside(t *testing.T) {
	node := &Node{
		Min: &Point{},
		Max: &Point{
			X: 10.0,
			Y: 10.0,
		},
	}

	testCases := map[string]struct {
		point    *Point
		expected bool
	}{
		"lowest": {
			point:    node.Min,
			expected: true,
		},
		"large": {
			point: &Point{
				X: node.Max.X - 0.0001,
				Y: node.Max.Y - 0.0001,
			},
			expected: true,
		},
		"ng/-x": {point: &Point{
			X: node.Min.X - 0.0001,
			Y: node.Min.Y,
		},
			expected: false,
		},
		"ng/+x": {point: &Point{
			X: node.Max.X - 0.0001,
			Y: node.Max.Y,
		},
			expected: false,
		},
		"ng/-y": {point: &Point{
			X: node.Min.X,
			Y: node.Min.Y - 0.0001,
		},
			expected: false,
		},
		"ng/+y": {point: &Point{
			X: node.Max.X,
			Y: node.Max.Y - 0.0001,
		},
			expected: false,
		},
	}

	for description, tc := range testCases {
		t.Run(description, func(t *testing.T) {
			actual := node.IsInside(tc.point)
			if tc.expected != actual {
				t.Errorf("expected %v, but actual %v", tc.expected, actual)
			}
		})
	}
}

func TestTree_Hash(t *testing.T) {
	tree := NewTree(
		&Point{},
		&Point{
			X: 128,
			Y: 128,
		})

	testCases := []struct {
		point    *Point
		depth    int32
		expected string
	}{
		{
			point:    &Point{},
			depth:    1,
			expected: "0",
		},
		{
			point:    &Point{},
			depth:    2,
			expected: "00",
		},
		{
			point: &Point{
				X: 32,
				Y: 32,
			},
			depth:    2,
			expected: "03",
		},
		{
			point: &Point{
				X: 127,
				Y: 127,
			},
			depth:    1,
			expected: "3",
		},
	}

	for _, tc := range testCases {
		node, actual := tree.Hash(tc.point, tc.depth)
		if !node.IsInside(tc.point) {
			t.Errorf("invalid position: %v", node)
		}
		if int32(len(actual)) != tc.depth {
			t.Errorf("invalid length: expected %v, but actual %v", tc.depth, len(actual))
		}
		if tc.expected != actual {
			t.Errorf("expected %v, but actual %v", tc.expected, actual)
		}
	}
}
