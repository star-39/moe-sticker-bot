import { useEffect, useState } from 'react';
import Edit from './Edit'
import Export from './Export';
import axios from 'axios';
import { Route, Routes, useLocation } from "react-router-dom"

function App() {
  let [isInitDataValid, setIsInitDataValid] = useState(null)
  const route = useLocation().pathname
  const qstring = window.location.search ? window.location.search + "&" : "?"

  useEffect(() => {
    axios.post(`/webapp/api/initData${qstring}cmd=${route}`,
      new URLSearchParams(window.Telegram.WebApp.initData))
      .then(res => {
        setIsInitDataValid(true)
      })
      .catch(err => {
        setIsInitDataValid(false)
      })
  }, [])

  if (isInitDataValid === null) {
    // initData not generated yet.
    return;
  } else if (!isInitDataValid) {
    // Bad initData
    return (<div className="App"><h1>Invalid WebApp initData!!!</h1></div>);
  } else {
    // initData OK
    window.Telegram.WebApp.ready();
    return (
      <div className='App'>
        <header className="App-header">
        </header>
        <Routes>
          <Route path="/webapp/edit" element={<Edit />} />
          <Route path="/webapp/export" element={<Export />} />
        </Routes>
      </div>
    );
  }
}

export default App;
