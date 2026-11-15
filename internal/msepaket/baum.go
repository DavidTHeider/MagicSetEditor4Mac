// Package msepaket liest und schreibt die MSE-Paketformate (AGENTS.md §2).
//
// Es kennt `internal/domain` und sonst nichts aus dem Projekt.
//
// # Woher das Format stammt
//
// **Aus den Datenpaketen, nicht aus dem MSE2-Quelltext**
// ([0001](../../docs/entscheidungen/0001-neubau-statt-portierung.md)). Jede
// Feststellung, auf der dieser Code steht, ist in `docs/FORMATE.md` mit Quelle
// belegt — welche Paketdatei, welche Stelle, welcher Sicherheitsgrad.
//
// Wer hier etwas ergänzt, ergänzt zuerst dort. Eine Erkenntnis ohne Quelle ist
// von einer abgeschriebenen nicht zu unterscheiden, und im Streitfall ist genau
// das die Frage.
package msepaket

import (
	"fmt"
	"strings"
)

// Stelle ist eine Fundstelle in einer Paketdatei.
//
// Zeile und Spalte zählen ab 1, wie ein Mensch sie liest — nicht ab 0, wie ein
// Feld sie ablegt. Wer eine Meldung bekommt, springt damit an die Stelle,
// statt sie um eins zu verschieben.
type Stelle struct {
	Datei  string
	Zeile  int
	Spalte int
}

// Knoten ist ein Eintrag im Datenbaum eines Pakets.
type Knoten struct {
	// Zeilenart sagt, ob dies ein Eintrag, eine Leerzeile oder ein Kommentar ist.
	Zeilenart  Zeilenart
	Schluessel string
	Wert       string
	Kinder     []*Knoten
	Stelle     Stelle

	// Roh ist die Zeile, wie sie in der Datei stand — samt Einrückung und
	// Abständen, ohne ihr Zeilenende.
	//
	// **Ohne sie ist die Rundlaufzusage nicht zu halten.** Der Zweitblick über
	// Phase 2 hat es belegt: ein Rohblock, der aus Schlüssel und Wert wieder
	// zusammengesetzt wird, verliert `key:  wert` gegenüber `key: wert`, verliert
	// gemischte Zeilenenden und den fehlenden letzten Umbruch. Er sieht dann aus
	// wie das Original und ist es nicht — und ein Byte-für-Byte-Vergleich, den
	// docs/entscheidungen/0002 zusagt, schlüge fehl.
	Roh string

	// Zeilenende ist die Form, in der die Zeile geendet hat — "\n", "\r\n"
	// oder ein einzelnes "\r" (FMT-06, wie MSE2s `read_utf8_line`).
	//
	// Sie wird gemerkt, nicht normalisiert. Das ist die halbe Rundlaufzusage
	// aus docs/entscheidungen/0002: was gelesen wurde, muss unverändert
	// zurückgeschrieben werden können. Eine Datei auf eine Form zu bringen wäre
	// eine Änderung, die niemand angeordnet hat — und in einem Byte-Vergleich
	// der Beweis, dass wir das Set angefasst haben.
	Zeilenende string
}

// Baum ist der Inhalt einer Paketdatei.
type Baum struct {
	Datei  string
	Wurzel []*Knoten

	// MitBOM sagt, ob die Datei mit einer Byte-Order-Mark begann.
	//
	// **Gemerkt, nicht normalisiert.** Der Zerteiler schnitt sie ab und vergaß
	// sie; ein Rundlauf hätte die Datei um drei Bytes kürzer zurückgegeben, als
	// sie hineinging — und ein Byte-Vergleich wäre der Beweis, dass wir das Set
	// angefasst haben.
	MitBOM bool
}

// Ersten gibt den ersten Knoten mit diesem Schlüssel, oder nil.
func (b *Baum) Ersten(schluessel string) *Knoten {
	return ersten(b.Wurzel, schluessel)
}

