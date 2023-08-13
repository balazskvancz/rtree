package rtree

import (
	"errors"
	"reflect"
	"testing"
)

type Route struct {
	name string
}

type getTreeFn func(t *testing.T) *Tree[*Route]

func getRoute() *Route {
	return &Route{}
}

func TestGetOffSets(t *testing.T) {
	type testCase struct {
		name       string
		storedKey  string
		searchKey  string
		isWildcard bool

		expectedOffset1    int
		expectedOffset2    int
		expectedIsWildcard bool
	}

	tt := []testCase{
		{
			name:               "no wildcard and no match",
			storedKey:          "foo",
			searchKey:          "bar",
			isWildcard:         false,
			expectedOffset1:    0,
			expectedOffset2:    0,
			expectedIsWildcard: false,
		},
		{
			name:               "no strict match, but wildcard by def (and still wildcard)",
			storedKey:          "foo",
			searchKey:          "bar",
			isWildcard:         true,
			expectedOffset1:    3,
			expectedOffset2:    0,
			expectedIsWildcard: true,
		},
		{
			name:               "no strict match, but wildcard by def (and not wildcard anymore)",
			storedKey:          "foo}",
			searchKey:          "bar",
			isWildcard:         true,
			expectedOffset1:    4,
			expectedOffset2:    3,
			expectedIsWildcard: false,
		},
		{
			name:               "no strict match, wildcard but not by def (and not wildcard anymore)",
			storedKey:          "{foo}",
			searchKey:          "bar",
			isWildcard:         false,
			expectedOffset1:    5,
			expectedOffset2:    3,
			expectedIsWildcard: false,
		},
		{
			name:               "no strict match, wildcard but not by def (and still wildcard)",
			storedKey:          "{foo",
			searchKey:          "bar",
			isWildcard:         false,
			expectedOffset1:    4,
			expectedOffset2:    0,
			expectedIsWildcard: true,
		},

		{
			name:               "strict match and wildcard but not by def (and not wildcard anymore)",
			storedKey:          "/foo/{id}",
			searchKey:          "/foo/5/6",
			isWildcard:         false,
			expectedOffset1:    9,
			expectedOffset2:    6,
			expectedIsWildcard: false,
		},

		{
			name:               "strict match and wildcard but not by def (and still wildcard)",
			storedKey:          "/foo/{id",
			searchKey:          "/foo/5/6",
			isWildcard:         false,
			expectedOffset1:    8,
			expectedOffset2:    5,
			expectedIsWildcard: true,
		},
		{
			name:               "strict match then wildcard then strict match again",
			storedKey:          "/{id}/bar",
			searchKey:          "/5/bar",
			isWildcard:         false,
			expectedOffset1:    9,
			expectedOffset2:    6,
			expectedIsWildcard: false,
		},
		{
			name:               "strict match then wildcard then not strict match",
			storedKey:          "/{id}/bar",
			searchKey:          "/5/foo",
			isWildcard:         false,
			expectedOffset1:    6,
			expectedOffset2:    3,
			expectedIsWildcard: false,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			gotOffset1, gotOffset2, gotIsWildcard := getOffsets(tc.storedKey, tc.searchKey, tc.isWildcard)

			if gotOffset1 != tc.expectedOffset1 {
				t.Errorf("expected offset1: %d; got: %d\n", tc.expectedOffset1, gotOffset1)
			}

			if gotOffset2 != tc.expectedOffset2 {
				t.Errorf("expected offset2: %d; got: %d\n", tc.expectedOffset2, gotOffset2)
			}

			if gotIsWildcard != tc.expectedIsWildcard {
				t.Errorf("expected isWildcard: %v; got: %v\n", tc.expectedIsWildcard, gotIsWildcard)
			}
		})
	}
}

