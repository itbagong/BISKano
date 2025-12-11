<template>
  <div class="w-full">
    <data-list
      class="card"
      ref="listControl"
      :title="data.titleForm"
      grid-config="/tenant/warehouse/gridconfig"
      form-config="/tenant/warehouse/formconfig"
      grid-read="/tenant/warehouse/gets"
      form-read="/tenant/warehouse/get"
      grid-mode="grid"
      grid-delete="/tenant/warehouse/delete"
      form-keep-label
      form-insert="/tenant/warehouse/save"
      form-update="/tenant/warehouse/save"
      grid-sort-field="Created"
      grid-sort-direction="desc"
      :grid-fields="['Site', 'PIC']"
      :form-fields="['Dimension', 'PIC']"
      :grid-hide-new="!profile.canCreate"
      :grid-hide-edit="!profile.canUpdate"
      :grid-hide-delete="!profile.canDelete"
      :init-app-mode="data.appMode"
      :init-form-mode="data.formMode"
      @controlModeChanged="onControlModeChanged"
      @alter-grid-config="alterGridConfig"
      @formNewData="newRecord"
      @formEditData="editRecord"
    >
      <template #grid_Site="{ item }">
        {{
          item.Dimension && item.Dimension.find((_dim) => _dim.Key === "Site")
            ? findSite(
                item.Dimension.find((_dim) => _dim.Key === "Site")["Value"]
              )
            : ""
        }}
      </template>
      <template #grid_PIC="{ item, config }">
        <s-input
          read-only
          v-model="item.PIC"
          class="w-full"
          use-list
          :lookup-url="`/tenant/employee/find`"
          lookup-key="_id"
          :lookup-labels="['Name']"
          :lookup-searchs="['_id', 'Name']"
        ></s-input>
      </template>
      <template #form_input_Dimension="{ item }">
        <dimension-editor
          v-model="item.Dimension"
          :default-list="profile.Dimension"
        ></dimension-editor>
      </template>
      <template #form_input_PIC="{ item, config }">
        <s-input
          label="PIC"
          v-model="item.PIC"
          class="w-full"
          use-list
          :lookup-url="`/tenant/employee/find`"
          lookup-key="_id"
          :lookup-labels="['Name']"
          :lookup-searchs="['_id', 'Name']"
          :lookup-payload-builder="
            (search) => lookupPayloadBuilder(search, config, item.PIC, item)
          "
        ></s-input>
      </template>
    </data-list>
  </div>
</template>
<script setup>
import { reactive, ref, inject, onMounted } from "vue";
import { layoutStore } from "@/stores/layout.js";
import { DataList, util, SInput } from "suimjs";
import DimensionEditor from "@/components/common/DimensionEditorVertical.vue";
import DimensionInventory from "@/components/common/DimensionInventory.vue";
import { authStore } from "@/stores/auth";

layoutStore().name = "tenant";
const featureID = "Warehouse";
const profile = authStore().getRBAC(featureID);
const listControl = ref(null);
const axios = inject("axios");
const roleID = [
  (v) => {
    if (v == 0) return "required";
    return "";
  },
];
const data = reactive({
  appMode: "grid",
  formMode: "edit",
  titleForm: "Warehouse",
  record: {
    _id: "",
  },
  listSite: [],
});

function newRecord(record) {
  data.formMode = "new";
  data.titleForm = `Create New warehouse`;
  record._id = "";
  openForm(record);
}

function editRecord(record) {
  data.formMode = "edit";
  data.titleForm = `Edit warehouse | ${record._id}`;
  if (!record.Dimension) {
    record.Dimension = [];
  }
  data.record = record;
  openForm(record);
}

function openForm() {
  util.nextTickN(2, () => {
    listControl.value.setFormFieldAttr("_id", "rules", roleID);
  });
}

function alterGridConfig(cfg) {
  const colm = {
    field: "Site",
    kind: "text",
    label: "Site",
    readType: "show",
    labelField: "",
    input: {
      field: "Site",
      label: "Site",
      hint: "",
      hide: false,
      placeHolder: "Site",
      kind: "text",
    },
  };

  cfg.fields.splice(5, 0, colm);
}

function onControlModeChanged(mode) {
  if (mode === "grid") {
    data.titleForm = "Warehouse";
  }
}

function getsSite(site) {
  const url = "/tenant/dimension/find?DimensionType=Site";
  axios
    .post(url, {
      Take: -1,
      Sort: ["Label"],
      Select: ["Label", "_id"],
    })
    .then(
      (r) => {
        const dt = r.data;
        data.listSite = dt;
      },
      (e) => {
        util.showError(e);
      }
    );
}

function findSite(siteID) {
  const res = data.listSite.find((e) => {
    return e._id == siteID;
  });

  if (res == undefined) {
    return "";
  }
  return res.Label || "";
}

function lookupPayloadBuilder(search, config, value, item) {
  const qp = {};
  if (search != "") data.filterTxt = search;
  qp.Take = 20;
  qp.Sort = [config.lookupLabels[0]];
  qp.Select = config.lookupLabels;
  let idInSelect = false;
  const selectedFields = config.lookupLabels.map((x) => {
    if (x == config.lookupKey) {
      idInSelect = true;
    }
    return x;
  });
  if (!idInSelect) {
    selectedFields.push(config.lookupKey);
  }
  qp.Select = selectedFields;

  //setting search
  const Site =
    item.Dimension &&
    item.Dimension.find((_dim) => _dim.Key === "Site") &&
    item.Dimension.find((_dim) => _dim.Key === "Site")["Value"] != ""
      ? item.Dimension.find((_dim) => _dim.Key === "Site")["Value"]
      : undefined;
  const querySite = [
    {
      Field: "Dimension.Key",
      Op: "$eq",
      Value: "Site",
    },
    {
      Field: "Dimension.Value",
      Op: "$eq",
      Value: Site,
    },
  ];
  if (Site) {
    qp.Where = {
      Op: "$and",
      items: querySite,
    };
  }
  if (search !== "" && search !== null) {
    let items = [
      {
        Op: "$or",
        items: [
          { Field: "_id", Op: "$contains", Value: [search] },
          { Field: "Name", Op: "$contains", Value: [search] },
        ],
      },
    ];
    if (Site) {
      items = [...items, ...querySite];
    }
    qp.Where = {
      Op: "$and",
      items: items,
    };
  }
  // if ((value !== '' && value !== null) && (search == '' || search == null)) {
  //   qp.Where = {
  //       Op: "$or",
  //       items: [
  //           { Field: "_id", Op: "$contains", Value: [value] },
  //           { Field: "Name", Op:"$contains", Value: [value] }
  //       ]
  //   }
  // }

  return qp;
}
onMounted(() => {
  getsSite();
});
</script>
