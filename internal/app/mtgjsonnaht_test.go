// **Die MTG-JSON-Naht ohne Fenster** —
// #1291.
package app_test

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// mtgjsonprobe gibt den Pfad zur eingecheckten MTG-JSON-Probedatei.
func mtgjsonprobe(t *testing.T) string {
	t.Helper()

	pfad, err := filepath.Abs(
		filepath.Join("..", "..", "testdaten", "mtgjson", "probe.json"))
	if err != nil {
		t.Fatalf("Pfad: %v", err)
	}

	return pfad
}

// **Aus einer Datei werden Sets — und die Kennungen kommen zurück.**
func TestMTGJSONLesenFaehrtEin(t *testing.T) {
	db := offeneProbedatenbank(t)

	eingefahren, fehler := app.MTGJSONLesen(probekontext(t), db,
		mtgjsonprobe(t), "magic")
	if fehler != nil {
		t.Fatalf("lesen: %v", fehler)
	}

	if len(eingefahren.Angelegt) == 0 {
		t.Fatal("es wurde kein Set angelegt")
	}

	if eingefahren.SchonOffen {
		t.Error("die erste Einfuhr galt als schon offen")
	}

	// **Die Befunde sind eine leere Liste und nicht `nil`** (§3).
	if eingefahren.Befunde == nil {
		t.Error("die Befunde sind nil statt einer leeren Liste")
	}

	// **Und ein zweiter Ruf legt nichts Neues an** — „eine Quelle, ein Set".
	nochmal, fehler := app.MTGJSONLesen(probekontext(t), db,
		mtgjsonprobe(t), "magic")
	if fehler != nil {
		t.Fatalf("zweites Lesen: %v", fehler)
	}

	if !nochmal.SchonOffen {
		t.Error("die zweite Einfuhr galt nicht als schon offen")
	}
}

// **Ein Spiel ohne Zuordnung ist eine Validierung — an beiden Wegen.**
func TestMTGJSONMitEinemUnbekanntenSpiel(t *testing.T) {
	db := offeneProbedatenbank(t)

	if _, fehler := app.MTGJSONLesen(probekontext(t), db,
		mtgjsonprobe(t), "kein-spiel"); !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("lesen: %v, erwartet ErrValidation", fehler)
	}

	if _, fehler := app.MTGJSONSchreiben(probekontext(t), db, 1, "kein-spiel",
		filepath.Join(t.TempDir(), "raus.json")); !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("schreiben: %v, erwartet ErrValidation", fehler)
	}
}

// **Und aus einem Set wird wieder MTG JSON.**
func TestMTGJSONSchreibenSchreibtEineDatei(t *testing.T) {
	db := offeneProbedatenbank(t)

	eingefahren, fehler := app.MTGJSONLesen(probekontext(t), db,
		mtgjsonprobe(t), "magic")
	if fehler != nil || len(eingefahren.Angelegt) == 0 {
		t.Fatalf("einfahren: %v", fehler)
	}

	ziel := filepath.Join(t.TempDir(), "raus.json")

	geschrieben, fehler := app.MTGJSONSchreiben(
		probekontext(t), db, eingefahren.Angelegt[0], "magic", ziel)
	if fehler != nil {
		t.Fatalf("schreiben: %v", fehler)
	}

	if geschrieben.Pfad != ziel {
		t.Errorf("der Pfad ist %q", geschrieben.Pfad)
	}

	if geschrieben.Befunde == nil {
		t.Error("die Befunde sind nil statt einer leeren Liste")
	}

	if _, err := os.Stat(ziel); err != nil {
		t.Errorf("die Datei fehlt: %v", err)
	}
}

// **Ein Pfad, der auf nichts zeigt, gibt kein Ergebnis und einen Fehler.**
//
// Der Zweig, der `nil` durchreicht: ohne ihn liefe die Naht in einen
// Nullzeiger, und zwar erst beim Aufrufer.
func TestMTGJSONLesenOhneQuelle(t *testing.T) {
	db := offeneProbedatenbank(t)

	eingefahren, fehler := app.MTGJSONLesen(probekontext(t), db,
		filepath.Join(t.TempDir(), "gibtsnicht.json"), "magic")
	if fehler == nil {
		t.Fatal("kein Fehler")
	}

	if eingefahren != nil {
		t.Error("es kam trotzdem ein Ergebnis")
	}
}

// **Die Spiele stehen im Programm und sind sortiert.**
func TestMTGJSONSpieleSindSortiert(t *testing.T) {
	spiele := app.MTGJSONSpiele()

	if len(spiele) == 0 {
		t.Fatal("kein Spiel")
	}

	if !slices.IsSorted(spiele) {
		t.Errorf("die Spiele sind nicht sortiert: %v", spiele)
	}

	if !slices.Contains(spiele, "magic") {
		t.Errorf("magic fehlt: %v", spiele)
	}
}

// **Eine Quelle ohne ein einziges Set gibt eine leere Liste und kein `nil`.**
//
// **Der Zweig ist erreichbar**, und das ist der Punkt: `Importeur.MTGJSON`
// legt `Angelegt` erst beim ersten Set an. Eine Datei mit leerem `data` —
// oder eine, deren erstes Set an der Prüfung scheitert — kommt mit einem
// Ergebnis zurück, dessen Liste `nil` ist. Über die Wails-Naht wäre das
// `undefined` statt einer leeren Liste (§3), und an der Kommandozeile
// `"angelegt": null`.
func TestMTGJSONOhneEinSingleSetGibtEineLeereListe(t *testing.T) {
	db := offeneProbedatenbank(t)

	leer := filepath.Join(t.TempDir(), "leer.json")
	if err := os.WriteFile(leer,
		[]byte(`{"meta":{"date":"2026-08-30","version":"5.2.2"},"data":{}}`), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	eingefahren, fehler := app.MTGJSONLesen(probekontext(t), db, leer, "magic")
	if fehler != nil {
		t.Fatalf("lesen: %v", fehler)
	}

	if eingefahren.Angelegt == nil {
		t.Error("Angelegt ist nil statt einer leeren Liste")
	}

	if len(eingefahren.Angelegt) != 0 {
		t.Errorf("es wurden %d Sets angelegt", len(eingefahren.Angelegt))
	}
}
