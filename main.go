package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"strings"
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

type ByDueDate []*Task
type ByPriority []*Task

const BODY_LIMIT int = 50

var new_entry = flag.Bool("c", false, "Create a new entry.")
var name = flag.String("n", "foo", "Name of new task.")
var body = flag.String("b", "bar", "Body of new task.")
var created_ts = flag.String("created", time.Now().Format(time.RFC850), "Time the task was created, default is now.")
var due_ts = flag.String("d", "", "Time the task is due, default is now.")
var priority = flag.Int("p", 5, "Priority of task")
var hide_completed = flag.Bool("hc", false, "Hide Completed task from printout")
var task_sub = flag.Bool("ts", false, "Print out a task like in the main view together with the sub tasks.")
var alert_duef = flag.Bool("ad", false, "Prints out tasks that are over due.")
var alert_due_short = flag.Bool("ads", false, "Prints out due tasks in short form.")
var status_filter = flag.String("ps", "", "Print out tasks based on their status.")

var task = flag.Int("i", -1, "ID of task to print.")
var status = flag.String("cs", "", "The status of the task, w=working, p=paused, s=stoped.")
var completed = flag.String("cC", "", "Set task as completed. t or 1 is completed, 0 or f is not completed.")
var new_priority = flag.Int("cp", -1, "Change the priority of a task.")

func (a ByDueDate) Len() int { return len(a) }
func (a ByDueDate) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}
func (a ByDueDate) Less(i, j int) bool {
	t1, err := time.Parse(time.RFC850, (a[i]).Due)
	check(err)
	t2, err := time.Parse(time.RFC850, (a[j]).Due)
	check(err)

	return t1.Before(t2)
}

func (a ByPriority) Len() int           { return len(a) }
func (a ByPriority) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByPriority) Less(i, j int) bool { return a[i].Priority < a[j].Priority }

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
	// If ident is 4 or higher don't do anything
	if indent >= 4 {
		return
	}

	// Show ... if the task has sub task and the task is already a sub sub task.
	if indent == 3 {
		for i := 0; i < indent; i++ {
			fmt.Print("\u2551\t\t")
		}
		fmt.Println("...")
		for i := 0; i < indent; i++ {
			fmt.Print("\u2551\t\t")
		}
		fmt.Println("...")
		return
	}

	for i := 0; i < indent; i++ {
		fmt.Print("\u2551\t\t")
	}

	task_status := task.Status
	if task.Completed {
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
	fmt.Print("\u2554")
	for i := 0; i < 79; i++ {
		fmt.Print("\u2550")
	}
	fmt.Print("\n")
	postfix := "\u2560"
	for idx, task := range storage_file.TaskList {
		if idx == (len(storage_file.TaskList) - 1) {
			postfix = "\u2560"
		}

		status := task.Status

		if task.Completed == true {
			status = "\u2713"
		}

		fmt.Println(postfix, task.Id, ": ", task.Priority, ": ", status, ": ", task.Name)

		for _, sub := range task.SubTasks {
			print_sub_task(sub, 1, "\u2560")
		}

		postfix = "\u2560"
	}
}

func check_subtask_completion(tasks []Task) bool {
	for _, task := range tasks {
		if task.Completed == false {
			return false
		}

		if !check_subtask_completion(task.SubTasks) {
			return false
		}
	}

	return true
}

