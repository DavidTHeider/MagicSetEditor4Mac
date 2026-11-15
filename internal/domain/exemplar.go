package domain

import "strings"

// **Die Sammlung** — Werkstück 2 von Phase 3.4
// (#1165).
//
// # Warum das hier steht und nicht an der Karte
//
// Eine Karte in dieser Anwendung ist ein **Entwurf**; sie existiert nicht als
// Papier. Ein Exemplar ist das Gegenteil: ein bestimmter Druck, in einer
// Sprache, mit einem Zustand, den jemand besitzt. Die beiden zusammenzulegen
// hieße, persönliche Bestandsdaten in eine Datei zu schreiben, die man
// weitergibt
// ([0052](../../docs/entscheidungen/0052-die-sammlung-ist-nicht-das-set.md)).
//
// # Warum `new` und nicht `ported`
//
// **MSE2 kennt nichts dergleichen**, und §15 verlangt, dass gesagt wird, wo
// gesucht wurde. Die Suche steht im Vermerk: das ganze Datenmodell (21
// Kopfdateien), die **sieben** Reiter aus `src/gui/set/window.cpp:148–154`
// (Fassung `a7336d6` vom 2020-06-14), und `src/` nach `collection`,
// `inventory`, `owned`, `price`, `wishlist`, `quantity`, `condition`, `foil`
// (Fassung `f9d9356`). Gesucht am 2026-08-27.

// Zustand ist der Erhaltungszustand eines Exemplars.
//
// **Die Aufzählung ist die von Cardmarket**, und sie ist damit ein **fremder**
// Begriff — deshalb englisch (§9,
// [0015](../../docs/entscheidungen/0015-bezeichner-folgen-dem-vokabular.md)).
// Eine eigene Skala erfände eine Sprache, die im Handel niemand spricht; wer
// eine Sammlung aus einem Marktplatz einliest, müsste sie übersetzen.
type Zustand string

// Die sieben Zustände, von unbenutzt bis abgegriffen.
const (
	ZustandMint        Zustand = "mint"
	ZustandNearMint    Zustand = "nearMint"
	ZustandExcellent   Zustand = "excellent"
	ZustandGood        Zustand = "good"
	ZustandLightPlayed Zustand = "lightPlayed"
	ZustandPlayed      Zustand = "played"
	ZustandPoor        Zustand = "poor"
)

// alleZustaende ist die Aufzählung in ihrer Reihenfolge — beste zuerst.
//
// **Die Reihenfolge ist fachlich**, nicht alphabetisch: eine Auswahlliste, die
// `excellent` vor `mint` stellte, ließe niemanden das Naheliegende zuerst
// finden.
var alleZustaende = []Zustand{
	ZustandMint, ZustandNearMint, ZustandExcellent,
	ZustandGood, ZustandLightPlayed, ZustandPlayed, ZustandPoor,
}

// Zustaende gibt die Zustände als Kopie.
//
// **Als Kopie**, wie [Fehlerschluessel] — eine Aufzählung, die ein Aufrufer
// ändern kann, ist keine.
func Zustaende() []Zustand {
	kopie := make([]Zustand, len(alleZustaende))
	copy(kopie, alleZustaende)

	return kopie
}

// IstZustand sagt, ob eine Angabe zur Aufzählung gehört.
func IstZustand(zustand Zustand) bool {
	for _, bekannt := range alleZustaende {
		if bekannt == zustand {
			return true
		}
	}

	return false
}

// Vorgabesprache ist die Sprache eines Drucks ohne weitere Angabe.
//
// **`en` und nicht die Sprache der Oberfläche.** Scryfall führt einen Druck
// unter `lang: "en"`, wenn nichts anderes dabeisteht; wer hier `de` vorgäbe,
// hätte eine Sammlung, die zu ihrer Quelle nicht passt.
const Vorgabesprache = "en"

