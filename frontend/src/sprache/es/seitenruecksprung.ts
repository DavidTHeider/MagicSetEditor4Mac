// The **going back** texts — work item O7 of phase 5.2
// (#1431).

export const ruecksprung = {
  titel: 'Volver atrás',
  hinweis:
    'Cuatro cosas pueden volver atrás, y cuestan cosas distintas. Dos de ellas ' +
    'esta aplicación no puede hacerlas por sí misma; entonces aquí se indica quién lo hace.',

  arten: {
    appfassung: 'La aplicación',
    schema: 'La estructura de la biblioteca',
    daten: 'Los datos',
    pakete: 'Los paquetes de plantillas',
  } as Record<string, string>,

  hindernisse: {
    nachProzessende:
      'Una persona sustituye el paquete en el Finder cuando esta aplicación ' +
      'ya no se está ejecutando. Un programa que se sustituye a sí mismo mientras ' +
      'se ejecuta necesitaría un segundo proceso con permiso de escritura en «Aplicaciones».',
    nurVorwaerts:
      'No hay camino hacia atrás a través de las migraciones. El camino de vuelta ' +
      'es una copia de seguridad: restaura el estado de entonces.',
    unbekannteArt: 'La aplicación no conoce este tipo.',
  } as Record<string, string>,

  gehtHier: 'Esta aplicación puede hacerlo por sí misma.',

  vergleichTitel: 'Lo que cuesta restaurar',
  datenstand: 'Estado restaurado',
  spalteGegenstand: 'Elemento',
  spalteSicherung: 'En la copia de seguridad',
  spalteJetzt: 'Ahora',
  spalteVerlust: 'Se perderá',

  gegenstaende: {
    sets: 'Sets',
    karten: 'Cartas',
    decks: 'Mazos',
    exemplare: 'Ejemplares',
  } as Record<string, string>,

  unbekannt: '—',

  vorherAusfahren:
    'Lo que aquí se pierde puedes exportarlo antes: un set con «Guardar como», la colección con «Exportar».',

  unklarTitel: 'Sin determinar después',
  unklarHinweis:
    'Estos ejemplares volverán a estar en tu colección tras la restauración. ' +
    'Nadie sabe si aún los tienes: pueden haberse entregado desde entonces. ' +
    'Quedan fuera del intercambio hasta que lo hayas comprobado; ' +
    'no se descuenta nada.',
  unklarZeile: '{menge} × {name}',
  ohneNamen: '{menge} × {schluessel}',

  handpruefung:
    'No se pudo obtener la comparación de las colecciones: esta copia de seguridad ' +
    'o tu biblioteca no la proporciona. Revisa tu colección a mano después de ' +
    'restaurar.',

  rueckkehrenTitel: 'Restauraciones registradas',
  keineRueckkehren:
    'Todavía ninguna. En cuanto restaures una copia de seguridad, aparecerá aquí, ' +
    'con el estado al que volvió y lo que dejó sin determinar.',
  rueckkehrZeile: 'El {zurueckgespielt}, de vuelta al estado del {datenstand}',
  ausDatei: 'desde {quelle}',
  offen: 'Pendiente',
  quittiert: 'Comprobado el {geklaertAm}',
  betrifft: 'afecta a {menge} ejemplares',
  quittieren: 'Comprobado',
  quittierenLaeuft: 'Registrando …',
  quittierenHinweis:
    'Con esto indicas que has revisado a mano los ejemplares afectados. ' +
    'No descuenta nada.',
}
