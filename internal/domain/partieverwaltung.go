// **Löschen und Kopieren einer Partie** — PLAY-03.AK-05
// (#1711):
// „Löschen betrifft eigene Entwürfe **nach definierter Regel**, Kopien werden
// neue Übungszustände **ohne alte Peer-Zustimmungen**."
//
// # Die Regel, und warum diese
//
// **Ein Entwurf ist eine Partie, die niemandem etwas zugesagt hat.** Zwei
// Dinge nehmen ihr das:
//
//   - **Sie ist beendet.** AK-05 sagt im letzten Satz: „Abgeschlossene
//     gemeinsame Partien bleiben nachvollziehbar." Was mit ihr geschieht,
//     entscheidet die **Datenpflege** nach ihrer Frist (`alteProtokolle`, 90
//     Tage) — und nicht ein Knopf, den jemand im Ärger drückt.
//   - **Jemand hat eine Einwilligung gegeben.** Eine Offenlegung ist die
//     Entscheidung eines Menschen über seine eigenen Geheimnisse (G5,
//     [0078](../../docs/entscheidungen/0078-der-wirksame-akteur.md)); eine
//     Partie, die eine trägt, ist kein Entwurf mehr.
//
// **Warum nicht „ohne Gäste"**: wer an einem Tisch saß, hinterlässt dort
// keine Zeile — eine Sitzung belegt einen **Rechner** und keinen Sitz (G5).
// Eine Regel über Gäste ließe sich damit nicht prüfen, und eine Regel, die
// man nicht prüfen kann, ist keine.

package domain

// Loeschbar sagt, ob eine Partie gelöscht werden darf.
//
// **`offenlegungen` ist ihre Zahl und nicht die Liste**: was darin steht,
// geht diese Entscheidung nichts an — dass es etwas gibt, schon.
func Loeschbar(partie Partie, offenlegungen int) bool {
	return partie.Stand != StandBeendet && offenlegungen == 0
}

// Kopiere gibt die Partie als neuen Übungszustand.
//
// # Was mitgeht und was nicht
//
// **Die Stellung geht mit** — Karten, Sitze, Leben, Zug: das ist der Sinn
// eines Übungszustands. Man stellt eine Lage nach und spielt sie anders.
//
// **Die Geschichte geht nicht mit.** Eine Kopie fängt bei Fassung 1 an; der
// Ereignisstrom der Vorlage gehört zu ihr und nicht hierher. Wer beides
// zusammenlegte, hätte eine Kette, deren Glieder aus zwei Partien stammen —
// und der Abdruck, der sie zusammenhält, wäre wertlos
// ([0044](../../docs/entscheidungen/0044-was-ein-werkzeug-schreiben-darf.md)).
//
// **Und die Einwilligungen erst recht nicht.** Sie liegen in einer eigenen
// Tafel und werden schlicht nicht mitkopiert — „ohne alte
// Peer-Zustimmungen", wörtlich.
//
// **Sie läuft, auch aus einer beendeten heraus.** Ein Übungszustand, der
// entschieden anfinge, wäre keiner.
func Kopiere(partie Partie, neueID string) Partie {
	kopie := partie

	kopie.ID = neueID
	kopie.Fassung = 1
	kopie.Stand = StandLaeuft
	kopie.Ausgang = Partieausgang{}

	// **Jede Liste wird gelegt und nicht geteilt.** Ein geteiltes Slice
	// machte aus zwei Partien eine: wer in der Kopie zöge, änderte das
	// Original.
	kopie.Sitze = append([]Sitz(nil), partie.Sitze...)
	kopie.Zuschauer = append([]Zuschauer(nil), partie.Zuschauer...)
	kopie.Objekte = append([]Spielobjekt(nil), partie.Objekte...)
	kopie.Legendenwahlen = append([]Legendenwahl(nil), partie.Legendenwahlen...)
	kopie.Ausgeloest = append([]Ausloeser(nil), partie.Ausgeloest...)
	kopie.Dauereffekte = append([]Dauereffekt(nil), partie.Dauereffekte...)
	kopie.Ersetzungen = append([]Ersetzungseffekt(nil), partie.Ersetzungen...)
	kopie.Mischpflichten = append([]Mischpflicht(nil), partie.Mischpflichten...)

	return kopie
}
