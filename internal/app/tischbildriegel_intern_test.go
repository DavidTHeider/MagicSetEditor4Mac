package app

// **Das eigene Tischbild an der Fassade** — Werkstück G3 von Phase 5.5
// (#1460).
//
// **Im Paket `app` und nicht daneben**, weil die Läufe `a.bilderort` setzen
// und `tischbilddatei` rufen — beides steht nicht draußen.

import (
	"context"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

// tischapp gibt eine Anwendung, deren Bilder in einem eigenen Verzeichnis
// liegen.
func tischapp(t *testing.T) (anwendung *App, bilderort string) {
	t.Helper()

	ort := t.TempDir()

	return &App{bilderort: func() (string, error) { return ort, nil }}, ort
}

// **Ohne abgelegtes Bild ist die Adresse leer** — und das ist keine Klage:
// die meisten Menschen bringen keines mit (§8).
func TestOhneTischbildIstDieAdresseLeer(t *testing.T) {
	t.Parallel()

	anwendung, _ := tischapp(t)

	antwort := anwendung.Tischbild(t.Context(), "7")
	switch {
	case antwort.Fehler != nil:
		t.Fatalf("Fehler: %v", antwort.Fehler)
	case antwort.Adresse != "":
		t.Errorf("Adresse %q, erwartet leer", antwort.Adresse)
	}
}

// **Eine Kennung, die keine ist, wird gemeldet** — an allen drei Wegen.
func TestEineUnlesbareKennungWirdGemeldet(t *testing.T) {
	t.Parallel()

	anwendung, _ := tischapp(t)

	for name, antwort := range map[string]TischbildAntwort{
		"Tischbild":          anwendung.Tischbild(t.Context(), "keine"),
		"TischbildEntfernen": anwendung.TischbildEntfernen(t.Context(), "keine"),
		"tischbildSetzen":    anwendung.tischbildSetzen(t.Context(), "keine", ""),
	} {
		if antwort.Fehler == nil {
			t.Errorf("%s nimmt eine unlesbare Kennung hin", name)
		}
	}
}

// **Was abgelegt ist, kommt unter seiner Adresse zurück** — und die Leitung
// gibt dieselben Bytes her.
func TestEinAbgelegtesTischbildKommtZurueck(t *testing.T) {
	t.Parallel()

	anwendung, ort := tischapp(t)
	pfad := filepath.Join(t.TempDir(), "bild.png")
	schreibeProbebild(t, pfad)

	gesetzt := anwendung.tischbildSetzen(t.Context(), "7", pfad)
	if gesetzt.Fehler != nil {
		t.Fatalf("setzen: %v", gesetzt.Fehler)
	}

	if gesetzt.Adresse != Tischbildpfad+"7.png" {
		t.Errorf("Adresse %q", gesetzt.Adresse)
	}

	if _, fehler := os.Stat(filepath.Join(ort, "nutzer", tischbilddatei(7))); fehler != nil {
		t.Errorf("die Datei liegt nicht da: %v", fehler)
	}

	bytes, da := Tischbilder(anwendung).Tischbild("7")
	switch {
	case !da:
		t.Error("die Leitung gibt nichts her")
	case len(bytes) == 0:
		t.Error("die Leitung gibt nichts als leere Bytes")
	}
}

// **Eine Datei, die kein Bild ist, wird abgewiesen** — und zwar **bevor** sie
// in der Ablage liegt (§8).
func TestEineDateiOhneBildKommtNichtAufDenTisch(t *testing.T) {
	t.Parallel()

	anwendung, ort := tischapp(t)
	pfad := filepath.Join(t.TempDir(), "kein.png")

	if fehler := os.WriteFile(pfad, []byte("kein Bild"), 0o600); fehler != nil {
		t.Fatalf("schreiben: %v", fehler)
	}

	if antwort := anwendung.tischbildSetzen(t.Context(), "7", pfad); antwort.Fehler == nil {
		t.Error("eine Datei ohne Bild wurde hingenommen")
	}

	if _, fehler := os.Stat(filepath.Join(ort, "nutzer", tischbilddatei(7))); fehler == nil {
		t.Error("die Ablage hat sie trotzdem bekommen")
	}
}

// **Weggenommen ist weggenommen**, und die Leitung sagt es auch.
func TestEinWeggenommenesTischbildIstFort(t *testing.T) {
	t.Parallel()

	anwendung, _ := tischapp(t)
	pfad := filepath.Join(t.TempDir(), "bild.png")
	schreibeProbebild(t, pfad)
	anwendung.tischbildSetzen(t.Context(), "7", pfad)

	weg := anwendung.TischbildEntfernen(t.Context(), "7")
	switch {
	case weg.Fehler != nil:
		t.Fatalf("entfernen: %v", weg.Fehler)
	case weg.Adresse != "":
		t.Errorf("Adresse %q, erwartet leer", weg.Adresse)
	}

	if _, da := Tischbilder(anwendung).Tischbild("7"); da {
		t.Error("die Leitung gibt es weiter her")
	}
}

// **Die Leitung schweigt zu einer Kennung, die keine ist.**
func TestDieLeitungSchweigtZuEinerUnlesbarenKennung(t *testing.T) {
	t.Parallel()

	anwendung, _ := tischapp(t)

	if _, da := Tischbilder(anwendung).Tischbild("keine"); da {
		t.Error("die Leitung antwortet auf eine unlesbare Kennung")
	}
}

// **Ohne Bilderort ist es ein Befund und kein Sturz.**
func TestOhneBilderortKommtEinBefund(t *testing.T) {
	t.Parallel()

	anwendung := &App{
		bilderort: func() (string, error) { return "", os.ErrNotExist },
	}

	if antwort := anwendung.Tischbild(t.Context(), "7"); antwort.Fehler == nil {
		t.Error("ohne Bilderort kam kein Befund")
	}

	if antwort := anwendung.TischbildEntfernen(t.Context(), "7"); antwort.Fehler == nil {
		t.Error("das Entfernen ohne Bilderort kam ohne Befund")
	}

	pfad := filepath.Join(t.TempDir(), "bild.png")
	schreibeProbebild(t, pfad)

	if antwort := anwendung.tischbildSetzen(t.Context(), "7", pfad); antwort.Fehler == nil {
		t.Error("das Setzen ohne Bilderort kam ohne Befund")
	}

	if _, da := Tischbilder(anwendung).Tischbild("7"); da {
		t.Error("die Leitung antwortet ohne Bilderort")
	}
}

// **Eine kaputte Datei in der Ablage ist ein Befund** und nicht „kein Bild"
// (§8): „es gibt keines" ist eine Auskunft, „da liegt etwas Unlesbares" ein
// Befund — wer beides gleich behandelte, ließe eine beschädigte Ablage
// unbemerkt.
func TestEineKaputteAblageWirdGemeldet(t *testing.T) {
	t.Parallel()

	anwendung, ort := tischapp(t)
	nutzerordner := filepath.Join(ort, "nutzer")

	if fehler := os.MkdirAll(nutzerordner, 0o750); fehler != nil {
		t.Fatalf("anlegen: %v", fehler)
	}

	kaputt := filepath.Join(nutzerordner, tischbilddatei(7))
	if fehler := os.WriteFile(kaputt, []byte("kein PNG"), 0o600); fehler != nil {
		t.Fatalf("schreiben: %v", fehler)
	}

	if antwort := anwendung.Tischbild(t.Context(), "7"); antwort.Fehler == nil {
		t.Error("eine kaputte Datei kam als „kein Bild\" zurück")
	}
}

// **Ein Verzeichnis, in das sich nicht schreiben lässt, wird gemeldet** (§8)
// — beim Ablegen wie beim Wegnehmen.
func TestEineUnbeschreibbareAblageWirdGemeldet(t *testing.T) {
	t.Parallel()

	ort := t.TempDir()
	nutzerordner := filepath.Join(ort, "nutzer")

	if fehler := os.MkdirAll(nutzerordner, 0o750); fehler != nil {
		t.Fatalf("anlegen: %v", fehler)
	}

	// **Erst ein Bild ablegen, dann zusperren** — sonst gäbe es beim
	// Wegnehmen nichts wegzunehmen, und das ist kein Fehler.
	anwendung := &App{bilderort: func() (string, error) { return ort, nil }}
	pfad := filepath.Join(t.TempDir(), "bild.png")
	schreibeProbebild(t, pfad)
	anwendung.tischbildSetzen(t.Context(), "7", pfad)

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

	if antwort := anwendung.tischbildSetzen(t.Context(), "7", pfad); antwort.Fehler == nil {
		t.Error("das Ablegen in ein zugesperrtes Verzeichnis kam ohne Befund")
	}

	if antwort := anwendung.TischbildEntfernen(t.Context(), "7"); antwort.Fehler == nil {
		t.Error("das Wegnehmen aus einem zugesperrten Verzeichnis kam ohne Befund")
	}
}

// schreibeProbebild legt ein winziges PNG ab.
func schreibeProbebild(t *testing.T, pfad string) {
	t.Helper()

	datei, fehler := os.Create(pfad) //#nosec G304 -- ein Pfad aus t.TempDir()
	if fehler != nil {
		t.Fatalf("anlegen: %v", fehler)
	}

	defer func() { _ = datei.Close() }() //nolint:errcheck // geschrieben ist geschrieben

	if fehler := png.Encode(datei, tischprobe()); fehler != nil {
		t.Fatalf("kodieren: %v", fehler)
	}
}

// **Der Dialog wird gefragt, und ein Abbruch ist keine Änderung.**
//
// Wer ihn schließt, hat nichts gewählt; eine Meldung darüber wäre ein Vorwurf
// für einen Handgriff, den jemand absichtlich getan hat.
func TestEinAbgebrochenerDialogLaesstAllesStehen(t *testing.T) {
	t.Parallel()

	anwendung, _ := tischapp(t)
	pfad := filepath.Join(t.TempDir(), "bild.png")
	schreibeProbebild(t, pfad)
	anwendung.tischbildSetzen(t.Context(), "7", pfad)

	anwendung.dateiwahl = func(context.Context, Dateiauswahl) (string, error) {
		return "", nil
	}

	antwort := anwendung.TischbildWaehlen(t.Context(), "7")
	switch {
	case antwort.Fehler != nil:
		t.Fatalf("Fehler: %v", antwort.Fehler)
	case antwort.Adresse != Tischbildpfad+"7.png":
		t.Errorf("Adresse %q — das vorige Bild sollte stehen bleiben", antwort.Adresse)
	}
}

// **Was der Dialog nennt, wird abgelegt.**
func TestDerDialogLegtAb(t *testing.T) {
	t.Parallel()

	anwendung, _ := tischapp(t)
	pfad := filepath.Join(t.TempDir(), "bild.png")
	schreibeProbebild(t, pfad)

	anwendung.dateiwahl = func(context.Context, Dateiauswahl) (string, error) {
		return pfad, nil
	}

	antwort := anwendung.TischbildWaehlen(t.Context(), "7")
	switch {
	case antwort.Fehler != nil:
		t.Fatalf("Fehler: %v", antwort.Fehler)
	case antwort.Adresse != Tischbildpfad+"7.png":
		t.Errorf("Adresse %q", antwort.Adresse)
	}

	// **Und danach sagt es auch der Leseweg** — der Erfolgsweg von
	// [App.Tischbild], den bis hierher kein Lauf betrat.
	if gelesen := anwendung.Tischbild(t.Context(), "7"); gelesen.Adresse == "" {
		t.Error("das abgelegte Bild wird nicht gefunden")
	}
}

// **Ein Dialog, der klagt, wird gemeldet.**
func TestEinKlagenderDialogWirdGemeldet(t *testing.T) {
	t.Parallel()

	anwendung, _ := tischapp(t)
	anwendung.dateiwahl = func(context.Context, Dateiauswahl) (string, error) {
		return "", os.ErrPermission
	}

	if antwort := anwendung.TischbildWaehlen(t.Context(), "7"); antwort.Fehler == nil {
		t.Error("ein klagender Dialog kam ohne Befund")
	}
}
