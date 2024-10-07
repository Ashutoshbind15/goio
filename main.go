package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type Todo struct {
	Id    string
	Title string
	Done  bool
}

type TodoFile struct {
	Todos []Todo
}

// JSON file server 
func readfile(path string) []byte {
	data, err := os.ReadFile(path)

	if err != nil {
		panic(err)
	}

	return data
} 

func convertTodos(bts []byte) []Todo {
	var jsonData TodoFile	

	json.Unmarshal(bts, &jsonData)
	fmt.Println(jsonData.Todos)

	return jsonData.Todos
}

func addTodo(todo Todo) {
	filebytes := readfile("todos.json")

	todoList := convertTodos(filebytes)

	todoList = append(todoList, todo)

	jsonf := TodoFile{
		Todos: todoList,
	}

	todoListBytes, err := json.Marshal(jsonf)

	if err != nil {
		panic(err)
	}

	os.WriteFile("todos1.json", todoListBytes, os.ModeAppend)
}

func initHandler(w http.ResponseWriter , req *http.Request) {
	if req.Method == "GET" {
		todos := readfile("todos.json")
		fmt.Fprintln(w, string(todos))
	} else {
		body := req.Body
		data, err := io.ReadAll(body)
		if err != nil {
			panic(err)
		}

		var todoBody Todo

		json.Unmarshal(data, &todoBody)

		fmt.Println(todoBody)

		// make a db call or something

		addTodo(todoBody)

		fmt.Fprintf(w, "done")
	}
}


func middleFunc(nxt http.HandlerFunc) http.HandlerFunc {

	hf := func(w http.ResponseWriter, r *http.Request) {
        fmt.Println("running the mware")
        nxt.ServeHTTP(w, r)
    }

	return hf
}

func hellofunc(w http.ResponseWriter, rq *http.Request) {
	fmt.Println("main func logic")
	fmt.Fprintf(w, "return")
}


func main() {

	// godotenv.Load()
	// dbclient := data.ConnectDb()

	// defer func() {
	// 	if err := dbclient.Disconnect(context.TODO()); err != nil {
	// 		panic(err)
	// 	}
	// }()


	// demoDbCall(dbclient)

	loggedHello := middleFunc(hellofunc)


	http.HandleFunc("/", loggedHello)
	http.ListenAndServe(":3000", nil)
}