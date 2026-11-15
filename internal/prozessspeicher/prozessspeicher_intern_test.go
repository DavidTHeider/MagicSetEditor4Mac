package prozessspeicher

import (
	"errors"
	"slices"
	"testing"
)

// erfunden ist eine Quelle mit festen Prozessen.
func erfunden(pfade map[int]string, verantwortliche map[int]int, bytes map[int]uint64) quelle {
	return quelle{
		pids: func() ([]int, error) {
			alle := make([]int, 0, len(pfade))
			for pid := range pfade {
				alle = append(alle, pid)
			}

			slices.Sort(alle)

			return alle, nil
		},
		pfad: func(pid int) (string, error) {
			if pfade[pid] == "" {
				return "", errors.New("beendet")
			}

			return pfade[pid], nil
		},
		verantwortlich: func(pid int) (int, error) { return verantwortliche[pid], nil },
		fussabdruck: func(pid int) (uint64, error) {
			if bytes[pid] == 0 {
				return 0, errors.New("beendet")
			}

			return bytes[pid], nil
		},
		// **Wer sich nicht messen ließ, ist verschwunden** — es sei denn,
		// ein Lauf sagt anderes.
		lebt: func(int) bool { return false },
	}
}

const webkit = "/System/Library/Frameworks/WebKit.framework/Versions/A/XPCServices/"

// **Zu dieser Anwendung gehört, wofür sie verantwortlich ist** — und nur das.
func TestDieWebprozesseGehoerenIhremVerantwortlichen(t *testing.T) {
	t.Parallel()

	pfade := map[int]string{
		1: "/Programme/MagicSetEditor4Mac.app/Contents/MacOS/MagicSetEditor4Mac",
		2: webkit + "com.apple.WebKit.WebContent.xpc/Contents/MacOS/com.apple.WebKit.WebContent",
		3: webkit + "com.apple.WebKit.GPU.xpc/Contents/MacOS/com.apple.WebKit.GPU",
		4: webkit + "com.apple.WebKit.WebContent.xpc/Contents/MacOS/com.apple.WebKit.WebContent",
		5: webkit + "com.apple.WebKit.Networking.xpc/Contents/MacOS/com.apple.WebKit.Networking",
		6: "/usr/bin/true",
		8: webkit + "com.apple.WebKit.WebContent.xpc/Contents/MacOS/com.apple.WebKit.WebContent.EnhancedSecurity",
	}
	// **6 ist ein eigenes Kind, aber kein WebKit** — es zählt nicht.
	verantwortliche := map[int]int{2: 1, 3: 1, 4: 99, 5: 1, 6: 1, 8: 1}
	bytes := map[int]uint64{1: 100, 2: 500, 3: 50, 4: 700, 8: 30} // 5 endet vor dem Messen

	stand, err := messenUeber(erfunden(pfade, verantwortliche, bytes), 1)
	if err != nil {
		t.Fatal(err)
	}

	erwartet := []Prozess{{2, ArtWebinhalt, 500}, {3, ArtGrafik, 50}, {8, ArtWebinhalt, 30}}
	if stand.Anwendung != (Prozess{1, ArtAnwendung, 100}) || !slices.Equal(stand.Web, erwartet) || stand.Ungemessen != "" {
		t.Errorf("Stand %+v, erwartet Web %+v", stand, erwartet)
	}
}

// **Was sich nicht fragen lässt, wird genannt** — und der eigene Speicher
// steht trotzdem da.
func TestEinUngemessenerWebspeicherWirdGenannt(t *testing.T) {
	t.Parallel()

	ohnePids := erfunden(map[int]string{}, nil, map[int]uint64{1: 100})
	ohnePids.pids = func() ([]int, error) { return nil, errors.New("keine Liste") }

	if stand, err := messenUeber(ohnePids, 1); err != nil || stand.Ungemessen != "keine Liste" || stand.Anwendung.FussabdruckBytes != 100 {
		t.Errorf("ohne Liste: %+v, %v", stand, err)
	}

	ohneVerantwortlichen := erfunden(map[int]string{2: webkit + "com.apple.WebKit.GPU"}, nil, map[int]uint64{1: 100})
	ohneVerantwortlichen.verantwortlich = func(int) (int, error) { return 0, errNichtGemessen }

	if stand, err := messenUeber(ohneVerantwortlichen, 1); err != nil || stand.Ungemessen != errNichtGemessen.Error() {
		t.Errorf("ohne Verantwortlichen: %+v, %v", stand, err)
	}

	if _, err := messenUeber(erfunden(nil, nil, nil), 1); err == nil {
		t.Error("ohne eigenen Fußabdruck kein Fehler")
	}
}

