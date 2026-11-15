package werkzeuge_test

// **Jede Ablage fragt nach dem Profil** —
// #1815.
//
// # Warum es diesen Wächter gibt
//
// `TestJedeWurzelabfrageKlemmtAufEinProfil` hält die zehn Wurzeltafeln; die
// **Abkömmlinge** — Karten, Deckplätze, Runden, Partieereignisse — standen
// ausdrücklich nicht darunter (0066). Am 2026-09-23 schrieb ein Nutzer an der
// Kommandozeile mit `mse feld 1 1 name …` den Kartennamen eines anderen um:
// die Kennung kam vom Aufrufer, und keine Ablage fragte, wem sie gehört.
//
// # Was er hält
//
// **Jede exportierte Methode in `internal/storage` erreicht eine Klemme** —
// selbst oder über eine Methode desselben Empfängers oder eine Funktion des
// Pakets: `ProfilAus` (eine Wurzel klemmt in ihrer eigenen Abfrage) oder einen
// Helfer aus `besitz.go` (ein Abkömmling fragt nach seiner Wurzel). Was keine
// erreicht, steht in [ohneBesitzfrage] mit Grund — eine neue Methode erzwingt
// damit die Entscheidung, statt still ohne Klemme durchzukommen.
//
// # Was er nicht hält
//
// **Ob die Klemme das Richtige klemmt.** Er sieht, dass gefragt wird, nicht
// wonach: eine Methode, die `ProfilAus` ruft und den Wert dann nicht in ihre
// Abfrage setzt, käme durch. Dafür stehen die Läufe
// `internal/storage/besitz*_intern_test.go`, die jede Methode aus einem fremden
// Profil rufen und danach nachzählen.
//
// **Und Rufe über ein Feld** (`e.ablage.Deck(…)`) verfolgt er nicht: dafür
// bräuchte er die Typen. Eine Methode, die nur so klemmt, steht mit Grund in
// der Liste.

