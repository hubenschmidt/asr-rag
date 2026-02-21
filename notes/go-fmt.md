# Go fmt Cheat Sheet

## Print functions — where does the output go?

| Function       | Destination              | Formatting     |
|----------------|--------------------------|----------------|
| `fmt.Println`  | stdout                   | no verbs       |
| `fmt.Printf`   | stdout                   | uses verbs     |
| `fmt.Fprintln` | you choose (first arg)   | no verbs       |
| `fmt.Fprintf`  | you choose (first arg)   | uses verbs     |
| `fmt.Errorf`   | returns an `error` value | uses verbs     |

## Verbs — what goes in the placeholder?

| Verb | Use for | Example |
|------|---------|---------|
| `%s` | strings only | `fmt.Printf("name: %s", name)` |
| `%d` | integers only | `fmt.Printf("count: %d", 42)` |
| `%v` | anything (Go picks format) | `fmt.Printf("value: %v", whatever)` |
| `%w` | wrapping errors (Errorf only) | `fmt.Errorf("failed: %w", err)` |

## Examples

### Printing to stdout

```go
// no formatting — just prints values separated by spaces, adds newline
fmt.Println("hello", name, age)
// output: hello Alice 30

// with verbs — you control the format
fmt.Printf("name: %s, age: %d\n", name, age)
// output: name: Alice, age: 30
```

### Printing to stderr

```go
// no verbs
fmt.Fprintln(os.Stderr, "something went wrong:", err)

// with verbs
fmt.Fprintf(os.Stderr, "config: %v\n", err)
```

### Returning errors

```go
// %w wraps the error — preserves the chain for errors.Is() / errors.As()
return nil, fmt.Errorf("read %s: %w", path, err)

// %v just stringifies — loses the chain (avoid for errors)
return nil, fmt.Errorf("read %s: %v", path, err)
```

### When to use what

```go
// telling the user something → Println
fmt.Println("seed complete:", count, "terms")

// error message to stderr → Fprintf
fmt.Fprintf(os.Stderr, "failed: %v\n", err)

// building an error to return → Errorf
return fmt.Errorf("embed %s: %w", term, err)

// don't care about verbs → just concat with Println
fmt.Println("failed: " + err.Error())
```

## Common mistakes

```go
// WRONG — Println doesn't use verbs, prints literal "%v"
fmt.Println("error: %v", err)

// WRONG — Errorf returns an error, doesn't print anything
fmt.Errorf("something failed: %w", err) // value thrown away

// WRONG — %s on a non-string
fmt.Printf("count: %s", 42) // prints "%!s(int=42)"

// RIGHT
fmt.Println("error:", err)
return fmt.Errorf("something failed: %w", err)
fmt.Printf("count: %d", 42)
```
