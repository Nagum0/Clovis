# BNF GRAMMAR

``` py
<program> ::= <statements> EOF

<statements> ::= { <statement> }
<statement> ::= <varDecl> |
                <varDefinition> |
                <blockStmt> |
                <ifStmt> |
                <whileStmt> |
                <forStmt> |
                <assert> |
                <expressionStmt>
<varDecl> ::= ( "uint" | "bool" ) ident ";" | 
              ( "uint" | "bool" ) ident "=" <expression> ";"
<varDefinition> ::= ident "=" <expression> ";"
<blockStmt> ::= "{" <statements> "}"
<ifStmt> ::= "if" <expression> <statement>
<whileStmt> ::= "while" <expression> <statement>
<forStmt> ::= "for" ident "=" <expression> ".." <expression> <statement> |
              "for" ident "=" <expression> ".." <expression> <expression> <statement>
<assert> ::= "assert" <expression> ";"
<expressionStmt> ::= <expression> ";"

<expression> ::= <equality>
<equality> ::= <comparison> { ("==" | "!=") <comparison> }
<comparison> ::= <term> { ("<" | "<=" | ">" | ">=") <term> }
<term> ::= <factor> { ("+" | "-") <factor> }
<factor> ::= <unary> { ("*" | "/") <unary> }
<unary> ::= ( "!" | "-" ) <unary> | <primary>
<primary> ::= <literal> | ident | <groupExpr> | <functionCall>
<groupExpr> ::= "(" <expression> ")"
<functionCall> ::= ident "(" [ <param> { "," <param> } ] ")"
<param> ::= <expression>
```
