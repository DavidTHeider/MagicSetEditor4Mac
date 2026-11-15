package service

// **Das Verzeichnis hinter dem Abgleich** — Werkstück `T4` von Phase 5.6
// (#1469).
//
// # Einmal lesen, dann nachschlagen
//
// **Beide Verzeichnisse entstehen vor der Rechnung**, nicht in ihr: das
// Konzept je Druck und der Preis je Druck und Ausführung. Ein Nachschlag je
// Vergleichszeile machte aus einer Rechnung über zwei Listen eine über ihr
// Produkt — genau das, was dieses Werkstück ausschließt.

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Preisverzeichnis ist, was der Abgleich von der Preisablage braucht.
type Preisverzeichnis interface {
	NeuesteAlle(ctx context.Context) ([]domain.Preisstand, error)
}

// verzeichnisLesen baut die Zuordnung Druck → Konzept.
//
// **Über den Scryfall-Alias und die lokale Zeile**, denn beides kann in einer
// Sammlungszeile stehen ([domain.Druckschluessel]). Ein Druck ohne
// Konzeptkennung steht nicht darin — und ein druckunabhängiger Wunsch auf ihn
// wird dann zum Befund statt zur Vermutung.
func (t *Tauschwerk) verzeichnisLesen(ctx context.Context) (map[string]string, error) {
	drucke, fehler := t.drucke.Alle(ctx)
	if fehler != nil {
		return nil, fehler
	}

	verzeichnis := make(map[string]string, len(drucke))

	for _, druck := range drucke {
		konzept, hat := druck.Kennung(domain.SystemScryfall, domain.SemantikKonzept)
		if !hat {
			continue
		}

		if alias := druck.Scryfall(); alias != "" {
			verzeichnis[alias] = konzept
		}

		if druck.ID != 0 {
			verzeichnis[domain.Druckschluessel(druck.ID, "")] = konzept
		}
	}

	return verzeichnis, nil
}

// preiseLesen baut die Zuordnung Druck und Ausführung → Preis.
//
// **Nur die gewählte Reihe.** Ein Stand, der zu einer anderen Quelle, einem
// anderen Markt oder einer anderen Richtung gehört, ist kein Preis für diesen
// Tausch — er ist ein Preis für eine andere Frage (A3).
func (t *Tauschwerk) preiseLesen(ctx context.Context) (map[string]domain.Preisstand, error) {
	if t.preise == nil {
		return nil, nil
	}

	staende, fehler := t.preise.NeuesteAlle(ctx)
	if fehler != nil {
		return nil, fehler
	}

	verzeichnis := map[string]domain.Preisstand{}

	for _, stand := range staende {
		if !passtZurReihe(stand, t.reihe) {
			continue
		}

		verzeichnis[preisschluessel(stand.Printing, stand.Finish)] = stand
	}

	return verzeichnis, nil
}

// passtZurReihe sagt, ob ein Stand zur gewählten Reihe gehört.
//
// **Was in der Wahl leer steht, wird nicht geprüft.** Wer nur die Quelle
// nennt, bekommt jede ihrer Reihen; wer alle vier nennt, genau eine. Eine
// Prüfung gegen die Leere wiese jeden Stand ab und machte die Wahl unbrauchbar.
func passtZurReihe(stand, reihe domain.Preisstand) bool {
	if reihe.Quelle != "" && stand.Quelle != reihe.Quelle {
		return false
	}

	if reihe.Markt != "" && stand.Markt != reihe.Markt {
		return false
	}

	if reihe.Art != "" && stand.Art != reihe.Art {
		return false
	}

	return reihe.Waehrung == "" || stand.Waehrung == reihe.Waehrung
}

// preisschluessel nennt Druck und Ausführung zusammen.
//
// **Beides, weil ein Foil anders kostet.** Ein Verzeichnis nur über den Druck
// gäbe dem gewöhnlichen Stück den Preis des veredelten.
func preisschluessel(printing string, finish domain.Finish) string {
	if finish == "" {
		finish = domain.FinishNonfoil
	}

	return printing + "\x00" + string(finish)
}

