package app

// **Apples Gerätemodell als Anbieter** — AI-05
// (#2042),
// Vermerk [0166](../../docs/entscheidungen/0166-was-fm-serve-belegt.md).
//
// # Drei Fragen, drei Stellen
//
//   - **Ist es aktiviert?** [App.geraeteeinrichtung] fragt `fm` nach Lizenz
//     und Modell — an den Exit-Codes, und für eine Minute gemerkt: der
//     Katalog fragt bei jedem Blick, und zwei Prozesse je Blick wären teuer.
//   - **Was kann es?** [geraetequelle.Modelle] nennt das Modell `system` mit
//     seinen **gemessenen** Fähigkeiten, **ohne den Server zu starten**.
//   - **Was sagt es?** [geraetequelle.Vorschlagen] startet `fm serve` bei
//     Bedarf ([applefm.Dienst]) und fragt — und `Beenden` beendet ihn.
//
// # Je Prozess, nicht je Anwendung
//
// **Kommandozeile, MCP und REST bauen die Anwendung je Aufruf**
// ([App.KIVorschlagInEinemZug]). Hinge der Dienst an ihr, startete jeder Ruf
// ein eigenes `fm serve` und ließe es stehen, und ein Sitzungsschalter
// vergäße sich nach dem Ruf, der ihn umlegt. **Sitzung heißt deshalb die
// Lebensdauer des Prozesses** (AI-05.AK-04): die Naht ist eine je Prozess
// ([prozessgeraet]), und wer den Prozess beendet, beendet mit
// [GeraetBeenden] auch `fm serve`. Ein Lauf setzt mit [App.MitGeraetemodell]
// eine eigene.
//
// # Was es nicht tut
//
// **Eine Lizenz annehmen**, `sudo` rufen, einen Port öffnen — siehe
// `internal/applefm`.

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/applefm"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// geraeteanbieter ist der Name im Katalog.
const geraeteanbieter = applefm.Anbieter

// erkennungsfrist ist, wie lange ein Stand der Erkennung gilt.
const erkennungsfrist = time.Minute

// geraetekontext ist das gemessene Fenster des Gerätemodells — mindestens so
// viele Eingangstokens gehen (Vermerk 0166).
const geraetekontext = 8009

// geraetenaht hält, was die Anwendung über das Gerätemodell weiß.
type geraetenaht struct {
	sperre sync.Mutex

	// schalten ordnet Schaltrufe: Schalter der Anwendung und Zustand des
	// Dienstes wechseln in **einem** Vorgang — N2 des Nachlaufs über Phase
	// 6.10; zwei gegenläufige Rufe ließen sonst die Anwendung „an" sagen,
	// während der Dienst jeden Start aufhielt. Eigene Sperre, weil der Dienst
	// beim Abschalten wartet und `sperre` dabei nicht gehalten werden soll.
	schalten sync.Mutex

	// erkenne, starter und ordner sind die Nähte für Läufe (§3) — leer heißt
	// die echten.
	erkenne func(context.Context) applefm.Einrichtung
	starter applefm.Starter
	ordner  string

	dienst    *applefm.Dienst
	erkannt   applefm.Einrichtung
	erkanntUm time.Time

	// sitzungAus schaltet es für **diesen** Prozess ab — nach einem Neustart
	// gilt wieder die Einstellung (AI-05.AK-04).
	sitzungAus bool

	// zwischenSchalten ist eine Naht für Läufe zwischen den zwei Schritten
	// eines Schaltrufs — nil im Betrieb.
	zwischenSchalten func()
}

// MitGeraetemodell setzt die Nähte — für Läufe ohne `/usr/bin/fm` (§3).
func (a *App) MitGeraetemodell(
	erkenne func(context.Context) applefm.Einrichtung, starter applefm.Starter, ordner string,
) *App {
	a.geraet = &geraetenaht{erkenne: erkenne, starter: starter, ordner: ordner}

	return a
}

// prozessgeraet ist die Naht dieses Prozesses — jede Anwendung ohne eigene
// benutzt sie.
var prozessgeraet = &geraetenaht{}

// GeraetBeenden beendet das `fm serve` dieses Prozesses — für die Flächen,
// die die Anwendung je Aufruf bauen und deshalb kein [App.Beenden] rufen.
func GeraetBeenden() { prozessgeraet.beenden() }

// geraetnaht gibt die Naht dieser Anwendung — die eigene oder die des
// Prozesses.
func (a *App) geraetnaht() *geraetenaht {
	if a.geraet != nil {
		return a.geraet
	}

	return prozessgeraet
}

