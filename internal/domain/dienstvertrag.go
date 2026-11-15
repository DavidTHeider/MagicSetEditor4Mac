package domain

// **Der Providervertrag je Fähigkeit** — Werkstück A1 von Phase 5.3
// (#1440),
// entschieden in [0082](../../docs/entscheidungen/0082-der-dienstkatalog.md).
//
// # Was ein Vertrag ist
//
// **Eine Fähigkeit ist ein Dienst und eine Frage** — »Scryfall kann suchen«,
// »magicthegathering.io kann suchen«, »Scryfall kann Regeln nennen«. Zu
// jeder steht hier, was über den Anbieter **nachgelesen** ist: wo er
// antwortet, welche Endpunkte dieses Programm ruft, wie es sich ausweist,
// was es kostet, wie viel erlaubt ist, wie lange eine Antwort als frisch
// gilt, unter welchen Bedingungen — und **wann** das gelesen wurde, mit der
// Adresse, unter der es sich prüfen lässt (§16, Regel 1).
//
// # Warum er im Quelltext steht und nicht nur im Vermerk
//
// Vermerk 0056 führt die drei neueren Dienste in einer Tabelle, 0052 die
// Bedingungen von Scryfall, und jedes Klientenpaket trägt seine Zahlen im
// Kommentar. **Vier Orte, und keiner davon ist eine Auskunft**: das Fenster
// konnte sagen »an« oder »aus«, aber nicht, was ein Dienst hier kann und
// unter welchen Bedingungen. Der Katalog ist die eine Stelle, die alle vier
// Flächen lesen — und ein Wächter (`internal/werkzeuge/dienstvertraege_test.go`)
// hält ihn gegen die Pakete: die Basisadresse **ist** deren `Adresse`, und
// jeder genannte Endpunkt steht in deren Quelltext.
//
// # Nur, was gebaut ist
//
// Ein Vertrag steht hier, weil es für Dienst und Frage einen **Weg** gibt —
// einen Klienten, der den Endpunkt ruft, und einen Verbraucher in `service`.
// Was ein Anbieter außerdem anbietet, steht hier nicht: ein Katalog, der
// Fähigkeiten nennt, die niemand rufen kann, wäre ein Universalanbieter mit
// Methoden, die »nicht unterstützt« melden (#1440).
//
// # Der Nachweis ist kein Gütesiegel
//
// [Livenachweis] sagt: **dieser** Lauf ist an **diesem** Tag gegen **diese**
// Fassung gelaufen. Mehr sagt er nicht — und die Oberfläche darf nicht mehr
// daraus machen. Ein Nachweis aus einer früheren Fassung ist ein Nachweis
// über die frühere Fassung; was er über den heutigen Bau sagt, ist, dass
// jemand ihn wiederholen kann.

// Die Fragen, die ein Dienst hier beantworten kann.
//
// **Die Suche steht in `fragen.go`**, weil sie die einzige mit einer Wahl
// ist. Die übrigen beantwortet heute nur Scryfall; sie stehen hier, damit
// der Katalog sie **als Fähigkeit** führen kann — mit Vertrag und Status —,
// ohne dass sie eine Einstellung bekommen (eine Wahl mit einer Möglichkeit
// ist keine).
const (
	// FrageKartendaten ist »nenne mir die Daten einer Karte« — nach Name,
	// Kennung, als Sammlung oder als ganzes Set.
	FrageKartendaten Frage = "kartendaten"
	// FrageBild ist »gib mir ein Bild dieser Karte«.
	FrageBild Frage = "bild"
	// FrageRegeln ist »welche Regelhinweise gibt es zu dieser Karte«.
	FrageRegeln Frage = "regeln"
	// FrageManakosten ist »welche Symbole gibt es, und was kosten sie« — die
	// Symbologie, gegen die ein Manakostenfeld geprüft wird.
	FrageManakosten Frage = "manakosten"
	// FrageSetangaben ist »was ist über dieses Set bekannt«.
	FrageSetangaben Frage = "setangaben"
	// FragePreise ist »was kostet dieser Druck heute«.
	FragePreise Frage = "preise"
	// FrageSammeldaten ist »gib mir alle Karten auf einmal« — die Bulk-Dateien.
	FrageSammeldaten Frage = "sammeldaten"

	// FrageKIVorschlag ist »entwirf mir Felder für diese Karte« — die erste
	// Frage, die kein Kartendienst beantwortet (A8, #1447).
	//
	// **Sie ist mehrstimmig** ([Frage.Mehrstimmig]): die Modelle aller
	// eingeschalteten Anbieter stehen nebeneinander im Katalog, und wer
	// fragt, nennt Anbieter und Modell in derselben Bedienhandlung. Eine
	// Wahl daneben wäre ein Schalter, den keine Stelle liest.
	FrageKIVorschlag Frage = "kivorschlag"

	// FrageKonto ist »was liegt in meinem Konto bei diesem Anbieter« — die
	// eigenen Angebote und die eigenen Bestellungen (T11,
	// #1476).
	//
	// **Einstimmig, und das bleibt sie vermutlich**: ein Konto hat man bei
	// **einem** Anbieter, und zwei Konten nebeneinander wären zwei Antworten
	// auf zwei Fragen und nicht zwei auf eine.
	FrageKonto Frage = "konto"
)

