// **Die Prüfsummen tragen eine Signatur, und sie wird geprüft**
// (#892,
// #1100).
//
// # Was ohne sie fehlte
//
// Der Ladeweg holte `pruefsummen.txt`, nahm die Zeile des Abbilds und verglich
// sie mit dem Geladenen. **Das prüft, ob die beiden Dateien zueinander
// passen** — nicht, woher sie kommen. Eine Gegenstelle, die beide Antworten
// beherrscht, liefert ein eigenes Abbild samt eigener Prüfsumme, und der
// Vergleich geht auf.
//
// `docs/RELEASE.md` und `README.md` sagen das seit dem 2026-08-25 selbst: die
// Signatur ist genau das, was den gemeinsamen Austausch verhindert. Sie lag
// jeder Freigabe bei, und niemand sah sie an — angestrichen vom Gesamtreview
// vor 0.2.0.
package aktualisierung_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/aktualisierung"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// signiert gibt eine abgetrennte Signatur über den Text und setzt ihren
// Schlüssel als Bund ein.
//
// **Sie wird wirklich erzeugt**, nicht nachgebildet: eine erfundene Signatur
// prüfte den Aufbau einer Zeichenkette und nicht die Prüfung.
//
// **Mit einem Probeschlüssel** (SEC03, #2356). Hier signierte `gpg` mit dem
// Freigabeschlüssel aus dem Schlüsselbund dieses Rechners — in jedem
// `make ci` auf dem Rechner, der auch freigibt. Dass eine Signatur aus `gpg`
// in dieser Prüfung besteht, hält `freigabesignatur_test.go` gegen die
// veröffentlichte Signatur einer echten Freigabe, ohne privaten Schlüssel.
func signiert(t *testing.T, text string) string {
	t.Helper()

	return probeschluessel(t)(text)
}

// signaturstelle gibt eine Gegenstelle mit Abbild, Prüfsummen und Signatur.
func signaturstelle(
	t *testing.T, inhalt []byte, pruefsummen, signatur string,
) *httptest.Server {
	t.Helper()

	dienst := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, anfrage *http.Request) {
			switch {
			case strings.HasSuffix(anfrage.URL.Path, "/signatur"):
				if signatur == "" {
					schreiben.WriteHeader(http.StatusNotFound)

					return
				}

				if _, fehler := schreiben.Write([]byte(signatur)); fehler != nil {
					t.Errorf("schreiben: %v", fehler)
				}
			case strings.HasSuffix(anfrage.URL.Path, "/pruefsummen"):
				if _, fehler := schreiben.Write([]byte(pruefsummen)); fehler != nil {
					t.Errorf("schreiben: %v", fehler)
				}
			default:
				schreiben.Header().Set("Content-Length", strconv.Itoa(len(inhalt)))

				if _, fehler := schreiben.Write(inhalt); fehler != nil {
					t.Logf("der Klient ging weg: %v", fehler)
				}
			}
		}))

	t.Cleanup(dienst.Close)

	return dienst
}

// mitSignatur baut die Freigabe zu einer Gegenstelle.
func mitSignatur(dienst *httptest.Server, mitAsc bool) aktualisierung.Freigabe {
	freigabe := aktualisierung.Freigabe{
		Fassung:     "0.2.0",
		Abbild:      beigabe("MagicSetEditor4Mac-0.2.0.dmg", dienst.URL+"/abbild"),
		Pruefsummen: beigabe("pruefsummen.txt", dienst.URL+"/pruefsummen"),
	}

	if mitAsc {
		freigabe.Signatur = beigabe("pruefsummen.txt.asc", dienst.URL+"/signatur")
	}

	return freigabe
}

