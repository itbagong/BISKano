<template>  
  <div class="flex gap-4 p-2">
    <button @click="reset" class="p-2 bg-primary text-white"> Reset</button>
    <button @click="generate" class="p-2 bg-primary text-white"> Generate (Random)</button>
  </div>
  <div class="dndflow" @drop="onDrop" >
    <VueFlow @connect="onConnect"  @dragover="onDragOver" fit-view-on-init class="vue-flow-basic-example">
      <template #node-input="nodeProps"> 
          <node :node="nodeProps"  :targets="['']" :sources="['bottom']" :node-data="nodeProps.data" kind="input"  v-model="nodeProps.label" @remove="onRemoveNode" />
      </template>
      <template #node-decision="nodeProps"> 
          <node :node="nodeProps"  :targets="['top']" :sources="['bottom','right']" :node-data="nodeProps.data" kind="decision"  v-model="nodeProps.label" @remove="onRemoveNode" />
      </template>
      <template #node-inputoutput="nodeProps"> 
          <node :node="nodeProps" :targets="['top']" :sources="['bottom']" :node-data="nodeProps.data" kind="inputouput"  v-model="nodeProps.label" @remove="onRemoveNode" />
      </template>
      <template #node-process="nodeProps"> 
          <node :node="nodeProps" :targets="['top']" :sources="['bottom']" :node-data="nodeProps.data" kind="process"  v-model="nodeProps.label" @remove="onRemoveNode" />
      </template> 
      <template #node-predefinedprocess="nodeProps"> 
          <node :node="nodeProps" :targets="['top']" :sources="['bottom']" :node-data="nodeProps.data" kind="predefinedprocess"  v-model="nodeProps.label" @remove="onRemoveNode" />
      </template>    
      <template #node-filedb="nodeProps"> 
          <node :node="nodeProps" :targets="['top']" :sources="['bottom']" :node-data="nodeProps.data" kind="filedb"  v-model="nodeProps.label" @remove="onRemoveNode" />
      </template>  
      <template #node-filedbmultiple="nodeProps"> 
          <node :node="nodeProps" :targets="['top']" :sources="['bottom']" :node-data="nodeProps.data" kind="filedbmultiple"  v-model="nodeProps.label" @remove="onRemoveNode" />
      </template>  
      <template #node-manualoperation="nodeProps">  
          <node :node="nodeProps" :targets="['top']" :sources="['bottom']" :node-data="nodeProps.data" kind="manualoperation"  v-model="nodeProps.label" @remove="onRemoveNode" />
      </template>      
      <template #node-manualinput="nodeProps">  
          <node :node="nodeProps" :targets="['top']" :sources="['bottom']" :node-data="nodeProps.data" kind="manualinput"  v-model="nodeProps.label" @remove="onRemoveNode" />
      </template>
      <template #node-preparation="nodeProps">  
          <node :node="nodeProps" :targets="['top']" :sources="['bottom']" :node-data="nodeProps.data" kind="preparation"  v-model="nodeProps.label" @remove="onRemoveNode" />
      </template>
      <template #node-offpage="nodeProps">  
          <node :node="nodeProps" :targets="['top']" :sources="['bottom']" :node-data="nodeProps.data" kind="offpage"  v-model="nodeProps.label" @remove="onRemoveNode" />
      </template>
 
      <template #edge-custom="props">
        <edge-custom v-bind="props" @remove="onRemoveEdge" />
      </template>
      <template #connection-line="{ sourceX, sourceY, targetX, targetY, sourcePosition, targetPosition }">
        <connection-line
          :source-x="sourceX"
          :source-y="sourceY"
          :target-x="targetX"
          :target-y="targetY"
          :source-position="sourcePosition"
          :target-position="targetPosition"
        />
      </template>
      <MiniMap />
      <Controls/>
    </VueFlow>
    <sidebar/>
  </div>

 
</template>
<script setup>
import { Position } from "@vue-flow/core";
import { reactive, ref, inject, onMounted, nextTick, watch } from "vue";
import { layoutStore } from "@/stores/layout.js";

import Sidebar from "./widget/Flow/Sidebar.vue";
import Node from "./widget/Flow/Node.vue";
import EdgeCustom from "./widget/Flow/EdgeCustom.vue";
import ConnectionLine from "./widget/Flow/ConnectionLine.vue";

