package domain_test

// **Welche Einstellungen frei sind** — aus `einstellungsreihe_test.go`
// gezogen, als die Wahl des Vorrangmodells (#2054) die Datei über die
// Zeilengrenze aus §2 hob. Gelesen wird die Liste dort, in
// `TestJedeEinstellungIstVollstaendig`.

import "github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"

// freieEinstellungen sind die, deren Wert sich nicht aufzählen lässt.
//
// **Eine Aufzählung und keine Regel über Namen.** Dass ein Pfad freier Text
// ist, ist eine **Entscheidung**; sie an einem Präfix festzumachen hieße, sie
// beim nächsten Namen zu vergessen — und dann bliebe eine Auswahlliste
// versehentlich leer, ohne dass es jemand merkt.
//
// **Die Anbieterkennungen kommen aus [domain.Anbieter] und nicht abgeschrieben**
// — Nacharbeit zu Werkstück 4 (#1351). Sie sind freier Text, weil jeder
// Anbieter eine andere Gestalt vergibt: GitHub `Ov23…`, Google eine Adresse,
// Microsoft eine UUID. Eine Prüfung darauf wäre eine Regel, die dieses
// Programm sich ausdenkt, und die erste Änderung beim Anbieter machte jede
// gültige Kennung unbrauchbar (§11).
//
// **Seit Werkstück O3 von Phase 5.2** (#1427) stehen drei weitere darin: zwei
// Spaltenbreiten und eine Kostengrenze. Sie sind **nicht** ungeprüft — für sie
// gilt eine Regel statt einer Liste (`einstellung.Pruefung`), und die prüft
// der Lauf darüber über [domain.PruefeEinstellung] mit.
var freieEinstellungen = append([]string{
	domain.VorgabeHochladepfad,
	domain.KIVorrangmodell,          // #2054: welches es gibt, sagt die Laufzeit
	domain.EinstellungEigeneSprache, // LOC-02 (#1688): welche es gibt, sagt der Bestand

	// **Und seit OPS-02 (#1719) drei Ordner** — geprüft, nicht aufgezählt.
	domain.VorgabeAusfuhrpfad,
	domain.VorgabeSicherungspfad,
	domain.VorgabeDiagnosepfad,

	// **Und der zuletzt gewählte Bildordner** — AP-10 (#2367), ebenso.
	domain.EinstellungBildordner,

	// **Und das Ziel der Ausgabe je Set** — AP-10 (#2367): eine Liste in JSON
	// mit einer Regel, wie die gespeicherten Filter.
	domain.EinstellungAusgabeziele,

	// **Und vier des Protokolls** — OPS-04 (#1721), ebenso.
	domain.ProtokollGrenze,
	domain.ProtokollAufbewahrung,
	domain.ProtokollKomponenten,
	domain.ProtokollOrt,

	// **Und vier der Messung und die Grenze** — OPS-05, OPS-08.
	domain.ZwischenspeicherGrenze,
	domain.MessungAbtastung,
	domain.MessungSpurgrenze,
	domain.MessungAufbewahrung,
	domain.MessungProfile,

	// **Und der Ordner des Plans** — OPS01 (#2347): ein Ordner, wie die der
	// Dialoge.
	domain.SicherungsplanOrdner,

	// **Und seit LOC-03 (#1689) Darstellung, Zeitzone und Vorgabewährung.**
	// Eine Locale, eine Zone und ein Code lassen sich prüfen, aber nicht
	// aufzählen — dafür gibt es `PruefeDarstellung`, `PruefeZeitzone` und
	// `PruefeWaehrungscode`.
	domain.EinstellungDarstellung,
	domain.EinstellungZeitzone,
	domain.VorgabeWaehrung,

	// **Und seit UX-09.AK-04 (#1681) die eigenen Merkmale**: zehn Farben
	// lassen sich prüfen (`PruefeGestaltung`), aber nicht aufzählen.
	domain.EinstellungGestaltung,

	// **Und seit UX-02.AK-03 (#1674) die Belegung der Ziffern, seit UX-12.AK-03
	// (#1684) die gespeicherten Filter**: geprüft (`PruefeZiffernbelegung`,
	// `PruefeGespeicherteFilter`), aber nicht als Liste aufzählbar.
	domain.EinstellungZiffernbelegung,
	domain.EinstellungGespeicherteFilter,

	// **Und seit G08 (#2272) die Lauschwahl** — Adressen lassen sich prüfen
	// (`PruefeLauschwahl`), aber nicht aufzählen: welche es gibt, sagt der
	// Rechner.
	domain.NetzLauschen,

	domain.ArbeitsplatzNavigatorbreite,
	domain.ArbeitsplatzInspektorbreite,
	domain.RichtlinieKostengrenze,

	// **Und drei des Handels** — Werkstück `T1` von Phase 5.6 (#1466). Eine
	// Menge, ein Wert und eine Frist sind Zahlen; eine Auswahlliste dafür
	// wäre eine Regel, die dieses Programm sich ausdenkt.
	domain.RichtlinieHandelMenge,
	domain.RichtlinieHandelWert,
	domain.RichtlinieHandelFrist,

	// **Und die zwei Fenstermaße** — UX-14.AK-01 (#1686). Eine Fenstergröße
	// entsteht am Fensterrand; eine Aufzählung von 1.437 Punkten wäre keine
	// Wahl. Für sie gilt `PruefeFenstermass` statt einer Liste.
	domain.EinstellungFensterbreite,
	domain.EinstellungFensterhoehe,

	// **Und seit A7 (#1446) die Adresse der KI-Laufzeit.** Auch sie hat eine
	// Regel statt einer Liste: eine Adresse lässt sich nicht aufzählen, aber
	// prüfen ([domain.PruefeKIAdresse]).
	domain.KIAdresse,

	// **Und seit N5 (#1571) die sechs Netzeinstellungen.** Ein Proxy ist
	// eine Adresse, eine Frist eine Zahl, und `netz.proxy.aus` eine Liste in
	// der Gestalt von `NO_PROXY` — keine davon lässt sich aufzählen. Fünf
	// haben eine **Regel** statt einer Liste; die sechste hat bewusst keine,
	// und der Grund steht bei ihr in `einstellung.go`. Dazu der Name einer
	// Schnittstelle (NET-02, #1706) und vier Pfade (NET-03, #1707).
	domain.NetzProxyHTTP, domain.NetzProxyHTTPS, domain.NetzProxyAus,
	domain.NetzFristVerbindung, domain.NetzFristKopf, domain.NetzFristStillstand,
	domain.NetzSchnittstelle,
	domain.DienstTLSZertifikat, domain.DienstTLSSchluessel, domain.DienstTLSKette, domain.DienstTLSAnker,
}, append(anbieterkennungen(), dienstadressen()...)...)

// dienstadressen gibt die Schlüssel aller Dienste, die hinter einem Gateway
// stehen dürfen — Werkstück 3 von A18
// (#1533).
//
// **Wie [domain.KIAdresse]**: eine Regel statt einer Liste, weil eine Adresse
// sich nicht aufzählen lässt. **Und über `Dienste()` und nicht abgeschrieben**
// — eine zweite Aufzählung liefe beim vierten Anbieter auseinander.
func dienstadressen() []string {
	schluessel := make([]string, 0, len(domain.Dienste()))

	for _, dienst := range domain.Dienste() {
		if dienst.AdresseEinstellbar() {
			schluessel = append(schluessel, domain.DienstAdresse(dienst))
		}
	}

	return schluessel
}

// anbieterkennungen gibt die Schlüssel aller Anbieter.
func anbieterkennungen() []string {
	schluessel := make([]string, 0, len(domain.Anbieter()))
	for _, anbieter := range domain.Anbieter() {
		schluessel = append(schluessel, domain.AnbieterKennung(anbieter))
	}

	return schluessel
}
