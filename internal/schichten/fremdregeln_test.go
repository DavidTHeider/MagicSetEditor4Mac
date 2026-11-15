// **Die Importe außerhalb des Projekts, die eine Schicht nicht führen darf** —
// abgetrennt von `regeln_test.go`, als die Datei mit den beiden Blättern des
// zweiten Faktors über die Zeilengrenze aus §2 kam
// (#1351).
//
// **Der Schnitt folgt der Frage**: nebenan steht, welche **eigenen** Pakete
// eine Schicht kennen darf, hier, welche **fremden** sie nicht kennen darf.
package schichten

// verboteneFremdpakete nennt Importe außerhalb des Projekts, die eine Schicht
// nicht führen darf.
//
// `internal/domain` und `database/sql` ist der benannte Fall aus §2: wer dort
// einen SQL-Import sieht, hat einen Fehler gefunden. Die Domäne beschreibt, was
// gilt — nicht, wie es gespeichert wird.
var verboteneFremdpakete = map[string][]string{
	"domain": {"database/sql"},
}