// Livenachweis ist ein Lauf gegen den echten Dienst — mit Datum und Fassung.
//
// **Datum und Fassung sind der Nachweis**, nicht der Lauf allein: eine
// Datei hinter dem Bautag `netzlauf` belegt, dass es den Weg gibt; wann er
// zuletzt gegangen wurde und gegen welchen Stand, steht hier. Wer den Lauf
// wiederholt, trägt beides nach.
type Livenachweis struct {
	// Lauf ist die Testdatei hinter dem Bautag `netzlauf`, vom Wurzelverzeichnis aus.
	Lauf string
	// Datum ist der Tag des letzten Laufs, JJJJ-MM-TT.
	Datum string
	// Fassung ist die Fassung der Anwendung, mit der er lief.
	Fassung string
	// Endpunkte sind die, die der Lauf wirklich gerufen hat.
	//
	// **Nicht alle des Vertrags**: ein Lauf, der die Suche fährt, sagt nichts
	// über die Bulk-Dateien — »pro Endpunkt« ist der Grad von Gewissheit, den
	// #1440 verlangt.
	Endpunkte []string
}

// Dienstvertrag ist, was über eine Fähigkeit eines Dienstes feststeht.
type Dienstvertrag struct {
	Dienst Dienst
	Frage  Frage

	// Wirkung ist die Klasse aus `akteur.go` — heute überall [WirkungExternSenden].
	//
	// **Als Feld und nicht als Konstante**, weil die erste schreibende
	// Fähigkeit nach draußen mit Phase 5.6 kommt und dann **hier** steht
	// statt in einem `if` über den Dienst.
	Wirkung Wirkung

	// Basisadresse ist die Grundadresse, Zeichen für Zeichen die des Pakets.
	Basisadresse string
	// Endpunkte sind die Wege, die dieses Programm für diese Frage ruft.
	Endpunkte []string
	// Zugang sagt, wie sich die Anwendung ausweist — oder dass sie es nicht muss.
	Zugang string
	// Tarif sagt, was es kostet.
	Tarif string
	// Grenze nennt die veröffentlichte Grenze und die Drossel dieses Programms.
	Grenze string
	// Frist sagt, wie lange eine Antwort hier als frisch gilt.
	Frist string
	// Bedingungen ist die Adresse der Nutzungsbedingungen.
	Bedingungen string
	// Abgerufen ist der Tag, an dem Bedingungen und Grenzen gelesen wurden.
	Abgerufen string
	// Beleg ist die Adresse, unter der sich die Zahlen nachlesen lassen.
	Beleg string

	// Nachweis ist der letzte Lauf gegen den echten Dienst — oder nil.
	//
	// **nil heißt: nie gegen sein Original gelaufen**, und die Oberfläche
	// sagt das dort, wo man ihn einschaltet (0056, Entscheidung 4).
	Nachweis *Livenachweis
}

// dienstvertraege ist der Katalog — in Anzeigereihenfolge.
//
// **Je Dienst eine Datei**, weil jede ihre eigene Quelle und ihr eigenes
// Abrufdatum hat: `dienstvertragscryfall.go`, `dienstvertragweitere.go`,
// `dienstvertragmarkt.go` und `dienstvertragki.go`.
var dienstvertraege = append(append(append(scryfallvertraege(),
	weitereVertraege()...), marktvertraege()...), kivertraege()...)

// Dienstvertraege gibt den Katalog — eine Kopie, in Anzeigereihenfolge.
func Dienstvertraege() []Dienstvertrag {
	kopie := make([]Dienstvertrag, len(dienstvertraege))
	copy(kopie, dienstvertraege)

	return kopie
}

// VertragFuer gibt den Vertrag eines Dienstes für eine Frage.
func VertragFuer(dienst Dienst, frage Frage) (Dienstvertrag, bool) {
	for _, vertrag := range dienstvertraege {
		if vertrag.Dienst == dienst && vertrag.Frage == frage {
			return vertrag, true
		}
	}

	return Dienstvertrag{}, false
}

// FragenVon gibt die Fragen, die ein Dienst beantworten kann — in Katalogreihenfolge.
func FragenVon(dienst Dienst) []Frage {
	var fragen []Frage

	for _, vertrag := range dienstvertraege {
		if vertrag.Dienst == dienst {
			fragen = append(fragen, vertrag.Frage)
		}
	}

	return fragen
}
