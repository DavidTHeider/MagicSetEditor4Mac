package werkzeuge

// **Jeder Lesezugriff auf eine ganze Datei ist begrenzt — oder begründet** —
// Befund 4 des Zweitblicks über Phase 6.1 (FMT-06.AK-01).
//
// # Der Befund, für den es diesen Wächter gibt
//
// FMT-06 hatte den **Leser** der Pakete begrenzt; der Schreibweg daneben,
// die Symbolzeichnung und die `game`-Datei beim Anlegen eines Stils lasen mit
// `os.ReadFile` und `io.ReadAll` ganz ein. Eine Zusage über „die Dateiwege",
// gehalten an einem von vieren.
//
// # Was er hält
//
// **Jeder Ruf von `os.ReadFile`, `fs.ReadFile` und `io.ReadAll` im
// Produktivcode** unter `internal/` und `cmd/` steht in [lesenOhneBudget] —
// mit Grund —, oder er liest über `io.LimitReader` oder
// `http.MaxBytesReader`. Ein neuer Ruf erzwingt die Entscheidung. Gezählt
// wird je Datei und Funktion, damit eine verschobene Zeile nichts ändert.
//
// # Was er nicht hält
//
// **Ob ein Grund trägt** — das Review. Und einen Leser, der selbst eine
// Schleife dreht, ohne eine dieser drei Funktionen zu rufen.

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// Die Gründe, aus denen eine Stelle ganz lesen darf — jeder sagt, **wessen**
// Datei es ist.
const (
	eingabeDesAufrufers = "die Kommandozeile liest, was ihr Aufrufer ihr gibt — seine Eingabe, " +
		"seine Datei; eine Grenze schützte ihn nur vor sich selbst"
	gewaehlteDatei = "eine Datei, die ein Mensch im Dialog gewählt hat, und kein Paket von außen"
	ausDemUmschlag = "liest aus dem Wegwerfort einer Sendung; das Auspacken hält jedes Objekt bei " +
		"256 MiB und alle zusammen bei 2 GiB (`austausch/quarantaene.go`)"
	eigeneDatei   = "eine Datei, die diese Anwendung selbst geschrieben hat"
	eingebettet   = "eine Wanderung aus `go:embed` — im Programm und nicht auf der Platte"
	werkzeugdatei = "ein Werkzeug der Pipeline liest die Dateien dieses Repositoriums"
	abnahmerumpf  = "nur im Abnahmebau: der Rumpf eines Bindungsrufs, den Wails gleich danach " +
		"selbst ganz liest; das Rufbuch braucht seinen Methodennamen, und der steht im JSON nicht vorn"
)