// **Ein Fehler, der jeden trifft, ist nicht „keine Webansicht"** (Zweitblick
// über Phase 4, Befund 9): fehlte `proc_pidpath`, fing der Riegel die Panik,
// die Schleife verwarf den Fehler, und die Oberfläche las eine leere Liste
// ohne Fehler als fehlende Webansicht. **Und ein Webprozess, der lebt und sich
// nicht messen lässt, macht die Messung unvollständig** — nur ein
// verschwundener zählt nicht.
func TestEinSystematischerMessfehlerIstNichtKeineWebansicht(t *testing.T) {
	t.Parallel()

	ohnePfad := erfunden(map[int]string{2: webkit + "com.apple.WebKit.GPU"}, map[int]int{2: 1}, map[int]uint64{1: 100, 2: 5})
	ohnePfad.pfad = func(int) (string, error) { return "", errNichtGemessen }

	if stand, err := messenUeber(ohnePfad, 1); err != nil || stand.Ungemessen != errNichtGemessen.Error() {
		t.Errorf("ohne Programmpfad: %+v, %v", stand, err)
	}

	pfade := map[int]string{
		2: webkit + "com.apple.WebKit.GPU.xpc/Contents/MacOS/com.apple.WebKit.GPU",
		3: webkit + "com.apple.WebKit.WebContent.xpc/Contents/MacOS/com.apple.WebKit.WebContent",
	}
	lebend := erfunden(pfade, map[int]int{2: 1, 3: 1}, map[int]uint64{1: 100, 2: 50})
	lebend.lebt = func(pid int) bool { return pid == 3 }

	stand, err := messenUeber(lebend, 1)
	if err != nil || len(stand.Web) != 1 || stand.Web[0].PID != 2 || stand.Ungemessen == "" {
		t.Errorf("ein lebender ohne Messung: %+v, %v", stand, err)
	}

	ohneFussabdruck := erfunden(pfade, map[int]int{2: 1, 3: 1}, map[int]uint64{1: 100})
	ohneFussabdruck.fussabdruck = func(pid int) (uint64, error) {
		if pid == 1 {
			return 100, nil
		}

		return 0, errNichtGemessen
	}

	if stand, err := messenUeber(ohneFussabdruck, 1); err != nil || stand.Ungemessen != errNichtGemessen.Error() {
		t.Errorf("ohne Fußabdruck der Webprozesse: %+v, %v", stand, err)
	}
}

// **Auch ein gewöhnlicher Pfadfehler eines eigenen Kindes ist eine Lücke**
// (Nachlauf über Phase 4, Befund 10): nur der Riegelfehler machte ihn
// sichtbar; lieferte `proc_pidpath` für einen lebenden Webprozess keinen Pfad,
// blieben Liste und `Ungemessen` leer. Gefragt wird deshalb zuerst, wer
// verantwortlich ist — ein fremder Prozess ohne Pfad bleibt kein Befund.
func TestEinPfadfehlerEinesEigenenKindesIstEineLuecke(t *testing.T) {
	t.Parallel()

	// 3 ist ein eigenes Kind ohne Pfad, 4 ein fremder ohne Pfad.
	quelle := erfunden(map[int]string{3: "", 4: ""}, map[int]int{3: 1, 4: 99}, map[int]uint64{1: 100})
	quelle.pids = func() ([]int, error) { return []int{1, 3, 4}, nil }
	quelle.lebt = func(pid int) bool { return pid == 3 || pid == 4 }

	if stand, err := messenUeber(quelle, 1); err != nil || stand.Ungemessen == "" || len(stand.Web) != 0 {
		t.Errorf("ein lebendes Kind ohne Pfad: %+v, %v", stand, err)
	}

	// **Ein verschwundenes Kind und ein fremder Prozess ohne Pfad sind keine.**
	quelle.lebt = func(pid int) bool { return pid == 4 }

	if stand, err := messenUeber(quelle, 1); err != nil || stand.Ungemessen != "" {
		t.Errorf("verschwunden oder fremd: %+v, %v", stand, err)
	}
}