layoutStore().name = "tenant";

import { VueFlow, useVueFlow } from "@vue-flow/core";
import { Background } from "@vue-flow/background";
import { MiniMap } from "@vue-flow/minimap";
import { Controls } from "@vue-flow/controls";

import "@vue-flow/controls/dist/style.css";

const data = reactive({
  nodes: [],
  flows: [
    {
      id: "input-1",
      type: "input",
      label: "Input 1",
      targets: ["decision-1"],
    },

    {
      id: "decision-1",
      type: "decision",
      label: "decision 1",
      targets: ["decision-11", "decision-12"],
    },

    {
      id: "decision-11",
      type: "decision",
      label: "decision 1",
      targets: ["decision-111", "decision-112"],
    },
    {
      id: "decision-12",
      type: "decision",
      label: "decision 12",
    },
    {
      id: "decision-111",
      type: "decision",
      label: "decision 111",
    },
    {
      id: "decision-112",
      type: "decision",
      label: "decision 112",
      targets: ["decision-1121", "decision-1122"],
    },

    {
      id: "decision-1121",
      type: "decision",
      label: "decision 1121",
    },
    {
      id: "decision-1122",
      type: "decision",
      label: "decision 1122",
      targets: ["decision-11221", "decision-11222"],
    },
    {
      id: "decision-11222",
      type: "decision",
      label: "decision 11222",
    },
    {
      id: "decision-11221",
      type: "decision",
      label: "decision 11221",
      targets: ["decision-112211", "decision-112212"],
    },
    {
      id: "decision-112211",
      type: "decision",
      label: "decision 112211",
    },
    {
      id: "decision-112212",
      type: "decision",
      label: "decision 112212",
    },
  ],
});
const {
  findNode,
  nodes,
  edges,
  addEdges,
  addNodes,
  project,
  vueFlowRef,
  removeNodes,
  findEdge,
  removeEdges,
  setNodes,
  setEdges,
  fitView,
} = useVueFlow({
  nodes: [],
});

function onDragOver(event) {
  event.preventDefault();

  if (event.dataTransfer) {
    event.dataTransfer.dropEffect = "move";
  }
}

function onDrop(event) {
  const type = event.dataTransfer?.getData("application/vueflow");

  const { left, top } = vueFlowRef.value.getBoundingClientRect();

  const position = project({
    x: event.clientX - left,
    y: event.clientY - top,
  });

  const newNode = {
    id: self.crypto.randomUUID(),
    type,
    position,
    label: `${type} node`,
  };

  addNodes([newNode]);
  nextTick(() => {
    const node = findNode(newNode.id);
    const stop = watch(
      () => node.dimensions,
      (dimensions) => {
        if (dimensions.width > 0 && dimensions.height > 0) {
          node.position = {
            x: node.position.x - node.dimensions.width / 2,
            y: node.position.y - node.dimensions.height / 2,
          };
          stop();
        }
      },
      { deep: true, flush: "post" }
    );
  });
}
function onConnect(param) {
  param.type = "custom";
  addEdges(param, edges);
}
function onRemoveEdge(edgeId) {
  const edge = findEdge(edgeId);
  removeEdges(edge);
}
function onRemoveNode(node) {
  removeNodes(node);
}
function reset() {
  setNodes([]);
  setEdges([]);
}
function getRandomFlow() {
  let r = [];
  const nodes = [
    "decision",
    "inputoutput",
    "process",
    "predefinedprocess",
    "manualoperation",
    "manualinput",
    "preparation",
    "filedb",
    "filedbmultiple",
    "offpage",
  ];
  const getRandom = (min, max) => Math.floor(Math.random() * (max - min)) + min;
  const getNode = (id, type) => {
    const label = type + "_" + id;
    return {
      id,
      type,
      label,
      targets: [],
    };
  };
  const setTargetById = (id, target) => {
    const idx = r.findIndex((e) => e.id == id);
    r[idx].targets.unshift(target);
  };
  const createArrRandomNodes = (max, parentId = "") => {
    let lastId = parentId;
    for (var i = 0; i < max; i++) {
      const id = self.crypto.randomUUID();
      const type =
        lastId == "" && i == 0
          ? "input"
          : nodes[getRandom(0, nodes.length - 1)];
      r.push(getNode(id, type));
      if (lastId == "") {
        lastId = id;
        continue;
      }
      if (type == "decision") createArrRandomNodes(4, id);

      setTargetById(lastId, id);
      lastId = id;
    }
    // const rand = getRandom(0, 2);
    // const nodeOutput = r.find((e) => e.type == "output");
    // let outputId = "";
    // if (rand == 0 || nodeOutput == undefined) {
    //   outputId = self.crypto.randomUUID();
    //   r.push(getNode(outputId, "output"));
    // } else {
    //   outputId = nodeOutput?.id;
    // }
    const outputId = self.crypto.randomUUID();
    r.push(getNode(outputId, "output"));
    setTargetById(lastId, outputId);
  };

  createArrRandomNodes(getRandom(5, 20));

  return r;
}

