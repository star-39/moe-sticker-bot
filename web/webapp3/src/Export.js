import React, { useEffect, useReducer, useState } from 'react';
import axios, { all } from 'axios';
import { sha256sum } from './utils';

import { StickerGrid } from './StickerGrid'
import { SortableSticker } from './SortableSticker'
import { Sticker } from './Sticker'

function Export() {
  const [ss, setSS] = useState([])
  const uid = window.Telegram.WebApp.initDataUnsafe.user.id
  const queryId = window.Telegram.WebApp.initDataUnsafe.query_id
  const params = new Proxy(new URLSearchParams(window.location.search), {
    get: (searchParams, prop) => searchParams.get(prop),
  });

  const exportLinkHttps = `https://${process.env.REACT_APP_HOST}/webapp/api/export?sn=${params.sn}&qid=${queryId}&hex=${params.hex}`
  const exportLinkMsb = `msb://app/export/{params.sn}/?qid=${queryId}&hex=${params.hex}`

  useEffect(() => {
    axios.get(`/webapp/api/ss?sn=${params.sn}&uid=${uid}&qid=${queryId}&hex=${params.hex}&cmd=export`)
      .then(res => {
        setSS(res.data.ss)
      })
      .catch(() => { })

    // Android specific bug.
    // Android does not support opening custom scheme link
    // as well as https link from MainButton.
    // Hence, we need to put a button inside WebPage and do not
    // generate MainButton.
    if (window.Telegram.WebApp.platform !== "android") {
      window.Telegram.WebApp.MainButton.setText('Export/匯出').show()
        .onClick(() => {
          window.open(exportLinkMsb)
        })
    }
  }, [])
  return (
    <div>
      <button onClick={() => window.open(exportLinkHttps)}>
        Export/匯出
      </button>
      <br />
      <h3>
        Preview 預覽:
      </h3>
      <StickerGrid columns={4}>
        {
          ss.map((item) => (
            <SortableSticker
              key={item.id}
              id={item.id}
              emoji={item.emoji}
              surl={item.surl} />
          ))
        }
      </StickerGrid>
      <br />
      <button onClick={() => window.open(exportLinkHttps)}>
        Export/匯出
      </button>
    </div>
  )
}

export default Export
