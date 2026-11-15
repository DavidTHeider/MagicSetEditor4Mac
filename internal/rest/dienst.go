package rest

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Dienst ist der REST-Zugang.
//
// **Er hält keinen Zustand außer seiner Marke und seiner Drossel.** Was er
// tut, tun die Anwendungsfälle dahinter; er übersetzt HTTP auf sie und ihre
// Fehler zurück auf HTTP.
type Dienst struct {
	marke string

	// pfadvertrag urteilt über jeden Pfad, der hereinkommt — OPS-02.AK-04
	// (#1719); `nil` heißt: niemand urteilt.
	pfadvertrag Pfadvertrag

	// eigeneNamen sind die zusätzlichen Namen, unter denen dieser Dienst
	// erreichbar ist — leer, solange er auf der Schleife bindet.
	//
	// **Sie kommen vom Composition Root** (N6, #1572): welche Adressen
	// dieser Rechner hat, weiß `internal/leitung` und nicht dieses Blatt.
	eigeneNamen []string

	// **Zwei Budgets, nicht eines** (Zweitblick über Phase U).
	//
	// Mit einem konnte jeder lokale Prozess 60 Anfragen **ohne** Marke
	// schicken und danach war der legitime Aufrufer für eine Minute draußen —
	// ohne dass der Angreifer je eine Marke gebraucht hätte.
	//
	// `fremde` zählt die Fehlversuche, `eigene` die Arbeit. Beide sind nötig:
	// die erste schützt das Raten der Marke, die zweite den Bestand.
	fremde *Drossel

	// **Seit O5 sind es Guthaben und keine zweite Drossel**
	// (#1429):
	// je Principal eines und eines für alle, mit Kosten je Ruf und einer
	// Wartezeit, die stimmt — und eine Sperre je **vorgezeigtem** Nachweis
	// statt eines Zählers für alle Ratenden zusammen.
	// **Beide kommen vom Wurzelknoten** (`MitHaushalt`), denn `rest` kennt
	// `service` nicht; ohne sie gibt es kein Guthaben und keine Sperre — nur
	// die Drossel für alle, wie vor O5.
	grenzen     domain.Haushaltsgrenzen
	kontingente Kontingentnaht
	sperre      Sperrnaht

	naehte Naehte

	// wache prüft Zugriffstoken — **oder ist nil, und dann gibt es keine**.
	//
	// **Nil ist der Normalfall und kein Versehen**: der Dienst läuft auch ohne
	// technische Profile, und dann ist die Prozessmarke der einzige Nachweis.
	// Wer sie setzt, sagt damit „dieser Bestand kennt Token" (Werkstück 10,
	// #1351).
	wache Tokenwache

	// messen nimmt jeden Ruf auf — OPS-05 (#1722); ohne sie misst niemand.
	messen func(context.Context, Messung)

	// schreibwache fragt vor jeder schreibenden Anfrage — `schreibwache.go`.
	schreibwache func(ctx context.Context) error

	// schemafreigabe hält jede Anfrage, bis sie fertig ist (`schreibwache.go`).
	schemafreigabe func(ctx context.Context) (func(), error)

	// hochladegrenze ersetzt [Hochladegrenze], wenn gesetzt —
	// `dateiuebertragung.go`.
	hochladegrenze int64

	// fortschritt legt einem Auftrag einen Melder in den Kontext —
	// `auftragswege.go` (FMT-03, #1692).
	fortschritt func(context.Context, domain.Fortschrittsmelder) context.Context

	// schalter fragt vor jedem Weg, ob sein Bereich an ist — `schalter.go`.
	schalter Schalternaht

	// ausgaenge merkt sich, wem eine Ausgabe gehört — `ausgangsbuch.go`
	// (SEC-P6, #2002).
	ausgaenge *ausgangsbuch
}

// Grenze und Fenster der Drossel.
//
// **60 in der Minute**, und die Zahl ist nach dem Zweck gewählt, nicht nach
// einer Erwartung an die Last: gegen das Raten einer 64-Zeichen-Marke ist jede
// Zahl in dieser Größenordnung dasselbe, und für einen Menschen oder ein
// Skript am selben Rechner ist eine pro Sekunde reichlich.
const (
	grenze  = 60
	fenster = time.Minute
)