// Exemplar ist ein Stück Papier im Besitz der nutzenden Person.
type Exemplar struct {
	// ID ist die Kennung im Bestand — 0 für eines, das es dort noch nicht gibt.
	ID int64

	// Fassung steigt mit jedem Schreiben.
	//
	// **Sie ist die Bedingung einer Änderung** (§7, Regel 2): wer ändert,
	// schickt die Fassung mit, die er gelesen hat. Stimmt sie nicht mehr, hat
	// jemand anderes geschrieben — und das gibt einen Konflikt statt eines
	// stillen Überschreibens. Angestrichen vom Zweitblick über Phase 3.4.
	Fassung int64

	// Printing ist die Scryfall-Kennung des **Drucks** — eine UUID.
	//
	// **Ein Druck und keine Karte.** „Lightning Bolt" gibt es dutzendfach; was
	// jemand besitzt, ist ein bestimmter davon. Der Name ist englisch, weil er
	// einer von Scryfall ist (§9).
	Printing string

	// DruckID ist die lokale Zeile des Drucks — Werkstück A2 von Phase 5.3
	// (#1441, 0086). **Neben [Exemplar.Printing], nicht an seiner Stelle**: der
	// Alias bleibt, was jede Fläche kennt; die Zeile ist, was ein Druck ohne
	// Scryfall hat. Beides leer ist kein Exemplar.
	DruckID int64

	// Menge ist die Stückzahl — mindestens eine.
	//
	// **Null ist kein Exemplar**, sondern seine Abwesenheit: wer nichts mehr
	// besitzt, löscht die Zeile, statt eine mit Menge 0 zu führen.
	Menge int

	Zustand Zustand

	// Sprache ist der Sprachcode des Drucks, wie Scryfall ihn führt.
	Sprache string

	// Finish ist die Ausführung des Drucks — **leer heißt `nonfoil`**.
	//
	// **Drei und nicht zwei** (T2,
	// #1467):
	// hier stand `Foil bool`, und der konnte `etched` nicht sagen. Der
	// Preisstand kennt es seit Phase 5.3, und ein Bestand, der eine Karte
	// als „foil" führt, weil sie nicht „nonfoil" ist, führt sie falsch.
	//
	// **Der Nullwert ist die alte Vorgabe**: `false` hieß nonfoil. Wer
	// daraus einen Fehler machte, erklärte jeden Bestand von gestern für
	// ungültig — lies ihn über [Exemplar.Ausfuehrung].
	Finish Finish

	// Erwerbsdatum ist der Tag des Erwerbs — leer, wenn niemand ihn weiß.
	//
	// **Als Zeichenkette in `JJJJ-MM-TT`** und nicht als Zeitpunkt: was hier
	// steht, ist ein **Kalendertag** ohne Uhrzeit und ohne Zone. Ein
	// `time.Time` daraus trüge eine Mitternacht, die niemand gemeint hat, und
	// eine Zone, die beim nächsten Lesen woanders liegt.
	Erwerbsdatum string

	// Erwerbspreis ist der bezahlte Preis in der **kleinsten Einheit** seiner
	// Währung — LOC-03.AK-02
	// (#1689).
	//
	// **Ganzzahlig und nicht als Gleitkommazahl.** 4,17 € sind 417; ein
	// `float64` daraus ergibt beim Summieren einer Sammlung Cent-Abweichungen,
	// die niemand erklären kann. Null heißt „kein Preis erfasst".
	//
	// **Wie viele Stellen eine Währung hat, sagt [Waehrungsstellen]** — und
	// nur sie.
	Erwerbspreis int64

	// Erwerbswaehrung ist der ISO-4217-Code zum Erwerbspreis.
	//
	// **Ohne sie ist ein Preis keine Auskunft**: 417 in welcher Währung?
	Erwerbswaehrung string

	Notiz string

	// Name, Setcode und Nummer sind der **lesbare Bezug** des Drucks —
	// #1200.
	//
	// **Der letzte gute Stand und nicht die Identität.** Die ist [Printing];
	// diese drei sagen, was dort steht, wenn kein Netz da ist (§16, Regel 5).
	// Dieselbe Entscheidung wie bei [Kartenverweis].
	//
	// **Sie dürfen leer sein.** Wer über die Kommandozeile eine Kennung
	// einträgt, hat keinen Namen zur Hand; die Zeile steht dann mit ihrer
	// Kennung da. Ein Zwang zum Namen hieße, jede Fläche ans Netz zu binden.
	Name    string
	Setcode string
	Nummer  string

	// Lagerort ist, wo das Stück Papier liegt — Freitext
	// ([0057](../../docs/entscheidungen/0057-wie-eine-sammlung-geordnet-wird.md)).
	//
	// **„Kiste 4", „Ordner Rot, Seite 12", „bei Anna".** Jede Aufzählung, die
	// wir vorgeben, ist bei jemandem falsch — und Ordner und Box **sind**
	// Lagerorte, keine eigenen Begriffe.
	Lagerort string

	// Wunsch sagt, dass diese Zeile **kein Besitz** ist (0057).
	//
	// Daraus folgen drei Dinge, und sie sind nicht verhandelbar:
	//
	//  1. **Das Portfolio zählt sie nicht.** Was man nicht hat, ist nichts wert.
	//  2. **[Exemplar.Besitzt] sagt Nein.**
	//  3. **Die Eindeutigkeit nimmt sie auf** — man kann denselben Druck
	//     zugleich besitzen und sich wünschen, und genau das tut, wer eine
	//     gespielte Fassung hat und eine schöne will.
	Wunsch bool

	// Wunschrang sagt, wie dringend dieser Wunsch ist — **leer heißt
	// `normal`**, und gelesen wird er über [Exemplar.Dringlichkeit].
	Wunschrang Wunschrang

	// Wunschart sagt, ob genau dieser Druck gemeint ist oder die Karte —
	// **leer heißt `genau`**, gelesen über [Exemplar.Druckbindung].
	Wunschart Wunschart

	// Herkunft sagt, woher dieses Exemplar kam — Werkstück 12 von Phase 4.12
	// (#1351).
	//
	// **Die leere heißt: hier entstanden.** Sie reist mit der Zeile und wird
	// nicht einzeln nachgefragt; der Grund steht bei `sammlungsquelle` in
	// `internal/storage/sammlung.go`.
	Herkunft Herkunft
}

