// **Was ein Archiv über sich selbst sagen darf** — die Grenzen des Auspackens.
//
// Aus `archiv.go` herausgelöst, als es über 400 Zeilen ging (§2). Der Schnitt
// folgt der Frage: dort steht, **was** eine Sicherung ist, hier, **wie viel**
// sie höchstens sein darf.
//
// # Warum es sie gibt
//
// Eine Sicherung sagt über sich selbst, wie groß ihre Teile sind — und das ist
// eine Angabe **aus** der Datei, nicht **über** sie. Wer ihr ohne Grenze
// glaubt, lässt sich von einer präparierten Datei sagen, wie viel er
// bereitstellen soll (Befunde 3 und 5 des Zweitblicks über
// #1162).
//
// **Die Zahlen sind großzügig gewählt und nicht knapp.** Sie sollen einen
// Angriff aufhalten, nicht eine große Sammlung.

package sicherung

// HoechstensManifest ist die Grenze für `sicherung.json`.
//
// **Es ist der eine Eintrag, dessen Größe niemand vorher kennt** — er *ist*
// die Auskunft darüber, wie groß etwas sein darf. Ohne Grenze wäre er der Weg,
// auf dem eine präparierte Datei den Arbeitsspeicher füllt, noch bevor
// irgendetwas geprüft ist (Befund 5 des Zweitblicks über
// #1162).
//
// **64 MiB ist reichlich**: ein Manifest mit einer Million Einträgen wiegt
// rund 130 MB — und eine Million Bilder hat niemand.
const HoechstensManifest int64 = 64 << 20

// HoechstensTeil ist die Grenze für einen einzelnen Eintrag.
//
// **Sie steht neben der Angabe aus dem Manifest, nicht statt ihr.** Gelesen
// wird höchstens die angekündigte Größe; diese Zahl begrenzt die Ankündigung
// selbst, denn sie stammt aus dem Archiv und damit von dem, der es geschrieben
// hat.
//
// **16 GiB**: der Bestand ist die größte Datei einer Sicherung, und eine
// SQLite-Datei dieser Größe hat niemand, der Karten entwirft.
const HoechstensTeil int64 = 16 << 30

// HoechstensEintraege ist die Grenze für die Zahl der Einträge.
//
// Sie fängt den Fall, den die beiden Zahlen darüber durchlassen: sehr viele
// sehr kleine Einträge.
//
// **Gezählt wird am ZIP-Verzeichnis, nicht am Manifest** — der Nachlauf über
// #1162 hat
// den Unterschied angestrichen: ein Archiv mit kurzem Manifest und Millionen
// **nicht aufgeführter** Einträge ging an der Manifestzählung vorbei. Das
// Manifest wird zusätzlich gezählt, denn seine Liste treibt die Schleife.
const HoechstensEintraege = 1 << 20

// HoechstensGesamt ist die Grenze für die Summe aller Teile.
//
// **Sie steht neben [HoechstensTeil] und nicht statt ihr**: hundert Teile zu
// je 15 GiB sind einzeln zulässig und zusammen anderthalb Terabyte. Was ohne
// sie fehlte, ist nicht der Arbeitsspeicher — geströmt wird — sondern der
// Platz auf dem Laufwerk, auf dem der laufende Bestand liegt.
//
// **64 GiB**: mehr als jede Sammlung wiegt, und weniger als jede Platte hat.
const HoechstensGesamt int64 = 64 << 30
