package domain

// **Die Partie** — Werkstück G1 von Phase 5.5
// (#1458).
//
// # Die Begriffe sind deutsch, und das ist eine Entscheidung
//
// **§9 sagt: das Vokabular entscheidet, nicht die Sichtbarkeit.** Eine Partie
// ist kein fremder Begriff aus MSE, MTG JSON, Go oder Wails — sie ist ein
// eigener dieses Projekts, so wie `Turnier`, `Paarung` und `Ergebnis` daneben.
// Das Issue verlangt es ausdrücklich: „Fachbegriffe an deutsche
// Repositorykonvention anpassen."
//
// **Zwei Namen weichen dabei vom offiziellen deutschen Magic-Wortschatz ab**,
// und beide mit Grund:
//
//   - **`Zaehler` statt „Marke"**: „Marke" ist in diesem Repositorium zweimal
//     vergeben — als Kartenmarke (`MarkeSetzen`) und als Zugang eines Dienstes
//     (`rest.NeueMarke`). Ein dritter Sinn desselben Wortes wäre nicht kürzer,
//     sondern unlesbar.
//   - **`Spielbefehl` statt „Befehl"**: `app.Befehl` ist der Eintrag der
//     Befehlspalette. Zwei `Befehl` in einem Projekt, die nichts miteinander
//     zu tun haben, sind eine Falle für jeden Leser.
//
// # Der Wirt ist offengelegt und nicht wegdefiniert
//
// **Ein beim Start benannter Sitz serialisiert die Befehle und hält den
// Vollzustand.** Er kann technisch die verdeckten Teile kennen — die Hand des
// Gegenübers, die Reihenfolge der Bibliothek —, und **genau das steht vor dem
// Beitritt da** (#1458: „diese Vertrauensgrenze wird vor Beitritt
// offengelegt").
//
// **Was daraus nicht folgt, ist ein Freibrief**: die Projektion je Rolle gilt
// auch für ihn, und der Vollzustand ist eine benannte Ausnahme mit einem
// eigenen Lauf — nicht der Normalfall, an dem sich niemand stört.

