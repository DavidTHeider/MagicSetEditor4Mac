//go:build !darwin

package prozessspeicher

// echteQuelle gibt auf einem anderen System **„nicht gemessen"** — und nicht
// null: eine Null sähe aus wie ein Prozess, der nichts braucht.
func echteQuelle() quelle {
	return quelle{
		pids:           func() ([]int, error) { return nil, errNichtGemessen },
		pfad:           func(int) (string, error) { return "", errNichtGemessen },
		verantwortlich: func(int) (int, error) { return 0, errNichtGemessen },
		fussabdruck:    func(int) (uint64, error) { return 0, errNichtGemessen },
		lebt:           func(int) bool { return false },
	}
}
