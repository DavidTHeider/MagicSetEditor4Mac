package service_test

// **Eine Leihe einlösen, widerrufen und auslaufen lassen** — Werkstück G6 von
// Phase 5.5 (#1463).
//
// **Die Gegenrichtung zu `deckleihe_test.go`**: dort wird eine Leihe
// **ausgestellt** und geprüft, hier wird sie **gebraucht**. Beides in einer
// Datei stand bei 541 Zeilen (§2).

import (
	"errors"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Vier Abweisungen vor dem Start** — die erste Abnahme des Issues.
func TestDieLeiheWirdVorDemStartGeprueft(t *testing.T) {
	t.Parallel()

	faelle := map[string]struct {
		jetzt time.Time
		woran string
		recht domain.Leihrecht
		gilt  bool
	}{
		"die freigegebene Partie": {
			jetzt: leihzeit, woran: "p1",
			recht: domain.RechtSpielen, gilt: true,
		},
		"eine andere Partie": {
			jetzt: leihzeit, woran: "p2", recht: domain.RechtSpielen,
		},
		"nach dem Ablauf": {
			jetzt: leihzeit.Add(48 * time.Hour), woran: "p1",
			recht: domain.RechtSpielen,
		},
		"ein Recht, das nicht darinsteht": {
			jetzt: leihzeit, woran: "p1", recht: domain.RechtEinreichen,
		},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			stelle, leihe := eineAngenommeneLeihe(t, fall.jetzt)

			abzug, fehler := stelle.Vorlegen(t.Context(), leihe.Kennung,
				domain.Leihbindung{
					Art: domain.LeiheFuerPartie, Woran: fall.woran,
				}, fall.recht)

			switch {
			case fall.gilt && fehler != nil:
				t.Fatalf("erwartet gültig, bekam %v", fehler)
			case fall.gilt && abzug.Abdruck != leihe.Abdruck:
				t.Fatalf("der Abzug ist %q", abzug.Abdruck)

			// **Seine Zeilen sind die des Verleihers** (P9, #2377): hier
			// nachgeschlagen fänden sie andere Karten.
			case fall.gilt && abzug.Quelle != leihe.Verleiher:
				t.Fatalf("die Quelle ist %q, verliehen hat %q", abzug.Quelle, leihe.Verleiher)
			case !fall.gilt && fehler == nil:
				t.Fatal("erwartet eine Abweisung")
			}
		})
	}
}

// **Ein anderer Peer bekommt den Abzug nicht** — auch nicht mit einem echten
// Dokument.
func TestEinFremderPeerLegtDieLeiheNichtVor(t *testing.T) {
	t.Parallel()

	anna, bert := zweiPeers(t)
	beim, _, _ := eineLeihstelle(t, anna)

	leihe, abzug, fehler := beim.Anbieten(t.Context(), einLeihwunsch(bert))
	if fehler != nil {
		t.Fatalf("anbieten: %v", fehler)
	}

	// **Cem hat das Dokument — und es ist echt.**
	_, cem := zweiPeers(t)
	beiCem, _, _ := eineLeihstelle(t, cem)

	if fehler := beiCem.Annehmen(t.Context(), leihe, abzug); !errors.Is(
		fehler, domain.ErrValidation) {
		t.Fatalf("erwartet ErrValidation, bekam %v", fehler)
	}
}

// **Ein fremder Abzug zu einem echten Dokument wird abgewiesen.**
//
// **Die Unterschrift steht über dem Abdruck und nicht über den Bytes** — ohne
// diese Prüfung käme ein echtes Dokument mit fremden Karten durch.
func TestEinFremderAbzugWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	anna, bert := zweiPeers(t)
	beimVerleiher, _, _ := eineLeihstelle(t, anna)

	leihe, _, fehler := beimVerleiher.Anbieten(t.Context(), einLeihwunsch(bert))
	if fehler != nil {
		t.Fatalf("anbieten: %v", fehler)
	}

	beimEntleiher, _, _ := eineLeihstelle(t, bert)

	fehler = beimEntleiher.Annehmen(t.Context(), leihe, domain.Partiedeck{
		Abdruck: "einfremderabdruck",
		Karten: []domain.Abzugskarte{{
			Kennung: "k1", Name: "Insel", Menge: 40,
		}},
	})

	if !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("erwartet ErrValidation, bekam %v", fehler)
	}
}

