package laufbeleg

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// Format ist die Fassung dieser Belegform. Wer ein Feld ändert oder
// entfernt, hebt sie — ein Leser weist eine fremde ab, statt sie falsch zu
// lesen.
const Format = 1

// Kopf ist, was der Lauf über sich selbst weiß — Stand, Werkzeuge und Zeit
// (AK-01).
type Kopf struct {
	Lauf   string `json:"lauf"`
	Commit string `json:"commit"`
	// Inhalt ist der Hash über das, was der Lauf misst — `git ls-files -s`
	// über die Pfade des Plans. **Er und nicht der Commit sagt, ob ein Beleg
	// noch passt**: jeder PR wird per Squash gemergt, und danach liegt der
	// gemessene Commit in keiner Geschichte mehr.
	Inhalt     string            `json:"inhalt"`
	BaumSauber bool              `json:"baumSauber"`
	Werkzeuge  map[string]string `json:"werkzeuge"`
	// Schalter sind die Umgebungswerte, die ändern, **was** gemessen wird —
	// `MSE_STICHPROBE` etwa —, mit leerem Wert, wenn sie nicht gesetzt waren.
	// Welche es sind, sagt der Plan.
	Schalter  map[string]string `json:"schalter,omitempty"`
	Anfang    time.Time         `json:"anfang"`
	Ende      time.Time         `json:"ende"`
	Befehl    []string          `json:"befehl"`
	Rueckgabe int               `json:"rueckgabe"`
}

// Beleg ist der maschinenlesbare Nachweis eines Laufs — QA-01.AK-01: Commit,
// Werkzeuge, Anfang und Ende, Eingaben mit Hash, erwartete und gemessene
// Fälle, Fehler, Sprünge und Artefakte.
type Beleg struct {
	Format int `json:"format"`
	Kopf
	Status        Status            `json:"status"`
	Grund         string            `json:"grund,omitempty"`
	Eingaben      []Eingabe         `json:"eingaben"`
	Erwartet      int               `json:"erwartet"`
	Gemessen      int               `json:"gemessen"`
	Faelle        []Fallergebnis    `json:"faelle"`
	Weitere       Zaehlung          `json:"weitere"`
	Fehler        []string          `json:"fehler"`
	Uebersprungen []string          `json:"uebersprungen"`
	Artefakte     []Eingabe         `json:"artefakte"`
	Kennzahlen    map[string]string `json:"kennzahlen"`
}

// Belegen setzt den Beleg aus seinen Teilen zusammen.
//
// **Eine fehlende Pflichteingabe macht ihn BLOCKED, ein fehlendes Artefakt
// FAIL** — und kein grüner Fall überstimmt das (AK-02). **Und ohne Commit
// ist er FAIL**: welcher Bau gemessen wurde, wüsste sonst niemand.
func Belegen(kopf Kopf, eingaben []Eingabe, auswertung Auswertung,
	artefakte []Eingabe,
) Beleg {
	beleg := Beleg{
		Format: Format, Kopf: kopf,
		Status: auswertung.Status, Grund: auswertung.Grund,
		Eingaben:      nichtNil(eingaben),
		Faelle:        nichtNil(auswertung.Faelle),
		Weitere:       auswertung.Weitere,
		Fehler:        nichtNil(auswertung.Fehler),
		Uebersprungen: []string{},
		Artefakte:     nichtNil(artefakte),
		Kennzahlen:    auswertung.Kennzahlen,
	}

	if beleg.Kennzahlen == nil {
		beleg.Kennzahlen = map[string]string{}
	}

	// **Eine leere Auswertung ist kein grüner Lauf** (AK-02) — wer keine
	// hat, hat nichts gemessen.
	if beleg.Status == "" {
		beleg.Status, beleg.Grund = NichtGelaufen, "keine Auswertung"
	}

	beleg.Werkzeuge = nichtNilKarte(kopf.Werkzeuge)
	beleg.Befehl = nichtNil(kopf.Befehl)

	for _, ergebnis := range beleg.Faelle {
		beleg.zaehle(ergebnis)
	}

	for _, eingabe := range beleg.Eingaben {
		if eingabe.Fehlt != "" {
			beleg.stufe(Blockiert, eingabe.Name+": "+eingabe.Fehlt)
		}
	}

	for _, artefakt := range beleg.Artefakte {
		if artefakt.Fehlt != "" {
			beleg.stufe(Gescheitert, "das Artefakt "+artefakt.Name+" fehlt: "+artefakt.Fehlt)
		}
	}

	if kopf.Commit == "" {
		beleg.stufe(Gescheitert, "ohne Commit ist der Stand nicht benannt")
	}

	if kopf.Inhalt == "" {
		beleg.stufe(Gescheitert, "ohne Inhaltshash ist nicht zu sagen, was gemessen wurde")
	}

	return beleg
}

