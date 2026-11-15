package msepaket

// **`statistics dimension:` — wonach ein Spielpaket auszählen lässt**
// (#1338,
// Werkstück C von Phase 4.11).
//
// # Was nachgelesen wurde (§15)
//
// `src/data/statistics.hpp`, Zeilen 25–45, und die Reflexion in
// `src/data/statistics.cpp`, Zeilen 64–78 (Fassung f9d9356 vom 2021-07-06).
// Elf Angaben, und zwei davon sind Mehrfachschlüssel: jede `group:`-Zeile
// hängt einen Namen an die Reihenfolge, `colors:` ist ein Unterblock aus
// Name und Farbe.
//
// # Was hier **nicht** steht
//
// **`statistics category:`.** MSE2 kennt sie (`StatsCategory`, Zeilen 51–70);
// sie kreuzt zwei Dimensionen zu einem Diagramm. Im ganzen Pack kommt der
// Schlüssel **null-mal** vor (gemessen am 2026-08-31 über
// `full-magic-pack/data`), und MSE2 baut die Kategorien dann selbst aus den
// Dimensionen. Ein Leser dafür hätte keinen Fall zu lesen — die Reichweite
// dieser Zahl ist dieses Pack, und das steht hier, damit niemand sie für eine
// Aussage über jedes Paket hält.
//
// # Und warum es hier steht und nicht in der Domäne
//
// **Anti-Korruptionsschicht** (§2, §16): [Auszaehldimension] ist die Gestalt
// der **fremden Datei**, samt Skripttext und Farbzeichenketten in MSEs
// Schreibweise. Was `service` davon braucht, bekommt es ausgerechnet.

import "strconv"

// Auszaehldimension ist ein `statistics dimension:`-Block.
type Auszaehldimension struct {
	// Name ist der Name der Dimension — er ist zugleich ihr Schlüssel.
	Name string

	// Beschreibung ist der Satz darunter. MSE zeigt ihn in der Statuszeile.
	Beschreibung string

	// Rang ist `position hint:` — **als Zeichenkette**.
	//
	// Im Pack steht dort `0000`, `0100`, `0210` … MSE liest ein `int`; wer
	// hier eine Zahl daraus machte, verlöre die führenden Nullen und damit
	// die Absicht, nach der Zeichenkette zu ordnen.
	Rang string

	// Sinnbild ist der Pfad zum Symbol im Paket — `icon:`.
	Sinnbild string

	// Skript rechnet den Wert **je Karte**.
	Skript string

	// Globalskript rechnet **einmal für das ganze Set** — `global_script:`.
	//
	// **MSE 2.1.2 kennt diesen Schlüssel nicht.**
	// `IMPLEMENT_REFLECTION_NO_GET_MEMBER(StatsDimension)`
	// (`src/data/statistics.cpp` Zeilen 64–77, Fassung `f9d9356` vom
	// 2021-07-06) führt elf Angaben, und diese ist keine davon. `magic.mse-game`
	// schreibt `mse version: 2.5.8`; der Schlüssel gehört einem neueren MSE,
	// dessen Quelltext hier nicht vorliegt
	// (#1355).
	//
	// **Gefragt statt erschlossen**: das Orakel scheitert an derselben Stelle.
	//
	//	ERROR: Variable not set: global_value
	//	         in function average_opening_hand_statistic
	//
	// **Genau eine der fünfzig Dimensionen trägt eines** — gemessen am
	// 2026-08-31 über `magic.mse-game`; im ganzen Pack kommt der Schlüssel in
	// keiner anderen Datei vor.
	Globalskript string

	// Numerisch sagt, ob die Werte Zahlen sind und als solche geordnet werden.
	Numerisch bool

	// Klassenbreite ist `bin size:` — benachbarte Zahlen zu einer Klasse
	// zusammenfassen. 0 heißt: nicht zusammenfassen.
	Klassenbreite float64

	// LeereZeigen ist `show empty:` — ob eine Karte ohne Wert mitzählt.
	LeereZeigen bool

	// Zerlegen ist `split list:` — ein Wert mit Kommas wird zu mehreren.
	Zerlegen bool

	// Gruppen ist die **erklärte** Reihenfolge der Werte.
	Gruppen []string

	// Farben ordnet einem Wert seine Farbe zu, in MSEs Schreibweise.
	Farben map[string]string
}

