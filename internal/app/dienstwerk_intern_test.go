package app

// **Wie ein Dienstposten ein Programm startet und seine Ansage liest** —
// Werkstück N6 von Phase 5.4
// (#1572).
//
// # Mit einem echten Prozess und nicht mit einem Doppel
//
// **Weil die Zusagen hier welche über einen Prozess sind**: dass die Ansage
// wirklich über ein Rohr ankommt, dass ein Beenden das Kind erreicht, und dass
// ein fehlendes Programm als Fehler und nicht als Warten endet. Ein Doppel
// belegte davon keine einzige.
//
// **Das Programm ist ein Skript** — geschrieben in `t.TempDir()`, unter dem
// Namen, den der Quelltext nennt. Was gemessen wird, ist der Weg dorthin und
// zurück; **was** `mse-rest` dann bedient, misst `internal/flaechen`.

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// imOrt sind die echten Nähte mit gesetztem Ort und großzügiger Frist.
//
// # Sechzig Sekunden, nicht zwei
//
// **Die Frist ist hier ein Netz gegen ein Hängen und keine Messung** —
// derselbe Satz wie bei `warteAufAnsage` in `cmd/mse-rest`. Zwei Sekunden
// reichten allein und nicht unter Last: am 2026-09-15 fiel
// `TestEinSterbenderDienstHinterlaesstEinenBefund` um, während die ganze
// Suite nebenher lief — `/bin/sh` zu starten und seine Ausgabe zu lesen
// dauerte länger als die Frist.
//
// **Wer die Frist selbst misst, setzt sie ausdrücklich** (siehe
// `TestEinStummesProgrammLaeuftInDieFrist`); für alle anderen ist eine kurze
// Frist ein Wettlauf mit der Maschine, den die Maschine gewinnt.
func imOrt(ort string) dienstnaehte {
	naehte := echteDienstnaehte()
	naehte.orten = func() (string, error) { return ort, nil }
	naehte.frist = 60 * time.Second

	return naehte
}

// einProgramm legt ein ausführbares Skript unter `name` ab.
func einProgramm(t *testing.T, ort, name, rumpf string) {
	t.Helper()

	pfad := filepath.Join(ort, name)
	skript := []byte("#!/bin/sh\n" + rumpf)

	// **0700, weil es laufen soll** — ohne Ausführungsrecht startet `exec` es
	// gar nicht. Nur dieser Nutzer darf es, und das ist die engste Fassung,
	// die noch tut, was hier gebraucht wird.
	//nolint:gosec // siehe darüber
	if fehler := os.WriteFile(pfad, skript, 0o700); fehler != nil {
		t.Fatalf("%s schreiben: %v", name, fehler)
	}
}

// **Ein echtes Programm, eine echte Ansage, ein echtes Ende.**
func TestDasProgrammDanebenWirdGefahrenUndGelesen(t *testing.T) {
	t.Parallel()

	ort := t.TempDir()

	// **Es hält die Standardeingabe** — so wie `mse-rest` es tut. `read`
	// endet, sobald das Rohr zugeht, und genau das ist die Elternwache.
	einProgramm(t, ort, "mse-rest", `
echo "Argumente $*"
echo "Dienst auf https://0.0.0.0:54591"
echo "Fingerabdruck AB:CD:EF"
echo "Marke geheim"
read -r _
`)

	lauf, fehler := starteDienstprogrammUeber(context.Background(), imOrt(ort),
		domain.DienstREST,
		domain.Bindung{Adresse: domain.JedeSchnittstelle, Port: 54591})
	if fehler != nil {
		t.Fatalf("starteDienstprogrammUeber: %v", fehler)
	}

	if lauf.adresse != "https://0.0.0.0:54591" || lauf.marke != "geheim" ||
		lauf.fingerabdruck != "AB:CD:EF" {
		t.Errorf("die Ansage kam unvollständig an: %+v", lauf)
	}

	// **Und das Beenden erreicht das Kind wirklich** — ohne das Rohr liefe
	// `read` weiter, und `<-fertig` käme nie zurück. Ein Ablauf dieses Tests
	// ist der Fehlschlag.
	lauf.beenden()
	<-lauf.fertig
}

