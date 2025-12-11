<template>
  <div>
    <div class="p-3 bg-gray-200">{{ props.titleHeader }}</div>
    <div class="p-3 border">
      <div class="flex flex-col gap-4">
        <s-form
          v-model="data.bgDimension"
          :config="data.frmCfg"
          :mode="readOnly ? 'view' : 'edit'"
          keep-label
          only-icon-top
          hide-submit
          hide-cancel
          ref="frmInventJurnal"
        >
          <template #input_WarehouseID="{ item }">
            <s-input-builder
              ref="inputWarehouseID"
              v-model="item.WarehouseID"
              useList
              label="Warehouse"
              class="w-full"
              lookup-url="/tenant/warehouse/find"
              lookup-key="_id"
              :lookup-labels="['Name', '_id']"
              :lookup-searchs="['_id', 'Name']"
              :query="data.queryParams"
              :query-protocol="
                data.warehouseparams ? data.warehouseparams : undefined
              "
              :multiple="false"
              :disabled="
                disabled ||
                readOnly ||
                props.disableField.includes('WarehouseID')
              "
              :required="props.mandatory.includes('WarehouseID')"
              @focus="selectOpen"
              @search="searchOption"
              @change="
                (field, v1, v2, old) => {
                  onChange('WarehouseID', v1, v2, old, item);
                }
              "
            ></s-input-builder>
          </template>
          <template #input_AisleID="{ item }">
            <s-input-builder
              ref="inputAisleID"
              v-model="item.AisleID"
              useList
              label="Aisle"
              class="w-full"
              lookup-url="/tenant/aisle/find"
              lookup-key="_id"
              :lookup-labels="['Name', '_id']"
              :lookup-searchs="['_id', 'Name']"
              :query="data.queryParams"
              :multiple="false"
              :disabled="
                disabled || readOnly || props.disableField.includes('AisleID')
              "
              :required="props.mandatory.includes('AisleID')"
              @change="
                (field, v1, v2, old) => {
                  onChange('AisleID', v1, v2, old, item);
                }
              "
            ></s-input-builder>
          </template>
          <template #input_SectionID="{ item }">
            <s-input-builder
              ref="inputSectionID"
              v-model="item.SectionID"
              useList
              label="Section"
              class="w-full"
              lookup-url="/tenant/section/find"
              lookup-key="_id"
              :lookup-labels="['Name', '_id']"
              :lookup-searchs="['_id', 'Name']"
              :query="data.queryParams"
              :multiple="false"
              :disabled="
                disabled || readOnly || props.disableField.includes('SectionID')
              "
              :required="props.mandatory.includes('SectionID')"
              @change="
                (field, v1, v2, old) => {
                  onChange('SectionID', v1, v2, old, item);
                }
              "
            ></s-input-builder>
          </template>
          <template #input_BoxID="{ item }">
            <s-input-builder
              ref="inputBoxID"
              v-model="item.BoxID"
              useList
              label="Box"
              class="w-full"
              lookup-url="/tenant/box/find"
              lookup-key="_id"
              :lookup-labels="['Name', '_id']"
              :lookup-searchs="['_id', 'Name']"
              :query="data.queryParams"
              :multiple="false"
              :disabled="
                disabled || readOnly || props.disableField.includes('BoxID')
              "
              :required="props.mandatory.includes('BoxID')"
              @change="
                (field, v1, v2, old) => {
                  onChange('BoxID', v1, v2, old, item);
                }
              "
            ></s-input-builder>
          </template>
        </s-form>
      </div>
    </div>
  </div>
</template>

<script setup>
import { reactive, watch, ref, inject, onMounted } from "vue";
import { util, loadFormConfig, SForm, SInput } from "suimjs";
import SInputBuilder from "./SInputBuilder.vue";
const inputs = ref([]);
const inputWarehouseID = ref(null);
const inputAisleID = ref(null);
const inputSectionID = ref(null);
const inputBoxID = ref(null);
const props = defineProps({
  modelValue: { type: Object, default: () => {} },
  titleHeader: { type: String, default: "Inventory Dimension" },
  mandatory: { type: Array, default: () => [] }, // WarehouseID, AisleID,  SectionID, BoxID
  disableField: { type: Array, default: () => [] }, // WarehouseID, AisleID,  SectionID, BoxID
  hideField: { type: Array, default: () => [] }, // WarehouseID, AisleID,  SectionID, BoxID
  hideSections: { type: Array, default: () => [] }, // WarehouseID, AisleID,  SectionID, BoxID
  column: { type: Number, default: 3 },
  readOnly: { type: Boolean, defaule: false },
  disabled: { type: Boolean, default: false },
  site: { type: [String, undefined], default: undefined },
  defaultList: { type: Array, default: () => [] },
});

