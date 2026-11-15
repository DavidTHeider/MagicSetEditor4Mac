package app

// **Was gilt und woher es kommt** — Werkstück N5 von Phase 5.4
// (#1571),
// Werkstück 3.
//
// # Warum die Herkunft mitreist
//
// **Weil „es steht ein Proxy" ohne „woher" keine Auskunft ist.** Wer im
// Fenster `proxy.firma:3128` liest, muss wissen, ob er ihn dort ändern kann
// oder ob er aus der Umgebung **dieses Prozesses** stammt — dort ändert ihn
// kein Fenster, und ein Feld, das sich einzutippen anbietet und nichts
// bewirkt, ist schlimmer als keines.
//
// # Und warum die Fristen dabeistehen
//
// **Weil sie sich anders verhalten als der Proxy**, und das gehört gesagt:
// der Proxy wird je Anfrage gelesen, die Fristen sitzen im Wähler und gelten
// ab dem **nächsten** Klienten. Eine Fläche, die beides gleich aussehen
// ließe, versprach mehr, als ihr Rumpf hält (§10).

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
)

// NetzfeldDTO ist ein wirksamer Wert samt seiner Herkunft.
type NetzfeldDTO struct {
	Wert string `json:"wert"`

	// Herkunft ist „einstellung", „umgebung" oder „ohne".
	Herkunft string `json:"herkunft"`
}

// NetzstandAntwort sagt, wie dieses Programm **jetzt** nach außen geht.
type NetzstandAntwort struct {
	ProxyHTTP  NetzfeldDTO `json:"proxyHTTP"`
	ProxyHTTPS NetzfeldDTO `json:"proxyHTTPS"`
	ProxyAus   NetzfeldDTO `json:"proxyAus"`

	// Fristen sind die drei Zeiten in **Sekunden**.
	FristVerbindung int `json:"fristVerbindung"`
	FristKopf       int `json:"fristKopf"`
	FristStillstand int `json:"fristStillstand"`

	// ProxyBenutzt sagt, ob überhaupt einer im Spiel ist.
	//
	// **Berechnet und nicht aus den Feldern erschlossen** — die Oberfläche
	// baute den Satz sonst aus zwei Bedingungen zusammen, und das ist die
	// Gestalt, die irgendwann jemand falsch zusammensetzt.
	ProxyBenutzt bool `json:"proxyBenutzt"`

	Fehler *Fehler `json:"fehler"`
}

// Netzstand gibt den wirksamen Netzweg — **rein lesend**.
func (a *App) Netzstand(ctx context.Context) NetzstandAntwort {
	ctx = a.imProfil(ctx)

	vorgaben := a.netzvorgaben(ctx)
	stand := netz.StandMit(ctx, func(context.Context) netz.Proxyvorgabe { return vorgaben.Proxy })

	return NetzstandAntwort{
		ProxyHTTP:       feldAus(domain.NetzProxyHTTP, stand.HTTP),
		ProxyHTTPS:      feldAus(domain.NetzProxyHTTPS, stand.HTTPS),
		ProxyAus:        feldAus(domain.NetzProxyAus, stand.Aus),
		FristVerbindung: int(vorgaben.Fristen.Verbindung.Seconds()),
		FristKopf:       int(vorgaben.Fristen.Kopf.Seconds()),
		FristStillstand: int(vorgaben.Fristen.Stillstand.Seconds()),
		ProxyBenutzt:    stand.Benutzt(),
	}
}

// feldAus übersetzt ein Feld samt Herkunft — **ohne Kennwort**, auch aus
// der Umgebung (S13.NAT-02, #1706).
func feldAus(schluessel string, feld netz.Proxyfeld) NetzfeldDTO {
	return NetzfeldDTO{
		Wert:     domain.Einstellungsanzeige(schluessel, feld.Wert),
		Herkunft: string(feld.Herkunft),
	}
}

// Netzstand am Fenster — **rein lesend** (`docs/SCHREIBWEGE.md`).
//
// **Geändert wird über `EinstellungSetzen`** wie jede andere Einstellung; ein
// zweiter Schreibweg daneben wäre eine zweite Fassung derselben Frage.
func (b *Bindungen) Netzstand(ctx context.Context) NetzstandAntwort {
	return b.app(ctx).Netzstand(b.ruf(ctx))
}
