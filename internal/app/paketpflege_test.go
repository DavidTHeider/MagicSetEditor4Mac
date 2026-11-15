// **Die Paketpflege ohne Fenster** —
// #1291.
//
// Eine eigene Datei neben den Läufen über den Katalog: dort steht, **was da
// ist**, hier, was **dazukommt und verschwindet**.
package app_test

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// pflegeweg gibt einen Paketweg über einem leeren Zuhause und einen Ort, an
// dem ein selbst geschriebenes Paket liegt (§11).
func pflegeweg(t *testing.T, name string) (weg *app.Paketweg, quelle string) {
	t.Helper()

	// **Ein eigenes Zuhause** (§3): `Paketverzeichnis()` liegt neben der
	// Datenbank, und die liegt unter `HOME`. Ohne diese Zeile legte der Lauf
	// Pakete in das **echte** Verzeichnis dieses Rechners — und wäre grün,
	// weil er funktioniert hat.
	t.Setenv("HOME", t.TempDir())

	weg, err := app.PaketnahtUeber(offeneProbedatenbank(t))
	if err != nil {
		t.Fatalf("Paketnaht: %v", err)
	}

	quelle = filepath.Join(t.TempDir(), name)
	if err := os.MkdirAll(quelle, 0o750); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	inhalt := "mse version: 2.0.0\nshort name: Probe\nfull name: Probeteil\nversion: 2026-08-30\n"
	if err := os.WriteFile(filepath.Join(quelle, "style"), []byte(inhalt), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	return weg, quelle
}

// **Ein Paket wird eingebunden und steht danach im Katalog.**
func TestPaketnahtBindetEinUndFuehrtEsDanach(t *testing.T) {
	weg, quelle := pflegeweg(t, "probe-ein.mse-style")

	paket, fehler := weg.Installieren(probekontext(t), quelle)
	if fehler != nil {
		t.Fatalf("einbinden: %v", fehler)
	}

	if paket.Name != "probe-ein" {
		t.Errorf("eingebunden wurde %+v", paket)
	}

	liste, _, lesefehler := weg.Liste(probekontext(t))
	if lesefehler != nil {
		t.Fatalf("Liste: %v", lesefehler)
	}

	if !fuehrt(liste, "probe-ein") {
		t.Errorf("das Paket fehlt im Katalog: %+v", liste)
	}

	// **Und danach wieder heraus.**
	if fehler := weg.Entfernen(probekontext(t), paket.Art, paket.Name, ""); fehler != nil {
		t.Fatalf("entfernen: %v", fehler)
	}

	danach, _, nachfehler := weg.Liste(probekontext(t))
	if nachfehler != nil {
		t.Fatalf("Liste danach: %v", nachfehler)
	}

	if fuehrt(danach, "probe-ein") {
		t.Errorf("das Paket steht noch im Katalog: %+v", danach)
	}
}

// **Eine Übernahme holt, was in einem Verzeichnis liegt.**
func TestPaketnahtUebernimmt(t *testing.T) {
	weg, quelle := pflegeweg(t, "probe-ueber.mse-style")

	uebernahme, fehler := weg.Uebernehmen(probekontext(t), filepath.Dir(quelle))
	if fehler != nil {
		t.Fatalf("übernehmen: %v", fehler)
	}

	if len(uebernahme.Uebernommen) != 1 {
		t.Errorf("übernommen: %+v", uebernahme.Uebernommen)
	}

	// **Die beiden Listen sind nie `nil`** (§3).
	if uebernahme.Uebergangen == nil || uebernahme.Befunde == nil {
		t.Error("eine Liste ist nil statt leer")
	}

	// **Und ein zweiter Lauf übergeht, statt zu überschreiben.**
	nochmal, fehler := weg.Uebernehmen(probekontext(t), filepath.Dir(quelle))
	if fehler != nil {
		t.Fatalf("zweite Übernahme: %v", fehler)
	}

	if len(nochmal.Uebergangen) != 1 {
		t.Errorf("übergangen: %+v", nochmal.Uebergangen)
	}
}

// **Ohne Quelle und ohne Ort ist es eine Validierung.**
//
// **Nicht „nicht gefunden"**: der Aufruf ist unvollständig, nicht das Ziel
// abwesend — ein Skript soll das unterscheiden (§8).
func TestPaketnahtWeistLeereAngabenAb(t *testing.T) {
	weg, _ := pflegeweg(t, "probe-leer.mse-style")

	if _, fehler := weg.Installieren(probekontext(t), "   "); !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("einbinden: %v, erwartet ErrValidation", fehler)
	}

	if _, fehler := weg.Uebernehmen(probekontext(t), ""); !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("übernehmen: %v, erwartet ErrValidation", fehler)
	}
}

