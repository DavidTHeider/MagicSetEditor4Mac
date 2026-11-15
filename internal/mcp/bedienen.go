package mcp

// **Die Bedienung über stdio** — herausgelöst aus `server.go`, als sie
// lernen musste, während eines Rufs weiter zuzuhören (FMT-03.AK-02,
// #1692).
//
// # Der Reihe nach, und trotzdem ansprechbar
//
// **Werkzeuge laufen weiter nacheinander.** Ein Ruf kann den Stand ändern,
// aus dem der nächste liest — `schreiben.profil.wechseln` tut genau das —,
// und zwei gleichzeitig wären eine Zusage, die kein Werkzeug gibt. Was sich
// ändert: **während** einer läuft, wird weiter gelesen. Ein
// `notifications/cancelled` hält ihn an, ein `ping` wird gleich beantwortet;
// jede andere Anfrage wartet, bis er fertig ist, und kommt in ihrer
// Reihenfolge dran.

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"
)

// ergebnis ist, was ein laufender Ruf zurückgibt.
type ergebnis struct {
	antwort   Antwort
	antworten bool
}

// Bediene liest Nachrichten aus `ein` und schreibt Antworten nach `aus`.
//
// **Eine Nachricht je Zeile.** Das ist die Gestalt, in der MCP über stdio
// spricht; die Länge steht nicht davor.
//
// Der Rückgabewert ist `nil`, wenn der Strom regulär endet — das Gegenüber hat
// dann aufgelegt, und das ist kein Fehler. **Was bis dahin kam, wird noch
// bedient**: `printf … | mse-mcp` schickt alles auf einmal und schließt.
//
// **Ein Abbruch des Kontexts beendet sie** — auch wenn die Eingabe offen
// bleibt (BK-07c, #2117), und danach geht keine Zeile mehr hinaus, auch nicht
// von einem Ruf, der noch läuft.
func (s *Server) Bediene(ctx context.Context, ein io.Reader, aus io.Writer) error {
	ausgang := &schliessbarerAusgang{aus: aus}
	defer ausgang.schliesse()

	aus = ausgang
	ctx = mitSender(ctx, func(meldung any) error { return s.schreibe(aus, meldung) })

	zeilen, lesefehler := leseZeilen(ein)

	// **Wer vorzeitig aufhört, lässt den Leser zu Ende lesen** — sonst hinge
	// er an einer Zeile, die niemand mehr abnimmt. Nach einem regulären Ende
	// ist der Kanal zu, und das Leeren endet sofort.
	defer func() { go leeren(zeilen) }()

	var (
		laeuft  chan ergebnis
		wartend [][]byte
	)

	for {
		for laeuft == nil && len(wartend) > 0 {
			var fehler error
			if laeuft, fehler = s.beginne(ctx, wartend[0], aus); fehler != nil {
				return fehler
			}

			wartend = wartend[1:]
		}

		if zeilen == nil && laeuft == nil {
			return <-lesefehler
		}

		select {
		// **Bis zum 2026-10-06 fehlte dieser Fall**: `mse-mcp` hängt SIGINT
		// und SIGTERM an den Kontext, und über stdio mit offener Eingabe
		// wartete die Bedienung trotzdem auf die nächste Zeile.
		case <-ctx.Done():
			return s.letzteAntwort(ctx, laeuft, aus)

		case zeile, offen := <-zeilen:
			switch {
			case !offen:
				zeilen = nil

			// **Ohne laufenden Ruf kommt die Zeile in die Reihe** — die
			// Schleife oben beginnt sie sofort.
			case laeuft == nil:
				wartend = append(wartend, zeile)

			default:
				sofort, fehler := s.nebenbei(ctx, zeile, aus)
				if fehler != nil {
					return fehler
				}

				if !sofort {
					wartend = append(wartend, zeile)
				}
			}

		case fertig := <-laeuft:
			laeuft = nil

			if fertig.antworten {
				if fehler := s.schreibe(aus, fertig.antwort); fehler != nil {
					return fehler
				}
			}
		}
	}
}

// beginne bedient eine Zeile. Ein Werkzeugruf läuft nebenher und meldet sich
// über den Kanal; alles andere wird gleich beantwortet.
func (s *Server) beginne(ctx context.Context, zeile []byte, aus io.Writer) (chan ergebnis, error) {
	var anfrage Anfrage
	if json.Unmarshal(zeile, &anfrage) == nil && !anfrage.istMeldung() && anfrage.Methode == "tools/call" {
		fertig := make(chan ergebnis, 1)
		lauf, beenden := s.beginneLauf(ctx, anfrage.ID)

		go func() {
			antwort, antworten := s.vollende(lauf, beenden, anfrage)
			fertig <- ergebnis{antwort: antwort, antworten: antworten}
		}()

		return fertig, nil
	}

	antwort, antworten := s.behandle(ctx, zeile)
	if !antworten {
		return nil, nil
	}

	return nil, s.schreibe(aus, antwort)
}

