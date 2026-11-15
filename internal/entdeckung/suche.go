package entdeckung

// **Wer sich meldet** — Werkstück N2 von Phase 5.4
// (#1455).
//
// **Eine eigene Datei**, weil `entdeckung.go` mit ihr über die Zeilengrenze
// aus §2 ginge. Der Schnitt folgt der Frage: dort, **wie man sich
// ankündigt** — hier, **wie man fragt** und was aus den Antworten wird.

import (
	"context"
	"errors"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/hashicorp/mdns"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/peer"
)

// Frage ist, wonach eine Suche fragt.
type Frage struct {
	// Frist ist, wie lange gewartet wird — `0` heißt [Hoechstfrist], und
	// länger geht nicht.
	Frist time.Duration

	// Schnittstelle ist, worüber gefragt wird — `nil` heißt: die des
	// Betriebssystems.
	Schnittstelle *net.Interface
}

// Fund ist ein Gegenüber, das geantwortet hat.
type Fund struct {
	// Kennung ist, was es über sich sagt — **geprüft ist sie noch nicht**.
	//
	// **Geprüft wird sie beim Handschlag**, und was sie darf, entscheidet
	// danach die Kontaktliste. Ein Fund ist eine Einladung, hinzusehen.
	Kennung peer.Kennung

	// Name ist der selbstgewählte — **kein Nachweis**.
	Name string

	// Wirt und Port sind, wohin gewählt würde.
	Wirt string
	Port int

	// Wirtname ist der `.local`-Name, unter dem geantwortet wurde — **leer,
	// wenn keiner mitkam**.
	//
	// **Neben dem Wirt und nicht an seiner Stelle**
	// (#1633):
	// zum **Wählen** ist die Adresse schneller, zum **Merken** ist der Name
	// haltbarer. Eine IP von heute ist nach dem nächsten DHCP-Wechsel
	// falsch, und der Kontakt zeigte dann auf einen Rechner, der inzwischen
	// jemand anderem gehört.
	//
	// **Keine eigene Auflösung und kein Eintrag in ein DNS**: was mDNS
	// ohnehin mitschickt, wird benutzt; mehr nicht.
	Wirtname string
}

// Adresse gibt, was [leitung.Waehle] entgegennimmt.
func (f Fund) Adresse() string {
	return net.JoinHostPort(f.Wirt, strconv.Itoa(f.Port))
}

// Merkadresse gibt, was sich aufzuschreiben lohnt.
//
// **Der Name, wo es einen gibt, sonst die Adresse.** Ein Fund ohne
// `.local`-Namen verlöre sonst seinen Weg — und das wäre schlechter als eine
// Adresse, die veralten kann.
func (f Fund) Merkadresse() string {
	if f.Wirtname == "" {
		return f.Adresse()
	}

	return net.JoinHostPort(f.Wirtname, strconv.Itoa(f.Port))
}

// Ergebnis ist, was eine Suche zurückbringt.
type Ergebnis struct {
	Funde []Fund

	// Verworfen zählt die Antworten, die keine Peerkennung trugen.
	//
	// **Gezählt und nicht verschwiegen** (§16): auf `_mse4mac._tcp` sollte
	// nur dieses Programm antworten, und wenn doch etwas anderes kommt, ist
	// das eine Auskunft. Wer sie unterschlüge, hätte eine leere Liste ohne
	// Erklärung.
	Verworfen int
}

