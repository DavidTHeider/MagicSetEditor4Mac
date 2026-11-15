package domain_test

// **Ordner, in denen Dialoge aufgehen** — OPS-02.AK-01
// (#1719):
// „relative Pfade und systemabhängige Expansion haben eine eindeutige
// Spezifikation." Hier ist sie: absolut oder mit `~/` vorn, sonst nichts.

import (
	"errors"
	"slices"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestPruefeOrdnerangabe(t *testing.T) {
	t.Parallel()

	faelle := []struct {
		name string
		wert string
		gut  bool
	}{
		{"leer heißt: wo das Betriebssystem ihn öffnet", "", true},
		{"ein absoluter Pfad", "/Volumes/Karten/Ausfuhr", true},
		{"das Zuhause", "~", true},
		{"unter dem Zuhause", "~/Dokumente/Sets", true},
		{"ein relativer Pfad — relativ wozu?", "Dokumente/Sets", false},
		{"ein Punkt vorn ist auch relativ", "./Sets", false},
		{"ein fremdes Zuhause wird nicht gedehnt", "~anna/Sets", false},
		{"eine Umgebungsvariable wird nicht gedehnt", "$HOME/Sets", false},
		{"ein Schritt nach oben verschleiert das Ziel", "/Volumes/Karten/../Privat", false},
		{"ein Nullbyte", "/tmp/a\x00b", false},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			fehler := domain.PruefeOrdnerangabe(fall.wert)
			if fall.gut && fehler != nil {
				t.Fatalf("%q abgewiesen: %v", fall.wert, fehler)
			}

			if !fall.gut && !errors.Is(fehler, domain.ErrValidation) {
				t.Fatalf("%q angenommen oder falsch abgewiesen: %v", fall.wert, fehler)
			}
		})
	}
}

func TestOrdnerAusDehntNurDasZuhause(t *testing.T) {
	t.Parallel()

	for angabe, erwartet := range map[string]string{
		"":                 "",
		"~":                "/Users/ich",
		"~/Sets":           "/Users/ich/Sets",
		"/Volumes/K/":      "/Volumes/K",
		"/Volumes/K//Sets": "/Volumes/K/Sets",
	} {
		if gedehnt := domain.OrdnerAus(angabe, "/Users/ich"); gedehnt != erwartet {
			t.Errorf("OrdnerAus(%q) = %q, erwartet %q", angabe, gedehnt, erwartet)
		}
	}
}

// **Vier Ordner, jede eine Einstellung wie jede andere** — auf dem Gerät,
// unter den Vorgaben, geprüft und sofort wirksam.
func TestDieOrdnerSindEinstellungenDesGeraets(t *testing.T) {
	t.Parallel()

	ordner := domain.Ordnerschluessel()
	if len(ordner) != 4 || !slices.Contains(ordner, domain.VorgabeHochladepfad) {
		t.Fatalf("Ordner: %v", ordner)
	}

	for _, schluessel := range ordner {
		ebene, _ := domain.EbeneVon(schluessel)
		if ebene != domain.EbeneGeraet || domain.EinstellungsBereich(schluessel) != domain.BereichVorgaben ||
			domain.EinstellungsArt(schluessel) != domain.ArtGeprueft ||
			domain.EinstellungsWirkung(schluessel) != domain.WirkungSofort {
			t.Errorf("%s: Ebene %q, Bereich %q, Art %q", schluessel, ebene,
				domain.EinstellungsBereich(schluessel), domain.EinstellungsArt(schluessel))
		}

		if fehler := domain.PruefeEinstellung(schluessel, "relativ"); !errors.Is(fehler, domain.ErrValidation) {
			t.Errorf("%s nimmt einen relativen Pfad: %v", schluessel, fehler)
		}
	}

	if !domain.IstOrdnerangabe(domain.VorgabeSicherungspfad) || domain.IstOrdnerangabe(domain.EinstellungSprache) {
		t.Error("IstOrdnerangabe unterscheidet nicht")
	}
}

// **Der Befund trägt seinen eigenen Schlüssel und die Angabe** — sonst zeigte
// das Fenster nur „Die Eingabe ist nicht gültig", und der Satz, der sagt,
// was ein Ordner sein muss, ginge verloren.
func TestEineFalscheOrdnerangabeSagtWasFehlt(t *testing.T) {
	t.Parallel()

	for _, wert := range []string{"relativ", "/a/../b"} {
		var befund *domain.FieldError
		if !errors.As(domain.PruefeOrdnerangabe(wert), &befund) ||
			befund.Key != domain.SchluesselOrdnerangabe || befund.Werte["wert"] != wert {
			t.Errorf("%q: %+v", wert, befund)
		}
	}
}
