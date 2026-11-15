package domain

// Die Einstellungen der Anwendung —
// #245.
//
// **„Eine Einstellung, die es nicht gibt, ist eine Entscheidung, die jemand
// anderes getroffen hat."** Bis hierher führte die Einstellungsseite die
// Sprache und sonst nichts — und die lag im Browserspeicher.
//
// # Warum die Werte hier stehen
//
// Ein Wert, den nur die Oberfläche kennt, kommt ungeprüft in den Bestand. Der
// MCP-Server schreibt auf dieselbe Datei ([0010](../../docs/entscheidungen/0010-mcp-als-eigener-prozess.md));
// was er hineinschreibt, muss dieselben Regeln bestehen. Und ein
// Erscheinungsbild `grau` machte die Anwendung nicht grau, sondern
// unbestimmt — die Klassen dafür gibt es nicht.

// Die Schlüssel, unter denen Einstellungen im Bestand stehen.
//
// **Deutsch, weil es eigene Begriffe sind** ([0015](../../docs/entscheidungen/0015-bezeichner-folgen-dem-vokabular.md)):
// weder MSE noch Wails noch Go kennt sie; sie sind die Wörter dieses
// Programms.
const (
	EinstellungSprache          = "sprache"
	EinstellungErscheinungsbild = "erscheinungsbild"
	EinstellungSchriftgroesse   = "schriftgroesse"

	// EinstellungBildkonflikt ist, was bei einem Namenskonflikt in der
	// Bildausgabe geschieht
	// (#491).
	EinstellungBildkonflikt = "bildkonflikt"

	// EinstellungBildordner ist der zuletzt gewählte Ordner der Bildausgabe —
	// AP-10 (#2367),
	// **ported** aus MSE2: `settings.default_export_dir`
	// (`src/data/settings.hpp`, Zeile 148), nach jeder Wahl fortgeschrieben
	// (`src/gui/images_export_window.cpp`, Zeilen 63–66, Fassung f9d9356 vom
	// 2021-07-06). **Einer je Installation**, wie dort — nicht je Set.
	EinstellungBildordner = "bildausfuhr.ordner"

	// EinstellungAktualisierung ist, ob und wie nach einer neueren Fassung
	// gesehen wird
	// (#1100).
	EinstellungAktualisierung = "aktualisierung"

	// EinstellungErsteSchritte ist, ob die Einführung beim Start erscheint
	// (#1325).
	//
	// **Sie war einmal eine Seitenhälfte.** Die Startseite zeigte *Erste
	// Schritte* **oder** *Dein Überblick*, je nachdem, ob der Bestand ein Set
	// führt — und weil „noch nicht geladen" wie „kein Set" aussah, sprang sie
	// beim Eintreffen der Antwort um. Eine Einführung, die verschwindet,
	// sobald man etwas hat, ist obendrein an dem Tag weg, an dem man sie zum
	// zweiten Mal bräuchte.
	EinstellungErsteSchritte = "ersteschritte"

	// EinstellungSeitenleiste ist, ob die Seitenleiste des Fensters steht —
	// Werkstück U14 von Phase 5.1
	// (#1490).
	// Eine Fensterfrage auf **Geräteebene**: sie reist nicht mit dem Profil.
	EinstellungSeitenleiste = "seitenleiste"

	// EinstellungLeistenlage ist, ob die Leiste links steht oder oben —
	// UX-14.AK-02 (#1686):
	// *„horizontale Anordnung ist ebenfalls möglich"*. Dieselben Einträge,
	// Kürzel und Fokuswege, nur quer. **Auf Geräteebene**, wie die
	// Seitenleiste: ein schmaler Bildschirm will sie oben, ein breiter links.
	EinstellungLeistenlage = "leistenlage"

	// LeisteSenkrecht ist die Leiste links — die Vorgabe, so wie bisher.
	LeisteSenkrecht = "senkrecht"

	// LeisteWaagerecht ist die Leiste oben, über dem Inhalt.
	LeisteWaagerecht = "waagerecht"

	// EinstellungFensterstart ist, wie das Fenster aufgeht — UX-14.AK-01
	// (#1686).
	//
	// **Eine Fensterfrage auf Geräteebene**, wie die Seitenleiste darüber:
	// ein Bildschirm hat eine Größe, und wer sein Profil auf einen anderen
	// Rechner mitnimmt, nimmt keine Fenstergröße mit.
	EinstellungFensterstart = "fensterstart"

	// EinstellungSchriftart ist die **Anzeigeschrift** — die für
	// Überschriften und Setnamen
	// (#1325,
	// Werkstück J).
	//
	// **Nicht die des Fließtexts.** Eine Kalligraphie über einem Absatz ist
	// keine Gestaltung, sondern eine Zumutung; §6 verlangt Lesbarkeit.
	EinstellungSchriftart = "schriftart"

	// EinstellungFarbwelt ist die Palette der Oberfläche — UX-09.AK-01
	// (#1681).
	//
	// **Nicht „Thema"**: so heißt im Fenster schon die Wahl zwischen hell und
	// dunkel (`erscheinungsbild`), und eine Farbwelt hat beide.
	EinstellungFarbwelt = "farbwelt"

	// EinstellungSchriftbereich sagt, ob die Anzeigeschrift nur die
	// Überschriften trägt oder die ganze Oberfläche — UX-09.AK-02.
	EinstellungSchriftbereich = "schriftbereich"

	// EinstellungHintergrund ist, ob und wie das eigene Fensterbild hinter
	// der Oberfläche liegt — UX-09.AK-01. **Das Bild gehört dem Menschen**,
	// wie das Tischbild; die Wahl, ob es erscheint, der Arbeitsumgebung.
	EinstellungHintergrund = "hintergrund"

	// EinstellungGestaltung sind die eigenen Merkmale — UX-09.AK-04: Farben
	// für die zehn Merkmale der Gestaltungssprache und sonst nichts
	// ([PruefeGestaltung]).
	EinstellungGestaltung = "gestaltung"

	// **Die Vorgaben eines Profils** — Werkstück 9 (#1351).
	//
	// VorgabeZustand ist der Zustand, mit dem eine neue Sammlungszeile
	// anfängt (Punkt 3).
	VorgabeZustand = "vorgabe.zustand"

	// VorgabeEinzelspiel sagt, ob ein neues Turnier Einzelspiele führt
	// (Punkt 2).
	VorgabeEinzelspiel = "vorgabe.einzelspiel"

	// VorgabeHochladepfad ist der Ordner, in dem ein Dateidialog aufgeht
	// (Punkt 8) — leer heißt: dort, wo das Betriebssystem ihn öffnet.
	VorgabeHochladepfad = "vorgabe.hochladepfad"

	// **Die lokale KI-Laufzeit** — Werkstück A7 von Phase 5.3
	// (#1446).
	//
	// KILaufzeit sagt, ob überhaupt eine gefragt wird. **Sie fängt aus an**:
	// eine Anwendung, die beim Start an einen Port klopft, den der Nutzende
	// nie genannt hat, tut etwas, um das niemand gebeten hat.
	KILaufzeit = "ki.laufzeit"

	// KIAdresse ist, wo sie lauscht — **leer heißt: die Schleifenadresse**.
	//
	// **Getrennt von der Frage darüber**, weil es zwei Fragen sind: „darf
	// gefragt werden" gehört dem Profil, „wo lauscht sie" diesem Rechner.
	// Und eine Adresse, die **nicht** auf der Schleife liegt, stuft jedes
	// Modell dahinter als nicht-lokal ein — das Issue verlangt für ein
	// entferntes Ollama „ausdrückliche Konfiguration und eindeutige
	// Einstufung".
	KIAdresse = "ki.adresse"

	// **Apples Gerätemodell** — AI-05
	// (#2042,
	// Vermerk 0166).
	//
	// KIGeraetemodell sagt, ob es für kleine Textaufgaben gefragt wird. **Es
	// fängt automatisch an** — anders als [KILaufzeit]: dort klopfte eine
	// Anwendung an einen Port, den niemand genannt hat. Hier gibt es keinen
	// Port (ein eigener Unix-Socket), das Modell gehört zum Betriebssystem,
	// und die Annahme seiner Lizenz ist eine eigene Handlung des Nutzers.
	KIGeraetemodell = "ki.geraetemodell"

	// Kontoschau sagt, ob dieses Programm das **eigene** Konto bei einem
	// Marktplatz lesen darf — Werkstück T11 von Phase 5.6
	// (#1476).
	//
	// # Warum es dafür einen zweiten Schalter gibt
	//
	// **Weil der erste eine andere Frage beantwortet.** Wer CardTrader
	// einschaltet, hat den **Marktplatz** gemeint: was fremde Verkäufer für
	// eine Karte verlangen. Daraus folgt nicht, dass dieses Programm auch
	// die **eigenen** Angebote und die **eigenen** Bestellungen lesen soll —
	// das sind andere Daten mit anderen Folgen, und das Issue verlangt
	// ausdrücklich ein „standardmäßig deaktiviertes Folgepaket".
	//
	// **Sie fängt aus an**, und dieselbe Marke bleibt es trotzdem: ein
	// zweiter Schlüsselbundeintrag für dasselbe Geheimnis wäre eine zweite
	// Stelle, an der es veralten kann (0122).
	Kontoschau = "konto.schau"
)

