package mcp

import (
	"context"
	"encoding/json"
)

// Art sagt, ob ein Werkzeug liest oder schreibt.
//
// **Nicht nur beschriftet, sondern getrennt** (#143): ein Werkzeug, das ein Feld
// setzt, gehört nicht in dieselbe Reihe wie eines, das Karten aufzählt. Die Art
// steht an drei Stellen — im Namen, im Text und in den Angaben, die MCP dafür
// vorsieht.
type Art string

const (
	Lesend     Art = "lesen"
	Schreibend Art = "schreiben"
)

// Werkzeug ist ein Werkzeug, wie MCP es kennt.
type Werkzeug struct {
	// Name trägt die Art als Präfix: `lesen.karten` oder `schreiben.feld`.
	Name string

	Art        Art
	Erklaerung string

	// Schema beschreibt die Argumente nach JSON Schema.
	Schema map[string]any

	// Idempotent sagt zu, dass ein zweiter Ruf **keine zusätzliche Wirkung**
	// hat.
	//
	// **Der Wortlaut ist der des Protokolls**, nachgelesen am 2026-08-23 in
	// `schema/2025-06-18/schema.ts`, Zeilen 904–912:
	//
	//	If true, calling the tool repeatedly with the same arguments
	//	will have no additional effect on the its environment.
	//
	// „Dasselbe Ergebnis" wäre **enger** — eine Zusage über die Antwort statt
	// über die Umgebung
	// (#915).
	// `schreiben.feld` gibt zweimal dieselbe Antwort und schreibt dabei
	// **zwei** Verlaufsschritte; unter der engen Lesart wäre das idempotent,
	// unter der richtigen nicht.
	//
	// **Am Werkzeug und nicht an der Art** (#895): `schreiben.karte.anlegen`
	// und `schreiben.karte.kopieren` legen bei jedem Ruf eine **neue Karte mit
	// neuer Kennung** an.
	//
	// **Kein schreibendes Werkzeug sagt sie heute zu**, und das ist keine
	// Bequemlichkeit: jede Änderung schreibt einen Verlaufsschritt
	// (`internal/storage/feldaenderung.go`), damit sie sich zurücknehmen
	// lässt — und ein zweiter Schritt ist eine zusätzliche Wirkung.
	//
	// **Der Hinweis ist eine Erlaubnis.** Ein Gegenüber, das nach einer
	// Zeitüberschreitung wiederholt, darf sich auf ihn verlassen — und genau
	// das ist der Fall, für den es ihn gibt. Es entstünden Doppelkarten.
	//
	// **Die Vorgabe ist `false`, und das ist Absicht.** Wer ein Werkzeug
	// hinzufügt, sagt Idempotenz ausdrücklich zu oder gar nicht; ein
	// vergessenes Feld ist dann eine zu schwache Angabe und keine falsche.
	// `idempotenz_test.go` **führt die Zusage aus**: es ruft zweimal und
	// vergleicht.
	Idempotent bool

	// Zeigerwechsel sagt, dass das Werkzeug den Zeiger bewegt — in ein anderes
	// Profil oder zu einem anderen Menschen. **Es läuft auch, wo nur gelesen
	// werden darf** (ID-03, #1702): sonst käme niemand aus einem fremden Profil
	// wieder heraus. Die Schreibwache ([Server.MitSchreibwache]) fragt es nicht.
	Zeigerwechsel bool

	// Ruf führt das Werkzeug aus und gibt seinen Text.
	Ruf func(ctx context.Context, argumente json.RawMessage) (string, error)

	// Namensfelder tragen einen **Namen**, keinen Ort auf der Platte — etwa
	// `datei` bei `lesen.paketdatei`, der Name einer Datei **im** Paket. Sie
	// gehen nicht durch den Pfadvertrag (Befund 4 des Zweitblicks über Phase
	// 6.4).
	Namensfelder []string

	// Pfadfelder tragen einen **Ort auf der Platte** unter einem Namen, der
	// nach keinem klingt — `quelle` beim Installieren eines Pakets, `ort`
	// beim Übernehmen. Sie gehen durch den Pfadvertrag wie die Namen aus
	// `domain.IstPfadfeld` (#2073); für alle Werkzeuge ginge es nicht, denn
	// `quelle` eines Preises ist ein Anbieter.
	Pfadfelder []string
}

