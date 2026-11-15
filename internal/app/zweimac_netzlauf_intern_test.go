//go:build netzlauf

package app

// **Zwei Macs auf einem** — die Messung zu Phase 6.7
// (#1668):
// tragen Verbindung und Leistung für Spielen und Handeln, wenn zwei
// Anwendungen einander über das Netz erreichen?
//
// # Was hier wirklich geschieht
//
// **Zwei Anwendungen mit je eigenem Bestand, eigener Identität und eigenen
// Kontakten** — die zwei Macs bis auf die Hardware. Dazwischen liegt alles,
// was der Betrieb hat: der Horchposten, TCP, TLS 1.3 festgenagelt auf die
// Kennung, der Handschlag, die Partie mit Wirt und Gast und ihr Tischbild.
// Keine eingesetzte Leitung und kein eingesetzter Wähler.
//
// **Gefunden wird über echtes mDNS** auf den tauglichen Schnittstellen dieses
// Rechners — die Pakete gehen ins Netz und kommen über die
// Multicast-Rückführung wieder an.
//
// **Die Wege**: die Schleife, die Adresse von `en0` und dieselbe über einen
// Vermittler, der jede Richtung um 10, 40 und 100 ms verzögert — ein Netz im
// Haus, eines im selben Land und eines über einen Ozean. Der Vermittler
// verzögert **jedes Stück um dieselbe Zeit, ohne den Fluss anzuhalten**: er
// stempelt beim Lesen und schreibt, wenn die Zeit um ist, wie eine Leitung
// mit Laufzeit und ohne Engpass.
//
// # Was er **nicht** belegt
//
// **Dass eine Verbindung zu `en0` dieses Rechners über die Netzkarte geht**:
// macOS stellt ein Paket an die eigene Adresse über die Rückschleife zu. Der
// Weg `en0` misst deshalb dasselbe wie die Schleife, nur mit der Adresse,
// die ein anderer Mac wählen würde. **Verlust, Engpass, Router, NAT, WLAN und
// Firewall** misst er nicht; die Laufzeit kommt vom Vermittler und nicht vom
// Netz. Das bleibt die Abnahme mit zwei Rechnern (QA-10.AK-04).
//
//	go test -tags netzlauf ./internal/app/ -run ZweiMacs -v -count=1

import (
	"context"
	"fmt"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/peer"
)

// zweiMacs baut Anna und Bert ohne eingesetzte Leitung.
func zweiMacs(t *testing.T) (anna, bert *seite) {
	t.Helper()

	anna, bert = &seite{}, &seite{}

	for _, wer := range []*seite{anna, bert} {
		wer.anwendung = PeerflaecheUeber(bestandFuerDenRiegel(t)).MitPeerbund(&peerprobebund{})
		wer.ctx = wer.anwendung.imProfil(riegelkontext(t))

		antwort := wer.anwendung.PeeridentitaetSicherstellen(wer.ctx)
		if antwort.Fehler != nil {
			t.Fatalf("Identität anlegen: %+v", antwort.Fehler)
		}

		wer.kennung = peer.Kennung(antwort.Kennung)
		wer.abdruck = antwort.Fingerabdruck
		wer.meldungen = make(chan string, 256)
		wer.nebenbei = make(chan NebenbeiDTO, 256)
		wer.anwendung.melden = wer.sammeln
	}

	anna.kennt(t, bert)
	bert.kennt(t, anna)

	return anna, bert
}

// ergebnis ist eine Zeile der Messung.
type ergebnis struct {
	weg                         string
	aufbau, chat, zug, beitritt messreihe
}

