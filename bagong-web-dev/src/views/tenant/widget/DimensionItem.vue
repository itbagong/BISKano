<template>
  <div class="flex gap-2">
    <s-input
      v-for="(value, key, index) in data.dimValues"
      v-model="data.dimValues[key]"
      :label="dimNames.find((v) => v.key == key).label"
      :read-only="!readOnly ? false : !mirrorValue[key]"
      kind="checkbox"
      class="w-full"
    ></s-input>
  </div>
</template>

<script setup>
import { reactive, watch } from "vue";
import { SInput } from "suimjs";

const props = defineProps({
  modelValue: { type: Object, default: () => {} },
  dimNames: {
    type: Array,
    default: () => [
      {
        key: "IsEnabledSpecVariant",
        label: "Variant",
      },
      {
        key: "IsEnabledSpecSize",
        label: "Size",
      },
      {
        key: "IsEnabledSpecGrade",
        label: "Grade",
      },
      {
        key: "IsEnabledItemBatch",
        label: "Batch",
      },
      {
        key: "IsEnabledItemSerial",
        label: "Serial",
      },
      {
        key: "IsEnabledLocationWarehouse",
        label: "Warehouse",
      },
      {
        key: "IsEnabledLocationAisle",
        label: "Aisle",
      },
      {
        key: "IsEnabledLocationSection",
        label: "Section",
      },
      {
        key: "IsEnabledLocationBox",
        label: "Box",
      },
    ],
  },
  column: { type: Number, default: 3 },
  readOnly: { type: Boolean, defaule: false },
  mirrorValue: {
    type: Object,
    default: {
      IsEnabledSpecVariant: false,
      IsEnabledSpecSize: false,
      IsEnabledSpecGrade: false,
      IsEnabledItemBatch: false,
      IsEnabledItemSerial: false,
      IsEnabledLocationWarehouse: false,
      IsEnabledLocationAisle: false,
      IsEnabledLocationSection: false,
      IsEnabledLocationBox: false,
    },
  },
});

const emit = defineEmits({
  "update:modelValue": null,
});

const data = reactive({
  val: true,
  dimValues: buildDimItem(props.dimNames),
});

function buildDimItem(dimNames) {
  const mv = props.modelValue ?? undefined;
  var object = dimNames.reduce(
    (obj, item) =>
      Object.assign(obj, { [item.key]: mv ? mv[item.key] : false }),
    {}
  );
  return object;
}

watch(
  () => props.dimNames,
  (nv) => {
    data.dimValues = buildDimItem(nv);
  }
);

watch(
  () => data.dimValues,
  (nv) => {
    emit("update:modelValue", nv);
  },
  { deep: true }
);

watch(
  () => props.modelValue,
  (nv) => {
    data.dimValues = buildDimItem(props.dimNames);
  }
);
</script>
