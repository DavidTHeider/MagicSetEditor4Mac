package service_test

// **Die Fehlerwege einer Sichtanfrage** — P11
// (#2379):
// jeder betreten, und keiner lässt eine halbe Antwort zurück (§10).

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/leitung"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// eineSichtanfrage ist eine Anfrage, wie Anna sie schickt.
func eineSichtanfrage(absender string) service.Handelsnachricht {
	return service.Handelsnachricht{Umschlag: domain.Handelsumschlag{
		Fassung: domain.Handelsprotokollfassung, Art: domain.NachrichtSichtanfrage,
		NachrichtID: "a-1", Vorgang: "sicht/a-1", Folge: 1,
		Absender: absender, Unterschrift: "unterschrieben",
	}}
}

func TestDerGefragteAntwortetNichtAufEineKrummeAnfrage(t *testing.T) {
	t.Parallel()

	anna, bert := sichtpeers(t)
	anfrage := eineSichtanfrage(string(anna.Kennung()))
	anfrage.Umschlag.Fassung = domain.Handelsprotokollfassung + 1

	post := bertsPost(anna, bert, sichtgeberprobe{zeilen: freigegebeneKarten()})
	if _, fehler := post.EmpfangeUeber(t.Context(), dortMit(t, anfrage), string(anna.Kennung())); fehler == nil {
		t.Error("eine Anfrage in fremder Fassung wurde beantwortet")
	}
}

func TestDerGefragteOhneAusweisAntwortetNicht(t *testing.T) {
	t.Parallel()

	anna, bert := sichtpeers(t)
	kaputt := errors.New("der Schlüsselbund ist zu")

	post := bertsPost(anna, bert, sichtgeberprobe{zeilen: freigegebeneKarten()}).
		MitAusweis(ausweisprobe{fehler: kaputt})
	if _, fehler := post.EmpfangeUeber(t.Context(), dortMit(t, eineSichtanfrage(string(anna.Kennung()))),
		string(anna.Kennung())); !errors.Is(fehler, kaputt) {
		t.Errorf("ohne Ausweis: %v", fehler)
	}
}

func TestEinLesefehlerDesGefragtenIstEineAbsage(t *testing.T) {
	t.Parallel()

	anna, bert := sichtpeers(t)
	hier, dort := leitung.Schleife(service.Handelskanal)
	post := bertsPost(anna, bert, sichtgeberprobe{klagt: errors.New("die Platte ist weg")})

	beantwortet := make(chan service.Zustellungsantwort, 1)
	go func() {
		antwort, _ := post.EmpfangeUeber(t.Context(), dort, string(anna.Kennung())) //nolint:errcheck // gemessen wird die Antwort
		beantwortet <- antwort
	}()

	zeilen, fehler := service.NeuerSichtfrager(ausweisprobe{ich: anna}).Fragen(t.Context(), hier, string(bert.Kennung()))
	if !errors.Is(fehler, domain.ErrValidation) || errors.Is(fehler, domain.ErrErlaubnis) || zeilen != nil {
		t.Errorf("nach einem Lesefehler drüben: %v, %+v", fehler, zeilen)
	}

	if antwort := <-beantwortet; antwort.Ergebnis != "abgesagt" {
		t.Errorf("drüben: %+v", antwort)
	}
}

func TestEineAbgerisseneStreckeTraegtKeineSicht(t *testing.T) {
	t.Parallel()

	anna, bert := sichtpeers(t)

	// **Beim Gefragten**: die Anfrage kam, die Antwort geht nicht hinaus.
	post := bertsPost(anna, bert, sichtgeberprobe{zeilen: freigegebeneKarten()})
	strecke := &antwortlosestrecke{anfrage: eineSichtanfrage(string(anna.Kennung()))}

	if _, fehler := post.EmpfangeUeber(t.Context(), strecke, string(anna.Kennung())); fehler == nil {
		t.Error("die abgerissene Antwort klagte nicht")
	}

	// **Beim Fragenden**: schon die Anfrage geht nicht hinaus.
	if zeilen, fehler := service.NeuerSichtfrager(ausweisprobe{ich: anna}).
		Fragen(t.Context(), &stolperstrecke{}, string(bert.Kennung())); fehler == nil || zeilen != nil {
		t.Errorf("die abgerissene Anfrage: %v, %+v", fehler, zeilen)
	}

	// **Und ohne Ausweis fragt er gar nicht.**
	kaputt := errors.New("der Schlüsselbund ist zu")
	if _, fehler := service.NeuerSichtfrager(ausweisprobe{fehler: kaputt}).
		Fragen(t.Context(), &stolperstrecke{traegt: 1}, string(bert.Kennung())); !errors.Is(fehler, kaputt) {
		t.Errorf("ohne Ausweis: %v", fehler)
	}
}

func TestEineUnlesbareAntwortIstKeineSicht(t *testing.T) {
	t.Parallel()

	anna, bert := sichtpeers(t)
	hier, dort := leitung.Schleife(service.Handelskanal)

	go func() {
		if _, fehler := dort.Empfange(t.Context()); fehler == nil {
			_ = dort.Sende(t.Context(), []byte("kein JSON")) //nolint:errcheck // gemessen wird der Fragende
		}
	}()

	if zeilen, fehler := service.NeuerSichtfrager(ausweisprobe{ich: anna}).
		Fragen(t.Context(), hier, string(bert.Kennung())); !errors.Is(fehler, domain.ErrValidation) || zeilen != nil {
		t.Errorf("eine unlesbare Antwort: %v, %+v", fehler, zeilen)
	}
}

// antwortlosestrecke liefert eine Anfrage und nimmt keine Antwort an.
type antwortlosestrecke struct {
	anfrage service.Handelsnachricht
}

func (a *antwortlosestrecke) Empfange(context.Context) ([]byte, error) {
	roh, _ := json.Marshal(a.anfrage) //nolint:errcheck // ein fester Verbund

	return roh, nil
}

func (a *antwortlosestrecke) Sende(context.Context, []byte) error {
	return errors.New("die Leitung riss ab")
}
