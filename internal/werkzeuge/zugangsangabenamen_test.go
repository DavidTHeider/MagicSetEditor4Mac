package werkzeuge_test

// **Eine Beschriftung je Dienst, wo eine je Angabeart nicht trägt** —
// #1534.
//
// # Der Befund
//
// **Bei CardTrader stand „Patreon-Marke".** `domain.ZugangMarke` bedient
// **zwei** Dienste — MTGJSON und CardTrader —, und der Katalog beschriftete
// nach der **Angabeart**. MTGJSONs Marke gibt es nur für Patreon-Unterstützer
// ([0056](../../docs/entscheidungen/0056-drei-weitere-kartendienste.md)),
// CardTraders kommt aus den Profileinstellungen seines eigenen Kontos
// ([0095](../../docs/entscheidungen/0095-was-cardtrader-sagt-und-was-davon-gilt.md)).
// Eine Zeichenkette kann nicht für beide stimmen.
//
// # Was hier gehalten wird
//
// **Zwei Zusagen, und die zweite gehört zur ersten.**
//
//  1. Eine Angabeart, die **mehr als einen** Dienst bedient, hat je Dienst
//     ihren Namen — **oder** sie steht in [geteilteAngaben] mit einem Grund.
//  2. Jeder Schlüssel in `dienstangaben` benennt einen **wirklichen** Dienst
//     und eine seiner Angaben.
//
// **Ohne die zweite wäre die erste eine Zusage ohne Deckung**: `'cardtradr.marke'`
// sähe aus wie ein Eintrag, fiele im Fenster still auf den geteilten Namen
// zurück — und der Befund wäre zurück, mit einem Wächter davor, der grün ist.
//
// # Was hier **nicht** gehalten wird
//
// **Ob der Name stimmt.** Dass CardTraders Marke aus den Profileinstellungen
// kommt, ist nachgelesen und steht im Vermerk; dieser Lauf sieht, dass es
// **einen eigenen** Namen gibt, nicht welchen.

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Und jede Zugangsangabe trägt einen übersetzten Namen** — Werkstück 8 von
// Phase 4.8.
//
// # Warum
//
// `Zugaenge.tsx` setzte die Angabe **roh** ins Fenster: neben „MTGJSON" stand
// `marke: fehlt`, neben „Scrydex" `schluessel: fehlt` und `team: fehlt`. Das
// sind Bezeichner aus `internal/domain/dienste.go`, keine Oberflächentexte —
// und in der englischen Fassung standen sie unverändert da.
//
// **`keintextimjsx.test.ts` sieht das nicht**: es weist Zeichenketten im JSX
// ab, und hier stand eine **Variable**. Gefunden hat es die Sichtprüfung am
// laufenden Fenster.
//
// # Was er nicht hält
//
// **Ob der Name den Kopfzeilennamen des Anbieters trifft.** Er sieht, dass ein
// Name dasteht. Dass `marke` bei MTGJSON die Patreon-Marke meint und `team`
// Scrydex' `X-Team-ID`, steht in
// [0056](../../docs/entscheidungen/0056-drei-weitere-kartendienste.md) und ist
// nachgelesen, nicht von hier.
//
// # Und was er seit dem 2026-09-12 sehr wohl hält
//
// **Dass eine Angabeart, die mehrere Dienste bedient, nicht mit einem Namen
// für alle davonkommt** (#1534).
// Bis dahin stand bei CardTrader „Patreon-Marke" — `domain.ZugangMarke`
// bedient MTGJSON **und** CardTrader, und der Katalog beschriftete nach der
// **Angabeart**. Der Satz darüber war ehrlich, und die Form darunter machte
// ihn unhaltbar: solange eine Beschriftung für zwei Anbieter gilt, kann kein
// Wächter mehr versprechen.
//
// **Das ist die Klasse und nicht der Fall** (§10): geprüft wird nicht, ob
// „Patreon" an der richtigen Stelle steht, sondern ob eine geteilte
// Beschriftung überhaupt geteilt sein darf.
func TestJedeZugangsangabeHatEinenNamen(t *testing.T) {
	t.Parallel()

	for _, sprache := range []string{"de", "en"} {
		t.Run(sprache, func(t *testing.T) {
			t.Parallel()

			angaben := angabenblock(t, sprache)
			jeDienst := dienstangabenblock(t, sprache)

			for _, dienst := range domain.Dienste() {
				for _, angabe := range dienst.Zugangsangaben() {
					// **Entweder je Dienst oder über den Rückfall** — beides
					// zusammen ist die Zusage, dass nie ein roher Bezeichner
					// im Fenster steht.
					if jeDienst[string(dienst)+"."+angabe] || angaben[angabe] {
						continue
					}

					t.Errorf("im Katalog %s hat %q bei %q keinen Namen — "+
						"weder unter `dienstangaben` noch als Rückfall "+
						"unter `angaben`; das Fenster zeigte sonst den "+
						"rohen Bezeichner", sprache, angabe, dienst)
				}
			}
		})
	}
}

