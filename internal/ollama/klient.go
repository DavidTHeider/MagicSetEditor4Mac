// Package ollama liest bei einer lokalen Ollama-Laufzeit — installierte
// Modelle, ihre Fähigkeiten und ihre Antworten.
//
// **Werkstück A7 von Phase 5.3**
// (#1446).
//
// # Was dieses Paket **nicht** kann
//
// **Es installiert nichts.** Die Positivliste [Wege] führt fünf Adressen:
// `/api/version`, `/api/tags`, `/api/show`, `/api/chat` und `/api/embed`.
// Es gibt kein
// `/api/pull`, kein `/api/create`, kein `/api/delete` — das Issue verlangt es
// wörtlich: „keine automatische Installation/Downloads/Modellpulls", „kein
// impliziter Modellpull".
//
// # Und was es nicht beweisen kann
//
// **Dass auf 11434 wirklich Ollama liegt.** Ein beliebiger Dienst kann dort
// antworten und weiterleiten; das Issue sagt es selbst: „Für beliebige fremde
// Loopback-Proxies **keine absolute Netzwerkgarantie** behaupten." Was dieses
// Paket hat, ist eine Auskunft von dem, was antwortet — und die steht mit
// ihrer Herkunft da, nicht als Gewissheit
// ([0098](../../docs/entscheidungen/0098-was-ollama-belegt-und-was-nicht.md)).
package ollama

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
)

// Vorgabeadresse ist, wo Ollama auf diesem Rechner lauscht.
//
// **Nachgelesen am 2026-09-11**: „Ollama binds 127.0.0.1 port 11434 by
// default." Die Schleifenadresse ist die Vorgabe und keine Zusage — wer
// `OLLAMA_HOST` setzt, bindet woandershin.
const Vorgabeadresse = "http://127.0.0.1:11434"

// Anbieter ist der Name, unter dem diese Laufzeit im Katalog steht.
const Anbieter = "ollama"

// Die vier Wege — die ganze Positivliste.
const (
	// Fassungsweg nennt die Fassung der Laufzeit.
	Fassungsweg = "/api/version"

	// Modellweg nennt die **installierten** Modelle.
	Modellweg = "/api/tags"

	// Schauweg nennt Fähigkeiten, Lizenz und Angaben eines Modells.
	Schauweg = "/api/show"

	// Redeweg ist die Unterhaltung — Text und strukturierte Ausgabe.
	Redeweg = "/api/chat"
)

// Wege gibt die Positivliste — für Wächter, die sie nachzählen.
func Wege() []string {
	return []string{Fassungsweg, Modellweg, Schauweg, Redeweg, Einbettungsweg}
}

// Cloudsuffix kennzeichnet ein Modell, das Ollama auslagert.
//
// **Nachgelesen am 2026-09-11**: „cloud models are automatically offloaded to
// Ollama's cloud service", und sie heißen `…-cloud`. **Die Adresse
// unterscheidet sie nicht** — sie liegen auf demselben Port wie die lokalen.
const Cloudsuffix = "-cloud"

// hoechstens ist die Größe, ab der eine Antwort abgewiesen wird.
const hoechstens int64 = 16 << 20

// Klient spricht mit einer Ollama-Laufzeit.
type Klient struct {
	adresse string
	klient  *netz.Klient
	jetzt   func() time.Time
}

// NeuerKlient baut ihn.
//
// **Adresse und Netzklient kommen von außen** (§16): sonst braucht jeder Lauf
// eine laufende Ollama-Installation.
func NeuerKlient(adresse string, klient *netz.Klient) *Klient {
	if adresse == "" {
		adresse = Vorgabeadresse
	}

	return &Klient{adresse: adresse, klient: klient, jetzt: time.Now}
}

// MitUhr setzt die Uhr — für Läufe, die ein festes Prüfdatum brauchen.
func (k *Klient) MitUhr(jetzt func() time.Time) *Klient {
	k.jetzt = jetzt

	return k
}

