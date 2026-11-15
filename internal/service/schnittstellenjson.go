package service

// **Ein Eintrag unter `mcpServers`** — Claude Desktop und Claude Code,
// OPS-09b (#1961).
//
// # Warum ein Gedächtnis
//
// **JSON kennt keinen Kommentar** und damit keinen markierten Block. Damit
// widerrufen trotzdem byte-gleich zurückgeht, merkt sich einrichten den
// Stand davor und eine Prüfsumme dessen, was es schrieb. Widerrufen legt den
// Stand davor zurück, wenn die Datei seither unverändert ist; sonst nimmt es
// nur den eigenen Eintrag heraus und sagt das.

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// mcpObjekt ist der Schlüssel, unter dem beide Klienten ihre Server führen.
const mcpObjekt = "mcpServers"

// mcpEintrag ist, was dieses Programm von einem Eintrag liest.
type mcpEintrag struct {
	Command string   `json:"command"`
	Args    []string `json:"args"`
}

// eintragsstandAus liest den Stand eines Eintrags unter `mcpServers`.
//
// **Der eigene ist der mit dem eigenen Werkzeugserver**: dasselbe Programm
// an einem anderen Ort ist veraltet (die Anwendung zog um), ein anderes
// Programm unter demselben Namen ist fremd.
func eintragsstandAus(ziel Schnittstellenziel, inhalt []byte) Schnittstellenstand {
	wert, da, fehler := jsonEintrag(inhalt, ziel.Schluessel)
	if fehler != nil {
		return SchnittstelleFremd
	}

	if !da {
		if !ziel.Vorhanden {
			return SchnittstelleNichtGefunden
		}

		return SchnittstelleOffen
	}

	befehl, eigen := eigeneGestalt(wert)

	switch {
	case !eigen:
		return SchnittstelleFremd
	case befehl == ziel.Programm:
		return SchnittstelleEingerichtet
	case filepath.Base(befehl) == filepath.Base(ziel.Programm) && EigenesBuendel(befehl):
		return SchnittstelleVeraltet
	}

	return SchnittstelleFremd
}

// EigenesBuendel sagt, ob ein Programm im Bündel dieser Anwendung liegt —
// oder in einem, das es nicht mehr gibt. **Für die Einträge der Klienten,
// für die Verknüpfungen in `~/.local/bin` und für einen alten Dienst**
// (Zweitblick über Phase 4, Befund 8, und sein Nachlauf).
//
// **Die Gestalt allein ist kein Nachweis** (Nachlauf über Phase 6.12, N3):
// ein `mse-mcp` mit leeren Argumenten außerhalb eines Bündels hat jemand
// anders hingelegt — ein eigenes Skript, eine zweite Kopie —, und ein
// Reparieren nähme es ihm. Diese Anwendung schreibt ihren Werkzeugort, und
// der liegt im Bündel (`app.Werkzeugort` löst seit #1961 auch den Link aus
// `~/.local/bin` auf).
//
// **Und ein Bündel, das es gibt, ist erst eigen, wenn es das sagt**
// ([Buendelzustand]); umbenannt darf es sein. Gibt es es nicht mehr, ist die
// Anwendung umgezogen, und mehr als die Gestalt ist nicht übrig.
func EigenesBuendel(programm string) bool {
	vorhanden, eigen := Buendelzustand(programm)

	return eigen || !vorhanden && strings.HasSuffix(filepath.Dir(programm), ".app/Contents/MacOS")
}

// Buendelzustand sagt, ob es das Bündel eines Programms gibt und ob es
// dieser Anwendung gehört — **nach seinem `Info.plist`**.
//
// **Fehlt die Datei, ist es nicht eigen** (Nachlauf über Phase 4, Befund 9):
// hier galt ein fehlendes `Info.plist` wie ein fehlendes Bündel, und ein
// vorhandenes fremdes `MagicSetEditor4Mac Fremd.app` ohne diese Datei ging
// als eigenes durch. Was sich nicht lesen lässt, belegt nichts.
func Buendelzustand(programm string) (vorhanden, eigen bool) {
	macos := filepath.Dir(programm)
	if !strings.HasSuffix(macos, ".app/Contents/MacOS") {
		return false, false
	}

	if _, fehler := os.Lstat(filepath.Dir(filepath.Dir(macos))); errors.Is(fehler, fs.ErrNotExist) {
		return false, false
	}

	plist, fehler := zielLesen(filepath.Join(filepath.Dir(macos), "Info.plist"))
	if fehler != nil || plist == nil {
		return true, false
	}

	treffer := buendelkennung.FindSubmatch(plist)

	return true, len(treffer) == 2 && eigeneKennungen[string(treffer[1])]
}

