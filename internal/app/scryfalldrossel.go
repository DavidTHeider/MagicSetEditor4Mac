// Die Drosseln, die alle Zugänge zu Scryfall teilen
// (#888).
//
// **Eine eigene Datei, weil `scryfall.go` mit ihr über 400 Zeilen ginge** (§2).
// Sie gehört zum Composition Root: `internal/schichten` lässt die Kante auf
// `internal/netz` nur in benannten Dateien zu, und diese ist eine davon.

package app

import (
	"os"
	"path/filepath"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
)

// drosselwerk hält die Drosseln, die **alle** Netzzugänge dieses Prozesses
// teilen — je Gegenstelle eine, über `Drosselwerk.Fuer`.
//
// **Eines, nicht eines je Dienst.** Ein zweites Werk hieße zwei Zustände über
// derselben Frage; ein Wächter zählt deshalb nach
// (`internal/werkzeuge/drosselwerk_test.go`), und er hat beim Freigabezugang
// (#1100) sofort gegriffen.
//
// **Scryfall und GitHub teilen es, ohne einander zu behindern**: die Drossel
// hängt am Schlüssel, und die Schlüssel sind verschieden.
//
// # Warum das ein Zustand des Prozesses ist
//
// **Das Budget gilt für Host und IP.** Vor dem 2026-08-24 baute dieses Paket
// vier Klienten — Abgleich, Bildweg, Auskunft, Sammeldaten —, jeden mit eigener
// Drossel; MCP und REST bauten ihre Naht sogar **je Aufruf** neu. Weil der
// erste Aufruf einer Drossel absichtlich nicht wartet, sah eine solche Drossel
// praktisch nie einen zweiten, und die Grenze war wirkungslos. Angestrichen vom
// Zweitblick über Phase V
// (#888).
//
// # Warum hier und nicht in `internal/netz`
//
// Weil ein Blatt keinen Prozesszustand halten soll (§2). `internal/app` **ist**
// der Composition Root; was für den ganzen Prozess gilt, gehört dorthin.
//
// # Über Prozessgrenzen hinweg gilt es seit Phase 3.2
//
// Fenster, `mse-mcp` und `mse-rest` sind eigene Prozesse; wer alle drei
// gleichzeitig fährt, hätte sonst dreimal das Budget. Das Schloss dafür liegt
// beim Zwischenspeicher
// (#1163,
// Befund Z7-03), und wie es arbeitet, steht in `netz/prozessdrossel.go`.
//
// # Und die Sperre aus einem 429 gilt mit
//
// Sie steht in derselben Nachbardatei wie der Abstand. Ohne sie sendete ein
// zweiter Prozess während `Retry-After` weiter — und ignorierte das 429 damit
// gemeinsam, auch wenn dieser es einhält.
//
// # Der Ort wird beim Gebrauch bestimmt, nicht beim Laden
//
// **Ein Paketwert wertete ihn beim Laden des Pakets aus** — also vor jedem
// `t.Setenv("HOME", …)`. Die Läufe legten damit Nachbardateien im **echten**
// Zuhause an und blieben dabei grün: genau der Fall, den §3 unter „ein Test,
// der nach draußen greift" führt. Angestrichen vom Nachlauf über Phase 3.2.
var drosselwerk = netz.NeuesDrosselwerk(time.Now, nil).
	MitAbsprache(ScryfallAbspracheort)

// Drosselordner ist das Verzeichnis der prozessübergreifenden Absprache.
const Drosselordner = "drosseln"

// ScryfallAbspracheort gibt das Verzeichnis der Absprache.
//
// # Neben dem Bestand und nicht im Scryfall-Zwischenspeicher
//
// Der erste Anlauf legte es dorthin — und **dasselbe Werk drosselt auch
// GitHub** (`Fassungsnaht`, eines je Prozess). Damit lagen die Nachbardateien
// des Aktualisierungswegs im Zwischenspeicher einer ganz anderen Gegenstelle.
// Angestrichen vom Nachlauf über Phase 3.2, an dem Weg, auf dem sie im
// **echten** Zuhause landeten.
//
// **Ein eigener Ordner**, weil es je Gegenstelle und Klasse eine Datei gibt:
// im Anwendungsverzeichnis nebeneinander wären es Fremdkörper zwischen
// Bestand, Paketen und Protokoll.
//
// **Je Rechner und nicht je Bestand** — die eine Stelle, die QA-10.NAT-01
// (#1741) bewusst teilt. Zwei Instanzen auf einem Rechner gehen über
// dieselbe Adresse hinaus; drosselte jede für sich, spräche der Rechner mit
// doppelter Rate mit Scryfall.
//
// **Ein leerer Ort schaltet sie ab.** Ohne Benutzerverzeichnis gibt es keinen
// Platz dafür; dann drosselt jeder Prozess für sich, wie vor Phase 3.2 — und
// das ist besser als eine Anwendung, die deswegen gar nicht startet (§8).
func ScryfallAbspracheort() string {
	pfad, err := Datenbankpfad()
	if err != nil {
		return ""
	}

	ort := filepath.Join(filepath.Dir(pfad), Drosselordner)

	// **Angelegt und nicht nur benannt.** `dateisperre.Oeffnen` legt die
	// Nachbardatei an, nicht ihr Verzeichnis.
	//
	// **Ein Fehler heißt: keine Absprache.** Dann drosselt jeder Prozess für
	// sich — dieselbe Antwort wie ohne Benutzerverzeichnis, und aus demselben
	// Grund (§8).
	if fehler := os.MkdirAll(ort, 0o700); fehler != nil {
		return ""
	}

	return ort
}
