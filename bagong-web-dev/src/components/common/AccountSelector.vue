<template>
  <div class="flex gap-1" :class="{ 'flex-row': row, 'flex-col': !row }">
    <s-input
      v-model="data.accountType"
      v-if="!hideAccountType"
      :disabled="readOnly || disabled"
      class="w-[150px]"
      :label="labelType"
      use-list
      :hide-label="hideLabel"
      :items="props.itemsType"
    />

    <s-input
      v-if="multipleAccountId"
      v-model="data.value.AccountIDs"
      :disabled="readOnly || disabled"
      class="w-full"
      :hide-label="hideLabel"
      :label="data.accountFieldSetup.label"
      :use-list="data.accountFieldSetup.useList"
      :lookup-url="data.accountFieldSetup.lookupUrl"
      :lookup-key="data.accountFieldSetup.lookupKey"
      :lookup-payload-builder="data.accountFieldSetup.lookupPayloadBuilder"
      :lookup-labels="data.accountFieldSetup.lookupLabels"
      :lookup-searchs="data.accountFieldSetup.lookupLabels"
      :multiple="props.multipleAccountId"
    />
    <s-input
      v-else
      v-model="data.value.AccountID"
      :disabled="readOnly || disabled"
      class="w-full"
      :hide-label="hideLabel"
      :label="data.accountFieldSetup.label"
      :use-list="data.accountFieldSetup.useList"
      :lookup-key="data.accountFieldSetup.lookupKey"
      :lookup-url="data.accountFieldSetup.lookupUrl"
      :lookup-payload-builder="data.accountFieldSetup.lookupPayloadBuilder"
      :lookup-labels="data.accountFieldSetup.lookupLabels"
      :lookup-searchs="data.accountFieldSetup.lookupLabels"
      :multiple="props.multipleAccountId"
    />
  </div>
</template>

<script setup>
import { reactive, onMounted, watch, computed } from "vue";
import { SInput } from "suimjs";

const props = defineProps({
  multipleAccountId: { type: Boolean, default: false },
  modelValue: { type: Object, default: () => {} },
  hideAccountType: { type: Boolean, default: false },
  hideLabel: { type: Boolean, default: false },
  labelType: { type: String, default: "Type" },
  labelAccount: { type: String, default: "" },
  row: { type: Boolean, default: true },
  disabled: { type: Boolean, default: false },
  itemsType: {
    type: Array,
    default: () => {
      return ["COA", "EXP", "VND", "CST", "AST", "CNB", "INV", 'ASM', 'EMP', 'SHE', 'HCM' ];
    },
  },
  readOnly: { type: Boolean, default: false },
  groupIdValue: { type: Array, default: () => [] },
  lookupUrl: { type: String, default: "" },
});

const subledger = {
  COA: { code: "LEDGERACCOUNT" },
  EXP: { code: "EXPENSE" },
  VND: { code: "VENDOR" },
  CST: { code: "CUSTOMER" },
  AST: { code: "ASSET" },
  CNB: { code: "CASHBANK" },
  INV: { code: "INVENTORY" },
  ASM: { code: "ASSETMOVEMENT"},
  EMP: { code: "EMPLOYEE"},
  SHE: { code: "SHE"},
  HCM: { code: "HCM"}
};

function getKeySubledger(code) {
  return Object.keys(subledger).reduce((acc, key) => {
    if (subledger[key].code === code) acc = key;
    return acc;
  }, "");
}

const emit = defineEmits({
  "update:modelValue": null,
  change: null,
});

const data = reactive({
  accountFieldSetup: buildSelector(
    props.modelValue == undefined
      ? ""
      : getKeySubledger(props.modelValue.AccountType)
  ),
  accountType:
    props.modelValue && props.modelValue.AccountType != ""
      ? getKeySubledger(props.modelValue.AccountType)
      : props.itemsType[0],
  value: {
    AccountType:
      props.modelValue == undefined
        ? props.itemsType[0]
        : props.modelValue.AccountType,
    AccountID: props.modelValue == undefined ? "" : props.modelValue.AccountID,
    AccountIDs:
      props.modelValue == undefined ? [] : props.modelValue.AccountIDs,
  },
});

watch(
  () => data.accountType,
  (nv) => {
    data.accountFieldSetup = buildSelector(nv);
    data.value.AccountID = "";
    data.value.AccountIDs = [];
  }
);

watch(
  () => data.value.AccountID,
  (nv) => {
    const subl = subledger[data.accountType];
    const account = {
      AccountType: subl == undefined ? "" : subl.code,
      AccountID: data.value.AccountID,
      AccountIDs: data.value.AccountIDs,
    };

    emit("update:modelValue", account);
    emit("change", account);
  }
);
watch(
  () => data.value.AccountIDs,
  (nv) => {
    const subl = subledger[data.accountType];
    const account = {
      AccountType: subl == undefined ? "" : subl.code,
      AccountID: data.value.AccountID,
      AccountIDs: data.value.AccountIDs,
    };

    emit("update:modelValue", account);
    emit("change", account);
  }
);

function buildSelector(code) {
  if ((code == "" || code == null) && props.itemsType.length > 0) code = props.itemsType[0];
  if (subledger[code] == undefined) return;
  const subType = subledger[code].code;
  const lowerName = subType.toLowerCase();
  let lookupUrl =
    lowerName == "expense"
      ? "expensetype"
      : lowerName == "inventory"
      ? "item"
      : lowerName;
  let buildUrl = 
    lowerName == "she"
      ? `bagong/menu?menu=she` : `tenant/${lookupUrl}/${code == "COA" ? "coa/find" : "find"}`;
  if (lowerName === 'hcm') {
    buildUrl = 'bagong/menu?menu=hcm'
  }
  const setup = {
    useList: true,
    label: lowerName,
    lookupItems: lowerName == "she" ? ["Label"] : ["Name"],
    lookupUrl: props.lookupUrl ? props.lookupUrl : buildUrl,
    lookupKey: "_id",
    lookupPayloadBuilder:
      code == "EXP" && props.groupIdValue.length > 0
        ? () => {
            return {
              Take: 20,
              Sort: ["Name"],
              Select: ["Name", "_id"],
              Where: {
                Op: "$contains",
                Field: "GroupID",
                Value: props.groupIdValue,
              },
            };
          }
        : undefined,
    lookupLabels: ["COA","CST", "EMP"].indexOf(code) >= 0 ? ["_id", "Name"] : ["SHE", 'HCM'].indexOf(code) >= 0 ? ["Label"] : ["Name"],
  };
  if (props.labelAccount != "") {
    setup.label = props.labelAccount;
  } else if (lowerName.length > 0)
    setup.label = subType[0] + lowerName.substr(1, lowerName.length - 1);

  return setup;
}

onMounted(() => {
  if (data.accountType != "")
    data.accountFieldSetup = buildSelector(data.accountType);
});
</script>
<style scoped>
.flexRow {
  @apply flex-row;
}
.flexCol {
  @apply flex-col;
}
</style>
