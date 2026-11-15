//go:build darwin

package app_test

// **Was die Metriken seit OPS-05 dazu sagen** — OPS-05.AK-01/04
// (#1722):
// Laufzeitverteilung, Fehlerrate, Datenbank, Zwischenspeicher und Prozess,
// jede Zahl mit ihrer Einheit im Namen — und der Bericht nennt Zeitraum,
// Vollständigkeit und Messaufwand.
//
// **Ohne `t.Parallel()`**, aus demselben Grund wie in `metriken_test.go`.

import (
	"slices"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

func TestDieMetrikenNennenLaufzeitenUndProzess(t *testing.T) {
	service.LaufzeitwerkZuruecksetzen()
	t.Cleanup(service.LaufzeitwerkZuruecksetzen)

	// **Über die Anwendung**, wie die Wurzeln von Dienst und Werkzeugserver.
	app.ErfasseRuf(t.Context(), app.Rufmessung{
		Flaeche: domain.ModusMCP, Weg: "lesen.sets", Dauer: 3 * time.Millisecond,
		Ausgang: domain.RufGescheitert,
	})

	antwort := app.NeueApp("").Metriken(t.Context())
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if len(antwort.Laufzeiten) != 1 || antwort.Laufzeiten[0].Flaeche != "mcp" ||
		antwort.Laufzeiten[0].Gescheitert != 1 || antwort.Laufzeiten[0].Eimer[1] != 1 {
		t.Errorf("Laufzeiten: %+v", antwort.Laufzeiten)
	}

	if !slices.Equal(antwort.EimerMs, service.Laufzeiteimer) {
		t.Errorf("Eimer: %v", antwort.EimerMs)
	}

	prozess := antwort.Prozess
	if prozess.HeapBytes <= 0 || prozess.Goroutinen <= 0 || prozess.SystemBytes < prozess.HeapBytes {
		t.Errorf("Prozess: %+v", prozess)
	}

	// **Zeitraum und Aufwand** (AK-04): von wann bis wann, und was das
	// Erheben selbst gekostet hat.
	seit, fehler := time.Parse(time.RFC3339Nano, antwort.Seit)
	bis, zweiter := time.Parse(time.RFC3339Nano, antwort.Bis)
	if fehler != nil || zweiter != nil || bis.Before(seit) || antwort.AufwandMikros < 0 {
		t.Errorf("Zeitraum %q bis %q, Aufwand %d", antwort.Seit, antwort.Bis, antwort.AufwandMikros)
	}

	// **Vollständigkeit**: was dieser Prozess nicht misst, steht in der
	// Antwort und nicht nur in der Doku.
	if !slices.Contains(antwort.Ungemessen, app.UngemessenFenster) {
		t.Errorf("ungemessen: %v", antwort.Ungemessen)
	}

	// **Ohne Bestand keine Datenbankzahlen** — und keine erfundenen Nullen.
	if antwort.Datenbank != nil || antwort.Spur.An || antwort.Spur.Punkte == nil {
		t.Errorf("Datenbank %+v, Spur %+v", antwort.Datenbank, antwort.Spur)
	}
}

func TestMitBestandNennenSieDieVerbindungen(t *testing.T) {
	anwendung, _ := gestarteteAnwendung(t)

	antwort := anwendung.Metriken(t.Context())
	if antwort.Datenbank == nil || antwort.Datenbank.Offen < 1 {
		t.Fatalf("Datenbank: %+v", antwort.Datenbank)
	}
}

// **Die Spur richtet sich nach den Einstellungen des Geräts** — gelesen mit
// dem gezeigten Profil, denn ohne es liest der Bestand keine.
func TestDieMessungRichtetSichNachDenEinstellungen(t *testing.T) {
	service.LaufzeitwerkZuruecksetzen()
	t.Cleanup(service.LaufzeitwerkZuruecksetzen)

	anwendung, bestand := gestarteteAnwendung(t)
	for schluessel, wert := range map[string]string{
		domain.MessungSpur: "an", domain.MessungSpurgrenze: "20",
	} {
		if antwort := anwendung.EinstellungSetzen(t.Context(), schluessel, wert); antwort.Fehler != nil {
			t.Fatalf("%s: %+v", schluessel, antwort.Fehler)
		}
	}

	// **Über eine eigene Verbindung**, wie Dienst und Werkzeugserver sie haben.
	db, fehler := storage.Oeffnen(t.Context(), bestand)
	if fehler != nil {
		t.Fatal(fehler)
	}

	t.Cleanup(func() {
		if schliessfehler := db.Close(); schliessfehler != nil {
			t.Error(schliessfehler)
		}
	})

	if fehler := app.MessungEinrichten(t.Context(), db); fehler != nil {
		t.Fatal(fehler)
	}

	if stand := service.Spurstand(); !stand.An || stand.Grenze != 20 {
		t.Errorf("%+v", stand)
	}

	// **Und Dienst und Werkzeugserver nennen ihre Verbindung** — dieselbe
	// Antwort, über die Verbindung, die sie halten.
	if antwort := app.MetrikenUeber(t.Context(), db); antwort.Datenbank == nil || !antwort.Spur.An {
		t.Errorf("über die Verbindung: %+v, %+v", antwort.Datenbank, antwort.Spur)
	}

	// **Ohne lesbaren Bestand bleibt die Spur, wie sie beim Start ist: aus** —
	// und der Fehler kommt zurück, damit die Wurzel entscheidet.
	service.LaufzeitwerkZuruecksetzen()

	if fehler := db.Close(); fehler != nil {
		t.Fatal(fehler)
	}

	if fehler := app.MessungEinrichten(t.Context(), db); fehler == nil || service.Spurstand().An {
		t.Errorf("ein geschlossener Bestand: %v, %+v", fehler, service.Spurstand())
	}
}
