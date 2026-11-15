package app

// **Das eigene Fensterbild an der Fassade** — UX-09.AK-01
// (#1681).
//
// **Im Paket `app`**, aus demselben Grund wie beim Tischbild: die Läufe
// setzen `a.bilderort` und rufen `fensterbildSetzen`.

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestOhneFensterbildIstDieAdresseLeer(t *testing.T) {
	t.Parallel()

	anwendung, _ := tischapp(t)

	antwort := anwendung.Fensterbild(t.Context(), "7")
	switch {
	case antwort.Fehler != nil:
		t.Fatalf("Fehler: %v", antwort.Fehler)
	case antwort.Adresse != "":
		t.Errorf("Adresse %q, erwartet leer", antwort.Adresse)
	}
}

func TestEineUnlesbareKennungWirdAmFensterbildGemeldet(t *testing.T) {
	t.Parallel()

	anwendung, _ := tischapp(t)

	for name, antwort := range map[string]TischbildAntwort{
		"Fensterbild":          anwendung.Fensterbild(t.Context(), "keine"),
		"FensterbildEntfernen": anwendung.FensterbildEntfernen(t.Context(), "keine"),
		"fensterbildSetzen":    anwendung.fensterbildSetzen(t.Context(), "keine", ""),
	} {
		if antwort.Fehler == nil {
			t.Errorf("%s nimmt eine unlesbare Kennung hin", name)
		}
	}
}

// **Was abgelegt ist, kommt unter seiner Adresse zurück** — über den Bildweg
// derselben Anwendung, und **nicht** als Tischbild.
func TestEinAbgelegtesFensterbildKommtUeberDenBildwegZurueck(t *testing.T) {
	t.Parallel()

	anwendung, _ := tischapp(t)
	pfad := filepath.Join(t.TempDir(), "bild.png")
	schreibeProbebild(t, pfad)

	gesetzt := anwendung.fensterbildSetzen(t.Context(), "7", pfad)
	if gesetzt.Fehler != nil {
		t.Fatalf("setzen: %v", gesetzt.Fehler)
	}

	if gesetzt.Adresse != Fensterbildpfad+"7.png" {
		t.Errorf("Adresse %q", gesetzt.Adresse)
	}

	if wieder := anwendung.Fensterbild(t.Context(), "7"); wieder.Adresse != gesetzt.Adresse {
		t.Errorf("gelesen %q, gesetzt %q", wieder.Adresse, gesetzt.Adresse)
	}

	if tisch := anwendung.Tischbild(t.Context(), "7"); tisch.Adresse != "" {
		t.Errorf("das Fensterbild gilt als Tischbild: %q", tisch.Adresse)
	}

	anfrage := httptest.NewRequestWithContext(
		t.Context(), http.MethodGet, gesetzt.Adresse, http.NoBody)
	blatt := httptest.NewRecorder()
	Bildweg(anwendung).ServeHTTP(blatt, anfrage)

	if blatt.Code != http.StatusOK || blatt.Body.Len() == 0 {
		t.Errorf("Bildweg: Code %d, %d Bytes", blatt.Code, blatt.Body.Len())
	}
}

func TestEineDateiOhneBildWirdKeinFensterbild(t *testing.T) {
	t.Parallel()

	anwendung, _ := tischapp(t)
	pfad := filepath.Join(t.TempDir(), "kein.png")

	if fehler := os.WriteFile(pfad, []byte("kein Bild"), 0o600); fehler != nil {
		t.Fatalf("schreiben: %v", fehler)
	}

	if antwort := anwendung.fensterbildSetzen(t.Context(), "7", pfad); antwort.Fehler == nil {
		t.Error("eine Datei ohne Bild wurde hingenommen")
	}

	if antwort := anwendung.Fensterbild(t.Context(), "7"); antwort.Adresse != "" {
		t.Errorf("die Ablage hat sie trotzdem bekommen: %q", antwort.Adresse)
	}
}

// **Weggenommen ist weggenommen** — das ist die Rücksetzung am Bild selbst.
func TestEinWeggenommenesFensterbildIstFort(t *testing.T) {
	t.Parallel()

	anwendung, _ := tischapp(t)
	pfad := filepath.Join(t.TempDir(), "bild.png")
	schreibeProbebild(t, pfad)
	anwendung.fensterbildSetzen(t.Context(), "7", pfad)

	if antwort := anwendung.FensterbildEntfernen(t.Context(), "7"); antwort.Fehler != nil {
		t.Fatalf("entfernen: %v", antwort.Fehler)
	}

	if _, da := Fensterbilder(anwendung).Fensterbild("7"); da {
		t.Error("die Leitung gibt noch ein Bild her")
	}

	if _, da := Fensterbilder(anwendung).Fensterbild("keine"); da {
		t.Error("eine unlesbare Kennung gibt ein Bild her")
	}
}

