package domain

// **Wege, die `POST` heißen und nur lesen** — Folgebefund F3 des Lastlaufs
// QA-08 (#2020),
// Marke `new`: MSE2 hat weder Dienst noch Werkzeugserver.
//
// # Der Befund
//
// Der Dienst las die Art eines Wegs an seiner Methode: alles außer `GET` und
// `HEAD` buchte als Schreiben. MCP liest sie am Werkzeug. Eine Vorschau, die
// einen Rumpf mitnimmt — ein Set, eine Liste, ein Ausdruck, eine Form — ist im
// Dienst ein `POST` und über MCP ein `lesen.*`: `POST /einstellungen/vorschau`
// kostete 2, `lesen.einstellungen.vorschau` 1. Einundzwanzig solcher Paare
// stehen in `docs/FAEHIGKEITEN.md`.
//
// # Die Entscheidung
//
// **Die Kosten kommen aus derselben Tafel wie bei den teuren Wegen** — je
// Fähigkeit eine Zeile mit ihren Wegen an beiden Flächen, hier mit der Klasse
// Lesen. Beide Flächen fragen [KostenklasseFuer]; der Weg steht in der Tafel,
// und die Methode entscheidet nur noch dort, wo die Tafel schweigt.
//
// **Die Methodenregel für Token bleibt, wie sie ist** (`rest.schreibend`):
// ein Token, das nur lesen darf, kommt an diese Wege weiter nicht heran. Das
// ist ein Satz über Schutz, nicht über Preise — und die Abweichung geht in die
// sichere Richtung. Wer beides an eine Regel hängte, gäbe demselben Weg zwei
// Arten ([0163](../../docs/entscheidungen/0163-zeichnen-belegt-einen-platz-und-lesen-kostet-lesen.md)).
//
// # Was hier nicht steht
//
// **Kommandozeile und Fenster.** Sie buchen nichts; die Spalte der
// Kommandozeile bliebe eine Zeile voller Befehle, die mehr als eine Fähigkeit
// bedienen (`mse symbol-rechnen` steht bei sechs), und der Wächter, der
// verlangt, dass kein Weg in zwei Zeilen steht, ließe das nicht zu.
//
// **Eine Fähigkeit, die zugleich liest und schreibt**, nennt hier nur ihre
// lesenden Wege: bei „Im Text ersetzen" ist es die Vorschau, bei den
// Zwischenablagen das Kopieren. Das Ersetzen selbst und das Ausschneiden
// kosten, was ein Schreiben kostet.
//
// `internal/werkzeuge/kostenparitaet_test.go` hält, dass jedes Paar aus der
// Tabelle der Fähigkeiten an beiden Flächen dasselbe kostet — eine neue
// lesende `POST`-Vorschau ohne Zeile hier wird rot und nennt das Paar.

// lesezeile baut eine Zeile der Klasse Lesen mit ihren Wegen an REST und MCP.
func lesezeile(faehigkeit string, rest, mcp []string) Kostenzeile {
	return Kostenzeile{
		Faehigkeit: faehigkeit, Klasse: KostenLesen,
		Wege: map[Flaeche][]string{FlaecheREST: rest, FlaecheMCP: mcp},
	}
}

// lesendePostwege sind die Fähigkeiten, deren `POST`-Wege kosten, was ihr
// Werkzeug kostet.
var lesendePostwege = []Kostenzeile{
	lesezeile("Ein Set vorab ansehen",
		[]string{"POST /sets/vorschau"}, []string{"lesen.set.vorschau"}),
	lesezeile("Sehen, was eine Berichtigung verschöbe",
		[]string{"POST /turniere/{}/paarungen/{}/folgen"}, []string{"lesen.turnier.folgen"}),
	lesezeile("Karten über die Zwischenablage bewegen",
		[]string{"POST /sets/{}/ablage/kopieren"}, []string{"lesen.ablage.kopieren"}),
	lesezeile("Eine Massenänderung vorher sehen",
		[]string{"POST /sets/{}/felder/vorschau"}, []string{"lesen.felder.vorschau"}),
	lesezeile("Ein Symbol zeichnen",
		[]string{"POST /symbol/grundform"}, []string{"lesen.symbol.grundform"}),
	lesezeile("Einen Bézierhandgriff ziehen",
		[]string{"POST /symbol/griff"}, []string{"lesen.symbol.griff"}),
	lesezeile("Einem Punkt seine Art geben",
		[]string{"POST /symbol/sperre"}, []string{"lesen.symbol.sperre"}),
	lesezeile("Einen Punkt einer Form einfügen oder löschen",
		[]string{"POST /symbol/punkt"}, []string{"lesen.symbol.punkt"}),
	lesezeile("Einen Booster ziehen",
		[]string{"POST /sets/{}/packarten/{}/ziehung"}, []string{"lesen.pack.ziehen"}),
	lesezeile("Im Text ersetzen",
		[]string{"POST /sets/{}/ersetzung/vorschau"}, []string{"lesen.ersetzen"}),
	lesezeile("Einen Ausdruck auswerten",
		[]string{"POST /karten/{}/skript"}, []string{"lesen.skript"}),
	lesezeile("Die Skripte einer Vorlage prüfen",
		[]string{"POST /paketdateien/pruefung"}, []string{"lesen.skriptbefund"}),
	lesezeile("Eine Kartenliste deuten",
		[]string{"POST /sammlung/liste/deuten"}, []string{"lesen.liste.deuten"}),
	lesezeile("Eine Kartenliste vorab ansehen",
		[]string{"POST /sammlung/liste/vorschau"}, []string{"lesen.liste.vorschau"}),
	lesezeile("Eine Deckliste deuten",
		[]string{"POST /decks/{}/liste/vorschau"}, []string{"lesen.deckliste"}),
	lesezeile("Eine CSV-Sammlung vorab ansehen",
		[]string{"POST /sammlung/vorschau"}, []string{"lesen.sammlung.vorschau"}),
	lesezeile("Ein Bild zuschneiden",
		[]string{"POST /symbol/zuschnitt"}, []string{"lesen.symbol.zuschnitt"}),
	lesezeile("Teile gruppieren, auflösen und umordnen",
		[]string{"POST /symbol/gruppe", "POST /symbol/reihenfolge"},
		[]string{"lesen.symbol.gruppe", "lesen.symbol.reihenfolge"}),
	lesezeile("Eine Symmetrie bedienen",
		[]string{"POST /symbol/symmetrie", "POST /symbol/symmetriekopien"},
		[]string{"lesen.symbol.symmetrie", "lesen.symbol.symmetriekopien"}),
	lesezeile("Eine Einfuhr von Einstellungen vorschauen",
		[]string{"POST /einstellungen/vorschau"}, []string{"lesen.einstellungen.vorschau"}),
	// **Der Probelauf vor dem Schreiben** (UXR-F7, #2015) — er schreibt
	// nichts und kostet deshalb, was sein Werkzeug kostet.
	lesezeile("Einen Druckbogen legen",
		[]string{"POST /sets/{}/bogen/probelauf"}, []string{"lesen.set.bogenprobelauf"}),
	lesezeile("Über eine Exportvorlage ausgeben",
		[]string{"POST /sets/{}/vorlagen/probelauf"}, []string{"lesen.vorlage.probelauf"}),
}
