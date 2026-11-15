package service_test

// **Eine Sichtanfrage über die Leitung** — P11
// (#2379),
// Befunde C-15 und Z-02: `sichtanfrage` stand im Wortschatz und hatte weder
// Sender noch Empfänger; wer sehen wollte, was ein Gegenüber zum Tausch
// freigibt, brauchte eine Datei oder eingefügtes JSON.
//
// # Was hier gemessen wird
//
// **Über eine echte Schleife** (`internal/leitung`): die Anfrage, die
// begrenzte Antwort aus der Freigabe des Gefragten, die Absage nach einem
// Widerruf — und was der Fragende **nicht** annimmt: eine fremde Fassung,
// eine Antwort auf eine andere Anfrage, einen anderen Absender, keine Antwort.

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/leitung"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/peer"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// sichtgeberprobe gibt eine feste Sicht — oder eine Absage.
type sichtgeberprobe struct {
	zeilen []domain.Sichtkarte
	klagt  error

	// gefragt hält, wer mit welcher Sitzung und welchem Zweck fragte.
	gefragt *[]string
}

func (s sichtgeberprobe) Sicht(
	_ context.Context, empfaenger, sitzung string, zweck domain.Zweck,
) ([]domain.Sichtkarte, error) {
	if s.gefragt != nil {
		*s.gefragt = append(*s.gefragt, empfaenger, sitzung, string(zweck))
	}

	return s.zeilen, s.klagt
}

// sichtpeers sind Anna, die fragt, und Bert, der antwortet.
func sichtpeers(t *testing.T) (anna, bert peer.Identitaet) {
	t.Helper()

	anna, fehler := peer.Erzeugen(nil)
	if fehler != nil {
		t.Fatal(fehler)
	}

	bert, fehler = peer.Erzeugen(nil)
	if fehler != nil {
		t.Fatal(fehler)
	}

	return anna, bert
}

// bertsPost beantwortet Sichtanfragen mit der Sicht aus `geber`.
func bertsPost(anna, bert peer.Identitaet, geber service.Sichtgeber) *service.Handelspost {
	return service.NeueHandelspost(&briefkastenprobe{}, standprobe{},
		kontaktnachschlagprobe{kontakt: domain.Peerkontakt{
			Kennung: string(anna.Kennung()), Stand: domain.KontaktBestaetigt,
			Rechte: []domain.Zweck{domain.ZweckHandel},
		}}).MitAusweis(ausweisprobe{ich: bert}).MitSichtgeber(geber)
}

func freigegebeneKarten() []domain.Sichtkarte {
	return []domain.Sichtkarte{
		{Referenz: "r1", Printing: "p-1", Menge: 2, Zustand: domain.ZustandNearMint, Finish: domain.FinishNonfoil, Bezeichnung: "Sturmrufer"},
		{Referenz: "r2", Printing: "p-2", Menge: 1, Zustand: domain.ZustandNearMint, Finish: domain.FinishFoil, Bezeichnung: "Aschenwacht"},
		// **Dieselbe Zeile zweimal** — der Fragende nimmt sie einmal.
		{Referenz: "r1", Printing: "p-1", Menge: 2, Zustand: domain.ZustandNearMint, Finish: domain.FinishNonfoil, Bezeichnung: "Sturmrufer"},
	}
}

