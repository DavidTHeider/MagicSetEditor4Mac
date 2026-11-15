package domain

// **Was auf dem Tisch liegt** — Werkstück G1 von Phase 5.5
// (#1458).
//
// # Ein Objekt ist nicht seine Karte
//
// **Das ist der Satz, an dem sich ein Spielkern entscheidet**, und das Issue
// nennt ihn als Abnahme: „Karten gleichen Namens/gleicher Textur bleiben
// verschiedene Objekte." Wer zwei Wälder über ihren Namen unterschiede, hätte
// beim zweiten Wald keinen Zustand mehr — welcher ist getappt, auf welchem
// liegt eine Marke, welcher kam gerade ins Spiel.
//
// **Deshalb trägt jedes Objekt eine eigene Kennung**, und `Karte` ist ein
// Verweis auf das, was darauf steht — nicht seine Identität.
//
// # Und eine Zone weiß, wer hineinsieht
//
// **Die Tabelle steht hier und nicht in der Projektion**, weil sie eine Regel
// ist und keine Übersetzung: die Hand gehört ihrem Beherrscher, die
// Bibliothek niemandem, der Tisch allen. Wer sie in der Sichtschicht
// unterbrächte, hätte sie an dem Ort, an dem man sie am leichtesten vergisst.
//
// **Eine unbekannte Zone ist verdeckt.** Das ist der sichere Ausgang: wer bei
// einem Tippfehler „offen" zurückgäbe, zeigte eine fremde Hand.

import "slices"

// Objektart sagt, woher ein Objekt kommt.
type Objektart string

const (
	// ObjektKarte ist eine Karte aus einem Deckabzug.
	ObjektKarte Objektart = "karte"

	// ObjektSpielstein ist etwas, das die Partie erzeugt hat.
	//
	// **„Spielstein" und nicht „Token"** — das ist das offizielle deutsche
	// Wort dafür, und `Token` ist in diesem Repositorium der Zugang zu einer
	// Fläche (O2).
	ObjektSpielstein Objektart = "spielstein"

	// ObjektEmblem ist eine bleibende Wirkung ohne Körper.
	ObjektEmblem Objektart = "emblem"

	// ObjektFaehigkeit ist eine ausgelöste Fähigkeit auf dem Stapel —
	// CR 603.3 (M9,
	// #1611).
	//
	// **Ein Objekt und keine Karte**, sagt CR 603.3 ausdrücklich: es liegt
	// auf dem Stapel, es hat einen Beherrscher, und beim Verrechnen geht es
	// **nirgendwohin** — es hört auf zu existieren (CR 608.2n).
	ObjektFaehigkeit Objektart = "faehigkeit"
)

// Objektarten gibt alle vier.
func Objektarten() []Objektart {
	return []Objektart{
		ObjektKarte, ObjektSpielstein, ObjektEmblem, ObjektFaehigkeit,
	}
}

// Zone ist ein Ort, an dem ein Objekt liegen kann.
type Zone string

const (
	ZoneBibliothek Zone = "bibliothek"
	ZoneHand       Zone = "hand"
	ZoneTisch      Zone = "tisch"
	ZoneFriedhof   Zone = "friedhof"
	ZoneExil       Zone = "exil"

	// ZoneStapel ist der Regelstapel — wo Zauber und Fähigkeiten warten.
	//
	// **Nicht zu verwechseln mit einem Haufen auf dem Tisch**: der eine ist
	// eine Regelzone, der andere eine Anordnung. Ein Haufen ist deshalb kein
	// eigener Ort, sondern eine Eigenschaft der Objekte, die auf ihm liegen
	// ([Spielobjekt.Haufen]).
	ZoneStapel Zone = "stapel"

	// ZoneKommando ist die Kommandozone.
	ZoneKommando Zone = "kommando"
)

// Zonen gibt alle sieben.
func Zonen() []Zone {
	return []Zone{
		ZoneBibliothek, ZoneHand, ZoneTisch, ZoneFriedhof,
		ZoneExil, ZoneStapel, ZoneKommando,
	}
}

// IstZone sagt, ob ein Wert eine der sieben ist.
//
// **Nachgetragen mit G4** (#1461):
// ein Regelprofil darf Zusatzzonen nennen, und ein Tippfehler darin wäre
// sonst eine Zone, in die nie etwas gelangt.
func IstZone(zone Zone) bool {
	return slices.Contains(Zonen(), zone)
}

// Zonensicht sagt, wer in eine Zone hineinsieht.
type Zonensicht string

