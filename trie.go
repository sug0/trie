/*
Implementation of an R-Way Trie data structure.

A Trie has a root Node which is the base of the tree.
Each subsequent Node has a letter and children, which are
nodes that have letter values associated with them.
*/

package trie

import (
	"bufio"
	"log"
	"os"
	"sort"
)

type Node struct {
	val      rune
	term     bool
	mask     uint64
	parent   *Node
	children map[rune]*Node
}

type Trie struct {
	root *Node
	size int
}

const nul = 0x0

func newNode(parent *Node, val rune, m uint64, term bool) *Node {
	return &Node{
		val:      val,
		mask:     m,
		term:     term,
		parent:   parent,
		children: make(map[rune]*Node),
	}
}

// Creates and returns a pointer to a new child for the node.
func (n *Node) NewChild(parent *Node, r rune, bitmask uint64, val rune, term bool) *Node {
	node := newNode(parent, val, bitmask, term)
	n.children[r] = node
	return node
}

func (n *Node) RemoveChild(r rune) {
	delete(n.children, r)

	n.recalculateMask()
	for parent := n.Parent(); parent != nil; parent = parent.Parent() {
		parent.recalculateMask()
	}
}

func (n *Node) recalculateMask() {
	n.mask = maskrune(n.Val())
	for k, c := range n.Children() {
		n.mask |= (maskrune(k) | c.Mask())
	}
}

// Returns the parent of this node.
func (n Node) Parent() *Node {
	return n.parent
}

// Returns the children of this node.
func (n Node) Children() map[rune]*Node {
	return n.children
}

func (n Node) Val() rune {
	return n.val
}

// Returns a uint64 representing the current
// mask of this node.
func (n Node) Mask() uint64 {
	return n.mask
}

// Creates a new Trie with an initialized root Node.
func CreateTrie() *Trie {
	node := newNode(nil, 0, 0, false)
	return &Trie{
		root: node,
		size: 0,
	}
}

// Returns the root node for the Trie.
func (t *Trie) Root() *Node {
	return t.root
}

// Adds the key to the Trie.
func (t *Trie) Add(key string) int {
	t.size++
	runes := []rune(key)
	return t.addrune(t.Root(), runes, 0)
}

// Removes a key from the trie.
func (t *Trie) Remove(key string) {
	var (
		i    int
		rs   = []rune(key)
		node = t.nodeAtPath(key)
	)

	t.size--
	for n := node.Parent(); n != nil; n = n.Parent() {
		i++
		if len(n.Children()) > 1 {
			idx := len(rs) - i
			r := rs[idx]
			n.RemoveChild(r)
			break
		}
	}
}

// Reads words from a file and adds them to the
// trie. Expects words to be seperated by a newline.
func (t *Trie) AddFromFile(path string) {
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}

	reader := bufio.NewScanner(file)

	for reader.Scan() {
		t.Add(reader.Text())
	}

	if reader.Err() != nil {
		log.Fatal(err)
	}
}

// Returns all the keys currently stored in the trie.
func (t *Trie) Keys() []string {
	return t.PrefixSearch("")
}

// Performs a fuzzy search against the keys in the trie.
func (t Trie) FuzzySearch(pre string) []string {
	var (
		keys []string
		pm   []rune
	)

	fuzzycollect(t.Root(), pm, []rune(pre), &keys)
	sort.Strings(keys)
	return keys
}

// Performs a prefix search against the keys in the trie.
func (t Trie) PrefixSearch(pre string) []string {
	var keys []string

	node := t.nodeAtPath(pre)
	if node == nil {
		return keys
	}

	collect(node, []rune(pre), &keys)
	return keys
}

func (t Trie) nodeAtPath(pre string) *Node {
	runes := []rune(pre)
	return findNode(t.Root(), runes, 0)
}

func findNode(node *Node, runes []rune, d int) *Node {
	if node == nil {
		return nil
	}

	if len(runes) == 0 {
		return node
	}

	upper := len(runes)
	if d == upper {
		return node
	}

	n, ok := node.Children()[runes[d]]
	if !ok {
		return nil
	}

	d++
	return findNode(n, runes, d)
}

func (t Trie) addrune(node *Node, runes []rune, i int) int {
	if len(runes) == 0 {
		node.NewChild(node, 0, 0, nul, true)
		return i
	}

	r := runes[0]
	c := node.Children()

	n, ok := c[r]
	bitmask := maskruneslice(runes)
	if !ok {
		n = node.NewChild(node, r, bitmask, r, false)
	}
	n.mask |= bitmask

	i++
	return t.addrune(n, runes[1:], i)
}

func maskruneslice(rs []rune) uint64 {
	var m uint64
	for _, r := range rs {
		m |= maskrune(r)
	}

	return m
}

func maskrune(r rune) uint64 {
	i := uint64(1)
	return i << (uint64(r) - 97)
}

func collect(node *Node, pre []rune, keys *[]string) {
	children := node.Children()
	for r, n := range children {
		if n.term {
			*keys = append(*keys, string(pre))
			continue
		}

		npre := append(pre, r)
		collect(n, npre, keys)
	}
}

func fuzzycollect(node *Node, partialmatch, partial []rune, keys *[]string) {
	partiallen := len(partial)

	if partiallen == 0 {
		collect(node, partialmatch, keys)
		return
	}

	m := maskruneslice(partial)
	children := node.Children()
	for v, n := range children {
		xor := n.Mask() ^ m
		if (xor & m) != 0 {
			continue
		}

		npartial := partial
		if v == partial[0] {
			if partiallen > 1 {
				npartial = partial[1:]
			} else {
				npartial = partial[0:0]
			}
		}

		fuzzycollect(n, append(partialmatch, v), npartial, keys)
	}
}
