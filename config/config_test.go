package config

import (
	"fmt"
	"testing"
)

func TestConfig_ParseConfFile(t *testing.T) {
	cfg := &Config{ConfigFile: "./test.conf"}
	err := cfg.ParseConfFile()
	if err != nil {
		t.Error(err)
	}

	if cfg.Host != "127.0.0.1" {
		t.Error(fmt.Sprintf("cfg.Host == %s, expect 127.0.0.1", cfg.Host))
	}
	if cfg.Port != 6399 {
		t.Error(fmt.Sprintf("cfg.Port == %d, expect 6399", cfg.Port))
	}
	if cfg.LogDir != "/log" {
		t.Error(fmt.Sprintf("cfg.LogDir == %s, expect /log", cfg.LogDir))
	}
	if cfg.LogLevel != "info" {
		t.Error(fmt.Sprintf("cfg.LogLevel == %s, expect info", cfg.LogLevel))
	}
	if cfg.DbNum != 16 {
		t.Error(fmt.Sprintf("cfg.ShardNum == %d, expect 16", cfg.DbNum))
	}
}
