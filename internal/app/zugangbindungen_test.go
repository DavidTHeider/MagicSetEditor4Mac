package app_test

// **Die Zugangsbindungen** — Werkstück 8 von Phase 4.8
// (#1321).
//
// **Kein Lauf hier fasst den echten Anmeldebund an** (§3): jeder setzt einen
// eigenen Speicher ein. Der Weg zum echten steht in `bundOderAnmeldebund` und
// wird hier nur daraufhin geprüft, dass er ohne gesetzten Bund gewählt wird.

import (
	"context"
	"slices"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// bundspeicher merkt sich, was abgelegt wurde.
type bundspeicher struct {
	werte  map[string]string
	fehler error
}

func neuerBundspeicher() *bundspeicher {
	return &bundspeicher{werte: map[string]string{}}
}

func (b *bundspeicher) Lies(_ context.Context, _, konto string) (string, error) {
	if b.fehler != nil {
		return "", b.fehler
	}

	wert, gibts := b.werte[konto]
	if !gibts {
		return "", domain.ErrNotFound
	}

	return wert, nil
}

func (b *bundspeicher) Schreibe(_ context.Context, _, konto, geheimnis string) error {
	if b.fehler != nil {
		return b.fehler
	}

	b.werte[konto] = geheimnis

	return nil
}

func (b *bundspeicher) Loesche(_ context.Context, _, konto string) error {
	if b.fehler != nil {
		return b.fehler
	}

	delete(b.werte, konto)

	return nil
}

// **Der Stand sagt, welche Angaben ein Dienst braucht und welche liegen.**
//
// **Und er sagt nie, wie sie lauten.** Ein Weg, der den Schlüssel ins Fenster
// zurückgäbe, machte aus dem Schlüsselbund eine Durchreiche.
func TestDerZugangsstandNenntNamenUndKeineWerte(t *testing.T) {
	t.Parallel()

	speicher := neuerBundspeicher()
	speicher.werte["scrydex.schluessel"] = "streng geheim"

	anwendung := app.NeueApp("")
	anwendung.MitSchluesselbund(speicher)

	antwort := anwendung.Zugaenge(t.Context())
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if len(antwort.Zugaenge) != len(domain.Dienste()) {
		t.Fatalf("%d Stände, erwartet %d",
			len(antwort.Zugaenge), len(domain.Dienste()))
	}

	for _, stand := range antwort.Zugaenge {
		switch stand.Dienst {
		case "scryfall", "mtgapi":
			if len(stand.Angaben) != 0 || !stand.Vollstaendig {
				t.Errorf("%s verlangt %v", stand.Dienst, stand.Angaben)
			}

		case "scrydex":
			if !slices.Equal(stand.Hinterlegt, []string{"schluessel"}) {
				t.Errorf("scrydex hat %v hinterlegt", stand.Hinterlegt)
			}

			if stand.Vollstaendig {
				t.Error("scrydex gilt als vollständig, obwohl das Team fehlt")
			}

		case "mtgjson":
			if len(stand.Hinterlegt) != 0 || stand.Vollstaendig {
				t.Errorf("mtgjson: %+v", stand)
			}
		}
	}

	// **Und nirgends steht der Wert.**
	if strings.Contains(erstFormatiert(antwort), "streng geheim") {
		t.Error("die Antwort trägt das Geheimnis")
	}
}

// erstFormatiert macht die Antwort zu einem Text — für die Suche darin.
func erstFormatiert(antwort app.ZugaengeAntwort) string {
	var gebaut strings.Builder

	for _, stand := range antwort.Zugaenge {
		gebaut.WriteString(stand.Dienst)
		gebaut.WriteString(strings.Join(stand.Angaben, ","))
		gebaut.WriteString(strings.Join(stand.Hinterlegt, ","))
	}

	return gebaut.String()
}

// **Ein Zugang lässt sich eintragen und wieder herausnehmen.**
func TestEinZugangLaesstSichEintragenUndLoeschen(t *testing.T) {
	t.Parallel()

	speicher := neuerBundspeicher()

	anwendung := app.NeueApp("")
	anwendung.MitSchluesselbund(speicher)

	if antwort := anwendung.ZugangSetzen(t.Context(),
		"scrydex", "team", "mannschaft"); antwort.Fehler != nil {
		t.Fatalf("setzen: %+v", antwort.Fehler)
	}

	if speicher.werte["scrydex.team"] != "mannschaft" {
		t.Errorf("abgelegt ist %+v", speicher.werte)
	}

	if antwort := anwendung.ZugangLoeschen(t.Context(),
		"scrydex", "team"); antwort.Fehler != nil {
		t.Fatalf("löschen: %+v", antwort.Fehler)
	}

	if _, liegt := speicher.werte["scrydex.team"]; liegt {
		t.Error("nach dem Löschen liegt der Zugang noch")
	}
}

// **Ein Dienst ohne Zugangsdaten wird abgewiesen.**
//
// „Scryfall braucht keinen Schlüssel" ist eine andere Auskunft als „diesen
// Dienst gibt es nicht" — und wer eine Angabe für einen Dienst ablegte, der
// sie nie liest, bekäme ein Geheimnis im Bund, das nichts tut (§8).
func TestEinDienstOhneZugangsdatenWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	anwendung := app.NeueApp("")
	anwendung.MitSchluesselbund(neuerBundspeicher())

	for _, fall := range []struct{ name, dienst, angabe, nennt string }{
		{"Scryfall braucht keine", "scryfall", "marke", "Zugangsdaten"},
		{"gibt es nicht", "gibtsnicht", "marke", "Zugangsdaten"},
		{"falsche Angabe", "scrydex", "passwort", "passwort"},
	} {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			antwort := anwendung.ZugangSetzen(
				probekontext(t), fall.dienst, fall.angabe, "egal")

			if antwort.Fehler == nil {
				t.Fatal("angenommen, erwartet abgewiesen")
			}

			if !strings.Contains(antwort.Fehler.Meldung, fall.nennt) {
				t.Errorf("die Meldung %q nennt %q nicht",
					antwort.Fehler.Meldung, fall.nennt)
			}

			// **Und derselbe Weg beim Löschen** — sonst könnte jemand eine
			// Angabe herausnehmen, die er nie ablegen durfte.
			if weg := anwendung.ZugangLoeschen(
				probekontext(t), fall.dienst, fall.angabe); weg.Fehler == nil {
				t.Error("das Löschen glückte, erwartet abgewiesen")
			}
		})
	}
}

