package service

// **Die Liste der eigenen Partien** —
// #1600.
//
// # Eine eigene Datei, und nicht wegen der Zeilengrenze
//
// **Weil sie eine andere Frage beantwortet** (§2): `partieleitung.go` führt
// eine Partie — anlegen, lesen, befehlen, beenden. Hier steht, **welche es
// gibt**, und das ist die Frage, die jemand stellt, **bevor** er eine führt.

import (
	"context"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Partien gibt die Partien dieses Profils, zuletzt geänderte zuerst.
//
// # Warum Zeilen und keine Partien
//
// **Weil eine Liste keine Sicht ist** — der Grund steht ausführlich bei
// [domain.Partiezeile]. Wer einsteigen will, ruft danach [Partieleitung.Partie]
// mit seinem Sitz, und **dort** rechnet die Projektion.
//
// **Die Reihenfolge kommt aus dem Speicher** und wird hier nicht noch einmal
// hergestellt: zwei Sortierungen derselben Liste liefen beim ersten neuen Feld
// auseinander.
//
// # Wer fragt, und warum das Profil mitkommt
//
// **Weil „dran“ eine Frage an einen Sitz ist** (UXR-F16,
// #2031), und
// welcher Sitz der eigene ist, sagt die Bindung an das Profil. Der Speicher
// kennt das Profil aus dem Kontext; der Dienst bekommt es **genannt** und
// importiert dafür keinen Speicher (§2). Rechnen tut [eigenerSitz] — hier
// wird nur jede Partie für sich gefragt.
func (l *Partieleitung) Partien(ctx context.Context, profilID int64) (
	[]domain.Partiezeile, error,
) {
	partien, fehler := l.partien.Partien(ctx)
	if fehler != nil {
		return nil, fmt.Errorf("die Partien lesen: %w", fehler)
	}

	zeilen := make([]domain.Partiezeile, 0, len(partien))
	for _, partie := range partien {
		zeile := partie.Zeile()
		zeile.EigenerSitz, zeile.Dran = eigenerSitz(partie, profilID)

		zeilen = append(zeilen, zeile)
	}

	return zeilen, nil
}
