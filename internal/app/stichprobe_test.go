//go:build fremdpakete

// **Wie groß muss eine Stichprobe sein, damit sie etwas sagt?**
//
// Die Läufe hinter `fremdpakete` messen an einer Auswahl: 25.148 Karten liegen
// im Kanon, und jede zu zeichnen kostet gut eine Viertelsekunde. Bis zum
// 2026-08-18 stand dort eine Zahl ohne Aussage — „25 je Set", weil das schnell
// geht.
//
// # Was eine fehlerfreie Stichprobe belegt
//
// Werden `n` Karten gezogen und keine liefert einen Befund, dann liegt die
// Befundquote `p` mit Vertrauen `c` unter
//
//	p = 1 − (1 − c)^(1/n)
//
// umgestellt: `n = ln(1 − c) / ln(1 − p)`. Das ist die übliche Schranke für
// ein Bernoulli-Experiment ohne Treffer.
//
//	Vertrauen | Quote unter |      n
//	    95 %  |       1 %   |    299
//	    99 %  |       1 %   |    459
//	  99,9 %  |       5 %   |    135
//	  99,9 %  |       1 %   |    688
//	  99,9 %  |     0,5 %   |  1.379
//	  99,9 %  |     0,1 %   |  6.905
//
// Bei einer **endlichen** Grundgesamtheit von 25.148 sinkt das leicht: 670 für
// 1 %, 1.308 für 0,5 %, 5.418 für 0,1 %.
//
// # Und was sie ausdrücklich **nicht** belegt
//
// **Sie sagt nichts über Richtigkeit, sondern über Befunde.** Eine Karte kann
// still falsch sein und keinen Befund liefern — genau das ist an diesem Tag
// dreimal passiert: die Maske, die nicht skaliert wurde
// (#471),
// `set_combine`, das seine Mischart wegwarf
// (#470), und
// die Saga, deren Textmaske unbeschnitten blieb
// (#473).
// Alle drei zeichneten befundfrei.
//
// **Wer „99,9 % sicher, dass die Portierung passt" sagen will, sagt zu viel.**
// Was sich sagen lässt: *mit 99,9 % Vertrauen liefern weniger als 1 % der
// Karten einen Befund* — und daneben der Abstand zu MSE2, der eine ganz andere
// Größe misst.
package app

import (
	"math"
	"os"
	"strconv"
	"strings"
	"testing"
)

// stichprobengroesse gibt die Zahl der Karten je Set.
//
// **Vorgabe ist die 99,9-Prozent-Schranke für 1 %**, über die Zahl der Sets
// verteilt und aufgerundet. `MSE_STICHPROBE` hebt sie an — `alle` nimmt jede
// Karte.
//
// # Was diese Zahl trägt und was nicht
//
// **Sie trägt die Aussage über die Vereinigung aller Sets, nicht über eines.**
// Der Bedarf wird für 25.148 Karten gerechnet und dann **geteilt**; jedes Set
// bekommt rund 68 Karten, und 68 aus 3.788 tragen keine 99,9 Prozent.
//
// Wer eine Aussage **je Set** will, nimmt [stichprobeFuer] — der Grund steht
// dort.
func stichprobengroesse(t *testing.T, sets int) int {
	t.Helper()

	roh := os.Getenv("MSE_STICHPROBE")

	switch roh {
	case "":
		return kohortengroesse(25148)/max(1, sets) + 1

	case "alle":
		return math.MaxInt32
	}

	zahl, err := strconv.Atoi(roh)
	if err != nil || zahl <= 0 {
		t.Fatalf("MSE_STICHPROBE ist %q — erwartet eine Zahl über null oder `alle`", roh)
	}

	return zahl
}

// stichprobeFuer gibt die Zahl der Karten für **eine** Grundgesamtheit.
//
// # Warum nicht die geteilte Zahl von oben
//
// **Weil eine geteilte Stichprobe eine geteilte Aussage trägt** — auf
// Anweisung des Auftraggebers vom 2026-09-07.
//
// `stichprobengroesse` rechnet den Bedarf für **alle** 25.148 Karten und
// verteilt ihn: rund 68 je Set. Damit gilt „99,9 Prozent Vertrauen, jeden
// Fehler zu treffen, der 1 Prozent betrifft" für den **Topf**, in den alle
// zehn Sets geschüttet sind. Für ein einzelnes Set mit 3.788 Karten tragen 68
// Karten nichts dergleichen: die Wahrscheinlichkeit, einen Fehler in 1 Prozent
// davon zu treffen, liegt bei rund 50 Prozent.
//
// **Ein Satz über „jedes Set" braucht die Schranke je Set.** Aus 3.788 Karten
// sind das 582 statt 68.
//
// # Die Rechnung
//
//	n∞ = ln(1 − Vertrauen) / ln(1 − Quote)
//	n  = n∞ / (1 + (n∞ − 1) / N)
//
// Die zweite Zeile ist die Korrektur für eine **endliche** Grundgesamtheit:
// aus 107 Karten zu ziehen ist genauer, als aus unendlich vielen zu ziehen.
// Bei einem Set, das kleiner ist als der Bedarf, kommt jede Karte dran.
//
// **`MSE_STICHPROBE` sticht auch hier**, damit ein schneller Lauf möglich
// bleibt — er sagt dann aber nichts über „jedes Set".
func stichprobeFuer(t *testing.T, karten int) int {
	t.Helper()

	if roh := os.Getenv("MSE_STICHPROBE"); roh != "" {
		return stichprobengroesse(t, 1)
	}

	if karten <= 0 {
		return 0
	}

	return min(kohortengroesse(karten), karten)
}

