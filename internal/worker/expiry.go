package worker

import (
	"log"
	"time"

	"github.com/imzami/routevpn-core/internal/models"
	"github.com/imzami/routevpn-core/internal/services"
	"github.com/imzami/routevpn-core/internal/wg"
)

type ExpiryWorker struct {
	svc    *services.Container
	ticker *time.Ticker
	done   chan struct{}
}

func NewExpiryWorker(svc *services.Container) *ExpiryWorker {
	return &ExpiryWorker{
		svc:  svc,
		done: make(chan struct{}),
	}
}

func (w *ExpiryWorker) Start() {
	w.ticker = time.NewTicker(1 * time.Hour)
	log.Println("[worker] expiry worker started (runs every hour)")

	// Run immediately on start
	w.run()

	for {
		select {
		case <-w.ticker.C:
			w.run()
		case <-w.done:
			w.ticker.Stop()
			log.Println("[worker] expiry worker stopped")
			return
		}
	}
}

func (w *ExpiryWorker) Stop() {
	close(w.done)
}

func (w *ExpiryWorker) run() {
	log.Println("[worker] checking for expired peers...")

	var peers []models.VPNPeer
	err := w.svc.DB.Select(&peers,
		"SELECT id, public_key FROM vpn_peers WHERE status='active' AND expiry_date < NOW()")
	if err != nil {
		log.Printf("[worker] query error: %v", err)
		return
	}

	if len(peers) == 0 {
		log.Println("[worker] no expired peers found")
		return
	}

	log.Printf("[worker] found %d expired peers, removing...", len(peers))

	for _, p := range peers {
		if err := wg.RemovePeer(w.svc.Cfg.WGInterface, p.PublicKey); err != nil {
			log.Printf("[worker] failed to remove peer %s from interface: %v", p.PublicKey[:8], err)
		}

		if _, err := w.svc.DB.Exec(
			"UPDATE vpn_peers SET status='expired' WHERE id=$1", p.ID); err != nil {
			log.Printf("[worker] failed to update peer %d status: %v", p.ID, err)
		} else {
			log.Printf("[worker] expired peer %d (key: %s...)", p.ID, p.PublicKey[:8])
		}
	}
}
