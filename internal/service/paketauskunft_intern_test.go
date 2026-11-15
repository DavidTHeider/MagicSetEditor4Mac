package service

// **Mehr als 300 Dateien** — G04
// (#2251).
//
// Gefunden am nativen Fenster (`g04bezuege`): `Magic-Mini-Packs` bringt mit
// einem Commit 435 Dateien, der Vergleich nannte 300, und die Vorschau war
// gekürzt — sie nannte nur einen Teil der Pakete. Wo der Vergleich gekürzt
// ist, fragt die Auskunft deshalb nach beiden Bäumen.

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/bezug"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// dreihundert ist ein Vergleich, den die Gegenstelle gekürzt hat.
func dreihundert() string {
	dateien := make([]string, 300)
	for i := range dateien {
		dateien[i] = fmt.Sprintf(`{"filename":"a.mse-style/%d","status":"added"}`, i)
	}

	return `{"status":"ahead","ahead_by":1,"files":[` + strings.Join(dateien, ",") + `]}`
}

// baumdienst antwortet auf Vergleich und Bäume; `baeume` nach dem Commit.
func baumdienst(t *testing.T, vergleich string, baeume map[string]string) (bezieher *Paketbezieher, gefragt *[]string) {
	t.Helper()
	t.Setenv("MSE_NETZ_SCHLEIFE", "an")

	gefragt = &[]string{}

	dienst := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*gefragt = append(*gefragt, r.URL.Path+"?"+r.URL.RawQuery)

		if strings.Contains(r.URL.Path, "/compare/") {
			_, _ = w.Write([]byte(vergleich)) //nolint:errcheck // ein Testserver

			return
		}

		rumpf, da := baeume[r.URL.Path[strings.LastIndex(r.URL.Path, "/")+1:]]
		if !da {
			w.WriteHeader(http.StatusNotFound)

			return
		}

		_, _ = w.Write([]byte(rumpf)) //nolint:errcheck // dito
	}))
	t.Cleanup(dienst.Close)

	bezieher = NeuerPaketbezieher(t.TempDir(), "probe")
	bezieher.api = dienst.URL + "/repos/"

	return bezieher, gefragt
}

func TestEinGekuerzterVergleichWirdAusDenBaeumenVollstaendig(t *testing.T) {
	bezieher, gefragt := baumdienst(t, dreihundert(), map[string]string{
		"alt": `{"tree":[{"path":"README.md","type":"blob","sha":"r"},{"path":"x.mse-style/style","type":"blob","sha":"1"}]}`,
		"neu": `{"tree":[{"path":"README.md","type":"blob","sha":"r"},{"path":"a.mse-style/style","type":"blob","sha":"2"},` +
			`{"path":"b.mse-style/style","type":"blob","sha":"3"}]}`,
	})

	vergleich, err := bezieher.Vergleich(t.Context(), "https://github.com/o/r", "alt", "neu")
	if err != nil {
		t.Fatalf("Vergleich: %v", err)
	}

	// **Lage und Vorsprung bleiben die des Vergleichs**; die Dateien kommen
	// aus den Bäumen, und gekürzt ist dann nichts mehr.
	if vergleich.Gekuerzt || vergleich.Lage != "ahead" || vergleich.Vorsprung != 1 || len(vergleich.Dateien) != 3 ||
		vergleich.Dateien[2] != (bezug.Aenderung{Datei: "x.mse-style/style", Art: "removed"}) {
		t.Errorf("gekürzt=%v, %d Dateien, Lage %s", vergleich.Gekuerzt, len(vergleich.Dateien), vergleich.Lage)
	}

	if strings.Join(*gefragt, " ") != "/repos/o/r/compare/alt...neu? /repos/o/r/git/trees/alt?recursive=1 /repos/o/r/git/trees/neu?recursive=1" {
		t.Errorf("gefragt: %v", *gefragt)
	}
}

func TestEinVollstaendigerVergleichFragtKeinenBaum(t *testing.T) {
	bezieher, gefragt := baumdienst(t, `{"status":"ahead","ahead_by":1,"files":[{"filename":"a","status":"added"}]}`, nil)

	if vergleich, err := bezieher.Vergleich(t.Context(), "https://github.com/o/r", "alt", "neu"); err != nil || vergleich.Gekuerzt {
		t.Errorf("Vergleich %+v, Fehler %v", vergleich, err)
	}

	if len(*gefragt) != 1 {
		t.Errorf("gefragt: %v", *gefragt)
	}
}

