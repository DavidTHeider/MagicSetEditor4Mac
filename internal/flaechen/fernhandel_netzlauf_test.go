//go:build netzlauf

package flaechen_test

// **Ein Handel über fremde Netze, direkt und über das Relais** — D04
// (#2235).
//
// Zwei Prozesse mit eigenem Bestand und eigener Identität erreichen einander
// über `mse fernhorchen` und die **Einladung** — tailcat, also WireGuard über
// Tailscales Datenebene, mit DERP als Rendezvous. Darüber derselbe Ablauf
// wie in `handelsabnahme_test.go`: eröffnen, anbieten, zustimmen, festhalten,
// vorbereiten, übergeben, und am Ende **beide Stände nebeneinander**.
//
// **Zweimal**: so, wie tailcat wählt (direkt, sobald die Pfadsuche ihn
// findet), und mit `TS_DEBUG_ALWAYS_USE_DERP` in beiden Prozessen — dann
// bindet magicsock kein UDP, und jedes Byte geht über ein Relais im Internet
// (`internal/fernleitung/rueckfall_netzlauf_test.go` misst, dass der Schalter
// das tut).
//
// # Was er nicht belegt
//
//   - **Zwei NAT-Grenzen.** Beide Prozesse laufen auf diesem Mac; der direkte
//     Weg ist die Schleife. Das ist die Abnahme mit zwei Rechnern hinter je
//     eigenem Router, und es gibt sie nicht (`out-of-scope`).
//   - **Die Erlaubnis „Lokales Netzwerk" der Oberfläche** — die fragt macOS
//     nur ein Fenster, und dieser Lauf hat keines.
//   - **Dass der Abschluss über den Draht reist** — er tut es mit Absicht
//     nicht (`docs/ABNAHME-T12.md`, „Wo der Draht aufhört"): `verbuchen`,
//     `abgleichen` und `bestaetigen` geht jede Seite selbst, mit einem
//     Nachweis, den nur das Fenster ausstellt. Der Lauf fährt sie deshalb
//     **auf jeder Maschine**, wie ein Mensch es täte (#2285), und
//     vergleicht danach die Bestände.
//
//	go test -tags netzlauf ./internal/flaechen/ -run FremdeNetze -v -count=1

