// Package applefm spricht mit Apples Gerätemodell über `fm serve` — AI-05
// (#2042),
// Vermerk [0166](../../docs/entscheidungen/0166-was-fm-serve-belegt.md).
//
// # Ein Blatt nach §16
//
// **Drei Schichten**: der Transport ist ein Unix-Socket aus `internal/netz`,
// das fremde DTO steht in diesem Paket ([antwortDTO]), und die Zuordnung auf
// das, was der Aufrufer braucht, auch — an einer Stelle. `domain` kennt es für
// die Fehlerarten und sonst nichts.
//
// # Was es nicht tut
//
// **Den Server starten**: das tut [Dienst] daneben. Und **keine Lizenz
// annehmen, nie `sudo`** — siehe [Erkennung].
package applefm

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
)

// Anbieter ist der Name, unter dem das Modell im Katalog steht.
const Anbieter = "apple"

// Modell ist das eine Modell, das `fm serve` führt.
const Modell = "system"

// Die Wege — gemessen am 2026-10-03 (Vermerk 0166).
const (
	Gesundheitsweg   = "/health"
	Modellweg        = "/v1/models"
	Unterhaltungsweg = "/v1/chat/completions"
)

// Socketadresse ist die Adresse, unter der der Socketklient fragt — der Name
// zählt nicht, der Socket entscheidet.
const Socketadresse = "http://fm"

// hoechstens ist die größte Antwort, die gelesen wird.
const hoechstens int64 = 1 << 20

// Klient fragt `fm serve`.
type Klient struct {
	adresse string
	klient  *netz.Klient
}

// NeuerKlient baut einen Klienten — über einem Socket mit [Socketadresse].
func NeuerKlient(klient *netz.Klient, adresse string) *Klient {
	return &Klient{adresse: adresse, klient: klient}
}

// Gesund fragt `/health` und verlangt das Modell `system` als verfügbar.
func (k *Klient) Gesund(ctx context.Context) error {
	var stand struct {
		Models []struct {
			Name      string `json:"name"`
			Available bool   `json:"available"`
		} `json:"models"`
	}

	if fehler := k.hole(ctx, Gesundheitsweg, nil, &stand); fehler != nil {
		return fehler
	}

	for _, modell := range stand.Models {
		if modell.Name == Modell && modell.Available {
			return nil
		}
	}

	return fmt.Errorf("%w: das Gerätemodell ist nicht verfügbar", domain.ErrValidation)
}

// Modelle nennt die Modelle, die `fm serve` führt.
func (k *Klient) Modelle(ctx context.Context) ([]string, error) {
	var liste struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}

	if fehler := k.hole(ctx, Modellweg, nil, &liste); fehler != nil {
		return nil, fehler
	}

	namen := make([]string, 0, len(liste.Data))
	for _, modell := range liste.Data {
		namen = append(namen, modell.ID)
	}

	return namen, nil
}

// hole stellt eine Anfrage und liest eine gute Antwort in `ziel`.
func (k *Klient) hole(ctx context.Context, weg string, rumpf, ziel any) error {
	antwort, fehler := k.frage(ctx, weg, rumpf)
	if fehler != nil {
		return fehler
	}

	if antwort.Status != netz.StatusGut {
		return fehlerAus(antwort, weg)
	}

	if fehler := json.Unmarshal(antwort.Rumpf, ziel); fehler != nil {
		return domain.MitGrund(domain.ErrValidation, fehler,
			"die Antwort des Gerätemodells auf %s lesen", weg)
	}

	return nil
}

func (k *Klient) frage(ctx context.Context, weg string, rumpf any) (netz.Antwort, error) {
	adresse, fehler := url.JoinPath(k.adresse, weg)
	if fehler != nil {
		return netz.Antwort{}, domain.MitGrund(domain.ErrValidation, fehler,
			"die Adresse für %s bauen", weg)
	}

	kenner := netz.Kenner{Zusatz: map[string]string{"content-type": "application/json"}}

	if rumpf == nil {
		return netz.Hole(ctx, k.klient, nil, kenner, adresse, hoechstens)
	}

	roh, fehler := json.Marshal(rumpf)
	if fehler != nil {
		return netz.Antwort{}, domain.MitGrund(domain.ErrValidation, fehler,
			"die Anfrage an das Gerätemodell bauen")
	}

	return netz.SendeJSON(ctx, k.klient, nil, kenner, adresse, roh, hoechstens)
}

// fehlerDTO ist die Gestalt eines Fehlers — `{"error":{"message","code","type"}}`.
type fehlerDTO struct {
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error"`
}

// meldungAus liest die Meldung eines Fehlers; leer, wenn der Rumpf keiner ist.
func meldungAus(rumpf []byte) string {
	var dto fehlerDTO
	if json.Unmarshal(rumpf, &dto) != nil {
		return ""
	}

	return strings.TrimSpace(dto.Error.Message)
}

// fehlerAus macht aus einer schlechten Antwort einen Fehler mit Grund.
func fehlerAus(antwort netz.Antwort, weg string) error {
	meldung := meldungAus(antwort.Rumpf)
	if meldung == "" {
		meldung = fmt.Sprintf("Status %d", antwort.Status)
	}

	return fmt.Errorf("%w: das Gerätemodell antwortete auf %s mit %s",
		domain.ErrValidation, weg, meldung)
}
