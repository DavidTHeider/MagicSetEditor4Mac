package app

// **Die Fehlercodes** — aus `fehler.go` herausgelöst, als D02 (#2233) die
// Datei über die Zeilengrenze aus §2 gebracht hätte. Der Wächter
// `nahtfehlervollstaendig_test.go` liest sie hier.

// Fehlercode ist die grobe Einordnung, an der das Frontend sein Verhalten
// ausrichtet: eine Validierung gehört ans Feld, ein Konflikt braucht ein
// Nachladen, ein Unbekannter einen Hinweis.
type Fehlercode string

const (
	CodeValidierung   Fehlercode = "validierung"
	CodeNichtGefunden Fehlercode = "nichtGefunden"
	CodeKonflikt      Fehlercode = "konflikt"
	CodeUnbekannt     Fehlercode = "unbekannt"

	// CodeBelegt heißt: ein anderer schreibt gerade am selben Bestand.
	//
	// **Nicht [CodeKonflikt], obwohl beides nach „jemand war schneller"
	// klingt.** Auf einen Konflikt lädt das Frontend nach (§7, Regel 3), und
	// hier liefe das Nachladen in dieselbe Sperre. Der Rat ist gleich und
	// nochmal (#886).
	CodeBelegt Fehlercode = "belegt"

	// CodeAbgebrochen heißt: der Lauf wurde **absichtlich** beendet, weil ein
	// neuer ihn abgelöst hat (§7).
	//
	// **Das ist kein Fehler, und die Oberfläche zeigt ihn nicht.** Über der
	// gezeichneten Karte stand „Es ist ein unerwarteter Fehler aufgetreten" —
	// die Meldung des abgebrochenen Laufs, der genau so abgebrochen wurde, wie
	// es vorgesehen ist
	// (#288).
	CodeAbgebrochen Fehlercode = "abgebrochen"

	// CodeErlaubnis heißt: der Nachweis stimmt, er reicht nur nicht so weit —
	// Werkstück 11 von Phase 4.12
	// (#1351).
	//
	// **Nicht [CodeValidierung], und der Unterschied ist der ganze Punkt.**
	// Der heißt „ändere deine Eingabe" und stellt dasselbe Formular noch
	// einmal hin. Hier ist die Eingabe richtig; was fehlt, ist eine Rolle —
	// und die bekommt man nicht, indem man etwas anders tippt.
	//
	// **Auch nicht [CodeNichtGefunden].** „Gibt es nicht" wäre die
	// vorsichtigere Auskunft, aber hier ist sie eine Lüge: das Profil steht in
	// derselben Übersicht, aus der die Kennung stammt.
	CodeErlaubnis Fehlercode = "erlaubnis"

	// CodeSchluesselFehlt heißt: der Bestand ist verschlüsselt, und sein
	// Schlüssel liegt nicht im Schlüsselbund — Werkstück 14 von Phase 4.12
	// ([0072](../../docs/entscheidungen/0072-ein-treiber-der-verschluesseln-kann.md)).
	//
	// **Nicht [CodeUnbekannt], und daran hängen Daten.** Der Rückfallsatz
	// lautet „Es ist ein unerwarteter Fehler aufgetreten" — wer das über einer
	// Datei liest, die sich nicht öffnen lässt, ersetzt sie. Sie ist in
	// Ordnung; es fehlt der Schlüssel, und ein Ersetzen macht den Verlust
	// endgültig.
	//
	// **Auch nicht [CodeNichtGefunden].** Gefunden ist der Bestand ja.
	CodeSchluesselFehlt Fehlercode = "schluesselFehlt"

	// CodeNurLesend heißt: in diesem Profil darf nur gelesen werden —
	// Werkstück 16 von Phase 4.12
	// ([0073](../../docs/entscheidungen/0073-ein-leser-darf-wirklich-lesen.md)).
	//
	// **Nicht [CodeErlaubnis], obwohl beides von einer Rolle kommt.** Der
	// heißt „dieser Ruf geht nicht" und steht an einer einzelnen Handlung;
	// dieser gilt für **alles**, solange der Zeiger dort steht. Die Oberfläche
	// tut daraufhin etwas anderes: nicht das Formular noch einmal zeigen,
	// sondern sagen, dass hier niemand schreibt.
	//
	// **Und nicht [CodeUnbekannt]**: dessen Rückfallsatz wäre über einem
	// Speichern-Knopf, der nichts tut, die unbrauchbarste aller Auskünfte.
	CodeNurLesend Fehlercode = "nurLesend"

	// CodeVoll heißt: auf dem Datenträger ist kein Platz mehr — #1812.
	//
	// **Nicht [CodeUnbekannt]**: dessen Satz schickt ins Protokoll, und der
	// Rat ist hier klar. **Nicht [CodeBelegt]**: der vergeht von allein,
	// dieser erst, wenn jemand Platz schafft. Neu und damit erweiternd (§16).
	CodeVoll Fehlercode = "voll"

	// CodeAbgeschaltet heißt: der Weg hängt an einem abgeschalteten Bereich
	// oder einer Funktion — OPS-10 (#1908). Neu und damit erweiternd (§16).
	CodeAbgeschaltet Fehlercode = "abgeschaltet"

	// CodeNetz heißt: der Fehler liegt **außerhalb** dieses Programms — an
	// einem fremden Dienst oder am Weg dorthin (D02, #2233).
	//
	// **Nicht [CodeUnbekannt]**: dessen Satz schickt ins Protokoll, und hier
	// ist der Rat klar — Verbindung, Proxy, später noch einmal. **Nicht
	// [CodeBelegt]**: der spricht vom eigenen Bestand. Der Schlüssel sagt,
	// welcher der Fälle aus `domain/gegenstelle.go` es war. Neu und damit
	// erweiternd (§16).
	CodeNetz Fehlercode = "netz"

	// CodeDrossel heißt: der Aufrufer liegt über **seinem eigenen** Budget an
	// dieser Anwendung — anders als [CodeNetz], wo eine fremde Gegenstelle
	// drosselt. Der Dienst nennt es seit #1429 so; MCP und Kommandozeile
	// tragen es seit dem Zweitblick über Phase 1 von 0.9.0 (Befund 5).
	CodeDrossel Fehlercode = "drossel"
)