// geraeteeinrichtung gibt den Stand — gemerkt für [erkennungsfrist].
func (a *App) geraeteeinrichtung(ctx context.Context) applefm.Einrichtung {
	naht := a.geraetnaht()

	naht.sperre.Lock()
	defer naht.sperre.Unlock()

	if naht.erkannt != "" && time.Since(naht.erkanntUm) < erkennungsfrist {
		return naht.erkannt
	}

	erkenne := naht.erkenne
	if erkenne == nil {
		erkenne = func(ctx context.Context) applefm.Einrichtung {
			return applefm.Erkenne(ctx, applefm.Vorhanden, applefm.Exitcode)
		}
	}

	naht.erkannt, naht.erkanntUm = erkenne(ctx), time.Now()

	return naht.erkannt
}

// geraetedienst gibt den Dienst, der `fm serve` hält — einen je Anwendung.
func (a *App) geraetedienst() *applefm.Dienst {
	naht := a.geraetnaht()

	naht.sperre.Lock()
	defer naht.sperre.Unlock()

	if naht.dienst == nil {
		starter, ordner := naht.starter, naht.ordner
		if starter == nil {
			starter = applefm.Prozessstarter
		}

		if ordner == "" {
			ordner = os.TempDir()
		}

		naht.dienst = applefm.NeuerDienst(ordner, starter)
	}

	return naht.dienst
}

// geraetBeenden beendet einen laufenden `fm serve` — beim Beenden der
// Anwendung.
func (a *App) geraetBeenden() { a.geraetnaht().beenden() }

// beenden beendet einen laufenden Dienst dieser Naht.
func (naht *geraetenaht) beenden() {
	naht.sperre.Lock()
	dienst := naht.dienst
	naht.sperre.Unlock()

	if dienst == nil {
		return
	}

	if fehler := dienst.Beende(); fehler != nil {
		meldeDienstbefund(fehler)
	}
}

// geraetebauplan gibt die Quelle des Gerätemodells.
func (a *App) geraetebauplan() func() KILaufzeitquelle {
	return func() KILaufzeitquelle { return geraetequelle{a: a} }
}

// geraetequelle ist das Gerätemodell als [KILaufzeitquelle].
type geraetequelle struct{ a *App }

// Fassung sagt, ob es aktiviert ist — und **wenn nicht, was fehlt**: der
// Satz landet als Befund am Anbieter (AI-05.AK-03).
func (g geraetequelle) Fassung(ctx context.Context) (string, error) {
	einrichtung := g.a.geraeteeinrichtung(ctx)
	if einrichtung == applefm.EinrichtungBereit {
		return "fm serve (macOS)", nil
	}

	return "", fmt.Errorf("%w: %s", domain.ErrValidation, geraetFehlt(einrichtung))
}

// geraetFehlt sagt, was fehlt, und den nächsten Schritt (AI-05.AK-03).
func geraetFehlt(einrichtung applefm.Einrichtung) string {
	switch einrichtung {
	case applefm.EinrichtungLizenz:
		return "die Lizenz von fm ist auf diesem Mac nicht angenommen — " +
			"ein Mensch nimmt sie im Terminal an: sudo fm license"
	case applefm.EinrichtungNichtVerfuegbar:
		return "das Gerätemodell ist nicht verfügbar — " +
			"Apple Intelligence einschalten (Systemeinstellungen)"
	default:
		return "dieser Mac hat kein Gerätemodell — es kommt mit macOS 27 (/usr/bin/fm)"
	}
}

// Modelle nennt das eine Modell mit seinen **gemessenen** Fähigkeiten — ohne
// den Server zu starten (Vermerk 0166).
func (geraetequelle) Modelle(context.Context) ([]domain.KIModell, []domain.Befund, error) {
	beleg := "gemessen über fm serve an einem eigenen Unix-Socket, Vermerk 0166"

	return []domain.KIModell{{
		Alias: applefm.Modell, Anbieter: geraeteanbieter, Ort: domain.KIOrtLokal,
		Laufzeit: "fm serve",
		Geprueft: []domain.KIGeprueft{
			{Faehigkeit: domain.KITextgenerierung, Datum: "2026-10-03", Beleg: beleg},
			{Faehigkeit: domain.KIStrukturierteGenerierung, Datum: "2026-10-03", Beleg: beleg},
		},

		// **Gemessen und nicht angegeben** (AI-01.AK-02): Apple nennt die
		// Zahl nicht; 8.009 Eingangstokens gingen, 9.009 nicht.
		Kontextfenster: geraetekontext,
		KontextBeleg:   "gemessen am 2026-10-03: 8.009 Eingangstokens gehen, 9.009 nicht — Vermerk 0166",
	}}, nil, nil
}

