package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Server bedient MCP über zwei Ströme.
type Server struct {
	werkzeuge []Werkzeug

	// pfadvertrag urteilt über Pfade in den Argumenten (OPS-02.AK-04).
	pfadvertrag func(ctx context.Context, pfad string) error

	// fassung ist die Fassung der Anwendung, wie die Vorstellung sie meldet.
	//
	// **Von außen und nicht als Konstante hier** (#896). `internal/mcp` ist ein
	// Blatt (§2) und kennt weder `wails.json` noch die Anwendung; die Fassung
	// kommt aus dem Composition Root, der sie ohnehin kennt.
	//
	// Ein fester Platzhalter wie `"0.0.0"` sähe aus wie eine Fassung und wäre
	// keine — ein Gegenüber, das Fähigkeitserkennung darauf stützt, bekäme eine
	// falsche Antwort statt gar keiner.
	fassung string

	// schloss deckt den Ausgabestrom.
	//
	// **Auch wenn heute nacheinander bedient wird.** Ein zweiter Aufrufer
	// entsteht in dem Augenblick, in dem jemand die Bedienung nebenläufig macht
	// — und dann schrieben zwei Antworten ineinander, was am Ausgabestrom nicht
	// mehr zu trennen ist.
	schloss sync.Mutex

	// schalter fragt vor jedem Werkzeug, ob sein Bereich an ist (OPS-10).
	schalter func(ctx context.Context, werkzeug string) error

	// uhr zählt die falschen Marken über HTTP ([fehlversuche]); nil heißt die
	// Wanduhr.
	uhr func() time.Time

	// ausstatten stattet den Kontext **je Ruf** aus.
	//
	// # Warum je Ruf und nicht einmal beim Start
	//
	// **Weil eine MCP-Sitzung lange lebt** und ein Werkzeug den Stand ändern
	// kann, aus dem der nächste liest — `schreiben.profil.wechseln` tut genau
	// das (#1351). Ein Kontext, der beim Start ausgestattet wurde, trüge
	// danach das **alte** Profil, und `lesen.sets` gäbe die Sets von vorhin.
	//
	// **Gefunden hat das der Flächenlauf, nicht ein Test dieses Pakets** — und
	// auch der erst, nachdem er beide Rufe in **eine** Sitzung gelegt hatte:
	// die Vorrichtung startete je Ruf einen eigenen Prozess und war damit
	// großzügiger als der Betrieb (§3).
	//
	// **`nil` ist erlaubt** — dann reist der Kontext unverändert weiter, so
	// wie jede andere Naht, die eine Fläche nicht einrichtet.
	ausstatten func(context.Context) (context.Context, error)

	// wache prüft **je Ruf**, ob dieser Aufruf das Werkzeug seiner Art
	// benutzen darf — Werkstück O2 von Phase 5.2 (#1426).
	//
	// **Je Ruf und nicht beim Start**, aus demselben Grund wie `ausstatten`:
	// eine Sitzung lebt lange, und ein Token, das zwischen zwei Rufen
	// zurückgezogen wird, abläuft oder seinen Umfang verliert, soll den
	// **nächsten** Ruf nicht mehr tragen. `nil` heißt: niemand wacht — der
	// lokale Altbetrieb ohne Token.
	wache func(ctx context.Context, art Art) (context.Context, error)

	// schreibwache fragt vor jedem schreibenden Werkzeug — `schreibwache.go`.
	schreibwache func(ctx context.Context) error

	// schemafreigabe hält jeder Ruf, bis er fertig ist (`schreibwache.go`).
	schemafreigabe func(ctx context.Context) (func(), error)

	// kontingent sagt vor jedem Werkzeugruf, ob das Guthaben reicht — und
	// sonst, wie lange es dauert (#1429). `nil` heißt: kein Guthaben, wie ein
	// Test es baut.
	//
	// **Mit dem Kontext des Rufs** — Befund 3 des Nachlaufs zum Zweitblick
	// über Phase 5.2. Ohne ihn kann keine Naht wissen, **wer** ruft, und ein
	// Guthaben ohne Rufer ist eines für den Prozess.
	kontingent func(ruf context.Context, art Art, name string) (bool, time.Duration)

	// kennung liest die Vorgangskennung aus dem Kontext eines Rufs —
	// Werkstück O9 von Phase 5.2
	// (#1433).
	//
	// **Eine Naht und kein Import.** `internal/mcp` ist ein Blatt (§2): es
	// kennt `domain` und sonst nichts aus diesem Projekt, und wo die Kennung
	// im Kontext liegt, weiß `internal/service`. Die Wurzel reicht die Frage
	// herein — dieselbe Gestalt wie [Server.MitAusstattung].
	kennung func(context.Context) string

	// fehlerform gibt einen Fehler maschinenlesbar (F03, #2342) — nil heißt:
	// nur der Satz.
	fehlerform func(context.Context, error) any

	// messen nimmt jeden Werkzeugruf auf — OPS-05 (#1722); ohne sie misst
	// niemand.
	messen func(context.Context, Messung)

	// fortschritt legt einen Melder in den Kontext — `fortschritt.go`.
	fortschritt func(context.Context, domain.Fortschrittsmelder) context.Context

	// laufende sind die Rufe, die gerade arbeiten — `abbruch.go`.
	laufsperre sync.Mutex
	laufende   map[string]*laufenderRuf
}

