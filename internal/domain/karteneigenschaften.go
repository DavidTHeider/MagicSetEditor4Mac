package domain

// **Was eine Karte ist** — Werkstück M2 der Regelmaschine
// (#1604).
//
// # Die Lücke, die das schließt
//
// **CR 109.3** (Fassung 2026-08-19) zählt auf, was ein Objekt hat: Name,
// Manakosten, Farbe, Kartentyp, Untertyp, Obertyp, Regeltext, Fähigkeiten,
// Stärke, Widerstandskraft, Loyalität, Verteidigung. Bis zum 2026-09-17 hatte
// [Spielobjekt] **keine davon** — es hatte Identität und Zone, und
// `docs/REGELABDECKUNG.md` sagte es so.
//
// # Woher die Werte kommen
//
// **Aus dem eingefrorenen Abzug und nicht aus dem Bestand.**
// [Abzugsseite.Felder] trägt seit G6
// (#1463)
// die Feldwerte, wie sie beim Anfang standen; `service.Partieaufbau`
// schreibt sie aus `karte.Values` hinein. Wer sie hier aus dem Bestand
// nachschlüge, bekäme eine Karte, die jemand seither geändert hat — und die
// Partie liefe mit einer anderen Karte weiter, als sie angefangen hat.
//
// **Und gedeutet wird beim Lesen.** Kein Bestandsschema ändert sich, keine
// Migration entsteht: eine Partie von gestern trägt dieselben Felder und
// bekommt heute Eigenschaften.
//
// # Die Feldnamen gehören dem Spielpaket
//
// **`docs/FORMATE.md` (Zeile 1832 ff.)** zählt sie für das Spiel `magic` auf,
// erhoben am `M15-Magic-Pack`. Ein anderes Spiel nennt seine Felder, wie es
// will; deshalb steht der Spielname im Ruf, und die Zuordnung liegt an
// **einer** Stelle je Spiel — dieselbe Regel wie in `internal/mtgjson` (§2).

import "strings"

// SpielMagic ist die Kennung des Spielpakets, dessen Felder hier gedeutet
// werden.
const SpielMagic = "magic"

const (
	// MarkePlusEins ist die +1/+1-Marke — CR 122.1a (Fassung 2026-08-19).
	//
	// **Der Schlüssel ist der Name der Marke**, so wie [Spielobjekt.Zaehler]
	// ihn führt: eine Marke hat einen Namen, und `+1/+1` ist ihrer.
	MarkePlusEins = "+1/+1"

	// MarkeMinusEins ist die −1/−1-Marke — CR 122.1a, zweiter Satz.
	MarkeMinusEins = "-1/-1"
)

// Karteneigenschaften sind die Eigenschaften einer Kartenseite — CR 109.3.
//
// **Nicht alle zwölf.** Regeltext und Fähigkeiten stehen nicht hier: ein Text
// ist eine Zeichenkette, solange ihn niemand ausführt, und das Ausführen ist
// `M8` (#1610).
// Die Farbe steht nicht als Feld, sondern folgt den Manakosten (CR 202.2) —
// ein zweites Feld daneben liefe auseinander.
type Karteneigenschaften struct {
	Name string

	Manakosten Manakosten
	Typzeile   Typzeile

	// Staerke und Widerstandskraft — CR 208.1.
	//
	// **Sie stehen auch an einer Nichtkreatur**, wenn die Karte sie gedruckt
	// trägt (ein Fahrzeug): CR 208.3 sagt, dass sie dann außerhalb des
	// Schlachtfelds gelten. Was auf dem Tisch gilt, entscheidet
	// [Partie.Werte].
	Staerke          Kartenwert
	Widerstandskraft Kartenwert

	// Loyalitaet ist die gedruckte Zahl eines Planeswalkers — CR 306.5b.
	//
	// **Sie ist eine Startzahl und kein Zustand**: CR 306.5b lässt das
	// Permanent mit so vielen Loyalitätsmarken ins Spiel kommen. Was davon
	// heute übrig ist, steht in [Spielobjekt.Zaehler].
	Loyalitaet Kartenwert

	// Befunde sind die Wörter der Typzeile, die CR 205.2a und CR 205.4a
	// nicht kennen — §8.
	//
	// **Sie werden mitgeführt und nicht weggeworfen.** Ein Wert, den
	// niemand liest, ist keine Zusage (§10): die dritte Abnahme von M2
	// lautet „ein Kartentyp, den die CR nicht kennt, ist ein **Befund** und
	// kein stiller Ausfall", und ein Befund, der in einer Zuweisung an `_`
	// endet, ist ein stiller Ausfall mit einem Test daneben.
	Befunde []Befund

	// Verteidigung ist die Zahl einer Kampfkarte — CR 310.1, CR 210.1.
	//
	// **Dieselbe Bauart wie die Loyalität**: sie sagt, mit wie vielen
	// Verteidigungsmarken das Permanent kommt (CR 310.4b).
	//
	// **Und ihr Feldname ist eine Annahme** (§15): `docs/FORMATE.md` führt
	// für das Spiel `magic` **vierzehn** Felder, und `defense` ist keines
	// davon — erhoben am 2026-08-15 am `M15-Magic-Pack`. Die Kampfkarte kam
	// 2023 dazu, das Paket ist älter. Wer ein Paket hat, das sie kennt,
	// bekommt sie; wer keines hat, bekommt eine leere Zahl. **Erfunden wird
	// sie nicht**: sie steht hier, weil CR 310 sie kennt, und sie bleibt
	// leer, solange kein Feld sie füllt.
	Verteidigung Kartenwert

	// Spielskript ist der MSE-Script-Text, den diese Karte beim Verrechnen
	// ausführt — CR 608.2c (M8,
	// #1610).
	//
	// **Der Text und nicht sein Ergebnis.** Die Domäne wertet ihn nicht aus
	// — MSE-Script liegt in `internal/msescript`, und `domain` importiert
	// nichts aus diesem Projekt (§2). Was dabei herauskommt, sind
	// [Spielwirkung]en, und die legt `internal/service` beim Ankündigen an
	// das Objekt auf dem Stapel.
	//
	// **Leer ist der Normalfall**: keine bezogene Vorlage kennt dieses Feld.
	Spielskript string
}

