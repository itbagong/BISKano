<template>
  <div>
    <div class="p-3 bg-gray-200">{{ props.titleHeader }}</div>
    <div class="p-3 border">
      <div class="flex flex-col gap-4">
        <div
          class="flex gap-1 items-center"
          v-for="(value, key, index) in data.dimValues"
        >
          <div class="flex flex-col gap-1 w-full">
            <div class="capitalize text-[0.7em] font-bold text-slate-500">
              {{ dimNames.find((v) => v.key == key).label }}
              <span
                v-if="props.mandatory.includes(key)"
                class="font-extrabold text-yellow-200"
                >*</span
              >
            </div>
            <s-input
              ref="inputs"
              hide-label
              v-model="data.dimValues[key]"
              :read-only="readOnly"
              :disabled="disabled"
              :required="props.mandatory.includes(key)"
              class="w-full"
              use-list
              :lookup-url="`/tenant/${
                props.dimNames.find((v) => v.key == key).point
              }/find`"
              lookup-key="_id"
              :lookup-labels="['Name']"
              :lookup-searchs="['_id', 'Name']"
              @change="
                (field, v1, v2, old, ctlRef) => {
                  onChange(v1, v2, data);
                }
              "
            ></s-input>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { reactive, watch, ref } from "vue";
import { util, SInput } from "suimjs";
const inputs = ref([]);
const props = defineProps({
  modelValue: { type: Object, default: () => {} },
  titleHeader: { type: String, default: "Inventory Dimension" },
  mandatory: { type: Array, default: () => [] }, // WarehouseID, AisleID,  SectionID, BoxID
  dimNames: {
    type: Array,
    default: () => [
      {
        key: "WarehouseID",
        label: "Warehouse",
        point: "warehouse",
      },
      {
        key: "SectionID",
        label: "Section",
        point: "section",
      },
      {
        key: "AisleID",
        label: "Aisle",
        point: "aisle",
      },
      {
        key: "BoxID",
        label: "Box",
        point: "box",
      },
    ],
  },
  column: { type: Number, default: 3 },
  readOnly: { type: Boolean, defaule: false },
  disabled: { type: Boolean, default: false },
});

const emit = defineEmits({
  "update:modelValue": null,
  onFieldChanged: null,
});

const data = reactive({
  val: true,
  dimValues: buildDimItem(props.dimNames),
});

function buildDimItem(dimNames) {
  const mv = props.modelValue ?? undefined;
  var object = dimNames.reduce(
    (obj, item) => Object.assign(obj, { [item.key]: mv ? mv[item.key] : "" }),
    {}
  );
  return object;
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

function onChange(v1, v2, item) {
  util.nextTickN(2, () => {
    emit("onFieldChanged", v1, v2, item);
  });
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

defineExpose({
  validate,
});
</script>
