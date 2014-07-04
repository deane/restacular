restacular
==========

## Router
Several things to fix before being usable:
- add more checks when adding route to prevent a bad tree (+ tests)
- count number of child having params

Perfs:
BenchmarkGettingRouteWithoutParam	50000000	        32.4 ns/op	       0 B/op	       0 allocs/op
BenchmarkGettingRouteWithOneParam	10000000	       247 ns/op	      97 B/op	       1 allocs/op
BenchmarkGettingRouteWithTwoParam	10000000	       285 ns/op	      97 B/op	       1 allocs/op

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

