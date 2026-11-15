// The texts of the **Play** area — work item G1 of phase 5.5
// (#1458).
import { abzugworte } from './spielenabzug'
import { zugworte } from './spielenzug'
import { menueworte } from './spielenmenue'
import { vorhangworte } from './spielenvorhang'
import { partienwoerter } from './spielenpartien'
import { regelwoerter } from './spielenregeln'
import { anfangworte } from './spielenanfang'
import { beginnworte } from './spielenbeginn'
import { gastworte } from './spielengast'
import { netzworte } from './spielennetz'
import { leihewoerter } from './spielenleihe'

export const spielen = {
  ...abzugworte,
  ...zugworte,
  ...menueworte,
  ...vorhangworte,
  ...partienwoerter,
  ...regelwoerter,
  ...anfangworte,
  ...beginnworte,
  ...gastworte,
  ...netzworte,
  ...leihewoerter,

  titel: 'Jugar',
  erklaerung:
    'Una partida es una mesa con mazos congelados. Lo que ocurre aquí queda fijado como un flujo de eventos: cada uno depende del anterior y ninguno se sobrescribe. Quien mueve una carta ve de inmediato lo que ven los demás de ella.',
  anfangTitel: 'Empezar una partida',
  anfangErklaerung:
    'Cada asiento aporta un mazo, y se congela: si después cambias tu mazo, esta partida no cambia. El juego viene del primer mazo elegido.',
  leerDecksTitel: 'Aún no hay mazo',
  leerDecks: 'Para jugar hace falta un mazo. Crea uno en Mazos.',
  sitz: 'Asiento',
  sitzName: 'Nombre en la mesa',
  sitzDeck: 'Mazo',
  sitzLeihe: 'Prestado por …{von} · {karten} cartas',
  deckWaehlen: 'Elegir un mazo',
  anschlussWartet:
    '«{name}» ha llegado y ocupa el primer asiento. El segundo está libre.',
  sitzHinzu: 'Añadir un asiento',
  sitzWeg: 'Quitar este asiento',
  sitzGruppe: 'Asiento {nummer}',
  spiel: 'Juego',
  anfangen: 'Empezar la partida',
  tischTitel: 'En la mesa',
  partieGegen: 'Partida contra {namen}',
  partieZwischen: 'Partida entre {namen}',
  partieVon: 'Partida en solitario: {namen}',
  fassung: 'Revisión',
  wirt: 'Anfitrión',
  wirtHinweis:
    'El anfitrión guarda el estado completo de esta partida y por eso ve más de lo que muestra su propia vista. Se dice aquí porque conviene saberlo antes de unirse, no después.',
  tischAufHaufen: '{zahl} en un montón',
  tischMarkeEine: '{zahl} contador {art}',
  tischMarkeViele: '{zahl} contadores {art}',
  tischArt: {
    spielstein: 'Ficha',
    emblem: 'Emblema',
    faehigkeit: 'Habilidad en la pila',
  },

  sicht: 'Vista de',
  sichtZuschauer: 'Espectador (sin asiento)',
  sichtWechseln: 'Cambiar de asiento',
  sichtLaedt: 'Cargando la vista elegida …',
  verlassen: 'Cerrar la partida',
  zonen: 'Qué hay dónde',
  zoneBibliothek: 'Biblioteca',
  zoneHand: 'Mano',
  zoneTisch: 'Campo de batalla',
  zoneFriedhof: 'Cementerio',
  zoneExil: 'Exilio',
  zoneStapel: 'Pila',
  zoneKommando: 'Zona de mando',
  zonekurzBibliothek: 'Bibl.',
  zonekurzFriedhof: 'Cem.',
  zonekurzExil: 'Exilio',
  zonekurzKommando: 'Mando',
  zoneLeer: 'vacía',

  unbekannterTyp: 'Tipo desconocido: {woerter}',
  geheim: 'boca abajo',
  getappt: 'girada',
  befehle: 'Acciones',
  regelnTitel: 'Qué se hace cumplir',
  regelwerk: 'Conjunto de reglas',
  wirksamAb: 'vigente desde',
  hoechsteStrenge: 'Rigor máximo',
  strenge: 'Rigor',
  andereSeite: 'Otra cara',
  grossAnsehen: 'Ver en grande',
  ersteSeite: 'Primera cara',
  stelle: 'A la posición',
  stelleOben: 'arriba',
  stelleUnten: 'abajo',
  stelleZufaellig: 'barajada dentro',
  mischpflichtTitel: 'Barajados obligatorios pendientes',

  legendeTitel: 'Elección pendiente: la regla de leyendas',
  legendeSatz:
    '{wer} debe elegir qué permanente legendario llamado «{name}» se queda (CR 704.5j).',
  legendeBehalten: 'Conservar este',

  reihungTitel: 'Elección pendiente: el orden de las habilidades',
  reihungSatz:
    '{wer} controla varias habilidades que se han disparado a la vez y indica en qué orden van a la pila (CR 603.3b). Lo que se pone primero queda abajo y se resuelve el último.',
  reihungLegen: 'Poner a continuación:',
  stand: 'Estado',
  standLaeuft: 'en curso',
  standPausiert: 'en pausa',
  standBeendet: 'terminada',
  pausieren: 'Pausar',
  fortsetzen: 'Reanudar',
  beenden: 'Terminar',
  beendenFrage:
    '¿Terminar esta partida? Terminar es definitivo: una partida terminada no se puede reanudar.',
  beendenJa: 'Terminar',
  beendenNein: 'Cancelar',
  ereignisse: 'Lo último que ha pasado',
  ereignisseLeer: 'Aún nada.',
  ereignisseLuecke:
    'Puede que falten jugadas aquí: mientras la página estaba oculta llegaron más avisos de los que se pueden recuperar.',
  ereignisPrivat: 'Solo para ti',
  ereignisSatz: '{sitz}: {handlung} — {objekt}',
  ereignisSatzOhneKarte: '{sitz}: {handlung}',

  // **Die Rückschau** (G5, #1462).
  verlauf: 'Lo que ha pasado hasta ahora',
  verlaufAnsehen: 'Ver historial',
  verlaufNachladen: 'Recargar',
  verlaufLaeuft: 'Obteniendo …',
  verlaufUngeholt:
    'El historial aún no se ha obtenido. Muestra cada jugada de esta partida, tal como este asiento puede verla.',
  verlaufLeer: 'En esta partida aún no ha pasado nada.',
  verlaufGrenze:
    'Lo que otro asiento vio solo para sí no está aquí, ni siquiera después.',

  // **Was von einer Partie übrig bleibt** (G5, #1462).
  protokoll: 'Lo que queda de esta partida',
  protokollErklaerung:
    'Cada jugada de esta partida está guardada en este ordenador, y cada asiento ve su parte. Puedes liberar tu parte para más adelante: para un único oponente o para todos.',
  protokollAnsehen: 'Ver liberaciones',
  protokollNiemand: 'Nadie ha liberado nada.',
  protokollFuer: 'Liberado para',
  protokollAlle: 'todos',
  protokollSitz: 'Asiento',
  protokollEmpfaenger: 'Destinatarios, separados por comas, o * para todos',
  protokollFreigeben: 'Liberar',
  protokollZuruecknehmen: 'Retirar',
  protokollNurSpaeter:
    'Una liberación solo surte efecto después de la partida. En la mesa cada uno sigue viendo únicamente lo suyo.',
  protokollKeineRueckholung:
    'Retirarla cierra la puerta, pero no recupera lo que alguien ya ha leído.',
  protokollAufbewahrung:
    'Un registro terminado se ofrece para su revisión en el mantenimiento de datos a los 90 días. Solo se elimina cuando tú lo dices.',
  protokollLoeschen: 'Eliminar registro',
  protokollLoeschenErklaerung:
    'Con esto se eliminan de este ordenador el estado, el historial y las liberaciones de esta partida. No dice nada sobre copias en otros sitios.',
  protokollGeloescht: 'El registro se ha eliminado.',
  protokollLoeschenFrage: '¿Eliminar el registro de esta partida?',
  protokollLoeschenJa: 'Eliminar registro',

  // **Die Diagnose der Leitung** (G5, #1462).
  diagnose: 'Estado de la conexión',
  diagnoseErklaerung:
    'Muestra por dónde habla esta mesa y cuánto está esperando. Nada de esto dice algo sobre cartas, ni las tuyas ni las de otros.',
  diagnoseAnsehen: 'Comprobar la conexión',
  diagnoseStill: 'Ahora mismo no hay ninguna mesa en marcha.',
  diagnoseAlsWirt: 'Este ordenador hospeda la mesa.',
  diagnoseAlsGast: 'Este ordenador está sentado en la mesa de otra persona.',
  diagnosePausiert: 'La conexión se ha cortado: el asiento se conserva.',
  diagnoseWeg: 'Ruta',
  diagnoseGegenueber: 'Interlocutor',
  diagnoseSeit: 'Activa desde hace',
  diagnoseWartend: 'En espera',
  diagnoseVerworfen: 'Descartados',
  diagnoseStand: 'En la secuencia',
  diagnoseRunde: 'Último viaje de ida y vuelta',
  diagnoseBildzeit: 'Intervalo entre fotogramas en esta ventana',
  diagnoseUnbekannt: 'desconocido',
  diagnoseOhneLeitung: 'No hay ninguna conexión activa.',

  // **Die Freigabe für Gäste** (G5, #1462).
  freigabe: 'Dejar que otros jueguen en esta mesa',
  freigabeErklaerung:
    'Ser accesible y dejar jugar son dos decisiones. Abrir la mesa permite a un interlocutor hacer jugadas en esta partida: los asientos los repartes tú, no él.',
  freigabeOffen: 'Esta mesa está abierta.',
  freigabeZu: 'Esta mesa no está abierta.',
  freigabeStarten: 'Abrir mesa',
  freigabeBeenden: 'Cerrar mesa',
  freigabeNachsehen: 'Comprobar de nuevo',
  freigabeSitze: 'Libres para',
  freigabeNiemand: 'Aún no se ha sentado nadie.',
  freigabePartner: 'Es',
  freigabePartnerFremd: 'Este interlocutor no está entre tus contactos.',
  freigabeBesetzt: 'Sentados:',
  freigabeOhneHorchposten:
    'Para que llegue un interlocutor, este ordenador también tiene que ser accesible; eso se configura en Ajustes.',

  // **Der gezeichnete Tisch** — Werkstück G3 von Phase 5.5
  // (#1460).
  tischBild: 'La mesa',
  tischListe: 'Como lista',
  tischGrafisch: 'Como mesa',
  tischDarstellung: 'Presentación',
  tischEinpassen: 'Mostrar todo',
  tischNaeher: 'Acercar',
  tischWeiter: 'Alejar',
  tischDrehen: 'Rotar la mesa',
  tischAuswahl: 'A la selección',
  tischEigenerSitz: 'A mi asiento',
  tischBlick: 'Vista de la mesa',
  tischZiel: 'Ir a',
  tischOhneWebgl:
    'Este ordenador no ofrece gráficos acelerados. La mesa aparece debajo como lista, y allí funciona todo lo que funciona en la mesa dibujada.',
  tischOhneRenderer:
    'No se pudo dibujar la mesa. La lista de abajo contiene todas las cartas, y allí funciona todo lo demás.',
  tischVerloren:
    'Se ha perdido la conexión gráfica. La mesa volverá cuando se vuelva a abrir la ventana; hasta entonces, la lista de abajo contiene todas las cartas.',
  tischHeimatlos:
    'Estos objetos están en una zona que este programa no conoce y por eso no aparecen en la mesa:',
  tischVerdeckt: 'boca abajo',
  tischGetappt: 'girada',
  tischHaengtAn: 'anexada a',
  tischbild: 'Fondo',
  tischbildHolz: 'Madera',
  tischbildDunklesHolz: 'Madera oscura',
  tischbildPicknick: 'Picnic',
  tischbildDungeon: 'Mazmorra',
  tischbildTaverne: 'Taberna',
  tischbildTurniertisch: 'Mesa de torneo',
  tischEigenes: 'Imagen propia …',
  tischEigenesWeg: 'Quitar imagen propia',
  gespraech: 'Charla en la mesa',
  gespraechVerlauf: 'Lo que se ha dicho en la mesa',
  gespraechEingabe: 'Decir algo',
  gespraechSenden: 'Decir',
  gespraechLeer: 'Aún no ha dicho nadie nada.',
  gespraechNiemandZeigt: 'Ahora mismo nadie señala nada.',
  gespraechZeigt: '{sitz}: {tut}',
  gespraechZeigtAuf: '{sitz}: {tut} {objekt}',
  zeigerZeiger: 'señala',
  zeigerGreift: 'arrastra',
  zeigerSchwebt: 'mira',
  zeigerPing: 'llama',
  spielart: 'Modo',
  spielartGoldfish: 'Practicar en solitario',
  spielartSelfplay: 'Varios asientos',
  format: 'Formato',
  fehltDeck: 'Elige un mazo para cada asiento: el primero determina el juego.',
  fehltEigenesDeck:
    'Un asiento necesita un mazo propio: un préstamo no trae juego, y el juego sale del primer mazo propio.',
  fehltName:
    'Cada asiento necesita un nombre; en la mesa distingue a los jugadores.',

  // **Die digitale Deckleihe** (G6, #1463).
  leihe: 'Prestar o tomar prestado un mazo',
  leiheOeffnen: 'Prestar o tomar prestado …',
  leiheVerleihenTitel: 'Prestar un mazo',
  leiheAnnehmenTitel: 'Aceptar un préstamo',
  leiheListeTitel: 'Tus préstamos',
  leiheErklaerung:
    'Al prestar, un oponente puede jugar con una instantánea congelada de tu mazo. El mazo se queda aquí y no cambia ni una sola carta de tu colección.',
  leiheKeinEigentum:
    'Es una cesión de uso, no una transferencia. Nada cambia de dueño, ni digital ni físicamente.',
  leiheKeinKopierschutz:
    'Y no es una protección contra copia: quien ha visto las cartas, las ha visto. Lo que regula el préstamo es con qué mazo inicia este programa una partida.',
  leiheAnsehen: 'Ver préstamos',
  leiheKeine: 'No hay préstamos.',
  leiheEntleiher: '¿A quién se lo prestas?',
  leiheEntleiherLeer:
    'Aún no hay ningún interlocutor confirmado. Añade uno en el recuadro «Añadir o confirmar un interlocutor» de más abajo.',
  leiheDeck: 'Qué mazo',
  leiheDeckWaehlen: 'Elegir un mazo …',
  leiheKeinDeck: 'Aún no hay mazo: se crea en Mazos.',
  leiheFachleute: 'Prestar solo para una partida concreta',
  leiheFachleuteErklaerung:
    'Sin indicarlo, el préstamo vale para la próxima partida del interlocutor. Para atarlo a una sola, introduce el identificador de esa partida; el interlocutor lo encuentra en su lista de partidas.',
  leiheWoran: 'ID de la partida — vacío significa: la siguiente',
  leiheStunden: 'Válido durante cuántas horas',
  leiheZiel: 'Dónde se deja el sobre',
  leiheAnbieten: 'Prestar y guardar',
  leiheAbgelegt: 'El sobre está aquí:',
  leiheUmschlag: 'Sobre de un préstamo',
  leiheAnnehmen: 'Aceptar el préstamo',
  leiheWiderrufen: 'Revocar',
  leiheZurueckgeben: 'Devolver',
  leiheVerliehen: 'prestado a',
  leiheGeliehen: 'tomado prestado de',
  leiheBis: 'válido hasta',
  leiheKarten: 'cartas',
  leiheVerbraucht: 'canjeado en',
  leiheOffen: 'aún sin canjear',
  leiheZurueckgenommen: 'revocado',
  leiheModusZuende: 'Una partida en curso termina de jugarse.',
  leiheModusAbbruch: 'Una partida en curso se detiene.',
  leiheRuecknahmeGrenze:
    'Revocar impide nuevos inicios. No quita nada de una partida en curso ni elimina ninguna copia en el otro extremo.',

  // **Wie die Partie ausgegangen ist** (M17, #1618).
  ausgangTitel: 'El resultado',
  ausgangSieger: 'Ha ganado',
  ausgangUnentschieden: 'Empate: todos los que quedaban han perdido a la vez.',
  grundAufgabe: 'ha concedido',
  standLeiste: 'Estado de los jugadores',
  standDu: 'Tú',
  standLeben: 'Vidas',
  standGift: 'Veneno',
  standMana: 'Maná',
  standStrahlung: 'Radiación',
  standSteuer: 'Impuesto',
  standCommanderschaden: '{name}: {schaden}',
  standCommanderTod: 'con 21 la partida termina',
  standLeergezogen: 'robó de una biblioteca vacía',
  grundLeben: 'tenía cero vidas o menos',
  grundBibliothek: 'iba a robar de una biblioteca vacía',
  grundGift: 'tenía diez contadores de veneno',
  aufgeben: 'Conceder',
  aufgebenFrage:
    '¿Conceder? La partida termina de inmediato para este asiento, y sus cartas se van con él.',
  aufgebenJa: 'Sí, conceder',
  aufgebenNein: 'Seguir jugando',
  aufgebenHinweis:
    'Puedes conceder en cualquier momento y surte efecto al instante. Sales de la partida y tus cartas se van contigo.',

  // **Der Weg zurück in die eigene Partie** (#1600).

  // **Ein Sitz an einem Profil** (G8, #1465).
  sitzprofil: 'Perfil',
  sitzprofilKeines: 'sin perfil',
  sitzprofilErklaerung:
    'Si tienes dos perfiles, puedes asignar uno a cada asiento. Entonces la partida empieza en la vista del asiento que corresponde a tu perfil actual; a los demás llegas desde el selector de vista en la mesa. Lo que mostraba la otra vista desaparece. Si no asignas perfil a ningún asiento, tu perfil actual ocupa el primer asiento.',
  sitzprofilBefunde: 'Hay una nota al respecto:',
}
