// **Wann eine Einstellung wirkt** — UX-04.AK-02
// (#1676).
//
// # Der Anlass
//
// **Das Kriterium verlangt es am Feld**: „Jede Einstellung zeigt
// Geltungsbereich, wirksamen Wert, Herkunft, Standard und **Zeitpunkt der
// Wirksamkeit**". Die ersten vier stehen da; der fünfte stand am 2026-09-21
// an **einer** Stelle im ganzen Programm — als Satz unter dem Netzweg:
// „Eine geänderte Frist gilt ab dem nächsten Aufruf; ein geänderter Proxy
// sofort." Er war richtig und unsichtbar.
//
// # Zwei Werte und nicht drei
//
// **Es geht um die Frage, ob jemand etwas tun muss.** „Wirkt beim nächsten
// Aufruf" heißt für den Menschen davor dasselbe wie „sofort" — er tut
// nichts. „Beim nächsten Start" heißt: er muss die Anwendung neu öffnen.
// Eine dritte Stufe dazwischen beschriebe die Maschine statt die Handlung.
//
// # Gemessen und nicht geraten (§15)
//
// Jeder Fall unten nennt die Fundstelle, an der es nachgelesen ist.
package domain_test

import (
	"slices"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestEineEinstellungSagtWannSieWirkt(t *testing.T) {
	t.Parallel()

	faelle := []struct {
		schluessel string
		wirkung    string
		belegt     string
	}{
		{
			domain.NetzProxyHTTP, domain.WirkungSofort,
			"internal/netz/vorgaben.go — proxyAus baut die Auswahl je Anfrage " +
				"aus StandMit(quelle)",
		},
		{
			domain.NetzFristVerbindung, domain.WirkungSofort,
			"dieselbe Stelle: die Fristen kommen aus Netzvorgaben, und die " +
				"liest App.netzvorgaben bei jedem Ruf neu",
		},
		{
			domain.EinstellungFensterstart, domain.WirkungNeustart,
			"main.go — anwendung.Fensterstand wird einmal gelesen und an " +
				"wails.Run gegeben",
		},
		{
			domain.EinstellungSprache, domain.WirkungSofort,
			"die Oberfläche zeichnet über Sprachraum neu; die Menüleiste " +
				"folgt über Menue.Neu",
		},
		{
			domain.DienstEinstellung(domain.DienstScryfall), domain.WirkungSofort,
			"internal/app/dienstschalterwirkt_test.go hält es in beide " +
				"Richtungen",
		},
	}

	for _, fall := range faelle {
		t.Run(fall.schluessel, func(t *testing.T) {
			t.Parallel()

			if wirkung := domain.EinstellungsWirkung(fall.schluessel); wirkung != fall.wirkung {
				t.Errorf("EinstellungsWirkung(%q) = %q, erwartet %q — %s",
					fall.schluessel, wirkung, fall.wirkung, fall.belegt)
			}
		})
	}
}

// **Ein unbekannter Schlüssel sagt nichts.** Er kann aus einem Bestand von
// morgen kommen; ihm „sofort" anzudichten wäre eine Behauptung über etwas,
// das dieses Programm nicht kennt (§10).
func TestEinUnbekannterSchluesselNenntKeineWirkung(t *testing.T) {
	t.Parallel()

	if wirkung := domain.EinstellungsWirkung("gibtesnicht"); wirkung != "" {
		t.Errorf("EinstellungsWirkung(unbekannt) = %q, erwartet leer", wirkung)
	}
}

// **Und jede Einstellung sagt es.** Ohne diesen Lauf behauptete jede neue
// stillschweigend „sofort" — der Nullwert einer Zeichenkette ist leer, und
// leer hieße hier „nicht gesagt". Dieselbe Gestalt wie
// `internal/werkzeuge/einstellungsgruppen_test.go`: eine neue Einstellung
// erzwingt die Entscheidung, statt in einen Rückfall zu rutschen.
//
// **Was er nicht prüft**: ob die genannte Wirkung die richtige ist. Er
// sieht, dass eine dasteht, nicht dass sie stimmt — das entscheidet, wer
// nachliest, wo der Schlüssel gelesen wird (§15). Die Fundstellen stehen im
// Lauf darüber.
func TestJedeEinstellungSagtIhreWirkung(t *testing.T) {
	t.Parallel()

	erlaubt := map[string]bool{
		domain.WirkungSofort:   true,
		domain.WirkungNeustart: true,
	}

	for _, schluessel := range domain.Einstellungsreihe() {
		wirkung := domain.EinstellungsWirkung(schluessel)

		if !erlaubt[wirkung] {
			t.Errorf("%q nennt die Wirkung %q — erwartet %q oder %q",
				schluessel, wirkung, domain.WirkungSofort,
				domain.WirkungNeustart)
		}
	}
}

// **Die vier, die einen Neustart brauchen — und nur die vier.**
//
// **Eine Aufzählung und kein „mindestens diese"**: wer eine fünfte dazu
// nimmt, ohne sie hier einzutragen, hat sie nicht gemessen. Und wer eine
// herausnimmt, sagt damit, dass sie jetzt sofort wirkt — auch das ist eine
// Aussage, die jemand geprüft haben muss.
//
// **Und seit OPS-04 (#1721) die neun des Protokolls** — gemessen an
// `main.go`: das Protokoll entsteht vor dem Bestand, und die Weiche richtet
// sich nach dem Öffnen einmal ein.
//
// **Und seit OPS-05 (#1722) vier der Messung** — die Spur richtet jeder
// Prozess beim Start ein. Die Zahl der Profile gilt sofort.
func TestGenauDieseEinstellungenBrauchenEinenNeustart(t *testing.T) {
	t.Parallel()

	erwartet := append([]string{
		domain.EinstellungAktualisierung,
		domain.EinstellungFensterstart,
		domain.EinstellungFensterbreite,
		domain.EinstellungFensterhoehe,
	}, domain.Protokollschluessel()...)
	erwartet = append(erwartet, domain.MessungSpur, domain.MessungAbtastung,
		domain.MessungSpurgrenze, domain.MessungAufbewahrung, domain.ZwischenspeicherGrenze)

	// **Und die Dateien des Dienstzeugnisses** (NET-03, #1707): der Dienst
	// liest sie beim Start.
	erwartet = append(erwartet, domain.Zeugnispfade()...)

	// **Und der Kontingentschalter** (QA-08.AK-03, #1739): Dienst und
	// Werkzeugserver bauen ihre Guthaben beim Start
	// (`cmd/mse-rest/dienst.go`, `cmd/mse-mcp/server.go`). **Und die
	// Lauschwahl** (G08, #2272): ein Dienst öffnet seine Lauscher beim Start.
	erwartet = append(erwartet, domain.GrenzenKontingente, domain.NetzLauschen)

	var gefunden []string

	for _, schluessel := range domain.Einstellungsreihe() {
		if domain.EinstellungsWirkung(schluessel) == domain.WirkungNeustart {
			gefunden = append(gefunden, schluessel)
		}
	}

	slices.Sort(gefunden)
	slices.Sort(erwartet)

	if !slices.Equal(gefunden, erwartet) {
		t.Errorf("Neustart verlangen %v, erwartet %v", gefunden, erwartet)
	}
}
