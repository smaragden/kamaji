package kamaji

import (
    "github.com/BurntSushi/toml"
    "fmt"
)

type dispatcher struct {
    Listen string
    Port   int
}

type database struct {
    Type string
    Host string
    Port int
}

type logging struct {
    Task           string
    Dispatcher     string
    Nodemanager    string
    Taskmanager    string
    Licensemanager string
}

type Configuration struct {
    Logging    logging `toml:"logging"`
    Dispatcher dispatcher `toml:"dispatcher"`
    Database   database `toml:"database"`
}


var Config Configuration

func init() {
    Config = Configuration{
        Logging:logging{
            Task:"info",
            Dispatcher:"info",
            Nodemanager:"info",
            Taskmanager:"info",
            Licensemanager:"info",
        },
        Dispatcher:dispatcher{
            Listen:"0.0.0.0",
            Port:1314,
        },
        Database:database{
            Type:"sqlite",
            Host:"localhost",
            Port:0,
        },
    }
    if _, err := toml.DecodeFile("kamaji.conf", &Config); err != nil {
        fmt.Println("Config parsing failed, using defaults.")
    }
    fmt.Printf("Conf: %+v\n", Config)
}



