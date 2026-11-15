package aktualisierung_test

// **Das Manifest einer Freigabe** — Werkstück O6 von Phase 5.2
// (#1430).

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/aktualisierung"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
)

// echtesManifest ist die Gestalt, die `release.sh` schreibt.
const echtesManifest = `{
  "fassung": "0.5.0",
  "plattform": "darwin/arm64",
  "artefakt": "abc123",
  "einstellungen": 1,
  "protokoll": 1,
  "wanderungen": [
    {"name": "0001_anfang.sql", "sha256": "aaa"}
  ],
  "ein_feld_das_wir_nicht_kennen": 42
}`

func TestEinManifestWirdGelesenUndFremdeFelderUebergangen(t *testing.T) {
	t.Parallel()

	manifest, err := aktualisierung.ManifestAus([]byte(echtesManifest))
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	switch {
	case manifest.Fassung != "0.5.0":
		t.Errorf("Fassung = %q", manifest.Fassung)
	case manifest.Plattform != "darwin/arm64":
		t.Errorf("Plattform = %q", manifest.Plattform)
	case len(manifest.Wanderungen) != 1 || manifest.Wanderungen[0].Summe != "aaa":
		t.Errorf("Wanderungen = %+v", manifest.Wanderungen)
	}
}

// **Pflichtfelder streng** (§16): ohne Fassung oder Plattform ist es keine
// Aussage, sondern eine Lücke.
func TestEinManifestOhnePflichtfelderWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	for name, roh := range map[string]string{
		"kein JSON":      `{`,
		"ohne Fassung":   `{"plattform":"darwin/arm64"}`,
		"ohne Plattform": `{"fassung":"0.5.0"}`,
		"leeres Objekt":  `{}`,
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if _, err := aktualisierung.ManifestAus([]byte(roh)); !errors.Is(
				err, domain.ErrValidation) {
				t.Errorf("Fehler = %v", err)
			}
		})
	}
}

// manifestdienst gibt eine Gegenstelle, die das Manifest ausliefert.
func manifestdienst(t *testing.T, rumpf string) *httptest.Server {
	t.Helper()

	dienst := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			if _, err := schreiben.Write([]byte(rumpf)); err != nil {
				t.Errorf("schreiben: %v", err)
			}
		}))
	t.Cleanup(dienst.Close)

	return dienst
}

// summenliste schreibt eine Zeile im Format von `shasum -a 256` — über das
// **echte** Manifest, denn genau seine Summe steht in der signierten Liste.
func summenliste() string {
	summe := sha256.Sum256([]byte(echtesManifest))

	return hex.EncodeToString(summe[:]) + "  ./" + aktualisierung.Manifestname + "\n"
}

// **Der ganze Weg**: geholt, gegen die signierte Liste geprüft, gelesen.
func TestDasManifestWirdGegenDieSignierteListeGeprueft(t *testing.T) {
	t.Parallel()

	dienst := manifestdienst(t, echtesManifest)
	klient := aktualisierung.NeuerKlient(dienst.URL,
		netz.NeuerKlient(netz.UeblicheFristen()),
		netz.NeuesDrosselwerk(time.Now, func(time.Duration) {}), "")

	freigabe := aktualisierung.Freigabe{
		Fassung: "0.5.0",
		Manifest: aktualisierung.Beigabe{
			Name: aktualisierung.Manifestname, Adresse: dienst.URL,
		},
	}

	manifest, err := klient.Manifest(t.Context(), freigabe,
		[]byte(summenliste()))
	if err != nil {
		t.Fatalf("holen: %v", err)
	}

	if manifest.Fassung != "0.5.0" {
		t.Errorf("Fassung = %q", manifest.Fassung)
	}
}

// **Ein manipuliertes Manifest fällt auf** — die erste Abnahme aus dem Issue.
func TestEinManipuliertesManifestWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	dienst := manifestdienst(t, `{"fassung":"9.9.9","plattform":"darwin/arm64"}`)
	klient := aktualisierung.NeuerKlient(dienst.URL,
		netz.NeuerKlient(netz.UeblicheFristen()),
		netz.NeuesDrosselwerk(time.Now, func(time.Duration) {}), "")

	freigabe := aktualisierung.Freigabe{
		Manifest: aktualisierung.Beigabe{
			Name: aktualisierung.Manifestname, Adresse: dienst.URL,
		},
	}

	// **Die Liste nennt die Summe des echten Manifests**, geliefert wird ein
	// anderes.
	_, err := klient.Manifest(t.Context(), freigabe, []byte(summenliste()))
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("das manipulierte Manifest kam durch: %v", err)
	}
}

// **Und eines, das die signierte Liste gar nicht nennt.**
func TestEinManifestOhneEintragInDerListeWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	dienst := manifestdienst(t, echtesManifest)
	klient := aktualisierung.NeuerKlient(dienst.URL,
		netz.NeuerKlient(netz.UeblicheFristen()),
		netz.NeuesDrosselwerk(time.Now, func(time.Duration) {}), "")

	freigabe := aktualisierung.Freigabe{
		Manifest: aktualisierung.Beigabe{
			Name: aktualisierung.Manifestname, Adresse: dienst.URL,
		},
	}

	_, err := klient.Manifest(t.Context(), freigabe,
		[]byte("aaa  ./MagicSetEditor4Mac.dmg\n"))
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("Fehler = %v", err)
	}
}