// **Ohne Ort für Bilder** gibt es eine Meldung und kein Bild — an allen
// Wegen, die einen brauchen.
func TestOhneBilderortMeldetDasFensterbild(t *testing.T) {
	t.Parallel()

	anwendung := &App{bilderort: func() (string, error) { return "", os.ErrNotExist }}

	if antwort := anwendung.Fensterbild(t.Context(), "7"); antwort.Fehler == nil {
		t.Error("Fensterbild ohne Ort meldet nichts")
	}

	if antwort := anwendung.fensterbildSetzen(t.Context(), "7", "/gibt/es/nicht.png"); antwort.Fehler == nil {
		t.Error("fensterbildSetzen ohne Datei meldet nichts")
	}

	if _, da := Fensterbilder(anwendung).Fensterbild("7"); da {
		t.Error("die Leitung gibt ohne Ort ein Bild her")
	}
}

// **Der Dialog wird gefragt** — was er nennt, wird abgelegt; ein Abbruch
// lässt das vorige Bild stehen, und ein Dialog, der klagt, wird gemeldet.
// Dieselben drei Wege wie beim Tischbild.
func TestDerFensterbilddialogLegtAbLaesstStehenOderMeldet(t *testing.T) {
	t.Parallel()

	anwendung, _ := tischapp(t)
	pfad := filepath.Join(t.TempDir(), "bild.png")
	schreibeProbebild(t, pfad)

	anwendung.dateiwahl = func(context.Context, Dateiauswahl) (string, error) {
		return pfad, nil
	}

	gewaehlt := anwendung.FensterbildWaehlen(t.Context(), "7")
	if gewaehlt.Fehler != nil || gewaehlt.Adresse != Fensterbildpfad+"7.png" {
		t.Fatalf("gewählt: %+v", gewaehlt)
	}

	anwendung.dateiwahl = func(context.Context, Dateiauswahl) (string, error) {
		return "", nil
	}

	abgebrochen := anwendung.FensterbildWaehlen(t.Context(), "7")
	if abgebrochen.Fehler != nil || abgebrochen.Adresse != gewaehlt.Adresse {
		t.Errorf("abgebrochen: %+v — das vorige Bild sollte stehen bleiben", abgebrochen)
	}

	anwendung.dateiwahl = func(context.Context, Dateiauswahl) (string, error) {
		return "", os.ErrPermission
	}

	if geklagt := anwendung.FensterbildWaehlen(t.Context(), "7"); geklagt.Fehler == nil {
		t.Error("ein klagender Dialog kam ohne Befund")
	}
}

// **Ohne Ort für Bilder** meldet auch der Weg, dessen Datei lesbar ist —
// und das Wegnehmen. Der Lauf darüber betritt nur den Lesefehler.
func TestOhneBilderortMeldenAblegenUndWegnehmen(t *testing.T) {
	t.Parallel()

	pfad := filepath.Join(t.TempDir(), "bild.png")
	schreibeProbebild(t, pfad)

	anwendung := &App{bilderort: func() (string, error) { return "", os.ErrNotExist }}

	if antwort := anwendung.fensterbildSetzen(t.Context(), "7", pfad); antwort.Fehler == nil {
		t.Error("das Ablegen ohne Ort meldet nichts")
	}

	if antwort := anwendung.FensterbildEntfernen(t.Context(), "7"); antwort.Fehler == nil {
		t.Error("das Wegnehmen ohne Ort meldet nichts")
	}
}

// **Ein Verzeichnis, in das sich nicht schreiben lässt, wird gemeldet** (§8)
// — beim Ablegen wie beim Wegnehmen.
func TestEineUnbeschreibbareAblageMeldetDasFensterbild(t *testing.T) {
	t.Parallel()

	anwendung, ort := tischapp(t)
	nutzerordner := filepath.Join(ort, "nutzer")
	pfad := filepath.Join(t.TempDir(), "bild.png")
	schreibeProbebild(t, pfad)

	// **Erst ein Bild ablegen, dann zusperren** — sonst gäbe es beim
	// Wegnehmen nichts wegzunehmen, und das ist kein Fehler.
	if antwort := anwendung.fensterbildSetzen(t.Context(), "7", pfad); antwort.Fehler != nil {
		t.Fatalf("ablegen: %v", antwort.Fehler)
	}

	//#nosec G302 -- absichtlich zugesperrt, um den Fehlerweg zu betreten
	if fehler := os.Chmod(nutzerordner, 0o500); fehler != nil {
		t.Fatalf("zusperren: %v", fehler)
	}

	t.Cleanup(func() {
		//#nosec G302 -- ein Verzeichnis, das der Lauf selbst angelegt hat
		if fehler := os.Chmod(nutzerordner, 0o750); fehler != nil {
			t.Errorf("aufsperren: %v", fehler)
		}
	})

	if antwort := anwendung.fensterbildSetzen(t.Context(), "7", pfad); antwort.Fehler == nil {
		t.Error("das Ablegen in ein zugesperrtes Verzeichnis kam ohne Befund")
	}

	if antwort := anwendung.FensterbildEntfernen(t.Context(), "7"); antwort.Fehler == nil {
		t.Error("das Wegnehmen aus einem zugesperrten Verzeichnis kam ohne Befund")
	}
}
