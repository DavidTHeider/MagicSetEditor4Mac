// **Mehrseitige Karten aus MTG JSON** —
// Abnahmekriterium aus #888.
package mtgjson_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mtgjson"
)

// **Der Name einer Seite ist ihr `faceName`, nicht der zusammengesetzte.**
//
// # Was dieser Lauf war und was er jetzt ist
//
// Er prüfte bis zum 2026-08-24, dass aus zwei Seiten **zwei** Karten mit je
// eigenem Namen werden. Der Zweitblick hat das angestrichen: MSE führt eine
// mehrseitige Karte als **eine** Karte mit zwei Feldsätzen, und
// `magic-set-generator` schreibt es genauso.
//
// **Die Zusicherung über den Namen bleibt** — sie war richtig. Sie steht jetzt
// nur an der zusammengelegten Karte, und der Lauf darüber
// ([TestZweiSeitenWerdenEineKarteMitZweitemFeldsatz]) hält die Struktur.
//
// **Warum er trotzdem hier steht**: er prüft die Wandlung `Seitenname` für
// sich. Wer sie entfernte, bekäme „Wear // Tear" in **beiden** Feldsätzen, und
// der Lauf darüber sähe es an `name`, dieser an beidem.
func TestEineKartenseiteBekommtIhrenEigenenNamen(t *testing.T) {
	t.Parallel()

	inhalt := `{"meta":{"version":"5.2.2"},"data":{"code":"PRF","name":"Prüfsatz","cards":[
		{"name":"Wear // Tear","faceName":"Wear","side":"a","layout":"split",
		 "manaCost":"{1}{R}","type":"Instant","types":["Instant"],
		 "otherFaceIds":["zweite-seite"],"uuid":"erste-seite"},
		{"name":"Wear // Tear","faceName":"Tear","side":"b","layout":"split",
		 "manaCost":"{W}","type":"Instant","types":["Instant"],
		 "otherFaceIds":["erste-seite"],"uuid":"zweite-seite"}
	]}}`

	quelle, fehler := mtgjson.LiesInhalt([]byte(inhalt), "probe.json", mtgjson.ZuordnungMagic)
	if fehler != nil {
		t.Fatalf("LiesInhalt: %v", fehler)
	}

	if len(quelle.Sets) != 1 || len(quelle.Sets[0].Karten) != 1 {
		t.Fatalf("Quelle = %+v", quelle)
	}

	werte := quelle.Sets[0].Karten[0].Werte

	erster, _ := wertVon(werte, "name")
	zweiter, _ := wertVon(werte, "name 2")

	if erster != "Wear" || zweiter != "Tear" {
		t.Errorf("Namen = [%q %q], erwartet [Wear Tear] — der zusammengesetzte "+
			"Name gehört keiner der beiden Seiten", erster, zweiter)
	}
}

// **Eine einseitige Karte behält ihren Namen.**
//
// Ohne diesen Fall prüfte der Test darüber nur, dass `faceName` **irgendwie**
// benutzt wird — und eine Fassung, die immer `faceName` nähme, ließe jede
// gewöhnliche Karte namenlos.
func TestEineEinseitigeKarteBehaeltIhrenNamen(t *testing.T) {
	t.Parallel()

	inhalt := `{"data":{"code":"PRF","name":"Prüfsatz","cards":[
		{"name":"Lightning Bolt","layout":"normal","types":["Instant"],"uuid":"x"}]}}`

	quelle, fehler := mtgjson.LiesInhalt([]byte(inhalt), "probe.json", mtgjson.ZuordnungMagic)
	if fehler != nil {
		t.Fatalf("LiesInhalt: %v", fehler)
	}

	if name, _ := wertVon(quelle.Sets[0].Karten[0].Werte, "name"); name != "Lightning Bolt" {
		t.Errorf("Name = %q", name)
	}
}

// **Und eine Seite ohne `faceName` fällt auf `name` zurück.**
//
// Eine von Hand geschriebene Datei kann `side` führen und `faceName` nicht.
// Sie deshalb namenlos zu lassen wäre die schlechtere Antwort — §8: „Ein Set,
// von dem 300 Karten lesbar sind und zwei nicht, wird mit 300 Karten geöffnet."
func TestEineSeiteOhneFaceNameFaelltZurueck(t *testing.T) {
	t.Parallel()

	inhalt := `{"data":{"code":"PRF","name":"Prüfsatz","cards":[
		{"name":"Halb // Fertig","side":"a","layout":"split","uuid":"x"}]}}`

	quelle, fehler := mtgjson.LiesInhalt([]byte(inhalt), "probe.json", mtgjson.ZuordnungMagic)
	if fehler != nil {
		t.Fatalf("LiesInhalt: %v", fehler)
	}

	if name, _ := wertVon(quelle.Sets[0].Karten[0].Werte, "name"); name != "Halb // Fertig" {
		t.Errorf("Name = %q — ohne faceName bleibt nur der zusammengesetzte", name)
	}
}

