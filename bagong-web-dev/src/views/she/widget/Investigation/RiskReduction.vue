<template>
  <s-grid
    class="RiskReductionLine tb-lines"
    ref="RiskReductionLine"
    :config="data.cfgGridRiskReductionLine"
    hide-search
    hide-sort
    hide-refresh-button
    hide-edit
    hide-select
    hide-paging
    hide-new-button
    editor
    auto-commit-line
    no-confirm-delete
  >
    <template #item_LineNo="{ item }">
      <div class="grid grid-cols-1 gap-2" v-if="item.Parent">
        {{ item.LineNo }}
      </div>
      <div v-else></div>
    </template>
    <template #item_IdentifiedCause="{ item }">
      <div class="grid grid-cols-1 gap-2" v-if="item.Parent">
        <s-input v-model="item.IdentifiedCause" :disabled="true" />
        <div class="grid grid-cols-2 gap-2 pb-4">
          <s-button
            label="Add Cause"
            tooltip="Add Cause"
            class="btn_success"
            @click="addDetail(item, true)"
          />
          <s-button
            v-if="false"
            label="delete"
            tooltip="delete"
            class="btn_primary"
            @click="deleteDetail(item)"
          />
        </div>
      </div>
      <div v-else></div>
    </template>
    <template #item_KontrolPengendalian="{ item }">
      <div class="grid grid-cols-1 gap-2" v-if="item.SubParent">
        <s-input
          ref="refKontrol"
          v-model="item.KontrolPengendalian"
          class="w-full"
          use-list
          allowAdd
          lookup-url="/tenant/masterdata/find?MasterDataTypeID=RiskReduction"
          lookup-key="_id"
          :lookup-labels="['Name']"
          :lookup-searchs="['_id', 'Name']"
        ></s-input>

        <div class="grid grid-cols-2 gap-2 pb-4">
          <s-button
            label="Add Control"
            tooltip="Add Control"
            class="btn_success"
            @click="addDetailCtr(item)"
          />
          <s-button
            v-if="false"
            label="delete"
            tooltip="delete"
            class="btn_primary"
            @click="deleteDetailCtr(item)"
          />
        </div>
      </div>
      <div v-else></div>
    </template>
    <template #item_button_delete="{ item, idx }">
      <a
        v-if="!item.SubParent"
        @click="deleteSubRisk(item)"
        class="delete_action"
      >
        <mdicon
          name="delete"
          width="16"
          alt="delete"
          class="cursor-pointer hover:text-primary"
        />
      </a>
      <div v-else></div>
    </template>

    <template #item_SubKontrolPengendalian="{ item, idx }">
      <s-input
        label=""
        v-model="item.SubKontrolPengendalian"
        use-list
        :lookup-url="
          item.KontrolPengendalian
            ? '/tenant/masterdata/find?MasterDataTypeID=SubRiskReduction&ParentID=' +
              item.KontrolPengendalian
            : '/tenant/masterdata/find?MasterDataTypeID=SubRiskReduction'
        "
        lookup-key="_id"
        :lookup-labels="['Name']"
        :lookup-searchs="['_id', 'Name']"
        :key="item.KontrolPengendalian"
      />
    </template>
  </s-grid>
</template>
<script setup>
import { onMounted, inject, reactive, ref } from "vue";
import { loadGridConfig, util, SGrid, SButton, SInput } from "suimjs";
const axios = inject("axios");

const props = defineProps({
  modelValue: { type: Object, default: () => {} },
  item: { type: Object, default: () => {} },
});

const emit = defineEmits({
  "update:modelValue": null,
  recalc: null,
});

const RiskReductionLine = ref(null);

const data = reactive({
  value: props.modelValue,
  record: props.modelValue,
  cfgGridRiskReductionLine: {},
});

function loadGridRiskReductionLine() {
  let url = `/she/investigasi/riskreduction/gridconfig`;
  loadGridConfig(axios, url).then(
    (r) => {
      data.cfgGridRiskReductionLine = r;
      updateGridLine(data.record.RiskReduction, "RiskReduction");
    },
    (e) => {}
  );
}

function updateGridLine(record, type) {
  //   record.map((obj, idx) => {
  //     obj.LineNo = parseInt(idx) + 1;
  //     return obj;
  //   });
  updateGridLines();
}