// Alle gibt alle Knoten mit diesem Schlüssel, in ihrer Reihenfolge.
//
// **Der wichtigere der beiden Zugriffe.** In echten Paketen wiederholt sich
// derselbe Schlüssel — `depends on:` steht mehrfach —, und wer nur `Ersten`
// benutzt, verliert den Rest, ohne es zu merken.
func (b *Baum) Alle(schluessel string) []*Knoten {
	return alle(b.Wurzel, schluessel)
}

// Skripttext gibt den Inhalt eines Blocks so, wie er als Skript dasteht.
//
// # Warum als Text und nicht als Baum
//
// Ein `init script:` ist ein **Programm**, kein `schlüssel: wert`-Baum
// ([0014](../../docs/entscheidungen/0014-eine-mse-datei-ist-nicht-notwendig-ein-baum.md)).
// Der Zerteiler hat es trotzdem in Knoten zerlegt, weil er jede Datei so liest;
// wer es auswerten will, braucht die Zeilen zurück.
//
// **Die Einrückung des Paketformats fällt weg.** Sie sagt dort, welcher Block
// offen ist; in MSE-Script bedeutet sie nichts.
//
// **Rekursiv**, denn ein Eintrag öffnet eine Ebene: `include file: …` steht auf
// der ersten, und was der Einbinder dafür einsetzt, hängt darunter.
func (k *Knoten) Skripttext() string {
	zeilen := k.Skriptzeilen()

	texte := make([]string, 0, len(zeilen))
	for _, zeile := range zeilen {
		texte = append(texte, zeile.Text)
	}

	return strings.Join(texte, "\n")
}

// Skriptzeile ist eine Zeile eines Skriptrumpfes samt ihrer Stelle in der
// **Datei** — Werkstück B von Phase 4.9
// (#1324).
//
// # Wozu die Stelle
//
// [Knoten.Skripttext] gibt den Rumpf ohne die Einrückung des Paketformats —
// richtig so, denn in MSE-Script bedeutet sie nichts. Wer aber einen Fehler
// **in der Datei** anstreichen will, braucht sie zurück: Zeile 2 des Rumpfes
// ist Zeile 47 der Datei, und ihre Spalte 1 ist dort Spalte 3.
//
// **Spalte ist die des ersten Zeichens nach der Einrückung** — dieselbe
// Zahl, die der Zerteiler dem Knoten gegeben hat.
type Skriptzeile struct {
	Text   string
	Zeile  int
	Spalte int
}

// Skriptzeilen gibt die Zeilen eines Skriptrumpfes samt ihrer Stelle.
//
// **Beide Formen, ein Weg.** Ein Skriptwert steht entweder hinter seinem
// Doppelpunkt (`script: 1 + 2`) oder eingerückt darunter — im vollen Pack
// **3.313-mal** so und **2.240-mal** so (gezählt am 2026-09-02 über
// `full-magic-pack`, an Zeilen mit einem `script:`-Schlüssel). Bis dahin gab
// diese Methode für die erste Form die leere Zeichenkette, und jeder
// Aufrufer hatte seinen eigenen Rückfall auf `Wert`; zwei Wege für eine
// Frage laufen auseinander, sobald einer eine Regel lernt.
func (k *Knoten) Skriptzeilen() []Skriptzeile {
	if strings.TrimSpace(k.Wert) != "" {
		return []Skriptzeile{{
			Text: k.Wert, Zeile: k.Stelle.Zeile, Spalte: k.wertspalte(),
		}}
	}

	var zeilen []Skriptzeile

	sammleSkript(&zeilen, k.Kinder)

	return zeilen
}

// sammleSkript hängt die Zeilen eines Teilbaums aneinander.
func sammleSkript(zeilen *[]Skriptzeile, knoten []*Knoten) {
	for _, kind := range knoten {
		*zeilen = append(*zeilen, Skriptzeile{
			Text:   strings.TrimLeft(kind.Zeile(), "\t"),
			Zeile:  kind.Stelle.Zeile,
			Spalte: kind.Stelle.Spalte,
		})

		sammleSkript(zeilen, kind.Kinder)
	}
}

