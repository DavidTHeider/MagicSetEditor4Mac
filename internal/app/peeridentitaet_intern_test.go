package app

// **Die Peeridentität an der Naht** — Werkstück N1 von Phase 5.4
// (#1454).
//
// # Was hier gemessen wird
//
// **Die Übersetzung und der Weg dorthin** — nicht die Regel: welcher der vier
// Stände gilt, entscheidet [service.Peerwerk], und dort steht der Lauf dazu.
// Hier steht, dass die Naht den Bestand und den Bund wirklich zusammenführt
// und dass die Antwort trägt, was eine Fläche braucht.
//
// **Mit einem eingesetzten Bund** (§3): ohne ihn schriebe dieser Lauf eine
// Identität in den Anmeldebund dieses Rechners.

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// peerprobebund hält ein Geheimnis im Speicher.
type peerprobebund struct {
	geheimnis string
	fehler    error
}

func (p *peerprobebund) Lies(context.Context, string) (string, error) {
	if p.fehler != nil {
		return "", p.fehler
	}

	if p.geheimnis == "" {
		return "", domain.MitGrund(domain.ErrNotFound, nil,
			"in diesem Bund liegt nichts")
	}

	return p.geheimnis, nil
}

func (p *peerprobebund) Schreibe(_ context.Context, _, geheimnis string) error {
	if p.fehler != nil {
		return p.fehler
	}

	p.geheimnis = geheimnis

	return nil
}

// einePeerwelt baut eine Anwendung mit Bestand und eingesetztem Bund.
func einePeerwelt(t *testing.T) (*App, *peerprobebund) {
	t.Helper()

	bund := &peerprobebund{}
	anwendung := PeerflaecheUeber(bestandFuerDenRiegel(t)).MitPeerbund(bund)

	return anwendung, bund
}

// **Ohne Identität sagt die Naht es — und legt keine an.**
func TestOhneIdentitaetSagtDieNahtEs(t *testing.T) {
	t.Parallel()

	anwendung, bund := einePeerwelt(t)

	antwort := anwendung.Peeridentitaet(riegelkontext(t))
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if antwort.Stand != string(domain.PeerstandOhne) {
		t.Fatalf("der Stand ist %q", antwort.Stand)
	}

	// **Eine leere Liste und kein `null`** — sonst verschwände sie aus der
	// Antwort, und die Seite bliebe ohne Grund leer (§3).
	if antwort.Befunde == nil {
		t.Error("die Befunde sind null statt leer")
	}

	if bund.geheimnis != "" {
		t.Error("das Fragen hat eine Identität angelegt")
	}
}

// **Angelegt steht sie danach an derselben Naht.**
func TestEineAngelegteIdentitaetStehtDanachDa(t *testing.T) {
	t.Parallel()

	anwendung, bund := einePeerwelt(t)

	angelegt := anwendung.PeeridentitaetSicherstellen(riegelkontext(t))
	if angelegt.Fehler != nil {
		t.Fatalf("%+v", angelegt.Fehler)
	}

	if angelegt.Kennung == "" || angelegt.Fingerabdruck == "" ||
		angelegt.Seit == "" {
		t.Fatalf("die Antwort ist %+v", angelegt)
	}

	if bund.geheimnis == "" {
		t.Fatal("im Bund liegt nichts")
	}

	gelesen := anwendung.Peeridentitaet(riegelkontext(t))
	if gelesen.Fehler != nil {
		t.Fatalf("%+v", gelesen.Fehler)
	}

	if gelesen.Kennung != angelegt.Kennung {
		t.Fatalf("gelesen wurde %q, angelegt %q",
			gelesen.Kennung, angelegt.Kennung)
	}

	// **Und der Stand ist bereit** — der Vermerk und der Bund sagen dasselbe.
	if gelesen.Stand != string(domain.PeerstandBereit) {
		t.Errorf("der Stand ist %q", gelesen.Stand)
	}
}

// **Eine Rotation nennt die abgelöste Kennung samt ihrem Abdruck.**
//
// **Beides, und das ist der Punkt**: ein Kontakt hat den **Abdruck**
// bestätigt, nicht die Kennung — ohne ihn könnte ihm niemand sagen, was nicht
// mehr gilt.
func TestEineRotationNenntDieAbgeloesteKennung(t *testing.T) {
	t.Parallel()

	anwendung, _ := einePeerwelt(t)

	erste := anwendung.PeeridentitaetSicherstellen(riegelkontext(t))
	if erste.Fehler != nil {
		t.Fatalf("%+v", erste.Fehler)
	}

	zweite := anwendung.PeeridentitaetRotieren(riegelkontext(t))
	if zweite.Fehler != nil {
		t.Fatalf("%+v", zweite.Fehler)
	}

	if zweite.Kennung == erste.Kennung {
		t.Fatal("nach der Rotation steht dieselbe Kennung da")
	}

	if zweite.Vorher != erste.Kennung ||
		zweite.VorherFingerabdruck != erste.Fingerabdruck {
		t.Fatalf("die Vorgängerin ist %q/%q, erwartet %q/%q",
			zweite.Vorher, zweite.VorherFingerabdruck,
			erste.Kennung, erste.Fingerabdruck)
	}

	if len(zweite.Befunde) != 1 {
		t.Errorf("die Befunde sind %v", zweite.Befunde)
	}
}

// **Jeder der drei Wege reicht seinen Fehler durch.**
func TestJederPeerwegReichtSeinenFehlerDurch(t *testing.T) {
	t.Parallel()

	wege := map[string]func(*App) PeeridentitaetAntwort{
		"lesen": func(a *App) PeeridentitaetAntwort {
			return a.Peeridentitaet(riegelkontext(t))
		},
		"anlegen": func(a *App) PeeridentitaetAntwort {
			return a.PeeridentitaetSicherstellen(riegelkontext(t))
		},
		"rotieren": func(a *App) PeeridentitaetAntwort {
			return a.PeeridentitaetRotieren(riegelkontext(t))
		},
	}

	for name, weg := range wege {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			anwendung, bund := einePeerwelt(t)
			bund.fehler = errors.New("der Schlüsselbund antwortet nicht")

			if antwort := weg(anwendung); antwort.Fehler == nil {
				t.Fatal("der Fehler kam nicht durch")
			}
		})
	}
}

// **Ohne eingesetzten Bund nimmt die Naht den echten.**
//
// **Und genau deshalb steht hier kein Ruf**: dieser Lauf prüft die **Wahl**
// und nicht den Zugriff — ein Ruf schriebe in den Anmeldebund dieses
// Rechners, und `internal/schluesselbund` sagt in seinem Kopf, warum das
// nicht sein darf.
func TestOhneEingesetztenBundNimmtDieNahtDenEchten(t *testing.T) {
	t.Parallel()

	if PeerflaecheUeber(nil).bundFuerDenPeer() == nil {
		t.Fatal("ohne eingesetzten Bund kam keiner")
	}
}
