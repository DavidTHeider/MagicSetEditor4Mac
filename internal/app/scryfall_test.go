package app_test

import (
	"context"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Ein Vorschlag kommt als DTO an, und er schreibt nichts**
// (#1005).
//
// # Was hier geprüft wird und was nicht
//
// **Die Fassade**: Kennungen, Fehlerübersetzung, DTO. Ob ein Vorschlag
// fachlich stimmt, hält `internal/service`; ob eine Scryfall-Antwort richtig
// gelesen wird, `internal/scryfall`.
func TestEinVorschlagKommtAlsDTOAn(t *testing.T) {
	t.Parallel()

	anwendung := &app.App{}
	anwendung.MitBestand(bestandMitSet())
	anwendung.MitEditor(nil, karteMitFeldern(), nil, nil)
	anwendung.MitScryfall(quelleMitVorschlag())

	antwort := anwendung.ScryfallVorschlag(t.Context(), "1", "2", "Prüfstein")
	if antwort.Fehler != nil {
		t.Fatalf("Fehler: %+v", antwort.Fehler)
	}

	if antwort.Vorschlag == nil {
		t.Fatal("kein Vorschlag")
	}

	if len(antwort.Vorschlag.Aenderungen) != 1 {
		t.Fatalf("%d Änderungen", len(antwort.Vorschlag.Aenderungen))
	}

	eine := antwort.Vorschlag.Aenderungen[0]
	if eine.Feld != "rule text" || eine.Alt != "Alt" || eine.Neu != "Neu" {
		t.Errorf("Änderung %+v", eine)
	}

	// **Das Alter kommt in Stunden an** — eine `time.Duration` käme als
	// Nanosekunden, und die Seite müsste rechnen.
	if antwort.Vorschlag.AlterStunden != 3 {
		t.Errorf("AlterStunden %d", antwort.Vorschlag.AlterStunden)
	}

	if antwort.Vorschlag.Herkunft != "https://beispiel.invalid/karte" {
		t.Errorf("Herkunft %q", antwort.Vorschlag.Herkunft)
	}

	// **Die leeren Listen bleiben Listen** — die Falle aus §3.
	if antwort.Vorschlag.Befunde == nil {
		t.Error("die Befunde kamen als null an")
	}
}

// **Eine unbrauchbare Kennung wird abgewiesen, bevor gefragt wird.**
func TestEineUnbrauchbareKennungWirdVorDemFragenAbgewiesen(t *testing.T) {
	t.Parallel()

	faelle := map[string][2]string{
		"unbrauchbares Set":  {"x", "2"},
		"unbrauchbare Karte": {"1", "x"},
	}

	for name, kennungen := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			gefragt := &zaehlendequelle{}

			anwendung := &app.App{}
			anwendung.MitScryfall(gefragt)

			antwort := anwendung.ScryfallVorschlag(t.Context(),
				kennungen[0], kennungen[1], "Prüfstein")
			if antwort.Fehler == nil {
				t.Fatal("die unbrauchbare Kennung kam durch")
			}

			if gefragt.rufe != 0 {
				t.Errorf("es wurde %dmal gefragt", gefragt.rufe)
			}
		})
	}
}

// **Ohne eingerichteten Zugang kommt eine Meldung, kein Absturz.**
//
// Wer kein Zuhause hat, bekommt beim Start keine Scryfall-Fähigkeit — das
// Fenster geht trotzdem auf, und die Meldung kommt beim ersten Versuch.
func TestOhneEingerichtetenZugangKommtEineMeldung(t *testing.T) {
	t.Parallel()

	anwendung := &app.App{}

	if antwort := anwendung.ScryfallVorschlag(t.Context(), "1", "2", "x"); antwort.Fehler == nil {
		t.Error("ohne Zugang kam ein Vorschlag")
	}

	if antwort := anwendung.ScryfallUebernehmen(t.Context(), "1", "2", "x", nil, ""); antwort.Fehler == nil {
		t.Error("ohne Zugang wurde übernommen")
	}
}

// **Eine Karte aus einem anderen Set wird abgewiesen.**
//
// # Warum das eine eigene Wache ist
//
// Sonst schriebe eine Übernahme in eine Karte, die zu einem anderen Set
// gehört — und die Zuordnung käme vom **falschen Spiel**. Das ist derselbe
// Grund, aus dem `FeldSetzen` beide Kennungen nimmt (#808).
func TestEineKarteAusEinemAnderenSetWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	anwendung := &app.App{}
	anwendung.MitBestand(bestandMitSet())
	anwendung.MitEditor(nil, karteAusAnderemSet(), nil, nil)
	anwendung.MitScryfall(quelleMitVorschlag())

	antwort := anwendung.ScryfallVorschlag(t.Context(), "1", "2", "Prüfstein")
	if antwort.Fehler == nil {
		t.Error("die fremde Karte kam durch")
	}
}

