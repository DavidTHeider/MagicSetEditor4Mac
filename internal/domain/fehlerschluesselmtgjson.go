package domain

// **Eine MTG-JSON-Datei, die sich nicht lesen lässt** — G01
// (#2248).
//
// Eigene Datei aus demselben Grund wie `fehlerschluesselsicherung.go`. **Zwei
// und nicht einer**: kaputtes JSON und JSON ohne Set verlangen verschiedene
// Sätze — das eine ist abgeschnitten oder falsch geformt, das andere eine
// ganz andere Datei. Beide tragen `datei` und `stelle` als Werte.
const (
	// SchluesselMTGJSONKeinJSON: die Datei ist kein gültiges JSON.
	SchluesselMTGJSONKeinJSON Schluessel = "mtgjsonKeinJSON"

	// SchluesselMTGJSONKeinSet: die Datei ist JSON, aber kein Set.
	SchluesselMTGJSONKeinSet Schluessel = "mtgjsonKeinSet"
)
