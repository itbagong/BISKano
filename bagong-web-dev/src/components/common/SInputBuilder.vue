<template>
  <s-input
    v-model="value"
    :items="items"
    :label="caption != '' ? caption : label != '' ? label : field"
    :show-clear-button="true"
    :allow-add="allowAdd"
    :lookup-url="lookupUrl"
    :lookup-key="lookupKey"
    :lookup-labels="lookupLabels"
    :lookup-searchs="lookupSearchs"
    :lookup-payload-builder="onQuery"
    :disabled="disabled"
    :multiple="multiple"
    class="w-[100%]"
    ref="control"
    :useList="useList"
    :read-only="readOnly"
    :required="required"
    @focus="focus"
    @change="onChange"
  ></s-input>
</template>
<script setup>
import { reactive, computed, onMounted, ref, nextTick } from "vue";
import { SInput } from "suimjs";
const control = ref(null);
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

  ctlRef: { type: [Object, String], default: () => {} },
  readOnly: { type: Boolean },
  allowAdd: { type: Boolean, default: true },
  disabled: { type: Boolean, default: false },
  multiple: { type: Boolean, default: false },
  useList: { type: Boolean, default: true },
  disableValidateOnChange: { type: Boolean, default: false },
  required: { type: Boolean, default: false },
  keepErrorSection: { type: Boolean, default: false },
  field: { type: String, default: "" },
  label: { type: String, default: "" },
  caption: { type: String, default: "" },
  lookupUrl: { type: String, default: "" },
  lookupKey: { type: String, default: "" },
  lookupLabels: { type: Array, default: () => [] },
  lookupSearchs: { type: Array, default: () => [] },
  rules: { type: Array, default: () => [] },
  query: { type: Array, default: () => [] },
  queryProtocol: { type: [Object, undefined], default: undefined },
  lookupPayloadBuilder: { type: Function },
});
const emit = defineEmits({
  validate: null,
  "update:modelValue": null,
  focus: null,
  change: null,
  search: null,
});
const state = reactive({
  errors: [],
  fieldLabel: props.modelValue,
  editorMode: "text",
  showPassword: false,
  payloadBuilder: undefined,
});
const value = computed({
  get() {
    return props.modelValue;
  },

  set(v) {
    handleChange(v, value2(v), props.modelValue, props.ctlRef);
    emit("update:modelValue", v);
    nextTick(() => {
      if (!props.disableValidateOnChange) validate();
    });
  },
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
}

function validate() {
  control.value.validate();
  emit("validate", props.field, control.value.validate());
  return control.value.validate();
}

function focus(field, modelValue) {
  emit("focus", field, modelValue);
}

function onChange(field, v1, v2, old, ctl) {
  emit("change", field, v1, v2, old, ctl);
}

function onQuery(search) {
  emit("search", search);
  let qp = {};
  qp.Take = 20;
  qp.Sort = [props.lookupLabels[0]];
  qp.Select = props.lookupLabels;
  let idInSelect = false;
  const selectedFields = props.lookupLabels.map((x) => {
    if (x == props.lookupKey) {
      idInSelect = true;
    }
    return x;
  });
  if (!idInSelect) {
    selectedFields.push(props.lookupKey);
  }
  qp.Select = selectedFields;
  if (props.query.length > 0) {
    qp.Where = {
      Op: "$or",
      items: props.query,
    };
  }

  if (search.length > 0 && props.lookupSearchs.length > 0) {
    if (props.lookupSearchs.length == 1) {
      if (props.query.length == 0) {
        qp.Where = {
          Field: props.lookupSearchs[0],
          Op: "$contains",
          Value: [search],
        };
      } else {
        qp.Where = {
          Op: "$or",
          items: props.query,
        };
      }
    } else {
      let filterWhare = {
        Op: "$or",
        Items: props.lookupSearchs.map((el) => {
          return { Field: el, Op: "$contains", Value: [search] };
        }),
      };

      if (props.query.length > 0) {
        filterWhare = {
          Op: "$and",
          items: [
            {
              Op: "$or",
              Items: props.lookupSearchs.map((el) => {
                return { Field: el, Op: "$contains", Value: [search] };
              }),
            },
            props.query[0],
          ],
        };
      }

      qp.Where = filterWhare;
    }
  }
  if (
    props.multiple &&
    props.modelValue &&
    props.modelValue.length > 0 &&
    qp.Where != undefined
  ) {
    const whereExisting =
      props.modelValue.length == 1
        ? { Op: "$eq", Field: props.lookupKey, Value: props.modelValue[0] }
        : {
            Op: "$or",
            items: props.modelValue.map((el) => {
              return { Field: props.lookupKey, Op: "$eq", Value: el };
            }),
          };
    qp.Where = { Op: "$or", items: [qp.Where, whereExisting] };
  }

  if (props.queryProtocol) {
    qp = { ...qp, ...props.queryProtocol };
  }

  return qp;
}

defineExpose({ validate, value2 });
</script>
