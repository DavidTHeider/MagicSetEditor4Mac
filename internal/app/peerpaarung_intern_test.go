package app

// **Zwei Anwendungen auf einem Rechner** — Werkstück N2 von Phase 5.4
// (#1455).
//
// # Was diese Vorrichtung ist
//
// **Der Lauf, der beantwortet, ob Spiele und Übertragungen klappen werden.**
// Zwei Anwendungen, jede mit eigenem Bestand, eigener Identität und eigenen
// Kontakten, verbunden über eine Schleife aus `internal/leitung`. Darüber
// läuft **jede** Schicht: Identität, Kontakte, Rechte, Handschlag, Hülle,
// Rahmung, Sitzungsverwaltung.
//
// **Was sie nicht zeigt** — und das gehört dazu, sonst reicht die Messung
// weiter als ihre Aussage (§10): die lokale Netzwerkberechtigung von macOS,
// die Firewall, isolierte Gäste-WLANs, Schlafen und Aufwachen, den
// Netzwechsel und zwei verschiedene Rechner. Das bleiben Handprüfungen.

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/leitung"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/peer"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// seite ist eine der beiden Anwendungen samt ihrem Ende der Leitung.
type seite struct {
	anwendung *App
	ctx       context.Context //nolint:containedctx // die Vorrichtung reicht ihn durch
	kennung   peer.Kennung
	abdruck   string
	leitung   *leitung.Leitung

	// meldungen und nebenbei fangen auf, was die Anwendung ans Fenster
	// meldet — **ab dem Aufbau und nicht ab dem ersten Ruf**.
	//
	// **Weil ein später gesetzter Meldeweg ein Rennen ist.** `App.melden` ist
	// ein Feld, das die Anwendung **einmal** bekommt, bevor etwas läuft; die
	// Sitzungen lesen es aus ihren eigenen Goroutinen. Ein Lauf, der es nach
	// dem Beitritt austauscht, schreibt daneben — und der Detektor hat es am
	// 2026-09-17 in der Stufe `test` gefunden, bei grüner Vorprüfung und
	// grünem Paketlauf. Eine Vorrichtung, die etwas tut, was der Betrieb nie
	// tut, prüft die Vorrichtung (§3).
	meldungen chan string
	nebenbei  chan NebenbeiDTO
}

// altesVerwerfen wirft weg, was vor diesem Punkt gemeldet wurde.
//
// **Damit eine Zusicherung nicht auf etwas von vorhin trifft** (§3): wer nach
// einem Zug auf eine Meldung wartet, will die **nächste** und nicht eine aus
// dem Aufbau. Vorher leisteten das die beiden Sammler, die genau hier gesetzt
// wurden — geblieben ist ihre Wirkung, weggefallen ist das Rennen.
//
// **Und was es heute wirklich wegwirft, ist nichts**: gemessen am 2026-09-17
// an `TestEinGastSetztSichUndSeinZugKommtAn` stehen an dieser Stelle **null**
// Meldungen im Kanal. Die Zeile ist damit kein Beleg und keine Absicherung
// gegen etwas Gesehenes — sie hält die Bedeutung des Laufs dieselbe wie vor
// der Umstellung, und mehr behauptet sie nicht (§10).
func (s *seite) altesVerwerfen() {
	for {
		select {
		case <-s.meldungen:
		case <-s.nebenbei:
		default:
			return
		}
	}
}

// zweiAnwendungen baut zwei Anwendungen, die einander kennen und vertrauen.
//
// **Beide Seiten geben einander den vollen Satz frei** — was eine Freigabe
// bewirkt, prüfen die Läufe in `service`; hier geht es um den Weg durch alle
// Schichten.
func zweiAnwendungen(t *testing.T) (hier, dort *seite) {
	t.Helper()

	hier, dort = &seite{}, &seite{}

	for _, wer := range []*seite{hier, dort} {
		// **Je Seite ein eingesetzter Bund** (§3): ohne ihn schriebe dieser
		// Lauf zwei Identitäten in den Anmeldebund dieses Rechners — genau
		// der Fehler, den `internal/schluesselbund` in seinem Kopf benennt.
		// Er ist hier auch aufgetreten: der erste Anlauf hing zehn Sekunden
		// und scheiterte, statt still zu schreiben.
		wer.anwendung = PeerflaecheUeber(
			bestandFuerDenRiegel(t)).MitPeerbund(&peerprobebund{})
		wer.ctx = wer.anwendung.imProfil(riegelkontext(t))

		antwort := wer.anwendung.PeeridentitaetSicherstellen(wer.ctx)
		if antwort.Fehler != nil {
			t.Fatalf("Identität anlegen: %+v", antwort.Fehler)
		}

		wer.kennung = peer.Kennung(antwort.Kennung)
		wer.abdruck = antwort.Fingerabdruck

		// **Ein Meldeweg, der nichts ans Fenster gibt, sondern sammelt.**
		//
		// Ohne einen eigenen ruft der Tischmelder `runtime.EventsEmit` mit
		// dem Kontext eines Laufs — und Wails lehnt ihn ab („An invalid
		// context was passed"), sichtbar als **rotes Paket ohne einen
		// einzigen roten Lauf**.
		//
		// **Und er steht hier, bevor irgendetwas läuft**, statt in jedem Lauf
		// nach dem Beitritt: siehe die beiden Kanäle an [seite].
		wer.meldungen = make(chan string, 64)
		wer.nebenbei = make(chan NebenbeiDTO, 64)
		wer.anwendung.melden = wer.sammeln
	}

	hiesige, dortige := leitung.Schleife("kanal-schleife")

	t.Cleanup(func() {
		_ = hiesige.Schliesse() //nolint:errcheck // nur ein Lauf
		_ = dortige.Schliesse() //nolint:errcheck // nur ein Lauf
	})

	hier.leitung, dort.leitung = hiesige, dortige

	hier.anwendung.MitWaehler(schleifenwaehler{leitung: hiesige})
	dort.anwendung.MitWaehler(schleifenwaehler{leitung: dortige})

	hier.kennt(t, dort)
	dort.kennt(t, hier)

	return hier, dort
}

