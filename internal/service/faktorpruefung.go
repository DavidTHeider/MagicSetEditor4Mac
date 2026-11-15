// **Der zweite Faktor: prüfen** — Werkstück 7 von Phase 4.12
// (#1351).
//
// **Eine eigene Datei**, damit `zweiterfaktor.go` unter der Zeilengrenze aus
// §2 bleibt. Der Schnitt folgt der Frage: dort steht, wie ein Faktor
// **entsteht** und verschwindet, hier, wie er **gilt**.
//
// # Die Reihenfolge ist die Aussage
//
// **Wer keinen Faktor hat, wird nicht gefragt** — die Abnahme aus dem Issue:
// „ein bestehender Aufruf ohne Anmeldung bricht nicht."
//
// **Wer einen hat und nichts eingibt, wird über den Sensor gefragt** — wenn
// es einen gibt. Sonst bekommt er die Aufforderung, seinen Code zu nennen.
//
// **Und ein Code, der kein Einmalkennwort ist, wird gegen die
// Wiederherstellungscodes gehalten.** Das ist der Weg für den, dessen
// Schlüsselbund das Geheimnis nicht kennt.

package service

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/totp"
)

// Pruefe lässt einen Nutzer den zweiten Schritt gehen — oder gar keinen.
//
// **Ein leerer Code heißt „über den Sensor"**, und das ist keine Abkürzung:
// wer Touch ID eingerichtet hat, tippt nichts. Wer etwas tippt, meint einen
// Code — aus der App oder von seinem Zettel.
func (z *Zweiterfaktor) Pruefe(ctx context.Context, nutzerID int64, code string) error {
	faktoren, fehler := z.ablage.Faktoren(ctx, nutzerID)
	if fehler != nil {
		return fehler
	}

	if len(faktoren) == 0 {
		return nil
	}

	if code == "" {
		return z.ueberDenSensor(ctx, faktoren)
	}

	_, fehler = imBudget(ctx, z, nutzerID, func() (struct{}, error) {
		return struct{}{}, z.ueberEinenCode(ctx, nutzerID, code, faktoren)
	})

	return fehler
}

// imBudget prüft einen Code **im Budget der Fehlversuche** (SEC-10, #2151,
// Vermerk 0178).
//
// **Jeder Codeweg geht hier durch** — Anmeldung, Wechsel, Entsperren und der
// Zettel beim vergessenen Kennwort ([Zweiterfaktor.CodeFinden]). Der Versuch
// zählt, bevor geprüft wird, und erst ein gelungener setzt zurück; ein
// Fehler unterwegs — ein Schlüsselbund, der nicht antwortet — bleibt
// gezählt, denn von außen ist er von einem falschen Code nicht zu
// unterscheiden. **Der Sensor geht nicht hindurch**: Touch ID lässt sich
// nicht raten, und wer ausgesperrt wurde, kommt am eigenen Rechner so herein.
func imBudget[T any](
	ctx context.Context, z *Zweiterfaktor, nutzerID int64, pruefen func() (T, error),
) (T, error) {
	var leer T

	bis, fehler := z.ablage.FaktorversuchBeginnen(ctx, nutzerID, z.uhr())
	if fehler != nil {
		return leer, fehler
	}

	if !bis.IsZero() {
		return leer, domain.FaktorGesperrt(bis)
	}

	ergebnis, pruefehler := pruefen()
	if pruefehler != nil {
		return leer, pruefehler
	}

	return ergebnis, z.ablage.FaktorversuchGelungen(ctx, nutzerID)
}

// PruefeBeimWechsel prüft den zweiten Faktor **beim Profilwechsel**.
//
// # Warum das ein eigener Weg ist
//
// **Weil der Nutzer ihn hier abschalten darf und beim Anmelden nicht** —
// Punkt 11 der Kommentierung, `sicherheit.faktorbeimwechsel`. Auf einem
// Rechner mit einem Menschen ist ein Einmalkennwort bei jedem Profilwechsel
// Reibung ohne Gegenwert; **beim Anmelden** steht der Faktor für das ein,
// wofür es ihn gibt.
//
// **Ohne Einstellungen wird geprüft.** Ein Doppel in einem Lauf, der nur den
// Wechsel meint, soll nicht an einer Naht scheitern, die zu seinem Fall
// nichts beiträgt — und die strengere Antwort ist die richtige, wenn niemand
// gefragt hat.
func (z *Zweiterfaktor) PruefeBeimWechsel(
	ctx context.Context, nutzerID int64, code string,
) error {
	if z.einstellungen != nil {
		wert, fehler := z.einstellungen.Wert(
			ctx, nutzerID, domain.NutzerFaktorbeimWechsel)
		if fehler != nil {
			return fehler
		}

		if wert == domain.FaktorNurAnmelden {
			return nil
		}
	}

	return z.Pruefe(ctx, nutzerID, code)
}

