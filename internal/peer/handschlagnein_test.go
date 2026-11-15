package peer_test

// **Die übrigen Neins des Handschlags** — Werkstück N1 von Phase 5.4
// (#1454).
//
// # Warum sie einen eigenen Lauf bekommen
//
// **Weil ein Nein eine Antwort ist** (§10). Wer nur den Erfolgsweg prüft, hat
// die Hälfte ausgelassen — und die Hälfte, die hier steht, ist die, auf die
// sich alles andere verlässt: dass ein halb ausgefüllter, verschobener oder
// abgeschnittener Schritt nicht weitergeht.

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/peer"
)

// kaputterZufall gibt nie etwas her.
type kaputterZufall struct{}

func (kaputterZufall) Read([]byte) (int, error) {
	return 0, errors.New("die Quelle ist zu")
}

// **Ein Angebot, dem etwas fehlt, entsteht gar nicht erst.**
func TestEinAngebotDemEtwasFehltEntstehtNicht(t *testing.T) {
	t.Parallel()

	ich := eineIdentitaet(t)

	faelle := map[string]peer.Wunsch{
		"ohne Zweck":     {Rolle: "gebend", Kanal: "k"},
		"fremder Zweck":  {Zweck: "alles", Rolle: "gebend", Kanal: "k"},
		"ohne Rolle":     {Zweck: domain.ZweckTeilen, Kanal: "k"},
		"ohne Kanal":     {Zweck: domain.ZweckTeilen, Rolle: "gebend"},
		"ohne beides":    {Zweck: domain.ZweckTeilen},
		"leere Bindung ": {Zweck: domain.ZweckPartie, Rolle: "spielend"},
	}

	for name, wunsch := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if _, _, fehler := ich.Anbieten(
				wunsch, &festerZufall{},
			); !errors.Is(fehler, domain.ErrValidation) {
				t.Fatalf("es kam durch: %v", fehler)
			}
		})
	}
}

// **Ohne Zufall gibt es keinen Handschlag — und keine halbe Sitzung.**
func TestOhneZufallGibtEsKeinenHandschlag(t *testing.T) {
	t.Parallel()

	ich, du := zweiIdentitaeten(t)

	if _, _, fehler := ich.Anbieten(peer.Wunsch{
		Zweck: domain.ZweckTeilen, Rolle: "gebend", Kanal: "kanal-1",
	}, kaputterZufall{}); !errors.Is(fehler, domain.ErrConflict) {
		t.Fatalf("das Angebot entstand ohne Zufall: %v", fehler)
	}

	angebot := einRohesAngebot(
		t, ich, peer.ProtokollTeilen, peer.ArtAngebot, dasAngebot(nil))

	if _, _, fehler := du.Entgegennehmen(
		angebot, peer.Bereitschaft{Rolle: "nehmend", Kanal: "kanal-1"},
		kaputterZufall{}, eineListe(t),
	); !errors.Is(fehler, domain.ErrConflict) {
		t.Fatalf("die Annahme entstand ohne Zufall: %v", fehler)
	}
}

// **Wer selbst nichts beiträgt, nimmt nichts entgegen.**
func TestWerSelbstNichtsBeitraegtNimmtNichtsEntgegen(t *testing.T) {
	t.Parallel()

	ich, du := zweiIdentitaeten(t)

	angebot := einRohesAngebot(
		t, ich, peer.ProtokollTeilen, peer.ArtAngebot, dasAngebot(nil))

	for name, bereit := range map[string]peer.Bereitschaft{
		"ohne eigene Rolle":  {Kanal: "kanal-1"},
		"ohne eigenen Kanal": {Rolle: "nehmend"},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if fehler := nimmt(t, du, angebot, bereit); !errors.Is(
				fehler, domain.ErrValidation) {
				t.Fatalf("es kam durch: %v", fehler)
			}
		})
	}

	// **Und wer sie nicht nennt, bietet auch nichts an.**
	angebotOhneRolle := einRohesAngebot(
		t, ich, peer.ProtokollTeilen, peer.ArtAngebot,
		dasAngebot(func(a *peer.Angebot) { a.Rolle = "" }))

	if fehler := nimmt(t, du, angebotOhneRolle, peer.Bereitschaft{
		Rolle: "nehmend", Kanal: "kanal-1",
	}); !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("die leere Rolle kam durch: %v", fehler)
	}
}

// **Kopf und Rumpf nennen dieselbe Sitzung.**
//
// **Sonst gäbe es zwei in einer Nachricht** — und welche davon im Gedächtnis
// landet, entschiede die Reihenfolge im Quelltext.
func TestKopfUndRumpfNennenDieselbeSitzung(t *testing.T) {
	t.Parallel()

	ich, du := zweiIdentitaeten(t)

	// **Beide für sich gültig**: 32 Byte Hex, nur eben nicht dieselben.
	angebot := einRohesAngebot(
		t, ich, peer.ProtokollTeilen, peer.ArtAngebot,
		dasAngebot(func(a *peer.Angebot) {
			a.Nonce = strings.Repeat("b", 64)
		}))

	if fehler := nimmt(t, du, angebot, peer.Bereitschaft{
		Rolle: "nehmend", Kanal: "kanal-1",
	}); !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("die zwei Sitzungen kamen durch: %v", fehler)
	}
}