// Auszaehldimensionen liest die `statistics dimension:`-Blöcke eines
// Spielpakets.
//
// **Über den ganzen Kopf und nicht über eine Datei.** `magic.mse-game` legt
// sie in `statistics_<sprache>` und bindet sie über `include localized file:`
// ein; ein Leser, der `game` allein läse, fände dort keinen einzigen Block.
func (p *Paket) Auszaehldimensionen() []Auszaehldimension {
	bloecke := p.Kopf.Alle("statistics dimension")
	dimensionen := make([]Auszaehldimension, 0, len(bloecke))

	for _, knoten := range bloecke {
		dimensionen = append(dimensionen, Auszaehldimension{
			Name:          wertOderLeer(knoten.Ersten("name")),
			Beschreibung:  wertOderLeer(knoten.Ersten("description")),
			Rang:          wertOderLeer(knoten.Ersten("position hint")),
			Sinnbild:      wertOderLeer(knoten.Ersten("icon")),
			Skript:        skripttextOderLeer(knoten.Ersten("script")),
			Globalskript:  skripttextOderLeer(knoten.Ersten("global_script")),
			Numerisch:     jaOderNein(knoten.Ersten("numeric")),
			Klassenbreite: kommazahlOderNull(knoten.Ersten("bin size")),
			LeereZeigen:   jaOderNein(knoten.Ersten("show empty")),
			Zerlegen:      jaOderNein(knoten.Ersten("split list")),
			Gruppen:       werteVon(knoten.Alle("group")),
			Farben:        farbtafelAus(knoten.Ersten("colors")),
		})
	}

	return dimensionen
}

// werteVon gibt die Werte mehrerer gleichnamiger Zeilen, in ihrer Reihenfolge.
//
// **Und eine leere Liste, wo keine dasteht** — dieselbe Zusage wie überall
// sonst (§3).
func werteVon(knoten []*Knoten) []string {
	werte := make([]string, 0, len(knoten))
	for _, einer := range knoten {
		werte = append(werte, einer.Wert)
	}

	return werte
}

// farbtafelAus liest den `colors:`-Unterblock — Name auf Farbe.
func farbtafelAus(knoten *Knoten) map[string]string {
	tafel := map[string]string{}

	if knoten == nil {
		return tafel
	}

	// **Ohne Prüfung auf die Zeilenart**, und das ist gemessen: die Kinder
	// eines Blocks sind **ausschließlich** Einträge. Leerzeilen,
	// Kommentarzeilen und Fortsetzungen legt der Zerteiler nicht als Kind ab
	// — nachgeprüft am 2026-08-31 an einem `colors:`-Block mit allen dreien.
	// Ein `if zeile.Zeilenart != ArtEintrag` wäre ein Zweig, den kein Aufruf
	// betreten kann (§3).
	for _, zeile := range knoten.Kinder {
		tafel[zeile.Schluessel] = zeile.Wert
	}

	return tafel
}

// kommazahlOderNull liest eine Maßangabe — und gibt null, wo keine steht.
//
// **Unlesbares gibt null und keinen Befund.** MSE liest `bin_size` über
// seinen Reader, der eine unbrauchbare Zahl meldet und den Vorgabewert
// stehen lässt; eine Dimension, die deshalb ganz ausfiele, wäre strenger als
// das Original (§15, Regel 4).
func kommazahlOderNull(knoten *Knoten) float64 {
	if knoten == nil {
		return 0
	}

	zahl, fehler := strconv.ParseFloat(knoten.Wert, 64)
	if fehler != nil {
		return 0
	}

	return zahl
}
