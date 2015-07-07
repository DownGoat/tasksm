package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"time"
)

type Task struct {
	Name      string
	Body      string
	Created   string
	Due       string
	Completed bool
	Status    string
	Id        int32
	Priority  int32
	SubTasks  []Task
}

type Tasks struct {
	NewestId int32
	TaskList []Task
}

const BODY_LIMIT int = 50

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func open_file(path string) []byte {
	file_content, err := ioutil.ReadFile(path)
	check(err)

	return file_content
}

func write_file(path string, content []byte) {
	err := ioutil.WriteFile(path, content, 0644)
	check(err)
}

func print_sub_task(task Task, indent int, postfix string) {
	for i := 0; i < indent; i++ {
		fmt.Print("\u2551\t\t")
	}

	task_status := task.Status
	if task.Status == "t" {
		task_status = "\u2713"
	}

	fmt.Println(postfix, task.Id, ": ", task.Priority, ": ", task_status, ": ", task.Name)

	sub_pos := "\u2560"
	for idx, sub := range task.SubTasks {
		if idx == len(task.SubTasks)-1 {
			sub_pos = "\u255a"
		}
		print_sub_task(sub, indent+1, sub_pos)
	}
}

func print_tasks(storage_file Tasks) {
	fmt.Println("ID  : Priority :  Status  :  Name")
	postfix := "\u2554"
	for idx, task := range storage_file.TaskList {
		if idx == (len(storage_file.TaskList) - 1) {
			postfix = "\u255a"
		}

		status := task.Status

		if task.Completed == true {
			status = "\u2713"
		}

		fmt.Println(postfix, task.Id, ": ", task.Priority, "      : ", status, "      : ", task.Name)

		for _, sub := range task.SubTasks {
			print_sub_task(sub, 1, "\u2560")
		}

		postfix = "\u2560"
	}
}

func nice_body(body string) string {
	last_space := 0
	nice_str := ""
	offset := 0
	limit := 0
	for idx, c := range body {
		if string(c) == " " {
			last_space = idx
		}

		if idx != 0 && idx%BODY_LIMIT == 0 {
			if string(c) == " " {
				nice_str += body[offset:limit+1] + "\n"
				offset = idx + 1
			} else {
				nice_str += body[offset:last_space] + "\n"
				offset = last_space + 1
			}

		}

		limit = idx
	}

	nice_str += body[offset : limit+1]

	return nice_str
}

func print_task(storage_file Tasks, id int) {
	for _, task := range storage_file.TaskList {
		if task.Id == int32(id) {
			fmt.Println("Name:    ", task.Name, "")
			fmt.Println("Created: ", task.Created)
			fmt.Println("Due:     ", task.Due)

			if task.Completed {
				fmt.Println("Status:  \u2713 Completed")
			} else {
				fmt.Println("Status:  ", task.Status)
			}

			fmt.Println("Priority:", task.Priority)
			fmt.Print("Sub Tasks:")

			for sub, _ := range task.SubTasks {
				fmt.Print(sub)
			}

			fmt.Print("\n")

			if len(task.Body) < BODY_LIMIT {
				fmt.Println("Body:    ", task.Body)
			} else {
				fmt.Println("Body:\n")
				fmt.Println(nice_body(task.Body))
			}
		}
	}
}

func change_status(storage_file *Tasks, status string, id int) {
	task := search_for_id(storage_file.TaskList, id)
	if task == nil {
		fmt.Println("Could not find task")
		os.Exit(1)
	}

	task.Status = status
}

func change_completion(storage_file *Tasks, completed string, id int) {
	task := search_for_id(storage_file.TaskList, id)
	if task == nil {
		fmt.Println("Could not find task")
		os.Exit(1)
	}

	if completed == "t" {
		task.Completed = true
	} else {
		task.Completed = false
	}

	for i := 0; i < len(storage_file.TaskList); i++ {
		if storage_file.TaskList[i].Id == int32(id) {
			if completed == "t" {
				storage_file.TaskList[i].Completed = true
			} else {
				storage_file.TaskList[i].Completed = false
			}
		}
	}
}

func change_priority(storage_file *Tasks, priority, id int) {
	for i := 0; i < len(storage_file.TaskList); i++ {
		if storage_file.TaskList[i].Id == int32(id) {
			storage_file.TaskList[i].Priority = int32(priority)
		}
	}
}

func add_sub_task(storage_file *Tasks, sub Task, id int) {
	for i := 0; i < len(storage_file.TaskList); i++ {
		if storage_file.TaskList[i].Id == int32(id) {
			if storage_file.TaskList[i].SubTasks == nil {
				storage_file.TaskList[i].SubTasks = []Task{sub}
			} else {
				storage_file.TaskList[i].SubTasks = append(storage_file.TaskList[i].SubTasks, sub)
			}
		}
	}
}

func search_for_id(tasks []Task, id int) *Task {
	for i := 0; i < len(tasks); i++ {
		if tasks[i].Id == int32(id) {
			return &(tasks[i])
		}
		return search_for_id(tasks[i].SubTasks, id)
	}

	return nil
}

func main() {
	new_entry := flag.Bool("c", false, "Create a new entry.")
	name := flag.String("n", "foo", "Name of new task.")
	body := flag.String("b", "bar", "Body of new task.")
	created_ts := flag.String("created", time.Now().Format(time.RFC850), "Time the task was created, default is now.")
	due_ts := flag.String("d", time.Now().Format(time.RFC850), "Time the task is due, default is now.")
	priority := flag.Int("p", 5, "Priority of task")

	task := flag.Int("i", -1, "ID of task to print.")
	status := flag.String("cs", "", "The status of the task, w=working, p=paused, s=stoped.")
	completed := flag.String("cC", "", "Set task as completed.")
	new_priority := flag.Int("cp", -1, "Change the priority of a task.")

	flag.Parse()

	tasks_f_content := open_file("/home/puse/tasks")

	var storage_file Tasks
	err := json.Unmarshal(tasks_f_content, &storage_file)
	check(err)

	if *task != -1 {
		if *status != "" {
			fmt.Println("Change status")
			change_status(&storage_file, *status, *task)
		} else if *completed != "" {
			fmt.Println("Change comp")
			change_completion(&storage_file, *completed, *task)
		} else if *new_priority != -1 {
			change_priority(&storage_file, *new_priority, *task)
		} else if *new_entry {
			new_task := Task{*name, *body, *created_ts, *due_ts, false, "w", storage_file.NewestId + 1, int32(*priority), nil}
			storage_file.NewestId += 1
			add_sub_task(&storage_file, new_task, *task)
		} else {
			print_task(storage_file, *task)
		}
	} else if *new_entry == false {
		print_tasks(storage_file)
	} else {
		new_task := Task{*name, *body, *created_ts, *due_ts, false, "w", storage_file.NewestId + 1, int32(*priority), nil}
		storage_file.NewestId += 1
		storage_file.TaskList = append(storage_file.TaskList, new_task)
		fmt.Println(new_task)
	}

	json_bytes, err := json.Marshal(storage_file)
	write_file("/home/puse/tasks", json_bytes)
}
