package service

// **Die Fehlerwege der Paketaktualisierung** — G04
// (#2251).
//
// Jeder Zug, der scheitern kann, scheitert hier einmal — und danach liegt
// jeder Baum wieder an seinem Platz. Ein Tausch, der auf halber Strecke
// stehen bleibt, hinterließe einen Bezug ohne Ort.

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/bezug"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

const aktualisierungsspitze = "c8e38c9b138c5e9c66199fd099a353bce6ff5ec7"

// zaehlenderKatalog scheitert ab dem n-ten Lesen der Liste — für den Weg,
// auf dem dieselbe Frage zweimal gestellt wird.
type zaehlenderKatalog struct {
	innerkatalog
	ab, gelesen int
}

func (k *zaehlenderKatalog) Pakete(ctx context.Context) ([]domain.Paket, error) {
	k.gelesen++
	if k.ab > 0 && k.gelesen >= k.ab {
		return nil, errors.New("der Katalog ist fort")
	}

	return k.innerkatalog.Pakete(ctx)
}

type innereAuskunft struct {
	vergleich bezug.Vergleich
	fehler    error
}

func (a innereAuskunft) Spitze(context.Context, string) (bezug.Spitzenstand, error) {
	return bezug.Spitzenstand{Commit: aktualisierungsspitze}, a.fehler
}

func (a innereAuskunft) Vergleich(context.Context, string, string, string) (bezug.Vergleich, error) {
	return a.vergleich, a.fehler
}

type scheiternderSetleser struct{}

func (scheiternderSetleser) Setkoepfe(context.Context) ([]domain.SetHead, error) {
	return nil, errors.New("der Bestand ist fort")
}

// innererBezug legt einen bezogenen Baum an und gibt die Aktualisierung
// darüber.
//
// **Die Adresse führt nirgends hin**, wo nicht ausdrücklich eine echte steht:
// ein Lauf, der bis zum Bezug kommt, lud sonst wirklich von GitHub — am
// 2026-10-10 geschehen, 26 Sekunden lang, weil ein Zähler um eins danebenlag.
func innererBezug(t *testing.T, katalog *zaehlenderKatalog, auskunft Quellauskunft) (aktualisierung *Paketaktualisierung, verwaltet string) {
	t.Helper()

	return innererBezugVon(t, katalog, auskunft, "https://beispiel.invalid/o/r")
}

func innererBezugVon(t *testing.T, katalog *zaehlenderKatalog, auskunft Quellauskunft, adresse string) (aktualisierung *Paketaktualisierung, verwaltet string) {
	t.Helper()

	verwaltet = filepath.Join(t.TempDir(), "pakete")
	if err := os.MkdirAll(filepath.Join(verwaltet, "Pack", "data"), 0o750); err != nil {
		t.Fatal(err)
	}

	katalog.pakete = []domain.Paket{
		{
			Art: "mse-game", Name: "magic", Pfad: filepath.Join(verwaltet, "Pack", "data", "magic.mse-game"),
			Herkunft: adresse + "@" + strings.Repeat("a", 40), Pruefsumme: "sha256:alt",
		},
		{Art: "mse-game", Name: "draussen", Pfad: filepath.Join(t.TempDir(), "x.mse-game"), Herkunft: "a@b"},
		{Art: "mse-game", Name: "relativ", Pfad: "relativ/x.mse-game", Herkunft: "a@b"},
		{Art: "mse-game", Name: "lose", Pfad: filepath.Join(verwaltet, "lose.mse-game"), Herkunft: "a@b"},
	}

	return NeuePaketaktualisierung(NeuerPaketverwalter(katalog, verwaltet), NeuerPaketbezieher(verwaltet, "probe"), auskunft), verwaltet
}

