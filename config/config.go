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

func initFlag() {
	flag.StringVar(&(Conf.ConfigFile), "config", "", "Select a config file")
	flag.StringVar(&(Conf.Host), "host", defaultHost, "Bind a server host")
	flag.IntVar(&(Conf.Port), "port", defaultPort, "Bind a server port")
	flag.StringVar(&(Conf.LogDir), "logdir", defaultLogDir, "Set log directory")
	flag.StringVar(&(Conf.LogLevel), "loglevel", defaultLogLevel, "Set log level")
	flag.IntVar(&(Conf.DbNum), "dbnum", defalutDbNum, "Set database number")
}

func Init() {
	_conf := &Config{
		Host:     defaultHost,
		Port:     defaultPort,
		LogDir:   defaultLogDir,
		LogLevel: defaultLogLevel,
		DbNum:    defalutDbNum,
	}

	initFlag()
	flag.Parse()

	if ip := net.ParseIP(_conf.Host); ip == nil {
		log.Panic(errors.New("given host invaild"))
	}
	if Conf.ConfigFile != "" {
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
		case "loglever":
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