// kennt trägt das Gegenüber ein, bestätigt es und gibt jeden Zweck frei.
func (s *seite) kennt(t *testing.T, anderer *seite) {
	t.Helper()

	if antwort := s.anwendung.PeerkontaktAufnehmen(
		s.ctx, string(anderer.kennung), "Der andere"); antwort.Fehler != nil {
		t.Fatalf("Aufnehmen: %+v", antwort.Fehler)
	}

	if antwort := s.anwendung.PeerkontaktBestaetigen(
		s.ctx, string(anderer.kennung), anderer.abdruck,
	); antwort.Fehler != nil {
		t.Fatalf("Bestaetigen: %+v", antwort.Fehler)
	}

	for _, zweck := range domain.DieZwecke {
		if zweck == domain.ZweckPaarung {
			continue
		}

		if antwort := s.anwendung.PeerzweckErlauben(
			s.ctx, string(anderer.kennung), string(zweck),
		); antwort.Fehler != nil {
			t.Fatalf("Erlauben %q: %+v", zweck, antwort.Fehler)
		}
	}
}

// eineSitzungZu führt den Handschlag und gibt den Griff dieser Seite.
func (s *seite) eineSitzungZu(
	t *testing.T, anderer *seite, zweck domain.Zweck,
) *service.Sitzungsgriff {
	t.Helper()

	fertig := make(chan error, 1)

	go func() {
		_, fehler := anderer.anwendung.sitzungsverwaltung().Annehmen(
			anderer.ctx, anderer.leitung)
		fertig <- fehler
	}()

	griff, fehler := s.anwendung.sitzungsverwaltung().Oeffne(
		s.ctx, anderer.kennung, zweck)

	if gegenfehler := <-fertig; gegenfehler != nil {
		t.Fatalf("die Gegenseite: %v", gegenfehler)
	}

	if fehler != nil {
		t.Fatalf("Oeffne: %v", fehler)
	}

	return griff
}

// **Zwei Anwendungen führen einen echten Handschlag und tauschen eine
// Nachricht.**
//
// # Der Lauf, für den es N2 Schritt 1 gibt
//
// **Er fährt jede Schicht außer den Bytes auf dem Draht.** Zwei Identitäten
// aus zwei Schlüsselbünden, zwei Bestände, zwei Kontaktlisten, zwei
// Rechteprüfungen, drei unterschriebene Nachrichten, eine Rahmung — und am
// Ende dieselbe Sitzungskennung auf beiden Seiten, die keine von beiden
// gewählt hat.
func TestZweiAnwendungenTauschenEineNachricht(t *testing.T) {
	t.Parallel()

	hier, dort := zweiAnwendungen(t)

	griff := hier.eineSitzungZu(t, dort, domain.ZweckTeilen)

	sitzung := griff.Sitzung()
	if sitzung.Kennung == "" || sitzung.Gegenueber != dort.kennung {
		t.Fatalf("die Sitzung ist %+v", sitzung)
	}

	// **Der Fingerabdruck kommt aus der Kennung** und stimmt mit dem
	// überein, den ein Mensch verglichen hat.
	if sitzung.Fingerabdruck != dort.abdruck {
		t.Errorf("der Abdruck ist %q", sitzung.Fingerabdruck)
	}

	if fehler := griff.Sende(
		hier.ctx, []byte("hallo"), service.Zuverlaessig); fehler != nil {
		t.Fatalf("Sende: %v", fehler)
	}

	angekommen := make(chan []byte, 1)

	go func() {
		gelesen, fehler := dort.leitung.Empfange(dort.ctx)
		if fehler != nil {
			t.Errorf("Empfange: %v", fehler)
		}

		angekommen <- gelesen
	}()

	if fehler := griff.Leere(hier.ctx); fehler != nil {
		t.Fatalf("Leere: %v", fehler)
	}

	if gelesen := <-angekommen; string(gelesen) != "hallo" {
		t.Fatalf("angekommen ist %q", gelesen)
	}
}

