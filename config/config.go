package config

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"

	"errors"
)

var Conf *Config

var (
	defaultHost     string = "127.0.0.1"
	defaultPort     int    = 6379
	defaultLogDir   string = "./"
	defaultLogLevel string = "info"
	defalutSegNum   int    = 100
)

type Config struct {
	ConfigFile string
	Host       string
	Port       int
	LogDir     string
	LogLevel   string
	SegNum     int // segmentation number
}

type CfgError struct {
	message string
}

func (e *CfgError) Error() string {
	return e.message
}

func initFlag(conf *Config) {
	flag.StringVar(&(conf.ConfigFile), "config", "", "Set a config file")
	flag.StringVar(&(conf.Host), "host", defaultHost, "Set a server host to listen")
	flag.IntVar(&(conf.Port), "port", defaultPort, "Set a server prot to listen")
	flag.StringVar(&(conf.LogDir), "logdir", defaultLogDir, "Set a log directory")
	flag.StringVar(&(conf.LogLevel), "loglevel", defaultLogLevel, "Set a log level")
	flag.IntVar(&(conf.SegNum), "segnum", defalutSegNum, "Set a segmentation number for cache database")
}

func Init() (*Config, error) {
	_conf := &Config{
		Host:     defaultHost,
		Port:     defaultPort,
		LogDir:   defaultLogDir,
		LogLevel: defaultLogLevel,
		SegNum:   defalutSegNum,
	}

	initFlag(_conf)
	flag.Parse()

	if ip := net.ParseIP(_conf.Host); ip == nil {
		ipErr := &CfgError{message: fmt.Sprintf("Ip address %s is invaild", _conf.Host)}
		return nil, ipErr
	}
	if _conf.ConfigFile != "" {
		err := _conf.ParseConfFile()
		if err != nil {
			return nil, err
		}
	}

	Conf = _conf
	return Conf, nil
}

func (conf *Config) ParseConfFile() error {
	file, err := os.Open(conf.ConfigFile)
	if err != nil {
		return err
	}

	defer file.Close()

	fileReader := bufio.NewReader(file)

	for {
		line, ioErr := fileReader.ReadString('\n')
		if ioErr != nil && ioErr != io.EOF {
			return err
		}

		argvs := strings.Fields(line)

		if len(argvs) == 0 {
			continue
		}

		switch argvs[0] {
		case "host":
			if ip := net.ParseIP(argvs[1]); ip == nil {
				return errors.New("given host invaild")
			}
			conf.Host = argvs[1]
		case "port":
			conf.Port, err = strconv.Atoi(argvs[1])
			if err != nil {
				return err
			}
		case "logdir":
			conf.LogDir = argvs[1]
		case "loglevel":
			conf.LogLevel = strings.ToLower(argvs[1])
		case "segnum":
			conf.SegNum, err = strconv.Atoi(argvs[1])
			if err != nil {
				return err
			}
		}

		if ioErr == io.EOF {
			break
		}
	}
	return nil
}
