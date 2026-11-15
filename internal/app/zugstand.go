package app

// **Der Zugstand an der Naht** — Werkstück G9 von Phase 5.5
// (#1592).
//
// # Eine eigene Datei, und der Schnitt folgt der Frage
//
// **Nebenan steht, wie eine Partie aussieht**; hier steht, **wo sie im Zug
// ist**. Das ist die Auskunft, die das Fenster braucht, um zu sagen „du bist
// dran" — und `partie.go` steht bei 352 Zeilen (§2).
//
// # Warum Leben und Marken am Sitz stehen und nicht hier
//
// **Weil sie zum Sitz gehören und nicht zum Zug.** Ein Spieler hat sein
// Leben, auch wenn keiner am Zug ist; der Zugstand sagt, wer handeln darf.

import "github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"

// ZugDTO ist, wo die Partie im Zug steht.
//
// **Für jeden derselbe.** Wer am Zug ist und wer Priorität hat, ist am Tisch
// öffentlich (CR 102.1, CR 117); ein Sitz, der es nicht wüsste, könnte nicht
// passen.
type ZugDTO struct {
	// Nummer ist der wievielte Zug — **null heißt: es geht nicht zugweise
	// zu**, und dann steht am Tisch kein Takt.
	Nummer int `json:"nummer"`

	// Aktiv ist die Sitzkennung dessen, dessen Zug es ist — CR 102.1.
	Aktiv string `json:"aktiv"`

	// Starter ist, wer den ersten Zug genommen hat — CR 103.1 (Fassung
	// 2026-09-25).
	Starter string `json:"starter"`

	Phase   string `json:"phase"`
	Schritt string `json:"schritt"`

	// Prioritaet ist, wer gerade handeln darf — **leer heißt: niemand**
	// (CR 502.4, CR 514.3).
	Prioritaet string `json:"prioritaet"`

	// Abwurf ist, wie viele Karten der aktive Spieler noch abwerfen muss —
	// CR 514.1.
	Abwurf int `json:"abwurf"`

	// Gepasst sind die, die seit der letzten Handlung gepasst haben —
	// CR 117.4.
	//
	// **Eine Liste und keine Zahl**: „alle nacheinander" ist eine Aussage
	// über **wer**, und am Tisch sieht man es einem Menschen an.
	Gepasst []string `json:"gepasst"`

	// Angriffe und Bloecke sind der Kampfstand — CR 508.1, CR 509.1
	// (#1623).
	//
	// # Warum sie erst hier dazukommen
	//
	// **Weil `M6` sie gebaut hat und niemand sie sehen konnte.** Der
	// Zugstand geht ungekürzt in die Sicht; kein DTO trug ihn weiter, und
	// damit stand an **keiner** der vier Flächen, wer angreift. Das ist
	// dieselbe Gestalt wie bei Leben und Vergiftungsmarken (M3) — eine
	// Regel, die greift, ohne dass jemand sieht, worauf (§10).
	//
	// **Und es ist mehr als eine Anzeige**: Blocker zu erklären setzt
	// voraus, zu wissen, wer angreift (CR 509.1a).
	Angriffe []KampfpaarDTO `json:"angriffe"`
	Bloecke  []KampfpaarDTO `json:"bloecke"`
}

// KampfpaarDTO ist eine Kreatur und das, wogegen sie steht.
//
// **Ein Verbund für beide Richtungen**, weil es dieselbe Gestalt ist: bei
// einem Angriff steht das Ziel, bei einem Block der Angreifer. Zwei Typen
// mit je einem belegten Feld wären zwei Listen, die niemand gegeneinander
// halten kann.
type KampfpaarDTO struct {
	Kreatur string `json:"kreatur"`

	// Ziel ist belegt bei einem Angriff — ein Sitz, ein Planeswalker oder
	// eine Kampfkarte (CR 508.1a).
	Ziel string `json:"ziel,omitempty"`

	// Angreifer ist belegt bei einem Block — CR 509.1a.
	Angreifer string `json:"angreifer,omitempty"`
}

// zugAus bildet den Zugstand auf sein DTO ab.
//
// **Leere Listen und kein `nil`** — derselbe Satz wie überall an dieser
// Grenze: im Fenster stünde sonst `null`, und jede Schleife darüber fiele um.
func zugAus(zug domain.Zugstand) ZugDTO {
	gepasst := zug.Gepasst
	if gepasst == nil {
		gepasst = []string{}
	}

	return ZugDTO{
		Nummer:     zug.Nummer,
		Aktiv:      zug.Aktiv,
		Starter:    zug.Starter,
		Phase:      string(zug.Abschnitt.Phase),
		Schritt:    string(zug.Abschnitt.Schritt),
		Prioritaet: zug.Prioritaet,
		Abwurf:     zug.Abwurf,
		Gepasst:    gepasst,
		Angriffe:   angriffeAus(zug.Kampf.Angriffe),
		Bloecke:    bloeckeAus(zug.Kampf.Bloecke),
	}
}

// angriffeAus bildet die erklärten Angriffe ab — CR 508.1a.
func angriffeAus(angriffe []domain.Angriff) []KampfpaarDTO {
	paare := make([]KampfpaarDTO, 0, len(angriffe))

	for _, angriff := range angriffe {
		paare = append(paare, KampfpaarDTO{
			Kreatur: angriff.Kreatur, Ziel: angriff.Ziel,
		})
	}

	return paare
}

// bloeckeAus bildet die erklärten Blöcke ab — CR 509.1a.
func bloeckeAus(bloecke []domain.Block) []KampfpaarDTO {
	paare := make([]KampfpaarDTO, 0, len(bloecke))

	for _, block := range bloecke {
		paare = append(paare, KampfpaarDTO{
			Kreatur: block.Kreatur, Angreifer: block.Angreifer,
		})
	}

	return paare
}

// AusgangDTO ist, wie die Partie ausgegangen ist — CR 104 (M17,
// #1618).
//
// **Für jeden derselbe.** Ein Ergebnis ist öffentlich: wer verloren hat,
// erfährt es, und wer zusieht, sieht es.
type AusgangDTO struct {
	// Sieger ist die Sitzkennung des Gewinners — **leer heißt: keiner**, und
	// solange auch `unentschieden` falsch ist, läuft die Partie noch.
	Sieger string `json:"sieger"`

	Unentschieden bool `json:"unentschieden"`

	// Ausgeschieden sind die, die gegangen sind, in der Reihenfolge, in der
	// sie gingen.
	Ausgeschieden []AusgeschiedenDTO `json:"ausgeschieden"`
}

// AusgeschiedenDTO ist ein Sitz, der die Partie verlassen hat.
type AusgeschiedenDTO struct {
	Sitz string `json:"sitz"`

	// Grund ist einer der vier aus CR 104.3 — `aufgabe`, `leben`,
	// `bibliothek` oder `gift`.
	Grund string `json:"grund"`
}

// ausgangAus bildet den Ausgang auf sein DTO ab.
//
// **Leere Liste und kein `nil`** — derselbe Satz wie überall an dieser Grenze.
func ausgangAus(ausgang domain.Partieausgang) AusgangDTO {
	gegangen := make([]AusgeschiedenDTO, 0, len(ausgang.Ausgeschieden))

	for _, einer := range ausgang.Ausgeschieden {
		gegangen = append(gegangen, AusgeschiedenDTO{
			Sitz: einer.Sitz, Grund: string(einer.Grund),
		})
	}

	return AusgangDTO{
		Sieger:        ausgang.Sieger,
		Unentschieden: ausgang.Unentschieden,
		Ausgeschieden: gegangen,
	}
}
