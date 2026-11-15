package app

// **Hin, zurück und wieder hin gibt denselben Fehler** — Befund 4 des
// Nachlaufs über Phase 1 von 0.9.0 (F03, #2342).
//
// `TestJederFehlercodeFindetSeineArtZurueck` fragt mit `errors.Is` nach der
// Klasse, und die stimmte. **Die Hinübersetzung verlangt bei drei Codes mehr**:
// `drossel` eine Kontingentgrenze, `abgeschaltet` den Schalter oder den
// Dienst, `netz` den Fall und den Wirt. Ohne sie wurde aus jedem beim
// zweiten Hinweg `unbekannt`.

import (
	"context"
	"maps"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestHinZurueckUndWiederHinGibtDenselbenFehler(t *testing.T) {
	t.Parallel()

	// **Je Code ein Fehler, wie er entsteht** — und bei den dreien mit
	// Einzelheiten jeder Schlüssel, den sie tragen können.
	vertreter := map[Fehlercode][]error{
		CodeNichtGefunden: {
			domain.ErrNotFound,
			// **Die genaueren reisen mit ihrem Schlüssel** (#2491): die
			// Kommandozeile gab nach dem Rückweg `nichtGefunden` aus, wo das
			// Fenster „nicht zu erreichen" bekam.
			domain.ErrKeinSetMitStil, domain.ErrOhnePeeridentitaet, domain.ErrPeerNichtErreichbar,
			domain.ErrPeerOhneAdresse,
		},
		CodeKonflikt:        {domain.ErrUeberholt},
		CodeValidierung:     {domain.ErrValidation},
		CodeBelegt:          {domain.ErrBelegt},
		CodeErlaubnis:       {domain.ErrErlaubnis},
		CodeSchluesselFehlt: {domain.ErrSchluesselFehlt},
		CodeNurLesend:       {domain.ErrNurLesend},
		CodeVoll:            {domain.ErrVoll},
		CodeAbgebrochen:     {context.Canceled, errAnwendungEndet},
		CodeAbgeschaltet:    {&domain.Abschaltung{Schalter: "netz.cardtrader"}, domain.DienstIstAus("scryfall")},
		CodeNetz: {
			domain.NeuerGegenstellenfehler(domain.ErrNameUnaufloesbar, nil, "api.beispiel.invalid:443", 0),
			domain.NeuerGegenstellenfehler(domain.ErrNichtErreichbar, nil, "api.beispiel.invalid:443", 0),
			domain.NeuerGegenstellenfehler(domain.ErrKeineAntwort, nil, "api.beispiel.invalid:443", 1500*time.Millisecond),
			domain.NeuerGegenstellenfehler(domain.ErrAntwortStockt, nil, "api.beispiel.invalid:443", 30*time.Second),
			domain.NeuerGegenstellenfehler(domain.ErrVerbindungAbgerissen, nil, "api.beispiel.invalid:443", 0),
			domain.NeuerGegenstellenfehler(domain.ErrProxyAnmeldung, nil, "proxy.beispiel.invalid:3128", 0),
			domain.NeuerGegenstellenfehler(domain.ErrGedrosselt, nil, "api.beispiel.invalid:443", 0),
			domain.NeueDrosselung(nil, "api.beispiel.invalid:443", time.Date(2026, 10, 14, 21, 30, 0, 0, time.Local)),
		},
		CodeDrossel: {domain.NeueKontingentgrenze(1500 * time.Millisecond), domain.NeueKontingentgrenze(42 * time.Second)},
	}

	for _, code := range fehlercodesImQuelltext(t) {
		if code == CodeUnbekannt {
			continue
		}

		fehler, bekannt := vertreter[code]
		if !bekannt {
			t.Errorf("%q hat hier keinen Vertreter — trag einen ein", code)

			continue
		}

		for _, ursprung := range fehler {
			hin := Uebersetze(t.Context(), ursprung)
			if hin.Code != code {
				t.Fatalf("der Vertreter %v kam als %q an, nicht als %q", ursprung, hin.Code, code)
			}

			wieder := Uebersetze(t.Context(), hin.AlsFehler())

			if wieder.Code != hin.Code {
				t.Errorf("%q kam beim zweiten Hinweg als %q an", code, wieder.Code)
			}

			// **Bei den vieren mit Einzelheiten reist auch der Satz**: der
			// Schlüssel und die Werte, die er einsetzt.
			if code == CodeDrossel || code == CodeAbgeschaltet || code == CodeNetz || code == CodeNichtGefunden {
				if wieder.Schluessel != hin.Schluessel || !maps.Equal(wieder.Werte, hin.Werte) {
					t.Errorf("%q: aus %s %v wurde %s %v", code, hin.Schluessel, hin.Werte, wieder.Schluessel, wieder.Werte)
				}
			}
		}
	}
}

// **Eine Uhrzeit, die sich nicht lesen lässt, bleibt eine Drosselung** — ohne
// Ende, und mit dem Satz dafür.
func TestEineUnlesbareUhrzeitBleibtEineDrosselung(t *testing.T) {
	t.Parallel()

	fehler := &Fehler{
		Code: CodeNetz, Schluessel: domain.SchluesselGedrosseltBis, Meldung: "der Satz",
		Werte: map[string]string{"wirt": "api.beispiel.invalid:443", "uhrzeit": "bald"},
	}

	wieder := Uebersetze(t.Context(), fehler.AlsFehler())
	if wieder.Code != CodeNetz || wieder.Schluessel != domain.SchluesselGedrosselt || wieder.Werte["wirt"] != "api.beispiel.invalid:443" {
		t.Errorf("kam als %s %s %v an", wieder.Code, wieder.Schluessel, wieder.Werte)
	}
}
