// **Das Tor der Sitzung an der Bindungsgrenze** — Zweitblick über ID-02,
// Befund 4 (#1701).
//
// # Warum an jeder Bindung und warum mit einer Panik
//
// **Weil Wails keinen Haken vor einem Aufruf kennt** (beta.23:
// `ServiceOptions` hat Name, Route und `MarshalError`, sonst nichts). Jede
// gebundene Methode kommt an die Anwendung nur über [Bindungen.app] — das
// Feld heißt anders —, und dort steht das Tor. **Eine Panik hält den Ruf an,
// bevor er etwas tut**, und Wails gibt sie als Aufruffehler zurück, statt die
// Anwendung zu beenden (`pkg/application/bindings.go`, `(*BoundMethod).Call`,
// nachgelesen am 2026-09-27).
//
// # Was vorher hier stand
//
// **Das Tor stand in [App.kontext] und erkannte seinen Rufer am Namen**
// (`runtime.Caller`). Der Nachlauf über ID-02 fand drei Wege daran vorbei:
// `IndexAbbrechen` und `KIBildAbbrechen` riefen den Kontext nie,
// `SpracheSetzen` rief ihn erst nach der Wirkung und über eine Methode der
// Anwendung, und `SetOeffnen` öffnete den Dialog davor. Gemessen am 2026-09-27:
// **32 von 458** Bindungen gingen gesperrt durch — darunter `KIVorschlagen`.
//
// # Was es nicht hält
//
// **Einen Weg ohne Fenster** — Kommandozeile, MCP, REST. Sie haben eigene
// Rechte (`domain.Akteur`); die Sitzung ist die des Fensters.

package app

import "errors"

// ErrSitzungVerschlossen ist, was eine gebundene Methode bei gesperrter oder
// beendeter Sitzung zurückbekommt.
var ErrSitzungVerschlossen = errors.New("die Anwendung ist gesperrt — entsperre sie zuerst")

// sitzungsfrei sind die gebundenen Methoden, die Sperranzeige und Anmeldung
// brauchen — und nichts sonst. **Nur sie rufen [Bindungen.frei]**;
// `sitzungstor_intern_test.go` hält beide Richtungen.
var sitzungsfrei = map[string]bool{
	"Sitzungsstand": true, "Entsperren": true, "EntsperrenMitSensor": true,
	"Abmelden": true, "LokalAnmelden": true, "KennwortWiederherstellen": true,
	// **Die Wahl, wer arbeitet**, und die Sprache der Sperranzeige.
	"Nutzer": true, "Einstellungen": true,
	// **Die Sprache der Menüleiste**: die Sperranzeige spricht die des
	// Fensters, und die Leiste soll es auch. Sie ändert keinen Bestand und
	// kein Profil — nur, in welcher Sprache die Leiste steht.
	"SpracheSetzen": true,
	// **Ob der Bestand aufging** — die Wurzel des Fensters fragt es vor dem
	// Tor, weil davon abhängt, ob überhaupt eine Sitzung dasteht. Gesperrt
	// ist die Antwort immer „offen“: ohne Bestand gibt es keine Sitzung.
	"Startbefund": true,
}

// sitzungstor hält eine gebundene Methode an, solange die Sitzung
// verschlossen ist.
func (a *App) sitzungstor() {
	if a.sitzungVerschlossen() {
		panic(ErrSitzungVerschlossen)
	}
}
