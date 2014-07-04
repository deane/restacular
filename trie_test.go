package restacular

import (
	"testing"
)

// Tests the trie independtly of the router

func createTrie() *node {
	tree := &node{path: "/"}

	// leading and trailing slash will have been removed by the router
	tree.addPath("users", true)
	tree.addPath("users/:id", true)
	tree.addPath("users/:id/files", true)
	tree.addPath("users/:id/friends", true)
	tree.addPath("ideas/:id", true)
	tree.addPath("images/:id", true)
	tree.addPath("images/:id/similar/:similarId", true)
	tree.addPath("images/:id/similar/:similarId/comments/:commentId", true)

	return tree
}

// Kind of integration test, verify that routes are properly added to the trie
func TestAddingRoutes(t *testing.T) {
	tree := createTrie()

	// we should have only one child, a static one
	numberChildren := len(tree.staticChildren)
	if tree.wildcardChild != nil {
		numberChildren += 1
	}
	if numberChildren > 2 {
		t.Log("\n" + tree.printTree("", ""))
		t.Errorf("Got more than 1 child node when adding routes, got %d", numberChildren)
	}

	// the node should have users as path
	child := tree.staticChildren[0]
	if child.path != "users" {
		t.Log("\n" + tree.printTree("", ""))
		t.Errorf("Got %s as path instead of users for child node", child.path)
	}

	// the node should have a priority of 4 and its child should have 3
	if child.priority != 4 {
		t.Log("\n" + tree.printTree("", ""))
		t.Errorf("Got wrong priorities: got %d (expected 4) for child", child.priority)
	}
}

func TestFindingRoutes(t *testing.T) {
	tree := createTrie()

	// Find a basic static one
	node, params := tree.find("/users") // Router will add a /

	if node.path != "users" {
		t.Log("\n" + tree.printTree("", ""))
		t.Errorf("Got %s as path instead of users when querying users", node.path)
	}

	// Find a wildcard path
	node, params = tree.find("/users/142/friends")
	if node.path != "riends" {
		t.Log("\n" + tree.printTree("", ""))
		t.Errorf("Got %s as path instead of riends when querying users", node.path)
	}

	if params.Get("id") != "142" {
		t.Log("\n" + tree.printTree("", ""))
		t.Errorf("Got %v as params but didn't get id=142", params)
	}

	// Try the ones from the benchmarks to make sure we don't benchmark 404
	node, params = tree.find("/images/1")
	if node == nil || params.Get("id") != "1" {
		t.Log("\n" + tree.printTree("", ""))
		t.Errorf("Could not find /images/1 in the trie or the params was not set properly")
	}

	node, params = tree.find("/images/1/similar/10")
	if node == nil {
		t.Log("\n" + tree.printTree("", ""))
		t.Errorf("Could not find /images/1/similar/10 in the trie")
	}

	node, params = tree.find("/hello/kitty")
	if node != nil {
		t.Log("\n" + tree.printTree("", ""))
		t.Errorf("Should not have found a node")
	}
}

func BenchmarkGettingPathWithoutParam(b *testing.B) {
	tree := createTrie()
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		tree.find("/users")
	}
}

func BenchmarkGettingPathWithOneParam(b *testing.B) {
	tree := createTrie()
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		tree.find("/images/1")
	}
}

func BenchmarkGettingPathWithTwoParam(b *testing.B) {
	tree := createTrie()
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		tree.find("/images/1/similar/10")
	}
}

func BenchmarkGettingPathWithThreeParam(b *testing.B) {
	tree := createTrie()
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		tree.find("/images/1/similar/10/comments/120")
	}
}
