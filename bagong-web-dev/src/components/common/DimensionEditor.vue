<template>
  <div class="flex gap-2">
    <s-input
      v-for="v in data.dimValues"
      ref="inputs"
      :field="v.Key"
      v-model="v.Value"
      :label="hideLabel ? '' : customLabels[v.Key] ?? v.Key"
      :read-only="readOnly"
      use-list
      :lookup-payload-builder="
        v.DefaultList?.length > 0
          ? (...args) =>
              helper.payloadBuilderDimension(
                v.DefaultList,
                v.Value,
                multiple,
                ...args
              )
          : undefined
      "
      :lookup-url="`/tenant/dimension/find?DimensionType=${v.Key}`"
      lookup-key="_id"
      :lookup-labels="['Label']"
      :lookup-searchs="['_id', 'Label']"
      class="w-full"
      :required="requiredFields.includes(v.Key)"
      :multiple="multiple"
      @change="onChange"
    ></s-input>
  </div>
</template>

<script setup>
import { reactive, watch, onMounted, ref } from "vue";
import helper from "@/scripts/helper.js";
import { SInput } from "suimjs";

const props = defineProps({
  modelValue: { type: Array, default: () => [] },
  //-- TODO: dim names perlu membaca dari store
  dimNames: { type: Array, default: () => ["PC", "CC", "Site", "Asset"] },
  requiredFields: { type: Array, default: () => ["PC", "CC", "Site"] },
  defaultList: { type: Array, default: () => [] },
  customLabels: { type: Array, default: () => [] },
  column: { type: Number, default: 2 },
  readOnly: { type: Boolean, defaule: false },
  sectionTitle: { type: String, default: "" },
  multiple: { type: Boolean, default: false },
  hideLabel: { type: Boolean, defaule: false },
});
console.log(props.customLabels)
const emit = defineEmits({
  "update:modelValue": null,
  change: null,
});

const inputs = ref([]);
const data = reactive({
  dimValues: buildDimValues(props.dimNames),
});

watch(
  () => props.modelValue,
  (nv) => {
    if (JSON.stringify(nv) !== JSON.stringify(data.dimValues)) {
      data.dimValues =
        props.modelValue?.length == 0
          ? buildDimValues(props.dimNames)
          : props.modelValue;
    }
  }
);

function buildDimValues(dimNames) {
  const mv = props.modelValue ?? [];

  const vs = dimNames.map((el) => {
    const f = mv.filter((v) => v.Key == el);
    const list = props.defaultList
      .filter((v) => v.Key == el)
      .map((e) => e.Value);

    let v;

    if (props.multiple)
      v = f.length == 0 ? (list.length === 1 ? list : []) : f[0].Value;
    else v = f.length == 0 ? (list.length === 1 ? list[0] : "") : f[0].Value;

    return {
      Key: el,
      Value: v,
      DefaultList: list,
    };
  });
  return vs;
}
function validate() {
  let isValid = true;
  inputs.value.forEach((el) => {
    if (!el.validate()) {
      isValid = false;
    }
  });
  return isValid;
}
function onChange(field, v1, v2, old, ctl) {
  emit("change", field, v1, v2, old, ctl);
}
watch(
  () => props.dimNames,
  (nv, old) => {
    if (JSON.stringify(nv) != JSON.stringify(old)) {
      data.dimValues = buildDimValues(nv);
    }
    //data.dimTypes = nv
  }
);

watch(
  () => data.dimValues,
  (nv) => {
    emit("update:modelValue", nv);
  },
  { deep: true }
);
defineExpose({
  validate,
});
onMounted(() => {
  emit("update:modelValue", data.dimValues);
});
</script>
