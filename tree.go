package rtree

import (
	"errors"
	"strings"
	"sync"
)

const (
	slash = '/'

	curlyStart = '{'
	curlyEnd   = '}'
)

type storeValue interface {
	any
}

type (
	predicateFunction[T storeValue] func(*node[T]) bool
	OptionFunc[T storeValue]        func(*Tree[T])
)

var (
	errBadPathParamSyntax = errors.New("[tree]: bad path param syntax")
	errKeyIsAlreadyStored = errors.New("[tree]: key is already stored")
	errKeyIsEmpty         = errors.New("[tree]: key is empty")
	errMissingSlashPrefix = errors.New("[tree]: urls must be started with a '/'")
	errNoCommonPrefix     = errors.New("[tree]: no commmon prefix in given strings")
	errPresentSlashSuffix = errors.New("[tree]: urls must not be ended with a '/'")
	errRootIsNil          = errors.New("[tree]: the root of the tree is <nil>")
	errTreeIsNil          = errors.New("[tree]: the tree is <nil>")
)

type Tree[T storeValue] struct {
	mu   sync.RWMutex
	root *node[T]
}

type node[T storeValue] struct {
	key    string
	value  T
	isLeaf bool

	children []*node[T]
}

// IsLeaf returns whether a node is a leaf.
func (n *node[T]) IsLeaf() bool {
	return n.isLeaf
}

// GetValue returns the stored value of a pointer to a node.
func (n *node[T]) GetValue() T {
	return n.value
}

func New[T storeValue](opts ...OptionFunc[T]) *Tree[T] {
	t := &Tree[T]{
		mu: sync.RWMutex{},
	}

	for _, o := range opts {
		o(t)
	}

	return t
}