// nebenbei bedient, was während eines Rufs nicht warten soll: Meldungen —
// darunter der Abbruch — und `ping`. Der erste Wert sagt, ob es geschah.
func (s *Server) nebenbei(ctx context.Context, zeile []byte, aus io.Writer) (bool, error) {
	var anfrage Anfrage

	lesbar := json.Unmarshal(zeile, &anfrage) == nil
	if !lesbar || (!anfrage.istMeldung() && anfrage.Methode != "ping") {
		return false, nil
	}

	antwort, antworten := s.behandle(ctx, zeile)
	if !antworten {
		return true, nil
	}

	return true, s.schreibe(aus, antwort)
}

// leseZeilen liest den Strom nebenher. Der Kanal der Zeilen schließt am
// Ende; danach steht im zweiten, woran es lag — `nil` für ein reguläres Ende.
//
// **Ohne Puffer**: solange ein Ruf läuft und niemand abnimmt, liest der Leser
// nicht weiter, und das Gegenüber spürt es an seiner Leitung — dieselbe
// Rückstauung wie vorher, als der Server erst nach der Antwort las.
func leseZeilen(ein io.Reader) (zeilenstrom <-chan []byte, fehlerstrom <-chan error) {
	zeilen := make(chan []byte)
	lesefehler := make(chan error, 1)

	go func() {
		defer close(zeilen)

		leser := bufio.NewScanner(ein)

		// **Eine großzügige Zeilengrenze.** Ein Kartentext mit Zeilenumbrüchen
		// kommt als eine JSON-Zeile; die Vorgabe von 64 KB reicht dafür, ein
		// ganzes Set in einer Antwort aber nicht.
		leser.Buffer(make([]byte, 0, 64*1024), Hoechstenachricht)

		for leser.Scan() {
			if len(leser.Bytes()) > 0 {
				zeilen <- append([]byte(nil), leser.Bytes()...)
			}
		}

		if fehler := leser.Err(); fehler != nil {
			lesefehler <- fmt.Errorf("MCP lesen: %w", fehler)

			return
		}

		lesefehler <- nil
	}()

	return zeilen, lesefehler
}

// leeren nimmt ab, was noch kommt.
func leeren(zeilen <-chan []byte) {
	for range zeilen { //nolint:revive // abnehmen ist hier der Zweck
	}
}

// schliessbarerAusgang nimmt Zeilen an, bis die Bedienung endet — danach
// keine mehr, auch nicht von einem Ruf, der noch läuft und Fortschritt meldet.
// **Prüfen und Schreiben unter einer Sperre**: sonst passte zwischen beides
// das Ende, und eine Zeile ginge doch noch hinaus.
type schliessbarerAusgang struct {
	sperre sync.Mutex
	zu     bool
	aus    io.Writer
}

// errBedienungBeendet sagt, dass die Bedienung schon endete.
var errBedienungBeendet = errors.New("die Bedienung ist beendet")

func (a *schliessbarerAusgang) Write(zeile []byte) (int, error) {
	a.sperre.Lock()
	defer a.sperre.Unlock()

	if a.zu {
		return 0, errBedienungBeendet
	}

	return a.aus.Write(zeile)
}

func (a *schliessbarerAusgang) schliesse() {
	a.sperre.Lock()
	a.zu = true
	a.sperre.Unlock()
}

// Nachfrist ist, wie lange ein Ende auf die Antwort des laufenden Rufs
// wartet.
var Nachfrist = 2 * time.Second

// letzteAntwort schreibt beim Ende noch die Antwort des laufenden Rufs — **wie
// `http.Server.Shutdown` am Dienst** (M02, #2346). Die Fassungswache bricht den
// Kontext ab und lehnt den Ruf ab, der sie fragte; ohne diese Frist ging die
// Ablehnung nie hinaus. Ein Ruf, der in der [Nachfrist] nicht fertig wird,
// bleibt ohne Antwort — sein Kontext ist ohnehin abgebrochen.
func (s *Server) letzteAntwort(ctx context.Context, laeuft chan ergebnis, aus io.Writer) error {
	if laeuft == nil {
		return ctx.Err()
	}

	select {
	case fertig := <-laeuft:
		if fertig.antworten {
			return errors.Join(ctx.Err(), s.schreibe(aus, fertig.antwort))
		}
	case <-time.After(Nachfrist):
	}

	return ctx.Err()
}
