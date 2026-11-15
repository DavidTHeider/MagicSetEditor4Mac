package domain_test

// **Der Kopf einer Handelsnachricht** — Werkstück `T5` von Phase 5.6
// (#1470).
//
// # Was hier gemessen wird
//
// **Dass nur die zwölf erlaubten Arten durchkommen.** Eine Fläche, die jede
// Zeichenkette als Art annimmt, ist ein frei wählbarer Aufrufkanal — genau
// das, was dieses Werkstück ausschließt.
//
// **Dass der Idempotenzschlüssel an Vorgang, Operation und Revision hängt.**
// Zwei Nachrichten, die sich in einem dieser drei unterscheiden, sind zwei;
// zweimal dieselben drei sind eine. Das Profil steht nicht darin — jede
// Wurzelabfrage klemmt ohnehin darauf (0066).
//
// **Und dass der Zeitstempel nicht unterschrieben wird.** Er ist ein
// Metadatum: eine Uhr, die vorgeht, ist kein Angriff, und eine Uhr, die
// nachgeht, darf keine Nachricht ungültig machen.

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// einUmschlag baut einen gültigen Kopf.
func einUmschlag(anders func(*domain.Handelsumschlag)) domain.Handelsumschlag {
	umschlag := domain.Handelsumschlag{
		Fassung:       domain.Handelsprotokollfassung,
		Art:           domain.NachrichtVorschlag,
		NachrichtID:   "n-1",
		Vorgang:       "v-1",
		Basisrevision: 2,
		Abdruck:       "abc123",
		Folge:         7,
		Absender:      "bgegenueber",
		Unterschrift:  "sig",
		Gesendet:      time.Date(2026, 9, 18, 12, 0, 0, 0, time.UTC),
	}

	if anders != nil {
		anders(&umschlag)
	}

	return umschlag
}

func TestEinGueltigerUmschlagKommtDurch(t *testing.T) {
	t.Parallel()

	if befunde := einUmschlag(nil).Validate(); len(befunde) != 0 {
		t.Errorf("es kamen Befunde: %v", befunde)
	}

	// **Jede erlaubte Art kommt durch** — die Gegenprobe zur Aufzählung
	// darunter (§10): ein Nein zu allem wäre kein Nein zur falschen Art.
	for _, art := range domain.Handelsnachrichtenarten() {
		umschlag := einUmschlag(func(u *domain.Handelsumschlag) { u.Art = art })

		if befunde := umschlag.Validate(); len(befunde) != 0 {
			t.Errorf("%q gab %v", art, befunde)
		}
	}
}

func TestWasEinenUmschlagUngueltigMacht(t *testing.T) {
	t.Parallel()

	for feld, anders := range map[string]func(*domain.Handelsumschlag){
		"fassung": func(u *domain.Handelsumschlag) {
			u.Fassung = domain.Handelsprotokollfassung + 1
		},
		"art":         func(u *domain.Handelsumschlag) { u.Art = "loescheAlles" },
		"nachrichtID": func(u *domain.Handelsumschlag) { u.NachrichtID = "  " },
		"vorgang":     func(u *domain.Handelsumschlag) { u.Vorgang = "" },
		"basisrevision": func(u *domain.Handelsumschlag) {
			u.Basisrevision = -1
		},
		"folge":        func(u *domain.Handelsumschlag) { u.Folge = 0 },
		"absender":     func(u *domain.Handelsumschlag) { u.Absender = "" },
		"unterschrift": func(u *domain.Handelsumschlag) { u.Unterschrift = "" },
	} {
		t.Run(feld, func(t *testing.T) {
			t.Parallel()

			befunde := einUmschlag(anders).Validate()
			if len(befunde) != 1 {
				t.Fatalf("es kamen %d Befunde: %v", len(befunde), befunde)
			}

			if befunde[0].Field != feld {
				t.Errorf("der Befund nennt %q", befunde[0].Field)
			}
		})
	}
}

// **Eine unbekannte Fassung ist ein Nein und kein Versuch.** Wer sie
// durchließe, deutete Felder, deren Bedeutung er nicht kennt.
func TestEineAeltereFassungIstEbenfallsEinNein(t *testing.T) {
	t.Parallel()

	alt := einUmschlag(func(u *domain.Handelsumschlag) { u.Fassung = 0 })

	if befunde := alt.Validate(); len(befunde) != 1 || befunde[0].Field != "fassung" {
		t.Errorf("es kam %v", befunde)
	}
}

