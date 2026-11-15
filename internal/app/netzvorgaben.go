package app

// **Die eine Stelle, an der die Netzeinstellungen gelesen werden** —
// Werkstück N5 von Phase 5.4
// (#1571).
//
// # Warum eine Stelle und nicht elf
//
// **Weil es elf `netz.NeuerKlient(…)` gibt.** Jede davon selbst lesen zu
// lassen hieße elf Gelegenheiten, es zu vergessen — und das Vergessen fiele
// niemandem auf, weil ein Klient ohne Proxy einfach direkt geht. Genau diese
// Gestalt hatte der Befund, aus dem `dienstklient` entstanden ist: „der
// Transport hält den Schalter ohnehin" stimmte für einen von vier Diensten.
//
// # Und warum je Ruf und nicht einmal
//
// **Weil eine Einstellung sich ändert.** Wer im Fenster einen Proxy einträgt,
// erwartet nicht, die Anwendung neu starten zu müssen. Ein festgehaltener
// Klient hätte den alten Transport; deshalb wird hier bei jedem Bau gelesen —
// derselbe Grund, aus dem `dienstklient` seine Erlaubnis je Anfrage fragt.

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/fernleitung"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
)

// Netzvorgaben sind Proxy und Fristen, wie sie **jetzt** gelten.
type Netzvorgaben struct {
	Proxy   netz.Proxyvorgabe
	Fristen netz.Fristen
}

// netzvorgaben liest sie aus dem Bestand — die üblichen, wo nichts steht.
//
// **Ein Lesefehler hält nichts auf.** Was hier zurückkommt, sind Vorgaben;
// ohne sie ginge eine Anfrage **direkt** hinaus statt gar nicht, und das ist
// das Verhalten, das dieses Programm ohne jede Einstellung ohnehin hat. Eine
// Netzanfrage an einem unlesbaren Bestand scheitern zu lassen wäre eine
// zweite, schlechtere Meldung über dasselbe Problem.
func (a *App) netzvorgaben(ctx context.Context) Netzvorgaben {
	if a.einstellungen == nil {
		return Netzvorgaben{Fristen: netz.UeblicheFristen()}
	}

	werte, fehler := a.einstellungen.Alle(ctx)
	if fehler != nil {
		return Netzvorgaben{Fristen: netz.UeblicheFristen()}
	}

	return NetzvorgabenAus(werte)
}

// NetzvorgabenAus deutet einen gelesenen Satz Einstellungen.
//
// **Exportiert, weil die Flächen ohne Fenster dieselbe Deutung brauchen**
// (§16, Werkstück 2 des Issues): `mse`, `mse-mcp` und `mse-rest` sind
// Klienten wie das Fenster. Eine zweite Deutung dort wäre die Gestalt, in der
// ein Proxy an einer Fläche gilt und an der anderen nicht.
func NetzvorgabenAus(werte map[string]string) Netzvorgaben {
	ueblich := netz.UeblicheFristen()

	return Netzvorgaben{
		Proxy: netz.Proxyvorgabe{
			HTTP:  werte[domain.NetzProxyHTTP],
			HTTPS: werte[domain.NetzProxyHTTPS],
			Aus:   werte[domain.NetzProxyAus],
		},
		Fristen: netz.Fristen{
			Verbindung: domain.Netzfrist(
				werte[domain.NetzFristVerbindung], ueblich.Verbindung),
			Kopf: domain.Netzfrist(werte[domain.NetzFristKopf], ueblich.Kopf),
			Stillstand: domain.Netzfrist(
				werte[domain.NetzFristStillstand], ueblich.Stillstand),
		},
	}
}

// Proxyquelle gibt die Quelle, die ein Klient **je Anfrage** fragt.
//
// **Und nicht die Vorgabe selbst.** Ein Klient entsteht beim Verdrahten; wer
// danach einen Proxy einträgt, erwartet nicht, die Anwendung neu starten zu
// müssen. Die Funktion liest bei jeder Anfrage neu.
//
// **Ohne Bestand gibt sie eine leere Vorgabe** — dann gilt die Umgebung, und
// das ist das Verhalten dieser Anwendung ohne jede Einstellung.
func (a *App) Proxyquelle() netz.Proxyquelle {
	// **Ohne die Umschaltsperre** (Nachlauf über Phase 6.7, N2): das Relais
	// des Fernhorchens fragt sie aus seinen eigenen Goroutinen — auch während
	// ein Wechsel unter der Sperre das Fernhorchen beendet und auf sie
	// wartet. Der Proxy ist eine Angabe dieses Geräts und nicht des Profils.
	return func(ctx context.Context) netz.Proxyvorgabe {
		return a.netzvorgaben(a.mitProfilFuerDenProxy(ctx)).Proxy
	}
}

