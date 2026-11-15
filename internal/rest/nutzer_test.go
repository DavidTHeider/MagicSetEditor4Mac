// **Nutzer und Profile über den Dienst** — Werkstück 1 von Phase 4.12
// (#1351).
package rest_test

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

// **Eine leere Nutzerliste kommt als `[]` und nicht als `null`** (§3).
//
// # Warum das eine eigene Prüfung ist
//
// `json:",omitempty"` und ein `nil`-Slice ergeben zusammen eine Antwort ohne
// das Feld — und die Seite bliebe leer, statt „noch niemand" zu zeigen. Das
// ist die Naht zwischen den Prüfebenen, vor der §3 ausdrücklich warnt: Go-Tests
// sehen die Marken nicht, Frontend-Tests gehen nicht durch die Serialisierung.
func TestEineLeereNutzerlisteKommtAlsListe(t *testing.T) {
	jetzt := time.Date(2026, time.September, 4, 12, 0, 0, 0, time.UTC)

	dienst := rest.NeuerDienst(probemarke, func() time.Time { return jetzt },
		rest.Naehte{Nutzer: leerernutzer{}})

	antwort := ruf(t, dienst, endpunkt{
		name: "Nutzer lesen", methode: http.MethodGet, weg: "/nutzer",
		erfolg: http.StatusOK,
	}, gueltig)

	if strings.Contains(antwort.Body.String(), `"nutzer":null`) {
		t.Errorf("die Liste ist null: %s", antwort.Body.String())
	}
}

// leerernutzer gibt gar keine Nutzer — und zwar als `nil`.
type leerernutzer struct{ williger }

func (leerernutzer) Nutzer(context.Context) ([]rest.Nutzer, error) { return nil, nil }

func (leerernutzer) GezeigtesProfil(context.Context) (rest.Profil, error) {
	return rest.Profil{}, nil
}

func (leerernutzer) ProfilWechseln(context.Context, string, string) (rest.Profil, error) {
	return rest.Profil{}, nil
}

func (leerernutzer) Anbieterkonten(context.Context, string) ([]rest.Konto, error) {
	return nil, nil
}

func (leerernutzer) KontoLoesen(context.Context, string, string) error { return nil }

// **Ein krummer Rumpf an einem der beiden Wechsel wird abgewiesen.**
//
// **Beide, seit Werkstück 17 von Phase 4.12** (#1400): sie sehen gleich aus
// und meinen Verschiedenes — der eine stellt das Profil um, der andere den
// Menschen. Ein Lauf über nur einen ließe den zweiten Rumpfleser ungeprüft.
func TestEinKrummerRumpfAmWechselWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	for name, weg := range map[string]string{
		"Profil wechseln": "/profil",
		"Nutzer wechseln": "/nutzer",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			antwort := ruf(t, neuerDienst(t), endpunkt{
				name: name, methode: http.MethodPut, weg: weg, rumpf: `{`,
			}, gueltig)

			if antwort.Code != http.StatusBadRequest {
				t.Errorf("endete mit %d statt 400: %s",
					antwort.Code, antwort.Body.String())
			}
		})
	}
}

// **Eine leere Kontenliste ist `[]`, nicht `null`** (§3).
func TestEineLeereKontenlisteIstKeineNull(t *testing.T) {
	t.Parallel()

	jetzt := time.Date(2026, time.September, 5, 12, 0, 0, 0, time.UTC)

	dienst := rest.NeuerDienst(probemarke, func() time.Time { return jetzt },
		rest.Naehte{Nutzer: leerernutzer{}})

	antwort := ruf(t, dienst, endpunkt{
		name: "Konten lesen", methode: http.MethodGet, weg: "/nutzer/1/konten",
		erfolg: http.StatusOK,
	}, gueltig)

	if strings.Contains(antwort.Body.String(), "null") {
		t.Errorf("die Liste ist null: %s", antwort.Body.String())
	}
}

// **Ein krummer Rumpf an den beiden Änderungswegen wird abgewiesen.**
func TestEinKrummerRumpfInDerPflegeWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	faelle := []endpunkt{
		{name: "Nutzer ändern", methode: http.MethodPut, weg: "/nutzer/1", rumpf: `{`},
		{name: "Profil ändern", methode: http.MethodPut, weg: "/profile/1", rumpf: `{`},

		// **Und an den beiden Kontowegen** — Werkstück 6 (#1351). Sie tragen
		// ein **Kennwort** im Rumpf; wer den nicht liest, ruft die Anwendung
		// mit zwei leeren Zeichenketten und bekommt eine Abweisung, die vom
		// falschen Feld spricht.
		{name: "Konto eintragen", methode: http.MethodPost, weg: "/konten", rumpf: `{`},
		{
			name: "Konto anmelden", methode: http.MethodPost,
			weg: "/konten/anmeldung", rumpf: `{`,
		},

		// **Und an den beiden Faktorwegen** — Werkstück 7 (#1351).
		{
			name: "Faktor einrichten", methode: http.MethodPost,
			weg: "/nutzer/1/faktoren", rumpf: `{`,
		},
		{
			name: "Faktor bestätigen", methode: http.MethodPut,
			weg: "/nutzer/1/faktoren/totp", rumpf: `{`,
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			antwort := ruf(t, neuerDienst(t), fall, gueltig)
			if antwort.Code != http.StatusBadRequest {
				t.Errorf("endete mit %d statt 400: %s", antwort.Code, antwort.Body.String())
			}
		})
	}
}

// **Eine leere Faktorliste ist eine Liste, kein `null`** (§3).
//
// **Und die Codes ebenso**: `json:",omitempty"` gibt es hier nicht, aber ein
// `nil`-Slice wird trotzdem zu `null` — und die Seite prüfte dann auf zwei
// Dinge statt auf eines.
func TestEineLeereFaktorlisteIstEineListe(t *testing.T) {
	t.Parallel()

	jetzt := time.Date(2026, time.September, 5, 12, 0, 0, 0, time.UTC)

	dienst := rest.NeuerDienst(probemarke, func() time.Time { return jetzt },
		rest.Naehte{Nutzer: leerefaktoren{}})

	for _, fall := range []endpunkt{
		{
			name: "Faktoren lesen", methode: http.MethodGet,
			weg: "/nutzer/1/faktoren", erfolg: http.StatusOK,
		},
		{
			name: "Faktor einrichten", methode: http.MethodPost,
			weg: "/nutzer/1/faktoren", rumpf: `{"art":"totp"}`,
			erfolg: http.StatusCreated,
		},
	} {
		antwort := ruf(t, dienst, fall, gueltig)
		if strings.Contains(antwort.Body.String(), "null") {
			t.Errorf("%s ist null: %s", fall.name, antwort.Body.String())
		}
	}
}

// leerefaktoren gibt gar nichts — und zwar als `nil`.
type leerefaktoren struct{ williger }

func (leerefaktoren) Faktoren(context.Context, string) (rest.Faktorstand, error) {
	return rest.Faktorstand{}, nil
}

func (leerefaktoren) FaktorEinrichten(context.Context, string) (
	rest.Faktoreinrichtung, error,
) {
	return rest.Faktoreinrichtung{Geheimnis: "ABC"}, nil
}