func TestDerVermerkKommtAusDerDateiOderAusDemKatalog(t *testing.T) {
	katalog := &zaehlenderKatalog{}
	aktualisierung, verwaltet := innererBezugVon(t, katalog, innereAuskunft{}, "https://github.com/MagicSetEditorPacks/M15-Magic-Pack")
	ort := filepath.Join(verwaltet, "Pack")

	staende, err := aktualisierung.Staende(t.Context(), false)
	if err != nil || len(staende) != 1 {
		t.Fatalf("Stände %+v, Fehler %v — ein Baum außerhalb, ein relativer und ein loser zählen nicht", staende, err)
	}

	vermerk, err := aktualisierung.vermerkVon(t.Context(), staende[0])
	if err != nil || vermerk.Pruefsumme != "sha256:alt" || vermerk.Adresse != "https://github.com/MagicSetEditorPacks/M15-Magic-Pack" {
		t.Errorf("aus dem Katalog: %+v, %v", vermerk, err)
	}

	if err := vermerkSchreiben(ort, Bezugsvermerk{Adresse: "aus der Datei"}); err != nil {
		t.Fatal(err)
	}

	if vermerk, err := aktualisierung.vermerkVon(t.Context(), staende[0]); err != nil || vermerk.Adresse != "aus der Datei" {
		t.Errorf("aus der Datei: %+v, %v", vermerk, err)
	}

	if err := os.WriteFile(filepath.Join(ort, bezugsvermerkDatei), []byte("{"), 0o600); err != nil {
		t.Fatal(err)
	}

	if _, err := vorigerVermerk(ort); err == nil || errors.Is(err, domain.ErrNotFound) {
		t.Errorf("ein kaputter Vermerk: %v", err)
	}

	if err := os.Remove(filepath.Join(ort, bezugsvermerkDatei)); err != nil {
		t.Fatal(err)
	}

	if err := os.Mkdir(filepath.Join(ort, bezugsvermerkDatei), 0o750); err != nil {
		t.Fatal(err)
	}

	if _, err := vorigerVermerk(ort); err == nil || errors.Is(err, domain.ErrNotFound) {
		t.Errorf("ein unlesbarer Vermerk: %v", err)
	}

	if err := vermerkSchreiben(filepath.Join(ort, "gibt", "es", "nicht"), Bezugsvermerk{}); err == nil {
		t.Error("ein Vermerk in einen Baum, den es nicht gibt")
	}

	katalog.ab = katalog.gelesen + 1
	if _, err := aktualisierung.vermerkVon(t.Context(), Bezugsstand{Name: "Pack", Ort: filepath.Join(verwaltet, "fehlt")}); err == nil {
		t.Error("ohne Katalog kein Vermerk")
	}
}

func TestEinAufgehobenerStandOhneVermerkLiegtTrotzdemBereit(t *testing.T) {
	aktualisierung, verwaltet := innererBezug(t, &zaehlenderKatalog{}, innereAuskunft{})

	if err := os.MkdirAll(filepath.Join(verwaltet+"-vorige-bezuege", "Pack"), 0o750); err != nil {
		t.Fatal(err)
	}

	if da, stand := aktualisierung.vorige("Pack"); !da || stand != "" {
		t.Errorf("da %v, Stand %q", da, stand)
	}
}

func TestDieFehlerwegeDerAktualisierung(t *testing.T) {
	t.Run("ohne Katalog keine Stände", func(t *testing.T) {
		aktualisierung, _ := innererBezug(t, &zaehlenderKatalog{ab: 1}, innereAuskunft{})

		if _, err := aktualisierung.Staende(t.Context(), false); err == nil {
			t.Error("kein Fehler")
		}

		if _, err := aktualisierung.Vorschau(t.Context(), "Pack", aktualisierungsspitze); err == nil {
			t.Error("Vorschau ohne Katalog")
		}
	})

	t.Run("dieselbe Spitze nach einem Vergleich ist aktuell", func(t *testing.T) {
		aktualisierung, _ := innererBezug(t, &zaehlenderKatalog{}, innereAuskunft{vergleich: bezug.Vergleich{Lage: "identical"}})

		if staende, err := aktualisierung.Staende(t.Context(), true); err != nil || staende[0].Lage != BezugAktuell {
			t.Errorf("Stände %+v, Fehler %v", staende, err)
		}
	})

	t.Run("ein scheiternder Vergleich hält die Vorschau an", func(t *testing.T) {
		aktualisierung, _ := innererBezug(t, &zaehlenderKatalog{}, innereAuskunft{fehler: domain.ErrKeineAntwort})

		if _, err := aktualisierung.Vorschau(t.Context(), "Pack", aktualisierungsspitze); !errors.Is(err, domain.ErrKeineAntwort) {
			t.Errorf("Fehler %v", err)
		}
	})

	t.Run("ohne Sets keine Vorschau", func(t *testing.T) {
		aktualisierung, _ := innererBezug(t, &zaehlenderKatalog{}, innereAuskunft{vergleich: bezug.Vergleich{Lage: "ahead"}})
		aktualisierung.verwalter.MitSetkoepfen(scheiternderSetleser{})

		if _, err := aktualisierung.Vorschau(t.Context(), "Pack", aktualisierungsspitze); err == nil {
			t.Error("kein Fehler")
		}
	})

	t.Run("ein Zweig wird nicht aktualisiert", func(t *testing.T) {
		aktualisierung, _ := innererBezug(t, &zaehlenderKatalog{}, innereAuskunft{vergleich: bezug.Vergleich{Lage: "ahead"}})

		if _, err := aktualisierung.Aktualisieren(t.Context(), "Pack", "main", true); err == nil {
			t.Error("kein Fehler")
		}
	})

	// **Gelesen wird dreimal**: in der Vorschau, beim zweiten Suchen des
	// Baums und für den Vermerk, solange der Baum keinen trägt.
	t.Run("ein Katalog, der zwischen Vorschau und Bezug verschwindet", func(t *testing.T) {
		katalog := &zaehlenderKatalog{ab: 2}
		aktualisierung, _ := innererBezug(t, katalog, innereAuskunft{vergleich: bezug.Vergleich{Lage: "ahead"}})

		if _, err := aktualisierung.Aktualisieren(t.Context(), "Pack", aktualisierungsspitze, true); err == nil {
			t.Error("kein Fehler")
		}
	})

	t.Run("ein Katalog, der vor dem Vermerk verschwindet", func(t *testing.T) {
		katalog := &zaehlenderKatalog{ab: 3}
		aktualisierung, _ := innererBezug(t, katalog, innereAuskunft{vergleich: bezug.Vergleich{Lage: "ahead"}})

		if _, err := aktualisierung.Aktualisieren(t.Context(), "Pack", aktualisierungsspitze, true); err == nil {
			t.Error("kein Fehler")
		}
	})

	t.Run("ein Baum, der nicht auf der Platte liegt", func(t *testing.T) {
		aktualisierung, verwaltet := innererBezug(t, &zaehlenderKatalog{}, innereAuskunft{vergleich: bezug.Vergleich{Lage: "ahead"}})

		if err := os.RemoveAll(filepath.Join(verwaltet, "Pack")); err != nil {
			t.Fatal(err)
		}

		if _, err := aktualisierung.Aktualisieren(t.Context(), "Pack", aktualisierungsspitze, true); err == nil {
			t.Error("kein Fehler")
		}
	})

	t.Run("ein Zurück, dessen Baum fehlt", func(t *testing.T) {
		aktualisierung, verwaltet := innererBezug(t, &zaehlenderKatalog{}, innereAuskunft{})
		vorige := filepath.Join(verwaltet+"-vorige-bezuege", "Pack")

		if err := os.MkdirAll(vorige, 0o750); err != nil {
			t.Fatal(err)
		}

		if err := vermerkSchreiben(vorige, Bezugsvermerk{Stand: "x"}); err != nil {
			t.Fatal(err)
		}

		if err := os.RemoveAll(filepath.Join(verwaltet, "Pack")); err != nil {
			t.Fatal(err)
		}

		if _, err := aktualisierung.Zurueck(t.Context(), "Pack"); err == nil {
			t.Error("kein Fehler")
		}
	})
}

