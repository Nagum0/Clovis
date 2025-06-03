# BNF GRAMMAR

``` py
<program> ::= <statements> EOF

<statements> ::= { <statement> }
<statement> ::= <varDecl> |
                <varDefinition> |
                <blockStmt> |
                <ifStmt> |
                <whileStmt> |
                <forStmt>
<varDecl> ::= ( "uint" | "bool" ) ident ";" | 
              ( "uint" | "bool" ) ident "=" <expression> ";"
<varDefinition> ::= ident "=" <expression> ";"
<blockStmt> ::= "{" <statements> "}"
<ifStmt> ::= "if" <expression> <statement>
<whileStmt> ::= "while" <expression> <statement>
<forStmt> ::= "for" ident "=" <expression> ".." <expression> <statement> |
              "for" ident "=" <expression> ".." <expression> <expression> <statement>

<expression> ::= <equality>
<equality> ::= <comparison> { ("==" | "!=") <comparison> }
<comparison> ::= <term> { ("<" | "<=" | ">" | ">=") <term> }
<term> ::= <factor> { ("+" | "-") <factor> }
<factor> ::= <unary> { ("*" | "/") <unary> }
<unary> ::= ( "!" | "-" ) <unary> | <primary>
<primary> ::= <literal> | ident | "(" <expression> ")" 
```
