// Der Wächter über `idempotentHint` — die Zusage wird **ausgeführt**.
//
// # Der Anlass
//
// `werkzeuge.go` setzte den Hinweis bis zum 2026-08-23 **unbedingt für jedes**
// Werkzeug, und der Kommentar daneben begründete ihn — richtig, aber nur für
// eines:
//
//	Ein zweites `schreiben.feld` mit denselben Argumenten setzt denselben Wert
//	noch einmal: das Ergebnis ist dasselbe.
//
// `schreiben.karte.anlegen` und `schreiben.karte.kopieren` legen bei jedem
// Aufruf eine **neue Karte mit neuer Kennung** an.
//
// **Der Hinweis ist eine Erlaubnis.** Ein Gegenüber, das nach einer
// Zeitüberschreitung wiederholt, darf sich auf ihn verlassen — und genau das
// ist der Fall, für den es ihn gibt: die Anlage kann geglückt sein, während die
// Antwort verlorenging. Es entstünden Doppelkarten
// (#895).
//
// # Warum die Zusage ausgeführt und nicht abgezählt wird
//
// Ein Wächter, der eine Liste von Namen gegen eine zweite Liste hält, prüft
// zwei Aufzählungen gegeneinander — und die zweite schreibt derselbe, der die
// erste ändert. Hier wird das Werkzeug **zweimal gerufen** und die Antwort
// verglichen; das ist die Zusage selbst, nicht ihre Beschriftung (§10).
//
// # Was er nicht prüft
//
//   - **Ob ein Werkzeug ohne den Hinweis wirklich keine Idempotenz hat.** Der
//     Hinweis ist eine Zusage, kein Ausschluss: `false` heißt „verlass dich
//     nicht darauf", nicht „es ist bestimmt anders".
//
//   - **Nebenwirkungen jenseits der Antwort.** Das Protokoll sagt „no
//     additional effect on the its environment"
//     (`schema/2025-06-18/schema.ts`, Zeilen 904–912) — und die Umgebung sieht
//     dieser Wächter nicht: er ruft gegen ein Double und vergleicht Antworten.
//
//     **Für die schreibenden Werkzeuge ist das heute folgenlos**, weil keines
//     die Zusage macht. Käme eines dazu, reichte dieser Wächter nicht: dann
//     gehörte der Fall gegen echtes Storage, mit gezählten Verlaufsschritten
//     vor und nach beiden Rufen
//     (#915).
//
//   - **Den Weg über die Datenbank.** Gerufen wird gegen das Double (§3).
package mcp_test

import (
	"encoding/json"
	"strings"
	"testing"
)

// **Was Idempotenz zusagt, wird zweimal gerufen und verglichen.**
func TestEinWerkzeugMitIdempotenzhinweisGibtZweimalDasselbe(t *testing.T) {
	geprueft := 0

	for _, werkzeug := range werkzeugliste(t) {
		name := zeichenkette(t, werkzeug["name"])

		if !sagtIdempotenzZu(t, werkzeug) {
			continue
		}

		argumente, da := idempotenzfaelle[name]
		if !da {
			t.Errorf("%s sagt Idempotenz zu, und es gibt keinen Fall dafür", name)

			continue
		}

		t.Run(name, func(t *testing.T) {
			// **Derselbe Bestand für beide Rufe.** Ein frischer beim zweiten
			// prüfte nichts: dann wäre jedes Werkzeug idempotent.
			bestand := probenbestand()
			werkzeug := werkzeugMitNamen(t, bestand, name)

			erste, err := werkzeug.Ruf(t.Context(), json.RawMessage(argumente))
			if err != nil {
				t.Fatalf("erster Ruf: %v", err)
			}

			zweite, err := werkzeug.Ruf(t.Context(), json.RawMessage(argumente))
			if err != nil {
				t.Fatalf("zweiter Ruf: %v", err)
			}

			// **Die rohe Antwort und keine zerlegte.** Manche Werkzeuge geben
			// eine Liste, manche einen Verbund; „dasselbe Ergebnis" ist die
			// Zeichenkette, die beim Gegenüber ankommt.
			if erste != zweite {
				t.Errorf("%s sagt Idempotenz zu, gibt aber\n  erst:  %s\n  dann:  %s",
					name, erste, zweite)
			}
		})

		geprueft++
	}

	if geprueft == 0 {
		t.Fatal("kein einziges Werkzeug sagt Idempotenz zu — hat sich der Hinweis bewegt?")
	}
}

