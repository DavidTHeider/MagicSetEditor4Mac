package rest_test

import (
	"sync"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

// **Guthaben und Sperre für die Tests des Dienstes** — Werkstück O5 von Phase
// 5.2 (#1429). `rest` kennt `service` nicht; was die Wurzel hereinreicht,
// stellen die Tests aus denselben Bausteinen der Domäne nach — `Guthaben`
// und `Sperrlage` —, damit die Zahlen hier dieselben sind wie im Betrieb.

type kontingentprobe struct {
	grenzen domain.Haushaltsgrenzen
	uhr     func() time.Time

	mu       sync.Mutex
	guthaben map[string]*domain.Guthaben
}

func neueKontingentprobe(grenzen domain.Haushaltsgrenzen, uhr func() time.Time) *kontingentprobe {
	return &kontingentprobe{grenzen: grenzen, uhr: uhr, guthaben: map[string]*domain.Guthaben{}}
}

func (k *kontingentprobe) Abbuchen(principal string, klasse domain.Kostenklasse, _ bool) (bool, time.Duration) {
	k.mu.Lock()
	defer k.mu.Unlock()

	guthaben, da := k.guthaben[principal]
	if !da {
		neu := domain.NeuesGuthaben(k.grenzen.Principal, k.uhr())
		guthaben = &neu
		k.guthaben[principal] = guthaben
	}

	return guthaben.Abbuchen(klasse.Kosten(), k.uhr())
}

type sperrprobe struct {
	uhr func() time.Time

	mu    sync.Mutex
	lagen map[string]domain.Sperrlage
}

func neueSperrprobe(uhr func() time.Time) *sperrprobe {
	return &sperrprobe{uhr: uhr, lagen: map[string]domain.Sperrlage{}}
}

func (s *sperrprobe) Gesperrt(nachweis string) (bool, time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.lagen[nachweis].Gesperrt(s.uhr())
}

func (s *sperrprobe) Fehlgeschlagen(nachweis string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.lagen[nachweis] = s.lagen[nachweis].Fehlgeschlagen(s.uhr())
}

func (s *sperrprobe) Gelungen(nachweis string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.lagen, nachweis)
}

// festeUhr ist die Uhr der Fabriken — ein fester Zeitpunkt.
func festeUhr() time.Time { return time.Date(2026, time.August, 23, 12, 0, 0, 0, time.UTC) }

// mitHaushalt gibt einem Dienst Guthaben und Sperre nach diesen Grenzen.
func mitHaushalt(t *testing.T, dienst *rest.Dienst, grenzen domain.Haushaltsgrenzen, uhr func() time.Time) *rest.Dienst {
	t.Helper()

	return dienst.MitGrenzen(grenzen).MitHaushalt(neueKontingentprobe(grenzen, uhr), neueSperrprobe(uhr))
}
