package service

// **Was ein Gegenüber zu sehen bekommt** — Werkstück `T3` von Phase 5.6
// (#1468).
//
// # Der Satz, um den herum das gebaut ist
//
// **Ohne Freigabe sieht niemand etwas.** Das Issue sagt es zweimal, und es
// verlangt die Prüfung vor **jeder** Projektion — vor der Antwort, vor dem
// Zwischenspeicher, vor der Wiederaufnahme, vor Ausfuhr und Beleg. Deshalb
// gibt es **einen** Weg zu einer Sicht, und er fragt am Anfang.
//
// # Zwei Neins, die gleich klingen müssen
//
// **„Diesen Kontakt gibt es nicht" und „du darfst nicht" sind dieselbe
// Antwort** (§8). Wer sie unterschiede, ließe die Kontaktliste dieses Profils
// abfragen, ohne je eine Sicht zu bekommen.

import (
	"context"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Freigabeablage ist, was die Pflege zum Lesen und Schreiben braucht.
//
// **Die Schnittstelle steht hier und wird in `storage` erfüllt** (§2).
type Freigabeablage interface {
	Erteilen(ctx context.Context, freigabe domain.Sammlungsfreigabe) error
	Widerrufen(
		ctx context.Context, empfaenger string, zweck domain.Zweck,
		umfang domain.Freigabeumfang,
	) error
	Fuer(ctx context.Context, empfaenger string) ([]domain.Sammlungsfreigabe, error)
}

// Kontaktnachschlag holt einen Peerkontakt nach seiner Kennung.
type Kontaktnachschlag interface {
	Lesen(ctx context.Context, kennung string) (domain.Peerkontakt, error)
}

// Freigabepflege ist der Anwendungsfall „was darf mein Gegenüber sehen?".
type Freigabepflege struct {
	freigaben Freigabeablage
	kontakte  Kontaktnachschlag
	sammlung  Sammlungsablage
	uhr       func() time.Time

	// reserviert ist die Naht zum Handeln (T2, #1467) — **wahlfrei**.
	reserviert Reservierungen

	// schranke ist die Handelsrichtlinie (#2435) — Teilen ist eine ihrer
	// vier Vertragsarten.
	schranke Handelsschranke
}

// NeueFreigabepflege baut sie.
func NeueFreigabepflege(
	freigaben Freigabeablage, kontakte Kontaktnachschlag,
	sammlung Sammlungsablage, uhr func() time.Time,
) *Freigabepflege {
	return &Freigabepflege{
		freigaben: freigaben, kontakte: kontakte, sammlung: sammlung, uhr: uhr,
		schranke: ohneEinstellung{},
	}
}

// MitHandelsschranke setzt die Handelsrichtlinie, die gilt (#2435).
func (f *Freigabepflege) MitHandelsschranke(schranke Handelsschranke) *Freigabepflege {
	f.schranke = schranke

	return f
}

// MitReservierungen setzt die Naht zum Handeln (T2, #1467, 0120).
func (f *Freigabepflege) MitReservierungen(
	reserviert Reservierungen,
) *Freigabepflege {
	f.reserviert = reserviert

	return f
}

// Erteilen gibt einem Gegenüber eine Sicht frei.
func (f *Freigabepflege) Erteilen(
	ctx context.Context, freigabe domain.Sammlungsfreigabe,
) error {
	if fehler := artErlaubt(ctx, f.schranke, domain.HandelTeilen); fehler != nil {
		return fehler
	}

	if freigabe.Erteilt.IsZero() {
		freigabe.Erteilt = f.uhr()
	}

	return f.freigaben.Erteilen(ctx, freigabe.MitVorgaben())
}

// Widerrufen nimmt sie zurück.
//
// **Künftige Antworten hören damit auf** — das ist alles, was dieses Programm
// zusagen kann. Was ein Gegenüber schon kopiert hat, ist bei ihm, und kein
// Widerruf holt es zurück; das Issue sagt es so, und die Oberfläche sagt es
// dem Menschen.
func (f *Freigabepflege) Widerrufen(
	ctx context.Context, empfaenger string, zweck domain.Zweck,
	umfang domain.Freigabeumfang,
) error {
	return f.freigaben.Widerrufen(ctx, empfaenger, zweck, umfang)
}

// Freigaben gibt, was einem Gegenüber erteilt ist — für die eigene Fläche.
func (f *Freigabepflege) Freigaben(
	ctx context.Context, empfaenger string,
) ([]domain.Sammlungsfreigabe, error) {
	return f.freigaben.Fuer(ctx, empfaenger)
}

// Sicht gibt, was ein Gegenüber sehen darf.
//
// **Die Sitzung benennt die Zeilen** ([domain.Sitzungsreferenz]): innerhalb
// eines Gesprächs zeigt derselbe Name auf dieselbe Zeile, und außerhalb sagt
// er nichts.
func (f *Freigabepflege) Sicht(
	ctx context.Context, empfaenger, sitzung string, zweck domain.Zweck,
) ([]domain.Sichtkarte, error) {
	gelten, fehler := f.erlaubnis(ctx, empfaenger, zweck)
	if fehler != nil {
		return nil, fehler
	}

	exemplare, lesefehler := f.sammlung.Liste(ctx)
	if lesefehler != nil {
		return nil, lesefehler
	}

	festgehalten, zusagenfehler := f.festgehalten(ctx)
	if zusagenfehler != nil {
		return nil, zusagenfehler
	}

	return zeigen(gelten, exemplare, festgehalten, sitzung), nil
}

// erlaubnis fragt Kontakt und Freigaben — und gibt auf beide dasselbe Nein.
func (f *Freigabepflege) erlaubnis(
	ctx context.Context, empfaenger string, zweck domain.Zweck,
) ([]domain.Sammlungsfreigabe, error) {
	kontakt, fehler := f.kontakte.Lesen(ctx, empfaenger)
	if fehler != nil {
		// **Derselbe Satz wie bei „darfst du nicht".** Ein eigener verriete,
		// ob dieses Profil das Gegenüber kennt.
		return nil, domain.MitGrund(domain.ErrErlaubnis, fehler,
			"Für diesen Zweck ist nichts freigegeben.")
	}

	erteilt, lesefehler := f.freigaben.Fuer(ctx, empfaenger)
	if lesefehler != nil {
		return nil, lesefehler
	}

	return domain.FreigabenFuer(kontakt, erteilt, zweck, f.uhr())
}

// festgehalten sagt, was je Exemplar einem Handel zugesagt ist.
func (f *Freigabepflege) festgehalten(
	ctx context.Context,
) (map[int64]int, error) {
	if f.reserviert == nil {
		return map[int64]int{}, nil
	}

	return f.reserviert.ReserviertJeExemplar(ctx)
}

// zeigen baut die Sicht aus den Zeilen, die eine Freigabe zeigt.
//
// **Zugesagte Stücke fallen ab** (T2, 0120): eine Sicht, die sie zeigte, lüde
// einen zweiten Partner ein, danach zu fragen — und die Absage käme erst beim
// Zusagen. Das Issue nennt den Fall nicht; er folgt aus `T2`.
func zeigen(
	gelten []domain.Sammlungsfreigabe, exemplare []*domain.Exemplar,
	festgehalten map[int64]int, sitzung string,
) []domain.Sichtkarte {
	sicht := []domain.Sichtkarte{}

	for _, exemplar := range exemplare {
		menge, zeigt := domain.Sichtmenge(gelten, *exemplar)
		if !zeigt {
			continue
		}

		frei := menge - festgehalten[exemplar.ID]
		if frei <= 0 {
			continue
		}

		gezeigt := *exemplar
		gezeigt.Menge = frei

		sicht = append(sicht, domain.AlsSichtkarte(
			gezeigt, domain.Sitzungsreferenz(sitzung, exemplar.ID)))
	}

	return sicht
}
