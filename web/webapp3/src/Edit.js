import React, { useEffect, useReducer, useState } from 'react';
import axios from 'axios';
import { sha256sum } from './utils';

import {
  DndContext,
  closestCenter,
  DragOverlay,
  MouseSensor,
  TouchSensor,
  useSensor,
  useSensors,
} from '@dnd-kit/core';

import {
  arrayMove,
  SortableContext,
  rectSortingStrategy,
} from '@dnd-kit/sortable';

import { StickerGrid } from './StickerGrid'
import { SortableSticker } from './SortableSticker'
import { Sticker } from './Sticker'

let resultArray;

function Edit(props) {
  const [items, setItems] = useState(props.ss);
  const [activeId, setActiveId] = useState(null);

  const [, forceUpdate] = useReducer(x => x + 1, 0);

  const sensors = useSensors(
    useSensor(MouseSensor, {
      activationConstraint: {
        delay: 250,
      },
    }),
    useSensor(TouchSensor, {
      // Press delay of 250ms, with tolerance of 5px of movement
      activationConstraint: {
        delay: 250,
        tolerance: 5,
      },
    })
  );

  useEffect(() => {
    console.log(window.Telegram.WebApp.platform);
    window.Telegram.WebApp.MainButton.setText('Done').show()
      .onClick(() => {
        sha256sum(JSON.stringify(resultArray))
          .then((sum) => {
            const uid = window.Telegram.WebApp.initDataUnsafe.user.id
            const queryId = window.Telegram.WebApp.initDataUnsafe.query_id
            axios.post(
              `/webapp/api/edit/result?uid=${uid}&qid=${queryId}&sha256sum=${sum}`,
              JSON.stringify(resultArray))
              .then(() => {
                window.Telegram.WebApp.close();
              })
              .catch((error) => {
                if (error.response) {
                  window.Telegram.WebApp.showAlert(error + "\n" + error.response.data)
                } else {
                  window.Telegram.WebApp.showAlert(error)
                }
              });
          })
      });
    // This is to address Android specific bug.
    // Expanding the webapp by swiping the content up
    // might cause dnd-context to freeze and cannot be recovered.
    // iOS is not affected.
    if (window.Telegram.WebApp.platform === "android") {
      window.Telegram.WebApp.expand();
    }
  }, [])


  return (
    <div>
      <h3>Please hold and drag to reorder</h3>
      <h3>請按住並拖拽來排序</h3>
      <DndContext
        sensors={sensors}
        collisionDetection={closestCenter}
        onDragStart={handleDragStart}
        onDragEnd={handleDragEnd}
        onDragCancel={handleDragCancel}
      >
        <SortableContext items={items} strategy={rectSortingStrategy}>
          <StickerGrid columns={4}>
            {
              items.map((item) => (
                <SortableSticker
                  key={item.id}
                  id={item.id}
                  emoji={item.emoji}
                  onEmojiChange={setEmoji}
                  surl={item.surl} />
              ))
            }
          </StickerGrid>
        </SortableContext>

        <DragOverlay adjustScale={true}>
          {activeId ?
            <Sticker
              id={activeId}
              surl={items[items.map(o => o.id).indexOf(activeId)].surl}
              emoji={items[items.map(o => o.id).indexOf(activeId)].emoji}
            />
            : null}
        </DragOverlay>
      </DndContext>
    </div>
  );

  function handleDragStart(event) {
    setActiveId(event.active.id);
    window.Telegram.WebApp.HapticFeedback.impactOccurred("heavy");
  }

  function handleDragEnd(event) {
    const { active, over } = event;
    if (active.id !== over.id) {
      setItems((items) => {
        const oldIndex = items.map(o => o.id).indexOf(active.id)
        const newIndex = items.map(o => o.id).indexOf(over.id)
        apiSubmitIndexChange(oldIndex, newIndex, items) //This items is before change commited.
        const newArray = arrayMove(items, oldIndex, newIndex);
        resultArray = newArray;
        return newArray
      });
    }
    setActiveId(null);
    window.Telegram.WebApp.HapticFeedback.impactOccurred("soft");
  }

  function handleDragCancel() {
    setActiveId(null);
  }

  function setEmoji(id, value) {
    let newItems = items
    let pos = newItems.map(o => o.id).indexOf(id)
    newItems[pos].emoji = value
    newItems[pos].emoji_changed = true
    setItems(newItems)
    resultArray = newItems
    forceUpdate()
  }

  function apiSubmitIndexChange(oldIndex, newIndex, items) {
    const uid = window.Telegram.WebApp.initDataUnsafe.user.id
    const qid = window.Telegram.WebApp.initDataUnsafe.query_id
    let form = new FormData()
    form.append("oldIndex", oldIndex)
    form.append("newIndex", newIndex)
    axios.post(`/webapp/api/edit/move?uid=${uid}&qid=${qid}`, form)
      .then((res) => {
        console.log("pos mov ok.")
      })
      .catch((err) => {
        console.log("pos mov failed!")
        setItems(items) //Revert items.
        window.Telegram.WebApp.showAlert(
          `${err}\n${err.response.data}\nChange reverted, please try again`
          )
      })
  }
};


export default Edit;
