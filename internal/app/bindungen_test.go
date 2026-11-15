package app_test

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// Jede gebundene Methode erreicht ihre Entsprechung an der Anwendung.
//
// **Der Zweck ist die Verdrahtung, nicht die Fachlichkeit.** Eine Fassade aus
// elf Durchreichern ist die Stelle, an der ein Abschreibfehler sitzt — die
// `Statistik`, die `Verteilung` ruft, sieht richtig aus und ist es nicht. Was
// die Methoden dahinter tun, prüfen ihre eigenen Tests.
func TestJedeBindungErreichtIhreAnwendung(t *testing.T) {
	stand := bindungsprobe(t)
	anwendung, bindungen := stand.Anwendung, stand.Bindungen
	schreiber := stand.Schreiber
	_ = schreiber

	// **Jeder Durchreicher wird an einem Ergebnis erkannt, nicht am Ausbleiben
	// eines Fehlers.**
	//
	// Der erste Anlauf prüfte bei der Hälfte nur `Fehler == nil`. Ein
	// Durchreicher, der sofort einen leeren Umschlag zurückgibt, wäre damit grün
	// geblieben, ohne die Anwendung je zu erreichen — der Nachlauf des
	// Zweitblicks hat es angestrichen.
	t.Run("MTGJSONOeffnen", func(t *testing.T) {
		probe := &importprobe{ergebnis: &service.Ergebnis{SetID: 6, Karten: 1}}
		anwendung.MitImporteur(probe)

		var gewaehlt app.Dateiauswahl

		anwendung.MitDateiwahl(func(_ context.Context, auswahl app.Dateiauswahl) (string, error) {
			gewaehlt = auswahl

			return "/tmp/PRB.json", nil
		})

		antwort := bindungen.MTGJSONOeffnen(t.Context(), "magic")
		if antwort.Fehler != nil {
			t.Fatalf("%+v", antwort.Fehler)
		}

		if antwort.SetID != "6" || probe.spiel != "magic" {
			t.Errorf("%+v, Spiel %q", antwort, probe.spiel)
		}

		// **Der Dateidialog, nicht der Ordnerdialog** — und mit einem Muster,
		// das JSON durchlässt. Ohne diese Prüfung wäre ein vertauschter Dialog
		// grün: die Vorrichtung gibt ihren Pfad ohnehin zurück.
		if gewaehlt.Verzeichnis || gewaehlt.Muster != "*.json" {
			t.Errorf("gefragt wurde nach %+v", gewaehlt)
		}
	})

	t.Run("MTGJSONVerzeichnisOeffnen", func(t *testing.T) {
		probe := &importprobe{ergebnis: &service.Ergebnis{SetID: 8, Karten: 3}}
		anwendung.MitImporteur(probe)

		var gewaehlt app.Dateiauswahl

		anwendung.MitDateiwahl(func(_ context.Context, auswahl app.Dateiauswahl) (string, error) {
			gewaehlt = auswahl

			return "/tmp/sets", nil
		})

		antwort := bindungen.MTGJSONVerzeichnisOeffnen(t.Context(), "magic")
		if antwort.Fehler != nil {
			t.Fatalf("%+v", antwort.Fehler)
		}

		if antwort.SetID != "8" || probe.pfad != "/tmp/sets" {
			t.Errorf("%+v, Pfad %q", antwort, probe.pfad)
		}

		if !gewaehlt.Verzeichnis {
			t.Errorf("gefragt wurde nach %+v — erwartet war der Ordnerdialog", gewaehlt)
		}
	})

	t.Run("MTGJSONSchreiben", func(t *testing.T) {
		probe := &ausfuhrprobe{ausfuhr: &service.Ausfuhr{Pfad: "/tmp/PRB.json", Karten: 4}}
		anwendung.MitExporteur(probe)

		var gewaehlt app.Dateiauswahl

		anwendung.MitDateiwahl(func(_ context.Context, auswahl app.Dateiauswahl) (string, error) {
			gewaehlt = auswahl

			return "/tmp/PRB.json", nil
		})

		antwort := bindungen.MTGJSONSchreiben(t.Context(), "7", "magic")
		if antwort.Fehler != nil {
			t.Fatalf("%+v", antwort.Fehler)
		}

		if antwort.Karten != 4 || probe.setID != 7 || probe.spiel != "magic" {
			t.Errorf("%+v, Set %d, Spiel %q", antwort, probe.setID, probe.spiel)
		}

		// **Der Sicherndialog, nicht der Öffnendialog** — und mit einem
		// Vorschlag. Wer eine Datei anlegt, sucht keinen vorhandenen Namen aus.
		if !gewaehlt.Speichern || gewaehlt.Vorschlag == "" {
			t.Errorf("gefragt wurde nach %+v", gewaehlt)
		}
	})

	t.Run("ein abgebrochener Sicherndialog", func(t *testing.T) {
		anwendung.MitExporteur(&ausfuhrprobe{})
		anwendung.MitDateiwahl(func(context.Context, app.Dateiauswahl) (string, error) {
			return "", errors.New("der Dialog ging nicht auf")
		})

		if antwort := bindungen.MTGJSONSchreiben(t.Context(), "7", "magic"); antwort.Fehler == nil {
			t.Error("der Fehler ging verloren")
		}
	})

	t.Run("MTGJSONSpiele", func(t *testing.T) {
		if len(bindungen.MTGJSONSpiele(t.Context())) == 0 {
			t.Error("keine einzige Zuordnung")
		}
	})
}

