import {useSortable} from '@dnd-kit/sortable';
import {CSS} from '@dnd-kit/utilities';

import {Sticker} from './Sticker';

export function SortableSticker(props) {
  const {
    attributes,
    listeners,
    setNodeRef,
    transform,
    transition,
  } = useSortable({id: props.id}); //id CANNOT be set to 0 !!!

  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
  };

  return (
    <Sticker
      ref={setNodeRef}
      style={style}
      {...props}
      {...attributes}
      {...listeners}
    />
  );
};
