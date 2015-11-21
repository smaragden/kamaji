package kamaji

import (
    "github.com/BurntSushi/toml"
    "fmt"
)

type dispatcher struct {
    Listen string
    Port   int
}

type licenseManager struct {
    Interval int // Interval in milliseconds between tries when no license is available
}

type database struct {
    Type string
    Host string
    Port int
}

type logging struct {
    Level string
}

type Configuration struct {
    Logging    logging `toml:"logging"`
    LicenseManager licenseManager `toml:"licenseManager"`
    Dispatcher dispatcher `toml:"dispatcher"`
    Database   database `toml:"database"`
}


var Config Configuration

func init() {
    Config = Configuration{
        Logging:logging{
            Level: "info",
        },
        Dispatcher:dispatcher{
            Listen: "0.0.0.0",
            Port: 1314,
        },
        LicenseManager:licenseManager{
           Interval: 200,
        },
        Database:database{
            Type: "sqlite",
            Host: "localhost",
            Port: 0,
        },
    }
    if _, err := toml.DecodeFile("kamaji.conf", &Config); err != nil {
        fmt.Println("Config parsing failed, using defaults.")
    }
    fmt.Printf("Conf: %+v\n", Config)
}



