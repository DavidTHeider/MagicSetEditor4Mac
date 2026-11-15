package service_test

// **Die gemeinsame Recherche** — Werkstück A16 von Phase 5.3
// (#1515).
//
// # Wogegen hier geprüft wird
//
// **Dass eine Recherche kein Sprachmodell fragt.** Das Issue nennt es im Ziel
// und in der Abnahme: die Kartenfrage liefert Quellenkarten „ohne
// Textgenerierungsaufruf". Hier gibt es keinen Anbieter und keinen Auftrag —
// und genau das ist die Zusage.
//
// **Und dass „nichts gefunden" seinen Grund nennt.**

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// recherchequellendoppel gibt Stand und Treffer vor.
type recherchequellendoppel struct {
	stand   service.Indexstand
	treffer []domain.Retrievaltreffer
	modus   service.Suchmodus
	grund   string
	fehler  error

	// gesehen hält den Wunsch, mit dem gesucht wurde.
	gesehen service.Suchwunsch
}

func (r *recherchequellendoppel) Stand(context.Context) service.Indexstand {
	return r.stand
}

func (r *recherchequellendoppel) Suchen(
	_ context.Context, wunsch service.Suchwunsch,
) (service.Suchergebnis, error) {
	r.gesehen = wunsch

	if r.fehler != nil {
		return service.Suchergebnis{}, r.fehler
	}

	modus := r.modus
	if modus == "" {
		modus = service.ModusHybrid
	}

	return service.Suchergebnis{
		Treffer: r.treffer, Modus: modus, Grund: r.grund,
	}, nil
}

// ausgangsdoppel gibt den Text einer Karte.
type ausgangsdoppel struct {
	text      string
	kennungen []string
	fehler    error

	// jetziger ist der Stand, auf dem die Karte **jetzt** steht — das, was
	// eine echte Quelle aus dem Bestand rechnet.
	jetziger string

	// gesehen hält Kennung und Stand, mit denen gefragt wurde.
	karteID int64
	stand   string
}

func (a *ausgangsdoppel) Ausgangstext(
	_ context.Context, karteID int64, stand string,
) (text, jetziger string, kennungen []string, fehler error) {
	a.karteID, a.stand = karteID, stand

	return a.text, a.jetziger, a.kennungen, a.fehler
}

// profildoppel nennt ein festes Profil.
type profildoppel struct {
	profil int64
	fehler error

	// keines lässt es null melden **ohne** Fehler — die Gestalt eines
	// Adapters, der nichts findet und es nicht sagt.
	keines bool
}

func (p profildoppel) Profil(context.Context) (int64, error) {
	if p.fehler != nil {
		return 0, p.fehler
	}

	if p.keines {
		return 0, nil
	}

	if p.profil == 0 {
		return 1, nil
	}

	return p.profil, nil
}

// dieRechercheuhr steht fest.
func dieRechercheuhr() time.Time {
	return time.Date(2026, 9, 12, 12, 0, 0, 0, time.UTC)
}

// einRecherchedienst baut ihn über den Doppeln.
func einRecherchedienst(
	quelle *recherchequellendoppel, ausgang *ausgangsdoppel,
) *service.Recherchedienst {
	return service.NeuerRecherchedienst(quelle, ausgang, profildoppel{}).
		MitUhr(dieRechercheuhr)
}

// einAktiverStand ist ein Index, gegen den sich suchen lässt.
func einAktiverStand() service.Indexstand {
	return service.Indexstand{
		Aktiv: true, Suchbar: true, Manifest: einRetrievalmanifest(),
	}
}

// einIndextreffer baut einen Treffer des Index.
func einIndextreffer(kennung, fachkennung, titel string, rang int) domain.Retrievaltreffer {
	return domain.Retrievaltreffer{
		Rang: rang, Ausschnitt: "…" + titel + "…",
		Dokument: domain.Retrievaldokument{
			Kennung: kennung, Quelle: domain.QuelleOracle,
			Kennungssystem: domain.SystemScryfall, Fachkennung: fachkennung,
			Semantik: domain.SemantikKonzept, Titel: titel,
			Text: titel + " macht etwas.", Sprache: "en",
			Quellrevision: "2026-09-01",
			Abgerufen:     dieRechercheuhr(),
		},
	}
}

