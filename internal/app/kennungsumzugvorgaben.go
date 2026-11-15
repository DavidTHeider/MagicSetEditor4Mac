package app

// **Die Vorgaben beim Umzug der Kennung** — R02
// (#2334),
// #2424, und der Zweitblick über Phase 4, Befunde 5 und 6.

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/dateisperre"
)

// vorgabenfrist ist, wie lange ein Umzug auf einen anderen wartet, der
// gerade dieselben Vorgaben zieht.
const vorgabenfrist = 30 * time.Second

// vorgabenZiehen gibt die Vorgaben der alten Kennung an die neue — **über
// cfprefsd und nach dem Namen der Domäne**, nicht über eine Datei. Der
// Gastlauf (R02) fand eine Vorgabe, die cfprefsd schon führte und die auf der
// Platte noch fehlte, und auf APFS ohne Unterscheidung der Schreibweise muss
// der Name der Datei nicht der der Domäne sein. Gefragt wird deshalb, ob es
// die neue Domäne schon gibt, und sonst je Schreibweise der alten, ob es sie
// gibt; die erste, die es gibt, zieht um. **Ein gescheitertes Fragen heißt:
// nicht da** — in einem Testlauf ruft niemand `defaults`.
//
// **Umgezogen heißt: die Marke steht da** ([umzugsmarke]) — und nicht: die
// neue Domäne gibt es. **Und gibt es sie ohne Marke, wird vereint**: ihre
// Schlüssel bleiben, aus der alten kommt nur, was fehlt (#2424,
// `vorgabenvereinen.go`). Der zweite Gastlauf fand sie leer vor, denn
// `defaults delete` lässt eine leere Domäne zurück, und der Umzug hielt sie
// für fertig.
//
// **Ein Umzug zur Zeit, jeder in seinem eigenen Ordner** (Zweitblick über
// Phase 4, Befund 5): zwei Anwendungen, die zugleich starten, teilten sich
// drei Namen unter `Caches/<Kennung>` — einer konnte die Zusammenführung des
// anderen importieren oder dessen Datei vor dem Import löschen. Lesen,
// Vereinen und Importieren laufen jetzt unter einer Sperre, und **unter ihr
// wird die Marke noch einmal gefragt**: wer eben umzog, hat sie gesetzt.
func (u kennungsumzug) vorgabenZiehen(ctx context.Context) error {
	if u.ausfuehren(ctx, []string{"defaults", "read", Kennung, umzugsmarke}) == nil {
		return nil
	}

	alt := ""

	for _, kennung := range altkennungen {
		if u.ausfuehren(ctx, []string{"defaults", "read", kennung}) == nil {
			alt = kennung

			break
		}
	}

	if alt == "" {
		return nil
	}

	// **Im eigenen Zwischenspeicher und nicht in TMPDIR** (#2424): dort
	// teilten sich zwei Läufe desselben Prozesses einen Namen.
	ablage := vorgabenablage(u.heim)
	if fehler := os.MkdirAll(ablage, 0o700); fehler != nil {
		return fmt.Errorf("die alten Vorgaben lesen: %w", fehler)
	}

	griff, fehler := dateisperre.Oeffnen(filepath.Join(ablage, "vorgaben.sperre"))
	if fehler != nil {
		return fmt.Errorf("die alten Vorgaben lesen: %w", fehler)
	}

	defer func() { _ = griff.Schliessen() }() //nolint:errcheck // der Anspruch fällt mit dem Griff

	if gewartet := griff.AlleinBis(ctx, vorgabenfrist, 50*time.Millisecond); gewartet != nil {
		return fmt.Errorf("auf einen anderen Umzug der Vorgaben warten: %w", gewartet)
	}

	if u.ausfuehren(ctx, []string{"defaults", "read", Kennung, umzugsmarke}) == nil {
		return nil
	}

	lauf, fehler := os.MkdirTemp(ablage, "umzug-")
	if fehler != nil {
		return fmt.Errorf("die alten Vorgaben lesen: %w", fehler)
	}

	defer func() { _ = os.RemoveAll(lauf) }() //nolint:errcheck // ein Arbeitsordner, den niemand sonst kennt

	return u.vorgabenVereinenUndImportieren(ctx, alt, lauf)
}

// vorgabenablage ist der Zwischenspeicher des Umzugs — Sperre und je Lauf
// ein Arbeitsordner. **Die Außerbetriebnahme nennt ihn** (Zweitblick über
// Phase 4, Befund 13); ihr Wächter liest ihn hier.
func vorgabenablage(heim string) string {
	return filepath.Join(heim, "Library", "Caches", Kennung)
}

