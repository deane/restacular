package restacular

import (
	"testing"
)

// Tests the trie independtly of the router

func createTrie() *node {
	tree := &node{path: "/"}

	// leading slash will have been removed by the router and trailing added
	tree.addPath("users/", true)
	tree.addPath("users/:id/", true)
	tree.addPath("users/:id/files/", true)
	tree.addPath("users/:id/friends/", true)
	tree.addPath("ideas/:id/", true)
	tree.addPath("images/:id/", true)

	return tree
}

// Kind of integration test, verify that routes are properly added to the trie
func TestAddingRoutes(t *testing.T) {
	tree := createTrie()

	// we should have only one child, a static one
	numberChildren := len(tree.staticChildren) + len(tree.wildcardChildren)
	if numberChildren > 2 {
		t.Log("\n" + tree.printTree("", ""))
		t.Errorf("Got more than 1 child node when adding routes, got %d", numberChildren)
	}

	// the node should have users as path
	child := tree.staticChildren[0]
	if child.path != "users/" {
		t.Log("\n" + tree.printTree("", ""))
		t.Errorf("Got %s as path instead of users/ for child node", child.path)
	}

	// the node should have a priority of 4 and its child should have 3
	if child.priority != 4 || child.wildcardChildren[0].priority != 3 {
		t.Log("\n" + tree.printTree("", ""))
		t.Errorf(
			"Got wrong priorities: got %d (expected 4) for child and %d (expected 3) for wildcard child",
			child.priority,
			child.wildcardChildren[0].priority,
		)
	}
}

func TestFindingRoutes(t *testing.T) {
	tree := createTrie()

	// Find a basic static one
	node, params := tree.find("/users/") // Router will add a /

	if node.path != "users/" {
		t.Log("\n" + tree.printTree("", ""))
		t.Errorf("Got %s as path instead of users when querying users", node.path)
	}

	// Find a wildcard path
	node, params = tree.find("/users/142/friends/")
	if node.path != "riends/" {
		t.Log("\n" + tree.printTree("", ""))
		t.Errorf("Got %s as path instead of riends when querying users", node.path)
	}
	if params["id"] != "142" {
		t.Log("\n" + tree.printTree("", ""))
		t.Log(params["id"])
		t.Errorf("Got %v as params but didn't get id=142", params)
	}
}