// **Jeder Fall sucht in seinen Korpora — und fragt kein Modell.**
//
// **Mit Freigabe für die privaten** (A13.4, #1452): seit dem Werkstück bleibt
// eine private Quelle ohne ausdrückliche Auswahl außen, und ohne diese Zeile
// prüfte der Lauf die Weglassung statt der Aufzählung. Dass die Weglassung
// wirkt, steht in `kontextfreigabewege_test.go`.
func TestJederRecherchefallSuchtInSeinenKorpora(t *testing.T) {
	t.Parallel()

	for _, fall := range domain.Recherchefaelle() {
		if fall == domain.FallAehnlich {
			continue
		}

		t.Run(string(fall), func(t *testing.T) {
			t.Parallel()

			quelle := &recherchequellendoppel{stand: einAktiverStand()}

			anfrage := domain.Rechercheanfrage{
				Fall: fall, Frage: "spielsteine", Grenze: 10,
				Freigegeben: privateQuellenVon(fall),
			}

			if _, fehler := einRecherchedienst(quelle, nil).Recherchieren(
				t.Context(), anfrage, ""); fehler != nil {
				t.Fatalf("Recherchieren: %v", fehler)
			}

			erwartet := fall.Quellen()
			if len(quelle.gesehen.Filter.Quellen) != len(erwartet) {
				t.Fatalf("gesucht wurde in %v", quelle.gesehen.Filter.Quellen)
			}

			// **Gefragt wird weiter als geantwortet** — gebündelt wird
			// danach.
			if quelle.gesehen.Grenze <= 10 {
				t.Errorf("die Grenze ist %d", quelle.gesehen.Grenze)
			}
		})
	}
}

// **Eine Ähnlichkeitsfrage sucht mit dem Text ihrer Ausgangskarte.**
//
// **Und die Karte kommt in ihren eigenen Treffern nicht vor.**
func TestEineAehnlichkeitsfrageSuchtMitIhremAusgangstext(t *testing.T) {
	t.Parallel()

	quelle := &recherchequellendoppel{
		stand: einAktiverStand(),
		treffer: []domain.Retrievaltreffer{
			einIndextreffer("o1:0", "o1", "Die Ausgangskarte", 1),
			einIndextreffer("o2:0", "o2", "Etwas Ähnliches", 2),
		},
	}

	ausgang := &ausgangsdoppel{
		text:      "Wenn du einen Spontanzauber wirkst, erzeuge einen Spielstein.",
		kennungen: []string{"o1:0"},
	}

	ergebnis, fehler := einRecherchedienst(quelle, ausgang).Recherchieren(
		t.Context(), domain.Rechercheanfrage{
			Fall: domain.FallAehnlich, AusgangskarteID: 7,
			Ausgangsstand: "rev-1", Grenze: 10,
		}, "")
	if fehler != nil {
		t.Fatalf("Recherchieren: %v", fehler)
	}

	// **Die Ausgangskarte wird mit ihrem Stand gelesen.**
	if ausgang.karteID != 7 || ausgang.stand != "rev-1" {
		t.Errorf("gelesen wurde %d in %q", ausgang.karteID, ausgang.stand)
	}

	if quelle.gesehen.Frage != ausgang.text {
		t.Errorf("gesucht wurde mit %q", quelle.gesehen.Frage)
	}

	if len(ergebnis.Treffer) != 1 ||
		ergebnis.Treffer[0].Kennung != "o2:0" {
		t.Fatalf("die Treffer sind %+v", ergebnis.Treffer)
	}
}

// **Ohne Ausgangsquelle gibt es keine Ähnlichkeitsfrage.**
func TestOhneAusgangsquelleGibtEsKeineAehnlichkeitsfrage(t *testing.T) {
	t.Parallel()

	if _, fehler := service.NeuerRecherchedienst(
		&recherchequellendoppel{stand: einAktiverStand()}, nil,
		profildoppel{}).Recherchieren(t.Context(), domain.Rechercheanfrage{
		Fall: domain.FallAehnlich, AusgangskarteID: 7, Ausgangsstand: "r",
		Grenze: 10,
	}, ""); !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("Recherchieren: %v", fehler)
	}
}

