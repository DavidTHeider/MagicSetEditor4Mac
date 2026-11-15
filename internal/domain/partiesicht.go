package domain

// **Die Projektion je Rolle** — Werkstück G1 von Phase 5.5
// (#1458).
//
// # Warum sie hier steht und nicht im Dienst
//
// **Weil sie die Regel selbst ist und nicht ihre Anwendung.** Ein Dienst, der
// sie rufen **müsste**, wäre ein Dienst, der sie vergessen kann — und das
// Vergessen fiele nicht auf: die Antwort sähe vollständiger aus, nicht
// falscher. Die Zonentabelle steht aus demselben Grund in `spielobjekt.go`.
//
// **Und weil sie nichts braucht.** Sie ist eine Rechnung über eine [Partie]
// und eine Kennung; kein Bestand, kein Netz, keine Uhr.
//
// # Das Hostprivileg ist eine Ausnahme mit Namen
//
// **Der Wirt sieht in seiner Projektion nicht mehr als jeder andere Sitz.**
// Technisch **kennt** er den Vollzustand — er hält ihn —, und dafür gibt es
// [Partie.Vollzustand] mit einem eigenen Namen und einem eigenen Lauf. Wer
// das umdrehte und dem Wirt eine reichere Sicht gäbe, hätte den Vollzustand
// zum Normalfall gemacht, und jede Prüfung darüber liefe ins Leere.
//
// # Und eine unbekannte Kennung bekommt die engste Sicht
//
// **Nicht die weiteste.** Wer nicht am Tisch sitzt, sieht so viel wie ein
// Zuschauer — das ist der sichere Ausgang, und er ist der Unterschied
// zwischen einem Tippfehler und einer offenen Hand.

// Partiesicht ist eine Partie, wie einer sie sehen darf.
type Partiesicht struct {
	PartieID string
	Fassung  int64

	Spiel     string
	Regelwerk Regelwerk
	Stand     Partiestand

	// Fuer ist die Kennung, für die gerechnet wurde.
	//
	// **Sie steht mit in der Antwort**, damit ein Empfänger merkt, wenn er
	// die Sicht eines anderen bekommen hat.
	Fuer string

	// Wirt steht offen da — die Vertrauensgrenze wird genannt, nicht
	// versteckt (#1458).
	Wirt string

	// Ausgang ist, wie die Partie ausgegangen ist — CR 104 (M17, #1618).
	//
	// **Für jeden derselbe.** Ein Ergebnis ist öffentlich: wer verloren hat,
	// erfährt es, und wer zusieht, sieht es. Ein Ausgang, den nur der Wirt
	// kennt, wäre ein Ergebnis ohne Zeugen.
	Ausgang Partieausgang

	// Zug ist, wo die Partie im Zug steht — CR 102.1, CR 117, CR 500
	// (G9, #1592).
	//
	// **Ungekürzt und für jeden derselbe.** Wer am Zug ist und wer Priorität
	// hat, ist am Tisch öffentlich; ein Sitz, der es nicht wüsste, könnte
	// nicht passen, und ein Zuschauer sähe einem Spiel zu, dessen Takt er
	// nicht kennt.
	//
	// **Eine Null in [Zugstand.Nummer] heißt: es geht nicht zugweise zu** —
	// die Partien aus G1 bis G8 sind Sandkästen und bleiben es.
	Zug Zugstand

	Sitze     []Sichtsitz
	Zuschauer []Zuschauer

	Objekte []Sichtobjekt

	// Mischpflichten sind die verborgenen Zonen, die jemand angesehen und
	// seitdem nicht gemischt hat — **für jeden dieselben**.
	Mischpflichten []Mischpflicht

	// Legendenwahlen sind die offenen Fragen aus CR 704.5j — **für jeden
	// dieselben** (M12,
	// #1614).
	//
	// **Ungekürzt und nicht nur für den Gefragten.** Wer wählen muss, ist am
	// Tisch öffentlich: bis er geantwortet hat, steht das Spiel, und ein
	// Gegenüber, das den Grund nicht sähe, hielte es für einen Ausfall.
	// Derselbe Grund wie beim Zugstand eine Zeile darüber.
	Legendenwahlen []Legendenwahl

	// Ausgeloest sind die Fähigkeiten, die ausgelöst haben und auf den
	// Stapel warten — CR 603.2 (M9,
	// #1611).
	//
	// **Für jeden dieselben**, und aus demselben Grund wie die
	// Legendenwahlen: eine ausgelöste Fähigkeit ist am Tisch öffentlich
	// (CR 603.2 lässt sie „warten", nicht verbergen), und solange geordnet
	// wird, steht das Spiel.
	Ausgeloest []Ausloeser

	// Ausloeserreihung ist der Sitz, der gerade ordnen soll — CR 603.3b.
	Ausloeserreihung string

	// Befunde sind Auffälligkeiten am Tisch, die die Partie **nicht**
	// unbenutzbar machen — §8 (M14,
	// #1616).
	//
	// **Heute steht darin eine Sache**: eine Schlüsselwortfähigkeit auf
	// einer Karte, die dieses Programm nicht durchsetzt. Ohne sie gälte an
	// derselben Kreatur das eine Wort und das andere nicht, und niemand
	// sähe den Unterschied.
	Befunde []string

	// Vorrat sind die Kartenkennungen des **eigenen** Decks (G2,
	// #1459).
	//
	// # Wofür es sie gibt
	//
	// **Damit die Bilder da sind, bevor die Karte kommt.** Das Issue
	// verlangt es wörtlich: „Eigene Decks, Rückseite … vorladen." Ohne diese
	// Liste gibt es nichts zum Vorladen — die Objekte in der Bibliothek
	// tragen keine Kennung, und das bleibt auch so.
	//
	// # Und warum das kein Widerspruch zur Bibliothek ist
	//
	// **Weil zwei verschiedene Dinge geheim sind.** Die **Reihenfolge** einer
	// Bibliothek ist das Geheimnis — wer sie kennt, kennt den nächsten Zug.
	// **Welche Karten im eigenen Deck liegen**, ist kein Geheimnis vor dem
	// eigenen Spieler: er hat das Deck gebaut und eingereicht.
	//
	// **Sie ist deshalb eine Menge und keine Reihe**: sortiert und ohne
	// Wiederholung, damit aus ihr nichts über die Lage einer Karte folgt.
	// Vier Wälder stehen einmal darin.
	//
	// **Und sie steht nur in der eigenen Sicht.** Ein Zuschauer bekommt
	// nichts, ein Gegenüber bekommt nichts — für eine unbekannte Kennung
	// ist sie leer.
	Vorrat []int64
}

