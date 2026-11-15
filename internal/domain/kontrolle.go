package domain

// **Was eine Kontrollnachricht meint** — Werkstück G5 von Phase 5.5
// (#1462).
//
// # Warum die Art nicht reicht
//
// **Weil „Kontrolle" in beide Richtungen geht und Verschiedenes bedeutet.**
// Vom Gast zum Wirt ist sie ein **Wunsch**; vom Wirt zum Gast ist sie das
// **Tischbild** — die Sicht, die für ihn gerechnet wurde. Wer beides unter
// einem Namen führte, müsste am Inhalt raten, und die erste neue Form ginge
// still als die alte durch (§8).
//
// # Und warum der Gast keine eigene Rechnung bekommt
//
// **Weil nur der Wirt rechnet.** Er hält den Vollzustand, er legt aus, er
// vergibt Folge und Kette; der Gast bekommt eine **Projektion** und zeichnet
// sie. Genau so steht es seit G1 in `Partie.Sicht`: die Sicht ist der einzige
// Weg von der Wahrheit zur Oberfläche — auch über eine Leitung.
//
// **Was das kostet, gehört ausgesprochen**: der Gast kann ohne den Wirt gar
// nichts, nicht einmal ansehen, was zuletzt galt. Die Alternative wäre ein
// zweiter Spielkern auf der Gastseite, und der hätte beim ersten Unterschied
// eine zweite Wahrheit.

import (
	"encoding/json"
	"fmt"
	"slices"
)

// Kontrollart sagt, was eine Kontrollnachricht will.
type Kontrollart string

const (
	// KontrolleWunsch ist ein Spielbefehl vom Gast an den Wirt.
	KontrolleWunsch Kontrollart = "wunsch"

	// KontrolleTischbild ist die Sicht, die der Wirt für einen Gast
	// gerechnet hat.
	KontrolleTischbild Kontrollart = "tischbild"

	// KontrolleRueckkehr ist der Wunsch eines Gastes, der zurückkommt.
	//
	// **Er sagt, wo der Gast stehengeblieben ist** — und er kommt auch beim
	// **ersten** Mal, mit Folge null. Ein Sonderweg „erstes Mal" wäre ein
	// zweiter Weg zum selben Ziel (§16).
	KontrolleRueckkehr Kontrollart = "rueckkehr"

	// KontrolleAbzug ist die Antwort darauf.
	//
	// **Sie heißt nach ihrem schwersten Fall.** Bei drei der vier Wege trägt
	// sie nur eine Auskunft; beim vierten trägt sie den **Abzug** — das
	// Tischbild, mit dem ein Gast wieder anfängt, dessen Stand sich nicht
	// mehr erklären lässt.
	KontrolleAbzug Kontrollart = "abzug"

	// KontrolleBildwunsch fragt nach dem Bild eines Objekts — P9 (#2377),
	// Vermerk 0198; KontrolleBild ist die Antwort darauf.
	KontrolleBildwunsch Kontrollart = "bildwunsch"
	KontrolleBild       Kontrollart = "bild"
)

// Kontrollarten gibt alle — für Läufe und für Oberflächen.
func Kontrollarten() []Kontrollart {
	return []Kontrollart{
		KontrolleWunsch, KontrolleTischbild,
		KontrolleRueckkehr, KontrolleAbzug,
		KontrolleBildwunsch, KontrolleBild,
	}
}

// Pruefe sagt, ob es die Art gibt.
func (a Kontrollart) Pruefe() error {
	if slices.Contains(Kontrollarten(), a) {
		return nil
	}

	return fmt.Errorf("%w: %q ist keine Kontrollart — es gibt %v",
		ErrValidation, string(a), Kontrollarten())
}

// Kontrolle ist der Umschlag einer Kontrollnachricht.
//
// **Ein Umschlag und kein Verbund aus allem** — wer jede mögliche Nutzlast als
// Feld führte, hätte bei jeder Nachricht alle bis auf eine leer, und niemand
// sähe, welche gemeint war.
type Kontrolle struct {
	Art Kontrollart `json:"art"`

	// Inhalt bleibt roh, bis die Art sagt, was er ist.
	//
	// **`json.RawMessage` und keine Bytes**: so steht die Nutzlast lesbar in
	// der Nachricht statt als Base64 — dasselbe wie bei [Partienachricht].
	Inhalt json.RawMessage `json:"inhalt"`
}

// Pruefe hält die Gestalt.
func (k Kontrolle) Pruefe() error {
	if fehler := k.Art.Pruefe(); fehler != nil {
		return fehler
	}

	if len(k.Inhalt) == 0 {
		return fmt.Errorf("%w: eine Kontrolle ohne Inhalt sagt nichts",
			ErrValidation)
	}

	return nil
}

