// Was ein Wurzelknoten verdrahten darf — die vier Composition Roots.
//
// Ausgelagert aus `regeln_test.go`, weil die Datei mit ihnen über die
// Zeilengrenze aus §2 ginge. Der Schnitt folgt der Frage: dort, **welche
// Schicht welche kennt** — hier, **wer zusammensetzen darf**. Es ist die
// benannte Ausnahme aus §2, und sie steht deshalb sichtbar für sich.
package schichten

// wurzelknoten sagt je Wurzelknoten, welche Schichten er verdrahten darf.
var wurzelknoten = map[string][]string{
	// main.go darf nur die Fassade kennen. Der Composition Root liegt in
	// internal/app, nicht im Wurzelpaket.
	// **Und `mcpnaehte`** seit U12 (#1423): der Wurzelknoten des Fensters
	// reicht der Kommandozentrale dieselben Nähte, die `cmd/mse-mcp` dem
	// Server gibt — ein Wurzelknoten darf zusammensetzen.
	//
	// **Und `fenster` seit ENG-02** (#1727): die Hülle um Wails 3. Sie öffnet
	// das Fenster und setzt die Nähte der Anwendung darauf; `main.go` ruft
	// sie, wie es vorher `wails.Run` rief.
	"main": {"app", "fenster", "mcpnaehte"},

	// **Der MCP-Server ist ein zweiter Composition Root** und darf deshalb
	// verdrahten: er kennt `mcp`, den Bestand und `app` für den Ort der
	// Datenbank. Fachliche Entscheidungen gehören auch hier nicht hinein —
	// welche Werkzeuge es gibt, sagt `internal/mcp`.
	//
	// **`domain` kam mit #52 dazu**, und der Grund gehört danebengeschrieben:
	// der Server muss sagen können, **wer** er ist (`domain.UrheberServer`).
	// Das ist keine fachliche Entscheidung, sondern der Kern der Verdrahtung —
	// ein Composition Root, der seinen eigenen Namen nicht nennen darf, schöbe
	// ihn in eine Schicht ab, die ihn nicht kennt. Die Zeile war ohnehin
	// ungleichmäßig: `storage` steht darin, und `storage` kennt `domain`.
	// **Dazu `rest` und `leitung` seit N6** (#1572): der zweite Transport
	// des Werkzeugservers spricht HTTP, und was er dafür braucht — einen
	// Lauscher, eine Marke, ein Zeugnis — gibt es schon. **Ein zweiter
	// Lauscher daneben wäre ein zweiter Satz Regeln über dieselbe Frage**:
	// jenseits der Schleife ist TLS Pflicht, und das steht in
	// `rest.LauschenMit` an **einer** Stelle.
	"cmd/mse-mcp": {"app", "domain", "leitung", "mcp", "mcpnaehte", "rest", "storage"},

	// **Der dritte Composition Root** (#886): die Kommandozeile. Sie kennt
	// dieselben Schichten wie der MCP-Server — `mcp` nicht, denn sie spricht
	// kein Protokoll.
	"cmd/mse": {"app", "domain", "storage"},

	// **Der vierte Composition Root**
	// (#997):
	// der Dienst auf der Rückschleife. Er kennt dieselben Schichten wie die
	// Kommandozeile, dazu `rest` — so wie `cmd/mse-mcp` sein `mcp` kennt.
	//
	// **Und er ist der einzige unter `cmd/`, der ins Netz darf.** Das steht
	// nicht hier, sondern in `netz_test.go`; hier steht nur, welche Schichten
	// er verdrahten darf. Zwei Wächter, zwei Fragen.
	// **Dazu `leitung` seit N6** (#1572): jenseits der Schleife ist TLS
	// Pflicht, und das Zeugnis trägt die Peeridentität — dasselbe Verfahren
	// wie bei der Leitung, also kein zweiter Krypto-Baukasten
	// ([0115](../../docs/entscheidungen/0115-die-schleife-ist-eine-vorgabe.md)).
	"cmd/mse-rest": {"app", "domain", "leitung", "rest", "storage"},
}
