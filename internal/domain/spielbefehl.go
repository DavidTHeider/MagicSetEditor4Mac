package domain

// **Was eine Partie bewegt, und was dabei herauskommt** — Werkstück G1 von
// Phase 5.5
// (#1458).
//
// # Die Folge ist maßgeblich, die Uhr ist ein Metadatum
//
// **Das Issue sagt es wörtlich**, und es ist mehr als eine Feinheit: zwei
// Rechner haben zwei Uhren, und keine davon darf entscheiden, was zuerst
// geschah. Die Folge zählt der Wirt, und sie geht in den Abdruck ein; der
// Zeitstempel steht daneben, **damit ein Mensch etwas ablesen kann**, und
// rechnet nicht mit.
//
// # Zwei Nachweise, und das ist der ganze Punkt
//
// **Ein gemeinsamer Vollzustandshash zwänge dazu, fremde Hände offenzulegen**
// — wer ihn nachrechnen wollte, bräuchte alles, was darin steckt. Deshalb
// gibt es zwei: der **öffentliche Strom** trägt eine Kette, die jeder prüfen
// kann, und jede **private Ergänzung** trägt ihren eigenen Abdruck, den nur
// ihr Empfänger nachrechnet.
//
// **Was das kostet, gehört ausgesprochen**: wer nur den öffentlichen Strom
// hat, weiß nicht, **ob** eine Ergänzung dazugehörte — nur, dass der Strom
// selbst unverändert ist. Die Alternative wäre, dass die Integrität das
// Geheimnis aufhebt, und das ist teurer.
//
// # Warum die Kodierung hier steht und nicht bei `internal/peer`
//
// **Weil `internal/domain` nichts aus dem Projekt importiert** (§2) — und
// `peer` kennt `domain`, also wäre es ein Kreis. Was hier steht, ist deshalb
// bewusst klein: `encoding/json` schreibt Struct-Felder in
// Deklarationsreihenfolge und Map-Schlüssel sortiert, und mehr braucht ein
// Abdruck nicht.

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// Spielbefehl ist ein Wunsch an eine Partie.
//
// **Ein Wunsch und keine Wirkung**: was daraus wird, entscheidet der Wirt,
// und was dabei herauskommt, sind [Ereignis]se.
type Spielbefehl struct {
	// Kennung ist der Schutz gegen die doppelte Wirkung.
	//
	// **Sie kommt vom Absender**, denn nur er weiß, ob er noch einmal
	// dasselbe meint oder etwas Neues. Ein zweiter Ruf mit derselben Kennung
	// erzeugt keinen zweiten Zug.
	Kennung string

	PartieID string

	// Absender ist eine Sitzkennung — **nicht** ein Nutzer.
	Absender string

	// Art sagt, was gewollt ist — `karte.ziehen`, `objekt.tappen`, …
	Art string

	// Erwartet ist die Fassung, gegen die gerechnet wird.
	//
	// **Null heißt: gegen die, die gerade gilt.** Das ist der bequeme Fall
	// für eine Fläche ohne Nebenläufigkeit; wer sicher gehen will, nennt
	// eine.
	Erwartet int64

	Nutzlast map[string]any

	// Gestellt ist die Wanduhr des Absenders — **Metadatum**, siehe Kopf.
	Gestellt time.Time
}

// Pruefe sagt, ob dieser Befehl überhaupt zugestellt werden kann.
func (b Spielbefehl) Pruefe() error {
	switch {
	case strings.TrimSpace(b.Kennung) == "":
		return fmt.Errorf(
			"%w: ein Befehl ohne Kennung lässt sich nicht von seiner "+
				"Wiederholung unterscheiden", ErrValidation)

	case strings.TrimSpace(b.PartieID) == "":
		return fmt.Errorf(
			"%w: ein Befehl ohne Partie weiß nicht, wen er meint",
			ErrValidation)

	case strings.TrimSpace(b.Absender) == "":
		return fmt.Errorf(
			"%w: ein Befehl ohne Absender gehört zu keinem Sitz",
			ErrValidation)

	case strings.TrimSpace(b.Art) == "":
		return fmt.Errorf(
			"%w: ein Befehl ohne Art sagt nicht, was geschehen soll",
			ErrValidation)

	case b.Erwartet < 0:
		return fmt.Errorf(
			"%w: die erwartete Fassung ist %d — die gab es nie",
			ErrValidation, b.Erwartet)
	}

	return nil
}

// Ereignis ist, was wirklich geschehen ist.
type Ereignis struct {
	PartieID string

	// Folge ist die Nummer im Strom — **sie ist maßgeblich** (siehe Kopf).
	Folge int64

	Art string

	// Oeffentlich ist, was jeder sehen darf — auch ein Zuschauer.
	Oeffentlich map[string]any

	// Privat sind die empfängergebundenen Ergänzungen.
	//
	// **Sie gehen nicht in [Ereignis.Abdruck] ein**, und das ist die Zusage
	// aus dem Kopf dieser Datei.
	Privat []Ergaenzung

	// Kette ist der Abdruck dieses Ereignisses samt seinem Vorgänger.
	//
	// **Er rechnet sich nicht selbst mit** — ein Feld, das in seinen eigenen
	// Hash einginge, ließe sich nie ausrechnen.
	Kette string

	// Geschehen ist die Wanduhr des Wirts — **Metadatum**.
	Geschehen time.Time
}

