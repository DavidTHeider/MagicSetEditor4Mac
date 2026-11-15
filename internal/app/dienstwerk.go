package app

// **Wie ein Dienstposten seinen Prozess startet** — Werkstück N6 von
// Phase 5.4
// (#1572).
//
// # Warum ein Unterprozess und kein Lauscher im Fenster
//
// **Weil es die Dienste schon gibt.** `mse-rest` und `mse-mcp` sind zwei
// fertige Composition Roots mit zusammen über hundert Nähten, und beide sind
// geprüft. Ein zweiter Aufbau derselben Fläche im Fenster wäre genau das, wovor
// §2 warnt: „sonst hätte der dritte Composition Root seine eigene Fassung
// davon, und drei Wurzeln liefen auseinander."
//
// **Und weil die Netzgrenze eine Grenze bleibt.** `internal/schichten` führt,
// wer `net/http` erreichen darf; ein REST-Dienst im Fenster hieße, dass die
// ganze Bindungsfassade ins Netz darf.
//
// # Was das kostet, und was dagegen steht
//
// **Ein Unterprozess kann seinen Elternprozess überleben.** Beim gewöhnlichen
// Beenden nicht — der Posten schickt ihm ein Signal und wartet. Bei einem
// **Absturz** des Fensters schon, und dagegen hilft kein `defer`.
//
// **Deshalb hält der Posten die Standardeingabe des Kindes offen.** Stirbt das
// Fenster, schließt das Betriebssystem das Rohr, und das Kind sieht ein EOF
// und endet von selbst — `cmd/mse-rest/elternwache.go` und die gleichnamige
// Datei bei `mse-mcp`. Ohne diese fünfzehn Zeilen wäre „das Fenster zu
// schließen beendet den Dienst" eine Behauptung mit einer Ausnahme, die
// niemand sieht.

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log/slog"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Startfrist ist, wie lange ein Dienst zum Ansagen hat.
//
// **Fünf Sekunden**, und sie sind großzügig: der Prozess öffnet einen
// Bestand, migriert ihn gegebenenfalls und macht einen Lauscher auf. Wer
// länger braucht, hängt — und ein Fenster, das dann noch wartet, sieht aus,
// als sei es abgestürzt.
const Startfrist = 5 * time.Second

// dienstlauf ist ein gestarteter Dienst samt seiner Ansage.
type dienstlauf struct {
	// adresse ist, was der Dienst über sich gesagt hat — mit Schema.
	adresse string

	// marke ist der Zugang, den er verlangt.
	marke string

	// fingerabdruck ist der seines Zeugnisses — leer ohne TLS.
	//
	// **Angesagt und nicht nachgerechnet**: das Zeugnis baut das Kind, und
	// nur es weiß, was darin steht. Ein zweiter Weg zu derselben Zahl liefe
	// bei der ersten Änderung des einen dem anderen davon (§10).
	fingerabdruck string

	// aussteller und gueltigBis kennt nur ein eigenes Zeugnis (NET-03).
	aussteller string
	gueltigBis time.Time

	// adressen sind alle, auf denen er lauscht — die erste ist [adresse];
	// fehlende die gewählten ohne Lauscher (G08, #2272).
	adressen []string
	fehlende []domain.Lauschbefund

	beenden func()
	fertig  chan struct{}
}

// Dienststarter startet einen Dienst — die Naht für Läufe ohne Prozess.
type Dienststarter func(
	ctx context.Context, art domain.Dienstart, bindungen []domain.Bindung,
) (dienstlauf, error)

// MitDienststarter setzt sie — für Läufe ohne echtes Programm (§3).
//
// **Ohne diese Naht bräuchte jeder Lauf ein gebautes `mse-rest`**, und was
// das wert ist, sagt §3: eine Vorrichtung, die den Bau voraussetzt, läuft in
// der Pipeline gar nicht.
func (a *App) MitDienststarter(starte Dienststarter) *App {
	a.dienststarter = starte

	return a
}

// dienststarten gibt die eingesetzte Naht — oder die echte.
func (a *App) dienststarten() Dienststarter {
	if a.dienststarter != nil {
		return a.dienststarter
	}

	return starteDienstprogramm
}