// **Übernommen wird über dieselbe Quelle, und der Fehler kommt durch.**
func TestEineUebernahmeReichtIhrenFehlerDurch(t *testing.T) {
	t.Parallel()

	anwendung := &app.App{}
	anwendung.MitBestand(bestandMitSet())
	anwendung.MitEditor(nil, karteMitFeldern(), nil, nil)
	anwendung.MitScryfall(quelleDieBeimSchreibenScheitert())

	antwort := anwendung.ScryfallUebernehmen(t.Context(), "1", "2", "Prüfstein",
		[]string{"rule text"}, "")
	if antwort.Fehler == nil {
		t.Error("der Schreibfehler kam nicht durch")
	}
}

// **Und eine geglückte Übernahme meldet keinen Fehler.**
func TestEineGegluecteUebernahmeMeldetKeinenFehler(t *testing.T) {
	t.Parallel()

	anwendung := &app.App{}
	anwendung.MitBestand(bestandMitSet())
	anwendung.MitEditor(nil, karteMitFeldern(), nil, nil)
	anwendung.MitScryfall(quelleMitVorschlag())

	antwort := anwendung.ScryfallUebernehmen(t.Context(), "1", "2", "Prüfstein",
		[]string{"rule text"}, "")
	if antwort.Fehler != nil {
		t.Errorf("Fehler: %+v", antwort.Fehler)
	}
}

// **Ein Vorschlag, der nicht mehr gilt, wird abgewiesen** —
// #1241.
//
// Dieser Weg liest die Karte neu und rechnet den Vorschlag ein zweites Mal;
// zwischen Zeigen und Übernehmen kann sie sich geändert haben — dann übernähme
// der Knopf **etwas anderes als das Gezeigte**.
func TestEinUeberholterVorschlagWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	anwendung := &app.App{}
	anwendung.MitBestand(bestandMitSet())
	anwendung.MitEditor(nil, karteMitFeldern(), nil, nil)
	anwendung.MitScryfall(quelleMitVorschlag())

	antwort := anwendung.ScryfallUebernehmen(t.Context(), "1", "2", "Prüfstein",
		[]string{"rule text"}, "ein-stand-von-vorhin")
	if antwort.Fehler == nil {
		t.Fatal("der überholte Vorschlag wurde übernommen")
	}

	if antwort.Fehler.Code != app.CodeKonflikt {
		t.Errorf("Code %q, erwartet einen Konflikt", antwort.Fehler.Code)
	}
}

// **Ein Set, das es nicht gibt, wird abgewiesen.**
func TestEinSetDasEsNichtGibtWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	anwendung := &app.App{}
	anwendung.MitBestand(bestandOhneSets())
	anwendung.MitEditor(nil, karteMitFeldern(), nil, nil)
	anwendung.MitScryfall(quelleMitVorschlag())

	if antwort := anwendung.ScryfallVorschlag(t.Context(), "1", "2", "x"); antwort.Fehler == nil {
		t.Error("das fehlende Set kam durch")
	}
}

// **Und ein Lesefehler ebenso.**
func TestEinLesefehlerBeimKartenstandKommtDurch(t *testing.T) {
	t.Parallel()

	for name, bauen := range map[string]func(*app.App){
		"die Karte lässt sich nicht lesen": func(a *app.App) {
			a.MitBestand(bestandMitSet())
			a.MitEditor(nil, karteleserScheitert{}, nil, nil)
		},
		"die Setköpfe lassen sich nicht lesen": func(a *app.App) {
			a.MitBestand(bestandScheitert{})
			a.MitEditor(nil, karteMitFeldern(), nil, nil)
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			anwendung := &app.App{}
			bauen(anwendung)
			anwendung.MitScryfall(quelleMitVorschlag())

			if antwort := anwendung.ScryfallVorschlag(t.Context(), "1", "2", "x"); antwort.Fehler == nil {
				t.Error("der Lesefehler kam nicht durch")
			}
		})
	}
}

// **Und ein Fehler beim Fragen ebenso.**
func TestEinFehlerBeimFragenKommtDurch(t *testing.T) {
	t.Parallel()

	anwendung := &app.App{}
	anwendung.MitBestand(bestandMitSet())
	anwendung.MitEditor(nil, karteMitFeldern(), nil, nil)
	anwendung.MitScryfall(quelleScheitert{})

	if antwort := anwendung.ScryfallVorschlag(t.Context(), "1", "2", "x"); antwort.Fehler == nil {
		t.Error("der Fehler kam nicht durch")
	}

	if antwort := anwendung.ScryfallUebernehmen(t.Context(), "1", "2", "x", nil, ""); antwort.Fehler == nil {
		t.Error("der Fehler kam bei der Übernahme nicht durch")
	}
}

// **Das Verzeichnis liegt neben dem Bestand.**
func TestDasScryfallverzeichnisLiegtNebenDemBestand(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	ort, fehler := app.Scryfallverzeichnis()
	if fehler != nil {
		t.Fatalf("Verzeichnis: %v", fehler)
	}

	if ort == "" {
		t.Error("der Ort ist leer")
	}
}

