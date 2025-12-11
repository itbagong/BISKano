<template>
  <template v-if="data.loading">
    <div class="min-h-full w-full flex items-center justify-center">
      <loader kind="circle" />
    </div>
  </template>
  <template v-else>
    <template v-if="props.stage != ''">
      <div class="border-b px-2 mb-4">
        <Header
          @apply="applyFilter"
          @check-uncheck="onCheckUncheckAll"
          :stage="stage"
        />
      </div>
      <s-list
        :key="data.key"
        ref="listCtl"
        class="w-full slist_tracking_candidate"
        hide-delete-button
        hide-sort
        hide-control
        v-model="data.recordsShowing"
      >
        <template #item="{ item }">
          <div
            class="flex p-2 items-center hover:bg-slate-200 rounded-md"
            :class="[data.selected.stageId == item._id ? 'bg-slate-200' : '']"
          >
            <div>
              <input v-if="item.Status == 'Passed'" type="checkbox" disabled />
              <input
                v-else
                type="checkbox"
                @click="onCheckUncheck(item)"
                :checked="Object.hasOwn(data.checkedMapIds, item._id)"
              />
            </div>
            <div class="grow grid grid-cols-1 gap-2 pl-2">
              <div class="flex gap-2 items-center" @click="onSelect(item)">
                <Photo v-model="item.CandidateID"></Photo>
                <div class="grid grid-cols-2 gap-2 w-full">
                  <div class="[&>*]:truncate grow">
                    <div>{{ item?.Name == "" ? "-" : item.Name }}</div>
                    <div
                      class="text-[0.8em] text-gray-500 flex items-center gap-1"
                    >
                      <mdicon
                        name="checkbox-blank-circle"
                        class="circle text-gray-400"
                        size="8"
                      />
                      {{ item?.Education == "" ? "-" : item.Education }}
                    </div>
                  </div>
                  <div class="[&>*]:truncate pl-[15px]">
                    <div :class="[item.Status]" class="flex items-center gap-1">
                      <mdicon
                        name="checkbox-blank-circle"
                        class="circle text-gray-400"
                        size="8"
                      />
                      {{ item.Status }}
                    </div>
                    <div
                      class="text-[0.8em] text-gray-500 flex items-center gap-1"
                    >
                      <mdicon
                        name="checkbox-blank-circle"
                        class="circle text-gray-400"
                        size="8"
                      />
                      {{ item.Age }} Years
                    </div>
                  </div>
                </div>
              </div>
              <div
                class="text-[0.76em] flex gap-1 items-center italic"
                v-if="stage == 'PshycologicalTest' && item.IsTestSent"
              >
                <mdicon name="send" size="10" />
                Psikotest is alreadey sent
              </div>
              <div
                class="text-[0.76em] flex gap-1 items-center italic"
                v-if="stage == 'MCU' && item.MCUTransactionID"
              >
                <mdicon name="medication" size="10" />
                MCU is created
              </div>
            </div>
          </div>
        </template>

        <template #paging="{}">
          <s-pagination
            :recordCount="pagination.recordCount"
            :pageCount="pagination.pageCount"
            :current-page="pagination.currentPage"
            :page-size="pagination.pageSize"
            @changePage="changePage"
          ></s-pagination>
        </template>

        <template #footer_2="{}">
          <buttons
            :map-ids="data.checkedMapIds"
            :stage="stage"
            :man-power-id="manPowerId"
            @refresh-list="refresh"
          />
        </template>
      </s-list>
    </template>
    <div
      v-else
      class="min-h-full flex items-center text-[1.5em] justify-center font-bold text-gray-400 mb-4"
    >
      Please select step
    </div>
  </template>
</template>
<script setup>
import { reactive, ref, inject, onMounted, watch } from "vue";
import { SList, util, SPagination, SButton } from "suimjs";
import helper from "@/scripts/helper.js";
import Loader from "@/components/common/Loader.vue";
import Header from "./TrackingCandidateHeader.vue";
import Buttons from "./TrackingCandidateButtons.vue";
import Photo from "../EmployeePhoto.vue";

const props = defineProps({
  modelValue: { type: Object, default: { id: "", stageId: "" } },
  manPowerId: { type: String, default: "" },
  stage: { type: String, default: "" },
});

const emit = defineEmits({
  select: null,
  "update:modelValue": null,
});

const axios = inject("axios");
const listCtl = ref(null);

const data = reactive({
  selected:
    props.modelValue == undefined
      ? { id: "", stageId: "", item: null }
      : props.modelValue,
  records: [],
  recordsShowing: [],
  isCheckedAll: false,
  checkedMapIds: {},
  filter: null,
  key: 0,
  loading: false,
});

