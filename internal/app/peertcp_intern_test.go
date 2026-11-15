package app

// **Zwei Anwendungen über eine echte Leitung** — Werkstück N2 von Phase 5.4
// (#1455).
//
// # Der Lauf, für den es den Weg „Adresse von Hand" gibt
//
// **Er fährt alles außer dem zweiten Rechner.** Zwei Anwendungen mit eigenem
// Bestand, eigener Identität und eigenen Kontakten; die eine öffnet einen
// **echten Port** auf der Rückschleife, die andere trägt dessen Adresse als
// Kontaktadresse ein und wählt sie. Dazwischen liegen TCP, TLS 1.3 mit einem
// selbst unterschriebenen Zeugnis, das Festnageln auf die erwartete Kennung,
// die Kanalbindung aus dem Schlüsselmaterial, die Rahmung und der
// Dreischritt-Handschlag.
//
// **Was er nicht zeigt** — und das gehört dazu, sonst reicht die Messung
// weiter als ihre Aussage (§10): die lokale Netzwerkberechtigung von macOS,
// die Firewall, isolierte Gäste-WLANs, Schlafen und Aufwachen, den Netzwechsel
// und zwei verschiedene Rechner. Das bleiben Handprüfungen.

import (
	"net"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/peer"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// zweiUeberTCP baut zwei Anwendungen **ohne** eingesetzte Leitung.
//
// **Der Unterschied zu [zweiAnwendungen]**: dort hängt eine Schleife im
// selben Prozess dazwischen, hier nichts — es gilt, was das Programm im
// Betrieb nähme.
func zweiUeberTCP(t *testing.T) (hier, dort *seite) {
	t.Helper()

	hier, dort = &seite{}, &seite{}

	for _, wer := range []*seite{hier, dort} {
		wer.anwendung = PeerflaecheUeber(
			bestandFuerDenRiegel(t)).MitPeerbund(&peerprobebund{})
		wer.ctx = wer.anwendung.imProfil(riegelkontext(t))

		antwort := wer.anwendung.PeeridentitaetSicherstellen(wer.ctx)
		if antwort.Fehler != nil {
			t.Fatalf("Identität anlegen: %+v", antwort.Fehler)
		}

		wer.kennung = peer.Kennung(antwort.Kennung)
		wer.abdruck = antwort.Fingerabdruck
	}

	hier.kennt(t, dort)
	dort.kennt(t, hier)

	return hier, dort
}

// **Eine Sitzung kommt über einen echten Port zustande.**
func TestEineSitzungKommtUeberEinenEchtenPortZustande(t *testing.T) {
	t.Parallel()

	hier, dort := zweiUeberTCP(t)

	// **Die annehmende Seite macht auf** — auf der Rückschleife und mit
	// einem Port, den das Betriebssystem aussucht.
	gestartet := dort.anwendung.HorchStarten(dort.ctx, "127.0.0.1:0")
	if gestartet.Fehler != nil {
		t.Fatalf("HorchStarten: %+v", gestartet.Fehler)
	}

	t.Cleanup(dort.anwendung.horchenBeenden)

	if !gestartet.Stand.Laeuft || gestartet.Stand.Adresse == "" {
		t.Fatalf("der Stand ist %+v", gestartet.Stand)
	}

	// **Und die wählende trägt ein, wo sie sitzt** — von Hand, so wie ein
	// Mensch es abtippte.
	if antwort := hier.anwendung.PeerkontaktAdresse(
		hier.ctx, string(dort.kennung), gestartet.Stand.Adresse,
	); antwort.Fehler != nil {
		t.Fatalf("PeerkontaktAdresse: %+v", antwort.Fehler)
	}

	griff, fehler := hier.anwendung.sitzungsverwaltung().Oeffne(
		hier.ctx, dort.kennung, domain.ZweckTeilen)
	if fehler != nil {
		t.Fatalf("Oeffne: %v", fehler)
	}

	sitzung := griff.Sitzung()
	if sitzung.Kennung == "" || sitzung.Gegenueber != dort.kennung {
		t.Fatalf("die Sitzung ist %+v", sitzung)
	}

	// **Der Fingerabdruck ist der, den ein Mensch verglichen hat.**
	if sitzung.Fingerabdruck != dort.abdruck {
		t.Errorf("der Abdruck ist %q", sitzung.Fingerabdruck)
	}

	if fehler := griff.Sende(
		hier.ctx, []byte("hallo"), service.Zuverlaessig); fehler != nil {
		t.Fatalf("Sende: %v", fehler)
	}

	if fehler := griff.Leere(hier.ctx); fehler != nil {
		t.Fatalf("Leere: %v", fehler)
	}
}

// **Ohne eingetragene Adresse sagt es die Sache und nicht die Leitung.**
//
// **Der Unterschied zählt** (§8): „für dieses Gegenüber ist keine Adresse
// eingetragen" ist eine Auskunft, mit der ein Mensch etwas anfangen kann;
// „Verbindung zu :0 fehlgeschlagen" wäre eine über den Bauzustand des
// Programms.
func TestOhneEingetrageneAdresseSagtEsDieSache(t *testing.T) {
	t.Parallel()

	hier, dort := zweiUeberTCP(t)

	_, fehler := hier.anwendung.sitzungsverwaltung().Oeffne(
		hier.ctx, dort.kennung, domain.ZweckTeilen)

	if fehler == nil {
		t.Fatal("die Sitzung kam ohne Adresse zustande")
	}

	for _, teil := range []string{"keine Adresse", dort.abdruck} {
		if !strings.Contains(fehler.Error(), teil) {
			t.Fatalf("der Satz lautet: %v — %q fehlt", fehler, teil)
		}
	}
}

// **Eine Adresse, an der niemand sitzt, endet mit einem Fehler.**
func TestEineAdresseAnDerNiemandSitztEndet(t *testing.T) {
	t.Parallel()

	hier, dort := zweiUeberTCP(t)

	// **Ein Port, der eben noch da war** — so ist er sicher frei, und der
	// Lauf hängt nicht von einer erfundenen Adresse ab.
	gestartet := dort.anwendung.HorchStarten(dort.ctx, "127.0.0.1:0")
	if gestartet.Fehler != nil {
		t.Fatalf("HorchStarten: %+v", gestartet.Fehler)
	}

	adresse := gestartet.Stand.Adresse

	dort.anwendung.horchenBeenden()

	if antwort := hier.anwendung.PeerkontaktAdresse(
		hier.ctx, string(dort.kennung), adresse); antwort.Fehler != nil {
		t.Fatalf("PeerkontaktAdresse: %+v", antwort.Fehler)
	}

	if _, fehler := hier.anwendung.sitzungsverwaltung().Oeffne(
		hier.ctx, dort.kennung, domain.ZweckTeilen); fehler == nil {
		t.Fatal("die Sitzung kam an einer geschlossenen Tür zustande")
	}
}

// **Ohne Identität wird kein Port geöffnet.**
//
// **Der Ausweis vor dem Port**: ein Profil ohne Identität kann keinen
// Handschlag führen, und ein offener Port ohne Identität sagte jedem, dass es
// dieses Programm gibt, um dann jeden abzuweisen.
func TestOhneIdentitaetWirdKeinPortGeoeffnet(t *testing.T) {
	t.Parallel()

	anwendung := PeerflaecheUeber(
		bestandFuerDenRiegel(t)).MitPeerbund(&peerprobebund{})
	ctx := anwendung.imProfil(riegelkontext(t))

	antwort := anwendung.HorchStarten(ctx, "127.0.0.1:0")
	if antwort.Fehler == nil || antwort.Fehler.Code != "nichtGefunden" {
		t.Fatalf("es kam %+v", antwort.Fehler)
	}
}

// **Und ohne Identität wird auch nicht gewählt.**
//
// **Derselbe Grund, die andere Seite**: wer wählt, ohne unterschreiben zu
// können, hat dem Gegenüber gesagt, dass es dieses Programm gibt, und steht
// dann da.
func TestOhneIdentitaetWirdAuchNichtGewaehlt(t *testing.T) {
	t.Parallel()

	anwendung := PeerflaecheUeber(
		bestandFuerDenRiegel(t)).MitPeerbund(&peerprobebund{})
	ctx := anwendung.imProfil(riegelkontext(t))

	// **Ein Kontakt mit Adresse, aber kein eigener Schlüssel.**
	fremd, fehler := peer.Erzeugen(nil)
	if fehler != nil {
		t.Fatalf("Erzeugen: %v", fehler)
	}

	if antwort := anwendung.PeerkontaktAufnehmen(
		ctx, string(fremd.Kennung()), "Der andere"); antwort.Fehler != nil {
		t.Fatalf("Aufnehmen: %+v", antwort.Fehler)
	}

	if antwort := anwendung.PeerkontaktAdresse(
		ctx, string(fremd.Kennung()), "127.0.0.1:1"); antwort.Fehler != nil {
		t.Fatalf("PeerkontaktAdresse: %+v", antwort.Fehler)
	}

	// **Über den Wähler und nicht über `Oeffne`** — der prüft den Ausweis
	// selbst und käme hier nie bis zur Leitung.
	if _, fehler := (adresswaehler{app: anwendung}).Waehle(
		ctx, fremd.Kennung(), ""); fehler == nil {
		t.Fatal("der Wähler kam ohne Identität durch")
	}
}

// **Ein Port, der schon belegt ist, wird gemeldet.**
//
// **Über den echten Lauscher und nicht über eine Naht** — was hier gemessen
// wird, ist der Weg von `leitung.Horche` zurück an die Oberfläche.
//
// # Warum ein belegter Port und kein erfundener Wirt
//
// **Weil ein Name aufgelöst werden müsste.** Der erste Anlauf gab
// `kein-wirt.invalid:41234` an, und das ist ein Griff nach draußen:
// `net.Listen` fragt dafür den Namensdienst dieses Rechners, und was dann
// geschieht, hängt am Netz — §3 verlangt ausdrücklich, dass ein gewöhnlicher
// Lauf das nicht tut.
//
// **Ein zweimal belegter Port ist derselbe Fehlerweg ohne diesen Preis**: er
// bleibt auf der Rückschleife, und die Antwort kommt sofort.
func TestEinBelegterPortWirdGemeldet(t *testing.T) {
	t.Parallel()

	anwendung := PeerflaecheUeber(
		bestandFuerDenRiegel(t)).MitPeerbund(&peerprobebund{})
	ctx := anwendung.imProfil(riegelkontext(t))

	if antwort := anwendung.PeeridentitaetSicherstellen(
		ctx); antwort.Fehler != nil {
		t.Fatalf("Identität anlegen: %+v", antwort.Fehler)
	}

	besetzer, fehler := (&net.ListenConfig{}).Listen(
		t.Context(), "tcp", "127.0.0.1:0")
	if fehler != nil {
		t.Fatalf("Listen: %v", fehler)
	}

	t.Cleanup(func() {
		_ = besetzer.Close() //nolint:errcheck // nur ein Lauf
	})

	antwort := anwendung.HorchStarten(ctx, besetzer.Addr().String())
	if antwort.Fehler == nil {
		t.Fatal("der Horcher stand auf einem belegten Port")
	}
}

// **Ein Profil ohne Identität bekommt den Satz, der zu ihm gehört.**
//
// # Der Befund, für den es diesen Lauf gibt
//
// **Die Sichtprüfung hat ihn gefunden** (2026-09-14): wer im Fenster
// „Erreichbar werden" drückte, ohne eine Identität zu haben, las „Der gesuchte
// Eintrag wurde nicht gefunden. Vielleicht hat ihn jemand anderes gelöscht —
// lade die Liste neu." Das ist der allgemeine Satz zu `nichtGefunden`, und er
// war hier dreifach falsch: nichts ist gelöscht worden, niemand anderes war
// beteiligt, und die Liste neu zu laden hilft nicht.
//
// **Kein Test hat es gezeigt**, weil jeder von ihnen den **Code** prüfte und
// keiner den Schlüssel, an dem das Fenster seinen Satz aussucht (§8).
func TestOhneIdentitaetKommtDerEigeneSchluessel(t *testing.T) {
	t.Parallel()

	anwendung := PeerflaecheUeber(
		bestandFuerDenRiegel(t)).MitPeerbund(&peerprobebund{})
	ctx := anwendung.imProfil(riegelkontext(t))

	for name, antwort := range map[string]Horchantwort{
		"starten": anwendung.HorchStarten(ctx, "127.0.0.1:0"),
	} {
		if antwort.Fehler == nil {
			t.Fatalf("%s kam durch", name)
		}

		if antwort.Fehler.Schluessel != "ohnePeeridentitaet" {
			t.Fatalf("%s gab den Schlüssel %q",
				name, antwort.Fehler.Schluessel)
		}
	}
}
