//go:build fremdpakete

// **Die drei Wege, die man wartet** —
// #891,
// Werkstück 6 von Phase Y.
//
// # Warum es sie gibt
//
// Im ganzen Projekt gab es **null** Benchmarks. Was schnell genug ist, war
// damit eine Meinung; ein Rückschritt fiel erst auf, wenn jemand die Anwendung
// benutzte — und dann wusste niemand, welche Änderung ihn gebracht hat.
//
// # Warum eine Schwelle und nicht nur eine Zahl
//
// Ein Benchmark, der nur berichtet, fängt keinen Rückschritt: seine Ausgabe
// liest niemand zweimal. Jeder Weg hat deshalb eine **benannte** Schwelle, und
// der Lauf wird rot, wenn er sie reißt.
//
// # Was die Schwelle **nicht** sagt
//
// **Sie gilt für diese Maschine.** Gemessen ist ein Rechner, keine Plattform
// ([0007](../../docs/entscheidungen/0007-kein-entfernter-lauf.md)) — eine
// langsamere reißt sie, ohne dass am Code etwas falsch wäre. Sie ist ein
// Wächter gegen **Rückschritt**, nicht eine Zusage an Nutzende.
//
// Deshalb steht neben jeder Schwelle, **wann und woran** sie gemessen wurde.
//
// **Und sie liegt knapp genug, um eine Verdopplung zu fangen.** Der erste
// Anlauf ließ drei- bis fünffach Luft — „gefangen werden soll die
// Größenordnung, nicht das Rauschen" —, und der Zweitblick über Phase Y hat
// nachgerechnet, was das durchläßt: 190 ms beim Rastern statt 71, also das
// 2,7-fache bei **jedem Tastendruck**, und 3,9 s statt 1,744 beim Einfahren.
// Das ist keine Größenordnung mehr, sondern ein spürbarer Rückschritt.
//
// **Und wie eng, ist je Weg verschieden** — die Zahl steht bei der Schwelle:
//
//	einfahren klein   44–52 ms →  90 ms   (Faktor 1,7)
//	einfahren gross 1,74–1,82 s → 2,5 s   (Faktor 1,4)
//	Liste             10–11 ms →  18 ms   (Faktor 1,6)
//	rastern           71–78 ms → 120 ms   (Faktor 1,5)
//
// **Hier stand „zwischen 40 und 70 Prozent", und das war falsch**: bei
// `klein` waren es 240 Prozent und bei der Liste 150. Beide ließen eine
// Verdopplung grün — angestrichen vom Nachlauf des Zweitblicks über Phase Y.
// Eine pauschale Aussage über vier Zahlen ist eine Behauptung über jede
// einzelne, und drei davon hatten sie nicht verdient.
//
// **Keine lässt jetzt eine Verdopplung durch.** Wird ein Lauf davon flatterig,
// ist die richtige Antwort **mehr Durchläufe** — nicht, die Schwelle zu heben.
//
// # Und was sie **auch** nicht sagt
//
// **Sie misst den Go-Weg, nicht das Warten am Fenster.** Zwischen beiden
// liegen die Wails-Naht, der JSON-Zerteiler und alles, was die Seite daneben
// noch lädt. Gemessen am 2026-08-25: die Liste braucht hier 10 ms und im
// Fenster 529 ms für dasselbe Set. Wer wissen will, was jemand wartet, misst
// **im** Fenster — das steht in `docs/RELEASE.md` und ist Handarbeit.
//
// # Wie sie gefahren werden
//
//	MSE_KANON=~/Library/Application\ Support/MagicSetEditor4Mac/kanonische-sets \
//	MSE_FREMDPAKETE=/…/full-magic-pack/data \
//	  go test -tags fremdpakete ./internal/app/ -run XXX -bench Weg -benchtime 1x -v
//
// **`-run XXX`**, damit nur die Benchmarks laufen und nicht die Prüfläufe
// daneben. Sie gehören auf die Liste in `docs/RELEASE.md`, nicht in `make ci`
// (§4): das Einfahren von `Black.mse-set` allein dauert Minuten.
package app

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Die Schwellen, und woran sie gemessen sind.
//
// Gemessen am 2026-08-25 auf einem Apple-Silicon-Rechner, je mit
// `-benchtime 1x` für das Einfahren und `-benchtime 5x` für die beiden
// anderen. Der Messwert steht dabei; die Schwelle liegt darüber.
const (
	// schwelleEinfahrenKlein gilt für ein Set mittlerer Größe.
	//
	// `Vanguard.mse-set` hat 107 Karten; gemessen wurden **44 ms**. Das
	// Einfahren liest die Datei, deutet jede Karte und schreibt sie in den
	// Bestand.
	schwelleEinfahrenKlein = 90 * time.Millisecond

	// schwelleEinfahrenGross gilt für ein großes Set.
	//
	// `Black.mse-set` hat 3.788 Karten; gemessen wurden **1,744 s**. Hier
	// sieht jemand zu, während es läuft.
	//
	// **Von 4 s auf 2,5 s** — Befund 8 des Zweitblicks über Phase Y: bei 4 s
	// blieben 3,9 s grün, also mehr als das Doppelte. 2,5 s lassen 43 %
	// Schwankung und fangen jede Verdopplung.
	schwelleEinfahrenGross = 2500 * time.Millisecond

	// schwelleListe gilt für die Liste **eines großen** Sets.
	//
	// `Black.mse-set` hat 3.788 Karten; gemessen wurden **10 ms**. Das ist der
	// Weg, den das Fenster bei **jedem** Öffnen geht.
	//
	// **Sie stand einmal bei 50 ms, und das war zu viel.** Die Gegenprobe —
	// `Karten` statt `Kartenzeilen`, also genau der Rückschritt, gegen den es
	// diese Schwelle gibt — braucht 35 ms und wäre durchgegangen. Eine
	// Schwelle, die den Rückschritt nicht fängt, gegen den sie geschrieben
	// wurde, ist keine (§10).
	schwelleListe = 18 * time.Millisecond

	// schwelleRastern gilt für **eine** Karte in Bildschirmauflösung.
	//
	// Gemessen wurden **71 ms** (dreimal 71, einmal 76, einmal 78). **Neu am
	// 2026-10-15 mit den Schriften des Packs** (#2416): 64–65 ms in drei Läufen. Das ist
	// der Weg, den jeder Tastendruck im Kartenformular anstößt — er wird
	// abgebrochen und neu begonnen (§7), aber wenn er zu lange braucht, steht
	// die Vorschau hinter dem Tippen zurück.
	//
	// **Von 200 ms auf 120 ms** — Befund 8 des Zweitblicks über Phase Y: bei
	// 200 ms blieben 190 ms grün, also das 2,7-fache, und das ist bei **jedem
	// Tastendruck** zu spüren. 120 ms lassen 54 % über dem höchsten Messwert
	// und fangen jede Verdopplung.
	schwelleRastern = 120 * time.Millisecond
)

