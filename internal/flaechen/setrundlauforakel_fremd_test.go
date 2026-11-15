//go:build fremdpakete

package flaechen_test

// **Das Orakel im Setdatei-Rundlauf** — F05
// (#2247):
// was das gebaute MSE2 über ein Set sagt — jeden Wert jeder Karte, eine
// setweite Stiloption und seine Warnungen beim Lesen.

import (
	"bytes"
	"context"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"
)

// orakelprogramm ist das gebaute MSE2 — Fassung f9d9356 vom 2021-07-06,
// gebaut nach Vermerk 0026.
const orakelprogramm = "/Users/david/mse2-quelle/build/magicseteditor"

// **Trenner, die in keinem Wert der Kohorte vorkommen** — Steuerbilder aus
// dem Unicode-Block „Control Pictures". Ein Kartenwert darf Tabulator und
// Zeilenende tragen; diese Zeichen trägt keiner.
const (
	feldtrenner   = "␟"
	kartentrenner = "␞"
	unlesbar      = "␀unlesbar"
)

// orakelantwort ist, was MSE2 über ein Set sagt.
type orakelantwort struct {
	karten []map[string]string

	// stil ist der Wert der gefragten setweiten Stiloption — oder leer.
	stil string

	// warnungen sind MSE2s Meldungen beim Lesen der Hauptdatei `set`, ohne
	// Zeilennummer: zwei Dateien mit verschieden vielen Zeilen trügen sonst
	// dieselbe Warnung zweimal verschieden.
	warnungen []string
}

// orakelAuszug fragt MSE2 nach jedem Feld jeder Karte eines Sets — dazu, wenn
// `stiloption` nicht leer ist, nach dem Wert dieser setweiten Stiloption
// (`styling.<name>`), und schreibt es, wenn `abschrift` nicht leer ist, mit
// `write_set_file` in `arbeit`.
//
// **Die Feldnamen in MSE2s Form** (`rule_text`): MSE2 legt einen Namen mit
// Unterstrich statt Leerzeichen ab, und `card["rule text"]` kennt es nicht
// (nachgefragt am 2026-10-10). **Ein Bildfeld ist `unlesbar`**: MSE2 macht aus
// einem Bild keinen Text; Bilder vergleicht [dateienVergleichen] an den
// Dateien.
func orakelAuszug(t *testing.T, setpfad, arbeit string, felder []string, stiloption, abschrift string) orakelantwort {
	t.Helper()

	namen := make([]string, len(felder))
	for i, feld := range felder {
		namen[i] = strconv.Quote(strings.ReplaceAll(feld, " ", "_"))
	}

	zeilen := []string{
		"rundlauf_felder := [" + strings.Join(namen, ", ") + "]",
		`"␂anfang"`,
		`for each c in set.cards do (for each f in rundlauf_felder do f + "` + feldtrenner +
			`" + (to_string(c[f]) or else "` + unlesbar + `") + "` + feldtrenner + `") + "` + kartentrenner + `"`,
	}
	if stiloption != "" {
		zeilen = append(zeilen, `"␂stil"`,
			"to_string(styling."+strings.ReplaceAll(stiloption, " ", "_")+`) or else "`+unlesbar+`"`)
	}
	if abschrift != "" {
		zeilen = append(zeilen, "write_set_file(set:set, file:"+strconv.Quote(abschrift)+")")
	}

	saetze, warnungen := orakelFragen(t, setpfad, arbeit, append(zeilen, ":quit"))
	antwort := orakelantwort{warnungen: warnungen}

	nach := func(marke string) (string, bool) {
		for i, satz := range saetze {
			if len(satz) == 1 && satz[0] == marke && i+1 < len(saetze) && len(saetze[i+1]) == 1 {
				return orakelEntquoten(t, saetze[i+1][0]), true
			}
		}

		return "", false
	}

	auszug, gefunden := nach(`"␂anfang"`)
	if !gefunden {
		t.Fatalf("das Orakel gab keinen Auszug für %s: %q", setpfad, saetze)
	}

	antwort.karten = kartenAusAuszug(t, auszug)

	if stiloption != "" {
		antwort.stil, _ = nach(`"␂stil"`)
	}

	return antwort
}