import (
	"go/ast"
	"go/parser"
	"go/token"
	"maps"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// klemmen sind die Funktionen, deren Ruf als Frage nach dem Profil zählt.
var klemmen = []string{
	"ProfilAus", "besitzt",
	"setImProfil", "karteImProfil", "kartenImProfil", "deckImProfil", "platzImDeck",
	"exemplarImProfil", "turnierImProfil", "teilnehmerImProfil", "teilnehmerImTurnier",
	"paarungImProfil", "paarungImTurnier", "partieImProfil",
}

// ohneBesitzfrage sind die Methoden, die bewusst keine Klemme erreichen — mit
// Grund. `Typ.*` nimmt einen ganzen Empfänger aus.
var ohneBesitzfrage = map[string]string{
	// **Die Vermerke gehören dem Bestand** (M01, #2345): eine Sicherung trägt
	// alle Profile, und ihr Träger ist der Bestand (`docs/SICHERUNG.md`,
	// Rollen).
	"Sicherungsvermerke.*": "eine Sicherung trägt alle Profile — ihr Träger ist der Bestand, nicht ein Profil",
	// **Die Wanderungen einer künftigen Fassung** (M01): ein Dateisystem für
	// Läufe, keine Ablage.
	"ueberlagert.*": "ein Dateisystem aus Wanderungen für Läufe — darunter liegt keine Ablage",
	// **Ein Ruf über einen neuen Empfänger** (G04, #2251): der Leser des
	// Setschreibers reicht an `Stilablage.Stilwerte` durch, und die klemmt
	// mit `setImProfil` — derselbe Weg, eine Stelle.
	"SetLeser.Stilwerte": "reicht an Stilablage.Stilwerte durch, und die fragt nach dem Profil",
	// **Die Hülle um den Treiber** (BK-03b, Vermerk 0175) — sie reicht
	// Anweisungen durch und fängt den Zusammenbruch; welches Profil eine
	// Anweisung meint, entscheidet die Ablage, die sie stellt.
	"fangverbinder.*":  "die Hülle um den Treiber — darunter liegt keine Ablage",
	"fangverbindung.*": "die Hülle um den Treiber — darunter liegt keine Ablage",
	"fanganweisung.*":  "die Hülle um den Treiber — darunter liegt keine Ablage",
	"fangzeilen.*":     "die Hülle um den Treiber — darunter liegt keine Ablage",
	"fanggeschaeft.*":  "die Hülle um den Treiber — darunter liegt keine Ablage",

	// **Was keinem Profil gehört.**
	"Betriebslage.*": "die Lage des Bestands selbst — Wanderung und Probe gehen über die " +
		"ganze Datei, nicht über ein Profil",
	"SetLeser.HerkunftVerwiesen": "fragt mit Absicht über jedes Profil und sagt nur ja oder nein — " +
		"dieselbe Ausnahme wie `herkunftverweis.go` im Wächter über die Wurzelabfragen (#1944)",
	"Standleser.*": "der Datenstand zählt Änderungen an der ganzen Datei; er sagt, dass sich " +
		"etwas geändert hat, nicht was",
	"Waisenzahlen.*": "die Diagnose zählt verwaiste Zeilen über jedes Profil — dieselbe " +
		"Ausnahme wie `waisenzaehler.go` im Wächter über die Wurzelabfragen",
	"Drucke.*": "der Katalog der Drucke gehört dem Rechner: ein Druck ist fremdes Material, " +
		"das jedes Profil gleich nachschlägt (0035)",
	"Preisspeicher.*": "Preise sind fremde Daten über einen Druck und gehören keinem Profil (0035)",
	"Paketkatalog.*":  "Pakete und Vorlagen gehören dem Rechner, nicht einem Profil (0035)",
	"Sprachpaketablage.*": "eigene Übersetzungen gehören dem Rechner wie ein MSE-Paket " +
		"(Wanderung 0092, Vermerk 0147); welche gilt, sagt die Einstellung des Menschen",
	"Wiederherstellungsbuch.*": "eine Wiederherstellung tauscht den ganzen Bestand; ihre Epochen " +
		"gelten über jedes Profil, und die Mengen darin klemmen beim Lesen",
	"Retrievalindex.*": "der Suchindex über Regeln und Hilfe ist öffentlich (Profil 0); was ein " +
		"Profil sperrt, nennt der Dienst ausdrücklich",
	"Nutzerablage.*": "Menschen und ihre Profile stehen über den Profilen; wer über wen schreiben " +
		"darf, entscheiden `UeberNutzer` und `UeberProfil` im Dienst (ID-03)",

	// **Was das Profil ausdrücklich nimmt** — gesetzt vom Dienst, nie von einer Fläche.
	"Kostenjournal.*": "nimmt das Profil als Wert; `service.Kostenwache` setzt es aus dem Akteur, " +
		"und die Kennung eines Postens entsteht beim Reservieren im selben Auftrag",
	"Modellprobenablage.*": "nimmt das Profil als Wert; die Anwendung setzt es aus dem Kontext " +
		"(`storage.ProfilAus`) und nie aus einer Angabe des Rufers",
	"Nachweisablage.*": "ein Nachweis ist eine Einmalkennung, die das Fenster ausstellt und " +
		"`service.Nachweise` an Vorgang, Stand und Nutzer bindet",
	"Freigabenablage.*": "Freigaben stehen über dem gezeigten Profil: wer sie setzt, braucht die " +
		"Rolle am **Zielprofil** (`App.alsVerwalter`)",
	"Zugriffstokenablage.*": "Token nehmen Profil oder Kennung ausdrücklich; die Rolle am " +
		"Zielprofil fragt die Anwendung (#1817), und die Wache sucht ein vorgezeigtes Token, " +
		"bevor es ein Profil gibt",
	"Vorgabenablage.*": "nimmt das Profil als Wert; die Rolle am Zielprofil fragt die " +
		"Anwendung (#1817)",

	// **Was eine Grenze bewusst überschreitet.**
	"Deckentnahme.DeckAusProfil": "liest ein Deck aus einem ausdrücklich genannten fremden " +
		"Profil; die Erlaubnis prüft `service.Deckuebernahme`, sein einziger Rufer",
	"Deckentnahme.DeckeAusProfil": "nennt die Decks eines ausdrücklich genannten fremden " +
		"Profils zur Wahl; die Erlaubnis prüft `service.Deckuebernahme` davor (UX-08.AK-04)",

	// **Erbauer** — sie geben eine Ablage zurück und lesen nichts.
	"Kartenschreiber.Als":      "ein Erbauer: gibt einen Schreiber mit anderem Urheber zurück",
	"Kartenschreiber.Von":      "ein Erbauer: gibt einen Schreiber mit anderer Herkunft zurück",
	"Schluesselwortablage.Als": "ein Erbauer: gibt eine Ablage mit anderem Urheber zurück",
	"SetSpeicher.MitUhr":       "ein Erbauer: setzt die Uhr und gibt die Ablage zurück",
}

// **Jede exportierte Methode der Ablage erreicht eine Klemme.**
func TestJedeAblagemethodeFragtNachDemProfil(t *testing.T) {
	t.Parallel()

	bestand := funktionenAus(t, "storage")

	var ohne []string

	for schluessel, funktion := range bestand.funktionen {
		if funktion.Recv == nil || !funktion.Name.IsExported() || ausgenommen(schluessel) {
			continue
		}

		if !bestand.erreicht(schluessel, klemmen, map[string]bool{}) {
			ohne = append(ohne, schluessel)
		}
	}

	slices.Sort(ohne)

	for _, schluessel := range ohne {
		t.Errorf("%s fragt nicht nach dem Profil — Klemme aus besitz.go rufen "+
			"oder mit Grund in ohneBesitzfrage eintragen", schluessel)
	}

	// **Und er muss etwas gesehen haben** — ein Pfad, der danebengreift, wäre
	// grün und prüfte nichts.
	if len(bestand.funktionen) < 500 {
		t.Errorf("nur %d Funktionen gelesen — der Pfad greift daneben", len(bestand.funktionen))
	}
}

// **Jede Ausnahme gibt es, und sie ist begründet.**
func TestJedeAusnahmeVonDerBesitzfrageIstBegruendet(t *testing.T) {
	t.Parallel()

	bestand := funktionenAus(t, "storage")
	empfaenger := map[string]bool{}

	for schluessel := range bestand.funktionen {
		if typ, _, hatTyp := strings.Cut(schluessel, "."); hatTyp {
			empfaenger[typ] = true
		}
	}

	for _, schluessel := range slices.Sorted(maps.Keys(ohneBesitzfrage)) {
		if len(ohneBesitzfrage[schluessel]) < 40 {
			t.Errorf("die Ausnahme %q ist nicht begründet", schluessel)
		}

		typ, methode, _ := strings.Cut(schluessel, ".")

		switch {
		case methode == "*" && !empfaenger[typ]:
			t.Errorf("die Ausnahme %q nennt einen Empfänger, den es nicht gibt", schluessel)
		case methode != "*" && bestand.funktionen[schluessel] == nil:
			t.Errorf("die Ausnahme %q nennt eine Methode, die es nicht gibt", schluessel)
		}
	}
}

// ausgenommen sagt, ob eine Methode oder ihr Empfänger in der Liste steht.
func ausgenommen(schluessel string) bool {
	typ, _, _ := strings.Cut(schluessel, ".")

	return ohneBesitzfrage[schluessel] != "" || ohneBesitzfrage[typ+".*"] != ""
}

// quellbestand sind die Funktionen eines Pakets, nach `Typ.Methode` oder
// `Funktion`, und die Verbünde, die ein Feld `ProfilID` tragen.
type quellbestand struct {
	funktionen map[string]*ast.FuncDecl
	mitProfil  map[string]bool
}

// funktionenAus liest jede Funktion aus `internal/<paket>`.
func funktionenAus(t *testing.T, paket string) quellbestand {
	t.Helper()

	dateien, fehler := filepath.Glob(filepath.Join("..", paket, "*.go"))
	if fehler != nil {
		t.Fatalf("die Dateien finden: %v", fehler)
	}

	bestand := quellbestand{funktionen: map[string]*ast.FuncDecl{}, mitProfil: map[string]bool{}}
	menge := token.NewFileSet()

	for _, pfad := range dateien {
		if strings.HasSuffix(pfad, "_test.go") {
			continue
		}

		datei, lesefehler := parser.ParseFile(menge, pfad, nil, 0)
		if lesefehler != nil {
			t.Fatalf("%s lesen: %v", pfad, lesefehler)
		}

		for _, erklaerung := range datei.Decls {
			switch knoten := erklaerung.(type) {
			case *ast.FuncDecl:
				bestand.funktionen[schluesselVon(knoten)] = knoten
			case *ast.GenDecl:
				bestand.verbuendeMitProfil(knoten)
			}
		}
	}

	return bestand
}

// verbuendeMitProfil merkt sich die Verbünde einer Erklärung, die ein Feld
// `ProfilID` tragen.
func (b quellbestand) verbuendeMitProfil(erklaerung *ast.GenDecl) {
	for _, angabe := range erklaerung.Specs {
		typ, istTyp := angabe.(*ast.TypeSpec)
		if !istTyp {
			continue
		}

		verbund, istVerbund := typ.Type.(*ast.StructType)
		if !istVerbund {
			continue
		}

		for _, feld := range verbund.Fields.List {
			for _, name := range feld.Names {
				if name.Name == "ProfilID" {
					b.mitProfil[typ.Name.Name] = true
				}
			}
		}
	}
}

// schluesselVon gibt `Typ.Methode` für eine Methode und den Namen für eine
// Funktion.
func schluesselVon(funktion *ast.FuncDecl) string {
	if funktion.Recv == nil || len(funktion.Recv.List) == 0 {
		return funktion.Name.Name
	}

	return empfaengertyp(funktion) + "." + funktion.Name.Name
}

// empfaengertyp gibt den Namen des Empfängertyps ohne Stern.
func empfaengertyp(funktion *ast.FuncDecl) string {
	ausdruck := funktion.Recv.List[0].Type
	if stern, istStern := ausdruck.(*ast.StarExpr); istStern {
		ausdruck = stern.X
	}

	if name, istName := ausdruck.(*ast.Ident); istName {
		return name.Name
	}

	return "?"
}

// empfaengername gibt den Namen, unter dem der Rumpf seinen Empfänger kennt.
func empfaengername(funktion *ast.FuncDecl) string {
	if funktion.Recv == nil || len(funktion.Recv.List[0].Names) == 0 {
		return ""
	}

	return funktion.Recv.List[0].Names[0].Name
}

// erreicht sagt, ob eine Funktion eines der Ziele ruft — selbst oder über
// eine Funktion des Pakets oder eine Methode desselben Empfängers.
func (b quellbestand) erreicht(schluessel string, ziele []string, gesehen map[string]bool) bool {
	if gesehen[schluessel] {
		return false
	}

	gesehen[schluessel] = true

	funktion := b.funktionen[schluessel]
	if funktion == nil || funktion.Body == nil {
		return false
	}

	for _, gerufen := range b.gerufene(funktion) {
		if slices.Contains(ziele, gerufen) || b.erreicht(gerufen, ziele, gesehen) {
			return true
		}
	}

	return false
}

// gerufene nennt die Funktionen des Pakets, die ein Rumpf ruft.
func (b quellbestand) gerufene(funktion *ast.FuncDecl) []string {
	var (
		gerufen []string
		selbst  = empfaengername(funktion)
		typ     = ""
	)

	if funktion.Recv != nil {
		typ = empfaengertyp(funktion)
	}

	ast.Inspect(funktion.Body, func(knoten ast.Node) bool {
		ruf, istRuf := knoten.(*ast.CallExpr)
		if !istRuf {
			return true
		}

		switch ziel := ruf.Fun.(type) {
		case *ast.Ident:
			gerufen = append(gerufen, ziel.Name)
		case *ast.SelectorExpr:
			if selbst != "" && vomEmpfaenger(ziel.X, selbst) {
				gerufen = append(gerufen, typ+"."+ziel.Sel.Name)
			}
		}

		return true
	})

	return gerufen
}

// vomEmpfaenger sagt, ob ein Ausdruck der Empfänger ist — oder ein Erbauer auf
// ihm, der denselben Typ zurückgibt (`k.Von(herkunft).FeldSetzenMitVorwert`).
//
// **Angenommen und nicht nachgesehen**: dass `k.Von(…)` einen Kartenschreiber
// gibt, wüsste nur die Typprüfung. Die Erbauer der Ablage heißen `Als`, `Von`
// und `MitUhr` und geben alle ihren eigenen Typ; ein Ruf, der hier fälschlich
// als eigener durchginge, fände seine Klemme nicht und stünde als Befund da.
func vomEmpfaenger(ausdruck ast.Expr, selbst string) bool {
	switch knoten := ausdruck.(type) {
	case *ast.Ident:
		return knoten.Name == selbst
	case *ast.CallExpr:
		if auswahl, istAuswahl := knoten.Fun.(*ast.SelectorExpr); istAuswahl {
			return vomEmpfaenger(auswahl.X, selbst)
		}
	}

	return false
}
