package app

// Die Beschriftungen der nativen Menüleiste.
//
// # Warum ein zweiter Katalog
//
// Die Oberflächentexte liegen unter `frontend/src/sprache/` (§9). Die
// Menüleiste ist **nativ**: sie entsteht in Go, bevor das Fenster steht, und
// kein React sieht sie je. Den Katalog von dort zu holen hieße, den Start der
// Anwendung von der Oberfläche abhängig zu machen — für vier Wörter.
//
// Also ein eigener, kleiner Katalog hier. Der Preis ist, dass zwei Kataloge
// nebeneinander bestehen; bezahlt wird er durch einen Wächter, der beide
// Sprachen vollständig hält, und durch einen zweiten, der eine Beschriftung im
// Quelltext von `menue.go` abweist.

// Menuetexte sind die Beschriftungen einer Menüleiste.
//
// **Als Struktur und nicht als Abbildung.** Ein fehlender Schlüssel in einer
// Abbildung ist eine leere Zeichenkette und damit ein Menüeintrag ohne
// Beschriftung; ein fehlendes Feld in einer Struktur ist ein Übersetzungsfehler.
type Menuetexte struct {
	Ablage         string
	BestandSichern string
	Sichern        string
	SichernUnter   string
	Einstellungen  string

	// Neu, Öffnen, Schließen und Drucken kamen mit #891 dazu.
	Neu        string
	Oeffnen    string
	Schliessen string
	Drucken    string

	// Zuletzt ist das Untermenü mit den jüngsten Sets (#891).
	Zuletzt string

	// **Das Setmenü und was in der Ablage dazukam** — Werkstück 4 von
	// Phase 3.7
	// (#1168,
	// Befund Z4-01).
	Set          string
	Rueckgaengig string

	// AllesAuswaehlen ist der Eintrag, der jede Karte der Liste anhakt
	// (#1242).
	AllesAuswaehlen  string
	Wiederherstellen string
	Suchen           string
	Ersetzen         string

	// Ausgeben ist das Untermenü der Ablage — MSEs `export`.
	Ausgeben      string
	Exportvorlage string
	Bilder        string
	Druckbogen    string

	// Hilfe ist das Menü, HilfeOeffnen sein Eintrag. Ueber steht seit #255
	// im Anwendungsmenü.
	Hilfe        string
	HilfeOeffnen string
	Ueber        string

	// Darstellung und Seitenleiste — Werkstück U14 von Phase 5.1 (#1490).
	//
	// **Je Teil zwei Wörter, nach dem Stand** — UXR-F14
	// (#2029):
	// „einblenden", wo es zu ist, „ausblenden", wo es steht (Apple, Human
	// Interface Guidelines, „The menu bar"). Bis dahin stand für die
	// Seitenleiste ein Eintrag „ein-/ausblenden", und Navigator und Inspector
	// standen in keinem Menü. Die Wörter sind die der Knöpfe im Fenster
	// (`sprache/*/allgemein.ts`, `bausteine.ts`), damit dieselbe Handlung
	// überall gleich heißt.
	Darstellung            string
	SeitenleisteEinblenden string
	SeitenleisteAusblenden string
	NavigatorEinblenden    string
	NavigatorAusblenden    string
	InspectorEinblenden    string
	InspectorAusblenden    string

	// DarstellungZuruecksetzen stellt das Aussehen zurück — UX-09.AK-03
	// (#1681): der Weg, der lesbar bleibt, wenn das Fenster es nicht mehr ist.
	DarstellungZuruecksetzen string

	// **Zurück und Vorwärts** — UX-03.AK-02 (#1675). Die Wörter, mit denen
	// die Systemeinstellungen „Back" und „Forward" übersetzen
	// (`Localizable.loctable` im Programmbündel, gelesen am 2026-09-24).
	SeiteZurueck string
	SeiteVor     string

	// **Die Menüs, die macOS erwartet** — seit #255 in der Sprache des
	// Fensters. Die Wörter sind die von macOS selbst, damit dieselbe Handlung
	// in jeder Anwendung gleich heißt; „Widerrufen" und nicht „Rückgängig",
	// denn das steht im Setmenü für den fachlichen Verlauf.
	Anwendung        string
	Dienste          string
	Ausblenden       string
	AndereAusblenden string
	AlleEinblenden   string
	Beenden          string

	Bearbeiten   string
	Widerrufen   string
	Wiederholen  string
	Ausschneiden string
	Kopieren     string
	Einsetzen    string
	Loeschen     string

	// AllesMarkieren ist „Alles auswählen" im Bearbeitenmenü — die Auswahl
	// im Text. [AllesAuswaehlen] hakt dagegen Karten an.
	AllesMarkieren  string
	Sprachausgabe   string
	SprechenStarten string
	SprechenStoppen string

	Fenster       string
	Minimieren    string
	Zoomen        string
	AlleNachVorne string
}

// Sprachen sind die Sprachen, die die Anwendung führt.
//
// **Deutsch zuerst**: es ist die Ausgangssprache (§9) und der Rückfall.
// **Seit LOC-01** (#1687) sechs weitere, maschinell übersetzt — ihr Stand
// steht in `docs/SPRACHINVENTAR.md` und am Fenster neben der Wahl.
var Sprachen = []string{"de", "en", "es", "fr", "it", "ja", "ru", "zh"}