// lesenOhneBudget sind die Stellen, die eine Datei ganz lesen — mit Grund.
var lesenOhneBudget = map[string]string{
	"internal/fenster/rufbuch.go:rufbuch.Huelle":         abnahmerumpf,
	"cmd/mse/deckeinfuhr.go:deckListeUebernehmenbefehl":  eingabeDesAufrufers,
	"cmd/mse/deckeinfuhr.go:deckListebefehl":             eingabeDesAufrufers,
	"cmd/mse/einstellungsaustausch.go:austauschbefehl":   eingabeDesAufrufers,
	"cmd/mse/handelsleitung.go:positionenLesen":          eingabeDesAufrufers,
	"cmd/mse/kartenablage.go:ablageeinfuegebefehl":       eingabeDesAufrufers,
	"cmd/mse/kontobefehle.go:kennwortVonDerEingabe":      eingabeDesAufrufers,
	"cmd/mse/listeneinfuhr.go:listeDeutenbefehl":         eingabeDesAufrufers,
	"cmd/mse/listeneinfuhr.go:listeEinlesenbefehl":       eingabeDesAufrufers,
	"cmd/mse/nutzerpflege.go:jsonVonDerEingabe":          eingabeDesAufrufers,
	"cmd/mse/paketdatei.go:paketdateisetzenbefehl":       eingabeDesAufrufers,
	"cmd/mse/paketdateibaum.go:paketdateianlegenbefehl":  eingabeDesAufrufers,
	"cmd/mse/sammlungsausfuhr.go:sammlungEinlesenbefehl": eingabeDesAufrufers,
	"cmd/mse/einfuhrvorschau.go:sammlungVorschaubefehl":  eingabeDesAufrufers,
	"cmd/mse/einfuhrvorschau.go:listeVorschaubefehl":     eingabeDesAufrufers,
	"cmd/mse/symbol.go:gelesenerText":                    eingabeDesAufrufers,
	"cmd/mse/tauschabgleich.go:fremdstandLesen":          eingabeDesAufrufers,

	"internal/app/einstellungsaustausch.go:App.EinstellungenEinlesen": gewaehlteDatei,

	"internal/app/deckleiheumschlag.go:leiheAusSendung": ausDemUmschlag,
	"internal/app/umschlag.go:setEntfalten":             ausDemUmschlag,
	"internal/app/umschlagannahme.go:lesenAusSendung":   ausDemUmschlag,
	"internal/austausch/lesen.go:Lies":                  ausDemUmschlag,

	"internal/app/kibildeinsetzen.go:App.bildvorlage": eigeneDatei + " — ein Bild im Beiwerk des Sets; " +
		"was von außen kam, ist beim Einfahren begrenzt worden",
	"internal/app/vorlagenbeigaben.go:ausVerzeichnis":                     eigeneDatei + " — das eben geschriebene Set",
	"internal/netz/speicher.go:Speicher.lies":                             eigeneDatei + " — der Zwischenspeicher, benannt nach einer Prüfsumme",
	"internal/ollama/cloud.go:LiesCloudstand":                             eigeneDatei + " — der Stand der Cloudanmeldung",
	"internal/service/ablagebilder.go:Kartenablage.bilderAnhaengen":       eigeneDatei + " — Bilder im Beiwerk des Sets",
	"internal/service/beiwerk.go:KopiereDatei":                            eigeneDatei + " — kopiert innerhalb des Beiwerks",
	"internal/service/bildvorrat.go:Bildvorrat.Lesen":                     eigeneDatei + " — ein wartender Bildvorschlag",
	"internal/service/epochenmarke.go:Bestandsweg.EpocheNachholen":        eigeneDatei + " — die Marke neben dem Bestand",
	"internal/service/rueckspielmarke.go:Bestandsweg.RueckspielVollenden": eigeneDatei + " — die Marke neben dem Bestand",
	"internal/service/tischbild.go:Bildablage.nutzerdateiBytes":           eigeneDatei + " — Tisch- und Fensterbild, als PNG auf 1920 und 2560 Punkte gebracht",

	"internal/mtgjson/lesen.go:Quelle.dateiLesen": "MTG JSON ist **ein** Dokument, und `AllPrintings` wiegt " +
		"ein halbes Gigabyte: eine Grenze darunter bräche den Zweck, eine darüber hielte nichts auf. " +
		"Der Weg wäre ein strömender Leser wie in `internal/mtgjsonpreise`",
	"internal/schriften/schriftlesen.go:schriftLesen": "eine Schrift des Systems — eine CJK-Schrift " +
		"wiegt 50 MB, und sie liegt dort, wo macOS sie hinlegt",

	"internal/storage/migrationsbuchfuehrung.go:anwenden":         eingebettet,
	"internal/storage/grundlinie.go:grundlinieAnwenden":           eingebettet,
	"internal/storage/migrationsbuchfuehrung.go:summeVergleichen": eingebettet,
	"internal/storage/mindestform.go:migrationstexte":             eingebettet,
	"internal/storage/grundlinie.go:grundlinienumfang":            eingebettet,

	"internal/laufbeleg/lauf/lauf.go:planLesen": werkzeugdatei,
	"internal/laufbeleg/markierte.go:Modul":     werkzeugdatei,
}

