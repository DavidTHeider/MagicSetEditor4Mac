// **Der Tausch von Code gegen Identität** — Werkstück 4 von Phase 4.12
// (#1351).

package anmeldung

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Die Zuordnung auf `domain` steht an einer Stelle je Gestalt** (§16) — in
// [Lauf.ausOIDC] und [Lauf.deuteNutzer], und nirgends sonst. Was dabei
// herauskommt, ist [domain.Fremdkennung]: ein fachlicher Begriff, kein Detail
// dieses Flusses.

// hoechstens ist, wie viel dieses Paket von einer Antwort liest.
//
// **Eine Grenze, weil die Gegenstelle fremd ist.** Eine Antwort ohne Ende
// füllte den Arbeitsspeicher, und ein Anbieter, dessen Verhalten sich ändert,
// ist genau der Fall, für den §16 „Drift wird erkannt, nicht erlitten" sagt.
const hoechstens = 1 << 20

// Einloesen tauscht den Code und gibt die Identität.
//
// **Der Klient wird eingesetzt und nicht erzeugt** (§16): sonst bräuchte jeder
// Test das Netz. Frist und Abbruch reisen im Kontext.
func (l *Lauf) Einloesen(ctx context.Context, rufer *Rufer, code string) (
	domain.Fremdkennung, error,
) {
	jetons, fehler := l.tausche(ctx, rufer, code)
	if fehler != nil {
		return domain.Fremdkennung{}, fehler
	}

	if l.anbieter.Gestalt == OIDC {
		return l.ausOIDC(jetons.IDToken)
	}

	if jetons.Zugangswort == "" {
		return domain.Fremdkennung{}, fmt.Errorf(
			"%w: %s gab kein Zugangswort zurück", domain.ErrValidation, l.anbieter.Name)
	}

	return l.ausNutzerruf(ctx, rufer, jetons.Zugangswort)
}

// jetonantwort ist die Gestalt, die ein Tauschendpunkt schickt.
//
// **Das fremde DTO** (§16): es trägt die Feldnamen des Anbieters. Was daraus
// wird, entscheiden die Zuordnungen darunter.
type jetonantwort struct {
	Zugangswort string `json:"access_token"`
	IDToken     string `json:"id_token"`
	Fehler      string `json:"error"`
	Erklaerung  string `json:"error_description"`
}

// tausche ruft den Tauschendpunkt.
func (l *Lauf) tausche(ctx context.Context, rufer *Rufer, code string) (
	jetonantwort, error,
) {
	werte := url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {code},
		"redirect_uri":  {l.rueckadresse},
		"client_id":     {l.anbieter.Kennung},
		"code_verifier": {l.pruefer},
	}

	// **Nur, wo es hingehört.** Microsoft weist einen öffentlichen Klienten
	// mit Geheimnis ab; GitHub verlangt eines. Beides steht am Anbieter.
	if l.anbieter.Geheimnis != "" {
		werte.Set("client_secret", l.anbieter.Geheimnis)
	}

	anfrage, baufehler := http.NewRequestWithContext(ctx, http.MethodPost,
		l.anbieter.Tauschadresse, strings.NewReader(werte.Encode()))
	if baufehler != nil {
		return jetonantwort{}, fmt.Errorf("die Tauschanfrage bauen: %w", baufehler)
	}

	anfrage.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// **GitHub gibt sonst `application/x-www-form-urlencoded` zurück** und
	// nicht JSON. Ein Kopf, den drei Anbieter ignorieren und einer braucht,
	// steht deshalb bei allen.
	anfrage.Header.Set("Accept", "application/json")
	anfrage.Header.Set("User-Agent", kenner)

	antwort, ruffehler := rufer.unten.Do(anfrage)
	if ruffehler != nil {
		return jetonantwort{}, fmt.Errorf("%s antwortet nicht: %w",
			l.anbieter.Name, ruffehler)
	}

	defer func() { _ = antwort.Body.Close() }() //nolint:errcheck // gelesen ist gelesen

	roh, lesefehler := io.ReadAll(io.LimitReader(antwort.Body, hoechstens))
	if lesefehler != nil {
		return jetonantwort{}, fmt.Errorf("die Antwort von %s lesen: %w",
			l.anbieter.Name, lesefehler)
	}

	var gelesen jetonantwort

	if err := json.Unmarshal(roh, &gelesen); err != nil {
		return jetonantwort{}, fmt.Errorf(
			"%w: %s antwortete nicht mit JSON (%d)",
			domain.ErrValidation, l.anbieter.Name, antwort.StatusCode)
	}

	// **Der Fehler im Rumpf zählt, nicht nur der Status.** GitHub antwortet
	// auf einen verbrauchten Code mit **200** und `error` im Rumpf.
	if gelesen.Fehler != "" {
		return jetonantwort{}, fmt.Errorf("%w: %s weist den Tausch ab (%s): %s",
			domain.ErrValidation, l.anbieter.Name, gelesen.Fehler, gelesen.Erklaerung)
	}

	if antwort.StatusCode < 200 || antwort.StatusCode >= 300 {
		return jetonantwort{}, fmt.Errorf("%w: %s weist den Tausch ab (%d)",
			domain.ErrValidation, l.anbieter.Name, antwort.StatusCode)
	}

	return gelesen, nil
}

// kenner ist der `User-Agent`, mit dem sich dieses Programm meldet (§16).
const kenner = "MagicSetEditor4Mac (+https://github.com/DavidTHeider/MagicSetEditor4Mac)"
