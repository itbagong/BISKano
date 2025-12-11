<template>
  <!-- <s-input
    v-model="value"
    :key="value"
    :items="items"
    :label="caption != '' ? caption : label != '' ? label : field"
    :show-clear-button="true"
    :allow-add="allowAdd"
    :lookup-url="value ? data.lookupUrl : `/tenant/item/gets-detail`"
    :lookup-key="lookupKey"
    :lookup-labels="lookupLabels"
    :lookup-searchs="lookupSearchs"
    :disabled="disabled"
    :multiple="multiple"
    class="sku-item"
    ref="control"
    :useList="useList"
    :read-only="readOnly"
    :required="required"
    :keepErrorSection="keepErrorSection"
    :lookup-payload-builder="lookupPayloadBuilder"
    @focus="focus"
    @change="
      (field, v1, v2, old, ctlRef) => {
        onChangeItem(v1, v2, props.record);
      }
    "
  ></s-input> -->
  <label class="input_label" v-if="!hideLabel">
    <div
      v-if="
        (value && value.length > 0) ||
        !isNaN(value) ||
        useList ||
        kind == 'md' ||
        kind == 'date' ||
        kind == 'time' ||
        kind == 'markdown' ||
        keepLabel
      "
    >
      {{ label }}

      <span v-if="required" class="is_required font-extrabold">*</span>
    </div>
    <div v-else>&nbsp;</div>
  </label>
  <s-select-debouncer
    v-model="value"
    :key="value"
    :items="items"
    :label="caption != '' ? caption : label != '' ? label : field"
    :show-clear-button="true"
    :allow-add="allowAdd"
    :lookup-url="
      lookupUrl
        ? lookupUrl
        : value == undefined || value === ''
        ? `/tenant/item/gets-detail`
        : `/tenant/item/gets-detail?_id=${value}`
    "
    :lookup-key="lookupKey"
    :lookup-labels="lookupLabels"
    :lookup-searchs="lookupSearchs"
    :disabled="disabled"
    :multiple="multiple"
    class="sku-item"
    ref="control"
    :useList="useList"
    :read-only="readOnly"
    :required="required"
    :keepErrorSection="keepErrorSection"
    :lookup-payload-builder="lookupPayloadBuilder"
    @focus="focus"
    @selectOpen="selectOpen"
  ></s-select-debouncer>
</template>
<script setup>
import { reactive, computed, onMounted, ref, nextTick, inject } from "vue";
import { SInput, util } from "suimjs";
import SSelectDebouncer from "./SSelectDebouncer.vue";
const control = ref(null);
const axios = inject("axios");
const props = defineProps({
  modelValue: {
    type: [String, Number, Boolean, Object, Array],
    default: () => {},
  },
  items: {
    type: Array,
    default() {
      return [];
    },
  },

  record: { type: [Object, String], default: () => {} },
  ctlRef: { type: [Object, String], default: () => {} },
  readOnly: { type: Boolean },
  allowAdd: { type: Boolean, default: true },
  disabled: { type: Boolean, default: false },
  multiple: { type: Boolean, default: false },
  useList: { type: Boolean, default: true },
  isUrl: { type: Boolean, default: false },
  hideLabel: { type: Boolean, default: false },
  disableValidateOnChange: { type: Boolean, default: false },
  required: { type: Boolean, default: false },
  keepErrorSection: { type: Boolean, default: false },
  keepLabel: { type: Boolean, default: false },
  field: { type: String, default: "" },
  label: { type: String, default: "" },
  caption: { type: String, default: "" },
  kind: { type: String, default: "text" },
  lookupUrl: { type: String, default: "/tenant/item/gets-detail" },
  lookupKey: { type: String, default: "ID" },
  lookupLabels: { type: Array, default: () => ["Text"] },
  lookupSearchs: { type: Array, default: () => ["Text", "ID"] },
  rules: { type: Array, default: () => [] },
  queryProtocol: { type: [Object, undefined], default: undefined },
  lookupPayloadBuilder: { type: Function },
});
const emit = defineEmits({
  validate: null,
  "update:modelValue": null,
  "update:record": null,
  focus: null,
  change: null,
  search: null,
  beforeOnChange: null,
  afterOnChange: null,
  cancelOnChange: null,
});

const value = computed({
  get() {
    const val = props.modelValue;
    return val;
  },

  set(v) {
    const val = v;
    handleChange(val, value2(val), props.modelValue, props.ctlRef);
    emit("update:modelValue", val);
  },
});

const lookupUrl = computed({
  get() {
    const val = props.lookupUrl.split("?");
    let url = val[0];
    let queryparm = [];
    let queryString = [];
    if (val.length > 1) {
      queryparm = val[1].split("&");
    }
    for (let idx = 0; idx < queryparm.length; idx++) {
      let q = queryparm[idx].split("=");
      if (!["null", "undefined", ""].includes(q[1])) {
        queryString.push(queryparm[idx]);
      }
    }
    if (queryString.length > 0) {
      url = url + "?" + queryString.join("&");
    }
    return url;
  },
  set(v) {
    const val = v;
    emit("update:lookupUrl", val);
  },
});

const data = reactive({
  lookupUrl: props.lookupUrl,
});
function value2(key) {
  if (key == undefined) key = props.modelValue;
  if (control.value && control.value.value2) {
    const v2 = control.value.value2(key);
    return v2 ? v2 : key;
  }
  return key;
}

function handleChange(v1, v2, old, ctlRef) {
  if (v1 == undefined) v1 = value;
  if (v2 == undefined) v2 = value2(v1);
  if (old == undefined) old = "";
  emit("change", props.field, v1, v2, old, props.ctlRef);
  onChangeItem(v1, v2, props.record);
}

function validate() {
  control.value.validate();
  emit("validate", props.field, control.value.validate());
  return control.value.validate();
}

function focus(field, modelValue) {
  emit("focus", field, modelValue);
}
function selectOpen() {}

function onChangeItem(v1, v2, item) {
  item.ItemID = "";
  item.SKU = "";
  item.UnitCost = 0;
  item.UnitID = "";
  if (v1 != "" && typeof v1 == "string") {
    emit("beforeOnChange", item);
    const payload = {
      Take: 1,
      Where: {
        Field: "_id",
        Op: "$eq",
        Value: v1,
      },
    };
    axios.post("/tenant/item/gets-detail", payload).then(
      (r) => {
        item.ItemID = r.data[0].ItemID;
        item.SKU = v1.split("~~").length == 1 ? "" : r.data[0].SKU;
        item.UnitCost = r.data[0].Item.CostUnit;
        item.UnitID = r.data[0].Item.DefaultUnitID;
        item.Item = r.data[0].Item;
        item.ItemSpec = r.data[0].ItemSpec;
        item.UoM = r.data[0].Item.DefaultUnitID;
        emit("afterOnChange", item);
      },
      (e) => {
        util.showError(e);
      }
    );
  } else {
    emit("cancelOnChange", item);
  }
}
onMounted(() => {
  util.nextTickN(2, () => {
    let url = `/tenant/item/gets-detail`;
    if (props.isUrl) {
      url = props.lookupUrl;
    }
    data.lookupUrl = url;
  });
});
defineExpose({ validate, value2 });
</script>
<style scoped>
.sku-item {
  /* width: max-content; */
  min-width: 100%;
}
.is_required {
  color: red !important;
  font-size: 1em;
}
</style>
