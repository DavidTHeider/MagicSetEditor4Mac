package app

// Menuhandlungen sind die Rückrufe, die die Menüleiste auslöst.
//
// Als Struktur und nicht als Parameterliste: die Leiste wächst mit jeder Phase,
// und eine Funktion mit sieben Rückrufen in Reihenfolge ist eine Fehlerquelle,
// die kein Übersetzer bemerkt.
type Menuhandlungen struct {
	// EinstellungenOeffnen wird von ⌘, ausgelöst.
	EinstellungenOeffnen func()

	// SetSichern wird von ⌘S ausgelöst, SetSichernUnter von ⇧⌘S (#162).
	SetSichern      func()
	SetSichernUnter func()

	// BestandSichern schreibt eine Sicherung der ganzen Datenbank (#799).
	//
	// **Ohne Kürzel.** ⌘S und ⇧⌘S sind vergeben, und ein drittes Kürzel für
	// eine Handlung, die man selten braucht, nähme eines weg, das man oft
	// braucht.
	BestandSichern func()

	// SetOeffnen liegt auf ⌘O, Drucken auf ⌘P (#891).
	//
	// **Beides sind gewohnte Bedeutungen unter macOS**, und beides tat bis
	// zum 2026-08-25 nichts.
	SetOeffnen func()
	Drucken    func()

	// SetNeu liegt auf ⌘N — ein leeres Set, ohne Datei (#891).
	//
	// **MSE2 hat den Eintrag ebenfalls ganz oben** (`src/gui/set/window.cpp`,
	// Fassung f9d9356 vom 2021-07-06, Zeile 58): `ID_FILE_NEW` vor
	// `ID_FILE_OPEN`.
	SetNeu func()

	// SetSchliessen liegt auf ⌘W und schließt das **Set**, nicht das Fenster.
	//
	// # Eine Abweichung von MSE2, und sie ist begründet (§15, Regel 7)
	//
	// MSE2 hat gar kein „Schließen": sein Ablagemenü führt Neu, Öffnen,
	// Sichern, Sichern unter, Ausfuhr, Drucken und Beenden
	// (`src/gui/set/window.cpp`, Fassung f9d9356 vom 2021-07-06, Zeilen
	// 57–77). Unter macOS ist ⌘W trotzdem belegt erwartet — und das
	// Fenstermenü führt nur Im Dock ablegen, Zoomen und Alle nach vorne.
	//
	// **Geschlossen wird deshalb das Dokument.** Diese Anwendung hat ein
	// Fenster, und es zu schließen hieße, sie zu beenden — dafür gibt es ⌘Q.
	SetSchliessen func()

	// HilfeOeffnen führt auf die Hilfeseite (#891).
	HilfeOeffnen func()

	// SetZuletzt schlägt eines der zuletzt benutzten Sets auf (#891).
	SetZuletzt func(setID string)

	// **Was Werkstück 4 von Phase 3.7 dazugebracht hat** (#1168, Z4-01).

	// Rueckgaengig und Wiederherstellen rufen den **fachlichen** Verlauf.
	//
	// # Warum sie nicht im Bearbeitenmenü stehen
	//
	// **Weil das macOS gehört.** Dessen „Widerrufen" hängt an der
	// Responder-Kette: es nimmt zurück, was jemand **getippt** hat, und weiß
	// von Karten nichts. Ohne die Einträge dort gäbe es kein ⌘C und ⌘V — sie
	// müssen also dableiben ([bearbeitenmenue]).
	//
	// **Also ein eigener Eintrag, und einer ohne Kürzel.** ⌘Z liegt weiter
	// bei macOS; dass es trotzdem den fachlichen Verlauf trifft, liegt daran,
	// dass AppKit den Eintrag außerhalb eines Textfeldes abschaltet und die
	// Taste ins Fenster durchfällt, wo `rahmen/tastatur.ts` sie deutet. Ein
	// zweites ⌘Z hier nähme AppKit diese Wahl ab — und meinte **immer** die
	// fachliche Rücknahme, auch mitten im Kartentext.
	Rueckgaengig     func()
	Wiederherstellen func()

	// SeiteOeffnen führt auf eine Lasche — Suchen, Ersetzen, Ausgeben.
	//
	// **Ein Rückruf mit Wert und nicht fünf ohne.** Was die Einträge tun, ist
	// dasselbe: eine Seite aufschlagen. Fünf Felder daneben wären fünf Wege
	// durch dieselbe Naht, und jeder müsste einzeln verdrahtet werden.
	SeiteOeffnen func(seite string)

	// AllesAuswaehlen hakt jede Karte der Kartenliste an (#1242).
	AllesAuswaehlen func()

	// UeberOeffnen führt auf den Abschnitt „Über" der Hilfe.
	//
	// **Ein eigener Rückruf und nicht `SeiteOeffnen("hilfe")`**: „Über" und
	// „Hilfe" sind zwei Einträge, und zwei Einträge, die dasselbe tun, sind
	// einer zu viel. Das Fenster rollt zum Abschnitt.
	UeberOeffnen func()

	// SeitenleisteUmschalten klappt die Seitenleiste ein oder aus — Werkstück
	// U14 von Phase 5.1 (#1490):
	// ⌥⌘S wie im Finder und in Mail, unter *Darstellung*, damit das Kürzel dort
	// sichtbar ist (§6: bedient, nicht nachgeahmt).
	SeitenleisteUmschalten func()

	// NavigatorUmschalten und InspectorUmschalten klappen die Spalten des
	// Arbeitsplatzes ein oder aus — UXR-F14
	// (#2029):
	// ⌥⌘[ und ⌥⌘], die Kürzel, die der Arbeitsplatz im Fenster schon hört.
	//
	// **Sie schalten um und stellen nichts**, wie die Seitenleiste: der
	// Eintrag meldet, das Fenster klappt, und wie die Spalte danach steht,
	// kommt über [App.DarstellungsstandMelden] zurück (`menuedarstellung.go`).
	NavigatorUmschalten func()
	InspectorUmschalten func()

	// SeiteZurueck und SeiteVor gehen den Gang der Seiten — UX-03.AK-02
	// (#1675):
	// ⌘[ und ⌘] — die Kürzel des Finders für den vorigen und den nächsten
	// Ordner (Apple, „Mac-Tastaturkurzbefehle", support.apple.com/de-de/102650,
	// nachgelesen am 2026-09-24). Unter *Darstellung*, weil dort schon die
	// Seitenleiste steht; den Gang führt das Fenster, der Eintrag meldet nur.
	SeiteZurueck func()
	SeiteVor     func()

	// DarstellungZuruecksetzen stellt das Aussehen des Fensters zurück —
	// UX-09.AK-03 (#1681):
	// *„ein sicherer Standard kann auch bei unlesbarer Konfiguration
	// wiederhergestellt werden."* **Die Menüleiste zeichnet macOS**, sie
	// bleibt lesbar, gleich welche Farben im Fenster stehen.
	DarstellungZuruecksetzen func()
}