import (
	"bufio"
	"context"
	"encoding/json"
	"maps"
	"os/exec"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// fernhorcheImHintergrund startet `mse fernhorchen` und gibt die Einladung.
func fernhorcheImHintergrund(t *testing.T, rechner einRechner) string {
	t.Helper()

	ctx, beenden := context.WithCancel(t.Context())

	// #nosec G204 -- gebautes Programm aus TestMain, feste Argumente
	lauf := exec.CommandContext(ctx, gebaut.mse, "fernhorchen")
	lauf.Env = rechner.umgebung
	lauf.Stderr = &mitschrift{}

	aus, fehler := lauf.StdoutPipe()
	if fehler != nil {
		t.Fatalf("die Ausgabe abgreifen: %v", fehler)
	}

	if gestartet := lauf.Start(); gestartet != nil {
		t.Fatalf("fernhorchen starten: %v", gestartet)
	}

	t.Cleanup(func() {
		beenden()
		_ = lauf.Wait() //nolint:errcheck // der Abbruch ist der Weg hinaus
	})

	zeile, fehler := bufio.NewReader(aus).ReadString('\n')
	if fehler != nil {
		t.Fatalf("die Einladung lesen: %v (%s)", fehler, lauf.Stderr)
	}

	einladung := strings.TrimSpace(zeile)
	if !domain.IstEinladung(einladung) {
		t.Fatalf("fernhorchen nannte %q und keine Einladung", einladung)
	}

	return einladung
}

// zweiRechnerUeberFremdeNetze stellt beide auf, lässt beide über fremde Netze
// horchen und trägt die Einladungen ein.
func zweiRechnerUeberFremdeNetze(t *testing.T, nurRelais bool) (anna, bert einRechner) {
	t.Helper()

	anna, bert = stelleAuf(t), stelleAuf(t)

	if nurRelais {
		anna.umgebung = append(anna.umgebung, "TS_DEBUG_ALWAYS_USE_DERP=true")
		bert.umgebung = append(bert.umgebung, "TS_DEBUG_ALWAYS_USE_DERP=true")
	}

	kenneUndErlaube(t, anna, bert)
	kenneUndErlaube(t, bert, anna)

	mussKommandozeile(t, bert.umgebung, "kontakt-adresse", anna.kennung, fernhorcheImHintergrund(t, anna))
	mussKommandozeile(t, anna.umgebung, "kontakt-adresse", bert.kennung, fernhorcheImHintergrund(t, bert))

	return anna, bert
}

func TestEinHandelGehtUeberFremdeNetze(t *testing.T) {
	for _, weg := range []struct {
		name      string
		nurRelais bool
	}{{"wie tailcat wählt", false}, {"nur über das Relais", true}} {
		t.Run(weg.name, func(t *testing.T) {
			anna, bert := zweiRechnerUeberFremdeNetze(t, weg.nurRelais)

			annasStueck := eineKennungAus(t, mussKommandozeile(t, anna.umgebung,
				"--json", "exemplar", "bolt-m10", "3", "nearMint", "de", "nonfoil"))
			bertsStueck := eineKennungAus(t, mussKommandozeile(t, bert.umgebung,
				"--json", "exemplar", "swamp-m10", "2", "nearMint", "de", "nonfoil"))

			positionen := `[{"von":"` + anna.kennung +
				`","printing":"bolt-m10","menge":1,"guete":"nearMint",` +
				`"sprache":"de","finish":"nonfoil"},{"von":"` + bert.kennung +
				`","printing":"swamp-m10","menge":1,"guete":"nearMint",` +
				`"sprache":"de","finish":"nonfoil"}]`

			mitEingabe(t, anna.umgebung, positionen,
				"handel", "eroeffnen", derAbnahmevorgang, "tausch", bert.kennung)
			mussKommandozeile(t, anna.umgebung, "handel", "anbieten", derAbnahmevorgang)

			trageHinueber(t, anna, bert.kennung)
			warteAufZustand(t, bert, "angeboten")

			mussKommandozeile(t, bert.umgebung, "handel", "zustimmen", derAbnahmevorgang)
			trageHinueber(t, bert, anna.kennung)

			mussKommandozeile(t, anna.umgebung, "handel", "zustimmen", derAbnahmevorgang)
			warteAufZustand(t, anna, "angenommen")

			trageHinueber(t, anna, bert.kennung)
			warteAufZustand(t, bert, "angenommen")

			mussKommandozeile(t, anna.umgebung, "handel-festhalten",
				derAbnahmevorgang, annasStueck, "1", "1")
			mussKommandozeile(t, bert.umgebung, "handel-festhalten",
				derAbnahmevorgang, bertsStueck, "1", "1")

			mussKommandozeile(t, anna.umgebung, "handel", "vorbereiten", derAbnahmevorgang)
			trageHinueber(t, anna, bert.kennung)
			warteAufZustand(t, bert, "vorbereitet")

			mussKommandozeile(t, bert.umgebung, "handel", "uebergeben", derAbnahmevorgang)
			trageHinueber(t, bert, anna.kennung)

			mussKommandozeile(t, anna.umgebung, "handel", "uebergeben", derAbnahmevorgang)
			warteAufZustand(t, anna, "uebergabeOffen")

			trageHinueber(t, anna, bert.kennung)
			warteAufZustand(t, bert, "uebergabeOffen")

			// **Beide Stände nebeneinander** — derselbe Vorgang, dieselbe
			// Revision, dieselben Positionen, auf zwei Beständen.
			annas := mussKommandozeile(t, anna.umgebung, "--json", "handel", "stand", derAbnahmevorgang)
			berts := mussKommandozeile(t, bert.umgebung, "--json", "handel", "stand", derAbnahmevorgang)

			if vergleichbar(t, annas, anna) != vergleichbar(t, berts, bert) {
				t.Errorf("die Stände weichen ab:\nAnna: %s\nBert: %s", annas, berts)
			}

			for _, wer := range []einRechner{anna, bert} {
				zusagen := mussKommandozeile(t, wer.umgebung, "--json", "handelszusagen", derAbnahmevorgang)
				if !strings.Contains(zusagen, `"uebergeben": true`) {
					t.Errorf("die Zusage gilt nicht als übergeben: %s", zusagen)
				}
			}

			// **Bis zum Abschluss** (#2285) — auf jeder Maschine dieselben
			// drei Schritte, jeder mit einem Nachweis vom Fenster.
			for _, wer := range []einRechner{anna, bert} {
				schliesseAb(t, wer, anna.kennung, bert.kennung)
			}

			annas = mussKommandozeile(t, anna.umgebung, "--json", "handel", "stand", derAbnahmevorgang)
			berts = mussKommandozeile(t, bert.umgebung, "--json", "handel", "stand", derAbnahmevorgang)

			if vergleichbar(t, annas, anna) != vergleichbar(t, berts, bert) {
				t.Errorf("die Stände weichen nach dem Abschluss ab:\nAnna: %s\nBert: %s", annas, berts)
			}

			// **Und die Karten haben den Besitzer gewechselt** — je ein
			// Stück, in beide Richtungen.
			for _, fall := range []struct {
				wer    einRechner
				mengen map[string]int
			}{
				{anna, map[string]int{"bolt-m10": 2, "swamp-m10": 1}},
				{bert, map[string]int{"bolt-m10": 1, "swamp-m10": 1}},
			} {
				if hat := mengenJeDruck(t, fall.wer); !maps.Equal(hat, fall.mengen) {
					t.Errorf("der Bestand ist %v, erwartet %v", hat, fall.mengen)
				}
			}
		})
	}
}

// vergleichbar macht aus einem Stand, was beide Seiten gleich sehen müssen —
// alles außer `ich`, und `ich` muss die eigene Kennung sein.
//
// **Gemessen und nicht angenommen**: am ersten Lauf wich genau dieses eine
// Feld ab.
func vergleichbar(t *testing.T, roh string, wer einRechner) string {
	t.Helper()

	var stand map[string]any
	if fehler := json.Unmarshal([]byte(roh), &stand); fehler != nil {
		t.Fatalf("den Stand lesen: %v (%s)", fehler, roh)
	}

	if stand["ich"] != wer.kennung {
		t.Errorf("der Stand nennt als ich %v, nicht %s", stand["ich"], wer.kennung)
	}

	delete(stand, "ich")

	geordnet, _ := json.Marshal(stand) //nolint:errcheck,errchkjson // eben gelesen

	return string(geordnet)
}
