package app

// **Der Plan der Zufallsstichprobe, und wie er gezogen wird** — QA-02.AK-01
// (#1734).
//
// Ohne Build-Tag; benutzt von den Läufen hinter `fremdpakete`, geprüft in
// `stichprobenplan_test.go`.
//
// # Zufällig und trotzdem wiederholbar
//
// **Einmal gezogen und dann gespeichert**: der Plan steht in
// `docs/belege/kanonstichprobe.json`, und jeder Lauf misst dieselben Karten.
// Das hebt den Einwand auf, mit dem die feste Kohorte systematisch gezogen
// wurde — eine Zahl, die bei jedem Lauf eine andere Stichprobe misst, rauscht.
// **Und nachrechnen lässt er sich**: aus der Saat und den Kartenzahlen folgt
// jede Stelle; wer an ihm zweifelt, zieht ihn nach.

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/rand/v2"
	"slices"
	"strings"
)

// planfassung ist die Form des Plans. Wer ein Feld ändert, hebt sie.
const planfassung = 1

// stichprobenplan ist die gezogene Zufallsstichprobe über die Sätze.
type stichprobenplan struct {
	Fassung   int           `json:"fassung"`
	Saat      [2]uint64     `json:"saat"`
	Vertrauen float64       `json:"vertrauen"`
	Quote     float64       `json:"quote"`
	Gemeinsam float64       `json:"gemeinsam"`
	Schichten []planschicht `json:"schichten"`
}

// planschicht ist eine Schicht — ein Satz — mit ihren gezogenen Stellen.
type planschicht struct {
	Satz                    string  `json:"satz"`
	Karten                  int     `json:"karten"`
	Kartenhash              string  `json:"kartenhash"`
	Betroffen               int     `json:"betroffen"`
	Findewahrscheinlichkeit float64 `json:"findewahrscheinlichkeit"`
	Stichprobe              []int   `json:"stichprobe"`
}

// ziehePlan zieht je Schicht so viele Karten, wie `noetigeStichprobe`
// verlangt — **ohne Zurücklegen und der Reihe nach aus einem Generator**, in
// der Ordnung der Schichten. Dieselbe Saat gibt denselben Plan.
func ziehePlan(saat [2]uint64, schichten []planschicht) stichprobenplan {
	zufall := rand.New(rand.NewPCG(saat[0], saat[1])) //nolint:gosec // eine Stichprobe und kein Geheimnis: die Saat steht offen im Plan

	plan := stichprobenplan{
		Fassung: planfassung, Saat: saat, Vertrauen: 0.999, Quote: 0.01,
		Schichten: make([]planschicht, 0, len(schichten)),
	}

	sicherheiten := make([]float64, 0, len(schichten))

	for _, schicht := range schichten {
		gezogen := noetigeStichprobe(plan.Vertrauen, plan.Quote, schicht.Karten)
		schicht.Stichprobe = zufall.Perm(schicht.Karten)[:gezogen]
		slices.Sort(schicht.Stichprobe)

		schicht.Betroffen = betroffeneBei(plan.Quote, schicht.Karten)
		schicht.Findewahrscheinlichkeit = findewahrscheinlichkeit(schicht.Karten,
			schicht.Betroffen, gezogen)

		sicherheiten = append(sicherheiten, schicht.Findewahrscheinlichkeit)
		plan.Schichten = append(plan.Schichten, schicht)
	}

	plan.Gemeinsam = gemeinsameSicherheit(sicherheiten)

	return plan
}

// schreiben schreibt den Plan als eingerücktes JSON.
func (p stichprobenplan) schreiben(ziel io.Writer) error {
	text, fehler := json.MarshalIndent(p, "", "  ")
	if fehler != nil {
		return fmt.Errorf("den Plan fassen: %w", fehler)
	}

	_, fehler = ziel.Write(append(text, '\n'))

	return fehler
}

// liesPlan liest einen Plan — und weist ab, was keiner dieser Form ist.
func liesPlan(quelle io.Reader) (stichprobenplan, error) {
	var plan stichprobenplan

	leser := json.NewDecoder(quelle)
	leser.DisallowUnknownFields()

	if fehler := leser.Decode(&plan); fehler != nil {
		return stichprobenplan{}, fmt.Errorf("kein Stichprobenplan: %w", fehler)
	}

	if plan.Fassung != planfassung {
		return stichprobenplan{}, fmt.Errorf("ein Plan der Form %d — gelesen wird %d",
			plan.Fassung, planfassung)
	}

	return plan, nil
}

// kartenhash hält die Kartenliste eines Satzes fest — **in ihrer Ordnung**,
// denn der Plan nennt Stellen und keine Namen.
//
// **Getrennt durch NUL**, das in keinem Namen steht: mit einem Zeilenende
// gäben „A\nB, C" und „A, B, C" denselben Hash.
func kartenhash(namen []string) string {
	summe := sha256.Sum256([]byte(strings.Join(namen, "\x00")))

	return "sha256:" + hex.EncodeToString(summe[:])
}
