package storage_test

// **Eine Zusage gilt einer Zeile, die sie erfüllt — je Variante** —
// #2492.
//
// Unter „Zusagen" ließ sich jede Besitzzeile des Drucks festhalten, auch eine
// in anderem Zustand. Verbucht ging das gewählte Exemplar weg, drüben kam die
// abgesprochene Zeile an, und die Gegenzeichnung fiel still aus (Hüllenprobe
// `nat04handel`). **Hier, in derselben Transaktion wie das Eintragen**: ein
// Exemplar, das zu keiner Zeile passt, wird nicht festgehalten, und jede
// Variante hat ihre eigene Grenze.

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// ueber ist ein Vertrag mit einer Zeile, die dieses Exemplar erfüllt.
func ueber(exemplar *domain.Exemplar, menge int) domain.Zusagevertrag {
	return domain.Zusagevertrag{{
		Printing: exemplar.Printing, DruckID: exemplar.DruckID, Menge: menge,
		Guete: string(exemplar.Zustand), Sprache: exemplar.Sprache,
		Finish: string(exemplar.Finish),
	}}
}

// wieDreiStueck ist ein Vertrag über die Variante von `dreiStueck` — für
// die Läufe, die das Exemplar erst im Aufruf anlegen.
func wieDreiStueck(menge int) domain.Zusagevertrag {
	return ueber(&domain.Exemplar{
		Printing: "abc-123", Zustand: domain.ZustandNearMint, Sprache: "de",
	}, menge)
}

// einExcellent legt neben `dreiStueck` eine Zeile desselben Drucks in
// anderem Zustand an.
func einExcellent(t *testing.T, db *storage.Sammlungsspeicher) *domain.Exemplar {
	t.Helper()

	exemplar := &domain.Exemplar{
		Printing: "abc-123", Menge: 1, Zustand: domain.ZustandExcellent, Sprache: "de",
	}

	kennung, fehler := db.Anlegen(probekontext(t), exemplar)
	if fehler != nil {
		t.Fatalf("anlegen: %v", fehler)
	}

	exemplar.ID, exemplar.Fassung = kennung, 1

	return exemplar
}

func TestEinExemplarInAnderemZustandWirdNichtFestgehalten(t *testing.T) {
	t.Parallel()

	db := frischeDatenbank(t)
	ctx := probekontext(t)
	bestand := storage.NeuerSammlungsspeicher(db)
	vereinbart, anderes := dreiStueck(t, bestand), einExcellent(t, bestand)
	ablage := storage.NeueReservierungsablage(db)

	fehler := ablage.Festhalten(ctx, eineZusage(anderes, "h1", 1), ueber(vereinbart, 1))

	var feld *domain.FieldError
	if !errors.As(fehler, &feld) || feld.Key != domain.SchluesselZusageUnpassend {
		t.Fatalf("es kam %v", fehler)
	}

	if gehalten, _ := ablage.Reservierungen(ctx, anderes.ID); len(gehalten) != 0 { //nolint:errcheck // gezählt wird
		t.Errorf("festgehalten wurde %+v", gehalten)
	}

	// **Und das vereinbarte geht** — sonst prüfte der Lauf nur, dass
	// überhaupt etwas abgewiesen wird.
	if fehler := ablage.Festhalten(ctx, eineZusage(vereinbart, "h1", 1), ueber(vereinbart, 1)); fehler != nil {
		t.Errorf("das vereinbarte: %v", fehler)
	}
}

func TestJedeVarianteHatIhreEigeneGrenze(t *testing.T) {
	t.Parallel()

	db := frischeDatenbank(t)
	ctx := probekontext(t)
	bestand := storage.NeuerSammlungsspeicher(db)
	nearMint, excellent := dreiStueck(t, bestand), einExcellent(t, bestand)
	ablage := storage.NeueReservierungsablage(db)

	// **Ein Stück jeder Variante** — zusammen zwei, aber nicht zwei von einer.
	vertrag := append(ueber(nearMint, 1), ueber(excellent, 1)...)

	if fehler := ablage.Festhalten(ctx, eineZusage(nearMint, "h1", 1), vertrag); fehler != nil {
		t.Fatalf("das erste: %v", fehler)
	}

	if fehler := ablage.Festhalten(ctx, eineZusage(nearMint, "h1", 1), vertrag); !errors.Is(fehler, domain.ErrConflict) {
		t.Errorf("ein zweites Near Mint: %v", fehler)
	}

	if fehler := ablage.Festhalten(ctx, eineZusage(excellent, "h1", 1), vertrag); fehler != nil {
		t.Errorf("das Excellent: %v", fehler)
	}
}