// **Eine Freigabe ohne Manifest ist keine ungültige** — sie ist eine von
// vorher, und der Nullwert ist die richtige Antwort.
func TestEineFreigabeOhneManifestGibtDenNullwert(t *testing.T) {
	t.Parallel()

	klient := aktualisierung.NeuerKlient("http://127.0.0.1:1",
		netz.NeuerKlient(netz.UeblicheFristen()),
		netz.NeuesDrosselwerk(time.Now, func(time.Duration) {}), "")

	manifest, err := klient.Manifest(t.Context(), aktualisierung.Freigabe{}, nil)

	switch {
	case err != nil:
		t.Fatalf("ohne Manifest kam ein Fehler: %v", err)
	case manifest.Fassung != "":
		t.Errorf("Manifest = %+v", manifest)
	}
}

// **Eine Freigabe, die ein Manifest mitbringt, führt es auch** — sonst wäre
// die Zuordnung in `zuFreigabe` eine Zeile, die kein Lauf betritt.
func TestEineFreigabeFuehrtIhrManifestAlsBeigabe(t *testing.T) {
	t.Parallel()

	mitManifest := `{
  "tag_name": "v0.9.0",
  "assets": [
    {"name": "manifest.json", "size": 120,
     "url": "https://api.github.com/repos/x/y/releases/assets/9"}
  ]
}`

	dienst, _ := gegenstelle(t, mitManifest, http.StatusOK)

	freigabe, err := probeklient(t, dienst.URL).NeuesteFreigabe(t.Context())
	if err != nil {
		t.Fatalf("holen: %v", err)
	}

	if freigabe.Manifest.Name != aktualisierung.Manifestname {
		t.Errorf("Manifest = %+v", freigabe.Manifest)
	}
}

// **Und ein Manifest, das die Gegenstelle mit einem Fehlerstatus schickt**,
// ist kein Manifest.
func TestEinManifestMitFehlerstatusWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	dienst := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			schreiben.WriteHeader(http.StatusInternalServerError)
		}))
	t.Cleanup(dienst.Close)

	klient := aktualisierung.NeuerKlient(dienst.URL,
		netz.NeuerKlient(netz.UeblicheFristen()),
		netz.NeuesDrosselwerk(time.Now, func(time.Duration) {}), "")

	freigabe := aktualisierung.Freigabe{
		Manifest: aktualisierung.Beigabe{
			Name: aktualisierung.Manifestname, Adresse: dienst.URL,
		},
	}

	if _, err := klient.Manifest(t.Context(), freigabe,
		[]byte(summenliste())); !errors.Is(err, domain.ErrConflict) {
		t.Errorf("Fehler = %v", err)
	}
}

// **Und eine Adresse, die es nicht gibt** — der Netzfehler kommt an.
func TestEinManifestVonEinerTotenAdresseMeldetSich(t *testing.T) {
	t.Parallel()

	klient := aktualisierung.NeuerKlient("http://127.0.0.1:1",
		netz.NeuerKlient(netz.UeblicheFristen()),
		netz.NeuesDrosselwerk(time.Now, func(time.Duration) {}), "")

	freigabe := aktualisierung.Freigabe{
		Manifest: aktualisierung.Beigabe{
			Name: aktualisierung.Manifestname, Adresse: "http://127.0.0.1:1/manifest.json",
		},
	}

	if _, err := klient.Manifest(t.Context(), freigabe,
		[]byte(summenliste())); err == nil {
		t.Error("die tote Adresse kam durch")
	}
}

// **Ein Manifest, das ein anderes Abbild nennt, wird abgewiesen.**
//
// Befund 6 des Zweitblicks über Phase 5.2: der Artefakthash stand im Manifest,
// und niemand verglich ihn. Ein Manifest, das zu einer **anderen** Freigabe
// gehört, sähe damit aus wie das dieser — obwohl die Abnahme von O6 sagt:
// „manipuliertes Manifest/Artefakt … werden vor Bestandstausch abgewiesen".
func TestEinManifestMitFremdemAbbildWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	dienst := manifestdienst(t, echtesManifest)
	klient := aktualisierung.NeuerKlient(dienst.URL,
		netz.NeuerKlient(netz.UeblicheFristen()),
		netz.NeuesDrosselwerk(time.Now, func(time.Duration) {}), "")

	freigabe := aktualisierung.Freigabe{
		Fassung: "0.5.0",
		Manifest: aktualisierung.Beigabe{
			Name: aktualisierung.Manifestname, Adresse: dienst.URL,
		},
		Abbild: aktualisierung.Beigabe{Name: "MagicSetEditor4Mac-0.5.0.dmg"},
	}

	// **Die Liste nennt eine andere Summe für das Abbild** als das Manifest
	// (`"artefakt": "abc123"`).
	liste := summenliste() + "deadbeef  ./MagicSetEditor4Mac-0.5.0.dmg\n"

	if _, err := klient.Manifest(t.Context(), freigabe, []byte(liste)); !errors.Is(
		err, domain.ErrConflict) {
		t.Fatalf("das fremde Abbild kam durch: %v", err)
	}

	// **Die Gegenprobe**: nennt die Liste dieselbe Summe, geht es durch.
	passend := summenliste() + "abc123  ./MagicSetEditor4Mac-0.5.0.dmg\n"

	if _, err := klient.Manifest(t.Context(), freigabe, []byte(passend)); err != nil {
		t.Errorf("mit passendem Abbild: %v", err)
	}
}