const (
	// SichtOffen heißt: jeder, auch ein Zuschauer.
	SichtOffen Zonensicht = "offen"

	// SichtBeherrscher heißt: nur, wer sie beherrscht.
	SichtBeherrscher Zonensicht = "beherrscher"

	// SichtVerdeckt heißt: niemand — auch ihr Beherrscher nicht.
	//
	// **Auch er nicht**, und das ist keine Härte: wer seine eigene
	// Bibliothek läse, hätte sie sortiert.
	SichtVerdeckt Zonensicht = "verdeckt"
)

// Sicht sagt, wer in diese Zone hineinsieht.
func (z Zone) Sicht() Zonensicht {
	switch z {
	case ZoneTisch, ZoneFriedhof, ZoneExil, ZoneStapel, ZoneKommando:
		return SichtOffen

	case ZoneHand:
		return SichtBeherrscher

	case ZoneBibliothek:
		return SichtVerdeckt
	}

	// **Was niemand kennt, sieht niemand.** Der Grund steht im Kopf.
	return SichtVerdeckt
}

// Spielobjekt ist ein Ding in einer Partie.
type Spielobjekt struct {
	// Kennung ist innerhalb der Partie eindeutig — **sie ist die Identität**.
	// **Nach außen geht sie nicht** (#2480): eine Sicht und ein Ereignis
	// tragen die [Partie.Sichtkennung].
	Kennung string

	// Wechsel ist der Stand der Sichtkennung — er steigt, wenn das Objekt
	// verborgen wird oder seine verborgene Lage sich ändert ([Sichtwechsel]).
	Wechsel int64

	Art Objektart

	// Beherrscher ist die Sitzkennung dessen, der es gerade beherrscht.
	//
	// **Gerade** — er wechselt, und das ist der Normalfall.
	Beherrscher string

	// Eigner ist, aus wessen Deck es kam — leer bei einem Spielstein.
	Eigner string

	Zone Zone

	// Karte ist der Verweis in den Deckabzug — leer bei allem, was keine ist.
	Karte string

	// Name steht dort, wo keine Karte dahintersteht.
	Name string

	// Ordnung ist der Platz innerhalb der Zone — **die Eins liegt oben**.
	//
	// `Sandkasten.platzFuer` gibt für `oben` die **Eins** und für `unten`
	// `anzahl+1`, `partieaufbau` legt die Bibliothek mit 1 beginnend, und
	// der Ziehschritt (G9, #1592) hängt genau davon ab (CR 121.1).
	Ordnung int

	// Getappt, Verdeckt und Beziehungen sind der sichtbare Zustand.
	Getappt  bool
	Verdeckt bool

	// Felder sind die **gedruckten** Eigenschaften eines Objekts ohne Karte
	// — CR 111.1 (Fassung 2026-08-19), Werkstück M12
	// (#1614).
	//
	// # Warum ein Spielstein sie mitbringt
	//
	// **Weil er aus keinem Deck kommt.** CR 111.1 nennt ihn eine Marke für
	// ein Permanent, das **keine Karte** vertritt; seine Eigenschaften
	// nennt der Effekt, der ihn erzeugt (CR 701.7a). Ein Verweis auf einen
	// Abzug ginge ins Leere.
	//
	// **Dieselben Feldnamen wie in einem MSE-Paket** (`name`, `super type`,
	// `power`, …), damit [EigenschaftenAus] sie ohne zweiten Weg liest —
	// eine zweite Deutung derselben Felder liefe beim ersten neuen Feld
	// auseinander (§2).
	//
	// **Leer bei einer Karte**: dort stehen sie im eingefrorenen Abzug.
	Felder map[string]string

	// Seite ist, welche Seite einer mehrseitigen Karte oben liegt —
	// CR 701.27a (M12).
	//
	// **Null ist die Vorderseite** (CR 712.8a: außerhalb von Schlachtfeld
	// und Stapel gelten ohnehin nur deren Merkmale), und das ist der
	// Normalfall. Eine einseitige Karte kennt nur sie.
	Seite int

	// Seit ist die Nummer des Zuges, seit dem dieser Beherrscher das Objekt
	// hat — CR 302.6 (Fassung 2026-08-19), Werkstück M6
	// (#1608).
	//
	// # Warum eine Zugnummer und kein Zeitstempel
	//
	// **Weil CR 508.1a nach dem Zug fragt** und nicht nach einem Moment:
	// „seit dem Beginn des laufenden Zuges unter derselben Kontrolle". Eine
	// Kreatur, die in Zug 5 gekommen ist, greift in Zug 5 nicht an und in
	// Zug 7 schon.
	//
	// **Null heißt: von Anfang an da.** Eine Partie ohne Zugbegriff (G1 bis
	// G8) hat keine Zugnummern, und ein Aufbau legt die Startobjekte hin,
	// bevor der erste Zug beginnt — beides ist keine Einsatzverzögerung.
	Seit int

	// Ziele sind die beim Ankündigen genannten Ziele — CR 115.1 (Fassung
	// 2026-08-19), Werkstück M4
	// (#1606).
	//
	// # Warum sie am Objekt hängen und nicht am Stapel
	//
	// **Weil ein Zauber sie mitnimmt.** CR 115.1 sagt, dass sie beim
	// Hinaufstellen festgelegt werden und sich danach nicht mehr ändern;
	// CR 608.2b prüft sie beim Verrechnen noch einmal. Eine Liste neben dem
	// Stapel liefe beim ersten Zonenwechsel auseinander.
	//
	// **Ein Ziel ist ein Objekt oder ein Sitz** (CR 115.1), und beide tragen
	// in dieser Partie eine Kennung. Jedes trägt zusätzlich die **Zone, in
	// der es beim Ansagen lag** — daran entscheidet CR 608.2b beim
	// Verrechnen, ob es noch zulässig ist.
	//
	// **Leer heißt: keine.** Die meisten Zauber haben keines, und `nil` ist
	// hier dasselbe wie eine leere Liste.
	Ziele []Ziel

	// HaengtAn ist die Kennung des Objekts, an dem dieses hängt.
	//
	// **Eine Beziehung und keine Liste**: eine Aura hängt an **einem** Ding,
	// und die Gegenrichtung ist eine Frage an die Partie und kein zweites
	// Feld, das auseinanderlaufen kann.
	HaengtAn string

	// Haufen ist die Anordnung auf dem Tisch — leer heißt: kein Haufen.
	Haufen string

	// Wirkungen sind die benannten Änderungen, die dieses Objekt beim
	// Verrechnen bewirkt — CR 608.2c (M8,
	// #1610).
	//
	// # Warum sie am Objekt stehen und nicht am Skript
	//
	// **Weil ein Skript ausgewertet werden muss und die Domäne das nicht
	// kann** (§2): MSE-Script liegt in `internal/msescript`, und `domain`
	// importiert nichts. Ausgewertet wird beim **Ankündigen**, in
	// `internal/service`; was dabei herauskommt, ist diese Liste.
	//
	// **Und das ist eine benannte Abweichung** (§15, Punkt 4): CR 608.2c
	// rechnet beim **Verrechnen**. Ein Skript, das zählt, was gerade auf
	// dem Tisch liegt, zählt hier den Stand beim Ankündigen. Für alles
	// Feste — „ziehe zwei Karten" — ist es dasselbe. Sie steht als Grenze
	// in der Regelmatrix, und der Weg da heraus hat ein Issue
	// (#1625).
	Wirkungen []Spielwirkung

	// Todesberuehrt sagt, ob seit der letzten zustandsbasierten Prüfung
	// Schaden aus einer Quelle mit Todesberührung angekommen ist —
	// CR 702.2b mit CR 704.5h (M14,
	// #1616).
	//
	// **Eine Marke am Objekt und keine Zahl**: CR 704.5h fragt nicht, wie
	// viel — jeder Schaden aus einer solchen Quelle genügt. Sie geht mit
	// dem markierten Schaden weg (CR 514.2) und mit dem Zonenwechsel
	// (CR 400.7).
	Todesberuehrt bool

	// Schaden ist der markierte Schaden — CR 120.3 (Fassung 2026-08-19),
	// Werkstück G9 von Phase 5.5 (#1592).
	//
	// **Am Objekt und nicht am Spieler**: Schaden an einem Spieler ändert
	// dessen Leben und bleibt nicht liegen; Schaden an einem Permanent wird
	// **markiert** und bleibt bis zum Aufräumschritt (CR 514.2).
	//
	// **Und er ist kein Zähler.** [Spielobjekt.Zaehler] führt Marken, die auf
	// dem Objekt liegen; markierter Schaden ist keine Marke, und CR 122
	// führt ihn auch nicht als eine.
	Schaden int

	// Zaehler sind die Marken darauf, je Art.
	//
	// **`Zaehler` und nicht „Marke"** — der Grund steht im Kopf von
	// `partie.go`: das Wort ist hier zweimal vergeben.
	Zaehler map[string]int
}

// Gleich sagt, ob zwei Werte dasselbe Objekt meinen.
//
// **Über die Kennung und über nichts sonst.** Der Grund steht im Kopf dieser
// Datei, und er ist die vierte Abnahme des Werkstücks.
func (o Spielobjekt) Gleich(anderes Spielobjekt) bool {
	return o.Kennung == anderes.Kennung
}