// vorgabenVereinenUndImportieren liest beide Domänen nach `lauf`, vereint sie
// und importiert das Ergebnis samt Marke.
func (u kennungsumzug) vorgabenVereinenUndImportieren(ctx context.Context, alt, lauf string) error {
	altroh, fehler := u.vorgabenAusfuehren(ctx, alt, filepath.Join(lauf, "vorgaben-alt.plist"))
	if fehler != nil {
		return fmt.Errorf("die alten Vorgaben lesen: %w", fehler)
	}

	neuroh, fehler := u.neueVorgaben(ctx, filepath.Join(lauf, "vorgaben-neu.plist"))
	if fehler != nil {
		return fmt.Errorf("die neuen Vorgaben lesen: %w", fehler)
	}

	// **Die Marke kommt mit dem Import und nicht danach** (Zweitblick über
	// Phase 3, Befund 9): ein eigenes `defaults write` konnte scheitern; der
	// Import bringt Werte und Marke in einem Zug.
	vereint, fehler := vorgabenVereinen(altroh, neuroh, alt)
	if fehler != nil {
		return fmt.Errorf("den Umzug der Vorgaben vermerken: %w", fehler)
	}

	vereinigt := filepath.Join(lauf, "vorgaben.plist")
	if fehler := os.WriteFile(vereinigt, vereint, 0o600); fehler != nil {
		return fmt.Errorf("den Umzug der Vorgaben vermerken: %w", fehler)
	}

	if fehler := u.ausfuehren(ctx, []string{"defaults", "import", Kennung, vereinigt}); fehler != nil {
		return fmt.Errorf("die Vorgaben übernehmen: %w", fehler)
	}

	return nil
}

// neueVorgaben liest, was unter der neuen Kennung schon steht.
//
// **Was die neue Fassung schon schrieb, bleibt** (#2424): scheiterte ein
// erster Umzug vor dem Import, gibt es die neue Domäne ohne Marke, und
// `defaults import` ersetzte sie ganz.
//
// **Und „leer" muss belegt sein** (Zweitblick über Phase 4, Befund 6, und
// sein Nachlauf, Befund 5): leer ist die Domäne nur, wenn `defaults read`
// scheitert **und dabei selbst sagt, dass es sie nicht gibt** — gemessen am
// 2026-10-16 unter macOS 27.0.1 `Domain '…' not found.`, ältere Fassungen
// schreiben `… does not exist`. Eine fehlende Datei unter
// `~/Library/Preferences` belegt nichts: cfprefsd führt Werte schon vor ihr
// (R02). **Jeder andere Fehler** — beim Lesen, Ausführen, Umwandeln — bricht
// den Umzug ab, statt vorhandene Werte durch alte zu ersetzen.
func (u kennungsumzug) neueVorgaben(ctx context.Context, datei string) ([]byte, error) {
	if gelesen := u.ausfuehren(ctx, []string{"defaults", "read", Kennung}); gelesen != nil {
		if domaeneFehlt(gelesen) {
			return []byte(leereVorgaben), nil
		}

		return nil, gelesen
	}

	return u.vorgabenAusfuehren(ctx, Kennung, datei)
}

// domaeneFehlt sagt, ob ein gescheitertes `defaults read` die Abwesenheit
// der Domäne meldet — gelesen in der Ausgabe, die im Grund des Fehlers reist
// ([befehlAusfuehren]).
func domaeneFehlt(fehler error) bool {
	for _, text := range fehlertexte(fehler) {
		if strings.Contains(text, "not found") || strings.Contains(text, "does not exist") {
			return true
		}
	}

	return false
}

// fehlertexte gibt die Texte eines Fehlers und aller, die er umschließt.
func fehlertexte(fehler error) []string {
	if fehler == nil {
		return nil
	}

	texte := []string{fehler.Error()}

	switch umschlossen := fehler.(type) { //nolint:errorlint // gelesen wird genau eine Ebene, rekursiv
	case interface{ Unwrap() []error }:
		for _, innen := range umschlossen.Unwrap() {
			texte = append(texte, fehlertexte(innen)...)
		}
	case interface{ Unwrap() error }:
		texte = append(texte, fehlertexte(umschlossen.Unwrap())...)
	}

	return texte
}

// leereVorgaben ist eine Domäne ohne Schlüssel.
const leereVorgaben = `<?xml version="1.0" encoding="UTF-8"?>
<plist version="1.0"><dict/></plist>
`

// vorgabenAusfuehren schreibt eine Domäne als XML-Plist nach `datei` und gibt
// sie — `defaults export` schreibt sie binär.
func (u kennungsumzug) vorgabenAusfuehren(ctx context.Context, domaene, datei string) ([]byte, error) {
	if fehler := u.ausfuehren(ctx, []string{"defaults", "export", domaene, datei}); fehler != nil {
		return nil, fehler
	}

	if fehler := u.ausfuehren(ctx, []string{"plutil", "-convert", "xml1", datei}); fehler != nil {
		return nil, fehler
	}

	return plistLesen(filepath.Dir(datei), filepath.Base(datei))
}
