package service_test

// **Entwürfe bearbeiten, kopieren, löschen — und das Archiv** —
// TRADE-01.AK-03 (#1715):
// was hinausgegangen ist, wird nicht mehr bearbeitet oder gelöscht, sondern
// weitergeführt, kopiert oder archiviert.

import (
	"errors"
	"slices"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func zeileMit(angebot domain.Handelsangebot) domain.Handelsvorgangzeile {
	return domain.Handelsvorgangzeile{Angebot: angebot}
}

func TestEinEntwurfWirdBearbeitet(t *testing.T) {
	t.Parallel()

	liste, ablage := verzeichnis(zeileMit(verzeichnisangebot("v1", "bdu", domain.HandelEntwurf)))

	// **Was der Rufer über Zustand, Revision und Initiator sagt, zählt nicht**
	// — ein Entwurf bleibt Revision 1 dessen, der ihn anlegte.
	wunsch := verzeichnisangebot("v1", "bdu", domain.HandelAngenommen)
	wunsch.Revision, wunsch.Initiator, wunsch.Zustimmungen = 7, "bdu", []string{"bdu"}
	wunsch.Positionen[0].Menge = 4

	gespeichert, fehler := liste.EntwurfAendern(t.Context(), wunsch)
	if fehler != nil {
		t.Fatalf("ändern: %v", fehler)
	}

	if len(ablage.ersetzt) != 1 || gespeichert.Revision != 1 || gespeichert.Zustand != domain.HandelEntwurf ||
		gespeichert.Initiator != "aich" || gespeichert.Zustimmungen != nil || gespeichert.Positionen[0].Menge != 4 {
		t.Errorf("gespeichert: %+v", gespeichert)
	}
}

func TestNurEinEntwurfWirdBearbeitet(t *testing.T) {
	t.Parallel()

	leer := verzeichnisangebot("v1", "bdu", domain.HandelEntwurf)
	leer.Positionen = nil

	for _, fall := range []struct {
		name     string
		wunsch   domain.Handelsangebot
		erwartet error
	}{
		{"ein angebotener", verzeichnisangebot("v2", "bdu", domain.HandelAngeboten), domain.ErrConflict},
		{"ein fehlender", verzeichnisangebot("v9", "bdu", domain.HandelEntwurf), domain.ErrNotFound},
		{"ein ungültiger", leer, domain.ErrValidation},
	} {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			liste, ablage := verzeichnis(
				zeileMit(verzeichnisangebot("v1", "bdu", domain.HandelEntwurf)),
				zeileMit(verzeichnisangebot("v2", "bdu", domain.HandelAngeboten)),
			)

			if _, fehler := liste.EntwurfAendern(t.Context(), fall.wunsch); !errors.Is(fehler, fall.erwartet) {
				t.Errorf("Fehler ist %v", fehler)
			}

			if len(ablage.ersetzt) != 0 {
				t.Error("es wurde trotzdem ersetzt")
			}
		})
	}
}

// **Eine Kopie ist ein neuer Entwurf** — ohne Stimmen, ohne Revisionen, mit
// mir als Initiator; auch wenn ich im Original die Gegenseite war
// (TRADE-03.AK-04: keine Wiederholung alter Zustimmung).
func TestEineKopieIstEinNeuerEntwurf(t *testing.T) {
	t.Parallel()

	original := verzeichnisangebot("v1", "bdu", domain.HandelAngenommen)
	original.Initiator, original.Parteien = "bdu", []string{"bdu", "aich"}
	original.Revision, original.Eltern, original.Zustimmungen = 3, "eltern", []string{"bdu"}

	liste, ablage := verzeichnis(zeileMit(original))

	kopie, fehler := liste.Kopieren(t.Context(), "aich", "v1")
	if fehler != nil {
		t.Fatalf("kopieren: %v", fehler)
	}

	if kopie.Vorgang != "neu-1" || kopie.Revision != 1 || kopie.Eltern != "" ||
		kopie.Zustand != domain.HandelEntwurf || kopie.Initiator != "aich" ||
		!slices.Equal(kopie.Parteien, []string{"aich", "bdu"}) || kopie.Zustimmungen != nil ||
		len(kopie.Positionen) != 1 || len(ablage.geschrieben) != 1 {
		t.Errorf("Kopie: %+v", kopie)
	}

	// **Die Positionen sind eine Kopie und keine geteilte Scheibe.**
	kopie.Positionen[0].Menge = 9
	if original.Positionen[0].Menge == 9 {
		t.Error("die Kopie teilt ihre Positionen mit dem Original")
	}
}

