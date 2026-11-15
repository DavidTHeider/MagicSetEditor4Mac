//go:build fremdpakete

package flaechen_test

// **Eine Setdatei, die beide Programme lesen** — F05
// (#2247).
//
// Zwei Richtungen, beide gegen das gebaute MSE2 (das Orakel aus §15) und die
// echte Kommandozeile:
//
//  1. **Schreiben → MSE2 öffnen**: `mse oeffnen` liest ein Set, `mse feld`
//     setzt den Namen der ersten Karte auf einen Unicode-Text, `mse sichern`
//     schreibt es. MSE2 liest Original und Abschrift, und jeder Wert jedes
//     Felds jeder Karte — die gerechneten eingeschlossen — ist derselbe, bis
//     auf die eine Änderung, und die steht so da, wie sie gesetzt wurde.
//  2. **MSE2 → MSE4Mac**: MSE2 schreibt dasselbe Set mit `write_set_file`,
//     `mse oeffnen` liest Original und Abschrift, und die Werte sind dieselben.
//
// Dazu je Richtung: welche Schlüssel der Setdatei verlorengehen (oberste und
// zweite Ebene), ob jede Datei neben `set` — die Bilder — byte-gleich
// mitgeht, und wie viele Werte Unicode tragen.
//
// **Die Kohorte**: vier kanonische Sets mit verschiedenen Stilen — Archenemy
// (Schemen), Planechase (Ebenen), Vanguard und White (M15 und Verwandte) —
// und eine **eigene Saat**, die dieser Lauf selbst schreibt: Unicode in drei
// Schriften, ein erfundenes Bild, unbekannte Schlüssel auf Set- und
// Kartenebene, ein mehrzeiliger Wert und eine geschützte spitze Klammer.
// Fremder und eigener Inhalt stehen damit getrennt.
//
// **Was MSE2 selbst verliert, ist kein Befund gegen diese Anwendung** — es
// wird gezählt und berichtet, aber nur, was **diese** Anwendung verliert,
// macht den Lauf rot.
//
// **Was er nicht belegt**: ob MSE2 die Abschrift **zeichnet** wie das
// Original — das belegt der Kanonlauf —, und Schlüssel unterhalb der zweiten
// Ebene.
//
//	MSE_FREMDPAKETE=~/full-magic-pack/data go test -tags fremdpakete \
//	  ./internal/flaechen/ -run Setrundlauf -v

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// unicodeprobe ist der Name, den Richtung 1 der ersten Karte gibt.
const unicodeprobe = "Rundlauf — Grüße 漢字 ✓ עברית"

// rundlaufmass ist, was ein Set im Rundlauf ergab.
type rundlaufmass struct {
	name                    string
	karten, werte, unicode  int
	gerechnet               int
	abweichungEins          []string
	abweichungZwei          []string
	verlorenEins            []string
	verlorenZweiMSE2        []string
	dateien                 int
	dateienFehlenEins       []string
	dateienAndersEins       []string
	dateienFehlenZweiMSE2   []string
	fremdeAbweichungZweiMSE []string

	// stiloption ist die setweite Option, die Richtung 1 setzt, und stilEins,
	// ob MSE2 sie aus der Abschrift liest — leer, wenn ja.
	stiloption string
	stilEins   []string

	// warnungenEins sind MSE2s Warnungen über die Abschrift, die es über
	// das Original nicht gab.
	warnungenEins []string
}

func TestSetrundlaufMitMSE2(t *testing.T) {
	pack := os.Getenv("MSE_FREMDPAKETE")
	if pack == "" {
		t.Skip("MSE_FREMDPAKETE ist nicht gesetzt")
	}

	if _, fehler := os.Stat(orakelprogramm); fehler != nil {
		t.Skipf("das Orakel ist nicht gebaut: %v", fehler)
	}

	umgebung := rundlaufumgebung(t, pack)
	kanon := filepath.Join(os.Getenv("HOME"), "Library", "Application Support",
		"MagicSetEditor4Mac", "kanonische-sets")

	kohorte := []struct{ name, pfad string }{
		{"eigene Saat (magic)", eigeneSaat(t, "magic", "m15", magicsaat)},
		{"eigene Saat (saga)", eigeneSaat(t, "saga", "standard", sagasaat)},
		{"Archenemy", filepath.Join(kanon, "Archenemy.mse-set")},
		{"Planechase", filepath.Join(kanon, "Planechase.mse-set")},
		{"Vanguard", filepath.Join(kanon, "Vanguard.mse-set")},
		{"White", filepath.Join(kanon, "White.mse-set")},
	}

	var bericht []rundlaufmass

	for _, set := range kohorte {
		t.Run(set.name, func(t *testing.T) {
			if _, fehler := os.Stat(set.pfad); fehler != nil {
				t.Skipf("%s liegt nicht da: %v", set.pfad, fehler)
			}

			mass := rundlauf(t, umgebung, pack, set.pfad)
			mass.name = set.name
			bericht = append(bericht, mass)

			pruefeRundlauf(t, mass)
		})
	}

	berichtSchreiben(t, bericht)
}