// zaehleSets zählt die `.mse-set`-Dateien in einem Verzeichnis.
func zaehleSets(eintraege []os.DirEntry) int {
	var sets int

	for _, eintrag := range eintraege {
		if strings.HasSuffix(eintrag.Name(), ".mse-set") {
			sets++
		}
	}

	return sets
}

// nenneReichweite sagt, was die gezogene Stichprobe belegt.
//
// **Bei einem Befund gilt die Schranke nicht mehr.** Die Rechnung setzt einen
// fehlerfreien Lauf voraus; sonst wäre die Quote zu schätzen und nicht zu
// beschränken, und dafür braucht es ein anderes Verfahren.
func nenneReichweite(t *testing.T, karten, befunde int) {
	t.Helper()

	if karten == 0 {
		return
	}

	if befunde > 0 {
		t.Logf("  %d Befunde bei %d Karten — die Schranke unten setzt einen "+
			"befundfreien Lauf voraus und gilt hier nicht", befunde, karten)

		return
	}

	for _, vertrauen := range []float64{0.95, 0.99, 0.999} {
		quote := 1 - math.Pow(1-vertrauen, 1/float64(karten))
		t.Logf("  mit %.1f %% Vertrauen liefern weniger als %.2f %% der Karten "+
			"einen Befund", vertrauen*100, quote*100)
	}

	t.Log("  **Das ist eine Aussage über Befunde, nicht über Richtigkeit.** " +
		"Eine still falsche Karte liefert keinen.")
}

// **Die Schranke je Set ist echt strenger als die geteilte** — auf Anweisung
// des Auftraggebers vom 2026-09-07.
//
// # Was dieser Test festhält
//
// Nicht eine Zahl, sondern die **Aussage**: aus 3.788 Karten müssen deutlich
// mehr gezogen werden als die rund 68, die die geteilte Rechnung gibt. Eine
// feste Zahl hier wäre ein zweiter Ort für dieselbe Formel; geprüft wird die
// Eigenschaft, um derentwillen es die Funktion gibt.
func TestDieSchrankeJeSetIstStrengerAlsDieGeteilte(t *testing.T) {
	t.Setenv("MSE_STICHPROBE", "")

	// **Das größte kanonische Set** — 3.788 Karten.
	jeSet := stichprobeFuer(t, 3788)
	geteilt := stichprobengroesse(t, 10)

	if jeSet <= geteilt {
		t.Errorf("je Set %d, geteilt %d — die Aufteilung wäre nicht "+
			"strenger, und die Zusage für ein einzelnes Set trüge nicht",
			jeSet, geteilt)
	}

	// **Und sie liegt unter der Vollerhebung**: eine Stichprobe, die alles
	// nimmt, ist keine.
	if jeSet >= 3788 {
		t.Errorf("je Set %d von 3788 — das ist keine Stichprobe mehr", jeSet)
	}
}

// **Ein kleines Set wird fast ganz gemessen — und nie mehr, als es hat.**
//
// # Was hier stand und warum es falsch war
//
// „Ein Set, das kleiner ist als die Schranke, wird **ganz** gemessen" — und
// dieser Test verlangte 107 von 107. Es sind **93**, und die Formel hat recht:
// die Korrektur für eine endliche Grundgesamtheit zieht 687 auf 93 herunter,
// weil aus 107 Karten zu ziehen genauer ist als aus unendlich vielen. Eine
// Vollerhebung wäre nicht nötig, um 99,9 Prozent zu tragen.
//
// **Der Test hat die Annahme korrigiert, nicht die Rechnung** (§10).
func TestEinKleinesSetWirdFastGanzGemessen(t *testing.T) {
	t.Setenv("MSE_STICHPROBE", "")

	// **Vanguard** hat 107 Karten.
	jeSet := stichprobeFuer(t, 107)

	if jeSet > 107 {
		t.Errorf("aus 107 Karten wurden %d gezogen — mehr, als es gibt", jeSet)
	}

	// **Der überwiegende Teil**: bei so kleinen Grundgesamtheiten bleibt von
	// der Schranke wenig übrig, und das ist der Sinn der Korrektur.
	if jeSet < 80 {
		t.Errorf("aus 107 Karten wurden nur %d gezogen — für 99,9 Prozent "+
			"bei einem Prozent ist das zu wenig", jeSet)
	}
}

// **Und ohne Karten wird nichts gezogen** — statt einer Division durch null.
func TestOhneKartenWirdNichtsGezogen(t *testing.T) {
	t.Setenv("MSE_STICHPROBE", "")

	if jeSet := stichprobeFuer(t, 0); jeSet != 0 {
		t.Errorf("aus null Karten wurden %d gezogen", jeSet)
	}
}
