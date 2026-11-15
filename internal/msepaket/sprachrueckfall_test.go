// **`include localized file:` fällt auf Englisch zurück, nicht auf den bloßen
// Namen** — gefunden beim Fremdpaketlauf zur Freigabe 0.3.0
// (#1170).
//
// # Der Befund
//
// `TestDieKanonischenSetsZeichnen` meldete **zwei Befunde je Karte**, über den
// ganzen Kanon 1.360 — und zwar immer dieselben zwei:
//
//	magic.mse-game/game:25: magic.mse-game/statistics ist nicht lesbar
//	magic.mse-game/game:28: magic.mse-game/keywords ist nicht lesbar
//
// Beide stehen im Spielpaket als `include localized file:`. Der Einbinder
// suchte `<name>_<sprache>` und fiel sonst auf `<name>` zurück — **und diese
// Datei gibt es nicht.**
//
// # Was gemessen ist
//
// In `full-magic-pack/data/magic.mse-game` liegen **vierzehn** Fassungen von
// `keywords` und dreizehn von `statistics` — `_ch-s`, `_da`, `_de`, `_en`,
// `_es`, `_fr`, `_it`, `_jp`, `_ko`, `_pl`, `_pt-br`, `_ru` und `_custom`.
// **Eine Datei ohne Zusatz gibt es zu keinem von beiden.**
//
// `_en` ist bei beiden dabei. Damit ist der Rückfall belegt: wer die
// Grundfassung sucht, sucht etwas, das kein Pack mitbringt.
//
// # Was weiterhin erschlossen bleibt
//
// **Dass `_en` die Vorgabe ist** und nicht etwa die erste Datei in
// alphabetischer Ordnung. MSE2 f9d9356 kennt den Schlüssel gar nicht
// (`Reader::unknownKey` behandelt allein `include_file`), das Pack trägt
// `mse version: 2.5.8`. Die Marke in `docs/PARITAET.md` bleibt deshalb
// `uncertain` — was sich geändert hat, ist die **Zahl der Belege**, nicht
// ihre Art.
package msepaket_test

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// nurSprachfassungen ist ein Paket, das `keywords` **nur** in Sprachfassungen
// führt — genau wie `magic.mse-game`.
var nurSprachfassungen = dateien{
	"a.mse-game/keywords_en": "keyword:\n\tkeyword: Flying\n",
	"a.mse-game/keywords_de": "keyword:\n\tkeyword: Fliegend\n",
}

// eingebunden löst einen Sprachverweis in einer Sprache auf.
func eingebunden(t *testing.T, quelle msepaket.Paketquelle, sprache string) (string, []msepaket.Befund) {
	t.Helper()

	baum, err := msepaket.Zerteile("game", []byte("include localized file: keywords\n"))
	if err != nil {
		t.Fatalf("zerteilen: %v", err)
	}

	einbinder := msepaket.NeuerEinbinder(quelle).ImPaket("a.mse-game").InSprache(sprache)

	return string(msepaket.Schreibe(einbinder.Einbinden(baum))), einbinder.Befunde()
}

func TestOhneSpracheGiltEnglisch(t *testing.T) {
	t.Parallel()

	baum, befunde := eingebunden(t, nurSprachfassungen, "")

	if len(befunde) != 0 {
		t.Fatalf("Befunde: %v — es gibt keine Grundfassung, und das ist der "+
			"Normalfall in jedem echten Pack", befunde)
	}

	if !strings.Contains(baum, "Flying") {
		t.Errorf("die englische Fassung fehlt im Baum:\n%s", baum)
	}
}

func TestEineGewaehlteSpracheSchlaegtEnglisch(t *testing.T) {
	t.Parallel()

	baum, befunde := eingebunden(t, nurSprachfassungen, "de")

	if len(befunde) != 0 {
		t.Fatalf("Befunde: %v", befunde)
	}

	if !strings.Contains(baum, "Fliegend") {
		t.Errorf("die deutsche Fassung fehlt im Baum:\n%s", baum)
	}
}

// **Und eine Sprache, die es nicht gibt, fällt auf Englisch** — nicht auf
// einen Befund.
func TestEineUnbekannteSpracheFaelltAufEnglisch(t *testing.T) {
	t.Parallel()

	baum, befunde := eingebunden(t, nurSprachfassungen, "kl")

	if len(befunde) != 0 {
		t.Fatalf("Befunde: %v", befunde)
	}

	if !strings.Contains(baum, "Flying") {
		t.Errorf("der englische Rückfall fehlt:\n%s", baum)
	}
}

// **Die Grundfassung bleibt der letzte Rückfall.** Ein Paket, das `keywords`
// ohne Zusatz führt, wird weiterhin gelesen — sonst bräche diese Änderung, was
// vorher ging.
func TestOhneSprachfassungGiltDieGrundfassung(t *testing.T) {
	t.Parallel()

	quelle := dateien{"a.mse-game/keywords": "keyword:\n\tkeyword: Schlicht\n"}

	baum, befunde := eingebunden(t, quelle, "de")

	if len(befunde) != 0 {
		t.Fatalf("Befunde: %v", befunde)
	}

	if !strings.Contains(baum, "Schlicht") {
		t.Errorf("die Grundfassung fehlt:\n%s", baum)
	}
}
