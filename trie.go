package restacular

import (
	"fmt"
	"strings"
)

type node struct {
	path string

	// how many leaves in all its children
	priority uint32

	// contains the first char of the childrens paths
	// index in that array match the index in the staticChildren array
	// and are ordered by their priority (highest first)
	Indices []byte

	// children nodes
	staticChildren []*node

	// wildcard nodes (those :param ones)
	wildcardChildren []*node

	// what are the actual handlers for that given path
	// those are set in the router directly rather carrying them around
	// while building the trie
	handlers map[string]HandlerFunc
}

// printTree is a util function to get a text representation of the trie for debugging/testing
func (n *node) printTree(indent string, nodeType string) string {
	line := fmt.Sprintf("%s %02d %s%s [%d]\n", indent, n.priority, nodeType, n.path, n.handlers)
	indent += "  "
	for _, node := range n.staticChildren {
		line += node.printTree(indent, "")
	}

	for _, node := range n.wildcardChildren {
		line += node.printTree(indent, ":")
	}

	return line
}

func (n *node) setHandler(method string, handler HandlerFunc) {
	if n.handlers == nil {
		n.handlers = make(map[string]HandlerFunc)
	}

	n.handlers[method] = handler
}

// sortChildren sorts from the given index to the front of the static children array
// in order to keep childre with the highest priority at the front
func (n *node) sortChildren(i int) {
	for i > 0 && n.staticChildren[i].priority > n.staticChildren[i-1].priority {
		n.staticChildren[i], n.staticChildren[i-1] = n.staticChildren[i-1], n.staticChildren[i]
		n.Indices[i], n.Indices[i-1] = n.Indices[i-1], n.Indices[i]
		i -= 1
	}
}

func (n *node) addWildcardNode(token string) *node {
	child := &node{
		path: token,
	}

	if n.wildcardChildren == nil {
		n.wildcardChildren = []*node{child}
	} else {
		n.wildcardChildren = append(n.wildcardChildren, child)
	}

	return child
}

func (n *node) addStaticNode(token string) *node {
	child := &node{
		path: token,
	}

	// append or create the first char array
	if n.Indices == nil {
		n.Indices = []byte{token[0]}
		n.staticChildren = []*node{child}
	} else {
		n.Indices = append(n.Indices, token[0])
		// and add it to static children (of men)
		n.staticChildren = append(n.staticChildren, child)
	}

	return child
}

// findCommonStaticChild looks up in the node static children to see
// whether one of them matches the new token.
// If so returns that node, its index and how many char they have in common
func (n *node) findCommonStaticChild(token string) (*node, int, int) {
	var commonUntil int
	var commonChild *node
	var indexChild int

	for i, char := range n.Indices {
		if char == token[0] {
			commonChild = n.staticChildren[i]
			// we want to know how many chars they have in common
			for commonUntil = range commonChild.path {
				if commonUntil == len(token) || token[commonUntil] != commonChild.path[commonUntil] {
					indexChild = i
					break
				}
			}
		}
	}

	return commonChild, commonUntil, indexChild
}

// addPath builds the trie and return the node for the path we just added
// The addPriority is a bit hackish way to ensure we don't double the priority
// when adding static nodes since we add priority right away in that case to
// reorder the list
func (n *node) addPath(path string, addPriority bool) *node {
	// Always increment the priority of the node we're going through
	if addPriority {
		n.priority++
	}

	// if we reached the end of the path, return the current node
	if len(path) == 0 {
		return n
	}

	isWilcard := path[0] == ':'
	nextSlash := strings.Index(path, "/")
	// what we are actually going to look at in that iteration
	token := path[0 : nextSlash+1]
	// what we will look at in the next iteration
	remainingPath := path[nextSlash+1:]

	// For now, don't bother doing a char by char trie for wildcards
	// It should almost never happen in practice anyway to have 2 different wilcards
	// from the same parent so doing by the whole token is fine
	// Not optimized on purpose
	if isWilcard {
		var child *node
		token = token[1:]

		// check if we already have it
		for _, wildcardNode := range n.wildcardChildren {
			if token == wildcardNode.path {
				child = wildcardNode
			}
		}

		// New wildcard node, create a node object and append it to that current node
		if child == nil {
			child = n.addWildcardNode(token)
		}

		return child.addPath(remainingPath, true)
	}

	// We got a normal string !
	// 2 things can happen here
	commonChild, commonUntil, indexChild := n.findCommonStaticChild(token)

	// 1 - some child nodes start with the same char as the current path
	// in that case we want to find the common prefix between both of them
	// and put them as child node of that common one
	if commonChild != nil {
		// 2 cases there as well
		// Either the path is fully the same and we can just continue our merry trip
		// TODO: check in which case we get commonUntil == 0, if we have only one char?
		if commonUntil == 0 || commonUntil == len(token)-1 {
			commonChild.priority++
			n.sortChildren(indexChild)
			return commonChild.addPath(path[commonUntil+1:], false)
		}

		// Or it's different and we need to do a NITM (Node In The Middle, I know...)
		commonPath := token[0:commonUntil]
		commonChild.path = commonChild.path[commonUntil:]

		middleNode := &node{
			path:           commonPath,
			priority:       commonChild.priority,
			staticChildren: []*node{commonChild},
			Indices:        []byte{commonChild.path[0]},
		}
		n.staticChildren[indexChild] = middleNode
		n.sortChildren(indexChild)
		return middleNode.addPath(path[commonUntil:], true)
	}

	// 2 - no common prefix with existing child so just append it
	child := n.addStaticNode(token)
	return child.addPath(remainingPath, true)
}

func findInStatic(n *node, path string) *node {
	for i, char := range n.Indices {
		if char == path[0] {
			child := n.staticChildren[i]
			if len(path) >= len(child.path) && child.path == path[:len(child.path)] {
				return child
			}
		}
	}

	return nil
}

func (n *node) find(path string) (*node, map[string]string) {
	var params map[string]string

	// gofmt is a bit weird here with the indentation
	for len(path) >= len(n.path) {
		path = path[len(n.path):]
		//fmt.Printf("Path: %s, node path: %s\n", path, n.path)

		if len(path) > 0 {
			child := findInStatic(n, path)

			if child != nil {
				//fmt.Printf("Continuing after finding static %v\n", child)
				n = child
			} else {
				// no luck in the static? check wildcard children
				for _, wildcardChildren := range n.wildcardChildren {
					nextSlash := strings.Index(path, "/")

					// check whether we have something after that path
					if len(path[nextSlash+1:]) == 0 {
						if params == nil {
							params = make(map[string]string)
						}
						params[wildcardChildren.path[:len(wildcardChildren.path)-1]] = path[:nextSlash]
						return wildcardChildren, params
					}
					// check if next token matches
					//fmt.Printf("Next token is %s\n", path[strings.Index(path, "/")+1:])
					child := findInStatic(wildcardChildren, path[nextSlash+1:])
					if child != nil {
						//fmt.Printf("Continuing after finding wildcard %v\n", wildcardChildren)
						n = wildcardChildren
						if params == nil {
							params = make(map[string]string)
						}
						params[wildcardChildren.path[:len(wildcardChildren.path)-1]] = path[:nextSlash]
						// Very stupid hack due to adding / everywhere
						path = path[1:]
					}
				}
			}

		} else {
			return n, params
		}

	}

	fmt.Printf("404: path: %s, node path: %s\n", path, n.path)
	// Ain't got nothing
	// TODO: 404 handling
	return n, params
}
