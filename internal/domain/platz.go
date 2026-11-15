package domain

// **Was Platz belegt** — Werkstück 4 von Phase 4.13
// (#1352).
//
// # Warum die Posten hier stehen
//
// **Weil „was gehört zusammen und was darf weg" eine fachliche Frage ist**
// (§2). Wo die Dateien liegen, weiß `internal/app`; **ob** ein Posten
// wegdarf, ist eine Entscheidung über die Daten — und die steht hier.
//
// # Der Anlass, gemessen
//
// **Am 2026-09-07 auf dem Rechner des Auftraggebers gezählt**: neben einem
// Bestand von 127 MB lagen **dreißig** Sicherungen vor Migrationen, zusammen
// über zwei Gigabyte. Sie entstehen bei jedem Start, der migriert, und niemand
// räumte sie weg.

// Platzposten ist eine Gruppe von Dateien, die zusammen Platz belegen.
//
// **Nicht `Posten`**, und das ist kein Geschmack: `domain.Posten` heißt schon
// die Funktion, die aus einem Deck seine Posten macht. Zwei Bedeutungen für
// ein Wort in derselben Schicht wären eine Verwechslung, die der Übersetzer
// findet und ein Leser nicht.
type Platzposten string

// Die sechs Platzposten.
//
// **Vier waren es bis zum 2026-09-07**; das Protokoll und der Rest kamen mit
// Werkstück 6 dazu. Diese Überschrift stand danach noch auf „vier" — der
// Nachlauf zum Zweitblick hat es angestrichen.
const (
	// PostenBestand ist die Bestandsdatei samt ihren Begleitern (`-wal`,
	// `-shm`). **Nicht räumbar** — sie ist die Anwendung.
	PostenBestand Platzposten = "bestand"

	// PostenBilder sind die eingesetzten Bilder. **Nicht räumbar**: sie
	// gehören zu Karten, und eine Karte ohne ihr Bild ist eine halbe Karte.
	PostenBilder Platzposten = "bilder"

	// PostenPakete sind die bezogenen Vorlagenpakete. **Nicht hier räumbar**,
	// und das ist eine Entscheidung: ein Paket entfernt man auf der Seite
	// *Pakete*, wo daneben steht, welche Sets es benutzen.
	PostenPakete Platzposten = "pakete"

	// PostenVorsicherungen sind die Sicherungen vor Migrationen. **Räumbar**,
	// bis auf den Rückweg — welche Datei das ist, rechnet `internal/app`
	// (Werkstück 5 von Phase 4.13).
	PostenVorsicherungen Platzposten = "vorsicherungen"

	// PostenProtokoll ist `mse.log` samt seiner Vorfassung. **Räumbar**, bis
	// auf die laufende Datei — in sie wird gerade geschrieben (Werkstück 6).
	PostenProtokoll Platzposten = "protokoll"

	// PostenSonstiges ist alles im Anwendungsordner, was kein anderer Posten
	// beansprucht. **Nicht räumbar**, und das ist der Punkt: was hier landet,
	// versteht die Anwendung nicht — es zu löschen wäre eine Entscheidung
	// über fremde Dateien.
	//
	// # Warum es diesen Posten gibt
	//
	// **Weil eine Aufzählung eine Zusage ist** (§10). Am 2026-09-07 wog der
	// Anwendungsordner des Auftraggebers 13 GB, und die vier Posten nannten
	// davon rund 1,5 GB — beiseitegelegte Bestände, ein Protokoll und zwei
	// Verzeichnisse aus früheren Läufen fielen durch. Ein längeres Verzeichnis
	// von Namen hätte dieselbe Lücke morgen wieder; ein Rest hat sie nie.
	PostenSonstiges Platzposten = "sonstiges"
)

// PostenAlle gibt alle Platzposten in der Reihenfolge, in der sie gezeigt
// werden.
//
// **Der Bestand zuerst**, weil er die Frage beantwortet, mit der jemand
// herkommt: „wie groß sind meine Daten?"
func PostenAlle() []Platzposten {
	return []Platzposten{
		PostenBestand, PostenBilder, PostenPakete, PostenVorsicherungen,
		PostenProtokoll, PostenSonstiges,
	}
}

// Raeumbar sagt, ob ein Posten sich von hier aus wegräumen lässt.
//
// **Zwei von sechs**, und beide lassen etwas stehen: bei den Vorsicherungen
// den Rückweg, beim Protokoll die laufende Datei. *Was* stehen bleibt, weiß
// `internal/app` — es hängt an den Dateien und nicht an der Fachlichkeit.
func (p Platzposten) Raeumbar() bool {
	return p == PostenVorsicherungen || p == PostenProtokoll
}

// PostenGefuehrt sagt, ob ein Name ein Platzposten ist.
func PostenGefuehrt(name string) bool {
	for _, posten := range PostenAlle() {
		if string(posten) == name {
			return true
		}
	}

	return false
}
