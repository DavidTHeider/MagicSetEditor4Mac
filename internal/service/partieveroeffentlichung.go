package service

// **Hinaus geht es in der Reihenfolge, in der der Kern annahm** — #2481
// (Zweitblick über Phase 7 von 0.9.0, Befund 8).
//
// # Was offen stand
//
// **Das Schloss `zug` ordnete die Züge, aber nicht, was danach hinausging.**
// Verkündet wurde außerhalb: ein späterer Zug konnte sein Tischbild
// verschicken, bevor der frühere seines verschickt hatte — und der Gast
// behielt das ältere, denn sein Standhalter übernimmt, was ankommt.
//
// # Wie es jetzt geht
//
// **Ein zweites Schloss, `reihe`, übernimmt, bevor `zug` fällt.** Dazwischen
// gibt es keinen Augenblick, in dem ein späterer Zug angenommen **und**
// verkündet sein könnte. Wer liest und schickt, ohne zu schreiben — die
// Absage, der Gruß, die Rückkehr —, nimmt dasselbe Schloss
// ([Partieleitung.InReihe]): auch sein Bild darf kein neueres überholen.
//
// # Was es nicht hält
//
//   - **Ein Schloss für alle Partien dieses Prozesses**, wie `zug`. Ein Wirt
//     hält eine Partie zur Zeit für Gäste offen (`app.PartieFreigeben`).
//   - **Ein Gast, der nicht liest, hält die Reihe an** — bis die Frist der
//     Leitung abläuft. Verkündet wird über `Leere`, und das schreibt. Bis
//     hierher hielt er nur den an, der gerade verkündete; jetzt wartet auch
//     der nächste Zug. Die Schleife zwischen zwei Anwendungen im selben
//     Prozess hat einen Puffer (`leitung.Schleifenpuffer`), wie ein Socket.
//   - **Zwischen Prozessen ordnet es nichts.** Dort verkündet niemand:
//     Kommandozeile, MCP und REST schreiben, und das Fenster erfährt es über
//     den Datenstand.

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Verkuendung trägt einen angenommenen Zug an den Tisch.
//
// **Ohne Fehler**: der Zug ist geschrieben (§7, Regel 3). Was beim Verkünden
// scheitert, entscheidet der Rufer — der Wirt beendet die Sitzung, das
// Fenster meldet es.
type Verkuendung func(Befehlsausgang)

// BefehlVerkuendet stellt einen Befehl und verkündet ihn, bevor ein späterer
// Zug verkündet wird.
//
// **Ein abgewiesener Befehl verkündet nichts** — seine Absage nimmt
// [Partieleitung.InReihe].
func (l *Partieleitung) BefehlVerkuendet(
	ctx context.Context, befehl domain.Spielbefehl, verkuende Verkuendung,
) (Befehlsausgang, error) {
	return l.inReihe(func() (Befehlsausgang, error) {
		return l.befehl(ctx, befehl)
	}, verkuende)
}

// InReihe führt `tun` aus, während keine andere Verkündung läuft.
//
// **Für Antworten, die lesen und schicken, ohne zu schreiben.** Ein Bild,
// das vor einer Verkündung gelesen und nach ihr geschickt würde, überholte
// ein neueres.
//
// **Nicht verschachteln**: wer darin einen Zug verkündet, wartet auf sich
// selbst.
func (l *Partieleitung) InReihe(tun func() error) error {
	l.reihe.Lock()
	defer l.reihe.Unlock()

	return tun()
}

// inReihe stellt unter `zug` und verkündet unter `reihe`.
//
// **Ohne Verkündung keine Reihe**: wer nur schreibt, hält niemanden auf, der
// verkündet — sein Zug geht nirgends hin.
func (l *Partieleitung) inReihe(
	stellen func() (Befehlsausgang, error), verkuende Verkuendung,
) (Befehlsausgang, error) {
	ausgang, fehler := l.unterZug(stellen, verkuende != nil)
	if fehler != nil || verkuende == nil {
		return ausgang, fehler
	}

	defer l.reihe.Unlock()

	verkuende(ausgang)

	return ausgang, nil
}

// unterZug stellt unter dem Schloss des Zuges und nimmt, wenn gewünscht, die
// Reihe, **bevor** es fällt.
//
// **Ein `defer` und kein Freigeben von Hand**: ein `panic` in `stellen` ließe
// sonst jeden späteren Zug warten.
func (l *Partieleitung) unterZug(
	stellen func() (Befehlsausgang, error), verkuendet bool,
) (Befehlsausgang, error) {
	l.zug.Lock()
	defer l.zug.Unlock()

	ausgang, fehler := stellen()
	if fehler == nil && verkuendet {
		l.reihe.Lock()
	}

	return ausgang, fehler
}
