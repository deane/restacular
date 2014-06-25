restacular
==========

## Router
Several things to fix before being usable:

- rewrite the find method of the trie to use a loop to avoid passing a pointer to a map
- stop adding / everywhere
- benchmark and find things to optimize
- add more checks when adding route to prevent a bad tree


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