// **Zwei Abschnitte derselben Karte werden ein Treffer.**
func TestZweiAbschnitteDerselbenKarteWerdenEinTreffer(t *testing.T) {
	t.Parallel()

	quelle := &recherchequellendoppel{
		stand: einAktiverStand(),
		treffer: []domain.Retrievaltreffer{
			einIndextreffer("o1:0", "o1", "Vorderseite", 1),
			einIndextreffer("o1:1", "o1", "Rückseite", 2),
			einIndextreffer("o2:0", "o2", "Etwas anderes", 3),
		},
	}

	ergebnis, fehler := einRecherchedienst(quelle, nil).Recherchieren(
		t.Context(), domain.Rechercheanfrage{
			Fall: domain.FallKarten, Frage: "etwas", Grenze: 10,
		}, "")
	if fehler != nil {
		t.Fatalf("Recherchieren: %v", fehler)
	}

	if len(ergebnis.Treffer) != 2 ||
		len(ergebnis.Treffer[0].Weitere) != 1 {
		t.Fatalf("die Treffer sind %+v", ergebnis.Treffer)
	}
}

// **Ein leerer Ausgangsstand wird zu dem, auf dem die Karte jetzt steht.**
//
// # Warum das ein eigener Lauf ist
//
// **Weil hier bis zum 2026-09-12 eine Pflicht stand, die keine Fläche
// einlösen konnte**: den gerechneten Stand einer Karte rechnet der Bestand,
// und weder Fenster noch Kommandozeile noch MCP bekommen ihn zu sehen. Die
// Ähnlichkeitsfrage wäre an allen vier Flächen abgewiesen worden.
//
// **Die Antwort nennt ihn jetzt**, und das ist keine Bequemlichkeit: ohne ihn
// bände der Fortsetzungszeiger den leeren Stand, und die zweite Seite gehörte
// zu einer Karte, die es so nicht mehr gibt.
func TestEinLeererAusgangsstandWirdZumJetzigen(t *testing.T) {
	t.Parallel()

	ausgang := &ausgangsdoppel{
		text: "Blitzschlag fügt 3 Schaden zu.", jetziger: "stand-von-jetzt",
		kennungen: []string{"eigenekarte:7:0"},
	}

	ergebnis, fehler := einRecherchedienst(&recherchequellendoppel{
		stand: einAktiverStand(),
	}, ausgang).Recherchieren(t.Context(), domain.Rechercheanfrage{
		Fall: domain.FallAehnlich, AusgangskarteID: 7, Grenze: 10,
	}, "")
	if fehler != nil {
		t.Fatalf("Recherchieren: %v", fehler)
	}

	if ergebnis.Ausgangsstand != "stand-von-jetzt" {
		t.Fatalf("die Antwort nennt %q", ergebnis.Ausgangsstand)
	}
}

// **Und der Zeiger bindet ihn — nicht den leeren.**
//
// **Die Gegenprobe zur Zeile darüber** (§10): dass eine Zeichenkette
// zurückkommt, sagt noch nicht, dass sie irgendwo wirkt. Hier wirkt sie —
// derselbe Ruf gegen eine Karte auf einem anderen Stand weist den Zeiger ab.
func TestDerZeigerBindetDenGerechnetenAusgangsstand(t *testing.T) {
	t.Parallel()

	quelle := &recherchequellendoppel{
		stand:   einAktiverStand(),
		treffer: vieleTreffer(30),
	}

	erste, fehler := einRecherchedienst(quelle, &ausgangsdoppel{
		text: "Blitzschlag", jetziger: "stand-eins",
	}).Recherchieren(t.Context(), domain.Rechercheanfrage{
		Fall: domain.FallAehnlich, AusgangskarteID: 7, Grenze: 5,
	}, "")
	if fehler != nil || erste.Fortsetzung == "" {
		t.Fatalf("die erste Seite ist %+v (%v)", erste, fehler)
	}

	// **Dieselbe Karte, ein anderer Stand** — der Zeiger gehört zu einer
	// Liste, die es so nicht mehr gibt.
	_, zweiter := einRecherchedienst(quelle, &ausgangsdoppel{
		text: "Blitzschlag, geändert", jetziger: "stand-zwei",
	}).Recherchieren(t.Context(), domain.Rechercheanfrage{
		Fall: domain.FallAehnlich, AusgangskarteID: 7, Grenze: 5,
	}, erste.Fortsetzung)

	if !errors.Is(zweiter, domain.ErrConflict) {
		t.Fatalf("der Zeiger überlebte den Standwechsel: %v", zweiter)
	}
}

// privateQuellenVon gibt die privaten Korpora eines Falls — A13.4 (#1452).
func privateQuellenVon(fall domain.Recherchefall) []domain.Retrievalquelle {
	privat := make([]domain.Retrievalquelle, 0, 2)

	for _, quelle := range fall.Quellen() {
		if quelle.IstPrivat() {
			privat = append(privat, quelle)
		}
	}

	return privat
}
