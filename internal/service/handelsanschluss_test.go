package service_test

// **Woran eine ankommende Revision anschließt** —
// #1634.
//
// # Warum das eine eigene Datei ist
//
// **Weil `handelspost_test.go` an der Zeilengrenze steht** (§2) — und weil
// die Frage eine eigene ist: nicht „darf dieser Absender", sondern „passt
// diese Revision auf das, was hier steht".

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// eineUnwissendePost kennt keinen einzigen Vorgang.
//
// **Die Lage des Empfängers beim ersten Angebot** — und bis zum 2026-09-19
// die Lage, in der jede Zustellung endete.
func eineUnwissendePost() *service.Handelspost {
	return service.NeueHandelspost(&briefkastenprobe{},
		standprobe{klagt: domain.MitGrund(domain.ErrNotFound, nil,
			"den Handelsvorgang gibt es nicht")},
		kontaktnachschlagprobe{kontakt: einHandelnder()})
}

// **Die erste Revision eröffnet auch beim Empfänger** —
// #1634.
//
// # Der Fund, für den es diesen Lauf gibt
//
// **Ein Handel kam bei dem, dem er angeboten wurde, nie an.** Die Post schlug
// den heutigen Stand nach, bekam „gibt es nicht" und endete dort — und ein
// Vorgang, den der Empfänger noch nicht kennt, ist **jeder erste**. Bis zum
// 2026-09-19 fiel das niemandem auf, weil die Läufe von `T5` die Gegenseite
// im selben Prozess aufstellten und beide denselben Vorgang schon hatten.
//
// **Gefunden hat es der erste Lauf über einen echten Draht** zwischen zwei
// Prozessen mit getrennten Beständen.
//
// # Warum das niemanden hereinlässt
//
// **Weil vorher schon drei Dinge feststehen** ([Handelspost.durchgelassen]):
// wer spricht, ist die geprüfte Peeridentität; er darf handeln; und Umschlag
// und Angebot meinen dieselbe Sache. Was hier dazukommt, ist die Prüfung des
// Angebots selbst — und die **Revision eins**: eine höhere ohne Vorgeschichte
// ist ein Konflikt und kein Anfang.
// einErstesVomGegenueber ist eine erste Revision, wie sie ankommt: vom
// Initiator — dem Gegenüber — geschickt (TRADE-03, #1717).
func einErstesVomGegenueber() domain.Handelsangebot {
	erstes := derHeutigeStand()
	erstes.Initiator = dasGegenueber

	return erstes
}

func TestDieErsteRevisionEroeffnetAuchBeimEmpfaenger(t *testing.T) {
	t.Parallel()

	briefkasten := &briefkastenprobe{}

	post := service.NeueHandelspost(briefkasten,
		standprobe{klagt: domain.MitGrund(domain.ErrNotFound, nil,
			"den Handelsvorgang gibt es nicht")},
		kontaktnachschlagprobe{kontakt: einHandelnder()})

	antwort, fehler := post.Empfangen(context.Background(), dasGegenueber,
		einePost(einErstesVomGegenueber(), nil))
	if fehler != nil {
		t.Fatalf("Empfangen: %v", fehler)
	}

	if antwort.Art != domain.ZustellungNeu {
		t.Errorf("die Zustellung gilt als %q", antwort.Art)
	}

	if len(briefkasten.angebote) != 1 {
		t.Fatalf("geschrieben wurden %d Angebote", len(briefkasten.angebote))
	}

	if briefkasten.angebote[0].Revision != 1 {
		t.Errorf("geschrieben wurde Revision %d",
			briefkasten.angebote[0].Revision)
	}
}

// **Eine zweite Revision ohne Vorgeschichte ist ein Konflikt.**
//
// **Die Gegenprobe**: ohne sie ließe der Weg oben jede Revision als Anfang
// durch — und ein Vorgang entstünde bei Revision 7, dessen sechs Vorgänger
// niemand gesehen hat.
func TestEineSpaetereRevisionOhneVorgeschichteIstEinKonflikt(t *testing.T) {
	t.Parallel()

	_, fehler := eineUnwissendePost().Empfangen(
		context.Background(), dasGegenueber, einePost(einNachfolger(), nil))
	if !errors.Is(fehler, domain.ErrConflict) {
		t.Fatalf("es kam %v", fehler)
	}
}

