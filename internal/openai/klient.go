// Package openai liest bei OpenAI — welche Modelle es gibt und was sie in
// fester Form antworten.
//
// **Werkstück A9 von Phase 5.3**
// (#1448).
//
// # Was dieses Paket **nicht** kann
//
// **Es belegt keine Fähigkeit aus einer Liste.** `/v1/models` nennt je Modell
// `id`, `object`, `created` und `owned_by` — und sonst nichts; kein
// `capabilities`, wie es Anthropic führt (A8). Aus dieser Liste entsteht
// deshalb **kein** Katalogeintrag, sondern je Modell ein **Befund**
// ([0100](../../docs/entscheidungen/0100-was-openai-nicht-sagt.md)).
//
// **Es erzeugt keine Bilder.** Die Bildwerkzeuge der Responses-API stehen
// hinter A12: ein Bild, das nirgends abgelegt werden kann, wäre ein Weg ohne
// Ziel (§10).
//
// **Und es legt nichts beim Anbieter ab.** `store` steht ausdrücklich auf
// `false` — die Vorgabe der Schnittstelle ist `true`, und das hieße, dass
// jede Kartenanfrage dort liegen bleibt.
package openai

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
)

// Vorgabeadresse ist die Schnittstelle von OpenAI.
const Vorgabeadresse = "https://api.openai.com"

// Anbieter ist der Name, unter dem diese Modelle im Katalog stehen.
const Anbieter = "openai"

// Die zwei Wege — die ganze Positivliste.
const (
	// Modellweg nennt die Modelle — **ohne** ihre Fähigkeiten.
	Modellweg = "/v1/models"

	// Antwortweg ist die Responses-API.
	Antwortweg = "/v1/responses"
)

// Wege gibt die Positivliste — für Wächter, die sie nachzählen.
//
// **Vier, und keiner davon legt beim Anbieter etwas ab.** Es gibt hier kein
// `/v1/files`, keine Stapel-API und keinen entfernten Werkzeugserver — die
// agentische Kette steht in A14 und ist dort ausdrücklich gesondert.
//
// **Die zwei Bildwege kamen mit A12 dazu** (#1451).
// Das ist eine **erweiternde** Änderung im Sinne von
// [0033](../../docs/entscheidungen/0033-was-ueber-fassungen-hinweg-zugesagt-wird.md):
// kein vorhandener Ruf ändert seine Bedeutung. Was **nicht** dazukam, ist
// `/v1/files` — der Bearbeitungsweg schickt eine Daten-URI und keine
// Dateikennung
// ([0102](../../docs/entscheidungen/0102-wo-ein-erzeugtes-bild-liegt.md) §5).
func Wege() []string {
	return []string{Modellweg, Antwortweg, Bildweg, Bildeditweg}
}

// Abstand ist der Mindestabstand zwischen zwei Anfragen.
//
// # Warum eine Sekunde
//
// **Nachgelesen am 2026-09-12**: die Grenzen stehen je Konto und Tarifstufe
// und werden von der Dokumentation nicht als Zahl genannt — sie stehen in
// den Kontoeinstellungen. Was **jede** Antwort mitbringt, sind die Kopfzeilen
// `x-ratelimit-*`; eine feste Zahl im Quelltext wäre geraten.
//
// Gedrosselt wird deshalb auf eine Anfrage je Sekunde: was dieses Programm
// fragt, ist ein Vorschlag je Bedienhandlung, und eine zu vorsichtige Grenze
// kostet nichts (§16, Regel 4).
const Abstand = time.Second

// hoechstens ist die Größe, ab der eine Antwort abgewiesen wird.
//
// **Vier Megabyte.** Eine Antwort dieser Größe gäbe es nur, wenn etwas
// grundlegend anders läuft als gedacht.
const hoechstens int64 = 4 << 20

