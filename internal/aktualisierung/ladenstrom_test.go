// **Was ein Strom anrichten kann, und was nicht**
// (#1100).
//
// **Eine eigene Datei**, seit `ladenfehler_test.go` über die Zeilengrenze aus
// §2 ging — und der Schnitt folgt der Frage: dort steht, was ein falscher
// **Inhalt** auslöst, hier, was eine Gegenstelle über den **Strom** versucht.
package aktualisierung_test

import (
	"errors"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/aktualisierung"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Ein Strom, der mittendrin abreißt, lässt nichts zurück.**
//
// Das ist der Fall, für den es die Nachbardatei gibt: 188 MB über eine
// Mobilverbindung, und auf halbem Weg ist die Leitung weg. Was dann im
// Downloadverzeichnis stünde, sähe aus wie ein fertiges Abbild.
func TestEinAbgerissenerStromLaesstNichtsZurueck(t *testing.T) {
	inhalt := abbild()

	signiere := probeschluessel(t)
	summen := summe(inhalt) + "  ./MagicSetEditor4Mac-0.2.0.dmg\n"
	signatur := signiere(summen)

	dienst := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, anfrage *http.Request) {
			if strings.HasSuffix(anfrage.URL.Path, "/signatur") {
				if _, fehler := schreiben.Write([]byte(signatur)); fehler != nil {
					t.Errorf("schreiben: %v", fehler)
				}

				return
			}

			if strings.HasSuffix(anfrage.URL.Path, "/pruefsummen") {
				if _, fehler := schreiben.Write([]byte(summen)); fehler != nil {
					t.Errorf("schreiben: %v", fehler)
				}

				return
			}

			// **Die Länge verspricht mehr, als kommt** — und dann reißt die
			// Verbindung ab. Genau das sieht ein Klient bei einem Abbruch.
			schreiben.Header().Set("Content-Length", strconv.Itoa(len(inhalt)))
			if _, fehler := schreiben.Write(inhalt[:4096]); fehler != nil {
				t.Logf("der Klient ging weg: %v", fehler)
			}

			panic(http.ErrAbortHandler)
		}))
	t.Cleanup(dienst.Close)

	dienst.Config.ErrorLog = log.New(io.Discard, "", 0)

	ziel := t.TempDir()

	_, err := probeklient(t, dienst.URL).Laden(t.Context(),
		aktualisierung.Freigabe{
			Fassung:     "0.2.0",
			Abbild:      beigabe("MagicSetEditor4Mac-0.2.0.dmg", dienst.URL+"/abbild"),
			Pruefsummen: beigabe("pruefsummen.txt", dienst.URL+"/pruefsummen"),
			Signatur:    beigabe("pruefsummen.txt.asc", dienst.URL+"/signatur"),
		}, ziel, nil)

	if err == nil {
		t.Fatal("der abgerissene Strom kam durch")
	}

	eintraege, lesefehler := os.ReadDir(ziel)
	if lesefehler != nil {
		t.Fatalf("lesen: %v", lesefehler)
	}

	if len(eintraege) != 0 {
		t.Errorf("es blieb etwas liegen: %+v", eintraege)
	}
}

// **Prüfsummen, die gar nicht erreichbar sind, werden gemeldet.**
func TestUnerreichbarePruefsummenWerdenGemeldet(t *testing.T) {
	_, err := probeklient(t, "http://127.0.0.1:1").Laden(t.Context(),
		aktualisierung.Freigabe{
			Fassung:     "0.2.0",
			Abbild:      beigabe("MagicSetEditor4Mac-0.2.0.dmg", "http://127.0.0.1:1/abbild"),
			Pruefsummen: beigabe("pruefsummen.txt", "http://127.0.0.1:1/pruefsummen"),
			Signatur:    beigabe("pruefsummen.txt.asc", "http://127.0.0.1:1/pruefsummen"+"/signatur"),
		}, t.TempDir(), nil)

	if err == nil {
		t.Error("unerreichbare Prüfsummen kamen durch")
	}
}

// **Beigaben werden als Inhalt angefragt, nicht als Auskunft.**
//
// Nachgelesen am 2026-08-25 (`https://docs.github.com/en/rest/releases/assets`,
// und nachgemessen im Issue): dieselbe Adresse gibt mit
// `Accept: application/vnd.github+json` die **Beschreibung** der Beigabe und
// erst mit `application/octet-stream` ihren **Inhalt**. Ohne diese Zeile lüde
// diese Anwendung ein Stück JSON, rechnete brav dessen Prüfsumme und meldete
// „stimmt nicht" — ein Fehler, der wie ein Angriff aussähe.
func TestBeigabenWerdenAlsInhaltAngefragt(t *testing.T) {
	inhalt := abbild()

	var angenommen []string

	signiere := probeschluessel(t)
	summen := summe(inhalt) + "  ./MagicSetEditor4Mac-0.2.0.dmg\n"
	signatur := signiere(summen)

	dienst := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, anfrage *http.Request) {
			angenommen = append(angenommen, anfrage.Header.Get("Accept"))

			if strings.HasSuffix(anfrage.URL.Path, "/signatur") {
				if _, fehler := schreiben.Write([]byte(signatur)); fehler != nil {
					t.Errorf("schreiben: %v", fehler)
				}

				return
			}

			if strings.HasSuffix(anfrage.URL.Path, "/pruefsummen") {
				if _, fehler := schreiben.Write([]byte(summen)); fehler != nil {
					t.Errorf("schreiben: %v", fehler)
				}

				return
			}

			schreiben.Header().Set("Content-Length", strconv.Itoa(len(inhalt)))
			if _, fehler := schreiben.Write(inhalt); fehler != nil {
				t.Logf("der Klient ging weg: %v", fehler)
			}
		}))
	t.Cleanup(dienst.Close)

	if _, err := probeklient(t, dienst.URL).Laden(t.Context(),
		aktualisierung.Freigabe{
			Fassung:     "0.2.0",
			Abbild:      beigabe("MagicSetEditor4Mac-0.2.0.dmg", dienst.URL+"/abbild"),
			Pruefsummen: beigabe("pruefsummen.txt", dienst.URL+"/pruefsummen"),
			Signatur:    beigabe("pruefsummen.txt.asc", dienst.URL+"/signatur"),
		}, t.TempDir(), nil); err != nil {
		t.Fatalf("Laden: %v", err)
	}

	for _, angabe := range angenommen {
		if angabe != "application/octet-stream" {
			t.Errorf("eine Beigabe wurde mit %q angefragt", angabe)
		}
	}

	// **Drei**: Prüfsummen, Signatur, Abbild. Alle drei sind Beigaben.
	if len(angenommen) != 3 {
		t.Errorf("es gingen %d Anfragen hinaus", len(angenommen))
	}
}