func TestTreeInsert(t *testing.T) {
	type testCase struct {
		name    string
		getTree getTreeFn
		input   string
		err     error
	}

	tt := []testCase{
		{
			name:    "error if the tree is <nil>",
			getTree: func(t *testing.T) *Tree[*Route] { return nil },
			input:   "",
			err:     errTreeIsNil,
		},
		{
			name: "error if given url (key) is empty",
			getTree: func(t *testing.T) *Tree[*Route] {
				return New[*Route]()
			},
			input: "",
			err:   errKeyIsEmpty,
		},
		{
			name: "error if given url (key) is not starting with a slash",
			getTree: func(t *testing.T) *Tree[*Route] {
				return New[*Route]()
			},
			input: "foo",
			err:   errMissingSlashPrefix,
		},
		{
			name: "error if given url (key) is ending with a slash",
			getTree: func(t *testing.T) *Tree[*Route] {
				return New[*Route]()
			},
			input: "/foo/",
			err:   errPresentSlashSuffix,
		},
		{
			name: "no error if insertion was successfull (empty tree)",
			getTree: func(t *testing.T) *Tree[*Route] {
				return New[*Route]()
			},
			input: "/foo",
			err:   nil,
		},
		{
			name: "no error if insertion was successful (not empty tree)",
			getTree: func(t *testing.T) *Tree[*Route] {
				tree := New[*Route]()

				if err := tree.Insert("/foo/bar", getRoute()); err != nil {
					t.Fatalf("unexpected error: %v\n", err)
				}

				if err := tree.Insert("/foo/baz", getRoute()); err != nil {
					t.Fatalf("unexpected error: %v\n", err)
				}

				return tree
			},
			input: "/foo",
			err:   nil,
		},
		{
			name: "no error if insertion successful similar keys (not empty tree)",
			getTree: func(t *testing.T) *Tree[*Route] {
				tree := New[*Route]()

				if err := tree.Insert("/foo", getRoute()); err != nil {
					t.Fatalf("unexpected error: %v\n", err)
				}

				return tree
			},
			input: "/fo",
			err:   nil,
		},
		{
			name: "error on duplicate keys",
			getTree: func(t *testing.T) *Tree[*Route] {
				tree := New[*Route]()

				if err := tree.Insert("/foo/bar", getRoute()); err != nil {
					t.Fatalf("unexpected error: %v\n", err)
				}

				if err := tree.Insert("/foo/baz", getRoute()); err != nil {
					t.Fatalf("unexpected error: %v\n", err)
				}

				if err := tree.Insert("/foo", getRoute()); err != nil {
					t.Fatalf("unexpected error: %v\n", err)
				}

				return tree
			},
			input: "/foo",
			err:   errKeyIsAlreadyStored,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			tree := tc.getTree(t)

			err := tree.Insert(tc.input, getRoute())

			if tc.err != nil && !errors.Is(err, tc.err) {
				t.Errorf("expected error: %v; got: %v\n", tc.err, err)
			}

			if tc.err == nil && err != nil {
				t.Errorf("unexpected error: %v\n", err)
			}

		})
	}

}