// Suche fragt, wer sich im Netz meldet.
//
// # Warum die Antworten erst gesammelt und dann gegeben werden
//
// **Weil eine Suche ein Ende hat.** mDNS kennt keine Vollständigkeit — es
// kommt, was in der Frist kommt. Wer einen Strom gäbe, müsste an jeder Fläche
// entscheiden, wann er aufhört; so entscheidet es diese eine Stelle.
//
// # Was der Kontext hier leistet — und was nicht
//
// **Er bricht das Warten nicht ab.** Die Bibliothek sagt über `QueryContext`,
// sie achte „auf den Kontext **und** die Frist"; gemessen am 2026-09-14 tut
// sie das nicht:
//
//	ctx 200 ms, Timeout 5 s   → 5,00 s
//	ctx 5 s,    Timeout 200 ms → 0,20 s
//
// **Die Frist entscheidet allein.** Der Kontext wird trotzdem durchgereicht —
// er erreicht die Stellen davor —, und der Satz steht hier, damit niemand von
// ihm erwartet, was er nicht tut (§10).
//
// **Und die Frist ist nicht die Wanduhr.** Dieselbe Messung:
//
//	Timeout 200 ms → 0,20 s     Timeout 2 s → 3,01 s
//	Timeout 1 s    → 1,00 s     Timeout 5 s → 6,00 s
//
// Ab zwei Sekunden kommt eine Sekunde hinzu — die Bibliothek fragt ein zweites
// Mal. Was ein Aufrufer erwarten darf, ist deshalb `Frist + `[Nachlauf].
func Suche(ctx context.Context, frage Frage) (Ergebnis, error) {
	frist := frage.Frist
	if frist <= 0 || frist > Hoechstfrist {
		frist = Hoechstfrist
	}

	// **Eine genannte Schnittstelle schlägt die Wahl.**
	// `MSE_ENTDECKUNGSSCHNITTSTELLE` ist für Läufe da und für den Menschen,
	// der weiß, was er tut — wer sie setzt, meint **diese** und keine
	// andere.
	if frage.Schnittstelle != nil {
		return ueberEine(ctx, frage.Schnittstelle, frist)
	}

	// **Sonst über jede taugliche** — NET-02.NAT-01
	// (#1706).
	//
	// `mdns` ruft `setInterface` nur, wenn eine genannt ist; sonst wählt das
	// Betriebssystem seine „system-assigned multicast interface". Auf einem
	// Rechner mit acht Schnittstellen — gemessen am 2026-09-21 — muss das
	// nicht das WLAN sein, und dann geht die Frage ins falsche Netz. Genau
	// deshalb musste `en0` von Hand eingetragen werden.
	taugliche, fehler := tauglicheLesen()
	if fehler != nil {
		return Ergebnis{}, fehler
	}

	// **Ohne taugliche bleibt es beim Betriebssystem** — das ist, was vorher
	// immer galt, und auf einem Rechner ohne Netz die einzige Antwort, die
	// nicht lügt.
	if len(taugliche) == 0 {
		return ueberEine(ctx, nil, frist)
	}

	return ueberAlle(ctx, taugliche, frist)
}

// ueberAlle fragt über jede Schnittstelle und legt die Antworten zusammen.
//
// # Nebenläufig und nicht nacheinander
//
// **Weil eine Suche bis zu 6,5 Sekunden dauert** (`Hoechstfrist` plus
// [Nachlauf]). Vier Schnittstellen nacheinander wären eine halbe Minute, in
// der niemand weiß, ob etwas geschieht.
//
// # Und ein Fehler nimmt die anderen nicht mit
//
// **Weil genau das der Fall ist, für den es diese Arbeit gibt**: ein VPN
// ohne Route darf nicht verhindern, dass das WLAN antwortet. Gescheitert ist
// die Suche erst, wenn **jede** Schnittstelle scheitert — dann trägt der
// erste Fehler die Auskunft.
func ueberAlle(
	ctx context.Context, taugliche []net.Interface, frist time.Duration,
) (Ergebnis, error) {
	ergebnisse := make([]Ergebnis, len(taugliche))
	fehler := make([]error, len(taugliche))

	var warten sync.WaitGroup

	for stelle := range taugliche {
		warten.Add(1)

		go func() {
			defer warten.Done()

			// **Ein Zusammenbruch ist ein Fehler dieser Schnittstelle**
			// (BK-05c, #2114) — und nicht das Ende des Prozesses. Die
			// anderen fragen weiter; gescheitert ist die Suche erst, wenn
			// jede scheitert.
			defer domain.Auffangen(func(zusammenbruch error) { fehler[stelle] = zusammenbruch })

			ergebnisse[stelle], fehler[stelle] = ueberEine(
				ctx, &taugliche[stelle], frist)
		}()
	}

	warten.Wait()

	return zusammen(ergebnisse, fehler)
}