// **Ein Manifest, das die Prüfsumme hält und trotzdem keines ist.**
//
// Die Summe sagt, dass niemand es unterwegs geändert hat — nicht, dass etwas
// Sinnvolles darin steht. Dieser Weg entstand mit Befund 6 des Zweitblicks
// über Phase 5.2: seither wird nach dem Lesen noch etwas verglichen, und das
// Lesen kann scheitern.
func TestEinUnlesbaresManifestHinterDerPruefsummeMeldetSich(t *testing.T) {
	t.Parallel()

	// **Gültiges JSON, aber ohne Plattform** — `ManifestAus` weist es ab.
	ohnePlattform := `{"fassung":"0.5.0"}`

	dienst := manifestdienst(t, ohnePlattform)
	klient := aktualisierung.NeuerKlient(dienst.URL,
		netz.NeuerKlient(netz.UeblicheFristen()),
		netz.NeuesDrosselwerk(time.Now, func(time.Duration) {}), "")

	freigabe := aktualisierung.Freigabe{
		Manifest: aktualisierung.Beigabe{
			Name: aktualisierung.Manifestname, Adresse: dienst.URL,
		},
	}

	summe := sha256.Sum256([]byte(ohnePlattform))
	liste := hex.EncodeToString(summe[:]) + "  ./" + aktualisierung.Manifestname + "\n"

	if _, err := klient.Manifest(t.Context(), freigabe, []byte(liste)); !errors.Is(
		err, domain.ErrValidation) {
		t.Fatalf("abgewiesen wurde mit %v", err)
	}
}

// **Ohne Bindung an ein Abbild gibt es kein Manifest.**
//
// Befund 4 des Nachlaufs zum Zweitblick über Phase 5.2: der Vergleich lief nur,
// wenn **beide** Angaben dastanden. Ein Manifest ohne `artefakt` oder eine
// Liste ohne Abbildzeile kam ohne Vergleich durch — der zugesagte Schutz war
// eine Wahl des Gegenübers.
func TestEinManifestOhneBindungAnDasAbbildWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	klientFuer := func(inhalt string) (*aktualisierung.Klient, aktualisierung.Freigabe) {
		dienst := manifestdienst(t, inhalt)

		return aktualisierung.NeuerKlient(dienst.URL,
				netz.NeuerKlient(netz.UeblicheFristen()),
				netz.NeuesDrosselwerk(time.Now, func(time.Duration) {}), ""),
			aktualisierung.Freigabe{
				Fassung: "0.5.0",
				Manifest: aktualisierung.Beigabe{
					Name: aktualisierung.Manifestname, Adresse: dienst.URL,
				},
				Abbild: aktualisierung.Beigabe{Name: "MagicSetEditor4Mac-0.5.0.dmg"},
			}
	}

	t.Run("das Manifest nennt keines", func(t *testing.T) {
		t.Parallel()

		ohneArtefakt := `{"fassung":"0.5.0","plattform":"darwin/arm64"}`
		klient, freigabe := klientFuer(ohneArtefakt)

		summe := sha256.Sum256([]byte(ohneArtefakt))
		liste := hex.EncodeToString(summe[:]) + "  ./" + aktualisierung.Manifestname +
			"\nabc123  ./MagicSetEditor4Mac-0.5.0.dmg\n"

		if _, err := klient.Manifest(t.Context(), freigabe, []byte(liste)); !errors.Is(
			err, domain.ErrValidation) {
			t.Fatalf("abgewiesen wurde mit %v", err)
		}
	})

	t.Run("die Liste nennt keines", func(t *testing.T) {
		t.Parallel()

		klient, freigabe := klientFuer(echtesManifest)

		if _, err := klient.Manifest(t.Context(), freigabe,
			[]byte(summenliste())); !errors.Is(err, domain.ErrConflict) {
			t.Fatalf("abgewiesen wurde mit %v", err)
		}
	})

	t.Run("ohne Abbild in der Freigabe gibt es nichts zu binden", func(t *testing.T) {
		t.Parallel()

		klient, freigabe := klientFuer(echtesManifest)
		freigabe.Abbild = aktualisierung.Beigabe{}

		if _, err := klient.Manifest(t.Context(), freigabe,
			[]byte(summenliste())); err != nil {
			t.Fatalf("die Vorprüfung ohne Abbild: %v", err)
		}
	})
}
