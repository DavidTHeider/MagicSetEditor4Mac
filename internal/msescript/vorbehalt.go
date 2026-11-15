package msescript

// Vorbehalt ist ein Befund, der die Auswertung **nicht** abbricht.
//
// # Wozu
//
// Ein Initskript ist eine **Folge von Vereinbarungen**, und jede steht für
// sich. `magic.mse-game` hat 14.741 Zeilen; scheitert eine davon, ist das kein
// Grund, die 4.000 danach zu verlieren — und in denen steht, was den Rahmen
// einer Karte zeichnet
// (#240).
//
// [VorbelegenMitVorbehalten] rechnet deshalb Satz für Satz und sammelt, was
// nicht ging.
//
// # Warum das kein stiller Weg ist
//
// Der stille Weg wäre, den gescheiterten Satz zu überspringen **ohne Meldung**.
// Genau der erzeugt eine Karte, die aussieht wie eine richtige. Ein Vorbehalt
// sagt beides: die Auswertung läuft weiter, **und** sie gibt hinterher an, wo
// und warum.
//
// Das ist §8 für die Sprache: „Ein Set, von dem 300 Karten lesbar sind und zwei
// nicht, wird mit 300 Karten und zwei benannten Befunden geöffnet."
//
// # Was ein Vorbehalt nicht ist
//
// **Kein Ersatz für einen Fehler.** Ein einzelner Ausdruck — ein Feldskript —
// bricht weiterhin ab: dort *ist* das Skript sein Ergebnis, und die Hälfte
// davon bedeutet nichts.
//
// **Und kein Notausgang für ein Muster.** Bis #240 gab es hier einen zweiten
// Weg: eine eingebaute Funktion durfte einen Ersatzwert zurückgeben, wenn Gos
// `regexp` ein Muster nicht übersetzen konnte. Diese Notlösung ist weg, weil
// die Ursache weg ist — [Muster] übersetzt die fünfzehn Muster, um die es ging.
// Eine Umgehung, deren Anlass behoben ist, gehört gelöscht und nicht behalten.
type Vorbehalt struct {
	Zeile   int
	Spalte  int
	Meldung string
}
