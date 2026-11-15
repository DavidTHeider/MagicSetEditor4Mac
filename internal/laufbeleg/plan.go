package laufbeleg

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"maps"
	"slices"
	"strings"
	"time"
)

// Eingabewunsch ist eine Pflichteingabe, wie der Plan sie nennt: der Name
// der Umgebungsvariablen, die sie trägt, und was darin liegen muss.
type Eingabewunsch struct {
	Name   string `json:"name"`
	Muster string `json:"muster,omitempty"`
}

// Plan ist, was ein Lauf misst, womit, und wohin sein Beleg geht.
//
// **Er steht als Datei neben dem Beleg** — `docs/belege/<lauf>.plan.json`
// —, damit eine Ausnahme dort begründet wird, wo man den Beleg liest, und
// nicht in einem Makefile.
type Plan struct {
	Lauf      string
	Tag       string
	Beleg     string
	Roh       string
	Frist     time.Duration
	Zeitlimit time.Duration
	Misst     []string
	Eingaben  []Eingabewunsch
	// Schalter sind die Umgebungswerte, die ändern, **was** gemessen wird —
	// der Beleg hält ihren Wert fest, und die Prüfung vergleicht ihn.
	Schalter  []string
	Erwartung Erwartung
}

// plandatei ist die Gestalt der Datei.
type plandatei struct {
	Lauf      string            `json:"lauf"`
	Tag       string            `json:"tag"`
	Beleg     string            `json:"beleg"`
	Roh       string            `json:"roh"`
	Frist     string            `json:"frist"`
	Zeitlimit string            `json:"zeitlimit"`
	Misst     []string          `json:"misst"`
	Eingaben  []Eingabewunsch   `json:"eingaben"`
	Schalter  map[string]string `json:"schalter"`
	// OhneWirkung sind Umgebungswerte, die ein Lauf liest und die **nicht**
	// ändern, was er misst — Ablageorte, Vorrichtungen —, je mit Grund. Der
	// Plan nennt sie, damit ein Wächter jede gelesene Variable einordnen kann.
	OhneWirkung map[string]string `json:"ohneWirkung"`
	Trend       map[string]string `json:"trend"`
	Optional    map[string]string `json:"optional"`
}

// LiesPlan liest einen Plan gegen die markierten Fälle.
//
// **Streng, und jede Abweisung sagt warum.** Ein Trend oder eine Ausnahme,
// die es nicht gibt, hieße: ein umbenannter Lauf verlöre still seine
// Ausnahme und würde rot — oder, schlimmer, ein neuer mit demselben Namen
// bekäme sie ungesehen. **Eine Ausnahme ohne Grund ist keine.**
//
// Die Schlüssel stehen relativ zum Modul — `internal/app.TestKanon` — und
// `modul` macht daraus den Pfad, den `go test` meldet.
func LiesPlan(quelle io.Reader, modul string, markiert []Fall) (Plan, error) {
	var datei plandatei

	leser := json.NewDecoder(quelle)
	leser.DisallowUnknownFields()

	if fehler := leser.Decode(&datei); fehler != nil {
		return Plan{}, fmt.Errorf("kein Plan: %w", fehler)
	}

	if fehler := datei.pruefe(); fehler != nil {
		return Plan{}, fehler
	}

	plan := Plan{
		Lauf: datei.Lauf, Tag: datei.Tag, Beleg: datei.Beleg, Roh: datei.Roh,
		Misst: datei.Misst, Eingaben: datei.Eingaben,
		Schalter:  slices.Sorted(maps.Keys(datei.Schalter)),
		Erwartung: Erwartung{Faelle: markiert},
	}

	if datei.Frist != "" {
		frist, fehler := time.ParseDuration(datei.Frist)
		if fehler != nil {
			return Plan{}, fmt.Errorf("die Frist %q: %w", datei.Frist, fehler)
		}

		plan.Frist = frist
	}

	// **Ein Zeitlimit ist Pflicht, sobald es dasteht** — und null ist keines:
	// `go test` liest `-timeout 0` als »ohne Grenze«, und ein hängender Lauf
	// hinge dann für immer.
	if datei.Zeitlimit != "" {
		zeitlimit, fehler := time.ParseDuration(datei.Zeitlimit)
		if fehler != nil || zeitlimit <= 0 {
			return Plan{}, fmt.Errorf("das Zeitlimit %q ist keine Dauer über null", datei.Zeitlimit)
		}

		plan.Zeitlimit = zeitlimit
	}

	var fehler error

	if plan.Erwartung.Trend, fehler = ausnahmen(datei.Trend, modul, markiert); fehler != nil {
		return Plan{}, fehler
	}

	if plan.Erwartung.Optional, fehler = ausnahmen(datei.Optional, modul, markiert); fehler != nil {
		return Plan{}, fehler
	}

	for einer := range plan.Erwartung.Trend {
		if _, auch := plan.Erwartung.Optional[einer]; auch {
			return Plan{}, fmt.Errorf("%s.%s ist Trend und Ausnahme zugleich", einer.Paket, einer.Test)
		}
	}

	return plan, nil
}