func TestEineKopieBrauchtEinenEigenenVorgang(t *testing.T) {
	t.Parallel()

	liste, _ := verzeichnis(zeileMit(verzeichnisangebot("v1", "bdu", domain.HandelEntwurf)))

	if _, fehler := liste.Kopieren(t.Context(), "aich", "v9"); !errors.Is(fehler, domain.ErrNotFound) {
		t.Errorf("ein fehlender: %v", fehler)
	}

	if _, fehler := liste.Kopieren(t.Context(), "fremd", "v1"); !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("einer, an dem ich nicht beteiligt bin: %v", fehler)
	}
}

func TestLoeschenUndArchiv(t *testing.T) {
	t.Parallel()

	liste, ablage := verzeichnis(
		zeileMit(verzeichnisangebot("v1", "bdu", domain.HandelEntwurf)),
		zeileMit(verzeichnisangebot("v2", "bdu", domain.HandelAbgelehnt)),
	)

	if fehler := liste.EntwurfLoeschen(t.Context(), "v1"); fehler != nil || !slices.Equal(ablage.geloescht, []string{"v1"}) {
		t.Errorf("löschen: %v, %v", fehler, ablage.geloescht)
	}

	if fehler := liste.Archivieren(t.Context(), "v2", true); fehler != nil || !ablage.archiv["v2"] {
		t.Errorf("archivieren: %v", fehler)
	}

	// **Was läuft, geht nicht ins Archiv** — herausholen geht immer.
	if fehler := liste.Archivieren(t.Context(), "v1", true); !errors.Is(fehler, domain.ErrConflict) {
		t.Errorf("ein laufender: %v", fehler)
	}

	if fehler := liste.Archivieren(t.Context(), "v1", false); fehler != nil {
		t.Errorf("herausholen: %v", fehler)
	}

	if fehler := liste.Archivieren(t.Context(), "v9", false); !errors.Is(fehler, domain.ErrNotFound) {
		t.Errorf("ein fehlender: %v", fehler)
	}
}

// **Ein Schreiben, das scheitert, meldet sich** — und eine Kopie aus einem
// Vorgang ohne Position ist keine.
func TestDieEntwurfswegeMeldenIhreFehler(t *testing.T) {
	t.Parallel()

	ohne := verzeichnisangebot("leer", "bdu", domain.HandelAbgelehnt)
	ohne.Positionen = nil

	liste, ablage := verzeichnis(
		zeileMit(verzeichnisangebot("v1", "bdu", domain.HandelEntwurf)),
		zeileMit(verzeichnisangebot("v2", "bdu", domain.HandelAbgelehnt)),
		zeileMit(ohne),
	)

	if _, fehler := liste.Kopieren(t.Context(), "aich", "leer"); !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("eine Kopie ohne Position: %v", fehler)
	}

	ablage.schreibklage = errors.New("klemmt")

	if _, fehler := liste.EntwurfAendern(t.Context(), verzeichnisangebot("v1", "bdu", domain.HandelEntwurf)); fehler == nil {
		t.Error("ändern: der Fehler ging verloren")
	}

	if _, fehler := liste.Kopieren(t.Context(), "aich", "v1"); fehler == nil {
		t.Error("kopieren: der Fehler ging verloren")
	}

	if fehler := liste.Archivieren(t.Context(), "v2", true); fehler == nil {
		t.Error("archivieren: der Fehler ging verloren")
	}
}
