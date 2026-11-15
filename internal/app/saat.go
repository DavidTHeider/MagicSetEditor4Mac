package app

import "os"

// Saatverzeichnis gibt den Ort der mitgelieferten Saat — oder nichts.
//
// # Warum neben dem Programm und nicht neben der Datenbank
//
// Die Saat ist **Teil der Anwendung**, kein Bestand der Nutzenden: sie kommt
// mit jeder Fassung neu, und wer sie ändert, ändert die Anwendung. Bezogene
// Pakete liegen unter `~/Library/Application Support/` und überleben eine
// Neuinstallation; die Saat soll das gerade nicht
// (#58).
//
// # Wer sie dorthin legt
//
// `scripts/saat.sh`, gerufen aus `scripts/bauen.sh`. Damit gilt derselbe Weg
// für die Freigabe und für `make dev` — und ausgerechnet der zweite ist der,
// auf dem die Sichtprüfung geht.
//
// # Was das nicht leistet
//
// **Ein `go run` findet nichts.** Wer die Anwendung ohne Bündel startet,
// bekommt einen leeren Pfad und damit keine Saat — dieselbe Lage wie vor #58,
// nur benannt. Ein Rückfall auf ein `saat/` im Arbeitsverzeichnis wäre ein
// relativer Pfad im Betrieb, und genau davor warnt der Paketbezieher.
func Saatverzeichnis() string { return saatNeben(os.Executable) }

// saatNeben rechnet vom Programmpfad auf das Saatverzeichnis im selben Bündel.
//
// **Dieselbe Rechnung wie für die mitgelieferten Vorlagen** ([nebenDemProgramm]) —
// eine zweite Fassung derselben Regel läuft irgendwann auseinander.
func saatNeben(programm func() (string, error)) string {
	return nebenDemProgramm(programm, "saat")
}
