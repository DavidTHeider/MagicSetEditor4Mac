package leitung

// **Ein eigenes Zeugnis, das sich tauschen lässt** — NET-03.AK-02
// (#1707),
// entschieden in
// [0154](../../docs/entscheidungen/0154-das-dienstzeugnis-ist-nicht-die-identitaet.md).
//
// # Warum am Handschlag und nicht per Signal
//
// **Weil ein Dienst, den das Fenster gestartet hat, kein Signal bekommt**,
// ohne dass jemand es ihm schickt — und wer eine PKI betreibt, tauscht die
// Dateien, nicht den Prozess. Am Handschlag nachzusehen kostet einen `stat`
// je Datei, und höchstens alle [Nachsehfrist].

import (
	"crypto/tls"
	"os"
	"sync"
	"time"
)

// Nachsehfrist ist, wie oft das Werk höchstens nach neuen Dateien sieht.
const Nachsehfrist = 5 * time.Second

// Zeugniswerk hält ein geprüftes eigenes Zeugnis und tauscht es, wenn sich
// seine Dateien ändern und das neue die Prüfung besteht.
type Zeugniswerk struct {
	dateien  Zeugnisdateien
	adressen []string
	uhr      func() time.Time
	melde    func(error)

	schloss sync.Mutex

	// aktuell wird nie verändert, nur ersetzt (Zweitblick über Phase 6.7,
	// Z4): ein Zeugnis, das an TLS ging, liest ein langsamer Handschlag noch,
	// wenn das nächste schon gilt — Go 1.27.1 verlangt, es dann nicht mehr
	// anzufassen (`crypto/tls/common.go`, `GetCertificate`).
	aktuell *tls.Certificate
	befund  Zeugnisbefund
	stand   dateistand
	zuletzt time.Time
}

// dateistand ist, woran eine Änderung zu erkennen ist.
type dateistand [4]struct {
	zeit    int64
	groesse int64
}

// NeuesZeugniswerk prüft die Dateien und gibt ein Werk — oder den Grund,
// warum der Dienst nicht startet.
//
// **melde** bekommt jeden Befund eines misslungenen Nachladens; das alte
// Zeugnis gilt dann weiter.
func NeuesZeugniswerk(
	dateien Zeugnisdateien, adressen []string, uhr func() time.Time, melde func(error),
) (*Zeugniswerk, error) {
	jetzt := uhr()

	stand, fehler := dateistandVon(dateien)
	if fehler != nil {
		return nil, fehler
	}

	paar, befund, fehler := PruefeZeugnisFuer(dateien, adressen, jetzt)
	if fehler != nil {
		return nil, fehler
	}

	return &Zeugniswerk{
		dateien: dateien, adressen: adressen, uhr: uhr, melde: melde,
		aktuell: &paar, befund: befund, stand: stand, zuletzt: jetzt,
	}, nil
}

// Einstellung ist die TLS-Einstellung des Dienstes.
//
// **TLS 1.3 und nichts darunter** — dieselbe Untergrenze wie beim Zeugnis
// auf der Peeridentität.
//
// **Und ohne Wiederaufnahme** (Zweitblick über Phase 6.7, Z5): eine Sitzung,
// die über ein Ticket wiederkommt, überspringt `GetCertificate`
// (`handshake_server_tls13.go`) — und damit das Nachladen und die Sperre für
// ein abgelaufenes Zeugnis. Ein voller Handschlag kostet eine Signatur mehr.
func (w *Zeugniswerk) Einstellung() *tls.Config {
	return &tls.Config{
		GetCertificate: w.zeugnisFuer, MinVersion: tls.VersionTLS13,
		SessionTicketsDisabled: true,
	}
}

// Befund beschreibt das Zeugnis, das gerade gilt.
func (w *Zeugniswerk) Befund() Zeugnisbefund {
	w.schloss.Lock()
	defer w.schloss.Unlock()

	return w.befund
}

func (w *Zeugniswerk) zeugnisFuer(*tls.ClientHelloInfo) (*tls.Certificate, error) {
	w.schloss.Lock()
	defer w.schloss.Unlock()

	jetzt := w.uhr()
	if jetzt.Sub(w.zuletzt) >= Nachsehfrist {
		w.zuletzt = jetzt
		w.nachsehen(jetzt)
	}

	// **Ein abgelaufenes Zeugnis wird nie vorgezeigt.** Ein Klient, der
	// prüft, wiese es ab; einer, der nicht prüft, sollte es nicht bekommen.
	if jetzt.After(w.aktuell.Leaf.NotAfter) {
		return nil, abweisen(nil, "das Zeugnis des Dienstes ist abgelaufen — seit %s",
			w.aktuell.Leaf.NotAfter.UTC().Format(time.RFC3339))
	}

	return w.aktuell, nil
}

// nachsehen tauscht, wenn sich die Dateien geändert haben und das neue Paar
// besteht.
//
// **Auch ein Stand, der die Prüfung nicht besteht, wird gemerkt** — sonst
// meldete jeder Blick denselben Befund neu. Geprüft wird wieder, sobald
// sich eine der Dateien noch einmal ändert: so kommt ein Austausch durch,
// der die zweite Datei erst nach der ersten schreibt.
func (w *Zeugniswerk) nachsehen(jetzt time.Time) {
	stand, fehler := dateistandVon(w.dateien)
	if fehler != nil {
		if stand != w.stand {
			w.stand = stand
			w.melde(fehler)
		}

		return
	}

	if stand == w.stand {
		return
	}

	w.stand = stand

	paar, befund, fehler := PruefeZeugnisFuer(w.dateien, w.adressen, jetzt)
	if fehler != nil {
		w.melde(fehler)

		return
	}

	w.aktuell, w.befund = &paar, befund
}

// dateistandVon liest Zeit und Größe jeder genannten Datei.
//
// **Eine fehlende Datei hinterlässt eine Null in ihrem Feld** — der Stand
// ist damit ein anderer als vorher, und der Fehler sagt, welche.
func dateistandVon(dateien Zeugnisdateien) (dateistand, error) {
	var (
		stand  dateistand
		ersten error
	)

	for stelle, pfad := range []string{dateien.Zertifikat, dateien.Schluessel, dateien.Kette, dateien.Anker} {
		if pfad == "" {
			continue
		}

		info, fehler := os.Stat(pfad)
		if fehler != nil {
			if ersten == nil {
				ersten = abweisen(fehler, "%s lässt sich nicht lesen", pfad)
			}

			continue
		}

		stand[stelle].zeit, stand[stelle].groesse = info.ModTime().UnixNano(), info.Size()
	}

	return stand, ersten
}
