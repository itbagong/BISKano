<template>
  <s-card class="rounded-md w-full" hide-title no-gap>
    <data-list
      class="card"
      ref="listControl"
      hide-title
      grid-config="/fico/cashjournal/gridconfig"
      form-config="/fico/cashjournal/formconfig"
      :grid-read="'/fico/cashjournal/gets?Status=READY'"
      grid-mode="grid"
      grid-delete="/fico/cashjournal/delete"
      :grid-fields="['References']"
      grid-sort-field="TrxDate"
      grid-sort-direction="desc"
      init-app-mode="grid"
      :grid-custom-filter="customFilter"
      @alter-grid-config="onAlterConfig"
      grid-hide-new
      grid-hide-delete
      grid-hide-edit
    >
      <template #grid_header_buttons_2>
        <s-button label="Submit" class="btn_primary" @click="onSubmit" />
      </template>
      <template #grid_References="{ item }">
        {{ showReferences(item.References) }}
      </template>
    </data-list>
  </s-card>
</template>
<script setup>
import { reactive, ref, computed } from "vue";
import { DataList, util, SButton } from "suimjs";

const props = defineProps({
  type: { type: String, default: "CASH IN" },
  siteId: { type: String, default: "" },
});

const mapType = {
  "SUBMISSION CASH OUT": "CASH OUT",
  "SUBMISSION CASH IN": "CASH IN",
};

const listControl = ref(null);
const emit = defineEmits({
  lineList: null,
});

const data = reactive({});

function onSubmit() {
  let Lines = [];
  let src = listControl.value.getGridSelected().value;

  for (let i in src) {
    let o = src[i];
    const newLines = o.Lines.map((e) => {
      e.References.push({
        Key: "HeaderID",
        Value: o._id,
      });
      return e;
    });
    Lines = [...Lines, ...newLines];
  }

  // set Line no
  for (let i in Lines) {
    let o = Lines[i];
    Lines[i].LineNo = parseInt(i) + 1;
    // if (props.type == "SUBMISSION CASH OUT") {
    //   o.Account = o.OffsetAccount;
    // }
  }

  emit("lineList", Lines);
}

const customFilter = computed(() => {
  let filters = null;
  const filterBySite = [
    {
      Op: "$eq",
      Field: "Dimension.Key",
      Value: "Site",
    },
    {
      Op: "$eq",
      Field: "Dimension.Value",
      Value: props.siteId,
    },
  ];
  const refer = [
    {
      Op: "$eq",
      Field: "References.Key",
      Value: "Submission Type",
    },
    {
      Op: "$eq",
      Field: "References.Value",
      Value: "Petty Cash",
    },
    {
      Op: "$ne",
      Field: "References.Key",
      Value: "SubmissionJournalID",
    },
    ...filterBySite,
  ];

  if (props.type == "SUBMISSION CASH IN") {
    filters = {
      Op: "$and",
      Items: [
        {
          Op: "$eq",
          Field: "CashJournalType",
          Value: "CASH IN",
        },
        ...refer,
      ],
    };
  } else {
    filters = {
      Op: "$or",
      Items: [
        {
          Op: "$and",
          Items: [
            {
              Op: "$eq",
              Field: "CashJournalType",
              Value: "CASH OUT",
            },
            {
              Op: "$ne",
              Field: "References.Key",
              Value: "SubmissionJournalID",
            },
            ...filterBySite,
          ],
        },
        {
          Op: "$and",
          Items: [
            {
              Op: "$eq",
              Field: "CashJournalType",
              Value: "CASH IN",
            },
            ...refer,
          ],
        },
      ],
    };
  }

  return filters;
});

function onAlterConfig(config) {
  config.setting.keywordFields = [
    ...config.setting.keywordFields,
    "CashJournalType",
  ];
  const CustomCol = ["CashJournalType", "References"];
  const numFields = [];
  let dimColl = [];
  for (let i in CustomCol) {
    let dim = CustomCol[i];
    dimColl.push({
      field: dim,
      kind: numFields.includes(dim) ? "number" : "text",
      label: dim.replace(/([A-Z])/g, " $1").trim(),
      readType: "show",
      labelField: "",
      input: {
        field: dim,
        label: dim,
        hint: "",
        hide: false,
        placeHolder: dim,
        kind: numFields.includes(dim) ? "number" : "text",
        disable: false,
        required: false,
        multiple: false,
      },
    });
  }
  config.fields = [...config.fields, ...dimColl];
}

function showReferences(params) {
  let f = params.find(
    (o) => o.Key == "Submission Type" && o.Value == "Petty Cash"
  );

  return f ? f.Key + " : " + f.Value : "-";
}
</script>
