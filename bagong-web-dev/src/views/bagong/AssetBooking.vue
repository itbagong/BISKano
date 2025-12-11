<template>
  <div class="w-full">
    <data-list
      v-if="!data.toAllocated"
      class="card"
      ref="listControl"
      title="Asset Booking"
      grid-config="/bagong/assetbooking/gridconfig"
      form-config="/bagong/assetbooking/formconfig"
      grid-read="/bagong/assetbooking/gets"
      form-read="/bagong/assetbooking/get"
      grid-mode="grid"
      grid-delete="/bagong/assetbooking/delete"
      form-keep-label
      form-insert="/bagong/assetbooking/insert"
      form-update="/bagong/assetbooking/update"
      :form-fields="['Dimension']"
      :form-tabs-new="data.tabs"
      :form-tabs-edit="data.tabs"
      :init-app-mode="data.appMode"
      :init-form-mode="data.formMode"
      stay-on-form-after-save
      @form-new-data="newRecord"
      @form-edit-data="openForm"
      @post-save="onPostSave"
      @grid-row-deleted="onGridRowDeleted"
    >
      <template #grid_header_buttons_1>
        <s-button
          class="btn_primary"
          label="Allocated"
          @click="onClickAllocated"
        />
      </template>

      <template #form_input_Dimension="{ item }">
        <dimension-editor-vertical
          v-model="item.Dimension"
        ></dimension-editor-vertical>
      </template>

      <template #form_tab_Lines="{ item }">
        <asset-booking-lines
          v-model="item.Lines"
          :AssetBookingID="item._id"
          :allocationAction="data.allocatedAction"
          @recalc="calcLines"
          @click-allocated-lines="clickLinesAllocatedForm"
        ></asset-booking-lines>
      </template>
    </data-list>

    <s-card
      v-else
      title="Asset Booking Allocation"
      class="w-full bg-white suim_datalist"
      hide-footer
    >
      <s-grid
        v-if="!data.toAllocatedEdit"
        class="w-full se-grid"
        ref="gridAllocation"
        hide-select
        hide-search
        hide-sort
        hide-refresh-button
        hide-new-button
        hide-detail
        hide-delete-button
        :config="data.gridCfg"
      >
        <template #header_buttons_2>
          <s-button
            class="btn_warning back_btn"
            label="Back"
            icon="rewind"
            @click="onBack"
          />
        </template>

        <template #item_buttons_1="{ item }">
          <s-button
            class="btn_primary mx-1 my-2"
            label="Allocated"
            @click="clickAllocatedForm(item)"
          />
        </template>
      </s-grid>

      <s-grid
        v-else
        class="w-full se-grid"
        ref="gridAllocationEdit"
        editor
        hide-select
        hide-search
        hide-sort
        hide-refresh-button
        hide-detail
        auto-commit-line
        no-confirm-delete
        :config="data.gridCfgAllocated"
        @delete-data="onDeleteAllocated"
      >
        <template #header_buttons>
          <s-button
            class="btn_primary"
            label="Add"
            icon="plus"
            @click="newRecordAllocated"
          />
          <s-button
            class="btn_primary"
            label="Save"
            icon="content-save"
            @click="onSaveAllocated"
          />
          <s-button
            class="btn_warning back_btn"
            label="Back"
            icon="rewind"
            @click="onBack"
          />
        </template>
      </s-grid>
    </s-card>
  </div>
</template>

<script setup>
import { reactive, ref, watch, inject, onMounted } from "vue";
import { layoutStore } from "@/stores/layout.js";
import { DataList, SCard, SGrid, SButton, util, loadGridConfig } from "suimjs";
import AssetBookingLines from "./widget/AssetBookingLines.vue";
import DimensionEditorVertical from "@/components/common/DimensionEditorVertical.vue";

layoutStore().name = "tenant";

const emit = defineEmits({
  "update:modelValue": null,
});

const listControl = ref(null);
const gridAllocation = ref(null);
const gridAllocationEdit = ref(null);

const axios = inject("axios");

