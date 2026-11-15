// **Wo eine Einstellung entschieden wird** — Werkstück O3 von Phase 5.2
// (#1427),
// nach [Vermerk 0079](../../docs/entscheidungen/0079-einstellungsebenen-und-portabler-import.md).
//
// # Was eine Ebene sagt und was nicht
//
// **Sie sagt, wem die Frage gehört — nicht, in welcher Tafel der Wert liegt.**
// Wo er liegt, ist unverändert: `settings` je Profil, `nutzereinstellungen` je
// Mensch. Ein Umzug wäre eine Verhaltensänderung ohne Anlass (§15, Satz 4),
// und er brächte nichts, was diese Zuordnung nicht schon bringt.
//
// **Wofür sie gebraucht wird**, sind zwei Fragen, die es vorher nicht gab:
//
//   - **Darf dieser Wert mitreisen?** Ein Pfad auf dieser Platte und die
//     Kennung, unter der sich diese Installation ausweist, dürfen es nicht —
//     sie sind `geraet`, und `geraet` bleibt hier.
//   - **In welchen Abschnitt eines Austauschs gehört er?** Der Abschnitt ist
//     die Ebene; ein Schlüssel im falschen Abschnitt ist ein Befund.
//
// # Warum die vier Erscheinungseinstellungen `nutzer` sind
//
// **Weil die Frage dem Menschen gehört und das Profil sie nur übersteuert** —
// das steht so in `uebersteuerung.go` und ist dort begründet. Dass ihre
// Übersteuerung trotzdem im Abschnitt `profil` mitreist, ist kein Widerspruch:
// die Ebene sagt, wo der **Boden** liegt, und ein Abschnitt trägt, was auf
// seiner Ebene **gesetzt** ist.

package domain

import (
	"slices"
	"strings"
)

// Ebene sagt, wo eine Einstellung entschieden wird.
type Ebene string

// Die vier Ebenen.
const (
	// EbeneGeraet ist dieser Rechner und diese Installation. **Nichts davon
	// reist mit** — es beschriebe sonst drüben eine Maschine, die es nicht
	// gibt.
	EbeneGeraet Ebene = "geraet"

	// EbeneNutzer ist der Mensch. Er nimmt es in jedes seiner Profile mit.
	EbeneNutzer Ebene = "nutzer"

	// EbeneProfil ist die Arbeitsumgebung.
	EbeneProfil Ebene = "profil"

	// EbeneArbeitsplatz ist das Fenster: welche Spalte steht und wie breit.
	// **Reist nur auf Wunsch** — ein Layout hängt an einem Bildschirm.
	EbeneArbeitsplatz Ebene = "arbeitsplatz"
)

// Ebenen gibt die vier, in der Reihenfolge von außen nach innen.
func Ebenen() []Ebene {
	return []Ebene{EbeneGeraet, EbeneNutzer, EbeneProfil, EbeneArbeitsplatz}
}

// KenntEbene sagt, ob dieser Name eine ist.
func KenntEbene(name string) bool { return slices.Contains(Ebenen(), Ebene(name)) }

