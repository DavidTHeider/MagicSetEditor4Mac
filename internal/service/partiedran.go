package service

// **Ob die Partie auf einen Sitz wartet** — UXR-F16
// (#2031),
// Folgebefund aus UXR-13
// (#1921).
//
// # Warum die Rechnung hier steht
//
// **Weil die Antwort Fachlogik ist** (§2): wer am Zug ist und wer Priorität
// hat, steht im Spielstand der Regelmaschine, und das Artboard *Spielen* zeigt
// es an jeder Partie. Das Fenster liest es von hier ab, statt es aus dem Zug
// zu erraten — und die drei Flächen ohne Fenster nennen dasselbe Feld (§16).
//
// # Was „dran" heißt
//
// **Ein Sitz ist dran, wenn die Partie auf ihn wartet.** Die Regelmaschine
// kennt dafür vier Stellen und eine fünfte davor, und jede steht schon
// irgendwo, wo sie etwas anderes beantwortet — hier werden sie nicht neu
// erfunden, nur gefragt (Fassung 2026-09-25 für alle Nummern):
//
//   - **im Beginn** der Sitz, der noch einen Schritt hat — CR 103.5, über
//     [domain.Partie.Eroeffnungsschritte], und vor allem anderen: während des
//     Beginns hat niemand Priorität (CR 103.8, [beginnHaelt]);
//   - **danach der Sitz mit der Priorität** — CR 117.1, [domain.Zugstand.Prioritaet];
//   - **der aktive Spieler mit offener Abwurfpflicht** — CR 514.1: im
//     Aufräumschritt hat niemand Priorität, und der Schritt wartet auf seine
//     Wahl;
//   - **wer eine Legende wählen soll** — CR 704.5j;
//   - **wer seine ausgelösten Fähigkeiten ordnen soll** — CR 603.3b.
//
// **Die letzten drei sind Fragen an einen Sitz, die das Spiel anhalten**, und
// dieselben, die [ohneUngestellteFragen] als Antworten anbietet. Sie gelten,
// wo sie stehen — auch dort, wo die Priorität nicht durchgesetzt wird.
//
// # Wo niemand dran ist
//
// **Im freien Modus.** Eine Partie ohne laufenden Zug, oder deren Profil die
// Priorität nicht durchsetzt, ist ein Sandkasten (§15, Punkt 4): jeder darf
// jederzeit alles, und die Regelmaschine kennt keine Priorität, die einem Sitz
// gehörte. Dort „dran" zu sagen, benennte einen, den niemand vergeben hat —
// ausgenommen eine offene Frage (siehe oben), denn die steht im Zustand und
// nicht in der Durchsetzung.
//
// **In einer angehaltenen oder beendeten Partie** ebenso: sie wartet auf
// niemanden. **Und für einen Sitz, der gegangen ist** (CR 800.4a): er stellt
// nichts mehr, und die Priorität ist an den Nächsten gegangen.

import "github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"

// eigenerSitz gibt den Sitz, den dieses Profil in der Partie hat, und ob die
// Partie auf ihn wartet.
//
// # Der Sitz, der dran ist, geht vor
//
// **Wer gegen sich selbst spielt, hat zwei Sitze** ([domain.Partie.SitzeVonProfil]).
// Die Zeile nennt dann den, an dem etwas zu tun ist — der Mensch, der „dran"
// liest, will wissen, wo; und wo keiner dran ist, den ersten am Tisch, damit
// die Antwort nicht von der Laune einer Reihenfolge abhängt.
//
// **Leer heißt: das Profil sitzt hier nicht** — und dann gibt es keine Frage,
// die mit „nein" zu beantworten wäre.
func eigenerSitz(partie domain.Partie, profilID int64) (string, bool) {
	sitze := partie.SitzeVonProfil(profilID)

	for _, sitz := range sitze {
		if sitzIstDran(partie, sitz) {
			return sitz, true
		}
	}

	if len(sitze) == 0 {
		return "", false
	}

	return sitze[0], false
}

// sitzIstDran sagt, ob die Partie gerade auf diesen Sitz wartet.
//
// **Der Sitz sitzt an diesem Tisch**: er kommt aus [eigenerSitz] und damit aus
// [domain.Partie.SitzeVonProfil]. Ein Zweig für einen, den es nicht gibt, stünde
// ungeprüft da — keine Eingabe erreicht ihn (§10).
func sitzIstDran(partie domain.Partie, sitz string) bool {
	if partie.Stand != domain.StandLaeuft || partie.Gegangen(sitz) {
		return false
	}

	// **Der Beginn geht vor** — wie in [Sandkasten.Erlaubt]: solange er
	// läuft, hat niemand Priorität, auch wenn der Zustand eine nennt.
	if beginnHaelt(partie) {
		return partie.Eroeffnungsschritte(sitz) != domain.Eroeffnungsschritte{}
	}

	if partie.Ausloeserreihung == sitz || wirdGefragt(partie, sitz) ||
		(partie.Zug.Abwurf > 0 && partie.Zug.Aktiv == sitz) {
		return true
	}

	return prioritaetGilt(partie) && partie.Zug.Prioritaet == sitz
}
