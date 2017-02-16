package hermes

var StructsMap map[string]interface{} = make(map[string]interface{})

func AddStructMap(key string, value interface{}) {
	StructsMap[key] = value
}

// holds the maping between collection type and controller that controlls that collection
var ControllerMap map[interface{}]Controlist = make(map[interface{}]Controlist)

func AddControllerMap(key interface{}, value Controlist) {
	ControllerMap[key] = value
}
