// Package mtgjson liest und schreibt Kartendefinitionen im Format **MTG JSON**.
//
// Es ist eine Fachbibliothek und damit ein Blatt (AGENTS.md §2): es kennt
// `internal/domain` und sonst nichts aus diesem Projekt.
//
// # Warum MTG JSON und nicht ein eigenes Schema
//
// Nachgesehen am 2026-08-09: `magic-set-generator` hat **kein** eigenes
// Kartenschema. Es liest MTG JSON — eine Datei im Format *AllSets* oder ein
// Verzeichnis mit *Individual Set*-Dateien — und schreibt daraus eine
// `.mse-set`. Wer „das JSON-Format" sagt, meint dieses.
//
// # Woher die Daten kommen dürfen
//
// MTG-JSON-Daten sind Kartentexte und Namen von Wizards of the Coast. Sie kommen
// von außen (§11); im Repository liegen nur selbst geschriebene Beispieldateien.
package mtgjson

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Quelle ist eine gelesene MTG-JSON-Quelle — eine Datei oder ein Verzeichnis.
type Quelle struct {
	Pfad string
	Sets []*Set

	// Befunde sind die Stellen, an denen die Zuordnung nicht weiterwusste.
	//
	// **Sie schließen den Erfolg nicht aus** (§8): eine Datei mit 300 lesbaren
	// Karten und zwei unbekannten Feldern wird gelesen, und die zwei stehen
	// hier.
	Befunde []domain.Befund
}

// Set ist ein Set aus MTG JSON.
type Set struct {
	Code   string
	Name   string
	Karten []*Karte
}

// Karte ist eine Karte aus MTG JSON, bereits auf Kartenfelder abgebildet.
type Karte struct {
	// Werte sind die Felder, die die Zuordnung kennt.
	Werte []domain.Value

	// Rest sind die Schlüssel, die sie nicht kennt — **roh mitgeführt**, damit
	// ein Rundlauf sie nicht verliert. Der Text ist das JSON selbst, samt
	// Anführungszeichen: nur so ist beim Schreiben `"1"` von `1` zu
	// unterscheiden.
	Rest []domain.Value
}

// rohSet ist ein Set, wie es in der Datei steht.
type rohSet struct {
	Code   string                       `json:"code"`
	Name   string                       `json:"name"`
	Karten []map[string]json.RawMessage `json:"cards"`
}

// ort sagt, wie dieses Set in einem Befund heißt.
//
// **Ein Befund, der sein Set nicht nennt, ist keiner** (§8: „gibt für jeden
// Fehler die Stelle her — Datei, Zeile, Feld"). Bei dreihundert Sets in einer
// Datei ist die Stelle das Set, und ein Ort, der mit `" → "` beginnt, sagt
// genauso wenig wie gar keiner.
//
// **Dieselbe Reihenfolge wie [ortVon] in `gueltigkeit.go`** — Kürzel, dann
// Name, dann der feste Satz. Der Zweitblick über Phase 4.2 hat angestrichen,
// dass die eine Stelle den Ersatz bekam und die drei hier nicht: der Fall
// `{"":{}}` war behoben, die **Klasse** nicht.
func (r rohSet) ort() string {
	return benannt(r.Code, benannt(r.Name, ohneCodeUndNamen))
}

// Lies liest eine MTG-JSON-Quelle: eine Datei oder ein Verzeichnis.
//
// **Beides, weil MTG JSON beides ist.** Eine *AllSets*-Datei führt alle Sets in
// einer Datei; wer einzelne Sets bezieht, bekommt ein Verzeichnis mit einer
// Datei je Set. `magic-set-generator` nimmt genauso beides entgegen.
func Lies(pfad string, zuordnung Zuordnung) (*Quelle, error) {
	angaben, err := os.Stat(pfad)
	if err != nil {
		return nil, fmt.Errorf("MTG JSON lesen: %w", err)
	}

	if angaben.IsDir() {
		return liesVerzeichnis(pfad, zuordnung)
	}

	quelle := &Quelle{Pfad: pfad}
	if err := quelle.dateiLesen(pfad, zuordnung); err != nil {
		return nil, err
	}

	return quelle, nil
}

