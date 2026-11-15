package ollama

// **Was über die Auslagerung feststeht** — Werkstück A7 von Phase 5.3
// (#1446).
//
// # Warum drei Stände und nicht zwei
//
// **Weil „nicht abgeschaltet" und „unbekannt" zwei verschiedene Auskünfte
// sind.** Ollama nennt zwei dokumentierte Wege, die Cloud abzuschalten
// (**nachgelesen am 2026-09-11** unter <https://docs.ollama.com/cloud> und
// <https://docs.ollama.com/faq>):
//
//   - `~/.ollama/server.json` mit `{"disable_ollama_cloud": true}` — **eine
//     Datei, die hier zu lesen ist**;
//   - `OLLAMA_NO_CLOUD=1` — **die Umgebung eines fremden Prozesses**, und die
//     ist es nicht.
//
// Wer eine fehlende Datei als „Cloud an" ausgäbe, behauptete etwas über eine
// Umgebung, die er nicht kennt. Deshalb sagt dieses Paket in dem Fall
// „ungewiss" und nicht mehr.
//
// # Was dieser Befund **nicht** ist
//
// **Keine Netzgarantie.** Er sagt, was in einer Datei steht — nicht, was der
// Prozess tut, der auf 11434 antwortet, und schon gar nicht, ob es Ollama ist
// ([0098](../../docs/entscheidungen/0098-was-ollama-belegt-und-was-nicht.md)).
// Die harte Seite dieser Frage ist der Ort eines Modells: ein Alias auf
// [Cloudsuffix] lagert aus, und `KIOrtsregel` weist ihn ab.

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Serverdatei ist der Ort der Servereinstellungen — **unter** dem Zuhause.
const Serverdatei = ".ollama/server.json"

// derCloudschalter ist der dokumentierte Schlüssel darin.
const derCloudschalter = "disable_ollama_cloud"

// Cloudstand ist, was über die Auslagerung feststeht.
type Cloudstand string

// Die drei Stände.
const (
	// CloudAus heißt: die Serverdatei schaltet sie ausdrücklich ab.
	CloudAus Cloudstand = "aus"

	// CloudAn heißt: sie schaltet sie ausdrücklich **nicht** ab.
	CloudAn Cloudstand = "an"

	// CloudUngewiss heißt: hier steht nichts dazu.
	CloudUngewiss Cloudstand = "ungewiss"
)

// Cloudbefund ist der Stand mit seiner Fundstelle.
type Cloudbefund struct {
	Stand Cloudstand

	// Beleg sagt, **woher** der Stand kommt — eine Aussage ohne Fundstelle
	// wäre eine Behauptung (§10).
	Beleg string
}

// LiesCloudstand liest die Serverdatei unter einem Zuhause.
//
// **Das Zuhause kommt von außen** und wird nicht erfragt: ein Blatt, das
// `os.UserHomeDir` riefe, läse in jedem Test den Stand **dieses** Rechners
// (§3).
//
// **Ohne Fehlerrückgabe, mit Absicht**: jeder Ausgang ist eine Auskunft. Eine
// fehlende Datei ist kein Fehler, eine kaputte auch nicht — beide heißen
// „ungewiss", und der Beleg sagt, warum.
func LiesCloudstand(zuhause string) Cloudbefund {
	ort := filepath.Join(zuhause, Serverdatei)

	// **Genannt wird die kurze Form und nicht der volle Pfad.**
	//
	// Der Beleg steht in der Oberfläche und geht über MCP und REST hinaus;
	// ein `/Users/<name>/…` darin nennt den Menschen, der hier arbeitet, und
	// den Ort seines Zuhauses. **Gesehen am 2026-09-11 an der laufenden
	// Anwendung** — kein Test hätte es gefunden, weil jeder sein eigenes
	// Zuhause setzt (§3).
	genannt := "~/" + Serverdatei

	roh, fehler := os.ReadFile(filepath.Clean(ort))
	if fehler != nil {
		return Cloudbefund{
			Stand: CloudUngewiss,
			Beleg: fmt.Sprintf("%s ist nicht zu lesen; die Abschaltung kann "+
				"auch über OLLAMA_NO_CLOUD gesetzt sein, und die Umgebung des "+
				"Servers ist von hier nicht einzusehen", genannt),
		}
	}

	var datei map[string]json.RawMessage

	if fehler := json.Unmarshal(roh, &datei); fehler != nil {
		return Cloudbefund{
			Stand: CloudUngewiss,
			Beleg: fmt.Sprintf("%s ist kein lesbares JSON: %v", genannt, fehler),
		}
	}

	roher, nennt := datei[derCloudschalter]
	if !nennt {
		return Cloudbefund{
			Stand: CloudUngewiss,
			Beleg: fmt.Sprintf("%s nennt %q nicht", genannt, derCloudschalter),
		}
	}

	var abgeschaltet bool

	if fehler := json.Unmarshal(roher, &abgeschaltet); fehler != nil {
		return Cloudbefund{
			Stand: CloudUngewiss,
			Beleg: fmt.Sprintf("%s setzt %q auf %s — das ist kein Ja oder Nein",
				genannt, derCloudschalter, string(roher)),
		}
	}

	if abgeschaltet {
		return Cloudbefund{
			Stand: CloudAus,
			Beleg: fmt.Sprintf("%s setzt %q auf true", genannt, derCloudschalter),
		}
	}

	return Cloudbefund{
		Stand: CloudAn,
		Beleg: fmt.Sprintf("%s setzt %q auf false", genannt, derCloudschalter),
	}
}
