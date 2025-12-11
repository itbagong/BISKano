

<template>
  <aside>
   
    <div class="nodes grid grid-cols-3 gap-3 [&>*]:flex [&>*]:justify-center [&>*]:items-center"> 
        <div :draggable="true" @dragstart="onDragStart($event, 'input')">
           <div class="w-[50px] h-[30px] rounded-full border-[1px] border-black ">&nbsp;</div>
        </div>
        <div :draggable="true" @dragstart="onDragStart($event, 'decision')">
            <shape-decision  />
        </div>
        <div :draggable="true" @dragstart="onDragStart($event, 'inputoutput')">
            <shape-input-output  />
        </div>
        <div :draggable="true" @dragstart="onDragStart($event, 'process')">
            <shape-process  />
        </div>  
        <div :draggable="true" @dragstart="onDragStart($event, 'predefinedprocess')">
            <shape-predefined-process  />
        </div>
      
        <div :draggable="true" @dragstart="onDragStart($event, 'manualoperation')">
            <shape-manual-operation  />
        </div>
        <div :draggable="true" @dragstart="onDragStart($event, 'manualinput')">
            <shape-manual-input  />
        </div>      
        <div :draggable="true" @dragstart="onDragStart($event, 'preparation')">
            <shape-preparation />
        </div>
        <div :draggable="true" @dragstart="onDragStart($event, 'filedb')">
            <shape-file-db />
        </div>
        <div :draggable="true" @dragstart="onDragStart($event, 'filedbmultiple')">
            <shape-file-db-multiple />
        </div>       
        <div :draggable="true" @dragstart="onDragStart($event, 'offpage')">
            <shape-off-page />
        </div>
    </div>
  </aside>
</template>
<script setup>
import ShapeDecision from "./Shape/Decision.vue";
import ShapeInputOutput from "./Shape/InputOutput.vue";
import ShapeProcess from "./Shape/Process.vue";
import ShapePredefinedProcess from "./Shape/PredefinedProcess.vue";
import ShapeManualOperation from "./Shape/ManualOperation.vue";
import ShapeManualInput from "./Shape/ManualInput.vue";
import ShapePreparation from "./Shape/Preparation.vue";
import ShapeFileDb from "./Shape/FileDb.vue";
import ShapeFileDbMultiple from "./Shape/FileDbMultiple.vue";
import ShapeOffPage from "./Shape/OffPage.vue";

function onDragStart(event, nodeType) {
  if (event.dataTransfer) {
    event.dataTransfer.setData("application/vueflow", nodeType);
    event.dataTransfer.effectAllowed = "move";
  }
}
</script>
<style>

</style>