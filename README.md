# CLOVIS

A simple compiler built to learn about compilers and assembly.

## Documentation

### Variables

#### Variable declaration

- Simple 32 bit unsigned integer type declaration.
``` go
uint32 x = 67;
```

- Separate declaration and definition is also supported.
``` go
uint32 x;
x = 420;
```

#### Array declaration

- An array of unsigned 8 bit integers (bytes).
- Currently array literals are not supported.
``` go
uint8[5] bytes;
```

#### Pointer declaration

- A pointer to x.
- All pointers are 64 bit unsigned integers.
``` go
uint32* x_ptr = &x;
```

#### Blocks

``` go
{
    {
        uint32 x = 67;
    }
    // x will be unaccessable here
}
```

- Shadowing
``` go
{
    uint32 x = 67;
    {
        uint32 x = 420;
        // x will be 420 here
    }
}
```

### Control flow

#### If statement

``` go
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

#### While loop

- While loops take a condition which must be of type bool and a
  lopp body.
``` py
uint64 i = 0;
while i < 10 {
    i = i + 1;
}
assert i == 10;
```

#### Assert statement

- If assert condition is true the execution continues,
  if it's not the program aborts.

``` go
uint32 x = 67;
assert x == 67;
```

### Arrays in depth

- Arrays are treated as pointers to their first element.
  However, they are NOT equivalent or can be used as pointers.
``` go
uint8[5] bytes; // bytes is a pointer to the first element
```

- Deep copy
``` go
uint32[3] xs;
xs[0] = 1;
xs[1] = 2;
xs[2] = 3;

// Here the values of xs are copied over byte by byte into ys.
uint32[3] ys = xs;
ys[1] = 67;
assert xs[1] != 67;
```

- Shallow copy
  - You need to create a pointer to the array you want to shallow copy.
``` go
uint32[3]* zs = &xs;
(*zs)[1] = 69;
assert xs[1] == 69;
assert (*zs)[1] == 69;
```

### Pointers in depth

#### Referencing

- Referencing can be achieved with the & unary operator used on a identifier.
``` go
uint32 x = 67;
uint32* x_ptr = &x;
```

#### Dereferencing

- Dereferencing can be achieved by using the * unary operator in a pointer type.
``` go
uint32 x = 67;
uint32* x_ptr = &x;
assert *x_ptr == 67;
```

### Functions