// NeuerDienst baut den Zugang.
//
// **Die Uhr wird eingesetzt** (§16) — die Drossel braucht sie, und ein Test
// mit echter Wartezeit flackert.
func NeuerDienst(marke string, jetzt func() time.Time, naehte Naehte) *Dienst {
	return &Dienst{
		marke:     marke,
		fremde:    NeueDrossel(grenze, fenster, jetzt),
		grenzen:   domain.Vorgabegrenzen(),
		naehte:    naehte,
		ausgaenge: neuesAusgangsbuch(jetzt),
	}
}

// MitHaushalt setzt Guthaben und Sperre — der Wurzelknoten reicht die
// Umsetzung aus `service` herein (#1429).
func (d *Dienst) MitHaushalt(kontingente Kontingentnaht, sperre Sperrnaht) *Dienst {
	d.kontingente = kontingente
	d.sperre = sperre

	return d
}

// abbuchen fragt das Guthaben — ohne eines ist alles erlaubt.
//
// **Und was nichts kostet, fragt gar nicht erst** (#2003): ein Abbruch nimmt
// Last, und er bekam bei leerem Guthaben ein 429, während der Auftrag zu Ende
// lief.
func (d *Dienst) abbuchen(principal string, klasse domain.Kostenklasse) (bool, time.Duration) {
	if d.kontingente == nil || klasse.Kosten() <= 0 {
		return true, 0
	}

	return d.kontingente.Abbuchen(principal, klasse, false)
}

// gesperrt fragt die Sperre — ohne eine ist niemand gesperrt.
func (d *Dienst) gesperrt(nachweis string) (bool, time.Duration) {
	if d.sperre == nil {
		return false, 0
	}

	return d.sperre.Gesperrt(nachweis)
}

func (d *Dienst) fehlgeschlagen(nachweis string) {
	if d.sperre != nil {
		d.sperre.Fehlgeschlagen(nachweis)
	}
}

func (d *Dienst) gelungen(nachweis string) {
	if d.sperre != nil {
		d.sperre.Gelungen(nachweis)
	}
}

// MitGrenzen setzt andere Grenzen — für Tests und für einen Prozess, der
// engere will. Die Guthaben beginnen damit von vorn.
func (d *Dienst) MitGrenzen(grenzen domain.Haushaltsgrenzen) *Dienst {
	d.grenzen = grenzen

	return d
}

// MitEigenenNamen nennt die Namen, unter denen dieser Dienst erreichbar ist.
//
// **Zusätzlich und nicht statt** `127.0.0.1` und `localhost`: die beiden
// gelten immer, und eine Bindung jenseits der Schleife kommt dazu. Der Grund
// steht bei [HerkunftTaugtUnter].
func (d *Dienst) MitEigenenNamen(namen ...string) *Dienst {
	d.eigeneNamen = namen

	return d
}

// MitTokenwache setzt die Prüfung für Zugriffstoken.
//
// **Als eigener Ruf und nicht als Parameter von [NeuerDienst]**, weil ein
// Dienst ohne sie vollständig ist: die Marke allein trägt jeden Aufrufer, der
// den Prozess gestartet hat. Ein Pflichtparameter zwänge jeden Test und jeden
// Aufrufer zu einem `nil`, das nichts aussagt.
func (d *Dienst) MitTokenwache(wache Tokenwache) *Dienst {
	d.wache = wache

	return d
}

