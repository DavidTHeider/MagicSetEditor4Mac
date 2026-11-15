// **Der technische Bereich** — Werkstück 5 von Phase 3.3
// (#1164).
//
// # Wofür
//
// Diese Anwendung bringt drei Programme ohne Fenster mit — `mse`, `mse-mcp`,
// `mse-rest`. Sie liegen **im Bündel** und damit in keinem `PATH`; wer sie
// benutzen will, muss ihren Ort kennen. Hier steht er, samt fertigen Zeilen
// zum Kopieren.
//
// # Was hier ausdrücklich nicht steht
//
// **Kein Ausführungsweg.** Das Issue sagt es: „Keine beliebige Shell. Was hier
// steht, ist eine Vorlage zum Kopieren." Eine Anwendung, die auf Knopfdruck
// beliebige Befehle führte, wäre ein Werkzeug für jeden, der ihr etwas
// unterschieben kann — und der Gewinn wäre ein gesparter Wechsel ins Terminal.
//
// # Warum die Zeilen hier entstehen und nicht im Fenster
//
// Weil sie **Pfade** enthalten, die nur dieses Programm kennt: den Ort seines
// eigenen Bündels und den seines Bestands. Eine Zeile, die das Fenster
// zusammensetzte, wäre eine zweite Fassung derselben Frage (§10).

package app

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Werkzeugzeile ist ein Befehl samt dem Satz, der sagt, wozu er gut ist.
type Werkzeugzeile struct {
	// Schluessel benennt den erklärenden Satz in den Sprachkatalogen (§9).
	//
	// **Kein deutscher Fließtext von hier**: er stünde in der englischen
	// Oberfläche deutsch — derselbe Fehler, den der Zweitblick über Phase 9
	// an den Bezugsquellen angestrichen hat.
	Schluessel string `json:"schluessel"`

	// Befehl ist die Zeile, wie sie in ein Terminal gehört.
	Befehl string `json:"befehl"`
}

// TechnikAntwort trägt den Ort der Werkzeuge und die Zeilen dazu.
type TechnikAntwort struct {
	// Ort ist das Verzeichnis, in dem die drei Programme liegen.
	Ort string `json:"ort"`

	// Vorhanden nennt die Werkzeuge, die dort wirklich liegen.
	//
	// **Gezählt und nicht behauptet**: wer die Anwendung aus dem Quelltext
	// startet, hat sie nicht — `make dev` baut nur das Fenster. Eine Liste,
	// die sie trotzdem nennte, schickte jemanden auf die Suche nach etwas,
	// das es nicht gibt.
	Vorhanden []string `json:"vorhanden"`

	Zeilen []Werkzeugzeile `json:"zeilen"`

	// Agenten sind fertige Konfigurationsbruchstücke für Claude Code und
	// Codex (A14, #1453) — kopierbar, ausführbar und **ohne Geheimnis**.
	Agenten []AgentenbruchstueckDTO `json:"agenten"`

	Fehler *Fehler `json:"fehler"`
}

// werkzeuge sind die drei Programme ohne Fenster.
var werkzeuge = []string{"mse", "mse-mcp", "mse-rest"}

// Technikangaben gibt den Ort der Werkzeuge und fertige Befehlszeilen.
//
// `setID` darf leer sein — dann bleiben die Zeilen, die ein Set brauchen, weg.
func (a *App) Technikangaben(ctx context.Context, setID string) TechnikAntwort {
	ctx = a.imProfil(ctx)

	return technikangabenNeben(ctx, os.Executable, setID)
}

// technikangabenNeben ist derselbe Weg über einer gereichten Programmsuche.
//
// **Gereicht und nicht gerufen** — dieselbe Gestalt wie `nebenDemProgramm` in
// `mitgeliefert.go`, und aus demselben Grund: `os.Executable` scheitert unter
// macOS nicht, und der Zweig daneben wäre einer, den kein Lauf betritt (§3).
func technikangabenNeben(ctx context.Context, programm func() (string, error), setID string) TechnikAntwort {
	pfad, err := programm()
	if err != nil {
		return TechnikAntwort{
			Vorhanden: []string{}, Zeilen: []Werkzeugzeile{},
			Agenten: []AgentenbruchstueckDTO{},
			Fehler:  Uebersetze(ctx, err),
		}
	}

	ort := filepath.Dir(pfad)

	return TechnikAntwort{
		Ort:       ort,
		Vorhanden: vorhandeneWerkzeuge(ort),
		Zeilen:    zeilenFuer(ort, setID),
		Agenten:   Agentenbruchstuecke(ort),
	}
}

// Werkzeugort ist das Verzeichnis, in dem die drei Programme liegen.
//
// **Neben diesem Programm** — dasselbe, das [App.Technikangaben] als `Ort`
// nennt. Exportiert seit N6
// (#1572):
// ein Dienstposten startet sie, und er braucht denselben Ort wie die Zeilen
// zum Kopieren. Zwei Wege dorthin wären zwei, die auseinanderlaufen.
func Werkzeugort() (string, error) {
	return werkzeugortNeben(os.Executable)
}

