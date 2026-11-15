// **Die Massennaht** — Werkstück 6 von Phase 4.5
// (#1292).
package app_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// dreiZeilen legt drei Exemplare an und gibt Naht und Zeilen.
func dreiZeilen(t *testing.T) (*app.Massennaht, []app.MassenzeileDTO) {
	t.Helper()

	t.Setenv("HOME", t.TempDir())

	db := frischerBestand(t)

	sammlung, err := app.SammlungsnahtUeber(db)
	if err != nil {
		t.Fatalf("SammlungsnahtUeber: %v", err)
	}

	zeilen := make([]app.MassenzeileDTO, 0, 3)

	for zaehler, zustand := range []string{"nearMint", "played", "good"} {
		id, fehler := sammlung.ExemplarAnlegen(probekontext(t), app.ExemplarDTO{
			Printing: "druck-" + string(rune('a'+zaehler)), Menge: 1,
			Zustand: zustand, Sprache: "en", Lagerort: "Kiste 4",
		})
		if fehler != nil {
			t.Fatalf("anlegen: %v", fehler)
		}

		zeilen = append(zeilen, app.MassenzeileDTO{ID: id, Fassung: 1})
	}

	return app.MassennahtUeber(db), zeilen
}

// **Der ganze Weg: ändern, löschen.**
func TestDieMassennahtFuehrtDenGanzenWeg(t *testing.T) {
	naht, zeilen := dreiZeilen(t)

	ordner := "Ordner Rot"
	zustand := "poor"
	menge := 3
	wunsch := true

	bericht, err := naht.Aendern(probekontext(t), app.MassenaenderungDTO{
		Zeilen: zeilen, Lagerort: &ordner, Zustand: &zustand,
		Menge: &menge, Wunsch: &wunsch,
	})
	if err != nil {
		t.Fatalf("Aendern: %v", err)
	}

	if bericht.Geaendert != 3 {
		t.Fatalf("%d geändert, erwartet 3: %v", bericht.Geaendert, bericht.Befunde)
	}

	// **Der Bericht nennt den Stand vorher samt neuer Fassung** — damit
	// sich die Änderung zurücknehmen lässt (DOM-03.AK-03, #1698).
	if len(bericht.Vorher) != 3 || bericht.Vorher[0].ID != zeilen[0].ID ||
		bericht.Vorher[0].Fassung != 2 || bericht.Vorher[0].Lagerort != "Kiste 4" ||
		bericht.Vorher[0].Zustand != "nearMint" || bericht.Vorher[0].Menge != 1 || bericht.Vorher[0].Wunsch {
		t.Errorf("vorher: %+v", bericht.Vorher)
	}

	// **Die Fassungen sind gestiegen** — zum Löschen braucht es die neuen.
	for stelle := range zeilen {
		zeilen[stelle].Fassung = 2
	}

	geloescht, err := naht.Loeschen(probekontext(t), zeilen)
	if err != nil {
		t.Fatalf("Loeschen: %v", err)
	}

	if geloescht.Geaendert != 3 {
		t.Errorf("%d gelöscht, erwartet 3: %v", geloescht.Geaendert, geloescht.Befunde)
	}

	// **Ein Löschen nennt keinen Stand vorher** — und die Liste ist trotzdem
	// eine Liste, kein `null`.
	if geloescht.Vorher == nil || len(geloescht.Vorher) != 0 {
		t.Errorf("vorher nach dem Löschen: %+v", geloescht.Vorher)
	}
}

// **Ein Preisschub geht über die Naht — gegen eine Gegenstelle.**
func TestDerPreisschubUeberDieNaht(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	dienst := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			// **Die Antwort nennt die erfragte Kennung.** Der Klient
			// vergleicht sie; eine feste ließe den Vergleich scheitern, und
			// der Lauf prüfte die Wache statt den Weg.
			schreibeRoh(t, schreiben,
				`{"object":"card","id":"probe-druck","name":"Probe","lang":"en",`+
					`"layout":"normal","set":"probe","set_name":"Probe",`+
					`"collector_number":"1","rarity":"common","type_line":"Instant",`+
					`"prices":{"eur":"4.17"}}`)
		}))
	t.Cleanup(dienst.Close)

	vorher := app.MitScryfalladresse(dienst.URL)
	t.Cleanup(func() { app.MitScryfalladresse(vorher) })

	db := frischerBestand(t)

	sammlung, err := app.SammlungsnahtUeber(db)
	if err != nil {
		t.Fatalf("SammlungsnahtUeber: %v", err)
	}

	if _, fehler := sammlung.ExemplarAnlegen(probekontext(t), app.ExemplarDTO{
		Printing: "probe-druck", Menge: 1, Zustand: "nearMint", Sprache: "en",
	}); fehler != nil {
		t.Fatalf("anlegen: %v", fehler)
	}

	schub, err := app.MassennahtUeber(db).PreiseInMasse(probekontext(t), 0)
	if err != nil {
		t.Fatalf("PreiseInMasse: %v", err)
	}

	if schub.Geholt != 1 || !schub.Fertig {
		t.Errorf("der Schub ist %+v", schub)
	}
}