// Partie ist eine laufende oder pausierte Partie.
type Partie struct {
	ID      string
	Fassung int64

	// Geheimnis ist der Schlüssel der Sichtkennungen (#2480, Vermerk 0201).
	// **Er verlässt den Wirt nie**: keine Sicht und kein Ereignis trägt ihn,
	// und mit ihm ließe sich jede Sichtkennung auf ihre Kennung zurückrechnen.
	Geheimnis string

	// Sichtfolge zählt die Wechsel der Sichtkennungen — der Wert, den ein
	// Objekt bekommt, wenn es verborgen wird ([Sichtwechsel]).
	Sichtfolge int64

	// Spiel ist die Kennung des Spielpakets — `magic`, `space`, …
	Spiel string

	// Regelwerk sagt, **welche** Regeln in welcher Strenge gelten.
	Regelwerk Regelwerk

	// Profil ist das **eingefrorene** Regelwerk (G4,
	// #1461).
	//
	// # Warum es mitgeht und nicht nachgeschlagen wird
	//
	// **Damit dieselbe Partie nach einem Neustart dieselben Entscheidungen
	// gibt** — auch dann, wenn draußen inzwischen eine neue Fassung gilt.
	// Das Issue verlangt es wörtlich: „Gleicher Ruleset-/Format-Snapshot
	// liefert nach Neustart dieselben Entscheidungen, auch wenn externe
	// Legality-Daten inzwischen aktualisiert wurden."
	//
	// **Dieselbe Überlegung wie beim Deckabzug** (G1): eingefroren wird,
	// was die Partie trägt, damit eine spätere Änderung sie nicht anfasst.
	//
	// **[Partie.Regelwerk] bleibt daneben stehen** und ist der kurze Griff —
	// Kennung, Fassung, Strenge. Die beiden müssen übereinstimmen, und
	// [Partie.Pruefe] hält es fest: ein Griff, der auf ein anderes Profil
	// zeigt als das eingefrorene, wäre eine Auskunft, die zweimal
	// verschieden ausfällt.
	Profil Regelprofil

	// Wirt ist die Sitzkennung dessen, der serialisiert.
	//
	// **Eine Sitzkennung und kein Nutzer.** Ein Nutzer kann die Anwendung
	// wechseln, ein Sitz nicht: er gehört zu dieser Partie.
	Wirt string

	Sitze     []Sitz
	Zuschauer []Zuschauer

	// Objekte ist alles, was in dieser Partie liegt — über alle Zonen.
	//
	// **Eine Liste und keine Zonen als Felder.** Ein Objekt wechselt die
	// Zone; wer es dafür zwischen zwei Listen umhängen müsste, hätte einen
	// Zustand, der zwischen den beiden Schritten falsch ist.
	Objekte []Spielobjekt

	// Legendenwahlen sind die offenen Fragen aus CR 704.5j (Fassung
	// 2026-08-19), Werkstück M12
	// (#1614).
	//
	// # Warum eine Aktion eine Frage hinterlässt
	//
	// **Weil CR 704.5j die einzige der fünfundzwanzig ist, die eine Wahl
	// verlangt**: der Beherrscher sagt, welche Legende bleibt. Die übrigen
	// vierundzwanzig geschehen; diese hier **wartet**.
	//
	// **An der Partie und nicht am Zugstand** — anders als [Zugstand.Abwurf]
	// trifft sie jeden Sitz und nicht den aktiven, und sie trifft auch eine
	// Partie, die nie einen Zug begonnen hat.
	Legendenwahlen []Legendenwahl

	// Ausgeloest sind die Fähigkeiten, die ausgelöst haben und auf den
	// Stapel warten — CR 603.2 (M9,
	// #1611).
	//
	// **Sie tun nichts, solange sie hier stehen.** CR 603.3 legt sie erst
	// auf den Stapel, wenn das nächste Mal jemand Priorität bekäme — und
	// zwischen Auslösen und Stapel liegt genau dieser Zustand.
	Ausgeloest []Ausloeser

	// Dauereffekte sind die Effekte, die eine Weile gelten — CR 611.1
	// (M10, #1612).
	//
	// **An der Partie und nicht am Objekt**: ein Effekt hat einen
	// Zeitstempel (CR 613.7), und der ordnet ihn gegen **alle** anderen —
	// eine Liste je Objekt hätte zehn Zählungen, die nichts voneinander
	// wissen.
	Dauereffekte []Dauereffekt

	// Ersetzungen sind die Effekte, die ein Ereignis durch ein anderes
	// ersetzen oder es verhindern — CR 614.1, CR 615.1 (M11,
	// #1613).
	//
	// **Getrennt von [Partie.Dauereffekte]**, und das ist der Unterschied,
	// um dessentwillen es CR 614 und CR 613 getrennt gibt: ein Dauereffekt
	// ändert, **was ein Objekt ist**; ein Ersetzungseffekt ändert, **was
	// geschieht**.
	Ersetzungen []Ersetzungseffekt

	// Ausloeserreihung ist der Sitz, der gerade die Reihenfolge seiner
	// wartenden Fähigkeiten nennen soll — CR 603.3b (M9, #1611).
	//
	// **Leer heißt: niemand wird gefragt**, und das ist der Normalfall. Die
	// Frage entsteht nur, wo **ein** Sitz **mehrere** Fähigkeiten zugleich
	// beherrscht; bei einer legt sie der Kern selbst.
	//
	// **Solange sie steht, bekommt niemand Priorität** — CR 603.3 stellt das
	// Legen ausdrücklich davor ([Partie.prioritaetFuer]).
	Ausloeserreihung string

	// Monarch und Initiative sind Bezeichnungen, die ein Spieler hat —
	// CR 725.1, CR 726.1 (M20,
	// #1621).
	//
	// **Leer heißt: niemand**, und das ist der Anfangszustand. CR 725.1
	// sagt es ausdrücklich: es gibt keinen Monarchen, bis ein Effekt einen
	// macht.
	Monarch    string
	Initiative string

	// Tageszeit ist, was die Partie selbst gerade ist — CR 731.1 (M20).
	//
	// **Leer heißt: weder noch**, und dann findet die Prüfung im
	// Enttappschritt gar nicht statt (CR 731.2c).
	Tageszeit Tageszeit

	// Stand sagt, ob gespielt wird.
	Stand Partiestand

	// Mischpflichten sind die verborgenen Zonen, die jemand angesehen und
	// seitdem nicht gemischt hat.
	//
	// **Im Zustand und nicht nur im Strom** — der Grund steht in
	// `mischpflicht.go`.
	Mischpflichten []Mischpflicht

	// Ausgang ist, wie die Partie ausgegangen ist — CR 104 (M17,
	// #1618).
	//
	// **Leer heißt: sie läuft noch.** `Stand` sagt, **ob** sie beendet ist;
	// hier steht, **warum** — und für das Turnier aus G7 sind das zwei
	// verschiedene Auskünfte.
	Ausgang Partieausgang

	// Zug ist, wo die Partie im Zug steht — CR 500.1 (G9, #1592).
	//
	// **Eine Null in [Zugstand.Nummer] heißt: es geht nicht zugweise zu.**
	// Die Partien aus G1 bis G8 sind Sandkästen ohne Zugbegriff, und sie
	// sollen weiterlaufen — ein Feld, das eine alte Partie plötzlich in den
	// Enttappen-Schritt stellte, änderte Verhalten ohne Zutun (§15, Punkt 4).
	Zug Zugstand

	// deuter liest die Kartenskripte — **beim Verrechnen** (CR 608.2c,
	// #1625).
	//
	// **Unexportiert, und das ist die Zusage**: er reist nicht in den
	// Bestand, und eine geladene Partie bekommt ihn neu. Der Grund steht bei
	// [Partie.MitDeuter].
	deuter Wirkungsdeuter

	// skriptbefund ist, was das zuletzt verrechnete Skript zu melden hatte.
	skriptbefund string
}

