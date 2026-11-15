package service_test

// **Ein erster Vorgang entsteht nur, wo dieses Profil Partei ist** — Befund
// Z3 des Zweitblicks über Phase 6.8
// (#1669).
//
// # Der Befund
//
// Die Post prüfte, dass der **Absender** unter den Parteien steht. Ob das
// **empfangende** Profil es tut, prüfte niemand: ein berechtigter Kontakt
// legte einen Vorgang zwischen sich und einem Dritten in einem fremden
// Bestand an, und die Übersicht blendete ihn dort aus.

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/peer"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// einePostMitAusweis ist eine Post ohne Vorgänge, die weiß, wer sie ist.
func einePostMitAusweis(briefkasten *briefkastenprobe, ausweis service.Ausweisstelle) *service.Handelspost {
	return service.NeueHandelspost(briefkasten,
		standprobe{klagt: domain.MitGrund(domain.ErrNotFound, nil,
			"den Handelsvorgang gibt es nicht")},
		kontaktnachschlagprobe{kontakt: einHandelnder()}).MitAusweis(ausweis)
}

func TestEinErsterVorgangMussDiesesProfilNennen(t *testing.T) {
	t.Parallel()

	ich, fehler := peer.Erzeugen(nil)
	if fehler != nil {
		t.Fatal(fehler)
	}

	fuerMich := einErstesVomGegenueber()
	fuerMich.Parteien = []string{dasGegenueber, string(ich.Kennung())}
	fuerMich.Positionen[0].Von = dasGegenueber

	briefkasten := &briefkastenprobe{}
	if antwort, fehler := einePostMitAusweis(briefkasten, ausweisprobe{ich: ich}).Empfangen(
		context.Background(), dasGegenueber, einePost(fuerMich, nil)); fehler != nil ||
		antwort.Art != domain.ZustellungNeu {
		t.Fatalf("für mich: %+v, %v", antwort, fehler)
	}

	// **Zwischen dem Absender und einem Dritten** — hier nicht.
	fuerAndere := einErstesVomGegenueber()
	fuerAndere.Parteien = []string{dasGegenueber, "bdritter"}
	fuerAndere.Positionen[0].Von = dasGegenueber

	andererKasten := &briefkastenprobe{}
	if _, fehler := einePostMitAusweis(andererKasten, ausweisprobe{ich: ich}).Empfangen(
		context.Background(), dasGegenueber, einePost(fuerAndere, nil)); !errors.Is(fehler, domain.ErrErlaubnis) {
		t.Errorf("für einen Dritten: %v", fehler)
	}

	if len(andererKasten.angebote) != 0 {
		t.Errorf("geschrieben wurde %+v", andererKasten.angebote)
	}

	// **Ohne Ausweis keine Antwort** — und keine Zeile.
	klemmt := errors.New("der Schlüsselbund schweigt")
	if _, fehler := einePostMitAusweis(&briefkastenprobe{}, ausweisprobe{fehler: klemmt}).Empfangen(
		context.Background(), dasGegenueber, einePost(fuerMich, nil)); !errors.Is(fehler, klemmt) {
		t.Errorf("ohne Ausweis: %v", fehler)
	}
}
