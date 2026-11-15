package service_test

// **Warum nichts kam** — Werkstück A16 von Phase 5.3
// (#1515).
//
// **Der Satz, gegen den hier geprüft wird**: „Leere Treffer unterscheiden von
// Index fehlt, Modell fehlt, unzureichender Quelle, gesperrtem Zugriff und
// Budgetgrenze." Fünf davon sagen jemandem, was zu tun ist; eine sagt, dass
// es die Antwort nicht gibt.

import (
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// eineKartenfrage ist die einfachste Anfrage.
func eineKartenfrage() domain.Rechercheanfrage {
	return domain.Rechercheanfrage{
		Fall: domain.FallKarten, Frage: "spielsteine", Grenze: 10,
	}
}

// **Ohne Index sagt die Recherche es — und sucht nicht.**
func TestOhneIndexSagtDieRechercheEs(t *testing.T) {
	t.Parallel()

	quelle := &recherchequellendoppel{stand: service.Indexstand{
		Grund: "Es ist kein Suchindex aktiv.",
	}}

	ergebnis, fehler := einRecherchedienst(quelle, nil).Recherchieren(
		t.Context(), eineKartenfrage(), "")
	if fehler != nil {
		t.Fatalf("Recherchieren: %v", fehler)
	}

	if ergebnis.Leere != domain.LeereOhneIndex || ergebnis.Grund == "" {
		t.Fatalf("das Ergebnis ist %+v", ergebnis)
	}

	// **Und gesucht wurde gar nicht** — ein Ruf gegen einen Index, den es
	// nicht gibt, kostet Zeit für eine Antwort, die feststeht.
	if quelle.gesehen.Frage != "" {
		t.Errorf("gesucht wurde mit %q", quelle.gesehen.Frage)
	}
}

// **Ohne Modell steht es am Ergebnis und in den Einschränkungen.**
func TestOhneModellStehtEsAmErgebnis(t *testing.T) {
	t.Parallel()

	stand := einAktiverStand()
	stand.Suchbar = false
	stand.Grund = "Das Modell dieses Index ist nicht mehr installiert."

	ergebnis, fehler := einRecherchedienst(&recherchequellendoppel{
		stand: stand, modus: service.ModusLexikalisch,
		grund: stand.Grund,
	}, nil).Recherchieren(t.Context(), eineKartenfrage(), "")
	if fehler != nil {
		t.Fatalf("Recherchieren: %v", fehler)
	}

	if ergebnis.Leere != domain.LeereOhneModell {
		t.Fatalf("die Leere ist %q", ergebnis.Leere)
	}

	// **Der Volltextrückfall steht unter den Einschränkungen** — neben dem
	// fehlenden Eigenkartenkorpus, den der Prüfindex auch nicht führt.
	if !strings.Contains(strings.Join(ergebnis.Einschraenkungen, " "),
		"nur im Volltext") {
		t.Fatalf("die Einschränkungen sind %+v", ergebnis.Einschraenkungen)
	}
}

// **Ein fehlender Korpus steht als Einschränkung da.**
//
// **Und nicht als Fehler**: wer nach Rulings sucht und keine im Index hat,
// bekommt nichts — aber er soll wissen, dass es an der Quelle liegt und nicht
// an der Frage.
func TestEinFehlenderKorpusStehtAlsEinschraenkungDa(t *testing.T) {
	t.Parallel()

	ergebnis, fehler := einRecherchedienst(&recherchequellendoppel{
		stand: einAktiverStand(),
	}, nil).Recherchieren(t.Context(), domain.Rechercheanfrage{
		Fall: domain.FallRulings, Frage: "trampelschaden", Grenze: 10,
	}, "")
	if fehler != nil {
		t.Fatalf("Recherchieren: %v", fehler)
	}

	if len(ergebnis.Einschraenkungen) != 1 ||
		!strings.Contains(ergebnis.Einschraenkungen[0], "ruling") {
		t.Fatalf("die Einschränkungen sind %+v", ergebnis.Einschraenkungen)
	}

	// **Und der Korpus, den es gibt, steht nicht darunter.**
	if strings.Contains(strings.Join(ergebnis.Einschraenkungen, " "),
		"oracle") {
		t.Errorf("der vorhandene Korpus wurde als fehlend genannt: %+v",
			ergebnis.Einschraenkungen)
	}
}

// **Ist alles da und findet sich nichts, sagt sie genau das.**
func TestIstAllesDaUndFindetSichNichts(t *testing.T) {
	t.Parallel()

	ergebnis, fehler := einRecherchedienst(&recherchequellendoppel{
		stand: einAktiverStand(),
	}, nil).Recherchieren(t.Context(), eineKartenfrage(), "")
	if fehler != nil {
		t.Fatalf("Recherchieren: %v", fehler)
	}

	if ergebnis.Leere != domain.LeereNichtsGefunden {
		t.Fatalf("die Leere ist %q", ergebnis.Leere)
	}

	// **Und das Ergebnis nennt trotzdem, wogegen gesucht wurde.**
	if ergebnis.Modell == "" || ergebnis.Suchkonfiguration == "" {
		t.Errorf("das Ergebnis ist %+v", ergebnis)
	}
}

// **Ein verlangter Korpusstand, den es nicht gibt, ist ein Konflikt.**
//
// **Und kein leeres Ergebnis**: wer einen bestimmten Stand nennt, meint ihn —
// eine Antwort aus einem anderen wäre die Mischung, die A15 ausschließt.
func TestEinFremderKorpusstandIstEinKonflikt(t *testing.T) {
	t.Parallel()

	anfrage := eineKartenfrage()
	anfrage.Korpusstand = "gibt-es-nicht"

	if _, fehler := einRecherchedienst(&recherchequellendoppel{
		stand: einAktiverStand(),
	}, nil).Recherchieren(t.Context(), anfrage, ""); !errors.Is(
		fehler, domain.ErrConflict) {
		t.Fatalf("Recherchieren: %v", fehler)
	}

	// **Der Stand, den es gibt, geht durch.**
	anfrage.Korpusstand = einRetrievalmanifest().Korpusstaende[0].Revision

	if _, fehler := einRecherchedienst(&recherchequellendoppel{
		stand: einAktiverStand(),
	}, nil).Recherchieren(t.Context(), anfrage, ""); fehler != nil {
		t.Fatalf("der richtige Stand: %v", fehler)
	}
}

// **Ein Fehler der Suche kommt als Fehler zurück.**
func TestEinFehlerDerSucheKommtAlsFehlerZurueck(t *testing.T) {
	t.Parallel()

	if _, fehler := einRecherchedienst(&recherchequellendoppel{
		stand:  einAktiverStand(),
		fehler: domain.MitGrund(domain.ErrConflict, nil, "der Bestand ist zu"),
	}, nil).Recherchieren(t.Context(), eineKartenfrage(), ""); fehler == nil {
		t.Fatal("der Suchfehler kam als leeres Ergebnis zurück")
	}
}

// **Eine Anfrage, die keine ist, wird abgewiesen — vor jedem Ruf.**
func TestEineAnfrageDieKeineIstWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	quelle := &recherchequellendoppel{stand: einAktiverStand()}

	if _, fehler := einRecherchedienst(quelle, nil).Recherchieren(
		t.Context(), domain.Rechercheanfrage{Fall: domain.FallKarten},
		""); !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("Recherchieren: %v", fehler)
	}

	if quelle.gesehen.Frage != "" {
		t.Errorf("gesucht wurde trotzdem mit %q", quelle.gesehen.Frage)
	}
}

