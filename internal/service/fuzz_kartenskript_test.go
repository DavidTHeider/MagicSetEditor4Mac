package service_test

// **Ein Kartenskript aus einem fremden Spielpaket** — QA-06c
// (#1976),
// Fläche 19 aus `docs/FUZZZIELE.md`.
//
// # Was es neben den Grenztests zu sehen gibt
//
// `kartenskriptgrenzen_test.go` hält drei genannte Fälle: eine Abbildung
// ohne `new_card`, eine Schleife, zu viele Karten. Hier kommen Skript **und**
// Spielskript aus fremder Hand — das Spielskript, weil es Funktionen
// vorgeben kann, die das Kartenskript ruft.
//
// # Was ein Fund ist — festgelegt **vor** dem ersten Lauf
//
//   - Panic · Hänger über die Frist hinaus
//   - **abgelehnt, und trotzdem ging etwas an die Ablage**
//   - eine Ablehnung ohne eine der drei Klassen
//   - **angelegt mit einem Feld, das `new_card` nicht setzen darf** — ein
//     Bildfeld oder eines, das das Spiel nicht kennt
//   - mehr Karten als [service.Kartenskriptgrenze]
//
// **Nicht** geprüft wird, dass zweimal dasselbe herauskommt: `random_shuffle`
// und seine Geschwister ziehen echten Zufall (`bibliothek_zufallsliste.go`).
//
//	go test ./internal/service -run Fuzz -fuzz FuzzKartenskript -fuzztime 2m -fuzzminimizetime 5s

import (
	"errors"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// kartenskriptfrist ist kurz, damit ein Lauf viele Eingaben schafft.
const kartenskriptfrist = 2 * time.Second

func FuzzKartenskript(f *testing.F) {
	for _, saat := range [][2]string{
		{`[ new_card([name: "Plains"]), new_card([name: "Island", rarity: "basic land"]) ]`, ``},
		{`[ [kunst: "/fremd/bild.png"] ]`, ``},
		{`[ new_card() + [kunst: "x"] ]`, ``},
		{`[ new_card([gibtsnicht: "x"]) ]`, ``},
		{`for x from 1 to 3 do [new_card([name: "k" + x])]`, ``},
		{`karten()`, `karten := { [ [name: "geschmuggelt", kunst: "/etc/passwd"] ] }`},
		{`random_shuffle([new_card(), new_card()])`, ``},
		{`filter_list([new_card(), [kunst: "x"]], filter: { true })`, ``},
		{`[ 1, "zwei", nil ]`, ``},
		{`{`, ``},
	} {
		f.Add(saat[0], saat[1])
	}

	felder := probefelder()
	setzbar := map[string]bool{"name": true, "rarity": true}

	f.Fuzz(func(t *testing.T, skript, spielskript string) {
		anleger := &anlegerprobe{}
		spiel := &spielprobe{
			skripte: []service.Kartenskriptangabe{{Name: "fremd", Skript: skript}},
			felder:  felder,
			skript:  spielskript,
		}

		beginn := time.Now()

		lauf, fehler := service.NeuerKartenskriptdienst(anleger, spiel).
			MitFrist(kartenskriptfrist).
			Ausfuehren(t.Context(), 1, "fremd")

		if dauer := time.Since(beginn); dauer > 5*kartenskriptfrist {
			t.Fatalf("das Skript lief %v — die Frist ist %v", dauer, kartenskriptfrist)
		}

		if fehler != nil {
			if anleger.rufe != 0 {
				t.Errorf("abgelehnt (%v), und trotzdem %d Rufe an die Ablage", fehler, anleger.rufe)
			}

			if !errors.Is(fehler, domain.ErrValidation) && !errors.Is(fehler, domain.ErrNotFound) &&
				!errors.Is(fehler, domain.ErrConflict) {
				t.Errorf("abgelehnt ohne Klasse: %v", fehler)
			}

			return
		}

		if len(anleger.karten) > service.Kartenskriptgrenze {
			t.Errorf("%d Karten — die Grenze ist %d", len(anleger.karten), service.Kartenskriptgrenze)
		}

		for stelle, karte := range anleger.karten {
			for _, stand := range karte {
				if !setzbar[stand.Feld] {
					t.Errorf("Karte %d trägt %q = %q — das darf new_card nicht setzen",
						stelle+1, stand.Feld, stand.Wert)
				}
			}
		}

		if lauf == nil {
			t.Error("angenommen ohne Lauf")
		}
	})
}
