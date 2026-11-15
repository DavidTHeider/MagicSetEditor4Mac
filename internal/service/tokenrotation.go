// **Rotation, Sammelwiderruf und Ablaufwahlen** — Werkstück O2 von Phase 5.2
// (#1426),
// nach [Vermerk 0080](../../docs/entscheidungen/0080-maschinenzugaenge-und-widerrufbare-sitzungen.md).
//
// # Was eine Rotation zusagt
//
// **Ein neues Geheimnis, dasselbe Token, und das alte gilt noch eine Stunde**
// ([domain.Tokenuebergang]) — lang genug, dass ein laufender Auftrag zu Ende
// kommt und ein Mensch das neue Geheimnis in die Einrichtung seines
// Gegenübers trägt; kurz genug, dass es keine zweite Tür bleibt. Wer noch
// einmal rotiert, wirft den Vorgänger sofort weg; wer das Token zurückzieht,
// sperrt beide.
//
// **Der Klartext kommt genau einmal heraus**, aus diesem Ruf — wie beim
// Anlegen. Im Bestand steht nur die gehütete Form, und zwar beide.

package service

import (
	"context"
	"crypto/rand"
	"strconv"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Rotieren gibt einem Token ein neues Geheimnis und den Klartext dazu.
func (v *Tokenverwaltung) Rotieren(
	ctx context.Context, tokenID int64,
) (domain.Zugriffstoken, string, error) {
	geheimnis := rand.Text()
	frist := v.uhr().Add(domain.Tokenuebergang).UTC().Format(time.RFC3339)

	rotiert, fehler := v.ablage.TokenRotieren(ctx, tokenID, huete(geheimnis), frist)
	if fehler != nil {
		return domain.Zugriffstoken{}, "", fehler
	}

	return rotiert, Tokenmarke + "_" + strconv.FormatInt(rotiert.ID, 10) +
		"_" + geheimnis, nil
}

// AlleWiderrufen zieht jedes Token eines Profils zurück und sagt, wie viele
// es waren.
func (v *Tokenverwaltung) AlleWiderrufen(ctx context.Context, profilID int64) (int64, error) {
	return v.ablage.TokenAlleLoeschen(ctx, profilID)
}

// Ablaufwahlen gibt die Vorschläge für den Ablauf, gerechnet aus der Uhr
// dieses Dienstes — damit alle vier Flächen dieselben Zeitpunkte nennen.
func (v *Tokenverwaltung) Ablaufwahlen() []domain.Ablaufwahl {
	return domain.Ablaufwahlen(v.uhr())
}
