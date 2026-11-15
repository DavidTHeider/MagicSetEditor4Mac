package rest_test

// **Das Auftragsbuch für sich** — FMT-03.AK-02 (#1692): Aufbewahrung,
// Obergrenze und ein Lauf, der nicht zu Ende kommt. Über HTTP prüft es
// `auftraege_test.go`.

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

// fertigerLauf antwortet gleich.
func fertigerLauf(context.Context, rest.Auftragseingabe, domain.Fortschrittsmelder) rest.Auftragsergebnis {
	return rest.Auftragsergebnis{Status: http.StatusCreated, Art: "application/json", Rumpf: []byte("{}")}
}

// bisBeendet fragt, bis der Auftrag nicht mehr läuft.
func bisBeendet(t *testing.T, buch *rest.Auftragsbuch, id string) rest.AuftragDTO {
	t.Helper()

	for range 400 {
		if auftrag, fehler := buch.Auftrag(t.Context(), id); fehler == nil && auftrag.Stand != "laeuft" {
			return auftrag
		}

		time.Sleep(5 * time.Millisecond)
	}

	t.Fatalf("der Auftrag %s endet nicht", id)

	return rest.AuftragDTO{}
}

func TestEinBeendeterAuftragBleibtEineStunde(t *testing.T) {
	jetzt := time.Date(2026, time.September, 29, 12, 0, 0, 0, time.UTC)
	uhr := jetzt

	buch := rest.NeuesAuftragsbuch(func() time.Time { return uhr })

	auftrag, fehler := buch.Annehmen(t.Context(), rest.Auftragseingabe{Weg: "POST /sets"}, fertigerLauf)
	if fehler != nil {
		t.Fatal(fehler)
	}

	bisBeendet(t, buch, auftrag.ID)

	uhr = jetzt.Add(59 * time.Minute)
	if _, fehler := buch.Auftrag(t.Context(), auftrag.ID); fehler != nil {
		t.Errorf("nach 59 Minuten: %v", fehler)
	}

	uhr = jetzt.Add(61 * time.Minute)
	if _, fehler := buch.Auftrag(t.Context(), auftrag.ID); !errors.Is(fehler, domain.ErrNotFound) {
		t.Errorf("nach 61 Minuten: %v", fehler)
	}
}

func TestDasBuchBehaeltSechzehnBeendete(t *testing.T) {
	buch := rest.NeuesAuftragsbuch(time.Now)

	var erster string

	for nummer := range 17 {
		auftrag, fehler := buch.Annehmen(t.Context(), rest.Auftragseingabe{Weg: "POST /sets"}, fertigerLauf)
		if fehler != nil {
			t.Fatal(fehler)
		}

		if nummer == 0 {
			erster = auftrag.ID
		}

		bisBeendet(t, buch, auftrag.ID)
	}

	liste, _ := buch.Auftraege(t.Context()) //nolint:errcheck // das Buch scheitert nicht
	if len(liste) != 16 || strings.Contains(auftragsliste(liste), erster) {
		t.Errorf("%d behalten, der erste dabei: %v", len(liste), strings.Contains(auftragsliste(liste), erster))
	}
}

func auftragsliste(liste []rest.AuftragDTO) string {
	kennungen := make([]string, 0, len(liste))
	for _, auftrag := range liste {
		kennungen = append(kennungen, auftrag.ID)
	}

	return strings.Join(kennungen, ",")
}

// **Eine zu große Antwort bricht ab wie am Draht** — und der Auftrag sagt es.
func TestEineZuGrosseAntwortLaesstDenAuftragScheitern(t *testing.T) {
	buch := rest.NeuesAuftragsbuch(time.Now)

	auftrag, fehler := buch.Annehmen(t.Context(), rest.Auftragseingabe{Weg: "POST /sets"},
		func(context.Context, rest.Auftragseingabe, domain.Fortschrittsmelder) rest.Auftragsergebnis {
			panic(http.ErrAbortHandler)
		})
	if fehler != nil {
		t.Fatal(fehler)
	}

	if beendet := bisBeendet(t, buch, auftrag.ID); beendet.Stand != "gescheitert" || beendet.Status != http.StatusInternalServerError {
		t.Fatalf("beendet: %+v", beendet)
	}

	ergebnis, _ := buch.Ergebnis(t.Context(), auftrag.ID) //nolint:errcheck // beendet
	if !strings.Contains(string(ergebnis.Rumpf), "größer als erlaubt") {
		t.Errorf("Ergebnis: %s", ergebnis.Rumpf)
	}
}

// **Jeder Auftragsweg ist ein eingetragener Weg** — ein Name, den der Dienst
// nicht kennt, träfe keine Anfrage, und die Bitte liefe still ins Leere.
func TestJederAuftragswegIstEinWeg(t *testing.T) {
	for _, weg := range rest.Auftragswege() {
		methode, muster, _ := strings.Cut(weg, " ")

		if !inDerListe(methode, muster) {
			t.Errorf("%s ist kein Weg des Dienstes", weg)
		}
	}
}