func TestTreeFind(t *testing.T) {
	type testCase struct {
		name      string
		getTree   func(t *testing.T) *Tree[*Route]
		searchKey string
		isExists  bool
	}

	tt := []testCase{
		{
			name:      "cant find, if tree is <nil>",
			getTree:   func(t *testing.T) *Tree[*Route] { return nil },
			searchKey: "/foo",
			isExists:  false,
		},
		{
			name: "cant find, if root of tree is <nil>",
			getTree: func(t *testing.T) *Tree[*Route] {
				return &Tree[*Route]{}
			},
			searchKey: "/foo",
			isExists:  false,
		},
		{
			name: "cant find, if the search key is empty",
			getTree: func(t *testing.T) *Tree[*Route] {
				tree := New[*Route]()

				if err := tree.Insert("/api/foo", getRoute()); err != nil {
					t.Fatalf("not expected error, but got: %v\n", err)
				}

				return tree
			},
			searchKey: "",
			isExists:  false,
		},

		// Simply not wildcard test.
		{
			name: "normal search without any wildcard (no match)",
			getTree: func(t *testing.T) *Tree[*Route] {
				tree := New[*Route]()

				if err := tree.Insert("/foo/bar/baz", getRoute()); err != nil {
					t.Fatalf("not expected error, but got: %v\n", err)
				}

				if err := tree.Insert("/foo/bar/bar", getRoute()); err != nil {
					t.Fatalf("not expected error, but got: %v\n", err)
				}

				return tree
			},
			searchKey: "/foo/bar/bak",
			isExists:  false,
		},
		{
			name: "normal search without any wildcard (no match, only common subpart)",
			getTree: func(t *testing.T) *Tree[*Route] {
				tree := New[*Route]()

				if err := tree.Insert("/foo/bar/baz", getRoute()); err != nil {
					t.Fatalf("not expected error, but got: %v\n", err)
				}

				if err := tree.Insert("/foo/bar/bar", getRoute()); err != nil {
					t.Fatalf("not expected error, but got: %v\n", err)
				}

				return tree
			},
			searchKey: "/foo/bar",
			isExists:  false,
		},
		{
			name: "normal search without any wildcard (match)",
			getTree: func(t *testing.T) *Tree[*Route] {
				tree := New[*Route]()

				if err := tree.Insert("/foo/bar/baz", getRoute()); err != nil {
					t.Fatalf("not expected error, but got: %v\n", err)
				}

				if err := tree.Insert("/foo/bar/bar", getRoute()); err != nil {
					t.Fatalf("not expected error, but got: %v\n", err)
				}

				return tree
			},
			searchKey: "/foo/bar/bar",
			isExists:  true,
		},

		// search with wildcard param
		{
			name: "wildcard search (no match)",
			getTree: func(t *testing.T) *Tree[*Route] {
				tree := New[*Route]()

				if err := tree.Insert("/foo/{id}/baz", getRoute()); err != nil {
					t.Fatalf("not expected error, but got: %v\n", err)
				}

				if err := tree.Insert("/foo/bar/bar", getRoute()); err != nil {
					t.Fatalf("not expected error, but got: %v\n", err)
				}

				return tree
			},
			searchKey: "/foo/bar/bak",
			isExists:  false,
		},
		{
			name: "wildcard search - param is at the start (match)",
			getTree: func(t *testing.T) *Tree[*Route] {
				tree := New[*Route]()

				if err := tree.Insert("/{id}/baz/bar", getRoute()); err != nil {
					t.Fatalf("not expected error, but got: %v\n", err)
				}

				if err := tree.Insert("/foo/bak/{id}", getRoute()); err != nil {
					t.Fatalf("not expected error, but got: %v\n", err)
				}

				if err := tree.Insert("/for/bak/{id}", getRoute()); err != nil {
					t.Fatalf("not expected error, but got: %v\n", err)
				}

				if err := tree.Insert("/foo/bar/bar", getRoute()); err != nil {
					t.Fatalf("not expected error, but got: %v\n", err)
				}

				return tree
			},
			searchKey: "/1/baz/bar",
			isExists:  true,
		},
		{
			name: "wildcard search - param is in the middle (match)",
			getTree: func(t *testing.T) *Tree[*Route] {
				tree := New[*Route]()

				if err := tree.Insert("/foo/{id}/baz", getRoute()); err != nil {
					t.Fatalf("not expected error, but got: %v\n", err)
				}

				if err := tree.Insert("/foo/bar/bar", getRoute()); err != nil {
					t.Fatalf("not expected error, but got: %v\n", err)
				}

				return tree
			},
			searchKey: "/foo/1/baz",
			isExists:  true,
		},
		{
			name: "wildcard search - param is at the end (match)",
			getTree: func(t *testing.T) *Tree[*Route] {
				tree := New[*Route]()

				if err := tree.Insert("/foo/baz/{id}", getRoute()); err != nil {
					t.Fatalf("not expected error, but got: %v\n", err)
				}

				if err := tree.Insert("/foo/bak/{id}", getRoute()); err != nil {
					t.Fatalf("not expected error, but got: %v\n", err)
				}

				if err := tree.Insert("/for/bak/{id}", getRoute()); err != nil {
					t.Fatalf("not expected error, but got: %v\n", err)
				}

				if err := tree.Insert("/foo/bar/bar", getRoute()); err != nil {
					t.Fatalf("not expected error, but got: %v\n", err)
				}

				return tree
			},
			searchKey: "/foo/baz/1",
			isExists:  true,
		},

		// Multiple params
		{
			name: "wildcard search - multiple params (no match)",
			getTree: func(t *testing.T) *Tree[*Route] {
				tree := New[*Route]()

				if err := tree.Insert("/api/{resource}/get/{id}", getRoute()); err != nil {
					t.Fatalf("not expected error, but got: %v\n", err)
				}

				if err := tree.Insert("/api/{resource}/delete/{id}", getRoute()); err != nil {
					t.Fatalf("not expected error, but got: %v\n", err)
				}

				if err := tree.Insert("/api/{resource}/update/{id}", getRoute()); err != nil {
					t.Fatalf("not expected error, but got: %v\n", err)
				}

				return tree
			},
			searchKey: "/api/products/insert/1",
			isExists:  false,
		},
		{
			name: "wildcard search - multiple params (match)",
			getTree: func(t *testing.T) *Tree[*Route] {
				tree := New[*Route]()

				if err := tree.Insert("/api/{resource}/get/{id}", getRoute()); err != nil {
					t.Fatalf("not expected error, but got: %v\n", err)
				}

				if err := tree.Insert("/api/{resource}/delete/{id}", getRoute()); err != nil {
					t.Fatalf("not expected error, but got: %v\n", err)
				}

				if err := tree.Insert("/api/{resource}/update/{id}", getRoute()); err != nil {
					t.Fatalf("not expected error, but got: %v\n", err)
				}

				return tree
			},
			searchKey: "/api/products/delete/1",
			isExists:  true,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			tree := tc.getTree(t)

			node := tree.Find(tc.searchKey)

			if tc.isExists && node == nil {
				t.Errorf("expected to find, but got <nil>")
			}

			if !tc.isExists && node != nil {
				t.Errorf("expected not to find, but got route")
			}
		})
	}
}

