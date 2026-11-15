package domain

// **Was von einem Abschluss übrig bleibt** — Werkstück `T6` von Phase 5.6
// (#1471).
//
// # Warum die Form mitgespeichert wird und nicht nur die Unterschrift
//
// **Weil eine nachgerechnete Form nicht die unterschriebene ist.** Wer sie
// beim Prüfen neu aus den Zeilen baute, prüfte die Unterschrift gegen das,
// was heute daraus folgt — und nicht gegen das, was damals unterschrieben
// wurde. Genau dieser Unterschied ist der Zweck einer Unterschrift.

import "time"

// Belegstand ist ein abgelegter Handelsbeleg.
type Belegstand struct {
	Vorgang string

	// Abdruck ist der Abdruck der Revision, die abgeschlossen wurde.
	Abdruck string

	// Form ist, **was** unterschrieben wurde — Zeichen für Zeichen.
	Form string

	Unterschrift string

	// Beidseitig sagt, ob auch die Gegenseite unterschrieben hat.
	//
	// **Eine eigene Spalte und keine Folgerung** (§10): zwei Unterschriften
	// vorzutäuschen, weil zwei Parteien dastehen, wäre die Zusage, die
	// dieses Werkstück ausdrücklich nicht gibt.
	//
	// **Und bis zum 2026-09-19 war sie trotzdem eine Folgerung**: gesetzt
	// wurde sie davon, dass **hier** jemand für beide Parteien bestätigt
	// hat (#1650).
	// Jetzt setzt sie nur, wer einen **geprüften** Fremdbeleg ablegt — und
	// die zwei Felder darunter sagen, wer und wann.
	Beidseitig bool

	// Gegenzeichner ist die Peerkennung dessen, dessen geprüfter Beleg hier
	// liegt — leer, solange keiner da ist.
	Gegenzeichner string

	// Gegengezeichnet ist der Augenblick, in dem sein Beleg ankam — die
	// Nullzeit, solange keiner da ist.
	Gegengezeichnet time.Time

	// Befund ist der jüngste abgewiesene Gegenbeleg — `nil`, solange keiner
	// abgewiesen wurde (#2500). **Er steht auch neben einem geprüften**: ein
	// späterer, passender Beleg löscht ihn nicht, er macht ihn alt.
	Befund *Gegenbelegbefund

	Gebucht time.Time
}
