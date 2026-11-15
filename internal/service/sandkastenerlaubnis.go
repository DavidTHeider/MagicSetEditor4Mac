package service

// **Was ein Sitz gerade stellen darf** — Werkstück `G1` von Phase 5.5
// (#1458).
//
// **Herausgezogen aus `sandkasten.go`**, weil die Datei an der Zeilengrenze
// stand (§2). Der Schnitt folgt der Frage: dort steht, was ein Befehl
// **bewirkt**, hier, welcher überhaupt **angeboten** ist.

import "github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"

// Erlaubt sagt, was `fuer` gerade stellen darf.
//
// **Wer keinen Sitz hat, darf nichts.** Ein Zuschauer sieht zu; das ist die
// Rolle, und eine Liste erlaubter Befehle für ihn wäre eine Einladung, die
// der Ausleger danach ablehnen müsste.
func (s *Sandkasten) Erlaubt(partie domain.Partie, fuer string) []string {
	// **Und wer gegangen ist, bekommt nichts mehr angeboten** (#1979) —
	// sonst stünden Knöpfe da, die der Ausleger abweist.
	if partie.SitzVon(fuer) == nil || partie.Gegangen(fuer) {
		return []string{}
	}

	// **Vor dem ersten Zug steht der Beginn** — CR 103.8 (Fassung
	// 2026-09-25), und zwar vor der Priorität: während des Beginns hat sie
	// niemand (siehe `eroeffnungsbefehle.go`).
	if beginnHaelt(partie) {
		return ohneUngestellteFragen(waehrendDesBeginns(), partie, fuer)
	}

	// **Wer nicht an der Reihe ist, darf nichts stellen, was Priorität
	// braucht** — CR 117.1, Werkstück G9 von Phase 5.5
	// (#1592).
	//
	// **Gefragt wird das eingefrorene Profil** und nicht die Fassung: die
	// Partien aus G1 bis G8 führen `prioritaet` als `manuell`, und sie
	// bleiben Sandkästen, in denen jeder jederzeit alles darf (§15, Punkt 4).
	// **Hier ist der Zweig erreichbar**, den der Aufbau nicht hat: er baut
	// immer mit Zug, gelesen werden aber auch alte Partien.
	// **Und eine Partie ohne laufenden Zug ist ein Sandkasten**, gleich was
	// ihr Profil sagt. [Zugstand.Nummer] ist dort null, es gibt keine
	// Priorität zu achten — wer hier nur das Profil fragte, sperrte jeden
	// aus, dessen Partie nie einen Zug begonnen hat.
	if !prioritaetGilt(partie) {
		return ohneUngestellteFragen(Sandkastenbefehle(), partie, fuer)
	}

	if partie.Zug.Prioritaet == fuer {
		return ohneUngestellteFragen(Sandkastenbefehle(), partie, fuer)
	}

	return ohneUngestellteFragen(erlaubtOhnePrioritaet(), partie, fuer)
}

// prioritaetGilt sagt, ob in dieser Partie Priorität vergeben wird.
//
// **Ein Zug läuft, und das eingefrorene Profil setzt sie durch** — sonst ist
// die Partie ein Sandkasten, in dem niemand zu achten ist und keiner die
// Priorität „hat“. Herausgezogen, weil die Frage zwei Fragende hat: was
// angeboten wird ([Sandkasten.Erlaubt]) und wer dran ist ([sitzIstDran]) —
// zwei Fassungen derselben Regel liefen auseinander (§10).
func prioritaetGilt(partie domain.Partie) bool {
	return partie.Zug.Nummer != 0 &&
		partie.Profil.Matrix.Grad(domain.BereichPrioritaet) ==
			domain.UnterstuetzungAutomatisch
}

// ohneUngestellteFragen nimmt die Antworten weg, zu denen es keine Frage gibt.
//
// # Warum das hierher gehört und nicht ins Fenster
//
// **Weil `Erlaubt` die einzige Auskunft darüber ist, was gestellt werden
// darf** — alle vier Flächen lesen sie (§16). Zwei Befehle sind **Antworten
// auf eine offene Frage** und sonst nichts: die Legendenwahl (CR 704.5j) und
// die Reihung ausgelöster Fähigkeiten (CR 603.3b). Wo die Frage nicht steht,
// weist die Domäne sie ab — und ein Knopf, den der Kern danach abwiese, ist
// „eine Einladung, die man zurücknimmt".
//
// **Gesehen am 2026-09-18 beim Bau von `M9`**: die Handlungsleiste zeigte
// „Fähigkeit als Nächste legen" mit einem **leeren** Auswahlfeld, in jedem
// Zug, in dem nichts ausgelöst hatte. Die Legendenwahl stand seit `M12`
// genauso da; es ist dieselbe Klasse, und sie wird hier als Klasse behoben
// (§10).
//
// **Die Frage selbst steht woanders**: beide werden über dem Tisch gestellt
// ([frontend/src/bausteine/Legendenwahlen.tsx] und `Ausloeserreihung.tsx`),
// weil sie das Spiel anhalten. Diese Liste sagt nur, ob der Befehl **jetzt**
// etwas beantwortet.
func ohneUngestellteFragen(
	arten []string, partie domain.Partie, fuer string,
) []string {
	uebrig := make([]string, 0, len(arten))

	for _, art := range arten {
		switch art {
		case BefehlReihung:
			if partie.Ausloeserreihung != fuer {
				continue
			}

		case BefehlLegende:
			if !wirdGefragt(partie, fuer) {
				continue
			}

		// **Und die vier des Beginns** (#1839): angeboten wird der Schritt,
		// den der Kern gerade annähme — eine Starthand nur einmal, das
		// Unterlegen nur mit Schuld.
		case BefehlStarthand, BefehlMulligan, BefehlUnterlegen, BefehlBehalten:
			if !eroeffnungsschritt(partie.Eroeffnungsschritte(fuer), art) {
				continue
			}
		}

		uebrig = append(uebrig, art)
	}

	return uebrig
}

// wirdGefragt sagt, ob für diesen Sitz eine Legendenwahl offensteht.
func wirdGefragt(partie domain.Partie, fuer string) bool {
	for _, wahl := range partie.Legendenwahlen {
		if wahl.Sitz == fuer {
			return true
		}
	}

	return false
}