// Fassung nennt die Fassung der Laufzeit.
//
// **Sie gehört zu jedem Beleg**: dieselbe Datei tut unter zwei Fassungen nicht
// dasselbe, und eine Fähigkeit ist immer die einer **Kombination**.
func (k *Klient) Fassung(ctx context.Context) (string, error) {
	var antwort struct {
		Fassung string `json:"version"`
	}

	if fehler := k.hole(ctx, Fassungsweg, nil, &antwort); fehler != nil {
		return "", fehler
	}

	return antwort.Fassung, nil
}

// hole ruft einen Weg der Positivliste.
//
// **Ein Rumpf macht daraus ein POST** — `/api/show` und `/api/chat` verlangen
// einen, `/api/tags` und `/api/version` nicht. Die Methode steht deshalb hier
// und nicht bei jedem Aufrufer.
func (k *Klient) hole(ctx context.Context, weg string, rumpf, ziel any) error {
	adresse, fehler := url.JoinPath(k.adresse, weg)
	if fehler != nil {
		return domain.MitGrund(domain.ErrValidation, fehler,
			"die Adresse für %s bauen", weg)
	}

	antwort, holfehler := k.frage(ctx, adresse, rumpf)
	if holfehler != nil {
		return holfehler
	}

	if antwort.Status != netz.StatusGut {
		return fmt.Errorf("%w: Ollama antwortet auf %s mit %d%s",
			domain.ErrValidation, weg, antwort.Status, grundAus(antwort.Rumpf))
	}

	if fehler := json.Unmarshal(antwort.Rumpf, ziel); fehler != nil {
		return domain.MitGrund(domain.ErrValidation, fehler,
			"die Antwort von Ollama auf %s lesen", weg)
	}

	return nil
}

// grundlaenge ist, wie viele Zeichen vom Grund der Laufzeit in eine Meldung
// kommen — genug für „"all-minilm:22m" does not support chat", zu wenig für
// einen Rumpf, den niemand gelesen hat.
const grundlaenge = 160

// grundAus liest das Feld `error` eines Fehlerrumpfs — **nur dieses**, auf
// einer Zeile und gekürzt. Ohne es kommt nichts dazu.
//
// **Gefunden in AI-02** (#1743): `ki-modell-pruefen` an einem
// Einbettungsmodell endete mit „mit 400", und Ollama hatte dazugeschrieben,
// dass es nicht reden kann.
func grundAus(rumpf []byte) string {
	var fehlerrumpf struct {
		Grund string `json:"error"`
	}

	if json.Unmarshal(rumpf, &fehlerrumpf) != nil {
		return ""
	}

	grund := strings.Join(strings.Fields(fehlerrumpf.Grund), " ")
	if grund == "" {
		return ""
	}

	if zeichen := []rune(grund); len(zeichen) > grundlaenge {
		grund = string(zeichen[:grundlaenge]) + "…"
	}

	return ": " + grund
}

// frage stellt die Anfrage — mit Rumpf als POST, ohne als GET.
func (k *Klient) frage(ctx context.Context, adresse string, rumpf any) (netz.Antwort, error) {
	if rumpf == nil {
		return netz.Hole(ctx, k.klient, nil, k.kenner(), adresse, hoechstens)
	}

	roh, fehler := json.Marshal(rumpf)
	if fehler != nil {
		return netz.Antwort{}, domain.MitGrund(domain.ErrValidation, fehler,
			"die Anfrage an Ollama bauen")
	}

	return netz.SendeJSON(ctx, k.klient, nil, k.kenner(), adresse, roh, hoechstens)
}

// kenner ist, womit sich dieses Programm vorstellt (§16, Satz 2).
//
// **Ohne Marke**: Ollama auf der Schleifenadresse kennt keine Anmeldung, und
// eine leere `Authorization` wäre etwas anderes als keine.
func (k *Klient) kenner() netz.Kenner {
	return netz.Kenner{
		Name:    "MagicSetEditor4Mac",
		Annahme: "application/json",
	}
}
