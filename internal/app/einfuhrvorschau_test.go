package app_test

// **Die Einfuhrvorschau an der Naht** — DOM-03.AK-01
// (#1698).

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func (stolpernde) Vorschauen(context.Context, []app.Listeneinfuhrzeile) (*app.Sammelvorschau, error) {
	return nil, errors.New("kein Netz")
}

func (stolpernde) CSVVorschauen(context.Context, string) (*app.Sammelvorschau, error) {
	return nil, errors.New("die Platte klemmt")
}

func (willige) Vorschauen(context.Context, []app.Listeneinfuhrzeile) (*app.Sammelvorschau, error) {
	return &app.Sammelvorschau{
		Posten:  []app.VorschaupostenDTO{{Zeile: 1, Art: "zusammen", Vorher: 2, Nachher: 3}},
		Befunde: []string{},
	}, nil
}

func (willige) CSVVorschauen(context.Context, string) (*app.Sammelvorschau, error) {
	return &app.Sammelvorschau{Posten: []app.VorschaupostenDTO{}, Befunde: []string{}}, nil
}

// **Beide Vorschauen gehen durch die App — geglückt, gescheitert und ohne
// Naht.**
func TestDieVorschauenGehenDurchDieApp(t *testing.T) {
	ohne := app.NeueApp(filepath.Join(t.TempDir(), "bestand.db"))
	if ohne.ListeVorschauen(t.Context(), nil).Fehler == nil ||
		ohne.SammlungCSVVorschauen(t.Context(), "x").Fehler == nil {
		t.Error("ohne Naht gab es keinen Fehler")
	}

	gut := app.NeueApp(filepath.Join(t.TempDir(), "bestand.db"))
	gut.MitListeneinfuhr(willige{})
	gut.MitSammlungsausfuhr(willige{})

	if antwort := gut.ListeVorschauen(t.Context(), nil); antwort.Fehler != nil ||
		antwort.Vorschau == nil || antwort.Vorschau.Posten[0].Art != "zusammen" {
		t.Errorf("Liste: %+v", antwort)
	}

	bindungen := app.NeueBindungen(gut)
	if antwort := bindungen.SammlungCSVVorschauen(t.Context(), "x"); antwort.Fehler != nil || antwort.Vorschau == nil {
		t.Errorf("CSV: %+v", antwort)
	}

	if antwort := bindungen.ListeVorschauen(t.Context(), nil); antwort.Fehler != nil {
		t.Errorf("Bindung Liste: %+v", antwort)
	}

	schlecht := app.NeueApp(filepath.Join(t.TempDir(), "bestand.db"))
	schlecht.MitListeneinfuhr(stolpernde{})
	schlecht.MitSammlungsausfuhr(stolpernde{})

	if schlecht.ListeVorschauen(t.Context(), nil).Fehler == nil ||
		schlecht.SammlungCSVVorschauen(t.Context(), "x").Fehler == nil {
		t.Error("ein Fehler der Naht ging verloren")
	}
}

// **Die Naht über einem echten Bestand**: die eigene Ausfuhr trifft beim
// zweiten Mal jede Zeile, und Auslassen lässt die Menge stehen.
func TestDieEigeneAusfuhrZweimalUeberDieNaht(t *testing.T) {
	naht := app.AusfuhrUeber(frischerBestand(t))
	text := "printing,menge,zustand,sprache\nprobe-druck,2," + string(domain.ZustandNearMint) + ",en\n"

	if _, err := naht.CSVEinlesen(probekontext(t), text, "", ""); err != nil {
		t.Fatalf("erstes Einlesen: %v", err)
	}

	vorschau, err := naht.CSVVorschauen(probekontext(t), text)
	if err != nil || len(vorschau.Posten) != 1 || vorschau.Posten[0].Art != "zusammen" ||
		vorschau.Posten[0].Vorher != 2 || vorschau.Posten[0].Nachher != 4 {
		t.Fatalf("Vorschau %+v (%v)", vorschau, err)
	}

	bericht, err := naht.CSVEinlesen(probekontext(t), text, "", "auslassen")
	if err != nil || bericht.Ausgelassen != 1 || bericht.Erhoeht != 0 {
		t.Errorf("zweites Einlesen %+v (%v)", bericht, err)
	}

	if _, err := naht.CSVVorschauen(probekontext(t), text); err != nil {
		t.Errorf("zweite Vorschau: %v", err)
	}

	if _, err := naht.CSVEinlesen(probekontext(t), text, "", "verdoppeln"); err == nil {
		t.Error("ein unbekannter Modus ging durch")
	}
}

// **Die Listenvorschau über der echten Naht** — und ein geschlossener Bestand
// ist ein Fehler, keine leere Vorschau.
func TestDieEinfuhrnahtSchautVor(t *testing.T) {
	zeilen := []app.Listeneinfuhrzeile{{
		Printing: "probe-druck", Menge: 1, Zustand: string(domain.ZustandNearMint), Sprache: "en",
	}}

	vorschau, err := einfuhrnaht(t).Vorschauen(probekontext(t), zeilen)
	if err != nil || len(vorschau.Posten) != 1 || vorschau.Posten[0].Art != "neu" {
		t.Fatalf("Vorschau %+v (%v)", vorschau, err)
	}

	// **Dieselbe Scryfall-Adresse und dasselbe Zuhause** — `einfuhrnaht`
	// hat beides schon gesetzt, und es gilt bis zum Ende des Laufs.
	db := frischerBestand(t)

	geschlossen, err := app.ListeneinfuhrUeber(db)
	if err != nil {
		t.Fatalf("ListeneinfuhrUeber: %v", err)
	}

	if fehler := db.Close(); fehler != nil {
		t.Fatalf("schließen: %v", fehler)
	}

	if _, err := geschlossen.Vorschauen(probekontext(t), zeilen); err == nil {
		t.Error("die Vorschau auf geschlossenem Bestand gab keinen Fehler")
	}

	ausfuhr := app.AusfuhrUeber(db)
	if _, err := ausfuhr.CSVVorschauen(probekontext(t), "printing,menge\nabc,1\n"); err == nil {
		t.Error("die CSV-Vorschau auf geschlossenem Bestand gab keinen Fehler")
	}
}