// Sichtsitz ist ein Sitz ohne das, was an ihm geheim ist.
type Sichtsitz struct {
	Kennung string
	Name    string

	// Deckabdruck ist der Abdruck des Abzugs — **nicht** der Abzug.
	//
	// **Er sagt, dass zwei Seiten dasselbe Deck meinen**, ohne es zu
	// zeigen; die Liste selbst ist im Turnier ein eingereichtes Dokument
	// und am Küchentisch das Geheimnis des Gegenübers.
	Deckabdruck string

	// Karten ist, wie viele im Abzug lagen — eine Zahl verrät keine Karte.
	Karten int

	// Wartet sagt, dass dieser Sitz auf das Deck seines Gastes wartet (P9b,
	// #2377) — **nicht, auf wen**: eine Peerkennung ist kein Name am Tisch.
	Wartet bool

	// Eingeladen ist die Peerkennung des eingeladenen Gastes — **nur in der
	// Sicht des Wirts**: er gibt den Sitz frei und muss ihn dem Richtigen
	// geben (P9b, #2377).
	Eingeladen string

	// Leben und Vergiftung machen den Sitz zum Spieler — CR 119.1 und
	// CR 122.1f (G9, #1592).
	//
	// **Ungekürzt und für jeden**: ein Lebensstand ist am Tisch öffentlich.
	Leben      int
	Vergiftung int

	// Manavorrat ist das ungenutzte Mana dieses Spielers — CR 106.4 (M3,
	// #1605).
	//
	// # Warum ihn jeder sieht
	//
	// **Weil CR 106.4a und CR 106.4b es verlangen**: wer Mana übrig hat, sagt
	// beim Bezahlen und beim Passen an, was noch dasteht. Ein Vorrat, den nur
	// sein Besitzer sähe, machte diese Ansage unmöglich — und sie ist eine
	// Regel und keine Höflichkeit. Derselbe Satz wie beim Lebensstand
	// darüber.
	Manavorrat Manavorrat

	// ProfilID ist das Profil, an das dieser Sitz gebunden ist (G8,
	// #1465)
	// — null heißt: keines.
	//
	// # Warum sie in der Sicht steht
	//
	// **Damit das Fenster weiß, welchen Sitz es zeigen soll.** Wer das
	// gezeigte Profil wechselt, bekommt die Sicht **dieses** Sitzes; ohne
	// diese Zahl müsste die Oberfläche raten oder eine zweite Liste führen.
	//
	// **Und sie ist kein Geheimnis.** Sie sagt, dass ein Sitz zu einer Ablage
	// auf **diesem** Rechner gehört — nicht, was darin steht, und über einen
	// Menschen gar nichts. Ein Gast an einem fremden Tisch sieht eine Zahl,
	// die bei ihm nichts bedeutet.
	ProfilID int64

	// Strahlung sind die Radioaktivitätsmarken dieses Sitzes — CR 728.1
	// (M20, #1621).
	//
	// **Für jeden sichtbar**, aus demselben Grund wie die Vergiftung
	// darüber: CR 728.1 nennt sie eine Art Zähler, die ein **Spieler**
	// haben kann, und Marken an einem Spieler liegen offen.
	Strahlung int

	// Commander ist die Kartenkennung des Commanders — CR 903.3 (M19,
	// #1620).
	//
	// **Kein Geheimnis**: er beginnt die Partie in der Kommandozone, und
	// die ist nach CR 400.2 eine öffentliche Zone. **Leer heißt: keiner**,
	// und das ist der Normalfall — es gibt eine Commandervariante und acht
	// andere Spielweisen.
	Commander string

	// Commandersteuer zählt, wie oft dieser Sitz seinen Commander schon aus
	// der Kommandozone gewirkt hat — CR 903.8.
	//
	// **Eine Zahl und kein Betrag**, wie am Sitz selbst: was sie kostet,
	// rechnet [Commandersteuer] aus.
	Commandersteuer int

	// Commanderschaden ist der Kampfschaden **je Commander** — CR 903.10a.
	//
	// **Je Kennung und nicht als Summe**: zwei mit je zwanzig sind vierzig
	// und trotzdem keine einundzwanzig.
	//
	// **Eine eigene Tabelle je Sicht** — dieselbe Regel wie beim Manavorrat
	// darüber: eine Sicht ist eine Projektion und kein Griff in den Zustand.
	Commanderschaden map[string]int

	// Leergezogen vermerkt, dass dieser Sitz aus einer **leeren** Bibliothek
	// ziehen wollte — CR 704.5b.
	//
	// **Er gehört in die Sicht, weil er am Tisch geschehen ist.** Wer zieht
	// und nichts findet, tut das vor allen; die nächste Prüfung der
	// zustandsbasierten Aktionen entscheidet, was daraus folgt.
	Leergezogen bool

	// Eroeffnung, Mulligans und Unterlegen sind der Stand im Beginn der
	// Partie — CR 103.5 (Fassung 2026-09-25),
	// #1839.
	//
	// **Keines ist geheim**: ob jemand einen Mulligan nimmt, wird am Tisch
	// erklärt, und wie viele Karten unter die Bibliothek gehen, zählt jeder
	// mit. **Welche**, bleibt verborgen — das steht in keinem der drei.
	Eroeffnung Eroeffnungsstand
	Mulligans  int
	Unterlegen int
}