// **Was keine Annahme ist, schließt nichts.**
func TestWasKeineAnnahmeIstSchliesstNichts(t *testing.T) {
	t.Parallel()

	faelle := map[string]func(*testing.T, peer.Identitaet, string) []byte{
		"kein JSON": func(*testing.T, peer.Identitaet, string) []byte {
			return []byte("hallo")
		},
		"andere Art": func(t *testing.T, du peer.Identitaet, _ string) []byte {
			return einRohesAngebot(
				t, du, peer.ProtokollTeilen, peer.ArtAngebot, dasAngebot(nil))
		},
		"fremder Rumpf": func(
			t *testing.T, du peer.Identitaet, nonce string,
		) []byte {
			return dieHuelle(t, du, peer.ArtAnnahme, nonce, []int{1, 2})
		},
		"andere Sitzung im Kopf": func(
			t *testing.T, du peer.Identitaet, nonce string,
		) []byte {
			return dieHuelle(t, du, peer.ArtAnnahme, strings.Repeat("c", 64),
				peer.Annahme{
					Fassung:   peer.Handschlagfassung,
					Zweck:     domain.ZweckTeilen,
					Rolle:     "nehmend",
					Kanal:     "kanal-1",
					DeinNonce: nonce,
					MeinNonce: strings.Repeat("d", 64),
				})
		},
	}

	for name, baue := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			anbahnung, du, echt := einLauf(t)
			nonce := dieAnnahmeAus(t, echt).DeinNonce

			if _, _, fehler := anbahnung.Schliessen(
				baue(t, du, nonce),
			); !errors.Is(fehler, domain.ErrValidation) {
				t.Fatalf("es kam durch: %v", fehler)
			}
		})
	}
}

// **Was kein Abschluss ist, schließt nichts.**
func TestWasKeinAbschlussIstSchliesstNichts(t *testing.T) {
	t.Parallel()

	faelle := map[string]func(*testing.T, peer.Identitaet) []byte{
		"kein JSON": func(*testing.T, peer.Identitaet) []byte {
			return []byte("hallo")
		},
		"andere Art": func(t *testing.T, ich peer.Identitaet) []byte {
			return einRohesAngebot(
				t, ich, peer.ProtokollTeilen, peer.ArtAnnahme, dasAngebot(nil))
		},
		"fremder Rumpf": func(t *testing.T, ich peer.Identitaet) []byte {
			return einRohesAngebot(
				t, ich, peer.ProtokollTeilen, peer.ArtAbschluss, []int{1, 2})
		},
		"andere Fassung": func(t *testing.T, ich peer.Identitaet) []byte {
			return einRohesAngebot(
				t, ich, peer.ProtokollTeilen, peer.ArtAbschluss,
				peer.Abschluss{Fassung: peer.Handschlagfassung + 1})
		},
	}

	for name, baue := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ich, du := zweiIdentitaeten(t)

			_, angebot, fehler := ich.Anbieten(peer.Wunsch{
				Zweck: domain.ZweckTeilen, Rolle: "gebend", Kanal: "kanal-1",
			}, &festerZufall{stand: 7})
			if fehler != nil {
				t.Fatalf("Anbieten: %v", fehler)
			}

			entgegennahme, _, fehler := du.Entgegennehmen(
				angebot, peer.Bereitschaft{Rolle: "nehmend", Kanal: "kanal-1"},
				&festerZufall{stand: 200}, eineListe(t))
			if fehler != nil {
				t.Fatalf("Entgegennehmen: %v", fehler)
			}

			if _, fehler := entgegennahme.Abschliessen(
				baue(t, ich),
			); !errors.Is(fehler, domain.ErrValidation) {
				t.Fatalf("es kam durch: %v", fehler)
			}
		})
	}
}

// **Das Gedächtnis vergisst, wenn seine Frist um ist.**
//
// # Was das kostet, und warum es trotzdem richtig ist
//
// **Nach der Frist käme dieselbe Zufallszahl wieder durch.** Getragen wird
// dann von der Kanalbindung: ein Angebot gilt auf der Verbindung, auf der es
// entstand. Ohne Vergessen wüchse die Liste unbegrenzt, und ein Peer, dem
// noch niemand traut, füllte den Speicher.
func TestDasGedaechtnisVergisstNachSeinerFrist(t *testing.T) {
	t.Parallel()

	jetzt := time.Date(2026, 9, 14, 12, 0, 0, 0, time.UTC)
	liste := peer.NeueNonceliste(func() time.Time { return jetzt }, time.Hour)

	if fehler := liste.Merke("abc"); fehler != nil {
		t.Fatalf("der erste Eintrag scheiterte: %v", fehler)
	}

	if fehler := liste.Merke("abc"); !errors.Is(
		fehler, domain.ErrValidation) {
		t.Fatalf("der zweite Eintrag kam durch: %v", fehler)
	}

	jetzt = jetzt.Add(2 * time.Hour)

	if fehler := liste.Merke("abc"); fehler != nil {
		t.Fatalf("nach der Frist scheiterte er: %v", fehler)
	}
}

