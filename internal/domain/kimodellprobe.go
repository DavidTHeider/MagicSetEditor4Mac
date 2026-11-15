package domain

// **Eine gemessene Fähigkeit** — Werkstück A9 von Phase 5.3
// (#1448).
//
// # Warum es sie gibt
//
// **Weil nicht jeder Anbieter seine Modelle beschreibt.** Anthropic nennt
// `capabilities` (A8), Ollama nennt sie in `/api/show` (A7) — OpenAI nennt
// **nichts**
// ([0100](../../docs/entscheidungen/0100-was-openai-nicht-sagt.md)).
//
// [0097](../../docs/entscheidungen/0097-was-ueber-ein-ki-modell-feststeht.md)
// lässt dafür genau einen Weg offen, und es ist derselbe, den A7 für den
// Einbettungsvertrag gegangen ist: **messen statt raten**. Eine Probe
// schickt einen kleinen Auftrag mit festem Schema; kommt etwas
// Schemakonformes zurück, ist die Fähigkeit belegt.
//
// # Was eine Probe **nicht** ist
//
// **Keine Zusage für morgen.** Sie sagt: an **diesem** Tag konnte **dieses**
// Modell unter **dieser** Laufzeit **dieses eine**. Ein Anbieter darf
// tauschen, was hinter einem Alias liegt — deshalb steht neben dem Alias,
// was wirklich geantwortet hat.

import (
	"slices"
	"strings"
)

// KIModellprobe ist eine Messung an einem Modell.
type KIModellprobe struct {
	// Anbieter ist, wen man gefragt hat.
	Anbieter string

	// Alias ist der Name, unter dem gefragt wurde.
	Alias string

	// Faehigkeit ist, was dabei herauskam.
	Faehigkeit KIFaehigkeit

	// Datum ist der Tag der Messung — `JJJJ-MM-TT`.
	Datum string

	// Beleg sagt, **wie** gemessen wurde.
	Beleg string

	// Laufzeit ist, womit geantwortet wurde.
	Laufzeit string

	// Modell ist, was der Anbieter als antwortendes Modell genannt hat.
	//
	// **Nicht dasselbe wie der Alias**: er zeigt auf eine Fassung, und wird
	// sie getauscht, ist die Messung eine über etwas anderes.
	Modell string
}

// AlsGeprueft macht aus der Messung einen Beleg für den Katalog.
func (p KIModellprobe) AlsGeprueft() KIGeprueft {
	return KIGeprueft{
		Faehigkeit: p.Faehigkeit, Datum: p.Datum, Beleg: p.Beleg,
	}
}

// GiltFuer sagt, ob diese Messung zu einem Modell gehört.
//
// **Anbieter **und** Alias**, denn zwei Anbieter dürfen denselben Namen
// führen — und genau deshalb steht der Anbieter seit A8 überall daneben.
func (p KIModellprobe) GiltFuer(modell KIModell) bool {
	return p.Anbieter == modell.Anbieter && p.Alias == modell.Alias
}

// Validate prüft Anbieter, Alias, Laufzeit und den Beleg.
func (p KIModellprobe) Validate() []*FieldError {
	var befunde []*FieldError

	for feld, wert := range map[string]string{
		"anbieter": p.Anbieter, "alias": p.Alias, "laufzeit": p.Laufzeit,
		"modell": p.Modell,
	} {
		if strings.TrimSpace(wert) == "" {
			befunde = append(befunde, NewFieldError(feld, SchluesselPflichtfeld,
				"Eine Messung nennt ihren %s.", feld))
		}
	}

	// **Der Beleg wird von [KIGeprueft] geprüft und nicht hier noch einmal.**
	// Zwei Prüfungen derselben Sache laufen auseinander, und die dort ist
	// die, an der der Katalog hängt.
	return append(befunde, p.AlsGeprueft().Validate()...)
}

