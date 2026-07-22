# Clovis — Language Features

Status of the language as currently implemented. Every example below was compiled and run against the compiler; the ones marked ❌ genuinely fail today.

A program is a sequence of statements. Execution starts at the first statement, and the program exits 0 when it falls off the end. A failed `assert` exits 1.

---

## Comments

**Not supported.** There is no comment syntax — the lexer has no rule for `//`, `#`, or `/* */`.

```go
uint32 x = 5; // this breaks the parse
```

> Note: `README.md` uses `//` comments in several examples. Those examples do not compile as written.

---

## Types

| Type | Size | Literal form | Notes |
|---|---|---|---|
| `uint8` | 1 byte | `0`–`255` | |
| `uint16` | 2 bytes | | |
| `uint32` | 4 bytes | | |
| `uint64` | 8 bytes | | |
| `bool` | 1 byte | `true`, `false` | |
| `T*` | 8 bytes | — | pointer, any depth (`uint32**`) |
| `T[N]` | `N * sizeof(T)` | — | array, no literal syntax |

Integer literals are untyped (`UINT_LIT`) until assigned, so they adapt to any unsigned width.

**There are no implicit conversions between widths.** Mixing them is a compile error:

```go
uint32 a = 5;
uint64 b = 5;
assert a == b;   // ❌ Cannot use operator '==' between types UINT32 and UINT64
```

---

## Variables

Declaration with initializer, or declaration then assignment:

```go
uint32 x = 67;

uint32 y;
y = 420;
assert y == 420;
```

