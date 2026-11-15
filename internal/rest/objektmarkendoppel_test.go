package rest_test

// **Die Doppel der Marken an jedem Fachobjekt** — UX-12 (#1684).

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func (williger) ObjektmarkeSetzen(context.Context, string, string, string) error    { return nil }
func (williger) ObjektmarkeEntfernen(context.Context, string, string, string) error { return nil }

func (williger) ObjektmarkenAn(context.Context, string, string) ([]string, error) {
	return []string{}, nil
}

func (williger) MarkenJeObjekt(context.Context, string) (map[string][]string, error) {
	return map[string][]string{}, nil
}

func (williger) MarkeVergeben(context.Context, string, []string, string) error { return nil }

func (williger) MarkenFiltern(context.Context, string, []string, string) ([]string, error) {
	return []string{}, nil
}

func (williger) Markenkatalog(context.Context) ([]domain.Markenposten, error) {
	return []domain.Markenposten{}, nil
}

func (williger) MarkeUmbenennen(context.Context, string, string, bool) (int, error) { return 1, nil }

func (williger) MarkeUeberallEntfernen(context.Context, string) (int, error) { return 1, nil }

func (z *zaehlender) ObjektmarkeSetzen(context.Context, string, string, string) error {
	z.zaehle()

	return nil
}

func (z *zaehlender) ObjektmarkeEntfernen(context.Context, string, string, string) error {
	z.zaehle()

	return nil
}

func (z *zaehlender) ObjektmarkenAn(context.Context, string, string) ([]string, error) {
	z.zaehle()

	return nil, nil
}

func (z *zaehlender) MarkenJeObjekt(context.Context, string) (map[string][]string, error) {
	z.zaehle()

	return map[string][]string{}, nil
}

func (z *zaehlender) MarkeVergeben(context.Context, string, []string, string) error {
	z.zaehle()

	return nil
}

func (z *zaehlender) MarkenFiltern(context.Context, string, []string, string) ([]string, error) {
	z.zaehle()

	return nil, nil
}

func (z *zaehlender) Markenkatalog(context.Context) ([]domain.Markenposten, error) {
	z.zaehle()

	return []domain.Markenposten{}, nil
}

func (z *zaehlender) MarkeUmbenennen(context.Context, string, string, bool) (int, error) {
	z.zaehle()

	return 1, nil
}

func (z *zaehlender) MarkeUeberallEntfernen(context.Context, string) (int, error) {
	z.zaehle()

	return 1, nil
}

func (u unwilliger) ObjektmarkeSetzen(context.Context, string, string, string) error { return u.fehler }

func (u unwilliger) ObjektmarkeEntfernen(context.Context, string, string, string) error {
	return u.fehler
}

func (u unwilliger) ObjektmarkenAn(context.Context, string, string) ([]string, error) {
	return nil, u.fehler
}

func (u unwilliger) MarkenJeObjekt(context.Context, string) (map[string][]string, error) {
	return nil, u.fehler
}

func (u unwilliger) MarkeVergeben(context.Context, string, []string, string) error { return u.fehler }

func (u unwilliger) MarkenFiltern(context.Context, string, []string, string) ([]string, error) {
	return nil, u.fehler
}

func (u unwilliger) Markenkatalog(context.Context) ([]domain.Markenposten, error) {
	return nil, u.fehler
}

func (u unwilliger) MarkeUmbenennen(context.Context, string, string, bool) (int, error) {
	return 0, u.fehler
}

func (u unwilliger) MarkeUeberallEntfernen(context.Context, string) (int, error) { return 0, u.fehler }