// wertspalte gibt die Spalte, in der der Wert einer Eintragszeile beginnt.
//
// **Aus der rohen Zeile gelesen und nicht gerechnet**: zwischen `:` und Wert
// dürfen Leerzeichen stehen, und wie viele, weiß nur die Datei.
func (k *Knoten) wertspalte() int {
	vorspann := len([]rune(k.Schluessel)) + 1
	roh := []rune(k.Zeile())

	if k.Stelle.Spalte-1 > len(roh) {
		return k.Stelle.Spalte
	}

	roh = roh[k.Stelle.Spalte-1:]

	for vorspann < len(roh) && (roh[vorspann] == ' ' || roh[vorspann] == '\t') {
		vorspann++
	}

	return k.Stelle.Spalte + vorspann
}

// Zeile gibt die Zeile eines Knotens — die rohe, wenn es sie noch gibt.
//
// **Sonst zusammengesetzt.** Wer einen Wert ändert, streicht `Roh`; genau daran
// erkennt der Schreiber, dass er neu setzen muss.
func (k *Knoten) Zeile() string {
	if k.Roh != "" {
		return k.Roh
	}

	if k.Zeilenart != ArtEintrag {
		return k.Wert
	}

	return k.Schluessel + ": " + k.Wert
}

// Ersten gibt das erste Kind mit diesem Schlüssel, oder nil.
//
// Auf einem nil-Knoten gibt sie nil zurück: `baum.Ersten("a").Ersten("b")`
// soll sich schreiben lassen, ohne dass jede Zwischenstufe geprüft wird — ein
// fehlender Schlüssel ist in fremden Daten der Normalfall, nicht der Absturz.
func (k *Knoten) Ersten(schluessel string) *Knoten {
	if k == nil {
		return nil
	}

	return ersten(k.Kinder, schluessel)
}

// Alle gibt alle Kinder mit diesem Schlüssel.
func (k *Knoten) Alle(schluessel string) []*Knoten {
	if k == nil {
		return nil
	}

	return alle(k.Kinder, schluessel)
}

func ersten(knoten []*Knoten, schluessel string) *Knoten {
	gesucht := Kanonisch(schluessel)

	for _, kandidat := range knoten {
		if kandidat.trifft(gesucht) {
			return kandidat
		}
	}

	return nil
}

func alle(knoten []*Knoten, schluessel string) []*Knoten {
	gesucht := Kanonisch(schluessel)

	var gefunden []*Knoten

	for _, kandidat := range knoten {
		if kandidat.trifft(gesucht) {
			gefunden = append(gefunden, kandidat)
		}
	}

	return gefunden
}

// trifft sagt, ob dieser Knoten der gesuchte Schlüssel ist.
//
// # Verglichen wird kanonisch
//
// **Beide Schreibweisen kommen wirklich vor**, gemessen am 2026-08-24: die 213
// Vorlagen in `M15-Magic-Pack` schreiben `card field:` mit **Leerzeichen** und
// keine einzige mit Unterstrich; die zehn kanonischen Sets schreiben
// `casting_cost:`, `rule_text:`, `set_info:` mit **Unterstrich** und keines mit
// Leerzeichen. Ein Nachschlag, der eine der beiden Formen nimmt, trifft die
// halbe Welt.
//
// MSE hat dieses Problem nicht, weil sein Leser **jeden** Schlüssel beim
// Einlesen kanonisch macht:
//
//	key = canonical_name_form(trim(key));
//
// `src/util/io/reader.cpp`, Zeile 217, Fassung f9d9356 vom 2021-07-06.
//
// # Warum nicht am Zerteiler, sondern hier
//
// Weil [Knoten.Schluessel] **roh** bleiben muss. Der Zerteiler dürfte ihn nicht
// umschreiben: die Rundlaufzusage gibt die Zeile aus, wie sie kam, und in einer
// Skriptdatei ist der Unterstrich Teil eines Namens, den ein Skript ruft
// (#396).
// Der Vergleich ist der einzige Ort, an dem beide Formen dasselbe heißen
// sollen.
//
// **Was es gekostet hat**: `namenAus` schlug `set info` nach, während jedes von
// MSE geschriebene Set `set_info` trägt — ein Set mit Titel hieß nach seiner
// Datei. Und `punkteAus` suchte `line after`, während MSE `line_after` schreibt
// (`src/util/io/writer.cpp` schreibt den Namen **ohne** Rückumwandlung, Zeile
// 56) — jede Bézierkurve eines echten Symbols wäre als Gerade gelesen worden.
func (k *Knoten) trifft(kanonisch string) bool {
	return k.Zeilenart == ArtEintrag && Kanonisch(k.Schluessel) == kanonisch
}

