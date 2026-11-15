package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"slices"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/laufbeleg"
)

// miss fährt den Lauf und schreibt seinen Beleg.
//
// **Fehlt eine Pflichteingabe, wird nicht gemessen** — der Beleg sagt
// BLOCKED, und `go test` startet gar nicht erst. Genau hier ist #1656
// durchgefallen: die Läufe starteten, fanden nichts und sprangen.
func miss(ctx context.Context, plan laufbeleg.Plan, planpfad string, zu *anzeige,
	kasten werkzeugkasten,
) int {
	eingaben, fehlt := erhebe(plan, kasten)
	argumente := testargumente(plan)
	vorher := standLesen(ctx, kasten, plan.Beleg)

	// **Ohne Inhaltshash wird der Beleg rot** — `Belegen` sagt es; gemessen
	// wird trotzdem, denn die Messung selbst hängt nicht an ihm.
	inhalt, _ := inhaltLesen(ctx, kasten, plan.Misst, planpfad) //nolint:errcheck // ein leerer Hash macht den Beleg rot

	kopf := laufbeleg.Kopf{
		Lauf: plan.Lauf, Commit: vorher.commit, Inhalt: inhalt, Anfang: kasten.jetzt(),
		Befehl: append([]string{"go"}, argumente...),
		Werkzeuge: map[string]string{
			"go": kasten.version(ctx), "system": kasten.system(ctx),
			"plattform": runtime.GOOS + "/" + runtime.GOARCH,
		},
		Schalter: schalterLesen(plan, kasten),
	}

	zu.sag("Lauf %s — %d Fälle erwartet, Stand %s\n",
		plan.Lauf, len(plan.Erwartung.Faelle), vorher.benannt())

	for _, eingabe := range eingaben {
		zeigeEingabe(zu, eingabe)
	}

	var (
		ereignisse []laufbeleg.Ereignis
		artefakte  []laufbeleg.Eingabe
	)

	if !fehlt && len(plan.Erwartung.Faelle) > 0 {
		var fehler error

		ereignisse, kopf.Rueckgabe, fehler = fahreTests(ctx, plan.Roh, argumente, zu, kasten)
		if fehler != nil {
			zu.klage("lauf: %v\n", fehler)

			return 2
		}

		artefakte = append(artefakte, laufbeleg.Erhebe("Rohprotokoll", plan.Roh, ""))
	}

	nachher := standLesen(ctx, kasten, plan.Beleg)
	kopf.Ende = kasten.jetzt()
	kopf.BaumSauber = vorher.sauber && vorher == nachher

	beleg := laufbeleg.Belegen(kopf, eingaben,
		laufbeleg.Auswerten(ereignisse, plan.Erwartung, kopf.Rueckgabe), artefakte)

	if fehler := beleg.SchreibeNach(plan.Beleg); fehler != nil {
		zu.klage("lauf: %v\n", fehler)

		return 2
	}

	zusammenfassen(zu, plan.Beleg, beleg)

	if beleg.Status == laufbeleg.Bestanden {
		return 0
	}

	return 1
}

// erhebe prüft die Pflichteingaben und sagt, ob eine fehlt.
func erhebe(plan laufbeleg.Plan, kasten werkzeugkasten) ([]laufbeleg.Eingabe, bool) {
	zuhause := kasten.umgebung("HOME")
	eingaben := make([]laufbeleg.Eingabe, 0, len(plan.Eingaben))
	fehlt := false

	for _, wunsch := range plan.Eingaben {
		eingabe := laufbeleg.Erhebe(wunsch.Name, kasten.umgebung(wunsch.Name), wunsch.Muster)
		eingabe.Pfad = ohneZuhause(eingabe.Pfad, zuhause)
		fehlt = fehlt || eingabe.Fehlt != ""
		eingaben = append(eingaben, eingabe)
	}

	return eingaben, fehlt
}

