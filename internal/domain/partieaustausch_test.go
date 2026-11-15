package domain_test

// **Eine Partie als Datei** — PLAY-03.AK-05
// (#1711).

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// eineAusfuhrpartie ist die kleinste, die durch [domain.Partie.Pruefe] geht.
func eineAusfuhrpartie() domain.Partie {
	abzug := domain.Partiedeck{
		Abdruck: "abcd",
		Karten:  []domain.Abzugskarte{{Kennung: "k1", Name: "Wald", Menge: 40}},
	}

	return domain.Partie{
		ID:    "p1",
		Spiel: "magic",
		Regelwerk: domain.Regelwerk{
			Kennung: "sandkasten",
			Fassung: domain.Grundprofilfassung,
			Strenge: domain.StrengeSandkasten,
		},
		Profil:  domain.Grundprofil("magic"),
		Wirt:    "sitz-1",
		Stand:   domain.StandLaeuft,
		Fassung: 9,
		Sitze: []domain.Sitz{
			{Kennung: "sitz-1", Name: "David", Deck: abzug},
			{Kennung: "sitz-2", Name: "Gegenüber", Deck: abzug},
		},
		Objekte: []domain.Spielobjekt{{
			Kennung: "o1", Art: domain.ObjektKarte, Beherrscher: "sitz-1",
			Eigner: "sitz-1", Zone: domain.ZoneHand, Karte: "k1",
		}},
	}.MitStartleben()
}

// **Was hinausgeht, kommt zurück** — und die Stellung ist dieselbe.
func TestEinePartieGehtDurchDieDateiUndZurueck(t *testing.T) {
	t.Parallel()

	gelesen, befunde, fehler := domain.ZerlegePartie(
		domain.FormePartie(eineAusfuhrpartie()))
	if fehler != nil {
		t.Fatalf("zerlegen: %v", fehler)
	}

	if len(befunde) != 0 {
		t.Errorf("unerwartete Befunde: %v", befunde)
	}

	if gelesen.Fassung != domain.PartieaustauschFassung {
		t.Errorf("Fassung %d", gelesen.Fassung)
	}

	if len(gelesen.Partie.Objekte) != 1 ||
		gelesen.Partie.Objekte[0].Kennung != "o1" {
		t.Errorf("die Objekte kamen nicht mit: %v", gelesen.Partie.Objekte)
	}

	if len(gelesen.Partie.Sitze) != 2 {
		t.Errorf("die Sitze kamen nicht mit: %v", gelesen.Partie.Sitze)
	}
}

// **Die Datei ist lesbar für Menschen** — eingerückt und mit Zeilenende.
//
// **Warum das eine Zusage ist und keine Zierde**: eine ausgeführte Partie
// landet in einem Verzeichnis, wird per Hand weitergereicht und irgendwann
// in einem Editor geöffnet. Eine einzige Zeile mit 20.000 Zeichen wäre dort
// unbrauchbar, und `git diff` über zwei davon sagte nichts.
func TestDieDateiIstLesbar(t *testing.T) {
	t.Parallel()

	roh := domain.FormePartie(eineAusfuhrpartie())

	if !strings.Contains(string(roh), "\n  \"fassung\"") {
		t.Errorf("die Datei ist nicht eingerückt: %.60q", roh)
	}

	if !strings.HasSuffix(string(roh), "\n") {
		t.Error("die Datei endet ohne Zeilenende")
	}
}

// **Vier Dateien, die abgewiesen werden** — und jede aus einem anderen
// Grund, den der Satz nennt (§8).
func TestZerlegePartieWeistAb(t *testing.T) {
	t.Parallel()

	faelle := []struct {
		name string
		roh  []byte
		teil string
	}{
		{
			name: "zu groß",
			roh:  make([]byte, domain.PartieaustauschHoechstens+1),
			teil: "groß",
		},
		// **Zwei Fälle und nicht einer**: `{"a":}` bricht schon beim
		// Zählen der Stufen ab — dort liest ein Entzifferer Marke für
		// Marke; `{` kommt bis dahin durch, weil ein zu Ende gegangener
		// Strom für ihn ein Ende ist und kein Fehler (gemessen am
		// 2026-09-22). Wer nur einen prüfte, hätte den anderen Weg nie
		// betreten.
		{name: "kein JSON", roh: []byte(`{"a":}`), teil: "kein JSON"},
		{name: "abgeschnitten", roh: []byte("{"), teil: "kein Objekt"},
		{name: "kein Objekt", roh: []byte(`"eine Zeichenkette"`), teil: "Objekt"},
		{
			name: "Fassung aus der Zukunft",
			roh:  []byte(`{"fassung":99,"partie":{}}`),
			teil: "Fassung 99",
		},
		{
			name: "Fassung ist keine Zahl",
			roh:  []byte(`{"fassung":"eins","partie":{}}`),
			teil: "keine Zahl",
		},
		{
			name: "Partie hat eine fremde Gestalt",
			roh:  []byte(`{"fassung":1,"partie":42}`),
			teil: "Gestalt",
		},
		{
			name: "zu tief verschachtelt",
			roh:  tiefeDatei(domain.PartieaustauschTiefe + 1),
			teil: "Stufen tief",
		},
		{
			name: "die Partie taugt nicht",
			roh:  []byte(`{"fassung":1,"partie":{"id":"p1"}}`),
			teil: "Spiel",
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			_, _, fehler := domain.ZerlegePartie(fall.roh)
			if !errors.Is(fehler, domain.ErrValidation) {
				t.Fatalf("erwartet ErrValidation, bekam %v", fehler)
			}

			if !strings.Contains(fehler.Error(), fall.teil) {
				t.Errorf("die Meldung %q nennt %q nicht", fehler, fall.teil)
			}
		})
	}
}

// **Ein fremdes Feld ist ein Befund und kein Fehler** — dieselbe Regel wie
// beim Einstellungsaustausch: die Datei bleibt lesbar, und dass etwas darin
// stand, das niemand kennt, wird gesagt statt verschluckt (§8, §16).
func TestEinFremdesFeldIstEinBefund(t *testing.T) {
	t.Parallel()

	gut := domain.FormePartie(eineAusfuhrpartie())

	var felder map[string]json.RawMessage

	if fehler := json.Unmarshal(gut, &felder); fehler != nil {
		t.Fatalf("lesen: %v", fehler)
	}

	felder["übermorgen"] = json.RawMessage(`true`)

	mitFremdem, fehler := json.Marshal(felder)
	if fehler != nil {
		t.Fatalf("schreiben: %v", fehler)
	}

	_, befunde, fehler := domain.ZerlegePartie(mitFremdem)
	if fehler != nil {
		t.Fatalf("zerlegen: %v", fehler)
	}

	if len(befunde) != 1 || !strings.Contains(befunde[0].Grund, "übermorgen") {
		t.Fatalf("erwartet ein Befund über das fremde Feld, bekam %v", befunde)
	}
}

// tiefeDatei baut ein JSON mit der gewünschten Verschachtelung.
func tiefeDatei(stufen int) []byte {
	gebaut := "1"

	for range stufen {
		gebaut = "[" + gebaut + "]"
	}

	return []byte(gebaut)
}
