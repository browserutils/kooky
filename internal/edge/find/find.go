package find

func GetEdgeRoots() (rootsFunc func() ([]string, error)) {
	return edgeRoots
}
