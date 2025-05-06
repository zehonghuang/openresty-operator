package pathtrie

import (
	"strings"
	"sync"
)

var PrefixTrie = newNamespaceTrie()

type Trie struct {
	root *node
	lock sync.RWMutex
}

type node struct {
	children map[string]*node
	terminal bool
	owner    string
}

func NewTrie() *Trie {
	return &Trie{
		root: &node{
			children: make(map[string]*node),
		},
	}
}

func (t *Trie) Insert(path, owner string) bool {
	t.lock.Lock()
	defer t.lock.Unlock()

	parts := cleanPath(path)
	n := t.root
	for i, part := range parts {
		if _, ok := n.children[part]; !ok {
			n.children[part] = &node{children: make(map[string]*node)}
		}
		n = n.children[part]
		if n.terminal && n.owner != owner {
			return true
		}
		if i == len(parts)-1 && len(n.children) > 0 {
			if n.terminal && n.owner != owner {
				return true
			}
			for _, child := range n.children {
				if child.terminal && child.owner != owner {
					return true
				}
			}
		}
	}

	n.terminal = true
	n.owner = owner
	return false
}

func cleanPath(path string) []string {
	path = strings.TrimPrefix(path, "/")
	path = strings.TrimSuffix(path, "/")
	if path == "" {
		return []string{}
	}
	return strings.Split(path, "/")
}

func (t *Trie) Clear() {
	t.lock.Lock()
	defer t.lock.Unlock()
	t.root = &node{children: make(map[string]*node)}
}

type NamespaceTrie struct {
	namespaces map[string]*Trie
	lock       sync.RWMutex
}

func newNamespaceTrie() *NamespaceTrie {
	return &NamespaceTrie{
		namespaces: make(map[string]*Trie),
	}
}

func (nt *NamespaceTrie) getOrCreateTrie(namespace string) *Trie {
	nt.lock.Lock()
	defer nt.lock.Unlock()
	if trie, ok := nt.namespaces[namespace]; ok {
		return trie
	}
	trie := NewTrie()
	nt.namespaces[namespace] = trie
	return trie
}

func (nt *NamespaceTrie) Insert(namespace, path, owner string) bool {
	trie := nt.getOrCreateTrie(namespace)
	return trie.Insert(path, owner)
}

func (nt *NamespaceTrie) ClearNamespace(namespace string) {
	nt.lock.Lock()
	defer nt.lock.Unlock()
	delete(nt.namespaces, namespace)
}

func (nt *NamespaceTrie) ClearAll() {
	nt.lock.Lock()
	defer nt.lock.Unlock()
	nt.namespaces = make(map[string]*Trie)
}