// Vorschlagen fragt das Modell — und startet `fm serve`, wenn es noch nicht
// läuft.
func (g geraetequelle) Vorschlagen(
	ctx context.Context, anweisung, _, auftrag string, schema map[string]any,
) (KIRohantwort, error) {
	klient, fehler := g.a.geraetedienst().Klient(ctx)
	if fehler != nil {
		return KIRohantwort{}, fehler
	}

	antwort, fehler := klient.Rede(ctx, applefm.Frage{System: anweisung, Auftrag: auftrag, Schema: schema})
	if fehler != nil {
		return KIRohantwort{}, fehler
	}

	// **Eine Ablehnung ist ein Ausgang** (AI-05.AK-06): der Lauf darüber
	// macht daraus einen Vorschlag ohne Felder samt Rückfällen.
	if antwort.Ausgang == applefm.AusgangAbgelehnt {
		return KIRohantwort{}, fmt.Errorf("%w (%s)", errKIAbgelehnt, antwort.Naeheres)
	}

	return KIRohantwort{
		Inhalt:        antwort.Inhalt,
		Abgeschnitten: antwort.Ausgang == applefm.AusgangUnvollstaendig,
		Modell:        applefm.Modell,
	}, nil
}

// geraetSitzungAus sagt, ob das Gerätemodell für diesen Prozess abgeschaltet ist.
func (a *App) geraetSitzungAus() bool {
	naht := a.geraetnaht()

	naht.sperre.Lock()
	defer naht.sperre.Unlock()

	return naht.sitzungAus
}

// KIGeraetSitzungAntwort sagt, ob das Gerätemodell in dieser Sitzung gefragt
// werden darf.
type KIGeraetSitzungAntwort struct {
	An bool `json:"an"`

	Fehler *Fehler `json:"fehler"`
}

// KIGeraetemodellFuerSitzung schaltet das Gerätemodell für **diesen Prozess**
// ab oder wieder an — AI-05.AK-04 (#2042).
//
// **Ab heißt sofort**: ein laufendes `fm serve` wird beendet, und bis zum
// Neustart oder bis es wieder angeschaltet wird, entsteht weder eine Anfrage
// noch ein Prozess. **Die Einstellung bleibt, wie sie ist** — wer es über
// Neustarts hinweg aus haben will, stellt `ki.geraetemodell` auf `aus`.
//
// **Dieselbe Rolle wie ein Vorschlag**: wer ein Modell fragen darf, darf es
// für seine Sitzung auch abschalten.
func (a *App) KIGeraetemodellFuerSitzung(ctx context.Context, an bool) KIGeraetSitzungAntwort {
	ctx = a.imProfil(ctx)

	if verweigert := a.darf(ctx, domain.WirkungSchreiben); verweigert != nil {
		return KIGeraetSitzungAntwort{An: !a.geraetSitzungAus(), Fehler: Uebersetze(ctx, verweigert)}
	}

	naht := a.geraetnaht()

	naht.schalten.Lock()
	defer naht.schalten.Unlock()

	naht.sperre.Lock()
	naht.sitzungAus = !an
	zwischen := naht.zwischenSchalten
	naht.sperre.Unlock()

	if zwischen != nil {
		zwischen()
	}

	// **Der Dienst selbst hält den Start auf** (Z2 des Zweitblicks über Phase
	// 6.10): der Schalter oben sagt dem Katalog, dass es aus ist; ein
	// Vorschlag, der den Anbieter davor gewählt hat, fragt den Dienst — und
	// der startet nicht mehr.
	dienst := a.geraetedienst()
	if an {
		dienst.Anschalten()
	} else if fehler := dienst.Abschalten(); fehler != nil {
		meldeDienstbefund(fehler)
	}

	return KIGeraetSitzungAntwort{An: an}
}

// geraeteanbieterAus baut die Zeile des Gerätemodells — AI-05 (#2042).
//
// **Eingeschaltet heißt: gewollt und eingerichtet.** Bis zur Pipeline vom
// 2026-10-04 galt es als eingeschaltet, sobald die Einstellung `automatisch`
// sagte — auf einem Mac ohne `fm` stand der Katalog dann auf
// „unerreichbar" statt „aus". Jetzt steht er auf „aus", und der Satz
// darunter sagt, was fehlt.
//
// **Gefragt wird die Einrichtung nur, wenn es gewollt ist** — ein
// abgeschaltetes Gerät startet keinen Prozess, auch keinen zum Nachsehen.
// **`gewaehlt` bleibt falsch**: »nur lokal« schaltet das Gerät nie ab, und
// der Satz über die Richtlinie träfe nicht zu.
func (a *App) geraeteanbieterAus(ctx context.Context, wahl service.Wahl) kianbieter {
	zeile := kianbieter{name: geraeteanbieter, sitzungAus: a.geraetSitzungAus(), bauen: a.geraetebauplan()}

	switch {
	case wahl.Wirksam != domain.KIGeraetemodellAutomatisch:
	case zeile.sitzungAus:
		zeile.hinweis = "für diese Sitzung abgeschaltet — nach einem Neustart gilt wieder die Einstellung"
	default:
		if einrichtung := a.geraeteeinrichtung(ctx); einrichtung == applefm.EinrichtungBereit {
			zeile.eingeschaltet = true
		} else {
			zeile.hinweis = geraetFehlt(einrichtung)
		}
	}

	return zeile
}
