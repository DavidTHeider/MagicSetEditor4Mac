package storage

// SicherungsschrittSetzen setzt die Naht `sicherungsschritt` — für die
// Absturzprobe in `sicherungsabbruch_test.go`, die als zweiter Prozess läuft
// und deshalb von außen an sie muss (BK-06a, #2130).
func SicherungsschrittSetzen(melden func(schritt, pfad string)) {
	sicherungsschritt = melden
}