// MitKontingent setzt die Naht zum Guthaben.
func (s *Server) MitKontingent(
	kontingent func(ruf context.Context, art Art, name string) (bool, time.Duration),
) *Server {
	s.kontingent = kontingent

	return s
}

// MitWache setzt die Prüfung, die jeder Ruf vor seinem Werkzeug besteht.
func (s *Server) MitWache(
	wache func(ctx context.Context, art Art) (context.Context, error),
) *Server {
	s.wache = wache

	return s
}

// MitAusstattung setzt die Naht, die jeden Ruf ausstattet.
func (s *Server) MitAusstattung(
	ausstatten func(context.Context) (context.Context, error),
) *Server {
	s.ausstatten = ausstatten

	return s
}

// NeuerServer baut den Server über einer Werkzeugliste.
func NeuerServer(werkzeuge []Werkzeug) *Server { return &Server{werkzeuge: werkzeuge} }

// MitFassung gibt dem Server die Fassung, die er meldet.
//
// **Eine Naht wie die übrigen `Mit…`**: der Composition Root verdrahtet sie,
// und ein Test setzt sie ohne die Anwendung.
func (s *Server) MitFassung(fassung string) *Server {
	s.fassung = fassung

	return s
}

// schreibe gibt eine Antwort oder eine Meldung als eine Zeile aus.
func (s *Server) schreibe(aus io.Writer, antwort any) error {
	s.schloss.Lock()
	defer s.schloss.Unlock()

	// **Ohne Fehlerweg für das Serialisieren**: eine Antwort besteht aus
	// Zeichenketten, Zahlen, Listen und Abbildungen — nichts davon kann
	// `Marshal` scheitern lassen. Ein Zweig, den kein Aufruf betreten kann,
	// wäre eine Behauptung darüber, was im Ernstfall geschieht (§3).
	roh, _ := json.Marshal(antwort) //nolint:errcheck // siehe darüber

	if _, err := aus.Write(append(roh, '\n')); err != nil {
		return fmt.Errorf("MCP antworten: %w", err)
	}

	return nil
}

// behandle beantwortet eine Nachricht; der zweite Wert sagt, ob zu antworten ist.
func (s *Server) behandle(ctx context.Context, zeile []byte) (Antwort, bool) {
	var anfrage Anfrage

	if err := json.Unmarshal(zeile, &anfrage); err != nil {
		return Antwort{
			JSONRPC: "2.0",
			Fehler:  &Protokollfehler{Code: codeZerteilung, Meldung: "die Nachricht ist kein JSON"},
		}, true
	}

	if anfrage.istMeldung() {
		// Eine Meldung bekommt keine Antwort — auch keine Fehlermeldung. Das
		// gilt für `notifications/initialized` ebenso wie für alles Unbekannte.
		if anfrage.Methode == "notifications/cancelled" {
			s.abbruchHoeren(anfrage.Parameter)
		}

		return Antwort{}, false
	}

	switch anfrage.Methode {
	case "initialize":
		return antwortAuf(anfrage, s.vorstellung()), true

	case "tools/list":
		return antwortAuf(anfrage, map[string]any{"tools": s.liste()}), true

	case "tools/call":
		return s.rufe(ctx, anfrage)

	case "ping":
		return antwortAuf(anfrage, map[string]any{}), true
	}

	return fehlerAuf(anfrage, codeUnbekannt, "unbekannte Methode %q", anfrage.Methode), true
}