const data = reactive({
  appMode: "grid",
  formMode: "edit",
  tabs: ["General", "Lines"],
  record: null,
  collectedLines: [],
  gridCfg: null,
  toAllocated: false,
  toAllocatedEdit: false,
  gridCfgAllocated: null,
  selectedLines: {},
  allocatedFrom: "general",
  allocatedAction: false,
});

function newRecord(record) {
  record._id = "";
  record.FromDate = new Date().toISOString();
  record.ToDate = new Date().toISOString();
  record.Unit = 0;
  record.Amount = 0;
  record.UnitBooked = 0;
  record.UnitAllocated = 0;

  openForm(record);
}

function openForm(record) {
  record.Lines = record.Lines
    ? record.Lines.map((item, index) => ({
        index,
        AssetBookingID: record._id,
        ...item,
      }))
    : [];
  data.record = record;

  util.nextTickN(2, () => {
    const formMode = listControl.value.getFormMode();
    if (formMode == "edit") {
      data.allocatedAction = true;
    } else {
      data.allocatedAction = false;
    }

    listControl.value.setFormFieldAttr("_id", "rules", [
      (v) => {
        let vLen = 0;
        let consistsInvalidChar = false;

        v.split("").forEach((ch) => {
          vLen++;
          const validCar =
            "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789abcdefghijklmnopqrstuvwxyz-_".indexOf(
              ch
            ) >= 0;
          if (!validCar) consistsInvalidChar = true;
        });

        if (vLen < 3 || consistsInvalidChar)
          return "minimal length is 3 and alphabet only";
        return "";
      },
    ]);
  });
}

function calcLines() {
  const record = data.record;
  record.Amount = record.Lines.reduce((a, b) => {
    return a + b.Total;
  }, 0);
  record.UnitBooked = record.Lines.reduce((a, b) => {
    return a + b.UnitBooked;
  }, 0);
}

function onClickAllocated() {
  const getSelected = listControl.value.getGridSelected();
  const selected = getSelected();
  const collectedLines = selected.value
    .map((item) =>
      item.Lines.map((line, index) => ({
        index,
        AssetBookingID: item._id,
        ...line,
      }))
    )
    .reduce((acc, curr) => acc.concat(curr), []);

  data.toAllocated = true;
  data.collectedLines = collectedLines;

  setTimeout(() => {
    gridAllocation.value.setRecords(data.collectedLines);
  }, 500);
}

function onBack() {
  data.toAllocated = false;
  data.toAllocatedEdit = false;
  data.collectedLines = [];

  if (data.allocatedFrom == "lines") {
    const records = gridAllocationEdit.value.getRecords();
    const recData = data.record;
    recData.Lines[data.selectedLines.index].UnitAllocated = records.length;
    recData.UnitAllocated = recData.Lines.reduce((acc, curr) => {
      return acc + curr.UnitAllocated;
    }, 0);

    setTimeout(() => {
      listControl.value.setControlMode("form");
      listControl.value.setFormMode("edit");
      listControl.value.setFormRecord(recData);
      data.allocatedFrom = "general";
    }, 500);
  }
}

function clickLinesAllocatedForm(item) {
  data.allocatedFrom = item.from;
  data.toAllocated = true;
  if (data.toAllocated) {
    data.toAllocatedEdit = true;
    data.selectedLines = item.data;
    axios
      .post(
        `/bagong/assetbooking-allocation/gets?AssetBookingID=${item.data.AssetBookingID}`,
        { Sort: ["_id"] }
      )
      .then(
        (r) => {
          const records = r.data.data.filter(
            (v) => v.LinesIdx == item.data.index
          );
          gridAllocationEdit.value.setRecords(records);
        },
        (e) => {
          util.showError(e);
        }
      );
  }
}

function clickAllocatedForm(item) {
  data.toAllocatedEdit = true;
  data.selectedLines = item;
  axios
    .post(
      `/bagong/assetbooking-allocation/gets?AssetBookingID=${item.AssetBookingID}`,
      { Sort: ["_id"] }
    )
    .then(
      (r) => {
        const records = r.data.data.filter((v) => v.LinesIdx == item.index);
        gridAllocationEdit.value.setRecords(records);
      },
      (e) => {
        util.showError(e);
      }
    );
}

