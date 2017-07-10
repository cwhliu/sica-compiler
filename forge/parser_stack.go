package forge

type parserStack struct {
	tokens         []string
	tokenPopCounts []int
	tokenArgCounts []int

	results []*node
}

// -----------------------------------------------------------------------------

/*
Check if a token is ready to be popped

A token is ready when the top of argument count stack is 0, meaning that
the token has all the arguments it needs
*/
func (s *parserStack) tokenReady() bool {
	return len(s.tokenArgCounts) > 0 && s.tokenArgCounts[len(s.tokenArgCounts)-1] == 0
}

/*
Push a non-leaf token to the stack
*/
func (s *parserStack) pushNonLeafToken(token string, argCount int) {
	s.tokens = append(s.tokens, token)

	// Add one to include the token itself
	s.tokenPopCounts = append(s.tokenPopCounts, argCount+1)
	s.tokenArgCounts = append(s.tokenArgCounts, argCount)
}

/*
Push a leaf token to the stack

A leaf token is the argument of a predecessor non-leaf token
*/
func (s *parserStack) pushLeafToken(token string) {
	s.tokens = append(s.tokens, token)

	// Decrement previous argument count by one
	count := s.tokenArgCounts[len(s.tokenArgCounts)-1] - 1
	s.tokenArgCounts = append(s.tokenArgCounts, count)
}

/*
Pop the top token and its arguments from the stack
*/
func (s *parserStack) popToken() (string, []string) {
	popCount := s.tokenPopCounts[len(s.tokenPopCounts)-1]

	token := s.tokens[len(s.tokens)-popCount]
	args := make([]string, popCount-1)
	copy(args, s.tokens[len(s.tokens)-popCount+1:len(s.tokens)])

	s.tokens = s.tokens[:len(s.tokens)-popCount]
	s.tokenPopCounts = s.tokenPopCounts[:len(s.tokenPopCounts)-1]
	s.tokenArgCounts = s.tokenArgCounts[:len(s.tokenArgCounts)-popCount]

	return token, args
}

// -----------------------------------------------------------------------------

/*
Push an intermediate result to the stack
*/
func (s *parserStack) pushResult(nd *node) {
	s.pushLeafToken("RESULT")

	s.results = append(s.results, nd)
}

/*
Pop the top intermediate result from the stack
*/
func (s *parserStack) popResult() *node {
	defer func() { s.results = s.results[:len(s.results)-1] }()

	return s.results[len(s.results)-1]
}
