package storage_test

// **Der Nachzug schreibt auf dem Stand, aus dem er rechnete, und unter der
// Bedingung nur, was gesetzt wurde** — Gesamtreview über 0.9.0, Befunde 1
// und 2 (#2382).
//
// Seit #2396 schreibt eine Eingabe auch die Felder, die ihr Wertskript
// nachzieht. Die Bedingung des Fensters (U02b, #2179) kennt nur das Feld, das
// es bearbeitet — ein nachgezogenes prüfte sie gegen fremde Werte und wies ab.
// Und gerechnet wurde vor der Transaktion: ein zweiter Schreiber dazwischen,
// und der nachgezogene Wert passte zu keinem Stand der Karte.

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

func TestEinNachgezogenerWertUmgehtDieBekanntenWerte(t *testing.T) {
	db := frischeDatenbank(t)
	karteID := eineKarte(t, db,
		domain.Value{Field: "zeile", Text: "[alt]"}, domain.Value{Field: "echo", Text: "[alt]!"})
	ctx := storage.MitBekanntenWerten(probekontext(t), []string{"[alt]", "neu"})

	if _, _, fehler := storage.NeuerKartenschreiber(db).FelderSetzenVonAussen(ctx, karteID, []domain.Value{
		{Field: "zeile", Text: "[neu]"},
		{Field: "echo", Text: "[neu]!", Nachgezogen: true},
	}, "", ""); fehler != nil {
		t.Fatalf("FelderSetzen: %v — der nachgezogene Wert galt als fremd", fehler)
	}

	karte, fehler := storage.NeuerSetLeser(db).KarteLesen(probekontext(t), karteID)
	if fehler != nil {
		t.Fatalf("lesen: %v", fehler)
	}

	if want := []domain.Value{{Field: "zeile", Text: "[neu]"}, {Field: "echo", Text: "[neu]!"}}; !gleicheWerte(karte.Values, want) {
		t.Errorf("im Bestand %v, erwartet %v", karte.Values, want)
	}
}

// **Ein gesetzter Wert bleibt unter der Bedingung** — auch der zweite eines
// Verbunds; sonst wäre die Korrektur ein Abschalten.
func TestEinGesetzterWertBleibtUnterDerBedingung(t *testing.T) {
	db := frischeDatenbank(t)
	karteID := eineKarte(t, db,
		domain.Value{Field: "zeile", Text: "[alt]"}, domain.Value{Field: "teil", Text: "fremd"})
	ctx := storage.MitBekanntenWerten(probekontext(t), []string{"[alt]", "neu"})

	_, _, fehler := storage.NeuerKartenschreiber(db).FelderSetzenVonAussen(ctx, karteID, []domain.Value{
		{Field: "zeile", Text: "[neu]"},
		{Field: "teil", Text: "neu"},
	}, "", "")

	var feldfehler *domain.FieldError
	if !errors.As(fehler, &feldfehler) || feldfehler.Key != domain.SchluesselFremderWert {
		t.Fatalf("FelderSetzen: %v, erwartet %q", fehler, domain.SchluesselFremderWert)
	}
}

func TestMehrereFelderSchreibenNurAufIhremStand(t *testing.T) {
	db := frischeDatenbank(t)
	karteID := eineKarte(t, db, domain.Value{Field: "zeile", Text: "[alt]"})
	leser := storage.NeuerSetLeser(db)
	schreiber := storage.NeuerKartenschreiber(db)

	vorher, fehler := leser.KarteLesen(probekontext(t), karteID)
	if fehler != nil {
		t.Fatalf("lesen: %v", fehler)
	}

	// **Ein zweiter Schreiber dazwischen.**
	if _, _, dazwischen := schreiber.FelderSetzenVonAussen(probekontext(t), karteID,
		[]domain.Value{{Field: "andere", Text: "dazwischen"}}, "", ""); dazwischen != nil {
		t.Fatalf("dazwischen: %v", dazwischen)
	}

	_, _, fehler = schreiber.FelderSetzenVonAussen(probekontext(t), karteID,
		[]domain.Value{{Field: "zeile", Text: "[neu]"}}, "", vorher.Stand())
	if !errors.Is(fehler, domain.ErrConflict) {
		t.Fatalf("auf überholtem Stand: %v, erwartet ErrConflict", fehler)
	}

	nachher, fehler := leser.KarteLesen(probekontext(t), karteID)
	if fehler != nil {
		t.Fatalf("lesen: %v", fehler)
	}

	if _, _, fehler := schreiber.FelderSetzenVonAussen(probekontext(t), karteID,
		[]domain.Value{{Field: "zeile", Text: "[neu]"}}, "", nachher.Stand()); fehler != nil {
		t.Fatalf("auf dem gelesenen Stand: %v", fehler)
	}
}