// starteDienstprogramm startet das Programm neben diesem und liest die Ansage.
func starteDienstprogramm(
	ctx context.Context, art domain.Dienstart, bindungen []domain.Bindung,
) (dienstlauf, error) {
	return starteDienstprogrammAufAllen(ctx, echteDienstnaehte(), art, bindungen)
}

// dienstnaehte sind die drei Stellen, an denen dieser Weg die Maschine
// anfasst.
//
// **Gereicht und nicht gerufen** — dieselbe Gestalt wie [technikangabenNeben]
// und aus demselben Grund: unter `go test` liegt neben dem Testprogramm kein
// `mse-rest`, und die zwei Rohre scheitern auf diesem Rechner nie. Ohne diese
// Nähte bliebe beides ungemessen, und §3 nennt eine unbequeme Verzweigung
// ausdrücklich keinen Grund für eine Ausnahme.
type dienstnaehte struct {
	// orten sagt, wo die Programme liegen.
	orten func() (string, error)

	// rohre macht Ein- und Ausgang zum Kind.
	rohre func(*exec.Cmd) (io.WriteCloser, io.ReadCloser, error)

	// frist ist, wie lange der Dienst zum Ansagen hat.
	frist time.Duration
}

// echteDienstnaehte sind die, die im Betrieb gelten.
func echteDienstnaehte() dienstnaehte {
	return dienstnaehte{
		orten: Werkzeugort,
		rohre: rohreAufmachen,
		frist: Startfrist,
	}
}

// rohreAufmachen gibt die Standardeingabe und die Standardausgabe des Kindes.
//
// **Beide zusammen und nicht einzeln**: sie scheitern aus demselben Grund —
// wenn jemand `Stdin` oder `Stdout` schon gesetzt hat —, und zwei
// Fehlerwege für einen Grund wären zwei Sätze, die dasselbe sagen.
func rohreAufmachen(
	befehl *exec.Cmd,
) (io.WriteCloser, io.ReadCloser, error) {
	eingabe, fehler := befehl.StdinPipe()
	if fehler != nil {
		return nil, nil, domain.MitGrund(domain.ErrValidation, fehler,
			"die Eingabe des Dienstes aufmachen")
	}

	ausgabe, fehler := befehl.StdoutPipe()
	if fehler != nil {
		return nil, nil, domain.MitGrund(domain.ErrValidation, fehler,
			"die Ausgabe des Dienstes aufmachen")
	}

	return eingabe, ausgabe, nil
}

// starteDienstprogrammUeber ist derselbe Weg über einer gereichten Ortssuche.
//
// **Gereicht und nicht gerufen** — dieselbe Gestalt wie [werkzeugortNeben],
// und aus demselben Grund: unter `go test` liegt neben dem Testprogramm kein
// `mse-rest`, und ohne diese Naht bliebe der ganze Weg ungemessen (§3).
//
// **Der Name bleibt ein Literal.** Was gereicht wird, ist das Verzeichnis —
// im Betrieb das des eigenen Bündels. Wer den **Namen** setzen könnte,
// entschiede, was läuft; genau das verhindert
// `internal/schichten/unterprozess_test.go`.
func starteDienstprogrammUeber(
	ctx context.Context, naehte dienstnaehte, art domain.Dienstart,
	bindung domain.Bindung,
) (dienstlauf, error) {
	return starteDienstprogrammAufAllen(ctx, naehte, art, []domain.Bindung{bindung})
}

