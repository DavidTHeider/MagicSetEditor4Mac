package domain

import (
	"strings"
	"time"
)

// Die Angaben zum Erwerb eines Exemplars — Preis, Währung und Tag.
//
// **Eine eigene Datei**, seit UX-06
// (#1678)
// die Gestalt der Währung und des Tages dazukam und `exemplar.go` über die
// Zeilengrenze aus §2 ging. Der Schnitt folgt der Frage: dort steht, was ein
// Exemplar ist, hier, was über seinen Erwerb gilt.

// Datumsbefunde prüft den Tag des Erwerbs — **ein Tag in kanonischer Form oder
// keiner**.
//
// Hier stand keine Prüfung, und `mse`, REST und MCP legten „gestern" ab; die
// Sammlung wird nach diesem Text geordnet (`storage/sammlungsfenster.go`).
//
// **Exportiert seit Befund G16 des Zweitblicks über Phase 6.3** (#1666), wie
// [Exemplar.Preisbefunde]: das Einlesen einer Sammlung leert eine Angabe zum
// Erwerb, die hier scheitert, statt die ganze Zeile zu verwerfen — und muss
// dafür wissen, **welche** Angabe scheiterte.
func (e Exemplar) Datumsbefunde() []*FieldError {
	datum := strings.TrimSpace(e.Erwerbsdatum)
	if datum == "" {
		return nil
	}

	if _, err := time.Parse(time.DateOnly, datum); err != nil {
		return []*FieldError{NewFieldError("erwerbsdatum", SchluesselValidierung,
			"»%s« ist kein Datum — erwartet wird Jahr-Monat-Tag wie 2026-09-24.", datum)}
	}

	return nil
}

// Preisbefunde prüft Erwerbspreis und Währung — sie gelten zusammen.
//
// **Zusammen und nicht einzeln.** Ein Preis ohne Währung ist eine Zahl ohne
// Bedeutung, eine Währung ohne Preis eine Angabe über nichts. Beide dürfen
// fehlen; **eines** von beiden darf es nicht. Exportiert aus demselben Grund
// wie [Exemplar.Datumsbefunde].
func (e Exemplar) Preisbefunde() []*FieldError {
	if e.Erwerbspreis < 0 {
		return []*FieldError{NewFieldError("erwerbspreis", SchluesselValidierung,
			"Ein Erwerbspreis ist nicht negativ.")}
	}

	waehrung := strings.TrimSpace(e.Erwerbswaehrung)

	if e.Erwerbspreis > 0 && waehrung == "" {
		return []*FieldError{NewFieldError("erwerbswaehrung", SchluesselPflichtfeld,
			"Zu einem Preis gehört seine Währung.")}
	}

	if e.Erwerbspreis == 0 && waehrung != "" {
		return []*FieldError{NewFieldError("erwerbspreis", SchluesselPflichtfeld,
			"Zu einer Währung gehört ihr Preis.")}
	}

	// **Die Gestalt eines Codes, wie an jeder anderen Stelle** — UX-06
	// (#1678): die Einstellung und das Kostenjournal weisen `EURO` und `€`
	// ebenso ab. Klein geschrieben bleibt erlaubt — `Normalisieren` hebt es an
	// (0057).
	if waehrung != "" && !istWaehrungscode(strings.ToUpper(waehrung)) {
		return []*FieldError{NewFieldError("erwerbswaehrung", SchluesselValidierung,
			"»%s« ist kein Währungscode — der hat drei Buchstaben, wie EUR.", waehrung)}
	}

	return nil
}