// **Wo die Bäume nicht helfen, bleibt der Vergleich gekürzt** — und sagt es:
// die Vorschau verlangt dann eine Bestätigung, statt vollständig auszusehen.
func TestOhneGanzeBaeumeBleibtDerVergleichGekuerzt(t *testing.T) {
	ganz := `{"tree":[{"path":"a","type":"blob","sha":"1"}]}`

	for name, baeume := range map[string]map[string]string{
		"der alte Baum fehlt":        {"neu": ganz},
		"der neue Baum fehlt":        {"alt": ganz},
		"der alte Baum ist gekürzt":  {"alt": `{"truncated":true,"tree":[]}`, "neu": ganz},
		"der neue Baum ist gekürzt":  {"alt": ganz, "neu": `{"truncated":true,"tree":[]}`},
		"ein Baum hat keine Gestalt": {"alt": `[]`, "neu": ganz},
	} {
		t.Run(name, func(t *testing.T) {
			bezieher, _ := baumdienst(t, dreihundert(), baeume)

			vergleich, err := bezieher.Vergleich(t.Context(), "https://github.com/o/r", "alt", "neu")
			if err != nil || !vergleich.Gekuerzt || len(vergleich.Dateien) != 300 {
				t.Errorf("gekürzt=%v, %d Dateien, Fehler %v", vergleich.Gekuerzt, len(vergleich.Dateien), err)
			}
		})
	}
}

func TestDieAuskunftKenntNurGitHub(t *testing.T) {
	t.Setenv("MSE_NETZ_SCHLEIFE", "an")

	var gefragt []string

	dienst := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gefragt = append(gefragt, r.URL.Path)

		if strings.Contains(r.URL.Path, "/compare/") {
			_, _ = w.Write([]byte(`{"status":"ahead","ahead_by":1}`)) //nolint:errcheck // ein Testserver

			return
		}

		_, _ = w.Write([]byte(`{"sha":"` + aktualisierungsspitze + `","commit":{"committer":{"date":"2026-10-01T10:00:00Z"}}}`)) //nolint:errcheck // dito
	}))
	t.Cleanup(dienst.Close)

	bezieher := NeuerPaketbezieher(t.TempDir(), "probe")
	bezieher.api = dienst.URL + "/repos/"

	if spitze, err := bezieher.Spitze(t.Context(), "https://github.com/o/r.git"); err != nil || spitze.Commit != aktualisierungsspitze {
		t.Errorf("Spitze %+v, Fehler %v", spitze, err)
	}

	if vergleich, err := bezieher.Vergleich(t.Context(), "https://github.com/o/r/", "v1.0", aktualisierungsspitze); err != nil || vergleich.Vorsprung != 1 {
		t.Errorf("Vergleich %+v, Fehler %v", vergleich, err)
	}

	if strings.Join(gefragt, " ") != "/repos/o/r/commits/HEAD /repos/o/r/compare/v1.0..."+aktualisierungsspitze {
		t.Errorf("gefragt: %v", gefragt)
	}

	for _, adresse := range []string{"https://gitlab.com/o/r", "https://github.com/nur", "https://github.com/o/r/tree/main"} {
		var feldfehler *domain.FieldError

		if _, err := bezieher.Spitze(t.Context(), adresse); !errors.As(err, &feldfehler) {
			t.Errorf("%s: Spitze %v", adresse, err)
		}

		if _, err := bezieher.Vergleich(t.Context(), adresse, "a", "b"); !errors.As(err, &feldfehler) {
			t.Errorf("%s: Vergleich %v", adresse, err)
		}
	}

	if NeuerPaketbezieher("/x", "probe").api != "" {
		t.Error("ohne Test fragt der Bezieher nicht GitHub")
	}
}

func TestOhneTestserverFragtDerBezieherGitHub(t *testing.T) {
	adresse, err := NeuerPaketbezieher("/x", "probe").auskunftsadresse("https://github.com/o/r")
	if err != nil || adresse != "https://api.github.com/repos/o/r" {
		t.Errorf("Adresse %q, Fehler %v", adresse, err)
	}
}
