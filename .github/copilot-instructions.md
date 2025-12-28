# GitHub Copilot Code Review Instructions

## Performance Anti-Patterns to Flag

When reviewing Go code, check for these performance issues that are difficult to catch with static analyzers.

### 1. Unbounded Slice Growth (Memory Leak)

Flag struct fields that are slices which only grow via `append()` but are never truncated, reset, or bounded.

```go
// BAD - memory leak in long-running service
type Cache struct {
    items []Item
}

func (c *Cache) Add(item Item) {
    c.items = append(c.items, item)  // Never shrinks!
}

// GOOD - bounded with eviction
func (c *Cache) Add(item Item) {
    if len(c.items) >= c.maxSize {
        c.items = c.items[1:]  // Evict oldest
    }
    c.items = append(c.items, item)
}
```

**Review guidance**: Check if the type has any method that truncates/resets the slice. If not, flag as potential memory leak.

---

### 2. Unnecessary Copy (Filter Then Iterate)

Flag patterns where code filters into a new slice only to iterate over it once.

```go
// BAD - O(n) extra memory for single-use slice
var recent []Fact
for _, f := range facts {
    if f.CreatedAt.After(cutoff) {
        recent = append(recent, f)
    }
}
for _, f := range recent {
    process(f)
}

// GOOD - single pass, no intermediate slice
for _, f := range facts {
    if f.CreatedAt.After(cutoff) {
        process(f)
    }
}
```

**Review guidance**: If a filtered slice is only used once in a subsequent loop, suggest combining into single pass.

---

### 3. Goroutine Leak (Blocked Forever)

Flag goroutines that send to channels without cancellation support.

```go
// BAD - goroutine blocks forever if caller abandons channel
func fetchAll(urls []string) <-chan Result {
    ch := make(chan Result)  // Unbuffered!
    for _, url := range urls {
        go func(u string) {
            result := fetch(u)
            ch <- result  // Blocks if no reader
        }(url)
    }
    return ch
}

// GOOD - context cancellation
func fetchAll(ctx context.Context, urls []string) <-chan Result {
    ch := make(chan Result, len(urls))  // Buffered
    for _, url := range urls {
        go func(u string) {
            result := fetch(u)
            select {
            case ch <- result:
            case <-ctx.Done():
                return  // Exit on cancel
            }
        }(url)
    }
    return ch
}
```

**Review guidance**: Check that goroutines with channel sends either:
1. Use buffered channels sized for expected writes, OR
2. Use `select` with `ctx.Done()` case

---

### 4. Interface Boxing in Hot Path

Flag passing concrete types to `interface{}`/`any` parameters inside loops.

```go
// BAD - allocation per iteration for boxing
for _, n := range numbers {
    fmt.Println(n)           // n boxed to interface{}
    cache[key] = n           // n boxed if cache is map[string]any
}

// GOOD - use typed alternatives
for _, n := range numbers {
    fmt.Printf("%d\n", n)    // Less boxing with format verb
}
// BETTER - use typed map
var cache map[string]int    // No boxing needed
```

**Review guidance**: Flag `fmt.Print`, `fmt.Sprint`, `any`/`interface{}` maps, or variadic `...any` calls inside loops. Suggest typed alternatives.

---

### 5. Linear Search Called in Loop (O(n²))

Flag when a function that performs linear search is called inside a loop.

```go
// BAD - O(n × m) = O(n²) when m ≈ n
func findByID(facts []Fact, id string) *Fact {
    for i := range facts {
        if facts[i].ID == id {
            return &facts[i]
        }
    }
    return nil
}

for _, ref := range references {
    fact := findByID(facts, ref.FactID)  // Linear search per item!
    process(fact)
}

// GOOD - O(n + m) with map
factMap := make(map[string]*Fact, len(facts))
for i := range facts {
    factMap[facts[i].ID] = &facts[i]
}
for _, ref := range references {
    fact := factMap[ref.FactID]  // O(1) lookup
    process(fact)
}
```

**Review guidance**: If a helper function loops to find by ID/key and is called inside another loop, suggest building a map first.

---

### 6. Sort for Min/Max Only (O(n log n) vs O(n))

Flag when code sorts an entire slice just to access the first or last element.

```go
// BAD - O(n log n) to find one element
sort.Slice(facts, func(i, j int) bool {
    return facts[i].CreatedAt.Before(facts[j].CreatedAt)
})
oldest := facts[0]

// GOOD - O(n) linear scan
var oldest Fact
for _, f := range facts {
    if oldest.CreatedAt.IsZero() || f.CreatedAt.Before(oldest.CreatedAt) {
        oldest = f
    }
}

// GOOD - use slices.MinFunc (Go 1.21+)
// Note: MinFunc panics on empty slice, check first
if len(facts) > 0 {
    oldest := slices.MinFunc(facts, func(a, b Fact) int {
        return a.CreatedAt.Compare(b.CreatedAt)
    })
    // use oldest
}
```

**Review guidance**: If only `[0]` or `[len-1]` is accessed after sort, suggest `slices.MinFunc`/`slices.MaxFunc` or linear scan.

---

## Project-Specific Patterns

### No API/Database Calls in Loops

This project uses batched operations. Flag individual calls inside loops:

```go
// BAD - N+1 pattern
for _, fact := range facts {
    embedding, _ := embedder.Embed(ctx, fact.Content)  // API call per item
    vectorDB.Save(ctx, fact)                            // DB call per item
}

// GOOD - batched
embeddings, _ := embedder.EmbedBatch(ctx, texts)
vectorDB.SaveBatch(ctx, facts)
```

**Methods to flag in loops**: `Embed`, `Save`, `Search`, `Delete`, `Extract`, `ExtractFacts`, `CheckConsistency`

---

## Severity Guidelines

| Pattern | Severity | Reason |
|---------|----------|--------|
| Unbounded slice | High | Memory leak in long-running service |
| Goroutine leak | High | Resource exhaustion |
| Linear search in loop | Medium | O(n²) performance |
| Sort for min/max | Medium | Unnecessary O(n log n) |
| Unnecessary copy | Low | Extra memory, rarely critical |
| Interface boxing | Low | Only matters in very hot paths |
