package service_test

// **Die Fehlerwege der Leihstelle** — Werkstück G6 von Phase 5.5
// (#1463).
//
// **Jeder betreten** (§3): ein Weg, den kein Lauf geht, ist eine Zusage ohne
// Prüfung — und zum Ergebnis gehört das Nein (§10).

import (
	"errors"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/peer"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// errProbe ist der Grund, den die Proben melden.
var errProbe = errors.New("die Probe klagt")

// eineKlagendeStelle gibt eine Leihstelle, deren Ablage auf jedem Weg klagt.
func eineKlagendeStelle(
	t *testing.T, ich peer.Identitaet,
) *service.Leihstelle {
	t.Helper()

	ablage := neueLeihprobe()
	ablage.klagt = errProbe

	return service.NeueLeihstelle(ablage, &deckprobe{
		decks: map[int64]domain.Deck{7: {ID: 7, Name: "W", Spiel: "magic"}},
	}, ausweisprobe{ich: ich}, func() time.Time { return leihzeit })
}

// **Ohne Ausweis geht nichts** — weder anbieten noch annehmen noch vorlegen.
func TestOhneAusweisGehtKeineLeihe(t *testing.T) {
	t.Parallel()

	stelle := service.NeueLeihstelle(neueLeihprobe(),
		&deckprobe{decks: map[int64]domain.Deck{}},
		ausweisprobe{fehler: errProbe}, func() time.Time { return leihzeit })

	_, bert := zweiPeers(t)

	wege := map[string]func() error{
		"anbieten": func() error {
			_, _, fehler := stelle.Anbieten(t.Context(), einLeihwunsch(bert))

			return fehler
		},
		"annehmen": func() error {
			return stelle.Annehmen(
				t.Context(), domain.Deckleihe{}, domain.Partiedeck{})
		},
	}

	for name, weg := range wege {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if fehler := weg(); !errors.Is(fehler, errProbe) {
				t.Fatalf("erwartet die Klage, bekam %v", fehler)
			}
		})
	}
}

// **Ein Deck, das es nicht gibt, wird gemeldet.**
func TestEinFehlendesDeckWirdBeimAnbietenGemeldet(t *testing.T) {
	t.Parallel()

	anna, bert := zweiPeers(t)

	stelle := service.NeueLeihstelle(neueLeihprobe(),
		&deckprobe{decks: map[int64]domain.Deck{}}, ausweisprobe{ich: anna},
		func() time.Time { return leihzeit })

	if _, _, fehler := stelle.Anbieten(
		t.Context(), einLeihwunsch(bert)); fehler == nil {
		t.Fatal("das fehlende Deck kam nicht an")
	}
}

// **Ein Wunsch, aus dem keine gültige Leihe wird, wird abgewiesen** — bevor
// irgendetwas unterschrieben ist.
func TestEinUngueltigerWunschWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	anna, bert := zweiPeers(t)
	stelle, _, _ := eineLeihstelle(t, anna)

	wunsch := einLeihwunsch(bert)
	wunsch.Dauer = 0

	if _, _, fehler := stelle.Anbieten(t.Context(), wunsch); !errors.Is(
		fehler, domain.ErrValidation) {
		t.Fatalf("erwartet ErrValidation, bekam %v", fehler)
	}
}

// **Eine klagende Ablage wird auf jedem Weg gemeldet.**
func TestEineKlagendeAblageWirdGemeldet(t *testing.T) {
	t.Parallel()

	anna, bert := zweiPeers(t)
	klagend := eineKlagendeStelle(t, anna)

	gute, _, _ := eineLeihstelle(t, anna)

	leihe, abzug, fehler := gute.Anbieten(t.Context(), einLeihwunsch(bert))
	if fehler != nil {
		t.Fatalf("anbieten: %v", fehler)
	}

	beiBert := eineKlagendeStelle(t, bert)

	wege := map[string]func() error{
		"anbieten": func() error {
			_, _, fehler := klagend.Anbieten(t.Context(), einLeihwunsch(bert))

			return fehler
		},
		"annehmen": func() error {
			return beiBert.Annehmen(t.Context(), leihe, abzug)
		},
		"widerrufen": func() error {
			return klagend.Widerrufen(t.Context(), leihe.Kennung)
		},
		"zurückgeben": func() error {
			return klagend.Zurueckgeben(t.Context(), leihe.Kennung)
		},
		"listen": func() error {
			_, fehler := klagend.Leihen(t.Context())

			return fehler
		},
		"vorlegen": func() error {
			_, fehler := klagend.Vorlegen(t.Context(), leihe.Kennung,
				leihe.Bindung, domain.RechtSpielen)

			return fehler
		},
		"fortdauer": func() error {
			_, _, fehler := klagend.Fortdauer(
				t.Context(), leihe.Kennung, "p1")

			return fehler
		},
	}

	for name, weg := range wege {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if fehler := weg(); !errors.Is(fehler, errProbe) {
				t.Fatalf("erwartet die Klage, bekam %v", fehler)
			}
		})
	}
}