// werkzeugortNeben ist derselbe Weg über einer gereichten Programmsuche.
//
// **Gereicht und nicht gerufen** — dieselbe Gestalt wie [technikangabenNeben]
// darüber und aus demselben Grund: `os.Executable` scheitert unter macOS
// nicht, und der Zweig daneben wäre einer, den kein Lauf betritt (§3).
func werkzeugortNeben(programm func() (string, error)) (string, error) {
	pfad, fehler := programm()
	if fehler != nil {
		return "", domain.MitGrund(domain.ErrValidation, fehler,
			"das eigene Programm finden")
	}

	// **Hinter einem Link liegt der Ort** (#1961, Hüllenprobe `ops09b`):
	// `os.Executable` nennt unter macOS den Pfad, über den gestartet wurde,
	// und `mse` steht nach OPS-03 als Link in `~/.local/bin`. Ohne diese
	// Zeile sah die Kommandozeile dort ein anderes `mse-mcp` als das
	// Fenster im Bündel. Ein Pfad, der sich nicht auflösen lässt, gilt, wie
	// er ist.
	if echt, aufgeloest := filepath.EvalSymlinks(pfad); aufgeloest == nil {
		pfad = echt
	}

	return filepath.Dir(pfad), nil
}

// vorhandeneWerkzeuge sagt, welche der drei dort wirklich liegen.
func vorhandeneWerkzeuge(ort string) []string {
	gefunden := make([]string, 0, len(werkzeuge))

	for _, name := range werkzeuge {
		if angaben, err := os.Stat(filepath.Join(ort, name)); err == nil && !angaben.IsDir() {
			gefunden = append(gefunden, name)
		}
	}

	return gefunden
}

// zeilenFuer baut die Befehle mit den Pfaden dieses Rechners.
//
// **Mit vollem Pfad und in Anführungszeichen.** Der Ort trägt unter macOS ein
// Leerzeichen (`Application Support`), und eine Zeile ohne Anführungszeichen
// zerfiele beim Einfügen in zwei Argumente.
func zeilenFuer(ort, setID string) []Werkzeugzeile {
	mse := zitiert(filepath.Join(ort, "mse"))

	// **Platz für sieben.** Fünf stehen immer da, zwei kommen mit einem
	// offenen Set dazu.
	zeilen := make([]Werkzeugzeile, 0, 7)

	zeilen = append(zeilen,
		Werkzeugzeile{Schluessel: "hilfe", Befehl: mse + " hilfe"},
		Werkzeugzeile{Schluessel: "sets", Befehl: mse + " sets"},
		Werkzeugzeile{
			Schluessel: "verknuepfen",
			Befehl:     "ln -s " + mse + " /usr/local/bin/mse",
		},
		Werkzeugzeile{
			Schluessel: "dienst",
			Befehl:     zitiert(filepath.Join(ort, "mse-rest")),
		},
		Werkzeugzeile{
			Schluessel: "mcp",
			Befehl:     zitiert(filepath.Join(ort, "mse-mcp")),
		})

	if setID == "" {
		return zeilen
	}

	// **Die Kennung des offenen Sets reist mit.** Wer sie abtippen müsste,
	// hätte eine Vorlage, die noch Arbeit macht.
	// **Auch sie wird zitiert.** Die Seite gibt heute eine Dezimalzahl, aber
	// die Bindung nimmt jede Zeichenkette — und was hier zusammengesetzt wird,
	// fügt jemand in ein Terminal ein. Angestrichen vom Nachlauf des
	// Zweitblicks über Phase 3.3.
	kennung := zitiert(setID)

	return append(zeilen,
		Werkzeugzeile{Schluessel: "karten", Befehl: mse + " karten " + kennung},
		Werkzeugzeile{Schluessel: "statistik", Befehl: mse + " statistik " + kennung},
	)
}

// zitiert setzt einen Pfad in Hochkommata, wenn er welche braucht.
//
// # Einfache und nicht doppelte
//
// In doppelten Anführungszeichen setzt eine Shell weiter ein: `$HOME` wird
// ersetzt, “ `…` “ ausgeführt. In **einfachen** ist jedes Zeichen es selbst;
// die einzige Ausnahme ist das Hochkomma, und dafür gibt es einen Weg — die
// Zeichenkette schließen, ein gefliehenes Hochkomma setzen, wieder öffnen.
//
// **`strconv.Quote` tut das nicht.** Es gibt ein **Go**-Literal, und das ist
// eine andere Sprache. Hier stand bis zum 2026-08-27 das Gegenteil, und ein
// Lauf schrieb es fest: `/x/$HOME/mse` sollte `"/x/$HOME/mse"` ergeben. Wer
// das einfügt, bekommt seinen eigenen Zuhause-Pfad mitten im Befehl.
// Angestrichen vom Zweitblick durch Codex über Phase 3.3.
//
// **Diese Anwendung führt nichts aus** — der Befehl geht in die Zwischenablage.
// Genau deshalb muss er stimmen: was hier falsch steht, führt jemand anderes
// aus.
func zitiert(pfad string) string {
	if !braucht(pfad) {
		return pfad
	}

	return "'" + strings.ReplaceAll(pfad, "'", `'\''`) + "'"
}

// braucht sagt, ob ein Pfad in einer Shell Hochkommata braucht.
//
// **Eine Erlaubnisliste und keine Verbotsliste.** Was hier nicht steht, wird
// zitiert. Eine Verbotsliste vergisst irgendwann ein Zeichen — die vorige
// vergaß die Tilde —, und ein vergessenes Zeichen fällt nicht auf, sondern
// wirkt.
func braucht(pfad string) bool {
	for _, zeichen := range pfad {
		harmlos := zeichen >= 'a' && zeichen <= 'z' ||
			zeichen >= 'A' && zeichen <= 'Z' ||
			zeichen >= '0' && zeichen <= '9' ||
			strings.ContainsRune("_@%+=:,./-", zeichen)

		if !harmlos {
			return true
		}
	}

	return false
}
