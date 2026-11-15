// Welche Schlüssel eines Setkopfs in den Bestand übernommen werden.
//
// **Eine eigene Datei**, weil `import.go` sonst über 400 Zeilen ginge (§2).

package service

// gespeichert sind die Schlüssel, die wir in den Bestand übernehmen.
//
// Alles andere wandert als Rohblock mit — **auch das, was wir lesen könnten**.
// `mse version` stand hier zuerst und war damit weder gespeichert noch roh
// mitgeführt: als „verstanden" markiert und danach verworfen. Der Zweitblick
// über Phase 2 hat es gefunden. Die Liste heißt deshalb nicht mehr „gedeutet":
// sie sagt, was **abgelegt** wird, und was hier fehlt, reist unverändert mit.
//
// # Die vier Klassen aus #88 sind geschlossen
//
// Ein Schlüssel aus dieser Liste schloss bis 2026-08-15 seinen ganzen Bereich
// vom Rohweg aus — auch das Vorkommen, das der Leser gar nicht verwertet. Vier
// Fälle standen hier aufgezählt, und jeder ist jetzt geprüft
// (`vierklassen_test.go`):
//
//   - ein zweites `game:` oder `stylesheet:` → seit `gedeutet` ein Rohblock
//   - ein zweites `set info:` → dito; das erste bleibt gedeutet
//   - Kommentare und Leerzeilen der Wurzelebene → `beiwerkAus`, denn der
//     Zerteiler hängt sie an den Knoten **davor**
//   - der skalare Rest eines `card: irgendwas` → ein Rohblock neben der Karte
//
// **Der Schreiber war der erste Schritt, nicht die Reparatur**, und das Issue
// hat es so vorhergesagt: seit #48 kommt die Datei byte-genau heraus, weil die
// Quelldatei mitreist (0011). Damit wurde aus „geht beim Schreiben etwas
// verloren" die andere Frage — **ist der Bestand blind?** Er war es, und Suche,
// Statistik und der MCP-Server waren es mit ihm.
var gespeichert = map[string]bool{
	"game":       true,
	"stylesheet": true,
	"set info":   true,
	"card":       true,

	// **Die Schlüsselwörter des Sets** (#1062) — gedeutet und nicht roh.
	"keyword": true,

	// **Die Packarten des Sets** (#1081) — dasselbe. MSE nennt das Feld
	// „Additional/replacement pack types" (`src/data/set.hpp`, Zeile 57,
	// Fassung f9d9356 vom 2021-07-06) und sucht **zuerst hier**, dann im
	// Spiel.
	"pack type": true,

	// **Die fünf Formatangaben des Setkopfs** (#1325) — gedeutet und nicht
	// roh. Sie standen in **jedem** der zehn kanonischen Sets unter „Nicht
	// gedeutet"; was sie bedeuten, steht in `domain/setkopf.go`, nachgelesen
	// in MSE2.
	//
	// **`version_control` ist ein Block**, die übrigen vier sind Zeilen. Der
	// Leser nimmt aus dem Block die Art (`type`) — was ein
	// Versionskontrollsystem täte, tut diese Anwendung nicht.
	"mse version":        true,
	"game version":       true,
	"stylesheet version": true,
	"version control":    true,
	"apprentice code":    true,

	// **Die setweiten Stiloptionen** — `Set::stylingDataFor(const StyleSheet&)`
	// (#557).
	//
	// Sie stehen als eigener Block neben `set info:`, eine Ebene je Stilpaket:
	//
	//	styling:
	//		magic-m15-saga:
	//			text_box_mana_symbols: magic-mana-small.mse-symbol-font
	//
	// **Jedes kanonische Set führt ihn**, und jedes von `magic-set-generator`
	// geschriebene ebenso (`src/mse.rs`, Zeilen 151–174, §11).
	"styling": true,
}
