package config

import (
	"bufio"
	"flag"
	"io"
	"log"
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
	defalutDbNum    int    = 16
)

type Config struct {
	ConfigFile string
	Host       string
	Port       int
	LogDir     string
	LogLevel   string
	DbNum      int
}

func initFlag(conf *Config) {
	flag.StringVar(&(conf.ConfigFile), "config", "", "Select a config file")
	flag.StringVar(&(conf.Host), "host", defaultHost, "Bind a server host")
	flag.IntVar(&(conf.Port), "port", defaultPort, "Bind a server port")
	flag.StringVar(&(conf.LogDir), "logdir", defaultLogDir, "Set log directory")
	flag.StringVar(&(conf.LogLevel), "loglevel", defaultLogLevel, "Set log level")
	flag.IntVar(&(conf.DbNum), "dbnum", defalutDbNum, "Set database number")
}

func Init() {
	_conf := &Config{
		Host:     defaultHost,
		Port:     defaultPort,
		LogDir:   defaultLogDir,
		LogLevel: defaultLogLevel,
		DbNum:    defalutDbNum,
	}

	initFlag(_conf)
	flag.Parse()

	if ip := net.ParseIP(_conf.Host); ip == nil {
		log.Panic(errors.New("given host invaild"))
	}
	if _conf.ConfigFile != "" {
		err := _conf.ParseConfFile()
		log.Panic(err)
	}

	Conf = _conf
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
		case "dbnum":
			conf.DbNum, err = strconv.Atoi(argvs[1])
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
