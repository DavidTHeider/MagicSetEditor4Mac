package app_test

// **Vorrang und Rückfall durch die ganze Anwendung** — AI-05.AK-05 und
// AK-06 (#2042).
//
// Das Gerätemodell ist eine Attrappe an einem Unix-Socket in diesem Prozess
// (die gemessene Gestalt, Vermerk 0166); Ollama ist das vorhandene Doppel.
// **Kein Lauf startet `/usr/bin/fm`** (§3).

import (
	"context"
	"net"
	"net/http"
	"os"
	"slices"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/applefm"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

const (
	// fmName ist eine Antwort mit einem Namen.
	fmName = `{"model":"system","choices":[{"finish_reason":"stop",` +
		`"message":{"content":"{\"name\":\"Funke\"}","refusal":null}}]}`

	// fmLeer ist ein Name ohne Inhalt — gemessen am 2026-10-03 (Vermerk
	// 0166): das Modell lässt das Feld manchmal leer.
	fmLeer = `{"model":"system","choices":[{"finish_reason":"stop",` +
		`"message":{"content":"{\"name\":\"\"}","refusal":null}}]}`

	// fmAblehnung ist die zweite gemessene Gestalt einer Ablehnung.
	fmAblehnung = `{"model":"system","choices":[{"finish_reason":"stop",` +
		`"message":{"content":"Diese Anfrage kann nicht bearbeitet werden.","refusal":null}}]}`
)

// mitGeraet setzt ein Gerätemodell an die Welt und zählt seine Starts.
func mitGeraet(t *testing.T, welt kiwelt, einrichtung applefm.Einrichtung, antwort string) *atomic.Int32 {
	t.Helper()

	ordner, fehler := os.MkdirTemp("", "fm")
	if fehler != nil {
		t.Fatalf("%v", fehler)
	}

	t.Cleanup(func() { _ = os.RemoveAll(ordner) }) //nolint:errcheck // nur zum Aufräumen

	gestartet := &atomic.Int32{}

	welt.anwendung.MitGeraetemodell(
		func(context.Context) applefm.Einrichtung { return einrichtung },
		func(ctx context.Context, argumente ...string) (applefm.Prozess, error) {
			gestartet.Add(1)

			horcher, err := (&net.ListenConfig{}).Listen(ctx, "unix", argumente[len(argumente)-1])
			if err != nil {
				return nil, err
			}

			t.Cleanup(func() { _ = horcher.Close() }) //nolint:errcheck // nur zum Aufräumen

			go func() {
				_ = http.Serve(horcher, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { //nolint:gosec,errcheck // Probe
					if r.URL.Path == applefm.Gesundheitsweg {
						_, _ = w.Write([]byte(`{"models":[{"available":true,"name":"system"}]}`)) //nolint:errcheck // Probe

						return
					}

					_, _ = w.Write([]byte(antwort)) //nolint:errcheck // Probe
				}))
			}()

			return fmProbe{horcher: horcher}, nil
		}, ordner)

	return gestartet
}

type fmProbe struct{ horcher net.Listener }

func (p fmProbe) Beende() error { return p.horcher.Close() }

// vorrang fragt mit Vorrang nach dem Namen.
func vorrang(t *testing.T, welt kiwelt, nur ...string) app.KIVorschlagAntwort {
	t.Helper()

	return welt.anwendung.KIVorschlagenMit(probekontext(t), app.KIVorschlagsbezug{
		KarteID: welt.karte, Anbieter: domain.KIVorrang, Auftrag: "Ein Name", Nur: nur,
	})
}

// ollamaAus schaltet die lokale Laufzeit aus.
func ollamaAus(t *testing.T, welt kiwelt) {
	t.Helper()

	if fehler := storage.NeueEinstellungen(welt.db).Setzen(probekontext(t),
		domain.KILaufzeit, domain.KILaufzeitAus); fehler != nil {
		t.Fatalf("%v", fehler)
	}
}

// **Eine kleine Aufgabe geht zuerst an das Gerät** — und Ollama wird nicht
// gefragt.
func TestDerVorrangFragtZuerstDasGeraet(t *testing.T) {
	t.Parallel()

	welt := einekiwelt(t)
	gestartet := mitGeraet(t, welt, applefm.EinrichtungBereit, fmName)

	antwort := vorrang(t, welt, "name")
	if antwort.Fehler != nil || antwort.Anbieter != "apple" || antwort.Modell != applefm.Modell ||
		antwort.Weg != string(domain.KIWegVorrang) || len(antwort.Felder) != 1 ||
		antwort.Felder[0].Werte[0] != "Funke" {
		t.Fatalf("%+v", antwort)
	}

	// **`auftraege` und nicht `gefragt`**: das Doppel zählt unter `gefragt`
	// die Blicke in den Katalog, und die gibt es bei jedem Vorschlag.
	if gestartet.Load() != 1 || len(welt.laufzeit.auftraege) != 0 {
		t.Errorf("gestartet %d, Ollama gefragt %d", gestartet.Load(), len(welt.laufzeit.auftraege))
	}
}