// starteDienstprogrammAufAllen startet das Programm mit **je einer Fahne**
// `--adresse` für jede Bindung der Wahl (G08, #2272); der Port ist der
// gemeinsame der ersten.
func starteDienstprogrammAufAllen(
	ctx context.Context, naehte dienstnaehte, art domain.Dienstart,
	bindungen []domain.Bindung,
) (dienstlauf, error) {
	if len(bindungen) == 0 {
		bindungen = []domain.Bindung{{}}
	}

	ort, fehler := naehte.orten()
	if fehler != nil {
		return dienstlauf{}, fehler
	}

	laufender, beenden := context.WithCancel(context.WithoutCancel(ctx))

	argumente := make([]string, 0, 2*len(bindungen)+2)
	for _, bindung := range bindungen {
		argumente = append(argumente, "--adresse", bindung.Wirkadresse())
	}

	argumente = append(argumente, "--port", strconv.Itoa(bindungen[0].Port))

	// **Der Programmname steht als Literal da und nicht in einer
	// Veränderlichen.** `internal/schichten/unterprozess_test.go` verlangt
	// es, und der Grund ist gut: wer den Namen setzen kann, entscheidet, was
	// läuft. Ein `dienstprogramme[art]` sähe kürzer aus und gäbe genau diese
	// Zusage auf.
	var befehl *exec.Cmd

	switch art {
	case domain.DienstREST:
		//#nosec G204 -- der Name ist ein Literal, der Ort kommt vom Bündel
		befehl = exec.CommandContext(laufender,
			filepath.Join(ort, "mse-rest"), argumente...)

	case domain.DienstMCP:
		//#nosec G204 -- dito
		befehl = exec.CommandContext(laufender,
			filepath.Join(ort, "mse-mcp"), argumente...)

	default:
		beenden()

		return dienstlauf{}, fmt.Errorf(
			"%w: %q ist keine Dienstart", domain.ErrValidation, string(art))
	}

	// **Seine Protokollzeilen gehen ins Protokoll des Fensters** — Nachlauf
	// über Phase 6.4, N3: ohne Empfänger gingen sie ins Nullgerät, und die
	// Vorgangskennung einer Antwort führte ins Leere.
	befehl.Stderr = neuesKindprotokoll(filepath.Base(befehl.Path), slog.Default())

	// **Die Standardeingabe bleibt am Fenster.** Stirbt es, schließt das
	// Betriebssystem das Rohr, und das Kind endet von selbst — siehe den Kopf
	// dieser Datei.
	eingabe, ausgabe, fehler := naehte.rohre(befehl)
	if fehler != nil {
		beenden()

		return dienstlauf{}, fehler
	}

	if gestartet := befehl.Start(); gestartet != nil {
		beenden()

		return dienstlauf{}, domain.MitGrund(domain.ErrValidation, gestartet,
			"%s starten", befehl.Path)
	}

	// **Gewartet wird am Kontext des Rufs und nicht am eigenen.** Der eigene
	// ist mit Absicht abgekoppelt — der Dienst soll den Ruf überleben, der ihn
	// gestartet hat. **Beim Warten gilt das nicht**: bricht der Ruf ab, weil
	// das Fenster schließt, hat niemand mehr ein Interesse an diesem Start,
	// und fünf Sekunden Warten hielten das Schließen auf.
	//
	// **Und ohne diese Zeile wäre der Zweig darunter unbetretbar** —
	// `laufender` bricht erst ab, wenn [dienstlauf.beenden] gerufen wird, und
	// das geschieht frühestens nach dieser Zeile (§3).
	ansage, fehler := leseAnsage(ctx, ausgabe, naehte.frist)

	// **Erst lesen, dann abholen.** `Wait` schließt die Ausgabe, sobald das
	// Kind endet, und es vor dem letzten Lesen zu rufen ist deshalb falsch
	// (`os/exec`, `Cmd.StdoutPipe`, Go 1.27.1). Hier stand es gleich nach
	// dem Start: endete ein Dienst schneller, als gelesen wurde, war seine
	// Ansage fort — gesehen am 2026-09-24 als roter Pipelinelauf unter Last.
	fertig := make(chan struct{})

	go func() {
		defer close(fertig)

		if gewartet := befehl.Wait(); gewartet != nil && laufender.Err() == nil {
			meldeDienstbefund(gewartet)
		}
	}()

	if fehler != nil {
		beenden()
		<-fertig

		return dienstlauf{}, fehler
	}

	ansage.beenden = func() {
		// **Erst das Rohr, dann das Signal.** Wer nur abbräche, verließe
		// sich darauf, dass das Signal ankommt; das Rohr schließt in
		// jedem Fall.
		_ = eingabe.Close() //nolint:errcheck // das Kind endet ohnehin

		beenden()
	}
	ansage.fertig = fertig

	return ansage, nil
}

