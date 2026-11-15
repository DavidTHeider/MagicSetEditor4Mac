package service

// **Ein Urteil, das es nicht gibt** — Werkstück G5 von Phase 5.5
// (#1462).
//
// # Warum dieser Lauf im Paket steht
//
// **Weil der Zweig von außen nicht erreichbar ist.** `Partieempfang.Nimm`
// gibt genau vier Urteile; ein fünftes entsteht erst, wenn jemand
// `domain.Folgeurteil` erweitert — und **dann** soll die Schleife anhalten
// statt die Nachricht fallen zu lassen. Ein Zweig, den kein Lauf betritt, ist
// eine Zusage, die niemand hält (§10), und ein `//nolint` darüber verdeckte
// nur, dass er ungemessen ist.
//
// **Was er nicht prüft**, steht daneben: dass die vier bekannten Urteile das
// Richtige tun. Das hält `partienetz_test.go` über die Schleife.

import (
	"context"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// innereKette ist ein Glied in der richtigen Länge.
//
// **Eine eigene Konstante**, weil `langeKette` in `service_test` steht: über
// die Paketgrenze reicht sie nicht, und eine zweite Datei dafür zu öffnen
// wäre mehr Aufwand als eine Zeile.
const innereKette = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

// stummerBote nimmt alles und sagt nichts.
type stummerBote struct{ bekommen int }

func (b *stummerBote) Angekommen(
	context.Context, domain.Partienachricht,
) error {
	b.bekommen++

	return nil
}

// **Ein unbekanntes Urteil hält an, statt die Nachricht zu verlieren.**
func TestEinUnbekanntesUrteilHaeltAn(t *testing.T) {
	t.Parallel()

	bote := &stummerBote{}
	netz := NeuesPartienetz(nil, NeuerPartieempfang(1), bote)

	fehler := netz.urteilen(context.Background(), Empfangsausgang{
		Nachricht: domain.Partienachricht{
			Art: domain.ArtEreignis, Partie: "p1", Epoche: 1, Folge: 1,
			Von: "sitz-2", Inhalt: []byte(`{}`), Kette: innereKette,
		},
		// **Eine Zahl, die kein Urteil ist** — genau der Fall, für den der
		// Zweig da ist.
		Urteil: domain.Folgeurteil(99),
	})

	switch {
	case fehler == nil:
		t.Fatal("das unbekannte Urteil kam durch")
	case bote.bekommen != 0:
		t.Error("die Nachricht ging trotzdem an den Boten")
	}
}
