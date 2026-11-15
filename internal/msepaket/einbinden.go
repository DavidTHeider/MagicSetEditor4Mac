package msepaket

import (
	"errors"
	"fmt"
	"slices"
	"strings"
)

// ErrPaketpfad meldet einen Verweis, der so nicht aufzulösen ist.
var ErrPaketpfad = errors.New("kein gültiger Paketpfad")

// Paketpfad ist ein Verweis der Form `/paket.mse-include/datei`.
//
// **Kein Dateisystempfad.** Er beginnt mit `/` und meint trotzdem nicht die
// Wurzel der Platte, sondern ein anderes **Paket**. Wer ihn an `os.Open` reicht,
// liest im günstigen Fall nichts und im ungünstigen etwas Fremdes.
//
// Zerlegt wird deshalb mit `path`, nie mit `filepath`: `path` kennt keine
// Laufwerksbuchstaben und keinen Gegenschrägstrich, und genau das ist hier
// richtig — ein Paketpfad sieht auf jeder Plattform gleich aus.

type Paketquelle interface {
	Lies(pfad Paketpfad) ([]byte, error)
}

// Einbinder löst `include file:` auf.
type Einbinder struct {
	quelle Paketquelle

	// eigen ist der Name des Pakets, in dem eingebunden wird.
	//
	// **Er ist nötig, weil ein Verweis ohne Schrägstrich das eigene Paket
	// meint** (siehe [ZerlegeEigenpfad]). Leer heißt: ein solcher Verweis
	// lässt sich nicht auflösen und wird benannt — so verhält sich ein
	// Einbinder, den jemand ohne Paketbezug baut.
	eigen string

	// sprache ist die bevorzugte Sprachfassung für
	// `include localized file:` — leer heißt: die Grundfassung.
	sprache string

	befunde []Befund
	imLauf  []string

	// einbindungen zählt die gelesenen Dateien eines Laufs — gegen
	// [GrenzeEinbindungen]. **Eine Raute ist kein Kreis**: dieselbe Datei
	// zehnmal, sechs Ebenen tief, ist eine Million Auflösungen, und die
	// Kreisprüfung sieht keine davon (FMT-06.AK-01).
	einbindungen int

	// ueberBudget sagt, ob die Grenze schon gemeldet ist — **einmal**, nicht
	// je übergangenem Verweis.
	ueberBudget bool
}

// NeuerEinbinder baut den Einbinder über einer Quelle.
func NeuerEinbinder(quelle Paketquelle) *Einbinder {
	return &Einbinder{quelle: quelle}
}

// InSprache sagt dem Einbinder, welche Sprachfassung er bevorzugt.
//
// # Was das heißt — und woher es kommt
//
// `include localized file: keywords` lädt `keywords_<sprache>`, wenn es die
// Datei gibt, und sonst `keywords`
// (#1065).
//
// **Das ist erschlossen, nicht nachgelesen.** `mse2-quelle` f9d9356 vom
// 2021-07-06 kennt den Schlüssel nicht — `Reader::unknownKey` behandelt
// `include_file` und sonst keinen, und eine Suche über den ganzen Baum nach
// „localized file" ist leer. Das Pack trägt `mse version: 2.5.8` und benutzt
// damit eine Fassung, die neuer ist als unsere Referenz.
//
// Die Bedeutung ist aus den **Dateinamen** geschlossen: neben `keywords`
// liegen im `M15-Magic-Pack` achtzehn Fassungen `keywords_de`, `keywords_fr`,
// `keywords_ja` … Die Erschlossenheit steht in `docs/FORMATE.md` mit
// Sicherheitsgrad und in `docs/PARITAET.md` als `uncertain`.
//
// **Leer heißt: die Grundfassung.** Ohne gewählte Sprache lädt der Einbinder
// die Datei ohne Zusatz — das ist der Zustand, in dem dieses Programm heute
// läuft.
func (e *Einbinder) InSprache(sprache string) *Einbinder {
	e.sprache = sprache

	return e
}

// ImPaket sagt dem Einbinder, in wessen Paket er arbeitet.
func (e *Einbinder) ImPaket(paket string) *Einbinder {
	e.eigen = paket

	return e
}

// Befunde sind die Verweise, die nicht aufzulösen waren.
func (e *Einbinder) Befunde() []Befund { return e.befunde }

