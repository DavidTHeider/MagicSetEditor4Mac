package msescript_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// netzAus zerteilt eine Reihe benannter Skripte.
func netzAus(t *testing.T, quellen map[string]string, karte map[string]msescript.Wert) *msescript.Netz {
	t.Helper()

	skripte := map[string]*msescript.Knoten{}

	for name, quelle := range quellen {
		knoten, err := msescript.Zerteile(quelle)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}

		skripte[name] = knoten
	}

	umgebung := msescript.Bibliothek()
	for name, wert := range karte {
		umgebung.Setze(name, wert)
	}

	return msescript.NeuesNetz(skripte, umgebung)
}

// kartenwert baut eine Abbildung, wie sie als `card` hereinkommt.
func kartenwert(felder map[string]string) msescript.Wert {
	wert := msescript.Wert{Art: msescript.ArtAbbildung}
	for name, text := range felder {
		wert.Eintraege = append(wert.Eintraege,
			msescript.Eintrag{Name: name, Wert: msescript.Text(text)})
	}

	return wert
}

// Ein Feldskript sieht die Karte und merkt sich, was es gelesen hat.
func TestEinFeldskriptSiehtDieKarteUndMerktSichWasEsLiest(t *testing.T) {
	netz := netzAus(t, map[string]string{
		"rahmen": `if card.rarity == "rare" then "gold" else "silber"`,
		"breite": `length(remove_tags(card.type_1)) * 17 + 2`,
	}, map[string]msescript.Wert{
		"card": kartenwert(map[string]string{"rarity": "rare", "type_1": "Drache"}),
	})

	wert, err := netz.Wert(t.Context(), "rahmen")
	if err != nil {
		t.Fatalf("abgewiesen: %v", err)
	}

	if wert.String() != `"gold"` {
		t.Errorf("%s", wert.String())
	}

	// **`card.rarity`, nicht `card`.** Wer nur die Abbildung protokollierte,
	// rechnete bei jeder Änderung an irgendeinem Feld alles neu.
	if gelesen := strings.Join(netz.Liest("rahmen"), ", "); gelesen != "card, card.rarity" {
		t.Errorf("gelesen: %q", gelesen)
	}
}

// Eine Änderung rechnet **nur** die abhängigen Felder neu.
func TestEineAenderungBetrifftNurDieAbhaengigenFelder(t *testing.T) {
	netz := netzAus(t, map[string]string{
		"seltenheit": `card.rarity`,
		"rahmen":     `if seltenheit() == "rare" then "gold" else "silber"`,
		"stempel":    `if rahmen() == "gold" then "geprägt" else "glatt"`,
		"breite":     `length(card.type_1)`,
	}, map[string]msescript.Wert{
		"card": kartenwert(map[string]string{"rarity": "rare", "type_1": "Drache"}),
	})

	for _, name := range []string{"rahmen", "stempel", "breite"} {
		if _, err := netz.Wert(t.Context(), name); err != nil {
			t.Fatalf("%s: %v", name, err)
		}
	}

	// `stempel` hängt über `rahmen` an `seltenheit` — drei Ecken weit.
	betroffen := netz.Geaendert("seltenheit")

	if gebaut := strings.Join(betroffen, ", "); gebaut != "rahmen, stempel" {
		t.Errorf("neu zu rechnen: %q — erwartet waren rahmen und stempel", gebaut)
	}

	// Und `breite` steht noch, ohne gerechnet worden zu sein.
	if netz.Liest("breite") == nil {
		t.Error("breite wurde mitverworfen, obwohl es nichts damit zu tun hat")
	}
}

// Ein Kreis wird benannt, statt endlos zu laufen.
//
// „a hängt von b ab, b von a" ist zu finden; „irgendwo im Set ist ein Kreis"
// nicht.
func TestEinKreisZwischenFeldernWirdBenannt(t *testing.T) {
	netz := netzAus(t, map[string]string{
		"a": `b()`,
		"b": `c()`,
		"c": `a()`,
	}, nil)

	_, err := netz.Wert(t.Context(), "a")

	var befund *msescript.Fehler
	if !errors.As(err, &befund) {
		t.Fatalf("Fehler %v, erwartet war ein *msescript.Fehler", err)
	}

	for _, teil := range []string{"a → b → c → a"} {
		if !strings.Contains(befund.Meldung, teil) {
			t.Errorf("die Meldung %q nennt den Kreis nicht", befund.Meldung)
		}
	}
}

// Ein Feld, das sich selbst liest, ist der kürzeste Kreis.
func TestEinFeldDasSichSelbstLiest(t *testing.T) {
	netz := netzAus(t, map[string]string{"a": `a()`}, nil)

	if _, err := netz.Wert(t.Context(), "a"); err == nil {
		t.Error("ein Feld, das sich selbst liest, lief durch")
	}
}

