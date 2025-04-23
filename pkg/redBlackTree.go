package pkg

import (
	"fmt"
	"strconv"
)

type color bool

const (
	black, red color = true, false
)

func Comparator(value1, value2 int) int {
	if value1 < value2 {
		return -1
	}
	if value1 > value2 {
		return 1
	}
	return 0
}

// Node and Tree structs
type Node struct {
	Key    int
	Value  interface{}
	color  color
	Left   *Node
	Right  *Node
	Parent *Node
}

type Tree struct {
	Root *Node
	size int
}

func NewRBTree() *Tree {
	return &Tree{Root: nil, size: 0}
}

// Color of node
func nodeColor(node *Node) color {
	if node == nil {
		return black
	}
	return node.color
}

// Relatives of nodes
func (node *Node) grandParent() *Node {
	if node != nil && node.Parent != nil {
		return node.Parent.Parent
	}
	return nil
}

func (node *Node) uncle() *Node {
	if node == nil || node.Parent == nil || node.Parent.Parent == nil {
		return nil
	}
	return node.Parent.sibling()
}

func (node *Node) sibling() *Node {
	if node == nil || node.Parent == nil {
		return nil
	}
	if node == node.Parent.Left {
		return node.Parent.Right
	}
	return node.Parent.Left
}

// Search for a node with the given key in the Red-Black Tree.
func (tree *Tree) Search(key int) *Node {
	node := tree.Root
	for node != nil {
		compare := Comparator(key, node.Key)
		switch {
		case compare == 0:
			return node
		case compare < 0:
			node = node.Left
		case compare > 0:
			node = node.Right
		}
	}
	return nil
}

func (node *Node) String() string {
	if node.color {
		return fmt.Sprintf("%v(B)", node.Key)
	} else {
		return fmt.Sprintf("%v(R)", node.Key)
	}
}

func Output(node *Node, prefix string, isTail bool, str *string) {
	if str == nil || node == nil {
		return
	}
	if node.Right != nil {
		newPrefix := prefix
		if isTail {
			newPrefix += "    "
		} else {
			newPrefix += "    "
		}
		Output(node.Right, newPrefix, false, str)
	}
	*str += prefix
	if isTail {
		*str += "└── "
	} else {
		*str += "┌── "
	}
	*str += node.String() + "\n"
	if node.Left != nil {
		newPrefix := prefix
		if isTail {
			newPrefix += "    "
		} else {
			newPrefix += "    "
		}
		Output(node.Left, newPrefix, true, str)
	}
}

// PreOrderTravers function performs a pre-order traversal of the Red-Black Tree.
func (tree *Tree) PreOrderTravers(node *Node) string {
	var str string
	if node != nil {
		str += strconv.Itoa(node.Key) + " "
		str += tree.PreOrderTravers(node.Left) + " "
		str += tree.PreOrderTravers(node.Right) + " "
	}
	return str
}

// InOrderTravers function performs an in-order traversal of the Red-Black Tree.
func (tree *Tree) InOrderTravers(node *Node) string {
	var str string
	if node != nil {
		str += tree.InOrderTravers(node.Left) + " "
		str += strconv.Itoa(node.Key) + " "
		str += tree.InOrderTravers(node.Right) + " "
	}
	return str
}

// PostOrderTravers function performs a post-order traversal of the Red-Black Tree.
func (tree *Tree) PostOrderTravers(node *Node) string {
	var str string
	if node != nil {
		str += tree.PostOrderTravers(node.Left) + " "
		str += tree.PostOrderTravers(node.Right) + " "
		str += strconv.Itoa(node.Key) + " "
	}
	return str
}

// LevelOrderTravers performs a level-order traversal of the Red-Black Tree.
func (tree *Tree) LevelOrderTravers(root *Node) string {
	var str string
	if root == nil {
		return str
	}
	queue := make([]*Node, 0)
	queue = append(queue, root)
	for len(queue) > 0 {
		node := queue[0]
		queue = queue[1:]
		str += strconv.Itoa(node.Key) + " "
		if node.Left != nil {
			queue = append(queue, node.Left)
		}
		if node.Right != nil {
			queue = append(queue, node.Right)
		}
	}
	return str
}

// Insert and its cases
func (tree *Tree) Insert(key int, value interface{}) {
	var insertedNode *Node
	if tree.Root == nil {
		tree.Root = &Node{Key: key, color: red, Value: value}
		insertedNode = tree.Root
	} else {
		node := tree.Root
		loop := true
		for loop {
			compare := Comparator(key, node.Key)
			switch compare {
			case 0:
				node.Key = key
				return
			case -1:
				if node.Left == nil {
					node.Left = &Node{Key: key, color: red, Value: value}
					insertedNode = node.Left
					loop = false
				} else {
					node = node.Left
				}
			case 1:
				if node.Right == nil {
					node.Right = &Node{Key: key, color: red, Value: value}
					insertedNode = node.Right
					loop = false
				} else {
					node = node.Right
				}
			}
		}
		insertedNode.Parent = node
	}
	tree.insertCase1(insertedNode)
	tree.size++
}

func (tree *Tree) insertCase1(node *Node) {
	if node.Parent == nil {
		node.color = black
	} else {
		tree.insertCase2(node)
	}
}

func (tree *Tree) insertCase2(node *Node) {
	if nodeColor(node.Parent) == black {
		return
	}
	tree.insertCase3(node)
}