Redeclaration in the same scope is an error; shadowing in a nested scope is allowed (see [Blocks](#blocks)).

Uninitialized variables are **not** zeroed — a declaration only moves the stack pointer, so reading one before assigning gives whatever was on the stack.

---

## Operators

### Arithmetic — `+` `-` `*` `/`

```go
uint32 a = 40;
assert a + 2 == 42;
assert a - 8 == 32;
assert a * 2 == 80;
assert a / 4 == 10;
```

Integer division truncates. There is no `%`.

### Comparison — `==` `!=` `<` `<=` `>`

```go
uint32 x = 5;
assert x == 5;
assert x != 6;
assert x < 9;
assert x <= 5;
assert x > 1;
```

All comparisons produce `bool`.

`>=` **is broken** — the lexer emits `>` followed by a stray `=`, so it always fails to parse:

```go
assert x >= 5;   // ❌ Invalid expression
```

Use `x > 4` or `x == 5` until [the lexer bug](#known-bugs) is fixed.

### Grouping

```go
uint32 x = (2 + 3) * 4;
assert x == 20;
```

### Not implemented: `!`, unary `-`, `++`, `--`

The tokens lex and parse, but `PrefixExpression`/`PostfixExpression` have empty `Semantics`/`EmitCode`, so the expression comes back as type `UNDEFINED` and gets rejected:

```go
bool b = false;
assert !b;       // ❌ Assert statement expects a boolean expression

uint32 y = -x;   // ❌ type UINT32 and right side type UNDEFINED do not match
```

---

## Blocks

A block introduces a scope. Symbols declared inside are popped at the closing brace and the stack space is reclaimed.

```go
{
    uint32 inner = 5;
    assert inner == 5;
}
// inner is not visible here
```

Shadowing works — lookup scans the symbol table backwards, so the innermost declaration wins:

```go
uint32 x = 67;
{
    uint32 x = 420;
    assert x == 420;
}
assert x == 67;
```

---

## Control flow

### `if` / `else if` / `else`

The condition must be exactly `bool` — there is no truthiness.

```go
uint32 x = 66;
uint32 y;

if x == 67 {
    y = 69;
}
else if x == 66 {
    y = 67;
}
else {
    y = 420;
}

assert y == 67;
```

### `while`

```go
uint64 i = 0;
while i < 10 {
    i = i + 1;
}
assert i == 10;
```

### `for` — not implemented

The grammar is in `bnf.md` and `for` is a reserved keyword, but `parseForStmt` has an empty body. Using `for` **hangs the compiler in an infinite loop** rather than reporting an error:

```go
for i = 0 .. 10 { }   // ❌ compiler never terminates
```

---

## `assert`

Takes a `bool`. If it holds, execution continues; otherwise the program exits 1 immediately via a raw `exit` syscall.

```go
uint32 x = 67;
assert x == 67;
```

This is the only observable output the language has — there is no I/O, so `assert` plus the process exit code is how you test a program.

---

## Arrays

Declared with a constant length. **There is no array literal**, so elements are assigned individually:

```go
uint32[3] xs;
xs[0] = 1;
xs[1] = 2;
xs[2] = 3;
assert xs[1] == 2;
```

An array name refers to the address of its first element, but arrays are *not* interchangeable with pointers.

### Deep copy

Assigning one array to another copies element bytes with `rep movsb`, so the two are independent:

```go
uint32[3] xs;
xs[0] = 1;
xs[1] = 2;
xs[2] = 3;

uint32[3] ys = xs;
ys[1] = 67;
assert xs[1] != 67;
assert ys[1] == 67;
```

### Shallow copy via pointer

Take a pointer to the array to alias it instead:

```go
uint32[3] xs;
xs[1] = 2;

uint32[3]* zs = &xs;
(*zs)[1] = 69;
assert xs[1] == 69;
assert (*zs)[1] == 69;
```

**Array lengths are not checked on assignment** — only the element type is compared, so `uint32[5] ys = xs;` where `xs` is `uint32[3]` compiles and copies 20 bytes out of a 12-byte source.

---

## Pointers

All pointers are 8 bytes. `&` takes an address, `*` dereferences; both read and write through the pointer work.

```go
uint32 x = 67;
uint32* p = &x;
assert *p == 67;

*p = 42;
assert x == 42;
```

Pointers nest to any depth:

```go
uint32 x = 67;
uint32* p = &x;
uint32** pp = &p;
assert **pp == 67;
```

There is no pointer arithmetic and no null.

---

## Known bugs

Verified against the current build. Listed because they produce confusing failures, not as a work list.

| # | Symptom | Cause |
|---|---|---|
| 1 | `>=` never parses | Lexer sees `>`, peeks `=`, but never consumes it (`lexer/lexer.go`, the `'>'` branch) |
| 2 | `x<9` fails; `x < 9` works | The `<` branch consumes an extra character when it isn't `<=`, eating whatever follows |
| 3 | Any unrecognized character (`@`, `,`, `.`) hangs the compiler | The lexer's `if/else if` chain has no final `else`, so `idx` never advances. `LexerError` is defined but never constructed |
| 4 | `for` hangs the compiler | `parseForStmt` is empty and consumes no tokens |
| 5 | Division after an overflowing multiply crashes with SIGFPE | `mul` leaves the high half in `rdx`; `div` is emitted without `xor rdx, rdx` first |
| 6 | Sub-64-bit values compare against stale high bits | Loads emit `mov al, BYTE [...]` with no zero-extension, then compare the full `rax`. `uint8 a = 300;` stores 44, yet **both** `assert a == 44` and `assert a == 300` pass |
| 7 | A semantic error inside a block is silently ignored | `BlockStmt.Semantics` discards inner errors and returns `nil`; `main.go` never reads `SemanticChecker.Errors`. `{ uint32 x = true; }` exits 0 and emits `mov DWORD [rbp - 0], eax` |

Bug 7 also means bugs 5 and 6 and the unimplemented operators fail *loudly* at top level but *silently* inside any block.

---

## Quick reference

```
types        uint8 uint16 uint32 uint64 bool  T*  T[N]
literals     123   true   false
operators    +  -  *  /  ==  !=  <  <=  >     &  *  [ ]  ( )
statements   <decl>  <assign>  { }  if/else  while  assert  <expr>;
reserved     if else while for uint64 uint32 uint16 uint8 bool true false assert
missing      comments  >=  !  unary -  ++  --  for  array literals  %  I/O
```