const pagination = reactive({
  recordCount: 0,
  pageCount: 0,
  currentPage: 1,
  pageSize: 10,
});

function applyFilter(filter) {
  pagination.currentPage = 1;
  data.filter = filter;
  util.nextTickN(2, () => {
    mappingRecords();
  });
}

function getFilterRecords() {
  return helper.cloneObject(data.records).filter((e) => {
    if (data.filter == null) return true;

    const filterAgeFrom = parseInt(data.filter.AgeFrom);
    const filterAgeTo = parseInt(data.filter.AgeTo);

    if (
      data.filter.Name != "" &&
      !e.Name?.toLowerCase()?.includes(data.filter.Name.toLowerCase())
    )
      return false;

    if (!isNaN(filterAgeFrom)) {
      if (filterAgeFrom > e.Age) return false;
    }

    if (!isNaN(filterAgeTo)) {
      if (filterAgeTo < e.Age) return false;
    }

    if (data.filter.Status.length > 0 && !data.filter.Status.includes(e.Status))
      return false;
    return true;
  });
}

function changePage(page) {
  pagination.currentPage = page;
  mappingRecords();
}

function mappingRecords() {
  const dt = getFilterRecords();
  pagination.recordCount = dt.length;
  pagination.pageCount = Math.ceil(
    pagination.recordCount / pagination.pageSize
  );

  const start = (pagination.currentPage - 1) * pagination.pageSize;
  const end =
    pagination.currentPage == pagination.pageCount
      ? pagination.recordCount
      : pagination.pageSize;

  const records = dt.splice(start, end);
  data.recordsShowing = [...records];

  util.nextTickN(2, () => {
    data.key++;
  });
}

function fetchRecords() {
  data.filter = null
  data.loading = true;

  pagination.currentPage = 1;

  axios
    .post("/hcm/tracking/get-applicant", {
      Where: { Stage: props.stage, JobID: props.manPowerId },
    })
    .then((r) => {
      data.records = r.data.data;
    })
    .catch((e) => {
      data.records = [];
      util.showError(e);
    })
    .finally(() => {
      data.loading = false;
      mappingRecords();
    });
}

function refresh() {
  data.selected = { id: "", stageId: "", item: null };
  data.checkedMapIds = [];
  data.isCheckedAll = false;

  if (props.stage != "") {
    fetchRecords();
  } else {
    data.records = [];
    util.nextTickN(2, () => {
      data.key++;
    });
  }
}

function onSelect(r) {
  data.selected = { id: r.CandidateID, stageId: r._id, item: r };
}

function onCheckUncheck(r) {
  if (Object.hasOwn(data.checkedMapIds, r._id)) {
    delete data.checkedMapIds[r._id];
  } else {
    let obj = { ...data.checkedMapIds };
    obj[r._id] = r;
    data.checkedMapIds = obj;
  }
}

function onCheckUncheckAll() {
  data.isCheckedAll = !data.isCheckedAll;
  data.checkedMapIds = {};
  if (data.isCheckedAll) {
    let obj = {};
    if (data.filter == null) {
      data.records.filter(o => o.Status !== 'Passed').forEach((e) => {
        obj[e._id] = e;
      });
    } else {
      const dt = getFilterRecords();
      dt.filter(o => o.Status !== 'Passed').forEach((e) => {
        obj[e._id] = e;
      });
    }
    
    data.checkedMapIds = obj;
  }
}

watch(
  () => props.stage,
  () => {
    refresh();
  }
);

watch(
  () => data.selected,
  (nv) => {
    emit("select", nv);
  }
);

defineExpose({
  refresh,
});

onMounted(() => {
  refresh();
});
</script>

<style>
.slist_tracking_candidate ul {
  @apply grid-cols-1 gap-0 overflow-auto max-h-[calc(100vh-450px)];
}
.slist_tracking_candidate ul li {
  @apply max-h-[90px];
}
/* .slist_tracking_candidate .suim_pagination .info_page{
    display: none;
} */
.slist_tracking_candidate .suim_pagination {
  @apply grid grid-cols-1 justify-items-center gap-y-2;
}
.slist_tracking_candidate .suim_pagination .pagesize {
  display: none;
}
</style>

<style scoped>
.Failed,
.Failed > span {
  color: #e74c3c;
}
.Passed,
.Passed > span {
  color: #27ae60;
}
.Not.Selected,
.Not.Selected > span {
  color: #392e4a;
}
</style>
