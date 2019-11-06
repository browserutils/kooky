package ordereddict

import "strings"

func GetString(event_map *Dict, members string) (string, bool) {
	value, pres := GetAny(event_map, members)
	if pres {
		switch t := value.(type) {
		case string:
			return t, true
		case *string:
			return *t, true
		}
	}

	return "", false
}

func GetMap(event_map *Dict, members string) (*Dict, bool) {
	value, pres := GetAny(event_map, members)
	if pres {
		switch t := value.(type) {
		case *Dict:
			return t, true
		}
	}
	return nil, false
}

func GetAny(event_map *Dict, members string) (interface{}, bool) {
	var value interface{} = event_map
	var pres bool

	for _, member := range strings.Split(members, ".") {
		if event_map == nil {
			return nil, false
		}

		value, pres = event_map.Get(member)
		if !pres {
			return nil, false
		}
		event_map, pres = value.(*Dict)
	}

	return value, true
}

func GetInt(event_map *Dict, members string) (int, bool) {
	value, pres := GetAny(event_map, members)
	if pres {
		switch t := value.(type) {
		case int:
			return t, true
		case uint8:
			return int(t), true
		case uint16:
			return int(t), true
		case uint32:
			return int(t), true
		case uint64:
			return int(t), true
		case int8:
			return int(t), true
		case int16:
			return int(t), true
		case int32:
			return int(t), true
		case int64:
			return int(t), true
		}
	}

	return 0, false
}