// **Ein Archiv kann den Vermerk nicht ersetzen** — und eines, das an seiner
// Stelle ein Verzeichnis trägt, hält den Bezug an, statt einen Baum ohne
// Vermerk einzusetzen.
func TestEinArchivMitVerzeichnisAnStelleDesVermerks(t *testing.T) {
	t.Setenv("MSE_NETZ_SCHLEIFE", "an")

	var gepackt bytes.Buffer

	packer := gzip.NewWriter(&gepackt)
	schreiber := tar.NewWriter(packer)

	for _, name := range []string{"wurzel/" + bezugsvermerkDatei + "/x", "wurzel/probe.mse-game/game"} {
		if err := schreiber.WriteHeader(&tar.Header{Name: name, Mode: 0o644, Size: 1, Typeflag: tar.TypeReg}); err != nil {
			t.Fatal(err)
		}

		if _, err := schreiber.Write([]byte("x")); err != nil {
			t.Fatal(err)
		}
	}

	if err := schreiber.Close(); err != nil {
		t.Fatal(err)
	}

	if err := packer.Close(); err != nil {
		t.Fatal(err)
	}

	dienst := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(gepackt.Bytes()) //nolint:errcheck // ein Testserver
	}))
	t.Cleanup(dienst.Close)

	wurzel := filepath.Join(t.TempDir(), "pakete")
	if err := os.MkdirAll(wurzel, 0o750); err != nil {
		t.Fatal(err)
	}

	if _, err := NeuerPaketbezieher(wurzel, "probe").beziehen(t.Context(), dienst.URL, strings.Repeat("b", 40), "", "Pack", ""); err == nil {
		t.Error("ein Baum ohne Vermerk wurde eingesetzt")
	}

	if _, err := os.Stat(filepath.Join(wurzel, "Pack")); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("es liegt etwas am Platz: %v", err)
	}
}

func TestZweiBaeumeStehenNachIhremNamen(t *testing.T) {
	katalog := &zaehlenderKatalog{}
	aktualisierung, verwaltet := innererBezug(t, katalog, innereAuskunft{})
	katalog.pakete = append(katalog.pakete, domain.Paket{
		Art: "mse-game", Name: "frueh", Pfad: filepath.Join(verwaltet, "Anfang", "frueh.mse-game"), Herkunft: "x@" + strings.Repeat("c", 40),
	})

	staende, err := aktualisierung.Staende(t.Context(), false)
	if err != nil || len(staende) != 2 || staende[0].Name != "Anfang" || staende[1].Name != "Pack" {
		t.Errorf("Stände %+v, Fehler %v", staende, err)
	}
}
