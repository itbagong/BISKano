<template>
  <div class="mb-8 w-full" :class="'pl-' + pl">
    <div v-for="(item, idx) in data.record" :key="idx">
      <div class="flex gap-2">
        <s-input
          keep-label
          label="Number"
          class="w-[250px]"
          v-model="item.Number"
        />
        <s-input
          keep-label
          label="Description"
          class="grow"
          v-model="item.Description"
          :multi-row="5"
        />
        <attachment label="Attachment" v-model="item.Attachment" :max="1" />
        <div class="flex gap-x-2 h-fit">
          <s-button
            class="flex btn_success mt-5"
            label="Add Child"
            @click="addChild(item)"
            v-if="item.level < maxLevel"
          />
          <mdicon
            name="delete"
            size="16"
            class="cursor-pointer mt-5"
            @click="data.record.splice(idx, 1)"
          />
        </div>
      </div>
      <parent-child
        v-model="item.Child"
        :pl="4"
        class="my-2"
        hide-add
        :max-level="maxLevel"
        v-if="item.Child.length > 0"
      />
    </div>
    <div class="flex justify-end my-4" v-if="!hideAdd">
      <s-button class="btn_success" label="Add Item" @click="addItems" />
    </div>
  </div>
</template>

<script setup>
import { reactive, ref, computed, watch } from "vue";
import { DataList, util, SInput, SButton } from "suimjs";
import Attachment from "./Attachment.vue";
import ParentChild from "./ItemTemplateChild.vue";

const props = defineProps({
  modelValue: { type: Array, default: () => [] },
  pl: { type: Number, default: 0 },
  hideAdd: { type: Boolean, default: false },
  maxLevel: { type: Number, default: 3 },
});

const data = reactive({
  record: props.modelValue == undefined ? [] : props.modelValue,
});

const emit = defineEmits({
  "update:modelValue": null,
});

function generateItem() {
  let res = JSON.parse(JSON.stringify(objEmpty));
  res.ID = util.uuid();
  return res;
}
const objEmpty = {
  ID: "",
  Number: "",
  Description: "",
  Attachment: [],
  Parent: "",
  Child: [],
  level: 1,
};

function addItems() {
  data.record.push(generateItem());
}
function addChild(item) {
  let ob = generateItem();
  ob.ID = item.ID + "#$" + ob.ID;
  ob.Parent = item.ID;
  ob.level = ob.ID.split("#$").length;
  item.Child.push(ob);
}

watch(
  () => data.record,
  (nv) => {
    emit("update:modelValue", nv);
  },
  { deep: true }
);
</script>

<style scoped>
.padding-left4 {
}
</style>
