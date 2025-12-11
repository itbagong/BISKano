<template>
    <svg :height="data.height" :width="data.width" :data-uuid="data.uuid"> 
       <path  stroke="black" stroke-width="1" fill="none" />
       <path  stroke="black" stroke-width="1" fill="none" />
       <path  stroke="black" stroke-width="1" fill="none" />
        
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
  gapY: 0.2,
});

function getPath() {
  return [
    "M0",
    data.height * 0.7,
    "L0",
    0,
    "L" + data.width,
    0,
    "L" + data.width,
    data.height * 0.7,
  ];
}

function getPath2() {
  return [
    "M0",
    data.height * 0.7,
    "Q" + data.width * 0.25,
    data.height,
    data.width * 0.5,
    data.height * 0.7,
  ];
}
function getPath3() {
  return [
    "M" + data.width * 0.5,
    data.height * 0.7,
    "Q" + data.width * 0.75,
    data.height * 0.4,
    data.width,
    data.height * 0.7,
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