// LiesInhalt liest MTG JSON aus dem Speicher statt aus einer Datei.
//
// **Für den Rundlauf, und nicht nur für ihn.** Was geschrieben wurde, muss sich
// wieder lesen lassen, ohne den Umweg über eine Datei — sonst prüfte der
// Rundlauftest nebenbei das Dateisystem. `name` steht in den Befunden an der
// Stelle, an der sonst der Pfad stünde.
func LiesInhalt(inhalt []byte, name string, zuordnung Zuordnung) (*Quelle, error) {
	quelle := &Quelle{Pfad: name}

	rohe, err := setsAus(inhalt, name)
	if err != nil {
		return nil, err
	}

	// **Die Schema- und Fassungsprüfung läuft mit** (Zweitblick über Phase V,
	// 2026-08-24).
	//
	// Sie war gebaut, vollständig getestet — und **ohne Aufrufer im
	// Importweg**. Wer eine Datei öffnete, bekam nie einen Schemabefund und
	// nie den Hinweis auf eine unbekannte Hauptfassung. Dieselbe Klasse wie
	// #852:
	// volle Abdeckung, kein Weg dorthin.
	//
	// **Sie steht vor der Zuordnung**, weil sie den Rahmen beschreibt: eine
	// Datei aus einer künftigen Hauptfassung erklärt, warum die Befunde
	// darunter aussehen, wie sie aussehen.
	quelle.Befunde = append(quelle.Befunde, Pruefe(inhalt, name)...)

	for _, roh := range rohe {
		set, befunde := zuordnung.set(roh, name)

		quelle.Sets = append(quelle.Sets, set)
		quelle.Befunde = append(quelle.Befunde, befunde...)
	}

	return quelle, nil
}

// liesVerzeichnis liest jede `.json`-Datei eines Verzeichnisses.
//
// In alphabetischer Reihenfolge, damit zwei Läufe dasselbe ergeben — die
// Reihenfolge des Dateisystems ist keine Zusage.
//
// **Was keine `.json` ist, wird übergangen und nicht bemängelt.** Ein bezogenes
// Verzeichnis führt eine LIESMICH und eine Prüfsummendatei; daran zu scheitern
// hieße, dass niemand es öffnen kann, ohne vorher aufzuräumen.
func liesVerzeichnis(pfad string, zuordnung Zuordnung) (*Quelle, error) {
	eintraege, err := os.ReadDir(pfad)
	if err != nil {
		return nil, fmt.Errorf("MTG JSON lesen: %w", err)
	}

	namen := make([]string, 0, len(eintraege))

	for _, eintrag := range eintraege {
		if !eintrag.IsDir() && strings.EqualFold(filepath.Ext(eintrag.Name()), ".json") {
			namen = append(namen, eintrag.Name())
		}
	}

	slices.Sort(namen)

	quelle := &Quelle{Pfad: pfad}

	for _, name := range namen {
		if err := quelle.dateiLesen(filepath.Join(pfad, name), zuordnung); err != nil {
			return nil, err
		}
	}

	if len(quelle.Sets) == 0 {
		return nil, domain.NewFieldError("pfad", domain.SchluesselValidierung,
			"%q enthält keine einzige lesbare JSON-Datei.", pfad)
	}

	return quelle, nil
}

// dateiLesen liest eine Datei und hängt ihre Sets an die Quelle.
//
// **Der Pfad kommt von außen, und das ist der Zweck.** Wer eine MTG-JSON-Datei
// öffnet, sagt, welche — über den nativen Dateidialog oder über eine
// Verzeichnisliste, die dieses Paket selbst gelesen hat. Eine Beschränkung auf
// ein Wurzelverzeichnis (`os.Root`) gäbe es nicht her: die Datei liegt dort, wo
// die Nutzenden sie haben, und nicht unter einem Ort, den wir bestimmen.
func (q *Quelle) dateiLesen(pfad string, zuordnung Zuordnung) error {
	inhalt, err := os.ReadFile(pfad) // #nosec G304 -- siehe darüber
	if err != nil {
		return fmt.Errorf("MTG JSON lesen: %w", err)
	}

	gelesen, err := LiesInhalt(inhalt, pfad, zuordnung)
	if err != nil {
		return err
	}

	q.Sets = append(q.Sets, gelesen.Sets...)
	q.Befunde = append(q.Befunde, gelesen.Befunde...)

	return nil
}

