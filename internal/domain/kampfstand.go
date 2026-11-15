package domain

// **Der Stand des Kampfes** — Werkstück M6 der Regelmaschine
// (#1608).
//
// # Warum er am Zug hängt und nicht am Objekt
//
// **Weil ein Kampf eine Beziehung ist und kein Zustand.** „Diese Kreatur
// greift jenen Spieler an" und „diese blockt jene" sind Paare; an einem
// Objekt ließe sich nur die eine Hälfte ablegen, und die andere müsste man
// suchen. CR 506.4 nimmt ein Permanent **aus dem Kampf**, wenn es das
// Schlachtfeld verlässt — auch das ist eine Aussage über die Beziehung.
//
// **Und weil er mit dem Zug endet.** CR 511.3: nach dem Kampfende ist niemand
// mehr Angreifer oder Blocker. Ein Feld am Objekt bliebe stehen und müsste
// einzeln geräumt werden.

// Kampfstand ist, wer wen angreift und wer wen blockt — CR 508, CR 509.
//
// **Reihen und keine Tabellen**: [Partie] ist ein Wert und wird kopiert; eine
// Tabelle darin wäre geteilt. Derselbe Grund wie bei [Zugstand.Gepasst].
type Kampfstand struct {
	// Angriffe sind die erklärten Angreifer — CR 508.1.
	Angriffe []Angriff

	// Bloecke sind die erklärten Blocker — CR 509.1.
	Bloecke []Block

	// Erklaert sagt, dass die Angreifer erklärt sind — auch **keine** ist
	// eine Erklärung (CR 508.1a: „if any").
	//
	// **Ohne dieses Feld wäre ein Kampf ohne Angreifer nicht von einem zu
	// unterscheiden, in dem niemand erklärt hat** — und der Schritt wartete
	// für immer.
	Erklaert bool

	// Geblockt sagt, dass die Blocker erklärt sind — dito für CR 509.1a.
	Geblockt bool

	// Schadensschritte zählt, wie oft in diesem Kampf schon Schaden
	// ausgeteilt wurde — CR 510.4.
	//
	// **Eine Zahl und kein Ja/Nein**, weil CR 510.4 **zwei** Schritte kennt
	// und nicht einen mit einer Marke: im ersten teilt aus, wer Erstschlag
	// hat, im zweiten der Rest **und** wer Doppelschlag hat. Ein Kampf ohne
	// Erstschläger hat genau einen.
	Schadensschritte int
}

// Angriff ist eine Kreatur und das, was sie angreift — CR 508.1a.
//
// **Ziel ist eine Kennung**: ein Sitz, ein Planeswalker oder eine Kampfkarte
// (CR 508.1a). Welches davon, sagt die Kennung selbst — dieselbe Wahl wie bei
// [Ziel].
type Angriff struct {
	Kreatur string
	Ziel    string
}

// Block ist eine Kreatur und der Angreifer, den sie blockt — CR 509.1a.
//
// **Eine Zeile je Paar**: eine Kreatur darf mehrere blocken (CR 509.1a nennt
// es ausdrücklich), und ein Angreifer darf mehrfach geblockt werden.
type Block struct {
	Kreatur   string
	Angreifer string
}

// greiftAn sagt, ob diese Kreatur angreift.
func (k Kampfstand) greiftAn(kreatur string) bool {
	for _, angriff := range k.Angriffe {
		if angriff.Kreatur == kreatur {
			return true
		}
	}

	return false
}

// blocker gibt die Kreaturen, die diesen Angreifer blocken — CR 509.1h.
func (k Kampfstand) blocker(angreifer string) []string {
	var wer []string

	for _, block := range k.Bloecke {
		if block.Angreifer == angreifer {
			wer = append(wer, block.Kreatur)
		}
	}

	return wer
}