func TestDerSchluesselHaengtAnDreiDingen(t *testing.T) {
	t.Parallel()

	grund := einUmschlag(nil).Idempotenzschluessel()

	if zweimal := einUmschlag(nil).Idempotenzschluessel(); zweimal != grund {
		t.Errorf("zweimal dieselbe Nachricht gab %q und %q", grund, zweimal)
	}

	for name, aendere := range map[string]func(*domain.Handelsumschlag){
		"ein anderer Vorgang": func(u *domain.Handelsumschlag) { u.Vorgang = "v-2" },
		"eine andere Operation": func(u *domain.Handelsumschlag) {
			u.Art = domain.NachrichtAnnahme
		},
		"eine andere Revision": func(u *domain.Handelsumschlag) {
			u.Basisrevision = 3
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if anderer := einUmschlag(aendere).
				Idempotenzschluessel(); anderer == grund {
				t.Errorf("der Schlüssel blieb %q", anderer)
			}
		})
	}

	// **Die Nachrichtenkennung gehört nicht dazu.** Zwei Zustellungen
	// derselben Nachricht tragen dieselbe; zwei **verschiedene** Nachrichten
	// derselben Operation gegen dieselbe Revision sind ein Konflikt und nicht
	// zwei Vorgänge.
	andereKennung := einUmschlag(func(u *domain.Handelsumschlag) {
		u.NachrichtID = "n-2"
	})

	if andereKennung.Idempotenzschluessel() != grund {
		t.Error("die Nachrichtenkennung ist in den Schlüssel geraten")
	}
}

func TestDieUnterschriftsformBindetJedesFeldAusserDerUhr(t *testing.T) {
	t.Parallel()

	grund := einUmschlag(nil).Unterschriftsform()

	for name, anders := range map[string]func(*domain.Handelsumschlag){
		"die Fassung": func(u *domain.Handelsumschlag) { u.Fassung = 99 },
		"die Art":     func(u *domain.Handelsumschlag) { u.Art = domain.NachrichtAbbruch },
		"die Kennung": func(u *domain.Handelsumschlag) { u.NachrichtID = "n-2" },
		"der Vorgang": func(u *domain.Handelsumschlag) { u.Vorgang = "v-2" },
		"die Revision": func(u *domain.Handelsumschlag) {
			u.Basisrevision = 3
		},
		"der Abdruck":  func(u *domain.Handelsumschlag) { u.Abdruck = "anders" },
		"die Folge":    func(u *domain.Handelsumschlag) { u.Folge = 8 },
		"der Absender": func(u *domain.Handelsumschlag) { u.Absender = "canderer" },
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if bytes.Equal(einUmschlag(anders).Unterschriftsform(), grund) {
				t.Error("die Form hat sich nicht geändert")
			}
		})
	}

	// **Die Uhr steht nicht drin** — sie ist ein Metadatum und kein
	// Replay-Schutz. Dagegen steht der Idempotenzschlüssel.
	spaeter := einUmschlag(func(u *domain.Handelsumschlag) {
		u.Gesendet = u.Gesendet.Add(time.Hour)
	})

	if !bytes.Equal(spaeter.Unterschriftsform(), grund) {
		t.Error("die Uhr ist in die Unterschrift geraten")
	}

	// **Und die Unterschrift selbst auch nicht** — sie entsteht über dieser
	// Form; wer sie hineinnähme, könnte sie nie prüfen.
	unterschrieben := einUmschlag(func(u *domain.Handelsumschlag) {
		u.Unterschrift = "eine andere"
	})

	if !bytes.Equal(unterschrieben.Unterschriftsform(), grund) {
		t.Error("die Unterschrift ist in ihre eigene Form geraten")
	}
}

// **Die Form trennt ihre Teile.** „ab" + „c" und „a" + „bc" dürfen nicht
// dieselbe Form ergeben — sonst ließe sich eine Unterschrift auf eine andere
// Nachricht übertragen.
func TestDieFormTrenntIhreTeile(t *testing.T) {
	t.Parallel()

	links := einUmschlag(func(u *domain.Handelsumschlag) {
		u.Vorgang, u.Abdruck = "ab", "c"
	})

	rechts := einUmschlag(func(u *domain.Handelsumschlag) {
		u.Vorgang, u.Abdruck = "a", "bc"
	})

	if bytes.Equal(links.Unterschriftsform(), rechts.Unterschriftsform()) {
		t.Error("zwei verschiedene Nachrichten haben dieselbe Form")
	}

	// **Und sie nennt sich selbst** — eine Form ohne Marke ließe sich als
	// Angebotsform ausgeben.
	if !strings.HasPrefix(string(links.Unterschriftsform()), "handelsumschlag") {
		t.Error("die Form nennt nicht, was sie ist")
	}
}
