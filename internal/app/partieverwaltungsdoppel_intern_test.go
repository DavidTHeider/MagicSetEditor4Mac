package app

// **Die vier Verwaltungswege an den beiden Leitungsdoppeln** —
// PLAY-03.AK-05
// (#1711).
//
// **Eine eigene Datei**, damit `partieleitungsdoppel_intern_test.go` unter
// der Zeilengrenze aus §2 bleibt. Die Typen bleiben zwei — Methoden dürfen
// daneben liegen.

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func (s *schweigsameLeitung) Loeschen(context.Context, string) error {
	return nil
}

func (s *schweigsameLeitung) Kopieren(
	_ context.Context, _, neueKennung string,
) (domain.Partie, error) {
	return domain.Partie{ID: neueKennung}, nil
}

func (s *schweigsameLeitung) Ausfuhr(
	context.Context, string,
) ([]byte, error) {
	return []byte("{}"), nil
}

func (s *schweigsameLeitung) Einfuhr(
	_ context.Context, _ []byte, neueKennung string,
) (domain.Partie, []domain.Einfuhrbefund, error) {
	return domain.Partie{ID: neueKennung},
		[]domain.Einfuhrbefund{{Grund: "ein fremdes Feld"}}, nil
}

func (k *klagendePartieleitung) Loeschen(context.Context, string) error {
	return k.stolpert("Loeschen")
}

func (k *klagendePartieleitung) Kopieren(
	_ context.Context, _, neueKennung string,
) (domain.Partie, error) {
	return domain.Partie{ID: neueKennung}, k.stolpert("Kopieren")
}

func (k *klagendePartieleitung) Ausfuhr(
	context.Context, string,
) ([]byte, error) {
	return []byte("{}"), k.stolpert("Ausfuhr")
}

func (k *klagendePartieleitung) Einfuhr(
	_ context.Context, _ []byte, neueKennung string,
) (domain.Partie, []domain.Einfuhrbefund, error) {
	return domain.Partie{ID: neueKennung}, nil, k.stolpert("Einfuhr")
}