// mitProfilFuerDenProxy gibt einen Kontext, der ein Profil trägt — das der
// Anfrage, das dieser Anwendung oder **das gezeigte des Bestands**.
//
// **Das dritte für die Flächen ohne Fenster** (Befund 1 des Zweitblicks über
// Phase D): eine Anwendung der Kommandozeile merkt sich kein Profil, und
// tailcat stellt die Anfrage an DERP mit einem eigenen Kontext. Ohne Profil
// scheitert das Lesen still, und die Anfrage ginge am Proxy vorbei.
//
// **Eine verschlossene Sitzung bekommt auch hier keines** (neuer Befund des
// Nachlaufs über Phase D): `imProfilOhneSperre` hängt dann gerade keines an
// (ID-02), und der Rückfall auf das gezeigte hätte dessen Einstellungen
// trotzdem gelesen. Gesperrt gilt die Umgebung — der Preis: ein Weg, der
// während der Sperre läuft, geht am eingestellten Proxy vorbei.
func (a *App) mitProfilFuerDenProxy(ctx context.Context) context.Context {
	ctx = a.imProfilOhneSperre(ctx)
	if _, fehlt := gezeigtesProfil(ctx); fehlt == nil || a.sitzungVerschlossen() {
		return ctx
	}

	db := a.laufendeVerbindung()
	if db == nil {
		return ctx
	}

	if mitProfil, fehler := mitGezeigtemProfil(ctx, db); fehler == nil {
		return mitProfil
	}

	return ctx
}

// fernproxywahl ist, über welchen Proxy der Weg zum Relais geht — dieselbe
// Wahl wie für jeden anderen Klienten (NET-02.AK-01, #1706). Bis dahin ging
// er über keinen, und hinter einem Proxy kam der Tunnel nie zustande.
func (a *App) fernproxywahl() fernleitung.Proxywahl {
	return netz.Proxyauswahl(a.Proxyquelle())
}

// proxyquelleAus baut eine Quelle über einem Einstellungsleser.
//
// **Für die Stellen ohne `*App`** — `dienstklient`, `mtgjsonbulk` und der
// Scryfall-Klient sind Funktionen des Pakets und bekommen ihre [Dienstwahl]
// hereingereicht. Kann sie Einstellungen lesen, kann sie auch den Proxy
// nennen; kann sie es nicht, gilt die Umgebung.
func proxyquelleAus(wahl Dienstwahl) netz.Proxyquelle {
	leser, kann := wahl.(Netzwegleser)
	if !kann {
		return nil
	}

	// **Mit dem Kontext der Anfrage** (D06, #2237): er trägt das Profil, und
	// ohne es scheitert das Lesen — still, denn ein Lesefehler heißt hier
	// „wie ohne Einstellung". Bis dahin stand hier `context.Background()`,
	// und Scryfall, die Kartendienste und der Preisbezug gingen am
	// eingestellten Proxy vorbei.
	return func(ctx context.Context) netz.Proxyvorgabe {
		return leser.Netzvorgaben(ctx).Proxy
	}
}

// Netzwegleser ist die freiwillige Zusatzfähigkeit einer [Dienstwahl].
//
// **Freiwillig wie [Riegeltraeger] daneben**, und aus demselben Grund: eine
// zweite Pflichtmethode an `Dienstwahl` bräche jede Vorrichtung, die nur den
// Schalter beantwortet — und die Antwort darauf wäre in jeder von ihnen
// dieselbe leere Vorgabe.
type Netzwegleser interface {
	Netzvorgaben(ctx context.Context) Netzvorgaben
}

// netzklientMit baut einen Klienten mit **eigenen** Fristen.
//
// **Es gibt nur diesen einen und keinen ohne Fristen.** Der erste Anlauf
// hatte beide; der ohne blieb **ungerufen**, weil jede Stelle in diesem Paket
// ihre eigenen Fristen mitbringt oder `netz.UeblicheFristen()` nennt. Ein
// Bauweg ohne Aufrufer ist keine Anwendungsfunktion (§2) — `golangci-lint`
// hat ihn angestrichen.
//
// **Für die KI-Dienste**: ihre Kopf- und Stillstandsfristen sind eine
// Eigenschaft *dieses* Dienstes — ein Modell, das eine Minute rechnet, ist
// kein hängender Server — und bleiben deshalb bei ihm. **Der Proxy ist es
// nicht** und zieht weg; das ist wörtlich Werkstück 1 des Issues.
func (a *App) netzklientMit(fristen netz.Fristen) *netz.Klient {
	return netz.NeuerKlientMit(fristen, a.Proxyquelle())
}
