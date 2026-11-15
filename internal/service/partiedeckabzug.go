package service

// **Aus einem Deck wird ein eingefrorener Abzug** — Werkstück G1 von
// Phase 5.5
// (#1458).
//
// # Eine eigene Datei, und der Schnitt folgt der Frage
//
// **In `partieaufbau.go` steht, wie aus Wünschen ein Tisch wird** — Sitze,
// Profil, Regelwerk, Bibliothek. Hier steht, **was ein Abzug ist**: wie ein
// Deck einfriert, wie sein Abdruck entsteht und wie daraus Objekte werden.
// `partieaufbau.go` stand mit beidem bei 444 Zeilen (§2).
//
// **Und `partieabzug.go` daneben beantwortet eine dritte**: was einen Abzug
// **selbsttragend** macht, wenn er auf einem fremden Rechner zeichnen soll
// (G6, #1463).

import (
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// abzugAus friert ein Deck ein.
func abzugAus(deck domain.Deck) domain.Partiedeck {
	abzug := domain.Partiedeck{
		Herkunft: strconv.FormatInt(deck.ID, 10),
		Karten:   make([]domain.Abzugskarte, 0, len(deck.Plaetze)),

		// **Leer und nicht `nil`** — der Grund steht im Kopf dieser Datei:
		// eine leere Liste sagt „noch keine", ein fehlendes Feld sagt nichts.
		Beigaben: []domain.Beigabe{},
	}

	for i, platz := range deck.Plaetze {
		// **Eine Seite je Karte, und sie trägt die auflösbare Abhängigkeit**
		// (G2, #1459).
		//
		// **`KarteID` und noch nicht `Felder`**: die Kennung macht das Bild
		// **erreichbar** — das braucht G2. Die eingefrorenen Feldwerte machen
		// den Abzug **selbsttragend**, sodass er auf einem fremden Rechner
		// ohne diesen Bestand zeichnet — das braucht G6
		// (#1463),
		// und dort gehört es hin. Zwei Fragen, zwei Werkstücke; sie hier
		// zusammen zu beantworten hieße, die zweite mit einer leeren Abbildung
		// anzudeuten.
		//
		// **Null heißt: ein echter Druck.** Ein Platz nennt entweder eine
		// eigene Karte oder einen Druck — dieselbe Regel wie an
		// [domain.Deckplatz], und ein Bild dieses Programms gibt es dann
		// nicht.
		seiten := []domain.Abzugsseite{{
			Name:    platz.Bezeichnung,
			KarteID: platz.KarteID,
		}}

		abzug.Karten = append(abzug.Karten, domain.Abzugskarte{
			// **Die Kennung ist die des Platzes und nicht die der Karte.**
			// Derselbe Druck darf zweimal im Deck liegen — im Haupt- und im
			// Seitendeck —, und zwei Plätze mit derselben Kennung wären
			// derselbe Eintrag.
			Kennung: platzkennung(platz, i),
			Name:    platz.Bezeichnung,
			Menge:   platz.Menge,
			Druck:   platz.Printing,
			Bereich: platz.Bereich,
			Seiten:  seiten,
		})
	}

	abzug.Abdruck = abdruckDesAbzugs(abzug)

	return abzug
}

// platzkennung gibt eine Kennung, die innerhalb des Abzugs eindeutig ist.
//
// **Die Stelle geht mit ein.** Ein Platz aus dem Bestand hat eine Kennung;
// einer, der nur im Arbeitsspeicher steht, hat die Null — und zwei davon
// wären sonst derselbe Eintrag.
func platzkennung(platz domain.Deckplatz, stelle int) string {
	return "p" + strconv.Itoa(stelle) + "-" + strconv.FormatInt(platz.ID, 10)
}

// abdruckDesAbzugs rechnet den Abdruck über den Inhalt.
//
// **Über den Inhalt und nicht über die Kennung des Decks**: zwei Seiten
// meinen dasselbe, wenn dieselben Karten in denselben Mengen darin liegen —
// und nicht, wenn sie dieselbe Zeile im Bestand nennen.
func abdruckDesAbzugs(abzug domain.Partiedeck) string {
	var gebaut strings.Builder

	for _, karte := range abzug.Karten {
		gebaut.WriteString(karte.Kennung)
		gebaut.WriteString("\x1f")
		gebaut.WriteString(karte.Name)
		gebaut.WriteString("\x1f")
		gebaut.WriteString(karte.Druck)
		gebaut.WriteString("\x1f")
		gebaut.WriteString(strconv.Itoa(karte.Menge))
		gebaut.WriteString("\x1f")
		gebaut.WriteString(string(karte.Bereich))
		gebaut.WriteString("\x1e")
	}

	summe := sha256.Sum256([]byte(gebaut.String()))

	return hex.EncodeToString(summe[:])
}

// bibliothekAus legt jede Karte des Abzugs als Objekt in die Bibliothek.
//
// **Je Stück ein Objekt und nicht je Zeile.** Vier Wälder sind vier Dinge —
// das ist derselbe Satz wie in `spielobjekt.go`, und hier wird er wahr.
//
// **Und das Seitendeck bleibt draußen**: es liegt neben der Partie und nicht
// in der Bibliothek. Wer es hineinlegte, hätte fünfzehn Karten zu viel im
// Deck und niemand wüsste, welche.
//
// **Geprüft wird hier und nicht beim Aufrufer** (Nachlauf zum Zweitblick über
// Phase 7, R1): je Stück entsteht ein Objekt, und eine Menge ohne Grenze
// hielte die Partieleitung auf. Bis dahin prüfte `Platznehmen` vorher und
// der Partieaufbau erst danach.
func bibliothekAus(
	abzug domain.Partiedeck, sitz string,
) ([]domain.Spielobjekt, error) {
	if fehler := abzug.Pruefe(); fehler != nil {
		return nil, fehler
	}

	gelegt := []domain.Spielobjekt{}
	ordnung := 0

	for _, karte := range abzug.Karten {
		if karte.Bereich == domain.BereichSideboard {
			continue
		}

		for stueck := range karte.Menge {
			ordnung++

			gelegt = append(gelegt, domain.Spielobjekt{
				Kennung: sitz + "/" + karte.Kennung + "/" +
					strconv.Itoa(stueck),
				Art:         domain.ObjektKarte,
				Beherrscher: sitz,
				Eigner:      sitz,
				Zone:        domain.ZoneBibliothek,
				Karte:       karte.Kennung,
				Ordnung:     ordnung,
			})
		}
	}

	return gelegt, nil
}

// regelwerkskennung sagt, welches Profil eine Partie bekommt.
//
// # Zwei Wege, und der genannte geht vor
//
