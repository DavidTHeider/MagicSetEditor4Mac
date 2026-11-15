package service_test

// **Ein Konflikt in der Übergabe führt in die Klärung** — TRADE-03.AK-04
// (#1717).
//
// **Dieselbe Operation gegen dieselbe Revision mit anderem Inhalt** hieß bis
// hierher: die Antwort von vorhin, und sonst nichts. Steht der Handel dort,
// wo die Matrix eine Klärung kennt — Übergabe, Verbuchen, Abgleich —, geht er
// jetzt hinein, und ein Mensch sieht es. **Anderswo nicht**: in der
// Verhandlung ist ein abweichender Vorschlag ein Gegenangebot und kein Streit.

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// klaerprobe merkt sich, wer was klären wollte.
type klaerprobe struct {
	gerufen []string
	klagt   error
}

func (k *klaerprobe) Schreiten(
	_ context.Context, vorgang string, schritt domain.Handelsschritt, wer string,
) (domain.Handelsangebot, error) {
	k.gerufen = append(k.gerufen, vorgang+"/"+string(schritt)+"/"+wer)

	return domain.Handelsangebot{}, k.klagt
}

// konfliktNach stellt eine Nachricht zu, dann eine abweichende — die zweite
// gegen einen Stand im genannten Zustand.
func konfliktNach(
	t *testing.T, zustand domain.Handelszustand, klaerung *klaerprobe,
	standklage ...error,
) (service.Zustellungsantwort, error) {
	t.Helper()

	briefkasten := &briefkastenprobe{}

	if _, fehler := einePoststelle(briefkasten, einHandelnder()).Empfangen(
		t.Context(), dasGegenueber, einePost(einNachfolger(), nil)); fehler != nil {
		t.Fatalf("das erste Empfangen: %v", fehler)
	}

	heute := derHeutigeStand()
	heute.Zustand = zustand

	probe := standprobe{stand: heute}
	if len(standklage) > 0 {
		probe.klagt = standklage[0]
	}

	post := service.NeueHandelspost(briefkasten, probe,
		kontaktnachschlagprobe{kontakt: einHandelnder()}).MitKlaerung(klaerung)

	anderer := einNachfolger()
	anderer.BewertungGeteilt = true

	return post.Empfangen(t.Context(), dasGegenueber,
		einePost(anderer, func(u *domain.Handelsumschlag) {
			u.NachrichtID, u.Abdruck = "n-2", anderer.Abdruck()
		}))
}

func TestEinKonfliktInDerUebergabeFuehrtInDieKlaerung(t *testing.T) {
	t.Parallel()

	klaerung := &klaerprobe{}

	antwort, fehler := konfliktNach(t, domain.HandelUebergabeOffen, klaerung)
	if fehler != nil || antwort.Art != domain.ZustellungKonflikt {
		t.Fatalf("Antwort %+v, %v", antwort, fehler)
	}

	// **Geklärt wird von dieser Seite** — nicht vom Absender.
	if len(klaerung.gerufen) != 1 || klaerung.gerufen[0] != derPostvorgang+"/klaeren/bich" {
		t.Fatalf("gerufen: %v", klaerung.gerufen)
	}
}

func TestEinKonfliktInDerVerhandlungKlaertNichts(t *testing.T) {
	t.Parallel()

	klaerung := &klaerprobe{}

	if _, fehler := konfliktNach(t, domain.HandelAngeboten, klaerung); fehler != nil {
		t.Fatal(fehler)
	}

	if len(klaerung.gerufen) != 0 {
		t.Fatalf("geklärt in der Verhandlung: %v", klaerung.gerufen)
	}
}

func TestEineGescheiterteKlaerungHaeltDieZustellungAn(t *testing.T) {
	t.Parallel()

	klaerung := &klaerprobe{klagt: errors.New("Platte voll")}

	// **Ein Fehler geht zurück**, damit der Absender es noch einmal versucht
	// — und die Klärung beim nächsten Anlauf geschieht.
	if _, fehler := konfliktNach(t, domain.HandelUebergabeOffen, klaerung); fehler == nil {
		t.Fatal("die gescheiterte Klärung ging still vorbei")
	}
}

func TestEinUnlesbarerStandHaeltDieKlaerungAn(t *testing.T) {
	t.Parallel()

	klaerung := &klaerprobe{}

	if _, fehler := konfliktNach(t, domain.HandelUebergabeOffen, klaerung,
		errors.New("Bestand zu")); fehler == nil || len(klaerung.gerufen) != 0 {
		t.Fatalf("Fehler %v, gerufen %v", fehler, klaerung.gerufen)
	}
}