// MitProben legt die gemessenen Fähigkeiten zu den gemeldeten.
//
// # Warum die gemeldeten zuerst stehen
//
// **Weil eine Messung die Auskunft des Anbieters nicht ersetzt, sondern
// ergänzt.** Sagt Anthropic, ein Modell könne feste Form, und eine Messung
// sagt dasselbe, ist das dieselbe Fähigkeit zweimal belegt — und
// [KIModell.Kann] fragt ohnehin nur, ob **ein** gültiger Beleg dasteht.
//
// # Und eine Messung ohne Modell dazu fällt hier weg
//
// **Sie gehört zu einem Alias, den es vielleicht nicht mehr gibt.** Ein
// Eintrag ohne Modell wäre ein Katalogeintrag ohne Anbieterantwort — genau
// das, was 0097 ausschließt.
//
// **Wo der Anbieter den Alias noch nennt, ist das anders** — dann gibt es
// eine Anbieterantwort, sie sagt nur nichts über Fähigkeiten. Dafür steht
// [MitGemessenen] daneben.
func MitProben(modelle []KIModell, proben []KIModellprobe) []KIModell {
	ergaenzt := make([]KIModell, 0, len(modelle))

	for _, modell := range modelle {
		for _, probe := range proben {
			if probe.GiltFuer(modell) && len(probe.Validate()) == 0 {
				modell.Geprueft = append(modell.Geprueft, probe.AlsGeprueft())
			}
		}

		ergaenzt = append(ergaenzt, modell)
	}

	return ergaenzt
}

// MitGemessenen nimmt Modelle auf, die der Anbieter **nennt**, über die er
// aber nichts sagt — und für die eine Messung vorliegt.
//
// # Warum es das neben [MitProben] gibt
//
// **Weil eine bezahlte Messung sonst folgenlos bleibt** —
// #1541.
// `MitProben` ergänzt nur, was schon in der Liste steht, und OpenAI gibt
// **immer eine leere Liste** zurück: jedes Modell wird zu einem Befund, denn
// „eine Modellliste ist keine belegte Fähigkeit"
// ([0100](../../docs/entscheidungen/0100-was-openai-nicht-sagt.md)). Damit kam
// dort **nie** ein gemessenes Modell in den Katalog; gemessen, gebucht,
// gemeldet — und zu nichts zu gebrauchen. Gefunden in der Sichtprüfung am
// laufenden Fenster (2026-09-13).
//
// # Und warum 0097 dabei bestehen bleibt
//
// **„Kein Katalogeintrag ohne Anbieterantwort" gilt weiter.** Aufgenommen
// wird nur, was in `genannt` steht — den Aliassen aus **dieser** Antwort des
// Anbieters. Wer ein Modell abschaltet, verschwindet damit aus dem Katalog,
// obwohl die Messung im Bestand liegt; das ist der Punkt jenes Vermerks.
//
// **Der Ort kommt vom Anbieter und nicht aus der Messung**: wo gerechnet
// wird, entscheidet nicht, wer geantwortet hat — und unter „nur lokal" darf
// ein Cloudmodell nicht plötzlich als lokal gelten.
func MitGemessenen(
	modelle []KIModell, proben []KIModellprobe, genannt []string,
	anbieter string, ort KIOrt,
) []KIModell {
	vorhanden := make(map[string]bool, len(modelle))
	for _, modell := range modelle {
		vorhanden[modell.Alias] = true
	}

	ergaenzt := slices.Clone(modelle)

	for _, alias := range genannt {
		if vorhanden[alias] {
			continue
		}

		geprueft, laufzeit := gemessenesFuer(proben, anbieter, alias)
		if len(geprueft) == 0 {
			continue
		}

		vorhanden[alias] = true

		ergaenzt = append(ergaenzt, KIModell{
			Anbieter: anbieter, Alias: alias, Ort: ort,
			Laufzeit: laufzeit, Geprueft: geprueft,
		})
	}

	return ergaenzt
}

// gemessenesFuer sammelt die gültigen Messungen eines Alias samt Laufzeit.
//
// **Beides in einem Durchgang, und das ist kein Sparen.** Eine gültige
// Messung **nennt** ihre Laufzeit — [KIModellprobe.Validate] verlangt sie —,
// und eine zweite Schleife dafür hätte einen Zweig „keine gefunden", den
// niemand betreten kann. Ein unbetretbarer Zweig ist eine Zusage, die keiner
// prüft (§10).
func gemessenesFuer(
	proben []KIModellprobe, anbieter, alias string,
) (geprueft []KIGeprueft, laufzeit string) {
	geprueft = make([]KIGeprueft, 0, len(proben))

	for _, probe := range proben {
		if probe.Anbieter != anbieter || probe.Alias != alias ||
			len(probe.Validate()) != 0 {
			continue
		}

		geprueft = append(geprueft, probe.AlsGeprueft())
		laufzeit = probe.Laufzeit
	}

	return geprueft, laufzeit
}
