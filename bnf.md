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
                <funcDeclaration> |
                <typeDeclaration>
<varDecl> ::= <typeID> { "*" | "[" UINT_LIT "]" } IDENT ( ";" | "=" <expression> ";" )
<varDefinition> ::= <lvalue> < "=" <expression> ";"
<blockStmt> ::= "{" <statements> "}"
<ifStmt> ::= "if" <expression> <statement> ( "else" <statement> )
<whileStmt> ::= "while" <expression> <statement>
<forStmt> ::= "for" ident "=" <expression> ".." <expression> <statement> |
              "for" ident "=" <expression> ".." <expression> <expression> <statement>
<assert> ::= "assert" <expression> ";"
<expressionStmt> ::= <expression> ";"

<funcDeclaration> ::= "fn" <ident> "(" [ <params> ] ")" [ "->" <typeID> ] "{" <statements> "}"
<params> ::= <param> { "," <param> }
<param> ::= <typeID> <ident>

<expression> ::= <equality>
<equality> ::= <comparison> { ("==" | "!=") <comparison> }
<comparison> ::= <term> { ("<" | "<=" | ">" | ">=") <term> }
<term> ::= <factor> { ("+" | "-") <factor> }
<factor> ::= <prefix> { ("*" | "/") <prefix> }
<prefix> ::= ( "!" | "-" | "*" | "&" ) <prefix> | 
            <postfix>
<postfix> ::= <primary> { ( "++" | "--" | <arrayAccess> | "(" [ <args> ] ")" ) }
<primary> ::= <literal> | <ident> | <groupExpr>
<arrayAccess> := "[" <expression> "]"
<args> ::= <args> { "," <arg> }
<arg> ::= <expression>
<groupExpr> ::= "(" <expression> ")"
```
