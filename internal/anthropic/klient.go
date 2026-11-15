// Package anthropic liest bei Anthropic — welche Modelle es gibt, was sie
// **belegt** können, und ihre Antworten in fester Form.
//
// **Werkstück A8 von Phase 5.3**
// (#1447).
//
// # Was dieses Paket **nicht** kann
//
// **Es rechnet nichts auf diesem Rechner.** Jedes Modell hier ist
// [domain.KIOrtCloud], und unter »nur lokal« wird gar nicht gefragt — anders
// als bei `internal/ollama`, wo die Adresse entscheidet. Hier gibt es nichts
// zu entscheiden: die Messages-API liegt bei Anthropic.
//
// **Und es erfindet keine Herkunft.** Anthropic nennt weder Artefakt noch
// Revision noch Prüfsumme; ein Modellname ist kein Digest. Die Kette bleibt
// leer, und das heißt „unbekannt"
// ([0099](../../docs/entscheidungen/0099-was-anthropic-belegt.md)).
package anthropic

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
)

// Vorgabeadresse ist die Schnittstelle von Anthropic.
const Vorgabeadresse = "https://api.anthropic.com"

// Anbieter ist der Name, unter dem diese Modelle im Katalog stehen.
const Anbieter = "anthropic"

// Fassung ist die Fassung der Schnittstelle.
//
// **Nachgelesen am 2026-09-11**: `anthropic-version` ist **Pflicht**, und
// `2023-06-01` ist die Fassung, mit der jede Seite der Dokumentation ruft.
// Ohne sie antwortet der Dienst mit einem 400, und die Meldung spräche von
// der Anfrage statt von der fehlenden Kopfzeile.
const Fassung = "2023-06-01"

// Die zwei Wege — die ganze Positivliste.
const (
	// Modellweg nennt die Modelle **und ihre Fähigkeiten**.
	Modellweg = "/v1/models"

	// Nachrichtenweg ist die Unterhaltung.
	Nachrichtenweg = "/v1/messages"
)

// Wege gibt die Positivliste — für Wächter, die sie nachzählen.
//
// **Zwei, und keiner davon schreibt bei Anthropic etwas.** Es gibt hier kein
// Konto, das dieses Programm anlegt, keine Datei, die es ablegt, und keinen
// Auftrag, den es einreiht (die Batch-API steht in A13, nicht hier).
func Wege() []string { return []string{Modellweg, Nachrichtenweg} }

// Abstand ist der Mindestabstand zwischen zwei Anfragen.
//
// # Warum eine Sekunde, wo 1.000 je Minute erlaubt wären
//
// **Nachgelesen am 2026-09-12** unter
// <https://platform.claude.com/docs/en/api/rate-limits>: die Anfragegrenze
// liegt in der untersten Tarifstufe bei **1.000 je Minute** je Modellklasse
// — also knapp 17 je Sekunde. Die Dokumentation warnt im selben Absatz
// davor, sie für ein Minutenbudget zu halten: „a rate of 60 requests per
// minute might be enforced as 1 request per second".
//
// Gedrosselt wird deshalb auf **eine je Sekunde**. Was dieses Programm
// fragt, ist ein Vorschlag je Bedienhandlung; zwischen 1/s und 17/s liegt
// hier keine spürbare Wartezeit, und eine zu vorsichtige Grenze kostet
// nichts (§16, Regel 4).
const Abstand = time.Second

// hoechstens ist die Größe, ab der eine Antwort abgewiesen wird.
//
// **Vier Megabyte.** Die Anfragegrenze des Dienstes liegt bei 32 MB
// (nachgelesen am 2026-09-11); eine **Antwort** dieser Größe gäbe es nur,
// wenn etwas grundlegend anders läuft als gedacht.
const hoechstens int64 = 4 << 20

// Zugang ist der Schlüssel, den Anthropic verlangt.
//
// **Nur einer.** Anders als Scrydex, das zwei Kopfzeilen verlangt, nimmt
// Anthropic `x-api-key` — und **keine** `Authorization` daneben.
type Zugang struct{ Schluessel string }

// Fehlt nennt, was fehlt — leer, wenn nichts fehlt.
//
// **Als Satz und nicht als `bool`**: „der Zugang fehlt" hilft nicht weiter,
// „der API-Schlüssel fehlt" schon.
func (z Zugang) Fehlt() string {
	if z.Schluessel == "" {
		return "der API-Schlüssel von Anthropic"
	}

	return ""
}

// Klient spricht mit der Schnittstelle von Anthropic.
type Klient struct {
	adresse string
	klient  *netz.Klient
	zugang  Zugang

	// drossel hält den Mindestabstand ein — nil heißt: kein Abstand.
	//
	// **Sie kommt von außen und wird nicht hier gebaut**, weil sie zur
	// **Gegenstelle** gehört und nicht zur Verbindung: zwei Nähte auf
	// dieselbe Adresse teilen sich eine (`internal/app/scryfalldrossel.go`).
	drossel *netz.Drossel

	// letzteGrenzen sind die Grenzen der zuletzt gesehenen Antwort.
	//
	// **Am Klienten und nicht an jedem Rückgabewert**: sie gehören zur
	// Verbindung, nicht zur Frage — und der Verbraucher in 5.2 will sie
	// nach **jedem** Ruf, auch nach einem gescheiterten.
	letzteGrenzen Grenzen
}

