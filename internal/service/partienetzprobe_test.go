package service_test

// **Die Vorrichtungen für die Schleife** — Werkstück G5 von Phase 5.5
// (#1462).
//
// **Eine eigene Datei**, weil `partienetz_test.go` mit ihnen an die
// Zeilengrenze aus §2 stößt. Der Schnitt folgt der Frage: dort steht, **was**
// geprüft wird, hier, **womit**.

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// errStromZuEnde beendet die Schleife in einem Lauf.
//
// **Ein eigener Fehler und kein `io.EOF`**: er sagt, dass der **Lauf** zu Ende
// ist, und nicht, dass eine Leitung abgerissen wäre — die beiden bedeuten
// Verschiedenes, und ein Lauf, der sie verwechselte, prüfte den falschen Weg.
var errStromZuEnde = errors.New("der Lauf hat nichts mehr")

// zugangsdoppel gibt vorgegebene Zustellungen aus.
type zugangsdoppel struct {
	zustellungen [][]byte
	gelesen      int

	// veraltet macht aus jeder Zustellung eine, die nicht mehr gilt.
	veraltet bool

	lesefehler error
	gesendet   [][]byte

	// geleert zählt, wie oft die Schlange hinausgeschickt wurde.
	geleert    int
	leerfehler error
}

func (z *zugangsdoppel) Sende(
	_ context.Context, nachricht []byte, _ service.Zustellung,
) error {
	z.gesendet = append(z.gesendet, nachricht)

	return nil
}

func (z *zugangsdoppel) Empfange(context.Context) ([]byte, error) {
	if z.lesefehler != nil {
		return nil, z.lesefehler
	}

	if z.gelesen >= len(z.zustellungen) {
		return nil, errStromZuEnde
	}

	z.gelesen++

	return z.zustellungen[z.gelesen-1], nil
}

func (z *zugangsdoppel) Empfangen([]byte) bool { return !z.veraltet }

func (z *zugangsdoppel) Leere(context.Context) error {
	z.geleert++

	return z.leerfehler
}

// botendoppel merkt sich, was ankam.
type botendoppel struct {
	bekommen []domain.Partienachricht
	klagt    error
}

func (b *botendoppel) Angekommen(
	_ context.Context, nachricht domain.Partienachricht,
) error {
	b.bekommen = append(b.bekommen, nachricht)

	return b.klagt
}

// verpackt macht aus einer Nachricht eine Zustellung.
func verpackt(t *testing.T, nachricht domain.Partienachricht) []byte {
	t.Helper()

	roh, fehler := json.Marshal(nachricht)
	if fehler != nil {
		t.Fatalf("verpacken: %v", fehler)
	}

	return roh
}

// einEreignis gibt ein Ereignis mit Nummer und Kette.
func einEreignis(folge uint64) domain.Partienachricht {
	return domain.Partienachricht{
		Art: domain.ArtEreignis, Partie: "p1", Epoche: 1, Folge: folge,
		Von: "sitz-2", Inhalt: []byte(`{"art":"objekt.tappen"}`),
		Kette: langeKette,
	}
}

// netzUeber baut die Schleife über einem Doppel — in der ersten Epoche.
func netzUeber(
	zugang service.Partiezugang, bote service.Partienbote,
) *service.Partienetz {
	return netzInEpoche(zugang, bote, 1)
}

// netzInEpoche baut sie in einer bestimmten Epoche.
//
// **Für die alte Epoche gebraucht**: sie ist nur alt im Verhältnis zu einer
// neueren, und eine Nachricht **ohne** Epoche ist etwas anderes — die weist
// schon `domain.Partienachricht.Pruefe` ab, und dann prüfte der Lauf den
// Weg für unlesbare Zustellungen statt den für alte Stände.
func netzInEpoche(
	zugang service.Partiezugang, bote service.Partienbote, epoche uint64,
) *service.Partienetz {
	return service.NeuesPartienetz(
		zugang, service.NeuerPartieempfang(epoche), bote)
}
