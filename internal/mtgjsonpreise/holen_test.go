package mtgjsonpreise_test

// **Der Bulkbezug gegen eine eigene Gegenstelle** — Werkstück A3 von Phase 5.3
// (#1442).
//
// **Nie gegen mtgjson.com** (§3, §16): die Datei wiegt ein Gigabyte, und ein
// Lauf, der sie holte, prüfte das Netz statt den Code. Was hier steht, ist
// eine eigene Gegenstelle mit selbst geschriebenen Gestalten.

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mtgjsonpreise"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
)

// gegenstelle antwortet auf die beiden Dateinamen.
func gegenstelle(t *testing.T, antwort func(http.ResponseWriter, *http.Request)) string {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(antwort))
	t.Cleanup(server.Close)

	return server.URL
}

// klientGegen baut einen Klienten gegen eine Adresse.
func klientGegen(adresse string, grenzen mtgjsonpreise.Grenzen) *mtgjsonpreise.Klient {
	return mtgjsonpreise.NeuerKlient(adresse,
		netz.NeuerKlient(netz.UeblicheFristen()),
		netz.Kenner{Name: "MagicSetEditor4Mac/prüfung", Annahme: "*/*"},
		grenzen)
}

// **Der heutige Stand wird geholt und gelesen.**
func TestDerHeutigeStandWirdGeholt(t *testing.T) {
	t.Parallel()

	var gefragt string

	adresse := gegenstelle(t, func(schreiben http.ResponseWriter, anfrage *http.Request) {
		gefragt = anfrage.URL.Path

		schreiben.Header().Set("Content-Type", "application/gzip")
		_, _ = schreiben.Write(gepackt(t)) //nolint:errcheck // die Probe
	})

	staende, befunde, fehler := klientGegen(adresse, mtgjsonpreise.Grenzen{}).
		Holen(t.Context(), string(mtgjsonpreise.UmfangHeute), gewollt())
	if fehler != nil {
		t.Fatalf("Holen: %v", fehler)
	}

	if !strings.HasSuffix(gefragt, "/AllPricesToday.json.gz") {
		t.Errorf("gefragt wurde %q", gefragt)
	}

	if len(staende) != 6 || len(befunde) != 0 {
		t.Errorf("%d Stände, %d Befunde", len(staende), len(befunde))
	}
}

// **Und der Neunzigtagestand von seiner eigenen Datei.**
//
// **Ohne diesen Fall stünde der Lauf darüber auch dann grün da, wenn beide
// Umfänge dieselbe Datei holten** (§10) — und dann wäre „neunzig Tage" eine
// Angabe ohne Wirkung.
func TestDerNeunzigtagestandHatSeineEigeneDatei(t *testing.T) {
	t.Parallel()

	var gefragt string

	adresse := gegenstelle(t, func(schreiben http.ResponseWriter, anfrage *http.Request) {
		gefragt = anfrage.URL.Path
		_, _ = schreiben.Write(gepackt(t)) //nolint:errcheck // die Probe
	})

	if _, _, fehler := klientGegen(adresse, mtgjsonpreise.Grenzen{}).
		Holen(t.Context(), string(mtgjsonpreise.UmfangNeunzigTage), gewollt()); fehler != nil {
		t.Fatalf("Holen: %v", fehler)
	}

	if !strings.HasSuffix(gefragt, "/AllPrices.json.gz") {
		t.Errorf("gefragt wurde %q", gefragt)
	}
}

// **Die Anwendung nennt sich** (§16, Satz 2).
func TestDerBezugNenntDieAnwendung(t *testing.T) {
	t.Parallel()

	var kenner string

	adresse := gegenstelle(t, func(schreiben http.ResponseWriter, anfrage *http.Request) {
		kenner = anfrage.Header.Get("User-Agent")
		_, _ = schreiben.Write(gepackt(t)) //nolint:errcheck // die Probe
	})

	if _, _, fehler := klientGegen(adresse, mtgjsonpreise.Grenzen{}).
		Holen(t.Context(), string(mtgjsonpreise.UmfangHeute), gewollt()); fehler != nil {
		t.Fatalf("Holen: %v", fehler)
	}

	if !strings.Contains(kenner, "MagicSetEditor4Mac") {
		t.Errorf("die Anwendung nennt sich %q", kenner)
	}
}

