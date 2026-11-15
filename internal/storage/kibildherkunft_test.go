package storage_test

// **Was über ein übernommenes Bild feststeht** — Werkstück A12 von Phase 5.3
// (#1451).
//
// # Was hier gemessen wird
//
// **Dass Wert, Feldherkunft und Bildangaben in einem Zug stehen.** Das ist
// die Zusage aus A11, eine Ebene tiefer: „Scheitert Provenienz/Verlauf in
// derselben Transaktion, stehen keine halben Inhaltsänderungen."
//
// **Und dass ein zweiter Lauf nichts vom ersten stehen lässt.** Nennt er
// keinen Seed, ist auch keiner mehr da — sonst stünde eine Angabe über ein
// Bild, das es nicht mehr gibt.

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// eineBildherkunft baut die Feldherkunft des Bildfeldes.
func eineBildherkunft() domain.KIFeldherkunft {
	return domain.KIFeldherkunft{
		Feld: "image", Anbieter: "openai", Modell: "gpt-image-2.5-flare",
		Aufgabe: "ein Blitz", Auftrag: "ki-4-1",
		Zeitpunkt: time.Date(2026, 9, 12, 9, 0, 0, 0, time.UTC),
		Bezug:     "sha256:abc",
	}
}

// mitBildkarte legt eine Karte an und gibt Schreiber, Leser und Kennung.
func mitBildkarte(t *testing.T) (
	context.Context, *storage.Kartenschreiber, *storage.Feldherkunft, int64,
) {
	t.Helper()

	db := frischeDatenbank(t)
	karte := eineKarte(t, db, domain.Value{Field: "image", Text: "alt.png"})

	return probekontext(t), storage.NeuerKartenschreiber(db),
		storage.NeueFeldherkunft(db), karte
}

// **Wert, Feldherkunft und Bildangaben stehen nach einem Zug da.**
func TestEineBilduebernahmeSchreibtAllesDrei(t *testing.T) {
	t.Parallel()

	ctx, schreiber, leser, karte := mitBildkarte(t)

	if _, _, fehler := schreiber.FelderSetzenMitHerkunft(ctx, karte,
		[]domain.Value{{Field: "image", Text: "bild1.png"}},
		"ki", []domain.KIFeldherkunft{eineBildherkunft()},
		map[string][]domain.KIBildparameter{"image": {
			{Name: "groesse", Wert: "1024x1536"},
			{Name: "guete", Wert: "medium"},
		}}, ""); fehler != nil {
		t.Fatalf("FelderSetzenMitHerkunft: %v", fehler)
	}

	herkuenfte, fehler := leser.Lesen(ctx, karte)
	if fehler != nil {
		t.Fatalf("Lesen: %v", fehler)
	}

	if len(herkuenfte) != 1 || herkuenfte[0].Feld != "image" {
		t.Fatalf("die Herkünfte sind %+v", herkuenfte)
	}

	angaben, fehler := leser.Bildangaben(ctx, karte, "image")
	if fehler != nil {
		t.Fatalf("Bildangaben: %v", fehler)
	}

	if len(angaben) != 2 || angaben[0].Name != "groesse" ||
		angaben[1].Wert != "medium" {
		t.Fatalf("die Angaben sind %+v", angaben)
	}
}

// **Ein zweiter Lauf lässt nichts vom ersten stehen.**
func TestEinZweiterLaufRaeumtDieAltenAngaben(t *testing.T) {
	t.Parallel()

	ctx, schreiber, leser, karte := mitBildkarte(t)

	setzen := func(angaben []domain.KIBildparameter) {
		t.Helper()

		if _, _, fehler := schreiber.FelderSetzenMitHerkunft(ctx, karte,
			[]domain.Value{{Field: "image", Text: "bild.png"}},
			"ki", []domain.KIFeldherkunft{eineBildherkunft()},
			map[string][]domain.KIBildparameter{"image": angaben}, ""); fehler != nil {
			t.Fatalf("FelderSetzenMitHerkunft: %v", fehler)
		}
	}

	setzen([]domain.KIBildparameter{
		{Name: "guete", Wert: "medium"}, {Name: "seed", Wert: "4711"},
	})

	setzen([]domain.KIBildparameter{{Name: "guete", Wert: "high"}})

	angaben, fehler := leser.Bildangaben(ctx, karte, "image")
	if fehler != nil {
		t.Fatalf("Bildangaben: %v", fehler)
	}

	if len(angaben) != 1 || angaben[0].Wert != "high" {
		t.Fatalf("die Angaben sind %+v — der Seed von vorhin steht noch da",
			angaben)
	}
}

