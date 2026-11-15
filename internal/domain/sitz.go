package domain

// **Ein Platz am Tisch** — Werkstück G1 von Phase 5.5
// (#1458).
//
// # Eine eigene Datei, und der Schnitt folgt der Frage
//
// **Nebenan steht, was eine Partie ist**; hier steht, **was ein Spieler
// hat**. Der Unterschied ist nicht akademisch: an einem Sitz hängt seit G9
// der Lebensstand, seit `M17` das Ausscheiden und seit `M3`
// (#1605)
// der Manavorrat — und `partie.go` stand damit bei 407 Zeilen (§2).

// Sitz ist ein Platz am Tisch — mit Deck und mit Hand.
type Sitz struct {
	// Kennung ist innerhalb dieser Partie eindeutig.
	Kennung string

	// Name ist, was auf dem Tisch steht — kein Nachweis über irgendwen.
	Name string

	// Deck ist der **eingefrorene** Abzug für diese Partie.
	Deck Partiedeck

	// Eingeladen ist die Peerkennung des Gastes, der an diesem Sitz spielt
	// und **sein Deck mitbringt** (P9b,
	// #2377).
	// Bis er sich setzt, steht der Sitz ohne Deck da — der einzige, der das
	// darf. **Leer heißt: niemand wird erwartet**, und das Deck lag beim
	// Anlegen.
	Eingeladen string

	// Leben ist der Stand dieses Spielers — CR 119.1 (Fassung 2026-08-19),
	// gesetzt aus [Regelprofil.Startleben] (G9, #1592).
	//
	// **Er darf unter null gehen, und dann scheidet der Sitz aus** —
	// CR 704.5a, geprüft von [Partie.Zustandspruefung] (M17, #1618).
	Leben int

	// Vergiftung sind die Vergiftungsmarken — CR 122.1f (Fassung
	// 2026-08-19).
	//
	// **Ab [Vergiftungsgrenze] verliert dieser Sitz** — CR 704.5c, geprüft
	// von [Partie.Zustandspruefung] (M17, #1618). G9 hat sie gezählt und
	// gezeigt; angewandt werden sie seither.
	Vergiftung int

	// Manavorrat ist das ungenutzte Mana dieses Spielers — CR 106.4 (Fassung
	// 2026-08-19), Werkstück M3
	// (#1605).
	//
	// **Er leert sich am Ende jedes Abschnitts und jeder Phase** (CR 500.5),
	// und das ist eine zugbasierte Handlung ohne Stapel (CR 703.4q). Wer ihn
	// über einen Abschnitt hinweg hielte, gäbe jedem Land unbegrenzte
	// Wirkung.
	//
	// **Leer ist die Vorgabe**, und `nil` ist leer: eine Partie aus einem
	// Bestand, der M3 nicht kannte, liest sich damit ohne Umrechnung.
	Manavorrat Manavorrat

	// Strahlung sind die Radioaktivitätsmarken dieses Sitzes — CR 728.1
	// (M20, #1621).
	//
	// **Am Spieler und nicht am Objekt**: CR 728.1 nennt sie eine Art
	// Zähler, die ein **Spieler** haben kann — dieselbe Gestalt wie die
	// Vergiftungsmarken (CR 122.1f).
	Strahlung int

	// Commander ist die Kartenkennung des Commanders — CR 903.3 (M19,
	// #1620).
	//
	// **Am Sitz und nicht an der Karte**: CR 903.3 sagt ausdrücklich, dass
	// die Bezeichnung **kein Merkmal** ist. Dieselbe Karte ist in einem
	// Deck der Commander und in einem anderen eine gewöhnliche Kreatur.
	//
	// **Leer heißt: keiner**, und das ist der Normalfall — es gibt eine
	// Commandervariante und acht andere Spielweisen.
	Commander string

	// Commandersteuer zählt, wie oft dieser Sitz seinen Commander schon aus
	// der Kommandozone gewirkt hat — CR 903.8.
	//
	// **Eine Zahl und kein Betrag**: was sie kostet, rechnet
	// [Commandersteuer] aus, und die Regel darf sich ändern, ohne dass
	// jeder gespeicherte Stand falsch wird.
	Commandersteuer int

	// Commanderschaden ist der Kampfschaden **je Commander**, den dieser
	// Sitz abbekommen hat — CR 903.10a.
	//
	// **Je Kennung und nicht als Summe**: CR 903.10a sagt „desselben
	// Commanders". Zwei mit je zwanzig sind vierzig und trotzdem keine
	// einundzwanzig.
	Commanderschaden map[string]int

	// Leergezogen vermerkt, dass dieser Sitz aus einer **leeren** Bibliothek
	// ziehen wollte — CR 704.5b (Fassung 2026-08-19).
	//
	// # Warum ein Vermerk und nicht der Verlust an Ort und Stelle
	//
	// **Weil CR 704.5b eine zustandsbasierte Aktion ist und kein Ereignis.**
	// Das Ziehen selbst hält nichts an; erst die nächste Prüfung stellt fest,
	// was daraus folgt. Der Unterschied ist messbar: dazwischen kann ein
	// Effekt Karten in die Bibliothek legen — und dann verliert niemand.
	//
	// **Er bleibt stehen, wenn er einmal gesetzt ist.** CR 704.5b spricht von
	// „seit der letzten Prüfung"; dass er zurückgesetzt würde, braucht einen
	// Effekt, der den Verlust verhindert — und die gibt es erst mit `M11`
	// (#1613).
	Leergezogen bool

	// Eroeffnung ist, wo dieser Sitz im Beginn der Partie steht — CR 103.5
	// (Fassung 2026-09-25), #1839.
	// **Leer heißt: eine Partie von vorher** — sie ist aufgebaut worden, bevor
	// es den Beginn gab, und beginnt nicht noch einmal. Der Aufbau setzt
	// [EroeffnungAusstehend].
	Eroeffnung Eroeffnungsstand

	// Mulligans ist, wie viele dieser Sitz genommen hat — **alle**, auch den
	// freien aus CR 103.5c; gezählt wird erst beim Unterlegen.
	Mulligans int

	// Unterlegen ist, wie viele Karten dieser Sitz nach seinem letzten
	// Mulligan noch unter die Bibliothek legen muss (CR 103.5). Solange es
	// mehr als null sind, gibt es weder Behalten noch einen weiteren
	// Mulligan.
	Unterlegen int

	// ProfilID bindet diesen Sitz an ein Profil **desselben Nutzers** (G8,
	// #1465).
	//
	// # Warum sie freiwillig ist
	//
	// **Weil die meisten Partien ohne sie auskommen.** Wer zu zweit am
	// Küchentisch sitzt, tippt zwei Namen. Wer gegen sich selbst spielt,
	// braucht eine Trennung, die nicht an einem getippten Wort hängt — und
	// genau dafür gibt es sie. **Null heißt: keines**, und das bleibt der
	// Normalfall.
	//
	// # Und warum sie die Sicht nicht bestimmt
	//
	// **Weil die Sicht ein Parameter bleibt.** Wer sie aus dem Profil
	// ableitete, gäbe einem Zuschauer die Sicht eines Sitzes, sobald beide
	// dasselbe Profil benutzen — und beim Spiel gegen sich selbst ist das der
	// Normalfall. Was hier steht, sagt dem **Fenster**, welchen Sitz es
	// zeigen soll; gerechnet wird weiter je Sitz ([Partie.Sicht]).
	//
	// # Und sie belegt niemanden
	//
	// **Ein Profil ist eine Ablage auf diesem Rechner**, kein Nachweis über
	// einen Menschen — derselbe Satz wie bei [Sitz.Name] darüber. Wer eine
	// belegte Identität will, nimmt die Peerkennung (G7,
	// #1464).
	ProfilID int64
}

// Wartet sagt, ob dieser Sitz auf das Deck seines Gastes wartet (P9b, #2377).
func (s Sitz) Wartet() bool {
	return s.Eingeladen != "" && s.Deck.Abdruck == ""
}
