package service

// **Wer geredet hat, sagt der Empfänger** — Werkstück G5 von Phase 5.5
// (#1462).
//
// # Warum Chat und Präsenz eine eigene Prüfung bekommen
//
// **Weil sie denselben Fehler zulassen wie ein Befehl.** Der Wirt setzt
// `Spielbefehl.Absender` aus der Sitzung, weil ein Gast sonst im Namen eines
// anderen zöge — und genau dieselbe Lücke stand bei Chat und Präsenz offen:
// „Von" kam ungeprüft aus der Nachricht. Ein Gast konnte im Namen eines
// anderen Sitzes reden und dessen Zeiger bewegen.
//
// **Das Issue verlangt es ausdrücklich**: Zeiger, Ziehen, Schweben und Pings
// bekommen „Autorisierung, monotonen Empfängerstand und Ablauf".
//
// # Die drei Zusagen und wo jede eingelöst wird
//
//   - **Autorisierung** — hier, beim Wirt: der Sprecher kommt aus der Sitzung.
//   - **Monotoner Empfängerstand** — hier, an beiden Enden: eine Präsenz, die
//     das andere Ende überholt hat, kommt nicht an ([domain.Praesenzkreis]).
//   - **Ablauf** — **nicht hier.** Er ist eine Frage an die Wanduhr und an
//     das, was jemand sieht; er steht im Fenster, wo er sichtbar wird. Dass er
//     woanders steht, ist eine Entscheidung und kein Versäumnis (§10).
//
// # Und warum der Gast den Sprecher glaubt
//
// **Weil sein einziges Gegenüber der Wirt ist.** Der hat ihn gesetzt, und er
// ist die Autorität am Tisch — dieselbe Aufteilung wie beim Tischbild: der
// Wirt rechnet, der Gast zeichnet.

import (
	"encoding/json"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Nebenbefund ist, was aus einer Nebennachricht geworden ist.
//
// **Die ausgepackte Nutzlast steht dabei**, und zwar aus einem handfesten
// Grund: der Wirt reicht sie weiter, und die Runde nummeriert eine Präsenz
// dabei um. Wer sie dort noch einmal auspackte, bekäme einen Fehlerzweig, den
// kein Lauf betreten kann — die Nutzlast kommt ja von hier (§10).
type Nebenbefund struct {
	Nachricht domain.Partienachricht

	// Chat und Praesenz — **genau eines** trägt etwas, und
	// `Nachricht.Art` sagt welches.
	Chat     domain.Chat
	Praesenz domain.Praesenz

	// Gilt sagt, ob sie wirkt.
	//
	// **Falsch ist kein Fehler**: eine überholte Präsenz ist der Normalfall
	// einer verwerfbaren Zustellung — dieselbe Überlegung wie bei
	// [Partieempfang].
	Gilt bool
}

// Nebenpruefer hält Chat und Präsenz gegen ihre Gestalt und ihren Stand.
//
// **Einer je Leitung.** Der Stand gehört dem Weg, über den er gezählt wurde;
// zwei Leitungen teilen ihn nicht.
type Nebenpruefer struct {
	kreis *domain.Praesenzkreis
}

// NeuerNebenpruefer fängt ohne einen einzigen Stand an.
func NeuerNebenpruefer() *Nebenpruefer {
	return &Nebenpruefer{kreis: domain.NeuerPraesenzkreis()}
}

// Beglaubige prüft eine Nebennachricht und setzt ihren Sprecher.
//
// **`sitz` leer heißt: der Sprecher steht schon fest.** Beim Wirt kommt er aus
// der Sitzung und überschreibt, was der Gast behauptet; beim Gast hat der Wirt
// ihn gesetzt, und der ist die Autorität (siehe den Kopf dieser Datei).
//
// **Ein Befund ohne `Gilt` ist kein Fehler** — der Grund steht bei
// [Nebenbefund.Gilt].
func (p *Nebenpruefer) Beglaubige(
	nachricht domain.Partienachricht, sitz string,
) (Nebenbefund, error) {
	switch nachricht.Art {
	case domain.ArtChat:
		return p.chat(nachricht, sitz)

	case domain.ArtPraesenz:
		return p.praesenz(nachricht, sitz)
	}

	return Nebenbefund{}, fmt.Errorf(
		"%w: %q ist weder Chat noch Präsenz",
		domain.ErrValidation, string(nachricht.Art))
}

// Vergiss streicht den Stand eines Sitzes, der aufgestanden ist.
//
// **Ohne diesen Weg käme niemand zurück** — der Grund steht bei
// [domain.Praesenzkreis.Geht].
func (p *Nebenpruefer) Vergiss(sitz string) { p.kreis.Geht(sitz) }

// chat prüft einen Satz und setzt seinen Sprecher.
func (p *Nebenpruefer) chat(
	nachricht domain.Partienachricht, sitz string,
) (Nebenbefund, error) {
	var gesagt domain.Chat

	if fehler := json.Unmarshal(nachricht.Inhalt, &gesagt); fehler != nil {
		return Nebenbefund{}, fmt.Errorf(
			"%w: der Inhalt ist kein Satz: %s",
			domain.ErrValidation, fehler.Error())
	}

	if sitz != "" {
		gesagt.Von = sitz
	}

	if fehler := gesagt.Pruefe(); fehler != nil {
		return Nebenbefund{}, fehler
	}

	return Nebenbefund{
		Nachricht: mitInhalt(nachricht, gesagt), Chat: gesagt, Gilt: true,
	}, nil
}

// praesenz prüft einen Zeiger, setzt seinen Sitz und urteilt über den Stand.
func (p *Nebenpruefer) praesenz(
	nachricht domain.Partienachricht, sitz string,
) (Nebenbefund, error) {
	var gezeigt domain.Praesenz

	if fehler := json.Unmarshal(nachricht.Inhalt, &gezeigt); fehler != nil {
		return Nebenbefund{}, fmt.Errorf(
			"%w: der Inhalt ist keine Präsenz: %s",
			domain.ErrValidation, fehler.Error())
	}

	if sitz != "" {
		gezeigt.Von = sitz
	}

	if fehler := gezeigt.Pruefe(); fehler != nil {
		return Nebenbefund{}, fehler
	}

	// **Erst prüfen, dann merken.** Ein Kreis, der eine ungültige Präsenz
	// mitzählte, spränge im Stand vor und wiese danach gültige ab.
	if !p.kreis.Nimm(gezeigt) {
		return Nebenbefund{}, nil
	}

	return Nebenbefund{
		Nachricht: mitInhalt(nachricht, gezeigt), Praesenz: gezeigt,
		Gilt: true,
	}, nil
}

// mitInhalt legt die beglaubigte Nutzlast zurück in die Nachricht.
//
// **Hier steht kein Fehlerzweig, und das ist eine Aussage.** [domain.Chat] und
// [domain.Praesenz] tragen ausschließlich Zeichenketten, Zahlen und
// Wahrheitswerte — `encoding/json` klagt darüber nie. Ein Zweig, den kein Lauf
// betreten kann, ist eine Zusage, die niemand hält (§10); gehalten wird die
// Voraussetzung von `internal/domain/sichtgestalt_test.go`, der beide
// Gestalten mitführt — dieselbe Lösung wie beim Tischbild.
func mitInhalt(
	nachricht domain.Partienachricht, nutzlast any,
) domain.Partienachricht {
	inhalt, _ := json.Marshal(nutzlast) //nolint:errcheck,errchkjson // der Grund steht darüber

	nachricht.Inhalt = inhalt

	return nachricht
}
