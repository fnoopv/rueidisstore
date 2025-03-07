# rueidisstore

A [rueidis](https://github.com/redis/rueidis) bases session store for [scs](https://github.com/alexedwards/scs)

## Setup

You should follow the instructions to [set a client](https://github.com/redis/rueidis?tab=readme-ov-file#getting-started),
and pass the client to `rueidisstore.New()` or `rueidisstore.NewWithPrefix()`
to establish the session store.

## Keys

Default key is `scs:session:`, you can change it via

```go
sessionManagerOne.Store = rueidisstore.NewWithPrefix(client, "scs:session:1:")
```