func TestZweiMacsAufEinem(t *testing.T) {
	anna, bert := zweiMacs(t)
	ip := en0(t)
	bilder := bilderVon(bert)

	// **Anna horcht auf jeder Schnittstelle** — dieselbe Form, die ein
	// Mensch wählt, der von überall erreichbar sein will.
	gestartet := anna.anwendung.HorchStarten(anna.ctx, ":0")
	if gestartet.Fehler != nil {
		t.Fatalf("HorchStarten: %+v", gestartet.Fehler)
	}

	t.Cleanup(anna.anwendung.horchenBeenden)

	_, port, fehler := net.SplitHostPort(gestartet.Stand.Adresse)
	if fehler != nil {
		t.Fatalf("die Horchadresse %q: %v", gestartet.Stand.Adresse, fehler)
	}

	entdeckung := findeUeberMDNS(t, anna, bert)

	MitPartienleitung(anna.anwendung, anna.anwendung.db)
	partieID, objekt := einTischBeimWirt(t, anna)

	if antwort := anna.anwendung.PartieFreigeben(
		anna.ctx, partieID, "anna", []string{"bert"}); antwort.Fehler != nil {
		t.Fatalf("PartieFreigeben: %+v", antwort.Fehler)
	}

	aufEn0 := net.JoinHostPort(ip, port)
	wege := []struct {
		name, adresse string
	}{
		{"Schleife", net.JoinHostPort("127.0.0.1", port)},
		{"en0", aufEn0},
		{"en0 + 10 ms je Richtung", verzoegerer(t, aufEn0, 10*time.Millisecond)},
		{"en0 + 40 ms je Richtung", verzoegerer(t, aufEn0, 40*time.Millisecond)},
		{"en0 + 100 ms je Richtung", verzoegerer(t, aufEn0, 100*time.Millisecond)},
	}

	zeilen := make([]ergebnis, 0, len(wege))
	zug := 0

	for _, weg := range wege {
		zeile := ergebnis{weg: weg.name}

		if antwort := bert.anwendung.PeerkontaktAdresse(
			bert.ctx, string(anna.kennung), weg.adresse); antwort.Fehler != nil {
			t.Fatalf("%s: PeerkontaktAdresse: %+v", weg.name, antwort.Fehler)
		}

		// **Aufbau**: Wählen, TLS, Handschlag — und Anna weist ab, weil
		// Teilen hier keinen Weg hat. Gemessen wird, bis der Handschlag steht.
		for range 10 {
			anfang := time.Now()

			griff, oeffnen := bert.anwendung.sitzungsverwaltung().Oeffne(
				bert.ctx, anna.kennung, domain.ZweckTeilen)
			if oeffnen != nil {
				t.Fatalf("%s: Oeffne: %v", weg.name, oeffnen)
			}

			zeile.aufbau = append(zeile.aufbau, time.Since(anfang))
			_ = griff.Schliesse() //nolint:errcheck // gemessen ist schon
		}

		// **Beitritt**: der Gast setzt sich an Annas Tisch.
		anfang := time.Now()
		if antwort := bert.anwendung.PartieBeitreten(
			bert.ctx, string(anna.kennung), partieID, PartiemitbringDTO{}); antwort.Fehler != nil {
			t.Fatalf("%s: PartieBeitreten: %+v", weg.name, antwort.Fehler)
		}

		zeile.beitritt = append(zeile.beitritt, time.Since(anfang))

		bert.altesVerwerfen()
		anna.altesVerwerfen()

		// **Ein Satz am Tisch**: vom Gast gesagt, beim Wirt angekommen.
		for runde := range 20 {
			text := fmt.Sprintf("%s-%d", weg.name, runde)
			anfang := time.Now()

			if antwort := bert.anwendung.PartieSagen(bert.ctx, text); antwort.Fehler != nil {
				t.Fatalf("%s: PartieSagen: %+v", weg.name, antwort.Fehler)
			}

			warteAufNebenbeiText(t, anna.nebenbei, text)
			zeile.chat = append(zeile.chat, time.Since(anfang))
		}

		// **Ein Zug hin und das Tischbild zurück** — was ein Spieler fühlt:
		// tappen beim Gast, wirken beim Wirt, sehen beim Gast.
		for runde := range 20 {
			art := "objekt.tappen"
			if zug%2 == 1 {
				art = "objekt.enttappen"
			}

			zug++
			anfang := time.Now()

			if antwort := bert.anwendung.PartieWuenschen(bert.ctx, domain.Spielbefehl{
				Kennung: fmt.Sprintf("z%d", zug), Art: art,
				Nutzlast: map[string]any{"objekt": objekt},
			}); antwort.Fehler != nil {
				t.Fatalf("%s: PartieWuenschen %d: %+v", weg.name, runde, antwort.Fehler)
			}

			warteAufBildMit(t, bilder, objekt, art == "objekt.tappen")
			zeile.zug = append(zeile.zug, time.Since(anfang))
		}

		if antwort := bert.anwendung.PartieBeitrittBeenden(bert.ctx); antwort.Fehler != nil {
			t.Fatalf("%s: PartieBeitrittBeenden: %+v", weg.name, antwort.Fehler)
		}

		zeilen = append(zeilen, zeile)
	}

	bericht := &strings.Builder{}
	fmt.Fprintf(bericht, "Entdeckung über mDNS: %s\n", entdeckung)

	for _, zeile := range zeilen {
		fmt.Fprintf(bericht, "\n%s\n  Aufbau   %s\n  Beitritt %s\n  Satz     %s\n  Zug      %s\n",
			zeile.weg, zeile.aufbau, zeile.beitritt, zeile.chat, zeile.zug)
	}

	t.Log("\n" + bericht.String())
}