func TestEineSichtanfrageBekommtDieFreigegebeneSicht(t *testing.T) {
	t.Parallel()

	anna, bert := sichtpeers(t)
	hier, dort := leitung.Schleife(service.Handelskanal)
	ctx := t.Context()

	var gefragt []string
	post := bertsPost(anna, bert, sichtgeberprobe{zeilen: freigegebeneKarten(), gefragt: &gefragt})

	beantwortet := make(chan error, 1)
	go func() {
		_, fehler := post.EmpfangeUeber(ctx, dort, string(anna.Kennung()))
		beantwortet <- fehler
	}()

	zeilen, fehler := service.NeuerSichtfrager(ausweisprobe{ich: anna}).
		Fragen(ctx, hier, string(bert.Kennung()))
	if fehler != nil {
		t.Fatalf("Fragen: %v", fehler)
	}

	if fehler := <-beantwortet; fehler != nil {
		t.Fatalf("Beantworten: %v", fehler)
	}

	if len(zeilen) != 2 || zeilen[0].Bezeichnung != "Sturmrufer" || zeilen[1].Referenz != "r2" {
		t.Errorf("Zeilen: %+v", zeilen)
	}

	// **Gefragt wurde für Anna, für den Handel, mit einer Sitzung je Anfrage.**
	if len(gefragt) != 3 || gefragt[0] != string(anna.Kennung()) || gefragt[2] != string(domain.ZweckHandel) || gefragt[1] == "" {
		t.Errorf("gefragt: %v", gefragt)
	}
}

func TestNachEinemWiderrufKommtKeineSicht(t *testing.T) {
	t.Parallel()

	anna, bert := sichtpeers(t)
	hier, dort := leitung.Schleife(service.Handelskanal)
	ctx := t.Context()

	widerrufen := domain.MitGrund(domain.ErrErlaubnis, nil, "Für diesen Zweck ist nichts freigegeben.")
	post := bertsPost(anna, bert, sichtgeberprobe{klagt: widerrufen})

	go func() { _, _ = post.EmpfangeUeber(ctx, dort, string(anna.Kennung())) }() //nolint:errcheck // gemessen wird die Antwort

	zeilen, fehler := service.NeuerSichtfrager(ausweisprobe{ich: anna}).
		Fragen(ctx, hier, string(bert.Kennung()))
	if !errors.Is(fehler, domain.ErrErlaubnis) || zeilen != nil {
		t.Fatalf("nach dem Widerruf: %v, %+v", fehler, zeilen)
	}
}

func TestOhneAntwortEndetDieAnfrageMitIhrerFrist(t *testing.T) {
	t.Parallel()

	anna, bert := sichtpeers(t)
	hier, _ := leitung.Schleife(service.Handelskanal)

	beginn := time.Now()
	zeilen, fehler := service.NeuerSichtfrager(ausweisprobe{ich: anna}).MitFrist(50*time.Millisecond).
		Fragen(t.Context(), hier, string(bert.Kennung()))

	if fehler == nil || zeilen != nil {
		t.Fatalf("ohne Antwort: %v, %+v", fehler, zeilen)
	}

	if dauer := time.Since(beginn); dauer > 5*time.Second {
		t.Errorf("die Frist hielt nicht: %v", dauer)
	}
}

// antwortVon schreibt eine selbstgebaute Antwort auf die nächste Anfrage.
func antwortVon(t *testing.T, dort *leitung.Leitung, bert peer.Identitaet, aendern func(*service.Handelsnachricht)) {
	t.Helper()

	roh, fehler := dort.Empfange(t.Context())
	if fehler != nil {
		t.Errorf("Anfrage lesen: %v", fehler)

		return
	}

	var anfrage service.Handelsnachricht
	if fehler := json.Unmarshal(roh, &anfrage); fehler != nil {
		t.Errorf("Anfrage deuten: %v", fehler)

		return
	}

	antwort := service.Handelsnachricht{
		Umschlag: domain.Handelsumschlag{
			Fassung: domain.Handelsprotokollfassung, Art: domain.NachrichtSicht,
			NachrichtID: "antwort-1", Vorgang: anfrage.Umschlag.Vorgang, Folge: 1,
			Absender: string(bert.Kennung()), Unterschrift: "unterschrieben",
		},
		Sicht: freigegebeneKarten(),
	}
	aendern(&antwort)

	ausgabe, _ := json.Marshal(antwort) //nolint:errcheck // ein fester Verbund
	if fehler := dort.Sende(t.Context(), ausgabe); fehler != nil {
		t.Errorf("Antwort senden: %v", fehler)
	}
}

