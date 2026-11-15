package werkzeuge_test

// **Wie ein Schlüssel zu den Namen seiner Werte kommt** — die Aufzählung
// hinter `TestJederEinstellungswertHatEinenNamen`
// (`einstellungstexte_test.go`).
//
// **Eine eigene Datei, weil sie die andere über die Zeilengrenze aus §2
// trug**, und weil sie eine andere Frage beantwortet: dort, *ob* jeder Wert
// einen Namen hat, hier, *woher* er ihn bekommt. Wer eine Einstellung
// aufnimmt, trägt sie hier ein.

import (
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// weg sagt, woher ein Schlüssel die Namen seiner Werte bekommt.
type weg int

const (
	// ausKatalog: je Wert eine Zeile `'<schluessel>.<wert>':` unter `werte`.
	ausKatalog weg = iota

	// ausSchalterpaar: die geteilten `dienstAus`/`dienstAn` — sie sind für
	// jeden Dienst dieselben, und je Dienst zwei Zeilen sagten alle dasselbe.
	ausSchalterpaar

	// ausEigennamen: `frontend/src/bausteine/dienstnamen.ts`. Eigennamen
	// werden nicht übersetzt (0063, Entscheidung 5) und stehen deshalb in
	// keinem Sprachkatalog.
	ausEigennamen

	// ausEigenerSeite: der Schlüssel wird nicht auf der Einstellungsseite
	// gezeigt. `bildkonflikt` steht in `Bildereinstellungen.tsx`, und dafür
	// gibt es `ausgabeformate_test.go` daneben.
	ausEigenerSeite
)

// schalternamen ordnet jedem Schalterwert seinen Katalogschlüssel zu.
//
// **Die Karte, die `Dienstschalter` aufspannt** (`werteNamen={werte}`): dort
// steht `{ aus: eigene.dienstAus, an: eigene.dienstAn }`. Wer in `domain`
// einen dritten Zustand aufnimmt, muss ihn hier und dort eintragen — sonst
// zeichnet das Fenster wieder eine Auswahl ohne Text.
var schalternamen = map[string]string{
	domain.DienstAus: "dienstAus",
	domain.DienstAn:  "dienstAn",
}

// namensweg ordnet jedem Schlüssel seinen Weg zu.
//
// **Eine Aufzählung und keine Regel über Präfixe.** Wer eine Einstellung
// aufnimmt, entscheidet hier, wie ihre Werte zu Namen kommen — der Lauf unten
// wird sonst rot und nennt den Schlüssel.
func namensweg(schluessel string) (weg, bool) {
	fest := map[string]weg{
		domain.EinstellungSprache: ausKatalog,
		// **Die eigene Übersetzung** (LOC-02, #1688) — ohne Werteliste; ihre
		// Wahl steht über der Liste der Pakete, ihr Name im Katalog.
		domain.EinstellungEigeneSprache: ausKatalog,

		// **Darstellung und Zeitzone** — LOC-03.AK-03
		// (#1689).
		// Beide haben **keine** Werteliste; die Schleife unten läuft über
		// null Werte und prüft damit nur ihren Namen, wie beim Hochladepfad.
		domain.EinstellungDarstellung: ausKatalog,
		domain.EinstellungZeitzone:    ausKatalog,

		// **Die Belegung der Ziffern** — UX-02.AK-03
		// (#1674).
		// Ohne Werteliste; ihre Tafel steht bei den Kürzeln, ihr Name im
		// Katalog wie bei den zwei darüber.
		domain.EinstellungZiffernbelegung:    ausKatalog,
		domain.EinstellungGespeicherteFilter: ausKatalog,

		domain.EinstellungErscheinungsbild: ausKatalog,
		domain.EinstellungSchriftgroesse:   ausKatalog,
		domain.EinstellungAktualisierung:   ausKatalog,
		domain.EinstellungErsteSchritte:    ausKatalog,
		domain.EinstellungSeitenleiste:     ausKatalog,
		domain.EinstellungLeistenlage:      ausKatalog,
		domain.EinstellungFarbwelt:         ausKatalog,
		domain.EinstellungSchriftbereich:   ausKatalog,
		domain.EinstellungHintergrund:      ausKatalog,
		domain.EinstellungGestaltung:       ausKatalog,

		// **Wie das Fenster aufgeht** — UX-14.AK-01
		// (#1686).
		// Die zwei Maße dahinter haben **keine** Werteliste — eine
		// Fenstergröße entsteht am Fensterrand; die Schleife unten läuft
		// über null Werte und prüft damit nur ihren Namen, wie beim
		// Hochladepfad.
		domain.EinstellungFensterstart:  ausKatalog,
		domain.EinstellungFensterbreite: ausKatalog,
		domain.EinstellungFensterhoehe:  ausKatalog,
		domain.EinstellungSchriftart:    ausKatalog,
		domain.EinstellungBildkonflikt:  ausEigenerSeite,

		// **Wie viel vom Kartenformular dasteht** — UX-05.AK-01
		// (#1677).
		domain.EinstellungFeldtiefe: ausKatalog,

		// **Ob der Tisch selbst ordnet** — PLAY-04.AK-03
		// (#1712).
		// Zwei Werte, beide aus demselben Katalog wie die Tiefe darüber.
		domain.EinstellungAutoanordnung: ausKatalog,

		// **Wie der Tisch beschriftet ist** — vier Stufen, aus demselben
		// Katalog wie die Anordnung darüber.
		domain.EinstellungTischbeschriftung: ausKatalog,

		// **Die Vorgaben eines Profils** — Werkstück 9 von Phase 4.12
		// (#1351), Punkte 2, 3 und 8. Sie stehen im Katalog der
		// Einstellungsseite wie alles darüber.
		domain.VorgabeZustand:     ausKatalog,
		domain.VorgabeWaehrung:    ausKatalog,
		domain.VorgabeEinzelspiel: ausKatalog,

		// **Der Hochladepfad hat keine Werte** — er ist ein Verzeichnis, und
		// ein Verzeichnis lässt sich nicht aufzählen. Sein **Name** steht
		// trotzdem im Katalog; die Schleife unten läuft über null Werte und
		// prüft damit nur ihn.
		domain.VorgabeHochladepfad: ausKatalog,

		// **Und die drei Ordner aus OPS-02** (#1719) — dieselbe Gestalt.
		domain.VorgabeAusfuhrpfad:    ausKatalog,
		domain.VorgabeSicherungspfad: ausKatalog,
		domain.VorgabeDiagnosepfad:   ausKatalog,

		// **Und der zuletzt gewählte Bildordner** — AP-10 (#2367), dieselbe
		// Gestalt: ein Name, keine Werte.
		domain.EinstellungBildordner: ausKatalog,

		// **Und das Ziel der Ausgabe je Set** — ebenso ein Name ohne Werte.
		domain.EinstellungAusgabeziele: ausKatalog,

		// **Das Protokoll** — OPS-04 (#1721), aus dem Katalog der Seite.
		domain.ProtokollStufe:        ausKatalog,
		domain.ProtokollDatei:        ausKatalog,
		domain.ProtokollSystem:       ausKatalog,
		domain.ProtokollAusgabe:      ausKatalog,
		domain.ProtokollFormat:       ausKatalog,
		domain.ProtokollGrenze:       ausKatalog,
		domain.ProtokollAufbewahrung: ausKatalog,
		domain.ProtokollKomponenten:  ausKatalog,
		domain.ProtokollOrt:          ausKatalog,

		// **Die Messung** — OPS-05 (#1722), ebenso.
		domain.MessungSpur:         ausKatalog,
		domain.MessungAbtastung:    ausKatalog,
		domain.MessungSpurgrenze:   ausKatalog,
		domain.MessungAufbewahrung: ausKatalog,
		domain.MessungProfile:      ausKatalog,

		// **Die Grenze der Zwischenspeicher** — OPS-08 (#1725), ebenso.
		domain.ZwischenspeicherGrenze: ausKatalog,

		// **Der Ordner des Plans** — OPS01 (#2347): ihn zeichnet die Fläche
		// „Letzte Sicherung" mit Kästchen und Dialog, nicht die Liste.
		domain.SicherungsplanOrdner: ausEigenerSeite,

		// **Der Arbeitsplatz und die zwei Schranken** — Werkstück O3 von
		// Phase 5.2
		// (#1427).
		// Die zwei Breiten und die Kostengrenze haben **keine** Werteliste;
		// die Schleife unten läuft über null Werte und prüft damit nur ihren
		// Namen — wie beim Hochladepfad.
		domain.ArbeitsplatzNavigator:       ausKatalog,
		domain.ArbeitsplatzInspektor:       ausKatalog,
		domain.ArbeitsplatzNavigatorbreite: ausKatalog,
		domain.ArbeitsplatzInspektorbreite: ausKatalog,

		// **Und wie eine Liste aussieht** — UX-11.AK-04
		// (#1683).
		domain.AnsichtSets:     ausKatalog,
		domain.AnsichtDecks:    ausKatalog,
		domain.AnsichtSammlung: ausKatalog,

		domain.RichtlinieNurLokal:     ausKatalog,
		domain.RichtlinieKostengrenze: ausKatalog,

		// **Und die fünf Schranken des Handels** — Werkstück `T1` von
		// Phase 5.6
		// (#1466),
		// [0118](../../docs/entscheidungen/0118-eine-handelsrichtlinie-ist-eine-schranke.md).
		// Drei von ihnen haben **keine** Werteliste — eine Menge, ein Wert
		// und eine Frist sind Zahlen; die Schleife unten läuft über null
		// Werte und prüft damit ihren Namen, wie beim Hochladepfad.
		domain.RichtlinieHandel:      ausKatalog,
		domain.RichtlinieHandelMarkt: ausKatalog,
		domain.RichtlinieHandelMenge: ausKatalog,
		domain.RichtlinieHandelWert:  ausKatalog,
		domain.RichtlinieHandelFrist: ausKatalog,

		// **Die lokale KI-Laufzeit** — Werkstück A7 von Phase 5.3
		// (#1446).
		// **Aus dem Katalog und nicht aus dem Schalterpaar**: das
		// geteilte `dienstAus`/`dienstAn` hängt an `dienstnamen.ts`, und
		// dort stehen die **Eigennamen der Kartendienste** — eine
		// KI-Laufzeit ist keiner. Die Adresse hat **keine** Werte; geprüft
		// wird ihr Name.
		domain.KILaufzeit:      ausKatalog,
		domain.KIGeraetemodell: ausKatalog,
		domain.KIVorrangmodell: ausKatalog,
		domain.KIAdresse:       ausKatalog,

		// **Und die Kontoschau** — Werkstück `T11` von Phase 5.6
		// (#1476).
		// Zwei Werte, beide im Katalog: „Aus" und „An" heißen hier dasselbe
		// wie bei der KI-Laufzeit, und zwei Schreibweisen für denselben
		// Zustand lernt niemand gern.
		domain.Kontoschau: ausKatalog,

		// **Die sechs Netzeinstellungen** — N5 von Phase 5.4
		// (#1571).
		// Aus dem Katalog: eine Proxyadresse und eine Frist haben **keine**
		// Werteliste, und geprüft wird deshalb ihr **Name** — genau die
		// Lücke, die #1540 an den Dienstadressen gefunden hat.
		domain.NetzProxyHTTP:       ausKatalog,
		domain.NetzProxyHTTPS:      ausKatalog,
		domain.NetzProxyAus:        ausKatalog,
		domain.NetzFristVerbindung: ausKatalog,
		domain.NetzFristKopf:       ausKatalog,
		domain.NetzFristStillstand: ausKatalog,
		domain.NetzSchnittstelle:   ausKatalog,
		domain.NetzLauschen:        ausKatalog,
		domain.DienstTLSZertifikat: ausKatalog,
		domain.DienstTLSSchluessel: ausKatalog,
		domain.DienstTLSKette:      ausKatalog,
		domain.DienstTLSAnker:      ausKatalog,

		// **Und die Kontingente der Dienste** — QA-08.AK-03 (#1739): zwei
		// Werte, beide im Katalog neben den Netzeinstellungen.
		domain.GrenzenKontingente: ausKatalog,
	}

	if gefunden, kennt := fest[schluessel]; kennt {
		return gefunden, true
	}

	// **Die Adresse eines Dienstes ist kein Schalter** —
	// #1540.
	//
	// Sie hat **keine** Werteliste (jede Adresse ist eine andere), und der
	// Schalterweg prüfte für sie deshalb **nichts**: die Schleife lief über
	// null Werte. Im Fenster stand darum `dienst.openai.adresse` als
	// Beschriftung — gefunden in der Sichtprüfung am 2026-09-13, nicht von
	// diesem Wächter.
	if unterschluesselEinesDienstes(schluessel) {
		return ausKatalog, true
	}

	if strings.HasPrefix(schluessel, "dienst.") {
		return ausSchalterpaar, true
	}

	if strings.HasPrefix(schluessel, "frage.") {
		return ausEigennamen, true
	}

	// **Die Anbieterkennungen** — Nacharbeit zu Werkstück 4 von Phase 4.12
	// (#1351). Sie haben **keine Werte** — jeder Anbieter vergibt eine andere
	// Gestalt —, und die Schleife läuft damit über null; geprüft wird ihr
	// **Name**, wie beim Hochladepfad.
	if strings.HasPrefix(schluessel, "anbieter.") {
		return ausKatalog, true
	}

	// **Die Schalter der Bereiche und Funktionen** — OPS-10 (#1908): je
	// Schalter ein Name und „an“/„aus“ im Katalog. **Aus dem Verzeichnis**,
	// damit ein elfter Schalter keine Zeile hier braucht.
	if domain.IstSchaltereinstellung(schluessel) {
		return ausKatalog, true
	}

	return 0, false
}