func (tree *Tree) insertCase3(node *Node) {
	uncle := node.uncle()
	if nodeColor(uncle) == red {
		node.Parent.color = black
		uncle.color = black
		node.grandParent().color = red
		tree.insertCase1(node.grandParent())
	} else {
		tree.insertCase4(node)
	}
}

func (tree *Tree) insertCase4(node *Node) {
	grandparent := node.grandParent()
	if node == node.Parent.Right && node.Parent == grandparent.Left {
		tree.rotateLeft(node.Parent)
		node = node.Left
	} else if node == node.Parent.Left && node.Parent == grandparent.Right {
		tree.rotateRight(node.Parent)
		node = node.Right
	}
	tree.insertCase5(node)
}

func (tree *Tree) insertCase5(node *Node) {
	node.Parent.color = black
	grandparent := node.grandParent()
	grandparent.color = red
	if node == node.Parent.Left && node.Parent == grandparent.Left {
		tree.rotateRight(grandparent)
	} else if node == node.Parent.Right && node.Parent == grandparent.Right {
		tree.rotateLeft(grandparent)
	}
}

// Rotates the tree
func (tree *Tree) rotateLeft(node *Node) {
	right := node.Right
	tree.replaceNode(node, right)
	node.Right = right.Left
	if right.Left != nil {
		right.Left.Parent = node
	}
	right.Left = node
	node.Parent = right
}

func (tree *Tree) rotateRight(node *Node) {
	left := node.Left
	tree.replaceNode(node, left)
	node.Left = left.Right
	if left.Right != nil {
		left.Right.Parent = node
	}
	left.Right = node
	node.Parent = left
}

func (tree *Tree) replaceNode(old *Node, new *Node) {
	if old.Parent == nil {
		tree.Root = new
	} else {
		if old == old.Parent.Left {
			old.Parent.Left = new
		} else {
			old.Parent.Right = new
		}
	}
	if new != nil {
		new.Parent = old.Parent
	}
}

// Clear removes all nodes from the Red-Black Tree.
func (tree *Tree) Clear() {
	tree.Root = nil
	tree.size = 0
}

// maximumNode finds and returns the maximum node in the subtree rooted at the given node.
func (node *Node) maximumNode() *Node {
	if node == nil {
		return nil
	}
	for node.Right != nil {
		node = node.Right
	}
	return node
}

// Delete removes a node with the given key from the Red-Black Tree
func (tree *Tree) Delete(key int) error {
	var child *Node
	node := tree.Search(key)
	if node == nil {
		return fmt.Errorf("node doesnt exist: %v", key)
	}
	if node.Left != nil && node.Right != nil {
		pred := node.Left.maximumNode()
		node.Key = pred.Key
		node = pred
	}
	if node.Left == nil || node.Right == nil {
		if node.Right == nil {
			child = node.Left
		} else {
			child = node.Right
		}
		if node.color == black {
			node.color = nodeColor(child)
			tree.deleteCase1(node)
		}
		tree.replaceNode(node, child)
		if node.Parent == nil && child != nil {
			child.color = black
		}
	}
	tree.size--
	return nil
}

func (tree *Tree) deleteCase1(node *Node) {
	if node.Parent == nil {
		return
	}
	tree.deleteCase2(node)
}

func (tree *Tree) deleteCase2(node *Node) {
	sibling := node.sibling()
	if nodeColor(sibling) == red {
		node.Parent.color = red
		sibling.color = black
		if node == node.Parent.Left {
			tree.rotateLeft(node.Parent)
		} else {
			tree.rotateRight(node.Parent)
		}
	}
	tree.deleteCase3(node)
}

func (tree *Tree) deleteCase3(node *Node) {
	sibling := node.sibling()
	if nodeColor(node.Parent) == black &&
		nodeColor(sibling) == black &&
		nodeColor(sibling.Left) == black &&
		nodeColor(sibling.Right) == black {
		sibling.color = red
		tree.deleteCase1(node.Parent)
	} else {
		tree.deleteCase4(node)
	}
}

func (tree *Tree) deleteCase4(node *Node) {
	sibling := node.sibling()
	if nodeColor(node.Parent) == red &&
		nodeColor(sibling) == black &&
		nodeColor(sibling.Left) == black &&
		nodeColor(sibling.Right) == black {
		sibling.color = red
		node.Parent.color = black
	} else {
		tree.deleteCase5(node)
	}
}

func (tree *Tree) deleteCase5(node *Node) {
	sibling := node.sibling()
	if node == node.Parent.Left &&
		nodeColor(sibling) == black &&
		nodeColor(sibling.Left) == red &&
		nodeColor(sibling.Right) == black {
		sibling.color = red
		sibling.Left.color = black
		tree.rotateRight(sibling)
	} else if node == node.Parent.Right &&
		nodeColor(sibling) == black &&
		nodeColor(sibling.Right) == red &&
		nodeColor(sibling.Left) == black {
		sibling.color = red
		sibling.Right.color = black
		tree.rotateLeft(sibling)
	}
	tree.deleteCase6(node)
}

func (tree *Tree) deleteCase6(node *Node) {
	sibling := node.sibling()
	sibling.color = nodeColor(node.Parent)
	node.Parent.color = black
	if node == node.Parent.Left && nodeColor(sibling.Right) == red {
		sibling.Right.color = black
		tree.rotateLeft(node.Parent)
	} else if nodeColor(sibling.Left) == red {
		sibling.Left.color = black
		tree.rotateRight(node.Parent)
	}
}