// beschreibung gibt das Werkzeug in der Form, die MCP erwartet.
func (w Werkzeug) beschreibung() map[string]any {
	schema := w.Schema
	if schema == nil {
		schema = map[string]any{"type": "object", "properties": map[string]any{}}
	}

	return map[string]any{
		"name":        w.Name,
		"description": w.Erklaerung,
		"inputSchema": schema,
		"annotations": map[string]any{
			"readOnlyHint": w.Art == Lesend,

			// **Ein schreibendes Werkzeug gilt als zerstörend, und das ist
			// die ehrlichere Angabe.** Es überschreibt einen Wert und kann ein
			// Feld wegnehmen; beides ist nicht rückgängig zu machen, ohne den
			// Vorwert zu kennen. Der erste Anlauf sagte hier `false` und
			// begründete es damit, dass nichts ein Set löscht — das ist wahr
			// und beantwortet die Frage nicht.
			//
			// Ein Gegenüber, das vor zerstörenden Werkzeugen nachfragt, soll
			// hier nachfragen.
			"destructiveHint": w.Art == Schreibend,

			// **Vom Werkzeug, nicht von der Art** (#895) — siehe
			// [Werkzeug.Idempotent].
			"idempotentHint": w.Idempotent,
		},
	}
}

// objekt baut ein JSON-Schema für ein Objekt mit Pflichtfeldern.
func objekt(felder map[string]any, pflicht ...string) map[string]any {
	schema := map[string]any{"type": "object", "properties": felder}
	if len(pflicht) > 0 {
		schema["required"] = pflicht
	}

	return schema
}

// text beschreibt ein Zeichenkettenfeld.
func text(erklaerung string) map[string]any {
	return map[string]any{"type": "string", "description": erklaerung}
}

// jaNein beschreibt ein Wahrheitsfeld.
func jaNein(erklaerung string) map[string]any {
	return map[string]any{"type": "boolean", "description": erklaerung}
}

// zahl beschreibt ein Zahlenfeld.
func zahl(erklaerung string) map[string]any {
	return map[string]any{"type": "integer", "description": erklaerung}
}

// objektliste beschreibt ein Feld mit mehreren Objekten.
//
// **Neu in Phase 4.5** (#1292): `schreiben.liste.einlesen` nimmt eine Liste
// von Zeilen, und `zahlenliste` daneben kann nur Zahlen. Ohne sie stünde das
// Schema dort ausgeschrieben — und das ist die Stelle, an der eine zweite
// Schreibweise für dasselbe entsteht.
func objektliste(erklaerung string, jeEintrag map[string]any) map[string]any {
	return map[string]any{
		"type": "array", "description": erklaerung, "items": jeEintrag,
	}
}

// textliste beschreibt ein Feld mit mehreren Zeichenketten.
//
// **Neu in A4** (#1443): `lesen.marktangebote` lässt mehrere Erhaltungen zu,
// und die tragen Leerzeichen — eine Liste in einer Zeichenkette bräuchte ein
// Trennzeichen, das in den Werten vorkommen kann.
func textliste(erklaerung, jeEintrag string) map[string]any {
	return map[string]any{
		"type": "array", "description": erklaerung, "items": text(jeEintrag),
	}
}

// zahlenliste beschreibt ein Feld mit mehreren Zahlen.
func zahlenliste(erklaerung, jeEintrag string) map[string]any {
	return map[string]any{
		"type": "array", "description": erklaerung, "items": zahl(jeEintrag),
	}
}

// bruch beschreibt ein Feld mit einer gebrochenen Zahl.
//
// **Neu am 2026-09-02** (#1391): der Symboleditor rechnet in Anteilen von
// null bis eins — eine Griffspitze auf halber Höhe ist `0.5`. `zahl` daneben
// sagt `integer`, und ein Modell, das dem Schema folgt, schickte dann `0`.
func bruch(erklaerung string) map[string]any {
	return map[string]any{"type": "number", "description": erklaerung}
}

// rohliste beschreibt ein Feld mit beliebig geformten Einträgen.
//
// **Für die Teile eines Symbols** (#1391). Ihre Gestalt ist rekursiv — eine
// Gruppe enthält Teile —, und ein JSON-Schema dafür wäre hier eine zweite,
// bald abweichende Beschreibung dessen, was `internal/app` ohnehin
// festlegt. Die Erklärung nennt die Felder; die Gestalt kommt aus
// »lesen.symbol« und geht unverändert zurück.
func rohliste(erklaerung string) map[string]any {
	return map[string]any{
		"type": "array", "description": erklaerung,
		"items": map[string]any{"type": "object"},
	}
}
