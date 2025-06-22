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
                <expressionStmt> |
                <typeDeclaration>
<varDecl> ::= typeId ( "*" ) ident ";" | 
              typeId ( "*" ) ident "=" <expression> ";"
<varDefinition> ::= <lvalue> < "=" <expression> ";"
<blockStmt> ::= "{" <statements> "}"
<ifStmt> ::= "if" <expression> <statement> ( "else" <statement> )
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
<unary> ::= ( "!" | "-" | "*" | "&" ) <unary> | 
            <postfix>
<postfix> ::= <primary> { ( "++" | "--" | <arrayAccess> | <funcCall> }
<primary> ::= <literal> | <ident> | <groupExpr>
<arrayAccess> := "[" <expression> "]"
<funcCall> ::= TODO
<groupExpr> ::= "(" <expression> ")"
```