// pruefe weist eine Datei ab, der eine Pflichtangabe fehlt.
func (d plandatei) pruefe() error {
	switch {
	case d.Lauf == "":
		return errors.New("der Plan nennt keinen Lauf")
	case d.Tag == "":
		return errors.New("der Plan nennt keinen Tag")
	case d.Beleg == "":
		return errors.New("der Plan nennt keinen Ort für den Beleg")
	case d.Roh == "":
		return errors.New("der Plan nennt keinen Ort für das Rohprotokoll")
	case len(d.Misst) == 0:
		return errors.New("der Plan nennt nicht, was er misst — dann wüsste niemand, " +
			"wann sein Beleg nicht mehr zum Stand passt")
	}

	genannt := map[string]bool{}

	for _, eingabe := range d.Eingaben {
		if eingabe.Name == "" {
			return errors.New("der Plan nennt eine Eingabe ohne Namen")
		}

		genannt[eingabe.Name] = true
	}

	// **Jede Umgebungsvariable an genau einer Stelle, und mit Grund** — eine,
	// die zugleich Eingabe und Schalter hieße, widerspräche sich.
	for _, gruppe := range []map[string]string{d.Schalter, d.OhneWirkung} {
		for _, name := range slices.Sorted(maps.Keys(gruppe)) {
			if strings.TrimSpace(gruppe[name]) == "" {
				return fmt.Errorf("der Plan nennt %s ohne Grund — eine Ausnahme ohne Grund ist keine", name)
			}

			if genannt[name] {
				return fmt.Errorf("%s steht zweimal im Plan — als Eingabe, Schalter oder ohne Wirkung", name)
			}

			genannt[name] = true
		}
	}

	return nil
}

// ausnahmen liest Trend oder Ausnahmen gegen die markierten Fälle.
func ausnahmen(eintraege map[string]string, modul string, markiert []Fall) (map[Fall]string, error) {
	gibt := map[Fall]bool{}
	for _, einer := range markiert {
		gibt[einer] = true
	}

	gelesen := map[Fall]string{}

	for schluessel, grund := range eintraege {
		paket, test, _ := strings.Cut(schluessel, ".")
		einer := Fall{Paket: modul + "/" + paket, Test: test}

		if !gibt[einer] {
			return nil, fmt.Errorf("der Plan nennt %s — %s gibt es nicht unter den markierten Fällen",
				schluessel, schluessel)
		}

		if strings.TrimSpace(grund) == "" {
			return nil, fmt.Errorf("der Plan nennt %s ohne Grund — eine Ausnahme ohne Grund ist keine",
				schluessel)
		}

		gelesen[einer] = grund
	}

	return gelesen, nil
}
