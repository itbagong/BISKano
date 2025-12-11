<template>
  <div>
    <data-list
      class="card"
      ref="listControl"
      grid-mode="grid"
      form-keep-label
      :title="data.titleForm"
      grid-hide-delete
      grid-config="/kara/worklocation/gridconfig"
      grid-read="/kara/worklocation/gets"
      form-config="/kara/worklocation/formconfig"
      form-read="/kara/worklocation/get"
      form-insert="/kara/admin/worklocation/create"
      form-update="/kara/admin/worklocation/update"
      stay-on-form-after-save
      :form-tabs-edit="['General', 'User']"
      :form-fields="['Location', 'DistanceTolerance', 'TimeLoc']"
      :grid-fields="['Location']"
      @formNewData="newRecord"
      @formEditData="editRecord"
      @controlModeChanged="onChangeMode"
      :grid-hide-new="!profile.canCreate"
      :grid-hide-edit="!profile.canUpdate" 
    >
    <template #grid_Location="{ item }">
      <div v-if="item?.Location?.Coordinates">{{ item?.Location?.Coordinates.toString() }}</div>
    </template>
      <template #form_input_Location="{ item }">
        <div class="w-full flex flex-col gap-2">
          <div class="w-full flex gap-2">
            <s-input class="w-full" v-model="data.Longitude" label="Longitude" kind="number" @change="(field, v1, v2, old, ctlRef) => {
              data.keyMap = util.uuid();
              if (data.formMode == 'edit') {
                item.Location.Coordinates[0] = v1;
              }
            }" />
            <s-input class="w-full" v-model="data.Latitude" label="Latitude" kind="number" @change="(field, v1, v2, old, ctlRef) => {
              data.keyMap = util.uuid();
              if (data.formMode == 'edit') {
                item.Location.Coordinates[1] = v1;
              }
            }" />
            <!-- <s-input
              label="Google Location Search"
              use-list
              caption="enter a google map location"
              class="w-full"
            /> -->
          </div>
          <div class="border p-2 h-[315px]">
            <Map
              :key="data.keyMap"
              :lat="data.Latitude"
              :long="data.Longitude"
              @on-change-marker="
                (loc) => {
                  item.Longitude = loc.lng;
                  item.Latitude = loc.lat;
                  data.Longitude = loc.lng;
                  data.Latitude = loc.lat;
                  if (data.formMode == 'edit') {
                    item.Location.Coordinates[0] = loc.lng;
                    item.Location.Coordinates[1] = loc.lat;
                  }
                }
              "
            />
          </div>
        </div>
      </template>
      <template #form_input_TimeLoc="{ item }">
        <s-input
          label="Time location"
          v-model="item.TimeLoc"
          class="w-full"
          use-list
          :items="moment.tz.names()"
        ></s-input>
      </template>
      <template #form_input_DistanceTolerance="{ item }">
        <s-input
          label="Distance tolerance (m)"
          v-model="item.DistanceTolerance"
          class="w-full"
          kind="number"
        ></s-input>
      </template>
      <template #form_tab_User="props">
        <UserLocation :work-location-id="props.item._id"></UserLocation>
      </template>
    </data-list>
  </div>
</template>

<script setup>
import { reactive, ref, inject } from "vue";
import { authStore } from "@/stores/auth.js";
import { layoutStore } from "@/stores/layout.js";
import { DataList, util, SInput } from "suimjs";
import UserLocation from "./widget/UserLocation.vue";
import Map from "./widget/Map.vue";
import moment from 'moment-timezone';

layoutStore().name = "tenant";

const FEATUREID = 'WorkLocation'
const profile = authStore().getRBAC(FEATUREID)

const axios = inject("axios");
const listControl = ref(null);
const data = reactive({
  formMode: "edit",
  titleForm: "Work Location",
  record: {
    workLocationID: null,
  },
  Latitude: 0,
  Longitude: 0,
  keyMap: Math.random(),
});
function newRecord(record) {
  data.formMode = "new";
  data.titleForm = `Create New Work Location`;
  record.DistanceTolerance = 0;
  data.record = record;
  data.Latitude = 0
  data.Longitude = 0
  data.record.Latitude = 0
  data.record.Longitude = 0
  openForm(record);
}

function editRecord(record) {
  data.formMode = "edit";
  data.record = record;
  if (record.Location !== null) {
    data.Longitude = record.Location.Coordinates[0]
    data.Latitude = record.Location.Coordinates[1]
  }
  data.titleForm = `Edit Work Location | ${record._id}`;
  openForm(record);
}
function openForm(record) {
  util.nextTickN(2, () => {
    listControl.value.setFormFieldAttr(
      "_id",
      "hide",
      data.formMode == "new" ? true : false
    );
    listControl.value.setFormLoading(false);
  });
}
function onChangeMode(mode) {
  if (mode === "grid") {
    data.titleForm = "Work Location";
  }
}
</script>
