// **Was ein offener Server an Pfaden annimmt** — OPS-02.AK-04
// (#1719),
// Marke `new`: MSE2 hat keinen Server.
//
// # Die Regel
//
// **Ein Pfad an den Dienst liegt im Zuhause oder in einem Ordner, den dieses
// Programm kennt — und nicht in einem verborgenen Winkel des Zuhauses.**
// Verborgen heißt: ein Glied, das mit einem Punkt beginnt (`~/.ssh`,
// `~/.zshrc`), und `~/Library`, wo Schlüsselbund und die Daten anderer
// Anwendungen liegen. Der Dienst liest Setpakete und schreibt Bögen; dafür
// braucht er keinen dieser Orte, und wer ein Token hat, soll über einen
// Pfad nicht herausfinden, was dort steht.
//
// **Die weiteren Ordner** sind die eigenen Orte der Anwendung (Bestand,
// Pakete), die vier eingestellten Ordner — soweit sie im Zuhause oder auf
// einem Laufwerk liegen ([PruefeDienstordner]) — und der
// Zwischenablageordner des Prozesses. Ein Laufwerk unter `/Volumes`, das
// jemand als Ausfuhrordner eingestellt hat, ist ein Ort, den er gemeint hat.
//
// # Was hier nicht geprüft wird
//
// **Verknüpfungen.** Ob `~/Karten` in Wahrheit nach `~/.ssh` zeigt, sagt nur
// die Platte; `app` löst Pfad **und** Ordner auf, bevor es hier fragt — der
// genannte Pfad selbst wird nicht noch einmal geprüft.

package domain

import (
	"encoding/json"
	"maps"
	"path/filepath"
	"slices"
	"strings"
)

// pfadfelder sind die Namen, unter denen ein Pfad über REST und MCP reist.
var pfadfelder = []string{"pfad", "ziel", "datei", "verzeichnis", "bildpfad"}

// IstPfadfeld sagt, ob ein Feld- oder Parametername einen Pfad trägt.
func IstPfadfeld(name string) bool { return slices.Contains(pfadfelder, name) }

// PfadeIn gibt die Pfade eines JSON-Objekts — die Zeichenketten unter einem
// Pfadfeld oder einem der weiteren Namen, **in der Ordnung der Namen**, damit
// zwei falsche Pfade immer dieselbe Meldung geben. Was kein Objekt ist, hat
// keine: dann sagt der Weg, was an ihm falsch ist.
//
// # Ohne Rücksicht auf die Schreibweise
//
// **Ein Name gilt so, wie `encoding/json` ihn zuordnet** (#2073): `"Ziel"`
// landet im Feld mit der Marke `json:"ziel"`, auch mit
// `DisallowUnknownFields`, und `strings.EqualFold` faltet wie der
// Dekodierer — bis zu »ſ«, das er als »s« liest. Wer nur den genauen Namen
// prüfte, ließ den Pfad unter dem groß geschriebenen durch.
//
// # Nur oben im Rumpf
//
// **Verschachteltes wird nicht gelesen**, und das ist Absicht: unter
// `nutzlast` eines Spielbefehls steht `ziel` für einen Spieler, nicht für
// eine Datei. Dass kein Weg ein Pfadfeld tiefer erwartet, hält
// `internal/werkzeuge/pfadstrukturen_test.go`.
func PfadeIn(roh []byte, weitere ...string) []string {
	var felder map[string]json.RawMessage
	if json.Unmarshal(roh, &felder) != nil {
		return nil
	}

	pfade := []string{}

	for _, name := range slices.Sorted(maps.Keys(felder)) {
		var pfad string
		if traegtPfad(name, weitere) && json.Unmarshal(felder[name], &pfad) == nil {
			pfade = append(pfade, pfad)
		}
	}

	return pfade
}

// traegtPfad sagt, ob ein Name im Rumpf ein Pfadfeld oder einen der weiteren
// Namen meint — in jeder Schreibweise, die der Dekodierer zuordnet.
func traegtPfad(name string, weitere []string) bool {
	gleich := func(bekannt string) bool { return strings.EqualFold(name, bekannt) }

	return slices.ContainsFunc(pfadfelder, gleich) || slices.ContainsFunc(weitere, gleich)
}

