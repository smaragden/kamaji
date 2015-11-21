package kamaji

import (
	"fmt"
//"html"
	log "github.com/Sirupsen/logrus"
	"net/http"
	"github.com/gorilla/mux"
	"encoding/json"
	"strings"
	"github.com/tylerb/graceful"
//"io/ioutil"
	"time"
)

type J_Node struct {
	Name  string    `json:"name"`
	Id    string    `json:"id"`
	State string   `json:"state"`
}

type J_Nodes []J_Node

type J_Job struct {
	Name       string    `json:"name"`
	Id         string    `json:"id"`
	State      string   `json:"state"`
	Completion float32 `json:"completion"`
}

type J_Jobs []J_Job

type J_Task struct {
	Name       string    `json:"name"`
	Id         string    `json:"id"`
	State      string   `json:"state"`
	Completion float32 `json:"completion"`
}

type J_Tasks []J_Task

type J_Command struct {
	Name       string    `json:"name"`
	Id         string    `json:"id"`
	State      string   `json:"state"`
	Completion float32 `json:"completion"`
}

type J_Commands []J_Command

type Service struct {
	s    *graceful.Server
	r    *mux.Router
	Addr string
	Port int
	lm   *LicenseManager
	nm   *NodeManager
	tm   *TaskManager

}

func NewService(address string, port int, lm *LicenseManager, nm *NodeManager, tm *TaskManager) *Service {
	s := new(Service)
	s.Addr = address
	s.Port = port
	s.lm = lm
	s.nm = nm
	s.tm = tm
	return s
}

func (s *Service) GetAddrStr() string {
	return fmt.Sprintf("%s:%d", s.Addr, s.Port)
}

func (s *Service) Start() {
	s.r = mux.NewRouter().StrictSlash(true)
	s.r.HandleFunc("/", s.Index)
	s.r.HandleFunc("/nodes", s.NodesIndex)
	s.r.HandleFunc("/jobs", s.JobsIndex)
	s.r.HandleFunc("/jobs/{jobId}", s.JobShow)
	s.r.HandleFunc("/jobs/{jobId}/tasks", s.TasksIndex)
	s.r.HandleFunc("/jobs/{jobId}/tasks/{taskId}", s.TaskShow)
	s.r.HandleFunc("/jobs/{jobId}/tasks/{taskId}/commands", s.CommandsIndex)
	s.r.HandleFunc("/jobs/{jobId}/tasks/{taskId}/commands/{commandId}", s.CommandShow)
	s.s = &graceful.Server{
		Timeout: 10 * time.Second,
		NoSignalHandling: true,
		Server: &http.Server{
			Addr: s.GetAddrStr(),
			Handler: s.r,
		},
	}
	s.s.ListenAndServe()
}

func (s *Service) Stop() {
	log.WithFields(log.Fields{
        "module":  "service",
        "action":  "stop",
    }).Info("Stopping Web Service")
	s.s.Stop(10 * time.Second)
}

func (s *Service) Index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Welcome!")
}

type test_struct struct {
	Type string
}

func (s *Service) NodesIndex(w http.ResponseWriter, r *http.Request) {
	callback := r.FormValue("callback")
	fmt.Println(callback)
	nodes := J_Nodes{}
	for _, node := range s.nm.Nodes {
		j_node := J_Node{Name: node.Name, Id: node.ID.String(), State: strings.ToLower(node.State.S())}
		nodes = append(nodes, j_node)
	}
	jsonBytes, _ := json.Marshal(nodes)
	if callback != "" {
		fmt.Fprintf(w, "%s(%s)", callback, jsonBytes)
	} else {
		w.Write(jsonBytes)
	}

}

func (s *Service) JobsIndex(w http.ResponseWriter, r *http.Request) {
	callback := r.FormValue("callback")
	fmt.Println(callback)
	jobs := J_Jobs{}
	for _, job := range s.tm.Jobs {
		j_job := J_Job{Name: job.Name, Id: job.ID.String(), State: strings.ToLower(job.State.S()), Completion: job.Completion}
		jobs = append(jobs, j_job)
	}
	jsonBytes, _ := json.Marshal(jobs)
	if callback != "" {
		fmt.Fprintf(w, "%s(%s)", callback, jsonBytes)
	} else {
		w.Write(jsonBytes)
	}

}

