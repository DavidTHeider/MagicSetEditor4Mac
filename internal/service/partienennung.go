package service

// **Welche Partie der Wirt meint — aus seinem ersten Wort** — UX-08.AK-04
// (#1680).
//
// # Wozu
//
// **Damit ein Gast keine Partiekennung tippt.** Der Wirt hält genau eine
// Partie für Gäste offen (`app.PartieFreigeben`: „Eine zur Zeit"), und welche
// es ist, weiß nur er. Bis hierher musste der Gast sie trotzdem nennen —
// abgelesen von einem anderen Bildschirm; und ein Tippfehler hieß, dass jede
// Nachricht des Wirts an [Partiegast.Angekommen] abprallte, ohne dass das
// Fenster es sagte.
//
// # Warum der Wirt zuerst spricht
//
// **Weil jede Nachricht ihre Partie trägt, auch die erste des Gastes** — und
// [domain.Partienachricht.Pruefe] weist eine ohne ab. Wer sie nicht kennt,
// kann also nichts sagen, bevor er sie gehört hat. Der Wirt schickt deshalb
// gleich nach der Aufnahme das Tischbild des vergebenen Sitzes
// ([Partiewirt.Begruesse]); sein Umschlag nennt die Partie, und das Bild ist
// ohnehin, was der Gast als Erstes sehen soll.
//
// # Und was ein Wirt vor 0.6.0 tut
//
// **Er schweigt, bis der Gast spricht.** Ein Gast ohne Partie wartet dann,
// bis er aufsteht; das Fenster sagt es dort, wo der Tisch fehlt, und die
// Partie lässt sich unter „Für Fachleute" nennen. **Umgekehrt ändert sich
// nichts**: ein älterer Gast, der die Partie nennt, bekommt ein Tischbild
// mehr und nimmt es wie jedes andere — das Protokoll bleibt `/mse4mac/game/1`.

import (
	"context"
	"encoding/json"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Begruesse schickt dem Gast sein Tischbild, bevor er etwas gesagt hat.
//
// **Es beantwortet die Frage, die er nicht stellen kann**: welche Partie hier
// gilt ([Partieerfragen]).
//
// **In der Reihe** (#2481), wie jedes Bild, das nicht aus einem Zug kommt.
func (w *Partiewirt) Begruesse(ctx context.Context) error {
	return w.nehmer.InReihe(func() error { return w.tischbildSchicken(ctx) })
}

// Partieerfragen liest, bis der Wirt etwas Gültiges sagt, und gibt dessen
// Partie.
//
// **Die Nachricht geht dabei nicht verloren**: der zurückgegebene Zugang
// liefert sie beim ersten Lesen noch einmal, und die Schleife danach urteilt
// über sie wie über jede andere — Epoche, Folge, Stand. Ein Vorzimmer, das sie
// selbst auswertete, wäre ein zweiter Empfang neben dem einen.
//
// **Was nicht mehr gilt oder sich nicht lesen lässt, wird übergangen** —
// dieselbe Haltung wie in [Partienetz.Bediene]: eine unlesbare Zustellung ist
// ein Befund über **eine** Nachricht und kein Grund, den Tisch aufzugeben.
func Partieerfragen(
	ctx context.Context, zugang Angenommene,
) (string, Angenommene, error) {
	for {
		roh, fehler := zugang.Empfange(ctx)
		if fehler != nil {
			return "", nil, fehler
		}

		if !zugang.Empfangen(roh) {
			continue
		}

		var nachricht domain.Partienachricht

		if json.Unmarshal(roh, &nachricht) != nil || nachricht.Pruefe() != nil {
			continue
		}

		return nachricht.Partie, &vorgelesen{Angenommene: zugang, erste: roh}, nil
	}
}

// vorgelesen liefert eine schon gelesene Nachricht noch einmal.
type vorgelesen struct {
	Angenommene

	erste []byte
}

// Empfange gibt zuerst die vorgelesene, danach, was die Leitung bringt.
//
// **Ohne Schloss, und das ist eine Aussage**: gelesen wird nur in der
// Schleife, die diesen Zugang bekommt — und die ist eine.
func (v *vorgelesen) Empfange(ctx context.Context) ([]byte, error) {
	if erste := v.erste; erste != nil {
		v.erste = nil

		return erste, nil
	}

	return v.Angenommene.Empfange(ctx)
}