// buendelkennung findet `CFBundleIdentifier` in einem `Info.plist` in XML —
// so legt `scripts/bauen.sh` es an.
var buendelkennung = regexp.MustCompile(`<key>CFBundleIdentifier</key>\s*<string>([^<]*)</string>`)

// eigeneKennungen sind die Kennungen dieser Anwendung: die aus
// `build/darwin/Info.plist` und die des Abnahmebaus aus `scripts/bauen.sh` —
// **und die alten vor 0.9.0** (Vermerk 0191, R02): ein Eintrag, der in ein
// altes eigenes Bündel zeigt, ist veraltet und lässt sich reparieren, nicht
// fremd.
var eigeneKennungen = map[string]bool{
	"io.github.davidtheider.magicseteditor4mac":         true,
	"io.github.davidtheider.magicseteditor4mac.abnahme": true,
	"com.wails.MagicSetEditor4Mac":                      true,
	"com.wails.MagicSetEditor4Mac.abnahme":              true,
}

// eigeneGestalt gibt den Befehl eines Eintrags — und ob er so aussieht, wie
// dieses Programm oder `claude mcp add` ihn schreibt.
//
// **Derselbe Programmname ist kein Nachweis** (Zweitblick über Phase 6.12,
// Z5): ein `mse-mcp` mit eigenen Argumenten oder eigener Umgebung hat
// jemand so gewollt, und ein Reparieren nähme ihm beides. Eigen ist ein
// `command`, dazu höchstens leere `args`, ein leeres `env` und
// `type: "stdio"` — genau das schreibt `claude mcp add -s user` (gemessen am
// 2026-10-01 an `~/.claude.json`). Jedes andere Feld macht ihn fremd.
func eigeneGestalt(wert json.RawMessage) (befehl string, eigen bool) {
	var eintrag struct {
		Command string            `json:"command"`
		Args    []string          `json:"args"`
		Env     map[string]string `json:"env"`
		Type    *string           `json:"type"`
	}

	leser := json.NewDecoder(bytes.NewReader(wert))
	leser.DisallowUnknownFields()

	if leser.Decode(&eintrag) != nil {
		return "", false
	}

	return eintrag.Command, eintrag.Command != "" && len(eintrag.Args) == 0 && len(eintrag.Env) == 0 &&
		(eintrag.Type == nil || *eintrag.Type == "stdio")
}

// gedaechtnis ist der Stand vor dem Einrichten.
type gedaechtnis struct {
	// Fehlte sagt, dass es die Datei vorher nicht gab.
	Fehlte bool `json:"fehlte"`

	Vorher string `json:"vorher"`

	// Nachher ist die Prüfsumme dessen, was einrichten schrieb.
	Nachher string `json:"nachher"`
}

// jsonEinrichten setzt den Eintrag und merkt sich den Stand davor — den
// ersten: wer einen veralteten repariert, will später zurück zum Anfang.
//
// # Wann das Gedächtnis gilt
//
// **Nur solange die Datei ist, was dieses Programm zuletzt schrieb**
// (Zweitblick über Phase 6.12, Z1). Hat der Mensch sie seither geändert,
// legte ein Widerruf den Stand von vor dem ersten Einrichten zurück, und
// seine Änderung wäre fort. Dann wird es vergessen: ohne Gedächtnis nimmt
// Widerrufen nur den eigenen Eintrag. Ein neues entsteht nur, wo noch kein
// Eintrag steht — sonst merkte es sich einen Stand, in dem der eigene
// Eintrag schon steckt.
func jsonEinrichten(ziel Schnittstellenziel, inhalt []byte) error {
	eintrag, _ := json.Marshal(mcpEintrag{Command: ziel.Programm, Args: []string{}}) //nolint:errcheck // aus Zeichenketten gebaut

	// **Ohne Fehlerzweig**: hierher kommt nur, wessen Datei sich als Objekt
	// lesen ließ — sonst wäre ihr Stand „fremd“, und einrichten hielte vorher.
	neu, _ := jsonEintragSetzen(inhalt, ziel.Schluessel, eintrag) //nolint:errcheck // siehe oben
	_, schonDa, _ := jsonEintrag(inhalt, ziel.Schluessel)         //nolint:errcheck // siehe oben

	alt, gemerkt := gedaechtnisLesen(ziel.Gedaechtnis)

	switch {
	case gemerkt && alt.Nachher == pruefsumme(inhalt):
		// **Ein gültiges Gedächtnis wird erst nach der Datei fortgeschrieben**
		// (Nachlauf über Phase 6.12, N5): scheitert das Ersetzen, bleibt es,
		// wie es war, und passt weiter zur unveränderten Datei. Scheitert
		// danach das Fortschreiben, passt es zu keiner Datei mehr, und
		// Widerrufen nimmt nur den eigenen Eintrag — nichts geht verloren.
		if fehler := zielErsetzen(ziel, inhalt, neu); fehler != nil {
			return fehler
		}

		return gedaechtnisSchreiben(ziel, gedaechtnis{Fehlte: alt.Fehlte, Vorher: alt.Vorher, Nachher: pruefsumme(neu)})
	case !schonDa:
		// **Ein neues Gedächtnis vor der Datei** — ohne Gedächtnis wird nicht
		// eingerichtet, sonst ginge der Stand davor verloren. Scheitert das
		// Ersetzen, geht es wieder.
		if fehler := gedaechtnisSchreiben(ziel, gedaechtnis{
			Fehlte: inhalt == nil, Vorher: string(inhalt), Nachher: pruefsumme(neu),
		}); fehler != nil {
			return fehler
		}

		if fehler := zielErsetzen(ziel, inhalt, neu); fehler != nil {
			return errors.Join(fehler, gedaechtnisVergessen(ziel))
		}

		return nil
	default:
		if fehler := gedaechtnisVergessen(ziel); fehler != nil {
			return fehler
		}

		return zielErsetzen(ziel, inhalt, neu)
	}
}

