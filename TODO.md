# TODO

Implementation backlog, grouped by language feature. Each item explains *what's
wrong* and *why*, so the fix is mine to write. Severity: 🔴 valid programs
miscompile · 🟡 narrower correctness gap · 🟢 cosmetic / harmless.

---

## Functions

Functions work for the common case — declarations with params and a return type,
calls, recursion, void functions, `main` enforcement, and returning any built-in
scalar (`uint8/16/32/64`, `bool`) or a pointer. The items below are what stands
between that and "done".

### Done

- [x] **Early `return;` in a void function was a no-op.** `ReturnStmt.EmitCode`
  only emitted `leave; ret` when the return had a value, so a bare `return;` in
  the middle of a void function fell through instead of returning. Now emits
  `leave; ret` unconditionally.

### 🔴 Blocks real programs

- [ ] **Parameter order is nondeterministic.** `semantics.Func.Params` is a
  `map[string]Type`. Both register assignment (`FuncDeclaration.Semantics`
  building `stmt.Params`) and argument type-checking iterate that map, and Go
  randomizes map iteration order — so the *same source* can compile to a
  *different binary* each run. Confirmed: `sub(a, b)` failed 1 in 8 identical
  recompiles.
  **Direction:** make params an ordered slice (e.g. `[]Param{Ident, Type}`)
  instead of a map; keep a map only if you also need name lookup.

- [ ] **Literal arguments are rejected.** `add(40, 2)` fails with *"Argument 0
  expected type UINT32 received UINT_LIT"*. `FuncCallExpression.Semantics`
  compares arg types with `Equals`, and an untyped `UINT_LIT` never equals a
  concrete width — even though variable initialization *does* adapt literals to
  the target width.
  **Direction:** in the arg-check loop, apply the same untyped-literal → concrete
  width adaptation that var-init already uses, before comparing.

- [ ] **The 6th argument is silently dropped.** The declaration side spills up to
  6 params (`rdi, rsi, rdx, rcx, r8, r9`), but the call side only fills 5:
  `argsLen := min(len(exp.Args), 5)` in `FuncCallExpression.EmitCode`. Confirmed:
  `sum6(...)` computes the wrong answer.
  **Direction:** change the call-side limit to 6 so it matches the declaration
  side.

- [ ] **Arguments 7+ (stack-passed) are unimplemented.** In-code TODO at
  `parser/node.go` (`FuncDeclaration.EmitCode`): *"Implement reading rest of the
  parameters from the stack"*. Beyond 6 args, neither the caller (push extra args)
  nor the callee (read them back from the stack) is wired up.
  **Direction:** caller pushes args 7+ right-to-left before the register args;
  callee reads them at positive `[rbp + N]` offsets (above the saved rbp / return
  address). Mind 16-byte alignment.

- [ ] **Forward references fail.** A function can only call functions declared
  *earlier* in the file — `main` calling a later `add` reports *"Undeclared
  symbol 'add'"*. Each function's symbol is only registered *during its own*
  `Semantics` pass, in source order, so a later callee doesn't exist yet.
  **Direction:** a hoisting pre-pass — walk all top-level statements and register
  every function *signature* in the symbol table before analyzing any function
  *body*.

- [ ] **Mutual recursion fails.** `isEven`/`isOdd` calling each other fails
  regardless of declaration order. Same root cause as forward references — fixed
  by the same hoisting pre-pass. (Self-recursion already works because a
  function's own symbol is pushed before its body is analyzed.)

- [ ] **Complex argument expressions clobber argument registers.** Args are
  evaluated and moved into `rdi/rsi/…` one at a time in
  `FuncCallExpression.EmitCode`. Evaluating a *later* arg can overwrite an
  *earlier* arg's register — any nested call sets `rdi`, so `sub(x, id(y))`
  silently passes the wrong values (`id`'s `rdi` overwrites `sub`'s). Confirmed:
  computes `3 - 3 = 0` instead of `10 - 3 = 7`.
  **Direction:** stop routing args register-by-register inline. Evaluate each arg
  and `push rax`, then `pop` into the target registers just before the `call`
  (or evaluate right-to-left). This makes any arg expression safe, not just
  simple loads.

### 🟡 Correctness gaps

- [ ] **Return-path analysis is shallow.** `FuncDeclaration.Semantics` only
  accepts a non-void function whose *last body statement* is literally a
  `ReturnStmt`. It rejects valid code where every branch returns but there's no
  trailing top-level return (e.g. `if c { return a; } else { return b; }`), and
  it doesn't actually verify that *all* paths return.
  **Direction:** replace the last-statement check with a small "does this
  statement always return?" recursion (a block returns if its last returning
  statement does; an if/else returns if both branches do).

- [ ] **Errors inside function-body blocks are swallowed.** `BlockStmt.Semantics`
  discards inner-statement errors and always returns `nil`, so a type error
  inside a `{ }` in a function body is silent. (Broader than functions, but it
  lands here too.)
  **Direction:** collect/propagate inner errors instead of ignoring the return
  value of `innerStmt.Semantics(s)`.

### 🟢 Cosmetic / harmless

- [ ] **`Align16` over-pads.** Returns `16` when `nextAddr` is already 16-aligned
  (`16 - (nextAddr % 16)` with `remainder == 0`), so an already-aligned call site
  still gets a full 16 bytes of dead padding. Wasteful, not wrong.
  **Direction:** return `0` when the remainder is `0`.

- [ ] **Duplicate epilogue.** A void function ending in `return;` now emits
  `leave; ret` twice — once from the return, once from the implicit epilogue
  `FuncDeclaration.EmitCode` always appends for void functions. The second is
  unreachable.
  **Direction:** only append the implicit epilogue when the body doesn't already
  end in a return.

- [ ] **Indirect call for a static callee.** `lea rax, [rel name]; call rax`
  where a direct `call name` would do — one extra instruction per call. A design
  choice (treats the callee as any expression yielding a function address), not a
  bug.

- [ ] **Stale sample.** `tests/misc/test1.clov` is `fn main() -> int {}`, but
  `int` isn't a real type, so the sample doesn't compile.