// feldnamen ordnet jeder Eigenschaft ihre MSE-Feldnamen zu — der erste
// gefundene gilt.
//
// **Die Reihenfolge ist die aus `docs/FORMATE.md`**: der eigentliche Name
// zuerst, die Alternativnamen dahinter. Sie sind keine Erfindung — MSE führt
// sie im `card field:`-Block seines Spielpakets.
var feldnamen = map[string][]string{
	"name":       {"name"},
	"kosten":     {"casting cost", "mana cost", "manacost", "cost", "cc"},
	"ober":       {"super type", "supertype"},
	"unter":      {"sub type", "subtype"},
	"zeile":      {"type", "typeline", "type line"},
	"staerke":    {"power"},
	"widerstand": {"toughness"},
	"geteilt":    {"pt", "p/t", "power/toughness"},
	"loyalitaet": {"loyalty"},
	"verteidig":  {"defense"},

	// **Das Spielskript ist neu** (§15, Marke `new`): MSE2 hat keinen
	// Spielkern, und deshalb gibt es in keinem Paket ein Feld dafür —
	// gesucht am 2026-09-18 in `docs/FORMATE.md`, das für `magic`
	// **vierzehn** Felder führt, und keines davon ist es (M8,
	// #1610).
	//
	// **Vier Namen, damit ein Paket wählen kann**, wie es sein eigenes Feld
	// nennt. Wer keines hat, bekommt keine Fähigkeit — und das ist der
	// Normalfall für jedes bezogene Paket.
	"skript": {"game script", "spielskript", "rules script", "game_script"},
}

// EigenschaftenAus deutet die Feldwerte einer Kartenseite.
//
// **Der Name gilt in jedem Spiel**, alles andere nur in `magic`: ein
// Kartentyp nach CR 205.2a ist eine Aussage über Magic, und sie auf ein
// fremdes Spielpaket anzuwenden hieße, dessen Felder zu erraten (§15).
func EigenschaftenAus(spiel string, felder map[string]string) Karteneigenschaften {
	eigen := Karteneigenschaften{Name: feldwert(felder, "name")}

	if spiel != SpielMagic {
		return eigen
	}

	eigen.Manakosten = ManakostenAus(feldwert(felder, "kosten"))

	ober, unter := feldwert(felder, "ober"), feldwert(felder, "unter")
	if strings.TrimSpace(ober) == "" {
		ober = feldwert(felder, "zeile")
	}

	eigen.Typzeile, eigen.Befunde = TypzeileAus(ober, unter)

	eigen.Staerke = KartenwertAus(feldwert(felder, "staerke"))
	eigen.Widerstandskraft = KartenwertAus(feldwert(felder, "widerstand"))

	// **Das geteilte Feld füllt, was leer geblieben ist.** Manche Vorlagen
	// führen `pt` statt der beiden Einzelfelder; wer es überginge, bekäme
	// eine Kreatur ohne Werte.
	if !eigen.Staerke.Steht && !eigen.Widerstandskraft.Steht {
		eigen.Staerke, eigen.Widerstandskraft = geteilteWerte(feldwert(felder, "geteilt"))
	}

	eigen.Loyalitaet = KartenwertAus(feldwert(felder, "loyalitaet"))
	eigen.Verteidigung = KartenwertAus(feldwert(felder, "verteidig"))

	// **Der Text und nicht sein Ergebnis** — was er bewirkt, entscheidet
	// `internal/service`; die Domäne kennt MSE-Script nicht (§2).
	eigen.Spielskript = feldwert(felder, "skript")

	return eigen
}

// feldwert gibt den ersten gefüllten der bekannten Namen.
func feldwert(felder map[string]string, name string) string {
	for _, kandidat := range feldnamen[name] {
		if wert := felder[kandidat]; strings.TrimSpace(wert) != "" {
			return wert
		}
	}

	return ""
}

// geteilteWerte zerlegt `3/4` in Stärke und Widerstandskraft.
//
// **Ohne Strich ist es keines von beidem**: ein `pt`, das nur `3` trägt, sagt
// nicht, ob die Drei die Stärke oder die Widerstandskraft ist — und eine
// Antwort zu erraten wäre schlimmer als keine (§10).
func geteilteWerte(roh string) (staerke, widerstand Kartenwert) {
	vor, nach, gefunden := strings.Cut(roh, "/")
	if !gefunden {
		return Kartenwert{}, Kartenwert{}
	}

	return KartenwertAus(vor), KartenwertAus(nach)
}
