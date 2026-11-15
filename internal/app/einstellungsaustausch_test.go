package app_test

// **Der Austausch der Einstellungen am Fenster** — Werkstück O3 von Phase 5.2
// (#1427).
//
// # Warum hier eine **angelaufene** Anwendung steht und keine Probe
//
// **Weil die Naht der Gegenstand ist.** Ein Doppel des Austauschdienstes
// prüfte, dass die Fassade weiterreicht — dass sie an einem echten Bestand
// überhaupt **verdrahtet** ist, prüft nur der Weg durch die Wurzel. Genau
// diese Lücke hat in Phase P ein ganzes Paket ohne Aufrufer stehen lassen
// (§10, erste Dimension).

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// austauschMitSprache ist eine gültige Datei mit einer Zeile.
const austauschMitSprache = `{"fassung":1,"bereiche":{"profil":{"sprache":"en"}}}`

// **Die Ausfuhr trägt die Fassung und keinen Pfad.**
func TestDieAusfuhrAmFenster(t *testing.T) {
	anwendung := angelaufeneAnwendung(t)

	antwort := anwendung.EinstellungenAusfuehren(t.Context(), nil)
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if !strings.Contains(antwort.Inhalt, `"fassung"`) {
		t.Errorf("die Fassung fehlt: %s", antwort.Inhalt)
	}

	if strings.Contains(antwort.Inhalt, "hochladepfad") {
		t.Errorf("ein Pfad reiste mit: %s", antwort.Inhalt)
	}

	// **Ein Abschnitt, den es nicht gibt, wird abgewiesen.**
	if abgewiesen := anwendung.EinstellungenAusfuehren(
		t.Context(), []string{"woanders"}); abgewiesen.Fehler == nil {
		t.Error("der erfundene Abschnitt kam durch")
	}
}

// **Geschrieben wird an den Ort, den jemand gewählt hat.**
func TestDieAusfuhrWirdGeschrieben(t *testing.T) {
	anwendung := angelaufeneAnwendung(t)
	ziel := filepath.Join(t.TempDir(), "einstellungen.json")

	anwendung.MitDateiwahl(func(_ context.Context, auswahl app.Dateiauswahl) (string, error) {
		if !auswahl.Speichern {
			t.Error("der Dialog fragt nicht nach einem Ziel")
		}

		return ziel, nil
	})

	if antwort := app.NeueBindungen(anwendung).EinstellungenSichern(t.Context(),
		nil); antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	roh, fehler := os.ReadFile(ziel) //#nosec G304 -- ein Pfad aus diesem Test
	if fehler != nil {
		t.Fatalf("die Datei lesen: %v", fehler)
	}

	if !strings.Contains(string(roh), `"fassung"`) {
		t.Errorf("die Datei trägt keine Fassung: %s", roh)
	}
}

// **Ein abgebrochener Dialog ist kein Fehler und schreibt nichts.**
func TestEinAbgebrochenerAustauschdialogSchreibtNichts(t *testing.T) {
	anwendung := angelaufeneAnwendung(t)

	anwendung.MitDateiwahl(func(context.Context, app.Dateiauswahl) (string, error) {
		return "", nil
	})

	if antwort := anwendung.EinstellungenSichern(
		t.Context(), nil); antwort.Fehler != nil {
		t.Errorf("%+v", antwort.Fehler)
	}
}

// **Und ein gescheiterter Dialog ist einer.**
func TestEinGescheiterterDialogWirdGemeldet(t *testing.T) {
	anwendung := angelaufeneAnwendung(t)

	anwendung.MitDateiwahl(func(context.Context, app.Dateiauswahl) (string, error) {
		return "", os.ErrPermission
	})

	if antwort := anwendung.EinstellungenSichern(
		t.Context(), nil); antwort.Fehler == nil {
		t.Error("der Fehler wurde verschluckt")
	}

	if antwort := anwendung.EinstellungenEinlesen(
		t.Context()); antwort.Fehler == nil {
		t.Error("der Fehler wurde beim Einlesen verschluckt")
	}
}

// **Ohne Ziel wird nicht geschrieben.**
func TestOhneZielWirdNichtGeschrieben(t *testing.T) {
	anwendung := angelaufeneAnwendung(t)

	if antwort := anwendung.EinstellungenSchreiben(
		t.Context(), nil, ""); antwort.Fehler == nil {
		t.Error("es kam durch")
	}

	// **Und ein Ort, an den sich nicht schreiben lässt, ist ein Fehler.**
	if antwort := anwendung.EinstellungenSchreiben(t.Context(), nil,
		filepath.Join(t.TempDir(), "gibtesnicht", "datei.json"),
	); antwort.Fehler == nil {
		t.Error("das unmögliche Ziel kam durch")
	}

	// **Ein abgewiesener Abschnitt kommt auch beim Schreiben nicht durch.**
	if antwort := anwendung.EinstellungenSchreiben(t.Context(),
		[]string{"woanders"}, filepath.Join(t.TempDir(), "datei.json"),
	); antwort.Fehler == nil {
		t.Error("der erfundene Abschnitt kam durch")
	}
}

