package app

import (
	"context"
)

// Die Ereignisse, die die native Menüleiste ins Fenster meldet.
//
// **Sie stehen einmal hier und einmal im Frontend** (`frontend/src/fenster.ts`).
// Eine gemeinsame Quelle gäbe es nur über die erzeugten Bindungen, und die
// tragen Methoden, keine Konstanten. Der Preis ist ein Wächter auf jeder Seite;
// bezahlt wird er dadurch, dass ein Tippfehler sonst eine Anwendung ergibt, in
// der ein Kürzel nichts tut — ohne Fehler, ohne Meldung, ohne roten Lauf.
const (
	EreignisEinstellungen   = "menue:einstellungen"
	EreignisSetSichern      = "menue:set-sichern"
	EreignisSetSichernUnter = "menue:set-sichern-unter"

	// **Und der Bestand selbst** — nicht das Set, sondern die Datenbank
	// (#799).
	//
	// `storage.Sichern` schreibt seit
	// #798 eine
	// vollständige Sicherung, während gearbeitet wird — **aufrufen konnte sie
	// niemand.** `docs/RELEASE.md` führte deshalb den Weg über `sqlite3` von
	// Hand, und „sichere deine Sets, indem du ein Terminal aufmachst" ist keine
	// Antwort für jemanden, der Karten baut.
	EreignisBestandSichern = "menue:bestand-sichern"

	// **Öffnen, Schließen, Drucken und die Hilfe** (#891). ⌘O, ⌘W und ⌘P
	// waren nicht belegt, und ein Hilfemenü gab es nicht.
	EreignisSetNeu        = "menue:set-neu"
	EreignisSetZuletzt    = "menue:set-zuletzt"
	EreignisSetOeffnen    = "menue:set-oeffnen"
	EreignisSetSchliessen = "menue:set-schliessen"
	EreignisDrucken       = "menue:drucken"
	EreignisHilfe         = "menue:hilfe"

	// **Das Setmenü und „Über"** — Werkstück 4 von Phase 3.7 (#1168, Z4-01).
	//
	// **`EreignisSeite` trägt einen Seitenschlüssel**; die übrigen tragen
	// nichts. Ein Ereignis je Lasche wären fünf Namen für eine Handlung.
	EreignisRueckgaengig = "menue:rueckgaengig"

	// EreignisAllesAuswaehlen hakt jede Karte der Kartenliste an
	// (#1242).
	EreignisAllesAuswaehlen  = "menue:alles-auswaehlen"
	EreignisWiederherstellen = "menue:wiederherstellen"
	EreignisSeite            = "menue:seite"
	EreignisUeber            = "menue:ueber"

	// EreignisSeitenleiste klappt die Seitenleiste ein oder aus (#1490).
	EreignisSeitenleiste = "menue:seitenleiste"

	// EreignisSeiteZurueck und EreignisSeiteVor gehen den Gang der Seiten
	// (UX-03.AK-02, #1675).
	EreignisSeiteZurueck = "menue:seite-zurueck"
	EreignisSeiteVor     = "menue:seite-vor"

	// EreignisNavigator und EreignisInspector klappen die Spalten des
	// Arbeitsplatzes ein oder aus (UXR-F14, #2029). **Ohne Nutzlast**, wie
	// [EreignisSeitenleiste]: der Eintrag sagt „umschalten", und was das
	// heißt, weiß die Spalte — sie steht oder nicht, und das Fenster hat es.
	EreignisNavigator = "menue:navigator"
	EreignisInspector = "menue:inspector"

	// EreignisDarstellungZuruecksetzen stellt das Aussehen zurück
	// (UX-09.AK-03, #1681) — das Fenster führt die Layoutrücksetzung aus.
	EreignisDarstellungZuruecksetzen = "menue:darstellung-zuruecksetzen"

	// EreignisObjekthandlung meldet einen Klick im Menü des gewählten Objekts
	// (GESTE, #2358) — **mit Nutzlast** [Objektausloesung]: Objekt, Epoche und
	// Schlüssel, damit das Fenster prüfen kann, ob der Klick noch zu dem
	// gehört, was gewählt ist (`objektmenue.go`).
	EreignisObjekthandlung = "menue:objekthandlung"
)