// **Die Bindung kommt als zwei Fahnen an.**
//
// **Ohne diesen Lauf stünde nur, dass ein Prozess startet** — nicht, dass er
// erfährt, wohin er sich binden soll (§10: die Reichweite einer Messung).
func TestDieBindungErreichtDasProgramm(t *testing.T) {
	t.Parallel()

	ort := t.TempDir()

	einProgramm(t, ort, "mse-mcp", `
echo "Dienst auf http://$2:$4/mcp"
echo "Marke geheim"
read -r _
`)

	lauf, fehler := starteDienstprogrammUeber(context.Background(), imOrt(ort),
		domain.DienstMCP,
		domain.Bindung{Adresse: domain.Schleifenadresse, Port: 8123})
	if fehler != nil {
		t.Fatalf("starteDienstprogrammUeber: %v", fehler)
	}

	t.Cleanup(func() { lauf.beenden(); <-lauf.fertig })

	if lauf.adresse != "http://127.0.0.1:8123/mcp" {
		t.Errorf("die Fahnen kamen als %q an", lauf.adresse)
	}

	// **Und ohne TLS bleibt der Abdruck leer** statt einer leeren Zeile.
	if lauf.fingerabdruck != "" {
		t.Errorf("woher kommt der Fingerabdruck? %q", lauf.fingerabdruck)
	}
}

// **Ein fehlendes Programm endet als Fehler, nicht als Warten.**
func TestOhneProgrammKommtEinFehler(t *testing.T) {
	t.Parallel()

	_, fehler := starteDienstprogrammUeber(context.Background(), imOrt(t.TempDir()),
		domain.DienstREST, domain.Bindung{Adresse: domain.Schleifenadresse})
	if fehler == nil {
		t.Fatal("ein Programm, das es nicht gibt, lief")
	}

	if !strings.Contains(fehler.Error(), "mse-rest") {
		t.Errorf("der Fehler nennt nicht, was fehlt: %v", fehler)
	}
}

// **Eine Dienstart, die es nicht gibt, fährt gar nichts.**
func TestEineUnbekannteDienstartFaehrtNichts(t *testing.T) {
	t.Parallel()

	ort := t.TempDir()

	einProgramm(t, ort, "mse-rest", "echo sollte nicht laufen")
	einProgramm(t, ort, "mse-mcp", "echo sollte nicht laufen")

	_, fehler := starteDienstprogrammUeber(context.Background(), imOrt(ort),
		domain.Dienstart("gopher"),
		domain.Bindung{Adresse: domain.Schleifenadresse})
	if !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("eine erfundene Art gab %v", fehler)
	}
}

// **Ein Programm, das nichts sagt, hält das Fenster nicht fest.**
func TestEinStummesProgrammLaeuftInDieFrist(t *testing.T) {
	t.Parallel()

	ort := t.TempDir()

	einProgramm(t, ort, "mse-rest", "read -r _")

	angefangen := time.Now()

	knapp := imOrt(ort)
	knapp.frist = 200 * time.Millisecond

	_, fehler := starteDienstprogrammUeber(context.Background(), knapp,
		domain.DienstREST, domain.Bindung{Adresse: domain.Schleifenadresse})
	if !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("ein stummes Programm gab %v", fehler)
	}

	// **Die Gegenprobe zur Frist**: ohne sie wäre nicht zu sehen, ob wirklich
	// sie es war oder ob das Kind einfach sofort starb.
	if gedauert := time.Since(angefangen); gedauert < 200*time.Millisecond {
		t.Errorf("die Frist wurde nicht abgewartet: %s", gedauert)
	}
}