// Behandler gibt den fertigen Weg samt aller Wachen.
//
// # Die Reihenfolge der Wachen ist die Reihenfolge ihrer Aussagekraft
//
//  1. **Herkunft.** Eine Seite im Browser soll gar nicht erst bis zur Marke
//     kommen; ihre Anfrage ist als solche zu erkennen.
//  2. **Marke — und je nach Ausgang ein anderes Budget.**
//
// **Hier stand bis zum 2026-08-23 eine Drossel vor der Marke**, und der
// Zweitblick hat gezeigt, was sie durchließ: jeder lokale Prozess konnte 60
// Anfragen ohne Marke schicken und danach war der legitime Aufrufer für eine
// Minute draußen — ohne je eine Marke gebraucht zu haben.
//
// **Die Reihenfolge war trotzdem richtig**, und deshalb ist sie nicht
// umgedreht worden: eine Drossel **hinter** der Marke zählte nur die
// erfolgreichen Versuche und ließe das Raten ungezählt. Was fehlte, war das
// zweite Budget.
func (d *Dienst) Behandler() http.Handler {
	weg := http.NewServeMux()
	d.wegeEintragen(weg)

	return d.mitMessung(weg, d.mitHerkunft(d.mitMarke(weg, d.mitSchaltern(weg,
		d.mitHochladen(weg, d.mitPfadvertrag(weg, d.mitAuftrag(weg)))))))
}

// mitHerkunft weist ab, was nach einem Browser aussieht.
func (d *Dienst) mitHerkunft(weiter http.Handler) http.Handler {
	return http.HandlerFunc(func(schreiben http.ResponseWriter, anfrage *http.Request) {
		if !HerkunftTaugtUnter(anfrage, d.eigeneNamen) {
			nein(schreiben, http.StatusForbidden, "herkunft",
				"diese Anfrage kommt nicht von einem Aufrufer, der gemeint ist")

			return
		}

		weiter.ServeHTTP(schreiben, anfrage)
	})
}

// zuViele weist eine Anfrage ab, die über ihr Budget geht.
func zuViele(schreiben http.ResponseWriter, warten time.Duration) {
	// **Die echte Restzeit, nicht eine feste 60** (#1429) — aufgerundet auf
	// ganze Sekunden, und nie null: `Retry-After: 0` hieße „sofort", und das
	// stimmte nicht.
	sekunden := max(int(math.Ceil(warten.Seconds())), 1)
	schreiben.Header().Set("Retry-After", strconv.Itoa(sekunden))
	nein(schreiben, http.StatusTooManyRequests, "drossel",
		fmt.Sprintf("zu viele Anfragen — in %d s wieder", sekunden))
}

// mitMarke prüft den Nachweis und zählt je nach Ausgang aus dem einen oder dem
// anderen Budget.
//
// **Zwei Arten von Nachweis, eine Kopfzeile** — die Weiche steht in
// `tokenwache.go`.
func (d *Dienst) mitMarke(weg *http.ServeMux, weiter http.Handler) http.Handler {
	return http.HandlerFunc(func(schreiben http.ResponseWriter, anfrage *http.Request) {
		// **Die Freigabe vor allem anderen** — auch vor dem Nachweis, der
		// schon schreibt (Zweitblick über Phase 2, Befund 1).
		frei, verwehrt := d.freigabeHolen(anfrage.Context())
		if verwehrt != nil {
			fehlerhaft(schreiben, verwehrt)

			return
		}

		defer frei()

		erlaubt, principal, gut := d.nachweisTaugt(schreiben, anfrage)
		if !gut {
			return
		}

		if verwehrt := d.schreibenVerwehrt(erlaubt); verwehrt != nil {
			fehlerhaft(schreiben, verwehrt)

			return
		}

		klasse := kostenklasseAus(weg, anfrage)
		if ok, warten := d.abbuchen(principal, klasse); !ok {
			zuViele(schreiben, warten)

			return
		}

		// **Rumpf und Antwort sind begrenzt** — dieselbe Zahl für beide, denn
		// beide binden Speicher, den ein Maschinenzugang nicht bestimmen soll.
		// **Und der Principal reist mit**: ein Auftrag gehört ihm (FMT-03).
		// **Eine übertragene Datei und eine abgeholte Ausgabe haben eigene
		// Grenzen** (FMT-03.AK-03): sie gehen auf die Platte statt in den
		// Speicher, und ein Bogen ist größer als jede JSON-Antwort.
		erlaubt = erlaubt.WithContext(mitPrincipal(erlaubt.Context(), principal))
		erlaubt.Body = http.MaxBytesReader(schreiben, erlaubt.Body, d.rumpfgrenzeFuer(erlaubt))

		if !istAbholung(erlaubt) {
			schreiben = begrenzt(schreiben, d.grenzen.Antwort)
		}

		weiter.ServeHTTP(schreiben, erlaubt)
	})
}