// bepreist hängt an jede Zeile ihren Preis — **oder keinen**.
//
// **Eine Zeile ohne Preis bleibt ohne Preis** (§8). Sie mit null zu bewerten
// hieße, sie sei nichts wert; sie auszulassen hieße, sie gäbe es nicht.
func bepreist(
	zeilen []domain.Angebotszeile, preise map[string]domain.Preisstand,
) []domain.Angebotszeile {
	if len(preise) == 0 {
		return zeilen
	}

	for stelle, zeile := range zeilen {
		stand, hat := preise[preisschluessel(zeile.Druck, zeile.Finish)]
		if !hat {
			continue
		}

		zeilen[stelle].Preis, zeilen[stelle].Waehrung = stand.Wert, stand.Waehrung
	}

	return zeilen
}

// alsAngebotszeile macht aus einem eigenen Exemplar eine Angebotszeile.
func alsAngebotszeile(
	exemplar *domain.Exemplar, verzeichnis map[string]string,
) domain.Angebotszeile {
	druck := exemplar.Druckschluessel()

	return domain.Angebotszeile{
		Druck: druck, Konzept: verzeichnis[druck], Menge: exemplar.Menge,
		Zustand: exemplar.Zustand, Sprache: exemplar.Sprache,
		Finish: exemplar.Ausfuehrung(), Bezeichnung: exemplar.Bezeichnung(),
	}
}

// alsWunschzeile macht aus einem eigenen Wunsch eine Wunschzeile.
func alsWunschzeile(
	exemplar *domain.Exemplar, verzeichnis map[string]string,
) domain.Wunschzeile {
	druck := exemplar.Druckschluessel()

	return domain.Wunschzeile{
		Druck: druck, Konzept: verzeichnis[druck], Menge: exemplar.Menge,
		Zustand: exemplar.Zustand, Sprache: exemplar.Sprache,
		Finish: exemplar.Ausfuehrung(), Art: exemplar.Druckbindung(),
		Rang: exemplar.Dringlichkeit(), Bezeichnung: exemplar.Bezeichnung(),
	}
}

// alsAngebotszeilen macht aus fremden Sichtkarten Angebotszeilen.
func alsAngebotszeilen(
	karten []domain.Sichtkarte, verzeichnis map[string]string,
) []domain.Angebotszeile {
	zeilen := make([]domain.Angebotszeile, 0, len(karten))

	for _, karte := range karten {
		zeilen = append(zeilen, domain.Angebotszeile{
			Druck: karte.Printing, Konzept: verzeichnis[karte.Printing],
			Menge: karte.Menge, Zustand: karte.Zustand, Sprache: karte.Sprache,
			Finish: karte.Finish, Bezeichnung: karte.Bezeichnung,
		})
	}

	return zeilen
}

// alsWunschzeilen macht aus fremden Sichtkarten Wunschzeilen.
//
// **Und zwar als `genau`.** Was ein Gegenüber will, sagt seine Zeile nicht:
// die freigegebene Sicht trägt sechs Felder, und eine Wunschart ist keines
// davon. Die engere Annahme ist die richtige — sie bietet niemandem etwas an,
// das er nicht genau so gewollt hat (T3, [#1468]).
//
// **Der Rang fehlt ihr ebenso**, und deshalb gilt `normal`: die Dringlichkeit
// eines Fremden ordnet nicht die eigene Liste.
func alsWunschzeilen(
	karten []domain.Sichtkarte, verzeichnis map[string]string,
) []domain.Wunschzeile {
	zeilen := make([]domain.Wunschzeile, 0, len(karten))

	for _, karte := range karten {
		zeilen = append(zeilen, domain.Wunschzeile{
			Druck: karte.Printing, Konzept: verzeichnis[karte.Printing],
			Menge: karte.Menge, Zustand: karte.Zustand, Sprache: karte.Sprache,
			Finish: karte.Finish, Art: domain.WunschGenau, Rang: domain.RangNormal,
			Bezeichnung: karte.Bezeichnung,
		})
	}

	return zeilen
}
