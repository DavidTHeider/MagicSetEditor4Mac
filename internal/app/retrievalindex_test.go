package app_test

// **Die Bindungen des Suchindex** — Werkstück A15 von Phase 5.3
// (#1514).
//
// **Was hier geprüft wird, ist die Übersetzung** — was `app` tut und sonst
// nichts: die Filterstruktur in die des Dienstes, das Manifest ins DTO, die
// Herkunft an den Treffer. Ob die Suche richtig sucht, steht in
// `internal/service`.

import (
	"context"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// retrievaldoppel hält, was die Bindungen bekommen sollen.
type retrievaldoppel struct {
	stand   service.Indexstand
	treffer service.Suchergebnis
	bericht service.Aufbaubericht

	suchfehler   error
	aufbaufehler error
	abgebrochen  bool

	// gesehen hält den Wunsch, mit dem gesucht wurde.
	gesehen service.Suchwunsch
	auftrag service.Aufbauauftrag
}

func (r *retrievaldoppel) Stand(context.Context) service.Indexstand {
	return r.stand
}

func (r *retrievaldoppel) Suchen(_ context.Context, wunsch service.Suchwunsch) (
	service.Suchergebnis, error,
) {
	r.gesehen = wunsch

	return r.treffer, r.suchfehler
}

func (r *retrievaldoppel) Aufbauen(
	_ context.Context, auftrag service.Aufbauauftrag,
) (service.Aufbaubericht, error) {
	r.auftrag = auftrag

	return r.bericht, r.aufbaufehler
}

func (r *retrievaldoppel) Abbrechen() bool { return r.abgebrochen }

// mitRetrieval baut eine Anwendung über einem Doppel.
func mitRetrieval(doppel *retrievaldoppel) *app.App {
	anwendung := app.NeueApp("")
	anwendung.MitRetrieval(doppel)

	return anwendung
}

// einProbemanifest ist ein vollständiges.
func einProbemanifest() domain.Indexmanifest {
	return domain.Indexmanifest{
		Schemafassung: 1, Zerteilfassung: "abschnitte-1", Indexfassung: 1,
		Modell: "all-minilm", Modellrevision: "aaaa",
		Laufzeit: "ollama 0.33.3", Abstandsmass: domain.AbstandCosinus,
		Vertrag: domain.KIEinbettungsvertrag{
			Dimension: 384, Pooling: domain.KIPoolingMittel,
			Normalisiert: true, Zerteiler: "bert",
		},
		Korpusstaende: []domain.Korpusstand{{
			Quelle: domain.QuelleOracle, Revision: "2026-09-01",
			Datum: "2026-09-01",
		}},
		Suchkonfiguration: service.Rangfusionsname,
	}
}

// **Der Stand kommt vollständig ans Fenster.**
func TestDerIndexstandKommtVollstaendigAnsFenster(t *testing.T) {
	t.Parallel()

	antwort := mitRetrieval(&retrievaldoppel{stand: service.Indexstand{
		Aktiv: true, Manifest: einProbemanifest(), Suchbar: true,
		Aufbau: service.Aufbaustand{Generation: 3, Geschrieben: 4711},
	}}).Indexstand(t.Context())

	if antwort.Fehler != nil {
		t.Fatalf("Indexstand: %+v", antwort.Fehler)
	}

	if !antwort.Aktiv || !antwort.Suchbar ||
		antwort.Manifest.Modell != "all-minilm" ||
		antwort.Manifest.Dimension != 384 ||
		antwort.Aufbau.Geschrieben != 4711 {
		t.Fatalf("die Antwort ist %+v", antwort)
	}

	// **Die Korpusstände kommen einzeln** — eine Zahl für alle wäre die
	// Mischung, die das Issue ausschließt.
	if len(antwort.Manifest.Korpusstaende) != 1 ||
		antwort.Manifest.Korpusstaende[0].Datum != "2026-09-01" {
		t.Errorf("die Stände sind %+v", antwort.Manifest.Korpusstaende)
	}
}

// **Die Filterstruktur kommt vollständig an — und die Farben bleiben zwei.**
func TestDieFilterstrukturKommtVollstaendigAn(t *testing.T) {
	t.Parallel()

	doppel := &retrievaldoppel{}
	von, bis := 1.0, 3.0

	mitRetrieval(doppel).Indexsuche(t.Context(), "bolt", 7,
		app.IndexsuchfilterDTO{
			Quellen: []string{"oracle"}, Sprachen: []string{"en"},
			Sets: []string{"neo"}, Farben: []string{"R"},
			Farbidentitaet: []string{"R", "G"}, Typen: []string{"instant"},
			ManawertVon: &von, ManawertBis: &bis,
			Format: "modern", Formatlage: "legal", Formatstand: "heute",
		})

	gesehen := doppel.gesehen

	if gesehen.Frage != "bolt" || gesehen.Grenze != 7 {
		t.Fatalf("der Wunsch ist %+v", gesehen)
	}

	if len(gesehen.Filter.Quellen) != 1 ||
		gesehen.Filter.Quellen[0] != domain.QuelleOracle {
		t.Errorf("die Quellen sind %+v", gesehen.Filter.Quellen)
	}

	// **Farbe und Farbidentität bleiben getrennt** — sie sind nicht dasselbe.
	if len(gesehen.Filter.Farben) != 1 ||
		len(gesehen.Filter.Farbidentitaet) != 2 {
		t.Errorf("die Farben sind %+v / %+v",
			gesehen.Filter.Farben, gesehen.Filter.Farbidentitaet)
	}

	if gesehen.Filter.ManawertVon == nil || *gesehen.Filter.ManawertVon != 1 ||
		gesehen.Filter.Formatstand != "heute" {
		t.Errorf("der Filter ist %+v", gesehen.Filter)
	}
}

// **Ein Treffer trägt seine Herkunft — mit der Abrufzeit als Zeitangabe.**
func TestEinTrefferTraegtSeineHerkunft(t *testing.T) {
	t.Parallel()

	abgerufen := time.Date(2026, 9, 11, 8, 0, 0, 0, time.UTC)

	antwort := mitRetrieval(&retrievaldoppel{treffer: service.Suchergebnis{
		Modus: service.ModusHybrid,
		Treffer: []domain.Retrievaltreffer{{
			Rang: 1, Ausschnitt: "…Lightning…",
			Dokument: domain.Retrievaldokument{
				Kennung: "scryfall-oracle:o1:0", Quelle: domain.QuelleOracle,
				Bezugsadresse: "https://beispiel.invalid/o1",
				Titel:         "Lightning Bolt", Text: "Deals 3 damage.",
				Sprache: "en", Quellrevision: "2026-09-01",
				Quelldatum: "2026-09-01", Abgerufen: abgerufen,
			},
		}},
	}}).Indexsuche(t.Context(), "bolt", 10, app.IndexsuchfilterDTO{})

	if len(antwort.Treffer) != 1 {
		t.Fatalf("die Treffer sind %+v", antwort.Treffer)
	}

	treffer := antwort.Treffer[0]

	if treffer.Kennung != "scryfall-oracle:o1:0" ||
		treffer.Ausschnitt != "…Lightning…" ||
		treffer.Quellrevision != "2026-09-01" ||
		treffer.Abgerufen != "2026-09-11T08:00:00Z" {
		t.Fatalf("der Treffer ist %+v", treffer)
	}

	if antwort.Modus != "hybrid" {
		t.Errorf("der Modus ist %q", antwort.Modus)
	}
}

// **Ein Aufbau reicht die vier Angaben durch — und die Abrufzeit ist jetzt.**
func TestEinAufbauReichtDieAngabenDurch(t *testing.T) {
	t.Parallel()

	doppel := &retrievaldoppel{bericht: service.Aufbaubericht{
		Generation: 3, Geschrieben: 4711, Dauer: 2 * time.Second,
		Manifest: einProbemanifest(),
	}}

	vorher := time.Now()

	antwort := mitRetrieval(doppel).IndexAufbauen(t.Context(),
		"all-minilm", "/tmp/oracle.jsonl.gz", "2026-09-01", "2026-09-01", true)

	if antwort.Fehler != nil {
		t.Fatalf("IndexAufbauen: %+v", antwort.Fehler)
	}

	if antwort.Generation != 3 || antwort.Geschrieben != 4711 ||
		antwort.Dauer != 2000 || !antwort.Fortgeschrieben {
		t.Fatalf("die Antwort ist %+v", antwort)
	}

	if doppel.auftrag.Modell != "all-minilm" ||
		doppel.auftrag.Quelle != "/tmp/oracle.jsonl.gz" ||
		doppel.auftrag.Revision != "2026-09-01" ||
		doppel.auftrag.Datum != "2026-09-01" ||
		!doppel.auftrag.Fortschreiben {
		t.Fatalf("der Auftrag ist %+v", doppel.auftrag)
	}

	if doppel.auftrag.Abgerufen.Before(vorher) {
		t.Errorf("die Abrufzeit ist %v", doppel.auftrag.Abgerufen)
	}
}

// **Ohne Naht antwortet jede Bindung mit einem Fehler statt mit einem Sturz.**
//
// **Der Fall ist nicht erfunden**: die drei Flächen ohne Fenster bauen ihre
// Anwendung je Ruf, und wer eine ohne Naht baut, bekommt genau das.
func TestOhneRetrievalnahtAntwortetJedeBindungMitFehler(t *testing.T) {
	t.Parallel()

	ohne := app.NeueApp("")

	if antwort := ohne.Indexstand(t.Context()); antwort.Fehler == nil {
		t.Error("Indexstand kam ohne Naht durch")
	}

	if antwort := ohne.Indexsuche(t.Context(), "x", 10,
		app.IndexsuchfilterDTO{}); antwort.Fehler == nil {
		t.Error("Indexsuche kam ohne Naht durch")
	}

	if antwort := ohne.IndexAufbauen(t.Context(), "m", "q", "r", "",
		false); antwort.Fehler == nil {
		t.Error("IndexAufbauen kam ohne Naht durch")
	}

	if ohne.IndexAbbrechen() {
		t.Error("IndexAbbrechen brach ohne Naht etwas ab")
	}
}

// **Ein Fehler der Suche und des Aufbaus kommt als Fehler zurück.**
func TestEinFehlerDerSucheKommtAlsFehlerZurueck(t *testing.T) {
	t.Parallel()

	doppel := &retrievaldoppel{
		suchfehler:   domain.MitGrund(domain.ErrConflict, nil, "der Bestand ist zu"),
		aufbaufehler: domain.MitGrund(domain.ErrValidation, nil, "kein Modell"),
		abgebrochen:  true,
	}

	anwendung := mitRetrieval(doppel)

	if antwort := anwendung.Indexsuche(t.Context(), "x", 10,
		app.IndexsuchfilterDTO{}); antwort.Fehler == nil {
		t.Error("der Suchfehler kam als Ergebnis zurück")
	}

	if antwort := anwendung.IndexAufbauen(t.Context(), "m", "q", "r", "",
		false); antwort.Fehler == nil {
		t.Error("der Aufbaufehler kam als Bericht zurück")
	}

	if !anwendung.IndexAbbrechen() {
		t.Error("der Abbruch kam nicht durch")
	}
}
