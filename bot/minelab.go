package bot

import (
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/racerxdl/minelab-bot/config"
	"github.com/racerxdl/minelab-bot/database"
	"github.com/racerxdl/minelab-bot/hockevent"
	"github.com/racerxdl/minelab-bot/models"
	"github.com/sandertv/gophertunnel/minecraft/text"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/openpgp/packet"
)

const ServerName = "MineLab"

type Minelab struct {
	log           *logrus.Logger
	players       map[string]*models.Player
	playerDeaths  map[string]int
	cfg           config.Config
	chatBroadcast chan packet.Packet
	// lastupdatetick uint64
	db database.DB

	playerLock sync.RWMutex
	// broadcastLock      sync.Mutex
	broadcastReceivers map[chan<- packet.Packet]struct{}
	globalstop         chan struct{}

	dg      *discordgo.Session
	discord struct {
		sync.RWMutex
		chatChannelId     string
		cachedPlayerUsers map[string]string
		lastmessages      []string
		lastupdate        time.Time
		deduplock         sync.Mutex
	}

	behock hockevent.HockClient
	sender chan hockevent.HockEvent
}

func MakeMinelab(cfg config.Config) (*Minelab, error) {
	db, err := database.MakeDB("minelab.db")
	if err != nil {
		return nil, err
	}
	return &Minelab{
		log:                logrus.New(),
		players:            make(map[string]*models.Player),
		cfg:                cfg,
		broadcastReceivers: make(map[chan<- packet.Packet]struct{}),
		chatBroadcast:      make(chan packet.Packet, 1),
		playerDeaths:       make(map[string]int),
		db:                 db,
	}, nil
}

func (lab *Minelab) Start() error {
	lab.log.Info("Server running")
	lab.globalstop = make(chan struct{})

	stop := make(chan struct{}, 3)
	defer func() {
		for i := 0; i < cap(stop); i++ {
			stop <- struct{}{}
		}
		close(stop)
	}()

	c, err := hockevent.Connect(lab.cfg.Bedhock.Address)
	if err != nil {
		return err
	}
	lab.behock = c
	lab.sender = make(chan hockevent.HockEvent, 100)
	c.Send(lab.sender)

	go lab.routine(stop)
	go lab.discordRoutine(stop)
	go lab.rxLoop(stop)

	go func() {
		time.Sleep(time.Second * 10)
		lab.RequestPlayerList()
		lab.RequestPlayerDeathCount()
	}()

	<-lab.globalstop
	lab.log.Info("Received global stop\n")

	lab.behock.Stop()
	for i := 0; i < cap(stop); i++ {
		stop <- struct{}{}
	}
	close(stop)

	close(lab.sender)
	_ = lab.db.Close()
	return nil
}

func (lab *Minelab) Stop() {
	if lab.globalstop != nil {
		lab.globalstop <- struct{}{}
		close(lab.globalstop)
	}
}

func (lab *Minelab) rxLoop(stop chan struct{}) {
	lab.log.Info("RX Loop Started")
	defer lab.log.Info("RX Loop Ended")
	rcv := lab.behock.Recv()

	loop := true
	for loop {
		select {
		case <-stop:
			loop = false
		case event := <-rcv:
			lab.HandlePacket(event)
		}
	}
}
func (lab *Minelab) routine(stop chan struct{}) {
	t := time.NewTicker(time.Hour)
	defer t.Stop()

	loop := true
	for loop {
		select {
		case <-stop:
			loop = false
		case <-t.C:
			lab.BroadcastMessage(ServerName, text.Colourf("<B>The time now is</B>: <red>%s</red>", time.Now().Format(time.RFC822Z)))
		}
	}
}

func (lab *Minelab) handleLog(event hockevent.LogEvent) {
	lab.log.Infof("[BDS] %s", event.Message)
}

func (lab *Minelab) HandlePacket(event hockevent.HockEvent) {
	switch event.GetType() {
	case hockevent.EventMessage:
		lab.handleChat(*event.(*hockevent.MessageEvent))
	case hockevent.EventPlayerJoin:
		lab.handlePlayerJoin(*event.(*hockevent.PlayerJoinEvent))
	case hockevent.EventPlayerLeft:
		lab.handlePlayerLeft(*event.(*hockevent.PlayerLeftEvent))
	case hockevent.EventPlayerDeath:
		lab.handlePlayerDeath(*event.(*hockevent.PlayerDeathEvent))
	case hockevent.EventPlayerUpdate:
		lab.handlePlayerUpdate(*event.(*hockevent.PlayerUpdateEvent))
	case hockevent.EventPlayerList:
		lab.handlePlayerList(*event.(*hockevent.PlayerListEvent))
	case hockevent.EventPlayerDimensionChange:
		lab.handlePlayerDimensionChanged(*event.(*hockevent.PlayerDimensionChangeEvent))
	case hockevent.EventPlayerDeathCountResponse:
		lab.handlePlayerDeathCount(*event.(*hockevent.PlayerDeathCountResponseEvent))
	case hockevent.EventLog:
		lab.handleLog(*event.(*hockevent.LogEvent))
	}
}
