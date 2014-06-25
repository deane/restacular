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

	// Always add a / to the end of a path if missing to simplify routing
	if !strings.HasSuffix(path, "/") {
		path += "/"
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
		if commonUntil == len(token)-1 {
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

// TODO: rewrite that method to use a for loop so we don't have to pass a pointer to a map
// TODO check for handlers,  test case sensitivity
func (n *node) find(path string, params *map[string]string) *node {
	pathLen := len(path)

	// end of the path, do we have a handler?
	if pathLen == 0 {
		return n
	}

	// Always add a / to the end of a path if missing to simplify routing
	if !strings.HasSuffix(path, "/") {
		path += "/"
		pathLen += 1
	}

	// First we try to find a match in the static children
	for i, char := range n.Indices {
		//fmt.Printf("%c - %c\n", char, path[0])
		//fmt.Printf("Node: %v\n", n)
		if char == path[0] {
			child := n.staticChildren[i]
			childPathLen := len(child.path)
			//fmt.Printf("Lengths: %d - %d, paths: %s - %s\n", pathLen, childPathLen, path[:childPathLen], child.path)
			// Compare numbers before comparing strings
			if pathLen >= childPathLen && child.path == path[:childPathLen] {
				remainingPath := path[childPathLen:]
				//fmt.Printf("Remaining: %s\n", remainingPath)
				return child.find(remainingPath, params)
			}
			break
		}
	}

	// Static path were not enough? Introducing wildcard path
	if len(n.wildcardChildren) > 0 {
		// Bench that against a basic loop iterating over the chars
		nextSlash := strings.Index(path, "/")
		token := path[:nextSlash]
		nextToken := path[nextSlash+1:]

		//fmt.Printf("Token: %s, next token: %s\n", token, nextToken)

		// So we got a token, but it was empty, that will be a 404
		if len(token) == 0 {
			return nil
		}

		for _, wildcardChildren := range n.wildcardChildren {
			found := wildcardChildren.find(nextToken, params)

			if found == nil {
				return nil
			}

			// Eh ! Caught one
			// TODO: optimize that part
			wildPathLen := len(wildcardChildren.path)
			var param string
			if wildcardChildren.path[wildPathLen-1] == '/' {
				param = wildcardChildren.path[:wildPathLen-1]
			} else {
				param = wildcardChildren.path
			}

			if *params == nil {
				*params = map[string]string{param: token}
			} else {
				(*params)[param] = token
			}
			return found
		}
	}

	return nil
}
