// **Ein Syntaxfehler wird an seiner Stelle angestrichen** — Werkstück 1 von
// Phase 3.3 (#1164).
//
// # Der Stand
//
// Der Zerteiler kennt Zeile und Spalte seit je (`internal/msescript`), und die
// Konsole warf beides weg: `service.Skriptkonsole.Auswerten` machte aus dem
// Fehler eine Zeichenkette, und das Fenster bekam einen Satz ohne Ort.
//
// **Eine Liste daneben ist keine Markierung.** Wer einen Ausdruck über drei
// Zeilen schreibt, sucht die Stelle sonst von Hand — und die Angabe stand die
// ganze Zeit da.
package app_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// stellenkonsole gibt einen Fehler mit Stelle zurück.
type stellenkonsole struct {
	fehler error
}

func (s stellenkonsole) Auswerten(
	context.Context, int64, string,
) (app.KonsolenDTO, error) {
	return app.KonsolenDTO{}, s.fehler
}

func (s stellenkonsole) Namen(context.Context, int64) ([]string, error) {
	return nil, nil
}

// **Die Antwort trägt Zeile und Spalte.**
func TestEinSkriptfehlerTraegtSeineStelle(t *testing.T) {
	anwendung := app.NeueApp("")
	anwendung.MitSkriptkonsole(stellenkonsole{fehler: fehlerMitStelle(2, 3)})

	antwort := anwendung.Skript(t.Context(), "1", "1 +\n  * 2")
	if antwort.Fehler == nil {
		t.Fatal("der Syntaxfehler kam als Erfolg durch")
	}

	if antwort.Stelle == nil {
		t.Fatal("die Antwort trägt keine Stelle")
	}

	if antwort.Stelle.Zeile != 2 || antwort.Stelle.Spalte != 3 {
		t.Errorf("die Stelle ist Zeile %d, Spalte %d — erwartet war 2/3",
			antwort.Stelle.Zeile, antwort.Stelle.Spalte)
	}
}

// **Ein Fehler ohne Stelle trägt keine.**
//
// Die Gegenprobe: ohne sie sagte der Lauf darüber auch dann ja, wenn **jeder**
// Fehler eine erfundene Stelle bekäme. Eine Karte, die es nicht gibt, hat
// keine Zeile.
func TestEinFehlerOhneStelleTraegtKeine(t *testing.T) {
	anwendung := app.NeueApp("")
	anwendung.MitSkriptkonsole(stellenkonsole{
		fehler: errors.New("die Karte gibt es nicht"),
	})

	antwort := anwendung.Skript(t.Context(), "1", "1 + 1")
	if antwort.Fehler == nil {
		t.Fatal("der Fehler kam als Erfolg durch")
	}

	if antwort.Stelle != nil {
		t.Errorf("es entstand die Stelle %+v", antwort.Stelle)
	}
}

// fehlerMitStelle baut einen Skriptfehler, wie der Zerteiler ihn gibt.
//
// **Über `msescript` und nicht nachgebaut**: eine eigene Fehlerart hier
// prüfte, dass `errors.As` auf sie greift — und nicht, dass es auf **die**
// greift, die der Zerteiler wirklich wirft (§3).
func fehlerMitStelle(zeile, spalte int) error {
	_, fehler := msescript.AuswertenQuelle(context.Background(), "1 +\n  * 2",
		msescript.Bibliothek())

	var stelle *msescript.Fehler
	if !errors.As(fehler, &stelle) || stelle.Zeile != zeile || stelle.Spalte != spalte {
		panic("der Zerteiler meldet nicht mehr Zeile 2, Spalte 3: " + fehler.Error())
	}

	return fmt.Errorf("%w: %w", domain.ErrValidation, fehler)
}

// **Die Namen kommen aus der wirklichen Umgebung.**
//
// Werkstück 1: „Vervollständigung aus der Bibliothek — **soweit belastbar**:
// was die Umgebung wirklich kennt, nicht was plausibel klingt."
func TestDieSkriptnamenKommenAusDerUmgebung(t *testing.T) {
	anwendung := app.NeueApp("")
	anwendung.MitSkriptkonsole(namenskonsole{namen: []string{"card", "length", "set"}})

	antwort := anwendung.Skriptnamen(t.Context(), "1")
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if len(antwort.Namen) != 3 {
		t.Errorf("es kamen %d Namen: %v", len(antwort.Namen), antwort.Namen)
	}
}

// **Ohne Konsole gibt es eine leere Liste und einen Fehler — kein `nil`.**
//
// Über die Wails-Bindung stünde dort sonst `null`, und das Fenster prüfte auf
// zwei Dinge statt auf eines.
func TestOhneKonsoleGibtEsEineLeereNamensliste(t *testing.T) {
	antwort := app.NeueApp("").Skriptnamen(probekontext(t), "1")
	if antwort.Fehler == nil {
		t.Fatal("ohne Konsole kam eine Antwort")
	}

	if antwort.Namen == nil {
		t.Error("die Namen sind nil statt leer")
	}
}