const emit = defineEmits({
  "update:modelValue": null,
  onFieldChanged: null,
  defaultWH: null,
  alterFormConfig: null,
});
const axios = inject("axios");
const frmInventJurnal = ref(null);
const data = reactive({
  val: true,
  frmCfg: {},
  bgDimension:
    props.modelValue == null || props.modelValue == undefined
      ? {}
      : props.modelValue,
  queryParams: [],
  warehouseparams: undefined,
  sectionparams: undefined,
  aisleparams: undefined,
  boxparams: undefined,
  params: undefined,
  defaultListWarehouse: [],
  defaultListSection: [],
  defaultListAisle: [],
  defaultListBox: [],
});

watch(
  () => props.site,
  (Site) => {
    if (Site) {
      const lookuplabels = ["Name", "_id"];
      data.warehouseparams = {};
      data.warehouseparams.Take = 20;
      data.warehouseparams.Sort = [lookuplabels[0]];
      data.warehouseparams.Select = lookuplabels;

      data.warehouseparams.Where = {
        Op: "$and",
        Items: [
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
        ],
      };

      // if (data.queryParams.length > 0) {
      //   data.warehouseparams.Where.Items.push({
      //     Op: "$or",
      //     Items: data.queryParams,
      //   })
      // }
    } else {
      setDefaultListWH(
        data.defaultListWarehouse.length > 0 ? data.defaultListWarehouse : []
      );
    }
  }
);

function setRequired() {
  util.nextTickN(2, () => {
    frmInventJurnal.value.getAllField().forEach((e) => {
      frmInventJurnal.value.setFieldAttr(
        e.field,
        "required",
        props.mandatory.includes(e.field)
      );
    });
  });
}

function setDisable() {
  frmInventJurnal.value.getAllField().forEach((e) => {
    frmInventJurnal.value.setFieldAttr(
      e.field,
      "disable",
      props.disableField.includes(e.field)
    );
    frmInventJurnal.value.setFieldAttr(
      e.field,
      "readOnly",
      props.disableField.includes(e.field)
    );
  });
}
function validate() {
  return (
    frmInventJurnal.value.validate() &&
    inputWarehouseID.value.validate() &&
    inputAisleID.value.validate() &&
    inputSectionID.value.validate() &&
    inputBoxID.value.validate()
  );
}

function selectOpen(v1, v2, item) {
  data.queryParams = [
    {
      Field: "IsActive",
      Op: "$eq",
      Value: true,
    },
  ];
  if (props.site) {
    const lookuplabels = ["Name", "_id"];
    data.warehouseparams = {};
    data.warehouseparams.Take = 20;
    data.warehouseparams.Sort = [lookuplabels[0]];
    data.warehouseparams.Select = lookuplabels;

    data.warehouseparams.Where = {
      Op: "$and",
      Items: [
        {
          Field: "Dimension.Key",
          Op: "$eq",
          Value: "Site",
        },
        {
          Field: "Dimension.Value",
          Op: "$eq",
          Value: props.site,
        },
      ],
    };
    data.queryParams = [];
  }
}
function searchOption(search) {
  let ActiveItem = [
    {
      Field: "IsActive",
      Op: "$eq",
      Value: true,
    },
  ];
  if (search.length > 0) {
    ActiveItem = [
      {
        Field: "IsActive",
        Op: "$eq",
        Value: [true],
      },
    ];
  }

  if (data.warehouseparams) {
    const lookupsearch = ["_id", "Name"];
    if (search.length > 0 && lookupsearch.length > 0) {
      if (lookupsearch.length == 1)
        if (data.queryParams.length == 0) {
          data.warehouseparams.Where.Items[2] = {
            Field: lookupsearch[0],
            Op: "$contains",
            Value: [search],
          };
        } else {
          data.warehouseparams.Where.Items[2] = {
            Op: "$or",
            items: data.queryParams,
          };
        }
      else
        data.warehouseparams.Where.Items[2] = {
          Op: "$or",
          items: [
            ...lookupsearch.map((el) => {
              return { Field: el, Op: "$contains", Value: [search] };
            }),
            ...data.queryParams,
          ],
        };
    }
  }
}

function onChange(field, v1, v2, old, item) {
  util.nextTickN(2, () => {
    emit("onChange", field, v1, v2, old, item);
  });
}
watch(
  () => data.bgDimension,
  (nv) => {
    emit("update:modelValue", nv);
    emit("onFieldChanged", nv);
  },
  { deep: true }
);

watch(
  () => props.modelValue,
  (nv) => {
    util.nextTickN(2, () => {
      data.bgDimension = nv;
    });
  },
  { deep: true }
);

//--- warehouse
watch(
  () => data.defaultListWarehouse,
  (nv) => {
    if (nv.length > 0) {
      setDefaultListWH(nv);
    }
  },
  { deep: true }
);
function setDefaultListWH(nv) {
  const list = nv.map((o) => o.Value);
  const lookuplabels = ["Name", "_id"];
  data.warehouseparams = {};
  data.warehouseparams.Take = 20;
  data.warehouseparams.Sort = [lookuplabels[0]];
  data.warehouseparams.Select = lookuplabels;

  data.warehouseparams.Where = {
    Op: "$or",
    Items: [
      {
        Field: "_id",
        Op: "$contains",
        Value: list.length > 0 ? list : [""],
      },
      {
        Field: "Name",
        Op: "$contains",
        Value: list.length > 0 ? list : [""],
      },
    ],
  };
}

