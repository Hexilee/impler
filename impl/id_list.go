package impl

type (
	IdList map[string]bool
)

func (list IdList) addKey(key string) (added bool) {
	if _, ok := list[key]; !ok {
		list[key] = true
		added = true
	}
	return
}

func (list IdList) deleteKey(key string) (exist bool) {
	if _, ok := list[key]; ok {
		delete(list, key)
		exist = true
	}
	return
}