// **Mit Fenster kommt der Kontext des Fensters.**
//
// Ohne Fenster ein leerer — der Zustand vor dem ersten `OnStartup`, in dem es
// noch keine Oberfläche gibt, die rufen könnte. Beide Wege gehören geprüft:
// ein Aufruf, der am `nil` scheitert, ist der Fall, den niemand sieht, bis die
// Anwendung ihn zeigt.

// bindungsprobe baut die Anwendung, gegen die beide Vertragstests laufen.
//
// **Eine Vorrichtung für zwei Dateien.** Der Test war eine Funktion mit
// dreißig Unterläufen und ging über 400 Zeilen (§2); geteilt ist er nur, wenn
// beide Hälften **dieselbe** Anwendung bekommen — zwei Aufbauten wären zwei
// Zusagen darüber, was verdrahtet ist.
// Bindungsstand ist, was ein Vertragstest zum Prüfen braucht.
//
// **Ein Verbund und keine sechs Rückgabewerte.** Sechs sind eine Reihenfolge,
// die man vertauschen kann, und zwei davon sind vom selben Typ.
type Bindungsstand struct {
	Suche     *suchprobe
	Statistik *statistikprobe
	Massen    *massenprobe
	Schreiber *schreibprobe
	Anwendung *app.App
	Bindungen *app.Bindungen
}

func bindungsprobe(t *testing.T) Bindungsstand {
	t.Helper()

	probe := &suchprobe{
		treffer: []domain.SearchHit{{CardID: 1, Field: "name", Value: "Feuerball"}},
		werte:   map[int64][]domain.Value{1: {{Field: "name", Text: "Feuerball"}}},
		felder:  []string{"name"},

		// **Das Set muss es geben** (#1336): eine Setkennung, die auf nichts
		// zeigt, wird seit dieser Änderung abgewiesen.
		sets: []domain.SetHead{{ID: 1}},
	}
	statistik := &statistikprobe{
		zahlen:     domain.SetStats{Cards: 12, Fields: 3},
		belegung:   []domain.FieldCount{{Field: "name", Cards: 12}},
		verteilung: []domain.ValueCount{{Value: "Zauber", Cards: 5}},
	}
	massen := &massenprobe{}
	schreiber := &schreibprobe{}

	anwendung := app.NeueApp("")
	anwendung.MitSucher(probe)

	// **Auch der Datenstand** (#156): ohne ihn steht die Naht auf `nil`, und
	// die Bindung dahinter greift ins Leere. Genau daran ist der erste Anlauf
	// gescheitert — eine Vorrichtung, die eine Naht auslässt, prüft sie nicht,
	// sie stürzt an ihr ab.
	anwendung.MitStand(standprobe{})
	anwendung.MitStatistiker(statistik)
	anwendung.MitMassenschreiber(massen)
	anwendung.MitBestand(leseprobe{
		koepfe:  []domain.SetHead{{ID: 1, Name: "Aus dem Bestand", Game: "probe"}},
		karten:  []*domain.CardRow{{ID: 7, Number: 0}},
		bloecke: []domain.RawBlock{{File: "set", Line: 3, Text: "roh"}},
	})
	// **Und die drei Wege zur Kartenliste** (#891). Sie sind eigene Nähte,
	// nicht Teil von [app.Bestandsleser] — ohne sie liefe die Bindung in ihre
	// Meldung statt in den Vertrag.
	liste := listenprobe{
		zeilen: []domain.CardLine{{ID: 7, Number: 0, Name: "Aus dem Bestand", Fields: 1}},
		karte:  &domain.CardRow{ID: 7, Number: 0},
		werte:  []domain.CardField{{ID: 7, Text: "Zauber"}},
	}
	// **Und die Marken** (#1208) — ohne die Naht griffe die Bindung ins Leere.
	anwendung.MitMarken(&markenpflegeprobe{})

	anwendung.MitKartenzeilen(liste)
	anwendung.MitEinzelkarte(liste)
	anwendung.MitKartenfeldwerten(liste)

	// **Und die Stilwahl** — sie trägt seit #560 auch die Feldbeschreibungen.
	anwendung.MitStilwahl(&stilwaehler{felder: []domain.Field{
		{Name: "name", Kind: domain.KindText},
	}})

	// **Und der Kartenanleger** (#562).
	anwendung.MitKartenanleger(&anlegeprobe{kennung: 9})

	// **Und der Kartenlöscher** (#621). Ohne ihn liefe die Bindung in einen
	// leeren Zeiger, statt den Vertrag zu prüfen.
	anwendung.MitKartenloescher(&loeschprobe{})
	anwendung.MitKartenkopierer(&kopierprobe{kennung: 11})
	anwendung.MitKartentauscher(&tauschprobe{})

	anwendung.MitEditor(
		schreiber,
		// **Mit einem Feldwert**, damit der Behelfsplan eine Fläche hat: an ihr
		// hängt `BildSetzen` — es schneidet auf die Fläche des Feldes zu.
		karteprobe{karte: &domain.CardRow{
			ID: 1, Values: []domain.Value{{Field: "kunst", Text: ""}},
		}},
		service.NeuerPlanbauer(),
		render.NeuerRasterer(ohneBilder{}),
	)

	return Bindungsstand{
		Suche: probe, Statistik: statistik, Massen: massen, Schreiber: schreiber,
		Anwendung: anwendung, Bindungen: app.NeueBindungen(anwendung),
	}
}
