<template>
    <svg :height="data.height" :width="data.width" :data-uuid="data.uuid"> 
       <path d="M0 60 L0 20 L80 20 L80 60 " stroke="black" stroke-width="1" fill="none" />
       <path d="M0 60 Q20 68 40 60 " stroke="black" stroke-width="1" fill="none" />
       <path d="M40 60 Q60 52 80 60 " stroke="black" stroke-width="1" fill="none" />
       <path d=" " stroke="black" stroke-width="1" fill="none" />  
        <path d=" " stroke="black" stroke-width="1" fill="none" />        
    </svg>
</template>
<script setup>
import { reactive, onMounted, defineExpose, nextTick } from "vue";
const props = defineProps({
  size: {
    type: Number,
    default: () => 50,
  },
});
const data = reactive({
  width: 0,
  height: 0,
  uuid: self.crypto.randomUUID(),
});

function getPath() {
  const w = data.width - 8;
  const y = 8;
  return [
    "M0",
    data.height * 0.7,
    "L0",
    y,
    "L" + w,
    y,
    "L" + w,
    data.height * 0.7,
  ];
}

function getPath2() {
  const w = data.width - 8;
  const y = 8;
  return [
    "M0",
    data.height * 0.7,
    "Q" + w * 0.25,
    data.height,
    w * 0.5,
    data.height * 0.7,
  ];
}
function getPath3() {
  const w = data.width - 8;
  return [
    "M" + w * 0.5,
    data.height * 0.7,
    "Q" + w * 0.75,
    data.height * 0.4,
    w,
    data.height * 0.7,
  ];
}
function getPath4() {
  return [
    "M7",
    4,
    "L7",
    0,
    "L" + data.width,
    0,
    "L" + data.width,
    data.height * 0.7 - 7,
    "L" + parseInt(data.width - 3),
    data.height * 0.7 - 7,
  ];
}

function getPath5() {
  return [
    "M4",
    7,
    "L4",
    4,
    "L" + parseInt(data.width - 4),
    4,
    "L" + parseInt(data.width - 4),
    data.height * 0.7 - 3,
    "L" + parseInt(data.width - 7),
    data.height * 0.7 - 3,
  ];
}

function draw() {
  const path = document.querySelector(
    "svg[data-uuid='" + data.uuid + "']>path:nth-child(1)"
  );
  if (path != null) {
    path.setAttribute("d", getPath().join(" "));
  }

  const path2 = document.querySelector(
    "svg[data-uuid='" + data.uuid + "']>path:nth-child(2)"
  );
  if (path2 != null) {
    path2.setAttribute("d", getPath2().join(" "));
  }

  const path3 = document.querySelector(
    "svg[data-uuid='" + data.uuid + "']>path:nth-child(3)"
  );
  if (path3 != null) {
    path3.setAttribute("d", getPath3().join(" "));
  }

  const path4 = document.querySelector(
    "svg[data-uuid='" + data.uuid + "']>path:nth-child(4)"
  );
  if (path4 != null) {
    path4.setAttribute("d", getPath4().join(" "));
  }

  const path5 = document.querySelector(
    "svg[data-uuid='" + data.uuid + "']>path:nth-child(5)"
  );
  console.log(path5, getPath5());
  if (path5 != null) {
    path5.setAttribute("d", getPath5().join(" "));
  }
}
function resize(w, h) {
  data.width = parseInt(w);
  data.height = parseInt(h);
  nextTick(() => {
    draw();
  });
}
onMounted(() => {
  resize(props.size, props.size * 0.6);
});
defineExpose({
  resize,
});
</script>