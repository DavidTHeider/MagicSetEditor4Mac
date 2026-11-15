package austausch_test

// **Der Umschlag, hin und zurück** — Werkstück N4 von Phase 5.4
// (#1457).
//
// # Was hier gemessen wird
//
// **Dass derselbe Inhalt herauskommt, der hineinging** — und dass **nichts**
// herauskommt, wenn unterwegs etwas geändert wurde. Beides an selbst
// geschriebenen Umschlägen; fremde Vorlagen liegen hier nicht (§11).

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/austausch"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/peer"
)

// eineIdentitaet gibt ein frisches Paar.
func eineIdentitaet(t *testing.T) peer.Identitaet {
	t.Helper()

	ich, fehler := peer.Erzeugen(nil)
	if fehler != nil {
		t.Fatalf("Erzeugen: %v", fehler)
	}

	return ich
}

// einUmschlag schreibt einen und gibt seinen Pfad.
func einUmschlag(
	t *testing.T, ich peer.Identitaet, inhalte ...austausch.Inhalt,
) string {
	t.Helper()

	pfad := filepath.Join(t.TempDir(), "probe"+austausch.Endung)

	if _, fehler := austausch.Schreibe(pfad, inhalte, ich); fehler != nil {
		t.Fatalf("Schreibe: %v", fehler)
	}

	return pfad
}

// **Was hineinging, kommt heraus.**
func TestEinUmschlagKommtHeilAn(t *testing.T) {
	t.Parallel()

	ich := eineIdentitaet(t)

	pfad := einUmschlag(t, ich,
		austausch.Inhalt{
			Art: austausch.ArtKarte, Name: "karten/eine.json",
			Bezeichnung: "Eine Karte", Bytes: []byte(`{"name":"Mausbraten"}`),
		},
		austausch.Inhalt{
			Art: austausch.ArtVorlage, Name: "vorlagen/probe.mse-style",
			Bezeichnung: "Die Vorlage", Bytes: []byte("style:\n\tname: probe"),
			Braucht: []string{"karten/eine.json"},
		})

	sendung, fehler := austausch.Lies(pfad, t.TempDir())
	if fehler != nil {
		t.Fatalf("Lies: %v", fehler)
	}

	if sendung.Urheber != ich.Kennung() {
		t.Errorf("Urheber ist %q", sendung.Urheber)
	}

	if len(sendung.Manifest.Objekte) != 2 {
		t.Fatalf("%d Objekte", len(sendung.Manifest.Objekte))
	}

	gelesen, fehler := os.ReadFile(
		filepath.Join(sendung.Verzeichnis, "karten", "eine.json"))
	if fehler != nil {
		t.Fatalf("lesen: %v", fehler)
	}

	if string(gelesen) != `{"name":"Mausbraten"}` {
		t.Errorf("gelesen wurde %q", gelesen)
	}
}

// **Ein leerer Umschlag ist keiner.**
func TestEinUmschlagOhneObjekteIstEinBefund(t *testing.T) {
	t.Parallel()

	_, fehler := austausch.Schreibe(
		filepath.Join(t.TempDir(), "leer"+austausch.Endung),
		nil, eineIdentitaet(t))

	if !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("erwartet war ein Befund, kam: %v", fehler)
	}
}

// **Eine unbekannte Objektart wird abgewiesen — beim Schreiben.**
//
// **Und nicht erst beim Lesen**: wer einen Umschlag schreibt, den niemand
// annehmen kann, erfährt es besser sofort.
func TestEineUnbekannteArtWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	_, fehler := austausch.Schreibe(
		filepath.Join(t.TempDir(), "krumm"+austausch.Endung),
		[]austausch.Inhalt{{
			Art: "zauberspruch", Name: "a", Bytes: []byte("x"),
		}}, eineIdentitaet(t))

	if !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("erwartet war ein Befund, kam: %v", fehler)
	}
}

// **Eine Leihe reist im selben Umschlag** — Werkstück G6 von Phase 5.5
// (#1463).
//
// **Die Abnahme verlangt es wörtlich**: „Datei-/LAN-/Internettransfer
// verwenden denselben N4-Vertrag."
func TestEineLeiheIstEineObjektart(t *testing.T) {
	t.Parallel()

	ziel := filepath.Join(t.TempDir(), "leihe.msepackage")

	manifest, fehler := austausch.Schreibe(ziel, []austausch.Inhalt{{
		Art: austausch.ArtLeihe, Name: "leihe.json",
		Bezeichnung: "Deckleihe", Bytes: []byte(`{"kennung":"l1"}`),
	}}, eineIdentitaet(t))
	if fehler != nil {
		t.Fatalf("schreiben: %v", fehler)
	}

	if len(manifest.Objekte) != 1 ||
		manifest.Objekte[0].Art != austausch.ArtLeihe {
		t.Fatalf("das Manifest ist %+v", manifest.Objekte)
	}

	sendung, fehler := austausch.Lies(ziel, t.TempDir())
	if fehler != nil {
		t.Fatalf("lesen: %v", fehler)
	}

	if sendung.Manifest.Objekte[0].Art != austausch.ArtLeihe {
		t.Fatalf("gelesen wurde %q", sendung.Manifest.Objekte[0].Art)
	}
}
