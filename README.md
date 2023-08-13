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
// Other work with found node.
```
