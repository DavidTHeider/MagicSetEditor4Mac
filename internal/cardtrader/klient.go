// Package cardtrader liest bei CardTrader — Stammdaten, Blueprints und die
// Angebote des Marktplatzes.
//
// **Werkstück A4 von Phase 5.3**
// (#1443).
//
// # Was dieses Paket ist
//
// Ein **Blatt** (§2): es kennt `internal/domain` und `internal/netz` und sonst
// nichts aus diesem Projekt. Die drei Schichten aus §16 liegen getrennt —
// Transport hier, fremde DTOs und Zuordnung je Gegenstand daneben.
//
// # Was es **nicht** ist
//
// **Kein Weg, der schreibt.** Die Aufzählung [wege] führt seit `T11`
// (#1476)
// **sieben** Adressen — vier zum Marktplatz, drei zum eigenen Konto —, und
// alle sieben sind `GET`: kein `POST`, kein `PUT`, kein `PATCH`, kein
// `DELETE`.
//
// **Was der Anbieter zum Schreiben anbietet, ist bekannt und bleibt
// ungebaut** — Angebote anlegen, Warenkorb, Kauf, Versand. Der Grund steht
// in [0122](../../docs/entscheidungen/0122-ein-fremdes-konto-wird-gelesen.md);
// die Kontowege dahinter stehen zusätzlich hinter einem eigenen Schalter,
// der aus beginnt.
package cardtrader

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
)

// Vorgabeadresse ist die Grundadresse der Schnittstelle.
//
// **Nachgelesen am 2026-09-11** (0095). Als Konstante und nicht fest im Code:
// ein Lauf gegen eine eigene Gegenstelle braucht sie, und ohne sie ginge jeder
// Test ins Netz (§3).
const Vorgabeadresse = "https://api.cardtrader.com/api/v2"

// Familie ist der Name, unter dem Preise von hier laufen.
const Familie domain.Preisquelle = "cardtrader"

// Abstand ist der Mindestabstand zwischen zwei Anfragen.
//
// # Warum eine Sekunde, wo zehn erlaubt wären
//
// **Nachgelesen** (0095): 200 Anfragen je 10 Sekunden über alle Endpunkte. Am
// Marktplatz stehen **beide** Zahlen im selben Abschnitt — „limited to 1 call
// per second" und „a rate limit of 10 requests per second" —, und dieser
// Widerspruch steht in der Quelle (nachgelesen am 2026-09-19 im Wortlaut,
// Befund 8 des Zweitblicks über Phase 5.6).
//
// Gedrosselt wird auf die engere: was dieses Programm fragt, sind
// wenige Abrufe je Bedienhandlung, und zwischen 1/s und 10/s liegt hier keine
// spürbare Wartezeit. Eine falsch geratene Grenze kostet einen Ausschluss,
// eine zu vorsichtige nichts (§16, Regel 4).
const Abstand = time.Second

// hoechstens ist die Größe, ab der eine Antwort abgewiesen wird.
//
// **Blueprints eines großen Sets sind einige Megabyte.** Zehn sind großzügig
// und trotzdem eine Grenze — ohne sie hinge die Anwendung an dem, was eine
// Gegenstelle schickt.
const hoechstens int64 = 10 << 20

// wege ist die Positivliste — alles, was dieses Paket überhaupt rufen kann.
//
// **Eine Aufzählung und kein Vorsatz.** Wer eine Adresse hinzufügt, tut es
// hier und sieht dabei, dass sie mit `GET` gerufen wird; ein Server-Spion in
// `internal/flaechen` weist nach, dass über alle vier Flächen zusammen genau
// null schreibende Methoden hinausgehen.
var wege = struct {
	Spiele, Erweiterungen, Blueprints, Markt string

	// Die drei Wege zum eigenen Konto — Werkstück T11 von Phase 5.6
	// (#1476).
	//
	// **Sie stehen in derselben Liste und sind aus demselben Holz**: alle
	// drei `GET`. Was der Anbieter zum **Schreiben** anbietet, ist bekannt
	// und bleibt ungebaut ([0122](../../docs/entscheidungen/0122-ein-fremdes-konto-wird-gelesen.md)).
	Konto, EigeneAngebote, Bestellungen string
}{
	Spiele:        "/games",
	Erweiterungen: "/expansions",
	Blueprints:    "/blueprints/export",
	Markt:         "/marketplace/products",

	Konto:          "/info",
	EigeneAngebote: "/products/export",
	Bestellungen:   "/orders",
}

// Wege gibt die Positivliste — für Wächter, die sie nachzählen.
func Wege() []string {
	return []string{
		wege.Spiele, wege.Erweiterungen, wege.Blueprints, wege.Markt,
		wege.Konto, wege.EigeneAngebote, wege.Bestellungen,
	}
}

// Klient liest bei CardTrader.
//
// **Adresse, Netzklient und Zugang kommen von außen** (§16): sonst braucht
// jeder Lauf das Netz und einen echten Schlüsselbund.
type Klient struct {
	adresse string
	klient  *netz.Klient
	zugang  Zugang
	drossel *netz.Drossel
}

// NeuerKlient baut ihn.
func NeuerKlient(adresse string, klient *netz.Klient, zugang Zugang) *Klient {
	if adresse == "" {
		adresse = Vorgabeadresse
	}

	return &Klient{adresse: strings.TrimRight(adresse, "/"), klient: klient, zugang: zugang}
}

// MitDrossel setzt die Drossel, die dieser Prozess mit den anderen teilt.
func (k *Klient) MitDrossel(drossel *netz.Drossel) *Klient {
	k.drossel = drossel

	return k
}

// hole ruft einen Weg der Positivliste und zerteilt seine Antwort.
//
// **Der Zugang wird vor der Anfrage geprüft** (§16, Regel 6 und 0063): ohne
// Marke antwortet CardTrader mit 401, und eine Anfrage, von der das feststeht,
// ist Last für den Anbieter und eine Wartezeit für nichts.
func (k *Klient) hole(ctx context.Context, weg string, abfrage url.Values, ziel any) error {
	if fehlt := k.zugang.Fehlt(); fehlt != "" {
		return fmt.Errorf("%w: %s", domain.ErrValidation, fehlt)
	}

	adresse, fehler := url.JoinPath(k.adresse, weg)
	if fehler != nil {
		return domain.MitGrund(domain.ErrValidation, fehler,
			"die Adresse für %s bauen", weg)
	}

	if len(abfrage) > 0 {
		adresse += "?" + abfrage.Encode()
	}

	antwort, holfehler := netz.Hole(ctx, k.klient, k.drossel, k.kenner(), adresse, hoechstens)
	if holfehler != nil {
		return holfehler
	}

	if antwort.Status != netz.StatusGut {
		return fmt.Errorf("%w: CardTrader antwortet auf %s mit %d",
			domain.ErrValidation, weg, antwort.Status)
	}

	if fehler := json.Unmarshal(antwort.Rumpf, ziel); fehler != nil {
		return domain.MitGrund(domain.ErrValidation, fehler,
			"die Antwort von CardTrader auf %s lesen", weg)
	}

	return nil
}

// kenner ist, womit dieses Programm sich nennt — samt Marke (§16, Satz 2).
func (k *Klient) kenner() netz.Kenner {
	return netz.Kenner{
		Name:    "MagicSetEditor4Mac",
		Annahme: "application/json",
		Marke:   k.zugang.Marke,
	}
}