// Besitzt sagt, ob diese Zeile ein Stück Papier meint, das jemand hat.
//
// **Eine Methode und kein `!e.Wunsch` an dreißig Stellen.** Die Regel „ein
// Wunsch ist kein Besitz" gehört an eine Stelle; wer sie verteilt, hat sie
// irgendwann an einer davon anders (0057).
func (e Exemplar) Besitzt() bool { return !e.Wunsch }

// Bezeichnung gibt den lesbaren Bezug — oder die Kennung.
//
// **Ein Fallback und keine leere Zeile**: ohne Namen ist die Kennung das
// Einzige, was diese Zeile über ihren Druck weiß, und sie zu verschweigen
// machte das Exemplar unidentifizierbar.
func (e Exemplar) Bezeichnung() string {
	return Kartenbezeichnung(e.Printing, e.Name, e.Setcode, e.Nummer)
}

// Kartenbezeichnung baut, wie eine Karte in einer Liste heißt.
//
// **Eine Stelle und nicht zwei.** Dieselbe Regel gilt für ein Exemplar der
// Sammlung und für eine Zeile im Journal, deren Namen erst beim Lesen
// nachgeschlagen wird (T6, #1471).
// Zwei Fassungen liefen auseinander, und die, die niemand ändert, wäre die im
// Journal — dort fiele es am spätesten auf.
//
// **Der Rückfall ist die Kennung des Drucks.** Sie sagt niemandem etwas, aber
// sie ist eine Auskunft; ein leerer Name wäre keine.
func Kartenbezeichnung(printing, name, setcode, nummer string) string {
	if name == "" {
		return printing
	}

	if setcode == "" {
		return name
	}

	if nummer == "" {
		return name + " · " + strings.ToUpper(setcode)
	}

	return name + " · " + strings.ToUpper(setcode) + " #" + nummer
}

