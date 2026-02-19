# Go Pointers

## Two operators

| Op  | Name                       | Does                                                            | Example                    |
| --- | -------------------------- | --------------------------------------------------------------- | -------------------------- |
| `&` | address-of                 | creates a pointer to a value                                    | `p := &myStruct{}`         |
| `*` | dereference / pointer type | declares a pointer type, or reads the value a pointer points to | `var p *int` / `val := *p` |

## When to use pointers

- **Avoid copying large structs** — pass `*Config` instead of `Config`
- **Mutate the original** — a func receiving `*Config` can modify the caller's value
- **Signal "optional/nil"** — a pointer can be `nil`, a value type cannot

## Common patterns

```go
// Return a pointer to a local (Go heap-allocates it automatically)
func newConfig() *Config {
    cfg := Config{URL: "localhost"}
    return &cfg       // &cfg escapes to heap, totally safe
}

// Shorthand: literal + address-of in one step
func newConfig() *Config {
    return &Config{URL: "localhost"}
}

// Pointer receiver: method can mutate the struct
func (c *Config) SetURL(u string) {
    c.URL = u
}

// Value receiver: method gets a copy, original unchanged
func (c Config) GetURL() string {
    return c.URL
}
```

## Gotchas

- **nil pointer dereference** — calling a method or reading a field on a `nil` pointer panics. Guard with `if p == nil { return ... }`.
- **No pointer arithmetic** — unlike C, Go doesn't allow `p++` or `p + offset`.
- **Auto-dereference** — `p.Field` works the same as `(*p).Field`. Go handles it implicitly.
- **new() vs &{}** — `new(Config)` returns `*Config` with zero values. `&Config{...}` lets you set fields inline. Prefer `&Config{...}`.