// **Eine gültige Signatur lässt das Laden durch.**
func TestEineEchteSignaturLaesstDasLadenDurch(t *testing.T) {
	inhalt := abbild()
	summen := summe(inhalt) + "  ./MagicSetEditor4Mac-0.2.0.dmg\n"

	dienst := signaturstelle(t, inhalt, summen, signiert(t, summen))

	pfad, err := probeklient(t, dienst.URL).Laden(t.Context(),
		mitSignatur(dienst, true), t.TempDir(), nil)
	if err != nil {
		t.Fatalf("Laden: %v", err)
	}

	if pfad == "" {
		t.Error("es kam kein Pfad zurück")
	}
}

// **Eine Signatur über einen anderen Text wird abgewiesen.**
//
// **Das ist der Angriff, den sie verhindert**: eine Gegenstelle, die Abbild
// **und** Prüfsummen beherrscht, kann beide zueinander passend machen — die
// Signatur darüber kann sie nicht.
func TestEineFremdeSignaturWirdAbgewiesen(t *testing.T) {
	inhalt := abbild()
	summen := summe(inhalt) + "  ./MagicSetEditor4Mac-0.2.0.dmg\n"

	// **Signiert wird ein anderer Text** — genau das, was ein Austausch ergibt.
	dienst, gefragt := zaehlendeStelle(t, inhalt, summen,
		signiert(t, "eine ganz andere Zeile\n"))

	_, err := probeklient(t, dienst.URL).Laden(t.Context(),
		mitSignatur(dienst, true), t.TempDir(), nil)

	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("der Fehler ist %v — will ErrConflict", err)
	}

	// **Und das Abbild ist gar nicht erst gefragt worden.**
	//
	// Das ist die Zusage aus [0051](../../docs/entscheidungen/0051-die-signatur-wird-geprueft.md):
	// erst die Signatur, dann die Deutung, dann der Abruf. Ohne diese Zeile
	// wäre der Test auch grün, wenn 224 MB geladen und danach verworfen
	// würden — und die Reihenfolge stünde nur im Vermerk.
	if gefragt() != 0 {
		t.Errorf("das Abbild wurde %d-mal gefragt, obwohl die Signatur nicht passt",
			gefragt())
	}
}

// **Eine Freigabe ohne Signatur wird abgewiesen — vor dem Laden.**
func TestEineFreigabeOhneSignaturWirdAbgewiesen(t *testing.T) {
	inhalt := abbild()
	summen := summe(inhalt) + "  ./MagicSetEditor4Mac-0.2.0.dmg\n"

	dienst := signaturstelle(t, inhalt, summen, "")

	_, err := probeklient(t, dienst.URL).Laden(t.Context(),
		mitSignatur(dienst, false), t.TempDir(), nil)

	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("der Fehler ist %v — will ErrNotFound", err)
	}
}

// **Und eine Signatur, die sich nicht lesen lässt, ebenso.**
func TestEineUnlesbareSignaturWirdAbgewiesen(t *testing.T) {
	inhalt := abbild()
	summen := summe(inhalt) + "  ./MagicSetEditor4Mac-0.2.0.dmg\n"

	dienst := signaturstelle(t, inhalt, summen, "das ist keine Signatur\n")

	_, err := probeklient(t, dienst.URL).Laden(t.Context(),
		mitSignatur(dienst, true), t.TempDir(), nil)

	if !errors.Is(err, domain.ErrConflict) {
		t.Errorf("der Fehler ist %v — will ErrConflict", err)
	}
}

// **Der eingebettete Schlüssel ist der Freigabeschlüssel** — seit 1.0.0 ein
// eigener, getrennt vom Commitschlüssel (Vermerke 0194 und 0205).
//
// **Sonst prüfte diese Anwendung eine Signatur gegen einen Schlüssel, den
// niemand kennt** — und das wäre schlimmer als keine Prüfung: es sähe aus wie
// eine.
func TestDerEingebetteteSchluesselIstDerRichtige(t *testing.T) {
	t.Parallel()

	abdruck, err := aktualisierung.Schluesselfingerabdruck()
	if err != nil {
		t.Fatalf("den Schlüssel lesen: %v", err)
	}

	if abdruck != "97721629ABDC00BA1711488244C59428DE1F766E" {
		t.Errorf("der eingebettete Schlüssel ist %s", abdruck)
	}
}