// **Und ohne Zuhause gibt es keines.**
func TestOhneZuhauseGibtEsKeinScryfallverzeichnis(t *testing.T) {
	t.Setenv("HOME", "")
	t.Setenv("MSE_BESTAND", "")

	if _, fehler := app.Scryfallverzeichnis(); fehler == nil {
		t.Error("ohne Zuhause kam ein Verzeichnis")
	}
}

// **Die Naht kommt zusammen, wenn es ein Zuhause gibt.**
func TestDieScryfallnahtKommtZusammen(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	abgleich, fehler := app.Scryfallnaht(app.ImmerAn{}, nil, nil)
	if fehler != nil {
		t.Fatalf("Naht: %v", fehler)
	}

	if abgleich == nil {
		t.Error("die Naht ist leer")
	}
}

// **Und ohne Zuhause nicht.**
func TestOhneZuhauseKommtDieScryfallnahtNichtZusammen(t *testing.T) {
	t.Setenv("HOME", "")
	t.Setenv("MSE_BESTAND", "")

	if _, fehler := app.Scryfallnaht(app.ImmerAn{}, nil, nil); fehler == nil {
		t.Error("ohne Zuhause kam eine Naht")
	}
}

// **Die Bindungen reichen durch — mit dem Kontext des Fensters.**
//
// # Warum sie einen eigenen Test brauchen
//
// Weil sie die einzige Stelle sind, an der das Fenster den Kontext bekommt.
// Eine Bindung, die `context.Background()` nähme, hörte auf kein Schließen —
// und das fiele erst auf, wenn jemand das Fenster während einer Anfrage
// zumacht.
func TestDieBindungenReichenDurch(t *testing.T) {
	t.Parallel()

	anwendung := &app.App{}
	anwendung.MitBestand(bestandMitSet())
	anwendung.MitEditor(nil, karteMitFeldern(), nil, nil)
	anwendung.MitScryfall(quelleMitVorschlag())

	bindungen := app.NeueBindungen(anwendung)

	antwort := bindungen.ScryfallVorschlag(t.Context(), "1", "2", "Prüfstein")
	if antwort.Fehler != nil || antwort.Vorschlag == nil {
		t.Fatalf("Vorschlag: %+v", antwort.Fehler)
	}

	if schreib := bindungen.ScryfallUebernehmen(t.Context(), "1", "2", "Prüfstein",
		[]string{"rule text"}, ""); schreib.Fehler != nil {
		t.Errorf("Übernehmen: %+v", schreib.Fehler)
	}
}

// **Die Befunde der Antwort kommen als Texte an.**
//
// Ein Befund trägt Datei, Ort und Meldung; das Fenster braucht einen Satz.
func TestDieBefundeKommenAlsTexteAn(t *testing.T) {
	t.Parallel()

	anwendung := &app.App{}
	anwendung.MitBestand(bestandMitSet())
	anwendung.MitEditor(nil, karteMitFeldern(), nil, nil)
	anwendung.MitScryfall(&feste{vorschlag: &service.Vorschlag{
		Befunde: []domain.Befund{{Ort: "layout", Meldung: "unbekannter Wert"}},
	}})

	antwort := anwendung.ScryfallVorschlag(t.Context(), "1", "2", "x")
	if antwort.Fehler != nil {
		t.Fatalf("Fehler: %+v", antwort.Fehler)
	}

	if len(antwort.Vorschlag.Befunde) != 1 {
		t.Fatalf("%d Befunde", len(antwort.Vorschlag.Befunde))
	}

	if antwort.Vorschlag.Befunde[0] == "" {
		t.Error("der Befund kam als leerer Text an")
	}
}

// **Und eine unbrauchbare Kennung bei der Übernahme ebenso.**
func TestEineUnbrauchbareKennungBeiDerUebernahmeWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	anwendung := &app.App{}
	anwendung.MitScryfall(quelleMitVorschlag())

	for _, kennungen := range [][2]string{{"x", "2"}, {"1", "x"}} {
		if antwort := anwendung.ScryfallUebernehmen(t.Context(),
			kennungen[0], kennungen[1], "x", nil, ""); antwort.Fehler == nil {
			t.Errorf("%v kam durch", kennungen)
		}
	}
}

// **Ein Lesefehler bei der Übernahme kommt durch.**
func TestEinLesefehlerBeiDerUebernahmeKommtDurch(t *testing.T) {
	t.Parallel()

	anwendung := &app.App{}
	anwendung.MitBestand(bestandMitSet())
	anwendung.MitEditor(nil, karteleserScheitert{}, nil, nil)
	anwendung.MitScryfall(quelleMitVorschlag())

	if antwort := anwendung.ScryfallUebernehmen(t.Context(), "1", "2", "x", nil, ""); antwort.Fehler == nil {
		t.Error("der Lesefehler kam nicht durch")
	}
}

func (karteleserScheitert) Herkuenfte(context.Context, int64) (map[string]string, error) {
	return nil, nil
}
