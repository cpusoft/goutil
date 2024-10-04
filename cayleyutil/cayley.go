package cayleyutils

import (
	"github.com/cayleygraph/cayley"
	"github.com/cayleygraph/cayley/graph"
	"github.com/cayleygraph/quad"
	"log"
)

type CayleyStore struct {
	store *cayley.Handle
}

// Initialize the store with BadgerDB backend
func NewCayleyStore() *CayleyStore {
	// Create a brand new graph
	store, err := cayley.NewMemoryGraph()
	if err != nil {
		log.Fatalln(err)
	}

	return &CayleyStore{store: store}
}

// AddQuad adds a quad (relation) to the graph
func (cs *CayleyStore) AddQuad(subject, predicate, object, label string) {
	cs.store.AddQuad(quad.Make(subject, predicate, object, label))
}

// RemoveQuad removes a quad (relation) from the graph
func (cs *CayleyStore) RemoveQuad(subject, predicate, object, label string) {
	cs.store.RemoveQuad(quad.Make(subject, predicate, object, label))
}

// QuerySubject performs a query for quads with the given subject
func (cs *CayleyStore) QuerySubject(subject string) []quad.Quad {
	var results []quad.Quad
	it := cs.store.QuadIterator(quad.Subject, cs.store.ValueOf(quad.String(subject)))
	for it.Next(nil) {
		q := cs.store.Quad(it.Result())
		results = append(results, q)
	}
	return results
}

// BuildPath builds a query path starting from a specific subject and following a predicate
func (cs *CayleyStore) BuildPath(subject, predicate string) *cayley.Path {
	return cayley.StartPath(cs.store, quad.String(subject)).Out(quad.String(predicate))
}

// IteratePath iterates over a path and applies a function to each result
func (cs *CayleyStore) IteratePath(path *cayley.Path, fn func(quad.Value)) {
	path.Iterate(nil).EachValue(nil, func(v quad.Value) {
		fn(v)
	})
}

// ApplyTransaction applies a transaction for batch operations (add/remove quads)
func (cs *CayleyStore) ApplyTransaction(txn *graph.Transaction) {
	err := cs.store.ApplyTransaction(txn)
	if err != nil {
		log.Fatalf("Failed to apply transaction: %v", err)
	}
}

// Close closes the Cayley store
func (cs *CayleyStore) Close() {
	cs.store.Close()
}
