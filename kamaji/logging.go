package kamaji
import "github.com/Sirupsen/logrus"

func init() {
	level, err := logrus.ParseLevel(Config.Logging.Level)
    if err == nil {
        logrus.SetLevel(level)
    }
}
