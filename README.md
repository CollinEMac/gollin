# gollin

## What is it?
 
Gollin is an opinionated superset of Go that adds some syntactic sugar. The main focus is on features that have been requested and denied by Go's maintainers.

## What does it do?

### Implemented ✅ 

- Try/catch blocks that compile to `if err != nil {}` syntax.
    - note that nested try/catch doesn't work right now.
    - example: 
    
```
    f := try {
        os.Open("test.txt")
    } catch {
        fmt.Println("I could not open that text file");
    }
```

- String interpolation

``` 
    cheese := "cheddar"
    n := 1000
    fmt.Printf("I like %s times %d", cheese, n)
```

- Ternary operator

```
    holes := 2
    type := holes >= 1 ? "swiss" : "cheddar"
    fmt.Println(type)
```

### Planned 📋

- Nil-coalescing operator  (`x ?? defaultValue`)
- Optional chaining like `foo?.bar`

## Why?

Mostly because I thought it would be fun. I also hope to use it for my Go projects moving forward.
    
## Is this project stable?

Hell no.
