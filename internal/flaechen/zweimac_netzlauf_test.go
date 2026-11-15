//go:build netzlauf

package flaechen_test

// **Zwei Macs auf einem: ein Handel über die Leitung** — die Messung zu
// Phase 6.7 (#1668).
//
// # Was hier wirklich geschieht
//
// **Zwei Prozesse je Seite, wie im Betrieb**: Anna horcht in einem eigenen
// `mse horchen`, Bert stellt mit `mse handelszustellung` zu — jeder mit
// eigenem Bestand, Bund und Schlüssel. Gemessen wird die Wanduhr eines
// Aufrufs, **abzüglich** eines Aufrufs, der dasselbe Programm startet und
// denselben Bestand öffnet, aber nichts hinüberträgt (`handelspost`). Was
// bleibt, ist Wählen, TLS, Handschlag, Tragen und die Antwort.
//
// **Die Wege**: die Schleife und derselbe Weg über den [vermittler], der jedes
// Stück um 10, 40 und 100 ms verzögert.
//
// # Was er **nicht** belegt
//
// **Ein Netz.** Der Vermittler verzögert ein Stück, bevor er es weitergibt;
// er kennt keinen Verlust, keinen Engpass und kein NAT — das bleibt die
// Abnahme mit zwei Rechnern (QA-10.AK-04).
//
//	go test -tags netzlauf ./internal/flaechen/ -run ZweiMacs -v -count=1

import (
	"fmt"
	"slices"
	"sort"
	"strings"
	"testing"
	"time"
)

// einHandelNr legt bei Bert einen Handel mit Anna in den Ausgang — mit
// eigener Kennung je Runde — und gibt die Kennung der Nachricht.
func einHandelNr(t *testing.T, anna, bert einRechner, vorgang string) string {
	t.Helper()

	positionen := `[{"von":"` + anna.kennung +
		`","printing":"bolt-m10","menge":1,"guete":"nearMint",` +
		`"sprache":"de","finish":"nonfoil"}]`

	mitEingabe(t, bert.umgebung, positionen,
		"handel", "eroeffnen", vorgang, "tausch", anna.kennung)
	mussKommandozeile(t, bert.umgebung, "handel", "anbieten", vorgang)
	return strings.TrimSpace(mussKommandozeile(t, bert.umgebung, "handelspost-schicken", vorgang, anna.kennung))
}

// gestoppt misst die Wanduhr eines Aufrufs und gibt seine Ausgabe.
func gestoppt(t *testing.T, rechner einRechner, argumente ...string) (dauer time.Duration, aus string) {
	t.Helper()

	anfang := time.Now()
	aus = mussKommandozeile(t, rechner.umgebung, argumente...)

	return time.Since(anfang), aus
}

func zusammengefasst(reihe []time.Duration) string {
	sortiert := slices.Clone(reihe)
	sort.Slice(sortiert, func(i, j int) bool { return sortiert[i] < sortiert[j] })

	return fmt.Sprintf("Median %s · max %s (n=%d)",
		sortiert[len(sortiert)/2].Round(100*time.Microsecond),
		sortiert[len(sortiert)-1].Round(100*time.Microsecond), len(sortiert))
}

func TestZweiMacsHandelnAufEinem(t *testing.T) {
	anna, bert := zweiImLabor(t)
	adresse, _ := horcheImHintergrund(t, anna)
	weg := neuerVermittler(t, adresse)

	mussKommandozeile(t, bert.umgebung, "kontakt-adresse", anna.kennung, weg.adresse())

	bericht := &strings.Builder{}
	runde := 0

	for _, laufzeit := range []time.Duration{0, 10 * time.Millisecond, 40 * time.Millisecond, 100 * time.Millisecond} {
		weg.verzoegern(laufzeit)

		getragen := make([]time.Duration, 0, 5)
		leer := make([]time.Duration, 0, 5)

		for range 5 {
			runde++
			nachricht := einHandelNr(t, anna, bert, fmt.Sprintf("v-mess-%d", runde))

			// **Der leere Aufruf zuerst**: derselbe Start, derselbe Bestand,
			// kein Weg hinüber.
			dauer, _ := gestoppt(t, bert, "handelspost", anna.kennung)
			leer = append(leer, dauer)

			// **Getragen heißt: einer ging hinüber** — sonst wäre die Zahl
			// eine über einen Weg ins Leere.
			dauer, aus := gestoppt(t, bert, "handelszustellung", anna.kennung)
			if !strings.Contains(aus, "\t1") {
				t.Fatalf("Runde %d: getragen wurde %q", runde, aus)
			}

			getragen = append(getragen, dauer)

			// **Und dann gilt sie als angekommen** — sonst trüge die nächste
			// Runde sie noch einmal: der Ausgang räumt sich erst mit der
			// Bestätigung von drüben (`T5`), und die nimmt hier die
			// Kommandozeile vorweg.
			mussKommandozeile(t, bert.umgebung, "handelspost-bestaetigen", nachricht)
		}

		netto := make([]time.Duration, len(getragen))
		for stelle := range getragen {
			netto[stelle] = getragen[stelle] - leer[stelle]
		}

		fmt.Fprintf(bericht, "\n+%s je Stück\n  Zustellung gesamt %s\n  Programmstart     %s\n  Leitung netto     %s\n",
			laufzeit, zusammengefasst(getragen), zusammengefasst(leer), zusammengefasst(netto))
	}

	if weg.durchgereicht.Load() == 0 {
		t.Fatal("die Zustellungen liefen nicht über den Vermittler")
	}

	// **Und bei Anna steht, was ankam** — verbucht wird drüben nebenher,
	// der Befehl kehrt zurück, wenn die Bytes draußen sind; deshalb wird
	// gewartet und nicht einmal gefragt.
	letzter := fmt.Sprintf("v-mess-%d", runde)

	for versuch := 0; ; versuch++ {
		aus, _, code := kommandozeile(t, anna.umgebung, "handel", "stand", letzter)
		if code == 0 && strings.Contains(aus, "angeboten") {
			break
		}

		if versuch == 50 {
			t.Fatalf("der letzte Handel steht bei Anna nicht: %q", aus)
		}

		time.Sleep(100 * time.Millisecond)
	}

	t.Log("\n" + bericht.String())
}
