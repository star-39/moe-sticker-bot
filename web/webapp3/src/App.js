import { useEffect, useState } from 'react';
import Edit from './Edit'
import axios from 'axios';

function App() {
  let [isInitDataValid, setIsInitDataValid] = useState(null)
  let [isBadWebAppState, setIsBadWebAppState] = useState(null)
  let [ss, setSS] = useState([])

  useEffect(() => {
    axios.post('/webapp/api/initData', new URLSearchParams(window.Telegram.WebApp.initData))
      .then(res => {
        setIsInitDataValid(true)
      })
      .catch(err => {
        setIsInitDataValid(false)
        if (err.response.data === "bad webapp state") {
          setIsBadWebAppState(true)
        }
        return
      })
      .then(() => {
        const uid = window.Telegram.WebApp.initDataUnsafe.user.id
        const queryId = window.Telegram.WebApp.initDataUnsafe.query_id
        axios.get(`/webapp/api/ss?uid=${uid}&qid=${queryId}`)
          .then(res => {
            setSS(res.data)
          })
      })
      .catch(e => {

      })
  }, [])

  if (isInitDataValid === null) {
    return;
  } else if (isInitDataValid) {
    if (ss.length === 0) {
      return;
    }
    window.Telegram.WebApp.ready();
    return (
      <div className="App">
        <header className="App-header">
        </header>
        <Edit ss={ss}></Edit>
      </div>
    );
  } else {
    if (isBadWebAppState) {
      return (
        <div className="App"><h3>Please launch WebApp through /manage command.</h3>
        <br/>
        <h3>請使用 /manage 指令後打開WebApp。</h3>
        </div>
      )
    }
    try {
      window.Telegram.WebApp.showAlert("Invalid initData!!");
      // window.Telegram.WebApp.close();
    } catch { }
    return (<div className="App"><h1>Invalid WebApp initData!!!</h1></div>);
  }
}


export default App;
