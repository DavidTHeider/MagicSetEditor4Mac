package domain

// **Was an einem Schalter hängt und läuft, endet mit ihm** — OPS-10b
// (#1978).
//
// # Die Regel
//
// **Ein Lauf endet, sobald sein Beenden gesperrt ist.** Nicht eine zweite
// Zuordnung von Schaltern zu Läufen neben der Wegetafel, sondern dieselbe
// Tafel: wer das Horchen nicht mehr beenden darf, weil `netz` aus ist, hätte
// sonst einen offenen Port, den das Fenster nicht mehr schließen kann. Damit
// folgt der Lauf jeder Änderung der Tafel von selbst.
//
// **Eingeschaltet startet keiner von selbst** — wer ihn wieder will, startet
// ihn (OPS-10.AK-03).

// Lauf ist etwas im Fenster, das läuft, bis jemand es beendet.
type Lauf string

// Die Läufe, deren Beenden an einem Schalter hängt.
const (
	LaufPartiebeitritt Lauf = "partiebeitritt"
	LaufPartiefreigabe Lauf = "partiefreigabe"
	LaufSichtbarkeit   Lauf = "sichtbarkeit"
	LaufHorchen        Lauf = "horchen"
	LaufFernhorchen    Lauf = "fernhorchen"
	LaufIndexaufbau    Lauf = "indexaufbau"
	LaufBildauftraege  Lauf = "bildauftraege"
	LaufREST           Lauf = "rest"
	LaufMCP            Lauf = "mcp"
)

// beendenwege nennt je Lauf die Bindung, die ihn beendet — in der Ordnung,
// in der sie enden: **erst die Sitzungen, dann die Ankündigung, dann die
// Ports**. Ein Gegenüber erfährt so vom Aufstehen, bevor der Weg zu ihm
// zugeht.
var beendenwege = []struct {
	lauf Lauf
	weg  string
}{
	{LaufPartiebeitritt, "PartieBeitrittBeenden"},
	{LaufPartiefreigabe, "PartieFreigabeBeenden"},
	{LaufSichtbarkeit, "SichtbarkeitBeenden"},
	{LaufHorchen, "HorchBeenden"},
	{LaufFernhorchen, "FernHorchBeenden"},
	{LaufIndexaufbau, "IndexAbbrechen"},
	{LaufBildauftraege, "KIBildAbbrechen"},
	{LaufREST, "RESTBeenden"},
	{LaufMCP, "MCPBeenden"},
}

// Laeufe nennt jeden Lauf, in der Ordnung, in der sie enden.
func Laeufe() []Lauf {
	laeufe := make([]Lauf, 0, len(beendenwege))
	for _, eintrag := range beendenwege {
		laeufe = append(laeufe, eintrag.lauf)
	}

	return laeufe
}

// Beendenweg nennt die Bindung des Fensters, die diesen Lauf beendet.
func (l Lauf) Beendenweg() string {
	for _, eintrag := range beendenwege {
		if eintrag.lauf == l {
			return eintrag.weg
		}
	}

	return ""
}

// GesperrteLaeufe nennt die Läufe, deren Beenden dieser Stand sperrt — die
// also jetzt enden müssen, weil es später niemand mehr könnte.
func (t Wegetafel) GesperrteLaeufe(stand Schalterstand) []Lauf {
	var laeufe []Lauf

	for _, eintrag := range beendenwege {
		if t.Pruefe(FlaecheFenster, eintrag.weg, stand) != nil {
			laeufe = append(laeufe, eintrag.lauf)
		}
	}

	return laeufe
}
