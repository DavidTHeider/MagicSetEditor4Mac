package domain

// **Die Balanceanalyse** — Werkstück A13.2b von Phase 5.3
// (#1452).
//
// # Warum sie mehr braucht als ein Regelhinweis
//
// **Weil „ähnlich" nichts sagt.** Ein Regelhinweis stützt sich auf eine
// Stelle und zitiert sie; eine Balanceaussage stützt sich auf **andere
// Karten**, und der Ähnlichkeitsrang, der sie gefunden hat, ist keine
// Aussage über sie: „eine Rangfusion ordnet, sie bewertet nicht"
// ([0103](../../docs/entscheidungen/0103-wie-zwei-trefferlisten-zusammenkommen.md)).
//
// **Also muss der Unterschied dastehen.** Das Issue sagt es als eigene Zeile:
// „nennt die … ausgewählten Vergleichskarten **und konkrete Unterschiede**
// (z.B. Manawert/Effekt/Zusatzfähigkeit)".
//
// # Und warum es keinen Balancewert gibt
//
// **Weil es kein Feld dafür gibt.** Das Issue verbietet „einen objektiven
// Balancewert aus einem Ähnlichkeitsscore"; eine Prüfung darauf wäre ein
// Wächter über einer Zahl, die niemand braucht. Hier gibt es sie nicht —
// und [Balancefelder] hält, dass das so bleibt
// ([0108](../../docs/entscheidungen/0108-eine-balanceanalyse-hat-keine-zahl.md)).

import "fmt"

// BefundOhneUnterschied steht für einen Vergleich, der keinen nennt.
const BefundOhneUnterschied = "ohneunterschied"

// BefundOhneVergleich steht für eine Balanceaussage ohne eine einzige
// Vergleichskarte.
const BefundOhneVergleich = "ohnevergleich"

// Vergleichskarte ist eine Karte, gegen die verglichen wurde.
type Vergleichskarte struct {
	// Beleg ist die Kennung, unter der sie im Kontextpaket steht.
	//
	// **Eine Kennung und kein Name**: zwei Karten dürfen denselben Namen
	// tragen (A2), und ein Name ließe sich erfinden. Die Kennung nicht.
	Beleg Belegkennung

	// Name ist, wie das Modell sie nennt — zum Lesen, nicht zum Prüfen.
	Name string

	// Unterschiede sind die konkreten: Manawert, Effekt, Zusatzfähigkeit.
	//
	// **Leer ist ein Befund.** Ein Vergleich ohne Unterschied sagt „ähnlich",
	// und das sagt der Rang schon — ohne etwas zu bedeuten.
	Unterschiede []string
}

// Balanceangaben sind, was eine Balanceanalyse außer ihrem Satz nennt.
type Balanceangaben struct {
	Vergleiche []Vergleichskarte

	// NichtBetrachtet sind die Synergien, die außen vor blieben.
	//
	// **Das Issue verlangt sie ausdrücklich.** Eine Balanceaussage über eine
	// Karte allein ist eine über die Karte in einem leeren Deck — und das
	// gibt es nicht.
	NichtBetrachtet []string
}

// Balancefelder nennt die Felder, die eine Balanceanalyse trägt.
//
// **Damit ein Lauf über sie gehen kann** — die Zusage „kein objektiver
// Balancewert" ist eine über die **Gestalt**, und eine Gestalt lässt sich
// aufzählen. Wer ein Feld hinzufügt, trägt es hier ein und stößt dabei auf
// die Frage, ob es eine Bewertung ist.
func Balancefelder() []string {
	return []string{
		"Vergleiche", "NichtBetrachtet",
		"Beleg", "Name", "Unterschiede",
	}
}

// PruefeBalance hält eine Balanceantwort gegen die vergebenen Belege.
//
// **Es verwirft nichts** — dieselbe Regel wie bei [PruefeEinschaetzung]
// ([0107](../../docs/entscheidungen/0107-eine-einschaetzung-ist-keine-regelauskunft.md)):
// was herauskommt, ist die Antwort **und** was an ihrem Anspruch nicht stimmt.
func PruefeBalance(
	satz string, angaben Balanceangaben, belege []Beleg,
) KIEinschaetzung {
	geprueft := PruefeEinschaetzung(EinschaetzungBalance, satz, belege)

	if len(angaben.Vergleiche) == 0 {
		geprueft.Befunde = append(geprueft.Befunde, Belegbefund{
			Art: BefundOhneVergleich, Stelle: "",
			Meldung: "Diese Einschätzung nennt keine Vergleichskarte. Eine " +
				"Aussage darüber, wie eine Karte dasteht, steht im " +
				"Vergleich — ohne ihn ist sie eine Meinung über eine Karte " +
				"allein.",
		})
	}

	geprueft.Befunde = append(geprueft.Befunde,
		vergleichsbefunde(angaben.Vergleiche, belege)...)

	return geprueft
}

// vergleichsbefunde hält jede Vergleichskarte gegen die Belege.
func vergleichsbefunde(
	vergleiche []Vergleichskarte, belege []Beleg,
) []Belegbefund {
	befunde := make([]Belegbefund, 0, len(vergleiche))

	for _, vergleich := range vergleiche {
		if !kennungGibtEs(vergleich.Beleg, belege) {
			befunde = append(befunde, Belegbefund{
				Art: BefundUnbekannteKennung, Stelle: string(vergleich.Beleg),
				Meldung: fmt.Sprintf(
					"Als Vergleich steht %q unter einer Kennung, die es in "+
						"diesem Kontextpaket nicht gibt.", vergleich.Name),
			})

			continue
		}

		if len(vergleich.Unterschiede) == 0 {
			befunde = append(befunde, Belegbefund{
				Art: BefundOhneUnterschied, Stelle: vergleich.Name,
				Meldung: "Zu dieser Vergleichskarte steht kein konkreter " +
					"Unterschied. „Ähnlich“ ist keine Aussage — das sagt " +
					"schon der Rang, der sie gefunden hat.",
			})
		}
	}

	return befunde
}

// kennungGibtEs sagt, ob ein Beleg unter dieser Kennung vergeben wurde.
func kennungGibtEs(gesucht Belegkennung, belege []Beleg) bool {
	for _, beleg := range belege {
		if beleg.Kennung == gesucht {
			return true
		}
	}

	return false
}
