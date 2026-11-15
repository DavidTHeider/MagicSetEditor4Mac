package domain

// **Die Ordnung in einer Zone und die Zonenstelle, an die gelegt wird** —
// Werkstück G1 von Phase 5.5
// (#1458).
//
// # Warum eine Zone „geordnet" heißt und nicht „sortiert"
//
// **Weil niemand sie sortiert.** In einer geordneten Zone bedeutet die Zonenstelle
// etwas — oben liegt, was als Nächstes kommt —, und genau deshalb ist die
// Reihenfolge einer Bibliothek ihr Geheimnis.
//
// # Und der Friedhof gehört dazu
//
// **CR 400.5** (Fassung 2026-08-19) nennt drei Zonen, deren Reihenfolge sich
// nur ändern lässt, wenn eine Regel oder ein Effekt es erlaubt: Bibliothek,
// **Friedhof** und Stapel — so zeichnet es auch der Tisch
// (`frontend/src/tisch/zonenfeld.ts`).
//
// **Was daran hängt, ist keine Kleinigkeit**: CR 404.1 legt eine Karte
// **oben** auf den Friedhof, und der Abwurf im Aufräumschritt (CR 514.1) tut
// genau das. Als ungeordnete Zone bekäme jede Karte darin die Ordnung null,
// und „oben auf den Friedhof" würde abgewiesen.
//
// # Und warum drei Zonenstellen und nicht zwei
//
// **Weil „eingemischt" etwas anderes ist als „irgendwohin".** Eine Karte, die
// zufällig einsortiert wird, ist für den Gegner ununterscheidbar von einer
// gemischten Bibliothek; eine, die „in die Mitte" gelegt wird, ist es nicht.
// Der Unterschied ist der ganze Zweck der dritten Zonenstelle.

// Geordnet sagt, ob die Reihenfolge in dieser Zone etwas bedeutet.
//
// **Eine unbekannte Zone ist keine geordnete** — derselbe Satz wie bei
// [Zone.Sicht]: was niemand kennt, bekommt die vorsichtige Antwort.
func (z Zone) Geordnet() bool {
	switch z {
	case ZoneBibliothek, ZoneStapel, ZoneFriedhof:
		return true

	case ZoneHand, ZoneTisch, ZoneExil, ZoneKommando:
		return false
	}

	return false
}

// Zonenstelle ist, wohin in einer geordneten Zone gelegt wird.
type Zonenstelle string

const (
	// ZonenstelleOben ist, was als Nächstes kommt.
	ZonenstelleOben Zonenstelle = "oben"

	// ZonenstelleUnten ist die andere Seite — und in einer Bibliothek die
	// Handlung, die eine Karte für diese Partie meist erledigt.
	ZonenstelleUnten Zonenstelle = "unten"

	// ZonenstelleZufaellig legt sie irgendwohin, ohne dass jemand weiß, wohin.
	//
	// **Nicht dasselbe wie Mischen**: hier bewegt sich **eine** Karte, und
	// die übrige Reihenfolge bleibt, wie sie war. Wer nach einem Blick in
	// die Bibliothek nur eine Karte zufällig einsteckt, hat den Rest
	// weiterhin gesehen — deshalb löst diese Zonenstelle **keine** Mischpflicht.
	ZonenstelleZufaellig Zonenstelle = "zufaellig"
)

// Zonenstellen gibt alle drei.
func Zonenstellen() []Zonenstelle {
	return []Zonenstelle{ZonenstelleOben, ZonenstelleUnten, ZonenstelleZufaellig}
}

// IstZonenstelle sagt, ob dieses Wort eine Zonenstelle benennt.
func IstZonenstelle(stelle Zonenstelle) bool {
	for _, bekannt := range Zonenstellen() {
		if stelle == bekannt {
			return true
		}
	}

	return false
}