// zusammen legt die Antworten zusammen — und entdoppelt über die Kennung.
//
// **Dieselbe Regel wie in [sammle]**: ein Gegenüber, das über zwei
// Schnittstellen erreichbar ist, antwortet zweimal und ist einmal da.
func zusammen(ergebnisse []Ergebnis, fehler []error) (Ergebnis, error) {
	var (
		zusammen Ergebnis
		gesehen  = map[peer.Kennung]bool{}
		gelungen bool
		erster   error
	)

	for stelle, ergebnis := range ergebnisse {
		if fehler[stelle] != nil {
			if erster == nil {
				erster = fehler[stelle]
			}

			continue
		}

		gelungen = true
		zusammen.Verworfen += ergebnis.Verworfen

		for _, fund := range ergebnis.Funde {
			if !gesehen[fund.Kennung] {
				gesehen[fund.Kennung] = true
				zusammen.Funde = append(zusammen.Funde, fund)
			}
		}
	}

	if !gelungen {
		return Ergebnis{}, erster
	}

	return zusammen, nil
}

// ueberEine fragt über eine Schnittstelle — und fällt auf IPv4 zurück.
func ueberEine(
	ctx context.Context, schnitt *net.Interface, frist time.Duration,
) (Ergebnis, error) {
	ergebnis, fehler := frageEinmal(ctx, schnitt, frist, false)
	if fehler == nil || !ohneRoute(fehler) {
		return ergebnis, fehler
	}

	// **Noch einmal, und diesmal nur über IPv4** — NET-02.NAT-01
	// (#1774).
	//
	// # Warum das hilft
	//
	// **Weil die Bibliothek beim ersten Schreibfehler abbricht.**
	// `hashicorp/mdns@v1.0.7`, `client.go`, `sendQuery`: erst IPv4, dann
	// IPv6, und jedes `WriteToUDP` mit Fehler kehrt zurück. Gibt macOS für
	// `ff02::fb` ein `EHOSTUNREACH`, weil für IPv6-Multicast keine Route
	// steht, scheitert die ganze Abfrage — **obwohl der IPv4-Ruf schon
	// hinausgegangen war**.
	//
	// # Warum nicht gleich ohne IPv6
	//
	// **Weil IPv6 nicht das Problem ist, sondern seine fehlende Route.** Wer
	// ihn immer abschaltete, fände ein Gegenüber nicht mehr, das **nur**
	// über IPv6 erreichbar ist.
	//
	// # Und warum der zweite Fehler den ersten ablöst
	//
	// **Weil er mehr sagt.** Scheitert auch IPv4 mit „unerreichbar", ist es
	// auf macOS die fehlende Erlaubnis **Lokales Netzwerk** — und das ist
	// etwas anderes als „IPv6 hat keine Route".
	return frageEinmal(ctx, schnitt, frist, true)
}

// frageEinmal stellt eine Abfrage und sammelt ihre Antworten.
func frageEinmal(
	ctx context.Context, schnitt *net.Interface, frist time.Duration,
	nurIPv4 bool,
) (Ergebnis, error) {
	antworten := make(chan *mdns.ServiceEntry, Schlangenlaenge)

	fehler := fragen(ctx, &mdns.QueryParam{
		Service: Dienstart, Domain: Bereich, Timeout: frist,
		Interface: schnitt, Entries: antworten,
		DisableIPv6: nurIPv4,
		Logger:      stilleLogger(),
	})

	close(antworten)

	if fehler != nil {
		return Ergebnis{}, deuteSuchfehler(fehler)
	}

	// **Gelesen wird erst, wenn die Bibliothek fertig ist** — der Grund
	// steht bei [Schlangenlaenge].
	return sammle(antworten), nil
}

// ohneRoute sagt, ob ein Fehler „das Ziel ist nicht erreichbar" bedeutet.
//
// **Dieselben drei wie in [deuteSuchfehler]**, und aus demselben Grund über
// `errors.Is` statt über den Text (§8): die Meldung ist plattformabhängig,
// übersetzt und ändert sich mit dem Betriebssystem.
func ohneRoute(fehler error) bool {
	return errors.Is(fehler, domain.ErrOhneNetzerlaubnis) || errors.Is(fehler, domain.ErrOhneNetzroute)
}