// **Ohne Zuhause gibt es keinen Preisschub — aber Massenvorgänge.**
//
// §16, Regel 5: ohne Netz arbeitet die Anwendung weiter. Die Preise brauchen
// ein Zuhause für den Zwischenspeicher, das Ändern nicht.
func TestOhneZuhauseKeinPreisschubAberMasse(t *testing.T) {
	t.Setenv("HOME", "")
	t.Setenv("XDG_CONFIG_HOME", "")

	naht := app.MassennahtUeber(frischerBestand(t))

	if _, err := naht.PreiseInMasse(probekontext(t), 0); err == nil {
		t.Error("ohne Zuhause kam ein Schub heraus")
	}

	// **Und das Ändern geht trotzdem** — es meldet nur, dass die Zeile fehlt.
	bericht, err := naht.Aendern(probekontext(t), app.MassenaenderungDTO{
		Zeilen: []app.MassenzeileDTO{{ID: 1, Fassung: 1}}, Menge: neueZahl(3),
	})
	if err != nil {
		t.Fatalf("Aendern: %v", err)
	}

	if len(bericht.Befunde) != 1 {
		t.Errorf("%d Befunde, erwartet 1", len(bericht.Befunde))
	}
}

// **Ein fachliches Nein kommt als Fehler heraus.**
func TestEinFachlichesNeinDerMassennaht(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	naht := app.MassennahtUeber(frischerBestand(t))

	if _, err := naht.Aendern(probekontext(t), app.MassenaenderungDTO{}); err == nil {
		t.Error("ohne Zeilen kam kein Fehler")
	}

	if _, err := naht.Loeschen(probekontext(t), nil); err == nil {
		t.Error("Löschen ohne Zeilen gab keinen Fehler")
	}
}

// **Ohne eingerichtete Naht sagen alle drei Bindungen es selbst.**
func TestOhneEingerichteteMasse(t *testing.T) {
	anwendung := app.NeueApp(filepath.Join(t.TempDir(), "bestand.db"))

	if anwendung.MasseAendern(t.Context(), app.MassenaenderungDTO{}).Fehler == nil {
		t.Error("MasseAendern gab keinen Fehler")
	}

	if anwendung.MasseLoeschen(t.Context(), nil).Fehler == nil {
		t.Error("MasseLoeschen gab keinen Fehler")
	}

	if anwendung.MassePreise(t.Context(), 0).Fehler == nil {
		t.Error("MassePreise gab keinen Fehler")
	}
}

// **Ein Fehler der Naht kommt an jeder Bindung an — und ein Erfolg auch.**
func TestDieBindungenDerMasse(t *testing.T) {
	anwendung := app.NeueApp(filepath.Join(t.TempDir(), "bestand.db"))
	anwendung.MitMasse(stolperndeMasse{})

	bindungen := app.NeueBindungen(anwendung)

	if bindungen.MasseAendern(t.Context(), app.MassenaenderungDTO{}).Fehler == nil {
		t.Error("MasseAendern gab keinen Fehler")
	}

	if bindungen.MasseLoeschen(t.Context(), nil).Fehler == nil {
		t.Error("MasseLoeschen gab keinen Fehler")
	}

	if bindungen.MassePreise(t.Context(), 0).Fehler == nil {
		t.Error("MassePreise gab keinen Fehler")
	}

	anwendung.MitMasse(willigeMasse{})

	if bindungen.MasseAendern(t.Context(), app.MassenaenderungDTO{}).Bericht == nil {
		t.Error("MasseAendern gab keinen Bericht")
	}

	if bindungen.MasseLoeschen(t.Context(), nil).Bericht == nil {
		t.Error("MasseLoeschen gab keinen Bericht")
	}

	if bindungen.MassePreise(t.Context(), 0).Schub == nil {
		t.Error("MassePreise gab keinen Schub")
	}
}

// neueZahl gibt einen Zeiger auf eine Zahl.
func neueZahl(wert int) *int { return &wert }

// stolperndeMasse scheitert auf allen drei Wegen.
type stolperndeMasse struct{}

func (stolperndeMasse) Aendern(context.Context, app.MassenaenderungDTO) (
	*app.MassenberichtDTO, error,
) {
	return nil, errors.New("die Platte klemmt")
}

func (stolperndeMasse) Loeschen(context.Context, []app.MassenzeileDTO) (
	*app.MassenberichtDTO, error,
) {
	return nil, errors.New("die Platte klemmt")
}

func (stolperndeMasse) PreiseInMasse(context.Context, int) (*app.PreisschubDTO, error) {
	return nil, errors.New("kein Netz")
}

// willigeMasse gibt auf allen drei Wegen ein Ergebnis.
type willigeMasse struct{}

func (willigeMasse) Aendern(context.Context, app.MassenaenderungDTO) (
	*app.MassenberichtDTO, error,
) {
	return &app.MassenberichtDTO{Geaendert: 1, Befunde: []string{}}, nil
}

func (willigeMasse) Loeschen(context.Context, []app.MassenzeileDTO) (
	*app.MassenberichtDTO, error,
) {
	return &app.MassenberichtDTO{Geaendert: 1, Befunde: []string{}}, nil
}

func (willigeMasse) PreiseInMasse(context.Context, int) (*app.PreisschubDTO, error) {
	return &app.PreisschubDTO{Fertig: true, Befunde: []string{}}, nil
}

// **Ein Fehler des Schubs kommt an der Naht an.**
//
// Der Zweig, den kein anderer Lauf betritt: das Zuhause steht, der Bestand
// klemmt.
func TestEinFehlerDesPreisschubsAnDerNaht(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	db := frischerBestand(t)
	naht := app.MassennahtUeber(db)

	if fehler := db.Close(); fehler != nil {
		t.Fatalf("schließen: %v", fehler)
	}

	if _, err := naht.PreiseInMasse(probekontext(t), 0); err == nil {
		t.Error("auf geschlossenem Bestand kam ein Schub heraus")
	}
}