// **Ein kaputtes erstes Angebot wird geprüft, bevor es entsteht.**
//
// # Die Zusage, die ohne diesen Lauf halb wäre
//
// **Ein unbekannter Vorgang kommt an keinem Stand vorbei** — es gibt keinen,
// den [domain.Handelsangebot.Nimmt] fragen könnte. Ohne diese Prüfung ginge
// jede erste Revision ungesehen durch, und der Empfänger schriebe einen
// Vorgang ohne Parteien, ohne Positionen, ohne Art.
func TestEinKaputtesErstesAngebotKommtNichtDurch(t *testing.T) {
	t.Parallel()

	kaputt := derHeutigeStand()
	kaputt.Positionen = nil

	_, fehler := eineUnwissendePost().Empfangen(
		context.Background(), dasGegenueber, einePost(kaputt, nil))
	if !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("es kam %v", fehler)
	}
}

// **Wer nicht mitspielt, schreibt auch nicht mit.**
//
// # Der Fund, für den es diesen Lauf gibt
//
// **Ein freigegebener Kontakt konnte Revisionen in einen fremden Handel
// schieben.** Geprüft war, dass er sprechen darf und dass er derselbe ist,
// den der Umschlag nennt — **nicht**, dass er an diesem Handel beteiligt
// ist. [domain.Handelsangebot.Nimmt] hält den Kopf fest, und dazu gehören
// die Parteien; ein Dritter konnte deshalb die **Positionen** eines Handels
// ändern, an dem er nicht teilnimmt.
//
// **Aufgefallen ist es beim Bau der ersten Revision** (#1634): sie entsteht
// aus nichts, und die Frage „wer darf einen Vorgang anlegen" führte auf die
// ältere Frage „wer darf ihn ändern".
func TestWerNichtMitspieltSchreibtAuchNichtMit(t *testing.T) {
	t.Parallel()

	briefkasten := &briefkastenprobe{}

	// **Ein Dritter, der handeln darf** — er ist bestätigt und freigegeben,
	// und trotzdem geht ihn dieser Vorgang nichts an.
	dritter := "bdritter"

	post := service.NeueHandelspost(briefkasten,
		standprobe{stand: derHeutigeStand()},
		kontaktnachschlagprobe{kontakt: domain.Peerkontakt{
			Kennung: dritter, Stand: domain.KontaktBestaetigt,
			Rechte: []domain.Zweck{domain.ZweckHandel},
		}})

	_, fehler := post.Empfangen(context.Background(), dritter,
		einePost(einNachfolger(), func(u *domain.Handelsumschlag) {
			u.Absender = dritter
		}))
	if !errors.Is(fehler, domain.ErrErlaubnis) {
		t.Fatalf("es kam %v", fehler)
	}

	if len(briefkasten.angebote) != 0 {
		t.Errorf("geschrieben wurden %d Angebote", len(briefkasten.angebote))
	}
}

// **Eine Nachricht zu derselben Revision bringt Stimmen** —
// #1477.
//
// # Der Fund, für den es diesen Lauf gibt
//
// **Eine Zustimmung konnte nicht reisen.** [domain.Handelsangebot.Nimmt]
// verlangt die **nächste** Revision; eine Stimme tritt zu derselben hinzu.
// Zwei Prozesse kamen deshalb nie über `angeboten` hinaus — jeder hatte seine
// eigene Stimme und keiner die des anderen. Gemessen am 2026-09-19 mit zwei
// `mse`-Prozessen über einen echten Draht.
func TestEineNachrichtZurSelbenRevisionBringtStimmen(t *testing.T) {
	t.Parallel()

	briefkasten := &briefkastenprobe{}

	hier := derHeutigeStand()

	post := service.NeueHandelspost(briefkasten,
		standprobe{stand: hier},
		kontaktnachschlagprobe{kontakt: einHandelnder()})

	drueben := derHeutigeStand()
	drueben.Zustimmungen = []string{dasGegenueber}

	antwort, fehler := post.Empfangen(context.Background(), dasGegenueber,
		einePost(drueben, nil))
	if fehler != nil {
		t.Fatalf("Empfangen: %v", fehler)
	}

	if antwort.Art != domain.ZustellungNeu {
		t.Errorf("die Zustellung gilt als %q", antwort.Art)
	}

	if len(briefkasten.angebote) != 1 {
		t.Fatalf("geschrieben wurden %d Angebote", len(briefkasten.angebote))
	}

	// **Geschrieben wird der hiesige Stand mit der fremden Stimme** und
	// nicht das, was ankam: was drüben gilt, ist eine Behauptung.
	geschrieben := briefkasten.angebote[0]

	if len(geschrieben.Zustimmungen) != 1 ||
		geschrieben.Zustimmungen[0] != dasGegenueber {
		t.Errorf("geschrieben wurden die Stimmen %v", geschrieben.Zustimmungen)
	}
}