// angabenblock liest die Namen aus dem Block `angaben` eines Katalogs.
//
// **Der Block und nicht die ganze Datei.** Ein Name wie `team` steht sonst
// womöglich anderswo, und der Wächter wäre grün, ohne dass ihn jemand als
// Angabe eingetragen hätte.
func angabenblock(t *testing.T, sprache string) map[string]bool {
	t.Helper()

	block := regexp.MustCompile(`(?s)\n {2}angaben: \{(.*?)\}`).
		FindStringSubmatch(zugangskatalog(t, sprache))
	if block == nil {
		t.Fatalf("im Zugangskatalog %s steht kein Block `angaben: {…}`", sprache)
	}

	gefunden := map[string]bool{}

	for _, zeile := range regexp.MustCompile(`(?m)^\s*([a-z][a-z0-9]*):`).
		FindAllStringSubmatch(block[1], -1) {
		gefunden[zeile[1]] = true
	}

	return gefunden
}

// geteilteAngaben nennt die Angabearten, deren Rückfallname für **jeden**
// Dienst stimmt, der sie benutzt — samt Grund.
//
// **Eine Aufzählung und keine Regel**, aus demselben Grund wie bei
// [namensweg]: wer eine Angabeart aufnimmt, die mehrere Dienste bedient,
// entscheidet hier — der Lauf wird sonst rot und nennt sie.
//
// **Der Eintrag ist eine Behauptung und wird als solche gelesen.** Wer eine
// Angabeart hier einträgt, sagt: dieser Name trifft jeden Anbieter, der sie
// benutzt. Für `marke` traf er das nicht, und deshalb steht sie nicht hier.
var geteilteAngaben = map[string]string{
	domain.ZugangSchluessel: "„API-Schlüssel“ sagt nichts über den Anbieter — " +
		"Scrydex, Anthropic, OpenAI und xAI verlangen alle einen, und keiner " +
		"nennt ihn anders",
}

// **Eine geteilte Beschriftung braucht je Dienst ihren Namen — oder einen
// Grund.**
func TestEineGeteilteAngabeartTraegtIhrenDienstnamen(t *testing.T) {
	t.Parallel()

	dienstejeAngabe := map[string][]domain.Dienst{}

	for _, dienst := range domain.Dienste() {
		for _, angabe := range dienst.Zugangsangaben() {
			dienstejeAngabe[angabe] = append(dienstejeAngabe[angabe], dienst)
		}
	}

	for _, sprache := range []string{"de", "en"} {
		t.Run(sprache, func(t *testing.T) {
			t.Parallel()

			jeDienst := dienstangabenblock(t, sprache)

			for angabe, dienste := range dienstejeAngabe {
				pruefeGeteilte(t, sprache, angabe, dienste, jeDienst)
			}
		})
	}
}

// pruefeGeteilte sieht eine Angabeart samt ihren Diensten an.
func pruefeGeteilte(
	t *testing.T, sprache, angabe string, dienste []domain.Dienst,
	jeDienst map[string]bool,
) {
	t.Helper()

	pruefeGeteilteMit(klagenAn{t}, sprache, angabe, dienste, jeDienst)
}

// klagend nimmt einen Befund entgegen.
//
// **Damit die Gegenprobe dieselbe Regel fahren kann**, ohne einen Lauf rot zu
// machen: `pruefeGeteilteMit` **ist** die Regel, und ein zweiter Nachbau
// daneben prüfte den Nachbau (§10).
type klagend interface{ klage(satz string) }

// klagenAn macht aus einem Lauf einen Empfänger.
type klagenAn struct{ t *testing.T }

func (k klagenAn) klage(satz string) {
	k.t.Helper()
	k.t.Error(satz)
}

// pruefeGeteilteMit ist die Regel selbst.
func pruefeGeteilteMit(
	an klagend, sprache, angabe string, dienste []domain.Dienst,
	jeDienst map[string]bool,
) {
	// **Eine Angabeart mit genau einem Dienst kann nicht danebenliegen** —
	// ihr Rückfallname gilt für ihn und für niemanden sonst.
	if len(dienste) < 2 {
		return
	}

	if _, geteilt := geteilteAngaben[angabe]; geteilt {
		return
	}

	for _, dienst := range dienste {
		if !jeDienst[string(dienst)+"."+angabe] {
			an.klage(fmt.Sprintf(
				"im Katalog %s benutzen %d Dienste die Angabeart %q, und %q "+
					"hat keinen eigenen Namen — ein Rückfall für alle kann "+
					"nicht für jeden Anbieter stimmen. Entweder %q.%s "+
					"eintragen oder die Angabeart in `geteilteAngaben` "+
					"begründen", sprache, len(dienste), angabe, dienst,
				dienst, angabe))
		}
	}
}

