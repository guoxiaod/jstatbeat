package beater

import (
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/elastic/beats/v7/libbeat/beat"
	"github.com/elastic/beats/v7/libbeat/common"
	"github.com/elastic/beats/v7/libbeat/logp"
	// "github.com/elastic/beats/libbeat/publisher"

	"github.com/guoxiaod/jstatbeat/config"
	"github.com/guoxiaod/jstatbeat/proc"
)

var blanks = regexp.MustCompile(`\s+`)
var cpuCores = proc.NumCores()
var prevCpuTick = int64(0)
var curCpuTick = int64(0)
var lastCpuStats = make(map[string]proc.Cpustat)
var curCpuStats = make(map[string]proc.Cpustat)

type Jstatbeat struct {
	done      chan struct{}
	config    config.Config
	client    beat.Client
	names     map[string]bool
	isExclude bool
}

type JpsResult struct {
	pid   string
	name  string
	alias string
	mx    string
	ms    string
}

// Creates beater
func New(b *beat.Beat, cfg *common.Config) (beat.Beater, error) {
	config := config.DefaultConfig
	if err := cfg.Unpack(&config); err != nil {
		return nil, fmt.Errorf("Error reading config file: %v", err)
	}
	names := make(map[string]bool)
	for _, v := range config.Names {
		names[v] = true
	}
	for _, v := range config.ExcludeNames {
		names[v] = false
	}

	bt := &Jstatbeat{
		done:      make(chan struct{}),
		config:    config,
		names:     names,
		isExclude: len(config.ExcludeNames) > 0,
	}
	return bt, nil
}

// publish sends events into the beats pipeline
func (bt *Jstatbeat) publishAll(content []common.MapStr, b *beat.Beat) error {
	logp.Debug("beat", "publishing %v event(s)", len(content))
	for _, evt := range content {
		// event CreationTime needs "Z" appended (unlike blob contentCreated)
		ts, err := time.Parse(time.RFC3339, evt["CreationTime"].(string)+"Z")
		if err != nil {
			logp.Error(err)
			return err
		}
		fs := common.MapStr{}
		for k, v := range evt {
			fs[k] = v
		}
		beatEvent := beat.Event{Timestamp: ts, Fields: fs}
		bt.client.Publish(beatEvent)
	}
	return nil
}

func (bt *Jstatbeat) Run(b *beat.Beat) error {
	logp.Info("jstatbeat is running! Hit CTRL-C to stop it.")

	var err error
	bt.client, err = b.Publisher.Connect()
	if err != nil {
		logp.Err("Error while connect to ES: %v", err)
		return err
	}
	ticker := time.NewTicker(bt.config.Period)

	for {
		select {
		case <-bt.done:
			return nil
		case <-ticker.C:
		}

		logp.Debug("beat", "Tick - Scanning for Named Java Process")

		output, err := exec.Command("jps", "-v").Output()
		if err != nil {
			logp.Err("Error get while run jps: %v", err)
			continue
		}
		if len(output) < 3 {
			logp.Debug("beat", "Get empty jps result: %v", output)
			continue
		}

		lines := strings.Split(string(output), "\n")
		names := make(map[string]JpsResult)
		for _, line := range lines {
			items := strings.Split(line, " ")
			if len(items) < 2 {
				continue
			}
			isInclude, exists := bt.names[items[1]]
			var alias string = ""
			var mx string = ""
			var ms string = ""
			for _, item := range items {
				if strings.HasPrefix(item, "-Xms") {
					ms = item[4:]
				} else if strings.HasPrefix(item, "-Xmx") {
					mx = item[4:]
				} else if strings.HasPrefix(item, "-DALIAS=") {
					alias = item[8:]
				}
			}
			if (!exists && bt.isExclude) || (exists && isInclude) {
				names[items[0]] = JpsResult{
					pid:   items[0],
					name:  items[1],
					ms:    ms,
					mx:    mx,
					alias: alias,
				}
			}
		}
		if len(names) == 0 {
			logp.Debug("beat", "Didn't find named Java Process - Check again on next Tick")
			continue
		}

		curCpuTick = proc.CpuTick()
		curCpuStats = make(map[string]proc.Cpustat)
		factor := 1. / (float32(curCpuTick-prevCpuTick) / float32(cpuCores) / 100.)
		for pid, _ := range names {
			pid_, _ := strconv.Atoi(pid)
			curCpuStats[pid] = proc.TimeFromPid(pid_)
		}

		events := []beat.Event{}
		logp.Debug("beat", "Starting Jstat Monitor for %v", names)
		for pid, jpsResult := range names {
			jstatoutput, err := exec.Command("jstat", "-gc", "-t", pid, bt.config.Freq, "1").Output()
			if err != nil {
				logp.Err("Error while run jstat: %v, and try next pid", err)
				continue
			}
			if len(jstatoutput) < 30 {
				logp.Debug("beat", "Got empty jstat output, and try next pid")
				continue
			}
			lines = strings.Split(string(jstatoutput), "\n")
			var keys []string
			var version string
			for _, line := range lines {
				logp.Debug("beat", "Line - %v", line)
				if len(line) < 30 {
					continue
				}

				values := blanks.Split(line, -1)
				if len(values) > 2 && values[0] == "Timestamp" {
					keys = values

					if strings.Contains(line, "CCSC") {
						version = "java8"
					} else {
						version = "java5"
					}

					continue
				}

				gcmetrics := common.MapStr{}

				for i, key := range keys {
					if len(key) == 0 {
						continue
					}
					gcmetrics[key], err = strconv.ParseFloat(values[i+1], 64)
				}

				cpu := common.MapStr{}

				lastCpuStat, exists := lastCpuStats[pid]
				if exists {
					curCpuStat, _ := curCpuStats[pid]
					utime := factor * float32(curCpuStat.Utime-lastCpuStat.Utime)
					stime := factor * float32(curCpuStat.Stime-lastCpuStat.Stime)
					cpu["iotime"] = factor * float32(curCpuStat.Iotime-lastCpuStat.Iotime)
					cpu["util"] = utime + stime
					cpu["utime"] = utime
					cpu["stime"] = stime
					cpu["cores"] = cpuCores
				}
				jstatmap := common.MapStr{
					"type":  version,
					"pid":   pid,
					"name":  jpsResult.name,
					"alias": jpsResult.alias,
					"mx":    jpsResult.mx,
					"ms":    jpsResult.ms,
					"gc":    gcmetrics,
					"cpu":   cpu,
				}
				ts := time.Now()
				event := common.MapStr{
					// "@timestamp": common.Time(ts),
					"jstat": jstatmap,
				}

				// beatEvent := beat.Event{Timestamp: ts, Fields: event}
				events = append(events, beat.Event{Timestamp: ts, Fields: event})
				// bt.client.Publish(beatEvent)
				// logp.Info("Event sent")
			}
		}
		if len(events) > 0 {
			bt.client.PublishAll(events)
			logp.Debug("beat", "Event sent: %v", len(events))
		}
		prevCpuTick = curCpuTick
		lastCpuStats = curCpuStats
		curCpuStats = nil
	}

	//	return nil
}

func (bt *Jstatbeat) Stop() {
	bt.client.Close()
	close(bt.done)
}
