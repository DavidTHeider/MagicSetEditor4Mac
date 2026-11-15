// Package xai liest bei xAI — welche Modelle es gibt, was sie **belegt**
// können, und ihre Antworten in fester Form.
//
// **Werkstück A10 von Phase 5.3**
// (#1449).
//
// # Warum dieses Paket nicht von `internal/openai` abschreibt
//
// **Das Issue verlangt es** — „ohne Annahmen aus OpenAI-API-Ähnlichkeiten" —,
// und das Nachlesen am 2026-09-12 hat gezeigt, wie berechtigt das ist:
//
//   - **Die Modellliste heißt anders und sagt mehr**: `/v1/language-models`
//     statt `/v1/models`, und je Modell `input_modalities`,
//     `output_modalities`, `aliases`, `fingerprint`, `version` und Preise.
//   - **Der Ausgang kennt ein Wort, das es sonst nicht gibt**: `end_turn`
//     neben `stop` und `length`.
//   - **Die Antwort nennt ihren Preis**: `usage.cost_in_usd_ticks`, und ein
//     Tick ist ein Zehnmilliardstel Dollar. Das tut sonst keiner.
//   - **Und derselbe Anbieter hat zwei Bildformen**: `/v1/responses` nimmt
//     `{"type":"input_image","image_url":"…"}`, `/v1/chat/completions`
//     dagegen `{"type":"image_url","image_url":{"url":"…"}}`. Wer eine davon
//     rät, rät falsch.
//
// # Was dieses Paket **nicht** kann
//
// **Es erzeugt keine Bilder.** `/v1/image-generation-models` gibt es, und
// wohin ein Bild danach gehört, entscheidet A12
// (#1451).
//
// **Und es legt nichts beim Anbieter ab.** `/v1/responses` hält Antworten
// **30 Tage** — deshalb geht dieses Paket über `/v1/chat/completions`
// ([0101](../../docs/entscheidungen/0101-was-xai-belegt.md)).
package xai

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
)

// Vorgabeadresse ist die Schnittstelle von xAI.
const Vorgabeadresse = "https://api.x.ai"

// Anbieter ist der Name, unter dem diese Modelle im Katalog stehen.
const Anbieter = "xai"

// Die zwei Wege — die ganze Positivliste.
const (
	// Modellweg nennt die Modelle **und ihre Modalitäten**.
	//
	// **`language-models` und nicht `models`**: der zweite gibt es auch, und
	// er nennt die Modalitäten nicht.
	Modellweg = "/v1/language-models"

	// Unterhaltungsweg ist die Unterhaltung.
	Unterhaltungsweg = "/v1/chat/completions"
)

// Wege gibt die Positivliste — für Wächter, die sie nachzählen.
//
// **Zwei, und keiner davon schreibt bei xAI etwas.** Es gibt hier kein
// `/v1/responses` (das hielte Antworten 30 Tage), keine Bilderzeugung (A12),
// keine Stapel-API (A13) und keinen entfernten Werkzeugserver.
func Wege() []string { return []string{Modellweg, Unterhaltungsweg} }

// Abstand ist der Mindestabstand zwischen zwei Anfragen.
//
// # Warum eine Sekunde, wo keine Zahl dasteht
//
// **Nachgelesen am 2026-09-12**: die Dokumentation nennt **keine** Zahl und
// **keine** Kopfzeilen über das Kontingent — sie sagt nur, dass ein 429
// kommt, wenn man zu schnell fragt, und rät zu wachsendem Warten.
//
// **Eine geratene Grenze wäre schlimmer als eine vorsichtige** (§16, Regel 4).
const Abstand = time.Second

// hoechstens ist die Größe, ab der eine Antwort abgewiesen wird.
const hoechstens int64 = 4 << 20

// Zugang ist der Schlüssel, den xAI verlangt.
type Zugang struct{ Schluessel string }

// Fehlt nennt, was fehlt — leer, wenn nichts fehlt.
func (z Zugang) Fehlt() string {
	if z.Schluessel == "" {
		return "der API-Schlüssel von xAI"
	}

	return ""
}

// Klient spricht mit der Schnittstelle von xAI.
type Klient struct {
	adresse string
	klient  *netz.Klient
	zugang  Zugang
	drossel *netz.Drossel
}

// NeuerKlient baut ihn.
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

// kenner ist die Aufschrift jeder Anfrage.
func (k *Klient) kenner() netz.Kenner {
	return netz.Kenner{
		Zusatz: map[string]string{
			"authorization": "Bearer " + k.zugang.Schluessel,
			"content-type":  "application/json",
		},
	}
}

// hole ruft einen Weg der Positivliste.
func (k *Klient) hole(ctx context.Context, weg string, rumpf, ziel any) error {
	if fehlt := k.zugang.Fehlt(); fehlt != "" {
		return fmt.Errorf("%w: es fehlt %s", domain.ErrValidation, fehlt)
	}

	adresse, fehler := url.JoinPath(k.adresse, weg)
	if fehler != nil {
		return domain.MitGrund(domain.ErrValidation, fehler,
			"die Adresse für %s bauen", weg)
	}

	antwort, holfehler := k.frage(ctx, adresse, rumpf)

	// **Ein 429 wird hier gedeutet und nicht durchgereicht** — dieselbe
	// Überlegung wie bei den anderen beiden Anbietern.
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
			"die Antwort von xAI auf %s lesen", weg)
	}

	return nil
}

// frage stellt die Anfrage — mit Rumpf als POST, ohne als GET.
func (k *Klient) frage(ctx context.Context, adresse string, rumpf any) (netz.Antwort, error) {
	if rumpf == nil {
		return netz.Hole(ctx, k.klient, k.drossel, k.kenner(), adresse, hoechstens)
	}

	roh, fehler := json.Marshal(rumpf)
	if fehler != nil {
		return netz.Antwort{}, domain.MitGrund(domain.ErrValidation, fehler,
			"die Anfrage an xAI bauen")
	}

	return netz.SendeJSON(ctx, k.klient, k.drossel, k.kenner(), adresse, roh, hoechstens)
}
