package msescript_test

// **Die Sprache gegen fremden Quelltext** —
// #1289.
//
// # Warum
//
// MSE-Script kommt aus fremden Paketen. Ein `.mse-style` aus dem Netz bringt
// Skripte mit, die dieses Programm zerteilt **und ausrechnet** — und es tut das
// beim bloßen Öffnen eines Sets, ohne dass jemand etwas bestätigt.
//
// # Was ein Fund ist — festgelegt **vor** dem ersten Lauf
//
//   - Panic
//   - Hänger
//   - zweimal dieselbe Eingabe, zwei Ergebnisse
//   - Erfolg gemeldet, Zustand strukturell ungültig — hier: ein Wert **und**
//     ein Fehler, oder keins von beidem
//
// **Nicht zugesichert**: Laufzeit und Speicher im Verhältnis zur Eingabe. Eine
// Zeitmessung im Fuzzlauf wäre eine Aussage über den Rechner (§10).
//
// # Warum der Auswerter eine Frist bekommt
//
// Weil ein Hänger sonst als Zeitüberschreitung des ganzen Fuzzlaufs erscheint
// und nicht als Fund. `Tiefengrenze` fängt die Rekursion; was sie **nicht**
// fängt, ist eine Bibliotheksfunktion, die über einer erzeugten Liste läuft.
// Die Frist macht daraus einen benannten Abbruch.
//
//	go test ./internal/msescript -run Fuzz -fuzz FuzzZerteileSprache -fuzztime 2m
//	go test ./internal/msescript -run Fuzz -fuzz FuzzAuswerten -fuzztime 2m