// **Eine offene Bindung wird beim ersten Start festgenagelt.**
func TestEineOffeneLeiheGiltNurEinerPartie(t *testing.T) {
	t.Parallel()

	stelle, leihe := eineAngenommeneLeihe(t, leihzeit, func(w *service.Leihwunsch) {
		w.Bindung.Woran = ""
	})

	if _, fehler := stelle.Vorlegen(t.Context(), leihe.Kennung,
		domain.Leihbindung{Art: domain.LeiheFuerPartie, Woran: "p1"},
		domain.RechtSpielen); fehler != nil {
		t.Fatalf("der erste Start: %v", fehler)
	}

	if _, fehler := stelle.Vorlegen(t.Context(), leihe.Kennung,
		domain.Leihbindung{Art: domain.LeiheFuerPartie, Woran: "p9"},
		domain.RechtSpielen); !errors.Is(fehler, domain.ErrConflict) {
		t.Fatalf("der zweite Start: erwartet ErrConflict, bekam %v", fehler)
	}

	// **Dieselbe Partie noch einmal geht** — ein Neustart nach einem Absturz
	// ist kein zweiter Start.
	if _, fehler := stelle.Vorlegen(t.Context(), leihe.Kennung,
		domain.Leihbindung{Art: domain.LeiheFuerPartie, Woran: "p1"},
		domain.RechtSpielen); fehler != nil {
		t.Fatalf("derselbe Start noch einmal: %v", fehler)
	}
}

// **Ein Widerruf verhindert neue Starts.**
func TestEinWiderrufVerhindertNeueStarts(t *testing.T) {
	t.Parallel()

	anna, bert := zweiPeers(t)
	beimVerleiher, verleiherablage, _ := eineLeihstelle(t, anna)

	leihe, abzug, anbietfehler := beimVerleiher.Anbieten(
		t.Context(), einLeihwunsch(bert))
	if anbietfehler != nil {
		t.Fatalf("anbieten: %v", anbietfehler)
	}

	if fehler := beimVerleiher.Widerrufen(
		t.Context(), leihe.Kennung); fehler != nil {
		t.Fatalf("widerrufen: %v", fehler)
	}

	stand, lesefehler := verleiherablage.Leihe(t.Context(), leihe.Kennung)
	if lesefehler != nil {
		t.Fatalf("lesen: %v", lesefehler)
	}

	if !stand.Widerrufen {
		t.Fatal("der Widerruf steht nicht in der Ablage")
	}

	// **Und beim Entleiher, sobald er ihn erfährt.**
	beimEntleiher, entleiherablage, _ := eineLeihstelle(t, bert)

	if fehler := beimEntleiher.Annehmen(t.Context(), leihe, abzug); fehler != nil {
		t.Fatalf("annehmen: %v", fehler)
	}

	if fehler := beimEntleiher.Zurueckgeben(
		t.Context(), leihe.Kennung); fehler != nil {
		t.Fatalf("zurückgeben: %v", fehler)
	}

	genommen, zweiterlesefehler := entleiherablage.Leihe(
		t.Context(), leihe.Kennung)
	if zweiterlesefehler != nil {
		t.Fatalf("lesen: %v", zweiterlesefehler)
	}

	if !genommen.Widerrufen {
		t.Fatal("die zurückgegebene Leihe steht nicht als widerrufen")
	}

	if _, fehler := beimEntleiher.Vorlegen(t.Context(), leihe.Kennung,
		domain.Leihbindung{Art: domain.LeiheFuerPartie, Woran: "p1"},
		domain.RechtSpielen); !errors.Is(fehler, domain.ErrConflict) {
		t.Fatalf("erwartet ErrConflict, bekam %v", fehler)
	}
}

// **Eine laufende Partie folgt dem vereinbarten Modus** — und niemals einem
// heimlichen Wegnehmen.
func TestEineLaufendeNutzungFolgtDemModus(t *testing.T) {
	t.Parallel()

	faelle := map[string]struct {
		modus  domain.Leihmodus
		bleibt bool
	}{
		"zu Ende spielen": {modus: domain.LaufZuEnde, bleibt: true},
		"abbrechen":       {modus: domain.LaufAbbruch},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			stelle, leihe := eineAngenommeneLeihe(t, leihzeit,
				func(w *service.Leihwunsch) { w.Modus = fall.modus })

			if _, fehler := stelle.Vorlegen(t.Context(), leihe.Kennung,
				domain.Leihbindung{Art: domain.LeiheFuerPartie, Woran: "p1"},
				domain.RechtSpielen); fehler != nil {
				t.Fatalf("vorlegen: %v", fehler)
			}

			if fehler := stelle.Zurueckgeben(
				t.Context(), leihe.Kennung); fehler != nil {
				t.Fatalf("zurückgeben: %v", fehler)
			}

			bleibt, grund, fehler := stelle.Fortdauer(
				t.Context(), leihe.Kennung, "p1")
			if fehler != nil {
				t.Fatalf("fortdauer: %v", fehler)
			}

			if bleibt != fall.bleibt {
				t.Fatalf("bleibt ist %v, Grund %q", bleibt, grund)
			}

			if grund == "" {
				t.Error("es steht kein Grund dabei")
			}
		})
	}
}