// Zuletzteintrag ist ein Set im Untermenü „Zuletzt benutzt".
type Zuletzteintrag struct {
	ID   string
	Name string
}

// hoechstensZuletzt ist die Zahl der Einträge.
//
// **Neun, wie MSE2** — `Settings::max_recent_sets` (`src/data/settings.hpp`,
// Fassung f9d9356 vom 2021-07-06, Zeile 139). Eine andere Zahl wäre eine
// stillschweigende Änderung bestehender Semantik (§15, Regel 4).
const hoechstensZuletzt = 9

// Menueleiste baut die Menüleiste im **Anfangsstand**: die Oberfläche hat noch
// nichts gemeldet, Seitenleiste, Navigator und Inspector stehen gegraut da.
//
// So beginnt die Leiste (`internal/fenster`), und so sieht sie aus, wo es
// keine Oberfläche gibt. Wer den Stand des Fensters kennt, baut sie mit
// [MenueleisteMitStand].
func Menueleiste(
	handlungen Menuhandlungen, texte Menuetexte, zuletzt []Zuletzteintrag,
) *Menue {
	return MenueleisteMitStand(handlungen, texte, zuletzt, Darstellungsstand{})
}

// MenueleisteMitStand baut die Menüleiste der Anwendung.
//
// **Der Stand führt das Menü *Darstellung*** (UXR-F14, #2029): Titel und
// Aktivierung seiner drei Teile — `menuedarstellung.go` sagt, woher er kommt.
//
// Sie ist reine Datenstruktur — erst Wails macht daraus etwas Sichtbares. Das
// ist der Grund, warum sie hier steht und nicht in `main.go`: so lässt sie sich
// prüfen, ohne ein Fenster zu öffnen.
//
// # Warum Anwendungs-, Bearbeiten- und Fenstermenü Bäume sind
//
// **Seit Wails 3 lassen sie sich selbst bauen** (#255). Unter Wails 2 waren es
// Rollen ohne Baum, mit englischen Beschriftungen in Objective-C (#822,
// Vermerk 0030): die Leiste stand halb deutsch da, und die Einstellungen
// lagen in der Ablage, weil sich ins Anwendungsmenü nichts hängen ließ.
// Wails 3 baut einen Eintrag mit Rolle und lässt ihn beschriften, ohne dass er
// seinen Selektor verliert (`menuitem_darwin.go`, `getSelectorForRole`,
// v3.0.0-beta.23). Die Leiste führt deshalb **Einträge** mit Rolle, und sie
// heißen in der Sprache des Fensters so, wie macOS sie nennt.
//
// **„Über" und die Einstellungen stehen im Anwendungsmenü**, wo macOS sie
// erwartet. „Über" auch deshalb, weil MSE2 es unter `wxID_ABOUT` führt
// (`src/util/window_id.hpp`, Zeile 73, Fassung f9d9356 vom 2021-07-06) und
// wxWidgets diese Kennung unter macOS ins Anwendungsmenü stellt (Dokumentation
// zu `wxMenu`, 3.2, nachgelesen am 2026-09-24). Die Einstellungen führt MSE2
// unter einer eigenen Kennung im Bearbeitenmenü (`ID_EDIT_PREFERENCES`); hier
// gilt, was §6 verlangt — die Stelle, an der man unter macOS sucht.
//
// # Woher die Beschriftungen kommen
//
// Aus `menuetexte.go` und nicht aus dem Quelltext hier. Bis Phase 5 standen sie
// als deutsche Zeichenketten mitten in dieser Funktion — gegen §9, und **kein
// Wächter sah es**: `keintextimjsx.test.ts` liest JSX, und die Menüleiste ist
// nativ. Jetzt weist `menue_intern_test.go` eine Beschriftung im Quelltext ab.
func MenueleisteMitStand(
	handlungen Menuhandlungen, texte Menuetexte, zuletzt []Zuletzteintrag, stand Darstellungsstand,
) *Menue {
	leiste := &Menue{}

	anwendungsmenue(leiste, handlungen, texte)

	ablage := leiste.Untermenue(texte.Ablage)

	// **Neu vor Öffnen**, wie in jeder Ablage unter macOS — und wie in MSE2
	// (`src/gui/set/window.cpp`, Fassung f9d9356 vom 2021-07-06, Zeilen
	// 58–59).
	ablage.Befehl(texte.Neu, Befehlstaste("n"), rueckruf(handlungen.SetNeu))
	ablage.Befehl(texte.Oeffnen, Befehlstaste("o"),
		rueckruf(handlungen.SetOeffnen))

	ablage.Trennlinie()

	// **Sichern vor Einstellungen**, denn so steht es in jeder Ablage unter
	// macOS: die Dateibefehle oben, das Übrige darunter.
	ablage.Befehl(texte.Sichern, Befehlstaste("s"),
		rueckruf(handlungen.SetSichern))
	ablage.Befehl(texte.SichernUnter,
		&Kuerzel{Taste: "s", Umschalt: true},
		rueckruf(handlungen.SetSichernUnter))

	// **Schließen unter dem Sichern**: erst das Set wegschreiben, dann
	// zumachen — die Reihenfolge, in der man es tut.
	ablage.Befehl(texte.Schliessen, Befehlstaste("w"),
		rueckruf(handlungen.SetSchliessen))

	// **Das Ausgabemenü zwischen Sichern und Zuletzt** — dort steht es in
	// MSE2 (`add_menu_item_tr(menuFile, wxID_ANY, "export", "export", …,
	// makeExportMenu())`, `src/gui/set/window.cpp`, Zeile 63, Fassung f9d9356
	// vom 2021-07-06).
	//
	// **Ohne Kürzel und ohne Dateidialog.** Die Einträge führen auf ihre
	// Lasche; was dort gewählt wird — Vorlage, Maße, Ziel —, entscheidet die
	// Seite. Ein Menüeintrag, der unmittelbar schriebe, müsste all das raten.
	ausgabe := ablage.Untermenue(texte.Ausgeben)
	ausgabe.Befehl(texte.Exportvorlage, nil, seitenruf(handlungen.SeiteOeffnen, "ausgeben"))
	ausgabe.Befehl(texte.Bilder, nil, seitenruf(handlungen.SeiteOeffnen, "bilder"))
	ausgabe.Befehl(texte.Druckbogen, nil, seitenruf(handlungen.SeiteOeffnen, "druckbogen"))

	// **„Zuletzt benutzt" zwischen Schließen und Drucken** — dort steht es in
	// MSE2 auch: nach den Dateibefehlen und vor dem Ende
	// (`src/gui/set/window.cpp`, Fassung f9d9356 vom 2021-07-06, Zeile 76).
	//
	// **Ohne Einträge steht es nicht da.** Ein leeres Untermenü ist ein
	// Versprechen, das beim Aufklappen zurückgenommen wird.
	if len(zuletzt) > 0 {
		liste := ablage.Untermenue(texte.Zuletzt)

		for _, eintrag := range zuletzt {
			liste.Befehl(eintrag.Name, nil, zuletztruf(handlungen.SetZuletzt, eintrag.ID))
		}
	}

	ablage.Trennlinie()
	ablage.Befehl(texte.Drucken, Befehlstaste("p"),
		rueckruf(handlungen.Drucken))
	// **Der Bestand steht unter dem Set und über den Einstellungen** (#799):
	// beides sind Dateibefehle, aber „das ganze Programm sichern" ist ein
	// anderer Gegenstand als „dieses Set sichern" — deshalb eine Trennlinie
	// dazwischen und nicht daneben.
	ablage.Trennlinie()
	ablage.Befehl(texte.BestandSichern, nil, rueckruf(handlungen.BestandSichern))

	// **Bearbeiten hängt an der Responder-Kette** — ohne diese Einträge gäbe
	// es kein ⌘C, ⌘V und ⌘Z in einem Textfeld.
	bearbeitenmenue(leiste, texte)

	// **Darstellung zwischen Bearbeiten und Fenster** — dort steht es unter
	// macOS, und dort sucht man ⌥⌘S (#1490). Was darin steht und wie die
	// Einträge heißen, steht in `menuedarstellung.go`.
	darstellungsmenue(leiste, handlungen, texte, stand)

	// **Das Setmenü** (#1168, Z4-01) — was MSE2 im Bearbeitenmenü führt und
	// hier nirgends stand.
	//
	// **Zwischen Darstellung und Fenster** — UXR-02 (#1910): dort stehen unter
	// macOS die eigenen Menüs einer Anwendung (Apple, Human Interface
	// Guidelines, „The menu bar", Abschnitt *App-specific menus*, nachgelesen
	// am 2026-10-02), und MSE2 fügt seine ebenso hinter Bearbeiten ein
	// (`mb->Insert(2, menuCard, …)`, `src/gui/set/cards_panel.cpp`, Zeile
	// 208, Fassung f9d9356 vom 2021-07-06). Bis dahin stand es zwischen
	// Ablage und Bearbeiten — gegen beide.
	//
	// **Neben und nicht in _Bearbeiten_**: dessen Einträge hängen an der
	// Responder-Kette von macOS und wirken im Textfeld ([bearbeitenmenue]).
	// Der Vermerk bei [Menuhandlungen.Rueckgaengig] sagt, was das für ⌘Z
	// bedeutet.
	//
	// **Was hier fehlt und warum** — die drei aus dem Audit, die nicht
	// gebaut sind:
	//
	//   - **Weitersuchen** (`ID_EDIT_FIND_NEXT`). MSEs Suche springt von
	//     Fund zu Fund durch die Karten; hier ist die Suche eine **Seite mit
	//     einer Trefferliste**, und „der nächste" ist die Zeile darunter. Ein
	//     Menüeintrag dafür hätte kein Gegenstück.
	//   - **Neu laden** (`ID_FILE_RELOAD`). Dort ist die Datei die Wahrheit
	//     und der Speicher eine Ansicht; hier ist es umgekehrt (0002). Ein
	//     „neu laden" würfe alles weg, was seit dem Einfahren geschrieben
	//     wurde — MSE fragt vorher (`askSaveAndContinue`), und die Frage
	//     hätte hier eine andere Bedeutung als dort.
	//
	//     **Was der Grund nicht sagt**: MSE lädt dabei nicht nur die
	//     Setdatei, sondern auch den Paket- und Einstellungsstand
	//     (`window.cpp`, Zeilen 715–735). Angestrichen als Befund 7 des
	//     Nachlaufs über Phase 3.7; die Anpassung trägt, die Beschreibung
	//     war unvollständig.
	//
	// **Und zwei Gründe standen hier, die nicht stimmten** — ebenfalls
	// Befund 7 des Nachlaufs. Sie sind jetzt Umfang und keine Begründung
	// (#1242):
	//
	//   - **Alles auswählen** (`ID_EDIT_SELECT_ALL`) ist bei MSE2 **nicht**
	//     nur Sache der Textmarke. `CardsPanel::doSelectAll` leitet je nach
	//     Fokus weiter, und `ItemList::doSelectAll` wählt dort **alle
	//     Karten** (`cards_panel.cpp`, Zeilen 404–410; `item_list.cpp`,
	//     Zeilen 43–64). Dieses Projekt hat mit den Anhakhaken ein
	//     fachliches Gegenstück — **und seit dem 2026-08-28 den Eintrag
	//     dazu**: er steht im Zweig *Set*, weil er das Set meint und nicht
	//     einen Text.
	//   - **Autoersetzung** (`ID_EDIT_AUTO_REPLACE`) ist nicht bloß eine
	//     Fläche neben dem Menü: der Eintrag **öffnet** genau die Verwaltung
	//     mit Aktivierung, Liste, Hinzufügen, Entfernen und Wortgrenze
	//     (`window.cpp`, Zeilen 797–799; `auto_replace_window.cpp`,
	//     Zeilen 158–180). Dass sie fehlt, ist eine Folge der fehlenden
	//     Verwaltung und kein eigener Grund.
	setzweig := leiste.Untermenue(texte.Set)
	setzweig.Befehl(texte.Rueckgaengig, nil, rueckruf(handlungen.Rueckgaengig))
	setzweig.Befehl(texte.Wiederherstellen, nil, rueckruf(handlungen.Wiederherstellen))
	setzweig.Trennlinie()
	// **Alles auswählen hakt jede Karte der Liste an** (#1242).
	//
	// **Im Zweig *Set* und nicht bei *Bearbeiten***: es meint das Set und
	// nicht einen Text. macOS hängt ⌘A dort an das Textfeld, das gerade den
	// Fokus hat; dieser Eintrag ist der zweite Weg zu derselben Handlung, wie
	// jeder andere hier.
	setzweig.Befehl(texte.AllesAuswaehlen, nil, rueckruf(handlungen.AllesAuswaehlen))
	setzweig.Trennlinie()
	setzweig.Befehl(texte.Suchen, nil, seitenruf(handlungen.SeiteOeffnen, "suche"))
	setzweig.Befehl(texte.Ersetzen, nil, seitenruf(handlungen.SeiteOeffnen, "ersetzen"))

	fenster := leiste.Rollenuntermenue(RolleFenstermenue, texte.Fenster)
	fenster.Rolleneintrag(RolleMinimieren, texte.Minimieren)
	fenster.Rolleneintrag(RolleZoomen, texte.Zoomen)
	fenster.Trennlinie()
	fenster.Rolleneintrag(RolleAlleNachVorne, texte.AlleNachVorne)

	// **Die Hilfe steht rechts außen** (§6). macOS erwartet sie dort; wer sie
	// da nicht findet, sucht sie nirgends sonst.
	//
	// **Ohne Kürzel in der Leiste.** ⌘? erreicht die Seite über die Tastatur
	// des Fensters (`navigation.ts`); ein zweites Kürzel hier zeigte macOS als
	// `⌘⇧/`, und das steht auf keiner Taste.
	hilfe := leiste.Untermenue(texte.Hilfe)
	hilfe.Befehl(texte.HilfeOeffnen, nil, rueckruf(handlungen.HilfeOeffnen))

	return leiste
}

