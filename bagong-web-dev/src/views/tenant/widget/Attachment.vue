<template>
  <div class="dummy-attachment">
    <label class="input_label"> {{ label }} </label>
    <div class="flex justify-center">
      <div class="box-upload" v-for="(dt, idx) in data.record" :key="idx">
        <mdicon
          name="close-circle"
          size="15"
          class="cursor-pointer place-self-start absolute top-0 right-0"
          @click="data.record.splice(idx, 1)"
        />
        <mdicon name="image-outline" size="30" />
      </div>
      <s-button
        class="btn_primary w-8"
        icon="plus"
        @click="data.record.push(JSON.parse(JSON.stringify(modelUpload)))"
        v-if="data.record.length < max"
      />
    </div>
  </div>
</template>
<script setup>
import { reactive, ref, watch } from "vue";
import { SButton } from "suimjs";

const props = defineProps({
  modelValue: { type: Array, default: [] },
  label: { type: String, default: "" },
  max: { type: Number, default: 3 },
});

const data = reactive({
  record: props.modelValue == undefined ? [] : props.modelValue,
});

const modelUpload = {
  ID: "",
  Name: "",
};

const emit = defineEmits({
  "update:modelValue": null,
});

watch(
  () => data.record,
  (nv) => {
    emit("update:modelValue", nv);
  },
  { deep: true }
);
</script>
<style scoped>
.dummy-attachment .box-upload {
  @apply border-2 border-dashed border-zinc-200 w-20 h-20 mr-2 rounded-md grid place-content-center relative;
}
</style>
