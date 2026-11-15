// **The script console** — `seiten.karten.konsole`, split out because
// `seitenkarten.ts` reached the line limit (§2, U13 of phase 5.1, #1424).
export const konsole = {
  ueberschrift: 'Consola de scripts',
  erklaerung:
    'Una expresión de script de MSE en el contexto de la carta abierta. El entorno contiene los mismos nombres que al dibujar: «card», «set», «stylesheet» y las funciones del paquete del juego. No cambia nada.',
  ausdruck: 'Expresión',
  platzhalter: 'card.name',
  auswerten: 'Evaluar',
  ergebnis: 'Resultado',
  art: 'Tipo: {art}',
  keineKarte: 'Para esto debe haber una carta abierta.',
  zuDenKarten: 'Ir a las cartas',
  vorbehalte: 'Salvedades al construir el entorno',
  stelle: '(línea {zeile}, columna {spalte})',
  namen: 'Nombres disponibles',
  namenErklaerung:
    'Lo que una expresión puede nombrar en esta carta: la biblioteca estándar y lo que añada la plantilla. Un clic inserta el nombre.',
  beispiele: 'Ejemplos',
  beispieleImmer: 'Se ejecutan en todas partes',
  beispieleImmerErklaerung:
    'Un clic inserta la expresión. Estas no necesitan más que el propio lenguaje.',
  beispieleVorlage: 'Necesitan una plantilla',
  beispieleVorlageErklaerung:
    'El paquete de estilo añade «card» y «set» al entorno. Si no hay ninguno instalado para esta carta, la respuesta lo indica; los ejemplos de arriba se ejecutan igualmente.',
  beispielListe: {
    rechnung: 'Aritmética',
    grossSchreiben: 'Una cadena en mayúsculas',
    laenge: 'Cuántos caracteres tiene algo',
    ausschnitt: 'Un fragmento de una cadena',
    bedingung: 'Una condición',
    name: 'El nombre de esta carta',
    grossName: 'Su nombre en mayúsculas',
    regeltext: 'La longitud de su texto de reglas',
    setangabe: 'El título del set',
  },
  vervollstaendigung: 'Sugerencias',
  uebernehmen: 'Aceptar',
  namenFiltern: 'Acotar los nombres',
  namenGezeigt: 'Se muestran {gezeigt} de {ganz}: escribe para acotarlos.',
  namenLeer: 'Ningún nombre coincide.',
  kuerzel: '⌘↩ evalúa.',
  abbrechen: 'Cancelar',
  laeuft: 'En curso …',
  dauer: '{ms} ms',
  abgebrochen:
    'Cancelado: la ejecución terminó antes de que hubiera un resultado. La expresión sigue en el campo.',
  baum: 'Resultado como árbol',
  verlauf: 'Historial de esta sesión',
  verlaufErklaerung:
    'Las últimas veinte expresiones, solo de esta sesión. Un clic vuelve a cargar la expresión; lo que suene a token o contraseña aparece aquí ocultado y no se puede volver a cargar.',
  verlaufLeer: 'Todavía no se ha evaluado nada.',
  verlaufLoeschen: 'Borrar historial',
  ausgangGelungen: 'correcto',
  ausgangGescheitert: 'fallido',
  ausgangAbgebrochen: 'cancelado',
}
