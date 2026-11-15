package leitung

// **Die Schleife mit Puffer** — Werkstück G5 von Phase 5.5
// (#1462).
//
// # Wogegen hier geprüft wird
//
// **Dass sie sich wie ein Transport verhält und nicht wie ein Treffpunkt.**
// `net.Pipe` lässt einen Schreibvorgang erst zurückkehren, wenn jemand liest —
// und das hat kein Socket dieser Welt. Genau daran ist der Lauf über beide
// Enden einer Partie stehengeblieben: Gast und Wirt hatten zugleich etwas zu
// sagen, beide schrieben, keiner las.
//
// **Und dass sie trotzdem alles kann, was `internal/leitung` von ihr
// verlangt**: Fristen, die ein laufendes Warten abbrechen, und ein Ende, das
// beide Seiten schließt.

import (
	"errors"
	"io"
	"net"
	"testing"
)

// **Geschrieben wird, auch wenn niemand liest.**
func TestDieSchleifePuffertOhneLeser(t *testing.T) {
	t.Parallel()

	hier, dort := gepuffertesPaar()

	t.Cleanup(func() {
		_ = hier.Close() //nolint:errcheck // nur ein Lauf
	})

	for range 8 {
		if _, fehler := hier.Write([]byte("hallo")); fehler != nil {
			t.Fatalf("schreiben: %v", fehler)
		}
	}

	// **Und gelesen wird es danach, Stück für Stück.**
	ziel := make([]byte, 5)

	for range 8 {
		gelesen, fehler := dort.Read(ziel)
		if fehler != nil {
			t.Fatalf("lesen: %v", fehler)
		}

		if string(ziel[:gelesen]) != "hallo" {
			t.Fatalf("gelesen wurde %q", ziel[:gelesen])
		}
	}
}

// **Ein Leser, der weniger holt, verliert nichts.**
//
// **Ein Strom hat keine Grenzen** — wer den Überhang wegwürfe, verlöre Bytes
// mitten in einem Rahmen.
func TestDieSchleifeVerliertKeinenUeberhang(t *testing.T) {
	t.Parallel()

	hier, dort := gepuffertesPaar()

	if _, fehler := hier.Write([]byte("abcdef")); fehler != nil {
		t.Fatalf("schreiben: %v", fehler)
	}

	erste := make([]byte, 2)

	if _, fehler := dort.Read(erste); fehler != nil {
		t.Fatalf("lesen: %v", fehler)
	}

	rest := make([]byte, 4)

	if _, fehler := io.ReadFull(dort, rest); fehler != nil {
		t.Fatalf("den Rest lesen: %v", fehler)
	}

	if string(erste)+string(rest) != "abcdef" {
		t.Errorf("zusammen kam %q%q", erste, rest)
	}
}

// **Wer eines niederlegt, legt beide nieder.**
func TestEinEndeSchliesstBeide(t *testing.T) {
	t.Parallel()

	for name, schliesse := range map[string]int{"hiesige": 0, "dortige": 1} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			hier, dort := gepuffertesPaar()
			enden := []net.Conn{hier, dort}

			if fehler := enden[schliesse].Close(); fehler != nil {
				t.Fatalf("schließen: %v", fehler)
			}

			// **Zweimal schließen ist kein Fehler** — und darf den geteilten
			// Kanal nicht zweimal schließen.
			if fehler := enden[1-schliesse].Close(); fehler != nil {
				t.Fatalf("das zweite Schließen klagte: %v", fehler)
			}

			if _, fehler := hier.Write([]byte("hallo")); !errors.Is(
				fehler, net.ErrClosed) {
				t.Errorf("schreiben kam mit %v", fehler)
			}

			if _, fehler := dort.Read(make([]byte, 4)); !errors.Is(
				fehler, net.ErrClosed) {
				t.Errorf("lesen kam mit %v", fehler)
			}
		})
	}
}

// **Was vor dem Schließen unterwegs war, kommt an — danach das Ende** —
// #2504.
// Ein Socket liefert Gesendetes, bevor er meldet, dass drüben niemand mehr
// ist. Die Schleife verwarf es: ein Handschlag, dessen letzter Schritt kurz
// vor dem Schließen hinausging, kam unter Last nicht an.
func TestWasVorDemSchliessenUnterwegsWarKommtAn(t *testing.T) {
	t.Parallel()

	hier, dort := gepuffertesPaar()

	if _, fehler := hier.Write([]byte("abschluss")); fehler != nil {
		t.Fatalf("schreiben: %v", fehler)
	}

	if fehler := hier.Close(); fehler != nil {
		t.Fatalf("schließen: %v", fehler)
	}

	gelesen := make([]byte, 16)

	anzahl, fehler := dort.Read(gelesen)
	if fehler != nil || string(gelesen[:anzahl]) != "abschluss" {
		t.Fatalf("gelesen %q mit %v", gelesen[:anzahl], fehler)
	}

	if _, fehler := dort.Read(gelesen); !errors.Is(fehler, io.EOF) {
		t.Errorf("nach dem Rest kam %v statt des Endes", fehler)
	}

	if _, fehler := dort.Write([]byte("x")); !errors.Is(fehler, net.ErrClosed) {
		t.Errorf("schreiben nach drüben kam mit %v", fehler)
	}
}

// **Wer sein eigenes Ende niederlegt, liest nichts mehr** — auch nicht, was
// noch im Puffer liegt. So hält es `net.Conn`.
func TestWerSeinEndeNiederlegtLiestNichtsMehr(t *testing.T) {
	t.Parallel()

	hier, dort := gepuffertesPaar()

	if _, fehler := dort.Write([]byte("spät")); fehler != nil {
		t.Fatalf("schreiben: %v", fehler)
	}

	if fehler := hier.Close(); fehler != nil {
		t.Fatalf("schließen: %v", fehler)
	}

	if _, fehler := hier.Read(make([]byte, 8)); !errors.Is(fehler, net.ErrClosed) {
		t.Errorf("lesen kam mit %v", fehler)
	}
}