// **Eine unlesbare Kennung wird abgewiesen.**
func TestEineUnlesbareKennungGibtKeineNamen(t *testing.T) {
	anwendung := app.NeueApp("")
	anwendung.MitSkriptkonsole(namenskonsole{})

	if antwort := anwendung.Skriptnamen(t.Context(), "keine Zahl"); antwort.Fehler == nil {
		t.Fatal("die unlesbare Kennung kam durch")
	}
}

// **Und ein Fehler der Umgebung kommt an.**
func TestEinFehlerBeimNamenlesenKommtAn(t *testing.T) {
	anwendung := app.NeueApp("")
	anwendung.MitSkriptkonsole(namenskonsole{fehler: errors.New("keine Karte")})

	antwort := anwendung.Skriptnamen(t.Context(), "1")
	if antwort.Fehler == nil {
		t.Fatal("der Fehler kam als Erfolg durch")
	}

	if antwort.Namen == nil {
		t.Error("die Namen sind nil statt leer")
	}
}

// namenskonsole gibt eine Namensliste oder einen Fehler.
type namenskonsole struct {
	namen  []string
	fehler error
}

func (n namenskonsole) Auswerten(
	context.Context, int64, string,
) (app.KonsolenDTO, error) {
	return app.KonsolenDTO{}, nil
}

func (n namenskonsole) Namen(context.Context, int64) ([]string, error) {
	return n.namen, n.fehler
}

// **Eine Umgebung ohne Namen gibt eine leere Liste, kein `nil`.**
//
// Der Lauf darüber deckt die Gegenrichtung: dort **gibt** es Namen. Diese
// hier deckt den Zweig, den die Wails-Bindung sichtbar macht — über sie
// stünde bei `nil` ein `null` im Fenster, und die Seite prüfte dann auf zwei
// Dinge statt auf eines.
func TestEineUmgebungOhneNamenGibtEineLeereListe(t *testing.T) {
	anwendung := app.NeueApp("")
	anwendung.MitSkriptkonsole(namenskonsole{namen: nil})

	antwort := anwendung.Skriptnamen(t.Context(), "1")
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if antwort.Namen == nil {
		t.Error("die Namen sind nil statt leer")
	}
}

// **Und die Bindung am Fenster geht denselben Weg.**
func TestDieBindungGibtDieSkriptnamen(t *testing.T) {
	anwendung := app.NeueApp("")
	anwendung.MitSkriptkonsole(namenskonsole{namen: []string{"length"}})

	antwort := app.NeueBindungen(anwendung).Skriptnamen(t.Context(), "1")
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if len(antwort.Namen) != 1 || antwort.Namen[0] != "length" {
		t.Errorf("es kamen %v", antwort.Namen)
	}
}

// **Der Satz des Zerteilers erreicht das Fenster.**
//
// # Warum das ein eigener Schlüssel sein muss
//
// `Uebersetze` gibt jedem `ErrValidation` den Schlüssel `unbekannt`, und das
// Fenster **bevorzugt den Katalog** (§3, „zwischen den Prüfebenen"). Am
// laufenden Programm nachgesehen: über dem angestrichenen `* 2` stand „Es ist
// ein unerwarteter Fehler aufgetreten. Zeile 2, Spalte 4" — eine Stelle ohne
// Aussage. Der Zerteiler weiß, was fehlt; er sagt es seit je, und niemand
// bekam es zu lesen.
//
// **Nicht nur die Konsole.** Jeder Weg, auf dem ein Skriptfehler an eine
// Fläche kommt, geht durch `Uebersetze` — deshalb steht der Zweig dort und
// nicht in der Konsolenbindung (§10, „nicht den Fall, sondern die Klasse").
func TestEinSkriptfehlerTraegtSeinenSatz(t *testing.T) {
	anwendung := app.NeueApp("")
	anwendung.MitSkriptkonsole(stellenkonsole{fehler: fehlerMitStelle(2, 3)})

	antwort := anwendung.Skript(t.Context(), "1", "1 +\n  * 2")
	if antwort.Fehler == nil {
		t.Fatal("der Syntaxfehler kam als Erfolg durch")
	}

	if antwort.Fehler.Schluessel != domain.SchluesselSkriptfehler {
		t.Errorf("der Schlüssel ist %q — will %q",
			antwort.Fehler.Schluessel, domain.SchluesselSkriptfehler)
	}

	// **Der Satz steht in den Werten, nicht nur in der Meldung.** Das Fenster
	// setzt den Katalogtext ein und füllt dessen Platzhalter aus `Werte`; wer
	// den Satz nur in `Meldung` legte, hätte ihn wieder verloren.
	if antwort.Fehler.Werte["meldung"] == "" {
		t.Errorf("die Werte tragen keinen Satz: %+v", antwort.Fehler.Werte)
	}
}
