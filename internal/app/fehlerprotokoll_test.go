// **Der Fehler, den die Oberfläche nicht bekommt, geht ins Protokoll** —
// #857.
//
// # Warum genau dieser
//
// `Uebersetze` ersetzt einen unbekannten Fehler durch den Rückfallsatz: „SQL
// logic error: no such table: cards" sagt Nutzenden nichts und verrät den
// inneren Aufbau (§8). **Damit ist er weg** — und das ist genau die Sorte
// Fehler, von der man später nur im Protokoll erfährt.
//
// Die anderen Zweige nicht: eine Validierung, ein fehlender Datensatz, ein
// Konflikt und ein Abbruch sind erwartete Ausgänge, und ein Protokoll, das sie
// mitschreibt, verdeckt den einen Eintrag, um dessentwillen es da ist.
package app_test

import (
	"bytes"
	"errors"
	"log/slog"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// mitProtokoll setzt für die Dauer eines Tests einen mitschreibenden Standard.
//
// **Nicht parallel.** `slog.SetDefault` ist ein Prozesszustand; zwei Tests, die
// ihn zugleich setzen, prüften einander.
func mitProtokoll(t *testing.T) *bytes.Buffer {
	t.Helper()

	var gesammelt bytes.Buffer

	vorher := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&gesammelt, nil)))
	t.Cleanup(func() { slog.SetDefault(vorher) })

	return &gesammelt
}

func TestEinUnbekannterFehlerStehtImProtokoll(t *testing.T) {
	gesammelt := mitProtokoll(t)

	fehler := app.Uebersetze(t.Context(), errors.New("SQL logic error: no such table: cards"))

	if fehler.Code != app.CodeUnbekannt {
		t.Fatalf("Code %q, erwartet war %q", fehler.Code, app.CodeUnbekannt)
	}

	// **Die Oberfläche bekommt den Rückfallsatz** — das bleibt so.
	if strings.Contains(fehler.Meldung, "no such table") {
		t.Errorf("die rohe Meldung erreicht die Oberfläche: %q", fehler.Meldung)
	}

	// **Und das Protokoll bekommt den Grund.**
	if !strings.Contains(gesammelt.String(), "no such table") {
		t.Errorf("der Grund fehlt im Protokoll:\n%s", gesammelt.String())
	}
}

// **Die erwarteten Ausgänge stehen nicht darin.**
//
// Ein Protokoll, das jede Eingabeprüfung mitschreibt, verdeckt den einen
// Eintrag, um dessentwillen es da ist.
func TestErwarteteFehlerStehenNichtImProtokoll(t *testing.T) {
	for name, fehler := range map[string]error{
		"Validierung":    domain.ErrValidation,
		"nicht gefunden": domain.ErrNotFound,
		"Konflikt":       domain.ErrConflict,
	} {
		t.Run(name, func(t *testing.T) {
			gesammelt := mitProtokoll(t)

			app.Uebersetze(t.Context(), fehler)

			if gesammelt.Len() != 0 {
				t.Errorf("es wurde protokolliert:\n%s", gesammelt.String())
			}
		})
	}
}

// **Und ein verdeckter Grund geht ins Protokoll, obwohl der Ausgang erwartet
// ist** — Werkstück 1 von Phase 4.13
// (#1352).
//
// # Warum das kein Widerspruch zum Lauf darüber ist
//
// **Der Absatz darüber sagt: erwartete Ausgänge werden nicht mitgeschrieben.**
// Das gilt weiter — mitgeschrieben wird nicht der Ausgang, sondern **der
// Grund, den es sonst nirgends mehr gibt**. Eine Validierung ohne Grund
// erzeugt weiterhin keine Zeile.
//
// **Wer einen setzt, hat ihn absichtlich aus dem Satz genommen** ([domain.MitGrund]).
// Ohne diese Zeile wäre er dann fort — und die Trennung machte aus einer
// unschönen Meldung eine unauffindbare Ursache.
func TestEinVerdeckterGrundStehtImProtokoll(t *testing.T) {
	gesammelt := mitProtokoll(t)

	fehler := app.Uebersetze(t.Context(), domain.MitGrund(
		domain.ErrValidation,
		errors.New("open /Users/wer/bild.png: no such file or directory"),
		"die Bilddatei ließ sich nicht öffnen"))

	if fehler.Code != app.CodeValidierung {
		t.Fatalf("Code %q, erwartet war %q", fehler.Code, app.CodeValidierung)
	}

	// **Der Satz bleibt der für Menschen** — kein Pfad, kein Go-Fehler (§8).
	if strings.Contains(fehler.Meldung, "no such file") {
		t.Errorf("der Grund erreicht die Oberfläche: %q", fehler.Meldung)
	}

	if !strings.Contains(fehler.Meldung, "die Bilddatei ließ sich nicht öffnen") {
		t.Errorf("der eigene Satz ging verloren: %q", fehler.Meldung)
	}

	// **Und der Grund steht im Protokoll.**
	if !strings.Contains(gesammelt.String(), "no such file") {
		t.Errorf("der Grund fehlt im Protokoll:\n%s", gesammelt.String())
	}
}

// **Ohne verdeckten Grund bleibt das Protokoll leer** — die Gegenprobe zum
// Lauf darüber.
func TestEineValidierungOhneGrundBleibtUnprotokolliert(t *testing.T) {
	gesammelt := mitProtokoll(t)

	app.Uebersetze(t.Context(), domain.MitGrund(domain.ErrValidation, nil, "der Name fehlt"))

	if gesammelt.Len() != 0 {
		t.Errorf("ein erwarteter Ausgang wurde mitgeschrieben:\n%s", gesammelt.String())
	}
}