// **Und ohne Freigabe kommt keine Sitzung zustande — über den ganzen Weg.**
//
// **Das ist derselbe Satz wie in `service`, aber hier steht er auf echten
// Kontakten**: die Freigabe wird über die Bindung zurückgenommen, und die
// nächste Sitzung scheitert.
func TestOhneFreigabeKommtKeineSitzungZustande(t *testing.T) {
	t.Parallel()

	hier, dort := zweiAnwendungen(t)

	if antwort := dort.anwendung.PeerzweckEntziehen(
		dort.ctx, string(hier.kennung), string(domain.ZweckHandel),
	); antwort.Fehler != nil {
		t.Fatalf("Entziehen: %+v", antwort.Fehler)
	}

	fertig := make(chan error, 1)

	go func() {
		_, fehler := dort.anwendung.sitzungsverwaltung().Annehmen(
			dort.ctx, dort.leitung)
		fertig <- fehler
	}()

	if _, fehler := hier.anwendung.sitzungsverwaltung().Oeffne(
		hier.ctx, dort.kennung, domain.ZweckHandel); fehler == nil {
		t.Error("die Sitzung kam ohne Freigabe zustande")
	}

	if fehler := <-fertig; fehler == nil {
		t.Fatal("die annehmende Seite kam durch")
	}
}

// klagendeLeitung reicht durch und klagt beim Niederlegen.
type klagendeLeitung struct{ innen *leitung.Leitung }

func (k klagendeLeitung) Kanal() string { return k.innen.Kanal() }

func (k klagendeLeitung) Gegenstelle() peer.Kennung { return k.innen.Gegenstelle() }

func (k klagendeLeitung) Sende(ctx context.Context, nachricht []byte) error {
	return k.innen.Sende(ctx, nachricht)
}

func (k klagendeLeitung) Empfange(ctx context.Context) ([]byte, error) {
	return k.innen.Empfange(ctx)
}

func (k klagendeLeitung) Schliesse() error {
	_ = k.innen.Schliesse() //nolint:errcheck // hier zählt die Klage

	return errors.New("die Leitung klagt beim Niederlegen")
}

// klagenderWaehler gibt eine Leitung, die beim Niederlegen klagt.
type klagenderWaehler struct{ leitung klagendeLeitung }

func (k klagenderWaehler) Waehle(
	context.Context, peer.Kennung, string,
) (service.Verbindung, error) {
	return k.leitung, nil
}

// **Eine Leitung, die beim Niederlegen klagt, hält den Profilwechsel nicht
// auf.**
//
// # Warum das so und nicht anders ist
//
// **Weil der Wechsel stattfindet.** Ein Profil offen zu lassen, weil eine
// Leitung klagt, hieße: das Gegenüber entscheidet, in welchem Profil dieser
// Rechner arbeitet. Die Klage gehört ins Protokoll und nicht in den Weg.
func TestEineKlagendeLeitungHaeltDenWechselNichtAuf(t *testing.T) {
	t.Parallel()

	hier, dort := zweiAnwendungen(t)

	hier.anwendung.MitWaehler(
		klagenderWaehler{leitung: klagendeLeitung{innen: hier.leitung}})

	hier.eineSitzungZu(t, dort, domain.ZweckTeilen)

	hier.anwendung.MitProfilwechsel(wechselprobe{})
	hier.anwendung.profilUmstellen(riegelkontext(t), 5)

	if len(hier.anwendung.sitzungsverwaltung().Alle()) != 0 {
		t.Fatal("die Sitzung überlebte den Wechsel")
	}
}

// sammeln legt jede Meldung in die Kanäle dieser Seite — ohne je zu klemmen.
//
// **Ein voller Kanal verwirft und hält nicht auf.** Ein Melder, der auf einen
// Leser wartet, hielte die Sitzung an, aus der er gerufen wird — und der Lauf
// endete in seiner Frist statt an seiner Zusicherung.
func (s *seite) sammeln(_ context.Context, ereignis string, daten ...any) {
	select {
	case s.meldungen <- ereignis:
	default:
	}

	if ereignis != EreignisPartienebenbei || len(daten) < 2 {
		return
	}

	eines, passt := daten[1].(NebenbeiDTO)
	if !passt {
		return
	}

	select {
	case s.nebenbei <- eines:
	default:
	}
}
