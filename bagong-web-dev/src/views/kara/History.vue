<template>
  <div>
    <s-modal
      :display="false"
      ref="confirmModal"
      hide-buttons
      title="Need review"
    >
      <div>
        Are you sure to approve the attendance?
        <div class="mt-5">
          <s-input
            class="w-full"
            label="Note"
            v-model="data.noteId"
            :lookup-url="`/tenant/masterdata/find?MasterDataTypeID=AttendanceNote`"
            field="_id"
            lookup-key="_id"
            :lookup-labels="['Name']"
            :lookup-searchs="['_id', 'Name']"
            use-list
          />
        </div>
        <div class="mt-5 flex gap-5">
          <s-button
            class="bg-primary text-white font-bold w-full flex justify-center"
            label="Approve"
            @click="onReview('OK')"
          ></s-button>
          <s-button
            class="btn_warning text-white font-bold w-full flex justify-center"
            label="Reject"
            @click="onReview('Rejected')"
          ></s-button>
        </div>
      </div>
    </s-modal>
    <s-modal
      v-if="data.showAddress"
      title="Position"
      display
      ref="address"
      @beforeHide="
        () => {
          data.showAddress = false;
        }
      "
      hideButtons
    >
      <div class="min-w-[500px] w-[500px] max-h-[600px] overflow-auto">
        <p>{{ data.address?.display_name }}</p>
      </div>
    </s-modal>
    <data-list
      class="card"
      ref="listControl"
      grid-mode="grid"
      form-keep-label
      title="Attendance History"
      grid-hide-new
      grid-hide-detail
      grid-hide-delete
      grid-config="/kara/trx/gridconfig"
      grid-read="/bagong/admin/trx/gets"
      :grid-fields="['Lat', 'Long']"
      :grid-custom-filter="data.customFilter"
      @gridResetCustomFilter="resetGridHeaderFilter"
    >
      <template #grid_header_search>
        <grid-header-filter
          ref="gridHeaderFilter"
          v-model="data.customFilter"
          hide-filter-text
          hide-filter-status
          @init-new-item="initNewItemFilter"
          @pre-change="changeFilter"
          @change="refreshGrid"
        >
          <template #filter_1="{ item }">
            <s-input
              class="w-[200px]"
              keep-label
              label="Search Name"
              v-model="item.Keyword"
            />
          </template>
          <template #filter_2="{ item }">
            <s-input
              class="w-[200px]"
              label="Op"
              v-model="item.Op"
              use-list
              :items="['Checkin', 'Checkout']"
            />
            <s-input
              class="w-[200px]"
              label="Direct Supervisor"
              v-model="item.DirectSupervisor"
              use-list
              lookup-url="/tenant/employee/find"
              lookup-key="_id"
              :lookup-labels="['Name']"
              :lookup-searchs="['_id', 'Name']"
              multiple
            />
          </template>
        </grid-header-filter>
      </template>
      <template #grid_Lat="{ item, header }">
        <span>{{ item.Lat }}</span>
      </template>
      <template #grid_Long="{ item, header }">
        <span>{{ item.Long }}</span>
      </template>
      <template #grid_item_buttons_1="{ item }">
        <a href="#" @click="getAddress(item)" class="edit_action">
          <mdicon
            name="information"
            width="16"
            alt="image"
            class="cursor-pointer hover:text-primary"
          />
        </a>
        <div>
          <div v-if="data.previewLoading && data.selectedId === item._id">
            loading...
          </div>
          <a v-else href="#" @click="showPhoto(item)" class="edit_action">
            <mdicon
              name="camera"
              width="16"
              alt="image"
              class="cursor-pointer hover:text-primary"
            />
          </a>
        </div>
        <a
          v-if="item.Status === 'Need review'"
          href="#"
          @click="showConfirm(item)"
          class="edit_action"
        >
          <mdicon
            name="check"
            width="16"
            alt="image"
            class="cursor-pointer hover:text-primary"
          />
        </a>
      </template>
    </data-list>
  </div>
</template>

<script setup>
import { reactive, ref, inject } from "vue";
import { authStore } from "@/stores/auth.js";
import { layoutStore } from "@/stores/layout.js";
import { DataList, SModal, util, SButton, SInput } from "suimjs";
import { api as viewerApi } from "v-viewer";
import GridHeaderFilter from "@/components/common/GridHeaderFilter.vue";

layoutStore().name = "tenant";

const FEATUREID = "Attendance";
const profile = authStore().getRBAC(FEATUREID);

const axios = inject("axios");
const api_url = import.meta.env.VITE_API_URL;

const listControl = ref(null);
const confirmModal = ref(null);
const gridHeaderFilter = ref(null);

const data = reactive({
  formCfg: null,
  previewLoading: false,
  selectedId: "",
  selectedRecord: {},
  customFilter: null,
  noteId: "",
  showAddress: false,
  address: null,
});

const showPhoto = (record) => {
  data.previewLoading = true;
  data.selectedId = record._id;
  axios
    .post("/asset/read-by-journal", {
      JournalType: "Attendance",
      JournalID: record._id,
    })
    .then(
      (r) => {
        const file = r.data[0];
        if (file) {
          onPreviewImg(file._id);
        }
        data.previewLoading = false;
      },
      (e) => {
        util.showError(e);
        data.previewLoading = false;
      }
    )
    .catch((e) => {
      util.showError(e);
      data.previewLoading = false;
    });
};
const showConfirm = (record) => {
  data.noteId = "";
  data.selectedRecord = record;
  util.nextTickN(2, () => {
    confirmModal.value.show();
  });
};
const getAddress = async (record) => {
  const url = `https://nominatim.openstreetmap.org/reverse?format=json&lat=${record.Lat}&lon=${record.Long}`;
  try {
    // const response = await axios.get(url);
    const response = await fetch(url);
    const resData = await response.json();
    data.address = resData;
    console.log(data, data.showAddress);
    data.showAddress = true;
  } catch (error) {
    console.error("Error fetching location details:", error);
  }
};
const onReview = (status) => {
  axios
    .post("/kara/admin/trx/save", {
      ...data.selectedRecord,
      Status: status,
      Message: data.noteId,
    })
    .then((r) => {
      listControl.value.refreshList();
    })
    .finally(() => confirmModal.value.hide());
};
function onPreviewImg(id) {
  viewerApi({
    images: [`${api_url}/asset/view?id=${id}`],
  });
}

function initNewItemFilter(item) {
  item.Keyword = "";
  item.Op = "";
  item.DirectSupervisor = [];
}

function changeFilter(item, filters) {
  if (item.Keyword && item.Keyword != "") {
    filters.push({
      Op: "$contains",
      Field: "Name",
      Value: [item.Keyword],
    });
  }
  if (item.Op && item.Op != "") {
    filters.push({
      Op: "$in",
      Field: "Op",
      Value: [item.Op],
    });
  }
  if (item.DirectSupervisor.length > 0) {
    filters.push({
      Op: "$in",
      Field: "DirectSupervisor",
      Value: [...item.DirectSupervisor],
    });
  }
}

function refreshGrid() {
  listControl.value.refreshGrid();
}

function resetGridHeaderFilter() {
  gridHeaderFilter.value.reset();
}
</script>