import (
	"context"
	"errors"
	"math"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// sprachsaat sind die Gestalten, mit denen der Fuzzer beginnt.
//
// **Aus dem Pack und aus `docs/FORMATE.md`**, nicht ausgedacht: ein
// Funktionswert, ein benannter Aufruf, eine Verkettung, `or else`, ein
// Zugriff über eine Stelle, eine Bedingung.
var sprachsaat = []string{
	"",
	"1",
	`"text"`,
	"a := 1\nb := a + 2",
	"{ card.name }",
	`sort_text@(order: "WUBRG")("U/W")`,

	// **Zwei Körner aus dem Fuzzlauf selbst**, beide über [gleicherWert]:
	// ein nackter Bibliotheksname gibt einen Funktionswert, und `0/0` gibt
	// `nan`. Warum beide keine Funde waren, steht dort.
	"sort_text",
	"0/0",
	`if a then b else c`,
	`a or else "ersatz"`,
	`base_hybrid[shape][color_count]`,
	`to_int("12") + max(3, 4)`,
	`combined_editor(field1: a, field2: b, separator: " // ")`,
	`{ if styling.unsorted_casting_cost then input else mana_sort(input) }`,
	`card["rarity_" + input]`,
	`a := { b := { c := 1 } }`,
	"(((((((((((((((((((((1)))))))))))))))))))))",
}

// **Der Zerteiler bricht nicht ab und antwortet zweimal gleich.**
func FuzzZerteileSprache(f *testing.F) {
	for _, saat := range sprachsaat {
		f.Add(saat)
	}

	f.Fuzz(func(t *testing.T, quelle string) {
		knoten, fehler := msescript.Zerteile(quelle)

		if fehler != nil {
			if knoten != nil {
				t.Errorf("abgelehnt **und** ein Baum zurück: %v", fehler)
			}

			// **Zum Ergebnis gehört auch das Nein** (Nachlauf über Phase 4.2).
			// Eine Ablehnung dieser Sprache nennt die Stelle im Quelltext; sie
			// dürfte nicht wandern.
			nochmal, zweiter := msescript.Zerteile(quelle)
			switch {
			case zweiter == nil:
				t.Errorf("dieselbe Quelle wurde erst abgelehnt (%v) und dann "+
					"angenommen: %+v", fehler, nochmal)

			case zweiter.Error() != fehler.Error():
				t.Errorf("dieselbe Quelle gab zwei Meldungen: %q gegen %q",
					fehler, zweiter)
			}

			return
		}

		if knoten == nil {
			t.Fatal("kein Fehler und kein Baum — das ist keine Antwort")
		}

		// **Zweimal dieselbe Quelle, derselbe Baum.** Der Zerteiler merkt sich
		// Bäume (`zerteilerspeicher.go`); ein Speicher, der etwas anderes
		// zurückgäbe, wäre die schlimmste Art von Fund — er träfe erst beim
		// zweiten Öffnen.
		nochmal, nochmalFehler := msescript.Zerteile(quelle)
		if nochmalFehler != nil {
			t.Fatalf("dieselbe Quelle wurde beim zweiten Mal abgelehnt: %v",
				nochmalFehler)
		}

		if knoten != nochmal {
			t.Error("dieselbe Quelle gab zwei verschiedene Bäume")
		}
	})
}

// **Der Auswerter bricht nicht ab und hängt nicht.**
func FuzzAuswerten(f *testing.F) {
	for _, saat := range sprachsaat {
		f.Add(saat)
	}

	f.Fuzz(func(t *testing.T, quelle string) {
		knoten, fehler := msescript.Zerteile(quelle)
		if fehler != nil {
			// Was der Zerteiler mit einer Ablehnung tut, prüft
			// [FuzzZerteileSprache]; hier geht es um den Auswerter.
			return
		}

		ctx, ende := context.WithTimeout(context.Background(), fuzzfrist)
		defer ende()

		wert, rechenfehler := msescript.Auswerten(ctx, knoten, fuzzumgebung())

		// **Ein Abbruch ist kein Fund.** Er ist die Antwort der Frist, und
		// dass sie greift, ist die Zusage.
		if errors.Is(rechenfehler, context.DeadlineExceeded) {
			return
		}

		if rechenfehler != nil {
			// **Ein Fehler ist erlaubt und ein Wert dann nicht.** Beides
			// zugleich wäre ein halber Stand (§8).
			if wert.Art != msescript.Nichts.Art {
				t.Errorf("Fehler **und** Wert: %v / %+v", rechenfehler, wert.Art)
			}

			auchDasNeinZweimal(t, knoten, rechenfehler)

			return
		}

		// **Ein Fehlerwert ist kein Fehler**, sondern die Art, wie diese
		// Sprache „das ging nicht" sagt — `or else` fängt ihn. Was nicht sein
		// darf, ist eine Wertart, die es nicht gibt.
		if wert.Art == "" {
			t.Errorf("ein Wert ohne Art: %+v", wert)
		}

		nochmalDasselbe(t, knoten, wert)
	})
}

// fuzzfrist ist die Frist, nach der ein Lauf als Hänger gilt.
//
// **Zwei Sekunden, und die Zahl ist großzügig gewählt.** Ein Skript aus einem
// Paket rechnet in Millisekunden; wer zwei Sekunden braucht, hat kein
// Skript, sondern eine Schleife. Enger gesetzt würde die Frist auf einem
// belasteten Rechner zum Fund erklären, was nur langsam war.
const fuzzfrist = 2 * time.Second

// fuzzumgebung gibt die Bibliothek plus eine Karte und ein Set.
//
// **Mit `card` und `set`, weil ein Paketskript sie liest.** Ohne sie führte
// jeder zweite Ausdruck sofort auf „unbekannter Name" und der Fuzzer käme
// nicht an den Auswerter heran.
func fuzzumgebung() *msescript.Umgebung {
	umgebung := msescript.Bibliothek().Darueber()

	umgebung.Setze("card", msescript.Wert{
		Art: msescript.ArtAbbildung,
		Eintraege: []msescript.Eintrag{
			{Name: "name", Wert: msescript.Text("Feuerball")},
			{Name: "rarity", Wert: msescript.Text("common")},
		},
	})

	umgebung.Setze("set", msescript.Wert{
		Art:       msescript.ArtAbbildung,
		Eintraege: []msescript.Eintrag{{Name: "title", Wert: msescript.Text("Probe")}},
	})

	umgebung.Setze("input", msescript.Text("W/U"))

	return umgebung
}

// nochmalDasselbe sichert zu, dass derselbe Baum denselben Wert gibt.
//
// **Der Kopf nannte „zweimal dieselbe Eingabe, zwei Ergebnisse" als Fund, und
// der Rumpf wertete einmal aus.** Angestrichen vom Zweitblick über Phase 4.2.
//
// Ausgewertet wird derselbe **Baum** und nicht derselbe Text: was der Zerteiler
// tut, prüft [FuzzZerteileSprache]. Hier steht die Frage, ob der Auswerter
// einen Zustand über den Lauf hinweg mitschleppt — ein Speicher, eine Uhr, eine
// Zufallsquelle.
//
// **Ein Abbruch beim zweiten Mal ist kein Fund.** Die Frist misst Zeit, und die
// ist zwischen zwei Läufen nicht dieselbe; ein Vergleich darüber wäre eine
// Aussage über den Rechner (§10).
func nochmalDasselbe(t *testing.T, knoten *msescript.Knoten, erster msescript.Wert) {
	t.Helper()

	ctx, ende := context.WithTimeout(context.Background(), fuzzfrist)
	defer ende()

	zweiter, fehler := msescript.Auswerten(ctx, knoten, fuzzumgebung())

	if errors.Is(fehler, context.DeadlineExceeded) {
		return
	}

	if fehler != nil {
		t.Errorf("derselbe Baum wurde beim zweiten Mal abgelehnt: %v", fehler)

		return
	}

	if zweiter.Art != erster.Art {
		t.Errorf("derselbe Baum gab zwei Wertarten: %q und dann %q",
			erster.Art, zweiter.Art)

		return
	}

	if !gleicherWert(erster, zweiter) {
		t.Errorf("derselbe Baum gab zwei Werte der Art %q: %q gegen %q",
			erster.Art, erster.AlsText(), zweiter.AlsText())
	}
}

// gleicherWert vergleicht zwei Werte so, wie diese Sprache sie versteht.
//
// **Warum nicht [reflect.DeepEqual]**, und warum das kein Nachlassen ist: zwei
// Werte dieser Sprache können gleich sein und strukturell verschieden — aus
// zwei Gründen, und beide sind belegtes Verhalten.
//
//   - **Eine Funktion trägt eine Go-Funktion.** `DeepEqual` gibt für zwei davon
//     `false`, auch wenn es dieselbe ist; Go vergleicht Funktionen nicht. Ein
//     Vergleich darüber prüfte eine Eigenschaft der Sprache, in der der
//     Auswerter geschrieben ist, und keine des Auswerters. Verglichen wird
//     deshalb ihr **Dasein**, und alles übrige am Wert wie sonst auch — der
//     erste Anlauf gab bei einer Funktion pauschal „gleich" zurück und ließ
//     damit Art, Feld und Marke ungeprüft (Nachlauf über Phase 4.2).
//   - **Eine Zahl darf `NaN` sein.** `0/0` ergibt hier `nan`, und das ist
//     portiertes Verhalten: MSEs `I_FDIV` rechnet zwei `double` durcheinander
//     und lässt IEEE 754 antworten (`src/script/context.cpp`, Zeilen 485–487,
//     Fassung f9d9356 vom 2021-07-06; am Orakel bestätigt,
//     #555).
//     Und `NaN != NaN` ist dieselbe Norm, nicht ein Fund.
//
// **Beide hat dieser Lauf selbst gefunden** — die verschärfte Zusicherung wurde
// zweimal rot, bei `sort_text` und bei `0/0`, und beide Male lag der Befund in
// der Zusicherung. Die Eingaben stehen seitdem als Saatkörner in [sprachsaat].
func gleicherWert(erster, zweiter msescript.Wert) bool {
	if erster.Art != zweiter.Art {
		return false
	}

	if erster.Zahl != zweiter.Zahl &&
		!(math.IsNaN(erster.Zahl) && math.IsNaN(zweiter.Zahl)) {
		return false
	}

	if erster.Text != zweiter.Text || erster.Wahrheit != zweiter.Wahrheit ||
		erster.Feld != zweiter.Feld || erster.Marke != zweiter.Marke ||
		erster.Verbund != zweiter.Verbund {
		return false
	}

	// **Eine Funktion nur an ihrem Dasein.** Zwei Funktionswerte sind
	// strukturell nie gleich, und ob **dieselbe** Funktion herauskam, ließe
	// sich nur über den Rumpfzeiger sagen — der bei einer eingebauten Funktion
	// leer ist. Was bleibt, ist: entweder beide tragen eine oder keiner.
	if (erster.Funktion == nil) != (zweiter.Funktion == nil) {
		return false
	}

	return gleicheGlieder(erster.Eintraege, zweiter.Eintraege) &&
		gleicheGlieder(erster.Glieder, zweiter.Glieder)
}

// gleicheGlieder vergleicht zwei Listen benannter Werte Stück für Stück.
//
// **Namen **und** Reihenfolge**: eine Abbildung dieser Sprache ist geordnet,
// und `to_text` wirft die Namen weg — ein Vergleich über die Textform ließe
// zwei verschiedene Abbildungen als gleich durchgehen.
//
// **Und `Glieder` genauso wie `Eintraege`.** Ein Wert kann als Liste etwas
// anderes sein als beim Zugriff über einen Namen (`wert.go`, #653); der erste
// Anlauf verglich nur die Einträge und ließ die zweite Hälfte des Werts außen
// vor — angestrichen vom Nachlauf über Phase 4.2.
func gleicheGlieder(erste, zweite []msescript.Eintrag) bool {
	if len(erste) != len(zweite) {
		return false
	}

	for i := range erste {
		if erste[i].Name != zweite[i].Name {
			return false
		}

		if !gleicherWert(erste[i].Wert, zweite[i].Wert) {
			return false
		}
	}

	return true
}

// auchDasNeinZweimal sichert zu, dass eine Ablehnung sich wiederholt.
//
// **Zum Ergebnis gehört auch das Nein** (Nachlauf über Phase 4.2). Der erste
// Anlauf kehrte bei einem Rechenfehler zurück, ohne ein zweites Mal zu rechnen.
//
// **Ein Abbruch auf einer der beiden Seiten ist kein Fund.** Die Frist misst
// Zeit, und die ist zwischen zwei Läufen nicht dieselbe.
func auchDasNeinZweimal(t *testing.T, knoten *msescript.Knoten, erster error) {
	t.Helper()

	if errors.Is(erster, context.DeadlineExceeded) {
		return
	}

	ctx, ende := context.WithTimeout(context.Background(), fuzzfrist)
	defer ende()

	wert, zweiter := msescript.Auswerten(ctx, knoten, fuzzumgebung())

	if errors.Is(zweiter, context.DeadlineExceeded) {
		return
	}

	if zweiter == nil {
		t.Errorf("derselbe Baum wurde erst abgelehnt (%v) und dann gerechnet: %+v",
			erster, wert.Art)

		return
	}

	if zweiter.Error() != erster.Error() {
		t.Errorf("derselbe Baum gab zwei Meldungen: %q gegen %q", erster, zweiter)
	}
}