// nachweisTaugt prüft Marke oder Zugriffstoken und gibt die Anfrage zurück,
// die weitergehen darf.
func (d *Dienst) nachweisTaugt(
	schreiben http.ResponseWriter, anfrage *http.Request,
) (*http.Request, string, bool) {
	nachweis, gefunden := nachweisAus(anfrage)

	if gefunden && d.wache != nil && istZugriffstoken(nachweis) {
		return d.tokenTaugt(schreiben, anfrage, nachweis)
	}

	if !MarkeStimmt(d.marke, anfrage) {
		d.keinNachweis(schreiben, nachweis)

		return nil, "", false
	}

	d.gelungen(nachweis)

	return anfrage, principalMarke, true
}

// tokenTaugt prüft ein Zugriffstoken und hängt sein Profil an die Anfrage.
func (d *Dienst) tokenTaugt(
	schreiben http.ResponseWriter, anfrage *http.Request, nachweis string,
) (*http.Request, string, bool) {
	ausgestattet, identitaet, fehler := d.wache.Pruefe(
		anfrage.Context(), nachweis, schreibend(anfrage))
	if fehler == nil {
		d.gelungen(nachweis)

		return anfrage.WithContext(ausgestattet),
			principalFuer(identitaet, nachweis), true
	}

	// **Eine zu knappe Erlaubnis ist kein Rateversuch.** Wer ein gültiges
	// Token hat, ist kein Fremder; sein Fehlschlag zählt deshalb aus dem
	// eigenen Budget und bekommt einen Satz, aus dem er lernen kann. Alles
	// andere zählt wie eine falsche Marke — und klingt auch so.
	if ErlaubnisZuKnapp(fehler) {
		d.gelungen(nachweis)

		if ok, warten := d.abbuchen(principalFuer(identitaet, nachweis), domain.KostenLesen); !ok {
			zuViele(schreiben, warten)

			return nil, "", false
		}

		nein(schreiben, http.StatusForbidden, "erlaubnis", fehler.Error())

		return nil, "", false
	}

	d.keinNachweis(schreiben, nachweis)

	return nil, "", false
}

// keinNachweis meldet einen fehlgeschlagenen Nachweis — **einen Satz für Marke
// und Token gleichermaßen**, damit aus der Antwort nicht hervorgeht, welche
// Art von Nachweis beinahe gestimmt hätte.
func (d *Dienst) keinNachweis(schreiben http.ResponseWriter, nachweis string) {
	// **Erst die Sperre dieses Nachweises, dann zählen, dann melden.** Wer
	// rät, soll die Zahl seiner Versuche spüren, bevor er erfährt, dass er
	// falsch lag — und wer gesperrt ist, zählt nicht weiter, sonst stimmte
	// die Auskunft „in n Sekunden" nie (#1429).
	// **Ohne vorgezeigten Nachweis gibt es nichts zu sperren**: eine Anfrage
	// ohne Kopfzeile rät nicht, sie fragt — und bekommt ihr 401. Für sie bleibt
	// die Drossel für alle; gesperrt wird, was jemand wirklich vorzeigt.
	if nachweis != "" {
		if gesperrt, rest := d.gesperrt(nachweis); gesperrt {
			zuViele(schreiben, rest)

			return
		}

		d.fehlgeschlagen(nachweis)
	}

	if !d.fremde.Durchlassen() {
		zuViele(schreiben, fenster)

		return
	}

	// **`WWW-Authenticate` gehört dazu**: ein 401 ohne diese Kopfzeile
	// sagt nicht, womit man es versuchen soll.
	schreiben.Header().Set("WWW-Authenticate", `Bearer realm="mse"`)
	nein(schreiben, http.StatusUnauthorized, "marke",
		"diese Anfrage trägt keine gültige Marke")
}