var menuekataloge = map[string]Menuetexte{
	"de": {
		Ablage:         "Ablage",
		Neu:            "Neu",
		Oeffnen:        "Öffnen …",
		Sichern:        "Sichern",
		SichernUnter:   "Sichern unter …",
		Schliessen:     "Schließen",
		Drucken:        "Drucken …",
		BestandSichern: "Bestand sichern …",
		Einstellungen:  "Einstellungen …",
		Zuletzt:        "Zuletzt benutzt",
		Hilfe:          "Hilfe",
		HilfeOeffnen:   "MagicSetEditor4Mac-Hilfe",

		Set:              "Set",
		Rueckgaengig:     "Rückgängig",
		AllesAuswaehlen:  "Alle Karten anhaken",
		Wiederherstellen: "Wiederherstellen",
		Suchen:           "Suchen …",
		Ersetzen:         "Ersetzen …",

		Ausgeben:      "Ausgeben",
		Exportvorlage: "Über eine Exportvorlage …",
		Bilder:        "Als Bilder …",
		Druckbogen:    "Druckbogen …",

		Ueber: "Über MagicSetEditor4Mac",

		Darstellung:            "Darstellung",
		SeitenleisteEinblenden: "Seitenleiste einblenden",
		SeitenleisteAusblenden: "Seitenleiste ausblenden",
		NavigatorEinblenden:    "Navigator einblenden",
		NavigatorAusblenden:    "Navigator ausblenden",
		InspectorEinblenden:    "Inspector einblenden",
		InspectorAusblenden:    "Inspector ausblenden",
		SeiteZurueck:           "Zurück",
		SeiteVor:               "Vorwärts",

		DarstellungZuruecksetzen: "Darstellung zurücksetzen",

		Anwendung:        "MagicSetEditor4Mac",
		Dienste:          "Dienste",
		Ausblenden:       "MagicSetEditor4Mac ausblenden",
		AndereAusblenden: "Andere ausblenden",
		AlleEinblenden:   "Alle einblenden",
		Beenden:          "MagicSetEditor4Mac beenden",

		Bearbeiten:      "Bearbeiten",
		Widerrufen:      "Widerrufen",
		Wiederholen:     "Wiederholen",
		Ausschneiden:    "Ausschneiden",
		Kopieren:        "Kopieren",
		Einsetzen:       "Einsetzen",
		Loeschen:        "Löschen",
		AllesMarkieren:  "Alles auswählen",
		Sprachausgabe:   "Sprachausgabe",
		SprechenStarten: "Sprachausgabe starten",
		SprechenStoppen: "Sprachausgabe stoppen",

		Fenster:       "Fenster",
		Minimieren:    "Im Dock ablegen",
		Zoomen:        "Zoomen",
		AlleNachVorne: "Alle nach vorne bringen",
	},
	"en": {
		Ablage:         "File",
		Neu:            "New",
		Oeffnen:        "Open …",
		Sichern:        "Save",
		SichernUnter:   "Save As …",
		Schliessen:     "Close",
		Drucken:        "Print …",
		BestandSichern: "Back Up Database …",
		Einstellungen:  "Settings …",
		Zuletzt:        "Open Recent",
		Hilfe:          "Help",
		HilfeOeffnen:   "MagicSetEditor4Mac Help",

		Set:              "Set",
		Rueckgaengig:     "Undo",
		AllesAuswaehlen:  "Check all cards",
		Wiederherstellen: "Redo",
		Suchen:           "Find …",
		Ersetzen:         "Replace …",

		Ausgeben:      "Export",
		Exportvorlage: "Using an Export Template …",
		Bilder:        "As Images …",
		Druckbogen:    "Print Sheet …",

		Ueber: "About MagicSetEditor4Mac",

		Darstellung:            "View",
		SeitenleisteEinblenden: "Show Sidebar",
		SeitenleisteAusblenden: "Hide Sidebar",
		NavigatorEinblenden:    "Show Navigator",
		NavigatorAusblenden:    "Hide Navigator",
		InspectorEinblenden:    "Show Inspector",
		InspectorAusblenden:    "Hide Inspector",
		SeiteZurueck:           "Back",
		SeiteVor:               "Forward",

		DarstellungZuruecksetzen: "Reset Appearance",

		Anwendung:        "MagicSetEditor4Mac",
		Dienste:          "Services",
		Ausblenden:       "Hide MagicSetEditor4Mac",
		AndereAusblenden: "Hide Others",
		AlleEinblenden:   "Show All",
		Beenden:          "Quit MagicSetEditor4Mac",

		Bearbeiten:      "Edit",
		Widerrufen:      "Undo",
		Wiederholen:     "Redo",
		Ausschneiden:    "Cut",
		Kopieren:        "Copy",
		Einsetzen:       "Paste",
		Loeschen:        "Delete",
		AllesMarkieren:  "Select All",
		Sprachausgabe:   "Speech",
		SprechenStarten: "Start Speaking",
		SprechenStoppen: "Stop Speaking",

		Fenster:       "Window",
		Minimieren:    "Minimize",
		Zoomen:        "Zoom",
		AlleNachVorne: "Bring All to Front",
	},
	"es": menuetexteEs,
	"fr": menuetexteFr,
	"it": menuetexteIt,
	"ja": menuetexteJa,
	"ru": menuetexteRu,
	"zh": menuetexteZh,
}

// MenuetexteFuer gibt den Katalog einer Sprache — oder den deutschen.
//
// **Ohne Fehler.** Eine unbekannte Sprache ist kein Grund, keine Menüleiste zu
// bauen; sie ist ein Grund, die Ausgangssprache zu nehmen. Wer eine Sprache
// setzt, bekommt den Befund an der Stelle, an der er sie setzt.
func MenuetexteFuer(sprache string) Menuetexte {
	if texte, da := menuekataloge[sprache]; da {
		return texte
	}

	return menuekataloge["de"]
}

// IstSprache sagt, ob die Anwendung diese Sprache führt.
func IstSprache(sprache string) bool {
	_, da := menuekataloge[sprache]

	return da
}