// **Lehnt das Gerät ab, fragt der Vorrang den nächsten lokalen** — und sagt,
// dass es ein Rückfall war.
func TestDerVorrangFaelltNachEinerAblehnungZurueck(t *testing.T) {
	t.Parallel()

	welt := einekiwelt(t)
	mitGeraet(t, welt, applefm.EinrichtungBereit, fmAblehnung)

	antwort := vorrang(t, welt, "name")
	if antwort.Fehler != nil || antwort.Anbieter != "ollama" ||
		antwort.Weg != string(domain.KIWegRueckfall) || len(welt.laufzeit.auftraege) != 1 {
		t.Fatalf("%+v", antwort)
	}
}

// **Ein leerer Wert ist kein Vorschlag** — Z1 des Zweitblicks über Phase
// 6.10: das Feld kam, aber ohne Inhalt, und der Rückfall blieb aus. Der
// nächste lokale Anbieter wird gefragt wie nach einer Ablehnung.
func TestEinLeererWertFaelltZurueck(t *testing.T) {
	t.Parallel()

	welt := einekiwelt(t)
	mitGeraet(t, welt, applefm.EinrichtungBereit, fmLeer)

	antwort := vorrang(t, welt, "name")
	if antwort.Fehler != nil || antwort.Anbieter != "ollama" ||
		antwort.Weg != string(domain.KIWegRueckfall) || len(welt.laufzeit.auftraege) != 1 {
		t.Fatalf("%+v", antwort)
	}
}

// **Eine große Aufgabe geht nie an das Gerät.**
func TestEineGrosseAufgabeGehtNieAnsGeraet(t *testing.T) {
	t.Parallel()

	welt := einekiwelt(t)
	gestartet := mitGeraet(t, welt, applefm.EinrichtungBereit, fmName)

	antwort := vorrang(t, welt)
	if antwort.Fehler != nil || antwort.Anbieter != "ollama" || antwort.Weg != string(domain.KIWegVorrang) {
		t.Fatalf("%+v", antwort)
	}

	if gestartet.Load() != 0 {
		t.Errorf("das Gerät wurde gestartet")
	}
}

// **Ohne bereiten lokalen Anbieter sagt der Vorrang es** — und fragt keine
// Cloud.
func TestOhneBereitenAnbieterSagtDerVorrangWasFehlt(t *testing.T) {
	t.Parallel()

	welt := einekiwelt(t)
	ollamaAus(t, welt)
	gestartet := mitGeraet(t, welt, applefm.EinrichtungLizenz, fmName)

	antwort := vorrang(t, welt, "name")
	if antwort.Fehler == nil || !strings.Contains(antwort.Fehler.Meldung, "kein lokaler Anbieter") {
		t.Fatalf("%+v", antwort)
	}

	if gestartet.Load() != 0 {
		t.Errorf("ein nicht eingerichtetes Gerät wurde gestartet")
	}
}

// **Mit dem Vorrang wählt das Programm das Modell** — ein genanntes ist ein
// Fehler, kein stilles Weglassen.
func TestDerVorrangNimmtKeinGenanntesModell(t *testing.T) {
	t.Parallel()

	welt := einekiwelt(t)

	antwort := welt.anwendung.KIVorschlagenMit(probekontext(t), app.KIVorschlagsbezug{
		KarteID: welt.karte, Anbieter: domain.KIVorrang, Alias: "gemma4:8b", Auftrag: "x",
	})
	if antwort.Fehler == nil {
		t.Fatalf("%+v", antwort)
	}
}

// **Eine Ablehnung bei ausdrücklicher Wahl ist ein Ausgang** — kein Fehler,
// keine Felder, und die Rückfälle stehen da (AK-06).
func TestEineAblehnungIstEinAusgangMitRueckfaellen(t *testing.T) {
	t.Parallel()

	welt := einekiwelt(t)
	mitGeraet(t, welt, applefm.EinrichtungBereit, fmAblehnung)

	antwort := welt.anwendung.KIVorschlagenMit(probekontext(t), app.KIVorschlagsbezug{
		KarteID: welt.karte, Anbieter: "apple", Alias: applefm.Modell, Auftrag: "x", Nur: []string{"name"},
	})
	if antwort.Fehler != nil || antwort.Ausgang != app.KIAusgangAbgelehnt || len(antwort.Felder) != 0 ||
		antwort.Weg != string(domain.KIWegGewaehlt) || !slices.Equal(antwort.Rueckfaelle, []string{"ollama"}) {
		t.Fatalf("%+v", antwort)
	}
}

