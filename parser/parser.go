package parser

import (
	"fmt"
	"monkey/ast"
	"monkey/lexer"
	"monkey/token"
	"strconv"
)

type Parser struct {
	l              *lexer.Lexer
	curToken       token.Token
	peekToken      token.Token
	errors         []string
	prefixParseFns map[token.TokenType]prefixParseFn
	infixParseFns  map[token.TokenType]infixParseFn
}

type (
	// 前缀
	prefixParseFn func() ast.Expression
	// 中缀
	infixParseFn func(ast.Expression) ast.Expression
)

// 区分优先级
const (
	_ int = iota
	LOWEST
	EQUALS      // == , !=
	LESSGREATER // > or <
	SUM         // +,-
	PRODUCT     // *,/
	PREFIX      // -X or !X
	CALL        // myFunction(X)
	INDEX       //array[index]
)

// 优先级表
var precedence = map[token.TokenType]int{
	token.EQ:       EQUALS,
	token.NOT_EQ:   EQUALS,
	token.LT:       LESSGREATER,
	token.GT:       LESSGREATER,
	token.PLUS:     SUM,
	token.MINUS:    SUM,
	token.SLASH:    PRODUCT,
	token.ASTERISK: PRODUCT,
	token.BANG:     PREFIX,
	token.LPAREN:   CALL,
	token.LBRACKET: INDEX,
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{l: l, errors: []string{}}
	// 普拉特
	//initial prefixParseFns
	p.prefixParseFns = make(map[token.TokenType]prefixParseFn)
	// 注册 解析函数
	p.registerPrefix(token.IDENT, p.parserIdentifier)
	p.registerPrefix(token.INT, p.parserIntegerLiteral)
	p.registerPrefix(token.IF, p.parserIfExpression)
	p.registerPrefix(token.FUNCTION, p.parserFunctionLiter)
	p.registerPrefix(token.STRING, p.parserStringLiteral)
	p.registerPrefix(token.LBRACKET, p.parserArrayLiteral)
	p.registerPrefix(token.LBRACE, p.parserHashLiteral)
	p.registerPrefix(token.MACRO, p.parserMacroLiteral)
	// 前缀解析函数
	p.registerPrefix(token.BANG, p.parserPrefixExpression)
	p.registerPrefix(token.MINUS, p.parserPrefixExpression)

	// 解析boolean
	p.registerPrefix(token.TRUE, p.parserBoolean)
	p.registerPrefix(token.FALSE, p.parserBoolean)

	p.infixParseFns = make(map[token.TokenType]infixParseFn)
	// 中缀表达式解析函数
	p.registerInfix(token.PLUS, p.parserInfixExpression)
	p.registerInfix(token.MINUS, p.parserInfixExpression)
	p.registerInfix(token.SLASH, p.parserInfixExpression)
	p.registerInfix(token.ASTERISK, p.parserInfixExpression)
	p.registerInfix(token.EQ, p.parserInfixExpression)
	p.registerInfix(token.NOT_EQ, p.parserInfixExpression)
	p.registerInfix(token.LT, p.parserInfixExpression)
	p.registerInfix(token.GT, p.parserInfixExpression)
	p.registerInfix(token.LPAREN, p.parserCallExpression)
	p.registerInfix(token.LBRACKET, p.parserIndexExpression)

	// group expression; let a = (1+2)*3;
	p.registerPrefix(token.LPAREN, p.parserGroupExpression)

	p.nextToken()
	p.nextToken()
	return p
}

// register function
func (p *Parser) registerPrefix(tokenType token.TokenType, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}
func (p *Parser) registerInfix(tokenType token.TokenType, fn infixParseFn) {
	p.infixParseFns[tokenType] = fn
}

// 关联解析函数
func (p *Parser) parserIdentifier() ast.Expression {
	ident := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	assignExpression := &ast.AssignExpression{}
	assignExpression.Name = ident
	if p.peekTokenIs(token.ASSIGN) {
		p.nextToken()
		p.nextToken()
		assignExpression.Value = p.parserExpression(LOWEST)
		return assignExpression
	}
	return ident
}

func (p *Parser) parserIntegerLiteral() ast.Expression {
	lit := &ast.IntegerLiteral{Token: p.curToken}
	value, err := strconv.ParseInt(p.curToken.Literal, 0, 64)
	if err != nil { //不等于nil，有错
		msg := fmt.Sprintf("could not parse %s as integer", p.curToken.Literal)
		p.errors = append(p.errors, msg)
		return nil
	}
	lit.Value = value
	return lit
}

func (p *Parser) parserPrefixExpression() ast.Expression {
	prefixExpr := &ast.PrefixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
	}

	p.nextToken()

	prefixExpr.Right = p.parserExpression(PREFIX)
	return prefixExpr
}

