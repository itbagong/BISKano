<template>
    <NodeResizer min-width="100" min-height="30" @resize="resize" />
 
    <NodeToolbar
        class="bg-white flex gap-3"
        :is-visible="nodeData.toolbarVisible"
        :position="nodeData.toolbarPosition"
    >
        <button @click="remove">Delete</button>
        <button>Rotate</button> 
    </NodeToolbar>
    <div class="flex justify-center items-center"> 
        <shape-decision v-if="kind == 'decision'" :size="100"  ref="shapeCtl"/>
        <shape-input-output v-else-if="kind =='inputouput' " :size="100"  ref="shapeCtl"/>
        <shape-process v-else-if="kind =='process' " :size="100" ref="shapeCtl"/>
        <shape-predefined-process v-else-if="kind =='predefinedprocess' "  :size="100" ref="shapeCtl"/>
        <shape-file-db v-else-if="kind =='filedb' " :size="100" ref="shapeCtl"/>
        <shape-file-db-multiple v-else-if="kind =='filedbmultiple' " :size="100" ref="shapeCtl"/>
        <shape-manual-operation v-else-if="kind =='manualoperation' " :size="100" ref="shapeCtl"/>
        <shape-manual-input v-else-if="kind =='manualinput' " :size="100"   ref="shapeCtl"/>
        <shape-preparation v-else-if="kind =='preparation' " :size="100"   ref="shapeCtl"/>
        <shape-off-page v-else-if="kind =='offpage' " :size="100"   ref="shapeCtl"/>
    </div> 
    <div class="absolute z-[10] w-[100%] h-[100%] top-0 left-0">
        <div class=" w-full h-full flex justify-center items-center">
            <input v-model="value"  type="text" rows="1" class="w-[80%] bg-transparent text-center" style="resize:both" />
        </div>   
    </div>
    <Handle type="target" :position="Position.Top" v-if="targets.indexOf('top') > -1" />
    <Handle type="target" :position="Position.Bottom" v-if="targets.indexOf('bottom') > -1" />
    <Handle type="target" :position="Position.Left" v-if="targets.indexOf('left') > -1"/>
    <Handle type="target" :position="Position.Right" v-if="targets.indexOf('right') > -1"/>

    <Handle type="source" :position="Position.Top" v-if="sources.indexOf('top') > -1" :connectable="'single'" />
    <Handle type="source" :position="Position.Bottom" v-if="sources.indexOf('bottom') > -1" :connectable="'single'"/>
    <Handle type="source" :position="Position.Left" v-if="sources.indexOf('left') > -1" :connectable="'single'"/>
    <Handle type="source" :position="Position.Right" v-if="sources.indexOf('right') > -1" :connectable="'single'"/>
 
</template>
<script setup>
import { computed, ref } from "vue";
import { NodeToolbar } from "@vue-flow/node-toolbar";
import { Handle, Position } from "@vue-flow/core";
import { NodeResizer } from "@vue-flow/node-resizer";
import "@vue-flow/node-resizer/dist/style.css";
import ShapeDecision from "./Shape/Decision.vue";
import ShapeInputOutput from "./Shape/InputOutput.vue";
import ShapeProcess from "./Shape/Process.vue";
import ShapePredefinedProcess from "./Shape/PredefinedProcess.vue";
import ShapeFileDb from "./Shape/FileDb.vue";
import ShapeFileDbMultiple from "./Shape/FileDbMultiple.vue";
import ShapeManualOperation from "./Shape/ManualOperation.vue";
import ShapeManualInput from "./Shape/ManualInput.vue";

import ShapePreparation from "./Shape/Preparation.vue";
import ShapeOffPage from "./Shape/OffPage.vue";

const props = defineProps({
  modelValue: {
    type: String,
    default: () => "",
  },
  nodeData: {
    type: Object,
    default: () => {
      return {
        toolbarVisible: false,
        toolbarPosition: "",
      };
    },
  },
  node: {
    type: Object,
    default: () => {
      return {};
    },
  },
  kind: {
    type: String,
    default: () => {
      return "";
    },
  },
  targets: {
    type: Array,
    default: () => {
      return [];
    },
  },
  sources: {
    type: Array,
    default: () => {
      return [];
    },
  },
});
const shapeCtl = ref(null);
const emit = defineEmits({
  "update:modelValue": null,
  remove: null,
});
const value = computed({
  get() {
    return props.modelValue;
  },
  set(v) {
    emit("update:modelValue", v);
  },
});
function resize(e) {
  const { width, height } = e.params;
  shapeCtl.value?.resize(width, height);
}
function remove() {
  emit("remove", props.node);
}
</script>
<style>
 
</style>
