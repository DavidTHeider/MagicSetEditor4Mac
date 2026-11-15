package main

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

// grenzenauskunft reicht die Grenzen des Prozesses an den Dienst (#1429) —
// **die beim Start gelesenen**: nach ihnen bucht er (QA-08.AK-03).
//
// **Gebaut von `rest.GrenzenVon` und nicht Feld für Feld hier** (#2003): eine
// Abschrift von Hand ließ ein neues Feld still fallen, und die Auskunft des
// Dienstes sagte weniger als die der anderen Flächen.
type grenzenauskunft struct {
	grenzen rest.Grenzen
}

func (g grenzenauskunft) Grenzen(context.Context) (rest.Grenzen, error) {
	return g.grenzen, nil
}