// Einbinden ersetzt jeden `include file:`-Knoten durch den Inhalt seiner Datei.
//
// # Ein fehlendes Paket ist ein Befund, kein Absturz
//
// Fremde Daten sind der Normalfall (AGENTS.md §8). Ein Stil, der ein Modul
// nennt, das hier nicht installiert ist, wird mit einem benannten Vorbehalt
// geöffnet — nicht abgewiesen. Wer ihn abwiese, machte aus einer fehlenden
// Abhängigkeit ein unlesbares Set.
func (e *Einbinder) Einbinden(baum *Baum) *Baum {
	// **Datei und BOM reisen mit.**
	//
	// Der erste Anlauf gab nur die Wurzel weiter. Solange niemand schrieb, fiel
	// das nicht auf; mit `Schreibe` verlor ein eingebundener Baum seine BOM —
	// und ein Rundlauf über ein Paket mit `include file:` hätte drei Bytes
	// weniger zurückgegeben, als er bekam. Gefunden hat es der Zweitblick über
	// Phase 5 als Naht zwischen zwei PRs, die je für sich richtig waren.
	return &Baum{
		Datei:  baum.Datei,
		MitBOM: baum.MitBOM,
		Wurzel: e.knotenliste(baum.Wurzel),
	}
}

// knotenliste baut eine **neue** Liste; der Eingabebaum bleibt unberührt.
//
// Der erste Anlauf schrieb `kind.Kinder` neu und teilte die Knoten mit dem
// Original — der Aufrufer hielt danach einen Baum, dessen Tiefen teils
// aufgelöst waren und teils nicht. Ein Ergebnis, das seine Eingabe verändert,
// ist an einer Stelle brauchbar und an jeder zweiten eine Überraschung.
// Gefunden vom Zweitblick über Phase 3.
func (e *Einbinder) knotenliste(knoten []*Knoten) []*Knoten {
	ergebnis := make([]*Knoten, 0, len(knoten))

	for _, kind := range knoten {
		if kind.Zeilenart == ArtEintrag && istVerweisschluessel(kind.Schluessel) {
			// **Was unter dem Verweis steht, bleibt** — und zwar hinter der
			// geladenen Datei, an der Stelle, an der es stand.
			//
			// # Der Befund
			//
			// `magic-blends.mse-include/new-blends` beginnt mit einer
			// eingerückten `include file: masks`-Zeile, und die fünfzehn
			// Vereinbarungen danach sind ebenso eingerückt. Im Baum hängen sie
			// damit **unter** dem Verweis. Hier stand nur die geladene Datei,
			// und die fünfzehn fielen weg.
			//
			// Nachgemessen am zusammengesetzten Spielskript von
			// `magic.mse-game`, 19.125 Zeilen: `trim_colors` wurde achtmal
			// gerufen und nirgends definiert. Das war der Grund, warum
			// `card color → image` in `magic-m15.mse-style` nicht zeichnete
			// (#240).
			ergebnis = append(ergebnis, e.aufloesen(kind)...)
			ergebnis = append(ergebnis, e.knotenliste(kind.Kinder)...)

			continue
		}

		kopie := *kind
		kopie.Kinder = e.knotenliste(kind.Kinder)

		// **Ein Verweis darf auch der Wert einer Zeile sein.**
		//
		// `magic.mse-game` schreibt `init script: include file: script` in
		// **einer** Zeile. Bis der Zweitblick über Phase 6 es fand, sah der
		// Einbinder nur `include file:` als eigenen Schlüssel — und das
		// Initskript des Spielpakets war 74 Zeichen lang statt der ganzen
		// Bibliothek (#195).
		//
		// **Der Wert wird geleert und die Datei zu Kindern.** So liest
		// [Knoten.Skripttext] sie — er sammelt Kinder, nicht Werte. Und `Roh`
		// fällt weg, denn die Zeile stimmt danach nicht mehr; genau daran
		// erkennt der Schreiber, dass er neu setzen muss.
		if verweis, da := eingebetteterVerweis(&kopie); da {
			geladen := e.aufloesen(&Knoten{
				Zeilenart:  ArtEintrag,
				Schluessel: verweisschluessel,
				Wert:       verweis,
				Stelle:     kopie.Stelle,
			})

			kopie.Wert = ""
			kopie.Roh = ""
			kopie.Kinder = append(geladen, kopie.Kinder...)
		}

		ergebnis = append(ergebnis, &kopie)
	}

	return ergebnis
}

