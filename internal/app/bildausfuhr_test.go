package app_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// bildprobe ist ein Double der Bildausfuhr.
type bildprobe struct {
	namen       []service.Bildname
	schreibung  *service.Bildschreibung
	fehler      error
	karteID     int64
	verzeichnis string
	datei       string
	format      service.Bildformat
	zielbreite  int

	// verhalten ist die Wahl beim Namenskonflikt (#491).
	verhalten string

	// guete ist die Wahl der Rastergüte (Ω.4, #1029).
	guete   render.Guete
	auswahl domain.Kartenauswahl
}

func (b *bildprobe) Namen(
	_ context.Context, _ int64, _ string, format service.Bildformat, verhalten string,
	auswahl domain.Kartenauswahl,
) ([]service.Bildname, error) {
	b.verhalten = verhalten

	b.format = format
	b.auswahl = auswahl

	return b.namen, b.fehler
}

func (b *bildprobe) Schreibe(
	_ context.Context, karteID int64, verzeichnis, datei string,
	format service.Bildformat, zielbreite int, verhalten string,
	guete render.Guete,
) (*service.Bildschreibung, error) {
	b.verhalten, b.guete = verhalten, guete

	b.karteID, b.verzeichnis, b.datei = karteID, verzeichnis, datei
	b.format, b.zielbreite = format, zielbreite

	return b.schreibung, b.fehler
}

func mitBildausfuhr(probe *bildprobe) *app.App {
	anwendung := app.NeueApp("")
	anwendung.MitBildausfuhr(probe)

	return anwendung
}

// **Die Namen kommen vor dem ersten Bild.**
//
// Die Oberfläche führt den Stapellauf; sie braucht die Liste vorher, um
// anzuzeigen, wie viele es werden — und um sie nicht selbst zu bilden.
func TestBildnamenKommenAlsListe(t *testing.T) {
	probe := &bildprobe{namen: []service.Bildname{
		{KarteID: 1, Nummer: 1, Datei: "001-Feuerball.png"},
		{KarteID: 2, Nummer: 2, Datei: "002-Eisball.png"},
	}}

	antwort := mitBildausfuhr(probe).Bildnamen(probekontext(t), "7", "{nummer}-{name}", "png", domain.BildkonfliktNummerierenNeu, nil)
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if len(antwort.Namen) != 2 {
		t.Fatalf("%d Namen", len(antwort.Namen))
	}

	// **Die Kennung als Zeichenkette**: eine `int64` überlebt den Weg durch
	// JavaScript nicht unbeschadet, und das gilt hier wie überall (§2).
	if antwort.Namen[0].KarteID != "1" || antwort.Namen[0].Datei != "001-Feuerball.png" {
		t.Errorf("%+v", antwort.Namen[0])
	}
}

// Ein Bild geht an das gewählte Verzeichnis, unter dem gewählten Namen.
func TestEinBildGehtAnDenGewaehltenOrt(t *testing.T) {
	probe := &bildprobe{schreibung: &service.Bildschreibung{
		Pfad: "/tmp/bilder/001-Feuerball.png", Breite: 750, Hoehe: 1046,
	}}

	antwort := mitBildausfuhr(probe).
		BildSchreiben(probekontext(t), "1", "/tmp/bilder", "001-Feuerball.png", "png", 750, domain.BildkonfliktUeberschreiben, "")
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if probe.karteID != 1 || probe.verzeichnis != "/tmp/bilder" ||
		probe.datei != "001-Feuerball.png" || probe.zielbreite != 750 {
		t.Errorf("beim Dienst kam %d, %q, %q, %d an",
			probe.karteID, probe.verzeichnis, probe.datei, probe.zielbreite)
	}

	if antwort.Breite != 750 || antwort.Hoehe != 1046 {
		t.Errorf("%+v", antwort)
	}
}

// **Was dem Bild fehlt, steht beim Erfolg** (§8).
func TestEinBildNenntSeineBefunde(t *testing.T) {
	probe := &bildprobe{schreibung: &service.Bildschreibung{
		Pfad:    "/tmp/001.png",
		Befunde: []domain.Befund{{Datei: "001.png", Ort: "Ebene 3", Meldung: "keine Schrift"}},
	}}

	antwort := mitBildausfuhr(probe).BildSchreiben(probekontext(t), "1", "/tmp", "001.png", "png", 0, domain.BildkonfliktUeberschreiben, "")

	if len(antwort.Befunde) != 1 || antwort.Befunde[0] == "" {
		t.Errorf("Befunde: %v", antwort.Befunde)
	}
}