// dieHuelle verpackt einen beliebigen Rumpf unter einer Art und Sitzung.
func dieHuelle(
	t *testing.T, wer peer.Identitaet, art, sitzung string, rumpf any,
) []byte {
	t.Helper()

	roh, fehler := json.Marshal(rumpf)
	if fehler != nil {
		t.Fatalf("Marshal: %v", fehler)
	}

	nachricht, fehler := wer.Verpacke(peer.Huelle{
		Protokoll: peer.ProtokollTeilen,
		Art:       art,
		ID:        "n-x",
		Sitzung:   sitzung,
		Absender:  string(wer.Kennung()),
		Folge:     1,
		Rumpf:     roh,
	})
	if fehler != nil {
		t.Fatalf("Verpacke: %v", fehler)
	}

	return nachricht
}

// **Ein Name, der die Nachricht sprengt, geht nicht hinaus.**
//
// # Warum ein Name das kann
//
// **Weil er im Rumpf steht und nicht im Kopf.** Die Kopffelder haben ihre
// Grenze ([peer.Hoechstwort]); was ein Mensch sich selbst nennt, hat sie
// nicht — und eine Zeichenkette von einem Megabyte ist immer noch ein Name.
// Die Grenze der **Nachricht** fängt sie, und zwar beim Absender.
func TestEinNameDerDieNachrichtSprengtGehtNichtHinaus(t *testing.T) {
	t.Parallel()

	ich, du := zweiIdentitaeten(t)
	zuLang := strings.Repeat("n", peer.Hoechstlaenge)

	if _, _, fehler := ich.Anbieten(peer.Wunsch{
		Zweck: domain.ZweckTeilen, Rolle: "gebend", Kanal: "kanal-1",
		Name: zuLang,
	}, &festerZufall{}); !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("das Angebot kam durch: %v", fehler)
	}

	// **Und auf der anderen Seite ebenso.**
	angebot := einRohesAngebot(
		t, ich, peer.ProtokollTeilen, peer.ArtAngebot, dasAngebot(nil))

	if fehler := nimmt(t, du, angebot, peer.Bereitschaft{
		Rolle: "nehmend", Kanal: "kanal-1", Name: zuLang,
	}); !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("die Annahme kam durch: %v", fehler)
	}
}

// **Ohne eingesetzte Quelle nimmt der Handschlag die des Betriebssystems.**
//
// # Der Fund
//
// **`nil` hieß hier Absturz und nicht „nimm die echte".** [peer.Erzeugen]
// behandelt es seit N1.1 richtig; `eineZufallszahl` tat es nicht, und
// `io.ReadFull(nil, …)` stürzt. Aufgefallen ist es, als `service` den
// Handschlag zum ersten Mal wirklich führte — der natürliche Ruf ist
// `Anbieten(wunsch, nil)`, und genau der war der kaputte.
//
// **Zwei Wege, und beide würfeln**: die anbietende Seite und die annehmende.
func TestOhneQuelleNimmtDerHandschlagDieDesBetriebssystems(t *testing.T) {
	t.Parallel()

	ich, du := zweiIdentitaeten(t)

	anbahnung, angebot, fehler := ich.Anbieten(peer.Wunsch{
		Zweck: domain.ZweckTeilen, Rolle: "gebend", Kanal: "kanal-1",
	}, nil)
	if fehler != nil {
		t.Fatalf("Anbieten ohne Quelle: %v", fehler)
	}

	entgegennahme, annahme, fehler := du.Entgegennehmen(angebot,
		peer.Bereitschaft{Rolle: "nehmend", Kanal: "kanal-1"},
		nil, eineListe(t))
	if fehler != nil {
		t.Fatalf("Entgegennehmen ohne Quelle: %v", fehler)
	}

	meine, abschluss, fehler := anbahnung.Schliessen(annahme)
	if fehler != nil {
		t.Fatalf("Schliessen: %v", fehler)
	}

	deine, fehler := entgegennahme.Abschliessen(abschluss)
	if fehler != nil {
		t.Fatalf("Abschliessen: %v", fehler)
	}

	if meine.Kennung == "" || meine.Kennung != deine.Kennung {
		t.Fatalf("zwei Sitzungen: %q und %q", meine.Kennung, deine.Kennung)
	}
}