// **Ein unlesbares Profil hält die Recherche auf.**
func TestEinUnlesbaresProfilHaeltDieRechercheAuf(t *testing.T) {
	t.Parallel()

	if _, fehler := service.NeuerRecherchedienst(
		&recherchequellendoppel{stand: einAktiverStand()}, nil,
		profildoppel{fehler: domain.MitGrund(domain.ErrValidation, nil,
			"kein Profil im Kontext")},
	).Recherchieren(t.Context(), eineKartenfrage(), ""); fehler == nil {
		t.Fatal("die Recherche kam ohne Profil durch")
	}
}

// **Und ein Fehler beim Lesen der Ausgangskarte ebenso.**
func TestEinFehlerAmAusgangHaeltDieRechercheAuf(t *testing.T) {
	t.Parallel()

	if _, fehler := einRecherchedienst(&recherchequellendoppel{
		stand: einAktiverStand(),
	}, &ausgangsdoppel{
		fehler: domain.MitGrund(domain.ErrNotFound, nil,
			"die Karte gibt es nicht mehr"),
	}).Recherchieren(t.Context(), domain.Rechercheanfrage{
		Fall: domain.FallAehnlich, AusgangskarteID: 7, Ausgangsstand: "r",
		Grenze: 10,
	}, ""); fehler == nil {
		t.Fatal("die fehlende Ausgangskarte kam durch")
	}
}