// VerpackeKontrolle macht aus Art und Nutzlast die fertigen Bytes.
//
// # Warum Bytes und nicht eine [Kontrolle]
//
// **Weil jeder Aufrufer sie ohnehin verpacken müsste** — und dabei einen
// Fehlerzweig bekäme, den kein Lauf betreten kann: eine Kontrolle, deren
// Inhalt aus `json.Marshal` stammt, lässt sich immer verpacken. Ein Zweig,
// den niemand erreicht, ist eine Zusage, die niemand hält (§10). So gibt es
// **eine** Stelle, an der etwas schiefgehen kann, und die ist erreichbar:
// eine Nutzlast, die JSON nicht kennt.
func VerpackeKontrolle(art Kontrollart, nutzlast any) ([]byte, error) {
	if fehler := art.Pruefe(); fehler != nil {
		return nil, fehler
	}

	inhalt, fehler := json.Marshal(nutzlast)
	if fehler != nil {
		return nil, fmt.Errorf(
			"die Nutzlast einer Kontrolle ließ sich nicht verpacken: %w",
			fehler)
	}

	// **Hier kann nichts mehr scheitern**: beide Felder sind gültiges JSON,
	// und `json.Marshal` über zwei davon ebenfalls. Deshalb steht hier kein
	// zweiter Fehlerzweig — er wäre unerreichbar.
	roh, _ := json.Marshal(Kontrolle{Art: art, Inhalt: inhalt}) //nolint:errcheck,errchkjson // der Grund steht darüber

	return roh, nil
}

// LiesKontrolle packt eine Kontrolle aus und hält sie gegen die erwartete Art.
//
// **Eine Stelle für beide Seiten.** Wirt und Gast erwarten je andere Arten,
// und beide stellen dieselben drei Fragen: ist es überhaupt eine Kontrolle,
// hält sie ihre Gestalt, und ist es eine erwartete? Zwei Kopien liefen beim
// ersten neuen Feld auseinander.
//
// **Mehrere erwartete Arten seit der Wiederaufnahme**, und die Art kommt
// zurück: beide Seiten nehmen jetzt zwei entgegen und müssen danach
// unterscheiden. Ein zweiter Ruf „welche war es denn?" packte den Umschlag ein
// zweites Mal aus — und bekäme dabei einen Fehlerweg, den kein Lauf betreten
// kann (§10).
//
// **Wer die falsche Art bekommt, bekommt keinen Formfehler.** Ein Wunsch an
// einen Gast oder ein Tischbild an den Wirt ist eine **Anmaßung**: der eine
// hat keinen Vollzustand, der andere rechnet selbst. Die Meldung sagt das.
func LiesKontrolle(
	roh []byte, von string, erwartet ...Kontrollart,
) (Kontrollart, json.RawMessage, error) {
	var umschlag Kontrolle

	if fehler := json.Unmarshal(roh, &umschlag); fehler != nil {
		return "", nil, fmt.Errorf("%w: der Inhalt ist keine Kontrolle: %s",
			ErrValidation, fehler.Error())
	}

	if fehler := umschlag.Pruefe(); fehler != nil {
		return "", nil, fehler
	}

	if !slices.Contains(erwartet, umschlag.Art) {
		return "", nil, fmt.Errorf(
			"%w: %q schickt %q, erwartet war %v — das nimmt die Gegenseite "+
				"entgegen und nicht diese", ErrValidation, von,
			string(umschlag.Art), erwartet)
	}

	return umschlag.Art, umschlag.Inhalt, nil
}

// Tischbild ist, was ein Gast nach jedem Zug bekommt.
//
// # Warum zwei Dinge und nicht eines
//
// **Weil es zwei Fragen sind.** Die Sicht sagt, **wie der Tisch aussieht** —
// das ist eine Aussage über den Zustand. Was erlaubt ist, sagt das
// **Regelwerk**, und es hängt nicht am Zustand allein: dieselbe Stellung
// erlaubt einem Sitz etwas, das sie einem anderen verwehrt.
//
// **Und die Aufteilung trägt weiter, als sie heute muss.** Heute gibt der
// Sandkasten jedem Sitz dieselben neun Befehle; wenn einmal eine Zugstruktur
// dazukommt, ist das hier die Stelle, an der ein Gast erfährt, dass er gerade
// nichts darf — ohne dass die Sicht sich ändert.
type Tischbild struct {
	Sicht Partiesicht `json:"sicht"`

	// Erlaubt sind die Befehlsarten, die dieser Sitz gerade stellen darf.
	//
	// **Eine leere Liste heißt: nichts.** Sie ist eine Antwort und sieht
	// anders aus als keine.
	Erlaubt []string `json:"erlaubt"`

	// Handlungen sagen je Objekt, was dieser Sitz daran tun kann (P3,
	// #2371). **Erweiternd**: ein Wirt einer älteren Fassung schickt sie
	// nicht, und der Gast sieht dann eine leere Liste.
	Handlungen []Objekthandlungen `json:"handlungen,omitempty"`

	// Absage sagt, dass der Kern den letzten Wunsch dieses Gastes abwies —
	// mit Grund (#2478). **Ein Zusatz und keine eigene Kontrollart**: ein Gast
	// vor 0.9.0 liest das Feld nicht und bekommt ein frisches Bild; eine neue
	// Art beendete gerade seine Sitzung (`Partienetz.Bediene`, §16).
	Absage *Wunschabsage `json:"absage,omitempty"`
}

// Wunschabsage ist ein Wunsch, den der Kern abwies, und warum.
type Wunschabsage struct {
	// Befehl ist die Kennung des Wunsches, Art seine Befehlsart.
	Befehl string `json:"befehl"`
	Art    string `json:"art"`

	// Grund ist der Satz des Kerns — für Menschen.
	Grund string `json:"grund"`
}
