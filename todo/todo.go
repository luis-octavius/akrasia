package todo 

import "time"

type Todo struct {
	ID 					int 
	Description string 
	Completed 	bool
	CreatedAt 	time.Time 
}

type TodoManager struct {
	todos  map[int]*Todo 
	nextID int 
}

func NewTodoManager() *TodoManager {
	return &TodoManager{
		todos: make(map[int]*Todo),
		nextID: 1,
	}
}

func (tm *TodoManager) AddTodo(description string) int {
	id := tm.nextID
	tm.todos[id] = &Todo{
		ID:						id, 
		Description:  description,
		Completed: 		false,
		CreatedAt: 		time.Now(),
	}
	tm.nextID++ 
	return id
}

func (tm *TodoManager) GetTodo(id int) (*Todo, bool) {
	todo, exists := tm.todos[id]
	return todo, exists 
}

func (tm *TodoManager) GetAllTodos() []*Todo {
	todos := make([]*Todo, 0, len(tm.todos))
	for _, todo := range tm.todos {
		todos = append(todos, todo)
	}
	return todos 
}


