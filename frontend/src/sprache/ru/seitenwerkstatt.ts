// The card workspace: workspace, quick filter, zoom, overlays and the problem list —
// part of the `karten` catalog, split off because `seitenkarten.ts` sits at the line limit.
export const werkstatt = {
  arbeitsplatz: 'Рабочее место карты',
  inspectorBezeichnung: 'Поля и проверка',
  // **Marks on cards** (#1208).
  schnellfilter: {
    bezeichnung: 'Фильтр карт',
    platzhalter: 'Фильтр — напр. type:creature -goblin',
    leeren: 'Очистить фильтр',
    hilfe:
      'Пробел разделяет части, «…» удерживает их вместе, поле:значение привязывает к полю, ведущий - означает отрицание. Регистр не важен.',
    treffer: 'Карты: {gezeigt} из {gesamt}',
  },
  zoom: {
    bezeichnung: 'Размер предпросмотра',
    anpassen: 'Вписать',
    druck: 'Размер печати',
    ohneDruck: 'В шаблоне не указано разрешение, поэтому размера печати нет.',
  },
  // **Das Kartenbild in den Finder ziehen** — AP-12 (#2369).
  ziehen: {
    beschriftung: 'Перетащить изображение',
    hinweis:
      'Перетащи изображение в Finder, чтобы сохранить карту как PNG. То же самое делает «{weg}».',
  },
  overlays: {
    bezeichnung: 'Наложения',
    bildgrenzen: 'Границы изображений',
    beschnitt: 'Припуск на обрез',
    diagnose: 'Диагностика',
    beschnittHinweis: 'То, что находится внутри линии, переживёт обрез в 3 мм.',
  },
  probleme: {
    ueberschrift: 'Проблемы',
    keine: 'На этой карте проблем нет.',
    zumFeld: 'Перейти к полю {feld}',
    ohneFeld: 'без поля',
  },
  // **Paging** (#1290) — ported from MSE's Card menu.
  blaettern: {
    vorige: 'Предыдущая карта',
    naechste: 'Следующая карта',
    stelle: '{stelle} из {gesamt}',
    kuerzelVorige: '⇞',
    kuerzelNaechste: '⇟',
  },
}