func TestDerFragendeNimmtKeineFremdeAntwort(t *testing.T) {
	t.Parallel()

	for name, aendern := range map[string]func(*service.Handelsnachricht){
		"eine andere Fassung": func(n *service.Handelsnachricht) { n.Umschlag.Fassung = domain.Handelsprotokollfassung + 1 },
		"eine andere Art":     func(n *service.Handelsnachricht) { n.Umschlag.Art = domain.NachrichtVorschlag },
		"ein anderer Absender": func(n *service.Handelsnachricht) {
			n.Umschlag.Absender = "bdritter"
		},
		"eine andere Anfrage": func(n *service.Handelsnachricht) { n.Umschlag.Vorgang = "sicht/eine-andere" },
		"ohne Unterschrift":   func(n *service.Handelsnachricht) { n.Umschlag.Unterschrift = "" },
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			anna, bert := sichtpeers(t)
			hier, dort := leitung.Schleife(service.Handelskanal)

			go antwortVon(t, dort, bert, aendern)

			zeilen, fehler := service.NeuerSichtfrager(ausweisprobe{ich: anna}).
				Fragen(t.Context(), hier, string(bert.Kennung()))
			if fehler == nil || zeilen != nil {
				t.Errorf("angenommen: %+v", zeilen)
			}
		})
	}

	// **Gegenprobe: unverändert geht sie durch** — sonst bewiese das Nein
	// darüber nur eine Antwort, die nie passt.
	anna, bert := sichtpeers(t)
	hier, dort := leitung.Schleife(service.Handelskanal)

	go antwortVon(t, dort, bert, func(*service.Handelsnachricht) {})

	if zeilen, fehler := service.NeuerSichtfrager(ausweisprobe{ich: anna}).
		Fragen(t.Context(), hier, string(bert.Kennung())); fehler != nil || len(zeilen) != 2 {
		t.Errorf("unverändert: %v, %+v", fehler, zeilen)
	}
}

func TestDieAntwortGehtNurAnDenSprecher(t *testing.T) {
	t.Parallel()

	anna, bert := sichtpeers(t)
	post := bertsPost(anna, bert, sichtgeberprobe{zeilen: freigegebeneKarten()})

	anfrage := service.Handelsnachricht{Umschlag: domain.Handelsumschlag{
		Fassung: domain.Handelsprotokollfassung, Art: domain.NachrichtSichtanfrage,
		NachrichtID: "a-1", Vorgang: "sicht/a-1", Folge: 1,
		Absender: "bdritter", Unterschrift: "unterschrieben",
	}}

	// **Wer im Umschlag steht, ist eine Behauptung** — gesprochen hat Anna.
	if _, fehler := post.EmpfangeUeber(t.Context(), dortMit(t, anfrage), string(anna.Kennung())); !errors.Is(fehler, domain.ErrErlaubnis) {
		t.Errorf("ein fremder Absender: %v", fehler)
	}

	// **Und ohne Sichtgeber wird nichts beantwortet.**
	anfrage.Umschlag.Absender = string(anna.Kennung())
	ohne := service.NeueHandelspost(&briefkastenprobe{}, standprobe{},
		kontaktnachschlagprobe{kontakt: einHandelnder()}).MitAusweis(ausweisprobe{ich: bert})

	if _, fehler := ohne.EmpfangeUeber(t.Context(), dortMit(t, anfrage), string(anna.Kennung())); fehler == nil {
		t.Error("eine Post ohne Sichtgeber beantwortete die Anfrage")
	}
}

// dortMit gibt eine Leitung, auf der die Anfrage schon wartet.
func dortMit(t *testing.T, anfrage service.Handelsnachricht) *leitung.Leitung {
	t.Helper()

	hier, dort := leitung.Schleife(service.Handelskanal)
	roh, _ := json.Marshal(anfrage) //nolint:errcheck // ein fester Verbund

	if fehler := hier.Sende(t.Context(), roh); fehler != nil {
		t.Fatal(fehler)
	}

	return dort
}