// Normalisieren setzt die Vorgaben und glättet, was aus einer Eingabe kommt.
//
// **Getrennt von [Exemplar.Validate].** Eine Prüfung, die nebenbei ändert,
// lässt den Aufrufer im Unklaren, was er gespeichert hat; und eine Vorgabe, die
// erst beim Prüfen entsteht, kommt für die Oberfläche zu spät.
func (e *Exemplar) Normalisieren() {
	e.Printing = strings.TrimSpace(e.Printing)
	e.Sprache = strings.TrimSpace(e.Sprache)
	e.Erwerbswaehrung = strings.ToUpper(strings.TrimSpace(e.Erwerbswaehrung))
	e.Erwerbsdatum = strings.TrimSpace(e.Erwerbsdatum)
	e.Notiz = strings.TrimSpace(e.Notiz)

	// **Der lesbare Bezug wird geglättet, nicht erfunden** (#1200): was aus
	// einer Eingabe kommt, trägt Ränder; was fehlt, bleibt leer.
	e.Lagerort = strings.TrimSpace(e.Lagerort)
	e.Name = strings.TrimSpace(e.Name)
	e.Setcode = strings.ToLower(strings.TrimSpace(e.Setcode))
	e.Nummer = strings.TrimSpace(e.Nummer)

	if e.Zustand == "" {
		e.Zustand = ZustandNearMint
	}

	if e.Sprache == "" {
		e.Sprache = Vorgabesprache
	}

	// **Was an einem Besitz nichts bedeutet, wird nicht aufbewahrt** (T4).
	// Sonst stünde die Angabe im Bestand und wartete darauf, dass jemand
	// sie liest.
	if e.Besitzt() {
		e.Wunschrang, e.Wunschart = "", ""
	}
}

// Validate gibt einen Befund je Feld, das nicht stimmt.
//
// **Eine Liste und kein einzelner Fehler** (§6): wer drei Felder falsch
// ausgefüllt hat, soll drei Meldungen sehen und nicht dreimal absenden.
func (e Exemplar) Validate() []*FieldError {
	var befunde []*FieldError

	if strings.TrimSpace(e.Printing) == "" && e.DruckID == 0 {
		befunde = append(befunde, NewFieldError("printing", SchluesselPflichtfeld,
			"Ein Exemplar gehört zu einem Druck."))
	}

	if e.Menge < 1 {
		befunde = append(befunde, NewFieldError("menge", SchluesselValidierung,
			"Die Menge ist %d — mindestens eines muss es sein.", e.Menge))
	}

	if e.Zustand != "" && !IstZustand(e.Zustand) {
		befunde = append(befunde, NewFieldError("zustand", SchluesselValidierung,
			"Den Zustand %q gibt es nicht.", e.Zustand))
	}

	if e.Finish != "" && !IstFinish(e.Finish) {
		befunde = append(befunde, NewFieldError("finish", SchluesselValidierung,
			"Die Ausführung %q gibt es nicht.", e.Finish))
	}

	if strings.TrimSpace(e.Sprache) == "" {
		befunde = append(befunde, NewFieldError("sprache", SchluesselPflichtfeld,
			"Ein Druck hat eine Sprache."))
	}

	if e.Wunschrang != "" && !IstWunschrang(e.Wunschrang) {
		befunde = append(befunde, NewFieldError("wunschrang", SchluesselValidierung,
			"Den Rang %q gibt es nicht.", e.Wunschrang))
	}

	if e.Wunschart != "" && !IstWunschart(e.Wunschart) {
		befunde = append(befunde, NewFieldError("wunschart", SchluesselValidierung,
			"Die Wunschart %q gibt es nicht.", e.Wunschart))
	}

	return append(append(befunde, e.Datumsbefunde()...), e.Preisbefunde()...)
}

// Vorlaeufig sagt, ob das Exemplar einen Druck ohne Alias nennt (0086).
func (e Exemplar) Vorlaeufig() bool { return druckVorlaeufig(e.DruckID, e.Printing) }

// Druckschluessel sagt, welchen Druck das Exemplar meint — für Gleichheit.
func (e Exemplar) Druckschluessel() string { return Druckschluessel(e.DruckID, e.Printing) }

// Ausfuehrung gibt die Ausführung — **leer heißt `nonfoil`**.
//
// **Eine Methode und kein `if e.Finish == ""` an dreißig Stellen.** Die Regel
// „ohne Angabe ist es nonfoil" gehört an eine Stelle; wer sie verteilt, hat
// sie irgendwann an einer davon anders — derselbe Satz wie bei
// [Exemplar.Besitzt].
func (e Exemplar) Ausfuehrung() Finish {
	if e.Finish == "" {
		return FinishNonfoil
	}

	return e.Finish
}

// FinishAus macht aus der alten Wahrheit eine Ausführung.
//
// **Und nie `etched`.** Aus `true` geht nicht hervor, welche der beiden es
// war; wer daraus riete, schriebe eine Angabe in den Bestand, die niemand
// gemacht hat (§8). Wer `etched` will, sagt es.
func FinishAus(foil bool) Finish {
	if foil {
		return FinishFoil
	}

	return FinishNonfoil
}