// **Ohne Herkunft ist es keine Übernahme.**
func TestOhneHerkunftKeineBilduebernahme(t *testing.T) {
	t.Parallel()

	ctx, schreiber, _, karte := mitBildkarte(t)

	if _, _, fehler := schreiber.FelderSetzenMitHerkunft(ctx, karte,
		[]domain.Value{{Field: "image", Text: "bild.png"}},
		"ki", nil, nil, ""); !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("die Übernahme ohne Herkunft kam durch: %v", fehler)
	}
}

// **Eine Angabe ohne Namen macht den ganzen Zug rückgängig.**
//
// **Das ist der Lauf, für den es die Transaktion gibt.** Der Wert steht schon
// in der Anweisung, die Feldherkunft auch — und trotzdem darf nichts
// zurückbleiben.
func TestEineAngabeOhneNamenDrehtDenZugZurueck(t *testing.T) {
	t.Parallel()

	ctx, schreiber, leser, karte := mitBildkarte(t)

	if _, _, fehler := schreiber.FelderSetzenMitHerkunft(ctx, karte,
		[]domain.Value{{Field: "image", Text: "bild.png"}},
		"ki", []domain.KIFeldherkunft{eineBildherkunft()},
		map[string][]domain.KIBildparameter{"image": {
			{Name: "", Wert: "ohne Namen"},
		}}, ""); !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("die namenlose Angabe kam durch: %v", fehler)
	}

	herkuenfte, fehler := leser.Lesen(ctx, karte)
	if fehler != nil {
		t.Fatalf("Lesen: %v", fehler)
	}

	if len(herkuenfte) != 0 {
		t.Fatalf("die Feldherkunft blieb stehen: %+v", herkuenfte)
	}
}

// **Ohne Profil im Kontext gibt es keine Auskunft.**
func TestOhneProfilKeineBildangaben(t *testing.T) {
	t.Parallel()

	leser := storage.NeueFeldherkunft(frischeDatenbank(t))

	if _, fehler := leser.Bildangaben(t.Context(), 1, "image"); fehler == nil {
		t.Fatal("die Auskunft kam ohne Profil")
	}
}

// **Ohne Tafel gibt es keine Auskunft und keine Übernahme.**
//
// # Warum die Tafel fällt und nicht etwas anderes
//
// **Weil sich ein SQL-Fehler sonst nicht herstellen lässt.** Die Wege hier
// sind kurz und die Anweisungen fest; was schieflaufen kann, ist die
// Datenbank selbst — und die Gestalt, in der das im Betrieb vorkommt, ist
// eine Migration, die nicht gelaufen ist.
//
// **Vier Zweige in einem Lauf**: räumen, schreiben, abfragen, lesen. Ohne
// ihn sähe keiner von ihnen je einen Fehler, und die vier Meldungen wären
// vier Zusagen, die niemand geprüft hat (§10).
func TestOhneTafelGibtEsKeineBildangaben(t *testing.T) {
	db := frischeDatenbank(t)
	karte := eineKarte(t, db, domain.Value{Field: "image", Text: "alt.png"})
	ctx := probekontext(t)

	if _, fehler := db.ExecContext(ctx, `DROP TABLE ki_bildherkunft`); fehler != nil {
		t.Fatalf("die Tafel ließ sich nicht fallen lassen: %v", fehler)
	}

	if _, fehler := storage.NeueFeldherkunft(db).Bildangaben(
		ctx, karte, "image"); fehler == nil {
		t.Error("die Auskunft kam ohne Tafel")
	}

	if _, _, fehler := storage.NeuerKartenschreiber(db).FelderSetzenMitHerkunft(
		ctx, karte, []domain.Value{{Field: "image", Text: "neu.png"}},
		"ki", []domain.KIFeldherkunft{eineBildherkunft()},
		map[string][]domain.KIBildparameter{"image": {
			{Name: "guete", Wert: "medium"},
		}}, ""); fehler == nil {
		t.Error("die Übernahme gelang ohne Tafel")
	}
}