// **Auch in einer Gruppe** — die Übernahme eines KI-Vorschlags über mehrere
// Karten schreibt nachgezogene Werte genauso.
func TestEinNachgezogenerWertUmgehtDieBedingungAuchInEinerGruppe(t *testing.T) {
	db := frischeDatenbank(t)
	karteID := eineKarte(t, db,
		domain.Value{Field: "zeile", Text: "[alt]"}, domain.Value{Field: "echo", Text: "[alt]!"})
	ctx := storage.MitBekanntenWerten(probekontext(t), []string{"[alt]", "neu"})

	if fehler := storage.NeuerKartenschreiber(db).GruppeSetzenMitHerkunft(ctx, []domain.Kartenschritt{{
		KarteID: karteID,
		Werte: []domain.Value{
			{Field: "zeile", Text: "[neu]"},
			{Field: "echo", Text: "[neu]!", Nachgezogen: true},
		},
		Herkunft:   "ki",
		Herkuenfte: []domain.KIFeldherkunft{eineFeldherkunft("zeile")},
	}}); fehler != nil {
		t.Fatalf("Gruppe: %v — der nachgezogene Wert galt als fremd", fehler)
	}

	if wert := kartenfeld(t, db, karteID, "echo"); wert != "[neu]!" {
		t.Errorf("echo trägt %q, erwartet [neu]!", wert)
	}
}

// **Und auf den Setangaben, die die Rechnung las** — #2533: ändert ein
// anderer Klient eine Setangabe zwischen Rechnung und Schreiben, bleibt der
// Kartenstand gleich, und ein Wertskript, das sie liest, schrieb sein
// überholtes Ergebnis.
func TestEineAenderungDerSetangabenUeberholtDenRechenstand(t *testing.T) {
	db := frischeDatenbank(t)
	karteID := eineKarte(t, db, domain.Value{Field: "zeile", Text: "[alt]"})
	leser := storage.NeuerSetLeser(db)
	schreiber := storage.NeuerKartenschreiber(db)

	rechenstand := func() string {
		karte, fehler := leser.KarteLesen(probekontext(t), karteID)
		if fehler != nil {
			t.Fatalf("lesen: %v", fehler)
		}

		setwerte, fehler := leser.Setwerte(probekontext(t), karte.SetID)
		if fehler != nil {
			t.Fatalf("Setangaben: %v", fehler)
		}

		return domain.Rechenstand(karte.Stand(), setwerte)
	}

	gerechnet := rechenstand()

	karte, _ := leser.KarteLesen(probekontext(t), karteID) //nolint:errcheck // oben gelesen
	if _, _, fehler := schreiber.SetwertSetzenMitVorwert(probekontext(t), karte.SetID, "copyright", "2026"); fehler != nil {
		t.Fatalf("Setangabe dazwischen: %v", fehler)
	}

	_, _, fehler := schreiber.FelderSetzenVonAussen(probekontext(t), karteID,
		[]domain.Value{{Field: "zeile", Text: "[neu]"}}, "", gerechnet)
	if !errors.Is(fehler, domain.ErrConflict) {
		t.Fatalf("auf überholten Setangaben: %v, erwartet ErrConflict", fehler)
	}

	if _, _, neu := schreiber.FelderSetzenVonAussen(probekontext(t), karteID,
		[]domain.Value{{Field: "zeile", Text: "[neu]"}}, "", rechenstand()); neu != nil {
		t.Fatalf("auf dem gelesenen Rechenstand: %v", neu)
	}
}

// gleicheWerte vergleicht Feld und Text, ohne die Marke des Nachzugs.
func gleicheWerte(ist, soll []domain.Value) bool {
	if len(ist) != len(soll) {
		return false
	}

	for i := range ist {
		if ist[i].Field != soll[i].Field || ist[i].Text != soll[i].Text {
			return false
		}
	}

	return true
}