// **Ein Wechsel über die Leitung sagt dem Bestand, was er bewirken soll** —
// Befund 3 des Zweitblicks über Phase 5.6 und Befund 3 seines Nachlaufs.
//
// # Warum das eine Weitergabe ist und keine Handlung
//
// **Weil Vermerk, Revision und Wirkung zusammen stehen müssen.** Führte der
// Dienst sie vorher für sich aus, blieb nach einem gescheiterten Schreiben
// eine Zusage als „übergeben" stehen, während der Vorgang noch auf
// `vorbereitet` stand — und `Freigeben` gab sie nie wieder frei. Der Dienst
// **entscheidet**, der Bestand **führt aus**, in einer Transaktion.
func TestEinWechselUeberDieLeitungSagtDemBestandWasErBewirkt(t *testing.T) {
	t.Parallel()

	briefkasten := &briefkastenprobe{}

	hier := derHeutigeStand()
	hier.Zustand = domain.HandelVorbereitet
	hier.Zustimmungen = []string{"bich"}

	post := service.NeueHandelspost(briefkasten,
		standprobe{stand: hier},
		kontaktnachschlagprobe{kontakt: einHandelnder()})

	drueben := hier
	drueben.Zustimmungen = []string{dasGegenueber}

	if _, fehler := post.Empfangen(context.Background(), dasGegenueber,
		einePost(drueben, nil)); fehler != nil {
		t.Fatalf("Empfangen: %v", fehler)
	}

	if briefkasten.wirkung != domain.ZusagenUebergeben {
		t.Errorf("weitergegeben wurde die Wirkung %q", briefkasten.wirkung)
	}
}

// **Und ein Abbruch über die Leitung sagt: freigeben.**
func TestEinAbbruchUeberDieLeitungSagtFreigeben(t *testing.T) {
	t.Parallel()

	briefkasten := &briefkastenprobe{}

	hier := derHeutigeStand()

	post := service.NeueHandelspost(briefkasten,
		standprobe{stand: hier},
		kontaktnachschlagprobe{kontakt: einHandelnder()})

	// **Der Initiator bricht ab** — `bich` hat den Vorgang eröffnet.
	drueben := hier
	drueben.Zustand = domain.HandelAbgebrochen

	if _, fehler := post.Empfangen(context.Background(), "bich",
		einePost(drueben, func(u *domain.Handelsumschlag) {
			u.Absender = "bich"
		})); fehler != nil {
		t.Fatalf("Empfangen: %v", fehler)
	}

	if briefkasten.wirkung != domain.ZusagenFreigeben {
		t.Errorf("weitergegeben wurde die Wirkung %q", briefkasten.wirkung)
	}
}

// **Ein Vorschlag lässt sie in Ruhe** — und das ist die Mehrzahl.
func TestEinVorschlagLaesstDieZusagenInRuhe(t *testing.T) {
	t.Parallel()

	briefkasten := &briefkastenprobe{}

	post := service.NeueHandelspost(briefkasten,
		standprobe{stand: derHeutigeStand()},
		kontaktnachschlagprobe{kontakt: einHandelnder()})

	if _, fehler := post.Empfangen(context.Background(), dasGegenueber,
		einePost(einNachfolger(), nil)); fehler != nil {
		t.Fatalf("Empfangen: %v", fehler)
	}

	if briefkasten.wirkung != domain.ZusagenBleiben {
		t.Errorf("weitergegeben wurde die Wirkung %q", briefkasten.wirkung)
	}
}
