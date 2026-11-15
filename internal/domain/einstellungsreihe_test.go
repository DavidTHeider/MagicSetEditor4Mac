package domain_test

import (
	"slices"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Die Einstellungen stehen in einer Liste, nicht in dreien.**
//
// Der Zweitblick über Phase G hat gefunden, was drei Listen kosten:
// `internal/service/einstellungen.go` zählte drei Schlüssel auf,
// `internal/app/einstellungen.go` vier, `internal/domain` kannte vier. Der
// vierte — `bildkonflikt` — fiel damit an der Dienstnaht heraus: die Antwort
// trug ihn mit **leerem** Wert, die Bildausgabe deutete das leere Verhalten
// wie „überschreiben", und ein gewähltes `altbehalten` fiel still weg.
//
// **Die Klasse, nicht der Fall** (§10): eine Liste, und die anderen lesen sie.
func TestDieEinstellungenStehenInEinerReihe(t *testing.T) {
	t.Parallel()

	reihe := domain.Einstellungsreihe()

	erwartet := make([]string, 0, 5+len(domain.Dienste())+len(domain.Fragen()))
	erwartet = append(append(erwartet,
		domain.EinstellungSprache,
		domain.EinstellungEigeneSprache,

		// **Darstellung und Zeitzone gleich hinter der Sprache** —
		// LOC-03.AK-03 (#1689). Dieselbe Frage für Zahlen und Uhren statt
		// für Wörter, und dieselbe Tragweite.
		domain.EinstellungDarstellung,
		domain.EinstellungZeitzone,
		domain.EinstellungGespeicherteFilter,
		domain.EinstellungZiffernbelegung,

		domain.EinstellungErscheinungsbild,
		domain.EinstellungSchriftgroesse,

		// **Die Anzeigeschrift steht bei der Schriftgröße** — beide betreffen
		// jede Zeile; jene ihre Größe, diese ihre Gestalt (#1325).
		domain.EinstellungSchriftart,

		// **Die Gestaltung aus UX-09** (#1681) gleich hinter der
		// Anzeigeschrift: wer dort liest, wie die Oberfläche aussieht, findet
		// Farbwelt, Reichweite der Schrift, Fensterbild und eigene Merkmale
		// beisammen.
		domain.EinstellungFarbwelt, domain.EinstellungSchriftbereich,
		domain.EinstellungHintergrund, domain.EinstellungGestaltung,

		// **Die Aktualisierung steht vor dem Namenskonflikt**: sie betrifft
		// das Programm, jener einen Ausgabelauf — dieselbe Ordnung nach
		// Tragweite wie darüber (#1100).
		domain.EinstellungAktualisierung,

		// **Die Einführung dahinter** (#1325): sie betrifft **einen
		// Augenblick** — den beim Start — und damit weniger als alles
		// darüber. Vor dem Namenskonflikt steht sie trotzdem, denn jener
		// betrifft nur einen Ausgabelauf.
		domain.EinstellungErsteSchritte,
		// **Die Seitenleiste dahinter** — Werkstück U14 von Phase 5.1 (#1490):
		// sie betrifft das Fenster als Ganzes, aber nur seine Gestalt, und
		// steht deshalb hinter dem, was die Anwendung tut, und vor den Vorgaben
		// für neue Datensätze.
		domain.EinstellungSeitenleiste,

		// **Die Lage der Leiste neben ihr** — UX-14.AK-02 (#1686): dieselbe
		// Frage nach der Gestalt der Navigation, auf Geräteebene.
		domain.EinstellungLeistenlage,

		// **Und der Fensterstart unmittelbar dahinter** — UX-14.AK-01
		// (#1686). Dieselbe Sorte Frage wie die Seitenleiste: die Gestalt des
		// Fensters, auf Geräteebene. Die zwei Zahlen folgen ihm, weil sie nur
		// für seinen dritten Wert gelten.
		domain.EinstellungFensterstart,
		domain.EinstellungFensterbreite,
		domain.EinstellungFensterhoehe,

		// **Die Vorgaben dahinter** — Werkstück 9 von Phase 4.12 (#1351),
		// Punkte 2, 3 und 8. Sie betreffen **einen neuen Datensatz** und
		// damit weniger als alles darüber, aber mehr als ein einzelner
		// Ausgabelauf.
		domain.VorgabeZustand,

		// **Die Vorgabewährung neben dem Vorgabezustand** — LOC-03 (#1689):
		// beide sagen, womit eine neue Sammlungszeile anfängt.
		domain.VorgabeWaehrung,

		domain.VorgabeEinzelspiel,
		domain.VorgabeHochladepfad,
		domain.VorgabeAusfuhrpfad,
		domain.VorgabeSicherungspfad,
		domain.VorgabeDiagnosepfad,

		// **Die lokale KI-Laufzeit** — A7 (#1446). Sie steht zwischen den
		// Vorgaben und dem Namenskonflikt, weil sie dort in der Reihe der
		// Einstellungen steht: hinter dem, womit ein Datensatz anfängt, und
		// vor dem, was beim Schreiben geschieht.
		domain.KILaufzeit, domain.KIGeraetemodell, domain.KIVorrangmodell, domain.KIAdresse,

		// **Und die Kontoschau dahinter** — T11 (#1476): sie beantwortet
		// dieselbe Frage wie die KI-Laufzeit, nur für einen Marktplatz.
		domain.Kontoschau,

		// **Und die sechs Netzeinstellungen dahinter** — N5
		// (#1571).
		// Sie stehen zwischen der KI-Laufzeit und dem Namenskonflikt, weil
		// sie dort in der Reihe stehen: hinter dem, **wen** dieses Programm
		// fragt, und vor dem, was beim Schreiben einer Datei geschieht.
		domain.NetzProxyHTTP, domain.NetzProxyHTTPS, domain.NetzProxyAus,
		domain.NetzFristVerbindung, domain.NetzFristKopf, domain.NetzFristStillstand, domain.NetzSchnittstelle,
		domain.NetzLauschen, // **und wo die eigenen Dienste lauschen** — G08 (#2272)
		domain.DienstTLSZertifikat, domain.DienstTLSSchluessel, domain.DienstTLSKette, domain.DienstTLSAnker,
		domain.GrenzenKontingente, // **und die Kontingente derselben Dienste** — QA-08.AK-03 (#1739)

		domain.EinstellungBildkonflikt,

		// **Die Anordnung des Tisches steht hier und nicht am Ende** —
		// PLAY-04.AK-03
		// (#1712).
		//
		// **Weil Go seine `init` nach Dateinamen ordnet**: `autoanordnung.go`
		// kommt vor `einstellungsreihe.go`, und jene hängt erst den
		// Arbeitsplatz und dann die Dienste an. Die Stelle ist damit eine
		// Eigenschaft der Dateinamen und keine Aussage über Wichtigkeit —
		// wer die Datei umbenennt, verschiebt die Zeile auf der
		// Einstellungsseite, und **dieser Lauf sagt es ihm**.
		domain.EinstellungAutoanordnung,

		// **Und die Beschriftung des Tisches gleich dahinter** — seit dem
		// 2026-09-24, aus demselben Grund: `beschriftung.go` kommt nach
		// `autoanordnung.go` und vor `einstellungsreihe.go`.
		domain.EinstellungTischbeschriftung,

		// **Und der zuletzt gewählte Bildordner** — AP-10 (#2367). Er hängt im
		// `init` von `einstellungsreihe.go` an, also hinter den zwei des
		// Tisches: `einstellungsliste.go` steht an der Zeilengrenze.
		domain.EinstellungBildordner,

		// **Und der Arbeitsplatz** — Werkstück O3 von Phase 5.2
		// (#1427).
		// Er betrifft **jede Fläche des Fensters**; in der Reihe steht er
		// hinter dem Namenskonflikt und vor den Dienstschaltern.
		domain.ArbeitsplatzNavigator, domain.ArbeitsplatzInspektor,
		domain.ArbeitsplatzNavigatorbreite, domain.ArbeitsplatzInspektorbreite,

		// **Und die gemerkten Ansichten dahinter** — UX-11.AK-04 (#1683).
		domain.AnsichtSets, domain.AnsichtDecks, domain.AnsichtSammlung,

		// **Und das Ziel der Ausgabe je Set** — AP-10 (#2367), dieselbe Frage
		// an das Fenster.
		domain.EinstellungAusgabeziele), schalterschluessel()...) // die Schalter: OPS-10 (#1908)

	// **Und die Dienstschalter am Ende** — Phase 4.8
	// (#1321).
	//
	// Sie werden **aus der Dienstaufzählung** angehängt und nicht hier
	// aufgeschrieben: eine zweite Liste ist genau das, wogegen dieser Lauf
	// steht. Was er prüft, ist die **Stelle** — hinter allem, was die
	// Anwendung selbst betrifft.
	for _, dienst := range domain.Dienste() {
		erwartet = append(erwartet, domain.DienstEinstellung(dienst))

		// **Und gleich hinter dem Schalter seine Adresse**, wo es eine geben
		// darf — Werkstück 3 von A18
		// (#1533).
		// Dieselbe Frage, eine Ebene tiefer: „darf er gefragt werden" und
		// „wen fragt er dann eigentlich".
		if dienst.AdresseEinstellbar() {
			erwartet = append(erwartet, domain.DienstAdresse(dienst))
		}
	}

	// **Und die Anbieterkennungen dahinter** — Nacharbeit zu Werkstück 4 von
	// Phase 4.12 (#1351). **Aus [domain.Anbieter]** und nicht hier
	// aufgeschrieben, aus demselben Grund wie oben.
	//
	// **Hinter den Diensten und vor den Fragen**, weil beides über dieselbe
	// `init` läuft: was die Anwendung nach außen tut, steht hinten, und
	// innerhalb davon steht „womit weise ich mich aus" vor „wen frage ich".
	erwartet = append(erwartet, anbieterkennungen()...)

	// **Und die Wahlen je Frage dahinter** — Werkstück 8 von Phase 4.8.
	// „Darf er gefragt werden" steht vor „wer beantwortet was".
	for _, frage := range domain.Fragen() {
		erwartet = append(erwartet, domain.FrageEinstellung(frage))
	}

	// **Protokoll, Messung, Zwischenspeicher** — OPS-04, OPS-05, OPS-08 —
	// **und der Ordner des Plans** (OPS01, #2347).
	erwartet = append(append(append(erwartet, domain.Protokollschluessel()...),
		domain.Messschluessel()...), domain.ZwischenspeicherGrenze, domain.SicherungsplanOrdner)

	// **Und ganz am Ende die Schranken** — Werkstück O3 von Phase 5.2
	// (#1427). Hinter den Dienstschaltern, weil sie über sie urteilen: wer
	// von oben nach unten liest, sieht erst, was einzeln geschaltet wird, und
	// dann, was das alles auf einmal abstellt.
	//
	// **Und die fünf des Handels dahinter** — Werkstück `T1` von Phase 5.6
	// (#1466). Sie stehen hinter den zwei allgemeinen, weil sie enger sind:
	// „kein fremder Dienst" gilt für alles, „kein Markt" für einen Weg.
	//
	// **Und die Tiefe des Kartenformulars ganz zuletzt** — UX-05.AK-01
	// (#1677).
	// Sie hängt sich wie die Dienstschalter über `init` an; wo sie steht,
	// sagt nichts über ihre Wichtigkeit, sondern dass sie ein Anhang ist.
	erwartet = append(erwartet,
		domain.RichtlinieNurLokal, domain.RichtlinieKostengrenze,
		domain.RichtlinieHandel, domain.RichtlinieHandelMarkt,
		domain.RichtlinieHandelMenge, domain.RichtlinieHandelWert,
		domain.RichtlinieHandelFrist,
		domain.EinstellungFeldtiefe)

	if len(reihe) != len(erwartet) {
		t.Fatalf("%v", reihe)
	}

	for i := range erwartet {
		if reihe[i] != erwartet[i] {
			t.Errorf("an Stelle %d steht %q, erwartet %q", i, reihe[i], erwartet[i])
		}
	}
}

// **Und jede Zeile ist vollständig.**
//
// Hier stand ein Wächter, der dieselben vier Schlüssel noch einmal aufzählte —
// der Nachlauf über Phase G hat gefunden, dass er damit die halbe Zusage hielt
// (§10). **Nachgestellt und gesehen**: ein fünfter Schlüssel allein in
// `einstellungswerte` ließ ihn grün.
//
// **Diesen Fall gibt es nicht mehr.** Eine Einstellung ist seither eine Zeile
// mit Schlüssel, Bereich, Werten und Vorgabe; eine halbe Zeile ist nicht
// darstellbar. Was ein Wächter noch prüfen kann, ist, ob eine Zeile leer
// bleibt — und ob eine Vorgabe unter ihren eigenen Werten steht.
func TestJedeEinstellungIstVollstaendig(t *testing.T) {
	t.Parallel()

	for _, schluessel := range domain.Einstellungsreihe() {
		// **Eine Einstellung ohne Werteliste nimmt freien Text** — Werkstück
		// 9 (#1351), Punkt 8. Ein Verzeichnis lässt sich nicht aufzählen, und
		// ein leeres `Werte` **sagt** das.
		//
		// **Die Zusage schrumpft dabei nicht, sie wird genauer**: geprüft
		// wird jetzt, dass genau die Schlüssel ohne Werte auskommen, die es
		// dürfen — eine Aufzählung, die versehentlich leer bliebe, ist
		// weiterhin ein Befund.
		werte := domain.ErlaubteEinstellungswerte(schluessel)
		if len(werte) == 0 && !slices.Contains(freieEinstellungen, schluessel) {
			t.Errorf("%s steht in der Reihe, hat aber keine Werte", schluessel)
		}

		if domain.EinstellungsBereich(schluessel) == "" {
			t.Errorf("%s steht in der Reihe, nennt aber keinen Bereich", schluessel)
		}

		// **Eine Vorgabe, die die eigene Prüfung nicht besteht, ist keine.**
		// Sie käme als Antwort heraus und fiele beim Zurückschreiben durch —
		// eine Einstellung, die sich nicht auf den Wert setzen lässt, den sie
		// zeigt.
		//
		// **Über [domain.PruefeEinstellung] und nicht über die Werteliste** —
		// Werkstück O3 von Phase 5.2 (#1427). Vorher stand hier
		// `slices.Contains`; das hielt die halbe Zusage, seit es Einstellungen
		// mit einer Regel statt einer Liste gibt. Die Prüfung **enthält** die
		// Liste und dazu die Regel.
		vorgabe := domain.EinstellungsVorgabe(schluessel)
		if vorgabe == "" {
			continue
		}

		if fehler := domain.PruefeEinstellung(schluessel, vorgabe); fehler != nil {
			t.Errorf("die eigene Vorgabe von %s wird abgewiesen: %v",
				schluessel, fehler)
		}
	}
}

// **Eine Kopie**: wer die Reihe änderte, änderte die Regel.
func TestDieReiheIstEineKopie(t *testing.T) {
	t.Parallel()

	erste := domain.Einstellungsreihe()
	erste[0] = "verbogen"

	if domain.Einstellungsreihe()[0] != domain.EinstellungSprache {
		t.Error("die Reihe ließ sich von außen ändern")
	}
}
