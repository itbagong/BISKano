<template>
    <svg :height="data.height" :width="data.width" :data-uuid="data.uuid"> 
        <path   stroke="black" stroke-width="1" fill="none" />
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
  return [
    "M" + 0,
    data.height * 0.6,
    "L" + 0,
    0,
    "L" + data.width,
    0,
    "L" + data.width,
    data.height * 0.6,
    "L" + data.width / 2,
    data.height,
    "L" + 0,
    data.height * 0.6,
  ];
}
function draw() {
  const path = document.querySelector(
    "svg[data-uuid='" + data.uuid + "']>path"
  );
  if (path != null) {
    path.setAttribute("d", getPath().join(" "));
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
  resize(props.size * 0.6, props.size * 0.8);
});
defineExpose({
  resize,
});
</script>