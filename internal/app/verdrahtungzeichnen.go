package app

// **Die Verdrahtung des Zeichenwerks**
// (#886).
//
// **Eine eigene Datei**, weil `verdrahtung.go` mit der Naht des Suchindex
// über die Zeilengrenze aus §2 ging. Der Schnitt folgt der Frage: dort hängt,
// **was der Bestand kann** — Sets, Karten, Decks, Nutzer —, hier, **was eine
// Karte zu einem Bild macht**.

import (
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// zeichenteile sind die Bausteine, die das Zeichenwerk von der Verdrahtung
// bekommt.
//
// **Ein Verbund und keine vier Parameter**: sie gehören zusammen, und eine
// Reihe von vier gleichartigen Zeigern ist die Gestalt, in der jemand zwei
// davon vertauscht.
type zeichenteile struct {
	db          *sql.DB
	leser       *storage.SetLeser
	wurzeln     []string
	ablage      *service.Bildablage
	stilwaehler *service.Stilwaehler
}

// zeichenwerkVerdrahten hängt Stile, Plan, Rasterung und Ausgabe ein — und
// gibt das Werk zurück, damit die Wurzel den Feldweg darüber rechnen lässt.
func (a *App) zeichenwerkVerdrahten(teile zeichenteile) *Zeichenwerk {
	db, leser, wurzeln := teile.db, teile.leser, teile.wurzeln
	ablage, stilwaehler := teile.ablage, teile.stilwaehler
	// **Mit den Stilen** (#206): bis hierher rief jeder Ausgabeweg den
	// Behelfsplan, und alles, was #120, #173 und #186 gebaut haben, war in der
	// laufenden Anwendung nicht zu sehen. Ohne Paket bleibt es beim Behelf —
	// aber mit einem Befund, der sagt, warum.
	//
	// **Das Zeichenwerk setzt zusammen, was eine Karte zu einem Bild macht**
	// (#886).
	//
	// Teile, die aneinanderhängen und **dieselben** sein müssen: zwei
	// Schriftsuchen wären zwei Antworten auf die Frage, wie breit ein Text
	// ist. Die Kommandozeile braucht dieselbe Zusammensetzung; sie zweimal
	// hinzuschreiben hieße, die Gründe zweimal zu haben.
	werk := NeuesZeichenwerk(leser, wurzeln, ablage.Wurzel()).MitPruefer(a.autorisierung)

	// **Der Stilwähler bekommt den Speicher des Werks** — nicht einen
	// zweiten: er sagt ihm, dass ein Wechsel sein Gedächtnis ungültig macht
	// (#693).
	stilwaehler.MitSpeicher(werk.Stile)

	// **Die Zusatzfelder einer Karte** (#2298) — über den Stil **der Karte**;
	// geschrieben wird der Block `extra data` auf den gelesenen Stand der
	// Karte (Befund 1 des Nachlaufs über E/F), mit Verlauf wie jedes Feld.
	a.zusatzwerte = service.NeueZusatzwertpflege(leser, werk.Stile, storage.NeuerKartenschreiber(db))

	// **Und die Schlüsselwörter des Sets** (#1062). Sie brauchen den
	// **Stilspeicher**, weil die des Spiels aus dem Spielpaket kommen — und
	// welches Paket das ist, sagt der Stil des Sets. Deshalb steht dieser Ruf
	// hier und nicht oben: den Speicher gibt es erst mit dem Zeichenwerk.
	// **Und die Werkzeugnähte** (#1062, #1070) — in `werkzeugnaehte.go`, weil
	// `app.go` sonst über die Zeilengrenze ginge (§2).
	werkzeugnaehte(a, db, leser, werk)

	// **Die Sammlung und die Preise** (#1165) — sie hängen am Bestand und
	// nicht an einem Set.
	sammlungsnaehte(a, db)

	// **Und das Löschen eines Sets** (#1218). Über dieselbe Naht wie die
	// Flächen ohne Fenster: was zusammengesetzt wird, gehört an eine Stelle.
	a.MitSetloeschung(SetloeschnahtUeber(db))

	a.planbauer = werk.Planbauer
	a.rasterer = werk.Rasterer

	// **Und die Problemprüfung der Sets** (AP-8, #2365) — über **demselben**
	// Zeichenwerk: geprüft wird, wie die Vorschau zeichnet.
	a.MitProblemweg(ProblemwegAus(werk.Problempruefung()))

	// **Und die Spieltexturen** (G2,
	// #1459) —
	// über **demselben** Planbauer und Rasterer. Zwei wären zwei Antworten auf
	// die Frage, wie eine Karte aussieht, und dann zeigte der Tisch etwas
	// anderes als die Vorschau.
	a.texturen = werk.Spielbilder
	a.spielbilder = werk.Spielbildfach
	a.vorlagen = werk.Vorlagen
	a.bildausfuhr = werk.Bildausfuhr

	// **Und das Teilen** (NET-04, #1708) — über dieselbe Bildausfuhr und
	// denselben Setschreiber; der Dialog kommt vom Fenster.
	a.teilenEinrichten(teilaufraeumenErlaubt)
	a.bogenleger = werk.Bogenleger

	// **Der Abzug des Tisches braucht nichts aus dem Werk** (PLAY-05.AK-03,
	// #1713): er bekommt ein fertiges Bild und schreibt es. Er steht hier,
	// weil eine Naht ohne Einsetzer beim ersten Aufruf ein `nil` wäre.
	a.tischabzug = tischabzugDienst()
	a.partiedatei = partiedateiDienst()
	a.turnierdatei = service.NeueTurnieraustauschdatei()

	// **Die Symbolzeichnungen kommen aus derselben Bildquelle** (#660): sie
	// liest `.mse-symbol` aus jedem Paket, auch aus dem Set.
	a.symbolzeichnungen = werk.Bildquelle

	// **Und die Vorschaubilder der Kartenliste ebenso** (#1166) — dieselbe
	// Quelle, die auch die Karte zeichnet. Zwei wären zwei Antworten auf die
	// Frage, wo das Bild einer Karte liegt.
	a.vorschaubilder = service.NeuesVorschaubildwerk(werk.Bildquelle)

	// **Und seine drei Zwischenspeicher zeigen sich** — OPS-08 (#1725).
	a.speicherquellen, a.speicherleerer = werkspeicher(werk)

	// **Und die Diagnose bekommt den Schemastand** (#1211, Befund 2): ein
	// Fehler nach einer Aktualisierung ist ohne ihn nicht einzuordnen.
	a.MitSchemastand(SchemastandUeber(db))

	// **Der Betriebszustand hängt am Bestand, den das Fenster geöffnet hat**
	// — Werkstück O9 von Phase 5.2 (#1433). **An `db` und nicht an `arbeit`**:
	// gefragt wird, ob **der Bestand** antwortet, und das ist eine Aussage
	// über die Datei, nicht über das gezeigte Profil.
	//
	// **Ohne Fremdschau**: was über fremde Dienste bekannt ist, führt der
	// Dienstkatalog (A1), und ihn hier zu fragen hieße, eine Zustandsabfrage
	// an eine Bestandsabfrage zu hängen. Wer die Auskunft will, ruft den
	// Katalog.
	a.MitGesundheit(service.NeueBetriebsschau(storage.NeueBetriebslage(db), nil))

	return werk
}