func (s *Service) JobShow(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	if origin := r.Header.Get("Origin"); origin != "" {
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers",
			"Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
	}
	vars := mux.Vars(r)
	jobId := vars["jobId"]
	job := s.tm.GetJobFromId(jobId)
	if job != nil {
		j_job := J_Job{Name: job.Name, Id: job.ID.String(), State: strings.ToLower(job.State.S()), Completion: job.Completion}
		json.NewEncoder(w).Encode(j_job)
	}else {
		fmt.Fprintln(w, "Not Found:", jobId)
	}

}

func (s *Service) TasksIndex(w http.ResponseWriter, r *http.Request) {
	callback := r.FormValue("callback")
	vars := mux.Vars(r)
	jobId := vars["jobId"]
	tasks := J_Tasks{}
	job := s.tm.GetJobFromId(jobId)
	if job != nil {
		for _, task := range job.GetTasks() {
			j_task := J_Task{Name: task.Name, Id: task.ID.String(), State: strings.ToLower(job.State.S()), Completion: task.Completion}
			tasks = append(tasks, j_task)
		}
		jsonBytes, _ := json.Marshal(tasks)
		if callback != "" {
			fmt.Fprintf(w, "%s(%s)", callback, jsonBytes)
		} else {
			w.Write(jsonBytes)
		}
	}else {
		fmt.Fprintln(w, "Not Found:", jobId)
	}
}

func (s *Service) TaskShow(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	jobId := vars["jobId"]
	taskId := vars["taskId"]
	job := s.tm.GetJobFromId(jobId)
	if job != nil {
		task := job.GetTaskFromId(taskId)
		if task != nil {
			j_task := J_Task{Name: task.Name, Id: task.ID.String(), State: strings.ToLower(job.State.S())}
			json.NewEncoder(w).Encode(j_task)
		}else {
			fmt.Fprintln(w, "Not Found:", taskId)
		}
	}else {
		fmt.Fprintln(w, "Not Found:", jobId)
	}

}

func (s *Service) CommandsIndex(w http.ResponseWriter, r *http.Request) {
	callback := r.FormValue("callback")
	vars := mux.Vars(r)
	jobId := vars["jobId"]
	taskId := vars["taskId"]
	commands := J_Commands{}
	job := s.tm.GetJobFromId(jobId)
	if job != nil {
		task := job.GetTaskFromId(taskId)
		if task != nil {
			for _, command := range task.getCommands() {
				j_command := J_Command{Name: command.Name, Id: command.ID.String(), State: strings.ToLower(job.State.S()), Completion: command.Completion}
				commands = append(commands, j_command)
			}
			jsonBytes, _ := json.Marshal(commands)
			if callback != "" {
				fmt.Fprintf(w, "%s(%s)", callback, jsonBytes)
			} else {
				w.Write(jsonBytes)
			}
		}else {
			fmt.Fprintln(w, "Not Found:", taskId)
		}
	}else {
		fmt.Fprintln(w, "Not Found:", jobId)
	}
}

func (s *Service) CommandShow(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	jobId := vars["jobId"]
	taskId := vars["taskId"]
	commandId := vars["commandId"]
	job := s.tm.GetJobFromId(jobId)
	if job != nil {
		task := job.GetTaskFromId(taskId)
		if task != nil {
			command := task.GetCommandFromId(commandId)
			if command != nil {
				j_command := J_Command{Name: command.Name, Id: command.ID.String(), State: job.State.S()}
				json.NewEncoder(w).Encode(j_command)
			}else {
				fmt.Fprintln(w, "Not Found:", commandId)
			}
		}else {
			fmt.Fprintln(w, "Not Found:", taskId)
		}
	}else {
		fmt.Fprintln(w, "Not Found:", jobId)
	}
}