// gedaechtnisSchreiben legt das Gedächtnis eines Ziels ab.
func gedaechtnisSchreiben(ziel Schnittstellenziel, merk gedaechtnis) error {
	roh, _ := json.Marshal(merk) //nolint:errcheck // aus Zeichenketten gebaut

	return zielSchreiben(ziel.Gedaechtnis, roh)
}

// jsonWiderrufen legt den Stand davor zurück — oder nimmt nur den eigenen
// Eintrag, wenn die Datei sich seither geändert hat. Der erste Wert sagt,
// ob es byte-gleich ging.
func jsonWiderrufen(ziel Schnittstellenziel, inhalt []byte) (bool, error) {
	alt, gemerkt := gedaechtnisLesen(ziel.Gedaechtnis)

	var fehler error

	switch {
	case gemerkt && alt.Nachher == pruefsumme(inhalt) && alt.Fehlte:
		fehler = zielEntfernen(ziel, inhalt)
	case gemerkt && alt.Nachher == pruefsumme(inhalt):
		fehler = zielErsetzen(ziel, inhalt, []byte(alt.Vorher))
	default:
		// **Ohne Fehlerzweig**, aus demselben Grund wie beim Einrichten:
		// widerrufen kommt nur hierher, wenn der Stand „eingerichtet“ oder
		// „veraltet“ war, die Datei sich also lesen ließ.
		uebrig, _ := jsonEintragEntfernen(inhalt, ziel.Schluessel) //nolint:errcheck // siehe oben

		gemerkt = false
		fehler = zielErsetzen(ziel, inhalt, uebrig)
	}

	if fehler != nil {
		return false, fmt.Errorf("%s zurücklegen: %w", ziel.Datei, fehler)
	}

	return gemerkt, gedaechtnisVergessen(ziel)
}

// gedaechtnisVergessen nimmt das Gedächtnis eines Ziels weg — eines, das es
// nicht gibt, ist schon vergessen.
func gedaechtnisVergessen(ziel Schnittstellenziel) error {
	if entfernt := os.Remove(ziel.Gedaechtnis); entfernt != nil && !errors.Is(entfernt, fs.ErrNotExist) {
		return fmt.Errorf("das Gedächtnis zu %s entfernen: %w", ziel.Name, entfernt)
	}

	return nil
}

// gedaechtnisLesen liest den gemerkten Stand — der zweite Wert sagt, ob es
// einen lesbaren gibt.
func gedaechtnisLesen(datei string) (gedaechtnis, bool) {
	roh, fehler := zielLesen(datei)

	var gemerkt gedaechtnis
	if fehler != nil || roh == nil || json.Unmarshal(roh, &gemerkt) != nil {
		return gedaechtnis{}, false
	}

	return gemerkt, true
}

// pruefsumme gibt SHA-256 als Hex.
func pruefsumme(inhalt []byte) string {
	summe := sha256.Sum256(inhalt)

	return hex.EncodeToString(summe[:])
}