// vorstellung ist die Antwort auf `initialize`.
func (s *Server) vorstellung() map[string]any {
	return map[string]any{
		"protocolVersion": fassung,
		"capabilities":    map[string]any{"tools": map[string]any{}},
		"serverInfo": map[string]any{
			"name":    "magicseteditor4mac",
			"version": s.gemeldeteFassung(),
		},
		// **Der Hinweis steht in der Vorstellung**, nicht nur an den Werkzeugen:
		// ein Gegenüber liest ihn, bevor es das erste Werkzeug sieht.
		"instructions": s.anweisung(),
	}
}

// gemeldeteFassung gibt die Fassung — oder sagt, dass sie fehlt.
//
// **„unbekannt" und nicht `0.0.0`.** Eine Vorgabe, die wie eine Fassung
// aussieht, ist schlimmer als gar keine: sie sieht nach einer Auskunft aus, und
// ein Gegenüber prüft sie nicht nach.
func (s *Server) gemeldeteFassung() string {
	if s.fassung == "" {
		return "unbekannt"
	}

	return s.fassung
}

// anweisung sagt, was die Werkzeugliste hergibt.
//
// **Abgeleitet und nicht danebengeschrieben** (#896): eine zweite Aufzählung
// neben der Werkzeugliste ist die, die niemand nachführt (§9).
func (s *Server) anweisung() string {
	var schreibende []string

	for _, werkzeug := range s.werkzeuge {
		if werkzeug.Art == Schreibend {
			schreibende = append(schreibende, werkzeug.Name)
		}
	}

	satz := "Lesende Werkzeuge heißen »lesen.…«, schreibende »schreiben.…«. "

	if len(schreibende) == 0 {
		satz += "Es gibt kein schreibendes Werkzeug. "
	} else {
		satz += "Diese ändern etwas: " + strings.Join(schreibende, ", ") + ". "
	}

	return satz + "Kein Werkzeug ändert mehr als die Karten, die es nennt. " +
		"Jede Änderung sagt, wie sie zurückzunehmen ist."
}

// liste gibt die Werkzeuge, wie MCP sie erwartet.
func (s *Server) liste() []map[string]any {
	beschreibungen := make([]map[string]any, 0, len(s.werkzeuge))
	for _, werkzeug := range s.werkzeuge {
		beschreibungen = append(beschreibungen, werkzeug.beschreibung())
	}

	return beschreibungen
}

