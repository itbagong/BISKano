<template>
  <s-form
    v-model="data.record"
    :config="data.frmConfig"
    :mode="readOnly ? 'view' : 'edit'"
    keep-label
    only-icon-top
    hide-submit
    hide-cancel
    ref="frmControl"
    class="w-[200px]"
    @field-change="onFieldChange"
  >
  </s-form>
</template>
<script setup>
import { reactive, watch, ref, inject, onMounted } from "vue";
import { util, SForm, SInput } from "suimjs";
import helper from "@/scripts/helper.js";

const frmControl = ref(null);

const props = defineProps({
  cfg: { type: Object, default: {} },
  modelValue: {
    type: Object,
    default: {},
  },
  readOnly: { type: Boolean, default: false },
  scope: { type: String, default: "" },
  item: { type: Object, default: {} },
});

const emit = defineEmits({
  "update:modelValue": null,
  fieldChange: null,
});

const data = reactive({
  record: props.modelValue,
  frmConfig: helper.cloneObject(props.cfg),
});

function kindScope(param) {
  if (!param) return;
  let obj = {
    TPY001: ["Engineering", "Administrasi", "APD"],
    TPY002: ["Proaktif", "PencegahanLimbah", "PengolahanLimbah", "Dilusi"],
  };

  data.frmConfig.sectionGroups.forEach((sg) => {
    sg.sections.forEach((s) => {
      let row = s.rows.filter((r) => obj[param].includes(r.inputs[0].field));
      s.rows = row;
    });
  });
}

function onFieldChange(name, v1, v2) {
  emit("fieldChange", name, v1, v2, props.item);
}

onMounted(() => {
  if (props.scope) kindScope(props.scope);
});
</script>
