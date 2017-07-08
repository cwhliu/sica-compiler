package forge

type graph struct {
	allNodes       map[string]*node
	inputNodes     map[string]*node
	outputNodes    map[string]*node
	internalNodes  map[string]*node
	operationNodes map[string]*node
	constantNodes  map[string]*node
}