// **Eine Sprache im Filter kommt am Index an.**
//
// **Und keine Sprache heißt: jede.** Ein leerer Filter mit einem leeren
// Eintrag darin schlösse alles aus.
func TestEineSpracheImFilterKommtAmIndexAn(t *testing.T) {
	t.Parallel()

	for name, fall := range map[string]struct {
		sprache  string
		erwartet int
	}{
		"eine Sprache": {"de", 1},
		"keine":        {"", 0},
		"nur Leerraum": {"  ", 0},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			quelle := &recherchequellendoppel{stand: einAktiverStand()}

			anfrage := eineKartenfrage()
			anfrage.Sprache = fall.sprache

			if _, fehler := einRecherchedienst(quelle, nil).Recherchieren(
				t.Context(), anfrage, ""); fehler != nil {
				t.Fatalf("Recherchieren: %v", fehler)
			}

			if len(quelle.gesehen.Filter.Sprachen) != fall.erwartet {
				t.Fatalf("gefiltert wurde nach %v",
					quelle.gesehen.Filter.Sprachen)
			}
		})
	}
}

// **Steht der Index, fehlt aber sein Modell, und kommen trotzdem Treffer,
// dann ist die Leere keine.**
//
// **Der Fall ist nicht erfunden**: die lexikalische Hälfte trägt weiter, und
// wer sie als „nichts gefunden" ausgäbe, verlöre die Treffer, die da sind.
func TestMitTreffernIstDieLeereKeine(t *testing.T) {
	t.Parallel()

	stand := einAktiverStand()
	stand.Suchbar = false

	ergebnis, fehler := einRecherchedienst(&recherchequellendoppel{
		stand: stand, modus: service.ModusLexikalisch,
		grund: "Kein Einbettungsmodell eingerichtet.",
		treffer: []domain.Retrievaltreffer{
			einIndextreffer("o1:0", "o1", "Ein Treffer", 1),
		},
	}, nil).Recherchieren(t.Context(), eineKartenfrage(), "")
	if fehler != nil {
		t.Fatalf("Recherchieren: %v", fehler)
	}

	if ergebnis.Leere != domain.LeereKeine || len(ergebnis.Treffer) != 1 {
		t.Fatalf("das Ergebnis ist %+v", ergebnis)
	}
}

// **Der Stand sagt suchbar, die Suche fiel trotzdem zurück.**
//
// # Warum es diesen Fall gibt
//
// **Weil es zwei Rufe sind.** `Stand` fragt, ob ein Modell da ist; `Suchen`
// fragt es wirklich. Zwischen beiden kann die Laufzeit ausfallen — und dann
// steht am Index „suchbar" und am Ergebnis „lexikalisch". Der Grund kommt
// dann aus der **Suche** und nicht aus dem Stand: er ist der jüngere.
func TestDerStandSagtSuchbarUndDieSucheFielZurueck(t *testing.T) {
	t.Parallel()

	ergebnis, fehler := einRecherchedienst(&recherchequellendoppel{
		stand: einAktiverStand(), modus: service.ModusLexikalisch,
		grund: "Das Einbettungsmodell hat nicht geantwortet.",
	}, nil).Recherchieren(t.Context(), eineKartenfrage(), "")
	if fehler != nil {
		t.Fatalf("Recherchieren: %v", fehler)
	}

	if ergebnis.Leere != domain.LeereOhneModell {
		t.Fatalf("die Leere ist %q", ergebnis.Leere)
	}

	if !strings.Contains(ergebnis.Grund, "nicht geantwortet") {
		t.Errorf("der Grund ist %q", ergebnis.Grund)
	}
}