// aufloesen liest die verwiesene Datei und gibt ihre Wurzelknoten.
func (e *Einbinder) aufloesen(knoten *Knoten) []*Knoten {
	pfad, err := e.pfadAus(knoten.Wert)
	if err != nil {
		e.melde(knoten, err.Error())

		return nil
	}

	if Kanonisch(knoten.Schluessel) == Kanonisch(sprachverweisschluessel) {
		pfad = e.sprachfassung(pfad)
	}

	schluessel := pfad.Paket + "/" + pfad.Datei

	if !e.imBudget(knoten, schluessel) {
		return nil
	}

	// **Der Kreis wird benannt.** `a` bindet `b` ein, `b` wieder `a`: ohne diese
	// Prüfung liest der Einbinder, bis der Stapel voll ist.
	if kreis := slices.Index(e.imLauf, schluessel); kreis >= 0 {
		e.melde(knoten, "diese Dateien binden sich im Kreis ein: "+
			strings.Join(append(e.imLauf[kreis:], schluessel), " → "))

		return nil
	}

	inhalt, err := e.quelle.Lies(pfad)
	if err != nil {
		e.melde(knoten, fmt.Sprintf("%s ist nicht lesbar: %v", schluessel, err))

		return nil
	}

	baum, fehler := Zerteile(schluessel, inhalt)
	if fehler != nil {
		e.melde(knoten, fmt.Sprintf("%s ließ sich nicht lesen: %v", schluessel, fehler))

		return nil
	}

	e.imLauf = append(e.imLauf, schluessel)
	defer func() { e.imLauf = e.imLauf[:len(e.imLauf)-1] }()

	// **Ab hier gilt das Paket der gelesenen Datei, nicht das des Aufrufers.**
	//
	// Ein `include file:` ohne Schrägstrich meint das eigene Paket (siehe
	// [ZerlegeEigenpfad]) — und „eigen" ist das Paket der **Datei, in der die
	// Zeile steht**, nicht das, in dem der Lauf begonnen hat. Bis
	// #241 trug
	// der Einbinder einen Namen über den ganzen Lauf; damit suchte er jede
	// eingebundene Datei im Ausgangspaket.
	//
	// Sichtbar wurde es an `magic.mse-game/script`, das
	// `/magic-blends.mse-include/new-blends` einbindet, dessen Zeile 10 wieder
	// `include file: masks` sagt: gesucht wurde `magic.mse-game/masks`, das es
	// nicht gibt, statt `magic-blends.mse-include/masks`, das dasteht. Der
	// Befund stand als „ein Verzeichnis wird nicht aufgelöst" im Bestand — die
	// Deutung war falsch, `masks` ist eine gewöhnliche Datei.
	//
	// **Zurückgesetzt wird, weil danach weitergelesen wird.** Auf die
	// eingebundene Zeile folgen in `script` weitere; ohne die Rückgabe suchten
	// sie im zuletzt besuchten Paket — derselbe Fehler, nur verschoben.
	vorher := e.eigen
	e.eigen = pfad.Paket

	defer func() { e.eigen = vorher }()

	return e.knotenliste(baum.Wurzel)
}

// imBudget sagt, ob noch eingebunden werden darf, und zählt die Einbindung.
//
// **Zwei Grenzen, zwei Befunde**: zu tief ist eine Kette, die den Stapel
// wachsen lässt; zu viele ist eine Raute, die die Zeit bestimmt. Beide werden
// benannt, und was dahinter läge, fehlt — ein Paket, das so gebaut ist, wird
// mit Befund geöffnet, nicht abgewiesen (§8).
func (e *Einbinder) imBudget(knoten *Knoten, schluessel string) bool {
	if len(e.imLauf) >= GrenzeEinbindetiefe {
		e.melde(knoten, fmt.Sprintf("%s liegt zu tief — mehr als %d Ebenen eingebunden: %s",
			schluessel, GrenzeEinbindetiefe, strings.Join(e.imLauf[:min(3, len(e.imLauf))], " → ")+" → …"))

		return false
	}

	if e.einbindungen >= GrenzeEinbindungen {
		if !e.ueberBudget {
			e.ueberBudget = true
			e.melde(knoten, fmt.Sprintf("mehr als %d Einbindungen in einem Lauf — %s und alles "+
				"danach fehlt", GrenzeEinbindungen, schluessel))
		}

		return false
	}

	e.einbindungen++

	return true
}

func (e *Einbinder) melde(knoten *Knoten, meldung string) {
	e.befunde = append(e.befunde, Befund{Stelle: knoten.Stelle, Meldung: meldung})
}

