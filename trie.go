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
	indices []byte

	// children nodes
	staticChildren []*node

	// wildcard nodes (those :param ones)
	wildcardChild *node

	// Is the current node a wildcard one?
	isWildcard bool

	// what are the actual handlers for that given path
	// those are set in the router directly rather carrying them around
	// while building the trie
	handlers map[string]HandlerFunc
}

// printTree is a util function to get a text representation of the trie for debugging/testing
func (n *node) printTree(indent string, nodeType string) string {
	line := fmt.Sprintf("%s %02d %s%s [%v]\n", indent, n.priority, nodeType, n.path, n.handlers)
	indent += "  "

	for _, node := range n.staticChildren {
		line += node.printTree(indent, "")
	}

	if n.wildcardChild != nil {
		line += n.wildcardChild.printTree(indent, ":")
	}

	return line
}

// setHandler is used in the router to assign the HTTP handlers to a leaf
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
		n.indices[i], n.indices[i-1] = n.indices[i-1], n.indices[i]
		i -= 1
	}
}

// addStaticNode takes a token and create a child node for the current node containg it
func (n *node) addStaticNode(token string) *node {
	child := &node{
		path: token,
	}

	// append or create the first char array
	if n.indices == nil {
		n.indices = []byte{token[0]}
		n.staticChildren = []*node{child}
	} else {
		n.indices = append(n.indices, token[0])
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

	for i, char := range n.indices {
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

// addPath builds the trie and return the leaf node for the path we just added
func (n *node) addPath(path string) *node {
	n.priority++

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

	// Wildchild path
	if firstChar == ':' {
		var child *node
		token = token[1:]

		// check if we already have it
		if n.wildcardChild != nil {
			if token == n.wildcardChild.path {
				child = n.wildcardChild
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

		return child.addPath(remainingPath)
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
		if commonUntil == 0 || commonUntil == len(token)-1 {
			// There's a bit of a hack here: we want to reorder the current node children and take into account that
			// the common child will have +1 prio so we temporarily increments his prio
			commonChild.priority++
			n.sortChildren(indexChild)
			// And put it back down 1 since it's going to get incremented by the addPath method immediately
			commonChild.priority--
			return commonChild.addPath(path[commonUntil+1:])
		}

		// Or it's different and we need to do a NITM (Node In The Middle, I know...)
		commonPath := token[0:commonUntil]
		commonChild.path = commonChild.path[commonUntil:]

		middleNode := &node{
			path:           commonPath,
			priority:       commonChild.priority,
			staticChildren: []*node{commonChild},
			indices:        []byte{commonChild.path[0]},
		}
		n.staticChildren[indexChild] = middleNode
		n.sortChildren(indexChild)
		return middleNode.addPath(path[commonUntil:])
	}

	// 2 - no common prefix with existing child so just append it
	child := n.addStaticNode(token)
	return child.addPath(remainingPath)
}

func (n *node) find(path string) (*node, Params) {
	var params Params

FIND:
	for len(path) >= len(n.path) {
		// for static nodes, we can just get the next path using the length of the path itself
		if n.isWildcard == false {
			path = path[len(n.path):]
		}

		// second part handles trailing slash
		if len(path) == 0 || (len(path) == 1 && path[0] == '/') {
			return n, params
		}

		if n.wildcardChild == nil {
			c := path[0]
			for i, index := range n.indices {
				if c == index {
					n = n.staticChildren[i]
					continue FIND
				}
			}

			// 404
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
			// 2 params sounds about right for an API
			// Small performance loss if have you > 2 params in the same URL
			params = make(Params, 0, 2)
		}
		params = append(params, Param{
			Name:  n.wildcardChild.path,
			Value: path[:nextSlash],
		})

		// Was it the end of the path?
		if len(nextToken) == 0 {
			return n.wildcardChild, params
		}

		// So we have something after the param, must be a static
		c := nextToken[0]
		for i, index := range n.wildcardChild.indices {
			if c == index {
				// We need to get the next token but can't use the length of a wilcard path obviously
				path = path[nextSlash:]
				n = n.wildcardChild.staticChildren[i]
				continue FIND
			}
		}

		// 404
		return nil, params
	}

	// 404
	fmt.Printf("Should never get there: path: %s, node path: %s\n", path, n.path)
	return nil, params
}
