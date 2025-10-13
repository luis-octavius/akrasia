package todo 

type Todo struct {
	ID int 
	Name string 
	Status bool 
}

func AddTodo(id int, name string, status bool) (Todo) {
	newTodo := Todo{
		ID: id, 
		Name: string, 
		Status: status 
	}

	return newTodo 
}

func ListTodo(m map[int]Todo) (string) {
	result := ``

	for _, todo := range m {
		result += 
	}
}
