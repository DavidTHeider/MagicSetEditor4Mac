package msescript_test

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// `match` — die meistgerufene Funktion der bezogenen Pakete
// (#214,
// #216).
//
// # Warum sie die wichtigste war
//
// `magic.mse-game` ruft sie **616-mal**, und der erste Fehler nach dem
// vollständigen Zerteilen des Initskripts war ihrer: `der Name "match" ist
// hier nicht bekannt`, Zeile 44. Drei der neunzehn Exportvorlagen hängen
// ebenfalls an ihr.
func TestMatchGibtEinenWahrheitswert(t *testing.T) {
	faelle := []struct {
		name    string
		quelle  string
		erwartt bool
	}{
		{"ein Treffer", `match("Feuerball", match: "ball")`, true},
		{"kein Treffer", `match("Feuerball", match: "wasser")`, false},

		// **`(?i)` ist die häufigste Form überhaupt**: 557 der 1645
		// verschiedenen Muster der bezogenen Pakete beginnen damit.
		{"ohne Rücksicht auf Groß und Klein", `match("Artefakt", match: "(?i)artefakt")`, true},

		// `match(param1.value, match: " land$")` aus `magic.mse-game`.
		{"ein Anker am Ende", `match("Basic Land", match: " Land$")`, true},
		{"der Anker greift", `match("Land Basic", match: " Land$")`, false},

		// **Die leere Eingabe ist kein Sonderfall.** Ein Feld ohne Wert ist der
		// Normalfall einer frischen Karte, und `match@(…)` läuft über jedes.
		{"leere Eingabe", `match("", match: "x")`, false},
		{"leeres Muster trifft alles", `match("irgendwas", match: "")`, true},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			wert, err := msescript.AuswertenQuelle(t.Context(), fall.quelle,
				msescript.Bibliothek())
			if err != nil {
				t.Fatalf("auswerten: %v", err)
			}

			if wert.Art != msescript.ArtWahrheitswert {
				t.Fatalf("Art %s — erwartet war ein Wahrheitswert", wert.Art)
			}

			if wert.Wahrheit != fall.erwartt {
				t.Errorf("%v, erwartet war %v", wert.Wahrheit, fall.erwartt)
			}
		})
	}
}

// **Teilweise angewandt, denn so steht sie fast immer da.**
//
// 606 der 616 Vorkommen in `magic.mse-game` sind `match@(match: "…")` — ein
// Prädikat, das später einen Text bekommt.
func TestMatchLaesstSichTeilweiseAnwenden(t *testing.T) {
	wert, err := msescript.AuswertenQuelle(t.Context(),
		`(match@(match: "(?i)^land"))("Landschaft")`, msescript.Bibliothek())
	if err != nil {
		t.Fatalf("auswerten: %v", err)
	}

	if wert.Art != msescript.ArtWahrheitswert || !wert.Wahrheit {
		t.Errorf("%+v", wert)
	}
}

// **Ein unbrauchbares Muster wird benannt, nicht verschluckt** (§8).
//
// # Warum das hier besonders zählt
//
// Ein stillschweigendes „kein Treffer" für ein Muster, das keine der beiden
// Maschinen liest, erzeugte eine Ausfuhr, die aussieht wie eine richtige.
func TestEinUnbrauchbaresMusterWirdBenannt(t *testing.T) {
	// **Ein Muster, das in jeder Sprache falsch ist, bricht ab.** Ein
	// Vorausblick dagegen läuft — er geht über die Rückfalllinie ([Muster]),
	// geprüft in `muster_test.go`.
	_, err := msescript.AuswertenQuelle(t.Context(),
		`match("x", match: "[")`, msescript.Bibliothek())
	if err == nil {
		t.Fatal("kein Fehler — ein kaputtes Muster ging stillschweigend durch")
	}

	// Die Meldung nennt das Muster. Ohne das steht in einem Lauf über
	// dreißig Module „ein Muster war unbrauchbar" und niemand weiß, welches.
	if !strings.Contains(err.Error(), "[") {
		t.Errorf("die Meldung nennt das Muster nicht: %v", err)
	}
}

// **Ohne `match:` ist es kein Aufruf, sondern ein Tippfehler.**
func TestMatchOhneMusterWirdAbgewiesen(t *testing.T) {
	_, err := msescript.AuswertenQuelle(t.Context(),
		`match("x")`, msescript.Bibliothek())
	if err == nil {
		t.Fatal("kein Fehler")
	}
}