// NeuerKlient baut ihn.
//
// **Adresse, Netzklient und Zugang kommen von außen** (§16): sonst braucht
// jeder Lauf einen echten Schlüssel und das Netz.
func NeuerKlient(adresse string, klient *netz.Klient, zugang Zugang) *Klient {
	if adresse == "" {
		adresse = Vorgabeadresse
	}

	return &Klient{adresse: adresse, klient: klient, zugang: zugang}
}

// MitDrossel setzt die Drossel.
func (k *Klient) MitDrossel(drossel *netz.Drossel) *Klient {
	k.drossel = drossel

	return k
}

// Grenzen gibt die Grenzen der zuletzt gesehenen Antwort.
func (k *Klient) Grenzen() Grenzen { return k.letzteGrenzen }

// hole ruft einen Weg der Positivliste.
//
// **Der Zugang wird vor dem Ruf geprüft** — eine Anfrage ohne Schlüssel
// erzeugte eine Last beim Dienst und einen 401, den diese Seite selbst kennt.
func (k *Klient) hole(
	ctx context.Context, weg string, frage url.Values, rumpf, ziel any,
) error {
	if fehlt := k.zugang.Fehlt(); fehlt != "" {
		return fmt.Errorf("%w: es fehlt %s", domain.ErrValidation, fehlt)
	}

	adresse, fehler := adresseFuer(k.adresse, weg, frage)
	if fehler != nil {
		return domain.MitGrund(domain.ErrValidation, fehler,
			"die Adresse für %s bauen", weg)
	}

	antwort, holfehler := k.frage(ctx, adresse, rumpf)

	// **Die Grenzen zuerst, auch bei einem Fehler.** Gerade ein 429 trägt
	// sie; wer sie erst nach dem Erfolgsweg läse, bekäme sie nie, wenn er
	// sie braucht.
	k.letzteGrenzen = grenzenAus(antwort.Kopf)

	// **Ein 429 wird hier gedeutet und nicht durchgereicht.**
	//
	// `netz` meldet es als [netz.ErrGedrosselt] — und zwei verschiedene
	// Dinge tragen bei Anthropic diesen Status: eine Drossel, die vorübergeht,
	// und die Sperre am Monatsdeckel, die es nicht tut. Der Unterschied steht
	// im Rumpf, und **den bringt `netz` seit A8 mit**.
	if antwort.Status == netz.StatusGedrosselt {
		return fehlerAus(antwort, weg)
	}

	if holfehler != nil {
		return holfehler
	}

	if antwort.Status != netz.StatusGut {
		return fehlerAus(antwort, weg)
	}

	if fehler := json.Unmarshal(antwort.Rumpf, ziel); fehler != nil {
		return domain.MitGrund(domain.ErrValidation, fehler,
			"die Antwort von Anthropic auf %s lesen", weg)
	}

	return nil
}

// adresseFuer baut die volle Adresse.
//
// **Der Fragenteil wird angehängt und nicht in den Weg geschrieben.**
// `url.JoinPath` behandelt sein Argument als **Wegteil** und schreibt ein
// `?` als `%3F` — die Anfrage ginge dann an einen Weg, den es nicht gibt.
// Gefunden am 2026-09-11 an einem Lauf, der null Modelle bekam.
func adresseFuer(grund, weg string, frage url.Values) (string, error) {
	voll, fehler := url.JoinPath(grund, weg)
	if fehler != nil {
		return "", fehler
	}

	if len(frage) == 0 {
		return voll, nil
	}

	return voll + "?" + frage.Encode(), nil
}

// frage stellt die Anfrage — mit Rumpf als POST, ohne als GET.
func (k *Klient) frage(ctx context.Context, adresse string, rumpf any) (netz.Antwort, error) {
	if rumpf == nil {
		return netz.Hole(ctx, k.klient, k.drossel, k.kenner(), adresse, hoechstens)
	}

	roh, fehler := json.Marshal(rumpf)
	if fehler != nil {
		return netz.Antwort{}, domain.MitGrund(domain.ErrValidation, fehler,
			"die Anfrage an Anthropic bauen")
	}

	return netz.SendeJSON(ctx, k.klient, k.drossel, k.kenner(), adresse, roh, hoechstens)
}

// kenner ist, womit sich dieses Programm vorstellt (§16, Satz 2).
//
// **Ohne Marke, mit eigenen Kopfzeilen**: Anthropic nimmt `x-api-key` und
// nicht `Authorization: Bearer`. Eine leere `Authorization` wäre etwas
// anderes als keine.
func (k *Klient) kenner() netz.Kenner {
	return netz.Kenner{
		Name:    "MagicSetEditor4Mac",
		Annahme: "application/json",
		Zusatz: map[string]string{
			"x-api-key":         k.zugang.Schluessel,
			"anthropic-version": Fassung,
		},
	}
}