// MitMeldung setzt die Naht, über die Go dem Fenster etwas sagt.
//
// **Als Feld und aus demselben Grund wie `menueSetzen`**: melden braucht ein
// laufendes Fenster, und ein Test hat keines. Ohne diese Naht wäre die ganze
// Menüleiste nur über ein Fenster prüfbar — also gar nicht. Gesetzt wird sie
// von `internal/fenster`, bevor das Fenster aufgeht (ENG-02).
func (a *App) MitMeldung(melden func(context.Context, string, ...any)) { a.melden = melden }

// Fensterhandlungen gibt die Rückrufe der Menüleiste.
//
// # Warum die Einträge nur melden
//
// **Der Menüeintrag löst aus; was zu tun ist, entscheidet das Frontend**
// (#162).
//
// Die andere Möglichkeit wäre gewesen, dass das Frontend nach jedem Wechsel
// meldet, welches Set offen ist, und Go es sich merkt. Das ist die größere
// Bauweise und die schlechtere: sie legt eine **fachliche** Entscheidung —
// „welches Set ist gemeint?" — in eine Schicht, die keine treffen soll (§2),
// und sie führt denselben Zustand zweimal. Zwei Stellen, die dasselbe wissen,
// laufen auseinander; die Frage ist nur, wann.
//
// **Der Preis gehört dazu:** Go erfährt nicht, ob etwas geschehen ist. Ein ⌘S
// ohne offenes Set löst ein Ereignis aus, das niemand beantwortet — und das ist
// genau, was das Issue verlangt: nichts tun, statt einen Fehler zu zeigen.
func (a *App) Fensterhandlungen() Menuhandlungen {
	return Menuhandlungen{
		EinstellungenOeffnen: a.meldet(EreignisEinstellungen),
		SetSichern:           a.meldet(EreignisSetSichern),
		SetSichernUnter:      a.meldet(EreignisSetSichernUnter),
		BestandSichern:       a.meldet(EreignisBestandSichern),
		SetNeu:               a.meldet(EreignisSetNeu),
		SetZuletzt:           a.meldetMit(EreignisSetZuletzt),
		SetOeffnen:           a.meldet(EreignisSetOeffnen),
		SetSchliessen:        a.meldet(EreignisSetSchliessen),
		Drucken:              a.meldet(EreignisDrucken),
		HilfeOeffnen:         a.meldet(EreignisHilfe),

		// **Die Rücknahme geht denselben Weg wie alles andere** (#1168): der
		// Eintrag meldet, das Fenster entscheidet. Es weiß, welches Set offen
		// ist — und es hat den Verlauf schon für ⌘Z.
		Rueckgaengig:     a.meldet(EreignisRueckgaengig),
		AllesAuswaehlen:  a.meldet(EreignisAllesAuswaehlen),
		Wiederherstellen: a.meldet(EreignisWiederherstellen),
		SeiteOeffnen:     a.meldetMit(EreignisSeite),
		UeberOeffnen:     a.meldet(EreignisUeber),

		SeitenleisteUmschalten: a.meldet(EreignisSeitenleiste),
		SeiteZurueck:           a.meldet(EreignisSeiteZurueck),
		SeiteVor:               a.meldet(EreignisSeiteVor),

		NavigatorUmschalten: a.meldet(EreignisNavigator),
		InspectorUmschalten: a.meldet(EreignisInspector),

		DarstellungZuruecksetzen: a.meldet(EreignisDarstellungZuruecksetzen),
	}
}

// meldetMit baut den Rückruf, der ein Ereignis **samt Wert** ins Fenster gibt.
//
// **Für „Zuletzt benutzt"** (#891): welches Set gemeint ist, weiß der Eintrag,
// und ohne die Kennung müsste das Fenster raten.
func (a *App) meldetMit(ereignis string) func(string) {
	return func(wert string) { a.melden(a.fenster, ereignis, wert) }
}

// meldet baut den Rückruf, der ein Ereignis ins Fenster gibt.
func (a *App) meldet(ereignis string) func() {
	return func() { a.melden(a.fenster, ereignis) }
}