// **Die Vorschau zeigt die Zeile samt Herkunft und ändert nichts.**
func TestDieVorschauAmFenster(t *testing.T) {
	anwendung := angelaufeneAnwendung(t)

	antwort := anwendung.EinstellungenVorschau(t.Context(), austauschMitSprache)
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if len(antwort.Sicht.Zeilen) != 1 {
		t.Fatalf("%+v", antwort.Sicht)
	}

	if antwort.Sicht.Zeilen[0].Herkunft == "" || antwort.Sicht.Revision == "" {
		t.Errorf("%+v", antwort.Sicht)
	}

	// **Und wann sie wirkt** — OPS-01.AK-02 (#1718); eine Abbildung, die das
	// Feld fallen ließe, gäbe dem Fenster eine leere Wirkung.
	if antwort.Sicht.Zeilen[0].Wirkung != domain.WirkungSofort {
		t.Errorf("Wirkung %q", antwort.Sicht.Zeilen[0].Wirkung)
	}

	// **Der Text kommt zurück** — er ist der, über den entschieden wird.
	if antwort.Sicht.Inhalt != austauschMitSprache {
		t.Errorf("der Text kam nicht zurück: %q", antwort.Sicht.Inhalt)
	}

	// **Ein Befund kommt mit** (§8) — was die Datei mitbringt und hier nicht
	// hingehört, wird genannt und nicht verschluckt.
	mitBefund := anwendung.EinstellungenVorschau(t.Context(),
		`{"fassung":1,"bereiche":{"profil":{"vorgabe.hochladepfad":"/tmp"}}}`)
	if mitBefund.Fehler != nil {
		t.Fatalf("%+v", mitBefund.Fehler)
	}

	if len(mitBefund.Sicht.Befunde) != 1 ||
		mitBefund.Sicht.Befunde[0].Grund == "" {
		t.Errorf("%+v", mitBefund.Sicht.Befunde)
	}

	// **Und eine unlesbare Datei ist ein Fehler mit leeren Listen** (§3).
	kaputt := anwendung.EinstellungenVorschau(t.Context(), "{")
	if kaputt.Fehler == nil {
		t.Fatal("sie kam durch")
	}

	if kaputt.Sicht.Zeilen == nil || kaputt.Sicht.Befunde == nil {
		t.Errorf("eine Liste kam als null: %+v", kaputt.Sicht)
	}
}

// **Eingelesen wird die Datei, die jemand gewählt hat.**
func TestDasEinlesenAmFenster(t *testing.T) {
	anwendung := angelaufeneAnwendung(t)
	quelle := filepath.Join(t.TempDir(), "austausch.json")

	if err := os.WriteFile(quelle, []byte(austauschMitSprache), 0o600); err != nil {
		t.Fatalf("die Datei schreiben: %v", err)
	}

	anwendung.MitDateiwahl(func(_ context.Context, auswahl app.Dateiauswahl) (string, error) {
		if auswahl.Speichern {
			t.Error("der Dialog fragt nach einem Ziel statt nach einer Quelle")
		}

		return quelle, nil
	})

	antwort := app.NeueBindungen(anwendung).EinstellungenEinlesen(t.Context())
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if len(antwort.Sicht.Zeilen) != 1 {
		t.Errorf("%+v", antwort.Sicht)
	}
}

// **Ein abgebrochener Dialog gibt eine leere Sicht ohne Fehler.**
func TestEinAbgebrochenesEinlesenIstKeinFehler(t *testing.T) {
	anwendung := angelaufeneAnwendung(t)

	anwendung.MitDateiwahl(func(context.Context, app.Dateiauswahl) (string, error) {
		return "", nil
	})

	antwort := anwendung.EinstellungenEinlesen(t.Context())
	if antwort.Fehler != nil || len(antwort.Sicht.Zeilen) != 0 {
		t.Errorf("%+v", antwort)
	}
}

// **Eine Datei, die es nicht gibt, wird gemeldet.**
func TestEineFehlendeDateiWirdGemeldet(t *testing.T) {
	anwendung := angelaufeneAnwendung(t)

	anwendung.MitDateiwahl(func(context.Context, app.Dateiauswahl) (string, error) {
		return filepath.Join(t.TempDir(), "gibtesnicht.json"), nil
	})

	if antwort := anwendung.EinstellungenEinlesen(t.Context()); antwort.Fehler == nil {
		t.Error("sie kam durch")
	}
}

