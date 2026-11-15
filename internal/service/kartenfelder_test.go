package service_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Die Feldbeschreibungen eines Sets** —
// #560.
//
// Der Karteneditor braucht die **Form** jedes Feldes, um das passende
// Steuerelement zu wählen: ein Regeltext ist mehrzeilig, eine Seltenheit hat
// Wahlen, ein Kunstfenster ist ein Bild. Ohne sie stellte er jedes Feld als
// einzeiliges Textfeld dar.
func TestKartenfelderKommenAusDemSpielpaket(t *testing.T) {
	waehler := service.NeuerStilwaehler(
		&koepfequelle{koepfe: []domain.SetHead{
			{ID: 1, Game: "probe", Style: "probe-schlicht"},
		}},
		saatkatalog(), &stilablage{}, &feldliste{}, festeSaatwurzel(t),
	)

	felder, err := waehler.Kartenfelder(t.Context(), 1)
	if err != nil {
		t.Fatalf("Kartenfelder: %v", err)
	}

	nach := map[string]domain.Field{}
	for i := range felder {
		nach[felder[i].Name] = felder[i]
	}

	if nach["name"].Kind != domain.KindText {
		t.Errorf("name: %+v", nach["name"])
	}

	if nach["farbe"].Kind != domain.KindChoice || len(nach["farbe"].Choices) != 2 {
		t.Errorf("farbe: %+v", nach["farbe"])
	}

	if nach["kunst"].Kind != domain.KindImage {
		t.Errorf("kunst: %+v", nach["kunst"])
	}
}

// **Ein Fehler des Bestands kommt durch.**
func TestKartenfelderMeldenDenBestandsfehler(t *testing.T) {
	kaputt := errors.New("der Bestand schweigt")

	waehler := service.NeuerStilwaehler(
		&koepfequelle{fehler: kaputt},
		saatkatalog(), &stilablage{}, &feldliste{}, festeSaatwurzel(t),
	)

	if _, err := waehler.Kartenfelder(t.Context(), 1); !errors.Is(err, kaputt) {
		t.Errorf("%v", err)
	}
}

// **Und ein Spiel, das nicht dasteht, ist `ErrNotFound`.**
//
// Der Fall ist im Betrieb erreichbar: ein Set nennt sein Spiel, und das Paket
// ist nicht installiert.
func TestKartenfelderMeldenEinFehlendesSpiel(t *testing.T) {
	waehler := service.NeuerStilwaehler(
		&koepfequelle{koepfe: []domain.SetHead{{ID: 1, Game: "gibtsnicht"}}},
		saatkatalog(), &stilablage{}, &feldliste{}, festeSaatwurzel(t),
	)

	_, err := waehler.Kartenfelder(t.Context(), 1)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("%v", err)
	}
}

// **Das Spielskript kommt aus demselben Paket** —
// #1238.
//
// Die Kartenliste braucht es, um berechnete Spalten zu füllen: `full name`
// heißt `script: card_full_name()`, und diese Funktion legt das `init script:`
// des Spiels an.
func TestDasSpielskriptKommtAusDemSpielpaket(t *testing.T) {
	waehler := service.NeuerStilwaehler(
		&koepfequelle{koepfe: []domain.SetHead{
			{ID: 1, Game: "probe", Style: "probe-schlicht"},
		}},
		saatkatalog(), &stilablage{}, &feldliste{}, festeSaatwurzel(t),
	)

	skript, err := waehler.Spielskript(t.Context(), 1)
	if err != nil {
		t.Fatalf("Spielskript: %v", err)
	}

	if strings.TrimSpace(skript) == "" {
		t.Error("das Spielskript ist leer — die Saat führt eines")
	}
}

// **Ein Fehler des Bestands kommt auch hier durch.**
func TestDasSpielskriptMeldetDenBestandsfehler(t *testing.T) {
	kaputt := errors.New("der Bestand schweigt")

	waehler := service.NeuerStilwaehler(
		&koepfequelle{fehler: kaputt},
		saatkatalog(), &stilablage{}, &feldliste{}, festeSaatwurzel(t),
	)

	if _, err := waehler.Spielskript(t.Context(), 1); !errors.Is(err, kaputt) {
		t.Errorf("%v", err)
	}
}

// **Und ein Spiel, das nicht dasteht, ist auch hier `ErrNotFound`.**
func TestDasSpielskriptMeldetEinFehlendesSpiel(t *testing.T) {
	waehler := service.NeuerStilwaehler(
		&koepfequelle{koepfe: []domain.SetHead{{ID: 1, Game: "gibtsnicht"}}},
		saatkatalog(), &stilablage{}, &feldliste{}, festeSaatwurzel(t),
	)

	if _, err := waehler.Spielskript(t.Context(), 1); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("%v", err)
	}
}