// insert tries to store a key-value pair in the tree.
// In case of unsuccessful insertion, we return the root of the error.
func (t *Tree[T]) Insert(key string, value T) error {
	if t == nil {
		return errTreeIsNil
	}

	if key == "" {
		return errKeyIsEmpty
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	if err := checkUrl(key); err != nil {
		return err
	}

	// If the root is still nil, then the new node is the root.
	if t.root == nil {
		t.root = createNewNode(key, value)
		return nil
	}

	return insertRec(t.root, key, value)
}

// iterateInsert iterates on the given node's children, and calls
// insertRec on each one. If there is no error during the recursive calls
// we successfully inserted the new node. Otherwise, if get an error that
// differs from errNoCommonPrefix, we return it. If none of those happaned, we
// simply return errNoCommonPrefix which indicates we were trying to
// insert on a wrong branch.
func iterateInsert[T storeValue](n *node[T], key string, value T) error {
	for _, ch := range n.children {
		insertErr := insertRec(ch, key, value)

		if insertErr == nil {
			return nil
		}

		if !errors.Is(insertErr, errNoCommonPrefix) {
			return insertErr
		}
	}

	return errNoCommonPrefix
}

func insertRec[T storeValue](n *node[T], key string, value T) error {
	lcp := longestCommonPrefix(n.key, key)

	// There is no chance of inserting in this branch.
	if lcp == 0 {
		return errNoCommonPrefix
	}

	var (
		currentKeyLen = len(n.key)
		keyLen        = len(key)
	)

	// If the length of the common part is equal to the inserting key,
	// then the current node is place we wanted to insert in the first place.
	if currentKeyLen == lcp && keyLen == lcp {
		// If it is already leaf, return error.
		if n.IsLeaf() {
			return errKeyIsAlreadyStored
		}
		// Otherwise we simply the store the value and we are done.
		n.value = value

		return nil
	}

	// Three other possibilities:
	// 		1) the current node's key is longer than the LCP => must split keys,
	// 		2) current node's are same as lcp, and new key is longer =>,
	// 		3) otherwise the new node should be amongs the children of the current node.
	if currentKeyLen > lcp {
		cNewNode := createNewNode(n.key[lcp:], n.value, n.children...)

		// If the key to be inserted is just as long as the stored key
		// then we have to store it here.
		keyRem := key[lcp:]
		if keyRem == "" {
			n.key = n.key[:lcp]
			n.value = value
			n.children = []*node[T]{cNewNode}
			n.isLeaf = true

			return nil
		}

		newNode := createNewNode(keyRem, value)

		var defValue T

		n.value = defValue
		n.isLeaf = false
		n.key = n.key[:lcp]
		n.children = []*node[T]{cNewNode, newNode}

		return nil
	}

	keyRem := key[lcp:]

	err := iterateInsert(n, keyRem, value)

	if err == nil {
		return nil
	}

	if !errors.Is(err, errNoCommonPrefix) {
		return err
	}

	addToChildren(n, createNewNode(keyRem, value))

	return nil
}

func addToChildren[T storeValue](n, newNode *node[T]) {
	n.children = append(n.children, newNode)
}

// checkUrl checks the given of errors such as missing slash prefix
// or bad path params.
func checkUrl(url string) error {
	// Leading slash.
	if url[0] != slash {
		return errMissingSlashPrefix
	}

	// Trailing slash.
	if url[len(url)-1] == slash {
		return errPresentSlashSuffix
	}

	// Check for path params, and check for its syntax.
	return checkPathParams(url)
}

func checkPathParams(url string) error {
	// If there is none of the curly brackets, we are good to go.
	if !strings.ContainsRune(url, curlyStart) && !strings.ContainsRune(url, curlyEnd) {
		return nil
	}

	var (
		insideParam = false
		counter     = 0
	)

	for counter < len(url) {
		if url[counter] == slash {
			// If we are inside a path param, there cant be a slash.
			if insideParam {
				return errBadPathParamSyntax
			}
		}

		if url[counter] == curlyStart {
			if insideParam {
				return errBadPathParamSyntax
			}

			insideParam = true
		}

		if url[counter] == curlyEnd {
			if !insideParam {
				return errBadPathParamSyntax
			}
			insideParam = false
		}

		counter++
	}

	// If we are still inside a path param
	// after the url is ended, means error.
	if insideParam {
		return errBadPathParamSyntax
	}

	return nil
}

// checkTree does a basic check on the given tree, returns error
// if either the tree or the root is nil.
func checkTree[T storeValue](t *Tree[T]) error {
	if t == nil {
		return errTreeIsNil
	}

	if t.root == nil {
		return errRootIsNil
	}

	return nil
}

// min returns the minimum of two given numbers.
func min(num1, num2 int) int {
	if num1 > num2 {
		return num2
	}

	return num1
}

// longestCommonPrefix returns the length of the
// longest common prefix of two given strings.
func longestCommonPrefix(str1, str2 string) int {
	var counter = 0

	maxVal := min(len(str1), len(str2))

	for counter < maxVal && str1[counter] == str2[counter] {
		counter += 1
	}

	return counter
}

// createNewNode is a factory for creating new nodes.
func createNewNode[T storeValue](key string, value T, children ...*node[T]) *node[T] {
	n := &node[T]{
		key:      key,
		value:    value,
		children: make([]*node[T], 0),
		isLeaf:   true,
	}

	if len(children) > 0 {
		n.children = children
	}

	return n
}

// find starts the search for given key and returns a pointer to
// the found node. If there is no match, it returns nil.
func (t *Tree[T]) Find(key string) *node[T] {
	if err := checkTree(t); err != nil {
		return nil
	}

	if key == "" {
		return nil
	}

	return findRec(t.root, key, false)
}

// findRec is the main logic for conducting the search in a recursive manner.
// It looks for match on the given node's level, and calls itself recursively
// amongs its children, until the search is over.
func findRec[T storeValue](n *node[T], key string, isWildcard bool) *node[T] {
	if n == nil {
		return nil
	}

	// If the current node's key contains curlyStart char,
	// that means there is a start of wildcard part.
	if strings.ContainsRune(n.key, curlyStart) {
		isWildcard = true
	}

	lcp := longestCommonPrefix(n.key, key)

	// If there is nothing in common and it is not wildcard, then we are off.
	if lcp == 0 && !isWildcard {
		return nil
	}

	// In case of non wildcard part, normal string comp.
	if !isWildcard {
		if key == n.key {
			return n
		}

		// If the current node's key is longer than the lcp, no match.
		if lcp < len(n.key) {
			return nil
		}

		// Otherwise have to look amongst the children recursively.
		for _, c := range n.children {
			if found := findRec(c, key[lcp:], isWildcard); found != nil {
				return found
			}
		}

		return nil
	}

	var (
		nodeKeyRem   = n.key[lcp:]
		searchKeyRem = key[lcp:]
	)

	offset1, offset2, isStillWildcard := getOffsets(nodeKeyRem, searchKeyRem, true)

	// Meaning we didnt shift until the last char, not a full match in this level.
	if len(nodeKeyRem) != offset1 {
		return nil
	}

	newSearchKey := searchKeyRem[offset2:]

	// If there is nothing from the original search key
	// we are on the exact node we were looking for.
	if newSearchKey == "" {
		// Only to check if this node is a leaf, or not.
		if n.IsLeaf() {
			return n
		}
		return nil
	}

	// Have to continue search on the next level.
	for _, ch := range n.children {
		if found := findRec(ch, newSearchKey, isStillWildcard); found != nil {
			return found
		}
	}

	return nil
}

// getOffsets returns the offset of the first and second given string and whether it is still
// a wildcard search. These offsets are displaying how far should each string be shifted, how long
// is the common part including wildcard option.
func getOffsets(storedKey, searchKey string, isWildcard bool) (int, int, bool) {
	var (
		i = 0
		j = 0

		storedKeyLen = len(storedKey)
		searchKeyLen = len(searchKey)
	)

	for {
		if i >= storedKeyLen {
			break
		}

		if j >= searchKeyLen && !isWildcard {
			break
		}

		if storedKey[i] == curlyStart {
			isWildcard = true
			i++
			continue
		}

		// In case of closing a {id} part, we have to
		// move forward in the search key aswell.
		if storedKey[i] == curlyEnd {
			isWildcard = false

			cSearchRem := searchKey[j:]

			nextSlashIdx := strings.IndexRune(cSearchRem, slash)

			j += func() int {
				// There is no other / remaining.
				if nextSlashIdx == -1 {
					return len(cSearchRem)
				}
				// Otherwise skip that amount.
				return nextSlashIdx
			}()

			i++

			continue
		}

		// If we are inside of a wildcard check,
		// we only increment the stored keys counter.
		if isWildcard {
			i++
			continue
		}

		if storedKey[i] != searchKey[j] {
			break
		}

		i++
		j++
	}

	return i, j, isWildcard
}

// FindLongestMatch is similar to find but it doesnt include storeValue wildcard params at all.
// And it is not looking for perfect match, rather it finds the longest „route” based on the given string.
// Used for storing services based on their prefixes.
func (t *Tree[T]) FindLongestMatch(key string) *node[T] {
	if err := checkTree(t); err != nil {
		return nil
	}

	if key == "" {
		return nil
	}

	return findLongestMatchRec(t.root, key)
}

func findLongestMatchRec[T storeValue](n *node[T], key string) *node[T] {
	if n == nil {
		return nil
	}

	lcp := longestCommonPrefix(n.key, key)

	if lcp == 0 {
		return nil
	}

	if lcp != len(n.key) {
		return nil
	}

	for _, ch := range n.children {
		if node := findLongestMatchRec(ch, key[lcp:]); node != nil {
			return node
		}
	}

	if !n.IsLeaf() {
		return nil
	}

	return n
}

// GetAllLeaf returns all the stored leafs.
func (t *Tree[T]) GetAllLeaf() []*node[T] {
	if err := checkTree(t); err != nil {
		return nil
	}

	return getAllLeafRec(t.root)
}

func getAllLeafRec[T storeValue](n *node[T]) []*node[T] {
	arr := make([]*node[T], 0)

	for _, c := range n.children {
		chArr := getAllLeafRec(c)

		if len(chArr) > 0 {
			arr = append(arr, chArr...)
		}
	}

	if n.IsLeaf() {
		arr = append(arr, n)
	}

	return arr
}

// GetByPredicate does a search in the tree based on given function.
// It uses DFS as the algorithm to traverse the tree.
func (t *Tree[T]) GetByPredicate(fn predicateFunction[T]) *node[T] {
	if err := checkTree(t); err != nil {
		return nil
	}

	return getByPredicateRec(t.root, fn)
}

func getByPredicateRec[T storeValue](n *node[T], fn predicateFunction[T]) *node[T] {
	if n == nil {
		return nil
	}

	if fn(n) {
		return n
	}

	for _, ch := range n.children {
		if match := getByPredicateRec(ch, fn); match != nil {
			return match
		}
	}

	return nil
}
