package beater

import (
	"bufio"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/elastic/beats/libbeat/beat"
	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/logp"
	"github.com/elastic/beats/libbeat/publisher"

	"github.com/sw/jstatbeat/config"
)

var blanks = regexp.MustCompile(`\s+`)

type Jstatbeat struct {
	done   chan struct{}
	config config.Config
	client publisher.Client
}

// Creates beater
func New(b *beat.Beat, cfg *common.Config) (beat.Beater, error) {
	config := config.DefaultConfig
	if err := cfg.Unpack(&config); err != nil {
		return nil, fmt.Errorf("Error reading config file: %v", err)
	}

	bt := &Jstatbeat{
		done:   make(chan struct{}),
		config: config,
	}
	return bt, nil
}

func (bt *Jstatbeat) Run(b *beat.Beat) error {
	logp.Info("jstatbeat is running! Hit CTRL-C to stop it.")

	bt.client = b.Publisher.Connect()
	ticker := time.NewTicker(bt.config.Period)

	for {
		select {
		case <-bt.done:
			return nil
		case <-ticker.C:
		}

		logp.Info("Tick - Scanning for Named Java Process")
		pid_id := ""

		cmd := exec.Command("jps")
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			logp.Err("Error get stdout pipe: %v", err)
			return err
		}

		cmd.Start()
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			items := strings.Split(line, " ")

			if len(items) == 2 && items[1] == bt.config.Name {
				pid_id = items[0]
				break
			}
		}
		cmd.Wait()
		if pid_id == "" {
			logp.Info("Didn't find named Java Process - Check again on next Tick")
			continue
		}

		logp.Info("Starting Jstat Monitor for %v", pid_id)
		jstatcmd := exec.Command("jstat", "-gc", "-t", pid_id, bt.config.Freq)
		jstatstdout, err := jstatcmd.StdoutPipe()
		if err != nil {
			logp.Err("Error get stdout pipe: %v", err)
			return err
		}

		jstatcmd.Start()
		jstatscanner := bufio.NewScanner(jstatstdout)
		var keys []string
		var version string
		for jstatscanner.Scan() {
			line := jstatscanner.Text()
			logp.Info("Line - %v", line)

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

			jstatmap := common.MapStr{
				"type": version,
				"pid":  pid_id,
				"name": bt.config.Name,
				"gc":   gcmetrics,
			}

			event := common.MapStr{
				"@timestamp": common.Time(time.Now()),
				"jstat":      jstatmap,
			}

			bt.client.PublishEvent(event)
			logp.Info("Event sent")
		}
		if err := scanner.Err(); err != nil {
			logp.Err("Scanner Error %v", err)
			return nil
		}

		logp.Info("Finished Scanning stdout - Check for another Java Process on next Tick")
	}

	//	return nil
}

func (bt *Jstatbeat) Stop() {
	bt.client.Close()
	close(bt.done)
}
