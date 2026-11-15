package service_test

// **Ein Deck aus einem fremden Profil übernehmen** — Werkstück 12 von
// Phase 4.12
// (#1351).
//
// **Das ist die Abnahme dieses Werkstücks**: nicht, dass sich eine Herkunft
// speichern lässt, sondern dass etwas ankommt, das woanders herkommt — und
// dass danach dasteht, woher.
//
// **`deckuebernahme_test.go` und nicht `uebernahme_test.go`**: der Name war
// seit Phase 4.10 vergeben — dort steht die Übernahme von **Vorlagen aus einer
// MSE2-Installation**. Zwei verschiedene Übernahmen, zwei Dateien.

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Ein Deck aus echten Drucken kommt vollständig an.**
func TestEinFremdesDeckKommtAn(t *testing.T) {
	t.Parallel()

	uebernahme, doppel := zweiProfileMitDeck(t, domain.RolleMitarbeiter)

	angekommen, befunde, fehler := uebernahme.DeckUebernehmen(t.Context(), 2, 7)
	if fehler != nil {
		t.Fatalf("%v", fehler)
	}

	if len(befunde) != 0 {
		t.Errorf("es kamen Befunde: %v", befunde)
	}

	if angekommen.Name != "Nachtlauf" || len(angekommen.Plaetze) != 2 {
		t.Errorf("angekommen ist %+v", angekommen)
	}

	// **Und die Herkunft steht daneben** — mit dem Namen des Quellprofils,
	// abgeschrieben und nicht als Kennung.
	if doppel.deckherkunft.Art != domain.HerkunftProfil ||
		doppel.deckherkunft.Quelle != "Zweiter · Fremd" ||
		doppel.deckherkunft.QuellProfil != 2 {
		t.Errorf("die Herkunft lautet %+v", doppel.deckherkunft)
	}

	if doppel.quellDeck != 7 {
		t.Errorf("das Quelldeck ist %d", doppel.quellDeck)
	}
}

// **Ohne Freigabe kommt gar nichts an.**
//
// **Und zwar bevor gelesen wird**: wer nicht hinein darf, soll auch nicht
// erfahren, was drinsteht.
func TestOhneFreigabeKommtKeinDeck(t *testing.T) {
	t.Parallel()

	uebernahme, doppel := zweiProfileMitDeck(t, "")

	if _, _, fehler := uebernahme.DeckUebernehmen(t.Context(), 2, 7); fehler == nil {
		t.Fatal("das Deck kam ohne Freigabe durch")
	}

	if doppel.entnommen {
		t.Error("das fremde Deck wurde trotzdem gelesen")
	}
}

// **Ein Leser übernimmt — seit Werkstück 16.**
//
// # Hier stand das Gegenteil, und die Begründung war schon damals dieselbe
//
// **„Dieselbe Grenze wie beim Wechsel"** — das stand hier, als der Wechsel
// einem `leser` verschlossen war (Vermerk
// [0070](../../docs/entscheidungen/0070-rechte-rollen-und-sichtbarkeit.md)).
// Die Grenze hat sich verschoben, der Satz gilt unverändert.
//
// **Und der Grund dafür steht in `uebernahme.go` selbst**: „wer aus einem
// Profil nehmen darf, ist genau, wer hineingehen darf — eine Übernahme, die
// mehr verlangte, wäre mit einem Wechsel zu umgehen: hineingehen, ansehen,
// abtippen." Seit ein `leser` hineingeht und liest, kann er genau das.
//
// **Geschrieben wird dabei ins eigene Profil**, nicht ins fremde: die
// Übernahme läuft, während der Zeiger im eigenen steht — `DeckUebernehmen`
// weist eine Quelle ab, die das gezeigte Profil ist.
func TestEinLeserUebernimmt(t *testing.T) {
	t.Parallel()

	uebernahme, _ := zweiProfileMitDeck(t, domain.RolleLeser)

	if _, _, fehler := uebernahme.DeckUebernehmen(t.Context(), 2, 7); fehler != nil {
		t.Errorf("ein Leser holte kein Deck: %v", fehler)
	}
}

// **Ein Platz auf eine eigene Karte bleibt zurück — und wird genannt.**
//
// # Warum er nicht mitkommt
//
// **Weil er auf eine Karte im Quellprofil zeigt**, und die gibt es hier
// nicht. Ihn mitzunehmen hieße, einen Verweis über die Profilgrenze zu legen —
// genau das, was Werkstück 2 verhindert.
//
// **Und weil ein stilles Weglassen schlimmer wäre als eine Absage** (§8): wer
// ein Deck mit vierzig Karten übernimmt und dreißig bekommt, muss es lesen
// können.
func TestEinPlatzAufEineEigeneKarteBleibtZurueck(t *testing.T) {
	t.Parallel()

	uebernahme, _ := zweiProfileMitDeck(t, domain.RolleMitarbeiter)

	angekommen, befunde, fehler := uebernahme.DeckUebernehmen(t.Context(), 3, 8)
	if fehler != nil {
		t.Fatalf("%v", fehler)
	}

	if len(angekommen.Plaetze) != 1 {
		t.Errorf("es kamen %d Plätze an", len(angekommen.Plaetze))
	}

	if len(befunde) != 1 {
		t.Fatalf("es kamen %d Befunde: %v", len(befunde), befunde)
	}
}

// **Das Format kommt nicht mit — und auch das wird genannt.**
//
// **Ein Format ist eine Zeile im Quellprofil**, und dieselbe Zeile gibt es
// hier nicht. Ein Deck ohne Format ist **ungeprüft, nicht falsch** (Vermerk
// 0061) — die Absage ist deshalb ein Befund und keine Ablehnung.
func TestDasFormatKommtNichtMit(t *testing.T) {
	t.Parallel()

	uebernahme, _ := zweiProfileMitDeck(t, domain.RolleMitarbeiter)

	angekommen, befunde, fehler := uebernahme.DeckUebernehmen(t.Context(), 4, 9)
	if fehler != nil {
		t.Fatalf("%v", fehler)
	}

	if angekommen.FormatID != 0 {
		t.Errorf("das Format kam mit: %d", angekommen.FormatID)
	}

	if len(befunde) != 1 {
		t.Errorf("es kamen %d Befunde: %v", len(befunde), befunde)
	}
}

// **Und aus dem eigenen Profil übernimmt niemand.**
//
// **Ein Deck, das schon hier ist, kommt nicht von woanders.** Eine Kopie
// davon wäre eine Kopie und keine Übernahme — und ihre Herkunft behauptete,
// sie käme von außen.
func TestAusDemEigenenProfilWirdNichtUebernommen(t *testing.T) {
	t.Parallel()

	uebernahme, _ := zweiProfileMitDeck(t, domain.RolleMitarbeiter)

	if _, _, fehler := uebernahme.DeckUebernehmen(
		t.Context(), 1, 7); !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("Fehler = %v, erwartet ErrValidation", fehler)
	}
}