// **Die Übernahme schreibt — und eine überholte Revision nicht.**
func TestDieUebernahmeAmFenster(t *testing.T) {
	anwendung := angelaufeneAnwendung(t)
	bindungen := app.NeueBindungen(anwendung)

	vorschau := anwendung.EinstellungenVorschau(t.Context(), austauschMitSprache)
	if vorschau.Fehler != nil {
		t.Fatalf("%+v", vorschau.Fehler)
	}

	// **Zuerst die überholte Revision** — sie darf nichts hinterlassen.
	ueberholt := bindungen.EinstellungenUebernehmen(t.Context(),
		austauschMitSprache, "vongestern", nil)
	if ueberholt.Fehler == nil {
		t.Fatal("die überholte Revision kam durch")
	}

	antwort := bindungen.EinstellungenUebernehmen(t.Context(),
		austauschMitSprache, vorschau.Sicht.Revision, nil)
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if antwort.Bericht.Uebernommen != 1 {
		t.Errorf("%+v", antwort.Bericht)
	}

	// **Und sie steht danach da** — die Antwort ist keine Behauptung über den
	// Bestand.
	gelesen := anwendung.Einstellungen(t.Context())
	if gelesen.Fehler != nil {
		t.Fatalf("%+v", gelesen.Fehler)
	}

	for _, eine := range gelesen.Einstellungen {
		if eine.Schluessel == "sprache" && eine.Wert != "en" {
			t.Errorf("die Sprache steht auf %q", eine.Wert)
		}
	}
}

// **Jede Einstellung nennt ihre Ebene und die Herkunft ihres Werts.**
func TestJedeEinstellungNenntEbeneUndHerkunft(t *testing.T) {
	antwort := angelaufeneAnwendung(t).Einstellungen(t.Context())
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	for _, eine := range antwort.Einstellungen {
		if eine.Ebene == "" {
			t.Errorf("%q nennt keine Ebene", eine.Schluessel)
		}

		if eine.Herkunft == "" {
			t.Errorf("%q nennt keine Herkunft", eine.Schluessel)
		}
	}
}

// **Und derselbe Weg über die Naht ohne Fenster.**
//
// # Warum das hier noch einmal steht
//
// **Weil es eine zweite Naht ist.** Das Fenster geht über `App`, die drei
// Flächen ohne Fenster über [app.Einstellungsweg] — und der baut seine
// Dienste selbst, samt Riegel. Ein Lauf, der nur die eine prüft, sagt über
// die andere nichts; genau diese Lücke hat in Phase P ein Paket ohne Aufrufer
// stehen lassen (§10).
//
// **Dass die drei Flächen darüber wirklich antworten**, misst
// `internal/flaechen/einstellungsaustausch_test.go` an den **gebauten**
// Programmen. Hier steht die Naht selbst.
func TestDerAustauschUeberDieNahtOhneFenster(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	weg := app.EinstellungsnahtUeber(frischerBestand(t))

	inhalt, fehler := weg.Ausfuhr(probekontext(t), nil)
	if fehler != nil {
		t.Fatalf("Ausfuhr: %v", fehler)
	}

	if !strings.Contains(inhalt, `"fassung"`) {
		t.Errorf("die Ausfuhr trägt keine Fassung: %s", inhalt)
	}

	sicht, fehler := weg.Einfuhrvorschau(probekontext(t), []byte(austauschMitSprache))
	if fehler != nil {
		t.Fatalf("Vorschau: %v", fehler)
	}

	if len(sicht.Zeilen) != 1 || sicht.Inhalt != austauschMitSprache {
		t.Fatalf("%+v", sicht)
	}

	bericht, fehler := weg.Einfuhruebernahme(
		probekontext(t), []byte(austauschMitSprache), sicht.Revision, nil)
	if fehler != nil {
		t.Fatalf("Übernahme: %v", fehler)
	}

	if bericht.Uebernommen != 1 {
		t.Errorf("%+v", bericht)
	}
}

// **Und jeder der drei Wege reicht seinen Fehler durch.**
func TestDieAustauschnahtReichtIhreFehlerDurch(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	weg := app.EinstellungsnahtUeber(frischerBestand(t))

	if _, fehler := weg.Ausfuhr(probekontext(t), []string{"woanders"}); fehler == nil {
		t.Error("der erfundene Abschnitt kam durch")
	}

	if _, fehler := weg.Einfuhrvorschau(probekontext(t), []byte("{")); fehler == nil {
		t.Error("die unlesbare Datei kam durch")
	}

	if _, fehler := weg.Einfuhruebernahme(
		probekontext(t), []byte(austauschMitSprache), "vongestern", nil,
	); fehler == nil {
		t.Error("die überholte Revision kam durch")
	}
}
