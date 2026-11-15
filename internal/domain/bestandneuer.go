package domain

// BestandNeuer ist die Abweisung des Bestands, den dieser Prozess öffnet —
// er stammt aus einer neueren Fassung. Nachlauf über Phase 2 von 0.9.0,
// Befund 2 (#2323).
//
// **Dieselbe [NeuereFassung] wie bei einer Sicherung**, und doch nicht
// dasselbe: eine Vorsicherung aus einer neueren Fassung ist ein Fehler der
// Vorbereitung, nach dem ein Dienst unter launchd neu anlaufen soll; ein
// laufender Bestand aus einer neueren Fassung heißt, dass dieses Programm
// nichts zu tun hat. Nur der Migrator setzt diese Hülle.
//
// **Sie hüllt die ganze Abweisung**: die Klasse der Migrationen samt der
// [NeuereFassung] mit ihren Namen — wer nach einer von beiden fragt, findet
// sie weiter.
type BestandNeuer struct {
	Grund error
}

func (b *BestandNeuer) Error() string { return b.Grund.Error() }

// Unwrap gibt die Abweisung — Satz und Werte bleiben, wie sie sind.
func (b *BestandNeuer) Unwrap() error { return b.Grund }
