// The **data hygiene** texts — work item O8 of phase 5.2
// (#1432).

export const datenpflege = {
  titel: 'Limpieza',
  hinweis:
    'Con el tiempo, una biblioteca acumula filas que ya no apuntan a nada. ' +
    'Aquí se indica qué son y cuáles de ellas se pueden eliminar sin riesgo.',

  nachsehen: 'Revisar',
  nachsehenLaeuft: 'Revisando …',
  sauber: 'No se ha encontrado nada. La biblioteca está en orden.',
  gueltigbis: 'Este resumen es válido hasta {zeit}.',

  arten: {
    kartenwerteOhneKarte: 'Valores de campo sin su carta',
    setangabenOhneSet: 'Datos de set sin su set',
    suchzeilenOhneKarte: 'Filas de búsqueda sin su carta',
    deckplaetzeOhneDeck: 'Posiciones de mazo sin su mazo',
    zuordnungenOhneExemplar: 'Asignaciones de mazo sin su ejemplar',
    verlaufOhneSet: 'Entradas del historial de sets que ya no existen',
    abgelaufeneNachweise: 'Comprobantes de autorización caducados',
    abgelaufeneToken: 'Tokens de acceso caducados',
    moeglicheDubletten: 'Parecen duplicados',
    offeneRueckkehren: 'Restauraciones abiertas',
  } as Record<string, string>,

  klassen: {
    sicher: 'Se puede eliminar sin riesgo',
    entscheidung: 'Decides tú',
    nurInformieren: 'No se tocará',
  } as Record<string, string>,

  klassenhinweise: {
    sicher:
      'Estas filas no apuntan a nada y no guardan nada. Están seleccionadas ' +
      'de antemano.',
    entscheidung:
      'Parece un resto, pero puede ser un comprobante. Mira los ejemplos ' +
      'antes de elegirlo.',
    nurInformieren:
      'Esto no se elimina. Los números de carta iguales no son duplicados: ' +
      'el estado, el idioma, el acabado, la adquisición y el lugar de ' +
      'almacenamiento los distinguen; y una restauración abierta es un ' +
      'comprobante.',
  } as Record<string, string>,

  anzahl: '{anzahl} lugares de la biblioteca',
  beispieleTitel: 'Por ejemplo',

  aufraeumen: 'Limpiar lo seleccionado',
  aufraeumenLaeuft: 'Limpiando …',
  aufraeumenHinweis:
    'Antes se hace una copia de seguridad de la biblioteca. Después se vuelve a revisar.',

  bilanzTitel: 'Qué ha pasado',
  gesichert: 'Copia de seguridad guardada en {ort}',

  ausgaenge: {
    erledigt: 'Eliminado',
    konflikt: 'Ha cambiado desde el resumen; no se ha tocado',
    schonWeg: 'Ya no estaba',
    uebersprungen: 'Omitido',
    gescheitert: 'No ha funcionado',
    bereit: 'Listo',
  } as Record<string, string>,

  bilanzZeile: '{art}: {ausgang} ({anzahl})',
  bilanzStaende: 'Antes {vorher} hallazgos, después {nachher}.',
}
