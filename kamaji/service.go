package kamaji

import (
    "fmt"
//"html"
    "log"
    "net/http"
    "github.com/gorilla/mux"
    "encoding/json"
    //"io/ioutil"
)

type J_Job struct {
    Name  string    `json:"name"`
    Id    string    `json:"id"`
    State string   `json:state`
}

type J_Jobs []J_Job

type J_Task struct {
    Name  string    `json:"name"`
    Id    string    `json:"id"`
    State string   `json:state`
}

type J_Tasks []J_Task

type J_Command struct {
    Name  string    `json:"name"`
    Id    string    `json:"id"`
    State string   `json:state`
}

type J_Commands []J_Command

type Service struct {
    r *mux.Router
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
    s.r.HandleFunc("/jobs", s.JobsIndex)
    s.r.HandleFunc("/jobs/{jobId}", s.JobShow)
    s.r.HandleFunc("/jobs/{jobId}/tasks", s.TasksIndex)
    s.r.HandleFunc("/jobs/{jobId}/tasks/{taskId}", s.TaskShow)
    s.r.HandleFunc("/jobs/{jobId}/tasks/{taskId}/commands", s.CommandsIndex)
    s.r.HandleFunc("/jobs/{jobId}/tasks/{taskId}/commands/{commandId}", s.CommandShow)
    log.Fatal(http.ListenAndServe(s.GetAddrStr(), s.r))
}

func (s *Service) Stop() {

}

func (s *Service) Index(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintln(w, "Welcome!")
}

type test_struct struct {
    Type string
}


func (s *Service) JobsIndex(w http.ResponseWriter, r *http.Request) {
    callback := r.FormValue("callback")
    fmt.Println(callback)
    jobs := J_Jobs{}
    for _, job := range s.tm.Jobs {
        j_job := J_Job{Name: job.Name, Id: job.ID.String(), State: job.State.S()}
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
        j_job := J_Job{Name: job.Name, Id: job.ID.String(), State: job.State.S()}
        json.NewEncoder(w).Encode(j_job)
    }else {
        fmt.Fprintln(w, "Not Found:", jobId)
    }

}

func (s *Service) TasksIndex(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    jobId := vars["jobId"]
    tasks := J_Tasks{}
    job := s.tm.GetJobFromId(jobId)
    if job != nil {
        for _, task := range job.GetTasks() {
            j_task := J_Task{Name: task.Name, Id: task.ID.String(), State: job.State.S()}
            tasks = append(tasks, j_task)
        }
        json.NewEncoder(w).Encode(tasks)
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
            j_task := J_Task{Name: task.Name, Id: task.ID.String(), State: job.State.S()}
            json.NewEncoder(w).Encode(j_task)
        }else {
            fmt.Fprintln(w, "Not Found:", taskId)
        }
    }else {
        fmt.Fprintln(w, "Not Found:", jobId)
    }

}

func (s *Service) CommandsIndex(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    jobId := vars["jobId"]
    taskId := vars["taskId"]
    commands := J_Commands{}
    job := s.tm.GetJobFromId(jobId)
    if job != nil {
        task := job.GetTaskFromId(taskId)
        if task != nil {
            for _, command := range task.getCommands() {
                j_command := J_Command{Name: command.Name, Id: command.ID.String(), State: job.State.S()}
                commands = append(commands, j_command)
            }
            json.NewEncoder(w).Encode(commands)
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