// zaehle nimmt einen Fall in die Zahlen des Kopfs auf.
func (b *Beleg) zaehle(ergebnis Fallergebnis) {
	if ergebnis.Art == Weitere {
		return
	}

	b.Erwartet++

	switch ergebnis.Ausfuehrung {
	case Bestanden:
		b.Gemessen++
	case Uebersprungen:
		b.Uebersprungen = append(b.Uebersprungen,
			ergebnis.Paket+"."+ergebnis.Test+": "+ergebnis.Grund)
	}
}

// stufe hebt den Stand, wenn der neue schwerer ist.
func (b *Beleg) stufe(neu Status, grund string) {
	if neu.rang() > b.Status.rang() {
		b.Status, b.Grund = neu, grund
	}
}

// Schreiben schreibt den Beleg als eingerücktes JSON.
func (b Beleg) Schreiben(ziel io.Writer) error {
	text, fehler := json.MarshalIndent(b, "", "  ")
	if fehler != nil {
		return fmt.Errorf("den Beleg fassen: %w", fehler)
	}

	_, fehler = ziel.Write(append(text, '\n'))

	return fehler
}

// SchreibeNach schreibt den Beleg an seinen Ort — **ganz oder gar nicht**.
//
// In eine Nachbardatei und dann umbenannt: ein Beleg wird eingecheckt, und
// ein halber sähe im Diff aus wie ein ganzer. Scheitert etwas, bleibt der
// alte stehen und nichts Halbes daneben.
func (b Beleg) SchreibeNach(pfad string) error {
	verzeichnis := filepath.Dir(pfad)

	if fehler := os.MkdirAll(verzeichnis, 0o750); fehler != nil {
		return fmt.Errorf("den Beleg %s schreiben: %w", pfad, fehler)
	}

	neben, fehler := os.CreateTemp(verzeichnis, ".beleg-*")
	if fehler != nil {
		return fmt.Errorf("den Beleg %s schreiben: %w", pfad, fehler)
	}

	fehler = errors.Join(b.Schreiben(neben), neben.Close())
	if fehler == nil {
		fehler = os.Rename(neben.Name(), pfad)
	}

	if fehler != nil {
		_ = os.Remove(neben.Name()) //nolint:errcheck // aufgeräumt wird nach dem Fehler, und der ist schon unterwegs

		return fmt.Errorf("den Beleg %s schreiben: %w", pfad, fehler)
	}

	return nil
}

// Lesen liest einen Beleg — und weist ab, was keiner dieser Gestalt ist.
//
// **Ein unbekanntes Feld ist ein Fehler**: ein Tippfehler im Namen hieße
// sonst, dass eine Angabe still fehlt.
func Lesen(quelle io.Reader) (Beleg, error) {
	var beleg Beleg

	leser := json.NewDecoder(quelle)
	leser.DisallowUnknownFields()

	if fehler := leser.Decode(&beleg); fehler != nil {
		return Beleg{}, fmt.Errorf("kein Beleg: %w", fehler)
	}

	if beleg.Format != Format {
		return Beleg{}, fmt.Errorf("ein Beleg der Form %d — gelesen wird %d", beleg.Format, Format)
	}

	if !bekannt(beleg.Status) {
		return Beleg{}, fmt.Errorf("%q ist keiner der fünf Stände", beleg.Status)
	}

	return beleg, nil
}

// bekannt sagt, ob ein Stand einer der fünf ist.
func bekannt(status Status) bool {
	switch status {
	case Bestanden, Gescheitert, Blockiert, Uebersprungen, NichtGelaufen:
		return true
	default:
		return false
	}
}

// nichtNil gibt eine leere Liste statt nil — damit JSON `[]` schreibt und
// nicht `null`: »nichts gelaufen« und »nichts gesagt« sind zweierlei.
func nichtNil[T any](liste []T) []T {
	if liste == nil {
		return []T{}
	}

	return liste
}

// nichtNilKarte ist dasselbe für eine Karte.
func nichtNilKarte(karte map[string]string) map[string]string {
	if karte == nil {
		return map[string]string{}
	}

	return karte
}