// **Regelhinweise und Identifiers werden nicht übernommen — mit Grund.**
//
// # Warum das eine Zusicherung ist und keine Auslassung
//
// §10: „Auch eine Ablehnung braucht Deckung." Ohne einen genannten Grund wäre
// ein bewusst übergangenes Feld von einem vergessenen nicht zu unterscheiden —
// und `TestJedesUebergangeneFeldNenntSeinenGrund` hält es fest.
func TestRulingsUndIdentifiersHabenIhrenGrund(t *testing.T) {
	t.Parallel()

	gruende := map[string]string{}

	for _, paar := range mtgjson.ZuordnungMagic.Felder {
		if paar.Feld == "" {
			for _, schluessel := range paar.JSON {
				gruende[schluessel] = paar.Grund
			}
		}
	}

	for _, schluessel := range []string{"rulings", "identifiers"} {
		grund, da := gruende[schluessel]
		if !da {
			t.Errorf("%q ist weder zugeordnet noch mit Grund übergangen", schluessel)

			continue
		}

		if !strings.Contains(grund, "MSE") && !strings.Contains(grund, "Scryfall") {
			t.Errorf("der Grund für %q sagt nicht, wohin die Angabe stattdessen "+
				"gehört: %q", schluessel, grund)
		}
	}
}

// **Ein unlesbarer Name wird gemeldet, nicht geraten.**
//
// Eine Zahl in `faceName` ist kein Name. Sie in eine Zeichenkette zu zwingen
// hieße, eine Eingabe zu erfinden — §8: fremde Daten sind der Normalfall, und
// der Fehler bekommt seine Stelle.
func TestEinUnlesbarerSeitennameWirdGemeldet(t *testing.T) {
	t.Parallel()

	inhalt := `{"data":{"code":"PRF","name":"Prüfsatz","cards":[
		{"name":"Gut","faceName":42,"layout":"normal","uuid":"x"}]}}`

	quelle, fehler := mtgjson.LiesInhalt([]byte(inhalt), "probe.json", mtgjson.ZuordnungMagic)
	if fehler != nil {
		t.Fatalf("LiesInhalt: %v", fehler)
	}

	// **Die Karte kommt trotzdem** — mit einem Befund, nicht mit einer Absage.
	if len(quelle.Sets[0].Karten) != 1 {
		t.Fatalf("%d Karten", len(quelle.Sets[0].Karten))
	}

	if len(quelle.Befunde) == 0 {
		t.Error("der unlesbare Name blieb ohne Befund")
	}
}

// **Eine Seite ganz ohne Namensangabe bleibt leer, statt zu scheitern.**
func TestEineSeiteOhneJedenNamenBleibtLeer(t *testing.T) {
	t.Parallel()

	inhalt := `{"data":{"code":"PRF","name":"Prüfsatz","cards":[
		{"layout":"normal","types":["Instant"],"uuid":"x"}]}}`

	quelle, fehler := mtgjson.LiesInhalt([]byte(inhalt), "probe.json", mtgjson.ZuordnungMagic)
	if fehler != nil {
		t.Fatalf("LiesInhalt: %v", fehler)
	}

	if name, da := wertVon(quelle.Sets[0].Karten[0].Werte, "name"); da && name != "" {
		t.Errorf("Name = %q, erwartet keinen", name)
	}
}

// **Die beiden Wandlungen, direkt geprüft.**
//
// # Warum direkt und nicht über den Leser
//
// Zwei ihrer Zweige sind über [mtgjson.Zuordnung.Karte] **nicht** erreichbar:
// `werteAus` überspringt ein Feldpaar, dessen Schlüssel die Karte gar nicht
// führt, und ruft die Wandlung dann nie. Sie stehen trotzdem im Code, weil die
// Wandlungen exportiert sind — wer sie benutzt, bekommt diese Antwort.
func TestDieWandlungenDirekt(t *testing.T) {
	t.Parallel()

	t.Run("Seitenname gibt nichts, wenn beide Angaben leer sind", func(t *testing.T) {
		t.Parallel()

		name, fehler := mtgjson.Seitenname([]json.RawMessage{
			json.RawMessage(`""`), json.RawMessage(`""`),
		})
		if fehler != nil {
			t.Fatalf("Seitenname: %v", fehler)
		}

		if name != "" {
			t.Errorf("Name = %q", name)
		}
	})

	t.Run("Seitenname überspringt eine fehlende Angabe", func(t *testing.T) {
		t.Parallel()

		name, fehler := mtgjson.Seitenname([]json.RawMessage{nil, json.RawMessage(`"Rückfall"`)})
		if fehler != nil {
			t.Fatalf("Seitenname: %v", fehler)
		}

		if name != "Rückfall" {
			t.Errorf("Name = %q", name)
		}
	})

	t.Run("Seitenname meldet eine unlesbare Angabe", func(t *testing.T) {
		t.Parallel()

		if _, fehler := mtgjson.Seitenname(
			[]json.RawMessage{json.RawMessage(`42`)}); fehler == nil {
			t.Error("die Zahl kam als Name durch")
		}
	})
}