//--- section
watch(
  () => data.defaultListSection,
  (nv) => {
    if (nv.length > 0) {
      setDefaultListSection(nv);
    }
  },
  { deep: true }
);
function setDefaultListSection(nv) {
  const list = nv.map((o) => o.Value);
  const lookuplabels = ["Name", "_id"];
  data.sectionparams = {};
  data.sectionparams.Take = 20;
  data.sectionparams.Sort = [lookuplabels[0]];
  data.sectionparams.Select = lookuplabels;

  data.sectionparams.Where = {
    Op: "$or",
    Items: [
      {
        Field: "_id",
        Op: "$contains",
        Value: list,
      },
      {
        Field: "Name",
        Op: "$contains",
        Value: list,
      },
    ],
  };
}

//--- aisle
watch(
  () => data.defaultListAisle,
  (nv) => {
    if (nv.length > 0) {
      setDefaultListAisle(nv);
    }
  },
  { deep: true }
);
function setDefaultListAisle(nv) {
  const list = nv.map((o) => o.Value);
  const lookuplabels = ["Name", "_id"];
  data.aisleparams = {};
  data.aisleparams.Take = 20;
  data.aisleparams.Sort = [lookuplabels[0]];
  data.aisleparams.Select = lookuplabels;

  data.aisleparams.Where = {
    Op: "$or",
    Items: [
      {
        Field: "_id",
        Op: "$contains",
        Value: list,
      },
      {
        Field: "Name",
        Op: "$contains",
        Value: list,
      },
    ],
  };
}

//--- box
watch(
  () => data.defaultListBox,
  (nv) => {
    if (nv.length > 0) {
      setDefaultListBox(nv);
    }
  },
  { deep: true }
);
function setDefaultListBox(nv) {
  const list = nv.map((o) => o.Value);
  const lookuplabels = ["Name", "_id"];
  data.boxparams = {};
  data.boxparams.Take = 20;
  data.boxparams.Sort = [lookuplabels[0]];
  data.boxparams.Select = lookuplabels;

  data.boxparams.Where = {
    Op: "$or",
    Items: [
      {
        Field: "_id",
        Op: "$contains",
        Value: list,
      },
      {
        Field: "Name",
        Op: "$contains",
        Value: list,
      },
    ],
  };
}
function setDefaultDimension(nv) {
  let site = nv.filter((item) => item.Key === "Site");
  if (site.length == 1) {
    getDefaultWarehouse(site[0].Value);
  }
  data.defaultListWarehouse = nv.filter((item) => item.Key === "WarehouseID");
  if (data.defaultListWarehouse.length === 1) {
    data.bgDimension.WarehouseID = data.defaultListWarehouse[0].Value;
  }
  data.defaultListSection = nv.filter((item) => item.Key === "SectionID");
  if (data.defaultListSection.length === 1) {
    data.bgDimension.SectionID = data.defaultListSection[0].Value;
  }

  data.defaultListAisle = nv.filter((item) => item.Key === "AisleID");
  if (data.defaultListAisle.length === 1) {
    data.bgDimension.AisleID = data.defaultListAisle[0].Value;
  }

  data.defaultListBox = nv.filter((item) => item.Key === "BoxID");
  if (data.defaultListBox.length === 1) {
    data.bgDimension.BoxID = data.defaultListBox[0].Value;
  }
}
function getDefaultWarehouse(site) {
  axios
    .post("/tenant/warehouse/find", {
      Where: {
        Op: "$and",
        Items: [
          {
            Field: "Dimension.Key",
            Op: "$eq",
            Value: "Site",
          },
          {
            Field: "Dimension.Value",
            Op: "$eq",
            Value: site,
          },
        ],
      },
    })
    .then((r) => {
      let WarehouseID = "";
      if (r.data.length > 0) {
        data.bgDimension.WarehouseID = r.data[0]._id;
        WarehouseID = r.data[0]._id;
        emit("defaultWH", WarehouseID);
      }
    });
}

onMounted(() => {
  loadFormConfig(
    axios,
    "/scm/movementtransaction/inventorydimension/formconfig"
  ).then(
    (r) => {
      r.sectionGroups.forEach((sg) => {
        sg.sections.forEach((s) => {
          let row = s.rows.filter(
            (r) => !props.hideField.includes(r.inputs[0].field)
          );
          s.rows = row;
        });
      });
      emit("alterFormConfig", r);
      data.frmCfg = r;
      util.nextTickN(2, () => {});
    },
    (e) => util.showError(e)
  );
  setDefaultDimension(props.defaultList);
});

defineExpose({
  validate,
});
</script>