// Zerteilfehler nennt, was nicht gelesen werden konnte, und wo.
//
// Fremde Daten sind der Normalfall (AGENTS.md §8): ein Paket kann von 2007
// stammen, halb geschrieben sein, in einer Kodierung, die niemand mehr benutzt.
// „Zerteilen fehlgeschlagen" hilft dabei niemandem — die Stelle schon.
type Zerteilfehler struct {
	Stelle  Stelle
	Meldung string

	// Grund ist die Fehlerart, wenn es eine benannte gibt — etwa
	// [ErrKodierung]. Sie ist über `errors.Is` erreichbar, die Stelle über
	// `errors.As`.
	Grund error
}

// Unwrap gibt den [Zerteilfehler.Grund].
func (z *Zerteilfehler) Unwrap() error { return z.Grund }

func (z *Zerteilfehler) Error() string {
	return fmt.Sprintf("%s:%d:%d: %s",
		z.Stelle.Datei, z.Stelle.Zeile, z.Stelle.Spalte, z.Meldung)
}

// Entkanonisch macht aus einem **Feldnamen** der Datei den des Modells.
//
//	void uncanonical_name_form_in_place(String& str) {
//	  for (String::iterator it = str.begin(); it != str.end(); ++it) {
//	    if (*it == '_') *it = ' ';
//	  }
//	}
//
// `src/util/string.cpp`, Zeilen 127–133, Fassung f9d9356 vom 2021-07-06; die
// Gegenrichtung heißt dort `canonical_name_form` und ist das, was **beim
// Schreiben** gilt.
//
// # Warum nicht am Zerteiler
//
// Weil MSE in die **andere** Richtung normiert: sein Leser macht **jeden**
// Schlüssel kanonisch (`key = canonical_name_form(trim(key))`,
// `src/util/io/reader.cpp`, Zeile 217), und sein Skriptzerteiler ebenso
// (`parser.cpp`, Zeile 200). Innen ist bei MSE alles unterstrichen; die
// Leerzeichen gibt es nur in den Dateien.
//
// Dieses Projekt hält es umgekehrt — Feldnamen mit Leerzeichen, wie die Vorlage
// sie erklärt. Beides ist verträglich, **solange die Umformung dort steht, wo
// ein Schlüssel ein Feldname ist**. Am Zerteiler stünde sie überall, auch an
// `mana_t: …` in einer Skriptdatei — und dort ist der Unterstrich Teil des
// Namens, den ein Skript ruft
// (#396).
//
// **Was es gekostet hat:** ein kanonisches Set trägt `rule_text:`, der Feldstil
// heißt `rule text` — der Regeltext kam nie an, und weil ein Textfeld ohne Wert
// gar nicht gezeichnet wird, ohne Befund.
func Entkanonisch(schluessel string) string {
	return strings.ReplaceAll(schluessel, "_", " ")
}

// Kanonisch ist die Gegenrichtung — so schreibt MSE einen Schlüssel in eine
// Datei zurück.
func Kanonisch(schluessel string) string {
	return strings.ReplaceAll(schluessel, " ", "_")
}