// **Die anlegenden Werkzeuge sagen sie ausdrücklich nicht zu.**
//
// Das ist die Gegenrichtung zum Test darüber und hier eigens benannt, weil
// genau diese beiden den Anlass geben.
func TestAnlegenUndKopierenSagenKeineIdempotenzZu(t *testing.T) {
	nichtIdempotent := map[string]bool{
		"schreiben.karte.anlegen":  true,
		"schreiben.karte.kopieren": true,
	}

	gesehen := 0

	for _, werkzeug := range werkzeugliste(t) {
		name := zeichenkette(t, werkzeug["name"])
		if !nichtIdempotent[name] {
			continue
		}

		gesehen++

		if sagtIdempotenzZu(t, werkzeug) {
			t.Errorf("%s sagt Idempotenz zu — es legt bei jedem Ruf eine neue Karte an", name)
		}
	}

	if gesehen != len(nichtIdempotent) {
		t.Errorf("%d von %d anlegenden Werkzeugen gesehen — heißen sie noch so?",
			gesehen, len(nichtIdempotent))
	}
}

// **Jedes Werkzeug hat einen Idempotenzfall**, ob es die Zusage macht oder
// nicht.
//
// Ohne diesen Wächter wäre [idempotenzfaelle] eine Liste, die still zu kurz
// wird: ein neues Werkzeug mit Zusage und ohne Fall bliebe ungeprüft.
func TestJedesWerkzeugHatEinenIdempotenzfall(t *testing.T) {
	namen := map[string]bool{}

	for _, werkzeug := range werkzeugliste(t) {
		name := zeichenkette(t, werkzeug["name"])
		namen[name] = true

		if _, da := idempotenzfaelle[name]; !da {
			t.Errorf("für %s steht kein Idempotenzfall in idempotenzfaelle", name)
		}
	}

	for name := range idempotenzfaelle {
		if !namen[name] {
			t.Errorf("idempotenzfaelle führt %s — das Werkzeug gibt es nicht mehr", name)
		}
	}
}

// sagtIdempotenzZu liest den Hinweis eines Werkzeugs.
//
// **Ein fehlender Hinweis ist ein Befund, kein `false`.** Ein Gegenüber, das
// ihn nicht findet, muss raten — und die Angabe gehört zu jedem Werkzeug, ob
// die Antwort ja oder nein lautet.
func sagtIdempotenzZu(t *testing.T, werkzeug map[string]any) bool {
	t.Helper()

	anmerkungen := abbildung(t, werkzeug["annotations"])

	zusage, istWahrheitswert := anmerkungen["idempotentHint"].(bool)
	if !istWahrheitswert {
		t.Fatalf("%v nennt keinen idempotentHint", werkzeug["name"])
	}

	return zusage
}

// **Kein schreibendes Werkzeug sagt Idempotenz zu.**
//
// Der Wächter darüber prüft, dass eine gemachte Zusage hält. Dieser prüft, dass
// sie gar nicht erst gemacht wird — und zwar aus einem Grund, der für **jedes**
// schreibende Werkzeug gilt: jede Änderung schreibt einen Verlaufsschritt
// (`internal/storage/feldaenderung.go`), damit sie sich zurücknehmen lässt.
// Ein zweiter Schritt ist eine zusätzliche Wirkung auf die Umgebung, und genau
// das schließt `idempotentHint` aus.
//
// **Wer das ändert, tut es mit einem Test gegen echtes Storage**, nicht gegen
// das Double: die Verlaufsschritte sieht nur die Datenbank.
func TestKeinSchreibendesWerkzeugSagtIdempotenzZu(t *testing.T) {
	gesehen := 0

	for _, werkzeug := range werkzeugliste(t) {
		name := zeichenkette(t, werkzeug["name"])
		if !strings.HasPrefix(name, "schreiben.") {
			continue
		}

		gesehen++

		if sagtIdempotenzZu(t, werkzeug) {
			t.Errorf("%s sagt Idempotenz zu — jede Änderung schreibt einen "+
				"Verlaufsschritt, und ein zweiter ist eine zusätzliche Wirkung", name)
		}
	}

	if gesehen == 0 {
		t.Fatal("kein einziges schreibendes Werkzeug — hat sich die Benennung bewegt?")
	}
}
