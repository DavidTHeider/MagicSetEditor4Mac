package app

// **Die zwei Bindungen mit Dialog** — Werkstück N4 von Phase 5.4
// (#1457).
//
// # Was hier geprüft wird und was nicht
//
// **Nicht der native Dialog** — er braucht ein Fenster und steht in
// `.covignore` (`dateiwahl.go`). Geprüft wird, was davor und danach steht:
// dass der Filter auf `.msepackage` zeigt, dass ein leerer Pfad ein
// **Abbruch** ist und kein Fehler, und dass ein Fehler des Dialogs eine leere
// Liste hinterlässt statt `null`.

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/austausch"
)

// **Ein weggeklickter Dialog ist ein Abbruch, kein Fehler.**
func TestEinWeggeklickterUmschlagdialogIstEinAbbruch(t *testing.T) {
	anwendung := mitSaat(t).MitPeerbund(&peerprobebund{})
	anwendung.MitDateiwahl(func(context.Context, Dateiauswahl) (string, error) {
		return "", nil
	})

	bindungen := NeueBindungen(anwendung)

	for name, antwort := range map[string]UmschlagAntwort{
		"schnüren": bindungen.UmschlagSchnueren(t.Context(), "", "1"),
		"wählen":   bindungen.UmschlagWaehlen(t.Context()),
	} {
		if !antwort.Abgebrochen {
			t.Errorf("%s: der Abbruch kam nicht an", name)
		}

		if antwort.Fehler != nil {
			t.Errorf("%s: ein Abbruch ist kein Fehler (%+v)", name, antwort.Fehler)
		}

		// **Eine leere Liste und kein `nil`** — sonst steht im Fenster
		// `objekte: null`, und jede Schleife darüber fällt um.
		if antwort.Objekte == nil {
			t.Errorf("%s: die Liste ist nil", name)
		}
	}
}

// **Ein Fehler des Dialogs kommt als Fehler an — mit leerer Liste.**
func TestEinFehlerDesUmschlagdialogsKommtAn(t *testing.T) {
	anwendung := mitSaat(t).MitPeerbund(&peerprobebund{})
	anwendung.MitDateiwahl(func(context.Context, Dateiauswahl) (string, error) {
		return "", errors.New("kein Fenster")
	})

	bindungen := NeueBindungen(anwendung)

	for name, antwort := range map[string]UmschlagAntwort{
		"schnüren": bindungen.UmschlagSchnueren(t.Context(), "", "1"),
		"wählen":   bindungen.UmschlagWaehlen(t.Context()),
	} {
		if antwort.Fehler == nil {
			t.Errorf("%s: der Fehler ging verloren", name)
		}

		if antwort.Objekte == nil {
			t.Errorf("%s: die Liste ist nil", name)
		}
	}
}

// **Der Dialog fragt nach `.msepackage` — hin wie zurück.**
func TestDerUmschlagdialogZeigtAufDieEndung(t *testing.T) {
	anwendung := mitSaat(t).MitPeerbund(&peerprobebund{})

	var gesehen []Dateiauswahl

	anwendung.MitDateiwahl(func(_ context.Context, auswahl Dateiauswahl) (string, error) {
		gesehen = append(gesehen, auswahl)

		return "", nil
	})

	bindungen := NeueBindungen(anwendung)
	bindungen.UmschlagSchnueren(t.Context(), "", "1")
	bindungen.UmschlagWaehlen(t.Context())

	if len(gesehen) != 2 {
		t.Fatalf("der Dialog ging %dmal auf", len(gesehen))
	}

	for _, auswahl := range gesehen {
		if auswahl.Muster != "*"+austausch.Endung {
			t.Errorf("das Muster ist %q", auswahl.Muster)
		}
	}

	// **Nur der Weg hinaus schreibt** — und nur er schlägt einen Namen vor.
	if !gesehen[0].Speichern || gesehen[0].Vorschlag == "" {
		t.Errorf("das Schnüren geht als %+v auf", gesehen[0])
	}

	if gesehen[1].Speichern {
		t.Error("das Ansehen ging als Speichern-Dialog auf")
	}
}

// **Der gewählte Pfad kommt zurück — das Übernehmen braucht ihn.**
func TestDieWahlGibtDenPfadZurueck(t *testing.T) {
	anwendung := mitSaat(t).MitPeerbund(&peerprobebund{})

	ctx := anwendung.imProfil(t.Context())

	if antwort := anwendung.PeeridentitaetSicherstellen(ctx); antwort.Fehler != nil {
		t.Fatalf("Identität: %+v", antwort.Fehler)
	}

	ich, fehler := anwendung.peerwerk().Ausweis(ctx)
	if fehler != nil {
		t.Fatalf("Ausweis: %v", fehler)
	}

	ziel := filepath.Join(t.TempDir(), "gewaehlt.msepackage")
	if _, fehler := austauschSchreiben(ziel, ich); fehler != nil {
		t.Fatalf("schreiben: %v", fehler)
	}

	anwendung.MitDateiwahl(func(context.Context, Dateiauswahl) (string, error) {
		return ziel, nil
	})

	antwort := NeueBindungen(anwendung).UmschlagWaehlen(t.Context())
	if antwort.Fehler != nil {
		t.Fatalf("wählen: %+v", antwort.Fehler)
	}

	// **Ohne diesen Pfad kennt das Fenster die Sendung nicht** — gewählt hat
	// sie der native Dialog, und das Übernehmen ist ein zweiter Ruf.
	if antwort.Pfad != ziel {
		t.Errorf("zurück kam %q statt %q", antwort.Pfad, ziel)
	}
}

// **Und das Schnüren geht wirklich durch den Dialog hindurch.**
func TestDerGewaehlteOrtWirdGeschnuert(t *testing.T) {
	anwendung := mitSaat(t).MitPeerbund(&peerprobebund{})

	ctx := t.Context()

	geoeffnet := anwendung.SetOeffnen(ctx, saatset(t))
	if geoeffnet.Fehler != nil {
		t.Fatalf("öffnen: %+v", geoeffnet.Fehler)
	}

	if antwort := anwendung.PeeridentitaetSicherstellen(
		anwendung.imProfil(ctx)); antwort.Fehler != nil {
		t.Fatalf("Identität: %+v", antwort.Fehler)
	}

	ziel := filepath.Join(t.TempDir(), "durch.msepackage")

	anwendung.MitDateiwahl(func(context.Context, Dateiauswahl) (string, error) {
		return ziel, nil
	})

	antwort := NeueBindungen(anwendung).UmschlagSchnueren(t.Context(), "", geoeffnet.SetID)
	if antwort.Fehler != nil {
		t.Fatalf("schnüren: %+v", antwort.Fehler)
	}

	angesehen := anwendung.UmschlagAnsehen(ctx, ziel)
	if angesehen.Fehler != nil {
		t.Fatalf("ansehen: %+v", angesehen.Fehler)
	}

	if len(angesehen.Objekte) == 0 {
		t.Error("der geschnürte Umschlag ist leer")
	}
}
