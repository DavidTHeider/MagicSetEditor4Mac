package domain

// **Die vier Kartendienste** — Phase 4.8
// (#1321),
// entschieden in [0056](../../docs/entscheidungen/0056-drei-weitere-kartendienste.md)
// und [0063](../../docs/entscheidungen/0063-was-ein-dienst-hier-ist.md).

import (
	"errors"
	"fmt"
)

// Dienst ist ein fremder Kartendienst.
//
// **Der Name wird nicht übersetzt** (0063, Entscheidung 5): „Scryfall" und
// „Scrydex" sind Eigennamen (§9). Übersetzt werden die **Fragen** und die
// **Zustände**.
type Dienst string

// Die sieben, und ihre Reihenfolge ist die ihrer Aufnahme.
const (
	DienstScryfall   Dienst = "scryfall"
	DienstMTGAPI     Dienst = "mtgapi"
	DienstMTGJSON    Dienst = "mtgjson"
	DienstScrydex    Dienst = "scrydex"
	DienstCardTrader Dienst = "cardtrader"

	// DienstAnthropic ist der erste KI-Anbieter — Werkstück A8 von
	// Phase 5.3
	// (#1447).
	//
	// **Er steht bei den Kartendiensten, und das ist eine Entscheidung.**
	// Was ihn dorthin bringt, ist nicht, dass er Karten liefert — er tut es
	// nicht —, sondern dass er dieselben drei Fragen beantwortet, für die es
	// diese Aufzählung gibt: darf er gefragt werden, welche Zugangsdaten
	// braucht er, und wo liegen sie. Eine zweite Aufzählung daneben liefe
	// bei der ersten Änderung auseinander (siehe den Kopf von
	// `einstellung.go`).
	DienstAnthropic Dienst = "anthropic"

	// DienstOpenAI ist der zweite KI-Anbieter — Werkstück A9 von Phase 5.3
	// (#1448).
	//
	// **Derselbe Grund wie bei Anthropic**: er beantwortet dieselben drei
	// Fragen, für die es diese Aufzählung gibt.
	DienstOpenAI Dienst = "openai"

	// DienstXAI ist der dritte KI-Anbieter — Werkstück A10 von Phase 5.3
	// (#1449).
	//
	// **Derselbe Grund wie bei den beiden darüber**, und ein eigener dazu:
	// sein Schlüssel liegt im Schlüsselbund und nicht im Bestand, und wo er
	// liegt, sagt genau diese Aufzählung.
	DienstXAI Dienst = "xai"
)

// Die zwei Stellungen eines Schalters.
//
// **„Aus" heißt: kein einziger Aufruf geht hinaus** (0063, Entscheidung 1) —
// auch nicht als Ausweichweg, auch nicht im Hintergrund. Eine Zusage, die
// Ausnahmen hätte, wäre nichts wert.
const (
	DienstAus = "aus"
	DienstAn  = "an"
)

// Die Angaben, die ein Zugang braucht.
//
// **Nachgelesen am 2026-08-30** (0056): MTGJSON nimmt `authorization: Bearer
// <Token>`, Scrydex verlangt **zwei** Kopfzeilen — `X-Api-Key` **und**
// `X-Team-ID`. Ein Dienst, der nur nach einem fragte, wäre eingerichtet und
// antwortete trotzdem nicht.
const (
	ZugangMarke      = "marke"
	ZugangSchluessel = "schluessel"
	ZugangTeam       = "team"
)

// dienstzeile ist, was über einen Dienst feststeht.
//
// **Eine Zeile je Dienst und keine zweite Liste daneben.** Dieselbe Gestalt
// wie bei den Einstellungen, und aus demselben Grund: dort ist eine zweite
// Aufzählung zweimal auseinandergelaufen, und beide Male hat es ein Nachlauf
// gefunden (siehe den Kopf von `einstellung.go`).
type dienstzeile struct {
	Dienst  Dienst
	Vorgabe string

	// Zugang sind die Angaben, die im Schlüsselbund liegen — leer, wenn der
	// Dienst ohne auskommt.
	Zugang []string

	// Adresse sagt, ob dieser Dienst hinter einem Gateway stehen darf —
	// Werkstück 3 von A18
	// (#1533).
	//
	// **Die Vorgabe steht nicht hier**, sondern im Blatt des Dienstes
	// (`openai.Vorgabeadresse`): `domain` kennt die Blätter nicht (§2), und
	// eine zweite Aufzählung liefe beim ersten Wechsel auseinander.
	Adresse bool
}