// leseAnsage liest die zwei Zeilen, mit denen sich ein Dienst meldet.
//
// **Zwei Zeilen und keine Datei** — der Grund steht in `cmd/mse-rest`: eine
// Datei müsste angelegt, gefunden, gelesen und aufgeräumt werden, und sie
// überlebte den Prozess, dem sie gehört.
//
// **Mit Frist**, weil ein Dienst, der nichts sagt, sonst das Fenster
// festhielte. **Gereicht und nicht gerufen** — sonst wartete der Lauf, der
// sie misst, wirklich fünf Sekunden (§3).
//
// # Und warum „Marke" die letzte Zeile ist
//
// **Weil sie das Ende der Ansage markiert.** Jenseits der Schleife steht der
// Fingerabdruck dazwischen; wer hier auf Adresse und Marke wartet, hat ihn
// dann schon gelesen. Ein Warten auf alle drei ginge nicht: auf der Schleife
// gibt es keinen, und der Leser weiß hier nicht, welche Bindung gilt.
func leseAnsage(
	ctx context.Context, ausgabe io.Reader, frist time.Duration,
) (dienstlauf, error) {
	gelesen := make(chan dienstlauf, 1)
	gescheitert := make(chan error, 1)

	go func() {
		// **Ein Leser, der zusammenbricht, ist eine gescheiterte Ansage**
		// (BK-05a, #2111) — und nicht das Ende der Anwendung, die den Dienst
		// gerade starten wollte.
		defer domain.Auffangen(func(zusammenbruch error) { gescheitert <- zusammenbruch })

		var gefunden dienstlauf

		leser := bufio.NewScanner(ausgabe)
		for leser.Scan() {
			zeile := leser.Text()

			if adresse, ja := strings.CutPrefix(zeile, "Dienst auf "); ja {
				gefunden.adressen = append(gefunden.adressen, strings.TrimSpace(adresse))
				gefunden.adresse = gefunden.adressen[0]
			}

			if fehlende, ja := strings.CutPrefix(zeile, "Ohne "); ja {
				adresse, grund, _ := strings.Cut(fehlende, ": ")
				gefunden.fehlende = append(gefunden.fehlende, domain.Lauschbefund{Adresse: adresse, Grund: grund})
			}

			if abdruck, ja := strings.CutPrefix(zeile, "Fingerabdruck "); ja {
				gefunden.fingerabdruck = strings.TrimSpace(abdruck)
			}

			if zeugnis, ja := strings.CutPrefix(zeile, "Zeugnis von "); ja {
				gefunden.aussteller, gefunden.gueltigBis = zeugnisAus(zeugnis)
			}

			if marke, ja := strings.CutPrefix(zeile, "Marke "); ja {
				gefunden.marke = strings.TrimSpace(marke)
			}

			if gefunden.adresse != "" && gefunden.marke != "" {
				gelesen <- gefunden

				return
			}
		}

		gescheitert <- fmt.Errorf(
			"%w: der Dienst hat sich nicht gemeldet", domain.ErrValidation)
	}()

	select {
	case ansage := <-gelesen:
		return ansage, nil

	case fehler := <-gescheitert:
		return dienstlauf{}, fehler

	case <-ctx.Done():
		return dienstlauf{}, fmt.Errorf(
			"%w: der Dienst hat sich nicht gemeldet", domain.ErrValidation)

	case <-time.After(frist):
		return dienstlauf{}, fmt.Errorf(
			"%w: der Dienst hat sich nicht innerhalb von %s gemeldet",
			domain.ErrValidation, frist)
	}
}

// zeugnisAus liest „<Aussteller> bis <Zeit>" — am **letzten** „ bis ", denn
// der Name einer Stelle darf es enthalten, die Zeit nicht.
func zeugnisAus(zeile string) (string, time.Time) {
	stelle := strings.LastIndex(zeile, " bis ")
	if stelle < 0 {
		return strings.TrimSpace(zeile), time.Time{}
	}

	bis, _ := time.Parse(time.RFC3339, strings.TrimSpace(zeile[stelle+len(" bis "):])) //nolint:errcheck // eine unlesbare Zeit bleibt null und steht nicht da

	return zeile[:stelle], bis
}