// ebenen ordnet jedem festen Schlüssel seine Ebene zu.
//
// **Eine Abbildung über Schlüssel und nicht über Zeilen**: vier Schlüssel
// stehen in beiden Reihen, und „genau eine Ebene" ist eine Aussage über die
// Frage, nicht über die Tafel, in der sie beantwortet wird.
//
// **Die drei erzeugten Familien stehen nicht darin** — `dienst.`, `frage.` und
// `anbieter.` entstehen aus `dienste.go` und `anbieterkonto.go`; eine zweite
// Aufzählung hier liefe beim fünften Dienst auseinander. Für sie gilt die
// Regel unten, und der Wächter über beide Reihen macht sichtbar, wenn eine
// vierte Familie dazukommt.
var ebenen = map[string]Ebene{
	// **Der Mensch**: wie die Anwendung aussieht und spricht, und was er an
	// Barrieren braucht.
	EinstellungSprache:          EbeneNutzer,
	EinstellungEigeneSprache:    EbeneNutzer,
	EinstellungDarstellung:      EbeneNutzer,
	EinstellungErscheinungsbild: EbeneNutzer,
	EinstellungSchriftgroesse:   EbeneNutzer,
	EinstellungSchriftart:       EbeneNutzer,
	NutzerBewegung:              EbeneNutzer,
	NutzerKontrast:              EbeneNutzer,
	NutzerFaktorbeimWechsel:     EbeneNutzer,

	// **Der Rechner.**
	//
	//   - `aktualisierung` fragt das Netz — eine Eigenschaft dieser Maschine,
	//     und die steht so schon in `uebersteuerung.go`.
	//   - `ersteschritte` ist „habe ich das hier schon gesehen"; drüben hat es
	//     niemand gesehen.
	//   - `vorgabe.hochladepfad` ist ein **Pfad auf dieser Platte**. Genau das,
	//     was nie in einen Austausch darf.
	EinstellungAktualisierung: EbeneGeraet,

	EinstellungErsteSchritte: EbeneGeraet,
	EinstellungSeitenleiste:  EbeneGeraet,
	EinstellungLeistenlage:   EbeneGeraet,

	// **Die Zeitzone sagt, wo der Rechner steht** — LOC-03 (#1689). Wer
	// sein Profil mitnimmt, nimmt keine Uhr mit.
	EinstellungZeitzone: EbeneGeraet,

	// **Die Tasten gehören der Hand, die sie drückt** (UX-02.AK-03, #1674):
	// wer ⌘3 für die Karten gelernt hat, will es in jedem seiner Profile.
	EinstellungZiffernbelegung: EbeneNutzer,

	// **Wie viel vom Kartenformular dasteht, gehört dem Profil**
	// (UX-05.AK-01, #1677) und nicht dem Gerät: wer erfahren genug für den
	// Expertenmodus ist, ist es an jedem Rechner. Das Kriterium sagt es
	// ausdrücklich — „ein je Nutzer oder Profil änderbarer Standardmodus".
	EinstellungFeldtiefe: EbeneProfil,

	// **Wie die Oberfläche aussieht, über hell und dunkel hinaus, gehört
	// der Arbeitsumgebung** (UX-09, #1681): wer ein Profil für Turniere und
	// eines für Entwürfe führt, sieht sie verschieden. **Die eigenen
	// Merkmale reisen mit** — es sind Farben, keine Dateien und keine
	// Adressen (AK-04).
	EinstellungFarbwelt:       EbeneProfil,
	EinstellungSchriftbereich: EbeneProfil,
	EinstellungHintergrund:    EbeneProfil,
	EinstellungGestaltung:     EbeneProfil,

	// **Gespeicherte Filter gehören dem Profil** (UX-12.AK-03, #1684): sie
	// fragen nach Marken, und Marken gehören dem Profil (0135).
	EinstellungGespeicherteFilter: EbeneProfil,

	// **Wie der Tisch ordnet, gehört dem Menschen** (PLAY-04.AK-03, #1712)
	// und nicht dem Profil: wer seine Kreaturen nach Angriffsreihen legt,
	// legt sie so, gleich mit welchem Deck er antritt. Dieselbe Überlegung
	// wie beim Tischbild — „das Tischbild gehört dem Menschen, nicht dem
	// Profil".
	//
	// **Und nicht dem Gerät**, anders als die Fenstergröße darunter: eine
	// Gewohnheit reist mit, eine Bildschirmbreite nicht.
	EinstellungAutoanordnung: EbeneNutzer,

	// **Wie der Tisch beschriftet ist, auch** — wer die Namen der Zonen nicht
	// braucht oder größer lesen will, will das mit jedem Deck.
	EinstellungTischbeschriftung: EbeneNutzer,

	// **Eine Fenstergröße gehört dem Bildschirm, nicht dem Profil**
	// (UX-14.AK-01, #1686). Wer sein Profil auf einen anderen Rechner
	// mitnimmt, nimmt keinen Bildschirm mit.
	EinstellungFensterstart:  EbeneGeraet,
	EinstellungFensterbreite: EbeneGeraet,
	EinstellungFensterhoehe:  EbeneGeraet,
	VorgabeHochladepfad:      EbeneGeraet,
	VorgabeAusfuhrpfad:       EbeneGeraet,
	VorgabeSicherungspfad:    EbeneGeraet,
	VorgabeDiagnosepfad:      EbeneGeraet,
	// **Der zuletzt gewählte Bildordner auch** (AP-10): ein Pfad dieser Platte.
	EinstellungBildordner: EbeneGeraet,

	// **Das Ziel der Ausgabe je Set** (AP-10, #2367) ist eine Frage an das
	// Fenster wie die Breite einer Spalte — es reist nur auf Wunsch.
	EinstellungAusgabeziele: EbeneArbeitsplatz,

	// **Und der Ordner des Plans** (OPS01, #2347): ein Pfad beschreibt einen
	// Rechner, und ein Stick, der hier steckt, steckt an keinem anderen.
	SicherungsplanOrdner: EbeneGeraet,

	// **Das Protokoll gehört dem Rechner** (OPS-04, #1721) — wie viel er
	// aufschreibt und wohin, beschreibt diese Installation.
	ProtokollStufe:        EbeneGeraet,
	ProtokollDatei:        EbeneGeraet,
	ProtokollSystem:       EbeneGeraet,
	ProtokollAusgabe:      EbeneGeraet,
	ProtokollFormat:       EbeneGeraet,
	ProtokollGrenze:       EbeneGeraet,
	ProtokollAufbewahrung: EbeneGeraet,
	ProtokollKomponenten:  EbeneGeraet,
	ProtokollOrt:          EbeneGeraet,

	// **Und die Messung auch** (OPS-05, #1722): eine Spur und ein Profil
	// beschreiben diesen Rechner und reisen nicht mit einem Profil.
	MessungSpur:         EbeneGeraet,
	MessungAbtastung:    EbeneGeraet,
	MessungSpurgrenze:   EbeneGeraet,
	MessungAufbewahrung: EbeneGeraet,
	MessungProfile:      EbeneGeraet,

	// **Und die Grenze des Zwischenspeichers** (OPS-08, #1725): die Platte
	// gehört dem Rechner.
	ZwischenspeicherGrenze: EbeneGeraet,

	// **Wo eine KI-Laufzeit lauscht, ist eine Eigenschaft dieser Maschine**
	// (A7, #1446) — drüben beschriebe die Adresse einen Rechner, den es
	// nicht gibt. Dass sie **gefragt werden darf**, gehört dem Profil.
	KIAdresse: EbeneGeraet,

	// **Und wie dieses Programm ins Netz geht, ebenso** (N5, #1571): ein
	// Proxy beschreibt das **Netz dieser Maschine**, und eine Frist hängt an
	// ihrer Leitung. Drüben beschriebe beides ein Netz, das es nicht gibt —
	// derselbe Grund wie bei `vorgabe.hochladepfad` darüber, und derselbe
	// Grund, aus dem ein Austausch sie nicht mitnimmt.
	NetzProxyHTTP:       EbeneGeraet,
	NetzProxyHTTPS:      EbeneGeraet,
	NetzProxyAus:        EbeneGeraet,
	NetzSchnittstelle:   EbeneGeraet,
	NetzLauschen:        EbeneGeraet,
	NetzFristVerbindung: EbeneGeraet,
	NetzFristKopf:       EbeneGeraet,
	NetzFristStillstand: EbeneGeraet,

	// **Und die Dateien des Dienstzeugnisses** (NET-03, #1707): ein Pfad auf
	// diesem Rechner beschriebe drüben einen Ort, den es nicht gibt.
	DienstTLSZertifikat: EbeneGeraet,
	DienstTLSSchluessel: EbeneGeraet,
	DienstTLSKette:      EbeneGeraet,
	DienstTLSAnker:      EbeneGeraet,

	// **Und ob die Dienste Kontingente führen** (QA-08.AK-03, #1739): eine
	// kontrollierte Umgebung ist eine Eigenschaft dieses Rechners.
	GrenzenKontingente: EbeneGeraet,

	// **Das Profil**: womit ein Datensatz dieser Arbeitsumgebung anfängt und
	// wie ein Lauf aus ihr herausschreibt.
	VorgabeZustand:          EbeneProfil,
	VorgabeWaehrung:         EbeneProfil,
	VorgabeEinzelspiel:      EbeneProfil,
	EinstellungBildkonflikt: EbeneProfil,
	KILaufzeit:              EbeneProfil,
	KIGeraetemodell:         EbeneProfil,
	KIVorrangmodell:         EbeneProfil,

	// **Die Kontoschau gehört dem Profil** (T11, #1476): welches Konto
	// gelesen werden darf, hängt an der Marke, und die hängt am Profil.
	// **Sie reist damit nicht mit einem Einstellungsaustausch** — wer eine
	// Datei weitergäbe, gäbe eine Freigabe weiter, die niemand erteilt hat.
	Kontoschau: EbeneProfil,
}

