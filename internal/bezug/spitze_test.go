package bezug_test

// **Wie weit eine Quelle ist** — G04
// (#2251).
//
// Bezogen wird an einem Commit, und der Stand veraltet nicht von selbst —
// aber niemand erfuhr, dass es einen neueren gibt (`docs/PARITAET.md`,
// „Paketaktualisierungen anzeigen"). [bezug.Spitze] löst den Hauptzweig zu
// **einem Commit** auf, [bezug.Vergleiche] sagt, welche Dateien sich zwischen
// zwei Commits ändern. Beide auf Aufruf, nie von selbst.

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/bezug"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
)

const spitzencommit = "0123456789abcdef0123456789abcdef01234567"

func antwortender(t *testing.T, status int, rumpf string) *httptest.Server {
	t.Helper()

	dienst := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(status)

		if _, fehler := w.Write([]byte(rumpf)); fehler != nil {
			t.Errorf("antworten: %v", fehler)
		}
	}))
	t.Cleanup(dienst.Close)

	return dienst
}

func TestDieSpitzeIstEinCommit(t *testing.T) {
	dienst := antwortender(t, http.StatusOK,
		`{"sha":"`+spitzencommit+`","commit":{"committer":{"date":"2026-10-01T10:00:00Z"}},"weiteres":1}`)

	spitze, fehler := bezug.Spitze(t.Context(), dienst.URL, "probe")
	if fehler != nil {
		t.Fatalf("Spitze: %v", fehler)
	}

	if spitze.Commit != spitzencommit || !spitze.Zeit.Equal(time.Date(2026, 10, 1, 10, 0, 0, 0, time.UTC)) {
		t.Errorf("Spitze %+v", spitze)
	}
}

func TestDieSpitzeScheitertBenannt(t *testing.T) {
	faelle := []struct {
		name   string
		status int
		rumpf  string
		feld   string
	}{
		{"ein Zweigname statt eines Commits bricht laut", http.StatusOK, `{"sha":"main","commit":{"committer":{"date":"2026-10-01T10:00:00Z"}}}`, "adresse"},
		{"ohne Datum bricht laut", http.StatusOK, `{"sha":"` + spitzencommit + `"}`, "adresse"},
		{"ein unbekanntes Repositorium", http.StatusNotFound, `{}`, "adresse"},
		{"eine gedrosselte Gegenstelle", http.StatusTooManyRequests, `{}`, "adresse"},
		{"kein JSON", http.StatusOK, `<html>`, "adresse"},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			_, fehler := bezug.Spitze(t.Context(), antwortender(t, fall.status, fall.rumpf).URL, "probe")

			var feldfehler *domain.FieldError
			if !errors.As(fehler, &feldfehler) || feldfehler.Field != fall.feld {
				t.Errorf("Fehler %v, erwartet am Feld %q", fehler, fall.feld)
			}
		})
	}
}

func TestEineUnerreichbareQuelleIstEineGegenstelle(t *testing.T) {
	dienst := antwortender(t, http.StatusOK, `{}`)
	dienst.Close()

	if _, fehler := bezug.Spitze(t.Context(), dienst.URL, "probe"); !errors.Is(fehler, domain.ErrGegenstelle) {
		t.Errorf("Fehler %v, erwartet war eine unerreichbare Gegenstelle", fehler)
	}
}

func TestDerVergleichNenntDieDateien(t *testing.T) {
	dienst := antwortender(t, http.StatusOK, `{"status":"ahead","ahead_by":3,"behind_by":0,"files":[
		{"filename":"data/magic-m15.mse-style/style","status":"modified"},
		{"filename":"data/magic-neu.mse-style/style","status":"added"},
		{"filename":"README.md","status":"removed"}]}`)

	vergleich, fehler := bezug.Vergleiche(t.Context(), dienst.URL, "probe")
	if fehler != nil {
		t.Fatalf("Vergleiche: %v", fehler)
	}

	if vergleich.Lage != "ahead" || vergleich.Vorsprung != 3 || vergleich.Gekuerzt || len(vergleich.Dateien) != 3 ||
		vergleich.Dateien[1] != (bezug.Aenderung{Datei: "data/magic-neu.mse-style/style", Art: "added"}) {
		t.Errorf("Vergleich %+v", vergleich)
	}
}