// Die Fehlerwege — jeder übersetzt, keiner durchgereicht.
func TestBildfehlerWerdenUebersetzt(t *testing.T) {
	faelle := map[string]func(*app.App) *Fehlerträger{
		"eine Kennung, die keine ist": func(a *app.App) *Fehlerträger {
			antwort := a.Bildnamen(context.Background(), "sieben", "{name}", "png", domain.BildkonfliktNummerierenNeu, nil)

			return &Fehlerträger{antwort.Fehler, antwort.Namen == nil}
		},
		"ein Dienst, der nicht antwortet": func(a *app.App) *Fehlerträger {
			antwort := a.Bildnamen(context.Background(), "7", "{name}", "png", domain.BildkonfliktNummerierenNeu, nil)

			return &Fehlerträger{antwort.Fehler, antwort.Namen == nil}
		},
		"eine Kartenkennung, die keine ist": func(a *app.App) *Fehlerträger {
			antwort := a.BildSchreiben(context.Background(), "eins", "/tmp", "001.png", "png", 0, domain.BildkonfliktUeberschreiben, "")

			return &Fehlerträger{antwort.Fehler, antwort.Befunde == nil}
		},
		"ein Schreiben, das scheitert": func(a *app.App) *Fehlerträger {
			antwort := a.BildSchreiben(context.Background(), "1", "/tmp", "001.png", "png", 0, domain.BildkonfliktUeberschreiben, "")

			return &Fehlerträger{antwort.Fehler, antwort.Befunde == nil}
		},
	}

	for name, ruf := range faelle {
		t.Run(name, func(t *testing.T) {
			getragen := ruf(mitBildausfuhr(&bildprobe{fehler: errors.New("es ging nicht")}))

			if getragen.fehler == nil {
				t.Fatal("kein Fehler")
			}

			if getragen.leer {
				t.Error("eine Liste ist null — das Frontend läuft darüber, ohne zu prüfen")
			}
		})
	}
}

// Fehlerträger hält, was ein Fehlerweg zurückgibt.
type Fehlerträger struct {
	fehler *app.Fehler
	leer   bool
}

// **Die Güte kommt als Name herein und als Wert an** — Ω.4 aus
// #1029.
//
// Über die Flächen reist eine Zeichenkette, im Rasterer steht ein `int`.
// Zwischen beiden liegt genau eine Übersetzung ([app.GueteAusName]); dieser
// Test hält sie an der Naht fest, an der sie benutzt wird.
//
// **Der Weg über Kommandozeile, MCP und REST steht daneben** und wird in
// `internal/flaechen/guete_test.go` gefahren. Was hier geprüft wird, ist der
// Weg des **Fensters** — der einzige, den kein Go-Prozess von außen fährt.
func TestDieGueteAusDemFensterErreichtDieAusfuhr(t *testing.T) {
	t.Parallel()

	for name, gewollt := range map[string]render.Guete{
		"fast":   render.GueteSchnell,
		"medium": render.GueteMittel,
		"high":   render.GueteHoch,
		"ultra":  render.GueteHoechst,

		// **Leer ist kein Fehler, sondern die Vorgabe.** Ein Aufruf ohne Wahl
		// ist der Normalfall.
		"": render.GueteMittel,
	} {
		t.Run("die Güte "+name, func(t *testing.T) {
			t.Parallel()

			probe := &bildprobe{schreibung: &service.Bildschreibung{}}

			antwort := mitBildausfuhr(probe).BildSchreiben(probekontext(t), "1",
				"/tmp", "001.png", "png", 0, domain.BildkonfliktUeberschreiben, name)
			if antwort.Fehler != nil {
				t.Fatalf("%+v", antwort.Fehler)
			}

			if probe.guete != gewollt {
				t.Errorf("die Ausfuhr bekam %d statt %d", int(probe.guete), int(gewollt))
			}
		})
	}
}

// **Eine erfundene Güte wird abgewiesen, und die Meldung nennt die vier.**
//
// Ohne diesen Test fiele ein Tippfehler still auf die Vorgabe — und der Test
// darüber bliebe grün, weil er nur richtige Namen kennt.
func TestEineErfundeneGueteWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	probe := &bildprobe{schreibung: &service.Bildschreibung{}}

	antwort := mitBildausfuhr(probe).BildSchreiben(probekontext(t), "1",
		"/tmp", "001.png", "png", 0, domain.BildkonfliktUeberschreiben, "superb")
	if antwort.Fehler == nil {
		t.Fatal("eine erfundene Güte wurde angenommen")
	}

	for _, guete := range render.Gueten {
		if !strings.Contains(antwort.Fehler.Meldung, guete.Name()) {
			t.Errorf("die Meldung nennt %q nicht: %s", guete.Name(), antwort.Fehler.Meldung)
		}
	}

	// **Und geschrieben wurde nichts.** Ein abgewiesener Aufruf, der die
	// Ausfuhr trotzdem erreicht, hätte die Prüfung an der falschen Stelle.
	if probe.karteID != 0 {
		t.Error("die Ausfuhr wurde trotz Abweisung gerufen")
	}
}
