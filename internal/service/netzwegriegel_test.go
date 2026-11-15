package service_test

// **Wohin gesendet wird, setzt nur, wer senden darf** — SEC-P4
// (#2000).
//
// # Warum ein Riegel an den Einstellungen
//
// **Die Adresse einer KI, das Gateway eines Anbieters und der Proxy
// bestimmen, wohin diese Anwendung sendet.** Wer sie setzt, wählt das Ziel
// jedes späteren Rufs — auch eines lesenden, den ein Token gehen darf. Ein
// Token reicht nie bis `externSenden` (0078); bis #2000 reichte es trotzdem
// bis zu der Einstellung, die entscheidet, wohin gesendet wird.
//
// **Am Schreiben und nicht am Wählen**, weil die Adresse im eigenen Netz
// liegen darf: ein Ollama auf `gpu.intern` ist ein beschriebener Weg
// (`docs/NETZEIGENSCHAFTEN.md`). Gesperrt wird, wer sie setzt — nicht, wohin
// sie zeigt.

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// netzwege sind die Schlüssel samt einem Wert, den die Domäne annimmt.
func netzwege() map[string]string {
	return map[string]string{
		domain.KIAdresse: "http://127.0.0.1:11434",
		domain.DienstAdresse(domain.DienstOpenAI): "https://gateway.example",
		domain.NetzProxyHTTP:                      "http://proxy.example:3128",
		domain.NetzProxyHTTPS:                     "http://proxy.example:3128",
		domain.NetzProxyAus:                       "intern.example",
	}
}

func TestWohinGesendetWirdSetztNurWerSendenDarf(t *testing.T) {
	t.Parallel()

	for schluessel, wert := range netzwege() {
		t.Run(schluessel, func(t *testing.T) {
			t.Parallel()

			ablage := &einstellungsablage{}
			nein := &neinsager{}

			fehler := service.NeuerEinstellungsverwalter(ablage).MitRiegel(nein).
				Setzen(t.Context(), schluessel, wert)
			if !errors.Is(fehler, domain.ErrNurLesend) {
				t.Errorf("%v statt der Absage des Riegels", fehler)
			}

			if !slices.Equal(nein.gefragt, []domain.Wirkung{domain.WirkungExternSenden}) {
				t.Errorf("gefragt wurde nach %v", nein.gefragt)
			}

			if len(ablage.werte) != 0 {
				t.Errorf("trotzdem geschrieben: %v", ablage.werte)
			}

			// **Die Gegenprobe**: wer senden darf, setzt ihn.
			erlaubt := &einstellungsablage{}
			if fehler := service.NeuerEinstellungsverwalter(erlaubt).MitRiegel(jasager{}).
				Setzen(t.Context(), schluessel, wert); fehler != nil || erlaubt.werte[schluessel] != wert {
				t.Errorf("mit Erlaubnis: %v, %v", fehler, erlaubt.werte)
			}
		})
	}
}

// **Ein Schlüssel, der nichts über das Ziel sagt, fragt den Riegel nicht** —
// sonst hielte die Sperre jede Einstellung eines Tokens an, und ein Token
// `schreiben` darf seine Sprache setzen.
func TestEineGewoehnlicheEinstellungFragtDenRiegelNicht(t *testing.T) {
	t.Parallel()

	ablage := &einstellungsablage{}
	nein := &neinsager{}

	if fehler := service.NeuerEinstellungsverwalter(ablage).MitRiegel(nein).
		Setzen(t.Context(), domain.EinstellungSprache, "en"); fehler != nil {
		t.Fatalf("%v", fehler)
	}

	if len(nein.gefragt) != 0 || ablage.werte[domain.EinstellungSprache] != "en" {
		t.Errorf("gefragt %v, geschrieben %v", nein.gefragt, ablage.werte)
	}
}

// sendeverweigerer erlaubt alles außer dem Senden.
type sendeverweigerer struct{ gefragt []domain.Wirkung }

func (s *sendeverweigerer) Pruefe(_ context.Context, wirkung domain.Wirkung) error {
	s.gefragt = append(s.gefragt, wirkung)

	if wirkung == domain.WirkungExternSenden {
		return fmt.Errorf("%w: ein Token sendet nicht", domain.ErrErlaubnis)
	}

	return nil
}

func (*sendeverweigerer) Akteur(context.Context) (domain.Akteur, error) {
	return domain.Akteur{NutzerID: 3, ProfilID: 7}, nil
}

// **Und eine Einfuhr ist kein Umweg** — dieselbe Einstellung aus einer
// Datei.
func TestEineEinfuhrSetztKeinenNetzwegOhneSenderecht(t *testing.T) {
	t.Parallel()

	bank := werkbankBauen()
	wache := &sendeverweigerer{}
	austausch := service.NeuerEinstellungsaustausch(
		bank.ablage, bank.nutzerwerte, bank.schreiber, wache)

	datei := alsDatei(t, map[string]map[string]string{
		string(domain.EbeneProfil): {
			domain.EinstellungSprache:                 "en",
			domain.DienstAdresse(domain.DienstOpenAI): "https://gateway.example",
		},
	})

	revision, fehler := austausch.Vorschau(t.Context(), datei)
	if fehler != nil {
		t.Fatalf("die Vorschau: %v", fehler)
	}

	if _, fehler = austausch.Uebernehmen(
		t.Context(), datei, revision.Revision, nil); !errors.Is(fehler, domain.ErrErlaubnis) {
		t.Errorf("%v statt der Absage des Riegels", fehler)
	}

	if bank.schreiber.rufe != 0 {
		t.Errorf("trotzdem %d-mal geschrieben", bank.schreiber.rufe)
	}

	// **Die Gegenprobe**: ohne die Adresse geht dieselbe Einfuhr durch —
	// die Absage galt dem Netzweg und nicht der Einfuhr.
	ohne := alsDatei(t, map[string]map[string]string{
		string(domain.EbeneProfil): {domain.EinstellungSprache: "en"},
	})

	vorschau, fehler := austausch.Vorschau(t.Context(), ohne)
	if fehler != nil {
		t.Fatalf("die Vorschau: %v", fehler)
	}

	if _, fehler := austausch.Uebernehmen(t.Context(), ohne, vorschau.Revision, nil); fehler != nil {
		t.Errorf("ohne Netzweg: %v", fehler)
	}
}
