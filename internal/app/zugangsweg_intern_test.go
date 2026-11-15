package app

// **Was eingetragen wird, muss ankommen** —
// #1537.
//
// # Der Befund
//
// **Ein Schlüssel für OpenAI, Anthropic oder xAI erreichte den Klienten nie.**
// `ZugangSetzen` schrieb nach `MagicSetEditor4Mac / openai.schluessel`,
// `cloudschluessel` las aus `openai / schluessel` — zwei Orte, die sich nicht
// treffen. Der Klient bekam eine leere Zeichenkette, und am Fenster sah alles
// in Ordnung aus, weil `Zugaenge()` die **richtige** Gestalt liest.
//
// # Warum kein Lauf es sah
//
// **Weil jeder beide Hälften selbst setzt.** Die Nahtläufe legen ihr
// Geheimnis mit der Gestalt ab, die die Naht danach liest; die Bindungsläufe
// prüfen, dass geschrieben wird. **Keiner fuhr den Weg von der Eingabe bis
// zum Klienten** — die Naht zwischen zwei Prüfebenen, vor der §3 warnt.
//
// # Was hier gehalten wird
//
// **Die Klasse und nicht die drei Fälle** (§10): für **jeden** Dienst mit
// Zugangsangaben wird eingetragen und dann nachgesehen, ob die Naht dieses
// Dienstes an **derselben** Stelle sucht. Ein vierter Dienst mit einer neuen
// Gestalt wird damit rot, statt still ins Leere zu lesen.

import (
	"context"
	"slices"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// zugangsmitschrift hält fest, wo geschrieben und wo gesucht wurde.
type zugangsmitschrift struct {
	inhalt map[string]string

	geschrieben []string
	gesucht     []string
}

func neueZugangsmitschrift() *zugangsmitschrift {
	return &zugangsmitschrift{inhalt: map[string]string{}}
}

// ort ist die Stelle im Bund, als eine Zeichenkette.
func ort(dienst, konto string) string { return dienst + " / " + konto }

func (z *zugangsmitschrift) Schreibe(
	_ context.Context, dienst, konto, wert string,
) error {
	z.geschrieben = append(z.geschrieben, ort(dienst, konto))
	z.inhalt[ort(dienst, konto)] = wert

	return nil
}

func (z *zugangsmitschrift) Lies(_ context.Context, dienst, konto string) (
	string, error,
) {
	z.gesucht = append(z.gesucht, ort(dienst, konto))

	if wert, da := z.inhalt[ort(dienst, konto)]; da {
		return wert, nil
	}

	return "", domain.ErrNotFound
}

func (z *zugangsmitschrift) Loesche(context.Context, string, string) error {
	return nil
}

// **Was für einen Cloudanbieter eingetragen wird, findet sein Klient.**
func TestEinEingetragenerCloudschluesselErreichtDenKlienten(t *testing.T) {
	for _, dienst := range []domain.Dienst{
		domain.DienstAnthropic, domain.DienstOpenAI, domain.DienstXAI,
	} {
		t.Run(string(dienst), func(t *testing.T) {
			buch := neueZugangsmitschrift()

			anwendung := NeueApp("")
			anwendung.bund = buch
			anwendung.marktbund = buch

			erwartet := "geheim-" + string(dienst)

			if antwort := anwendung.ZugangSetzen(t.Context(), string(dienst),
				domain.ZugangSchluessel, erwartet); antwort.Fehler != nil {
				t.Fatalf("ZugangSetzen: %+v", antwort.Fehler)
			}

			if gefunden := anwendung.cloudschluesselFuer(dienst, ""); gefunden != erwartet {
				t.Fatalf("der Klient bekam %q statt %q — geschrieben nach "+
					"%v, gesucht in %v", gefunden, erwartet,
					buch.geschrieben, buch.gesucht)
			}
		})
	}
}

// **Und jede Naht sucht dort, wo das Fenster ablegt** — für jeden Dienst.
//
// **Die Klasse**: der Lauf darüber nennt drei Dienste beim Namen, dieser
// keinen. Wer einen vierten aufnimmt und ihn woanders lesen lässt, wird hier
// rot.
func TestJedeZugangsnahtSuchtWoDasFensterAblegt(t *testing.T) {
	for _, dienst := range domain.Dienste() {
		if len(dienst.Zugangsangaben()) == 0 {
			continue
		}

		t.Run(string(dienst), func(t *testing.T) {
			buch := neueZugangsmitschrift()

			anwendung := NeueApp("")
			anwendung.bund = buch
			anwendung.marktbund = buch

			// **Alle Angaben eines Dienstes und nicht eine.**
			// `zugangslageVon` hält bei der **ersten** fehlenden an — das ist
			// Absicht, und ein Lauf, der nur eine ablegte, käme nie bis zur
			// zweiten.
			for _, angabe := range dienst.Zugangsangaben() {
				if antwort := anwendung.ZugangSetzen(t.Context(),
					string(dienst), angabe, "geheim"); antwort.Fehler != nil {
					t.Fatalf("ZugangSetzen: %+v", antwort.Fehler)
				}
			}

			// **Der Katalog liest jeden Zugang** und ist damit die eine
			// Naht, die alle Dienste anfasst.
			anwendung.Dienstkatalog(t.Context())

			for _, abgelegt := range buch.geschrieben {
				if !slices.Contains(buch.gesucht, abgelegt) {
					t.Errorf("abgelegt in %q, gesucht wurde in %v",
						abgelegt, buch.gesucht)
				}
			}
		})
	}
}
