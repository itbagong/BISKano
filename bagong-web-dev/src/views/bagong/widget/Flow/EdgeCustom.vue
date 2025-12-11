<template>
    <BaseEdge :path="data.path" :markerEnd="markerEnd" :style="style" />
    <EdgeLabelRenderer>
         <div
          :style="{
                position: 'absolute',
                transform: `translate(-50%, -50%) translate(${data.labelX}px,${data.labelY}px)`,
                fontSize: 12,
                pointerEvents: 'all'
            }
          "
          className="nodrag nopan"
        >
        <div class="flex flex-col justify-center">
          <div class="text-center">
          <button class="w-[20px] h-[20px] rounded-[50%] bg-[#ddd] hover:bg-[white]  hover:font-bold" @click="remove">
            ×
          </button> 
          </div>

          <label v-if="sourceNode.type == 'decision' && sourcePosition == 'right'" class="bg-[#F1F4F8] p-1">No</label>
          <label v-if="sourceNode.type == 'decision' && sourcePosition == 'bottom'"  class="bg-[#F1F4F8] p-1">Yes</label>
        </div>
        </div>
    </EdgeLabelRenderer>
</template>
<script setup>
import { onMounted, reactive, watch } from "vue";
import {
  BaseEdge,
  EdgeLabelRenderer,
  getBezierPath,
  getSmoothStepPath,
  StepEdge,
} from "@vue-flow/core";
const data = reactive({
  path: "",
  labelX: 0,
  labelY: 0,
});
const props = defineProps({
  id: {
    type: Number,
    default: () => {
      return "";
    },
  },
  sourceX: {
    type: Number,
    required: true,
  },
  sourceY: {
    type: Number,
    required: true,
  },
  sourcePosition: {
    type: Number,
    required: true,
  },
  targetX: {
    type: Number,
    required: true,
  },
  targetY: {
    type: Number,
    required: true,
  },
  targetPosition: {
    type: Number,
    required: true,
  },
  style: {
    type: Object,
    default: {},
  },
  markerEnd: {
    type: String,
    default: "",
  },
  curvature: {
    type: Number,
    default: 0,
  },
  sourceNode: {
    type: Object,
    default: () => {},
  },
  targetNode: {
    type: Object,
    default: () => {},
  },
});
const emit = defineEmits({
  remove: null,
});
function init() {
  const [path, labelX, labelY] = getSmoothStepPath({
    sourceX: props.sourceX,
    sourceY: props.sourceY,
    sourcePosition: props.sourcePosition,
    targetX: props.targetX,
    targetY: props.targetY,
    targetPosition: props.targetPosition,
  });
  data.path = path;
  data.labelX = labelX;
  data.labelY = labelY;
}
function remove() {
  emit("remove", props.id);
}
onMounted(() => {
  init();
});
watch(
  () => props,
  () => {
    init();
  },
  { deep: true }
);
</script>