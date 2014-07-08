package restacular

import (
	"reflect"
	"testing"
)

// Tests the trie independtly of the router

func getPaths() []string {
	// leading and trailing slash will have been removed by the router
	return []string{
		"users",
		"users/:id",
		"users/:id/files",
		"users/:id/friends",
		"ideas/:id",
		"images/:id",
		"images/:id/similar/:similarId",
		"images/:id/similar/:similarId/comments/:commentId",
		"users/:id/filesystem",
		"users/:id/filet",
	}

}

func createTrie() *node {
	tree := &node{path: "/"}

	for _, path := range getPaths() {
		tree.addPath(path)
	}

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

	// the node should have a priority of 6
	if child.priority != 6 {
		t.Log("\n" + tree.printTree("", ""))
		t.Errorf("Got wrong priorities: got %d (expected 6) for child", child.priority)
	}
}

type testPaths struct {
	path        string
	shouldBeNil bool // True if checking for 404
	params      Params
}

func testMultiplePaths(t *testing.T, root *node, pathsToTest []testPaths) {
	for _, pathToTest := range pathsToTest {
		node, params := root.find(pathToTest.path)

		if node == nil && !pathToTest.shouldBeNil {
			t.Errorf("Didn't find a handler for %s", pathToTest.path)
		}
		if !pathToTest.shouldBeNil && !reflect.DeepEqual(params, pathToTest.params) {
			t.Errorf("Params mismatch for route '%s': %v (expected) - %v (found)", pathToTest.path, pathToTest.params, params)
		}
	}
}

func TestFindingRoutes(t *testing.T) {
	tree := createTrie()

	testMultiplePaths(t, tree, []testPaths{
		{"/users", false, nil},
		{"/users/42", false, Params{Param{"id", "42"}}},
		{"/users/42/files", false, Params{Param{"id", "42"}}},
		{"/users/42/friends", false, Params{Param{"id", "42"}}},
		{"/ideas/21", false, Params{Param{"id", "21"}}},
		{"/images/2", false, Params{Param{"id", "2"}}},
		{"/images/2/similar/12", false, Params{Param{"id", "2"}, Param{"similarId", "12"}}},
		{"/images/2/similar/12/comments/1234", false, Params{Param{"id", "2"}, Param{"similarId", "12"}, Param{"commentId", "1234"}}},
		{"/users/21/filesystem", false, Params{Param{"id", "21"}}},
		{"/users/21/filet", false, Params{Param{"id", "21"}}},
		{"/users/123/something", true, nil},
		{"/hello/kitty", true, nil},
	})
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

// This will be a bit slower than the others
func BenchmarkGettingPathWithThreeParam(b *testing.B) {
	tree := createTrie()
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		tree.find("/images/1/similar/10/comments/120")
	}
}
