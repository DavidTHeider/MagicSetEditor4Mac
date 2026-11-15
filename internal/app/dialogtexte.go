package app

// **Die Beschriftungen der nativen Dialoge** — aus `menuetexte.go`
// herausgelöst, als die Menüs von macOS in die Sprache des Fensters kamen
// (#255) und die Datei über die Zeilengrenze aus §2 wuchs.
//
// **Der Schnitt folgt der Frage**: dort steht, wie die Menüleiste heißt; hier,
// was ein Dialog fragt und was er zeigen darf. Die Sätze selbst stehen in
// `dialogkataloge.go`.

// Dialogtexte sind die Beschriftungen der nativen Dialoge.
//
// **Zweisprachig wie alles andere** (§9). Der erste Anlauf des Dateidialogs trug
// „Set öffnen" und „MSE-Sets (*.mse-set)" als deutsche Zeichenketten im
// Quelltext — dieselbe Lücke, die `menue.go` eine ganze Phase lang hatte, und
// derselbe Grund: kein Wächter sieht einen nativen Text.
type Dialogtexte struct {
	SetOeffnen   string
	SetFilter    string
	SetSpeichern string
	SetVorschlag string

	// Die vier für den portablen Umschlag (N4, #1457).
	//
	// **Vier und nicht drei**: beim Umschlag gibt es einen Weg hinaus *und*
	// einen herein, und beide nehmen denselben Filter.
	UmschlagSchnueren string
	UmschlagOeffnen   string
	UmschlagFilter    string

	// Uebernommen hängt an einem angenommenen Deck, dessen Name hier schon
	// vergeben ist (T7, #1472).
	//
	// **Sichtbar und nicht still.** Ein Deck, das unter fremdem Namen
	// hereinkäme, ließe zwei gleich heißende nebeneinanderstehen; eines
	// stillschweigend umzubenennen wäre schlimmer. Der Zusatz sagt, was
	// geschehen ist.
	Uebernommen       string
	UmschlagVorschlag string

	// Die zwei für das Ziel einer Leihe (UXR-F15,
	// #2030):
	// der Titel des Sichern-Dialogs und der Name, den er vorschlägt.
	//
	// **Eigene Sätze und nicht die des Sets daneben.** „Set als Umschlag
	// ablegen" fragte nach einem Set, wo eine Leihe hinausgeht. Der Filter
	// ist der des Umschlags — dieselbe Endung, dieselbe Sache.
	LeiheAblegen   string
	LeiheVorschlag string

	// Die drei für die Sicherung des Bestands (#799) und der vierte für den
	// Weg zurück (#940).
	BestandSpeichern string
	BestandFilter    string
	BestandVorschlag string
	BestandOeffnen   string

	// Die zwei für eine vCard-Datei (ID-04, #1703).
	KontakteOeffnen string
	KontakteFilter  string

	// Die drei für die Diagnose (#1166, Werkstück 4).
	DiagnoseSpeichern string
	DiagnoseFilter    string
	DiagnoseVorschlag string

	// Die drei für das Diagnosepaket (O10, #1434).
	DiagnosepaketSpeichern string
	DiagnosepaketVorschlag string
	DiagnosepaketFilter    string

	// Die Aufschriften **im** Diagnosetext (#1215, Befund 4).
	//
	// # Warum sie hierher gehören
	//
	// Der Text kommt aus Go und geht durch keinen der fünf Sprachkataloge —
	// `keintextimjsx` sieht ihn nicht, und die englische Oberfläche zeigte
	// deutsche Bezeichner mitten in dem Bereich, der für wenig technische
	// Menschen gebaut ist. Diese Struktur ist die Stelle, an der Go schon
	// zweisprachig ist.
	DiagnoseBestand      string
	DiagnoseProtokollort string

	// Die Kanäle des Protokolls — OPS-04 (#1721).
	DiagnoseKanaele      string
	DiagnoseKanalDatei   string
	DiagnoseKanalSystem  string
	DiagnoseKanalAusgabe string
	DiagnoseKanalAn      string
	DiagnoseKanalAus     string
	DiagnoseKanalFehlt   string
	DiagnoseStufe        string
	DiagnoseAusfuehrlich string

	// DiagnoseProtokollVerdeckt steht anstelle der Protokollzeilen, wenn die
	// Rolle nicht bis dorthin reicht (O9, #1433).
	DiagnoseProtokollVerdeckt string

	// DiagnoseTeilUnbekannt steht bei einem Teil, für den es keinen Zweig
	// gibt — er fällt auf, statt still leer zu bleiben (O10, #1434).
	DiagnoseTeilUnbekannt string
	DiagnoseSchema        string

	// DiagnoseBau ist die Aufschrift der Zeile mit dem Stand des Baus,
	// DiagnoseVeraendert der Vermerk, wenn der Baum dabei nicht sauber war,
	// und DiagnoseBindungsvertrag die Zeile mit dem Abdruck des Vertrags
	// zwischen Fenster und Go (ENG-06.AK-04, #1731).
	DiagnoseBau             string
	DiagnoseVeraendert      string
	DiagnoseBindungsvertrag string

	// DiagnoseWaisen ist die Aufschrift der Zeile über verwaiste Zeilen im
	// Bestand (#1220).
	DiagnoseWaisen          string
	DiagnoseMigrationen     string
	DiagnoseKeineSchema     string
	DiagnoseNichtLesbar     string
	DiagnosePakete          string
	DiagnoseWeitere         string
	DiagnoseKeinBestand     string
	DiagnoseKeinProtokoll   string
	DiagnoseKeineVerwaltung string

	// BildOeffnen und BildFilter sind der Dialog für ein Kartenbild
	// (#246).
	//
	// **Ohne BMP und TIFF**, die MSE anbietet (`src/gui/value/image.cpp`,
	// Zeile 23, Fassung f9d9356 vom 2021-07-06): die Standardbibliothek liest
	// beide nicht, und ein Filter, der
	// eine Datei zulässt, die danach abgewiesen wird, ist schlechter als
	// keiner.
	BildOeffnen string
	BildFilter  string

	// PaketdateiEinfuegen und PaketdateiFilter sind der Dialog, der eine
	// Datei in ein Paket der Werkstatt legt (FMT-02,
	// #1691):
	// Bilder und Schriften — das Beiwerk, das sich nicht als Text anlegen
	// lässt.
	PaketdateiEinfuegen string
	PaketdateiFilter    string

	// PaketWeitergeben, InstallerEinlesen und InstallerFilter sind die
	// Dialoge, über die ein Paket als `.mse-installer` hinausgeht und
	// hereinkommt (FMT-02.AK-04, #1691).
	PaketWeitergeben  string
	InstallerEinlesen string
	InstallerFilter   string

	// Die drei für den Abzug eines Tisches (PLAY-05.AK-03,
	// #1713).
	//
	// **Nur PNG**: ein Tischbild trägt Flächen und scharfe Kanten, und JPEG
	// setzt an jede davon einen Schleier. Ein Filter, der etwas zulässt, was
	// dieser Weg nicht schreibt, ist schlechter als keiner — derselbe Satz
	// wie beim Kartenbild darüber.
	TischabzugSpeichern string
	TischabzugFilter    string
	TischabzugVorschlag string

	// Und die drei für dasselbe Bild auf einem Blatt Papier.
	TischabzugPDFSpeichern string
	TischabzugPDFFilter    string
	TischabzugPDFVorschlag string

	// Die vier für die Partiedatei (PLAY-03.AK-05, #1711).
	PartieSpeichern string
	PartieOeffnen   string
	PartieFilter    string
	PartieVorschlag string

	// Die vier für die Turnierdatei und die drei für ihren Bericht
	// (DOM-04.AK-03, #1699).
	TurnierSpeichern string
	TurnierOeffnen   string
	TurnierFilter    string
	TurnierVorschlag string
	BerichtSpeichern string
	BerichtFilter    string
	BerichtVorschlag string

	// Die vier für ein Symbol der Bibliothek (DOM-02, #1697).
	SymbolAusgeben  string
	SymbolEinlesen  string
	SymbolFilter    string
	SymbolVorschlag string

	// Die drei für eine eigene Übersetzung (LOC-02, #1688) — der Vorschlag
	// ist ihre Kennung und wird nicht übersetzt.
	SprachpaketAusgeben string
	SprachpaketEinlesen string
	SprachpaketFilter   string

	MTGJSONDatei       string
	MTGJSONFilter      string
	MTGJSONVerzeichnis string
	MTGJSONSchreiben   string
	MTGJSONVorschlag   string

	VorlageSchreiben string
	VorlageFilter    string

	// **Die Auskunft über einen Menschen** — Werkstück 8 von Phase 4.12
	// (#1351). `JSONFilter` heißt nicht `MTGJSONFilter`: dieselbe Endung,
	// eine andere Sache.
	PersonendatenSichern string
	JSONFilter           string

	// **Der Austausch der Einstellungen** — Werkstück O3 von Phase 5.2
	// (#1427). Er nimmt `JSONFilter` mit: dieselbe Endung, dieselbe Sache.
	EinstellungenSichern   string
	EinstellungenOeffnen   string
	EinstellungenVorschlag string

	Bildordner     string
	BogenSchreiben string
	BogenVorschlag string
	PDFFilter      string

	PaketWaehlen string
	MSE2Waehlen  string

	// Planordner ist der Titel, unter dem der Ordner des freiwilligen Plans
	// gewählt wird (OPS01, #2347).
	Planordner string
}