func (p *Parser) parserInfixExpression(left ast.Expression) ast.Expression {
	expression := &ast.InfixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
		Left:     left,
	}
	precedence := p.curPrecedence()
	p.nextToken()
	expression.Right = p.parserExpression(precedence)
	// +右关联 | 右结合
	// if expression.Operator == "+" {
	// 	expression.Right = p.parserExpression(precedence - 1)
	// } else {
	// 	expression.Right = p.parserExpression(precedence)
	// }
	return expression
}

func (p *Parser) parserBoolean() ast.Expression {
	return &ast.Boolean{Token: p.curToken, Value: p.curTokenIs(token.TRUE)}
}

func (p *Parser) parserGroupExpression() ast.Expression {
	p.nextToken()

	exp := p.parserExpression(LOWEST)

	if !p.expectPeek(token.RPAREN) {
		return nil
	}
	return exp
}

// if else expression
func (p *Parser) parserIfExpression() ast.Expression {
	expression := &ast.IfExpression{}
	if !p.expectPeek(token.LPAREN) {
		return nil
	}
	p.nextToken()
	expression.Condition = p.parserExpression(LOWEST)
	if !p.expectPeek(token.RPAREN) {
		return nil
	}
	if !p.expectPeek(token.LBRACE) {
		return nil
	}
	expression.Consequence = p.parserBlockStatement()

	if p.peekTokenIs(token.ELSE) {
		p.nextToken() //skip } token
		//skip else token
		if !p.expectPeek(token.LBRACE) {
			return nil
		}
		expression.Alternative = p.parserBlockStatement()
	}
	return expression
}

// function `fn(x, y) { x + y;}`
func (p *Parser) parserFunctionLiter() ast.Expression {
	fnLiteral := &ast.FunctionLiteral{Token: p.curToken}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}
	fnLiteral.Parameters = p.parserFunctionParameters()

	if !p.expectPeek(token.LBRACE) {
		return nil
	}
	fnLiteral.Body = p.parserBlockStatement()
	return fnLiteral
}

func (p *Parser) parserFunctionParameters() []*ast.Identifier {
	identifiers := []*ast.Identifier{}

	if p.peekTokenIs(token.RPAREN) {
		p.nextToken()
		return identifiers
	}
	//else
	p.nextToken() //skip ( curToken = ident->"x"
	ident := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	identifiers = append(identifiers, ident)

	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken() //skip "," ident -> "y"
		ident := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
		identifiers = append(identifiers, ident)
	}
	if !p.expectPeek(token.RPAREN) {
		return nil
	}
	return identifiers
}

func (p *Parser) parserCallExpression(function ast.Expression) ast.Expression {
	exp := &ast.CallExpression{Token: p.curToken, Function: function}
	exp.Arguments = p.parserCallArguments()
	return exp
}

// parserExpressionList 复用
func (p *Parser) parserCallArguments() []ast.Expression {
	args := []ast.Expression{}

	if p.peekTokenIs(token.RPAREN) {
		p.nextToken()
		return args
	}

	p.nextToken()
	args = append(args, p.parserExpression(LOWEST))

	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		args = append(args, p.parserExpression(LOWEST))
	}

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return args
}

func (p *Parser) parserStringLiteral() ast.Expression {
	return &ast.StringLiteral{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) parserArrayLiteral() ast.Expression {
	array := &ast.ArrayLiteral{Token: p.curToken}

	array.Elements = p.parserExpressionList(token.RBRACKET)
	return array
}

func (p *Parser) parserExpressionList(end token.TokenType) []ast.Expression {
	list := []ast.Expression{}

	if p.peekTokenIs(end) {
		p.nextToken()
		return list
	}

	p.nextToken()
	list = append(list, p.parserExpression(LOWEST))
	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		list = append(list, p.parserExpression(LOWEST))
	}

	if !p.expectPeek(end) {
		return nil
	}

	return list
}

func (p *Parser) parserIndexExpression(left ast.Expression) ast.Expression {
	exp := &ast.IndexExpression{Token: p.curToken, Left: left}

	p.nextToken()
	exp.Index = p.parserExpression(LOWEST)

	if !p.expectPeek(token.RBRACKET) {
		return nil
	}

	return exp
}

func (p *Parser) parserHashLiteral() ast.Expression {
	hash := &ast.HashLiteral{Token: p.curToken}
	hash.Pairs = make(map[ast.Expression]ast.Expression)

	for !p.peekTokenIs(token.RBRACE) {
		p.nextToken()
		key := p.parserExpression(LOWEST)

		if !p.expectPeek(token.COLON) {
			return nil
		}
		p.nextToken()
		value := p.parserExpression(LOWEST)
		hash.Pairs[key] = value

		if !p.peekTokenIs(token.RBRACE) && !p.expectPeek(token.COMMA) {
			return nil
		}
	}

	if !p.expectPeek(token.RBRACE) {
		return nil
	}
	return hash
}