func TestCheckPathParams(t *testing.T) {
	type testCase struct {
		name  string
		input string
		err   error
	}

	tt := []testCase{
		{
			name:  "no error, if there is no path params at all",
			input: "/foo/bar/baz",
			err:   nil,
		},
		{
			name:  "error if there no closing of param",
			input: "/foo/bar/{baz",
			err:   errBadPathParamSyntax,
		},
		{
			name:  "error if there no start of param",
			input: "/foo/bar/baz}",
			err:   errBadPathParamSyntax,
		},
		{
			name:  "error if multiple start of param",
			input: "/foo/bar/{{baz",
			err:   errBadPathParamSyntax,
		},
		{
			name:  "error if multiple end of param",
			input: "/foo/bar/baz}}",
			err:   errBadPathParamSyntax,
		},
		{
			name:  "error if there is a slash inside a param",
			input: "/foo/bar{/baz}",
			err:   errBadPathParamSyntax,
		},
		{
			name:  "no error if one path param",
			input: "/foo/bar/{baz}",
			err:   nil,
		},
		{
			name:  "no error if multiple path param",
			input: "/{foo}/bar/{baz}",
			err:   nil,
		},
		{
			name:  "error if one good and one bad param",
			input: "/{foo}/bar/baz}",
			err:   errBadPathParamSyntax,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			if err := checkPathParams(tc.input); !errors.Is(err, tc.err) {
				t.Errorf("expected error: %v; got: %v\n", tc.err, err)
			}
		})
	}
}