func hide_completed_print(storage_file Tasks) {
	fmt.Println("ID  : Priority :  Status  :  Name")
	fmt.Print("\u2554")
	for i := 0; i < 79; i++ {
		fmt.Print("\u2550")
	}
	fmt.Print("\n")
	postfix := "\u2560"

	for idx, task := range storage_file.TaskList {
		if task.Completed && len(task.SubTasks) == 0 {
			continue
		}
		if task.Completed && len(task.SubTasks) != 0 && check_subtask_completion(task.SubTasks) {
			continue
		}

		if idx == (len(storage_file.TaskList) - 1) {
			postfix = "\u2560"
		}

		status := task.Status

		if task.Completed == true {
			status = "\u2713"
		}

		fmt.Println(postfix, task.Id, ": ", task.Priority, ": ", status, ": ", task.Name)

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
	task := search_for_id(storage_file.TaskList, id)
	if task == nil {
		fmt.Println("Could not find task")
		os.Exit(1)
	}

	fmt.Println("Name:    ", task.Name, "")
	fmt.Println("Created: ", task.Created)
	fmt.Println("Due:     ", task.Due)

	if task.Completed {
		fmt.Println("Status:  \u2713 Completed")
	} else {
		fmt.Println("Status:  ", task.Status)
	}

	fmt.Println("Priority:", task.Priority)
	fmt.Print("Sub Tasks: ")

	for _, sub := range task.SubTasks {
		fmt.Print(sub.Id, " ")
	}

	fmt.Print("\n")

	if len(task.Body) < BODY_LIMIT {
		fmt.Println("Body:    ", task.Body)
	} else {
		fmt.Println("Body:\n")
		fmt.Println(nice_body(task.Body))
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
}

func change_priority(storage_file *Tasks, priority, id int) {
	task := search_for_id(storage_file.TaskList, id)
	if task == nil {
		fmt.Println("Could not find task")
		os.Exit(1)
	}

	task.Priority = int32(priority)
}

func add_sub_task(storage_file *Tasks, sub Task, id int) {
	task := search_for_id(storage_file.TaskList, id)
	if task == nil {
		fmt.Println("Could not find task")
		os.Exit(1)
	}

	if task.SubTasks == nil {
		task.SubTasks = []Task{sub}
	} else {
		task.SubTasks = append(task.SubTasks, sub)
	}
}

func search_for_id(tasks []Task, id int) *Task {
	for i := 0; i < len(tasks); i++ {
		if tasks[i].Id == int32(id) {
			return &(tasks[i])
		}

		ret := search_for_id(tasks[i].SubTasks, id)
		if ret != nil && ret.Id == int32(id) {
			return ret
		}
	}

	return nil
}

func print_task_and_subs(storage_file Tasks, id int) {
	task := search_for_id(storage_file.TaskList, id)
	if task == nil {
		fmt.Println("Could not find task")
		os.Exit(1)
	}

	fmt.Println("ID  : Priority :  Status  :  Name")
	fmt.Print("\u2554")
	for i := 0; i < 79; i++ {
		fmt.Print("\u2550")
	}
	fmt.Print("\n")
	postfix := "\u2560"

	status := task.Status

	if task.Completed == true {
		status = "\u2713"
	}

	fmt.Println(postfix, task.Id, ": ", task.Priority, ": ", status, ": ", task.Name)

	for _, sub := range task.SubTasks {
		print_sub_task(sub, 1, "\u2560")
	}
}

/**
 *  @brief This function prints out the tasks that are overdue.
 *  The tasks are printed out in the same fashion as when a
 *  single task is selected.
 */
func print_alert_due(tasks []Task) {
	now := time.Now()
	for _, task := range tasks {
		task_due, err := time.Parse(time.RFC850, task.Due)
		check(err)

		if !task.Completed && task_due.Before(now) {
			fmt.Println("Name:    ", task.Name, "")
			fmt.Println("Created: ", task.Created)
			fmt.Println("Due:     ", task.Due)

			if task.Completed {
				fmt.Println("Status:  \u2713 Completed")
			} else {
				fmt.Println("Status:  ", task.Status)
			}

			fmt.Println("Priority:", task.Priority)
			fmt.Print("Sub Tasks: ")

			for _, sub := range task.SubTasks {
				fmt.Print(sub.Id, " ")
			}

			fmt.Print("\n")

			if len(task.Body) < BODY_LIMIT {
				fmt.Println("Body:    ", task.Body)
			} else {
				fmt.Println("Body:")
				fmt.Println(nice_body(task.Body))
			}

			for i := 0; i < BODY_LIMIT; i++ {
				fmt.Print("\u2500")
			}
			fmt.Print("\n")
		}
		if task.SubTasks != nil {
			print_alert_due(task.SubTasks)
		}
	}
}

/**
 * @brief This function prints out the tasks that are overdue
 * in short form. In short form only the name and the ID of the
 * task is printed out.
 */
func print_alert_due_short(tasks []Task) {
	now := time.Now()
	for _, task := range tasks {
		task_due, err := time.Parse(time.RFC850, task.Due)
		check(err)

		if !task.Completed && task_due.Before(now) {
			fmt.Println("Id: ", task.Id, "Name: ", task.Name)
		}
		if task.SubTasks != nil {
			print_alert_due_short(task.SubTasks)
		}
	}
}

func print_status_filter(tasks []Task) {
	for _, task := range tasks {
		if task.Status == *status_filter && !task.Completed {
			fmt.Println("Id: ", task.Id, "Name :", task.Name)
		}

		if task.SubTasks != nil {
			print_status_filter(task.SubTasks)
		}
	}
}

func parse_date(input_date string) string {
	date_tod := strings.Split(input_date, " ")

	if len(date_tod) != 2 {
		fmt.Println("Input date should be in the format: YYYY.MM.DD hh:mm")
		os.Exit(1)
	}

	date := date_tod[0]
	time_of_day := date_tod[1]

	date_splitted := strings.Split(date, ".")
	if len(date_splitted) != 3 {
		fmt.Println("Input date should be in the format: YYYY.MM.DD hh:mm")
		os.Exit(1)
	}

	year := date_splitted[0]
	month := date_splitted[1]
	day := date_splitted[2]

	tod_splitted := strings.Split(time_of_day, ":")
	if len(tod_splitted) != 2 {
		fmt.Println("Input date should be in the format: YYYY.MM.DD hh:mm")
		os.Exit(1)
	}

	hours := tod_splitted[0]
	minutes := tod_splitted[1]

	the_old_switcharoo := fmt.Sprintf("%s-%s-%s %s:%s:00 +0000 UTC", year, month, day, hours, minutes)

	the_time, err := time.Parse("2006-01-02 15:04:05 +0000 UTC", the_old_switcharoo)
	check(err)

	return the_time.Format(time.RFC850)
}

func add_new_task(storage_file *Tasks) {
	if *due_ts != "" {
		*due_ts = parse_date(*due_ts)
	} else {
		*due_ts = "Not set"
	}

	new_task := Task{
		*name,
		*body,
		*created_ts,
		*due_ts,
		false,
		"ns",
		storage_file.NewestId + 1,
		int32(*priority),
		nil,
	}

	if *task != -1 {
		add_sub_task(storage_file, new_task, *task)
	} else {
		storage_file.TaskList = append(storage_file.TaskList, new_task)
	}

	storage_file.NewestId += 1
}

func build_tasks_array(tasks *[]Task, tasks_array *[]*Task) {
	for i := 0; i < len(*tasks); i++ {
		if (*tasks)[i].Due == "Not set" || (*tasks)[i].Completed {
			continue
		}
		*tasks_array = append(*tasks_array, &((*tasks)[i]))
		build_tasks_array(&((*tasks)[i].SubTasks), tasks_array)
	}
}

func arrange_priorities(tasks *[]*Task) {
	for i := 0; i < len(*tasks); i++ {

	}
}

func priority_scheduler(storage_file *Tasks) {
	tasks := []*Task{}
	build_tasks_array(&storage_file.TaskList, &tasks)

	sort.Sort(ByPriority(tasks))

	arrange_priorities(&tasks)
}

func main() {

	flag.Parse()
	parse_date("2015.05.11 15:15")

	tasks_f_content := open_file("/home/puse/tasks")

	var storage_file Tasks
	err := json.Unmarshal(tasks_f_content, &storage_file)
	check(err)

	if *new_entry {
		add_new_task(&storage_file)
	}

	if *task != -1 && !*new_entry {
		if *status != "" {
			fmt.Println("Change status")
			change_status(&storage_file, *status, *task)
		} else if *completed != "" {
			fmt.Println("Change comp")
			change_completion(&storage_file, *completed, *task)
		} else if *new_priority != -1 {
			change_priority(&storage_file, *new_priority, *task)
		} else if *task_sub {
			print_task_and_subs(storage_file, *task)
		} else {
			print_task(storage_file, *task)
		}
	} else if *status_filter != "" {
		fmt.Println("Searching for tasks with the status: ", *status_filter)
		print_status_filter(storage_file.TaskList)
	} else if *alert_due_short {
		fmt.Println("Tasks that are overdue:")
		for i := 0; i < BODY_LIMIT; i++ {
			fmt.Print("\u2500")
		}
		fmt.Print("\n")

		print_alert_due_short(storage_file.TaskList)
	} else if *alert_duef {
		print_alert_due(storage_file.TaskList)
	} else if *hide_completed {
		hide_completed_print(storage_file)
	} else if *new_entry == false {
		print_tasks(storage_file)
	}

	priority_scheduler(&storage_file)

	json_bytes, err := json.Marshal(storage_file)
	write_file("/home/puse/tasks", json_bytes)
}