// **`replace:` darf eine Funktion sein — dann wird sie je Treffer gerufen.**
//
//	if (replacement_function) {
//	  for (UInt sub = 0 ; sub < results.size() ; ++sub) {
//	    String name = sub == 0 ? _("input") : String(_("_")) << sub;
//	    ctx.setVariable(name, to_script(results.str(sub)));
//	  }
//	  inside = replacement_function->eval(ctx)->toString();
//	}
//
// `RegexReplacer::apply`, `src/script/functions/regex.cpp`, Zeilen 78–99,
// Fassung f9d9356 vom 2021-07-06.
//
// **Der ganze Treffer heißt `input`, die Gruppen `_1` bis `_9`.**
//
// # Was das gekostet hat
//
// `magic.mse-game` baut damit seine Typzeile und den Regeltext um. Ohne diesen
// Weg wurde die Funktion in Text verwandelt, und auf *Firebrand Incinerator*
// stand `Creature — Goblinfunktion<class-en>…` statt `Creature — Goblin Mage`
// und `deal Funktion damage` statt `deal 2 damage`
// (#345).
func TestReplaceDarfEineFunktionSein(t *testing.T) {
	faelle := []struct{ name, quelle, erwartet string }{
		{
			name:     "der ganze Treffer heißt input",
			quelle:   `replace("abc", match: "b", replace: { "[" + input + "]" })`,
			erwartet: "a[b]c",
		},
		{
			name:     "die Gruppen heißen _1 und _2",
			quelle:   `replace("ab", match: "(a)(b)", replace: { _2 + _1 })`,
			erwartet: "ba",
		},
		{
			name:     "jeder Treffer wird einzeln gerufen",
			quelle:   `replace("a1b2", match: "[0-9]", replace: { to_string(to_int(input) + 1) })`,
			erwartet: "a2b3",
		},
		{
			// **Eine Zeichenkette bleibt eine Zeichenkette** — die alte Form
			// mit `\1` gilt weiter.
			name:     "ein Text bleibt ein Text",
			quelle:   `replace("ab", match: "(a)(b)", replace: "\\2\\1")`,
			erwartet: "ba",
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			wert, err := msescript.AuswertenQuelle(t.Context(), fall.quelle, msescript.Bibliothek())
			if err != nil {
				t.Fatalf("ausrechnen: %v", err)
			}

			if wert.AlsText() != fall.erwartet {
				t.Errorf("ergibt %q, erwartet %q", wert.AlsText(), fall.erwartet)
			}
		})
	}
}

// **`recursive:` wendet die Ersetzung auf ihr eigenes Ergebnis an.**
//
//	if (recursive && level < 20) {
//	  ret += apply(ctx, inside, level + 1);
//	}
//
// Zeilen 100–104. **Zwanzig Ebenen tief**, dann bricht MSE ab und nimmt das
// Ergebnis, wie es ist — keine Meldung, keine Schleife.
func TestRecursiveWendetDieErsetzungAufSichSelbstAn(t *testing.T) {
	wert, err := msescript.AuswertenQuelle(t.Context(),
		`replace("aaa", match: "a", replace: "b", recursive: true)`, msescript.Bibliothek())
	if err != nil {
		t.Fatalf("ausrechnen: %v", err)
	}

	if wert.AlsText() != "bbb" {
		t.Errorf("ergibt %q, erwartet %q", wert.AlsText(), "bbb")
	}

	// **Eine Ersetzung, die sich selbst wieder trifft, endet an der Grenze.**
	tiefer, err := msescript.AuswertenQuelle(t.Context(),
		`replace("a", match: "a", replace: "aa", recursive: true)`, msescript.Bibliothek())
	if err != nil {
		t.Fatalf("ausrechnen: %v", err)
	}

	if len(tiefer.AlsText()) < 2 {
		t.Errorf("ergibt %q — die Grenze hat mehr weggenommen als sie soll", tiefer.AlsText())
	}
}

// **Was die Ersatzfunktion nicht kann, wird gemeldet.**
//
// Kein halb ersetzter Text: MSE lässt den Fehler aus dem Rumpf durch, und
// `Value::update` fängt ihn je Feld (§8). Ein Text, dem in der Mitte etwas
// fehlt, wäre schlimmer als eine Meldung.
func TestEinFehlerInDerErsatzfunktionWirdGemeldet(t *testing.T) {
	faelle := map[string]string{
		"die Funktion scheitert":   `replace("ab", match: "a", replace: { nirgends_bekannt })`,
		"auch rekursiv":            `replace("ab", match: "a", replace: { nirgends_bekannt }, recursive: true)`,
		"und `replace:` ist keine": `replace("ab", match: "a", replace: 5, recursive: true)`,
	}

	for name, quelle := range faelle {
		t.Run(name, func(t *testing.T) {
			wert, err := msescript.AuswertenQuelle(t.Context(), quelle, msescript.Bibliothek())

			// **Der dritte Fall ist kein Fehler**: eine Zahl ist keine
			// Funktion, sie wird als Text eingesetzt — auch rekursiv.
			if name == "und `replace:` ist keine" {
				if err != nil || wert.AlsText() != "5b" {
					t.Errorf("ergibt %q / %v, erwartet \"5b\"", wert.AlsText(), err)
				}

				return
			}

			if err == nil {
				t.Errorf("ergibt %q statt eines Fehlers", wert.AlsText())
			}
		})
	}
}