func TestFindLongestMatch(t *testing.T) {
	type testCase struct {
		name         string
		getTree      getTreeFn
		input        string
		expectedName string
	}

	tt := []testCase{
		{
			name:  "empty tree, no find",
			input: "/api/products/list-all",
			getTree: func(t *testing.T) *Tree[*Route] {
				return New[*Route]()
			},
			expectedName: "",
		},
		{
			name:  "not empty tree, but not matching leaf",
			input: "/api/products/list-all",
			getTree: func(t *testing.T) *Tree[*Route] {
				tree := New[*Route]()

				tree.Insert("/api/users", &Route{
					name: "users",
				})

				tree.Insert("/api/categories", &Route{
					name: "categories",
				})

				return tree
			},
			expectedName: "",
		},
		{
			name:  "not empty tree, and matching leaf",
			input: "/api/products/list-all",
			getTree: func(t *testing.T) *Tree[*Route] {
				tree := New[*Route]()

				tree.Insert("/api/users", &Route{
					name: "users",
				})

				tree.Insert("/api/categories", &Route{
					name: "categories",
				})

				tree.Insert("/api/products", &Route{
					name: "products",
				})

				return tree
			},
			expectedName: "products",
		},
		{
			name:  "not empty tree, and matching leaf with similiar services",
			input: "/api/professions/list-all",
			getTree: func(t *testing.T) *Tree[*Route] {
				tree := New[*Route]()

				tree.Insert("/api/professions", &Route{
					name: "professions",
				})

				tree.Insert("/api/products", &Route{
					name: "products",
				})

				return tree
			},
			expectedName: "professions",
		},
		{
			name:  "not empty tree, no matching leaf nearly good key",
			input: "/api/profession/list-all",
			getTree: func(t *testing.T) *Tree[*Route] {
				tree := New[*Route]()

				tree.Insert("/api/professions", &Route{
					name: "professions",
				})

				tree.Insert("/api/products", &Route{
					name: "products",
				})

				return tree
			},
			expectedName: "",
		},
		{
			name:  "not empty tree, no matching leaf nearly good key (not leaf)",
			input: "/api/products/list-all",
			getTree: func(t *testing.T) *Tree[*Route] {
				tree := New[*Route]()

				tree.Insert("/api/products/outer", &Route{
					name: "outer",
				})

				tree.Insert("/api/products/inner", &Route{
					name: "inner",
				})

				return tree
			},
			expectedName: "",
		},
		{
			name:  "not empty tree, matching similiar prefixes",
			input: "/api/products/list-all",
			getTree: func(t *testing.T) *Tree[*Route] {
				tree := New[*Route]()

				tree.Insert("/api/products", &Route{
					name: "products",
				})

				tree.Insert("/api/products/outer", &Route{
					name: "outer",
				})

				tree.Insert("/api/products/inner", &Route{
					name: "inner",
				})

				return tree
			},
			expectedName: "products",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			tree := tc.getTree(t)

			match := tree.FindLongestMatch(tc.input)

			if tc.expectedName == "" && match != nil {
				t.Error("expected not to find any node, but found")
			}

			if tc.expectedName != "" {
				if match == nil {
					t.Error("expected to find node, but found none")
					return
				}

				c := match.value

				if tc.expectedName != c.name {
					t.Errorf("expected value: %s; got %s\n", tc.expectedName, c.name)
				}
			}
		})
	}
}

