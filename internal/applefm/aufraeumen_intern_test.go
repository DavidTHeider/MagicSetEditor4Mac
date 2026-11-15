package applefm

// **Ein Waise wird an seiner genauen Befehlszeile erkannt** — Z4 des
// Zweitblicks über Phase 6.10 (#1671). Bis dahin ging der Socketpfad als
// Muster an `pkill -f`: ein `.` im Pfad traf jedes Zeichen, ein `[` machte
// das Muster ungültig, und ein Beenden, das scheiterte, löschte trotzdem den
// Socket — der Waise lief weiter, und niemand fand ihn wieder.

import (
	"context"
	"errors"
	"net"
	"net/http"
	"os"
	"os/exec"
	"slices"
	"testing"
	"time"
)

func TestNurDieGenaueBefehlszeileZaehlt(t *testing.T) {
	t.Parallel()

	befehl := Programm + " serve --socket /tmp/a.b[1]/mse-fm/fm-11.sock"
	liste := "  101 " + befehl + "\n" +
		"  102 " + Programm + " serve --socket /tmp/aXb[1]/mse-fm/fm-11.sock\n" +
		"  103 " + befehl + " --weiteres\n" +
		"  104 /bin/zsh -c " + befehl + "\n" +
		"kaputt\n" +
		"  105\n" +
		"  zehn " + befehl + "\n"

	if kennungen := kennungenMitBefehl(liste, befehl); !slices.Equal(kennungen, []int{101}) {
		t.Errorf("%v", kennungen)
	}
}

// **Ein Beenden, das scheitert, lässt den Socket stehen** — der nächste Start
// findet den Waisen wieder.
func TestEinScheiterndesBeendenLaesstDenSocketStehen(t *testing.T) {
	t.Parallel()

	ordner, fehler := os.MkdirTemp("", "fm")
	if fehler != nil {
		t.Fatalf("%v", fehler)
	}

	t.Cleanup(func() { _ = os.RemoveAll(ordner) }) //nolint:errcheck // nur zum Aufräumen

	unter := ordner + "/mse-fm"
	if fehler := os.Mkdir(unter, 0o700); fehler != nil {
		t.Fatalf("%v", fehler)
	}

	if fehler := os.WriteFile(unter+"/fm-11.sock", nil, 0o600); fehler != nil {
		t.Fatalf("%v", fehler)
	}

	dienst := NeuerDienst(ordner, func(context.Context, ...string) (Prozess, error) {
		return nil, os.ErrNotExist
	})
	dienst.lebt = func(int) bool { return false }
	dienst.verwaisteBeenden = func(context.Context, string) error { return errors.New("lebt weiter") }

	_, _ = dienst.Klient(t.Context()) //nolint:errcheck // der Start scheitert absichtlich

	if _, fehler := os.Stat(unter + "/fm-11.sock"); fehler != nil {
		t.Errorf("der Socket ist fort: %v", fehler)
	}
}