// schalterLesen gibt die Schalter des Plans mit ihrem Wert — **leer, wo einer
// nicht gesetzt ist**, damit ein späterer Lauf, der ihn setzt, anders heißt.
func schalterLesen(plan laufbeleg.Plan, kasten werkzeugkasten) map[string]string {
	werte := make(map[string]string, len(plan.Schalter))
	for _, name := range plan.Schalter {
		werte[name] = ohneZuhause(kasten.umgebung(name), kasten.umgebung("HOME"))
	}

	return werte
}

// ohneZuhause ersetzt das Zuhause durch `~` — **der Beleg wird eingecheckt**,
// und wer ihn liest, braucht den Namen dieses Kontos nicht.
func ohneZuhause(pfad, zuhause string) string {
	if zuhause != "" && (pfad == zuhause || strings.HasPrefix(pfad, zuhause+string(filepath.Separator))) {
		return "~" + strings.TrimPrefix(pfad, zuhause)
	}

	return pfad
}

// testargumente baut den Aufruf von `go test`.
//
// **`-count=1`, denn ein Treffer aus dem Zwischenspeicher ist keine Messung.**
// Und nur die erwarteten Fälle, in den Paketen, in denen sie stehen: was
// sonst läuft, belegt dieser Lauf nicht.
func testargumente(plan laufbeleg.Plan) []string {
	namen := make([]string, 0, len(plan.Erwartung.Faelle))
	pakete := make([]string, 0, len(plan.Erwartung.Faelle))

	for _, einer := range plan.Erwartung.Faelle {
		namen = append(namen, regexp.QuoteMeta(einer.Test))
		pakete = append(pakete, einer.Paket)
	}

	slices.Sort(namen)
	slices.Sort(pakete)

	argumente := []string{
		"test", "-json", "-count=1", "-tags", plan.Tag,
		"-run", "^(" + strings.Join(slices.Compact(namen), "|") + ")$",
	}

	if plan.Zeitlimit > 0 {
		argumente = append(argumente, "-timeout", plan.Zeitlimit.String())
	}

	return append(argumente, slices.Compact(pakete)...)
}

// fahreTests startet `go test`, schreibt das Rohprotokoll und liest den Strom.
//
// **Ein Fehler hier heißt: das Protokoll ließ sich nicht anlegen.** Dass
// `go test` nicht startet oder mitten im Lauf abbricht, ist dagegen ein
// Ergebnis — Rückgabe -1, und der Beleg sagt es.
func fahreTests(ctx context.Context, rohpfad string, argumente []string, zu *anzeige,
	kasten werkzeugkasten,
) ([]laufbeleg.Ereignis, int, error) {
	if fehler := os.MkdirAll(filepath.Dir(rohpfad), 0o750); fehler != nil {
		return nil, 0, fmt.Errorf("das Rohprotokoll %s anlegen: %w", rohpfad, fehler)
	}

	roh, fehler := os.Create(rohpfad) // #nosec G304 -- den Ort nennt der Plan
	if fehler != nil {
		return nil, 0, fmt.Errorf("das Rohprotokoll %s anlegen: %w", rohpfad, fehler)
	}

	var strom, nebenausgabe bytes.Buffer

	fortschrittsanzeige := &fortschritt{ziel: zu.aus, modul: zu.modul}

	rueckgabe, startfehler := kasten.gotest(ctx, argumente,
		io.MultiWriter(roh, &strom, fortschrittsanzeige), &nebenausgabe)
	fortschrittsanzeige.Schliessen()

	if startfehler = errors.Join(startfehler, roh.Close()); startfehler != nil {
		zu.klage("lauf: go test lief nicht zu Ende: %v\n", startfehler)
		rueckgabe = -1
	}

	if nebenausgabe.Len() > 0 {
		zu.klage("go test schrieb daneben:\n%s", nebenausgabe.String())
	}

	ereignisse, fremd, lesefehler := laufbeleg.LiesEreignisse(&strom)
	if lesefehler != nil {
		zu.klage("lauf: der Strom ist nicht zu Ende zu lesen: %v\n", lesefehler)
		rueckgabe = -1
	}

	for _, zeile := range fremd {
		zu.klage("keine Zeile von go test: %s\n", zeile)
	}

	return ereignisse, rueckgabe, nil
}
