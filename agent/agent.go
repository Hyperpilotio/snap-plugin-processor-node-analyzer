package agent

import (
	"errors"
	"os"
	"path"

	"github.com/intelsdi-x/snap-plugin-lib-go/v1/plugin"
	logging "github.com/op/go-logging"
)

var log = logging.MustGetLogger("processor")
var logFormatter = logging.MustStringFormatter(
	` %{level:.1s}%{time:0102 15:04:05.999999} %{pid} %{shortfile}] %{message}`,
)

type FileLog struct {
	Name    string
	Logger  *logging.Logger
	LogFile *os.File
}

// Processor test processor
type SnapProcessor struct {
	Log *FileLog
}

// NewProcessor generate processor
func NewProcessor() plugin.Processor {
	return &SnapProcessor{}
}

func NewLogger(filesPath string, name string) (*FileLog, error) {
	logDirPath := path.Join(filesPath, "log")
	if _, err := os.Stat(logDirPath); os.IsNotExist(err) {
		os.Mkdir(logDirPath, 0777)
	}

	logFilePath := path.Join(logDirPath, name+".log")
	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		return nil, errors.New("Unable to create log file:" + err.Error())
	}

	fileLog := logging.NewLogBackend(logFile, "["+name+"]", 0)
	fileLogLevel := logging.AddModuleLevel(fileLog)
	fileLogLevel.SetLevel(logging.ERROR, "")
	fileLogBackend := logging.NewBackendFormatter(fileLog, logFormatter)

	log.SetBackend(logging.SetBackend(fileLogBackend))

	return &FileLog{
		Name:    name,
		Logger:  log,
		LogFile: logFile,
	}, nil
}

// Process test process function
func (p *SnapProcessor) Process(mts []plugin.Metric, cfg plugin.Config) ([]plugin.Metric, error) {
	metrics := []plugin.Metric{}

	// TODO: put your process logic here

	return metrics, nil
}

/*
	GetConfigPolicy() returns the configPolicy for your plugin.
	A config policy is how users can provide configuration info to
	plugin. Here you define what sorts of config info your plugin
	needs and/or requires.
*/
func (p *SnapProcessor) GetConfigPolicy() (plugin.ConfigPolicy, error) {
	policy := plugin.NewConfigPolicy()
	return *policy, nil
}