// VerlangtBeimWechsel sagt, ob [Zweiterfaktor.PruefeBeimWechsel] für diesen
// Nutzer fragen würde — ohne selbst zu fragen.
//
// **Dieselbe Reihenfolge wie dort**: wer den Faktor beim Wechsel abgeschaltet
// hat, wird beim Wechsel nicht gefragt, und wer keinen hat, auch nicht. Wer
// hier strenger antwortete als der Wechsel selbst, sperrte jemanden aus, zu dem
// jeder ohnehin wechseln kann — eine Hürde, die man mit einem Umweg nimmt.
func (z *Zweiterfaktor) VerlangtBeimWechsel(ctx context.Context, nutzerID int64) (bool, error) {
	if z.einstellungen != nil {
		wert, fehler := z.einstellungen.Wert(ctx, nutzerID, domain.NutzerFaktorbeimWechsel)
		if fehler != nil {
			return false, fehler
		}

		if wert == domain.FaktorNurAnmelden {
			return false, nil
		}
	}

	art, fehler := z.Verlangt(ctx, nutzerID)

	return art != "", fehler
}

// Verlangt sagt, welchen zweiten Faktor dieser Nutzer verlangt.
//
// **Ohne etwas anzustoßen** — die Fläche fragt vorher, um das richtige Feld
// zu zeigen. Ein leerer Wert heißt: keinen.
func (z *Zweiterfaktor) Verlangt(ctx context.Context, nutzerID int64) (
	domain.Faktorart, error,
) {
	faktoren, fehler := z.ablage.Faktoren(ctx, nutzerID)
	if fehler != nil {
		return "", fehler
	}

	if len(faktoren) == 0 {
		return "", nil
	}

	return bevorzugte(faktoren), nil
}

// ueberDenSensor fragt die Biometrie — wenn dieser Nutzer sie hat und dieser
// Rechner sie kann.
func (z *Zweiterfaktor) ueberDenSensor(
	ctx context.Context, faktoren []domain.Faktor,
) error {
	if hat(faktoren, domain.FaktorBiometrie) && z.sensor.Vorhanden(ctx) {
		return z.sensor.Bestaetige(ctx, grundPruefen)
	}

	// **Sonst wird gesagt, was fehlt** — und **welche Art**, damit die Fläche
	// das richtige Feld anbietet. Wer Touch ID eingerichtet hat und den
	// Bestand auf einem Mac ohne Sensor öffnet, landet hier und kommt über
	// seinen Code oder seinen Zettel herein (§16, Regel 5, sinngemäß).
	return domain.FehlenderFaktor(bevorzugte(faktoren))
}

// ueberEinenCode hält den Code gegen das Einmalkennwort und den Zettel.
func (z *Zweiterfaktor) ueberEinenCode(
	ctx context.Context, nutzerID int64, code string, faktoren []domain.Faktor,
) error {
	// **Erst das Einmalkennwort**, weil es der Normalfall ist und weil es
	// nichts verbraucht: ein Fehlversuch kostet keinen Wiederherstellungscode.
	if hat(faktoren, domain.FaktorTOTP) {
		zaehler, fehler := z.zaehlerFuer(ctx, nutzerID)

		switch {
		case fehler == nil:
			// **Und gilt nur einmal** (SEC-09, #2150): verbraucht wird der
			// Schritt, zu dem der Code passt, und ein zweiter Prozess mit
			// demselben Code trifft null Zeilen. Ein verbrauchter fällt
			// weiter zum Zettel und scheitert dort.
			if schritt, stimmt := totp.PassenderSchritt(zaehler, code, z.uhr()); stimmt == nil {
				gilt, verbrauchfehler := z.ablage.SchrittVerbrauchen(ctx, nutzerID, schritt)
				if verbrauchfehler != nil || gilt {
					return verbrauchfehler
				}
			}

		// **Kein Geheimnis ist kein Fehler, sondern der Anlass für den
		// Zettel**: genau dann, wenn der Schlüsselbund eines anderen Macs es
		// nicht kennt, braucht jemand den Rückweg.
		case !istNichtGefunden(fehler):
			return fehler
		}
	}

	return z.ueberDenZettel(ctx, nutzerID, code)
}

// ueberDenZettel hält den Code gegen die Wiederherstellungscodes.
//
// **Jeder freie wird geprüft, nicht der erste passende gesucht.** Argon2id
// gibt für jeden Code einen anderen Salzwert; einen davon zu finden geht nur,
// indem man jeden rechnet.
func (z *Zweiterfaktor) ueberDenZettel(
	ctx context.Context, nutzerID int64, code string,
) error {
	codeID, fehler := z.codeFinden(ctx, nutzerID, code)
	if fehler != nil {
		return fehler
	}

	// **Verbraucht wird erst, wenn er passt** — und das Verbrauchen
	// entscheidet der Bestand: zwei gleichzeitige Anmeldungen mit demselben
	// Zettel treffen null Zeilen beim zweiten (§7, Regel 4).
	return z.ablage.CodeVerbrauchen(ctx, codeID)
}

// hat sagt, ob diese Art unter den Faktoren ist.
func hat(faktoren []domain.Faktor, art domain.Faktorart) bool {
	for _, einer := range faktoren {
		if einer.Art == art {
			return true
		}
	}

	return false
}

// bevorzugte sagt, welche Art die Fläche anbieten soll.
//
// **Biometrie vor TOTP**, weil sie ohne Tippen geht — und wer sie nicht
// bedienen kann, bekommt vom Prüfer trotzdem beide Wege: der Code aus der App
// und der Zettel gelten weiter.
func bevorzugte(faktoren []domain.Faktor) domain.Faktorart {
	if hat(faktoren, domain.FaktorBiometrie) {
		return domain.FaktorBiometrie
	}

	return domain.FaktorTOTP
}