// rueckruf macht aus einer möglicherweise fehlenden Handlung einen Rückruf.
//
// Eine nicht gesetzte Handlung ist ein Programmierfehler — aber einer, der sich
// beim Klick zeigen würde, also zur Laufzeit und bei den Nutzenden. Ein Menü,
// das nichts tut, ist ärgerlich; eines, das die Anwendung beendet, ist schlimmer.
func rueckruf(handlung func()) func() {
	return func() {
		if handlung == nil {
			return
		}

		handlung()
	}
}

// seitenruf bindet einen Seitenschlüssel an den Rückruf.
//
// **Der Schlüssel steht hier und wird nicht erfunden.** Was ein Menüeintrag
// mitgibt, muss `frontend/src/navigation.ts` kennen — sonst führte er
// nirgendwohin. Ein Wächter hält beide Seiten zusammen
// (`internal/werkzeuge/menueseiten_test.go`).
func seitenruf(handlung func(string), seite string) func() {
	return func() {
		if handlung == nil {
			return
		}

		handlung(seite)
	}
}

// zuletztruf bindet eine Setkennung an den Rückruf.
//
// **Die Kennung wird gefangen, nicht gelesen.** Ein Rückruf, der sie aus der
// Schleifenvariablen läse, bekäme bei jedem Klick die letzte — der Fehler, den
// Go seit 1.22 nicht mehr hat und den ein Leser trotzdem sucht.
func zuletztruf(handlung func(string), setID string) func() {
	return func() {
		if handlung == nil {
			return
		}

		handlung(setID)
	}
}
