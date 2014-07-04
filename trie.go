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
	isWildcard    bool
	wildcardChild *node

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

	if n.wildcardChild != nil {
		line += n.wildcardChild.printTree(indent, ":")
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
	// TODO: remove need for this hack
	// Always increment the priority of the node we're going through
	if addPriority {
		n.priority++
	}

	// if we reached the end of the path, return the current node
	if len(path) == 0 {
		return n
	}

	firstChar := path[0]
	// what we are actually going to look at in that iteration
	var token string
	// what we will look at in the next iteration
	var remainingPath string

	// If the first char is a /, we want to know the next one
	nextSlash := strings.Index(path, "/")

	if firstChar == '/' {
		token = "/"
		remainingPath = path[1:]
	} else if nextSlash != -1 {
		token = path[0:nextSlash]
		remainingPath = path[nextSlash:]
	} else {
		token = path
		// No need for remaining path if we're at a leaf node
	}

	//fmt.Printf("%s - %s - %s\n", path, token, remainingPath)

	// Wildchild path
	if firstChar == ':' {
		var child *node
		token = token[1:]

		// check if we already have it
		if n.wildcardChild != nil {
			if token == n.wildcardChild.path {
				child = n.wildcardChild
			} else {
				if n.isWildcard {
					panic("Can't have 2 wildcard nodes at the same level")

				}
			}
		}

		// New wildcard node, create a node object and assign it to the current node
		if child == nil {
			n.wildcardChild = &node{
				path:       token,
				isWildcard: true,
			}
			child = n.wildcardChild
		}

		return child.addPath(remainingPath, true)
	}

	// We got a normal string !
	// 2 things can happen here
	commonChild, commonUntil, indexChild := n.findCommonStaticChild(token)
	//fmt.Printf("Common child: %v, common until: %d, node path:%s\n\n", commonChild, commonUntil, n.path)

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

func (n *node) find(path string) (*node, map[string]string) {
	var params map[string]string

FIND:
	for len(path) >= len(n.path) {
		if n.isWildcard == false {
			path = path[len(n.path):]
		}

		// second part handles trailing slash
		if len(path) == 0 || (len(path) == 1 && path[0] == '/') {
			return n, params
		}

		if n.wildcardChild == nil {
			c := path[0]
			for i, index := range n.Indices {
				if c == index {
					n = n.staticChildren[i]
					continue FIND
				}
			}

			// TODO: handle 404
			return nil, params
		}

		// no luck in the static? check wildcard child
		// Faster than strings.Index
		nextSlash := 0
		for nextSlash < len(path) && path[nextSlash] != '/' {
			nextSlash++
		}

		nextToken := path[nextSlash:]
		if params == nil {
			params = map[string]string{
				n.wildcardChild.path: path[:nextSlash],
			}
		} else {
			params[n.wildcardChild.path] = path[:nextSlash]
		}

		// Was it the end of the path?
		if len(nextToken) == 0 {
			return n.wildcardChild, params
		}

		// So we have something after the param, must be a static
		c := nextToken[0]
		for i, index := range n.wildcardChild.Indices {
			if c == index {
				path = path[nextSlash:]
				n = n.wildcardChild.staticChildren[i]
				continue FIND
			}
		}

		// TODO: handle 404
		return nil, params
	}

	fmt.Printf("404: path: %s, node path: %s\n", path, n.path)
	// Ain't got nothing
	// TODO: 404 handling
	return nil, params
}
