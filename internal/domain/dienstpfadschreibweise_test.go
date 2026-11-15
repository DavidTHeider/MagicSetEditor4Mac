package domain_test

// **Der Pfadvertrag liest einen Namen so, wie der Dekodierer ihn liest** —
// #2073.
// `encoding/json` legt `"Ziel"` in das Feld mit der Marke `json:"ziel"`, auch
// mit `DisallowUnknownFields`; der Vertrag sah nur `"ziel"`, und ein Pfad
// unter dem groß geschriebenen Namen kam ungeprüft an den Weg.

import (
	"bytes"
	"encoding/json"
	"slices"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// dekodierterRumpf hat, was ein Weg dekodiert: zwei Pfadfelder, eines mit
// einem Namen, der nicht nach einem Pfad klingt, und eines, das keinen trägt.
type dekodierterRumpf struct {
	Ziel        string `json:"ziel"`
	Verzeichnis string `json:"verzeichnis"`
	Quelle      string `json:"quelle"`
	Name        string `json:"name"`
}

// **Die Gegenprobe ist der Dekodierer selbst**, nicht eine Liste von
// Schreibweisen, die jemand für vollständig hält: was er in ein Pfadfeld
// legt, muss der Vertrag sehen — und was er in `name` legt, nicht. Darunter
// »ſ«, das er zu »s« faltet.
func TestPfadeInSiehtJedeSchreibweiseDesDekodierers(t *testing.T) {
	t.Parallel()

	for _, name := range []string{
		"ziel", "Ziel", "ZIEL", "zIeL", "verzeichnis", "VerzeichNis", "verzeichniſ",
		"quelle", "Quelle", "QUELLE", "name", "Name", "ziele", "zie l", "pfad",
	} {
		roh, _ := json.Marshal(map[string]string{name: "/fremd"}) //nolint:errcheck,errchkjson // eine Abbildung von Zeichenketten

		var rumpf dekodierterRumpf

		entzifferer := json.NewDecoder(bytes.NewReader(roh))
		entzifferer.DisallowUnknownFields()

		if entzifferer.Decode(&rumpf) != nil {
			continue
		}

		traegtPfad := rumpf.Ziel != "" || rumpf.Verzeichnis != "" || rumpf.Quelle != ""
		gesehen := slices.Contains(domain.PfadeIn(roh, "quelle"), "/fremd")

		if gesehen != traegtPfad {
			t.Errorf("%q: der Dekodierer legt den Wert in ein Pfadfeld: %v — der Vertrag sieht ihn: %v",
				name, traegtPfad, gesehen)
		}
	}
}

// **Zweimal derselbe Name in zwei Schreibweisen** — der Dekodierer nimmt den
// letzten, und welcher das ist, entscheidet der Rumpf. Der Vertrag prüft
// deshalb beide.
func TestPfadeInPrueftJedeSchreibweiseEinesRumpfs(t *testing.T) {
	t.Parallel()

	for roh, erwartet := range map[string]string{
		`{"ziel":"/gut","Ziel":"/fremd"}`:   "/fremd|/gut",
		`{"Ziel":"/fremd","ziel":"/gut"}`:   "/fremd|/gut",
		`{"PFAD":"/a","bildPfad":"/b"}`:     "/a|/b",
		`{"angriffe":[{"ziel":"spieler"}]}`: "",
	} {
		if gelesen := strings.Join(domain.PfadeIn([]byte(roh)), "|"); gelesen != erwartet {
			t.Errorf("%s: %q, erwartet %q", roh, gelesen, erwartet)
		}
	}
}

// **Ein Pfad unter einem Namen, der nach keinem klingt**, gehört dem Weg,
// der ihn so nennt — `quelle` beim Einbinden eines Pakets ist ein
// Verzeichnis, `quelle` eines Preises ein Anbieter.
func TestPfadeInKenntDieWeiterenNamenEinesWegs(t *testing.T) {
	t.Parallel()

	roh := []byte(`{"Quelle":"/q","ziel":"/z","ort":"/o"}`)

	for weitere, erwartet := range map[string]string{
		"":            "/z",
		"quelle":      "/q|/z",
		"quelle|ort":  "/q|/o|/z",
		"ORT|unbenut": "/o|/z",
	} {
		namen := strings.Split(weitere, "|")
		if weitere == "" {
			namen = nil
		}

		if gelesen := strings.Join(domain.PfadeIn(roh, namen...), "|"); gelesen != erwartet {
			t.Errorf("weitere %q: %q, erwartet %q", weitere, gelesen, erwartet)
		}
	}
}
