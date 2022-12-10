import React, { forwardRef } from 'react';
// import axios from 'axios';
import Img from "react-cool-img";
// import './StickerStyle.css'
import loading_gif from './loading.gif'

// const mediaStyle = {
//   marginLeft: "auto",
//   marginRight: "auto",
//   display: "block"
// }

export const Sticker = forwardRef(({ id, faded, style, emoji, surl, onEmojiChange, ...props }, ref) => {

  if (surl.endsWith(".webp")) {
    return (
      <div ref={ref} style={style} {...props}>
        {/* <div style={{width: 64, height: 64}}> */}
        <Img src={surl} placeholder={loading_gif} alt="Loading..."
          retry={{ count: 10, delay: 2, acc: false }}
          width="auto" height="64" max-width="64" 
        ></Img>
        {/* </div> */}
        <br />
        <div>
          <label>{id}</label>
          <input type="text" value={emoji} size="6"
            onChange={(e) => onEmojiChange(id, e.target.value)}></input>
        </div>
      </div>
    );
  } else {
    return (
      <div ref={ref} style={style} {...props}>
        <video loop muted autoplay playsinline width="auto" height="64" max-width="64" >
          <source src={surl} type="video/webm" />
        </video>
        <br />
        <label>{id}</label>
        <input type="text" value={emoji} size="6"
          onChange={(e) => onEmojiChange(id, e.target.value)}></input>
      </div>
    )
  }
});