// **Eine Signatur, die gar nicht erreichbar ist, wird gemeldet.**
func TestEineUnerreichbareSignaturWirdGemeldet(t *testing.T) {
	inhalt := abbild()
	summen := summe(inhalt) + "  ./MagicSetEditor4Mac-0.2.0.dmg\n"

	dienst := signaturstelle(t, inhalt, summen, "")

	freigabe := mitSignatur(dienst, true)
	freigabe.Signatur = beigabe("pruefsummen.txt.asc", "http://127.0.0.1:1/signatur")

	if _, err := probeklient(t, dienst.URL).Laden(t.Context(),
		freigabe, t.TempDir(), nil); err == nil {
		t.Error("eine unerreichbare Signatur kam durch")
	}
}

// **Und eine, die mit einem Fehlschlag kommt, ebenso.**
func TestEineSignaturMitFehlschlagWirdGemeldet(t *testing.T) {
	inhalt := abbild()
	summen := summe(inhalt) + "  ./MagicSetEditor4Mac-0.2.0.dmg\n"

	// **Leere Signatur heißt hier 404** — so baut `signaturstelle` sie.
	dienst := signaturstelle(t, inhalt, summen, "")

	_, err := probeklient(t, dienst.URL).Laden(t.Context(),
		mitSignatur(dienst, true), t.TempDir(), nil)

	if !errors.Is(err, domain.ErrConflict) {
		t.Errorf("der Fehler ist %v — will ErrConflict", err)
	}
}

// **Die Signatur einer Freigabe wird als Beigabe erkannt.**
func TestDieSignaturWirdAlsBeigabeErkannt(t *testing.T) {
	dienst := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			if _, fehler := schreiben.Write([]byte(
				`{"tag_name":"v9.9.9","assets":[` +
					`{"name":"x.dmg","url":"u/abbild"},` +
					`{"name":"pruefsummen.txt","url":"u/summen"},` +
					`{"name":"pruefsummen.txt.asc","url":"u/signatur"}]}`,
			)); fehler != nil {
				t.Errorf("schreiben: %v", fehler)
			}
		}))
	t.Cleanup(dienst.Close)

	freigabe, err := probeklient(t, dienst.URL).NeuesteFreigabe(t.Context())
	if err != nil {
		t.Fatalf("NeuesteFreigabe: %v", err)
	}

	if freigabe.Signatur.Adresse != "u/signatur" {
		t.Errorf("die Signatur ist %+v", freigabe.Signatur)
	}
}

// zaehlendeStelle ist [signaturstelle] und zählt die Abrufe des Abbilds.
func zaehlendeStelle(
	t *testing.T, inhalt []byte, pruefsummen, signatur string,
) (dienst *httptest.Server, abbildabrufe func() int) {
	t.Helper()

	var abrufe atomic.Int64

	dienst = httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, anfrage *http.Request) {
			switch {
			case strings.HasSuffix(anfrage.URL.Path, "/signatur"):
				if _, fehler := schreiben.Write([]byte(signatur)); fehler != nil {
					t.Errorf("schreiben: %v", fehler)
				}
			case strings.HasSuffix(anfrage.URL.Path, "/pruefsummen"):
				if _, fehler := schreiben.Write([]byte(pruefsummen)); fehler != nil {
					t.Errorf("schreiben: %v", fehler)
				}
			default:
				abrufe.Add(1)

				schreiben.Header().Set("Content-Length", strconv.Itoa(len(inhalt)))

				if _, fehler := schreiben.Write(inhalt); fehler != nil {
					t.Logf("der Klient ging weg: %v", fehler)
				}
			}
		}))

	t.Cleanup(dienst.Close)

	return dienst, func() int { return int(abrufe.Load()) }
}
