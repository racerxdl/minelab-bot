package bot

import (
	"github.com/bwmarrin/discordgo"
	"github.com/racerxdl/minelab-bot/config"
	"github.com/racerxdl/minelab-bot/hockevent"
	"github.com/racerxdl/minelab-bot/models"
	"github.com/sandertv/gophertunnel/minecraft/text"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/openpgp/packet"
	"sync"
	"time"
)

const ServerName = "MineLab"

type Minelab struct {
	log            *logrus.Logger
	players        map[string]*models.Player
	cfg            config.Config
	chatBroadcast  chan packet.Packet
	lastupdatetick uint64

	playerLock         sync.RWMutex
	broadcastLock      sync.Mutex
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

func MakeMinelab(cfg config.Config) *Minelab {
	return &Minelab{
		log:                logrus.New(),
		players:            make(map[string]*models.Player),
		cfg:                cfg,
		broadcastReceivers: make(map[chan<- packet.Packet]struct{}),
		chatBroadcast:      make(chan packet.Packet, 1),
	}
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
	}()

	<-lab.globalstop
	lab.log.Info("Received global stop\n")

	lab.behock.Stop()
	for i := 0; i < cap(stop); i++ {
		stop <- struct{}{}
	}
	close(stop)

	close(lab.sender)

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
	for {
		select {
		case <-stop:
			break
		case event := <-rcv:
			lab.HandlePacket(event)
		}
	}
}
func (lab *Minelab) routine(stop chan struct{}) {
	t := time.NewTicker(time.Hour)
	defer t.Stop()

	for {
		select {
		case <-stop:
			break
		case <-t.C:
			lab.BroadcastMessage(ServerName, text.Colourf("<B>The time now is</B>: <red>%s</red>", time.Now().Format(time.RFC822Z)))
		}
	}
}

func (lab *Minelab) HandlePacket(event hockevent.HockEvent) {
	switch event.GetType() {
	case hockevent.EVENT_MESSAGE:
		lab.handleChat(*event.(*hockevent.MessageEvent))
	case hockevent.EVENT_PLAYER_JOIN:
		lab.handlePlayerJoin(*event.(*hockevent.PlayerJoinEvent))
	case hockevent.EVENT_PLAYER_LEFT:
		lab.handlePlayerLeft(*event.(*hockevent.PlayerLeftEvent))
	case hockevent.EVENT_PLAYER_DEATH:
		lab.handlePlayerDeath(*event.(*hockevent.PlayerDeathEvent))
	case hockevent.EVENT_PLAYER_UPDATE:
		lab.handlePlayerUpdate(*event.(*hockevent.PlayerUpdateEvent))
	case hockevent.EVENT_PLAYER_LIST:
		lab.handlePlayerList(*event.(*hockevent.PlayerListEvent))
	}
}