function newRecordAllocated() {
  const record = {};
  record._id = "";
  record.AssetBookingID = data.selectedLines.AssetBookingID;
  record.AssetID = "";
  record.Utilization = "";
  record.Notes = "";

  gridAllocationEdit.value.setRecords([
    ...gridAllocationEdit.value.getRecords(),
    record,
  ]);
}

async function onSaveAllocated() {
  const records = gridAllocationEdit.value.getRecords();
  records.forEach(async (item) => {
    await axios
      .post("/bagong/assetbooking-allocation/save", {
        ...item,
        LinesIdx: data.selectedLines.index,
      })
      .then(
        (r) => {},
        (e) => {
          util.showError(e);
        }
      );
  });

  await axios
    .post("/bagong/assetbooking/get", [data.selectedLines.AssetBookingID])
    .then(
      async (r) => {
        const recData = r.data;
        recData.Lines[data.selectedLines.index].UnitAllocated = records.length;
        recData.UnitAllocated = recData.Lines.reduce((acc, curr) => {
          return acc + curr.UnitAllocated;
        }, 0);

        await axios.post("/bagong/assetbooking/update", recData).then(
          (r) => {
            util.showInfo("Data has been saved");
          },
          (e) => {
            util.showError(e);
          }
        );
      },
      (e) => {
        util.showError(e);
      }
    );
}

async function onDeleteAllocated(_record, index) {
  const records = gridAllocationEdit.value.getRecords();

  await axios
    .post("/bagong/assetbooking-allocation/delete", records[index])
    .then(
      (r) => {},
      (e) => {
        util.showError(e);
      }
    );

  const newRecords = records.filter((dt, idx) => {
    return idx != index;
  });

  await axios
    .post("/bagong/assetbooking/get", [data.selectedLines.AssetBookingID])
    .then(
      async (r) => {
        const recData = r.data;
        recData.Lines[data.selectedLines.index].UnitAllocated =
          newRecords.length;
        recData.UnitAllocated = recData.Lines.reduce((acc, curr) => {
          return acc + curr.UnitAllocated;
        }, 0);

        await axios.post("/bagong/assetbooking/update", recData).then(
          (r) => {},
          (e) => {
            util.showError(e);
          }
        );
      },
      (e) => {
        util.showError(e);
      }
    );

  gridAllocationEdit.value.setRecords(newRecords);
}

function onPostSave() {
  setTimeout(() => {
    const formMode = listControl.value.getFormMode();
    if (formMode == "edit") {
      data.allocatedAction = true;
    } else {
      data.allocatedAction = false;
    }
  }, 500);
}

function onGridRowDeleted(params) {
  const url =
    "/bagong/assetbooking-allocation/find?AssetBookingID=" + params._id;
  axios.post(url).then(
    (r) => {
      if (r.data.length > 0) {
        r.data.forEach((item) => {
          onDeleteDetail(item);
        });
      }
    },
    (e) => {
      util.showError(e);
    }
  );
}

function onDeleteDetail(id) {
  const url = "/bagong/assetbooking-allocation/delete";
  axios.post(url, id).then(
    (r) => {},
    (e) => {
      util.showError(e);
    }
  );
}

onMounted(() => {
  loadGridConfig(axios, "/bagong/assetbooking/lines/gridconfig").then(
    (r) => {
      const newField = {
        field: "AssetBookingID",
        kind: "text",
        label: "BookingID",
        labelField: "",
        readType: "show",
        input: {
          field: "AssetBookingID",
          label: "BookingID",
          hint: "",
          hide: false,
          placeHolder: "BookingID",
          kind: "text",
          disable: false,
          required: false,
          multiple: false,
        },
      };
      r.fields = [{ ...newField }, ...r.fields];
      data.gridCfg = r;
    },
    (e) => util.showError(e)
  );

  loadGridConfig(axios, "/bagong/assetbooking-allocation/gridconfig").then(
    (r) => {
      data.gridCfgAllocated = r;
    },
    (e) => util.showError(e)
  );
});
</script>