// Ergaenzung ist, was nur einer sehen darf.
type Ergaenzung struct {
	// Empfaenger ist eine Sitzkennung.
	Empfaenger string

	Inhalt map[string]any
}

// abdruckbaresEreignis ist die Gestalt, über die gerechnet wird.
//
// **Eine eigene Struktur und kein Zeroing der Felder.** Wer stattdessen eine
// Kopie mit geleerten Feldern hashte, hätte beim nächsten neuen Feld
// stillschweigend etwas dazugenommen — hier steht ausdrücklich da, **was**
// zählt.
type abdruckbaresEreignis struct {
	Vorher      string         `json:"vorher"`
	PartieID    string         `json:"partie"`
	Folge       int64          `json:"folge"`
	Art         string         `json:"art"`
	Oeffentlich map[string]any `json:"oeffentlich"`
}

// Abdruck rechnet den Kettenabdruck über den öffentlichen Teil.
//
// **`vorher` ist der Abdruck des Vorgängers** — beim ersten Ereignis leer.
// Ohne ihn wäre es keine Kette, sondern eine Liste von Abdrücken, und ein
// entferntes Ereignis fiele nicht auf.
//
// # Warum er scheitern kann
//
// **Weil `Oeffentlich` ein `map[string]any` ist und `any` alles sein darf.**
// Wer einen Kanal oder eine Funktion hineinlegt, hat etwas hineingelegt, das
// sich nicht kodieren lässt — und das ist kein unmöglicher Fall, sondern
// einer, den ein Aufrufer aus fremder Eingabe bauen kann.
//
// **Ein `panic` wäre hier falsch**: er machte aus einem Eingabefehler einen
// Absturz, und ein Zweig, den kein Lauf betritt, ist ohnehin keiner (§3).
func (e Ereignis) Abdruck(vorher string) (string, error) {
	return kanonischerAbdruck(abdruckbaresEreignis{
		Vorher:      vorher,
		PartieID:    e.PartieID,
		Folge:       e.Folge,
		Art:         e.Art,
		Oeffentlich: e.Oeffentlich,
	}, "ein Ereignis")
}

// kanonischerAbdruck kodiert kanonisch und hasht.
//
// **Nicht `abdruckVon`** — den Namen trägt schon der Abdruck des
// Retrievalindex (`indexmanifest.go`), und er fasst Zeichenketten zusammen
// statt zu kodieren. Zwei Funktionen desselben Namens wären zwei Antworten
// auf verschiedene Fragen.
//
// **Eine Stelle für beide Abdrücke.** Zwei Kodierungen für dieselbe Frage
// liefen bei der ersten Änderung auseinander, und der Unterschied fiele erst
// auf, wenn eine Kette nicht mehr prüft.
func kanonischerAbdruck(was any, benennung string) (string, error) {
	roh, fehler := json.Marshal(was)
	if fehler != nil {
		return "", fmt.Errorf(
			"%w: %s ließ sich nicht kodieren — im Inhalt steht etwas, das "+
				"kein JSON werden kann: %s", ErrValidation, benennung,
			fehler.Error())
	}

	summe := sha256.Sum256(roh)

	return hex.EncodeToString(summe[:]), nil
}

// abdruckbareErgaenzung ist die Gestalt, über die eine Ergänzung rechnet.
//
// **Sie hat dieselben Felder wie [Ergaenzung], und das ist Absicht** — was
// zählt, steht ausdrücklich da statt sich aus dem Typ zu ergeben. Ein
// zusätzliches Feld an der Ergänzung ginge sonst still in den Nachweis ein,
// und niemand hätte es entschieden. `staticcheck` schlägt an dieser Stelle
// eine Umwandlung vor (S1016); sie wäre kürzer und nähme genau diese Zusage
// weg.
type abdruckbareErgaenzung struct {
	Empfaenger string         `json:"empfaenger"`
	Inhalt     map[string]any `json:"inhalt"`
}

// Abdruck rechnet den eigenen Nachweis einer Ergänzung.
//
// **Der Empfänger geht ein.** Dieselbe Karte an jemand anderen ist eine
// andere Ergänzung — sonst ließe sich eine an den falschen Sitz hängen, ohne
// dass der Nachweis es merkte.
func (e Ergaenzung) Abdruck() (string, error) {
	//nolint:staticcheck // der Grund steht bei abdruckbareErgaenzung
	return kanonischerAbdruck(abdruckbareErgaenzung{
		Empfaenger: e.Empfaenger,
		Inhalt:     e.Inhalt,
	}, "eine Ergänzung")
}
