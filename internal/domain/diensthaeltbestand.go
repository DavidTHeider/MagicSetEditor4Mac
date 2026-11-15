package domain

// DienstHaeltBestand sagt, dass der eingerichtete Dienst den Bestand offen
// hält — der Grund, aus dem ein Zurückspielen nicht allein ist (M02,
// #2346,
// Befund A-11). Bis dahin hieß es „von einem anderen Programm geöffnet —
// schließe es", und den Dienst sieht niemand als Programm.
//
// **Die Klasse ist [ErrConflict]**, wie die Meldung, die er ersetzt; `Grund`
// trägt, was SQLite sagte.
type DienstHaeltBestand struct {
	Grund error
}

func (d *DienstHaeltBestand) Error() string {
	return "der Dienst für Skripte hält den Bestand offen — widerrufe ihn unter Einstellungen › Diagnose " +
		"und versuche es erneut"
}

// Unwrap gibt die Klasse und den Grund.
func (d *DienstHaeltBestand) Unwrap() []error { return []error{ErrConflict, d.Grund} }
