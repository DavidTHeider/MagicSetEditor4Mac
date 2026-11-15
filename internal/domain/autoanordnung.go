// **Ob der Tisch die Karten selbst ordnet** — PLAY-04.AK-03
// (#1712).
//
// # Der Anlass
//
// **Heute liegt jede Karte auf einem gerechneten Platz.** `zonenlage.ts`
// legt Reihen, `schlachtfeldlagen` gruppiert nach CR 403.1 — und eine von
// Hand gewählte Lage gibt es nirgends. Das ist für die meisten das Richtige
// und für manche das Gegenteil: wer seine Kreaturen nach Angriffsreihen
// sortiert oder seine Länder links legt, will, dass es so bleibt.
//
// **Das Kriterium verlangt beides**: „Autoanordnung ist abschaltbar und
// berücksichtigt Anhänge und Gruppen; manuelle Positionen bleiben bei
// zulässigen Updates erhalten."
//
// # Warum eine Einstellung und kein Knopf am Tisch
//
// **Weil es eine Gewohnheit ist und keine Handlung.** Wer den Tisch von Hand
// ordnet, will das nicht je Partie neu sagen — dieselbe Überlegung wie bei
// der Tiefe des Kartenformulars ([EinstellungFeldtiefe]).

package domain

// EinstellungAutoanordnung sagt, ob der Tisch die Karten selbst legt.
const EinstellungAutoanordnung = "tisch.autoanordnung"

// Die zwei Werte von [EinstellungAutoanordnung].
const (
	// AutoanordnungAn heißt: der Tisch legt — die Vorgabe.
	//
	// **Sie ist die Vorgabe, weil sie ohne Zutun trägt.** Ein Tisch, der
	// nichts ordnet, bis jemand jede Karte angefasst hat, wäre beim ersten
	// Öffnen ein Haufen.
	AutoanordnungAn = "an"

	// AutoanordnungAus heißt: was jemand hinlegt, bleibt liegen.
	//
	// **Und was niemand angefasst hat, liegt weiterhin gerechnet.** „Aus"
	// heißt nicht „ohne Ordnung", sondern „ohne **Nach**ordnung": eine neu
	// gezogene Karte braucht einen Platz, und den kann nur die Rechnung
	// geben.
	AutoanordnungAus = "aus"
)

// init hängt die Anordnung an die Einstellungsreihe — wie [init] in
// `feldtiefe.go`, und aus demselben Grund: `einstellung.go` ist voll.
func init() {
	einstellungen = append(einstellungen, einstellung{
		Schluessel: EinstellungAutoanordnung,
		Wirkung:    WirkungSofort,
		Bereich:    BereichAnwendung,
		Werte:      []string{AutoanordnungAn, AutoanordnungAus},
		Vorgabe:    AutoanordnungAn,
	})
}
