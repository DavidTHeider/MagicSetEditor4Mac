package domain

// **Welche Vorschau für welche Datei sicher ist** — UX-13.AK-01
// (#1685),
// Marke `new`: MSE2 zeigt Dateien nicht an, es öffnet sie.
//
// # Die Regel
//
// **Angesehen wird, was sich ansehen lässt, ohne dass etwas ausgeführt wird.**
// Ein Bild und ein SVG erscheinen als Bild — im Fenster über `<img>`, wo
// WebKit weder Skripte ausführt noch fremde Adressen lädt. Text erscheint als
// Text, HTML eingeschlossen. Ein Paket oder Archiv erscheint als Liste seiner
// Einträge — **ausgepackt wird nichts**. Ein PDF nennt seine Seiten.
//
// **Und eine unbekannte Datei zeigt nur, was über sie bekannt ist** — Name,
// Größe, Zeit. Ein `.command`, ein `.app` oder eine Datei ohne Endung kann
// ein Programm sein; es anzusehen hieße, es zu raten, und es zu öffnen,
// hieße, es zu starten.
//
// # Warum eine Grenze je Art
//
// **Weil eine Vorschau keine Kopie ist** (AK-03): gelesen wird höchstens so
// viel, wie die Ansicht braucht. Wer ein Gigabyte ansieht, sieht Name, Größe
// und dass es zu groß ist.
//
// # Und warum die Paketarten hier ein zweites Mal stehen
//
// **Weil `domain` nichts aus dem Projekt kennt** (§2) und `msepaket` sie
// führt. Ein Wächter hält beide Listen gleich
// (`internal/werkzeuge/dateiart_test.go`).

import (
	"path/filepath"
	"strings"
)

// Dateiart ist die Art einer Datei, nach der sich ihre Ansicht richtet.
type Dateiart string

// Die Dateiarten.
const (
	DateiartPaket     Dateiart = "paket"
	DateiartBild      Dateiart = "bild"
	DateiartVektor    Dateiart = "vektor"
	DateiartPDF       Dateiart = "pdf"
	DateiartText      Dateiart = "text"
	DateiartArchiv    Dateiart = "archiv"
	DateiartUnbekannt Dateiart = "unbekannt"
)

// Vorschauart sagt, wie eine Datei erscheint.
type Vorschauart string

// Die Vorschauarten.
const (
	VorschauBild  Vorschauart = "bild"
	VorschauText  Vorschauart = "text"
	VorschauListe Vorschauart = "liste"
	VorschauKeine Vorschauart = "keine"
)

// Ansichtsregel sagt, was für eine Datei sicher ist.
type Ansichtsregel struct {
	Art      Dateiart
	Vorschau Vorschauart

	// Hoechstens ist, was für die Vorschau gelesen wird — Bytes, bei einer
	// Liste die Zahl der Einträge.
	Hoechstens int64

	// Oeffnen sagt, ob das Fenster die Datei dem Standardprogramm des
	// Systems übergeben darf.
	Oeffnen bool
}

// Paketendungen sind die sieben MSE-Paketarten.
func Paketendungen() []string {
	return []string{
		"mse-set", "mse-game", "mse-style", "mse-symbol-font",
		"mse-include", "mse-locale", "mse-export-template",
	}
}

// endungen ordnet die bekannten Endungen ihrer Art zu.
var endungen = func() map[string]Dateiart {
	arten := map[string]Dateiart{}

	for _, endung := range Paketendungen() {
		arten[endung] = DateiartPaket
	}

	for art, liste := range map[Dateiart][]string{
		DateiartBild:   {"png", "jpg", "jpeg", "gif", "webp", "bmp", "tif", "tiff"},
		DateiartVektor: {"svg"},
		DateiartPDF:    {"pdf"},
		DateiartText:   {"txt", "json", "csv", "tsv", "md", "dec", "dck", "mwdeck", "xml", "log", "html", "htm"},
		DateiartArchiv: {"zip", "msesicherung", "msepackage"},
	} {
		for _, endung := range liste {
			arten[endung] = art
		}
	}

	return arten
}()

// ohneOeffnen sind Endungen einer sonst öffenbaren Art, die das Fenster
// trotzdem nicht übergibt — ein Browser führte ihre Skripte aus.
var ohneOeffnen = map[string]bool{"html": true, "htm": true, "xml": true}

// AnsichtsregelFuer gibt die Regel für einen Dateinamen.
func AnsichtsregelFuer(name string) Ansichtsregel {
	endung := strings.ToLower(strings.TrimPrefix(filepath.Ext(name), "."))
	art, bekannt := endungen[endung]

	if !bekannt {
		return Ansichtsregel{Art: DateiartUnbekannt, Vorschau: VorschauKeine}
	}

	regel := map[Dateiart]Ansichtsregel{
		DateiartPaket:  {Vorschau: VorschauListe, Hoechstens: 2000},
		DateiartBild:   {Vorschau: VorschauBild, Hoechstens: 20 << 20, Oeffnen: true},
		DateiartVektor: {Vorschau: VorschauBild, Hoechstens: 1 << 20, Oeffnen: true},
		DateiartPDF:    {Vorschau: VorschauKeine, Hoechstens: 20 << 20, Oeffnen: true},
		DateiartText:   {Vorschau: VorschauText, Hoechstens: 256 << 10, Oeffnen: !ohneOeffnen[endung]},
		DateiartArchiv: {Vorschau: VorschauListe, Hoechstens: 2000},
	}[art]
	regel.Art = art

	return regel
}
