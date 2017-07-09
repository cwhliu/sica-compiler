package forge

type parserStack struct {
	token    []string
	popCount []int
	argCount []int
}

// Push a non-leaf token onto stack
func (s *parserStack) pushNonLeaf(token string, count int) {
	s.token = append(s.token, token)

	s.popCount = append(s.popCount, count)
	s.argCount = append(s.argCount, count)
}

// Push a leaf token onto stack
func (s *parserStack) pushLeaf(token string) {
	// Decrement previous argument count by one
	count := s.argCount[len(s.argCount)-1] - 1

	s.token = append(s.token, token)

	s.argCount = append(s.argCount, count)
}
