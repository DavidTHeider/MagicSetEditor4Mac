package mtgjson

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Fassungsangabe ist, was der Umschlag einer MTG-JSON-Datei über sich sagt.
//
// **Beide Felder dürfen leer sein.** Eine Datei ohne Umschlag hat keine
// Fassung, und das ist der Normalfall für eine von Hand geschriebene (§8).
//
// Gelesen am 2026-08-24 unter `https://mtgjson.com/data-models/meta/`:
//
//	date    — The current release date in ISO 8601 format for the MTGJSON build.
//	version — The current SemVer version for the MTGJSON build appended with
//	          the build date.
type Fassungsangabe struct {
	// Datum ist das Baudatum, **unverändert**.
	//
	// **Keine Umrechnung**: es steht in ISO 8601 da, und wer eine Zeitzone
	// erfände, verschöbe es um bis zu einen Tag (§3).
	Datum string

	// Version ist die SemVer-Fassung samt angehängtem Baudatum.
	Version string
}

// bekannteHauptfassung ist die Hauptfassung, gegen die dieses Programm liest.
//
// **Fünf**, nachgelesen am 2026-08-24: `identifiers` ist mit v5.0.0 gekommen,
// und der Umschlag `{meta, data}` gehört zu derselben Reihe.
const bekannteHauptfassung = 5

// Fassung liest den Umschlag und meldet, wenn er anders aussieht als erwartet.
//
// # Warum das kein Fehler ist, sondern ein Befund
//
// Zwei Gründe, und beide stehen in AGENTS.md:
//
//   - §8: „Fremde Daten sind der Normalfall." Eine Datei aus einer künftigen
//     Hauptfassung soll geöffnet werden, nicht verweigert — was sich geändert
//     haben mag, gehört daneben, nicht an ihre Stelle.
//   - §16: „Pflichtfelder streng, Zusatzfelder tolerant." Die Fassung ist kein
//     Pflichtfeld; sie fehlt in jeder von Hand geschriebenen Datei.
//
// **Was diese Prüfung nicht kann**: sagen, *was* sich geändert hat. Sie sagt
// nur, dass die Datei aus einer Reihe stammt, gegen die hier niemand gelesen
// hat — und das ist der Anlass, die Befunde darunter genauer anzusehen.
func Fassung(inhalt []byte, name string) (Fassungsangabe, []domain.Befund) {
	var umschlag struct {
		Meta *struct {
			Datum   string `json:"date"`
			Version string `json:"version"`
		} `json:"meta"`
	}

	if err := json.Unmarshal(inhalt, &umschlag); err != nil || umschlag.Meta == nil {
		return Fassungsangabe{}, nil
	}

	angabe := Fassungsangabe{Datum: umschlag.Meta.Datum, Version: umschlag.Meta.Version}

	return angabe, fassungsbefunde(angabe, name)
}

// fassungsbefunde meldet eine Fassung, gegen die hier niemand gelesen hat.
func fassungsbefunde(angabe Fassungsangabe, name string) []domain.Befund {
	if angabe.Version == "" {
		return nil
	}

	haupt, lesbar := hauptfassungAus(angabe.Version)
	if !lesbar {
		return []domain.Befund{{
			Datei: name,
			Ort:   "meta.version",
			Art:   domain.ArtUmgebung,
			Meldung: fmt.Sprintf(
				"die Fassungsangabe %q lässt sich nicht als Nummer lesen. Die Datei "+
					"wird trotzdem gelesen; ob ihr Aufbau passt, sagt sie damit nicht.",
				angabe.Version),
		}}
	}

	if haupt == bekannteHauptfassung {
		return nil
	}

	return []domain.Befund{{
		Datei: name,
		Ort:   "meta.version",
		Art:   domain.ArtUmgebung,
		Meldung: fmt.Sprintf(
			"die Datei nennt die Hauptfassung %d; dieses Programm liest gegen %d. "+
				"Sie wird trotzdem gelesen — was sich geändert hat, zeigt sich an den "+
				"übrigen Befunden.", haupt, bekannteHauptfassung),
	}}
}

// hauptfassungAus liest die erste Zahl einer SemVer-Angabe.
//
// **`5.2.2+20260824` ergibt 5.** Der angehängte Bauteil hinter dem `+` gehört
// zur Fassung und nicht zur Nummer; ihn mitzulesen ergäbe keine Zahl.
func hauptfassungAus(version string) (int, bool) {
	erste, _, _ := strings.Cut(version, ".")

	haupt, fehler := strconv.Atoi(strings.TrimSpace(erste))
	if fehler != nil || haupt <= 0 {
		return 0, false
	}

	return haupt, true
}