// Sichtobjekt ist ein Spielobjekt, so weit es gezeigt werden darf.
type Sichtobjekt struct {
	// KarteID ist die Zeile im Bestand, aus der diese Karte stammt (G2,
	// #1459) — **null, wo nicht hineingesehen werden darf**.
	//
	// Sie ist der Weg zum Bild. Sie an einem geheimen Objekt mitzugeben
	// verriete die Karte über einen Umweg, und die Projektion daneben sähe
	// weiter richtig aus.
	KarteID int64

	// Eigenschaften sind, was diese Karte **ist** — **nil, wo nicht
	// hineingesehen werden darf** (M2,
	// #1604).
	//
	// **Unter derselben Regel wie [Sichtobjekt.KarteID]**, und aus demselben
	// Grund: wer die Stärke eines Objekts auf der gegnerischen Hand läse,
	// wüsste, welche Karte es ist — und die Projektion daneben sähe weiter
	// richtig aus.
	Eigenschaften *Sichteigenschaften

	Kennung string
	Art     Objektart

	Beherrscher string
	Eigner      string
	Zone        Zone

	// Bild ist die undurchsichtige Kennung des Kartenbilds — P9 (#2377),
	// Vermerk 0198. **Leer, wo nicht hineingesehen werden darf**, wie
	// [Sichtobjekt.KarteID]; anders als sie zeigt sie in keinen Bestand und
	// ist an jedem Ende dieselbe: wer ein Bild nicht selbst zeichnen kann,
	// fragt den, dem das Deck gehört, und erkennt die Antwort an ihr.
	Bild string

	// Karte und Name sind leer, wo nicht hineingesehen werden darf.
	Karte string
	Name  string

	// Ordnung ist **nur in offenen Zonen** gefüllt.
	//
	// **Die Reihenfolge ist das eigentliche Geheimnis einer Bibliothek** —
	// wer sie kennt, kennt den nächsten Zug, auch ohne eine einzige Karte
	// zu lesen.
	Ordnung int

	Getappt bool

	// Verdeckt heißt: es **liegt** verdeckt — eine Tatsache auf dem Tisch,
	// die jeder sieht.
	Verdeckt bool

	// Geheim heißt: **dieser** Empfänger darf nicht hineinsehen.
	//
	// # Warum das zwei Felder sind und nicht eines
	//
	// **Weil es zwei verschiedene Tatsachen sind.** Eine verdeckt liegende
	// Kreatur liegt für **alle** verdeckt — und ihr Beherrscher weiß
	// trotzdem, was sie ist. Wer beides in ein Feld legte, könnte die zwei
	// Fälle nicht mehr auseinanderhalten: „liegt verdeckt, und ich weiß, was
	// es ist" gegen „liegt offen, aber ich darf nicht hinsehen".
	//
	// **Und es macht diesen Typ von [Spielobjekt] unterscheidbar.** Wären
	// beide feldgleich, ließe sich der eine in den anderen umwandeln — und
	// genau das soll nicht gehen: die Projektion ist der einzige Weg von der
	// Wahrheit zu dem, was jemand sehen darf.
	Geheim bool

	HaengtAn string
	Haufen   string
	Zaehler  map[string]int
}