// **Ein Muster, das RE2 nicht übersetzt, nimmt denselben Weg.**
//
// Rückwärtsverweise kennt RE2 nicht; dafür steht `regexp2` bereit
// (#240).
// Die Ersatzfunktion muss auch dort je Treffer laufen — und die Stellen
// zählen dann in **Zeichen** statt in Bytes.
func TestEineErsatzfunktionLaeuftAuchUeberDieRueckfalllinie(t *testing.T) {
	wert, err := msescript.AuswertenQuelle(t.Context(),
		`replace("Ötzi aa bb", match: "(\\w)\\1", replace: { "[" + input + "]" })`,
		msescript.Bibliothek())
	if err != nil {
		t.Fatalf("ausrechnen: %v", err)
	}

	if wert.AlsText() != "Ötzi [aa] [bb]" {
		t.Errorf("ergibt %q, erwartet %q", wert.AlsText(), "Ötzi [aa] [bb]")
	}
}

// **Eine Gruppe, die nicht getroffen hat, ist leer und kein Fehler.**
func TestEineNichtGetroffeneGruppeIstLeer(t *testing.T) {
	wert, err := msescript.AuswertenQuelle(t.Context(),
		`replace("ab", match: "a(x)?", replace: { "[" + _1 + "]" })`, msescript.Bibliothek())
	if err != nil {
		t.Fatalf("ausrechnen: %v", err)
	}

	if wert.AlsText() != "[]b" {
		t.Errorf("ergibt %q, erwartet %q", wert.AlsText(), "[]b")
	}
}

// **Nach dem ersten Fehler wird nicht weiterersetzt.**
//
// Zwei Treffer, und der erste scheitert: der zweite Ruf unterbleibt. Ein
// halber Text mit einer Meldung daneben wäre schlimmer als kein Text.
func TestNachDemErstenFehlerWirdNichtWeitergemacht(t *testing.T) {
	if _, err := msescript.AuswertenQuelle(t.Context(),
		`replace("aa", match: "a", replace: { nirgends_bekannt })`,
		msescript.Bibliothek()); err == nil {
		t.Error("erwartet einen Fehler")
	}
}

// **`RufeFunktion` weist zurück, was keine Funktion ist.**
//
// Sie steht eingebauten Funktionen offen — `replace` ruft sie —, und was von
// dort kommt, ist ein Wert aus einer Vorlage. Die Prüfung gehört deshalb
// hierher und nicht in den Aufrufer.
func TestRufeFunktionWeistZurueckWasKeineIst(t *testing.T) {
	if _, err := msescript.RufeFunktion(t.Context(),
		msescript.Text("kein Rumpf"), msescript.Bibliothek()); err == nil {
		t.Error("erwartet einen Fehler")
	}
}

// **Eine eingebaute Funktion ohne Umgebung fällt auf die Bibliothek zurück.**
//
// Wer sie von Hand ruft — ein Test, ein Werkzeug —, gibt keine mit. Der Rumpf
// eines übergebenen Blocks sieht dann die Bibliothek und sonst nichts; das ist
// weniger als im Betrieb, aber es ist kein Absturz.
func TestEineEingebauteOhneUmgebungFaelltAufDieBibliothekZurueck(t *testing.T) {
	wert, da := msescript.Bibliothek().Suche("tag_contents")
	if !da || wert.Funktion == nil || wert.Funktion.Eingebaut == nil {
		t.Fatal("tag_contents steht nicht als eingebaute Funktion bereit")
	}

	ergebnis, err := wert.Funktion.Eingebaut(t.Context(), msescript.Ruf{
		Stellung: []msescript.Wert{msescript.Text("<b>x</b>")},
		Benannt: []msescript.Eintrag{
			{Name: "tag", Wert: msescript.Text("<b")},
			{Name: "contents", Wert: msescript.NeueFunktion(zerteiltesX(t))},
		},
	})
	if err != nil {
		t.Fatalf("ausrechnen: %v", err)
	}

	if ergebnis.AlsText() != "<b>y</b>" {
		t.Errorf("ergibt %q, erwartet %q", ergebnis.AlsText(), "<b>y</b>")
	}
}

// zerteiltesX gibt den Rumpf `"y"` als Baum.
func zerteiltesX(t *testing.T) *msescript.Knoten {
	t.Helper()

	baum, err := msescript.Zerteile(`"y"`)
	if err != nil {
		t.Fatalf("zerteilen: %v", err)
	}

	return baum
}
