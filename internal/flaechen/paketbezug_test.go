// **Ein Vorlagenpaket beziehen — über jede Fläche ohne Fenster**
// (#1169).
//
// # Warum kein Aufruf hinausgeht
//
// Dieselbe Regel wie bei Scryfall (§3, seit #1021): eine **umgelenkte
// Adresse** und ein **eigenes Zuhause**, und beides einzeln reicht nicht. Die
// Gegenstelle hier liefert ein `tar.gz`, wie GitHub es ausliefert — alles
// unter einem Wurzelverzeichnis, dessen Name den Stand trägt.
//
// **Das Archiv ist selbst geschrieben** (§11): eigener Inhalt, dieselbe
// Gestalt. Ein bezogenes Paket kommt hier nicht ins Repositorium.
package flaechen_test

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// bezugsstand ist der Commit, unter dem das Probearchiv liegt.
//
// **Ein Stand und kein Zweig** — das ist die Regel, die `internal/bezug`
// durchsetzt, und der Aufruf muss sie an jeder Fläche einhalten.
const bezugsstand = "0123456789abcdef0123456789abcdef01234567"

// **Jede Fläche bezieht ihr Paket, und danach steht es in ihrer Liste.**
//
// # Die Zusage
//
// Nicht „derselbe Rückgabewert" — bezogen wird an jeder Fläche in ein
// **eigenes** Zuhause. Die Zusage ist: **alle drei können es**, und danach
// führt jede das bezogene Paket in ihrer Liste.
//
// # Warum drei Umgebungen und nicht eine
//
// Der erste Anlauf ließ alle drei in dasselbe Zuhause beziehen, unter drei
// Namen — und die Liste führte danach **ein** `bezogen`. Das ist kein Fehler
// der Anwendung: ein Paketname ist eindeutig, und drei Ablagen desselben
// Pakets sind ein Paket. Der Lauf hatte eine Zusage erfunden, die es nicht
// gibt.
func TestEinPaketBeziehenUeberJedeFlaeche(t *testing.T) {
	server := bezugsgegenstelle(t)

	const unter = "pakete"

	for name, beziehe := range map[string]func([]string){
		"Kommandozeile": func(umgebung []string) {
			mussKommandozeile(t, umgebung, "paket-beziehen",
				server.URL, bezugsstand, unter, "bezogen")
		},
		"MCP": func(umgebung []string) {
			if text, istFehler := mcpRuf(t, umgebung, "schreiben.paket.beziehen",
				map[string]any{
					"adresse": server.URL, "stand": bezugsstand,
					"unterverzeichnis": unter, "name": "bezogen",
				}); istFehler {
				t.Fatalf("schreiben.paket.beziehen: %s", text)
			}
		},
		"REST": func(umgebung []string) {
			an := starteDienst(t, umgebung)

			status, rumpf := an.ruf(t, http.MethodPost, "/pakete",
				`{"adresse":"`+server.URL+`","stand":"`+bezugsstand+
					`","unterverzeichnis":"`+unter+`","name":"bezogen"}`)
			if status != http.StatusOK && status != http.StatusCreated {
				t.Fatalf("POST pakete: %d (%s)", status, rumpf)
			}
		},
	} {
		t.Run(name, func(t *testing.T) {
			umgebung := mitEigenerSaat(t)

			// **Vorher steht es nicht da** — sonst sagte der Lauf darunter
			// auch dann ja, wenn gar nichts bezogen würde.
			if aus := mussKommandozeile(t, umgebung, "pakete"); strings.Contains(aus, "bezogen") {
				t.Fatalf("das Paket lag schon vorher da:\n%s", aus)
			}

			beziehe(umgebung)

			if aus := mussKommandozeile(t, umgebung, "pakete"); !strings.Contains(aus, "bezogen") {
				t.Errorf("das bezogene Paket steht nicht in der Liste:\n%s", aus)
			}
		})
	}
}

// **Ein Zweig statt eines Standes wird überall gleich abgewiesen.**
//
// # Warum gerade dieser Fall
//
// Er ist die Zusage, für die es `internal/bezug` gibt: „Der Stand ist ein
// Commit, kein Zweig" — ein Zweig zeigt morgen woanders hin, und ein Paket,
// das sich unter der Hand ändert, macht jede Prüfsumme wertlos. Die
// Gegenstelle antwortet auf `main` mit 404.
func TestEinZweigStattEinesStandsWirdUeberallAbgewiesen(t *testing.T) {
	server := bezugsgegenstelle(t)
	umgebung := mitEigenerSaat(t)

	const unter = "pakete"

	an := starteDienst(t, umgebung)

	abweisungGleich(t, "ein Zweig statt eines Standes", abweisung{
		kern:   "Ein Zweig ist kein Stand",
		klasse: klasseValidierung,
		cli: func() (string, int) {
			aus, fehler, code := kommandozeile(t, umgebung, "paket-beziehen",
				server.URL, "main", unter, "aus-der-zeile")

			return aus + fehler, code
		},
		mcp: func() (string, bool) {
			return mcpRuf(t, umgebung, "schreiben.paket.beziehen",
				map[string]any{
					"adresse": server.URL, "stand": "main",
					"unterverzeichnis": unter, "name": "aus-dem-modell",
				})
		},
		rest: func() (int, string) {
			return an.ruf(t, http.MethodPost, "/pakete",
				`{"adresse":"`+server.URL+`","stand":"main","unterverzeichnis":"`+unter+
					`","name":"aus-der-schleife"}`)
		},
		fenster: func() *app.Fehler {
			return fensterAus(t, umgebung).PaketBeziehen(t.Context(), server.URL, "main", unter, "aus-dem-fenster").Fehler
		},
	})
}

// bezugsgegenstelle liefert ein Probearchiv unter genau einem Stand.
//
// **Nur unter einem**, damit der Fall „ein Zweig" wirklich scheitert: eine
// Gegenstelle, die jeden Pfad bedient, machte aus der Zusage ein
// Wunschdenken.
func bezugsgegenstelle(t *testing.T) *httptest.Server {
	t.Helper()

	inhalt := probearchiv(t)

	server := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, anfrage *http.Request) {
			if !strings.HasSuffix(anfrage.URL.Path, "/"+bezugsstand) {
				schreiben.WriteHeader(http.StatusNotFound)

				return
			}

			schreibeBytes(t, schreiben, inhalt)
		}))

	t.Cleanup(server.Close)

	return server
}

// probearchiv baut ein `tar.gz` mit einem winzigen Stilpaket darin.
func probearchiv(t *testing.T) []byte {
	t.Helper()

	var gepackt bytes.Buffer

	packer := gzip.NewWriter(&gepackt)
	schreiber := tar.NewWriter(packer)

	const stil = "mse version: 2.0.0\ngame: probe\nshort name: Bezogen\n"

	kopf := &tar.Header{
		Name: filepath.Join("wurzel-"+bezugsstand, "pakete",
			"bezogen.mse-style", "style"),
		Mode:     0o644,
		Size:     int64(len(stil)),
		Typeflag: tar.TypeReg,
	}

	if err := schreiber.WriteHeader(kopf); err != nil {
		t.Fatalf("Kopf schreiben: %v", err)
	}

	if _, err := schreiber.Write([]byte(stil)); err != nil {
		t.Fatalf("Inhalt schreiben: %v", err)
	}

	if err := schreiber.Close(); err != nil {
		t.Fatalf("tar schließen: %v", err)
	}

	if err := packer.Close(); err != nil {
		t.Fatalf("gzip schließen: %v", err)
	}

	return gepackt.Bytes()
}