// **Eine Zeile, die sich nicht lesen lässt, ist ein Fehler.**
//
// # Warum die Tafel dafür neu gebaut wird
//
// **Weil `STRICT` und `NOT NULL` es sonst verhindern** — und das ist gut so.
// Der Fall entsteht nicht durch dieses Programm, sondern durch einen Bestand,
// den jemand von Hand bearbeitet hat; der Bestand ist eine Datei, die
// weitergegeben wird (§8). Der Weg soll das melden und nicht eine leere
// Angabe zurückgeben.
func TestEineUnlesbareBildangabeIstEinFehler(t *testing.T) {
	db := frischeDatenbank(t)
	karte := eineKarte(t, db, domain.Value{Field: "image", Text: "alt.png"})
	ctx := probekontext(t)

	if _, fehler := db.ExecContext(ctx,
		`DROP TABLE ki_bildherkunft;
		 CREATE TABLE ki_bildherkunft (
		   profil_id INTEGER, karte_id INTEGER, feld TEXT, name TEXT, wert TEXT);
		 INSERT INTO ki_bildherkunft VALUES (1, `+
			`(SELECT id FROM cards LIMIT 1), 'image', 'guete', NULL)`); fehler != nil {
		t.Fatalf("die Tafel ließ sich nicht neu bauen: %v", fehler)
	}

	if _, fehler := storage.NeueFeldherkunft(db).Bildangaben(
		ctx, karte, "image"); fehler == nil {
		t.Error("die leere Angabe kam als Auskunft zurück")
	}
}

// **Die drei Stellen, an denen eine Bildübernahme scheitern darf.**
//
// Jede von ihnen trägt eine eigene Meldung, und jede von ihnen lässt den
// Bestand so zurück, wie er war. Ohne diesen Lauf wären es drei Zusagen, die
// niemand geprüft hat (§10).
func TestEineBilduebernahmeScheitertAnDreiStellen(t *testing.T) {
	faelle := map[string]struct {
		ohneProfil bool
		verbiegen  string
		doppelt    bool
	}{
		// **Ohne Profil weiß niemand, wessen Bild das wäre.**
		"ohne Profil": {ohneProfil: true},

		// **Eine doppelte Feldherkunft** — dieselbe Gestalt wie in A11: eine
		// Übernahme gibt einem Feld einen Wert und nicht zwei.
		"mit doppelter Feldherkunft": {doppelt: true},

		// **Eine Tafel, die löschen lässt und nicht schreiben.** Damit
		// scheitert die Angabe **nach** dem Räumen — der Zweig, den ein
		// fehlendes `ki_bildherkunft` nie erreicht.
		"mit halber Bildtafel": {verbiegen: `DROP TABLE ki_bildherkunft;
			 CREATE TABLE ki_bildherkunft (
			   profil_id INTEGER, karte_id INTEGER, feld TEXT, name TEXT)`},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			db := frischeDatenbank(t)
			karte := eineKarte(t, db, domain.Value{Field: "image", Text: "alt.png"})
			ctx := probekontext(t)

			if fall.verbiegen != "" {
				if _, fehler := db.ExecContext(ctx, fall.verbiegen); fehler != nil {
					t.Fatalf("verbiegen: %v", fehler)
				}
			}

			herkuenfte := []domain.KIFeldherkunft{eineBildherkunft()}
			if fall.doppelt {
				herkuenfte = append(herkuenfte, eineBildherkunft())
			}

			if fall.ohneProfil {
				ctx = t.Context()
			}

			if _, _, fehler := storage.NeuerKartenschreiber(db).
				FelderSetzenMitHerkunft(ctx, karte,
					[]domain.Value{{Field: "image", Text: "neu.png"}},
					"ki", herkuenfte,
					map[string][]domain.KIBildparameter{"image": {
						{Name: "guete", Wert: "medium"},
					}}, ""); fehler == nil {
				t.Fatalf("%q kam durch", name)
			}

			// **Und der Wert steht noch, wie er war** — alles oder nichts.
			karten, _ := storage.NeuerSetLeser(db).KarteLesen( //nolint:errcheck // unten geprüft
				probekontext(t), karte)
			if wert := wertVon(*karten, "image"); wert != "alt.png" {
				t.Errorf("das Feld trägt %q", wert)
			}
		})
	}
}