// orakelFragen startet MSE2 mit `--cli -q --raw` und gibt die Datensätze —
// samt den Warnungen über die Hauptdatei des Sets.
//
// **Das Rohformat ist ein Protokoll** (`src/cli/text_io_handler.cpp`,
// `flushRaw`, Fassung f9d9356 vom 2021-07-06): je Eingabe eine Zeile Status,
// eine Zeile mit der Zahl der folgenden Zeilen, dann diese Zeilen.
func orakelFragen(t *testing.T, setpfad, arbeit string, zeilen []string) (antworten [][]string, warnungen []string) {
	t.Helper()

	frist, abbrechen := context.WithTimeout(t.Context(), 20*time.Minute)
	defer abbrechen()

	// #nosec G204 -- das Programm ist eine Konstante dieser Datei, das Set
	// kommt aus der Kohorte, und der Lauf steht hinter `fremdpakete`.
	befehl := exec.CommandContext(frist, orakelprogramm, "--cli", setpfad, "-q", "--raw")
	befehl.Dir = arbeit
	befehl.Stdin = strings.NewReader(strings.Join(zeilen, "\n") + "\n")

	var aus, fehlerstrom bytes.Buffer
	befehl.Stdout, befehl.Stderr = &aus, &fehlerstrom

	if fehler := befehl.Run(); fehler != nil {
		t.Fatalf("das Orakel scheiterte an %s: %v", setpfad, fehler)
	}

	// **Warnungen beim Laden stehen außerhalb des Protokolls** — „newer
	// version" und ihre Fortsetzungszeilen, vor dem ersten Datensatz. Ein
	// Datensatz beginnt deshalb erst dort, wo zwei Zahlenzeilen aufeinander
	// folgen; was dazwischen steht, wird übergangen.
	var saetze [][]string

	ausgabe := strings.Split(strings.TrimSuffix(aus.String(), "\n"), "\n")

	for i := 0; i+1 < len(ausgabe); {
		_, keinStatus := strconv.Atoi(strings.TrimSpace(ausgabe[i]))
		zahl, keineZahl := strconv.Atoi(strings.TrimSpace(ausgabe[i+1]))

		if keinStatus != nil || keineZahl != nil || i+2+zahl > len(ausgabe) {
			i++

			continue
		}

		saetze = append(saetze, ausgabe[i+2:i+2+zahl])
		i += 2 + zahl
	}

	// **Die Warnungen stehen im Rohmodus auf der Standardausgabe**, außerhalb
	// der Datensätze — gesehen am 2026-10-10: „On line 18: Unexpected key:
	// 'styling'" kam dort und nicht im Fehlerstrom. Der erste Anlauf las nur
	// den Fehlerstrom und meldete null Warnungen über eine Datei, die MSE2
	// abwies; die Gegenprobe gegen den alten Schreiber hat es gezeigt.
	return saetze, setwarnungen(aus.String() + "\n" + fehlerstrom.String())
}

// warnzeile ist eine Meldung des Lesers von MSE2 (`src/util/io/reader.cpp`,
// Fassung f9d9356 vom 2021-07-06): „On line 18: Unexpected key: 'styling'".
var warnzeile = regexp.MustCompile(`^\s*On line \d+:\s*(.*)$`)

var farbfolge = regexp.MustCompile("\x1b\\[[0-9;]*m")

// setwarnungen liest aus MSE2s Fehlerstrom die Meldungen über die Hauptdatei
// eines Sets. Ein Block beginnt mit „Warnings while reading file:", die Zeile
// danach nennt die Datei.
func setwarnungen(fehlerstrom string) []string {
	var (
		warnungen []string
		imSet     bool
	)

	zeilen := strings.Split(farbfolge.ReplaceAllString(fehlerstrom, ""), "\n")

	for i, zeile := range zeilen {
		switch {
		case strings.Contains(zeile, "Warnings while reading file:"):
			imSet = i+1 < len(zeilen) && filepath.Base(strings.TrimSpace(zeilen[i+1])) == "set"
		case strings.HasPrefix(zeile, "WARNING:") || strings.HasPrefix(zeile, "ERROR:"):
			imSet = false
		case imSet:
			if treffer := warnzeile.FindStringSubmatch(zeile); treffer != nil {
				warnungen = append(warnungen, strings.TrimSpace(treffer[1]))
			}
		}
	}

	return warnungen
}

// neueWarnungen nennt, was `nachher` warnt und `vorher` nicht — als Menge
// mit Vielfachheit.
func neueWarnungen(vorher, nachher []string) []string {
	gesehen := make(map[string]int)
	for _, warnung := range vorher {
		gesehen[warnung]++
	}

	var neu []string

	for _, warnung := range nachher {
		if gesehen[warnung] > 0 {
			gesehen[warnung]--

			continue
		}

		neu = append(neu, warnung)
	}

	return neu
}

// orakelEntquoten liest eine Zeichenkette, wie MSE2 sie zitiert
// (`quote_string`, `src/script/value.cpp`, Fassung f9d9356 vom 2021-07-06).
//
// `\<` bleibt stehen: es ist MSE2s Zeichen für eine geschützte spitze
// Klammer, und beide Seiten eines Vergleichs tragen es gleich.
func orakelEntquoten(t *testing.T, zitat string) string {
	t.Helper()

	if len(zitat) < 2 || zitat[0] != '"' || zitat[len(zitat)-1] != '"' {
		t.Fatalf("keine zitierte Zeichenkette: %.200q", zitat)
	}

	ersatz := strings.NewReplacer(`\\`, `\`, `\"`, `"`, `\n`, "\n", `\r`, "\r", `\t`, "\t")

	return ersatz.Replace(zitat[1 : len(zitat)-1])
}

func kartenAusAuszug(t *testing.T, auszug string) []map[string]string {
	t.Helper()

	stuecke := strings.Split(strings.TrimSuffix(auszug, kartentrenner), kartentrenner)
	karten := make([]map[string]string, 0, len(stuecke))

	for _, karte := range stuecke {
		teile := strings.Split(strings.TrimSuffix(karte, feldtrenner), feldtrenner)
		if len(teile)%2 != 0 {
			t.Fatalf("ein Kartenauszug ohne Paare: %.200q", karte)
		}

		werte := make(map[string]string, len(teile)/2)
		for i := 0; i < len(teile); i += 2 {
			werte[teile[i]] = teile[i+1]
		}

		karten = append(karten, werte)
	}

	return karten
}