// findeUeberMDNS kündigt Anna an und lässt Bert suchen, bis er sie findet.
func findeUeberMDNS(t *testing.T, anna, bert *seite) messreihe {
	t.Helper()

	if antwort := anna.anwendung.SichtbarWerden(anna.ctx, "Messlauf Anna"); antwort.Fehler != nil {
		t.Fatalf("SichtbarWerden: %+v", antwort.Fehler)
	}

	t.Cleanup(func() { anna.anwendung.SichtbarkeitBeenden(context.WithoutCancel(anna.ctx)) })

	reihe := make(messreihe, 0, 3)

	for range 3 {
		anfang := time.Now()
		gefunden := false

		for !gefunden && time.Since(anfang) < 20*time.Second {
			antwort := bert.anwendung.Nachbarn(bert.ctx)
			if antwort.Fehler != nil {
				t.Fatalf("Nachbarn: %+v", antwort.Fehler)
			}

			for _, nachbar := range antwort.Nachbarn {
				gefunden = gefunden || nachbar.Kennung == string(anna.kennung)
			}
		}

		if !gefunden {
			t.Fatal("Bert fand Anna nicht über mDNS")
		}

		reihe = append(reihe, time.Since(anfang))
	}

	return reihe
}

// bilderVon sammelt die Tischbilder, die beim Gast ankommen.
//
// **Gezählt wird das Bild mit dem erwarteten Zustand, nicht irgendeines.**
// Der erste Lauf wartete auf die nächste Meldung `partie:tischbild` — und
// fand oft eine, die noch zum vorigen Zug gehörte: Median 2 ms bei 200 ms
// Umlaufzeit.
func bilderVon(wer *seite) chan PartieAntwort {
	bilder := make(chan PartieAntwort, 1024)
	vorher := wer.anwendung.melden

	wer.anwendung.melden = func(ctx context.Context, ereignis string, daten ...any) {
		vorher(ctx, ereignis, daten...)

		if ereignis != EreignisPartietischbild || len(daten) < 2 {
			return
		}

		if bild, passt := daten[1].(PartieAntwort); passt {
			select {
			case bilder <- bild:
			default:
			}
		}
	}

	return bilder
}

// warteAufBildMit wartet auf ein Tischbild, in dem `objekt` so getappt ist.
func warteAufBildMit(t *testing.T, bilder chan PartieAntwort, objekt string, getappt bool) {
	t.Helper()

	frist := time.After(10 * time.Second)

	for {
		select {
		case bild := <-bilder:
			for _, eines := range bild.Partie.Objekte {
				if eines.Kennung == objekt && eines.Getappt == getappt {
					return
				}
			}
		case <-frist:
			t.Fatalf("kein Tischbild mit %s getappt=%v", objekt, getappt)
		}
	}
}
