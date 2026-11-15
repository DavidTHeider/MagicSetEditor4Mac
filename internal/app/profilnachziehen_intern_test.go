package app

// **Die zwei stillen Wege von `profilNachziehen`** — Nachlauf zum Zweitblick
// über die Phasen 4.0 bis 4.13 (2026-09-07), Befund 1.
//
// # Was der Weg tut
//
// Nach dem Löschen eines Profils kennt das Fenster den neuen Zeiger nicht: der
// Fachkern hat ihn in derselben Transaktion gesetzt. `profilNachziehen` liest
// ihn nach und stellt darauf um — Kennung **und** Leserverbindung, unter der
// Umschaltsperre.
//
// # Warum diese zwei Zweige geprüft gehören
//
// **Beide enden mit `return` und schweigen** — das eine Mal absichtlich, das
// andere Mal mit einer Zeile im Protokoll. Ein stiller Weg, den kein Lauf
// betritt, ist der, bei dem später niemand merkt, dass er sich geändert hat.
//
// **Paketintern**, weil `profilNachziehen` und `profil` es sind.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Ohne Bestand geschieht nichts** — der Fall der drei Flächen ohne Fenster.
//
// Sie bauen ihre Anwendung über `NeueApp("")` und reichen den Bestand je Naht
// herein; ihr Profil steht im Wurzelkontext und nicht in diesem Feld. Ohne
// diese Wache stürzte `ProfilLoeschen` an jeder von ihnen ab — gefunden vom
// Flächenlauf zu Werkstück 5 (#1351), und dieselbe Wache steht in
// `profilLaden` nebenan.
func TestOhneBestandZiehtNichtsNach(t *testing.T) {
	t.Parallel()

	anwendung := &App{}
	anwendung.profil.Store(7)

	anwendung.profilNachziehen(t.Context())

	if gesehen := anwendung.profil.Load(); gesehen != 7 {
		t.Errorf("Profil %d — ohne Bestand darf sich nichts ändern", gesehen)
	}
}

// **Ein unlesbarer Zeiger hält das Löschen nicht auf.**
//
// Dieselbe Zusage wie beim Start (`profilLaden`): das Profil ist gelöscht, und
// das bleibt es. Was nicht gelingt, ist das Nachziehen — es wird protokolliert
// und nicht zurückgenommen (§7, Regel 3).
func TestEinUnlesbarerZeigerHaeltDasNachziehenNichtAuf(t *testing.T) {
	t.Parallel()

	db, err := storage.FuerLaeufe("file:" + t.TempDir() + "/zu.db")
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	if schliessfehler := db.Close(); schliessfehler != nil {
		t.Fatalf("schließen: %v", schliessfehler)
	}

	anwendung := &App{db: db}
	anwendung.profil.Store(7)

	anwendung.profilNachziehen(t.Context())

	// **Der alte Wert bleibt stehen.** Eine Null einzutragen sähe aus wie ein
	// Profil und wäre keines — dieselbe Überlegung wie bei `profilLaden`.
	if gesehen := anwendung.profil.Load(); gesehen != 7 {
		t.Errorf("Profil %d — ein Lesefehler darf den Zeiger nicht bewegen",
			gesehen)
	}
}

// **Und `profilLaden` schweigt ohne Bestand ebenso.**
//
// # Warum dieser Test heute dazukam
//
// **Er hatte seinen Aufrufer verloren.** `ProfilLoeschen` rief `profilLaden`,
// und der Flächenlauf zu Werkstück 5 fuhr ihn über `mse profil <id>` — dort
// ist `a.db` leer, und ohne die Wache stürzte der Befehl ab. Seit dem
// Nachlauf zum Zweitblick über die Phasen 4.0 bis 4.13 geht `ProfilLoeschen`
// durch `profilNachziehen`, und die Wache in `profilLaden` stand ohne Lauf da.
//
// **Eine Wache ohne Lauf ist eine Wache, von der niemand weiß, ob sie noch
// hält** — und die Abdeckung hat es angestrichen (85,7 Prozent).
func TestOhneBestandLaedtNichtsNach(t *testing.T) {
	t.Parallel()

	anwendung := &App{}
	anwendung.profil.Store(7)

	anwendung.profilLaden(t.Context())

	if gesehen := anwendung.profil.Load(); gesehen != 7 {
		t.Errorf("Profil %d — ohne Bestand darf sich nichts ändern", gesehen)
	}
}
