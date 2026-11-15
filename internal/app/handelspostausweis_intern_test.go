package app

// **Die Post der Anwendung weiß, wer sie ist** — Befund Z3 des Zweitblicks
// über Phase 6.8 (#1669).
//
// Der Dienst prüft nur, wenn ihm jemand einen Ausweis gibt; hier steht, dass
// die Anwendung es tut — an der echten Naht und nicht an einer Probe.

import (
	"errors"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

func TestDiePostDerAnwendungNimmtNurErsteVorgaengeFuerSich(t *testing.T) {
	anwendung, ich := eineLeitungsflaeche(t)
	ctx := anwendung.imProfil(riegelkontext(t))

	// **Ein Gegenüber, das handeln darf.**
	kontakte := storage.NeueKontaktablage(anwendung.db)
	jetzt := time.Now()

	for _, schritt := range []func() error{
		func() error { return kontakte.Begegnet(ctx, dasLeitungsgegenueber, "AB-CD", "Turfer", jetzt) },
		func() error { return kontakte.Stand(ctx, dasLeitungsgegenueber, domain.KontaktBestaetigt) },
		func() error { return kontakte.Erlauben(ctx, dasLeitungsgegenueber, domain.ZweckHandel, jetzt) },
	} {
		if fehler := schritt(); fehler != nil {
			t.Fatal(fehler)
		}
	}

	erste := func(vorgang string, parteien ...string) service.Handelsnachricht {
		angebot := domain.Handelsangebot{
			Vertragsfassung: domain.Handelsvertragsfassung, Vorgang: vorgang, Revision: 1,
			Art: domain.HandelTausch, Initiator: dasLeitungsgegenueber, Parteien: parteien,
			Zustand: domain.HandelAngeboten, Zustimmungen: []string{},
			Positionen: []domain.Handelsposition{{
				Von: dasLeitungsgegenueber, Printing: "shivan-lea", Menge: 1,
				Guete: string(domain.ZustandNearMint), Sprache: "en", Finish: "nonfoil",
			}},
		}

		return service.Handelsnachricht{
			Umschlag: domain.Handelsumschlag{
				Fassung: domain.Handelsprotokollfassung, Art: domain.NachrichtVorschlag,
				NachrichtID: "n-" + vorgang, Vorgang: vorgang, Basisrevision: 1,
				Abdruck: angebot.Abdruck(), Folge: 1, Absender: dasLeitungsgegenueber,
				Unterschrift: "sig",
			},
			Angebot: angebot,
		}
	}

	// **Zwischen dem Gegenüber und einem Dritten** — abgewiesen, und nichts
	// steht danach im Bestand.
	if _, fehler := anwendung.handelspost().Empfangen(ctx, dasLeitungsgegenueber,
		erste("v-fremd", dasLeitungsgegenueber, "bdritter")); !errors.Is(fehler, domain.ErrErlaubnis) {
		t.Errorf("für einen Dritten: %v", fehler)
	}

	if _, fehler := anwendung.handelsleitung().Stand(ctx, "v-fremd"); !errors.Is(fehler, domain.ErrNotFound) {
		t.Errorf("der fremde Vorgang steht da: %v", fehler)
	}

	// **Für dieses Profil** — angenommen.
	if antwort, fehler := anwendung.handelspost().Empfangen(ctx, dasLeitungsgegenueber,
		erste("v-meins", dasLeitungsgegenueber, ich)); fehler != nil || antwort.Art != domain.ZustellungNeu {
		t.Errorf("für mich: %+v, %v", antwort, fehler)
	}
}
