package cayleyutils

import (
	"github.com/cayleygraph/cayley"
	"github.com/cayleygraph/cayley/graph"
	"github.com/cayleygraph/quad"

	_ "github.com/cayleygraph/cayley/graph/kv/bolt"

	"log"
)

// 需要加上读写锁
type CayleyStore struct {
	store *cayley.Handle
}

// Initialize the store with BadgerDB backend

func NewCayleyStore(useMem bool, path string) *CayleyStore {

	if !useMem {
		// Initialize the database
		err := graph.InitQuadStore("bolt", path, nil)
		if err != nil && err != graph.ErrDatabaseExists {
			log.Fatalln(err)
			return nil
		}

		// Open and use the database
		store, err := cayley.NewGraph("bolt", path, nil)
		return &CayleyStore{store: store}
	}

	//// Create a brand new graph
	store, err := cayley.NewMemoryGraph()
	if err != nil {
		log.Fatalln(err)
	}

	return &CayleyStore{store: store}
}

// AddQuad adds a quad (relation) to the graph
func (cs *CayleyStore) AddQuad(subject, predicate, object, label string) error {
	return cs.store.AddQuad(quad.Make(subject, predicate, object, label))
}

// RemoveQuad removes a quad (relation) from the graph
func (cs *CayleyStore) RemoveQuad(subject, predicate, object, label string) error {
	return cs.store.RemoveQuad(quad.Make(subject, predicate, object, label))
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

func (cs *CayleyStore) BuildRecursivePath(subject, predicate string, maxDepth int) *cayley.Path {
	return cayley.StartPath(cs.store, quad.String(subject)).FollowRecursive(quad.String(predicate), maxDepth, nil)

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

// BFS
func (cs *CayleyStore) BFS(startNode, relationship string) [][]string {
	queue := []string{startNode}
	visited := make(map[string]bool)
	visited[startNode] = true

	var result [][]string

	// 层次遍历
	for len(queue) > 0 {

		currentLevelSize := len(queue)
		var currentLevel []string // 当前层的节点

		for i := 0; i < currentLevelSize; i++ {
			node := queue[0]
			queue = queue[1:]
			currentLevel = append(currentLevel, node)
			path := cayley.StartPath(cs.store, quad.String(node)).Out(quad.String(relationship))

			_ = path.Iterate(nil).EachValue(nil, func(value quad.Value) {
				nextNode := quad.NativeOf(value).(string)
				if !visited[nextNode] {
					visited[nextNode] = true
					queue = append(queue, nextNode)
				}
			})
		}

		result = append(result, currentLevel)
	}

	return result
}

func (cs *CayleyStore) BFSByLevelWithMap(startNode, relationship string) [][]map[string][]string {
	// 使用队列进行层次遍历
	queue := []string{startNode}
	visited := make(map[string]bool)
	visited[startNode] = true

	var result [][]map[string][]string // 存储每一层的父子节点

	// 层次遍历
	for len(queue) > 0 {
		// 获取当前层的所有节点
		currentLevelSize := len(queue)
		var currentLevel []map[string][]string // 当前层存储父节点及其子节点

		// 遍历当前层的所有节点
		for i := 0; i < currentLevelSize; i++ {
			// 从队列中取出当前节点
			currentNode := queue[0]
			queue = queue[1:]

			// 存储当前节点的子节点
			children := []string{}

			// 找到当前节点的所有后继节点（下一层）
			path := cayley.StartPath(cs.store, quad.String(currentNode)).Out(quad.String(relationship))

			// 将所有未访问的后继节点放入队列
			_ = path.Iterate(nil).EachValue(nil, func(value quad.Value) {
				nextNode := quad.NativeOf(value).(string)
				if !visited[nextNode] {
					visited[nextNode] = true
					queue = append(queue, nextNode)
					children = append(children, nextNode)
				}
			})

			// 将父节点和其子节点组成的 map 放入当前层次
			if len(children) > 0 {
				currentLevel = append(currentLevel, map[string][]string{
					currentNode: children,
				})
			}
		}

		// 如果当前层有数据，加入结果中
		if len(currentLevel) > 0 {
			result = append(result, currentLevel)
		}
	}

	return result
}
