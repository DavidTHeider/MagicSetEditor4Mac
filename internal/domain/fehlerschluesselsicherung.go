package domain

// **Die Befunde am Kennwort einer Sicherung** — OPS-06.AK-02
// (#1723).
//
// Eigene Datei, weil `fehlerschluessel.go` mit ihnen an die Zeilengrenze aus
// §2 ginge. **Vier und nicht einer**: „zu kurz", „fehlt", „passt nicht" und
// „der Bestand ist verschlüsselt" verlangen vier verschiedene Handlungen, und
// der allgemeine Satz zur Validierung nennt keine davon.
const (
	// SchluesselSicherungskennwortKurz: unter zwölf Zeichen.
	SchluesselSicherungskennwortKurz Schluessel = "sicherungskennwortKurz"

	// SchluesselSicherungskennwortFehlt: die Sicherung ist verschlüsselt und
	// kam ohne Kennwort.
	SchluesselSicherungskennwortFehlt Schluessel = "sicherungskennwortFehlt"

	// SchluesselSicherungskennwortFalsch: das Kennwort öffnet sie nicht.
	SchluesselSicherungskennwortFalsch Schluessel = "sicherungskennwortFalsch"

	// SchluesselSicherungImKlartext: der Bestand ist verschlüsselt, und die
	// Sicherung käme ohne Kennwort und ohne ausdrückliches „im Klartext".
	SchluesselSicherungImKlartext Schluessel = "sicherungImKlartext"

	// SchluesselSicherungUnvermerkt gehört zu [ErrSicherungUnvermerkt]: die
	// Sicherung liegt geprüft da, nur ihr Zeitpunkt ist nicht vermerkt
	// (OPS01, #2347).
	SchluesselSicherungUnvermerkt Schluessel = "sicherungUnvermerkt"

	// SchluesselPlanVerschluesselt: der Bestand ist verschlüsselt, und der
	// Plan sichert ohne Kennwort — er sichert ihn deshalb nicht (Zweitblick
	// über Phase 4, Befund 4).
	SchluesselPlanVerschluesselt Schluessel = "planVerschluesselt"
)
