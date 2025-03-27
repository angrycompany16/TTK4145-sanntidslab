package mapfunctions

func DuplicateMap[V any, X comparable](input map[X]V) map[X]V {
	result := make(map[X]V, 0)
	for key, item := range input {
		result[key] = item
	}
	return result
}