// **Ein kaputter Schlüsselbund kommt als Fehler an.**
func TestEinKaputterBundKommtAn(t *testing.T) {
	t.Parallel()

	speicher := neuerBundspeicher()
	speicher.fehler = domain.ErrConflict

	anwendung := app.NeueApp("")
	anwendung.MitSchluesselbund(speicher)

	if antwort := anwendung.Zugaenge(t.Context()); antwort.Fehler == nil {
		t.Error("der Stand kam ohne Fehler")
	}

	if antwort := anwendung.ZugangSetzen(
		probekontext(t), "scrydex", "team", "x"); antwort.Fehler == nil {
		t.Error("das Setzen glückte")
	}

	if antwort := anwendung.ZugangLoeschen(
		probekontext(t), "scrydex", "team"); antwort.Fehler == nil {
		t.Error("das Löschen glückte")
	}
}

// **Ohne gesetzten Bund wird der Anmeldebund genommen.**
//
// **Er wird nicht angefasst** — geprüft wird, dass es ihn gibt, nicht was
// darin steht. Ein Lauf, der in den echten Bund schriebe, wäre keiner (§3).
func TestOhneGesetztenBundKommtDerAnmeldebund(t *testing.T) {
	t.Parallel()

	if !app.NeueApp("").NimmtDenAnmeldebund() {
		t.Error("ohne gesetzten Bund kommt nicht der Anmeldebund")
	}

	anwendung := app.NeueApp("")
	anwendung.MitSchluesselbund(neuerBundspeicher())

	if anwendung.NimmtDenAnmeldebund() {
		t.Error("mit gesetztem Bund kommt trotzdem der Anmeldebund")
	}
}
