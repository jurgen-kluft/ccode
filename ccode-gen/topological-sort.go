package ccode_gen

import (
	"errors"
	"fmt"
)

type Vertex int

const (
	InvalidVertex = -1
)

type Edge struct {
	S Vertex // Source node
	D Vertex // Destination node
}

// Toposort performs a topological sort of the DAG defined by given edges.
//
// Takes a slice of Edge, where each element is a vertex pair representing an
// edge in the graph.  Each pair can also be considered a dependency
// relationship where Edge[0] must happen before Edge[1].
//
// To include a node that is not connected to the rest of the graph, include a
// node with one nil vertex.  It can appear anywhere in the sorted output.
//
// Returns an ordered list of vertices where each vertex occurs before any of
// its destination vertices.  An error is returned if a cycle is detected.
func Toposort(edges []Edge) ([]Vertex, error) {
	g := make(map[Vertex][]Vertex, len(edges)+1)
	for i := range edges {
		u, v := edges[i].S, edges[i].D
		if u == v {
			return nil, errors.New("nodes in edge cannot be the same")
		}
		if u == InvalidVertex {
			if _, ok := g[v]; !ok { // Add vertex only (empty destination list)
				g[v] = nil
			}
		} else if v == InvalidVertex {
			if _, ok := g[u]; !ok {
				g[u] = nil
			}
		} else {
			g[u] = append(g[u], v)
		}
	}

	sorted := make([]Vertex, 0, len(g))

	// Create map of vertices to incoming edge count, and set counts to 0
	inDegree := make(map[Vertex]Vertex, len(g))
	for n := range g {
		inDegree[n] = 0
	}

	for _, adjacent := range g { // For each vertex u, get adjacent list
		for _, v := range adjacent { // For each vertex v adjacent to u
			inDegree[v]++ // Increment inDegree[v]
		}
	}

	// Make a list next consisting of all vertices u such that inDegree[u] = 0
	var next []Vertex
	for u, deg := range inDegree {
		if deg == 0 {
			next = append(next, u)
		}
	}

	for len(next) > 0 { // While next is not empty...
		// Pop a vertex from next and call it vertex u
		u := next[len(next)-1]
		next = next[:len(next)-1]

		sorted = append(sorted, u) // Add u to the end sorted list

		// For each vertex v adjacent to sorted vertex u
		for _, v := range g[u] {
			inDegree[v]--         // Decrement count of incoming edges
			if inDegree[v] == 0 { // Enqueue nodes with no incoming edges
				next = append(next, v)
			}
		}
	}

	if len(sorted) < len(g) { // Check for cycles
		var cycleNodes []Vertex
		for u, deg := range inDegree {
			if deg != 0 {
				cycleNodes = append(cycleNodes, u)
			}
		}
		return nil, fmt.Errorf("graph contains cycle in nodes %q", cycleNodes)
	}

	return sorted, nil
}