// **Eine Unterschrift, die kein Hex ist, wird gemeldet** — und nicht als
// falscher Schlüssel.
func TestEineUnterschriftOhneHexWirdGemeldet(t *testing.T) {
	t.Parallel()

	fehler := service.PruefeLeihe(domain.Deckleihe{Unterschrift: "zzz"})

	if !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("erwartet ErrValidation, bekam %v", fehler)
	}
}

// **Eine eigene Leihe legt niemand vor** — wer verleiht, spielt mit seinem
// eigenen Deck.
func TestEineEigeneLeiheWirdNichtVorgelegt(t *testing.T) {
	t.Parallel()

	anna, bert := zweiPeers(t)
	stelle, _, _ := eineLeihstelle(t, anna)

	leihe, _, fehler := stelle.Anbieten(t.Context(), einLeihwunsch(bert))
	if fehler != nil {
		t.Fatalf("anbieten: %v", fehler)
	}

	if _, fehler := stelle.Vorlegen(t.Context(), leihe.Kennung,
		leihe.Bindung, domain.RechtSpielen); !errors.Is(
		fehler, domain.ErrConflict) {
		t.Fatalf("erwartet ErrConflict, bekam %v", fehler)
	}
}

// **Eine fremde Leihe nimmt niemand zurück** — zurücknehmen kann nur, wer
// ausgestellt hat.
func TestEineFremdeLeiheWirdNichtWiderrufen(t *testing.T) {
	t.Parallel()

	stelle, leihe := eineAngenommeneLeihe(t, leihzeit)

	if fehler := stelle.Widerrufen(
		t.Context(), leihe.Kennung); !errors.Is(fehler, domain.ErrConflict) {
		t.Fatalf("erwartet ErrConflict, bekam %v", fehler)
	}
}

// **Ein Abzug, der keiner ist, wird abgewiesen** — auch mit passendem
// Abdruck.
func TestEinUngueltigerAbzugWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	anna, bert := zweiPeers(t)
	beimVerleiher, _, _ := eineLeihstelle(t, anna)

	leihe, _, fehler := beimVerleiher.Anbieten(
		t.Context(), einLeihwunsch(bert))
	if fehler != nil {
		t.Fatalf("anbieten: %v", fehler)
	}

	beimEntleiher, _, _ := eineLeihstelle(t, bert)

	// **Derselbe Abdruck, aber keine Karte darin.**
	if fehler := beimEntleiher.Annehmen(t.Context(), leihe,
		domain.Partiedeck{Abdruck: leihe.Abdruck}); !errors.Is(
		fehler, domain.ErrValidation) {
		t.Fatalf("erwartet ErrValidation, bekam %v", fehler)
	}
}

// **Fortdauer über eine Leihe, die diese Partie nicht trägt, wird gemeldet.**
func TestFortdauerUeberEineFremdePartieWirdGemeldet(t *testing.T) {
	t.Parallel()

	stelle, leihe := eineAngenommeneLeihe(t, leihzeit)

	if _, _, fehler := stelle.Fortdauer(
		t.Context(), leihe.Kennung, "p9"); !errors.Is(
		fehler, domain.ErrValidation) {
		t.Fatalf("erwartet ErrValidation, bekam %v", fehler)
	}
}

// **Eine abgelaufene Leihe endet nach ihrem Modus** — ohne dass jemand
// widerrufen hat.
func TestEineAbgelaufeneLeiheEndetNachDemModus(t *testing.T) {
	t.Parallel()

	stelle, leihe := eineAngenommeneLeihe(t, leihzeit)

	if _, fehler := stelle.Vorlegen(t.Context(), leihe.Kennung,
		leihe.Bindung, domain.RechtSpielen); fehler != nil {
		t.Fatalf("vorlegen: %v", fehler)
	}

	spaeter, _ := eineAngenommeneLeihe(t, leihzeit)
	_ = spaeter

	bleibt, grund, fehler := stelle.Fortdauer(
		t.Context(), leihe.Kennung, "p1")
	if fehler != nil {
		t.Fatalf("fortdauer: %v", fehler)
	}

	if !bleibt || grund == "" {
		t.Fatalf("bleibt ist %v, Grund %q", bleibt, grund)
	}
}

