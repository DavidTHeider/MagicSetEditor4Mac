package app

// Fassung ist die Fassung dieser Anwendung.
//
// # Warum sie im Quelltext steht und nicht nur in `wails.json`
//
// Das Fenster bekommt seine Fassung beim Bauen: Wails setzt
// `productVersion` in `Info.plist`, und macOS zeigt sie an. **Der MCP-Server
// hat kein `Info.plist`** — er ist ein eigenes Binärprogramm über stdio
// ([0010](../../docs/entscheidungen/0010-mcp-als-eigener-prozess.md)) und muss
// die Fassung mitbringen, wenn er sie in seiner Vorstellung melden soll.
//
// Bis zum 2026-08-23 meldete er fest `"0.0.0"`
// (#896) —
// eine Angabe, die wie eine Fassung aussieht und keine ist.
//
// # Wer sie zusammenhält
//
// `internal/werkzeuge/fassung_test.go` hält sie gegen `wails.json`, in **jedem**
// Lauf. Die übrigen drei Orte — `frontend/package.json`, die oberste datierte
// Überschrift in `CHANGELOG.md` und der Tag — hält `versionen_pruefen` im
// Freigabelauf.
//
// **Wer die Fassung erhöht, ändert fünf Orte.** Vier davon merkt die Freigabe,
// diesen einen die Pipeline.
const Fassung = "1.0.0"

// Kenner ist die Selbstnennung dieser Anwendung im Netz.
//
// **`Name/Fassung`**, wie es sich gehört und wie Scryfall es ausdrücklich
// verlangt: „the header should be the name of your application, such as
// MTGExampleApp/1.0" (nachgelesen am 2026-08-23). Für den Paketbezug ist es
// Höflichkeit, für jeden weiteren Dienst Pflicht (§16).
//
// **Hier und nicht in `internal/bezug`**: das Paket ist ein Blatt und kennt die
// Fassung der Anwendung nicht (§2). Sie dort ein zweites Mal zu schreiben wäre
// derselbe Fehler wie der, den
// #896
// behoben hat.
func Kenner() string { return "MagicSetEditor4Mac/" + Fassung }