// Zuschauer sieht zu und hat keinen Sitz.
//
// **Eine Rolle ohne belegten Spielersitz** (#1458). Der Unterschied ist nicht
// höflich gemeint: wer keinen Sitz hat, hat keine Hand — und bekommt deshalb
// auch keine zu sehen.
type Zuschauer struct {
	Kennung string
	Name    string
}

// Partiestand ist, was mit dieser Partie gerade geschieht.

// SichtFuerProfil gibt die Sitzkennung, die dieses Profil sehen soll.
//
// **Das ist die zweite Zusage von Werkstück G8**
// (#1465):
// „Das Fenster arbeitet in einem Profil; wer wechselt, bekommt die Sicht
// **dieses** Sitzes."
//
// **Die Null bekommt den Wirt.** Ein Sitz ohne Profil trägt sie, und wer mit
// ihr suchte, bekäme den ersten ungebundenen — das wäre eine Sicht, die
// niemand gewählt hat.
//
// **Und ein Profil, das hier nicht sitzt, ebenso.** Wer in einem dritten
// Profil arbeitet, ist am Tisch ein Zuschauer und kein Sitz; die Wahl bleibt
// ihm, und sie geht über die Sichtwahl.
//
// # Warum die Rechnung hier steht und nicht in `app`
//
// **Weil sie eine Entscheidung über eine Partie ist** (§2). Ein `if` über
// einen Sitz in der Bindungsschicht wäre dieselbe Regel an einer Stelle, an
// der niemand sie sucht — und die drei kopflosen Flächen kämen nur dann
// daran vorbei, wenn jemand sie dort noch einmal schriebe.
func (p Partie) SichtFuerProfil(profilID int64) string {
	if profilID == 0 {
		return p.Wirt
	}

	for stelle := range p.Sitze {
		if p.Sitze[stelle].ProfilID == profilID {
			return p.Sitze[stelle].Kennung
		}
	}

	return p.Wirt
}

// SitzVon gibt den Sitz zu einer Kennung — `nil`, wenn es keinen gibt.
//
// **`nil` und kein leerer Sitz**: „kein Sitz" ist die Antwort, an der die
// Sichtbarkeit hängt, und ein leerer Sitz sähe aus wie einer.
func (p Partie) SitzVon(kennung string) *Sitz {
	for i := range p.Sitze {
		if p.Sitze[i].Kennung == kennung {
			return &p.Sitze[i]
		}
	}

	return nil
}

// IstWirt sagt, ob diese Kennung serialisieren darf.
//
// **Die leere Kennung ist keiner.** Ohne diese Zeile wäre jeder Ruf ohne
// Absender der Wirt — und das ist der Ruf, den ein Angreifer am leichtesten
// stellt.
func (p Partie) IstWirt(kennung string) bool {
	return kennung != "" && p.Wirt == kennung
}