// Was die kanonischen Sets tragen — **gezählt, nicht geschätzt**.
//
// **Sie stehen hier, weil eine Schwelle allein nach oben prüft.** Ein Weg, der
// nichts tut, ist der schnellste; ohne diese Zahlen sagte ein leeres Ergebnis,
// die Anwendung sei schneller geworden. Das ist der zweite nachgeschärfte
// Punkt von #891:
// zu jeder zugesagten Zahl gehört eine Ober- **und** eine Untergrenze.
const (
	kartenKlein = 107
	kartenGross = 3788
)

// kanonsatz gibt den Pfad zu einem kanonischen Set oder überspringt.
func kanonsatz(b *testing.B, name string) string {
	b.Helper()

	kanon := os.Getenv("MSE_KANON")
	if kanon == "" {
		b.Skip("MSE_KANON nicht gesetzt — die kanonischen Sets liegen neben dem Repositorium (§11)")
	}

	// **Der Pfad kommt aus der Umgebung dessen, der misst.** Ein Messlauf
	// hinter `fremdpakete` läuft nie in der Pipeline und nie auf fremde
	// Eingabe.
	pfad := filepath.Join(kanon, name)
	if _, err := os.Stat(pfad); err != nil { // #nosec G703 -- siehe darüber
		b.Skipf("%s: %v", pfad, err)
	}

	return pfad
}