// rundlauf fährt beide Richtungen über ein Set.
func rundlauf(t *testing.T, umgebung []string, pack, quelle string) rundlaufmass {
	t.Helper()

	arbeit := t.TempDir()
	original := paketinhalt(t, quelle)
	spiel := spielfelder(t, pack, string(original["set"]))

	// **Richtung 1**: lesen, ändern, schreiben — und MSE2 fragen.
	eins := mitEigenemBestand(umgebung, arbeit, "eins")
	mussMse(t, eins, "pakete")
	setID := setkennungAus(t, mussMse(t, eins, "oeffnen", quelle, "--json"))
	kennungen, werte := kartenwerte(t, eins, setID)

	if len(kennungen) == 0 {
		t.Fatal("das Set hat keine Karte")
	}

	felder := feldliste(spiel, werte)
	mussMse(t, eins, "feld", setID, kennungen[0], "name", unicodeprobe)

	// **Und eine setweite Stiloption** — sie kam bis zum 2026-10-10 gar nicht
	// in die Datei (`internal/service/setschreiberstil_test.go`).
	option, optionswert := stiloptionWaehlen(t, eins, setID)
	if option != "" {
		mussMse(t, eins, "stiloption", setID, "setzen", option, optionswert)
	}

	abschrift := filepath.Join(arbeit, "abschrift.mse-set")
	mussMse(t, eins, "sichern", setID, abschrift)

	vorherGefragt := orakelAuszug(t, quelle, arbeit, felder, "", "mse2.mse-set")
	nachherGefragt := orakelAuszug(t, abschrift, arbeit, felder, option, "")
	vorher, nachher := vorherGefragt.karten, nachherGefragt.karten

	mass := rundlaufmass{karten: len(vorher), stiloption: option}
	mass.abweichungEins = vergleiche(vorher, nachher, func(karte int, feld string) (string, bool) {
		return unicodeprobe, karte == 0 && feld == "name"
	})
	mass.warnungenEins = neueWarnungen(vorherGefragt.warnungen, nachherGefragt.warnungen)

	if option != "" && nachherGefragt.stil != optionswert {
		mass.stilEins = []string{fmt.Sprintf("styling.%s ist %q, gesetzt war %q", option, nachherGefragt.stil, optionswert)}
	}

	// **Gerechnet ist, was MSE2 nennt und die Datei nicht trägt** — die
	// Skripte des Spiels (`card color`, `type`). Sie stehen im Vergleich der
	// Richtung 1, denn MSE2 rechnet sie aus der Abschrift neu.
	for i, karte := range vorher {
		for feld, wert := range karte {
			if _, gespeichert := werte[i][msepaket.Entkanonisch(feld)]; !gespeichert && wert != unlesbar {
				mass.gerechnet++
			}
		}
	}

	// **Richtung 2**: was MSE2 schreibt, liest diese Anwendung gleich.
	zwei := mitEigenemBestand(umgebung, arbeit, "zwei")
	mussMse(t, zwei, "pakete")
	_, werteMSE2 := kartenwerte(t, zwei, setkennungAus(t,
		mussMse(t, zwei, "oeffnen", filepath.Join(arbeit, "mse2.mse-set"), "--json")))

	for _, karte := range werte {
		for _, wert := range karte {
			mass.werte++
			if hatUnicode(wert) {
				mass.unicode++
			}
		}
	}

	vonMSE2Gelesen := orakelAuszug(t, filepath.Join(arbeit, "mse2.mse-set"), arbeit, felder, "", "")
	eigen, fremd := vergleicheWerte(werte, werteMSE2, vonMSE2Gelesen.karten)
	mass.abweichungZwei, mass.fremdeAbweichungZweiMSE = eigen, fremd

	// **Die Schlüssel und die Dateien.**
	geschrieben := paketinhalt(t, abschrift)
	vonMSE2 := paketinhalt(t, filepath.Join(arbeit, "mse2.mse-set"))
	mass.verlorenEins = verloreneSchluessel(schluesselpfade(string(original["set"])), schluesselpfade(string(geschrieben["set"])))
	mass.verlorenZweiMSE2 = verloreneSchluessel(schluesselpfade(string(original["set"])), schluesselpfade(string(vonMSE2["set"])))
	mass.dateien, mass.dateienFehlenEins, mass.dateienAndersEins = dateienVergleichen(original, geschrieben)
	_, mass.dateienFehlenZweiMSE2, _ = dateienVergleichen(original, vonMSE2)

	return mass
}

// stiloptionWaehlen sucht eine setweite Stiloption, die sich setzen lässt —
// zuerst eine Textoption (Wert „7"), sonst eine Auswahl mit einer anderen
// Möglichkeit als der jetzigen. Kennt der Stil keine, bleibt der Name leer.
func stiloptionWaehlen(t *testing.T, umgebung []string, setID string) (name, wert string) {
	t.Helper()

	var antwort struct {
		Optionen []struct {
			Name           string `json:"name"`
			Art            string `json:"art"`
			Wert           string `json:"wert"`
			Moeglichkeiten []struct {
				Wert string `json:"wert"`
			} `json:"moeglichkeiten"`
		} `json:"optionen"`
	}

	if fehler := json.Unmarshal([]byte(mussMse(t, umgebung, "stiloption", setID, "--json")), &antwort); fehler != nil {
		t.Fatalf("stiloption: %v", fehler)
	}

	for _, option := range antwort.Optionen {
		if option.Art == "text" {
			return option.Name, "7"
		}
	}

	for _, option := range antwort.Optionen {
		for _, moeglich := range option.Moeglichkeiten {
			if option.Art == "choice" && moeglich.Wert != option.Wert {
				return option.Name, moeglich.Wert
			}
		}
	}

	return "", ""
}
