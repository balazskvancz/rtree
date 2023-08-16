# Radix-Tree

A simple radix/prefix tree (trie) implementation in Go. The type of the stored value is determined by the given generic template, which is based upon `interface{}`. Besides normal key-value pairs, it is possible to store keys with wildcard params as well. The main objective of this package is to store URLs which compatible with REST principles.

## Features

There is a generic factory method for creating a tree instance.

```go
tree := rtree.New[*Route]()

// Do something with the tree...
```

### Insert

To store a new node is by calling the `Insert` method on the Tree pointer. The first parameter of the function is the URL, second is the value – or a pointer to it – based on type that given as generic template at the tree creation. It could return an error, so it is advisable to handle it.

```go
tree := rtree.New[*Route]()

if err := tree.Insert("/foo/bar/baz", &Route{}); err != nil {
  fmt.Printf("tree inserting error: %v\n", err)
}
```

### Find

The `Find` searches for the given key in the tree and returns the corresponding node – if there was a match. By default, the search is conducted by involving wildcard path params as well.

To check whether the search was successfull – thus found the node – this only thing is to do, to check if the node is `nil` or not. 

The example below demonstrates the abilities:

```go
tree := rtree.New[*Route]()

if err := tree.Insert("/foo/{id}/baz", &Route{}); err != nil {
  fmt.Printf("tree inserting error: %v\n", err)
}

node := t.Find("/foo/5/baz")

if node == nil {
  fmt.Println("no match")
  return
}

// It returns the stored value with the original type – in this scenario the type of `value` would be *Route.
value := node.GetValue()

// GetParams returns ALL the params in a map[string]string format.
params := node.GetParams()

// eg. it would be "5" NOT 5.
id := params["id"]
```