// **Und jeder Eintrag im neuen Block benennt einen wirklichen Dienst.**
//
// **Sonst fiele er still auf den Rückfall zurück** — also auf genau den
// Namen, wegen dem es diesen Block gibt.
func TestJederDienstangabenschluesselNenntEinenWirklichenDienst(t *testing.T) {
	t.Parallel()

	echt := echteDienstangaben()

	for _, sprache := range []string{"de", "en"} {
		t.Run(sprache, func(t *testing.T) {
			t.Parallel()

			pruefeSchluesselMit(klagenAn{t}, sprache,
				dienstangabenblock(t, sprache), echt)
		})
	}
}

// echteDienstangaben gibt jedes wirkliche Paar aus Dienst und Angabe.
func echteDienstangaben() map[string]bool {
	echt := map[string]bool{}

	for _, dienst := range domain.Dienste() {
		for _, angabe := range dienst.Zugangsangaben() {
			echt[string(dienst)+"."+angabe] = true
		}
	}

	return echt
}

// pruefeSchluesselMit ist die Regel selbst — siehe [pruefeGeteilteMit].
func pruefeSchluesselMit(
	an klagend, sprache string, jeDienst, echt map[string]bool,
) {
	for schluessel := range jeDienst {
		if !echt[schluessel] {
			an.klage(fmt.Sprintf(
				"im Katalog %s steht unter `dienstangaben` der Schlüssel "+
					"%q — dazu gibt es keinen Dienst mit dieser Angabe. Im "+
					"Fenster fiele er still auf den Rückfall zurück",
				sprache, schluessel))
		}
	}
}

// **Und die Aufzählung selbst nennt nur Angabearten, die es gibt.**
//
// **Die dritte Richtung** (§10): eine Aufzählung, deren Einträge ins Leere
// zeigen, entschärft den Lauf darüber, ohne dass jemand es sieht — wer
// `domain.ZugangSchluessel` umbenennt, schaltet die Zusage still ab.
func TestGeteilteAngabenNenntNurWirklicheAngabearten(t *testing.T) {
	t.Parallel()

	echt := map[string]bool{}

	for _, dienst := range domain.Dienste() {
		for _, angabe := range dienst.Zugangsangaben() {
			echt[angabe] = true
		}
	}

	for angabe, grund := range geteilteAngaben {
		if !echt[angabe] {
			t.Errorf("`geteilteAngaben` nennt %q — diese Angabeart benutzt "+
				"kein Dienst; der Eintrag entschärft nichts und sieht aus, "+
				"als täte er es", angabe)
		}

		if strings.TrimSpace(grund) == "" {
			t.Errorf("`geteilteAngaben` nennt %q ohne Grund — ein Eintrag "+
				"ohne ihn ist eine Ausnahme, die niemand nachprüft", angabe)
		}
	}
}

// dienstangabenblock liest die Namen aus dem Block `dienstangaben`.
//
// **Der Block und nicht die ganze Datei** — dieselbe Überlegung wie bei
// [angabenblock]: `'mtgjson.marke'` stünde sonst womöglich anderswo, und der
// Wächter wäre grün, ohne dass ihn jemand als Beschriftung eingetragen hat.
func dienstangabenblock(t *testing.T, sprache string) map[string]bool {
	t.Helper()

	return dienstangabenAus(t, zugangskatalog(t, sprache), sprache)
}

// zugangskatalog liest die Datei, in der die Zugangstexte stehen.
//
// **`einstellungenzugaenge.ts` und nicht `seiteneinstellungen.ts`**: dort
// standen sie bis zum 2026-09-12, dann ging die Datei über die Zeilengrenze
// aus §2 (#1534). **Ein stilles Leerlesen gibt es nicht** — wer den Block
// woandershin verschiebt, bekommt das `t.Fatalf` der beiden Leser darunter
// und nicht einen grünen Lauf über null Einträge.
func zugangskatalog(t *testing.T, sprache string) string {
	t.Helper()

	roh, err := os.ReadFile(filepath.Join( //nolint:gosec // ein fester Pfad im Repositorium
		"..", "..", "frontend", "src", "sprache", sprache,
		"einstellungenzugaenge.ts"))
	if err != nil {
		t.Fatalf("den Zugangskatalog %s lesen: %v", sprache, err)
	}

	return string(roh)
}

// dienstangabenAus liest den Block aus einem Katalogtext.
//
// **Getrennt vom Lesen der Datei**, damit die Gegenprobe ihn mit einem
// erfundenen Katalog füttern kann — ohne eine Datei im Baum zu ändern.
func dienstangabenAus(t *testing.T, katalog, sprache string) map[string]bool {
	t.Helper()

	block := regexp.MustCompile(`(?s)dienstangaben: \{(.*?)\}`).
		FindStringSubmatch(katalog)
	if block == nil {
		t.Fatalf("im Katalog %s steht kein Block `dienstangaben: {…}`", sprache)
	}

	gefunden := map[string]bool{}

	for _, zeile := range regexp.MustCompile(`(?m)^\s*'([^']+)':`).
		FindAllStringSubmatch(block[1], -1) {
		gefunden[zeile[1]] = true
	}

	return gefunden
}
