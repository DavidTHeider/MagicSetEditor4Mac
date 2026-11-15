// **Ordner, in denen Dialoge aufgehen** — OPS-02
// (#1719),
// Marke `new`: MSE2 merkt sich den letzten Ordner eines Dialogs, eine
// Einstellung dafür kennt es nicht.
//
// # Die Spezifikation (AK-01)
//
// **Absolut, oder mit `~/` vorn — sonst nichts.** Ein relativer Pfad wäre
// relativ zu einem Arbeitsverzeichnis, und das ist beim Fenster ein anderes
// als an der Kommandozeile oder im Dienst; derselbe Wert zeigte auf drei
// Flächen auf drei Orte. `~` wird zum Zuhause dessen, der die Anwendung
// ausführt, und nur vorn: `~anna` und `$HOME` bleiben, was sie sind, und
// werden abgewiesen — eine Dehnung, die eine Shell macht und diese Anwendung
// nicht, sähe aus wie ein Fehler der Anwendung. Ein `..` verschleiert das
// Ziel und wird ebenfalls abgewiesen.
//
// # Was hier nicht geprüft wird
//
// **Ob es den Ordner gibt.** Das ist eine Frage an die Platte und nicht an
// den Wert; sie stellt `app` beim Setzen und beim Benutzen (AK-01, AK-03).

package domain

import (
	"path/filepath"
	"slices"
	"strings"
)

// Die drei Ordner, die mit OPS-02 dazukamen — der vierte, der Ordner fürs
// Hereinholen, steht als [VorgabeHochladepfad] seit Phase 4.12.
const (
	// VorgabeAusfuhrpfad ist der Ordner, in dem ein Sicherndialog aufgeht,
	// wenn die Handlung keinen eigenen Ort hat.
	VorgabeAusfuhrpfad = "vorgabe.ausfuhrpfad"

	// VorgabeSicherungspfad ist der Ordner für Sicherungen des Bestands.
	VorgabeSicherungspfad = "vorgabe.sicherungspfad"

	// VorgabeDiagnosepfad ist der Ordner für Diagnosetext und -paket.
	VorgabeDiagnosepfad = "vorgabe.diagnosepfad"
)

// Ordnerschluessel gibt die Einstellungen, deren Wert ein Ordner ist.
func Ordnerschluessel() []string {
	return []string{VorgabeHochladepfad, VorgabeAusfuhrpfad, VorgabeSicherungspfad, VorgabeDiagnosepfad}
}

// IstOrdnerangabe sagt, ob der Wert eines Schlüssels ein Ordner ist — die
// vier der Dialoge, der Ort des Protokolls (OPS-04, #1721), der Ordner des
// Plans (OPS01, #2347) und der zuletzt gewählte der Bildausgabe (AP-10).
func IstOrdnerangabe(schluessel string) bool {
	return slices.Contains(Ordnerschluessel(), schluessel) ||
		schluessel == ProtokollOrt || schluessel == SicherungsplanOrdner ||
		schluessel == EinstellungBildordner
}

// PruefeOrdnerangabe urteilt über den Wert — leer heißt: wo das
// Betriebssystem den Dialog öffnet.
func PruefeOrdnerangabe(wert string) error {
	if wert == "" {
		return nil
	}

	if strings.ContainsRune(wert, 0) {
		return NewFieldError("wert", SchluesselValidierung,
			"Diese Angabe darf kein Nullbyte enthalten.")
	}

	vollstaendig := wert == "~" || strings.HasPrefix(wert, "~/") || filepath.IsAbs(wert)
	if !vollstaendig || slices.Contains(strings.Split(wert, "/"), "..") {
		return NewFieldError("wert", SchluesselOrdnerangabe,
			"»%s« ist kein vollständiger Ort. Trage einen ein, der mit / oder mit ~/ beginnt, ohne .. dazwischen.",
			wert).MitWerten(map[string]string{"wert": wert})
	}

	return nil
}

// OrdnerAus dehnt `~` zum Zuhause und glättet den Pfad — leer bleibt leer.
func OrdnerAus(angabe, zuhause string) string {
	switch {
	case angabe == "":
		return ""
	case angabe == "~":
		return filepath.Clean(zuhause)
	case strings.HasPrefix(angabe, "~/"):
		return filepath.Join(zuhause, angabe[2:])
	default:
		return filepath.Clean(angabe)
	}
}