func (p *Parser) parserMacroLiteral() ast.Expression {
	macroLit := &ast.MacroLiteral{Token: p.curToken}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}
	macroLit.Parameters = p.parserFunctionParameters()

	if !p.expectPeek(token.LBRACE) {
		return nil
	}
	macroLit.Body = p.parserBlockStatement()

	return macroLit
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

// 语法分析，提示错误
func (p *Parser) Errors() []string {
	return p.errors
}

func (p *Parser) peekError(t token.TokenType) {
	msg := fmt.Sprintf("expected next token to be '%s' got='%s'", t, p.peekToken.Type)
	p.errors = append(p.errors, msg)
}

func (p *Parser) ParserProgram() *ast.Program {
	program := &ast.Program{}
	program.Statements = []ast.Statement{}

	for !p.curTokenIs(token.EOF) {
		stmt := p.parserStatement()
		// if stmt != nil {
		program.Statements = append(program.Statements, stmt)
		// }
		p.nextToken()
	}
	return program
}

// statement
func (p *Parser) parserStatement() ast.Statement {
	switch p.curToken.Type {
	case token.LET:
		return p.parserLetStatement()
	case token.RETURN:
		return p.parserReturnStatement()
	case token.WHILE:
		return p.parserWhileStatement()
	default:
		/* !!5 | !!true | !！false */
		// if p.curToken.Literal == "!" && p.peekTokenIs(token.BANG) {
		// 	p.nextToken()
		// 	return p.parserExpressionStatement()
		// }
		return p.parserExpressionStatement()
	}
}

func (p *Parser) parserLetStatement() *ast.LetStatement {
	stmt := &ast.LetStatement{Token: p.curToken}
	if !p.expectPeek(token.IDENT) {
		return nil
	}
	// if strings.Contains(p.curToken.Literal,"1") {return nil}
	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	if !p.expectPeek(token.ASSIGN) {
		return nil
	}

	p.nextToken()

	stmt.Value = p.parserExpression(LOWEST)
	if fn, ok := stmt.Value.(*ast.FunctionLiteral); ok {
		fn.Name = stmt.Name.Value
	}

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parserReturnStatement() *ast.ReturnStatement {
	returnStmt := &ast.ReturnStatement{Token: p.curToken}
	p.nextToken()

	returnStmt.ReturnValue = p.parserExpression(LOWEST)
	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return returnStmt
}

func (p *Parser) parserWhileStatement() *ast.WhileStatement {
	whileStmt := &ast.WhileStatement{Token: p.curToken}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}
	p.nextToken()

	whileStmt.Condition = p.parserExpression(LOWEST)

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	whileStmt.Body = *p.parserBlockStatement()

	return whileStmt
}

func (p *Parser) parserExpressionStatement() *ast.ExpressionStatement {
	stmt := &ast.ExpressionStatement{Token: p.curToken}
	stmt.Expression = p.parserExpression(LOWEST)

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}
	return stmt
}

func (p *Parser) parserBlockStatement() *ast.BlockStatement {
	block := &ast.BlockStatement{Token: p.curToken}
	block.Statements = []ast.Statement{}

	p.nextToken()
	for !p.curTokenIs(token.RBRACE) && !p.curTokenIs(token.EOF) {
		stmt := p.parserStatement()
		// if stmt != nil {
		block.Statements = append(block.Statements, stmt)
		// }
		p.nextToken()
	}
	return block
}

// 沃恩·普拉特优先级解析法
func (p *Parser) parserExpression(precedence int) ast.Expression {
	prefix := p.prefixParseFns[p.curToken.Type]
	if prefix == nil {
		p.noPrefixFnError(p.curToken.Type)
		return nil
	}
	leftExp := prefix()

	for !p.peekTokenIs(token.SEMICOLON) && precedence < p.peekPrecedence() {
		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil {
			return leftExp
		}
		p.nextToken()
		leftExp = infix(leftExp)
	}
	return leftExp
}

func (p *Parser) expectPeek(t token.TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	} else {
		p.peekError(t)
		return false
	}
}

func (p *Parser) peekTokenIs(t token.TokenType) bool {
	return p.peekToken.Type == t
}

func (p *Parser) curTokenIs(t token.TokenType) bool {
	return p.curToken.Type == t
}

// 没找到解析函数时错误
func (p *Parser) noPrefixFnError(t token.TokenType) {
	msg := fmt.Sprintf("prefix parse function for %s not found", t)
	p.errors = append(p.errors, msg)
}

func (p *Parser) peekPrecedence() int {
	if p, ok := precedence[p.peekToken.Type]; ok {
		return p
	}
	return LOWEST
}

func (p *Parser) curPrecedence() int {
	if p, ok := precedence[p.curToken.Type]; ok {
		return p
	}
	return LOWEST
}