// Ein gerechneter Wert bleibt liegen, bis ihn jemand verwirft.
//
// Ohne das rechnete eine Vorschau bei jedem Tastendruck jedes Feld neu.
func TestEinGerechneterWertBleibtLiegen(t *testing.T) {
	netz := netzAus(t, map[string]string{"a": `1`}, nil)

	for lauf := 1; lauf <= 2; lauf++ {
		if _, err := netz.Wert(t.Context(), "a"); err != nil {
			t.Fatalf("Lauf %d: %v", lauf, err)
		}
	}

	if netz.Liest("a") == nil {
		t.Error("der Wert wurde nicht behalten")
	}

	netz.Geaendert("a")

	if netz.Liest("a") != nil {
		t.Error("der Wert blieb trotz Änderung liegen")
	}
}

// Ein Name, den es nicht gibt, ist ein Fehler und kein leerer Wert.
func TestEinUnbekanntesSkriptWirdGemeldet(t *testing.T) {
	netz := netzAus(t, map[string]string{"a": `1`}, nil)

	if _, err := netz.Wert(t.Context(), "gibtsnicht"); err == nil {
		t.Error("ein unbekanntes Skript ergab einen Wert")
	}
}

// Ein Fehler im Skript reist durch das Netz nach oben.
func TestEinFehlerImSkriptKommtDurchsNetz(t *testing.T) {
	netz := netzAus(t, map[string]string{"a": `gibtsnicht`}, nil)

	if _, err := netz.Wert(t.Context(), "a"); err == nil {
		t.Error("ein unbekannter Name im Skript kam durch")
	}
}

// Ohne eigene Umgebung nimmt das Netz die Bibliothek.
func TestEinNetzOhneUmgebungNimmtDieBibliothek(t *testing.T) {
	knoten, err := msescript.Zerteile(`to_lower("A")`)
	if err != nil {
		t.Fatalf("zerteilen: %v", err)
	}

	netz := msescript.NeuesNetz(map[string]*msescript.Knoten{"a": knoten}, nil)

	wert, err := netz.Wert(t.Context(), "a")
	if err != nil || wert.String() != `"a"` {
		t.Errorf("%v / %v", wert, err)
	}
}

// Der Behälter rechnet wirklich nur einmal.
//
// Der erste Anlauf dieses Tests rief `Wert` zweimal und prüfte den Fehler und
// `Liest` — beides sähe bei doppelter Rechnung genauso aus. Er war grün, ohne
// etwas zu zeigen. Gefunden vom Zweitblick über Phase 3.
func TestDerBehaelterRechnetNurEinmal(t *testing.T) {
	laeufe := 0

	umgebung := msescript.Bibliothek()
	umgebung.Setze("zaehle", msescript.Wert{
		Art: msescript.ArtFunktion,
		Funktion: &msescript.Funktionsteil{
			Eingebaut: func(_ context.Context, _ msescript.Ruf) (msescript.Wert, error) {
				laeufe++

				return msescript.Zahl(float64(laeufe)), nil
			},
		},
	})

	knoten, err := msescript.Zerteile(`zaehle()`)
	if err != nil {
		t.Fatalf("zerteilen: %v", err)
	}

	netz := msescript.NeuesNetz(map[string]*msescript.Knoten{"a": knoten}, umgebung)

	for range 3 {
		if _, err := netz.Wert(t.Context(), "a"); err != nil {
			t.Fatalf("rechnen: %v", err)
		}
	}

	if laeufe != 1 {
		t.Errorf("%d Läufe, erwartet war einer — der Behälter greift nicht", laeufe)
	}

	netz.Geaendert("a")

	if _, err := netz.Wert(t.Context(), "a"); err != nil {
		t.Fatalf("rechnen: %v", err)
	}

	if laeufe != 2 {
		t.Errorf("%d Läufe, erwartet waren zwei — nach einer Änderung wird neu gerechnet", laeufe)
	}
}

// Ein Netzskript sieht die Namen des Aufrufers **nicht**.
//
// Das ist der Unterschied zur geschweiften Klammer, und er ist die Bedingung
// dafür, dass ein Wert liegen bleiben darf: der Behälter ist nach dem Namen
// geführt, und ein Wert, der von den Namen des Aufrufers abhinge, wäre für den
// nächsten Aufrufer falsch.
func TestEinNetzskriptSiehtDieNamenDesAufrufersNicht(t *testing.T) {
	netz := netzAus(t, map[string]string{
		"liest": `eigen`,
		"ruft":  "eigen := \"x\"\nliest()",
	}, nil)

	if _, err := netz.Wert(t.Context(), "ruft"); err == nil {
		t.Error("das Netzskript sah einen Namen des Aufrufers — dann ist der Behälter falsch")
	}
}
