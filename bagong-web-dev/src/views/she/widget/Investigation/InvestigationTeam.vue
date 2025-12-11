<template>
  <s-form
    ref="InvestigationTeamFormCtl"
    v-model="data.record"
    :config="data.formCfg"
    keep-label
    only-icon-top
    :buttons-on-top="false"
    hide-cancel
    auto-focus
  >
    <template #input_TeamMember="{ item, idx }">
      <s-grid
        class="team-member-line"
        ref="TeamMemberLine"
        :config="data.cfgGridTeamMemberLine"
        hide-search
        hide-sort
        hide-refresh-button
        hide-edit
        hide-select
        hide-paging
        editor
        auto-commit-line
        no-confirm-delete
        @new-data="newTeamMemberLine"
        @row-field-changed="handleChanged"
      >
        <template #item_JobTittle="{ item, idx }">
          <s-input
            label=""
            v-model="item.JobTittle"
            use-list
            lookup-url="/tenant/masterdata/find?MasterDataTypeID=PTE"
            lookup-key="_id"
            :lookup-labels="['Name']"
            :lookup-searchs="['_id', 'Name']"
            :key="data.keyJobTitle"
          />
        </template>
        <template #item_button_delete="{ item, idx }">
          <a @click="deleteTeamMemberLine(item)" class="delete_action">
            <mdicon
              name="delete"
              width="16"
              alt="delete"
              class="cursor-pointer hover:text-primary"
            />
          </a>
        </template>
      </s-grid>
    </template>
    <template #input_CheckListDirection="{ item, idx }">
      <s-grid
        class="team-member-line"
        ref="CheckListDirectionLine"
        :config="data.cfgGridCheckListDirectionLine"
        hide-search
        hide-sort
        hide-refresh-button
        hide-edit
        hide-select
        hide-paging
        editor
        auto-commit-line
        no-confirm-delete
        @new-data="newCheckListDirectionLine"
      >
        <template #item_button_delete="{ item, idx }">
          <a @click="deleteCheckListDirectionLine(item)" class="delete_action">
            <mdicon
              name="delete"
              width="16"
              alt="delete"
              class="cursor-pointer hover:text-primary"
            />
          </a>
        </template>
      </s-grid>
    </template>
  </s-form>
</template>
<script setup>
import { onMounted, inject, reactive, ref } from "vue";
import {
  loadGridConfig,
  loadFormConfig,
  util,
  SGrid,
  SForm,
  SInput,
} from "suimjs";
const axios = inject("axios");

const props = defineProps({
  modelValue: { type: Object, default: () => {} },
  item: { type: Object, default: () => {} },
});

const emit = defineEmits({
  "update:modelValue": null,
  recalc: null,
});

const InvestigationTeamFormCtl = ref(null);
const TeamMemberLine = ref(null);
const CheckListDirectionLine = ref(null);

const data = reactive({
  record: props.modelValue,
  formCfg: {},
  cfgGridTeamMemberLine: {},
  cfgGridCheckListDirectionLine: {},
  keyJobTitle: 0,
});

function loadFromInvestigationTeam() {
  let url = `/she/investigasi/investigationteam/formconfig`;
  loadFormConfig(axios, url).then(
    (r) => {
      data.formCfg = r;
      loadGridTeamMemberLine();
      loadGridCheckListDirectionline();
    },
    (e) => {}
  );
}

function loadGridTeamMemberLine() {
  let url = `/she/investigasi/teammemberline/gridconfig`;
  loadGridConfig(axios, url).then(
    (r) => {
      console.log(r);
      data.cfgGridTeamMemberLine = r;
      updateGridLine(data.record.TeamMember, "TeamMember");
    },
    (e) => {}
  );
}

function loadGridCheckListDirectionline() {
  let url = `/she/investigasi/checklistdirectionline/gridconfig`;
  loadGridConfig(axios, url).then(
    (r) => {
      data.cfgGridCheckListDirectionLine = r;
      updateGridLine(data.record.CheckListDirection, "CheckListDirection");
    },
    (e) => {}
  );
}

function newTeamMemberLine() {
  let r = {};
  const noLine = data.record.TeamMember.length + 1;
  r._id = util.uuid();
  r.LineNo = noLine;
  r.Role = "";
  r.Name = "";
  r.JobTittle = "";
  r.Description = "";
  data.record.TeamMember.push(r);
  updateGridLine(data.record.TeamMember, "TeamMember");
}
function deleteTeamMemberLine(r) {
  data.record.TeamMember = data.record.TeamMember.filter(
    (obj) => obj._id !== r._id
  );
  updateGridLine(data.record.TeamMember, "TeamMember");
}
function newCheckListDirectionLine() {
  let r = {};
  const noLine = data.record.CheckListDirection.length + 1;
  r._id = util.uuid();
  r.LineNo = noLine;
  r.Role = "";
  r.Name = "";
  r.JobTittle = "";
  r.Description = "";
  data.record.CheckListDirection.push(r);
  updateGridLine(data.record.CheckListDirection, "CheckListDirection");
}
function deleteCheckListDirectionLine(r) {
  data.record.CheckListDirection = data.record.CheckListDirection.filter(
    (obj) => obj._id !== r._id
  );
  updateGridLine(data.record.CheckListDirection, "CheckListDirection");
}

function updateGridLine(record, type) {
  record.map((obj, idx) => {
    obj.LineNo = parseInt(idx) + 1;
    return obj;
  });
  if (type == "TeamMember") {
    TeamMemberLine.value.setRecords(record);
  } else if (type == "CheckListDirection") {
    CheckListDirectionLine.value.setRecords(record);
  }
}

function handleChanged(field, v1, v2, old, record) {
  switch (field) {
    case "Name":
      record.JobTittle = "";
      if (v1) {
        axios.post("/bagong/employee/get", [v1]).then(
          (r) => {
            let emp = r.data;
            record.JobTittle = emp.Detail.Position;
            data.keyJobTitle++;
          },
          (e) => util.showError(e)
        );
      }
      break;
      record.Unit = "";
      record.Type = "";
      record.Damage = "";
      record.PartEquipment = "";
      record.TotalCost = 0;
      record.Remark = "";
      if (v1) {
        axios.post("/tenant/asset/get", [v1]).then(
          (r) => {
            let asset = r.data;
            record.Unit = asset.AssetType;
            record.Type = asset.DriveType;
          },
          (e) => util.showError(e)
        );
      }
      break;
    default:
      break;
  }
}

onMounted(() => {
  loadFromInvestigationTeam();
});
defineExpose({});
</script>
