restacular
==========

## Router
Few things to think of:
- reverse urls

Perfs:
BenchmarkGettingRouteWithoutParam	50000000	        32.2 ns/op	       0 B/op	       0 allocs/op
BenchmarkGettingRouteWithOneParam	10000000	       216 ns/op	      65 B/op	       1 allocs/op
BenchmarkGettingRouteWithTwoParam	10000000	       254 ns/op	      65 B/op	       1 allocs/op

Performance will degrade a bit if you have more than 2 params in your URL (about 200ns more per additional param).

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

