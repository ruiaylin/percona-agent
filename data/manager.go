package data

import (
	"encoding/json"
	"errors"
	"github.com/percona/cloud-protocol/proto"
	"github.com/percona/cloud-tools/pct"
	"time"
)

type Manager struct {
	logger   *pct.Logger
	hostname string
	client   pct.WebsocketClient
	// --
	config    *Config
	configDir string
	sz        Serializer
	spooler   Spooler
	sender    *Sender
	status    *pct.Status
}

func NewManager(logger *pct.Logger, hostname string, client pct.WebsocketClient) *Manager {
	m := &Manager{
		logger:   logger,
		hostname: hostname,
		client:   client,
		// --
		status: pct.NewStatus([]string{"data"}),
	}
	return m
}

/////////////////////////////////////////////////////////////////////////////
// Interface
/////////////////////////////////////////////////////////////////////////////

// @goroutine[0]
func (m *Manager) Start(cmd *proto.Cmd, config []byte) error {
	if m.config != nil {
		err := pct.ServiceIsRunningError{Service: "data"}
		return err
	}

	// proto.Cmd[Service:agent, Cmd:StartService, Data:proto.ServiceData[Name:data Config:data.Config]]
	c := &Config{}
	if err := json.Unmarshal(config, c); err != nil {
		return err
	}

	if err := pct.MakeDir(c.Dir); err != nil {
		return err
	}

	var sz Serializer
	switch c.Encoding {
	case "":
		sz = NewJsonSerializer()
	case "gzip":
		sz = NewJsonGzipSerializer()
	default:
		return errors.New("Unknown encoding: " + c.Encoding)
	}

	spooler := NewDiskvSpooler(
		pct.NewLogger(m.logger.LogChan(), "data-spooler"),
		c.Dir,
		sz,
		m.hostname,
	)
	if err := spooler.Start(); err != nil {
		return err
	}
	m.spooler = spooler
	m.logger.Info("Started spooler")

	sender := NewSender(
		pct.NewLogger(m.logger.LogChan(), "data-sender"),
		m.client,
		m.spooler,
		time.Tick(time.Duration(c.SendInterval)*time.Second),
	)
	if err := sender.Start(); err != nil {
		return err
	}
	m.sender = sender
	m.logger.Info("Started sender")

	m.config = c

	m.status.Update("data", "Ready")
	m.logger.Info("Ready")
	return nil
}

// @goroutine[0]
func (m *Manager) Stop(cmd *proto.Cmd) error {
	// Can't stop data yet.
	return nil
}

// @goroutine[0]
func (m *Manager) Handle(cmd *proto.Cmd) *proto.Reply {
	defer m.status.Update("data", "Ready")
	switch cmd.Cmd {
	case "GetConfig":
		// proto.Cmd[Service:data, Cmd:GetConfig]
		return cmd.Reply(m.config)
	case "Status":
		// proto.Cmd[Service:data, Cmd:Status]
		status := m.InternalStatus()
		return cmd.Reply(status)
	default:
		// todo: dynamic config
		return cmd.Reply(pct.UnknownCmdError{Cmd: cmd.Cmd})
	}
}

// @goroutine[0:1]
func (m *Manager) Status() string {
	return m.status.Get("data", true)
}

// @goroutine[0]
func (m *Manager) InternalStatus() map[string]string {
	s := make(map[string]string)
	s["data"] = m.Status()
	return s
}

func (m *Manager) Spooler() Spooler {
	return m.spooler
}

func (m *Manager) Sender() *Sender {
	return m.sender
}

func (m *Manager) LoadConfig(configDir string) (interface{}, error) {
	m.configDir = configDir
	v, err := pct.ReadConfig(configDir + "/" + CONFIG_FILE)
	if err != nil {
		return nil, err
	}
	config := v.(Config)
	if config.Dir == "" {
		config.Dir = DEFAULT_DATA_DIR
	}
	if config.SendInterval <= 0 {
		config.SendInterval = DEFAULT_DATA_SEND_INTERVAL
	}
	return config, nil
}

func (m *Manager) WriteConfig(config interface{}, name string) error {
	if m.configDir == "" {
		return nil
	}
	file := m.configDir + "/" + CONFIG_FILE
	m.logger.Info("Writing", file)
	return pct.WriteConfig(file, config)
}

func (m *Manager) RemoveConfig(name string) error {
	if m.configDir == "" {
		return nil
	}
	file := m.configDir + "/" + CONFIG_FILE
	m.logger.Info("Removing", file)
	return pct.RemoveFile(file)
}
