package leitung

// **Ein eigenes Zeugnis für den Dienst** — NET-03
// (#1707),
// entschieden in
// [0154](../../docs/entscheidungen/0154-das-dienstzeugnis-ist-nicht-die-identitaet.md).
//
// # Warum es hier steht
//
// **Aus demselben Grund wie [Serverzeugnis]**: es ist die TLS-Einstellung
// eines HTTPS-Dienstes, und ein zweiter Bauweg in `internal/rest` wäre eine
// zweite Fassung derselben Frage. Anders ist nur, **wer** das Zeugnis
// ausstellt — eine Stelle des Hauses statt dieses Profils.
//
// # Warum vor dem Start geprüft wird
//
// **Weil ein Klient nur sieht, dass der Handschlag scheitert.** Ein falsches
// Paar, eine abgelaufene Gültigkeit oder eine Kette, die nicht bis zum Anker
// reicht, fiele sonst erst beim ersten Zugriff auf — und dann ohne den Satz,
// der sagt, was fehlt.

import (
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"net/netip"
	"os"
	"strings"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Zeugnisdateien sind die vier Pfade einer eigenen PKI.
type Zeugnisdateien struct {
	// Zertifikat ist das Blatt, als PEM — die Zwischenkette darf dahinter
	// mitstehen, wie viele Stellen es ausliefern.
	Zertifikat string

	// Schluessel ist der private Schlüssel des Blatts, als PEM, für niemanden
	// sonst lesbar.
	Schluessel string

	// Kette sind Zwischenzertifikate, als PEM — leer heißt: keine, oder sie
	// stehen im Zertifikat.
	Kette string

	// Anker sind die Vertrauensanker, gegen die die Kette geprüft wird —
	// leer heißt: die des Systems.
	Anker string
}

// Eigene sagt, ob es eine eigene PKI ist — ohne Zertifikat und Schlüssel
// bleibt es beim Zeugnis auf der Peeridentität.
func (d Zeugnisdateien) Eigene() bool {
	return d.Zertifikat != "" && d.Schluessel != ""
}

// Zeugnisbefund beschreibt ein geprüftes Zeugnis.
type Zeugnisbefund struct {
	// Fingerabdruck ist SHA-256 über das Blatt, in der Form von
	// `openssl x509 -fingerprint -sha256`.
	Fingerabdruck string

	GueltigBis time.Time
	Aussteller string

	// Namen sind die Namen und Adressen des Blatts (SAN).
	Namen []string
}

// PruefeZeugnis liest die Dateien und prüft sie für eine Bindungsadresse.
//
// **Die Reihenfolge ist die der Auskunft**: erst, ob die Dateien da und
// lesbar sind, dann, ob das Paar passt, dann Gültigkeit, Namen und Kette.
// Wer ein abgelaufenes Zertifikat hat, soll „abgelaufen" lesen und nicht,
// dass die Kette nicht trägt — die trägt dann auch nicht, aber das ist die
// Folge und nicht der Grund.
func PruefeZeugnis(
	dateien Zeugnisdateien, adresse string, jetzt time.Time,
) (tls.Certificate, Zeugnisbefund, error) {
	return PruefeZeugnisFuer(dateien, []string{adresse}, jetzt)
}

// PruefeZeugnisFuer liest die Dateien **einmal** und prüft das Blatt gegen
// jede Bindungsadresse.
//
// **Einmal und nicht je Adresse** — neuer Befund des Nachlaufs über Phase G:
// wer je Adresse neu liest, bestätigt bei einem Austausch dazwischen ein
// anderes Zertifikat als das, das danach gilt. Geprüft wird das Paar, das
// zurückgegeben wird. **Ohne Adresse wird keine geprüft** — das Paar selbst
// schon.
func PruefeZeugnisFuer(
	dateien Zeugnisdateien, adressen []string, jetzt time.Time,
) (tls.Certificate, Zeugnisbefund, error) {
	if !dateien.Eigene() {
		return tls.Certificate{}, Zeugnisbefund{}, abweisen(nil,
			"ein eigenes Zeugnis braucht Zertifikat und Schlüssel")
	}

	zertifikatPEM, blatt, zwischen, fehler := zertifikateLesen(dateien)
	if fehler != nil {
		return tls.Certificate{}, Zeugnisbefund{}, fehler
	}

	paar, fehler := paarBilden(dateien, zertifikatPEM)
	if fehler != nil {
		return tls.Certificate{}, Zeugnisbefund{}, fehler
	}

	// **Ohne Adresse prüft `blattPruefen` Gültigkeit und Namen**, aber an
	// keiner Bindung.
	if len(adressen) == 0 {
		adressen = []string{""}
	}

	for _, adresse := range adressen {
		if fehler := blattPruefen(blatt, adresse, jetzt); fehler != nil {
			return tls.Certificate{}, Zeugnisbefund{}, fehler
		}
	}

	if fehler := ketteBilden(dateien, blatt, zwischen, jetzt); fehler != nil {
		return tls.Certificate{}, Zeugnisbefund{}, fehler
	}

	paar.Leaf = blatt

	return paar, befundVon(blatt), nil
}

// zertifikateLesen liest das Blatt und die Zwischenstellen — aus dem
// Zertifikat und aus der Kette. Das PEM, das `tls` braucht, trägt beide.
func zertifikateLesen(
	dateien Zeugnisdateien,
) ([]byte, *x509.Certificate, *x509.CertPool, error) {
	roh, fehler := dateiLesen(dateien.Zertifikat)
	if fehler != nil {
		return nil, nil, nil, fehler
	}

	zertifikate, fehler := pemZertifikate(roh, dateien.Zertifikat)
	if fehler != nil {
		return nil, nil, nil, fehler
	}

	zwischen := x509.NewCertPool()
	for _, zertifikat := range zertifikate[1:] {
		zwischen.AddCert(zertifikat)
	}

	if dateien.Kette != "" {
		kette, lesefehler := dateiLesen(dateien.Kette)
		if lesefehler != nil {
			return nil, nil, nil, lesefehler
		}

		weitere, kettenfehler := pemZertifikate(kette, dateien.Kette)
		if kettenfehler != nil {
			return nil, nil, nil, kettenfehler
		}

		for _, zertifikat := range weitere {
			zwischen.AddCert(zertifikat)
		}

		roh = append(append(roh, '\n'), kette...)
	}

	return roh, zertifikate[0], zwischen, nil
}

// paarBilden prüft die Rechte des Schlüssels und ob er zum Blatt passt.
func paarBilden(dateien Zeugnisdateien, zertifikatPEM []byte) (tls.Certificate, error) {
	// **Nur der Eigentümer liest ihn.** Ein Schlüssel, den jeder auf diesem
	// Rechner lesen kann, weist den Dienst gegenüber niemandem mehr aus.
	info, fehler := os.Stat(dateien.Schluessel)
	if fehler != nil {
		return tls.Certificate{}, abweisen(fehler, "%s lässt sich nicht lesen", dateien.Schluessel)
	}

	if rechte := info.Mode().Perm(); rechte&0o077 != 0 {
		return tls.Certificate{}, abweisen(nil,
			"der Schlüssel %s ist für andere lesbar (%04o) — erlaubt ist 0600",
			dateien.Schluessel, rechte)
	}

	schluessel, fehler := dateiLesen(dateien.Schluessel)
	if fehler != nil {
		return tls.Certificate{}, fehler
	}

	paar, fehler := tls.X509KeyPair(zertifikatPEM, schluessel)
	if fehler != nil {
		return tls.Certificate{}, abweisen(fehler,
			"der Schlüssel %s passt nicht zum Zertifikat %s",
			dateien.Schluessel, dateien.Zertifikat)
	}

	return paar, nil
}

// blattPruefen hält Gültigkeit und Namen.
func blattPruefen(blatt *x509.Certificate, adresse string, jetzt time.Time) error {
	if jetzt.Before(blatt.NotBefore) {
		return abweisen(nil, "das Zertifikat ist noch nicht gültig — ab %s",
			blatt.NotBefore.UTC().Format(time.RFC3339))
	}

	if jetzt.After(blatt.NotAfter) {
		return abweisen(nil, "das Zertifikat ist abgelaufen — seit %s",
			blatt.NotAfter.UTC().Format(time.RFC3339))
	}

	if len(blatt.DNSNames)+len(blatt.IPAddresses) == 0 {
		return abweisen(nil, "das Zertifikat nennt keinen Namen und keine Adresse (SAN)")
	}

	// **Eine bestimmte Adresse muss darin stehen**; eine Bindung an jede
	// Schnittstelle kann keine verlangen — der Klient wählt, über welche er
	// kommt.
	if gebunden, fehler := netip.ParseAddr(adresse); fehler == nil && !gebunden.IsUnspecified() {
		if fehler := blatt.VerifyHostname(adresse); fehler != nil {
			return abweisen(fehler, "das Zertifikat nennt die Bindungsadresse %s nicht", adresse)
		}
	}

	return nil
}

// ketteBilden prüft, dass die Kette zum Serverzweck bis zu einem Anker reicht.
func ketteBilden(
	dateien Zeugnisdateien, blatt *x509.Certificate, zwischen *x509.CertPool,
	jetzt time.Time,
) error {
	var anker *x509.CertPool

	if dateien.Anker != "" {
		roh, fehler := dateiLesen(dateien.Anker)
		if fehler != nil {
			return fehler
		}

		zertifikate, fehler := pemZertifikate(roh, dateien.Anker)
		if fehler != nil {
			return fehler
		}

		anker = x509.NewCertPool()
		for _, zertifikat := range zertifikate {
			anker.AddCert(zertifikat)
		}
	}

	if _, fehler := blatt.Verify(x509.VerifyOptions{
		Roots: anker, Intermediates: zwischen, CurrentTime: jetzt,
		KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}); fehler != nil {
		return abweisen(fehler,
			"die Kette von %s reicht nicht bis zu einem Vertrauensanker", dateien.Zertifikat)
	}

	return nil
}

func befundVon(blatt *x509.Certificate) Zeugnisbefund {
	summe := sha256.Sum256(blatt.Raw)

	teile := make([]string, 0, len(summe))
	for _, wert := range summe {
		teile = append(teile, fmt.Sprintf("%02X", wert))
	}

	namen := append([]string{}, blatt.DNSNames...)
	for _, adresse := range blatt.IPAddresses {
		namen = append(namen, adresse.String())
	}

	return Zeugnisbefund{
		Fingerabdruck: strings.Join(teile, ":"),
		GueltigBis:    blatt.NotAfter,
		Aussteller:    blatt.Issuer.CommonName,
		Namen:         namen,
	}
}

// Zeugnisgrenze ist, wie groß eine Datei der PKI höchstens sein darf.
//
// **Ein Zeugnis ist ein paar Kilobyte**, eine lange Kette ein paar Dutzend.
// Ein Pfad, der auf etwas anderes zeigt, soll den Dienst nicht mit einem
// Gigabyte aufhalten.
const Zeugnisgrenze = 1 << 20

// dateiLesen liest eine Datei der PKI — **eine Variable**, damit ein Test
// eine Datei zwischen zwei Lesevorgängen tauschen kann
// (`export_test.go`, Nachlauf über Phase G).
var dateiLesen = dateiVonDerPlatte

func dateiVonDerPlatte(pfad string) ([]byte, error) {
	datei, fehler := os.Open(pfad) //#nosec G304 -- der Pfad ist die Einstellung dieses Rechners
	if fehler != nil {
		return nil, abweisen(fehler, "%s lässt sich nicht lesen", pfad)
	}

	defer func() { _ = datei.Close() }() //nolint:errcheck // gelesen ist gelesen

	roh, fehler := io.ReadAll(io.LimitReader(datei, Zeugnisgrenze+1))
	if fehler != nil {
		return nil, abweisen(fehler, "%s lässt sich nicht lesen", pfad)
	}

	if len(roh) > Zeugnisgrenze {
		return nil, abweisen(nil, "%s ist größer als %d Bytes — ein Zeugnis ist das nicht",
			pfad, Zeugnisgrenze)
	}

	return roh, nil
}

// pemZertifikate liest jedes Zertifikat einer PEM-Datei.
func pemZertifikate(roh []byte, pfad string) ([]*x509.Certificate, error) {
	var zertifikate []*x509.Certificate

	for {
		var block *pem.Block

		block, roh = pem.Decode(roh)
		if block == nil {
			break
		}

		if block.Type != "CERTIFICATE" {
			continue
		}

		zertifikat, fehler := x509.ParseCertificate(block.Bytes)
		if fehler != nil {
			return nil, abweisen(fehler, "%s enthält ein Zertifikat, das sich nicht lesen lässt", pfad)
		}

		zertifikate = append(zertifikate, zertifikat)
	}

	if len(zertifikate) == 0 {
		return nil, abweisen(nil, "%s enthält kein Zertifikat (PEM)", pfad)
	}

	return zertifikate, nil
}

func abweisen(grund error, satz string, werte ...any) error {
	return domain.MitGrund(domain.ErrValidation, grund, satz, werte...)
}