// **Ein Strom, der länger ist als angekündigt, wird abgebrochen.**
//
// # Der Befund
//
// `io.Copy` las bis zum Dateiende, und die Frist darunter ist eine
// **Stillstands**frist: sie wird vor jedem Lesen erneuert. Eine Gegenstelle,
// die stetig weitersendet, konnte die Nachbardatei wachsen lassen, bis die
// Platte voll ist — die Prüfsumme greift erst danach, und bis dahin ist auch
// das Aufräumen unerreichbar.
//
// Angestrichen vom Gesamtreview vor 0.2.0.
func TestEinZuLangerStromWirdAbgebrochen(t *testing.T) {
	inhalt := abbild()

	signiere := probeschluessel(t)
	summen := summe(inhalt) + "  ./MagicSetEditor4Mac-0.2.0.dmg\n"
	signatur := signiere(summen)

	var geschrieben atomic.Int64

	dienst := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, anfrage *http.Request) {
			switch {
			case strings.HasSuffix(anfrage.URL.Path, "/signatur"):
				if _, fehler := schreiben.Write([]byte(signatur)); fehler != nil {
					t.Errorf("schreiben: %v", fehler)
				}
			case strings.HasSuffix(anfrage.URL.Path, "/pruefsummen"):
				if _, fehler := schreiben.Write([]byte(summen)); fehler != nil {
					t.Errorf("schreiben: %v", fehler)
				}
			default:
				// **Zwanzigmal so viel wie angekündigt** — und gezählt wird,
				// wie viel davon wirklich hinausging. Das ist der Unterschied,
				// den der Deckel macht: ohne ihn liest der Klient alles.
				for range 20 {
					anzahl, fehler := schreiben.Write(inhalt)

					geschrieben.Add(int64(anzahl))

					if fehler != nil {
						return
					}
				}
			}
		}))
	t.Cleanup(dienst.Close)

	ziel := t.TempDir()

	_, err := probeklient(t, dienst.URL).Laden(t.Context(),
		aktualisierung.Freigabe{
			Fassung: "0.2.0",
			Abbild: aktualisierung.Beigabe{
				Name:    "MagicSetEditor4Mac-0.2.0.dmg",
				Groesse: int64(len(inhalt)),
				Adresse: dienst.URL + "/abbild",
			},
			Pruefsummen: beigabe("pruefsummen.txt", dienst.URL+"/pruefsummen"),
			Signatur:    beigabe("pruefsummen.txt.asc", dienst.URL+"/signatur"),
		}, ziel, nil)

	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("der Fehler ist %v — will ErrConflict", err)
	}

	eintraege, lesefehler := os.ReadDir(ziel)
	if lesefehler != nil {
		t.Fatalf("lesen: %v", lesefehler)
	}

	if len(eintraege) != 0 {
		t.Errorf("es blieb etwas liegen: %+v", eintraege)
	}

	// **Und der Klient hat wirklich früh aufgehört.**
	//
	// Ohne diesen Satz prüfte der Test nur die **Meldung** — und die kommt
	// auch, wenn erst alles gelesen und danach verglichen wird. Genau das war
	// der erste Anlauf: die Gegenprobe (den Deckel entfernen) blieb grün.
	//
	// **Vier Vielfache als Spielraum**: die Puffer von TCP nehmen dem Sender
	// noch etwas ab, nachdem der Empfänger aufgehört hat. Zwanzig wären ohne
	// Deckel hinausgegangen.
	if grenze := int64(4 * len(inhalt)); geschrieben.Load() > grenze {
		t.Errorf("die Gegenstelle brachte %d Byte los, erwartet waren höchstens %d",
			geschrieben.Load(), grenze)
	}
}

// **Ohne angekündigte Größe gilt die feste Grenze.**
func TestOhneAngekuendigteGroesseGiltDieFesteGrenze(t *testing.T) {
	t.Parallel()

	if aktualisierung.Deckel(0) != aktualisierung.HoechstensAbbild {
		t.Errorf("ohne Größe gilt %d", aktualisierung.Deckel(0))
	}

	// **Und eine unglaubwürdig große ebenso.** Wer 40 GB ankündigt, bekommt
	// nicht 40 GB Deckel — die Zahl kommt von derselben Gegenstelle.
	if aktualisierung.Deckel(40<<30) != aktualisierung.HoechstensAbbild {
		t.Errorf("bei 40 GB gilt %d", aktualisierung.Deckel(40<<30))
	}

	if aktualisierung.Deckel(4711) != 4711 {
		t.Errorf("bei 4711 gilt %d", aktualisierung.Deckel(4711))
	}
}