// fuehreAus sucht das Werkzeug und führt es hinter allen Wachen aus.
func (s *Server) fuehreAus(ctx context.Context, anfrage Anfrage) Antwort {
	lauf := neuerMesslauf(ctx)
	defer s.melden(lauf)

	var parameter struct {
		Name      string          `json:"name"`
		Argumente json.RawMessage `json:"arguments"`
	}

	if err := json.Unmarshal(anfrage.Parameter, &parameter); err != nil {
		return fehlerAuf(anfrage, codeParameter, "die Argumente sind kein JSON: %v", err)
	}

	// **Die Freigabe vor allem anderen** — auch vor der Tokenwache, die schon
	// schreibt (Zweitblick über Phase 2, Befund 1).
	frei, verwehrt := s.freigabeHolen(ctx)
	if verwehrt != nil {
		lauf.messung.Ausgang = domain.RufGescheitert

		return antwortAuf(anfrage, s.fehlerInhalt(ctx, verwehrt, verwehrt.Error()))
	}

	defer frei()

	for _, werkzeug := range s.werkzeuge {
		if werkzeug.Name != parameter.Name {
			continue
		}

		lauf.werkzeug(werkzeug)

		gerichtet, ausstattungsfehler := s.ausgestattet(ctx)
		if ausstattungsfehler != nil {
			lauf.messung.Ausgang = domain.RufGescheitert

			return antwortAuf(anfrage, s.fehlerInhalt(ctx, ausstattungsfehler, ausstattungsfehler.Error()))
		}

		lauf.ctx = gerichtet

		if s.wache != nil {
			bewacht, wachfehler := s.wache(gerichtet, werkzeug.Art)
			if wachfehler != nil {
				return antwortAuf(anfrage, s.fehlerInhalt(gerichtet, wachfehler, wachfehler.Error()))
			}

			gerichtet = bewacht
		}

		// **Hinter der Wache**, wie am Dienst hinter der Marke: wer sich
		// nicht ausweist, erfährt nicht, was abgeschaltet ist (OPS-10).
		if abgeschaltet := s.abgeschaltet(gerichtet, werkzeug); abgeschaltet != nil {
			return antwortAuf(anfrage, s.fehlerInhalt(gerichtet, abgeschaltet, abgeschaltet.Error()))
		}

		if verwehrt := s.schreibenVerwehrt(gerichtet, werkzeug); verwehrt != nil {
			return antwortAuf(anfrage, s.fehlerInhalt(gerichtet, verwehrt, verwehrt.Error()))
		}

		// **Das Guthaben vor der Arbeit** (#1429): wer über seinem Budget
		// liegt, bekommt die Wartezeit — als Fehler des Werkzeugs, denn das
		// Protokoll war in Ordnung. **Und mit seiner Form** (Zweitblick über
		// Phase 1 von 0.9.0, Befund 5): bis dahin baute dieser Weg sein
		// Ergebnis von Hand, und `structuredContent` fehlte.
		if s.kontingent != nil {
			if ok, warten := s.kontingent(gerichtet, werkzeug.Art, werkzeug.Name); !ok {
				grenze := domain.NeueKontingentgrenze(warten)

				return antwortAuf(anfrage, s.fehlerInhalt(gerichtet, grenze, grenze.Error()))
			}
		}

		if verwehrt := s.pfadeVerwehrt(gerichtet, werkzeug, parameter.Argumente); verwehrt != nil {
			return antwortAuf(anfrage, s.fehlerInhalt(gerichtet, verwehrt, verwehrt.Error()))
		}

		text, err := werkzeug.Ruf(gerichtet, parameter.Argumente)

		// **Geschrieben ist geschrieben** (§7, Regel 3; #1970): ein
		// [Nachgeladen] ist ein geglückter Ruf, nach dem nur das Nachlesen
		// scheiterte — kein `isError`, sonst versuchte es das Modell noch
		// einmal.
		var nachgeladen *Nachgeladen
		if errors.As(err, &nachgeladen) {
			lauf.messung.Ausgang = domain.RufGelungen

			return antwortAuf(anfrage, nachgeladenerInhalt(nachgeladen))
		}

		if err != nil {
			lauf.messung.Ausgang = ausgangVon(err)

			// **Ein fachlicher Fehler ist kein Protokollfehler.** MCP will ihn
			// im Ergebnis mit `isError`, damit das Gegenüber ihn dem Modell
			// zeigen kann statt die Verbindung für kaputt zu halten.
			//
			// Was der Text sagt, steht in `fehlertext.go` (#1812, QA-09).
			return antwortAuf(anfrage, s.fehlerInhalt(gerichtet, err, fehlertext(err)))
		}

		lauf.messung.Ausgang = domain.RufGelungen

		return antwortAuf(anfrage, inhalt(text, false))
	}

	return fehlerAuf(anfrage, codeUnbekannt, "unbekanntes Werkzeug %q", parameter.Name)
}

// ErrKeinWerkzeug sagt, dass ein Aufruf kein Werkzeug getroffen hat.
var ErrKeinWerkzeug = errors.New("kein solches Werkzeug")

// ausgestattet gibt den Kontext, mit dem ein Werkzeug arbeitet.
func (s *Server) ausgestattet(ctx context.Context) (context.Context, error) {
	if s.ausstatten == nil {
		return ctx, nil
	}

	return s.ausstatten(ctx)
}