// **Ohne Rückfall bleibt die Ablehnung stehen** — als Ausgang des Vorrangs.
func TestEinAbgelehnterVorrangOhneRueckfall(t *testing.T) {
	t.Parallel()

	welt := einekiwelt(t)
	ollamaAus(t, welt)
	mitGeraet(t, welt, applefm.EinrichtungBereit, fmAblehnung)

	antwort := vorrang(t, welt, "name")
	if antwort.Fehler != nil || antwort.Ausgang != app.KIAusgangAbgelehnt || antwort.Weg != string(domain.KIWegVorrang) {
		t.Fatalf("%+v", antwort)
	}
}

// **Ein Anbieter ohne Modell, das strukturiert antwortet, wird übersprungen**
// — und ist er der letzte, sagt der Fehler, woran es lag.
func TestEinAnbieterOhneTauglichesModell(t *testing.T) {
	t.Parallel()

	welt := einekiwelt(t)
	welt.laufzeit.modelle = []domain.KIModell{einModell("nurtext")}

	antwort := vorrang(t, welt)
	if antwort.Fehler == nil || !strings.Contains(antwort.Fehler.Meldung, "strukturiert") {
		t.Fatalf("%+v", antwort)
	}
}

// **Ein Katalog, der sich nicht bauen lässt, ist ein Fehler** — auch im
// Vorrang.
func TestEinUnbaubarerKatalogImVorrang(t *testing.T) {
	t.Parallel()

	welt := einekiwelt(t)
	welt.laufzeit.modelle = []domain.KIModell{einKoenner("gemma4:8b"), einKoenner("gemma4:8b")}

	if antwort := vorrang(t, welt); antwort.Fehler == nil {
		t.Fatalf("%+v", antwort)
	}
}

// **Nach einem Feld, das es nicht gibt, wird nicht gefragt.**
func TestNurEinUnbekanntesFeldIstEinFehler(t *testing.T) {
	t.Parallel()

	welt := einekiwelt(t)

	antwort := welt.anwendung.KIVorschlagenMit(probekontext(t), app.KIVorschlagsbezug{
		KarteID: welt.karte, Anbieter: "ollama", Alias: "gemma4:8b", Auftrag: "x",
		Nur: []string{"gibt es nicht"},
	})
	if antwort.Fehler == nil || len(welt.laufzeit.auftraege) != 0 {
		t.Fatalf("%+v, gefragt %d", antwort, len(welt.laufzeit.auftraege))
	}
}

// **Wonach ausdrücklich gefragt wird, ist Pflicht im Schema** — gemessen am
// echten Gerätemodell (Vermerk 0166): die Kartenanweisung erlaubt, Felder
// wegzulassen, und ohne `required` kam am Kartenformular mehrmals `{}`.
func TestDasGefragteFeldIstPflicht(t *testing.T) {
	t.Parallel()

	welt := einekiwelt(t)

	antwort := welt.anwendung.KIVorschlagenMit(probekontext(t), app.KIVorschlagsbezug{
		KarteID: welt.karte, Anbieter: "ollama", Alias: "gemma4:8b", Auftrag: "x", Nur: []string{"name"},
	})
	if antwort.Fehler != nil || len(welt.laufzeit.schemen) != 1 {
		t.Fatalf("%+v", antwort)
	}

	pflicht, _ := welt.laufzeit.schemen[0]["required"].([]string) //nolint:errcheck // fehlt es, ist die Liste leer
	if !slices.Equal(pflicht, []string{"name"}) {
		t.Errorf("required: %v", welt.laufzeit.schemen[0]["required"])
	}

	// **Ohne `nur` bleibt alles frei** — die ganze Karte darf Lücken haben.
	welt.anwendung.KIVorschlagenMit(probekontext(t), app.KIVorschlagsbezug{
		KarteID: welt.karte, Anbieter: "ollama", Alias: "gemma4:8b", Auftrag: "x",
	})
	if _, gibt := welt.laufzeit.schemen[1]["required"]; gibt {
		t.Errorf("ohne nur: %v", welt.laufzeit.schemen[1])
	}
}

// **Ein Vorschlag mit Vorrang lässt sich übernehmen** — derselbe zweite
// Schritt wie bei jeder Wahl, und er schreibt den Wert (AK-06).
func TestEinVorrangvorschlagLaesstSichUebernehmen(t *testing.T) {
	t.Parallel()

	welt := einekiwelt(t)
	mitGeraet(t, welt, applefm.EinrichtungBereit, fmName)

	vorschlag := vorrang(t, welt, "name")
	if vorschlag.Fehler != nil || vorschlag.Kennung == "" {
		t.Fatalf("%+v", vorschlag)
	}

	uebernahme := welt.anwendung.KIVorschlagUebernehmen(probekontext(t), vorschlag.Kennung, []string{"name"})
	if uebernahme.Fehler != nil || uebernahme.Felder != 1 {
		t.Fatalf("%+v", uebernahme)
	}

	if !strings.Contains(welt.kartenwerte(t), "Funke") {
		t.Errorf("die Karte: %s", welt.kartenwerte(t))
	}
}