// eineAngenommeneLeihe gibt eine Stelle beim Entleiher samt angenommener
// Leihe — die Uhr steht dabei auf `jetzt`.
func eineAngenommeneLeihe(
	t *testing.T, jetzt time.Time, aendere ...func(*service.Leihwunsch),
) (*service.Leihstelle, domain.Deckleihe) {
	t.Helper()

	anna, bert := zweiPeers(t)
	beimVerleiher, _, _ := eineLeihstelle(t, anna)

	wunsch := einLeihwunsch(bert)
	for _, aenderung := range aendere {
		aenderung(&wunsch)
	}

	leihe, abzug, fehler := beimVerleiher.Anbieten(t.Context(), wunsch)
	if fehler != nil {
		t.Fatalf("anbieten: %v", fehler)
	}

	beimEntleiher := service.NeueLeihstelle(neueLeihprobe(),
		&deckprobe{decks: map[int64]domain.Deck{}}, ausweisprobe{ich: bert},
		func() time.Time { return jetzt })

	if fehler := beimEntleiher.Annehmen(t.Context(), leihe, abzug); fehler != nil {
		t.Fatalf("annehmen: %v", fehler)
	}

	return beimEntleiher, leihe
}

// **Eine Partie startet mit einem geliehenen Abzug** — und ohne Leihe nicht.
//
// **Das ist die Stelle, an der eine Leihe etwas tut.** Ein Dokument, das
// niemand vorlegt, ist eine Zusage ohne Einlösung (§2, §10).
func TestEinePartieStartetMitEinemGeliehenenAbzug(t *testing.T) {
	t.Parallel()

	stelle, leihe := eineAngenommeneLeihe(t, leihzeit)

	aufbau := service.NeuerPartieaufbau(
		&deckprobe{decks: map[int64]domain.Deck{}}).MitLeihen(stelle)

	partie, fehler := aufbau.Aufbauen(t.Context(), "p1", "magic",
		domain.Regelwerk{Strenge: domain.StrengeSandkasten}, "anna",
		[]service.Sitzwunsch{{
			Kennung: "anna", Name: "Anna", LeiheID: leihe.Kennung,
		}})
	if fehler != nil {
		t.Fatalf("aufbauen: %v", fehler)
	}

	if partie.Sitze[0].Deck.Abdruck != leihe.Abdruck {
		t.Fatalf("der Abzug ist %q", partie.Sitze[0].Deck.Abdruck)
	}
}

// **Eine Leihe für eine andere Partie trägt diese nicht.**
func TestEineLeiheFuerEineAnderePartieTraegtNicht(t *testing.T) {
	t.Parallel()

	stelle, leihe := eineAngenommeneLeihe(t, leihzeit)

	aufbau := service.NeuerPartieaufbau(
		&deckprobe{decks: map[int64]domain.Deck{}}).MitLeihen(stelle)

	_, fehler := aufbau.Aufbauen(t.Context(), "p9", "magic",
		domain.Regelwerk{Strenge: domain.StrengeSandkasten}, "anna",
		[]service.Sitzwunsch{{
			Kennung: "anna", Name: "Anna", LeiheID: leihe.Kennung,
		}})

	if fehler == nil {
		t.Fatal("die Leihe trug eine fremde Partie")
	}
}

// **Ohne eingehängte Leihstelle wird eine Leihkennung gemeldet** — statt
// still auf ein Deck zurückzufallen, das es nicht gibt.
func TestOhneLeihstelleWirdEineLeihkennungGemeldet(t *testing.T) {
	t.Parallel()

	_, fehler := service.NeuerPartieaufbau(
		&deckprobe{decks: map[int64]domain.Deck{}}).Aufbauen(
		t.Context(), "p1", "magic",
		domain.Regelwerk{Strenge: domain.StrengeSandkasten}, "anna",
		[]service.Sitzwunsch{{
			Kennung: "anna", Name: "Anna", LeiheID: "l1",
		}})

	if !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("erwartet ErrValidation, bekam %v", fehler)
	}
}
