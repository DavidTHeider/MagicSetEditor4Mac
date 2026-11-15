// Die Texte der **Turniere** — Werkstück 5 von Phase 4.7
// (#1294),
// nach Vermerk 0062.
//
// **Ein eigener Baustein**, wie `seitendecks.ts` und aus demselben Grund:
// `seiten.ts` steht an der Zeilengrenze aus §2.
export const turniere = {
  alleTurniere: 'Todos los torneos ({anzahl})',
  titel: 'Torneos',
  erklaerung:
    'Un torneo es un desarrollo a lo largo del tiempo: rondas, emparejamientos, resultados. La clasificación se calcula a partir de él y no se guarda en ningún sitio: si corriges un resultado, ves la nueva clasificación al instante.',

  leer: 'Todavía no hay ningún torneo. «Crear torneo …», arriba, empieza uno; «Importar torneo …» lo trae de un archivo.',

  neu: 'Crear torneo',
  neuOeffnen: 'Crear torneo …',
  name: 'Nombre',
  spiel: 'Juego',
  notiz: 'Nota',
  einzelspiel: 'Una partida por enfrentamiento',
  einzelspielHinweis:
    'Entonces el porcentaje de partidas ganadas y el de los oponentes no dicen nada: son copias de las cifras del enfrentamiento. Se calculan igualmente.',
  anlegen: 'Crear',

  // **The wizard** — DOM-04.AK-01 (#1699).
  assistent: 'Un torneo nuevo en cuatro pasos',
  schrittTurnier: 'Torneo',
  schrittRegeln: 'Reglas',
  schrittTeilnehmende: 'Participantes',
  schrittErsteRunde: 'Primera ronda',
  weiter: 'Siguiente',
  zurueck: 'Atrás',
  verlassen: 'Continuar más tarde',
  vorlage: 'Plantilla',
  ohneVorlage: 'Sin plantilla',
  vorlageHinweis:
    'Una plantilla trae juego, formato, reglas y nota. Si quieres, también los nombres, pero nunca los mazos: una lista presentada pertenece a un solo torneo.',
  mitTeilnehmenden: 'Adoptar los nombres de los jugadores',
  nameWarum:
    'Obligatorio, para que lo encuentres de nuevo en la lista. Dos torneos de un mismo juego tienen nombres distintos.',
  spielWarum: 'Obligatorio: determina qué formatos y mazos encajan.',
  format: 'Formato',
  ohneFormat: 'Sin formato',
  formatWarum: 'Opcional: puedes elegir entre los formatos de este juego.',
  regelnAusVorlage: 'Las reglas proceden de la plantilla.',
  schweizerSystem:
    'Los emparejamientos siguen el sistema suizo: quienes tienen los mismos puntos juegan entre sí, y nadie se enfrenta dos veces al mismo oponente. La aplicación sugiere cuántas rondas jugar según el número de jugadores.',
  zuWenige:
    'Una ronda necesita a dos personas que jueguen entre sí: inscribe al menos a dos.',
  ersteRundeHinweis:
    'Así quedaría la primera ronda. No se empareja nada hasta que pulses el botón.',
  ersteRundeAuslosen: 'Emparejar la ronda 1',

  waehlen: 'Abrir',
  loeschen: 'Eliminar torneo',
  handlungenAn: 'Acciones — {name}',
  handlungenKurz: '…',
  turniermenue: 'Torneo',
  loeschenFrage:
    '¿Eliminar este torneo? Con él se van las rondas, los emparejamientos y los resultados. Los mazos de los jugadores se conservan.',
  loeschenJa: 'Eliminar',
  loeschenNein: 'Cancelar',

  teilnehmer: 'Participantes',
  teilnehmerLeer: 'Todavía no se ha inscrito nadie.',
  anmelden: 'Inscribir',
  gebiete: 'Áreas de este torneo',
  gebietTeilnehmer: 'Participantes',
  gebietRunden: 'Rondas',
  gebietRangliste: 'Clasificación',

  // **Closing and log** — DOM-04.AK-04 (#1699).
  abschliessen: 'Cerrar',
  abgeschlossenAm:
    'Cerrado el {datum}. Los jugadores, las rondas y los resultados son definitivos.',
  wiedereroeffnen: 'Reabrir',
  nichtWiedereroeffnen: 'No reabrir',
  wiedereroeffnenFrage: '¿Reabrir este torneo?',
  wiedereroeffnenErklaerung:
    'Después se podrán volver a cambiar jugadores, rondas y resultados. El registro deja constancia de que ocurrió.',
  gebietProtokoll: 'Registro',
  protokollLeer: 'Todavía no ha pasado nada.',
  protokollAngelegt: 'Creado como «{angaben}»',
  protokollAngemeldet: 'Inscripción de {angaben}',
  protokollAusgestiegen: 'Baja de {angaben}',
  protokollEingestiegen: 'Nueva inscripción de {angaben}',
  protokollRunde: 'Ronda {angaben} emparejada',
  protokollErgebnis: 'Resultado {paarung}: {angaben}',
  protokollZurueckgenommen: 'Resultado {paarung} retirado',
  protokollAbgeschlossen: 'Cerrado',
  protokollWiedereroeffnet: 'Reabierto',
  protokollEingelesen: 'Importado como «{angaben}»',

  // **File and report** — DOM-04.AK-03 (#1699).
  ausgeben: 'Exportar …',
  berichtSichern: 'Guardar informe …',
  einlesen: 'Importar torneo …',
  gesichertUnter: 'Guardado en {pfad}.',
  eingelesen: 'Importado y abierto.',
  einlesebefunde: 'Al importar se observó:',

  // **Header and links after the artboard** — DOM-04.AK-02 (#1699).
  kopfzeile: 'Sistema suizo · Ronda {runde} de {von} · {anzahl} jugadores',
  kopfzeileOhneRunde: 'Sistema suizo · {anzahl} jugadores',
  kopfzeileOhneEmpfehlung: 'Sistema suizo · Ronda {runde} · {anzahl} jugadores',
  gebietDecklisten: 'Listas de mazo',
  decklistenLeer:
    'Todavía nadie ha presentado una lista: al elegir un mazo al inscribirse, se presenta su lista.',
  deckAufschlagen: 'Abrir mazo',
  partieOeffnen: 'Abrir partida',
  deck: 'Mazo',
  ohneDeck: 'Sin mazo',
  deckHinweis:
    'Si se indica un mazo, su lista se copia en lugar de referenciarse: una lista de mazo en un torneo es un documento presentado.',
  deckabzug: 'Lista presentada',
  deckabzugLeer: 'No se ha presentado ningún mazo.',
  deckabzugUnbenannt: 'Carta',
  aussteigen: 'Darse de baja',
  einsteigen: 'Volver a inscribirse',
  ausgestiegen: 'de baja',
  ausgestiegenHinweis:
    'Ya no se le empareja y sigue en la clasificación: su porcentaje aún cuenta para el de sus oponentes.',

  runden: 'Rondas',
  rundenLeer: 'Todavía no se ha emparejado ninguna ronda.',
  runde: 'Ronda {nummer}',
  empfehlung:
    '{gelegt} de {empfohlen} rondas: las reglas de torneo recomiendan {empfohlen} para {spieler} jugadores. Una sugerencia, no una norma.',
  vorschlag: 'Previsualizar la próxima ronda',
  vorschlagTitel: 'Así quedaría la próxima ronda',
  vorschlagHinweis:
    'Solo previsualizada, no creada. Únicamente el botón de abajo la crea.',
  rundeLegen: 'Crear ronda',
  bye: 'Bye (2–0)',
  gegen: 'contra',
  offen: 'aún sin jugar',
  spiele: 'Partidas',
  spieleA: 'Partidas del primero',
  spieleB: 'Partidas del segundo',
  unentschiedenSpiele: 'Empates',
  eintragen: 'Registrar',
  zuruecknehmen: 'Deshacer',
  ergebnis: 'Resultado',

  // **The effects of a correction** — DOM-04.AK-02 (#1699).
  folgenFrage: 'Esta corrección cambia la clasificación.',
  folgenPlatz:
    '{name}: puesto {vorher} → {nachher}, {punkteVorher} → {punkteNachher} puntos',
  folgenSpaeter:
    'La ronda {runden} se emparejó con la clasificación antigua y se queda como está.',
  folgenEintragen: 'Registrar igualmente',
  folgenNicht: 'No registrar',
  ergebnisZahlen:
    'Las partidas se cuentan con números enteros a partir de cero.',
  ergebnisOhneSpiel:
    'Un resultado sin ni una sola partida no es un resultado: quien no ha jugado todavía no tiene ninguno.',
  ergebnisZuViele:
    'Aquí un enfrentamiento dura como máximo {hoechstens} partidas.',

  rangliste: 'Clasificación',
  ranglisteLeer: 'Todavía no se ha jugado nada.',
  platz: 'Puesto',
  bilanz: 'Balance',
  matchpunkte: 'Puntos',
  gegnermatchquote: 'OMW',
  spielquote: 'GW',
  gegnerspielquote: 'OGW',
  quotenHinweis:
    'Cuatro desempates en el orden de las Magic Tournament Rules: puntos, porcentaje de enfrentamientos ganados de los oponentes, porcentaje de partidas ganadas y porcentaje de partidas ganadas de los oponentes. Cada porcentaje tiene un mínimo de 0,33.',

  // **Eine Paarung, an eine Partie gebunden** (G7, #1464).
  match: 'Jugado aquí',
  matchErklaerung:
    'Si este emparejamiento se juega en este programa, vincúlalo a la partida. El resultado lo comunican entonces ambas partes y solo se registra cuando las dos dicen lo mismo.',
  matchNichtGebunden: 'Este emparejamiento no tiene ninguna partida asociada.',
  matchAnsehen: 'Comprobar',
  matchPartie: 'ID de la partida',
  matchSitzA: 'Asiento del primer jugador',
  matchSitzB: 'Asiento del segundo jugador',
  matchBinden: 'Vincular a esta partida',
  matchGebunden: 'Vinculado a la partida',
  matchSitz: 'Asiento',
  matchDeckabdruck: 'Huella del mazo',
  matchKeineMeldung: 'Todavía nadie ha comunicado nada.',
  matchMeldungVon: 'comunicado por',
  matchMelden: 'Comunicar para este asiento',
  matchAusDerPartie: 'Comunicar el resultado desde la partida',
  matchAusDerPartieHilfe:
    'El marcador y la prueba proceden de la partida terminada: solo tienes que decir quién lo comunica.',
  matchNachweis: 'Prueba (último eslabón de la cadena)',
  matchUebernehmen: 'Registrar resultado',
  matchUebernommen: 'El resultado está en el torneo.',
  matchGrenze:
    'La prueba dice que ambas partes vieron el mismo desarrollo. No dice que se jugara correctamente, y una ronda privada no es una homologación oficial de torneo.',
}