// **Dreihundert Dateien sind die Grenze der Gegenstelle** — mehr nennt sie
// nicht. Eine Vorschau, die das verschwiege, sähe vollständig aus.
func TestEinVergleichMitDreihundertDateienIstGekuerzt(t *testing.T) {
	dateien := make([]string, 300)
	for i := range dateien {
		dateien[i] = `{"filename":"data/a.mse-style/b","status":"modified"}`
	}

	dienst := antwortender(t, http.StatusOK, `{"status":"ahead","ahead_by":40,"files":[`+strings.Join(dateien, ",")+`]}`)

	vergleich, fehler := bezug.Vergleiche(t.Context(), dienst.URL, "probe")
	if fehler != nil || !vergleich.Gekuerzt {
		t.Errorf("Vergleich gekürzt=%v, Fehler %v", vergleich.Gekuerzt, fehler)
	}
}

func TestEinVergleichOhneLageBrichtLaut(t *testing.T) {
	_, fehler := bezug.Vergleiche(t.Context(), antwortender(t, http.StatusOK, `{"files":[]}`).URL, "probe")

	var feldfehler *domain.FieldError
	if !errors.As(fehler, &feldfehler) || feldfehler.Field != "adresse" {
		t.Errorf("Fehler %v, erwartet am Feld adresse", fehler)
	}
}

// **Ohne den Schalter der Läufe ist die Schleife gesperrt** — dieselbe Sperre
// wie beim Bezug (#2000): wer die Adresse nennt, erreicht damit nicht diesen
// Rechner.
func TestEinGesperrtesZielIstEinFeldbefund(t *testing.T) {
	dienst := antwortender(t, http.StatusOK, `{}`)
	t.Setenv(netz.SchleifeUmgebung, "")

	for name, ruf := range map[string]func() error{
		"Spitze":     func() error { _, f := bezug.Spitze(t.Context(), dienst.URL, "probe"); return f },
		"Vergleiche": func() error { _, f := bezug.Vergleiche(t.Context(), dienst.URL, "probe"); return f },
	} {
		var feldfehler *domain.FieldError
		if fehler := ruf(); !errors.As(fehler, &feldfehler) || feldfehler.Field != "adresse" {
			t.Errorf("%s: Fehler %v, erwartet am Feld adresse", name, fehler)
		}
	}
}

func TestEinVergleichScheitertWieDieSpitze(t *testing.T) {
	_, fehler := bezug.Vergleiche(t.Context(), antwortender(t, http.StatusNotFound, `{}`).URL, "probe")

	var feldfehler *domain.FieldError
	if !errors.As(fehler, &feldfehler) {
		t.Errorf("Fehler %v, erwartet am Feld adresse", fehler)
	}
}

// **Eine Umbenennung nennt ihren vorigen Namen** — Befund 3 des Zweitblicks
// über Phase G, zweiter Teil: ohne ihn verschwand ein Paket, aus dem eine
// Datei in ein anderes wanderte, aus der Vorschau.
func TestEineUmbenennungNenntIhrenVorigenNamen(t *testing.T) {
	dienst := antwortender(t, http.StatusOK, `{"status":"ahead","ahead_by":1,"files":[
		{"filename":"anders.mse-style/style","status":"renamed","previous_filename":"benutzt.mse-style/style"}]}`)

	vergleich, fehler := bezug.Vergleiche(t.Context(), dienst.URL, "probe")
	if fehler != nil || len(vergleich.Dateien) != 1 ||
		vergleich.Dateien[0] != (bezug.Aenderung{Datei: "anders.mse-style/style", Art: "renamed", Vorher: "benutzt.mse-style/style"}) {
		t.Errorf("Vergleich %+v, Fehler %v", vergleich, fehler)
	}
}