// **Der echte Weg, an einem harmlosen Programm**: ein `sleep` mit einer
// Dauer, die niemand sonst wählt, wird an seiner Befehlszeile gefunden und
// beendet — ein zweites mit anderer Dauer bleibt.
func TestGenauBeendenTrifftNurDenGemeinten(t *testing.T) {
	t.Parallel()

	gemeint := exec.CommandContext(t.Context(), "/bin/sleep", "31.4159")
	anderer := exec.CommandContext(t.Context(), "/bin/sleep", "31.41592")

	for _, befehl := range []*exec.Cmd{gemeint, anderer} {
		if fehler := befehl.Start(); fehler != nil {
			t.Fatalf("%v", fehler)
		}
	}

	t.Cleanup(func() { _ = anderer.Process.Kill() }) //nolint:errcheck // nur zum Aufräumen

	if fehler := genauBeenden(t.Context(), "/bin/sleep 31.4159"); fehler != nil {
		t.Fatalf("%v", fehler)
	}

	ende := make(chan error, 1)
	go func() { ende <- gemeint.Wait() }()

	select {
	case fehler := <-ende:
		// **Durch ein Signal und nicht von selbst** — `sleep 31.4159`
		// endete sonst erst nach einer halben Minute.
		var abbruch *exec.ExitError
		if !errors.As(fehler, &abbruch) || abbruch.Exited() {
			t.Errorf("der gemeinte endete nicht durch ein Signal: %v", fehler)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("der gemeinte lebt noch")
	}

	if !prozessLebt(anderer.Process.Pid) {
		t.Error("der andere ist mit beendet worden")
	}

	// **Ohne Treffer ist es kein Fehler.**
	if fehler := verwaistenBeenden(t.Context(), "/gibt/es/nicht/fm-1.sock"); fehler != nil {
		t.Errorf("%v", fehler)
	}
}

// **Die zwei Fehlerwege**: die Prozessliste lässt sich nicht lesen, und ein
// Prozess lässt sich nicht beenden — beides kommt beim Rufer an.
func TestDieFehlerwegeDesGenauenBeendens(t *testing.T) {
	t.Parallel()

	abgebrochen, abbrechen := context.WithCancel(t.Context())
	abbrechen()

	if fehler := genauBeenden(abgebrochen, "/bin/sleep 1"); fehler == nil {
		t.Error("ohne Prozessliste kein Fehler")
	}

	if fehler := toete(1 << 30); fehler == nil {
		t.Error("ein Prozess, den es nicht gibt, beendet sich nicht")
	}

	// **Und ein Signal, das scheitert, kommt beim Rufer an** — an der
	// eigenen Kennung mit ihrer wirklichen Befehlszeile, und das Signal ist
	// eine Attrappe.
	eigener := befehlVon(t.Context(), os.Getpid())
	if eigener == "" {
		t.Fatal("die eigene Befehlszeile ist nicht zu lesen")
	}

	if fehler := beendeKennungen(t.Context(), eigener, []int{os.Getpid()},
		func(int) error { return errors.New("kein Recht") }); fehler == nil {
		t.Error("ein gescheitertes Signal ist kein Fehler")
	}

	// **Eine Kennung, die es nicht gibt, hat keine Befehlszeile.**
	if befehl := befehlVon(t.Context(), 1<<30); befehl != "" {
		t.Errorf("%q", befehl)
	}
}

// **Die Kennung wird unmittelbar vor dem Signal noch einmal geprüft** — N4
// des Nachlaufs über Phase 6.10: zwischen der Liste und dem Signal kann ein
// Waise enden und seine Kennung an einen fremden Prozess gehen. Hier trägt
// die Kennung schon einen anderen Befehl, und sie bleibt unberührt.
func TestEineWiedervergebeneKennungBleibtUnberuehrt(t *testing.T) {
	t.Parallel()

	fremd := exec.CommandContext(t.Context(), "/bin/sleep", "31.41593")
	if fehler := fremd.Start(); fehler != nil {
		t.Fatalf("%v", fehler)
	}

	t.Cleanup(func() { _ = fremd.Process.Kill() }) //nolint:errcheck // nur zum Aufräumen

	getroffen := 0
	if fehler := beendeKennungen(t.Context(), "/bin/sleep 99.1", []int{fremd.Process.Pid},
		func(int) error { getroffen++; return nil }); fehler != nil || getroffen != 0 {
		t.Errorf("%v, getroffen %d", fehler, getroffen)
	}

	if !prozessLebt(fremd.Process.Pid) {
		t.Error("der fremde Prozess ist fort")
	}
}

// **Ohne Treffer bleibt ein Socket, an dem noch jemand lauscht** — N5 des
// Nachlaufs: weicht die Darstellung der Befehlszeile ab, fände die Liste
// den Waisen nicht, und sein Socket wäre fort, während er weiterläuft.
func TestEinSocketMitHoererBleibt(t *testing.T) {
	t.Parallel()

	ordner, fehler := os.MkdirTemp("", "fm")
	if fehler != nil {
		t.Fatalf("%v", fehler)
	}

	t.Cleanup(func() { _ = os.RemoveAll(ordner) }) //nolint:errcheck // nur zum Aufräumen

	unter := ordner + "/mse-fm"
	if fehler := os.Mkdir(unter, 0o700); fehler != nil {
		t.Fatalf("%v", fehler)
	}

	horcher, err := (&net.ListenConfig{}).Listen(t.Context(), "unix", unter+"/fm-11.sock")
	if err != nil {
		t.Fatalf("%v", err)
	}

	// **Ein Waise, der antwortet wie `fm serve`** — und den die Liste nicht
	// fand.
	go func() {
		_ = http.Serve(horcher, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { //nolint:gosec,errcheck // Probe
			_, _ = w.Write([]byte(`{"models":[{"available":true,"name":"system"}]}`)) //nolint:errcheck // Probe
		}))
	}()

	t.Cleanup(func() { _ = horcher.Close() }) //nolint:errcheck // nur zum Aufräumen

	dienst := NeuerDienst(ordner, func(context.Context, ...string) (Prozess, error) {
		return nil, os.ErrNotExist
	})
	dienst.lebt = func(int) bool { return false }
	dienst.verwaisteBeenden = func(context.Context, string) error { return nil }

	_, _ = dienst.Klient(t.Context()) //nolint:errcheck // der Start scheitert absichtlich

	if _, fehler := os.Stat(unter + "/fm-11.sock"); fehler != nil {
		t.Errorf("der Socket ist fort, obwohl dort jemand lauscht: %v", fehler)
	}
}