// EbeneVon gibt die Ebene eines Schlüssels — der zweite Wert sagt, ob es sie
// gibt.
func EbeneVon(schluessel string) (Ebene, bool) {
	if ebene, kennt := ebenen[schluessel]; kennt {
		return ebene, true
	}

	if slices.Contains(arbeitsplatzschluessel(), schluessel) {
		return EbeneArbeitsplatz, true
	}

	// **Die zwei Schranken gehören dem Menschen.** Er nimmt sie in jedes
	// Profil mit, und ein Profil kann sie nur verschärfen — siehe
	// `richtlinie.go`.
	if slices.Contains(richtlinienschluessel(), schluessel) {
		return EbeneNutzer, true
	}

	// **Die Kennung eines Anbieters ist eine lokale Kennung**: sie sagt, als
	// welche Anwendung sich **diese Installation** beim Anbieter ausweist.
	// Wer sie mitgäbe, gäbe ein Stück Zugang mit.
	if strings.HasPrefix(schluessel, "anbieter.") {
		return EbeneGeraet, true
	}

	// **Ein Schalter gehört dem Rechner** (OPS-10): er sagt, was diese
	// Installation anbietet, nicht was ein Profil darf.
	if IstSchaltereinstellung(schluessel) {
		return EbeneGeraet, true
	}

	// **Ein Dienstschalter und die Frage, wer ihn beantwortet, gehören dem
	// Profil**: sie sagen, was **diese Arbeitsumgebung** nach draußen tut.
	if strings.HasPrefix(schluessel, "dienst.") ||
		strings.HasPrefix(schluessel, "frage.") {
		return EbeneProfil, true
	}

	return "", false
}

// Ebenenschluessel gibt die Schlüssel, die in der festen Tabelle stehen.
//
// **Für den Wächter daneben**, der die Gegenrichtung hält: eine Zeile, die
// einen Schlüssel nennt, den es nicht mehr gibt, fiele sonst niemandem auf.
// Die erzeugten Familien stehen nicht darin — sie können keinen Schlüssel
// nennen, den es nicht gibt.
func Ebenenschluessel() []string {
	alle := make([]string, 0, len(ebenen))
	for schluessel := range ebenen {
		alle = append(alle, schluessel)
	}

	slices.Sort(alle)

	return alle
}