// **Ein Umfang, den es nicht gibt, geht gar nicht erst ins Netz.**
func TestEinUnbekannterUmfangGehtNichtInsNetz(t *testing.T) {
	t.Parallel()

	var gefragt bool

	adresse := gegenstelle(t, func(http.ResponseWriter, *http.Request) { gefragt = true })

	_, _, fehler := klientGegen(adresse, mtgjsonpreise.Grenzen{}).
		Holen(t.Context(), "voriges Jahr", gewollt())

	if !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("abgewiesen wurde mit %v", fehler)
	}

	if gefragt {
		t.Error("es wurde trotzdem gefragt")
	}
}

// **Was die Gegenstelle ablehnt, kommt als Fehler zurück — ohne ihren Rumpf.**
//
// **Eine Fehlerseite kann ein Megabyte HTML sein**, und in einer Meldung steht
// sie schlecht. Was der Mensch braucht, ist die Zahl.
func TestEineAbgelehnteAntwortWirdGemeldet(t *testing.T) {
	t.Parallel()

	adresse := gegenstelle(t, func(schreiben http.ResponseWriter, _ *http.Request) {
		schreiben.WriteHeader(http.StatusNotFound)
		_, _ = schreiben.Write([]byte(strings.Repeat("<html>weg</html>", 1000))) //nolint:errcheck // die Probe
	})

	_, _, fehler := klientGegen(adresse, mtgjsonpreise.Grenzen{}).
		Holen(t.Context(), string(mtgjsonpreise.UmfangHeute), gewollt())

	if !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("abgewiesen wurde mit %v", fehler)
	}

	if !strings.Contains(fehler.Error(), "404") ||
		strings.Contains(fehler.Error(), "<html>") {
		t.Errorf("die Meldung ist %q", fehler.Error())
	}
}

// **Ein Abbruch bricht ab** (§16, Satz 3).
func TestEinAbbruchHaeltDenBezugAn(t *testing.T) {
	t.Parallel()

	adresse := gegenstelle(t, func(schreiben http.ResponseWriter, _ *http.Request) {
		_, _ = schreiben.Write(gepackt(t)) //nolint:errcheck // die Probe
	})

	abgebrochen, abbrechen := context.WithCancel(t.Context())
	abbrechen()

	if _, _, fehler := klientGegen(adresse, mtgjsonpreise.Grenzen{}).
		Holen(abgebrochen, string(mtgjsonpreise.UmfangHeute), gewollt()); fehler == nil {
		t.Error("der abgebrochene Ruf kam durch")
	}
}

// **Der Deckel gilt auch für das, was aus dem Netz kommt.**
//
// **Und das ist die Stelle, an der er zählt**: ein fremder Server kann
// schicken, was er will, und eine Grenze, die nur auf einer Datei von der
// Platte greift, schützt vor nichts.
func TestDerDeckelGiltAuchAmNetz(t *testing.T) {
	t.Parallel()

	adresse := gegenstelle(t, func(schreiben http.ResponseWriter, _ *http.Request) {
		_, _ = schreiben.Write(gepackt(t)) //nolint:errcheck // die Probe
	})

	_, _, fehler := klientGegen(adresse, mtgjsonpreise.Grenzen{HoechstensBytes: 16}).
		Holen(t.Context(), string(mtgjsonpreise.UmfangHeute), gewollt())

	if !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("abgewiesen wurde mit %v", fehler)
	}
}

// **Die Quelle sagt, wer sie ist.**
func TestDerBulkklientNenntSeineQuelle(t *testing.T) {
	t.Parallel()

	if name := klientGegen("", mtgjsonpreise.Grenzen{}).Name(); name != "mtgjson" {
		t.Errorf("der Klient nennt sich %q", string(name))
	}
}

// **Eine Adresse, die keine ist, geht nicht ins Netz.**
func TestEineKrummeAdresseWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	_, _, fehler := klientGegen("://kein-schema", mtgjsonpreise.Grenzen{}).
		Holen(t.Context(), string(mtgjsonpreise.UmfangHeute), gewollt())

	if fehler == nil {
		t.Error("die krumme Adresse kam durch")
	}
}