// Dateiauswahl beschreibt einen Dialog, ohne einen zu kennen.
//
// **Der Aufrufer sagt, wonach gefragt wird**, nicht der Dialog. Sonst stünde in
// `dateiwahl.go` — der einzigen Datei, die kein Test ausführt — eine Fallunter-
// scheidung darüber, welches Format gerade geöffnet wird.
type Dateiauswahl struct {
	Titel  string
	Filter string
	Muster string

	// Verzeichnis wählt den Ordner- statt den Dateidialog.
	Verzeichnis bool

	// Speichern wählt den Sicherndialog statt des Öffnendialogs.
	Speichern bool

	// Vorschlag ist der Name, den der Sicherndialog vorbelegt.
	Vorschlag string

	// Startverzeichnis ist der Ordner, in dem der Dialog aufgeht — Werkstück
	// 9 von Phase 4.12 (#1351), Punkt 8.
	//
	// **Leer heißt: dort, wo das Betriebssystem ihn öffnet.** Gesetzt wird er
	// nicht hier, sondern vom Mantel aus `hochladepfad.go` — und **nur beim
	// Öffnen**: wohin etwas hinausgeht, ist eine andere Frage.
	Startverzeichnis string

	// Ablage sagt, welcher eingestellte Ordner gilt — leer heißt: beim
	// Öffnen der fürs Hereinholen, beim Sichern der fürs Hinausgeben
	// (OPS-02, #1719). [AblageSicherung] und [AblageDiagnose] haben eigene.
	Ablage string
}

// Die Ablagen mit eigenem Ordner — und seit AP-10 (#2367) die der Bilder.
const (
	AblageSicherung = "sicherung"
	AblageDiagnose  = "diagnose"
	AblageBilder    = "bilder"
)

// DialogtexteFuer gibt den Katalog einer Sprache — oder den deutschen.
func DialogtexteFuer(sprache string) Dialogtexte {
	if texte, da := dialogkataloge[sprache]; da {
		return texte
	}

	return dialogkataloge["de"]
}
