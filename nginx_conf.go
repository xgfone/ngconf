// Copyright 2019 xgfone
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ngconf

import (
	"fmt"
	"io"
	"sort"
	"strings"
)

// Pre-defines some errors.
var (
	ErrRootDirective      = fmt.Errorf("root node must have the directive")
	ErrNonRootNoDirective = fmt.Errorf("non-root node have no directive")
)

type nodes []*Node

func (ns nodes) Len() int           { return len(ns) }
func (ns nodes) Swap(i, j int)      { ns[i], ns[j] = ns[j], ns[i] }
func (ns nodes) Less(i, j int) bool { return ns[j] == nil }

type nodeStack struct {
	nodes []*Node
}

func (ns *nodeStack) Push(node *Node) { ns.nodes = append(ns.nodes, node) }
func (ns *nodeStack) Pop() *Node {
	_len := len(ns.nodes) - 1
	node := ns.nodes[_len]
	ns.nodes = ns.nodes[:_len]
	return node
}

// Node represents a set of the key-value configurations of Nginx.
type Node struct {
	Root      bool
	Directive string
	Args      []string
	Children  []*Node
}

func newNode(stmts []string, root ...bool) (*Node, error) {
	var isRoot bool
	if len(root) > 0 && root[0] {
		isRoot = true
	}

	var directive string
	var args []string
	switch len(stmts) {
	case 0:

	case 1:
		directive = stmts[0]
	default:
		directive = stmts[0]
		args = stmts[1:]
	}

	if directive == "" && !isRoot {
		return nil, ErrNonRootNoDirective
	}
	if directive != "" && isRoot {
		return nil, ErrRootDirective
	}

	return &Node{Directive: directive, Args: args, Root: isRoot}, nil
}

func (n *Node) appendChild(node *Node) { n.Children = append(n.Children, node) }

// String is equal to n.Dump(0).
func (n *Node) String() string {
	return n.Dump(0)
}

// WriteTo implements the interface WriterTo to write the configuration to file.
func (n *Node) WriteTo(w io.Writer) (int64, error) {
	m, err := io.WriteString(w, n.String())
	return int64(m), err
}

func (n *Node) getChildren(indent int, ctx *nodeDumpCtx) string {
	ss := make([]string, len(n.Children))
	for i, node := range n.Children {
		if i == 0 {
			ctx.LastBlockEnd = false
		} else {
			ctx.BlockStart = false
		}
		ss[i] = node.dump(indent, ctx, n)
	}
	return strings.Join(ss, "\n")
}

// Dump converts the Node to string.
func (n *Node) Dump(indent int) string {
	return n.dump(indent, &nodeDumpCtx{FirstBlock: true}, n)
}

type nodeDumpCtx struct {
	HasComment   bool
	FirstBlock   bool
	BlockStart   bool
	LastBlockEnd bool
}

func (n *Node) dump(indent int, ctx *nodeDumpCtx, parent *Node) string {
	var prefix, spaces string
	for i := indent; i > 0; i-- {
		spaces += "    "
	}

	lastComment := ctx.HasComment
	ctx.HasComment = strings.HasPrefix(n.Directive, "#")
	if ctx.HasComment {
		if !lastComment {
			if !ctx.BlockStart {
				prefix = "\n"
			}
		} else if !parent.Root && ctx.LastBlockEnd {
			prefix = "\n"
		}
	} else if ctx.LastBlockEnd {
		prefix = "\n"
	}

	if len(n.Children) == 0 {
		if ctx.HasComment {
			if len(n.Args) == 0 {
				return fmt.Sprintf("%s%s%s", prefix, spaces, n.Directive)
			}
			return fmt.Sprintf("%s%s%s %s", prefix, spaces, n.Directive, strings.Join(n.Args, " "))
		}

		if len(n.Args) == 0 {
			return fmt.Sprintf("%s%s%s;", prefix, spaces, n.Directive)
		}
		return fmt.Sprintf("%s%s%s %s;", prefix, spaces, n.Directive, strings.Join(n.Args, " "))
	} else if n.Root {
		return n.getChildren(indent, ctx)
	} else if args := strings.Join(n.Args, " "); args != "" {
		if ctx.FirstBlock {
			ctx.FirstBlock = false
			if prefix == "" {
				prefix = "\n"
			}
		}
		ctx.BlockStart = true
		s := fmt.Sprintf("%s%s%s %s {\n%s\n%s}", prefix, spaces, n.Directive, args,
			n.getChildren(indent+1, ctx), spaces)
		ctx.LastBlockEnd = true
		return s
	} else {
		if ctx.FirstBlock {
			ctx.FirstBlock = false
			if prefix == "" {
				prefix = "\n"
			}
		}
		ctx.BlockStart = true
		s := fmt.Sprintf("%s%s%s {\n%s\n%s}", prefix, spaces, n.Directive,
			n.getChildren(indent+1, ctx), spaces)
		ctx.LastBlockEnd = true
		return s
	}
}

