package werkzeuge

// **Die Ausnahmen der Meldungswege** — aus `meldungswege_test.go`
// herausgelöst, als sie mit der Partiedatei über die Zeilengrenze aus §2
// ging (PLAY-03.AK-05,
// #1711).
//
// **Der Schnitt folgt der Frage**: dort steht, **wie** gesucht wird und was
// ein Befund ist; hier steht, **was** erlaubt ist — und jede Zeile trägt
// ihren Grund.

// meldungsausnahmen sind sie alle.
var meldungsausnahmen = []meldungsausnahme{
	// --- Ein „Pfad", der keiner ist -------------------------------------

	{
		"service/kartenablage.go", "gehören zum Spiel",
		"`ziel` ist hier ein **Spielname** und kein Pfad — `spielVon` gibt das " +
			"Spiel des Sets. Der Wächter geht nach dem Bezeichner, und der ist " +
			"an dieser Stelle unglücklich gewählt",
	},

	{
		"rest/lauscher.go", "bindet nur an",
		"`adresse` ist die Lauschadresse, die der Aufrufer selbst mitgebracht " +
			"hat — sie zu nennen ist die Auskunft, um die es geht",
	},

	{
		"domain/peeradresse.go", "ist keine Adresse",
		"`adresse` ist die Adresse, die ein Mensch **gerade getippt hat** — " +
			"die Meldung sagt, dass sie keine ist, und muss sie dafür zeigen. " +
			"Ein Satz „das ist keine Adresse“ ohne den Wert ließe jemanden vor " +
			"einem Feld sitzen, in dem etwas steht, das er für richtig hält " +
			"(§6: der Fehler erscheint am verursachenden Feld)",
	},

	{
		"domain/lauschwahl.go", "ist keine Adresse — gemeint",
		"`adresse` ist ein Eintrag der Lauschwahl, die ein Mensch **gerade " +
			"gewählt oder getippt hat** — derselbe Grund wie bei `peeradresse.go` (G08, #2272)",
	},

	{
		"domain/lauschwahl.go", "steht zweimal da",
		"`adresse` ist der doppelte Eintrag der Lauschwahl selbst — die Meldung " +
			"nennt, welcher es ist (G08, #2272)",
	},

	// --- Ein Name **im** Paket, kein Ort auf der Platte ------------------
	//
	// Diese vier prüfen, ob eine Zeichenkette ein einfacher Dateiname ist —
	// und melden die, die es nicht ist. Der Wert **ist** die Auskunft; ihn zu
	// kürzen nähme der Meldung ihren Inhalt, und `filepath.Base` wäre hier
	// obendrein sinnlos: geprüft wird gerade, dass er ihm gleicht.

	{
		"app/vorlagenbeigaben.go", "ist kein Dateiname",
		"`datei` ist der geprüfte Name selbst — die Meldung sagt, **dass** er " +
			"keiner ist, und muss ihn dafür zeigen",
	},
	{
		"service/bildausfuhr.go", "ist kein Dateiname",
		"dieselbe Prüfung, an der anderen Schicht",
	},
	{
		"service/exportsetsymbol.go", "das Bild für",
		"`datei` ist der Name einer Beigabe **im Paket** — `beigaben/kosten.png` " +
			"und kein Ort auf der Platte",
	},
	{
		"service/paketwerkstatt.go", "keine Vorlage, sondern Beiwerk",
		"`datei` ist der Name **im Paket** — die Meldung sagt, welche der " +
			"enthaltenen Dateien keine Vorlage ist",
	},

	// --- Ein Paketort, den jemand selbst eingerichtet hat ----------------
	//
	// **Ein Paket liegt dort, wo jemand es hingestellt hat**, und der Ort ist
	// die Auskunft: „welches der installierten Pakete fehlt?" beantwortet nur
	// der ganze Pfad. `filepath.Base` machte aus zwei Wurzeln mit gleichnamigen
	// Paketen dieselbe Meldung — und `msepaket/bildquelle_test.go` unterscheidet
	// genau daran, ob eine **leere** Wurzel durchsucht wurde. Gekürzt wäre der
	// Wächter dort blind (§10).

	{"msepaket/paket.go", "ist nicht zu finden", "der Ort des Pakets, das fehlt"},
	{"msepaket/schreibweg.go", "ist nicht zu finden", "derselbe Ort beim Schreiben"},
	{
		"msepaket/schreibweg.go", "hat sich seit dem Öffnen geändert",
		"derselbe Ort beim Prüfen des Stands",
	},
	{"msepaket/schreibweg.go", "ließ sich nicht lesen", "derselbe Ort beim Lesen"},
	{
		"msepaket/schreibsperre.go", "schreibt gerade jemand anderes",
		"der Ort, an dem die Sperre liegt",
	},
	{
		"msepaket/archivanlegen.go", "trägt keine der sieben Paketendungen",
		"der Ort, an den geschrieben werden soll",
	},
	{
		"msepaket/archivanlegen.go", "fehlt seine Hauptdatei",
		"derselbe Ort, eine Prüfung weiter",
	},
	{
		"service/bezugssperre.go", "schreibt gerade jemand anderes",
		"derselbe Ort wie in `msepaket/schreibsperre.go`, eine Schicht höher",
	},
	{
		"service/partieaustauschdatei.go", "gibt es nicht",
		"der Ort, den ein Mensch gerade gewählt hat (#1711)",
	},

	// --- Ein zweites `%w`, das keinen fremden Text trägt -----------------
	//
	// Diese wickeln einen Fehler aus **diesem** Projekt ein, und dessen Satz
	// ist genau das, was jemand lesen soll: eine Stelle im Skript, eine Zeile
	// in einem Paket, der Name einer Migration. [domain.MitGrund] verstellte
	// hier die Auskunft, statt sie zu schützen.

	{
		"service/skriptkonsole.go", "%w: %w",
		"`fehler` ist ein `msescript`-Fehler und trägt **die Stelle im " +
			"Skript** — sie ist der Inhalt der Konsole",
	},
	{
		"service/kartenskriptlauf.go", "das Kartenskript",
		"derselbe Fall: ein Skriptfehler mit Zeile und Spalte",
	},
	{
		"service/kartenablage.go", "enthält keine Karten",
		"`err` kommt aus `msepaket.Zerteile` und nennt die Zeile in der " +
			"Zwischenablage",
	},
	{
		"service/paketwerkstatt.go", "%w: %w",
		"`lesefehler` kommt aus `msepaket` und benennt die fehlende Datei",
	},
	{
		"storage/sicherungspruefung.go", "%w: %q: %w",
		"`unbekannteAbweisen` nennt **den Namen der Migration**, die diese " +
			"Fassung nicht kennt — ohne ihn bliebe nur, dass etwas nicht stimmt",
	},
	{
		"app/kilaufzeit.go", "liegt nicht auf diesem Rechner",
		"`adresse` ist **keine Datei, sondern die URL, die der Nutzende " +
			"selbst eingestellt hat** — und genau sie ist der Inhalt der " +
			"Meldung: unter »nur lokal« wird dort nicht gefragt. Ohne sie " +
			"bliebe „irgendeine Adresse ist falsch\" (A7, #1446)",
	},
	{
		"scrydex/klient.go", "kein gültiges JSON",
		"`json.Unmarshal` gibt zurück, was ein eigenes `UnmarshalJSON` " +
			"gemeldet hat — `regeltext.go` sagt dort, welches Feld welchen " +
			"Unsinn trug",
	},
}