// **Ein Programm, das sofort stirbt, endet als Fehler.**
//
// **Der andere Weg durch [leseAnsage]** — hier endet der Strom, statt dass
// die Frist abläuft.
func TestEinSofortSterbendesProgrammEndetAlsFehler(t *testing.T) {
	t.Parallel()

	ort := t.TempDir()

	einProgramm(t, ort, "mse-mcp", "echo 'Dienst auf http://127.0.0.1:1'")

	_, fehler := starteDienstprogrammUeber(context.Background(), imOrt(ort),
		domain.DienstMCP, domain.Bindung{Adresse: domain.Schleifenadresse})
	if !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("eine halbe Ansage gab %v", fehler)
	}
}

// **Ein abgebrochener Kontext wartet die Frist nicht ab.**
func TestEinAbgebrochenerRufWartetNicht(t *testing.T) {
	t.Parallel()

	ort := t.TempDir()

	einProgramm(t, ort, "mse-rest", "read -r _")

	ctx, abbrechen := context.WithCancel(context.Background())
	abbrechen()

	geduldig := imOrt(ort)
	geduldig.frist = time.Hour

	_, fehler := starteDienstprogrammUeber(ctx, geduldig, domain.DienstREST,
		domain.Bindung{Adresse: domain.Schleifenadresse})
	if !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("der Abbruch gab %v", fehler)
	}
}

// **Ohne eingesetzte Naht fährt der Posten das Programm daneben.**
//
// **Die Gegenprobe zu [App.MitDienststarter]**: ohne sie bliebe offen, ob die
// Naht den Betriebsfall überhaupt vertritt oder ob dort etwas anderes steht.
func TestOhneNahtGiltDasProgrammDaneben(t *testing.T) {
	t.Parallel()

	anwendung := NeueApp("")

	if reflektiert(anwendung.dienststarten()) !=
		reflektiert(Dienststarter(starteDienstprogramm)) {
		t.Error("ohne Naht wird etwas anderes gefahren")
	}

	gesetzt := anwendung.MitDienststarter(func(
		context.Context, domain.Dienstart, []domain.Bindung,
	) (dienstlauf, error) {
		return dienstlauf{}, nil
	})

	if gesetzt != anwendung {
		t.Error("MitDienststarter gibt eine andere Anwendung zurück")
	}

	if reflektiert(anwendung.dienststarten()) ==
		reflektiert(Dienststarter(starteDienstprogramm)) {
		t.Error("die gesetzte Naht kommt nicht an")
	}
}

// **Und ohne eigenes Programm gibt es keinen Ort.**
func TestOhneEigenesProgrammGibtEsKeinenOrt(t *testing.T) {
	t.Parallel()

	_, fehler := werkzeugortNeben(func() (string, error) {
		return "", errors.New("kein Pfad")
	})
	if !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("ein Fehlschlag gab %v", fehler)
	}

	ort, fehler := Werkzeugort()
	if fehler != nil || ort == "" {
		t.Errorf("Werkzeugort: %q, %v", ort, fehler)
	}
}

// reflektiert gibt die Stelle, an der eine Funktion steht.
//
// **Weil sich Funktionswerte nicht vergleichen lassen** — Go verbietet `==`
// darauf, und ohne diesen Umweg ließe sich „es ist der echte" gar nicht sagen.
func reflektiert(was Dienststarter) uintptr {
	return reflect.ValueOf(was).Pointer()
}

// **Ohne Ort wird nichts gestartet.**
//
// **Der Zweig, den [Werkzeugort] im Betrieb nicht betritt** — `os.Executable`
// scheitert unter macOS nicht. Dass er trotzdem dasteht, ist richtig; dass
// niemand ihn je gesehen hätte, wäre es nicht (§10).
func TestOhneOrtWirdNichtsGestartet(t *testing.T) {
	t.Parallel()

	ohneOrt := echteDienstnaehte()
	ohneOrt.orten = func() (string, error) { return "", errors.New("kein Pfad") }

	_, fehler := starteDienstprogrammUeber(context.Background(), ohneOrt,
		domain.DienstREST, domain.Bindung{Adresse: domain.Schleifenadresse})
	if fehler == nil {
		t.Fatal("ohne Ort lief etwas")
	}
}