function addDetail(r, SubParent) {
  let obj = {};
  obj._id = util.uuid();
  obj.Parent = false;
  obj.LineNo = r.LineNo;
  obj.SourceId = r.SourceId;
  obj.IdentifiedCause = r.IdentifiedCause;
  obj.SubParent = SubParent;
  obj.KontrolPengendalian = "";
  obj.SubKontrolPengendalian = "";
  obj.Remark = "";
  obj.ControlNo = util.uuid();

  const index = data.value.RiskReduction.findIndex(
    (item) => item._id === r._id
  );

  const raw = data.value.RiskReduction.filter(
    (item) => item.SourceId === r.SourceId
  ).length;

  data.value.RiskReduction.splice(index + raw, 0, obj);
  updateGridLines();
}

function deleteDetail(r) {
  data.value.RiskReduction = data.value.RiskReduction.filter(
    (obj) => obj.LineNo !== r.LineNo
  );
  let no = 1;
  for (let i in data.value.RiskReduction) {
    let obj = data.value.RiskReduction[i];
    if (obj.Parent) {
      let LineNo = no++;
      obj.LineNo = LineNo;
    } else {
      obj.LineNo = data.value.RiskReduction[parseInt(i) - 1].LineNo;
    }
  }
  updateGridLines();
}

function addDetailCtr(r) {
  let obj = JSON.parse(JSON.stringify(r));
  obj._id = util.uuid();
  obj.SubKontrolPengendalian = "";
  obj.Remark = "";
  obj.Parent = false;
  obj.SubParent = false;

  let RiskReduction = data.value.RiskReduction.sort((a, b) => {
    if (a.LineNo !== b.LineNo) {
      return a.LineNo - b.LineNo;
    }
    return b.Parent - a.Parent;
  });

  const grouped = RiskReduction.reduce((acc, item) => {
    const key = item.ControlNo;
    if (!acc[key]) {
      acc[key] = [];
    }
    acc[key].push(item);
    return acc;
  }, {});
  let dt = [];
  for (const key in grouped) {
    if (key == obj.ControlNo) {
      grouped[key].push(obj);
    }
    dt = [...dt, ...grouped[key]];
  }

  data.value.RiskReduction = dt;
  RiskReductionLine.value.setRecords(data.value.RiskReduction);
  util.nextTickN(3, () => {
    formatRowSpan();
  });
}

function deleteDetailCtr(r) {
  deleteDetail(r);
}

function deleteSubRisk(r) {
  data.value.RiskReduction = data.value.RiskReduction.filter(
    (obj) => obj._id !== r._id
  );
  updateGridLines();
}

function updateGridLines() {
  //   data.value.RiskReduction = data.value.RiskReduction.sort((a, b) => {
  //     if (a.LineNo !== b.LineNo) {
  //       return a.LineNo - b.LineNo;
  //     }
  //     return b.Parent - a.Parent;
  //   });
  RiskReductionLine.value.setRecords(data.value.RiskReduction);
  util.nextTickN(3, () => {
    formatRowSpan();
  });
}

function formatRowSpan() {
  const myTable = document.querySelectorAll(
    ".tb-lines .suim_table [name='grid_body']"
  );
  if (myTable.length == 0) {
    return;
  }
  let irow = 0;
  let rrow = 0;
  for (let i = 0, row; (row = myTable[0].rows[i]); i++) {
    const firstCell = row.cells[0];
    firstCell.classList.remove("hidden");
    if (data.value.RiskReduction[i].Parent) {
      let lengthRow = data.value.RiskReduction.filter(
        (o) => o.LineNo == data.value.RiskReduction[i].LineNo
      );
      firstCell.rowSpan = lengthRow.length;

      const grouped = lengthRow.reduce((acc, item) => {
        const key = item.ControlNo;
        if (!acc[key]) {
          acc[key] = [];
        }
        acc[key].push(item);
        return acc;
      }, {});

      for (const key in grouped) {
        for (let idx = 0; idx < grouped[key].length; idx++) {
          myTable[0].rows[rrow].cells[1].classList.remove("hidden");
          rrow++;
        }
      }

      for (const key in grouped) {
        let rowSpan = grouped[key].length;
        for (let idx = 0; idx < grouped[key].length; idx++) {
          if (idx == 0) {
            myTable[0].rows[irow].cells[1].rowSpan = rowSpan;
          } else {
            myTable[0].rows[irow].cells[1].classList.add("hidden");
          }
          irow++;
        }
      }
    } else {
      firstCell.classList.add("hidden");
    }
  }
}

onMounted(() => {
  loadGridRiskReductionLine();
});
defineExpose({
  updateGridLine,
  updateGridLines,
});
</script>
