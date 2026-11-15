package domain

// **Zurück ist nicht eine Handlung, sondern vier** — Werkstück O7 von Phase
// 5.2 (#1431).
//
// # Warum die Trennung der Ertrag ist
//
// Bis hierher gab es **einen** Knopf: *Sicherung zurückspielen*. Er tauscht
// die Daten — und wer ihn drückt, weil eine neue Fassung nicht tut, was sie
// soll, hat die Fassung damit nicht gewechselt. Vier Dinge können zurück, sie
// kosten Verschiedenes, und **zwei davon kann dieses Programm nicht selbst**.
//
// Das auszusprechen ist die Arbeit dieses Stücks: eine Fläche, die alle vier
// gleich aussehen lässt, verspricht dreimal etwas, das nur einmal stimmt (§10).

// Ruecksprungart ist eine der vier Arten.
type Ruecksprungart string

const (
	// RuecksprungAppfassung ist das Programm selbst.
	RuecksprungAppfassung Ruecksprungart = "appfassung"

	// RuecksprungSchema wäre die Rückwärtswanderung des Bestands.
	RuecksprungSchema Ruecksprungart = "schema"

	// RuecksprungDaten ist der Bestand samt Bildern — das Zurückspielen.
	RuecksprungDaten Ruecksprungart = "daten"

	// RuecksprungPakete ist der Stand der Vorlagenpakete.
	RuecksprungPakete Ruecksprungart = "pakete"
)

// Ruecksprungarten gibt die Reihe — **in der Ordnung, in der sie zusammenhängen**:
// erst das Programm, dann sein Schema, dann die Daten, dann das Beiwerk.
func Ruecksprungarten() []Ruecksprungart {
	return []Ruecksprungart{
		RuecksprungAppfassung, RuecksprungSchema,
		RuecksprungDaten, RuecksprungPakete,
	}
}

// Die Hindernisse — **Marken und keine Sätze**, wie bei jedem Ausgang, den
// vier Flächen verschieden formulieren (O6, O9).
const (
	// RuecksprungNachProzessende: das Bündel tauscht ein Mensch, wenn dieses
	// Programm nicht mehr läuft.
	//
	// **Ein Programm, das sich selbst ersetzt, während es läuft**, braucht auf
	// macOS einen Helferprozess, der das Bündel nach dem Ende überschreibt —
	// und der wäre ein zweites, unsigniertes Stück Software, das
	// Schreibrechte auf `/Applications` hat. Was stattdessen dasteht, ist der
	// Ablauf: `docs/ADMINISTRATION.md` nennt ihn Schritt für Schritt, und
	// dieses Programm prüft nach dem Start, ob er geglückt ist.
	RuecksprungNachProzessende = "nachProzessende"

	// RuecksprungNurVorwaerts: es gibt keine Rückwärtswanderung, und das ist
	// eine Entscheidung.
	//
	// **Der Weg zurück ist eine Sicherung**, nicht eine zweite Wanderung.
	// Rückwärts zu wandern hieße, für jede Migration einen zweiten Satz
	// Wahrheit über dasselbe Schema zu führen — und der zweite ist der, den
	// niemand je ausführt, bis es darauf ankommt. Das Issue sagt es
	// ausdrücklich: Vorwärtsmigration plus Sicherungswiederherstellung, keine
	// Pflicht zu Down-Migrationen.
	RuecksprungNurVorwaerts = "nurVorwaerts"

	// RuecksprungUnbekannteArt steht für eine Art, die es nicht gibt.
	//
	// **Sie ist ein Hindernis und kein leerer Hinweis.** Ein leerer bedeutet
	// „geht hier"; eine erfundene Art als machbar durchzulassen wäre die
	// stille Zusage, dass etwas geschieht, wofür es keinen Weg gibt.
	RuecksprungUnbekannteArt = "unbekannteArt"
)

// RuecksprungHindernis sagt, was einer Art im Weg steht — leer heißt: **diese
// Anwendung kann sie selbst**.
func RuecksprungHindernis(art Ruecksprungart) string {
	switch art {
	case RuecksprungAppfassung:
		return RuecksprungNachProzessende
	case RuecksprungSchema:
		return RuecksprungNurVorwaerts
	case RuecksprungDaten, RuecksprungPakete:
		return ""
	default:
		return RuecksprungUnbekannteArt
	}
}