// messanwendung gibt eine gestartete Anwendung über dem vollen Pack.
//
// **Mit dem Paketverzeichnis**, denn ohne Stilpaket zeichnet der Rasterer den
// Behelfsplan — und der misst nichts, was jemand zu sehen bekommt.
func messanwendung(b *testing.B) *App {
	b.Helper()

	// **Ein eigenes Zuhause** (§3). `Starten` legt Bilder- und
	// Symbolverzeichnis unter `os.UserConfigDir()` an, und das ist auf darwin
	// `$HOME/Library/Application Support`. Ohne diese Zeile schriebe ein
	// Messlauf in den **echten** Bestand dessen, der misst — und läse dessen
	// Stand mit.
	//
	// Gefunden am 2026-08-24 in
	// #888:
	// ein Test hat zweimal 25 MB von Scryfall geladen, weil ihm beides fehlte.
	b.Setenv("HOME", b.TempDir())

	wurzel := os.Getenv("MSE_FREMDPAKETE")
	if wurzel == "" {
		b.Skip("MSE_FREMDPAKETE nicht gesetzt — ohne Stilpaket misst der Rasterer den Behelfsplan")
	}

	// **Und die Schriften des Packs** (#2416), über die eine Stelle, die sie
	// kennt ([schriftwurzelDesLaufs]) — vor dem Start, denn das Zeichenwerk
	// liest die Wurzel dabei. Bis hierher maß jede Messung, die zeichnet,
	// Karten ohne Text; [keineFehlendeSchrift] hält das jetzt.
	ort := schriftwurzelDesLaufs()
	if ort == "" {
		b.Skip("MSE_SCHRIFTWURZEL nicht gesetzt — ohne die Schriften des Packs misst der Rasterer Karten ohne Text")
	}

	b.Cleanup(MitSchriftwurzel(func() string { return ort }))

	anwendung := NeueApp(filepath.Join(b.TempDir(), "bestand.db"))
	anwendung.paketort = func() (string, error) { return wurzel, nil }

	if err := anwendung.Starten(context.Background()); err != nil {
		b.Fatalf("starten: %v", err)
	}

	b.Cleanup(func() {
		if fehler := anwendung.Beenden(); fehler != nil {
			b.Errorf("beenden: %v", fehler)
		}
	})

	return anwendung
}

// keineFehlendeSchrift macht den Lauf rot, wenn einer Karte eine Schrift
// fehlt — #2416.
//
// **Eine Messung ohne die Schriften des Packs misst Karten ohne Text**: sie
// rastert weniger, ist schneller, und ihre Schwelle hält nach oben, was
// niemand zu sehen bekommt. Gefunden an der Problemprüfung über `Black`
// (PERF02), die ohne Schriften 3.788 von 3.788 Karten mit Problem fand.
func keineFehlendeSchrift(b *testing.B, befunde []string) {
	b.Helper()

	for _, befund := range befunde {
		if strings.Contains(befund, "die Schrift") && strings.Contains(befund, "nicht zu finden") {
			b.Fatalf("der Messung fehlt eine Schrift — %s", befund)
		}
	}
}

// haelt macht den Lauf rot, wenn ein Weg seine Schwelle reißt.
//
// **Je Durchlauf und nicht über die Summe.** `b.Elapsed()` ist die Zeit aller
// Durchläufe; geteilt durch ihre Zahl ist es die eines einzelnen — und die ist
// das, was jemand wartet.
func haelt(b *testing.B, name string, schwelle time.Duration) {
	b.Helper()

	haeltZeit(b, name, b.Elapsed()/time.Duration(max(b.N, 1)), schwelle)
}

// haeltZeit ist dasselbe für eine **selbst gestoppte** Dauer.
//
// **Für einen Weg, den es nur einmal gibt.** Ein Kaltstart lässt sich nicht
// wiederholen — das zweite Mal ist warm. `b.Elapsed()/b.N` wäre dort trügerisch:
// wenn die Messung unter der Standard-`benchtime` bleibt, ruft der Rahmen die
// Funktion **noch einmal** und teilt durch eine Zahl, die nur für den warmen
// Weg gilt.
func haeltZeit(b *testing.B, name string, gemessen, schwelle time.Duration) {
	b.Helper()

	b.Logf("%s: %v, Schwelle %v", name, gemessen.Round(time.Millisecond), schwelle)

	if gemessen > schwelle {
		b.Errorf("%s braucht %v und reißt damit die Schwelle von %v — "+
			"entweder ist etwas langsamer geworden, oder diese Maschine ist "+
			"eine andere als die, an der gemessen wurde (siehe Kopf)",
			name, gemessen.Round(time.Millisecond), schwelle)
	}
}