// Die zwei Stellungen der Kontoschau — dieselben Worte wie bei der
// KI-Laufzeit, damit niemand zwei Schreibweisen für dasselbe lernt.
const (
	KontoschauAus = "aus"
	KontoschauAn  = "an"
)

// Die zwei Stellungen der KI-Laufzeit.
const (
	KILaufzeitAus = "aus"
	KILaufzeitAn  = "an"

	// Die Werte von [KIGeraetemodell].
	KIGeraetemodellAutomatisch = "automatisch"
	KIGeraetemodellAus         = "aus"
)

// einstellung ist eine Einstellung mit allem, was zu ihr gehört.
//
// **Eine Liste und nicht vier Abbildungen.** Hier standen `einstellungsbereiche`,
// `einstellungswerte`, `einstellungsvorgaben` und `einstellungsreihe` — vier
// Aufzählungen über demselben Schlüsselvorrat. Der Zweitblick über Phase G
// fand, was **drei** davon kosten: `bildkonflikt` fiel an der Dienstnaht
// heraus, die Antwort trug ihn mit leerem Wert, und ein ausdrücklich gewähltes
// `altbehalten` wurde stillschweigend übergangen.
//
// **Der Nachlauf fand, dass die Korrektur ihre Klasse nicht traf**: der
// Wächter darüber zählte dieselben vier Schlüssel noch einmal auf, und ein
// fünfter nur in `einstellungswerte` blieb wieder unbemerkt — nachgestellt und
// gesehen. Jetzt ist er nicht mehr **darstellbar**: eine Einstellung ist eine
// Zeile, und eine halbe Zeile gibt es nicht.
type einstellung struct {
	Schluessel string
	Bereich    string

	// Werte ist, was erlaubt ist — **in Anzeigereihenfolge**. Sie steht hier
	// und nicht im Fenster, weil sonst zwei Listen dieselbe Frage
	// beantworteten und die im Fenster die ungeprüfte wäre.
	Werte []string

	// Pruefung ist die Regel, die eine Werteliste nicht ausdrücken kann —
	// `nil`, wo es keine gibt.
	//
	// **Für Werte, die sich nicht aufzählen lassen**: eine Spaltenbreite
	// entsteht am Ziehgriff, eine Kostengrenze am Taschenrechner. Ohne sie
	// gäbe es für beide nur „Liste oder alles", und „alles" hieße, dass ein
	// Fenster mit einer Spalte von zwei Punkten Breite eine gültige
	// Einstellung ist.
	Pruefung func(wert string) error

	// Wirkung sagt, **wann** ein geänderter Wert gilt — [WirkungSofort] oder
	// [WirkungNeustart] (UX-04.AK-02).
	//
	// **Leer heißt „nicht gesagt" und ist ein Fehler**, kein Rückfall:
	// `TestJedeEinstellungSagtIhreWirkung` macht die Reihe rot, sobald eine
	// Einstellung ihn leer lässt. Ohne den Lauf behauptete jede neue
	// stillschweigend „sofort", denn der Nullwert einer Zeichenkette ist leer.
	Wirkung string

	// Vorgabe ist, was ohne Wahl gilt — leer heißt: keine.
	//
	// **Der zurückhaltendste Wert gewinnt**: `system` und `normal` heißen
	// „nimm, was der Rechner sagt". Die Sprache hat **keine** — sie hängt an
	// `navigator.language`, und das weiß nur das Fenster. Hier eine zu setzen
	// hieße, die Systemsprache zu überschreiben, sobald jemand die Anwendung
	// zum ersten Mal öffnet.
	Vorgabe string
}

// OhneVorgabe heißt: das Feld bleibt leer, wie es immer war.
//
// **Es ist die Vorgabe und kein Zustand.** Ein frischer Bestand verhält sich,
// wie er sich immer verhalten hat — eine Vorgabe, die beim Einbau etwas
// ändert, ändert es für jeden, der nie danach gefragt wurde.
const OhneVorgabe = "ohne"

// zustandswerte gibt »ohne« und dahinter jeden Zustand.
//
// **Aus [Zustaende] und nicht abgeschrieben**: die Liste steht in
// `exemplar.go`, und eine zweite hier ließe beim nächsten Zustand eine Wahl
// zu, die der Bestand abweist.
func zustandswerte() []string {
	werte := make([]string, 0, len(Zustaende())+1)
	werte = append(werte, OhneVorgabe)

	for _, zustand := range Zustaende() {
		werte = append(werte, string(zustand))
	}

	return werte
}

// einstellungen ist die Reihe, in der die Einstellungen stehen.
//