// **Und ein Ort, den es nicht gibt, ebenso.**
func TestPaketnahtWeistEinenFehlendenOrtAb(t *testing.T) {
	weg, _ := pflegeweg(t, "probe-fehlt.mse-style")

	fehlt := filepath.Join(t.TempDir(), "gibtsnicht")

	if _, fehler := weg.Uebernehmen(probekontext(t), fehlt); fehler == nil {
		t.Error("ein fehlender Ort ging durch")
	}

	if _, fehler := weg.Installieren(probekontext(t), fehlt); fehler == nil {
		t.Error("eine fehlende Quelle ging durch")
	}
}

// fuehrt sagt, ob der Katalog ein Paket mit diesem Namen führt.
func fuehrt(pakete []app.Paketzeile, name string) bool {
	for _, zeile := range pakete {
		if strings.Contains(zeile.Paket.Name, name) {
			return true
		}
	}

	return false
}

// **Eine Übernahme ohne Übergangene gibt eine leere Liste, kein `nil`** (§3).
//
// Der Zweig ist über die Naht erreichbar: [app.Paketweg] nimmt jeden
// `Paketfuehrer`, und der Dienst ist nur **einer** davon. Über die Wails-Naht
// wäre `nil` ein `undefined` statt einer leeren Liste.
func TestEineUebernahmeOhneUebergangeneGibtEineLeereListe(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	weg, err := app.PaketnahtUeber(offeneProbedatenbank(t))
	if err != nil {
		t.Fatalf("Paketnaht: %v", err)
	}

	weg.MitVerwalter(&paketfuehrer{uebernahme: &service.Uebernahme{}})

	uebernahme, fehler := weg.Uebernehmen(probekontext(t), "/irgendwo")
	if fehler != nil {
		t.Fatalf("übernehmen: %v", fehler)
	}

	if uebernahme.Uebergangen == nil {
		t.Error("Uebergangen ist nil statt einer leeren Liste")
	}
}

// **Und ein Auffrischen, das scheitert, nimmt die Übernahme nicht zurück.**
//
// „Geschrieben ist geschrieben" (§7, Regel 3): die Pakete **liegen** im
// Verzeichnis. Wer nur den Fehler zeigte, ließe sie zurück, ohne dass jemand
// von ihnen wüsste.
func TestEinGescheitertesAuffrischenNimmtDieUebernahmeNichtZurueck(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	weg, err := app.PaketnahtUeber(offeneProbedatenbank(t))
	if err != nil {
		t.Fatalf("Paketnaht: %v", err)
	}

	weg.MitVerwalter(&paketfuehrer{
		uebernahme:      &service.Uebernahme{Uebernommen: []domain.Paket{{Name: "probe"}}},
		auffrischfehler: errors.New("der Katalog ist zu"),
	})

	uebernahme, fehler := weg.Uebernehmen(probekontext(t), "/irgendwo")
	if fehler == nil {
		t.Fatal("der Fehler beim Auffrischen wurde verschluckt")
	}

	if uebernahme == nil || len(uebernahme.Uebernommen) != 1 {
		t.Errorf("die Übernahme ging verloren: %+v", uebernahme)
	}
}
