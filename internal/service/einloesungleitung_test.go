package service_test

// **Die Einlösung über die Leitung, beide Richtungen** — NET-01b
// (#1954).

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/peer"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// geantwortet liest die eine Antwort, die Anna geschickt hat.
func geantwortet(t *testing.T, leitung *probeleitung) domain.Einloeseergebnis {
	t.Helper()

	if len(leitung.gesendet) != 1 {
		t.Fatalf("gesendet: %d", len(leitung.gesendet))
	}

	antwort, fehler := domain.ZerlegeEinloeseantwort(leitung.gesendet[0])
	if fehler != nil {
		t.Fatalf("die Antwort: %v", fehler)
	}

	return antwort.Ergebnis
}

func TestAnnaBeantwortetEineEinloesung(t *testing.T) {
	t.Parallel()

	_, werk, ablage, _, nonce := annaStelltAus(t)
	bert := mussIdentitaet(t)

	leitung := &probeleitung{
		stand: service.Sitzungsstand{Gegenueber: bert.Kennung()},
		antworten: [][]byte{kodiert(t, domain.Einloesung{
			Nonce: nonce, Adresse: "192.0.2.8:41234", Name: "Bert",
		}.Kodiere)},
	}

	if fehler := werk.EinloesungBeantworten(t.Context(), leitung); fehler != nil {
		t.Fatalf("EinloesungBeantworten: %v", fehler)
	}

	if ergebnis := geantwortet(t, leitung); ergebnis != domain.EinloesungAufgenommen {
		t.Errorf("geantwortet: %q", ergebnis)
	}

	if kontakt := ablage.kontakte[string(bert.Kennung())]; kontakt.SelbstgewaehlterName != "Bert" ||
		kontakt.Adresse != "192.0.2.8:41234" {
		t.Errorf("Kontakt: %+v", kontakt)
	}
}

// **Jedes Nein wird beantwortet** — mit seinem Ergebnis und nicht mit einem
// Abriss, der drüben wie eine Zeitüberschreitung aussähe (AK-03).
func TestJedesNeinWirdBeantwortet(t *testing.T) {
	t.Parallel()

	for name, fall := range map[string]struct {
		nachricht []byte
		ergebnis  domain.Einloeseergebnis
		klasse    error
	}{
		"fremd":          {[]byte(`{"fassung":1,"nonce":"` + eineFremdeNonce() + `"}`), domain.EinloesungFremd, domain.ErrEinladungFremd},
		"krumm":          {[]byte(`{`), domain.EinloesungGescheitert, domain.ErrValidation},
		"neuere Fassung": {[]byte(`{"fassung":2,"nonce":"x"}`), domain.EinloesungZuNeu, domain.ErrEinloesefassung},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, werk, _, _, _ := annaStelltAus(t)
			leitung := &probeleitung{
				stand:     service.Sitzungsstand{Gegenueber: mussIdentitaet(t).Kennung()},
				antworten: [][]byte{fall.nachricht},
			}

			if fehler := werk.EinloesungBeantworten(t.Context(), leitung); !errors.Is(fehler, fall.klasse) {
				t.Errorf("Befund: %v", fehler)
			}

			if ergebnis := geantwortet(t, leitung); ergebnis != fall.ergebnis {
				t.Errorf("geantwortet: %q", ergebnis)
			}
		})
	}
}

// **Was die Leitung nicht trägt, wird gemeldet** — und ohne Nachricht
// gibt es nichts zu beantworten.
func TestEineKlemmendeLeitungWirdGemeldet(t *testing.T) {
	t.Parallel()

	kaputt := errors.New("klemmt")

	for name, einrichten := range map[string]func(*probeleitung){
		"Empfange": func(l *probeleitung) { l.empfangefehler = kaputt },
		"Sende":    func(l *probeleitung) { l.sendefehler = kaputt },
		"Leere":    func(l *probeleitung) { l.leerefehler = kaputt },
	} {
		_, werk, _, _, nonce := annaStelltAus(t)
		leitung := &probeleitung{
			stand:     service.Sitzungsstand{Gegenueber: mussIdentitaet(t).Kennung()},
			antworten: [][]byte{kodiert(t, domain.Einloesung{Nonce: nonce}.Kodiere)},
		}
		einrichten(leitung)

		if fehler := werk.EinloesungBeantworten(t.Context(), leitung); !errors.Is(fehler, kaputt) {
			t.Errorf("%s: %v", name, fehler)
		}
	}
}

// bertMeldetSich gibt Berts Werk und einen Boten, der eine Probeleitung
// reicht — und aufschreibt, wen er wählen sollte.
func bertMeldetSich(t *testing.T, leitung *probeleitung, fehler error) (
	*service.Einladungswerk, service.Einloesebote, *[]peer.Kennung,
) {
	t.Helper()

	gewaehlt := &[]peer.Kennung{}
	werk := service.NeuesEinladungswerk(festerAusweis{ich: mussIdentitaet(t)}, festeRotation{},
		neueEinladungsablage(), func() time.Time { return einladungsuhr })

	return werk, func(_ context.Context, kennung peer.Kennung) (service.Einloeseleitung, error) {
		*gewaehlt = append(*gewaehlt, kennung)
		if fehler != nil {
			return nil, fehler
		}

		return leitung, nil
	}, gewaehlt
}