// dienstzeilen ist die Aufzählung, aus der alles Übrige folgt.
var dienstzeilen = []dienstzeile{
	// **Scryfall bleibt an.** Es ist heute schon an, und die Anwendung könnte
	// ohne es weniger als gestern (0063, Verworfen).
	{Dienst: DienstScryfall, Vorgabe: DienstAn},

	// **Die drei neuen fangen aus an.** Ein Dienst, dessen Bedingungen
	// jemand nicht gelesen hat, soll nicht in seinem Namen fragen.
	{Dienst: DienstMTGAPI, Vorgabe: DienstAus},
	{Dienst: DienstMTGJSON, Vorgabe: DienstAus, Zugang: []string{ZugangMarke}},
	{
		Dienst:  DienstScrydex,
		Vorgabe: DienstAus,
		Zugang:  []string{ZugangSchluessel, ZugangTeam},
	},

	// **CardTrader fängt ebenso aus an** — und es ist der erste Dienst hier
	// mit einem **Marktplatz**: seine Antworten sind Angebote fremder
	// Verkäufer, nicht eine Schätzung je Karte (A4, #1443,
	// [0095](../../docs/entscheidungen/0095-was-cardtrader-sagt-und-was-davon-gilt.md)).
	//
	// **Eine Marke, wie MTGJSON**: `Authorization: Bearer <Marke>`,
	// nachgelesen am 2026-09-11.
	{Dienst: DienstCardTrader, Vorgabe: DienstAus, Zugang: []string{ZugangMarke}},

	// **Anthropic fängt ebenso aus an** und verlangt **einen** Schlüssel:
	// `x-api-key`, nachgelesen am 2026-09-11. Keine `Authorization`
	// daneben — ein zweiter Zugang, den niemand vergibt, wäre eine
	// Kopfzeile, die der Dienst abweist.
	{
		Dienst:  DienstAnthropic,
		Vorgabe: DienstAus,
		Zugang:  []string{ZugangSchluessel},
		Adresse: true,
	},

	// **OpenAI fängt ebenso aus an** und verlangt **einen** Schlüssel:
	// `Authorization: Bearer <Schlüssel>`, nachgelesen am 2026-09-12.
	//
	// **Und er ist nicht das Abonnement.** Das Issue sagt es ausdrücklich:
	// „ChatGPT/Codex-Abonnement nicht als API-Berechtigung oder verfügbares
	// Kontingent voraussetzen." Wer ChatGPT bezahlt, hat kein API-Guthaben —
	// und bekäme hier einen 429 mit `credit_balance_exhausted`.
	{
		Dienst:  DienstOpenAI,
		Vorgabe: DienstAus,
		Zugang:  []string{ZugangSchluessel},
		Adresse: true,
	},

	// **xAI fängt ebenso aus an** und verlangt **einen** Schlüssel:
	// `Authorization: Bearer <Schlüssel>`, nachgelesen am 2026-09-12.
	//
	// **Und er ist nicht das Abonnement.** Wer X Premium bezahlt oder Grok
	// im Browser benutzt, hat damit kein API-Guthaben — das ist derselbe
	// Satz wie bei OpenAI, und er steht hier, weil er bei beiden zutrifft
	// und bei Anthropic ebenso.
	{
		Dienst:  DienstXAI,
		Vorgabe: DienstAus,
		Zugang:  []string{ZugangSchluessel},
		Adresse: true,
	},
}

// Dienste gibt sie in ihrer Reihenfolge.
func Dienste() []Dienst {
	alle := make([]Dienst, 0, len(dienstzeilen))
	for _, zeile := range dienstzeilen {
		alle = append(alle, zeile.Dienst)
	}

	return alle
}