// Zugang ist der Schlüssel, den OpenAI verlangt.
//
// **Ein eigener Zugang und kein Abonnement.** Das Issue sagt es ausdrücklich:
// „ChatGPT/Codex-Abonnement nicht als API-Berechtigung oder verfügbares
// Kontingent voraussetzen." Wer ChatGPT bezahlt, hat damit **kein**
// API-Guthaben.
type Zugang struct{ Schluessel string }

// Fehlt nennt, was fehlt — leer, wenn nichts fehlt.
func (z Zugang) Fehlt() string {
	if z.Schluessel == "" {
		return "der API-Schlüssel von OpenAI"
	}

	return ""
}

// Klient spricht mit der Schnittstelle von OpenAI.
type Klient struct {
	adresse string
	klient  *netz.Klient
	zugang  Zugang
	drossel *netz.Drossel

	// letzteGrenzen sind die Grenzen der zuletzt gesehenen Antwort.
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

// kenner ist die Aufschrift jeder Anfrage.
//
// **`Authorization: Bearer …`**, anders als bei Anthropic (`x-api-key`).
func (k *Klient) kenner() netz.Kenner {
	return netz.Kenner{
		Zusatz: map[string]string{
			"authorization": "Bearer " + k.zugang.Schluessel,
			"content-type":  "application/json",
		},
	}
}

// hole ruft einen Weg der Positivliste.
//
// **Der Zugang wird vor dem Ruf geprüft** — eine Anfrage ohne Schlüssel
// erzeugte eine Last beim Dienst und einen 401, den diese Seite selbst kennt.
func (k *Klient) hole(
	ctx context.Context, weg string, rumpf, ziel any,
) error {
	return k.holeGross(ctx, weg, rumpf, ziel, hoechstens)
}

// holeGross ist derselbe Weg mit einer eigenen Größengrenze.
//
// **Weil ein Bild anders wiegt als ein Satz** (A12, #1451). Vier Mebibyte
// reichen für jede Antwort der Responses-API und für keine Bildantwort; eine
// Grenze für beide wäre für das eine zu weit und für das andere zu eng.
func (k *Klient) holeGross(
	ctx context.Context, weg string, rumpf, ziel any, hoechstens int64,
) error {
	if fehlt := k.zugang.Fehlt(); fehlt != "" {
		return fmt.Errorf("%w: es fehlt %s", domain.ErrValidation, fehlt)
	}

	adresse, fehler := url.JoinPath(k.adresse, weg)
	if fehler != nil {
		return domain.MitGrund(domain.ErrValidation, fehler,
			"die Adresse für %s bauen", weg)
	}

	antwort, holfehler := k.frage(ctx, adresse, rumpf, hoechstens)

	// **Die Grenzen zuerst, auch bei einem Fehler.** Gerade ein 429 trägt
	// sie; wer sie erst nach dem Erfolgsweg läse, bekäme sie nie, wenn er
	// sie braucht.
	k.letzteGrenzen = grenzenAus(antwort.Kopf)

	// **Ein 429 wird hier gedeutet und nicht durchgereicht.** `netz` meldet
	// es als [netz.ErrGedrosselt] — und bei OpenAI tragen drei verschiedene
	// Dinge diesen Status, von denen zwei nicht vorübergehen.
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
			"die Antwort von OpenAI auf %s lesen", weg)
	}

	return nil
}

// frage stellt die Anfrage — mit Rumpf als POST, ohne als GET.
func (k *Klient) frage(
	ctx context.Context, adresse string, rumpf any, hoechstens int64,
) (netz.Antwort, error) {
	if rumpf == nil {
		return netz.Hole(ctx, k.klient, k.drossel, k.kenner(), adresse, hoechstens)
	}

	roh, fehler := json.Marshal(rumpf)
	if fehler != nil {
		return netz.Antwort{}, domain.MitGrund(domain.ErrValidation, fehler,
			"die Anfrage an OpenAI bauen")
	}

	return netz.SendeJSON(ctx, k.klient, k.drossel, k.kenner(), adresse, roh, hoechstens)
}