function generate() {
  data.flows = getRandomFlow();

  const nodes = getArrNode()
    .sort((a, b) => b.index - a.index)
    .map((e) => {
      e.position = getPositionNode(e.xIndex, e.yIndex);
      return e;
    });
  const edges = nodes.reduce((acc, obj) => {
    const r =
      obj.targets?.map((e, i) => {
        return getObjEdge(obj.type, obj.id, e, i);
      }) ?? [];

    acc = [...acc, ...r];
    return acc;
  }, []);
  setNodes(nodes);
  setEdges(edges);
  setTimeout(() => {
    fitView();
  }, 100);
}
function getArrNode(flow) {
  let result = [];
  let xIndex = 0;

  function generateIndexNode(ctx, flow, yIndex = 0) {
    ctx.push({ ...flow, xIndex: xIndex, yIndex: yIndex });
    yIndex++;

    const targets = !flow.targets
      ? []
      : flow.targets
          .map((e) => data.flows.find((x) => x.id == e))
          .reduce((acc, obj) => {
            const idx = ctx.findIndex((e) => obj.id == e.id);
            if (idx == -1) acc.push(obj);
            return acc;
          }, []);

    targets.forEach((f, i) => {
      generateIndexNode(ctx, f, yIndex);
      if (i == 0 && flow.type == "decision") xIndex++;
    });
  }

  generateIndexNode(result, data.flows[0]);
  return result;
}

function getObjEdge(type, source, target, index) {
  const obj = {
    id: self.crypto.randomUUID(),
    type: "custom",
    source: source,
    target: target,
  };
  if (type == "decision") {
    obj.sourceHandle =
      source + "__handle" + (index == 0 ? "-bottom" : "-right");
  }
  return obj;
}
function getPositionNode(xIndex, yIndex) {
  return {
    x: calcX(xIndex),
    y: calcY(yIndex),
  };
}

function calcX(index) {
  const width = 120;
  const space = 120;
  return (width + space) * index;
}
function calcY(index) {
  const height = 120;
  const space = 60;
  return (height + space) * index;
}

onMounted(() => {});
</script>
<style>
.vue-flow__node-default, .vue-flow__node-input, .vue-flow__node-output{
  padding: 0;
  background: transparent;
}
.dndflow {
  flex-direction: column;
  display: flex;
  height: 100%;
}
.dndflow aside {
  color: #fff;
  font-weight: 700;
  border-right: 1px solid #eee;
  padding: 15px 10px;
  font-size: 12px;
  background: white;
  -webkit-box-shadow: 0px 5px 10px 0px rgba(0, 0, 0, 0.3);
  box-shadow: 0 5px 10px #0000004d;
}
.dndflow aside .nodes > * {
  margin-bottom: 10px;
  cursor: grab; 
}
.dndflow aside .description {
  margin-bottom: 10px;
}
.dndflow .vue-flow-wrapper {
  flex-grow: 1;
  height: 100%;
}
@media screen and (min-width: 640px) {
  .dndflow {
    flex-direction: row;
  }
  .dndflow aside {
    min-width: 15%;
  }
}
@media screen and (max-width: 639px) {
  .dndflow aside .nodes {
    display: flex;
    flex-direction: row;
    gap: 5px;
  }
}
.vue-flow__node-input {
  border: 1px solid black !important;
  border-radius: 99px;
}
.vue-flow__node .vue-flow__resize-control.line{
  border-style:none;
}
.vue-flow__node.selected .vue-flow__resize-control.line{
  border-style: dashed;
}
</style>