// **Eine Leihe mit gebrochener Unterschrift wird abgewiesen** — beim
// Annehmen und beim Vorlegen.
//
// **Zwei Stellen und nicht eine**: eine Leihe kommt herein und liegt danach
// im Bestand. Wer nur beim Hereinkommen prüfte, glaubte danach der Ablage
// (§10) — und eine Zeile im Bestand lässt sich von Hand ändern.
func TestEineGebrocheneUnterschriftWirdZweimalAbgewiesen(t *testing.T) {
	t.Parallel()

	anna, bert := zweiPeers(t)
	beimVerleiher, _, _ := eineLeihstelle(t, anna)

	leihe, abzug, fehler := beimVerleiher.Anbieten(
		t.Context(), einLeihwunsch(bert))
	if fehler != nil {
		t.Fatalf("anbieten: %v", fehler)
	}

	gebrochen := leihe
	gebrochen.Bis = gebrochen.Bis.Add(time.Hour)

	beimEntleiher, ablage, _ := eineLeihstelle(t, bert)

	if fehler := beimEntleiher.Annehmen(
		t.Context(), gebrochen, abzug); !errors.Is(
		fehler, domain.ErrValidation) {
		t.Fatalf("annehmen: erwartet ErrValidation, bekam %v", fehler)
	}

	// **Und jetzt liegt sie trotzdem da** — so, wie ein fremdes Werkzeug sie
	// hineinschriebe.
	if fehler := ablage.LeiheSchreiben(t.Context(), domain.Leihstand{
		Leihe: gebrochen, Abzug: abzug,
	}); fehler != nil {
		t.Fatalf("legen: %v", fehler)
	}

	if _, fehler := beimEntleiher.Vorlegen(t.Context(), gebrochen.Kennung,
		gebrochen.Bindung, domain.RechtSpielen); !errors.Is(
		fehler, domain.ErrValidation) {
		t.Fatalf("vorlegen: erwartet ErrValidation, bekam %v", fehler)
	}
}

// **Ohne Ausweis legt niemand vor** — auch nicht mit einer gültigen Leihe.
func TestOhneAusweisWirdNichtVorgelegt(t *testing.T) {
	t.Parallel()

	anna, bert := zweiPeers(t)
	beimVerleiher, _, _ := eineLeihstelle(t, anna)

	leihe, abzug, fehler := beimVerleiher.Anbieten(
		t.Context(), einLeihwunsch(bert))
	if fehler != nil {
		t.Fatalf("anbieten: %v", fehler)
	}

	ablage := neueLeihprobe()
	ablage.leihen[leihe.Kennung] = domain.Leihstand{Leihe: leihe, Abzug: abzug}

	ohne := service.NeueLeihstelle(ablage,
		&deckprobe{decks: map[int64]domain.Deck{}},
		ausweisprobe{fehler: errProbe}, func() time.Time { return leihzeit })

	if _, fehler := ohne.Vorlegen(t.Context(), leihe.Kennung,
		leihe.Bindung, domain.RechtSpielen); !errors.Is(fehler, errProbe) {
		t.Fatalf("erwartet die Klage, bekam %v", fehler)
	}
}

// **Ein gescheitertes Eintragen des Verbrauchs gibt den Abzug nicht heraus.**
//
// **Sonst startete eine Partie mit einer Leihe, von der niemand weiß, dass
// sie eingelöst ist** — und die zweite bekäme sie noch einmal.
func TestEinGescheitertesEintragenGibtDenAbzugNichtHeraus(t *testing.T) {
	t.Parallel()

	anna, bert := zweiPeers(t)
	beimVerleiher, _, _ := eineLeihstelle(t, anna)

	leihe, abzug, fehler := beimVerleiher.Anbieten(
		t.Context(), einLeihwunsch(bert))
	if fehler != nil {
		t.Fatalf("anbieten: %v", fehler)
	}

	ablage := neueLeihprobe()
	ablage.leihen[leihe.Kennung] = domain.Leihstand{Leihe: leihe, Abzug: abzug}
	ablage.schreibklage = errProbe

	stelle := service.NeueLeihstelle(ablage,
		&deckprobe{decks: map[int64]domain.Deck{}}, ausweisprobe{ich: bert},
		func() time.Time { return leihzeit })

	if _, fehler := stelle.Vorlegen(t.Context(), leihe.Kennung,
		leihe.Bindung, domain.RechtSpielen); !errors.Is(fehler, errProbe) {
		t.Fatalf("erwartet die Klage, bekam %v", fehler)
	}
}