func TestBertMeldetSichBeiAnna(t *testing.T) {
	t.Parallel()

	anna, _, _, code, nonce := annaStelltAus(t)
	leitung := &probeleitung{antworten: [][]byte{domain.AntwortAuf(nil).Kodiere()}}
	werk, bote, gewaehlt := bertMeldetSich(t, leitung, nil)

	if fehler := werk.Melden(t.Context(), bote, code,
		domain.Einloesung{Adresse: "192.0.2.8:41234", Name: "Bert"}); fehler != nil {
		t.Fatalf("Melden: %v", fehler)
	}

	if len(*gewaehlt) != 1 || (*gewaehlt)[0] != anna.Kennung() || !leitung.geschlossen {
		t.Errorf("gewählt %v, geschlossen %v", *gewaehlt, leitung.geschlossen)
	}

	einloesung, fehler := domain.ZerlegeEinloesung(leitung.gesendet[0])
	if fehler != nil || einloesung.Nonce != nonce || einloesung.Adresse != "192.0.2.8:41234" ||
		einloesung.Name != "Bert" {
		t.Errorf("gesendet: %+v, %v", einloesung, fehler)
	}
}

// **Jede Antwort und jedes Schweigen wird zu einem benannten Fehler.**
func TestBertHoertJedeAntwort(t *testing.T) {
	t.Parallel()

	kaputt := errors.New("keine Route")

	for name, fall := range map[string]struct {
		leitung *probeleitung
		wahl    error
		klasse  error
	}{
		"verbraucht": {&probeleitung{antworten: [][]byte{
			[]byte(`{"fassung":1,"ergebnis":"verbraucht"}`),
		}}, nil, domain.ErrEinladungVerbraucht},
		"gedrosselt": {&probeleitung{antworten: [][]byte{
			[]byte(`{"fassung":1,"ergebnis":"gedrosselt"}`),
		}}, nil, domain.ErrEinloesenGedrosselt},
		"krumme Antwort":  {&probeleitung{antworten: [][]byte{[]byte(`{`)}}, nil, domain.ErrValidation},
		"nicht zu wählen": {&probeleitung{}, kaputt, domain.ErrEinloesungUnbeantwortet},
		"aufgelegt":       {&probeleitung{}, nil, domain.ErrEinloesungUnbeantwortet},
		"Senden klemmt":   {&probeleitung{sendefehler: kaputt}, nil, domain.ErrEinloesungUnbeantwortet},
		"Leeren klemmt":   {&probeleitung{leerefehler: kaputt}, nil, domain.ErrEinloesungUnbeantwortet},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, _, _, code, _ := annaStelltAus(t)
			werk, bote, _ := bertMeldetSich(t, fall.leitung, fall.wahl)

			if fehler := werk.Melden(t.Context(), bote, code, domain.Einloesung{}); !errors.Is(fehler, fall.klasse) {
				t.Errorf("Befund: %v", fehler)
			}
		})
	}
}

// **Eine Einladung, die nicht einlösbar ist, wählt niemanden an** — und
// eine, die nicht gilt, auch nicht.
func TestOhneEinloesbarkeitWirdNiemandGewaehlt(t *testing.T) {
	t.Parallel()

	anna := mussIdentitaet(t)
	annaWerk := service.NeuesEinladungswerk(festerAusweis{ich: anna}, festeRotation{},
		neueEinladungsablage(), func() time.Time { return einladungsuhr })

	ohne, fehler := annaWerk.Erstellen(t.Context(), service.Einladungsbitte{Name: "Anna", Zweck: domain.ZweckPaarung})
	if fehler != nil {
		t.Fatal(fehler)
	}

	werk, bote, gewaehlt := bertMeldetSich(t, &probeleitung{}, nil)

	if fehler := werk.Melden(t.Context(), bote, ohne.Code, domain.Einloesung{}); !errors.Is(fehler, domain.ErrEinladungNichtEinloesbar) {
		t.Errorf("nicht einlösbar: %v", fehler)
	}

	if fehler := werk.Melden(t.Context(), bote, "kein Code", domain.Einloesung{}); !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("kein Code: %v", fehler)
	}

	// **Und eine eigene Adresse, die keine ist**, geht nicht hinaus.
	_, _, _, einloesbar, _ := annaStelltAus(t)
	if fehler := werk.Melden(t.Context(), bote, einloesbar,
		domain.Einloesung{Adresse: "keine Adresse"}); !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("krumme eigene Adresse: %v", fehler)
	}

	if len(*gewaehlt) != 0 {
		t.Errorf("gewählt: %v", *gewaehlt)
	}
}