// PruefeDienstpfad urteilt über einen Pfad, den ein offener Server bekommt.
func PruefeDienstpfad(pfad, zuhause string, weitere []string) error {
	if !filepath.IsAbs(pfad) {
		return NewFieldError("pfad", SchluesselValidierung,
			"»%s« ist kein vollständiger Ort — der Dienst nimmt nur Pfade, die mit / beginnen.", pfad)
	}

	sauber := filepath.Clean(pfad)

	// **Ein verborgener Winkel ist überall einer**, auch in einem erlaubten
	// Ordner — ein `.ssh`, ein `.git`, ein `.zshrc`. Gefunden in der
	// Pipeline: dort liegt das Zuhause selbst unter `$TMPDIR`, und `~/.ssh`
	// kam über die Erlaubnis für den Zwischenablageordner durch.
	if slices.ContainsFunc(strings.Split(sauber, string(filepath.Separator)), func(glied string) bool {
		return strings.HasPrefix(glied, ".")
	}) {
		return NewFieldError("pfad", SchluesselValidierung,
			"»%s« liegt in einem verborgenen Ordner — der Dienst fasst keinen an.", pfad)
	}

	// **Die Bibliothek bleibt zu, auch in einem erlaubten Ordner** — es sei
	// denn, der Ordner liegt selbst dort: die Pakete dieses Programms stehen
	// unter `~/Library/Application Support`. Ein Zwischenablageordner, der
	// das ganze Zuhause umfasst — in der Pipeline liegt es unter `$TMPDIR` —,
	// öffnet sie nicht (Befund 2 des Zweitblicks über Phase 6.4).
	heim := filepath.Clean(zuhause)
	for _, ordner := range weitere {
		erlaubt := filepath.Clean(ordner)
		if ordner != "" && innerhalb(sauber, erlaubt) && (!verborgen(sauber, heim) || verborgen(erlaubt, heim)) {
			return nil
		}
	}

	if innerhalb(sauber, heim) && !verborgen(sauber, heim) {
		return nil
	}

	return NewFieldError("pfad", SchluesselValidierung,
		"»%s« liegt außerhalb dessen, was der Dienst anfassen darf — dem Zuhause ohne seine verborgenen Ordner und den eingestellten Ordnern.", pfad)
}

// innerhalb sagt, ob ein Pfad der Ordner selbst ist oder in ihm liegt — ein
// gemeinsamer Anfang allein genügt nicht (`/a/b2` liegt nicht in `/a/b`).
func innerhalb(pfad, ordner string) bool {
	return pfad == ordner || strings.HasPrefix(pfad, ordner+string(filepath.Separator))
}

// verborgen sagt, ob ein Pfad unter dem Zuhause in `~/Library` liegt — wo
// Schlüsselbund und die Daten anderer Anwendungen stehen. Glieder mit Punkt
// vorn weist [PruefeDienstpfad] schon davor ab, überall.
//
// **Gerechnet und nicht über `filepath.Rel`**: der Pfad liegt schon im
// Zuhause, der Rest ist, was hinter ihm steht.
//
// **In jeder Schreibweise** — Befund 1 des Zweitblicks über Phase 6.4: APFS
// unterscheidet Groß- und Kleinschreibung in der Vorgabe nicht, und
// `~/library/LaunchAgents` ist dieselbe Bibliothek.
func verborgen(pfad, zuhause string) bool {
	rest := strings.TrimPrefix(strings.TrimPrefix(pfad, zuhause), string(filepath.Separator))
	erstes, _, _ := strings.Cut(rest, string(filepath.Separator))

	return innerhalb(pfad, zuhause) && strings.EqualFold(erstes, "Library")
}

// Laufwerke ist der Ort, an dem macOS weitere Laufwerke einhängt.
const Laufwerke = "/Volumes"

// PruefeDienstordner urteilt über einen **Ordner**, den ein offener Server
// als seinen gelten lassen soll — Befund 2 des Zweitblicks über Phase 6.4.
//
// **Im Zuhause ohne seine verborgenen Winkel, oder auf einem eingehängten
// Laufwerk** — und sonst nirgends. Hier galt jeder eingestellte Ordner, und
// über den Dienst ließ sich jeder einstellen: `~/Library/LaunchAgents` oder
// gleich `/`. Ein Laufwerk unter `/Volumes` ist ein Ort, den jemand gemeint
// hat; `/Volumes` selbst ist keiner.
func PruefeDienstordner(ordner, zuhause string) error {
	return PruefeDienstordnerUnter(ordner, zuhause, Laufwerke)
}

// PruefeDienstordnerUnter ist [PruefeDienstordner] mit einem anderen Ort für
// Laufwerke — für Läufe, die kein Laufwerk einhängen können.
func PruefeDienstordnerUnter(ordner, zuhause, laufwerke string) error {
	sauber := filepath.Clean(ordner)
	if filepath.IsAbs(sauber) && innerhalb(sauber, laufwerke) && sauber != filepath.Clean(laufwerke) {
		return PruefeDienstpfad(sauber, zuhause, []string{sauber})
	}

	return PruefeDienstpfad(ordner, zuhause, nil)
}
