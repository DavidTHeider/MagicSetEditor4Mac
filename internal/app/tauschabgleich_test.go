package app_test

// **Der Tauschabgleich an der Naht** — Werkstück `T4` von Phase 5.6
// (#1469).
//
// # Was hier gemessen wird
//
// **Dass die Naht überträgt, ohne zu entscheiden.** Was ein Treffer ist,
// steht in `domain`; woraus er gebaut wird, in `service`. Hier steht, dass
// die Angaben heil durchkommen — samt der **Erklärung**, warum eine Zeile ein
// Treffer ist.
//
// **Und dass der fremde Stand dieselbe Gestalt hat, die T3 hergibt.** Was
// `Freigabesicht` ausgibt, geht hier hinein; zwei Gestalten für dieselbe
// Sache liefen auseinander.

import (
	"database/sql"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// eineDublette legt drei Stück an, von denen zwei überzählig sind.
//
// **Über den Bestand und nicht über eine zweite Naht**: dieser Lauf misst den
// Abgleich, und eine Zeile, die über die Sammlungsnaht käme, brächte deren
// Wege mit in die Messung.
func eineDublette(t *testing.T, db *sql.DB, printing string) {
	t.Helper()

	schreibe(t, db, &domain.Exemplar{
		Printing: printing, Menge: 3, Zustand: domain.ZustandNearMint,
		Sprache: "de", Finish: domain.FinishNonfoil, Name: "Shivan Dragon",
	})
}

// einWunschMitRang legt einen Wunsch mit Rang und Art an.
func einWunschMitRang(t *testing.T, db *sql.DB, printing string) {
	t.Helper()

	schreibe(t, db, &domain.Exemplar{
		Printing: printing, Menge: 2, Zustand: domain.ZustandNearMint,
		Sprache: "de", Finish: domain.FinishNonfoil, Name: "Lightning Bolt",
		Wunsch: true, Wunschrang: domain.RangHoch, Wunschart: domain.WunschGenau,
	})
}

// schreibe legt eine Zeile in den Bestand.
func schreibe(t *testing.T, db *sql.DB, exemplar *domain.Exemplar) {
	t.Helper()

	if _, fehler := storage.NeuerSammlungsspeicher(db).
		Anlegen(probekontext(t), exemplar); fehler != nil {
		t.Fatalf("anlegen: %v", fehler)
	}
}

// eineFremdkarte baut eine Zeile, wie T3 sie hergibt.
func eineFremdkarte(printing string, menge int) app.SichtkarteDTO {
	return app.SichtkarteDTO{
		Printing: printing, Menge: menge, Zustand: "nearMint",
		Sprache: "de", Finish: "nonfoil", Bezeichnung: printing,
	}
}

func TestDerAbgleichNenntBeideRichtungen(t *testing.T) {
	db := frischerBestand(t)
	flaeche := app.PeerflaecheUeber(db)

	eineDublette(t, db, "shivan-lea")
	einWunschMitRang(t, db, "bolt-m10")

	antwort := flaeche.Tauschabgleich(probekontext(t), "bgegenueber", app.FremdstandDTO{
		Angebot:  []app.SichtkarteDTO{eineFremdkarte("bolt-m10", 4)},
		Wuensche: []app.SichtkarteDTO{eineFremdkarte("shivan-lea", 1)},
	}, 1, 0)
	if antwort.Fehler != nil {
		t.Fatalf("Tauschabgleich: %+v", antwort.Fehler)
	}

	if len(antwort.Gebe) != 1 || antwort.Gebe[0].Menge != 1 {
		t.Fatalf("Gebe: %+v", antwort.Gebe)
	}

	if len(antwort.Bekomme) != 1 || antwort.Bekomme[0].Menge != 2 {
		t.Fatalf("Bekomme: %+v", antwort.Bekomme)
	}

	// **Der Druck des Angebots reist mit** (P11, #2379): aus ihm entsteht
	// eine Handelszeile, wenn jemand den Vorschlag übernimmt.
	if antwort.Bekomme[0].AngebotDruck != "bolt-m10" || antwort.Gebe[0].AngebotDruck != "shivan-lea" {
		t.Errorf("die Drucke der Angebote: %q und %q", antwort.Bekomme[0].AngebotDruck, antwort.Gebe[0].AngebotDruck)
	}

	// **Und seine Variante** (Befund 2 des Zweitblicks über Phase 8): ohne
	// Zustand, Sprache und Ausführung entstand aus „Vorschlag übernehmen"
	// eine Zeile, die drüben nicht verbucht werden konnte.
	for _, kandidat := range []app.TauschkandidatDTO{antwort.Gebe[0], antwort.Bekomme[0]} {
		if kandidat.AngebotZustand != "nearMint" || kandidat.AngebotSprache != "de" ||
			kandidat.AngebotFinish != "nonfoil" {
			t.Errorf("die Variante des Angebots: %+v", kandidat)
		}
	}

	// **Die Erklärung reist mit** — Rang und Art sagen, warum diese Zeile
	// ein Treffer ist und wie dringend sie ist.
	if antwort.Bekomme[0].Rang != "hoch" || antwort.Bekomme[0].Art != "genau" {
		t.Errorf("die Erklärung ist %q und %q",
			antwort.Bekomme[0].Rang, antwort.Bekomme[0].Art)
	}

	if !antwort.HatVorschlag {
		t.Fatal("es kam kein Vorschlag")
	}

	if antwort.Vorschlag.Gebe.Stueck != 1 || antwort.Vorschlag.Bekomme.Stueck != 2 {
		t.Errorf("der Vorschlag ist %+v", antwort.Vorschlag)
	}
}

// **Ohne Gegenleistung gibt es keinen Vorschlag** — und die Listen kommen
// trotzdem als Listen an und nicht als `null` (§3).
func TestOhneGegenleistungKommtKeinVorschlag(t *testing.T) {
	db := frischerBestand(t)
	flaeche := app.PeerflaecheUeber(db)

	eineDublette(t, db, "shivan-lea")

	antwort := flaeche.Tauschabgleich(probekontext(t), "bgegenueber", app.FremdstandDTO{
		Wuensche: []app.SichtkarteDTO{eineFremdkarte("shivan-lea", 1)},
	}, 1, 0)

	if antwort.HatVorschlag {
		t.Error("es kam ein einseitiger Vorschlag")
	}

	if antwort.Bekomme == nil || antwort.Befunde == nil {
		t.Errorf("es kam null statt einer Liste: %+v", antwort)
	}
}

// **Ein Fehler kommt als Fehler an** — und nicht als leerer Abgleich.
func TestEinFehlerKommtAlsFehlerAn(t *testing.T) {
	flaeche := app.PeerflaecheUeber(frischerBestand(t))

	// **Ohne Profil gibt es keinen Bestand** — derselbe Weg, den jede
	// andere Naht nimmt.
	antwort := flaeche.Tauschabgleich(t.Context(), "bgegenueber",
		app.FremdstandDTO{}, 1, 0)

	if antwort.Fehler == nil {
		t.Fatal("es kam kein Fehler")
	}

	if antwort.Gebe == nil || antwort.Bekomme == nil || antwort.Befunde == nil {
		t.Errorf("es kam null statt einer Liste: %+v", antwort)
	}
}

// **Ein Befund kommt als Satz an** — nicht als Verbund und nicht gar nicht.
//
// **Eine fremde Zeile ohne Druck kann keinen genauen Wunsch bedienen.** Sie
// stillschweigend auszulassen verschwiege einen möglichen Tausch; sie zu raten
// böte einen falschen an (§8).
func TestEinBefundKommtAlsSatzAn(t *testing.T) {
	db := frischerBestand(t)
	flaeche := app.PeerflaecheUeber(db)

	einWunschMitRang(t, db, "bolt-m10")

	ohneDruck := eineFremdkarte("", 1)
	ohneDruck.Bezeichnung = "Eine Zeile ohne Druck"

	antwort := flaeche.Tauschabgleich(probekontext(t), "bgegenueber", app.FremdstandDTO{
		Angebot: []app.SichtkarteDTO{ohneDruck},
	}, 1, 0)
	if antwort.Fehler != nil {
		t.Fatalf("Tauschabgleich: %+v", antwort.Fehler)
	}

	if len(antwort.Befunde) != 1 {
		t.Fatalf("es kamen %d Befunde: %v", len(antwort.Befunde), antwort.Befunde)
	}

	if !strings.Contains(antwort.Befunde[0], "Eine Zeile ohne Druck") {
		t.Errorf("der Befund lautet %q", antwort.Befunde[0])
	}
}

// **Rang und Art gehen durch die Naht** — hin und zurück.
func TestRangUndArtGehenDurchDieNaht(t *testing.T) {
	naht := sammlungsnaht(t)
	ctx := probekontext(t)

	if _, fehler := naht.ExemplarAnlegen(ctx, app.ExemplarDTO{
		Printing: "bolt-m10", Menge: 2, Wunsch: true,
		Wunschrang: "hoch", Wunschart: "jederdruck",
	}); fehler != nil {
		t.Fatalf("ExemplarAnlegen: %v", fehler)
	}

	gelesen, fehler := naht.Sammlung(ctx)
	if fehler != nil {
		t.Fatalf("Sammlung: %v", fehler)
	}

	if len(gelesen) != 1 {
		t.Fatalf("es kamen %d Zeilen", len(gelesen))
	}

	if gelesen[0].Wunschrang != "hoch" || gelesen[0].Wunschart != "jederdruck" {
		t.Errorf("es steht %q und %q", gelesen[0].Wunschrang, gelesen[0].Wunschart)
	}
}
