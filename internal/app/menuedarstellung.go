package app

// **Das Menü *Darstellung*, nach dem Stand des Fensters** — UXR-F14
// (#2029),
// Folge aus UXR-02 (#1910).
//
// # Was hier neu ist
//
// Seitenleiste, Navigator und Inspector lassen sich ein- und ausblenden, und
// ihre Einträge heißen **nach dem Stand**: „einblenden", wenn das Teil zu ist,
// „ausblenden", wenn es steht. So verlangt es Apple (Human Interface
// Guidelines, „The menu bar", am 2026-10-02 nachgelesen für UXR-02; der Befund
// steht in #2029). Wo die Seite das Teil gar nicht hat, ist der Eintrag
// gegraut.
//
// **Bis hierher hieß die Seitenleiste „ein-/ausblenden"**, Navigator und
// Inspector standen in keinem Menü — nur ihre Kürzel waren im Fenster
// verdrahtet (`rahmen/tastatur.ts`).
//
// # Woher der Stand kommt
//
// **Aus dem Fenster, und Go merkt ihn sich nur.** Ob die Seite einen
// Arbeitsplatz hat, ob die Spalte steht oder gerade unter einer Schublade
// liegt, weiß allein die Oberfläche; ein Stand, den Go selbst führte, liefe
// auseinander (der Vermerk bei [App.Fensterhandlungen] sagt es für alle
// Einträge). Der Eintrag **meldet** deshalb ein Ereignis, das Fenster schaltet,
// und der neue Stand kommt über [App.DarstellungsstandMelden] zurück — derselbe
// Weg für alle drei Teile.
//
// # Und MSE2?
//
// **Navigator und Inspector gibt es dort nicht**: `SetWindow` führt seine
// Bereiche über eine Reiterleiste (`src/gui/set/window.cpp`, Fassung f9d9356
// vom 2021-07-06), und beide Wörter sind Begriffe dieses Entwurfs
// (`docs/PROTOTYP-0.6.0.md`). Gesucht am 2026-10-03 in `src/gui/set/` nach
// „navigator" und „inspector": kein Treffer. Diese Einträge sind **new**.
//
// **Dass ein Eintrag nach dem Stand heißt und grau ist, wo er nichts täte,
// gibt es dort schon**: `SetWindow::onUpdateUI` (dieselbe Datei, Zeilen
// 498–541) setzt je Kennung `Enable` und `SetText` — Rückgängig mit dem Namen
// des Schrittes, `ID_WINDOW_KEYWORDS` nur, wo das Spiel Schlüsselwörter hat.
// Das Verhalten ist **adapted**, und der Weg ist ein anderer: wxWidgets fragt
// das Fenster beim Öffnen des Menüs (ein Update-UI-Ereignis); diese Leiste ist
// nativ, Wails schaltet das Selbstaktivieren ab (`fenster/menue.go`), und Go
// kann das Fenster nicht fragen — also **meldet** das Fenster.

// Fensterteil sagt, wie ein Teil des Fensters dasteht, das *Darstellung*
// ein- und ausblendet.
type Fensterteil struct {
	// Vorhanden sagt, dass es dort etwas umzulegen gibt. Ohne das ist der
	// Eintrag gegraut: die Seite hat keinen Arbeitsplatz, die Leiste steht
	// oben, oder eine Schublade liegt über dem Platz und fängt die Tasten.
	Vorhanden bool `json:"vorhanden"`

	// Offen sagt, dass das Teil steht — neben dem Zentrum, bei Enge darüber.
	// Der Eintrag heißt dann „ausblenden".
	Offen bool `json:"offen"`
}

// Darstellungsstand ist, was das Fenster der Menüleiste über sich sagt.
//
// **Der Nullwert ist der Anfang**: nichts ist da, alles ist zu — so steht das
// Wiederherstellungsfenster da, so steht der Sperrbildschirm da, und so
// beginnt die Leiste, bevor die Oberfläche etwas gemeldet hat.
type Darstellungsstand struct {
	Seitenleiste Fensterteil `json:"seitenleiste"`
	Navigator    Fensterteil `json:"navigator"`
	Inspector    Fensterteil `json:"inspector"`
}

// darstellungsmenue hängt das Menü *Darstellung* an.
//
// **Zwischen Bearbeiten und dem Setmenü**, wie macOS es führt (UXR-02): die
// Stelle bestimmt der Aufruf. Hier steht, was darin ist — erst die drei Teile,
// die das Fenster aufteilen, dann der Gang der Seiten, dann das Zurücksetzen,
// durch Trennlinien getrennt.
func darstellungsmenue(
	leiste *Menue, handlungen Menuhandlungen, texte Menuetexte, stand Darstellungsstand,
) {
	darstellung := leiste.Untermenue(texte.Darstellung)

	// **⌥⌘S wie im Finder und in Mail** (#1490) — und ⌥⌘[ und ⌥⌘] sind die
	// Kürzel, die der Arbeitsplatz schon hört (`bausteine/Arbeitsplatz.tsx`).
	// Ein Kürzel, das nur im Fenster galt, stand in keinem Menü und war so
	// nirgends zu finden (§6).
	darstellung.umlegen(stand.Seitenleiste, &Kuerzel{Taste: "s", Wahl: true},
		texte.SeitenleisteAusblenden, texte.SeitenleisteEinblenden,
		rueckruf(handlungen.SeitenleisteUmschalten))
	darstellung.umlegen(stand.Navigator, &Kuerzel{Taste: "[", Wahl: true},
		texte.NavigatorAusblenden, texte.NavigatorEinblenden,
		rueckruf(handlungen.NavigatorUmschalten))
	darstellung.umlegen(stand.Inspector, &Kuerzel{Taste: "]", Wahl: true},
		texte.InspectorAusblenden, texte.InspectorEinblenden,
		rueckruf(handlungen.InspectorUmschalten))

	darstellung.Trennlinie()
	darstellung.Befehl(texte.SeiteZurueck, Befehlstaste("["), rueckruf(handlungen.SeiteZurueck))
	darstellung.Befehl(texte.SeiteVor, Befehlstaste("]"), rueckruf(handlungen.SeiteVor))
	darstellung.Trennlinie()
	darstellung.Befehl(texte.DarstellungZuruecksetzen, nil,
		rueckruf(handlungen.DarstellungZuruecksetzen))
}

// umlegen hängt den Eintrag eines Fensterteils an: er heißt nach dem Stand,
// und er ist gegraut, wo es nichts umzulegen gibt.
//
// **Der Titel folgt `Offen` auch dann, wenn nichts vorhanden ist.** Liegt eine
// Schublade über dem Arbeitsplatz, steht die Spalte darunter weiter da, und
// der graue Eintrag sagt, wie — statt auf „einblenden" zu springen und etwas
// zu behaupten, das nicht stimmt.
func (m *Menue) umlegen(teil Fensterteil, kuerzel *Kuerzel, ausblenden, einblenden string, handlung func()) {
	beschriftung := einblenden
	if teil.Offen {
		beschriftung = ausblenden
	}

	m.BefehlWenn(teil.Vorhanden, beschriftung, kuerzel, handlung)
}