// DienstEinstellung gibt den Einstellungsschlüssel eines Dienstes.
//
// **Ein Präfix und kein eigener Bereich im Schema**: die Einstellungen sind
// Schlüssel und Wert (Migration 0010), und eine Spalte je Dienst hieße eine
// Wanderung je Dienst.
func DienstEinstellung(dienst Dienst) string {
	return "dienst." + string(dienst)
}

// BrauchtSchluessel sagt, ob dieser Dienst ohne Zugangsdaten arbeitet.
func (d Dienst) BrauchtSchluessel() bool { return len(d.Zugangsangaben()) > 0 }

// Zugangsangaben gibt die Angaben, die im Schlüsselbund liegen müssen.
func (d Dienst) Zugangsangaben() []string {
	for _, zeile := range dienstzeilen {
		if zeile.Dienst == d {
			return zeile.Zugang
		}
	}

	return nil
}

// ZugangKonto gibt den Kontonamen, unter dem eine Angabe im Schlüsselbund
// liegt.
//
// **Er muss ohne Leerzeichen und Sonderzeichen auskommen**, weil er auf der
// Befehlszeile von `security` mitreist — `internal/schluesselbund` weist alles
// andere ab, und das ist dort gemessen und begründet.
func ZugangKonto(dienst Dienst, angabe string) string {
	return fmt.Sprintf("%s.%s", dienst, angabe)
}

// ErrDienstAus sagt, dass ein Dienst abgeschaltet ist.
//
// **Ein eigener Ausgang und kein `ErrValidation`.** Die Abhilfe ist eine
// andere: nicht „berichtige die Eingabe", sondern „schalte ihn ein, wenn du
// ihn willst". Wer beide zusammenlegte, schickte jemanden auf die falsche
// Suche (§8).
var ErrDienstAus = errors.New("der Dienst ist abgeschaltet")

// SchluesselDienstAus ist der Schlüssel dieser Abweisung — mit dem Dienst
// als Wert `dienst`, wie `schalter` bei [SchluesselBereichAbgeschaltet].
const SchluesselDienstAus Schluessel = "dienstAus"

// AbgeschalteterDienst ist die Abweisung eines abgeschalteten Dienstes —
// **mit dem Dienst**, damit die Oberfläche ihn im Rat nennen kann (I05,
// #2263). Er ist ein [ErrDienstAus], und seine Meldung ist dieselbe wie
// vorher.
type AbgeschalteterDienst struct {
	Dienst Dienst
}

func (a *AbgeschalteterDienst) Error() string {
	return fmt.Sprintf("%s: %s darf nicht gefragt werden — unter Einstellungen "+
		"lässt er sich einschalten", ErrDienstAus, a.Dienst)
}

func (a *AbgeschalteterDienst) Unwrap() error { return ErrDienstAus }

// DienstIstAus baut die Meldung, die jemand zu sehen bekommt.
func DienstIstAus(dienst Dienst) error {
	return &AbgeschalteterDienst{Dienst: dienst}
}

// ZugangFehlt sagt, dass eine Angabe im Schlüsselbund fehlt.
//
// **Ein eigener Grund und nicht Schweigen** — Zusatzbefund 1 des Nachlaufs
// über Phase 4.8. Vorher verschwand ein Dienst ohne Zugang schlicht aus der
// Karte, und die Meldung riet danach zwischen zwei Möglichkeiten: „fehlt sein
// Zugang, oder ist er abgeschaltet?" Das ist eine Frage zurück an den
// Lesenden, keine Auskunft (§8).
func ZugangFehlt(dienst Dienst, angabe string) error {
	return fmt.Errorf(
		"%w: %s braucht %s — unter Einstellungen lässt sich der Zugang "+
			"eintragen", ErrValidation, dienst, angabe)
}

// IstDienstAus sagt, ob ein Fehler daher kommt, dass ein Dienst aus ist.
func IstDienstAus(err error) bool { return errors.Is(err, ErrDienstAus) }