// pfadAus deutet einen Verweis — fremdes Paket oder eigenes.
//
// **Der Schrägstrich entscheidet**, und nur er: ein Verweis mit ihm nennt ein
// Paket, einer ohne ihn eine Datei nebenan.
func (e *Einbinder) pfadAus(verweis string) (Paketpfad, error) {
	if strings.HasPrefix(strings.TrimSpace(verweis), "/") {
		return ZerlegePaketpfad(verweis)
	}

	if e.eigen == "" {
		return Paketpfad{}, fmt.Errorf(
			"%w: %q meint das eigene Paket, und dieser Einbinder kennt keines",
			ErrPaketpfad, verweis)
	}

	return ZerlegeEigenpfad(e.eigen, verweis)
}

// eingebetteterVerweis liest ein `include file:` aus dem **Wert** einer Zeile.
//
// **Nur bei einem Eintrag, und nur wenn der Wert damit beginnt.** Ein
// Skripttext, in dem die Zeichenfolge irgendwo mitten vorkommt, ist keiner —
// er ist ein Skript, das sie erwähnt.
func eingebetteterVerweis(knoten *Knoten) (string, bool) {
	if knoten.Zeilenart != ArtEintrag {
		return "", false
	}

	rest, da := strings.CutPrefix(strings.TrimSpace(knoten.Wert), "include file:")
	if !da {
		return "", false
	}

	verweis := strings.TrimSpace(rest)

	return verweis, verweis != ""
}

// Die beiden Schlüssel, die einen Verweis auf eine Datei tragen.
//
// **`include localized file:` ist erschlossen, nicht nachgelesen** — siehe
// [Einbinder.InSprache].
const (
	verweisschluessel       = "include file"
	sprachverweisschluessel = "include localized file"
)

// istVerweisschluessel sagt, ob ein Schlüssel eine Datei einbindet — **in
// kanonischer Form**, wie MSE2 nach `include_file` fragt
// (`src/util/io/reader.hpp` Zeile 175, Fassung f9d9356 vom 2021-07-06).
// Bis QA-06b (#1967) blieb `include_file:` als gewöhnlicher Eintrag stehen.
func istVerweisschluessel(schluessel string) bool {
	return Kanonisch(schluessel) == Kanonisch(verweisschluessel) ||
		Kanonisch(schluessel) == Kanonisch(sprachverweisschluessel)
}

// sprachfassung wählt die Datei zu einem `include localized file:`.
//
// # Die Reihenfolge
//
// Gewählte Sprache, dann **Englisch**, dann die Grundfassung ohne Zusatz.
// Keiner der drei Rückfälle ist ein Befund: ein Pack, das `keywords_de` nicht
// mitbringt, soll seine Schlüsselwörter trotzdem haben — englisch statt gar
// nicht (§8).
//
// # Warum Englisch dazwischen steht
//
// **Hier fehlte es**, und der Fremdpaketlauf zur Freigabe 0.3.0 hat es
// aufgedeckt: `magic.mse-game` bindet `statistics` und `keywords` so ein, und
// **eine Datei ohne Zusatz gibt es zu keinem von beiden**. Der Rückfall auf
// die Grundfassung suchte also etwas, das kein Pack mitbringt — zwei Befunde
// je Karte, über den Kanon 1.360
// (#1170).
//
// **Gezählt in `full-magic-pack/data/magic.mse-game`**: vierzehn Fassungen von
// `keywords` und dreizehn von `statistics` — `_ch-s`, `_ch-t`, `_da`, `_de`,
// `_en`, `_es`, `_fr`, `_it`, `_jp`, `_ko`, `_pl`, `_pt-br`, `_ru`, dazu
// `keywords_custom` und `statistics_script`. `_en` ist bei beiden dabei.
//
// # Was erschlossen bleibt
//
// **Dass `_en` die Vorgabe ist.** MSE2 f9d9356 kennt den Schlüssel gar nicht;
// das Pack trägt `mse version: 2.5.8`. Die Marke in `docs/PARITAET.md` bleibt
// `uncertain` — was sich geändert hat, ist die Zahl der Belege, nicht ihre
// Art.
func (e *Einbinder) sprachfassung(pfad Paketpfad) Paketpfad {
	for _, zusatz := range []string{e.sprache, "en"} {
		if zusatz == "" {
			continue
		}

		versuch := pfad
		versuch.Datei = pfad.Datei + "_" + zusatz

		if _, lesbar := e.quelle.Lies(versuch); lesbar == nil {
			return versuch
		}
	}

	return pfad
}
