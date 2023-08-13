// URLs are stored in the form of: /api/foo/{resource}/{id}
// where „resource” and „id” are the two path params
// of the request.
//
// Examples of matching urls for the scheme above:
//
//	/api/foo/products/1										-> resource="products" 		id="1"
//	/api/foo/categories/example-category 	-> resource="categories"  id="example-category"
//
// Due to the way we store these routes, there is a chance of trying to store
// ambigoues routes aswell. That obviously would cause a bad behaviour.
// Example for these routes:
//
// /api/{resource}/get
// /api/products/get
//
// The two above would cause an error, however the two 2 below would not:
//
// /api/{resource}/get
// /api/products/get-all
package rtree
