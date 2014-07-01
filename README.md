restacular
==========

## Router
Several things to fix before being usable:
- stop adding / everywhere => moved that part to the router for now (still want a proper fix)
- rewrite the find method of the trie to use a loop to avoid passing a pointer to a map
- add more checks when adding route to prevent a bad tree

Perfs:
Initial:
	BenchmarkGettingRouteWithoutParam	20000000	       105 ns/op	       8 B/op	       0 allocs/op
	BenchmarkGettingRouteWithParam	 1000000	      1663 ns/op	     785 B/op	       6 allocs/op
After toLower(path)
	BenchmarkGettingRouteWithoutParam	10000000	       169 ns/op	       8 B/op	       0 allocs/op
	BenchmarkGettingRouteWithParam	 1000000	      1769 ns/op	     794 B/op	       7 allocs/op
After rewriting find()
	BenchmarkGettingRouteWithoutParam	10000000	       217 ns/op	      16 B/op	       1 allocs/op
	BenchmarkGettingRouteWithOneParam	 1000000	      1156 ns/op	     436 B/op	       5 allocs/op
	BenchmarkGettingRouteWithTwoParam	 1000000	      1809 ns/op	     487 B/op	       5 allocs/op

## TODO now:
- create interface for a Resource as defined in the gist
- add routes to the router from a resource object (use https://github.com/julienschmidt/httprouter probably)
- automatic JSON encoding of response
- settings: able to load some from env and toml
- middlewares

## TODO soon:
- CORS
- logging
- metrics
- reverse for HATEOAS
- validation
- example app copying github API to see how well it works and tweak things to make it nicer to use