func TestJedesGanzeLesenIstBegrenztOderBegruendet(t *testing.T) {
	t.Parallel()

	gefunden := map[string]bool{}

	for _, wurzel := range []string{"../../internal", "../../cmd"} {
		err := filepath.WalkDir(wurzel, func(pfad string, eintrag fs.DirEntry, fehler error) error {
			if fehler != nil {
				return fehler
			}

			if eintrag.IsDir() || !strings.HasSuffix(pfad, ".go") || strings.HasSuffix(pfad, "_test.go") {
				return nil
			}

			for _, stelle := range ganzesLesenIn(t, pfad) {
				gefunden[stelle] = true
			}

			return nil
		})
		if err != nil {
			t.Fatalf("%s durchsuchen: %v", wurzel, err)
		}
	}

	var neu []string

	for stelle := range gefunden {
		if lesenOhneBudget[stelle] == "" {
			neu = append(neu, stelle)
		}
	}

	sort.Strings(neu)

	for _, stelle := range neu {
		t.Errorf("%s liest eine Datei ganz — begrenzen (msepaket.LeseDateiBegrenzt, io.LimitReader) "+
			"oder mit Grund in lesenOhneBudget eintragen", stelle)
	}

	for stelle := range lesenOhneBudget {
		if !gefunden[stelle] {
			t.Errorf("%s steht in lesenOhneBudget, liest aber nicht mehr ganz", stelle)
		}
	}

	if len(gefunden) < 10 {
		t.Fatalf("nur %d Stellen gefunden — greift der Wächter ins Leere?", len(gefunden))
	}
}

// ganzesLesenIn gibt die Stellen einer Datei, die ohne Budget ganz lesen —
// als `pfad:Funktion`.
func ganzesLesenIn(t *testing.T, pfad string) []string {
	t.Helper()

	baum, fehler := parser.ParseFile(token.NewFileSet(), pfad, nil, 0)
	if fehler != nil {
		t.Fatalf("%s lesen: %v", pfad, fehler)
	}

	relativ := strings.TrimPrefix(filepath.ToSlash(pfad), "../../")

	var stellen []string

	for _, erklaerung := range baum.Decls {
		funktion, istFunktion := erklaerung.(*ast.FuncDecl)
		if !istFunktion || funktion.Body == nil {
			continue
		}

		name := funktion.Name.Name
		if funktion.Recv != nil && len(funktion.Recv.List) == 1 {
			name = empfaengername(funktion.Recv.List[0].Type) + "." + name
		}

		ast.Inspect(funktion.Body, func(knoten ast.Node) bool {
			ruf, istRuf := knoten.(*ast.CallExpr)
			if istRuf && liestGanz(ruf) {
				stellen = append(stellen, relativ+":"+name)
			}

			return true
		})
	}

	return stellen
}

// liestGanz sagt, ob ein Ruf eine Datei oder einen Strom ohne Grenze liest.
func liestGanz(ruf *ast.CallExpr) bool {
	auswahl, istAuswahl := ruf.Fun.(*ast.SelectorExpr)
	if !istAuswahl {
		return false
	}

	paket, istName := auswahl.X.(*ast.Ident)
	if !istName {
		return false
	}

	switch paket.Name + "." + auswahl.Sel.Name {
	case "os.ReadFile", "fs.ReadFile":
		return true
	case "io.ReadAll":
		return len(ruf.Args) != 1 || !begrenzt(ruf.Args[0])
	}

	return false
}

// begrenzt sagt, ob ein Strom durch eine Grenze geht.
func begrenzt(ausdruck ast.Expr) bool {
	innen, istRuf := ausdruck.(*ast.CallExpr)
	if !istRuf {
		return false
	}

	auswahl, istAuswahl := innen.Fun.(*ast.SelectorExpr)

	return istAuswahl && (auswahl.Sel.Name == "LimitReader" || auswahl.Sel.Name == "MaxBytesReader")
}

// empfaengername gibt den Typnamen eines Empfängers, ohne Stern.
func empfaengername(ausdruck ast.Expr) string {
	if stern, istStern := ausdruck.(*ast.StarExpr); istStern {
		ausdruck = stern.X
	}

	if name, istName := ausdruck.(*ast.Ident); istName {
		return name.Name
	}

	return "?"
}
