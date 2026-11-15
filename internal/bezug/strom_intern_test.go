// **Der geströmte Bezug** — Befund Z7-01, letztes Werkstück
// (#1163).
package bezug

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

// laden liest einen Strom vollständig — **für Tests**.
//
// # Warum es diesen Helfer gibt
//
// Bis Phase 3.2 hieß so eine Funktion im Produktivcode, und die Läufe über
// Fristen und Grenzen riefen sie an. `stroemen` gibt den Rumpf jetzt
// **ungelesen** heraus; ein Test, der eine Leerlauffrist misst, muss ihn
// trotzdem lesen — sonst misst er nichts.
//
// **Er ist nicht großzügiger als der Betrieb** (§3): `Hole` liest denselben
// Strom bis zum Ende, nur reicht es ihn dabei durch gzip und tar.
func laden(ctx context.Context, adresse, kenner string) ([]byte, error) {
	strom, err := stroemen(ctx, adresse, kenner)
	if err != nil {
		return nil, err
	}

	defer func() { _ = strom.Close() }() //nolint:errcheck // nur gelesen

	return io.ReadAll(strom)
}

// **Ein Bezug hält das Archiv nicht im Arbeitsspeicher.**
//
// # Der Befund
//
// `laden` gab `[]byte` zurück — bis zu zwei Gigabyte, vollständig gesammelt,
// **bevor** das Auspacken begann. Danach lief `sha256` ein zweites Mal darüber
// und `gzip` ein drittes Mal.
//
// # Was gemessen wird
//
// Der Weg **von der Gegenstelle bis auf die Platte**, an einem Archiv, das
// sich nicht verdichten lässt. Bliebe es unterwegs liegen, schlüge es mit
// seiner ganzen Größe zu Buche.
//
// **Was nicht gemessen wird**: was gzip und tar an Puffern brauchen. Die
// Schwelle ist deshalb ein Bruchteil und nicht null.
func TestEinBezugSammeltDasArchivNicht(t *testing.T) {
	const inhaltsgroesse = 16 << 20

	gepackt := archivOhneVerdichtung(t, "probe/data/datei.txt", inhaltsgroesse)

	gegenstelle := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			if _, err := schreiben.Write(gepackt); err != nil {
				t.Errorf("die Gegenstelle konnte nicht schreiben: %v", err)
			}
		}))
	t.Cleanup(gegenstelle.Close)

	belegt := belegtWaehrend(t, func() {
		strom, err := stroemen(t.Context(), gegenstelle.URL, "prüfung")
		if err != nil {
			t.Fatalf("strömen: %v", err)
		}

		defer func() { _ = strom.Close() }() //nolint:errcheck // nur gelesen

		if _, fehler := auspacken(strom, "", t.TempDir()); fehler != nil {
			t.Fatalf("auspacken: %v", fehler)
		}
	})

	// **Ein Achtel der Archivgröße.** Eine gesammelte Kopie schlüge mit der
	// ganzen zu Buche und risse die Schwelle um das Achtfache.
	schwelle := uint64(len(gepackt)) / 8

	if belegt > schwelle {
		t.Errorf("der Bezug belegte %d Byte bei einem Archiv von %d — mehr als "+
			"die Schwelle %d. Wird es gesammelt?", belegt, len(gepackt), schwelle)
	}

	t.Logf("Archiv %d Byte, belegt %d Byte (%.1f %%)",
		len(gepackt), belegt, float64(belegt)*100/float64(len(gepackt)))
}

// **Die Prüfsumme gilt dem ganzen Rumpf, auch mit Anhängsel.**
//
// tar hört am Ende seines letzten Eintrags auf. Was danach kommt, hat es nie
// gesehen — und eine Prüfsumme, die es auch nicht sieht, ist eine andere als
// die, die vorher im Katalog stand.
func TestDiePruefsummeGiltDemGanzenRumpf(t *testing.T) {
	gepackt := archivOhneVerdichtung(t, "probe/data/datei.txt", 64)
	mitAnhang := append(bytes.Clone(gepackt), []byte("und noch etwas hinterher")...)

	ohne := bezugsSumme(t, gepackt)
	mit := bezugsSumme(t, mitAnhang)

	if ohne == mit {
		t.Error("das Anhängsel änderte die Prüfsumme nicht — sie ging nur über das, was tar las")
	}
}

// bezugsSumme bezieht ein Archiv von einer Gegenstelle und gibt die Prüfsumme.
func bezugsSumme(t *testing.T, inhalt []byte) string {
	t.Helper()

	gegenstelle := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, anfrage *http.Request) {
			if anfrage.URL.Path == "/marken/v1" {
				return
			}

			if _, err := schreiben.Write(inhalt); err != nil {
				t.Errorf("die Gegenstelle konnte nicht schreiben: %v", err)
			}
		}))
	t.Cleanup(gegenstelle.Close)

	ergebnis, err := Hole(t.Context(), Repositorium{
		Adresse: gegenstelle.URL,
		Marken:  gegenstelle.URL + "/marken/",
		Stand:   "v1",
	}, t.TempDir(), "prüfung")
	if err != nil {
		t.Fatalf("beziehen: %v", err)
	}

	return ergebnis.Pruefsumme
}

// **Ein Anhängsel über der Grenze wird gemeldet, nicht überlesen.**
//
// Der Fall entsteht, weil die Prüfsumme dem **ganzen** Rumpf gilt: tar hört am
// Ende seines letzten Eintrags auf, der Rest wird leergelesen — und dabei gilt
// die Grenze weiter. Ohne diesen Weg käme ein Bezug durch, dessen Rumpf die
// Grenze reißt, solange nur der tar-Teil darunter liegt.
func TestEinAnhaengselUeberDerGrenzeWirdGemeldet(t *testing.T) {
	gepackt := archivOhneVerdichtung(t, "probe/data/datei.txt", 64)

	// **Knapp über dem Archiv**: tar kommt durch, das Anhängsel nicht.
	mitGrenze(t, int64(len(gepackt))+8)

	mitAnhang := append(bytes.Clone(gepackt), make([]byte, 4096)...)

	gegenstelle := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, anfrage *http.Request) {
			if anfrage.URL.Path == "/marken/v1" {
				return
			}

			if _, err := schreiben.Write(mitAnhang); err != nil {
				t.Errorf("die Gegenstelle konnte nicht schreiben: %v", err)
			}
		}))
	t.Cleanup(gegenstelle.Close)

	_, err := Hole(t.Context(), Repositorium{
		Adresse: gegenstelle.URL,
		Marken:  gegenstelle.URL + "/marken/",
		Stand:   "v1",
	}, t.TempDir(), "prüfung")
	if err == nil {
		t.Fatal("der Rumpf über der Grenze kam durch")
	}

	if !bytes.Contains([]byte(err.Error()), []byte("größer als")) {
		t.Errorf("die Meldung nennt die Grenze nicht: %v", err)
	}
}