func TestNewGetOffSets(t *testing.T) {
	type testCase struct {
		name       string
		storedKey  string
		searchKey  string
		isWildcard bool

		expectedOffset1    int
		expectedOffset2    int
		expectedIsWildcard bool
	}

	tt := []testCase{
		{
			name:               "no wildcard and no match",
			storedKey:          "foo",
			searchKey:          "bar",
			isWildcard:         false,
			expectedOffset1:    0,
			expectedOffset2:    0,
			expectedIsWildcard: false,
		},
		{
			name:               "no strict match, but wildcard by def (and still wildcard)",
			storedKey:          "foo",
			searchKey:          "bar",
			isWildcard:         true,
			expectedOffset1:    3,
			expectedOffset2:    0,
			expectedIsWildcard: true,
		},
		{
			name:               "no strict match, but wildcard by def (and not wildcard anymore)",
			storedKey:          "foo}",
			searchKey:          "bar",
			isWildcard:         true,
			expectedOffset1:    4,
			expectedOffset2:    3,
			expectedIsWildcard: false,
		},
		{
			name:               "no strict match, wildcard but not by def (and not wildcard anymore)",
			storedKey:          "{foo}",
			searchKey:          "bar",
			isWildcard:         false,
			expectedOffset1:    5,
			expectedOffset2:    3,
			expectedIsWildcard: false,
		},
		{
			name:               "no strict match, wildcard but not by def (and still wildcard)",
			storedKey:          "{foo",
			searchKey:          "bar",
			isWildcard:         false,
			expectedOffset1:    4,
			expectedOffset2:    0,
			expectedIsWildcard: true,
		},

		{
			name:               "strict match and wildcard but not by def (and not wildcard anymore)",
			storedKey:          "/foo/{id}",
			searchKey:          "/foo/5/6",
			isWildcard:         false,
			expectedOffset1:    9,
			expectedOffset2:    6,
			expectedIsWildcard: false,
		},

		{
			name:               "strict match and wildcard but not by def (and still wildcard)",
			storedKey:          "/foo/{id",
			searchKey:          "/foo/5/6",
			isWildcard:         false,
			expectedOffset1:    8,
			expectedOffset2:    5,
			expectedIsWildcard: true,
		},
		{
			name:               "strict match then wildcard then strict match again",
			storedKey:          "/{id}/bar",
			searchKey:          "/5/bar",
			isWildcard:         false,
			expectedOffset1:    9,
			expectedOffset2:    6,
			expectedIsWildcard: false,
		},
		{
			name:               "strict match then wildcard then not strict match",
			storedKey:          "/{id}/bar",
			searchKey:          "/5/foo",
			isWildcard:         false,
			expectedOffset1:    6,
			expectedOffset2:    3,
			expectedIsWildcard: false,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			gotOffset1, gotOffset2, gotIsWildcard := getOffsets(tc.storedKey, tc.searchKey, tc.isWildcard)

			if gotOffset1 != tc.expectedOffset1 {
				t.Errorf("expected offset1: %d; got: %d\n", tc.expectedOffset1, gotOffset1)
			}

			if gotOffset2 != tc.expectedOffset2 {
				t.Errorf("expected offset2: %d; got: %d\n", tc.expectedOffset2, gotOffset2)
			}

			if gotIsWildcard != tc.expectedIsWildcard {
				t.Errorf("expected isWildcard: %v; got: %v\n", tc.expectedIsWildcard, gotIsWildcard)
			}
		})
	}
}

func TestGetPathParams(t *testing.T) {
	type testCase struct {
		name     string
		input    string
		expected []paramInfo
	}

	tt := []testCase{
		{
			name:     "no params in input, empty slice",
			input:    "/foo/bar/baz",
			expected: []paramInfo{},
		},
		{
			name:  "params in input, proper slice",
			input: "/foo/{bar}/baz/{id}",
			expected: []paramInfo{
				{
					key: "bar",
					pos: 2,
				},
				{
					key: "id",
					pos: 4,
				},
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			got := getPathParams(tc.input)

			if !reflect.DeepEqual(tc.expected, got) {
				t.Error("bad; todo fix: more informative error")
			}
		})
	}
}

func TestMatchParams(t *testing.T) {
	type testCase struct {
		name   string
		params []paramInfo
		input  string

		expected matchedParams
	}

	tt := []testCase{
		{
			name:     "empty map, if no params",
			params:   []paramInfo{},
			input:    "/foo/bar/baz",
			expected: map[string]string{},
		},
		{
			name: "returns the good params",
			params: []paramInfo{
				{
					key: "first-one",
					pos: 2,
				},
				{
					key: "second-one",
					pos: 3,
				},
			},
			input: "/foo/bar/baz",
			expected: map[string]string{
				"first-one":  "bar",
				"second-one": "baz",
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			got := matchParams(tc.params, tc.input)

			if !reflect.DeepEqual(tc.expected, got) {
				t.Error("bad; todo fix: more informative error")
			}
		})
	}
}
