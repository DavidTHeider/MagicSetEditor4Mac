package app

// **Die Nähte der nativen Menüleiste und der Stand, den das Fenster ihr
// meldet** — UXR-F14
// (#2029).
//
// # Warum eine eigene Datei, und warum sie drei Felder mitnimmt
//
// `anwendung.go` stand bei 400 Zeilen, und der Stand der Darstellung hätte sie
// überschritten (§2). Der Schnitt folgt der Frage: dort, **woraus** die
// Anwendung besteht; hier, **was die Menüleiste braucht** — die Rückrufe, die
// Naht zum Fenster, der Meldeweg und seit dieser Änderung der Stand. Die drei
// Felder standen schon dort, wo die Leiste entsteht, und werden hier nur
// zusammengefasst; **angesprochen werden sie weiter als `a.menueSetzen`**,
// denn eingebettete Felder steigen auf.

import (
	"context"
	"sync"
)

// menuenaht ist der Verbund, den [App] einbettet.
type menuenaht struct {
	// menuhandlungen sind die Rückrufe der Menüleiste — sie überleben den
	// Sprachwechsel, weil die Leiste dabei neu gebaut wird.
	menuhandlungen Menuhandlungen

	// menueSetzen ist die Naht zum Fenster.
	//
	// Als Feld und aus demselben Grund wie `oeffnen`: ein Test hätte sonst nur
	// die Wahl zwischen einem echten Fenster und gar keiner Prüfung.
	menueSetzen func(context.Context, *Menue)

	// melden ist die Naht, über die Go dem Fenster etwas sagt — aus demselben
	// Grund als Feld wie `menueSetzen` (#162).
	melden func(context.Context, string, ...any)

	// menuesperre hält die Leiste und das, woraus sie entsteht, zusammen.
	//
	// **Gebaut und gesetzt wird unter ihr**, nicht nur gelesen: zwei Meldungen
	// kurz nacheinander kommen aus zwei Goroutinen, und wer die Leiste außerhalb
	// der Sperre setzte, ließe die ältere die jüngere überschreiben — die
	// Leiste stünde dann im Stand von vorhin. Der Fehler ist keiner, den ein
	// Lauf zeigt; `-race` sieht ihn nur, wenn die Felder darunter ungeschützt
	// wären.
	menuesperre sync.Mutex

	// darstellung ist der letzte Stand, den das Fenster gemeldet hat.
	//
	// **Der Nullwert ist der Anfang** (siehe [Darstellungsstand]).
	darstellung Darstellungsstand

	// zuletztliste ist die Liste der jüngsten Sets, mit der die Leiste zuletzt
	// gebaut wurde.
	//
	// **Gemerkt, damit ein Stand nicht den Bestand liest.** Die Leiste wird bei
	// jedem Wechsel des Standes neu gebaut — beim Wechsel einer Seite, beim
	// Klappen einer Spalte —, und die Liste ändert sich dabei nicht: sie
	// ändert sich, wenn ein Set aufgeschlagen wird, das Profil wechselt oder
	// die Sprache (`menueNeu`). Bei einem Stand eine Abfrage über alle Sets zu
	// stellen, hieße, für jeden Klick auf eine Spalte den Bestand zu lesen.
	zuletztliste []Zuletzteintrag
	// objekthandlungen ist die Handlungsliste des gewählten Objekts, wie das
	// Fenster sie zuletzt gemeldet hat (`objektmenue.go`).
	objekthandlungen Objekthandlungen
}

// DarstellungsstandMelden nimmt den Stand entgegen, den das Fenster der
// Menüleiste über seine Teile sagt, und baut die Leiste neu.
//
// # Eine Meldung, kein Befehl
//
// **Go stellt nichts.** Die Einträge melden ein Ereignis ins Fenster
// ([App.Fensterhandlungen]); das Fenster klappt die Seitenleiste oder eine
// Spalte und sagt hier, wie es danach dasteht. Ein Stand, den Go selbst
// umlegte, wäre eine zweite Wahrheit — und die laufen auseinander, sobald die
// Spalte aus einem anderen Grund zuklappt (Enge, eine Schublade, ein anderer
// Weg wie ⌥⌘[).
//
// # Nichts Neues bei gleichem Stand
//
// **Gleicher Stand, keine neue Leiste.** Das Fenster meldet bei jedem Wechsel
// der Seite, und zwischen zwei Arbeitsplätzen ändert sich nichts; macOS die
// Leiste dabei jedes Mal neu bauen zu lassen, wäre Arbeit für nichts.
//
// # Und gesperrt?
//
// **Der Stand wird gemerkt, aber nicht gezeigt.** Hinter dem Sperrbildschirm
// gibt es weder Seitenleiste noch Arbeitsplatz ([menueAufbauen]); eine späte
// Meldung eines Fensters, das gesperrt wurde, öffnet nichts wieder.
func (a *App) DarstellungsstandMelden(stand Darstellungsstand) SchreibAntwort {
	a.menuesperre.Lock()
	defer a.menuesperre.Unlock()

	if a.darstellung == stand {
		return SchreibAntwort{}
	}

	a.darstellung = stand
	a.menueAufbauen()

	return SchreibAntwort{}
}

// menueAufbauen baut die Leiste aus dem, was gemerkt ist, und setzt sie.
//
// **Wer es ruft, hält die Sperre** ([menuenaht.menuesperre]).
//
// **Gesperrt oder abgemeldet gilt der Anfangsstand.** [App.inDerSitzung] baut
// die Leiste ohnehin neu, wenn gesperrt, entsperrt oder abgemeldet wird (sie
// zeigt dann keine Sets); hier kommt dazu, dass auch Seitenleiste, Navigator und
// Inspector gegraut sind. Die Sperre ist eine Eigenschaft der Sitzung und kein
// Teil dessen, was das Fenster meldet — **abgeleitet beim Bauen und nicht
// gespeichert**, damit das Entsperren den Stand von vorhin zurückbringt und
// kein Weg ihn vergisst (das Fenster meldet beim Wiederaufbau ohnehin neu).
//
// **Und gesperrt gibt es kein Objektmenü**: hinter dem Sperrbildschirm ist
// nichts gewählt.
func (a *App) menueAufbauen() {
	stand, objekt := a.darstellung, a.objekthandlungen
	if a.sitzungVerschlossen() {
		stand, objekt = Darstellungsstand{}, Objekthandlungen{}
	}

	leiste := MenueleisteMitStand(a.menuhandlungen, MenuetexteFuer(a.sprache), a.zuletztliste, stand)
	objektmenueEinsetzen(leiste, objekt, func(ausloesung Objektausloesung) {
		a.melden(a.fenster, EreignisObjekthandlung, ausloesung)
	})
	a.menueSetzen(a.fenster, leiste)
}
