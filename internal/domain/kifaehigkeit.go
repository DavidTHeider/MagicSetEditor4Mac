package domain

// **Was ein KI-Modell kann** — Werkstück A6 von Phase 5.3
// (#1445).
//
// # Warum sieben und nicht eine
//
// **Weil sie verschiedene Dinge liefern.** Das Issue sagt es wörtlich:
// „Embeddings liefern Vektoren, Reranker Kandidatenbewertungen, Generierung
// Vorschläge — kein generischer »KI-Antwort«-Blob." Ein gemeinsamer Typ
// darüber hieße, dass eine Stelle im Code entscheidet, was gerade
// zurückkommt — und das ist die Stelle, an der ein Vektor als Vorschlag in
// ein Kartenfeld gerät.
//
// # Und warum ein Modell sie **belegen** muss
//
// **Weil eine Liste keine Fähigkeit ist.** „Unbekannt bleibt unbekannt;
// Modelllisten garantieren keine Capability" — ein Name in einem Katalog sagt,
// dass es das Modell gibt, nicht, dass es dieses kann. Deshalb trägt jede
// Fähigkeit ein **Datum** und einen **Beleg**; ohne beides gilt sie als
// ungeprüft und nicht als vorhanden.

import "time"

// KIFaehigkeit ist etwas, das ein Modell kann.
type KIFaehigkeit string

// Die sieben.
const (
	// KITextgenerierung ist freier Text — ein Vorschlag für ein Feld.
	KITextgenerierung KIFaehigkeit = "textgenerierung"

	// KIStrukturierteGenerierung ist Text in einer zugesagten Form.
	//
	// **Nicht dasselbe wie Textgenerierung mit einer Bitte im Prompt**: hier
	// hält die Laufzeit die Form, dort hofft man darauf.
	KIStrukturierteGenerierung KIFaehigkeit = "strukturiert"

	// KIVision ist „sieh dir dieses Bild an und sag, was darauf steht".
	KIVision KIFaehigkeit = "vision"

	// KIBildgenerierung macht aus Text ein Bild.
	KIBildgenerierung KIFaehigkeit = "bildgenerierung"

	// KIBildbearbeitung ändert ein vorhandenes Bild.
	KIBildbearbeitung KIFaehigkeit = "bildbearbeitung"

	// KIEinbettung gibt einen **Vektor** — die Grundlage des Index aus A15.
	KIEinbettung KIFaehigkeit = "einbettung"

	// KINeuordnung bewertet Kandidaten gegen eine Frage.
	//
	// **Sie gibt Zahlen und keine Reihenfolge**: wer aus ihnen eine Liste
	// macht, entscheidet über Schwellen — und das ist eine Frage des
	// Aufrufers.
	KINeuordnung KIFaehigkeit = "neuordnung"
)

var kiFaehigkeiten = []KIFaehigkeit{
	KITextgenerierung, KIStrukturierteGenerierung, KIVision,
	KIBildgenerierung, KIBildbearbeitung, KIEinbettung, KINeuordnung,
}

// KIFaehigkeiten gibt die sieben in ihrer Reihenfolge.
func KIFaehigkeiten() []KIFaehigkeit {
	kopie := make([]KIFaehigkeit, len(kiFaehigkeiten))
	copy(kopie, kiFaehigkeiten)

	return kopie
}

// IstKIFaehigkeit sagt, ob es die Fähigkeit gibt.
func IstKIFaehigkeit(faehigkeit KIFaehigkeit) bool {
	return kennt(kiFaehigkeiten, faehigkeit)
}

// KIGeprueft ist eine Fähigkeit, die jemand nachgesehen hat.
//
// **Datum und Beleg sind Pflicht**, und das ist der ganze Zweck des Typs: ohne
// sie wäre er ein zweiter Name für [KIFaehigkeit].
type KIGeprueft struct {
	Faehigkeit KIFaehigkeit

	// Datum ist der Tag der Prüfung — `YYYY-MM-DD`.
	//
	// **Eine Prüfung altert.** Ein Anbieter tauscht ein Modell hinter seinem
	// Namen aus, eine Laufzeit ändert ihre Vorgaben; was am 2026-09-11 ging,
	// geht deshalb nicht auf ewig. Das Datum sagt, wie alt die Auskunft ist —
	// es sagt **nicht**, dass sie noch stimmt.
	Datum string

	// Beleg sagt, **wie** geprüft wurde — ein eigener Lauf, eine Seite des
	// Anbieters mit Abrufdatum, eine Messung.
	//
	// **Als Satz und nicht als Häkchen**: „geprüft" ohne das Wie ist eine
	// Behauptung, die niemand nachvollziehen kann (§10).
	Beleg string
}

// Validate prüft Fähigkeit, Datum und Beleg.
func (g KIGeprueft) Validate() []*FieldError {
	var befunde []*FieldError

	if !IstKIFaehigkeit(g.Faehigkeit) {
		befunde = append(befunde, NewFieldError("faehigkeit", SchluesselValidierung,
			"%q ist keine bekannte KI-Fähigkeit.", string(g.Faehigkeit)))
	}

	if g.Datum == "" {
		befunde = append(befunde, NewFieldError("datum", SchluesselPflichtfeld,
			"Eine geprüfte Fähigkeit nennt den Tag der Prüfung."))
	} else if _, fehler := time.Parse(time.DateOnly, g.Datum); fehler != nil {
		befunde = append(befunde, NewFieldError("datum", SchluesselValidierung,
			"%q ist kein Datum in der Form JJJJ-MM-TT.", g.Datum))
	}

	if g.Beleg == "" {
		befunde = append(befunde, NewFieldError("beleg", SchluesselPflichtfeld,
			"Eine geprüfte Fähigkeit sagt, wie geprüft wurde — „geprüft“ "+
				"allein ist eine Behauptung."))
	}

	return befunde
}
