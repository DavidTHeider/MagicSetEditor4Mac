package app_test

import "github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"

// freieEinstellungen sind die Schlüssel, die freien Text nehmen — die
// Begründung je Gruppe steht hier, die Regel dahinter am Aufrufer
// (`TestJedeEinstellungBringtIhreWerteMit`).
func freieEinstellungen() map[string]bool {
	frei := map[string]bool{
		domain.VorgabeHochladepfad:         true,
		domain.ArbeitsplatzNavigatorbreite: true,
		domain.ArbeitsplatzInspektorbreite: true,
		domain.RichtlinieKostengrenze:      true,

		// **Und die Adresse der KI-Laufzeit** — A7 (#1446), aus demselben
		// Grund: eine Regel ([domain.PruefeKIAdresse]) statt einer Liste.
		domain.KIAdresse: true,

		// **Und das Modell des Vorrangs** — #2054: welches es gibt, sagt die
		// Laufzeit; [domain.PruefeKIVorrangmodell] prüft die Gestalt.
		domain.KIVorrangmodell: true,

		// **Und die zwei Fenstermaße** — UX-14.AK-01 (#1686). Eine
		// Fenstergröße entsteht am Fensterrand; für sie gilt
		// [domain.PruefeFenstermass] statt einer Liste.
		domain.EinstellungFensterbreite: true,
		domain.EinstellungFensterhoehe:  true,

		// **Und die Belegung von ⌘0 bis ⌘9** — UX-02.AK-03 (#1674): zehn
		// Paare, geprüft von [domain.PruefeZiffernbelegung].
		domain.EinstellungZiffernbelegung: true,
	}

	for _, anbieter := range domain.Anbieter() {
		frei[domain.AnbieterKennung(anbieter)] = true
	}

	// **Und die sechs Netzeinstellungen** — N5 (#1571). Ein Proxy ist eine
	// Adresse, eine Frist eine Zahl; fünf haben eine Regel statt einer Liste,
	// und `netz.proxy.aus` hat bewusst keine.
	for _, schluessel := range []string{
		domain.NetzProxyHTTP, domain.NetzProxyHTTPS, domain.NetzProxyAus,
		domain.NetzFristVerbindung, domain.NetzFristKopf,
		domain.NetzFristStillstand, domain.NetzSchnittstelle,
		domain.DienstTLSZertifikat, domain.DienstTLSSchluessel, domain.DienstTLSKette, domain.DienstTLSAnker,
	} {
		frei[schluessel] = true
	}

	// **Und drei Schranken des Handels** — `T1` von Phase 5.6 (#1466):
	// eine Menge, ein Wert und eine Frist sind Zahlen und keine Auswahl.
	for _, schluessel := range []string{
		domain.RichtlinieHandelMenge, domain.RichtlinieHandelWert,
		domain.RichtlinieHandelFrist,
	} {
		frei[schluessel] = true
	}

	// **Und Darstellung, Zeitzone und Vorgabewährung** — LOC-03.AK-03
	// (#1689): eine Locale, eine Zone und ein Code lassen sich prüfen, aber
	// nicht aufzählen.
	for _, schluessel := range []string{
		domain.EinstellungDarstellung, domain.EinstellungZeitzone,
		domain.VorgabeWaehrung,

		// **Und die eigenen Farben** — UX-09.AK-04 (#1681): zehn Merkmale
		// mit Hexfarben lassen sich prüfen, aber nicht aufzählen.
		domain.EinstellungGestaltung,

		// **Und die drei Ordner aus OPS-02** (#1719) — geprüft, nicht
		// aufgezählt.
		domain.VorgabeAusfuhrpfad,
		domain.VorgabeSicherungspfad,
		domain.VorgabeDiagnosepfad,

		// **Und der zuletzt gewählte Bildordner** — AP-10 (#2367), mit
		// derselben Ordnerprüfung.
		domain.EinstellungBildordner,
		domain.EinstellungAusgabeziele, // und das Ziel je Set, eine Liste in JSON

		// **Und vier des Protokolls** — OPS-04 (#1721).
		domain.ProtokollGrenze,
		domain.ProtokollAufbewahrung,
		domain.ProtokollKomponenten,
		domain.ProtokollOrt,

		// **Und vier der Messung** — OPS-05 (#1722): Zahlen mit Grenzen.
		domain.MessungAbtastung,
		domain.MessungSpurgrenze,
		domain.MessungAufbewahrung,
		domain.MessungProfile,
		domain.ZwischenspeicherGrenze,

		// **Und die gespeicherten Filter** — UX-12 (#1684): eine Liste in JSON.
		domain.EinstellungGespeicherteFilter,

		// **Und die eigene Übersetzung** — LOC-02 (#1688): welche es gibt,
		// sagt der Bestand.
		domain.EinstellungEigeneSprache,

		// **Und die Lauschwahl** — G08 (#2272): welche Adressen es gibt,
		// sagt der Rechner.
		domain.NetzLauschen,

		// **Und der Ordner des Plans** — OPS01 (#2347): ein Ordner, wie die
		// Vorgabepfade, der sich prüfen und nicht aufzählen lässt.
		domain.SicherungsplanOrdner,
	} {
		frei[schluessel] = true
	}

	// **Und die Adresse je Dienst mit Gateway** — Werkstück 3 von A18
	// (#1533), aus demselben Grund wie [domain.KIAdresse]: eine Adresse
	// lässt sich nicht aufzählen, aber prüfen.
	for _, dienst := range domain.Dienste() {
		if dienst.AdresseEinstellbar() {
			frei[domain.DienstAdresse(dienst)] = true
		}
	}

	return frei
}