// **Ein Set einfahren** — in zwei Größen.
//
// **Zwei, weil die Frage zwei ist.** Ein Set mittlerer Größe fährt jemand
// beiläufig ein; bei einem großen sieht er zu. Eine Schwelle über beide wäre
// entweder für das kleine bedeutungslos oder für das große falsch.
func BenchmarkWegEinfahren(b *testing.B) {
	for _, fall := range []struct {
		name     string
		satz     string
		karten   int
		schwelle time.Duration
	}{
		{"klein", "Vanguard.mse-set", kartenKlein, schwelleEinfahrenKlein},
		{"gross", "Black.mse-set", kartenGross, schwelleEinfahrenGross},
	} {
		b.Run(fall.name, func(b *testing.B) {
			quelle := kanonsatz(b, fall.satz)

			for b.Loop() {
				// **Eine frische Anwendung je Durchlauf.** Dasselbe Set
				// zweimal einzufahren gäbe beim zweiten Mal einen anderen
				// Weg — es liegt dann schon im Bestand.
				b.StopTimer()

				anwendung := messanwendung(b)

				b.StartTimer()

				geoeffnet := anwendung.SetOeffnen(context.Background(), quelle)
				if geoeffnet.Fehler != nil {
					b.Fatalf("einfahren: %+v", geoeffnet.Fehler)
				}

				// **Und es sind Karten dabei.** Ein Einfahren, das nichts
				// liest, ist schnell — und ohne diese Zeile grün.
				if geoeffnet.Karten < fall.karten {
					b.Fatalf("%d Karten eingefahren, erwartet mindestens %d",
						geoeffnet.Karten, fall.karten)
				}
			}

			haelt(b, "ein Set einfahren ("+fall.name+")", fall.schwelle)
		})
	}
}

// **Die Liste eines großen Sets laden.**
func BenchmarkWegListe(b *testing.B) {
	anwendung := messanwendung(b)
	setID := eingefahren(b, anwendung, "Black.mse-set")

	for b.Loop() {
		zeilen := anwendung.Kartenzeilen(context.Background(), setID)
		if zeilen.Fehler != nil {
			b.Fatalf("Kartenzeilen: %+v", zeilen.Fehler)
		}

		// **Und es sind Zeilen dabei.** Eine leere Antwort ist schnell; eine
		// Schwelle, die nur nach oben prüft, wäre dann grün und sagte, die
		// Liste sei schneller geworden (§10, zweiter nachgeschärfter Punkt
		// von #891).
		if len(zeilen.Zeilen) != kartenGross {
			b.Fatalf("%d Zeilen, erwartet %d", len(zeilen.Zeilen), kartenGross)
		}
	}

	haelt(b, "die Liste laden", schwelleListe)
}

// **Eine Karte rastern.**
func BenchmarkWegRastern(b *testing.B) {
	anwendung := messanwendung(b)
	setID := eingefahren(b, anwendung, "Black.mse-set")

	zeilen := anwendung.Kartenzeilen(context.Background(), setID)
	if zeilen.Fehler != nil {
		b.Fatalf("Kartenzeilen: %+v", zeilen.Fehler)
	}

	karteID := zeilen.Zeilen[0].ID

	// **Einmal vorab.** Der erste Lauf lädt Stilpaket, Schriften und Bilder;
	// gemessen wird, was **danach** jeder Tastendruck kostet.
	erste := anwendung.Vorschau(context.Background(), karteID, 1)
	if erste.Fehler != nil {
		b.Fatalf("Vorschau: %+v", erste.Fehler)
	}

	keineFehlendeSchrift(b, erste.Vorschau.Befunde)

	for b.Loop() {
		bild := anwendung.Vorschau(context.Background(), karteID, 1)
		if bild.Fehler != nil {
			b.Fatalf("Vorschau: %+v", bild.Fehler)
		}

		// **Und es ist ein Bild dabei.** Ein Rasterer, der nichts zeichnet,
		// ist der schnellste — und ohne diese Zeile der beste Messwert.
		//
		// **Was die Gegenprobe hier trifft**, ist der Fehlerzweig darüber: eine
		// Karte, die es nicht gibt, und eine Vergrößerung von null melden beide
		// einen Fehler. Belegt ist damit, dass ein fehlendes Bild auffällt —
		// nicht, welcher der beiden Zweige es fängt. Der Zustand „kein Bild und
		// kein Fehler" gibt es (`Karteneditor` meldet ihn als Befund); von
		// außen ist er nicht zu erzwingen.
		if bild.Vorschau == nil || bild.Vorschau.Breite == 0 {
			b.Fatalf("kein Bild: %+v", bild.Vorschau)
		}
	}

	haelt(b, "eine Karte rastern", schwelleRastern)
}

// eingefahren fährt ein kanonisches Set ein und gibt seine Kennung.
func eingefahren(b *testing.B, anwendung *App, name string) string {
	b.Helper()

	quelle := kanonsatz(b, name)

	// **Außerhalb der Messung**, denn gemessen wird, was das Fenster **bei
	// jedem** Öffnen tut, nicht was es einmal tut.
	b.StopTimer()

	geoeffnet := anwendung.SetOeffnen(context.Background(), quelle)
	if geoeffnet.Fehler != nil {
		b.Fatalf("einfahren: %+v", geoeffnet.Fehler)
	}

	b.Logf("%s mit %d Karten", name, geoeffnet.Karten)
	b.StartTimer()

	return geoeffnet.SetID
}