// set macht aus einem rohen Set eines mit abgebildeten Karten.
//
// **Die Befunde über unbekannte Felder werden über das Set gebündelt.** Eine
// echte MTGJSON-Datei führt an fast jeder Karte fast jedes unbekannte Feld; je
// Karte gemeldet ergibt das für 295 Karten 7.669 gleichlautende Zeilen — und
// wer die sieht, liest die eine nicht mehr, die von einem wirklichen Verlust
// handelt. Gebündelt sind es 26, jede mit der Zahl der betroffenen Karten.
//
// Der Wert bleibt dabei an **jeder** Karte: gebündelt wird die Meldung, nicht
// das Mitführen.
func (z Zuordnung) set(roh rohSet, pfad string) (*Set, []domain.Befund) {
	set := &Set{Code: roh.Code, Name: roh.Name}

	var (
		befunde     []domain.Befund
		gebuendelt  = map[string]*buendel{}
		reihenfolge []string
	)

	for nummer, rohKarte := range roh.Karten {
		karte, hinweise := z.Karte(rohKarte)

		set.Karten = append(set.Karten, karte)

		for _, hinweis := range hinweise {
			if hinweis.JeKarte {
				befunde = append(befunde, domain.Befund{
					Datei: pfad,
					Ort: roh.ort() + " → Karte " + strconv.Itoa(nummer+1) +
						" → " + benannt(hinweis.Feld, feldOhneNamen),
					Meldung: hinweis.Meldung,
				})

				continue
			}

			vorhanden, schon := gebuendelt[hinweis.Feld]
			if !schon {
				gebuendelt[hinweis.Feld] = &buendel{meldung: hinweis.Meldung, erste: nummer + 1, karten: 1}
				reihenfolge = append(reihenfolge, hinweis.Feld)

				continue
			}

			vorhanden.karten++
		}
	}

	// **Die Seiten einer Karte werden zusammengelegt** (Zweitblick über Phase
	// V, 2026-08-24). MSE führt eine mehrseitige Karte als **eine** Karte mit
	// zwei Feldsätzen; MTGJSON liefert je Seite ein eigenes Kartenobjekt.
	//
	// **Nach der Zuordnung und vor den Befunden**: die Zusammenlegung arbeitet
	// auf fertig abgebildeten Karten, und ihre eigenen Hinweise sollen in
	// derselben Liste landen wie die übrigen.
	gelegt, seitenhinweise := fuehreSeitenZusammen(set.Karten, roh.Karten)
	set.Karten = gelegt

	for _, hinweis := range seitenhinweise {
		befunde = append(befunde, domain.Befund{
			Datei:   pfad,
			Ort:     roh.ort() + " → " + benannt(hinweis.Feld, feldOhneNamen),
			Meldung: hinweis.Meldung,
		})
	}

	for _, feld := range reihenfolge {
		eines := gebuendelt[feld]

		befunde = append(befunde, domain.Befund{
			Datei: pfad,
			Ort:   roh.ort() + " → " + benannt(feld, feldOhneNamen),
			Meldung: fmt.Sprintf("%s — %d Karten führen es, zuerst Karte %d",
				eines.meldung, eines.karten, eines.erste),
		})
	}

	return set, befunde
}

// buendel zählt die Karten, an denen derselbe Hinweis aufgetreten ist.
type buendel struct {
	meldung string
	erste   int
	karten  int
}

// setsAus holt die Sets aus dem Inhalt einer Datei.
//
// # Drei Gestalten, eine Auswertung
//
// MTGJSON 5 legt einen Umschlag um alles: `{"meta": …, "data": …}`. Belegt ist
// er dort, wo ihn jemand benutzt — `magic.mse-game` nennt in `json_paths` den
// Pfad `data:cards:` (docs/FORMATE.md). Darunter steht **entweder** ein
// einzelnes Set (*Individual Set*) **oder** eine Abbildung Setcode → Set
// (*AllSets*). Ältere Dateien haben den Umschlag nicht.
//
// Unterschieden wird an der Gestalt, nicht an einer Versionsangabe: eine Datei
// von Hand ist der Normalfall (§8), und `meta.version` fehlt dort.
func setsAus(inhalt []byte, pfad string) ([]rohSet, error) {
	var umschlag struct {
		Daten json.RawMessage `json:"data"`
	}

	if err := json.Unmarshal(inhalt, &umschlag); err != nil {
		return nil, fmt.Errorf("MTG JSON lesen (%s): %w", pfad, err)
	}

	kern := umschlag.Daten
	if kern == nil {
		kern = inhalt
	}

	var eines rohSet
	if err := json.Unmarshal(kern, &eines); err == nil && eines.Code != "" {
		return []rohSet{eines}, nil
	}

	var viele map[string]rohSet
	if err := json.Unmarshal(kern, &viele); err != nil {
		return nil, fmt.Errorf("MTG JSON lesen (%s): %w", pfad, err)
	}

	codes := make([]string, 0, len(viele))
	for code := range viele {
		codes = append(codes, code)
	}

	slices.Sort(codes)

	sets := make([]rohSet, 0, len(codes))

	for _, code := range codes {
		set := viele[code]

		// Der Schlüssel der Abbildung ist der Setcode. Ein Set, das ihn im
		// Rumpf wiederholt, gewinnt — dort steht er in der Schreibweise, die
		// die Datei selbst benutzt.
		if set.Code == "" {
			set.Code = code
		}

		sets = append(sets, set)
	}

	return sets, nil
}
