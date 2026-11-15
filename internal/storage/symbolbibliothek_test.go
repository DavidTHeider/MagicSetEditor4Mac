package storage_test

// **Die Symbolbibliothek im Bestand** — DOM-02
// (#1697):
// je Profil, ein Name einmal, und wenn ein Symbol geht, gehen seine Marken.

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

func symbol(name string) domain.Bibliothekssymbol {
	return domain.Bibliothekssymbol{Name: name, Inhalt: "mse version: 0.3.5\n", Abdruck: "abdruck:" + name}
}

func TestEinSymbolGehtInDieBibliothekUndKommtWieder(t *testing.T) {
	db := frischeDatenbank(t)
	ablage := storage.NeueSymbolbibliothek(db)
	ctx := probekontext(t)

	krone, err := ablage.SymbolAnlegen(ctx, symbol("Krone"))
	if err != nil || krone.ID == 0 || krone.Angelegt.IsZero() {
		t.Fatalf("anlegen: %+v, %v", krone, err)
	}

	if _, anlegefehler := ablage.SymbolAnlegen(ctx, symbol("anker")); anlegefehler != nil {
		t.Fatalf("anlegen: %v", anlegefehler)
	}

	// **Nach Namen, ohne Rücksicht auf Groß und Klein.**
	symbole, err := ablage.Symbole(ctx)
	if err != nil || len(symbole) != 2 || symbole[0].Name != "anker" || symbole[1].Inhalt != "mse version: 0.3.5\n" {
		t.Errorf("die Liste: %+v, %v", symbole, err)
	}

	gelesen, err := ablage.Symbol(ctx, krone.ID)
	if err != nil || gelesen.Abdruck != "abdruck:Krone" {
		t.Errorf("gelesen: %+v, %v", gelesen, err)
	}

	// **Ein Name einmal je Profil** — auch in anderer Schreibung, und die
	// Absage steht am Namensfeld (ENG-05.AK-03, #1730).
	var feldfehler *domain.FieldError
	if _, err := ablage.SymbolAnlegen(ctx, symbol("KRONE")); !errors.Is(err, domain.ErrDoppelt) ||
		!errors.As(err, &feldfehler) || feldfehler.Field != "name" ||
		feldfehler.Key != domain.SchluesselDoppelterName {
		t.Errorf("doppelt: %v", err)
	}

	if err := ablage.SymbolUmbenennen(ctx, krone.ID, "Diadem"); err != nil {
		t.Errorf("umbenennen: %v", err)
	}

	if err := ablage.SymbolUmbenennen(ctx, krone.ID, "Anker"); !errors.Is(err, domain.ErrDoppelt) {
		t.Errorf("umbenennen auf einen vorhandenen: %v", err)
	}

	// **Der Inhalt wechselt, und der vorige bleibt** — der Weg zurück.
	if err := ablage.SymbolInhaltSetzen(ctx, krone.ID, "neu", "abdruck:neu"); err != nil {
		t.Errorf("Inhalt: %v", err)
	}

	if geaendert, lesefehler := ablage.Symbol(ctx, krone.ID); lesefehler != nil || geaendert.Inhalt != "neu" || geaendert.Vorher != "mse version: 0.3.5\n" ||
		geaendert.Name != "Diadem" {
		t.Errorf("nach dem Ändern: %+v", geaendert)
	}
}

func TestEinSymbolGehoertSeinemProfil(t *testing.T) {
	db := frischeDatenbank(t)
	ablage := storage.NeueSymbolbibliothek(db)

	krone, err := ablage.SymbolAnlegen(probekontext(t), symbol("Krone"))
	if err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	fremd := storage.MitProfil(t.Context(), zweitesProfil(t, storage.NeueNutzerablage(db), "Zweites"))

	if symbole, err := ablage.Symbole(fremd); err != nil || len(symbole) != 0 {
		t.Errorf("das fremde Profil sieht: %+v, %v", symbole, err)
	}

	for name, ruf := range map[string]func() error{
		"lesen":      func() error { _, e := ablage.Symbol(fremd, krone.ID); return e },
		"umbenennen": func() error { return ablage.SymbolUmbenennen(fremd, krone.ID, "X") },
		"Inhalt":     func() error { return ablage.SymbolInhaltSetzen(fremd, krone.ID, "x", "y") },
		"löschen":    func() error { return ablage.SymbolLoeschen(fremd, krone.ID) },
	} {
		if err := ruf(); !errors.Is(err, domain.ErrNotFound) {
			t.Errorf("%s über das fremde Profil: %v", name, err)
		}
	}

	// **Ohne Profil kein Zugriff.**
	if _, err := ablage.Symbole(t.Context()); !errors.Is(err, domain.ErrValidation) {
		t.Errorf("ohne Profil: %v", err)
	}
}

// **Wenn ein Symbol geht, gehen seine Marken** — derselbe Auslöser wie an
// jedem Fachobjekt (0135).
func TestEinGeloeschtesSymbolNimmtSeineMarkenMit(t *testing.T) {
	db := frischeDatenbank(t)
	ablage := storage.NeueSymbolbibliothek(db)
	marken := storage.NeueObjektmarkenablage(db)
	ctx := probekontext(t)

	krone, err := ablage.SymbolAnlegen(ctx, symbol("Krone"))
	if err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	art := domain.MarkenartSymbol
	if da, err := marken.Vorhanden(ctx, art, "1"); !da || err != nil {
		t.Errorf("das Symbol ist nicht vorhanden: %v", err)
	}

	if err := marken.Setzen(ctx, art, "1", "gold"); err != nil {
		t.Fatalf("Marke: %v", err)
	}

	if err := ablage.SymbolLoeschen(ctx, krone.ID); err != nil {
		t.Fatalf("löschen: %v", err)
	}

	if je, err := marken.JeObjekt(ctx, art); err != nil || len(je) != 0 {
		t.Errorf("die Marken nach dem Löschen: %+v, %v", je, err)
	}

	// **Und ein gelöschtes Symbol ist nicht mehr vorhanden.**
	if da, err := marken.Vorhanden(ctx, art, "1"); da || err != nil {
		t.Errorf("vorhanden nach dem Löschen: %v, %v", da, err)
	}

	if err := ablage.SymbolLoeschen(ctx, krone.ID); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("zweimal löschen: %v", err)
	}
}