// Get returns the child node by the given directive with the args.
func (n *Node) Get(directive string, args ...string) []*Node {
	results := make([]*Node, 0, 1)
	for _, child := range n.Children {
		if child.Directive == directive {
			results = append(results, child)
		}
	}

	_len := len(results)
	if _len == 0 {
		return nil
	}

	if argslen := len(args); argslen > 0 {
		var count int
		for i, node := range results {
			if len(node.Args) < argslen {
				results[i] = nil
				count++
			}
		}
		if count == _len {
			return nil
		} else if count > 0 {
			sort.Sort(nodes(results))
			results = results[:_len-count]
		}

		oldResults := results
		results = make([]*Node, 0, len(oldResults))
		for _, node := range oldResults {
			ok := true
			for i, arg := range args {
				if arg != node.Args[i] {
					ok = false
					break
				}
			}
			if ok {
				results = append(results, node)
			}
		}
	}

	return results
}

// Add adds and returns the child node with the directive and the args.
//
// If the child node has existed, it will be ignored and return the first old.
func (n *Node) Add(directive string, args ...string) *Node {
	if nodes := n.Get(directive, args...); len(nodes) > 0 {
		return nodes[0]
	}

	node := &Node{Directive: directive, Args: args}
	n.appendChild(node)
	return node
}

// Del deletes the child node by the directive with the args.
//
// If args is nil, it will delete all the child nodes.
func (n *Node) Del(directive string, args ...string) {
	_len := len(n.Children)
	if _len == 0 {
		return
	}

	var count int
	for i, child := range n.Children {
		if child.Directive == directive {
			if _len := len(args); _len == 0 {
				n.Children[i] = nil
				count++
			} else if _len <= len(child.Args) {
				ok := true
				for j, arg := range args {
					if arg != child.Args[j] {
						ok = false
						break
					}
				}
				if ok {
					n.Children[i] = nil
					count++
				}
			}
		}
	}

	if count > 0 {
		sort.Stable(nodes(n.Children))
		n.Children = n.Children[:_len-count]
	}
}

// Decode decodes the string s to Node.
func Decode(s string) (*Node, error) {
	var err error
	var node *Node
	var isComment bool

	stack := &nodeStack{}
	currentWord := []rune{}
	currentStmt := []string{}
	currentBlock, _ := newNode(nil, true)

	for _, char := range s {
		if isComment {
			switch char {
			case '\n':
				currentStmt = append(currentStmt, string(currentWord))
				if node, err = newNode(currentStmt); err != nil {
					return nil, err
				}

				isComment = false
				currentWord = nil
				currentStmt = nil
				currentBlock.appendChild(node)
			case ' ', '\t':
				// End the current word.
				currentStmt = append(currentStmt, string(currentWord))
				currentWord = nil
			default:
				currentWord = append(currentWord, char)
			}

			continue
		}

		switch char {
		case '{':
			// Put the current block on the stack, start a new block.
			// Also, if we are in a word, "finish" that off, and end
			// the current statement.
			stack.Push(currentBlock)
			if len(currentWord) > 0 {
				currentStmt = append(currentStmt, string(currentWord))
				currentWord = nil
			}
			if currentBlock, err = newNode(currentStmt); err != nil {
				return nil, err
			}
			currentStmt = nil
		case '}':
			// Finalize the current block, pull the previous (outer) block off
			// of the stack, and add the inner block to the previous block's
			// map of blocks.
			innerBlock := currentBlock
			currentBlock = stack.Pop()
			currentBlock.appendChild(innerBlock)
		case ';':
			// End the current word and statement.
			currentStmt = append(currentStmt, string(currentWord))
			currentWord = nil

			if len(currentStmt) > 0 {
				if node, err = newNode(currentStmt); err != nil {
					return nil, err
				}
				currentBlock.appendChild(node)
			}
			currentStmt = nil
		case '\n', ' ', '\t':
			// End the current word.
			if len(currentWord) > 0 {
				currentStmt = append(currentStmt, string(currentWord))
				currentWord = nil
			}
		case '#':
			isComment = true
			currentWord = append(currentWord, char)
		default:
			// Add current character onto the current word.
			currentWord = append(currentWord, char)
		}
	}

	// Support the last line is the comment
	if len(currentWord) > 0 {
		currentStmt = append(currentStmt, string(currentWord))
	}
	if len(currentStmt) > 0 {
		if node, err = newNode(currentStmt); err == nil {
			currentBlock.appendChild(node)
		}
	}

	return currentBlock, nil
}
